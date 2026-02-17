<script lang="ts">
  import { fade, fly, scale } from 'svelte/transition';
  import { onMount, getContext } from 'svelte';

  // Imports for API and Types
  import { gameApi } from '$lib/features/game/api';
  import { loadMockGames } from '$lib/features/game/mockData';
  import { authApi } from '$lib/features/auth/api';     
  import { authStore } from '$lib/features/auth/model'; 
  import type { User } from '$lib/features/auth/types';
  import { toGameUI, toMatchUI } from './adapter';
  import type { GameDTO, GameUI, MatchDTO, MatchUI } from '$lib/features/game/types';

  const themeColor = "#FF4D00";
  const theme = getContext<{ isDark: boolean }>('theme');

  // ----------------------------------------------------------------
  // State Management (Svelte 5 Runes)
  // ----------------------------------------------------------------

  let games = $state<GameUI[]>([]);
  let matches = $state<MatchUI[]>([]);
  
  let selectedGame = $state<GameUI | null>(null);
  let showGameModal = $state(false);
  
  let isLoading = $state(true);
  let activeCategory = $state<'all' | 'action' | 'adventure' | 'puzzle' | 'strategy'>('all');
  let activeSection = $state<'games' | 'matches'>('games');

  let filteredGames = $derived(
    activeCategory === 'all' 
      ? games 
      : games.filter(g => g.tags?.includes(activeCategory))
  );

  let isDarkMode = $derived(theme.isDark);

  // ----------------------------------------------------------------
  // Lifecycle & Logic
  // ----------------------------------------------------------------

  onMount(async () => {
    try {
      // Restore user session
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
      } catch (refreshErr) {
        console.warn('Failed to restore user session:', refreshErr);
      }

      let rawGames: GameDTO[] = [];
      let rawMatches: MatchDTO[] = [];

      try {
        const gamesRes = await gameApi.getGames();
        const apiGames = gamesRes.data;
        if (Array.isArray(apiGames) && apiGames.length > 0) {
          rawGames = apiGames;
        } else {
          rawGames = await loadMockGames();
        }
      } catch (gamesError) {
        console.warn('Games API failed. Using mock data.', gamesError);
        rawGames = await loadMockGames();
      }

      try {
        const matchesRes = await gameApi.getMyMatches();
        rawMatches = matchesRes.data;
      } catch (matchesError) {
        console.warn('Matches API failed. Using empty list.', matchesError);
        rawMatches = [];
      }

      games = rawGames.map(toGameUI);
      matches = rawMatches.map(m => toMatchUI(m, rawGames));

    } catch (error) {
      console.error("Failed to load lobby data:", error);
    } finally {
      isLoading = false;
    }
  });

  function openGameModal(game: GameUI) {
    selectedGame = game;
    showGameModal = true;
  }

  async function startNewMatch(game: GameUI) {
    try {
      const res = await gameApi.createMatch(game.id);
      const newMatchUI = toMatchUI(res.data, games); 
      matches = [newMatchUI, ...matches];
      showGameModal = false;
      activeSection = 'matches';
    } catch (error) {
      console.error("Failed to create match:", error);
    }
  }

</script>

