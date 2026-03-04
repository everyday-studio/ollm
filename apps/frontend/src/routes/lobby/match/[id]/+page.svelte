<script lang="ts">
  import { fade, fly, scale } from 'svelte/transition';
  import { onMount, getContext } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';

  import { gameApi } from '$lib/features/game/api';
  import { messageApi } from '$lib/features/game/messageApi';
  import { authApi } from '$lib/features/auth/api';
  import { authStore } from '$lib/features/auth/model';
  import type { User } from '$lib/features/auth/types';
  import type { MatchDTO, GameDTO, MessageDTO, MatchStatus } from '$lib/features/game/types';

  const theme = getContext<{ isDark: boolean }>('theme');
  let isDarkMode = $derived(theme.isDark);

  // ----------------------------------------------------------------
  // Route param (reactive — updates on navigation within same component)
  // ----------------------------------------------------------------
  let matchId = $derived($page.params.id);

  // ----------------------------------------------------------------
  // State
  // ----------------------------------------------------------------
  let match = $state<MatchDTO | null>(null);
  let game = $state<GameDTO | null>(null);
  let messages = $state<MessageDTO[]>([]);
  let siblingMatches = $state<MatchDTO[]>([]);
  let inputText = $state('');
  let isLoading = $state(true);
  let isSending = $state(false);
  let errorMessage = $state('');
  let showResignModal = $state(false);
  let showSidebar = $state(false);
  let sessionRestored = $state(false);

  // Derived helpers
  let visibleMessages = $derived(messages.filter(m => m.is_visible));
  let isMatchActive = $derived(match?.status === 'active');
  let isGenerating = $derived(match?.status === 'generating');
  let isTerminal = $derived(
    match?.status === 'won' ||
    match?.status === 'lost' ||
    match?.status === 'resigned' ||
    match?.status === 'expired' ||
    match?.status === 'error'
  );
  let turnDisplay = $derived(
    match ? `${match.turn_count} / ${match.max_turns}` : '— / —'
  );
  let statusLabel = $derived(getStatusLabel(match?.status));
  let statusColor = $derived(getStatusColor(match?.status));

  // ----------------------------------------------------------------
  // Chat container ref for auto-scroll
  // ----------------------------------------------------------------
  let chatContainer = $state<HTMLDivElement | null>(null);

  function scrollToBottom() {
    const el = chatContainer;
    if (el) {
      requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight;
      });
    }
  }

  // ----------------------------------------------------------------
  // Lifecycle
  // ----------------------------------------------------------------
  onMount(async () => {
    // Restore session
    try {
      const refreshRes = await authApi.refresh();
      if (refreshRes?.data) {
        const { access_token, id, name, email } = refreshRes.data as any;
        if (access_token && email) {
          const user: User = {
            id: id || '',
            name: name || 'Player',
            email,
            role: 'USER',
            created_at: new Date().toISOString()
          };
          authStore.loginSuccess(access_token, user);
        }
      }
    } catch {
      console.warn('Failed to restore session');
    }
    sessionRestored = true;
  });

  // Reload match data whenever matchId changes (reactive navigation)
  $effect(() => {
    if (sessionRestored && matchId) {
      loadMatchData(matchId);
    }
  });

  async function loadMatchData(id: string) {
    isLoading = true;
    errorMessage = '';

    try {
      // Fetch match and messages in parallel
      const [matchRes, messagesRes] = await Promise.all([
        gameApi.getMatchById(id),
        messageApi.getHistory(id)
      ]);

      match = matchRes.data;
      messages = messagesRes.data ?? [];

      // Fetch game info and sibling matches
      if (match) {
        try {
          const [gameRes, siblingsRes] = await Promise.all([
            gameApi.getGameById(match.game_id),
            gameApi.getMyMatchesByGame(match.game_id)
          ]);
          game = gameRes.data;
          siblingMatches = (siblingsRes.data ?? []).sort(
            (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
          );
        } catch {
          console.warn('Failed to fetch game info or sibling matches');
        }
      }

      scrollToBottom();
    } catch (e: any) {
      const status = e?.response?.status;
      if (status === 403) {
        errorMessage = '이 매치에 접근할 권한이 없습니다.';
      } else if (status === 404) {
        errorMessage = '매치를 찾을 수 없습니다.';
      } else {
        errorMessage = '매치 정보를 불러오는 데 실패했습니다.';
      }
    } finally {
      isLoading = false;
    }
  }

  // ----------------------------------------------------------------
  // Send message
  // ----------------------------------------------------------------
  async function handleSendMessage() {
    const currentMatchId = matchId;
    if (!currentMatchId || !inputText.trim() || isSending || !isMatchActive) return;

    const userContent = inputText.trim();
    inputText = '';
    isSending = true;
    errorMessage = '';

    // Optimistic: add user message to UI immediately
    const optimisticUserMsg: MessageDTO = {
      id: `temp-${Date.now()}`,
      match_id: currentMatchId,
      role: 'user',
      content: userContent,
      is_visible: true,
      turn_count: (match?.turn_count ?? 0) + 1,
      token_count: 0,
      created_at: new Date().toISOString()
    };
    messages = [...messages, optimisticUserMsg];

    // Update local match status to generating
    if (match) {
      match = { ...match, status: 'generating' as MatchStatus };
    }

    scrollToBottom();

    try {
      const res = await messageApi.sendMessage(currentMatchId, { content: userContent });
      const aiMessage = res.data;

      messages = [...messages, aiMessage];

      // Refresh match state to get updated status/turn_count
      try {
        const matchRes = await gameApi.getMatchById(currentMatchId);
        match = matchRes.data;
        siblingMatches = siblingMatches.map(m => m.id === currentMatchId ? matchRes.data : m);
      } catch {
        if (match) {
          match = {
            ...match,
            turn_count: aiMessage.turn_count,
            status: 'active' as MatchStatus
          };
        }
      }

      scrollToBottom();
    } catch (e: any) {
      const status = e?.response?.status;
      if (status === 409) {
        errorMessage = '매치가 이미 종료되었거나 AI가 응답 중입니다.';
        try {
          const matchRes = await gameApi.getMatchById(currentMatchId);
          match = matchRes.data;
        } catch { /* ignore */ }
      } else if (status === 403) {
        errorMessage = '이 매치에 접근할 권한이 없습니다.';
      } else {
        errorMessage = '메시지 전송에 실패했습니다. 다시 시도해주세요.';
      }
      messages = messages.filter(m => m.id !== optimisticUserMsg.id);

      if (match && match.status === 'generating') {
        match = { ...match, status: 'active' as MatchStatus };
      }
    } finally {
      isSending = false;
    }
  }

  // ----------------------------------------------------------------
  // Resign
  // ----------------------------------------------------------------
  async function handleResign() {
    showResignModal = false;
    const currentMatchId = matchId;
    if (!currentMatchId) return;
    try {
      await gameApi.resignMatch(currentMatchId);
      const matchRes = await gameApi.getMatchById(currentMatchId);
      match = matchRes.data;
      siblingMatches = siblingMatches.map(m => m.id === currentMatchId ? matchRes.data : m);
    } catch (e: any) {
      const status = e?.response?.status;
      if (status === 409) {
        errorMessage = '이미 종료된 매치입니다.';
      } else {
        errorMessage = '기권 처리에 실패했습니다.';
      }
    }
  }

  // ----------------------------------------------------------------
  // Create new match and navigate
  // ----------------------------------------------------------------
  async function handleRetry() {
    if (!game) return;
    try {
      const res = await gameApi.createMatch(game.id);
      // Navigate to the new match (triggers $effect reload)
      await goto(`/lobby/match/${res.data.id}`);
    } catch {
      errorMessage = '새 매치 생성에 실패했습니다.';
    }
  }

  // ----------------------------------------------------------------
  // Keyboard shortcut: Enter to send, Shift+Enter for newline
  // ----------------------------------------------------------------
  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
      e.preventDefault();
      handleSendMessage();
    }
  }

  // ----------------------------------------------------------------
  // Status helpers
  // ----------------------------------------------------------------
  function getStatusLabel(status: MatchStatus | undefined): string {
    switch (status) {
      case 'active': return '진행 중';
      case 'generating': return 'AI 응답 중...';
      case 'won': return '승리';
      case 'lost': return '패배';
      case 'resigned': return '기권';
      case 'expired': return '만료';
      case 'error': return '오류';
      default: return '—';
    }
  }

  function getStatusColor(status: MatchStatus | undefined): string {
    switch (status) {
      case 'active': return 'bg-green-500/20 text-green-400 border-green-500/30';
      case 'generating': return 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30';
      case 'won': return 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30';
      case 'lost': return 'bg-red-500/20 text-red-400 border-red-500/30';
      case 'resigned': return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
      case 'expired': return 'bg-orange-500/20 text-orange-400 border-orange-500/30';
      case 'error': return 'bg-red-500/20 text-red-400 border-red-500/30';
      default: return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
    }
  }

  function getSidebarStatusBadge(status: MatchStatus): string {
    switch (status) {
      case 'active': return 'bg-green-500/20 text-green-400';
      case 'generating': return 'bg-yellow-500/20 text-yellow-400';
      case 'won': return 'bg-emerald-500/20 text-emerald-300';
      case 'lost': return 'bg-red-500/20 text-red-400';
      case 'resigned': return 'bg-gray-500/20 text-gray-400';
      default: return 'bg-gray-500/20 text-gray-400';
    }
  }

  function getShortStatusLabel(status: MatchStatus): string {
    switch (status) {
      case 'active': return '진행 중';
      case 'generating': return '생성 중';
      case 'won': return '승리';
      case 'lost': return '패배';
      case 'resigned': return '기권';
      case 'expired': return '만료';
      case 'error': return '오류';
      default: return status;
    }
  }

  // ----------------------------------------------------------------
  // Relative time (Instagram DM style)
  // ----------------------------------------------------------------
  function getRelativeTime(dateStr: string): string {
    const now = Date.now();
    const created = new Date(dateStr).getTime();
    const diffMs = now - created;
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);
    const diffWeek = Math.floor(diffDay / 7);
    const diffMonth = Math.floor(diffDay / 30);
    const diffYear = Math.floor(diffDay / 365);

    if (diffSec < 60) return '방금';
    if (diffMin < 60) return `${diffMin}분 전`;
    if (diffHour < 24) return `${diffHour}시간 전`;
    if (diffDay === 1) return '어제';
    if (diffDay < 7) return `${diffDay}일 전`;
    if (diffWeek === 1) return '1주 전';
    if (diffWeek === 2) return '2주 전';
    if (diffWeek === 3) return '3주 전';
    if (diffMonth < 12) return `${Math.max(1, diffMonth)}달 전`;
    return `${diffYear}년 전`;
  }
