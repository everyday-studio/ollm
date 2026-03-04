<script lang="ts">
  import { fly } from 'svelte/transition';
  import { onMount, getContext } from 'svelte';

  import { gameApi } from '$lib/features/game/api';
  import { loadMockGames } from '$lib/features/game/mockData';
  import { authApi } from '$lib/features/auth/api';
  import { authStore } from '$lib/features/auth/model';
  import type { User } from '$lib/features/auth/types';
  import type { GameDTO, LeaderboardEntry } from '$lib/features/game/types';

  const theme = getContext<{ isDark: boolean }>('theme');
  let isDarkMode = $derived(theme.isDark);

  // ----------------------------------------------------------------
  // State
  // ----------------------------------------------------------------
  let games = $state<GameDTO[]>([]);
  let selectedGameId = $state<string>('');
  let entries = $state<LeaderboardEntry[]>([]);
  let isLoading = $state(true);
  let isTableLoading = $state(false);

  // ----------------------------------------------------------------
  // Lifecycle
  // ----------------------------------------------------------------
  onMount(async () => {
    try {
      // Restore access token (layout handles getMe for user info)
      try {
        const refreshRes = await authApi.refresh();
        if (refreshRes?.data?.access_token) {
          authStore.updateToken(refreshRes.data.access_token);
        }
      } catch { /* ignore */ }

      // Fetch games
      try {
        const gamesRes = await gameApi.getGames();
        games = Array.isArray(gamesRes.data) && gamesRes.data.length > 0 ? gamesRes.data : await loadMockGames();
      } catch {
        games = await loadMockGames();
      }

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

<div class={`h-[calc(100vh-64px)] overflow-y-auto transition-colors ${isDarkMode ? 'bg-gradient-to-br from-black to-gray-950' : 'bg-gradient-to-br from-gray-50 to-gray-100'}`}>
  <main class="max-w-[1800px] mx-auto px-4 py-6 md:px-8 md:py-10 lg:px-10 lg:py-12">

    {#if isLoading}
      <div class="flex items-center justify-center h-[400px]">
        <div class="animate-spin rounded-full h-12 w-12 border-4 border-[#FF4D00] border-t-transparent"></div>
      </div>
    {:else}
      <!-- Header -->
      <div class="mb-8">
        <h1 class={`text-3xl md:text-4xl font-black mb-2 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>리더보드</h1>
        <p class={`text-sm ${isDarkMode ? 'text-gray-500' : 'text-gray-500'}`}>게임별 최고 기록 Top 10</p>
      </div>

      <!-- Game Selector -->
      <div class="flex gap-2 mb-6 flex-wrap">
        {#each games as g (g.id)}
          <button
            onclick={() => selectedGameId = g.id}
            class={`px-4 py-2 rounded-full text-sm font-semibold transition-all ${
              selectedGameId === g.id
                ? 'bg-[#FF4D00] text-white shadow-lg'
                : isDarkMode ? 'bg-gray-900 text-gray-400 hover:bg-gray-800 border border-gray-800' : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'
            }`}
          >
            {g.title}
          </button>
        {/each}
      </div>

      <!-- Table -->
      <div class={`rounded-2xl border overflow-hidden shadow-lg ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`} in:fly={{ y: 20, duration: 300 }}>
        {#if isTableLoading}
          <div class="flex items-center justify-center py-20">
            <div class="animate-spin rounded-full h-10 w-10 border-4 border-[#FF4D00] border-t-transparent"></div>
          </div>
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
                <th class="px-5 py-3.5 text-right w-28">토큰</th>
                <th class="px-5 py-3.5 text-right w-32 hidden sm:table-cell">달성일</th>
              </tr>
            </thead>
            <tbody>
              {#each entries as entry, i (entry.user_id)}
                <tr class={`transition-colors ${isDarkMode ? 'hover:bg-gray-900/60' : 'hover:bg-gray-50'} ${i > 0 ? (isDarkMode ? 'border-t border-gray-800/50' : 'border-t border-gray-100') : ''}`}>
                  <td class="px-5 py-3.5">
                    {#if entry.rank === 1}
                      <span class="inline-flex items-center justify-center w-7 h-7 rounded-full bg-yellow-500/20 text-yellow-400 font-black text-sm">1</span>
                    {:else if entry.rank === 2}
                      <span class="inline-flex items-center justify-center w-7 h-7 rounded-full bg-gray-400/20 text-gray-300 font-black text-sm">2</span>
                    {:else if entry.rank === 3}
                      <span class="inline-flex items-center justify-center w-7 h-7 rounded-full bg-amber-600/20 text-amber-500 font-black text-sm">3</span>
                    {:else}
                      <span class={`inline-flex items-center justify-center w-7 h-7 text-sm font-semibold ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>{entry.rank}</span>
                    {/if}
                  </td>
                  <td class={`px-5 py-3.5 font-semibold ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}>
                    {entry.username}
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