<div class={`min-h-screen transition-colors ${isDarkMode ? 'bg-gradient-to-br from-black to-gray-950' : 'bg-gradient-to-br from-gray-50 to-gray-100'}`}>
  <main class="max-w-[1800px] mx-auto px-4 py-6 md:px-8 md:py-10 lg:px-10 lg:py-12">
    
    {#if isLoading}
      <div class="flex items-center justify-center h-[700px]">
        <div class="animate-spin rounded-full h-16 w-16 border-4 border-[#FF4D00] border-t-transparent"></div>
      </div>
    {:else}
      <!-- Hero Banner -->
      <section class="mb-8">
        <div class="relative h-[320px] md:h-[400px] rounded-3xl overflow-hidden shadow-2xl group">
          {#if games.length > 0}
            <img 
              src={games[0].image} 
              alt={games[0].title}
              class="absolute inset-0 w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
            />
            <div class="absolute inset-0 bg-gradient-to-r from-black/80 via-black/40 to-transparent"></div>
            
            <div class="absolute inset-0 flex flex-col justify-center px-8 md:px-16">
              <div class="flex gap-2 mb-4 flex-wrap">
                {#each games[0].tags as tag}
                  <span class="bg-white/20 backdrop-blur-md px-3 py-1 rounded-full text-sm font-bold text-white border border-white/20">
                    #{tag}
                  </span>
                {/each}
              </div>
              <h1 class="text-4xl md:text-6xl lg:text-7xl font-black text-white mb-3 md:mb-4 drop-shadow-2xl">
                {games[0].title}
              </h1>
              <p class="text-base md:text-lg lg:text-xl text-white/90 max-w-2xl mb-4 md:mb-6 line-clamp-2">
                {games[0].description}
              </p>
              <button 
                onclick={() => openGameModal(games[0])}
                class="self-start px-6 py-3 md:px-8 md:py-4 bg-[#FF4D00] text-white rounded-full font-bold text-base md:text-lg hover:bg-[#ff3300] transition-all hover:scale-105 active:scale-95 shadow-xl"
              >
                지금 플레이
              </button>
            </div>
          {/if}
        </div>
      </section>

      <!-- Section Toggle -->
      <div class="flex gap-4 mb-6">
        <button 
          onclick={() => activeSection = 'games'}
          class="px-5 md:px-6 py-2 md:py-3 rounded-full font-bold text-base md:text-lg transition-all
                 {activeSection === 'games' 
                   ? 'bg-[#FF4D00] text-white shadow-lg' 
                   : isDarkMode ? 'bg-gray-900 text-gray-300 hover:bg-gray-800 border border-gray-800' : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'}"
        >
          게임 둘러보기
        </button>
        <button 
          onclick={() => activeSection = 'matches'}
          class="px-5 md:px-6 py-2 md:py-3 rounded-full font-bold text-base md:text-lg transition-all
                 {activeSection === 'matches' 
                   ? 'bg-[#FF4D00] text-white shadow-lg' 
                   : isDarkMode ? 'bg-gray-900 text-gray-300 hover:bg-gray-800 border border-gray-800' : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'}"
        >
          내 매치 {matches.length > 0 ? `(${matches.length})` : ''}
        </button>
      </div>

      <!-- Games Section -->
      {#if activeSection === 'games'}
        <div in:fly={{ y: 20, duration: 300 }}>
          <!-- Category Filters -->
          <nav class="flex gap-2 md:gap-3 mb-6 md:mb-8 flex-wrap">
            <button 
              onclick={() => activeCategory = 'all'}
              class={`px-4 md:px-5 py-2 rounded-full font-semibold text-sm transition-all ${activeCategory === 'all' 
                       ? isDarkMode ? 'bg-gray-800 text-white' : 'bg-gray-800 text-white' 
                       : isDarkMode ? 'bg-gray-900 text-gray-400 hover:bg-gray-800 border border-gray-800' : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'}`}
            >
              모든 게임
            </button>
            <button 
              onclick={() => activeCategory = 'action'}
              class={`px-4 md:px-5 py-2 rounded-full font-semibold text-sm transition-all ${activeCategory === 'action' 
                       ? isDarkMode ? 'bg-gray-800 text-white' : 'bg-gray-800 text-white' 
                       : isDarkMode ? 'bg-gray-900 text-gray-400 hover:bg-gray-800 border border-gray-800' : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'}`}
            >
              액션
            </button>
            <button 
              onclick={() => activeCategory = 'adventure'}
              class={`px-4 md:px-5 py-2 rounded-full font-semibold text-sm transition-all ${activeCategory === 'adventure' 
                       ? isDarkMode ? 'bg-gray-800 text-white' : 'bg-gray-800 text-white' 
                       : isDarkMode ? 'bg-gray-900 text-gray-400 hover:bg-gray-800 border border-gray-800' : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'}`}
            >
              어드벤처
            </button>
            <button 
              onclick={() => activeCategory = 'puzzle'}
              class={`px-4 md:px-5 py-2 rounded-full font-semibold text-sm transition-all ${activeCategory === 'puzzle' 
                       ? isDarkMode ? 'bg-gray-800 text-white' : 'bg-gray-800 text-white' 
                       : isDarkMode ? 'bg-gray-900 text-gray-400 hover:bg-gray-800 border border-gray-800' : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'}`}
            >
              퍼즐
            </button>
            <button 
              onclick={() => activeCategory = 'strategy'}
              class={`px-4 md:px-5 py-2 rounded-full font-semibold text-sm transition-all ${activeCategory === 'strategy' 
                       ? isDarkMode ? 'bg-gray-800 text-white' : 'bg-gray-800 text-white' 
                       : isDarkMode ? 'bg-gray-900 text-gray-400 hover:bg-gray-800 border border-gray-800' : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'}`}
            >
              전략
            </button>
          </nav>

          <!-- Games Grid -->
          <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4 md:gap-6">
            {#each filteredGames as game}
              <button 
                onclick={() => openGameModal(game)}
                class={`group rounded-2xl overflow-hidden shadow-lg hover:shadow-2xl transition-all duration-300 hover:scale-105 active:scale-100 flex flex-col border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
              >
                <div class="relative aspect-[16/10] bg-gray-200 overflow-hidden">
                  <img 
                    src={game.image} 
                    alt={game.title}
                    class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                  />
                  <div class="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
                  
                  <div class="absolute top-2 left-2 flex gap-1 flex-wrap">
                    {#each game.tags as tag}
                      <span class="bg-black/60 backdrop-blur-sm px-2 py-0.5 rounded text-xs font-bold text-white">
                        {tag}
                      </span>
                    {/each}
                  </div>
                  
                  <div class="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                    <div class="w-12 h-12 md:w-14 md:h-14 rounded-full bg-[#FF4D00] flex items-center justify-center shadow-xl">
                      <svg class="w-5 h-5 md:w-6 md:h-6 fill-white ml-1" viewBox="0 0 24 24">
                        <path d="M8 5v14l11-7z"/>
                      </svg>
                    </div>
                  </div>
                </div>
                
                <div class="p-3 md:p-4 flex-1 flex flex-col">
                  <h3 class={`font-bold text-base md:text-lg group-hover:text-[#FF4D00] transition-colors mb-1 line-clamp-1 ${isDarkMode ? 'text-gray-100' : 'text-gray-800'}`}>
                    {game.title}
                  </h3>
                  <p class={`text-xs line-clamp-2 flex-1 ${isDarkMode ? 'text-gray-400' : 'text-gray-600'}`}>
                    {game.description}
                  </p>
                </div>
              </button>
            {/each}
          </div>

          {#if filteredGames.length === 0}
            <div class={`text-center py-20 ${isDarkMode ? 'text-gray-500' : 'text-gray-600'}`}>
              <p class="text-lg font-semibold">이 카테고리에 게임이 없습니다.</p>
            </div>
          {/if}
        </div>

      {:else}
        <!-- Matches Section -->
        <div in:fly={{ y: 20, duration: 300 }}>
          {#if matches.length === 0}
            <div class={`text-center py-16 md:py-20 rounded-2xl shadow-lg border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}>
              <p class={`text-lg mb-4 ${isDarkMode ? 'text-gray-400' : 'text-gray-600'}`}>아직 활성 매치가 없습니다.</p>
              <button 
                onclick={() => activeSection = 'games'}
                class="px-6 py-3 bg-[#FF4D00] text-white rounded-full font-bold hover:bg-[#ff3300] transition-colors"
              >
                첫 게임 시작하기
              </button>
            </div>
          {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {#each matches as match}
                <div class={`rounded-2xl p-5 md:p-6 shadow-lg hover:shadow-xl transition-all hover:scale-105 active:scale-100 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}>
                  <div class="flex justify-between items-start mb-3">
                    <h3 class={`font-bold text-lg md:text-xl line-clamp-1 ${isDarkMode ? 'text-gray-100' : 'text-gray-800'}`}>
                      {match.gameTitle}
                    </h3>
                    <span class={`text-xs font-medium px-2 py-1 rounded uppercase shrink-0 ml-2 ${isDarkMode ? 'bg-green-900 text-green-300' : 'bg-green-100 text-green-700'}`}>
                      {match.status}
                    </span>
                  </div>
                  <p class={`text-sm line-clamp-2 mb-3 ${isDarkMode ? 'text-gray-400' : 'text-gray-600'}`}>
                    {match.lastMessage}
                  </p>
                  <div class={`text-xs ${isDarkMode ? 'text-gray-500' : 'text-gray-500'}`}>
                    {match.displayTime}
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    {/if}
  </main>
</div>

<!-- Game Detail Modal -->
{#if showGameModal && selectedGame}
  <div 
    class="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4" 
    transition:fade={{ duration: 200 }}
    onclick={() => showGameModal = false}
    onkeydown={(e) => e.key === 'Escape' && (showGameModal = false)}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
  >
    <div 
      class={`rounded-3xl max-w-4xl w-full max-h-[90vh] overflow-hidden shadow-2xl border transition-colors ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`} 
      transition:fly={{ y: 50, duration: 300 }}
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="presentation"
    >
      <div class="relative h-[250px] md:h-[300px] bg-gray-200">
        <img 
          src={selectedGame.image} 
          alt={selectedGame.title}
          class="w-full h-full object-cover"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent"></div>
        
        <button 
          onclick={() => showGameModal = false}
          class="absolute top-4 right-4 w-10 h-10 bg-black/50 hover:bg-black/70 rounded-full flex items-center justify-center text-white transition-colors"
          aria-label="모달 닫기"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
        
        <div class="absolute bottom-4 md:bottom-6 left-4 md:left-6">
          <div class="flex gap-2 mb-2 md:mb-3 flex-wrap">
            {#each selectedGame.tags as tag}
              <span class="bg-white/20 backdrop-blur-md px-2 md:px-3 py-0.5 md:py-1 rounded-full text-xs md:text-sm font-bold text-white border border-white/20">
                #{tag}
              </span>
            {/each}
          </div>
          <h2 class="text-3xl md:text-4xl font-black text-white drop-shadow-lg">
            {selectedGame.title}
          </h2>
        </div>
      </div>
      
      <div class="p-6 md:p-8">
        <h3 class="text-xs font-bold text-[#FF4D00] uppercase tracking-widest mb-3">
          게임 소개
        </h3>
        <p class={`text-base md:text-lg leading-relaxed mb-6 ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}>
          {selectedGame.description}
        </p>
        
        <div class="flex justify-end gap-3">
          <button 
            onclick={() => showGameModal = false}
            class={`px-5 md:px-6 py-2 md:py-3 rounded-full font-bold transition-colors ${isDarkMode ? 'text-gray-300 hover:bg-gray-900' : 'text-gray-600 hover:bg-gray-100'}`}
          >
            취소
          </button>
          <button 
            onclick={() => { if (selectedGame) startNewMatch(selectedGame); showGameModal = false; }}
            class="px-6 md:px-8 py-2 md:py-3 bg-[#FF4D00] text-white rounded-full font-bold hover:bg-[#ff3300] transition-all hover:scale-105 active:scale-95 shadow-lg flex items-center gap-2"
          >
            <svg class="w-5 h-5 fill-current" viewBox="0 0 24 24">
              <path d="M8 5v14l11-7z"/>
            </svg>
            게임 시작
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}