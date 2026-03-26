<script lang="ts">
  import { fly } from 'svelte/transition';
  import { onMount, getContext } from 'svelte';

  import { gameApi } from '$lib/features/game/api';
  import { loadMockGames } from '$lib/features/game/mockData';
  import { ensureSession } from '$lib/features/auth/session';
  import { getCachedGames } from '$lib/cache/gameCache';
  import { DEFAULT_USER_PROFILE } from '$lib/utils/imageFallback';

  function profileFallback(img: HTMLImageElement) {
    const handler = () => {
      img.onerror = null;
      img.src = DEFAULT_USER_PROFILE;
    };
    img.addEventListener('error', handler);
    return { destroy() { img.removeEventListener('error', handler); } };
  }
  import type { GameDTO, LeaderboardEntry } from '$lib/features/game/types';
  import { disassemble } from 'es-hangul';

  const theme = getContext<{ isDark: boolean }>('theme');
  let isDarkMode = $derived(theme.isDark);

  /** Korean-aware search: regular includes + jamo decomposition matching */
  function matchesQuery(title: string, query: string): boolean {
    const q = query.trim();
    if (!q) return true;
    if (title.toLowerCase().includes(q.toLowerCase())) return true;
    // Decompose both to jamo (ㄷㅣㅂㅅㅣㅋㅡ) for partial syllable matching
    if (disassemble(title).includes(disassemble(q))) return true;
    return false;
  }

  // ----------------------------------------------------------------
  // State
  // ----------------------------------------------------------------
  let games = $state<GameDTO[]>([]);
  let selectedGameId = $state<string>('');
  let entries = $state<LeaderboardEntry[]>([]);
  let isLoading = $state(true);
  let isTableLoading = $state(false);
  let showTokenInfo = $state(false);


  // Combobox state
  let searchQuery = $state('');
  let isComboOpen = $state(false);
  let highlightedIndex = $state(-1);

  let filteredGames = $derived(
    searchQuery.trim() === ''
      ? games
      : games.filter(g => matchesQuery(g.title, searchQuery.trim()))
  );

  let selectedGameLabel = $derived(
    games.find(g => g.id === selectedGameId)?.title ?? '게임 선택'
  );

  function selectGame(id: string) {
    selectedGameId = id;
    searchQuery = '';
    isComboOpen = false;
    highlightedIndex = -1;
  }

  function handleComboKeydown(e: KeyboardEvent) {
    if (!isComboOpen) {
      if (e.key === 'ArrowDown' || e.key === 'Enter') {
        isComboOpen = true;
        e.preventDefault();
      }
      return;
    }
    if (e.key === 'ArrowDown') {
      highlightedIndex = Math.min(highlightedIndex + 1, filteredGames.length - 1);
      e.preventDefault();
    } else if (e.key === 'ArrowUp') {
      highlightedIndex = Math.max(highlightedIndex - 1, 0);
      e.preventDefault();
    } else if (e.key === 'Enter' && highlightedIndex >= 0 && highlightedIndex < filteredGames.length) {
      selectGame(filteredGames[highlightedIndex].id);
      e.preventDefault();
    } else if (e.key === 'Escape') {
      isComboOpen = false;
      searchQuery = '';
      highlightedIndex = -1;
      e.preventDefault();
    }
  }

  // ----------------------------------------------------------------
  // Lifecycle
  // ----------------------------------------------------------------
  onMount(async () => {
    try {
      await ensureSession();

      // Fetch games (cached)
      games = await getCachedGames();

      if (games.length > 0) {
        selectedGameId = games[0].id;
      }
    } finally {
      isLoading = false;
    }
  });

  // Fetch leaderboard when game selection changes
  $effect(() => {
    if (selectedGameId) {
      fetchLeaderboard(selectedGameId);
    }
  });

  async function fetchLeaderboard(gameId: string) {
    isTableLoading = true;
    try {
      const res = await gameApi.getLeaderboard(gameId);
      entries = (res.data as any)?.data ?? res.data ?? [];
    } catch {
      entries = [];
    } finally {
      isTableLoading = false;
    }
  }
</script>

