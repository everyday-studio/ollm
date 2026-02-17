<script lang="ts">
  import { fade, slide } from 'svelte/transition';
  import { onMount, tick } from 'svelte';
  // Import 'invalidateAll' to force reload server data on logout
  import { goto, invalidateAll } from '$app/navigation'; 

  // Imports for API and Types
  import { gameApi } from '$lib/features/game/api';
  import { authApi } from '$lib/features/auth/api';     
  import { authStore } from '$lib/features/auth/model'; 
  import { toGameUI, toMatchUI } from './adapter';
  import type { GameUI, MatchUI } from '$lib/features/game/types';

  const themeColor = "#FF4D00";

  // ----------------------------------------------------------------
  // State Management (Svelte 5 Runes)
  // ----------------------------------------------------------------

  let games = $state<GameUI[]>([]);
  let matches = $state<MatchUI[]>([]);
  
  let activeTab = $state<'matches' | 'scenarios'>('matches');
  let selectedGame = $state<GameUI | null>(null);
  let selectedMatch = $state<MatchUI | null>(null);
  
  let isLoading = $state(true);
  let isImageLoaded = $state(false);
  let imgEl: HTMLImageElement | null = null;

  // ----------------------------------------------------------------
  // Lifecycle & Logic
  // ----------------------------------------------------------------

  onMount(async () => {
    try {
      const [gamesRes, matchesRes] = await Promise.all([
        gameApi.getGames(),
        gameApi.getMyMatches()
      ]);

      const rawGames = gamesRes.data;
      const rawMatches = matchesRes.data;

      // Transform DTOs to UI models
      games = rawGames.map(toGameUI);
      matches = rawMatches.map(m => toMatchUI(m, rawGames));

      // Set initial selection
      if (games.length > 0) selectedGame = games[0];
      if (matches.length > 0) selectedMatch = matches[0];

    } catch (error) {
      console.error("Failed to load lobby data:", error);
    } finally {
      isLoading = false;
    }
  });

  // Effect to handle image fade-in
  $effect(() => {
    if (selectedGame && imgEl?.complete) {
      isImageLoaded = true;
    }
  });

  function switchTab(tab: 'matches' | 'scenarios') {
    activeTab = tab;
  }

  async function selectGame(game: GameUI) {
    isImageLoaded = false;
    selectedGame = game;
    await tick(); 
    if (imgEl?.complete) isImageLoaded = true;
  }

  function selectMatch(match: MatchUI) {
    selectedMatch = match;
  }

  async function startNewMatch() {
    if (!selectedGame) return;

    try {
      const res = await gameApi.createMatch(selectedGame.id);
      
      const newMatchUI = toMatchUI(res.data, games); 
      matches = [newMatchUI, ...matches];
      
      selectedMatch = newMatchUI;
      activeTab = 'matches';

    } catch (error) {
      console.error("Failed to create match:", error);
    }
  }

  // [FIXED] Logout Handler
  async function handleLogout() {
    try {
      // 1. Request server to delete HTTP-only cookie
      await authApi.logout();
    } catch (e) {
      // Even if API fails (e.g. network error), force client logout
      console.warn("Logout request failed", e);
    } finally {
      // 2. Clear client-side store
      authStore.logout();
      
      // 3. Force reload all data to trigger server-side hooks (+layout.server.ts)
      // This ensures the browser realizes the cookie is gone
      
      // 4. Redirect to login page
      await goto('/login', { invalidateAll: true });
    }
  }
</script>