</script>

<!-- ============================================================== -->
<!-- TEMPLATE                                                        -->
<!-- ============================================================== -->
<div class={`flex flex-col h-[calc(100vh-64px)] transition-colors ${isDarkMode ? 'bg-gray-950' : 'bg-gray-50'}`}>
  <!-- Same container as lobby: max-w + lobby margins -->
  <div class="max-w-[1800px] mx-auto w-full flex-1 flex flex-col min-h-0 px-4 md:px-8 lg:px-10 py-4">

  {#if isLoading}
    <!-- ======================== Loading ======================== -->
    <div class="flex-1 flex items-center justify-center">
      <div class="flex flex-col items-center gap-3">
        <div class="w-10 h-10 border-[3px] border-[#FF4D00]/30 border-t-[#FF4D00] rounded-full animate-spin"></div>
        <span class={`text-sm ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>불러오는 중…</span>
      </div>
    </div>

  {:else if errorMessage && !match}
    <!-- ======================== Error ======================== -->
    <div class="flex-1 flex items-center justify-center p-6">
      <div class="text-center max-w-sm">
        <div class={`w-14 h-14 mx-auto mb-4 rounded-full flex items-center justify-center ${isDarkMode ? 'bg-red-500/10' : 'bg-red-50'}`}>
          <svg class="w-7 h-7 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z"/>
          </svg>
        </div>
        <p class={`font-semibold mb-2 ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}>{errorMessage}</p>
        <button
          onclick={() => goto('/lobby')}
          class="mt-2 px-5 py-2.5 bg-[#FF4D00] text-white rounded-lg font-semibold text-sm hover:bg-[#ff3300] transition-colors"
        >
          로비로 돌아가기
        </button>
      </div>
    </div>

  {:else if match}
    <!-- Chat column: full-width -->
    <div class="flex-1 flex flex-col min-h-0">

    <!-- ==================== Chat area (full-width, centered) ==================== -->

      <!-- Top bar -->
      <div class={`shrink-0 flex items-center justify-between py-2.5 border-b transition-colors ${
        isDarkMode ? 'border-gray-800/70' : 'border-gray-200'
      }`}>
        <!-- Left group -->
        <div class="flex items-center gap-2 min-w-0">
          <!-- Mobile only: back + hamburger -->
          <button
            onclick={() => goto('/lobby')}
            class={`lg:hidden shrink-0 p-1.5 rounded-lg transition-colors ${
              isDarkMode ? 'text-gray-400 hover:bg-gray-800' : 'text-gray-500 hover:bg-gray-100'
            }`}
            aria-label="뒤로가기"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
            </svg>
          </button>
          <button
            onclick={() => showSidebar = !showSidebar}
            class={`lg:hidden shrink-0 p-1.5 rounded-lg transition-colors ${
              isDarkMode ? 'text-gray-400 hover:bg-gray-800' : 'text-gray-500 hover:bg-gray-100'
            }`}
            aria-label="매치 목록"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7"/>
            </svg>
          </button>

          <!-- Mobile: game title -->
          <span class={`lg:hidden font-semibold text-sm truncate ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}>
            {game?.title ?? '게임'}
          </span>

          <!-- Status badge -->
          <div class="flex items-center gap-2.5">
            <span class={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-semibold border ${statusColor}`}>
              {#if match?.status === 'generating'}
                <span class="w-1.5 h-1.5 rounded-full bg-yellow-400 animate-pulse"></span>
              {:else if match?.status === 'active'}
                <span class="w-1.5 h-1.5 rounded-full bg-green-400"></span>
              {/if}
              {statusLabel}
            </span>
            <span class={`text-xs tabular-nums font-medium hidden sm:inline ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
              턴 {turnDisplay}
            </span>
          </div>
        </div>

        <!-- Right group -->
        <div class="flex items-center gap-2 shrink-0">
          <span class={`sm:hidden text-[11px] tabular-nums ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
            {turnDisplay}
          </span>
          {#if isMatchActive}
            <button
              onclick={() => showResignModal = true}
              class={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors ${
                isDarkMode
                  ? 'text-gray-500 hover:text-red-400 hover:bg-red-500/10'
                  : 'text-gray-400 hover:text-red-500 hover:bg-red-50'
              }`}
            >
              기권
            </button>
          {/if}
        </div>
      </div>

      <!-- Body row: sidebar + bordered chat column -->
      <div class="flex-1 flex min-h-0">

        <!-- ===== Left sidebar (in-flow) ===== -->
        <aside class={`hidden lg:flex flex-col w-[272px] shrink-0 transition-colors ${
          isDarkMode ? 'bg-gray-900' : 'bg-white'
        }`}>
          <!-- Back link -->
          <div class={`shrink-0 px-4 pt-4 pb-2`}>
            <button
              onclick={() => goto('/lobby')}
              class={`inline-flex items-center gap-1 text-[11px] font-medium transition-colors ${
                isDarkMode ? 'text-gray-500 hover:text-gray-300' : 'text-gray-400 hover:text-gray-600'
              }`}
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M15 19l-7-7 7-7"/>
              </svg>
              로비
            </button>
          </div>

          <!-- Match list -->
          <nav class="flex-1 overflow-y-auto py-1 pr-0 scrollbar-hide">
            {#each siblingMatches as sibling (sibling.id)}
              {@const isActive = sibling.id === matchId}
              {@const progress = sibling.max_turns > 0 ? (sibling.turn_count / sibling.max_turns) * 100 : 0}
              <a
                href="/lobby/match/{sibling.id}"
                class={`match-tab group relative flex items-center gap-3 ml-2 my-1.5 px-3 py-3 text-sm transition-all ${
                  isActive
                    ? 'match-tab-active rounded-l-xl border border-r-0 border-[#FF4D00]/40 ' + (isDarkMode
                        ? 'bg-gray-950'
                        : 'bg-gray-50')
                    : 'rounded-xl mr-2 border ' + (isDarkMode
                        ? 'border-gray-800 hover:bg-gray-800/50 hover:border-gray-700'
                        : 'border-gray-200 hover:bg-gray-50 hover:border-gray-300')
                }`}
                data-dark={isDarkMode ? '' : undefined}
              >
                {#if isActive}
                  <span class="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-r-full bg-[#FF4D00]"></span>
                {/if}
                <span class={`shrink-0 w-2 h-2 rounded-full ${
                  sibling.status === 'active' ? 'bg-green-400' :
                  sibling.status === 'generating' ? 'bg-yellow-400 animate-pulse' :
                  sibling.status === 'won' ? 'bg-emerald-400' :
                  sibling.status === 'lost' ? 'bg-red-400' :
                  sibling.status === 'resigned' ? 'bg-gray-400' : 'bg-gray-500'
                }`}></span>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between mb-1">
                    <span class={`text-xs font-semibold ${
                      isActive ? 'text-[#FF4D00]' : isDarkMode ? 'text-gray-300' : 'text-gray-700'
                    }`}>{getRelativeTime(sibling.created_at)}</span>
                    <span class={`text-[10px] font-medium ${
                      isActive ? 'text-[#FF4D00]/70' : isDarkMode ? 'text-gray-600' : 'text-gray-400'
                    }`}>{getShortStatusLabel(sibling.status)}</span>
                  </div>
                  <div class={`h-1 rounded-full overflow-hidden ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}>
                    <div
                      class={`h-full rounded-full transition-all duration-300 ${
                        sibling.status === 'won' ? 'bg-emerald-400' :
                        sibling.status === 'lost' || sibling.status === 'resigned' ? isDarkMode ? 'bg-gray-600' : 'bg-gray-400' :
                        'bg-[#FF4D00]'
                      }`}
                      style="width: {progress}%"
                    ></div>
                  </div>
                  <div class={`text-[10px] mt-0.5 tabular-nums ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}>
                    {sibling.turn_count} / {sibling.max_turns} 턴
                  </div>
                </div>
              </a>
            {/each}
          </nav>

          <!-- New match button -->
          {#if game}
            <div class={`shrink-0 px-3 py-3 border-t ${isDarkMode ? 'border-gray-800/70' : 'border-gray-100'}`}>
              <button
                onclick={handleRetry}
                class={`w-full flex items-center justify-center gap-2 py-2.5 rounded-xl text-sm font-semibold transition-all ${
                  isDarkMode
                    ? 'text-gray-400 hover:bg-gray-800 hover:text-gray-200'
                    : 'text-gray-500 hover:bg-gray-100 hover:text-gray-700'
                }`}
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                </svg>
                새 매치
              </button>
            </div>
          {/if}
        </aside>
        <!-- ===== End sidebar ===== -->

        <!-- Chat column: orange border, right-rounded only on desktop -->
        <div class="flex-1 flex flex-col min-h-0 border border-[#FF4D00]/50 rounded-xl lg:rounded-l-none overflow-hidden">
          <div
            bind:this={chatContainer}
            class="flex-1 overflow-y-auto min-h-0"
          >
          {#key matchId}
          <div class="max-w-2xl mx-auto px-4 py-6 md:px-8 space-y-5" in:fade={{ duration: 150, delay: 50 }}>

            <!-- Error banner -->
            {#if errorMessage}
              <div
                class="p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm text-center"
                transition:fade={{ duration: 200 }}
              >
                {errorMessage}
              </div>
            {/if}

            <!-- Empty state -->
            {#if visibleMessages.length === 0 && !isSending}
              <div class="flex flex-col items-center justify-center py-24">
                <div class={`w-20 h-20 mb-6 rounded-2xl flex items-center justify-center ${
                  isDarkMode ? 'bg-gray-800/40' : 'bg-gray-100'
                }`}>
                  <svg class={`w-10 h-10 ${isDarkMode ? 'text-gray-700' : 'text-gray-300'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
                  </svg>
                </div>
                <p class={`font-semibold text-base ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>
                  대화를 시작하세요
                </p>
                <p class={`text-sm mt-1 ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}>
                  첫 메시지를 보내 게임을 시작해보세요
                </p>
              </div>
            {/if}

            <!-- Message bubbles -->
            {#each visibleMessages as msg (msg.id)}
              <div in:fly={{ y: 6, duration: 180 }}>
                {#if msg.role === 'user'}
                  <!-- User -->
                  <div class="flex justify-end">
                    <div class="max-w-[80%] md:max-w-[70%]">
                      <div class="bg-[#FF4D00] text-white rounded-2xl rounded-br-md px-4 py-2.5 shadow-sm shadow-orange-500/10">
                        <p class="text-[15px] leading-relaxed whitespace-normal break-words">{msg.content}</p>
                      </div>
                      <div class="flex justify-end mt-1 pr-1">
                        <span class={`text-[10px] tabular-nums ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}>
                          턴 {msg.turn_count}
                        </span>
                      </div>
                    </div>
                  </div>
                {:else}
                  <!-- AI -->
                  <div class="flex gap-2.5">
                    <div class={`shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-[10px] font-black tracking-tight ${
                      isDarkMode
                        ? 'bg-gradient-to-br from-gray-700 to-gray-800 text-orange-400 ring-1 ring-gray-700'
                        : 'bg-gradient-to-br from-gray-100 to-gray-200 text-orange-500 ring-1 ring-gray-200'
                    }`}>AI</div>
                    <div class="max-w-[80%] md:max-w-[70%]">
                      <div class={`rounded-2xl rounded-tl-md px-4 py-2.5 ${
                        isDarkMode
                          ? 'bg-gray-800/60 text-gray-200 ring-1 ring-gray-800'
                          : 'bg-white text-gray-800 shadow-sm ring-1 ring-gray-100'
                      }`}>
                        <p class="text-[15px] leading-relaxed whitespace-normal break-words">{msg.content}</p>
                      </div>
                      <div class="mt-1 pl-1">
                        <span class={`text-[10px] tabular-nums ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}>
                          턴 {msg.turn_count}
                        </span>
                      </div>
                    </div>
                  </div>
                {/if}
              </div>
            {/each}

            <!-- Generating indicator -->
            {#if isSending || isGenerating}
              <div in:fade={{ duration: 200 }}>
                <div class="flex gap-2.5">
                  <div class={`shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-[10px] font-black tracking-tight ${
                    isDarkMode
                      ? 'bg-gradient-to-br from-gray-700 to-gray-800 text-orange-400 ring-1 ring-gray-700'
                      : 'bg-gradient-to-br from-gray-100 to-gray-200 text-orange-500 ring-1 ring-gray-200'
                  }`}>AI</div>
                  <div class={`rounded-2xl rounded-tl-md px-4 py-3 ${
                    isDarkMode ? 'bg-gray-800/60 ring-1 ring-gray-800' : 'bg-white shadow-sm ring-1 ring-gray-100'
                  }`}>
                    <div class="flex items-center gap-1.5">
                      <span class="w-1.5 h-1.5 rounded-full bg-[#FF4D00] animate-bounce" style="animation-delay: 0ms"></span>
                      <span class="w-1.5 h-1.5 rounded-full bg-[#FF4D00] animate-bounce" style="animation-delay: 150ms"></span>
                      <span class="w-1.5 h-1.5 rounded-full bg-[#FF4D00] animate-bounce" style="animation-delay: 300ms"></span>
                    </div>
                  </div>
                </div>
              </div>
            {/if}

            <!-- Terminal result (inline in chat flow) -->
            {#if isTerminal}
              <div
                class={`rounded-2xl p-6 text-center ring-1 ${
                  match?.status === 'won'
                    ? isDarkMode ? 'bg-emerald-500/5 ring-emerald-500/20' : 'bg-emerald-50 ring-emerald-200'
                    : match?.status === 'lost'
                      ? isDarkMode ? 'bg-red-500/5 ring-red-500/20' : 'bg-red-50 ring-red-200'
                      : isDarkMode ? 'bg-gray-800/30 ring-gray-700/50' : 'bg-gray-50 ring-gray-200'
                }`}
                in:fly={{ y: 12, duration: 300 }}
              >
                <div class="text-4xl mb-3">
                  {#if match?.status === 'won'}🎉{:else if match?.status === 'lost'}💀{:else if match?.status === 'resigned'}🏳️{:else}⚠️{/if}
                </div>
                <h3 class={`text-xl font-bold mb-1 ${
                  match?.status === 'won' ? 'text-emerald-400' :
                  match?.status === 'lost' ? isDarkMode ? 'text-red-400' : 'text-red-600' :
                  isDarkMode ? 'text-gray-300' : 'text-gray-700'
                }`}>
                  {#if match?.status === 'won'}클리어!
                  {:else if match?.status === 'lost'}게임 오버
                  {:else if match?.status === 'resigned'}기권
                  {:else if match?.status === 'expired'}만료
                  {:else}오류
                  {/if}
                </h3>
                <p class={`text-sm mb-5 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
                  {#if match?.status === 'won'}{match?.turn_count}턴 만에 클리어
                  {:else if match?.status === 'lost'}{match?.max_turns}턴을 모두 소진했습니다
                  {:else}턴 {match?.turn_count} / {match?.max_turns}
                  {/if}
                </p>
                <div class="flex items-center justify-center gap-2.5">
                  <button
                    onclick={() => goto('/lobby')}
                    class={`px-5 py-2 rounded-xl text-sm font-semibold transition-colors ${
                      isDarkMode ? 'bg-gray-700 text-gray-300 hover:bg-gray-600' : 'bg-gray-200 text-gray-600 hover:bg-gray-300'
                    }`}
                  >
                    로비
                  </button>
                  {#if game}
                    <button
                      onclick={handleRetry}
                      class="px-5 py-2 bg-[#FF4D00] text-white rounded-xl text-sm font-semibold hover:bg-[#ff3300] transition-colors shadow-sm shadow-orange-500/20"
                    >
                      다시 도전
                    </button>
                  {/if}
                </div>
              </div>
            {/if}
          </div>
          {/key}
          </div>

          <!-- Input area (inside bordered chat column) -->
          {#if !isTerminal}
            <div class={`shrink-0 border-t py-3 transition-colors ${
              isDarkMode ? 'border-gray-800/70' : 'border-gray-200'
            }`}>
              <div class="max-w-2xl mx-auto">
                <div class={`flex items-end gap-2 rounded-2xl px-2 py-1.5 transition-colors ${
                  isMatchActive
                    ? (isDarkMode ? 'bg-gray-800/50 ring-1 ring-gray-700/60 focus-within:ring-[#FF4D00]/40' : 'bg-gray-100 ring-1 ring-gray-200 focus-within:ring-[#FF4D00]/30')
                    : (isDarkMode ? 'bg-gray-800/30' : 'bg-gray-100')
                }`}>
                  <textarea
                    bind:value={inputText}
                    onkeydown={handleKeydown}
                    placeholder={isMatchActive ? '메시지를 입력하세요…' : 'AI 응답을 기다리는 중…'}
                    disabled={!isMatchActive || isSending}
                    rows={1}
                    class={`flex-1 resize-none rounded-xl px-3 py-2.5 text-sm md:text-[15px] outline-none bg-transparent transition-colors ${
                      isDarkMode ? 'text-gray-200 placeholder-gray-500' : 'text-gray-900 placeholder-gray-400'
                    } disabled:opacity-50 disabled:cursor-not-allowed`}
                    style="max-height: 120px"
                  ></textarea>
                  <button
                    onclick={handleSendMessage}
                    disabled={!inputText.trim() || !isMatchActive || isSending}
                    class="shrink-0 w-9 h-9 mb-0.5 rounded-xl bg-[#FF4D00] text-white flex items-center justify-center hover:bg-[#ff3300] transition-all active:scale-95 disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-[#FF4D00]"
                    aria-label="전송"
                  >
                    {#if isSending}
                      <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                    {:else}
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4.5 10.5L12 3m0 0l7.5 7.5M12 3v18"/>
                      </svg>
                    {/if}
                  </button>
                </div>
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
  </div>
</div>

<!-- ==================== Mobile sidebar ==================== -->
{#if showSidebar}
  <div
    class="fixed inset-0 bg-black/40 backdrop-blur-sm z-30 lg:hidden"
    onclick={() => showSidebar = false}
    transition:fade={{ duration: 150 }}
    role="presentation"
  ></div>
  <aside
    class={`fixed top-16 left-0 bottom-0 w-[280px] z-40 flex flex-col lg:hidden shadow-2xl ${
      isDarkMode ? 'bg-gray-900' : 'bg-white'
    }`}
    transition:fly={{ x: -280, duration: 200 }}
  >
    <!-- Header -->
    <div class={`shrink-0 px-4 pt-4 pb-3 border-b ${isDarkMode ? 'border-gray-800' : 'border-gray-200'}`}>
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <h2 class={`font-bold text-sm truncate ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>
            {game?.title ?? '게임'}
          </h2>
          <p class={`text-xs mt-0.5 line-clamp-1 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
            {game?.description ?? ''}
          </p>
        </div>
        <button
          onclick={() => showSidebar = false}
          class={`shrink-0 p-1 rounded-lg ${isDarkMode ? 'text-gray-500 hover:bg-gray-800' : 'text-gray-400 hover:bg-gray-100'}`}
          aria-label="닫기"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Match list -->
    <nav class="flex-1 overflow-y-auto py-1.5 scrollbar-hide">
      {#each siblingMatches as sibling (sibling.id)}
        {@const isActive = sibling.id === matchId}
        {@const progress = sibling.max_turns > 0 ? (sibling.turn_count / sibling.max_turns) * 100 : 0}
        <a
          href="/lobby/match/{sibling.id}"
          onclick={() => showSidebar = false}
          class={`group relative flex items-center gap-3 mx-2 my-0.5 px-3 py-3 rounded-xl text-sm transition-all ${
            isActive
              ? 'border border-[#FF4D00]/40 ' + (isDarkMode ? 'bg-[#FF4D00]/10' : 'bg-[#FF4D00]/5')
              : isDarkMode ? 'hover:bg-gray-800/50' : 'hover:bg-gray-50'
          }`}
        >
          {#if isActive}
            <span class="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-r-full bg-[#FF4D00]"></span>
          {/if}

          <span class={`shrink-0 w-2 h-2 rounded-full ${
            sibling.status === 'active' ? 'bg-green-400' :
            sibling.status === 'generating' ? 'bg-yellow-400 animate-pulse' :
            sibling.status === 'won' ? 'bg-emerald-400' :
            sibling.status === 'lost' ? 'bg-red-400' :
            sibling.status === 'resigned' ? 'bg-gray-400' :
            'bg-gray-500'
          }`}></span>

          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between mb-1">
              <span class={`text-xs font-semibold ${
                isActive ? 'text-[#FF4D00]' : isDarkMode ? 'text-gray-300' : 'text-gray-700'
              }`}>{getRelativeTime(sibling.created_at)}</span>
              <span class={`text-[10px] font-medium ${
                isActive ? 'text-[#FF4D00]/70' : isDarkMode ? 'text-gray-600' : 'text-gray-400'
              }`}>{getShortStatusLabel(sibling.status)}</span>
            </div>
            <div class={`h-1 rounded-full overflow-hidden ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}>
              <div
                class={`h-full rounded-full transition-all duration-300 ${
                  sibling.status === 'won' ? 'bg-emerald-400' :
                  sibling.status === 'lost' || sibling.status === 'resigned' ? isDarkMode ? 'bg-gray-600' : 'bg-gray-400' :
                  'bg-[#FF4D00]'
                }`}
                style="width: {progress}%"
              ></div>
            </div>
            <div class={`text-[10px] mt-0.5 tabular-nums ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}>
              {sibling.turn_count} / {sibling.max_turns} 턴
            </div>
          </div>
        </a>
      {/each}
    </nav>

    <!-- New match button -->
    {#if game}
      <div class={`shrink-0 px-3 py-3 border-t ${isDarkMode ? 'border-gray-800' : 'border-gray-200'}`}>
        <button
          onclick={() => { handleRetry(); showSidebar = false; }}
          class={`w-full flex items-center justify-center gap-2 py-2.5 rounded-xl text-sm font-semibold transition-all ${
            isDarkMode ? 'text-gray-400 hover:bg-gray-800 hover:text-gray-200' : 'text-gray-500 hover:bg-gray-100 hover:text-gray-700'
          }`}
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
          </svg>
          새 매치
        </button>
      </div>
    {/if}
  </aside>
{/if}

<!-- ==================== Resign modal ==================== -->
{#if showResignModal}
  <div
    class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    onclick={() => showResignModal = false}
    onkeydown={(e) => e.key === 'Escape' && (showResignModal = false)}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    transition:fade={{ duration: 200 }}
  >
    <div
      class={`w-full max-w-sm rounded-2xl shadow-2xl p-7 ring-1 ${
        isDarkMode ? 'bg-gray-900 ring-gray-800' : 'bg-white ring-gray-200'
      }`}
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="presentation"
      transition:scale={{ duration: 200, start: 0.95 }}
    >
      <h2 class={`text-xl font-bold mb-2 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>
        기권하시겠습니까?
      </h2>
      <p class={`text-sm mb-6 leading-relaxed ${isDarkMode ? 'text-gray-400' : 'text-gray-600'}`}>
        기권하면 이 매치는 종료됩니다. 되돌릴 수 없습니다.
      </p>
      <div class="flex justify-end gap-2.5">
        <button
          onclick={() => showResignModal = false}
          class={`px-5 py-2.5 rounded-xl font-semibold text-sm transition-colors ${
            isDarkMode ? 'text-gray-300 hover:bg-gray-800' : 'text-gray-600 hover:bg-gray-100'
          }`}
        >
          취소
        </button>
        <button
          onclick={handleResign}
          class="px-5 py-2.5 bg-red-500 text-white rounded-xl font-semibold text-sm hover:bg-red-600 transition-colors"
        >
          기권하기
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* ===== Active match tab: right edge connects to chat column border ===== */
  .match-tab-active {
    position: relative;
    z-index: 1;
    margin-right: 0;
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
  }
</style>