<div class={`h-[calc(100dvh-64px)] overflow-y-auto transition-colors ${isDarkMode ? 'bg-gradient-to-br from-black to-gray-950' : 'bg-gradient-to-br from-gray-50 to-gray-100'}`}>
  <main class="max-w-[1800px] mx-auto px-4 py-6 md:px-8 md:py-10 lg:px-10 lg:py-12">

    {#if isLoading}
      <!-- Skeleton: Header -->
      <div class="mb-8">
        <div class={`h-10 w-48 rounded-lg skeleton mb-2 ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
        <div class={`h-4 w-36 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
      </div>
      <!-- Skeleton: Game selector -->
      <div class="flex gap-2 mb-6">
        {#each Array(3) as _}
          <div class={`h-10 w-24 rounded-full skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
        {/each}
      </div>
      <!-- Skeleton: Table -->
      <div class={`rounded-2xl border overflow-hidden shadow-lg ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}>
        <div class={`px-5 py-3.5 border-b ${isDarkMode ? 'border-gray-800' : 'border-gray-200'}`}>
          <div class={`h-3 w-full rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
        </div>
        {#each Array(5) as _, i}
          <div class={`px-5 py-3.5 flex items-center gap-4 ${i > 0 ? (isDarkMode ? 'border-t border-gray-800/50' : 'border-t border-gray-100') : ''}`}>
            <div class={`w-7 h-7 rounded-full shrink-0 skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
            <div class={`h-4 flex-1 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
            <div class={`h-4 w-12 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
            <div class={`h-4 w-16 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
          </div>
        {/each}
      </div>
    {:else}
      <!-- Header -->
      <div class="mb-8">
        <h1 class={`text-3xl md:text-4xl font-black mb-2 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>리더보드</h1>
        <p class={`text-sm ${isDarkMode ? 'text-gray-500' : 'text-gray-500'}`}>게임별 최고 기록 Top 10</p>
      </div>

      <!-- Game Selector: Searchable Combobox -->
      <div class="relative mb-6 max-w-xs">
        <!-- svelte-ignore a11y_role_has_required_aria_props -->
        <button
          onclick={() => { isComboOpen = !isComboOpen; highlightedIndex = -1; searchQuery = ''; }}
          class={`w-full flex items-center justify-between gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold transition-all border ${
            isDarkMode
              ? 'bg-gray-900 text-gray-200 border-gray-700 hover:border-gray-600'
              : 'bg-white text-gray-800 border-gray-300 hover:border-gray-400'
          } shadow-sm`}
          role="combobox"
          aria-expanded={isComboOpen}
        >
          <span class="truncate">{selectedGameLabel}</span>
          <svg class={`w-4 h-4 shrink-0 transition-transform ${isComboOpen ? 'rotate-180' : ''} ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
          </svg>
        </button>

        {#if isComboOpen}
          <!-- Backdrop to close -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div class="fixed inset-0 z-40" onclick={() => { isComboOpen = false; searchQuery = ''; }} onkeydown={() => {}}></div>

          <div class={`absolute z-50 mt-1.5 w-full rounded-xl border shadow-xl overflow-hidden ${
            isDarkMode ? 'bg-gray-900 border-gray-700' : 'bg-white border-gray-200'
          }`}>
            <!-- Search input -->
            <div class={`px-3 py-2 border-b ${isDarkMode ? 'border-gray-800' : 'border-gray-100'}`}>
              <div class="relative">
                <svg class={`absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                </svg>
                <input
                  type="text"
                  bind:value={searchQuery}
                  onkeydown={handleComboKeydown}
                  placeholder="게임 검색…"
                  class={`w-full pl-8 pr-3 py-1.5 text-sm rounded-lg border-none focus:ring-1 focus:ring-[#FF4D00]/50 outline-none ${
                    isDarkMode ? 'bg-gray-800 text-gray-200 placeholder-gray-500' : 'bg-gray-50 text-gray-800 placeholder-gray-400'
                  }`}
                />
              </div>
            </div>
            <!-- Options -->
            <ul class="max-h-52 overflow-y-auto py-1" role="listbox">
              {#each filteredGames as g, i (g.id)}
                <li>
                  <button
                    onclick={() => selectGame(g.id)}
                    onmouseenter={() => highlightedIndex = i}
                    class={`w-full px-4 py-2 text-sm text-left transition-colors flex items-center gap-2 ${
                      g.id === selectedGameId
                        ? 'text-[#FF4D00] font-bold'
                        : isDarkMode ? 'text-gray-300' : 'text-gray-700'
                    } ${
                      highlightedIndex === i
                        ? isDarkMode ? 'bg-gray-800' : 'bg-gray-100'
                        : ''
                    }`}
                    role="option"
                    aria-selected={g.id === selectedGameId}
                  >
                    {#if g.id === selectedGameId}
                      <svg class="w-4 h-4 shrink-0 text-[#FF4D00]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                      </svg>
                    {:else}
                      <span class="w-4 shrink-0"></span>
                    {/if}
                    <span class="truncate">{g.title}</span>
                  </button>
                </li>
              {:else}
                <li class={`px-4 py-3 text-sm text-center ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
                  검색 결과가 없습니다
                </li>
              {/each}
            </ul>
          </div>
        {/if}
      </div>

      <!-- Table -->
      <div class={`rounded-2xl border overflow-hidden shadow-lg ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`} in:fly={{ y: 20, duration: 300 }}>
        {#if isTableLoading}
          <div class={`px-5 py-3.5 border-b ${isDarkMode ? 'border-gray-800' : 'border-gray-200'}`}>
            <div class={`h-3 w-full rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
          </div>
          {#each Array(5) as _, i}
            <div class={`px-5 py-3.5 flex items-center gap-4 ${i > 0 ? (isDarkMode ? 'border-t border-gray-800/50' : 'border-t border-gray-100') : ''}`}>
              <div class={`w-7 h-7 rounded-full shrink-0 skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
              <div class={`h-4 flex-1 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
              <div class={`h-4 w-12 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
              <div class={`h-4 w-16 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
            </div>
          {/each}
        {:else if entries.length === 0}
          <div class="text-center py-16">
            <p class={`text-lg font-semibold mb-1 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>아직 기록이 없습니다</p>
            <p class={`text-sm ${isDarkMode ? 'text-gray-600' : 'text-gray-500'}`}>첫 번째 클리어 기록의 주인공이 되어보세요!</p>
          </div>
        {:else}
          <table class="w-full">
            <thead>
              <tr class={`text-xs uppercase tracking-wider ${isDarkMode ? 'text-gray-500 border-b border-gray-800' : 'text-gray-400 border-b border-gray-200'}`}>
                <th class="px-5 py-3.5 text-left w-16">순위</th>
                <th class="px-5 py-3.5 text-left">플레이어</th>
                <th class="px-5 py-3.5 text-right w-24">턴</th>
                <th class="px-5 py-3.5 text-right w-28">
                  <span class="inline-flex items-center justify-end gap-1">
                    토큰
                    <span class="relative">
                      <button
                        onclick={(e) => { e.stopPropagation(); showTokenInfo = !showTokenInfo; }}
                        class={`w-4 h-4 rounded-full text-[10px] font-bold flex items-center justify-center transition-colors ${
                          showTokenInfo
                            ? isDarkMode ? 'bg-orange-500/30 text-orange-300' : 'bg-orange-100 text-orange-600'
                            : isDarkMode ? 'bg-gray-700 text-gray-400 hover:bg-gray-600' : 'bg-gray-200 text-gray-500 hover:bg-gray-300'
                        }`}
                        aria-label="토큰 설명"
                      >?</button>
                      {#if showTokenInfo}
                        <!-- svelte-ignore a11y_no_static_element_interactions -->
                        <div
                          class="fixed inset-0 z-40"
                          onclick={() => (showTokenInfo = false)}
                          onkeydown={() => {}}
                        ></div>
                        <div class={`absolute z-50 right-0 top-6 w-64 rounded-xl shadow-xl p-3.5 text-left ${
                          isDarkMode ? 'bg-gray-900 border border-gray-700' : 'bg-white border border-gray-200'
                        }`}>
                          <div class={`text-xs font-bold mb-1 ${ isDarkMode ? 'text-orange-300' : 'text-orange-600' }`}>에이전트(AI)가 사용한 토큰 수</div>
                          <p class={`text-xs leading-relaxed ${ isDarkMode ? 'text-gray-400' : 'text-gray-600' }`}>토큰은 AI가 대화를 이해하고 응답을 생성하는 데 쓴 비용으로, 읽은 토큰 + 입력한 토큰의 합산입니다. 토큰 수가 적을수록 짧고 직접적인 대화로 클리어한 것으로, 더 효율적인 플레이를 나타냅니다.</p>
                        </div>
                      {/if}
                    </span>
                  </span>
                </th>
                <th class="px-5 py-3.5 text-right w-32 hidden sm:table-cell">달성일</th>
              </tr>
            </thead>
            <tbody>
              {#each entries as entry, i (entry.user_id)}
                <tr class={`transition-colors ${isDarkMode ? 'hover:bg-gray-900/60' : 'hover:bg-gray-50'} ${i > 0 ? (isDarkMode ? 'border-t border-gray-800/50' : 'border-t border-gray-100') : ''}`}>
                  <td class="px-5 py-3.5">
                    {#if entry.rank <= 3}
                      <img src={`/medal-${entry.rank}.svg`} alt={`${entry.rank}위`} class="w-8 h-8" />
                    {:else}
                      <span class={`inline-flex items-center justify-center w-7 h-7 text-sm font-semibold ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>{entry.rank}</span>
                    {/if}
                  </td>
                  <td class={`px-5 py-3.5 font-semibold ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}>
                    <div class="flex items-center gap-3">
                      <div class="relative w-8 h-8 shrink-0 bg-gray-700 rounded-full flex items-center justify-center text-xs font-bold text-gray-300">
                        <img
                          src={`https://storage.googleapis.com/ollm-assets-prod/user/${entry.user_id}.png`}
                          alt={entry.username}
                          class="w-8 h-8 rounded-full object-cover"
                          use:profileFallback
                        />
                      </div>
                      <span>{entry.username}</span>
                    </div>
                  </td>
                  <td class={`px-5 py-3.5 text-right tabular-nums font-mono text-sm ${isDarkMode ? 'text-gray-300' : 'text-gray-700'}`}>
                    {entry.turn_count}
                  </td>
                  <td class={`px-5 py-3.5 text-right tabular-nums font-mono text-sm ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>
                    {entry.total_tokens.toLocaleString()}
                  </td>
                  <td class={`px-5 py-3.5 text-right text-sm hidden sm:table-cell ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
                    {new Date(entry.achieved_at).toLocaleDateString()}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </div>
    {/if}
  </main>
</div>