<div class="min-h-screen bg-gray-50 text-[#333] font-sans">
  <main class="max-w-7xl mx-auto px-4 py-8 lg:px-6 lg:py-12">
    
    {#if isLoading}
      <div class="flex items-center justify-center h-[700px]">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#FF4D00]"></div>
      </div>
    {:else}
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 h-[750px] lg:h-[700px]">
        
        <aside class="lg:col-span-4 flex flex-col h-full bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
          
          <div class="flex border-b border-gray-100">
            <button 
              onclick={() => switchTab('matches')}
              class="flex-1 py-4 text-sm font-extrabold tracking-wider uppercase transition-colors relative
                     {activeTab === 'matches' ? 'text-[#FF4D00] bg-gray-50' : 'text-gray-400 hover:text-gray-600'}"
            >
              Matches
              {#if activeTab === 'matches'}
                <div class="absolute bottom-0 left-0 right-0 h-0.5 bg-[#FF4D00]" transition:slide={{ axis: 'x', duration: 200 }}></div>
              {/if}
            </button>
            
            <button 
              onclick={() => switchTab('scenarios')}
              class="flex-1 py-4 text-sm font-extrabold tracking-wider uppercase transition-colors relative
                     {activeTab === 'scenarios' ? 'text-[#FF4D00] bg-gray-50' : 'text-gray-400 hover:text-gray-600'}"
            >
              New Match
              {#if activeTab === 'scenarios'}
                <div class="absolute bottom-0 left-0 right-0 h-0.5 bg-[#FF4D00]" transition:slide={{ axis: 'x', duration: 200 }}></div>
              {/if}
            </button>
          </div>

          <div class="flex-1 overflow-y-auto custom-scrollbar p-2">
            {#if activeTab === 'matches'}
              <div in:fade={{ duration: 200 }} class="flex flex-col gap-1">
                {#each matches as match}
                  <button 
                    onclick={() => selectMatch(match)}
                    class="w-full text-left p-4 rounded-xl transition-all duration-200 border border-transparent group
                           {selectedMatch?.id === match.id 
                             ? 'bg-[#FF4D00]/5 border-[#FF4D00]/20 shadow-sm' 
                             : 'hover:bg-gray-50 hover:border-gray-100'}"
                  >
                    <div class="flex justify-between items-start mb-1">
                      <span class="font-bold text-gray-800 text-sm line-clamp-1 group-hover:text-[#FF4D00] transition-colors">
                        {match.gameTitle}
                      </span>
                      <span class="text-xs text-gray-400 shrink-0 ml-2">{match.displayTime}</span>
                    </div>
                    <div class="text-xs text-gray-500 line-clamp-2">
                      {match.lastMessage}
                    </div>
                  </button>
                {/each}

                {#if matches.length === 0}
                  <div class="text-center py-10 text-gray-400 text-sm">
                    No active matches found.<br>Start a new game!
                  </div>
                {/if}
              </div>

            {:else}
              <div in:fade={{ duration: 200 }} class="flex flex-col gap-2">
                {#each games as game}
                  <button 
                    onclick={() => selectGame(game)}
                    class="group w-full text-left p-4 rounded-xl transition-all duration-200 border-2 relative overflow-hidden
                           {selectedGame?.id === game.id 
                             ? 'bg-white border-[#FF4D00] shadow-md' 
                             : 'bg-white border-transparent hover:bg-gray-50 hover:border-gray-200'}"
                  >
                    {#if selectedGame?.id === game.id}
                      <div class="absolute left-0 top-0 bottom-0 w-1.5 bg-[#FF4D00]"></div>
                    {/if}
                    <div class="pl-2">
                      <div class="text-xs font-bold uppercase tracking-wider mb-1 transition-colors"
                           class:text-[#FF4D00]={selectedGame?.id === game.id}
                           class:text-gray-400={selectedGame?.id !== game.id}
                      >
                        {game.subtitle}
                      </div>
                      <div class="font-bold text-lg text-gray-800 group-hover:text-black">
                        {game.title}
                      </div>
                    </div>
                  </button>
                {/each}
              </div>
            {/if}
          </div>

          <div class="p-4 border-t border-gray-100 bg-gray-50/50 mt-auto">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-full bg-gradient-to-tr from-green-400 to-green-600 flex items-center justify-center text-white font-bold shadow-sm">
                  {$authStore.user?.email?.[0].toUpperCase() || 'U'}
                </div>
                
                <div class="flex flex-col">
                  <span class="text-sm font-bold text-gray-800 leading-tight">
                    {$authStore.user?.name || 'Player'}
                  </span>
                  <span class="text-[10px] text-gray-500 font-mono">
                    {$authStore.user?.email || 'Guest'}
                  </span>
                </div>
              </div>

              <button 
                type="button"
                onclick={handleLogout}
                class="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors cursor-pointer"
                title="Logout"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                </svg>
              </button>
            </div>
          </div>
        </aside>

        <section class="lg:col-span-8 bg-white rounded-2xl shadow-xl border border-gray-100 overflow-hidden flex flex-col relative h-full">
          
          {#if activeTab === 'matches' && selectedMatch}
            <div class="flex flex-col h-full" in:fade={{ duration: 200 }}>
              <div class="h-16 border-b border-gray-100 flex items-center px-6 bg-white shrink-0 justify-between">
                <h2 class="font-bold text-lg text-gray-800 flex items-center gap-2">
                  <span class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
                  {selectedMatch.gameTitle}
                </h2>
                <span class="text-xs font-medium px-2 py-1 rounded bg-gray-100 text-gray-600 uppercase">
                  {selectedMatch.status}
                </span>
              </div>

              <div class="flex-1 bg-gray-50 p-6 overflow-y-auto flex flex-col gap-4 items-center justify-center text-gray-400">
                <p>Chat history will be implemented here.</p>
                <p class="text-xs">Match ID: {selectedMatch.id}</p>
              </div>

              <div class="p-4 bg-white border-t border-gray-100 shrink-0">
                <div class="relative">
                  <input 
                    type="text" 
                    placeholder="Enter command..." 
                    class="w-full bg-gray-100 text-gray-700 rounded-full pl-5 pr-12 py-3 focus:outline-none focus:ring-2 focus:ring-[#FF4D00]/50 transition-all text-sm"
                    disabled
                  />
                  <button class="absolute right-2 top-1.5 p-1.5 bg-[#FF4D00] text-white rounded-full hover:bg-[#ff3300] transition-colors disabled:opacity-50">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M12 5l7 7-7 7"></path></svg>
                  </button>
                </div>
              </div>
            </div>

          {:else if activeTab === 'scenarios' && selectedGame}
            {#key selectedGame.id}
              <div class="flex flex-col h-full" in:fade={{ duration: 300 }}>
                <div class="relative h-[55%] w-full bg-gray-200 group overflow-hidden shrink-0">
                  <div class="absolute inset-0 flex items-center justify-center text-gray-400">Loading...</div>
                  
                  <img 
                    bind:this={imgEl} 
                    src={selectedGame.image} 
                    alt={selectedGame.title} 
                    class="absolute inset-0 w-full h-full object-cover transition-opacity duration-700
                           {isImageLoaded ? 'opacity-100' : 'opacity-0'}"
                    onload={() => isImageLoaded = true}
                  />
                  
                  <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent pointer-events-none"></div>
                  
                  <div class="absolute bottom-6 left-8 text-white">
                    <div class="flex gap-2 mb-2">
                      {#each selectedGame.tags as tag}
                        <span class="bg-white/20 backdrop-blur-md px-2 py-0.5 rounded text-xs font-bold border border-white/10">
                          #{tag}
                        </span>
                      {/each}
                    </div>
                    <h1 class="text-4xl md:text-5xl font-black tracking-tight drop-shadow-lg">
                      {selectedGame.title}
                    </h1>
                  </div>
                </div>

                <div class="flex-1 p-8 md:p-10 flex flex-col justify-between bg-white overflow-y-auto">
                  <div>
                    <h3 class="text-xs font-bold text-[#FF4D00] uppercase tracking-widest mb-2">
                      Mission Briefing
                    </h3>
                    <p class="text-gray-600 text-lg leading-relaxed">
                      {selectedGame.description}
                    </p>
                  </div>

                  <div class="flex items-center justify-end mt-4 pt-4 border-t border-gray-100 shrink-0">
                    <button 
                      onclick={startNewMatch}
                      class="pl-6 pr-8 py-3 rounded-full font-black text-xl border-2 transition-all duration-300 flex items-center gap-2 active:scale-95 cursor-pointer
                             bg-transparent text-[var(--theme-color)] border-transparent
                             hover:bg-[var(--theme-color)] hover:text-white hover:border-transparent hover:shadow-md
                             focus:outline-none" 
                      style="--theme-color: {themeColor};"
                    >
                      <svg class="w-6 h-6 fill-current" viewBox="0 0 24 24">
                        <path d="M8 5v14l11-7z"/>
                      </svg>
                      START NEW MATCH
                    </button>
                  </div>
                </div>
              </div>
            {/key}
          {/if}

        </section>
      </div>
    {/if}
  </main>
</div>

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 4px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background-color: #f3f4f6;
    border-radius: 20px;
  }
  .custom-scrollbar:hover::-webkit-scrollbar-thumb {
    background-color: #d1d5db;
  }
</style>