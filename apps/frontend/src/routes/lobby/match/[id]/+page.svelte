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
          // Sort by created_at descending (newest first)
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
        // Also update this match in the sidebar list
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
    if (e.key === 'Enter' && !e.shiftKey) {
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
      case 'won': return '🎉 승리!';
      case 'lost': return '😔 패배';
      case 'resigned': return '🏳️ 기권';
      case 'expired': return '⏰ 만료';
      case 'error': return '⚠️ 오류';
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
</script>

<div class={`flex flex-col h-[calc(100vh-64px)] transition-colors ${isDarkMode ? 'bg-gray-950' : 'bg-gray-50'}`}>

  {#if isLoading}
    <!-- Loading State -->
    <div class="flex-1 flex items-center justify-center">
      <div class="animate-spin rounded-full h-12 w-12 border-4 border-[#FF4D00] border-t-transparent"></div>
    </div>

  {:else if errorMessage && !match}
    <!-- Error State (no match loaded) -->
    <div class="flex-1 flex items-center justify-center p-6">
      <div class={`text-center max-w-md rounded-2xl p-8 border ${isDarkMode ? 'bg-gray-900 border-gray-800' : 'bg-white border-gray-200'}`}>
        <p class="text-red-400 text-lg font-semibold mb-4">{errorMessage}</p>
        <button
          onclick={() => goto('/lobby')}
          class="px-6 py-3 bg-[#FF4D00] text-white rounded-full font-bold hover:bg-[#ff3300] transition-colors"
        >
          로비로 돌아가기
        </button>
      </div>
    </div>

  {:else if match}
    <!-- ============================================================ -->
    <!-- Match Header Bar (full width)                                -->
    <!-- ============================================================ -->
    <div class={`shrink-0 border-b px-4 py-3 md:px-6 md:py-4 flex items-center justify-between ${isDarkMode ? 'bg-gray-900/80 border-gray-800' : 'bg-white border-gray-200'}`}>
      <div class="flex items-center gap-2 min-w-0">
        <button
          onclick={() => goto('/lobby')}
          class={`shrink-0 p-2 rounded-lg transition-colors ${isDarkMode ? 'hover:bg-gray-800 text-gray-400' : 'hover:bg-gray-100 text-gray-600'}`}
          aria-label="뒤로가기"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
          </svg>
        </button>
        <!-- Match list toggle (mobile/tablet) -->
        <button
          onclick={() => showSidebar = !showSidebar}
          class={`shrink-0 p-2 rounded-lg transition-colors xl:hidden ${isDarkMode ? 'hover:bg-gray-800 text-gray-400' : 'hover:bg-gray-100 text-gray-600'}`}
          aria-label="매치 목록 열기"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7"/>
          </svg>
        </button>
        <div class="min-w-0">
          <h1 class={`font-bold text-lg truncate ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>
            {game?.title ?? '게임'}
          </h1>
          <div class="flex items-center gap-2 text-sm">
            <span class={`px-2 py-0.5 rounded-full text-xs font-semibold border ${statusColor}`}>
              {statusLabel}
            </span>
            <span class={isDarkMode ? 'text-gray-500' : 'text-gray-400'}>
              턴 {turnDisplay}
            </span>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-2 shrink-0">
        {#if isMatchActive}
          <button
            onclick={() => showResignModal = true}
            class={`px-3 py-1.5 rounded-lg text-sm font-semibold transition-colors border ${isDarkMode ? 'border-gray-700 text-gray-400 hover:text-red-400 hover:border-red-500/50' : 'border-gray-300 text-gray-500 hover:text-red-500 hover:border-red-300'}`}
          >
            기권
          </button>
        {/if}
      </div>
    </div>

    <!-- ============================================================ -->
    <!-- Content: Chat (centered) + Match Panel (overlaid in margin)  -->
    <!-- ============================================================ -->
    <div class="flex-1 flex flex-col min-h-0">

      <!-- Scrollable zone: chat messages + overlaid match panel -->
      <div class="flex-1 relative min-h-0">

        <!-- Mobile overlay backdrop -->
        {#if showSidebar}
          <div
            class="fixed inset-0 bg-black/50 z-30 xl:hidden"
            onclick={() => showSidebar = false}
            transition:fade={{ duration: 150 }}
            role="presentation"
          ></div>
        {/if}

        <!-- Match panel: overlays the left dead-space, stops above input -->
        <div class={`
          flex flex-col overflow-hidden z-40
          ${isDarkMode ? 'bg-gray-900 xl:bg-transparent' : 'bg-white xl:bg-transparent'}
          fixed xl:absolute left-0 w-72 xl:w-[280px]
          top-[calc(64px+57px)] bottom-0 xl:top-0 xl:bottom-0 xl:left-3
          border-r xl:border-0 ${isDarkMode ? 'border-gray-800' : 'border-gray-200'}
          transition-transform duration-200
          ${showSidebar ? 'translate-x-0' : '-translate-x-full xl:translate-x-0'}
        `}>
          <!-- Panel header -->
          <div class={`shrink-0 px-3 py-2.5 flex items-center justify-between xl:pt-4 ${isDarkMode ? 'border-b xl:border-0 border-gray-800' : 'border-b xl:border-0 border-gray-200'}`}>
            <span class={`text-xs font-semibold ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>
              {siblingMatches.length}개 매치
            </span>
            {#if game}
              <button
                onclick={handleRetry}
                class="shrink-0 w-7 h-7 rounded-lg bg-[#FF4D00] text-white flex items-center justify-center hover:bg-[#ff3300] transition-colors shadow-sm"
                title="새 매치 생성"
                aria-label="새 매치 생성"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                </svg>
              </button>
            {/if}
          </div>

          <!-- Match cards -->
          <nav class="flex-1 overflow-y-auto p-2 xl:p-1.5 space-y-2">
            {#each siblingMatches as sibling (sibling.id)}
              <a
                href="/lobby/match/{sibling.id}"
                class={`group block rounded-2xl border-2 px-4 py-3.5 transition-all cursor-pointer ${
                  sibling.id === matchId
                    ? 'border-[#FF4D00] shadow-lg shadow-orange-500/10 ' + (isDarkMode ? 'bg-gray-800/90' : 'bg-orange-50')
                    : isDarkMode
                      ? 'border-gray-800 hover:border-gray-600 bg-gray-900/80 hover:bg-gray-800/90 hover:shadow-md'
                      : 'border-gray-200 hover:border-gray-300 bg-white hover:bg-gray-50 hover:shadow-md'
                }`}
                onclick={() => showSidebar = false}
              >
                <div class="flex items-center justify-between mb-2">
                  <span class={`text-xs font-mono tracking-wider ${
                    sibling.id === matchId
                      ? 'text-[#FF4D00] font-bold'
                      : isDarkMode ? 'text-gray-400 group-hover:text-gray-300' : 'text-gray-500 group-hover:text-gray-700'
                  }`}>
                    #{sibling.id.slice(-6)}
                  </span>
                  <span class={`text-xs font-bold px-2 py-0.5 rounded-full ${getSidebarStatusBadge(sibling.status)}`}>
                    {getShortStatusLabel(sibling.status)}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class={`text-xs ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
                    턴 {sibling.turn_count}/{sibling.max_turns}
                  </span>
                  <span class={`text-xs ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}>
                    {new Date(sibling.updated_at).toLocaleDateString()}
                  </span>
                </div>
              </a>
            {/each}

            {#if siblingMatches.length === 0 && !isLoading}
              <div class={`px-3 py-6 text-center text-xs ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}>
                매치가 없습니다
              </div>
            {/if}
          </nav>
        </div>

        <!-- Chat Messages Area (scrollable, fills this zone) -->
        <div
          bind:this={chatContainer}
          class="absolute inset-0 overflow-y-auto px-4 py-4 md:px-6 md:py-6 space-y-4"
        >
          <!-- Inline error banner -->
          {#if errorMessage}
            <div class="mx-auto max-w-2xl p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm text-center" transition:fade={{ duration: 200 }}>
              {errorMessage}
            </div>
          {/if}

          {#if visibleMessages.length === 0 && !isSending}
            <!-- Empty state -->
            <div class="flex-1 flex items-center justify-center h-full">
              <div class="text-center">
                <div class={`text-6xl mb-4 ${isDarkMode ? 'opacity-30' : 'opacity-20'}`}>💬</div>
                <p class={`text-lg font-semibold mb-1 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
                  대화를 시작하세요
                </p>
                <p class={`text-sm ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}>
                  AI에게 메시지를 보내 게임을 진행해보세요
                </p>
              </div>
            </div>
          {/if}

          {#each visibleMessages as msg (msg.id)}
            <div
              class="flex {msg.role === 'user' ? 'justify-end' : 'justify-start'} max-w-3xl mx-auto w-full"
              in:fly={{ y: 12, duration: 250 }}
            >
              {#if msg.role === 'assistant'}
                <!-- AI Avatar -->
                <div class="shrink-0 mr-3 mt-1">
                  <div class="w-8 h-8 rounded-full bg-[#FF4D00] flex items-center justify-center shadow-md">
                    <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
                    </svg>
                  </div>
                </div>
              {/if}

              <div class={`max-w-[80%] rounded-2xl px-4 py-3 shadow-sm ${
                msg.role === 'user'
                  ? 'bg-[#FF4D00] text-white rounded-br-md'
                  : isDarkMode
                    ? 'bg-gray-800 text-gray-200 border border-gray-700 rounded-bl-md'
                    : 'bg-white text-gray-800 border border-gray-200 rounded-bl-md'
              }`}>
                <p class="text-sm md:text-base whitespace-pre-wrap leading-relaxed">{msg.content}</p>
                <div class={`text-xs mt-1.5 ${
                  msg.role === 'user'
                    ? 'text-white/60'
                    : isDarkMode ? 'text-gray-500' : 'text-gray-400'
                }`}>
                  턴 {msg.turn_count}
                </div>
              </div>

              {#if msg.role === 'user'}
                <!-- User Avatar -->
                <div class="shrink-0 ml-3 mt-1">
                  <div class={`w-8 h-8 rounded-full flex items-center justify-center shadow-md text-xs font-bold ${isDarkMode ? 'bg-gray-700 text-gray-300' : 'bg-gray-200 text-gray-600'}`}>
                    U
                  </div>
                </div>
              {/if}
            </div>
          {/each}

          <!-- Generating indicator -->
          {#if isSending || isGenerating}
            <div class="flex justify-start max-w-3xl mx-auto w-full" in:fade={{ duration: 200 }}>
              <div class="shrink-0 mr-3 mt-1">
                <div class="w-8 h-8 rounded-full bg-[#FF4D00] flex items-center justify-center shadow-md">
                  <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
                  </svg>
                </div>
              </div>
              <div class={`rounded-2xl rounded-bl-md px-4 py-3 shadow-sm ${isDarkMode ? 'bg-gray-800 border border-gray-700' : 'bg-white border border-gray-200'}`}>
                <div class="flex items-center gap-1.5">
                  <span class="w-2 h-2 rounded-full bg-[#FF4D00] animate-bounce" style="animation-delay: 0ms"></span>
                  <span class="w-2 h-2 rounded-full bg-[#FF4D00] animate-bounce" style="animation-delay: 150ms"></span>
                  <span class="w-2 h-2 rounded-full bg-[#FF4D00] animate-bounce" style="animation-delay: 300ms"></span>
                </div>
              </div>
            </div>
          {/if}
        </div>
        <!-- end chat messages area -->
      </div>
      <!-- end scrollable zone -->

      <!-- Terminal match result banner -->
      {#if isTerminal}
        <div
          class={`shrink-0 px-4 py-4 md:px-6 text-center border-t ${isDarkMode ? 'border-gray-800' : 'border-gray-200'}`}
          in:fly={{ y: 20, duration: 300 }}
        >
          <div class="max-w-3xl mx-auto">
            <p class="text-2xl font-black mb-2">
              {#if match?.status === 'won'}
                🎉 축하합니다! 승리했습니다!
              {:else if match?.status === 'lost'}
                😔 아쉽습니다. 턴이 모두 소진되었습니다.
              {:else if match?.status === 'resigned'}
                🏳️ 기권했습니다.
              {:else if match?.status === 'expired'}
                ⏰ 매치가 만료되었습니다.
              {:else}
                ⚠️ 오류가 발생했습니다.
              {/if}
            </p>
            <p class={`text-sm mb-4 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
              사용 턴: {match?.turn_count} / {match?.max_turns}
            </p>
            <div class="flex justify-center gap-3">
              <button
                onclick={() => goto('/lobby')}
                class={`px-5 py-2.5 rounded-full font-bold text-sm transition-colors ${isDarkMode ? 'bg-gray-800 text-gray-300 hover:bg-gray-700' : 'bg-gray-200 text-gray-700 hover:bg-gray-300'}`}
              >
                로비로 돌아가기
              </button>
              {#if game}
                <button
                  onclick={handleRetry}
                  class="px-5 py-2.5 bg-[#FF4D00] text-white rounded-full font-bold text-sm hover:bg-[#ff3300] transition-colors"
                >
                  다시 도전하기
                </button>
              {/if}
            </div>
          </div>
        </div>
      {/if}

      <!-- Input Area -->
      {#if !isTerminal}
        <div class={`shrink-0 border-t px-4 py-3 md:px-6 md:py-4 ${isDarkMode ? 'bg-gray-900/80 border-gray-800' : 'bg-white border-gray-200'}`}>
          <div class="max-w-3xl mx-auto flex items-end gap-3">
            <div class="flex-1 relative">
              <textarea
                bind:value={inputText}
                onkeydown={handleKeydown}
                placeholder={isMatchActive ? '메시지를 입력하세요...' : 'AI 응답을 기다리는 중...'}
                disabled={!isMatchActive || isSending}
                rows={1}
                class={`w-full resize-none rounded-xl px-4 py-3 text-sm md:text-base outline-none transition-colors border ${
                  isDarkMode
                    ? 'bg-gray-800 border-gray-700 text-gray-200 placeholder-gray-500 focus:border-[#FF4D00]'
                    : 'bg-gray-100 border-gray-200 text-gray-900 placeholder-gray-400 focus:border-[#FF4D00]'
                } disabled:opacity-50 disabled:cursor-not-allowed`}
                style="max-height: 120px"
              ></textarea>
            </div>
            <button
              onclick={handleSendMessage}
              disabled={!inputText.trim() || !isMatchActive || isSending}
              class="shrink-0 w-11 h-11 rounded-xl bg-[#FF4D00] text-white flex items-center justify-center hover:bg-[#ff3300] transition-all active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-[#FF4D00]"
              aria-label="메시지 전송"
            >
              {#if isSending}
                <div class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
              {:else}
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19V5m0 0l-7 7m7-7l7 7"/>
                </svg>
              {/if}
            </button>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Resign Confirmation Modal -->
{#if showResignModal}
  <div
    class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    onclick={() => showResignModal = false}
    onkeydown={(e) => e.key === 'Escape' && (showResignModal = false)}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    transition:fade={{ duration: 200 }}
  >
    <div
      class={`w-full max-w-md rounded-2xl shadow-2xl p-8 border ${isDarkMode ? 'bg-gray-900 border-gray-800' : 'bg-white border-gray-200'}`}
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="presentation"
      transition:scale={{ duration: 200, start: 0.95 }}
    >
      <h2 class={`text-2xl font-bold mb-2 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>기권하시겠습니까?</h2>
      <p class={`text-sm mb-6 ${isDarkMode ? 'text-gray-400' : 'text-gray-600'}`}>
        기권하면 이 매치는 종료됩니다. 이 작업은 되돌릴 수 없습니다.
      </p>
      <div class="flex justify-end gap-3">
        <button
          onclick={() => showResignModal = false}
          class={`px-5 py-2.5 rounded-full font-bold text-sm transition-colors ${isDarkMode ? 'text-gray-300 hover:bg-gray-800' : 'text-gray-600 hover:bg-gray-100'}`}
        >
          취소
        </button>
        <button
          onclick={handleResign}
          class="px-5 py-2.5 bg-red-500 text-white rounded-full font-bold text-sm hover:bg-red-600 transition-colors"
        >
          기권하기
        </button>
      </div>
    </div>
  </div>
{/if}
