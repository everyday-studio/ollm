<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { onMount, getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { SvelteMap } from 'svelte/reactivity';

	import { gameApi } from '$lib/features/game/api';
	import { ensureSession } from '$lib/features/auth/session';
	import { getCachedGames, getCachedMyMatches, invalidateMatchesCache } from '$lib/cache/gameCache';
	import { toGameUI, toMatchUI } from './adapter';
	import type { GameUI, MatchUI } from '$lib/features/game/types';
	import GameCard from '$lib/components/ui/GameCard.svelte';
	import { getJudgeBadgeStyle } from '$lib/utils/gameHelpers';
	import { handleImageError, DEFAULT_GAME_THUMBNAIL } from '$lib/utils/imageFallback';

	const theme = getContext<{ isDark: boolean }>('theme');

	// ----------------------------------------------------------------
	// State Management (Svelte 5 Runes)
	// ----------------------------------------------------------------

	let games = $state<GameUI[]>([]);
	let allGames = $state<GameUI[]>([]);
	let matches = $state<MatchUI[]>([]);

	let selectedGame = $state<GameUI | null>(null);
	let showGameModal = $state(false);

	let isLoading = $state(true);
	let activeSection = $state<'games' | 'matches'>('games');
	let judgeFilter = $state<'all' | 'target_word' | 'llm_judge' | 'format_break'>('all');

	const judgeFilters = [
		{ id: 'all' as const, label: '전체' },
		{ id: 'target_word' as const, label: 'Target Word' },
		{ id: 'llm_judge' as const, label: 'LLM Judge' },
		{ id: 'format_break' as const, label: 'Format Break' }
	];

	let filteredGames = $derived(
		judgeFilter === 'all' ? games : games.filter((g) => g.judge_type === judgeFilter)
	);

	let modalJudgeBadge = $derived(
		selectedGame ? getJudgeBadgeStyle(selectedGame.judge_type) : getJudgeBadgeStyle('unknown')
	);

	// Hero Carousel
	let currentSlide = $state(0);
	let slideInterval: ReturnType<typeof setInterval> | null = null;
	let slideDirection = $state<1 | -1>(1);

	function startAutoSlide() {
		stopAutoSlide();
		slideInterval = setInterval(() => {
			if (games.length > 1) {
				slideDirection = 1;
				currentSlide = (currentSlide + 1) % games.length;
			}
		}, 6000);
	}

	function stopAutoSlide() {
		if (slideInterval) {
			clearInterval(slideInterval);
			slideInterval = null;
		}
	}

	function goToSlide(index: number) {
		slideDirection = index > currentSlide ? 1 : -1;
		currentSlide = index;
		startAutoSlide();
	}

	function prevSlide() {
		slideDirection = -1;
		currentSlide = (currentSlide - 1 + games.length) % games.length;
		startAutoSlide();
	}

	function nextSlide() {
		slideDirection = 1;
		currentSlide = (currentSlide + 1) % games.length;
		startAutoSlide();
	}

	let isDarkMode = $derived(theme.isDark);

	// Group matches by game for the "내 매치" section
	interface GameMatchGroup {
		gameId: string;
		gameTitle: string;
		matches: MatchUI[];
		total: number;
		active: number;
		won: number;
		lost: number;
		resigned: number;
		other: number;
		latestMatch: MatchUI;
	}

	let matchGroups = $derived.by<GameMatchGroup[]>(() => {
		const groupMap = new SvelteMap<string, MatchUI[]>();
		for (const m of matches) {
			const key = m.game_id;
			if (!groupMap.has(key)) groupMap.set(key, []);
			groupMap.get(key)!.push(m);
		}
		const groups: GameMatchGroup[] = [];
		for (const [gameId, groupMatches] of groupMap) {
			// Sort newest first within group
			groupMatches.sort(
				(a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
			);
			groups.push({
				gameId,
				gameTitle: groupMatches[0].gameTitle,
				matches: groupMatches,
				total: groupMatches.length,
				active: groupMatches.filter((m) => m.status === 'active' || m.status === 'generating')
					.length,
				won: groupMatches.filter((m) => m.status === 'won').length,
				lost: groupMatches.filter((m) => m.status === 'lost').length,
				resigned: groupMatches.filter((m) => m.status === 'resigned').length,
				other: groupMatches.filter((m) => m.status === 'expired' || m.status === 'error').length,
				latestMatch: groupMatches[0]
			});
		}
		// Sort groups by latest activity
		groups.sort(
			(a, b) =>
				new Date(b.latestMatch.updated_at).getTime() - new Date(a.latestMatch.updated_at).getTime()
		);
		return groups;
	});

	let filteredMatchGroups = $derived.by(() => {
		if (judgeFilter === 'all') return matchGroups ?? [];
		return (matchGroups ?? []).filter((group) => {
			const gameUI = allGames?.find((g) => g.id === group.gameId);
			return gameUI?.judge_type === judgeFilter;
		});
	});

	// ----------------------------------------------------------------
	// Lifecycle & Logic
	// ----------------------------------------------------------------

	onMount(() => {
		(async () => {
			try {
				await ensureSession();

				const [rawGames, rawMatches] = await Promise.all([getCachedGames(), getCachedMyMatches()]);

				allGames = rawGames.map(toGameUI);
				// Filter out inactive or private games from being shown in the main games list.
				games = allGames.filter((g) => g.status === 'active' && g.is_public);
				matches = rawMatches.map((m) => toMatchUI(m, rawGames));
			} catch (error) {
				console.error('Failed to load lobby data:', error);
			} finally {
				isLoading = false;
			}

			// Start carousel after data loads
			startAutoSlide();
		})();

		return () => stopAutoSlide();
	});

	function openGameModal(game: GameUI) {
		selectedGame = game;
		showGameModal = true;
	}

	async function redirectToLatestMatch(gameId: string): Promise<boolean> {
		let latestMatchId = matchGroups.find((group) => group.gameId === gameId)?.latestMatch.id;

		if (!latestMatchId) {
			try {
				const res = await gameApi.getMyMatchesByGame(gameId);
				const latest = (res.data ?? []).sort(
					(a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
				)[0];
				latestMatchId = latest?.id;
			} catch (fetchError) {
				console.error('Failed to fetch latest match by game:', fetchError);
			}
		}

		if (!latestMatchId) {
			return false;
		}

		showGameModal = false;
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(`/lobby/match/${latestMatchId}`);
		return true;
	}

	async function startNewMatch(game: GameUI) {
		try {
			const res = await gameApi.createMatch(game.id);
			invalidateMatchesCache(); // bust cache so lobby refreshes on return
			showGameModal = false;
			// Navigate directly to the match chat page
			// eslint-disable-next-line svelte/no-navigation-without-resolve
			await goto(`/lobby/match/${res.data.id}`);
		} catch (error) {
			const status = (error as { response?: { status?: number } })?.response?.status;
			if (status === 409) {
				const redirected = await redirectToLatestMatch(game.id);
				if (!redirected) {
					console.error('Match limit reached but no existing match found for game:', game.id);
				}
				return;
			}
			console.error('Failed to create match:', error);
		}
	}

	function goToMatch(matchId: string) {
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		goto(`/lobby/match/${matchId}`);
	}
</script>

<div
	class={`h-[calc(100vh-64px)] overflow-y-auto transition-colors ${isDarkMode ? 'bg-gradient-to-br from-black to-gray-950' : 'bg-gradient-to-br from-gray-50 to-gray-100'}`}
>
	<main class="max-w-[1800px] mx-auto px-4 py-6 md:px-8 md:py-10 lg:px-10 lg:py-12">
		{#if isLoading}
			<!-- Skeleton: Hero banner -->
			<section class="mb-8">
				<div
					class={`h-[320px] md:h-[400px] rounded-3xl skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
				></div>
			</section>
			<!-- Skeleton: Section toggle -->
			<div class="flex gap-4 mb-6">
				<div
					class={`h-11 w-36 rounded-full skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
				></div>
				<div
					class={`h-11 w-28 rounded-full skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
				></div>
			</div>
			<!-- Skeleton: Game grid -->
			<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4 md:gap-6">
				{#each Array.from({ length: 5 }, (__, i) => i) as i (i)}
					<div
						class={`rounded-2xl overflow-hidden border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
					>
						<div
							class={`aspect-[16/10] skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
						></div>
						<div class="p-3 md:p-4 space-y-2">
							<div
								class={`h-5 w-3/4 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
							></div>
							<div
								class={`h-3 w-full rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
							></div>
							<div
								class={`h-3 w-2/3 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
							></div>
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<!-- Hero Carousel -->
			<section class="mb-8">
				<div
					class="relative h-[320px] md:h-[400px] rounded-3xl overflow-hidden shadow-2xl group"
					onmouseenter={stopAutoSlide}
					onmouseleave={startAutoSlide}
					role="region"
					aria-label="추천 게임 슬라이더"
				>
					{#if games.length > 0}
						{#key currentSlide}
							<div
								class="absolute inset-0"
								in:fly={{ x: 300 * slideDirection, duration: 500 }}
								out:fly={{ x: -300 * slideDirection, duration: 500 }}
							>
								<img
									src={games[currentSlide].image}
									alt={games[currentSlide].title}
									class="w-full h-full object-cover"
									onerror={handleImageError(DEFAULT_GAME_THUMBNAIL)}
								/>
								<div
									class="absolute inset-0 bg-gradient-to-r from-black/80 via-black/40 to-transparent"
								></div>

								<!-- Content (inside the sliding container to move together) -->
								<div class="absolute inset-0 flex flex-col justify-center px-8 md:px-16">
									<div class="flex gap-2 mb-4 flex-wrap">
										{#each games[currentSlide].tags as tag (tag)}
											<span
												class="bg-white/20 backdrop-blur-md px-3 py-1 rounded-full text-sm font-bold text-white border border-white/20"
											>
												#{tag}
											</span>
										{/each}
									</div>
									<h1
										class="text-4xl md:text-6xl lg:text-7xl font-black text-white mb-3 md:mb-4 drop-shadow-2xl"
									>
										{games[currentSlide].title}
									</h1>
									<p
										class="text-base md:text-lg lg:text-xl text-white/90 max-w-2xl mb-4 md:mb-6 line-clamp-2"
									>
										{games[currentSlide].description}
									</p>
									<button
										onclick={() => openGameModal(games[currentSlide])}
										class="self-start px-6 py-3 md:px-8 md:py-4 bg-[#FF4D00] text-white rounded-full font-bold text-base md:text-lg hover:bg-[#ff3300] transition-all hover:scale-105 active:scale-95 shadow-xl"
									>
										지금 플레이
									</button>
								</div>
							</div>
						{/key}

						<!-- Arrow buttons -->
						{#if games.length > 1}
							<button
								onclick={prevSlide}
								class="absolute left-3 top-1/2 -translate-y-1/2 w-10 h-10 md:w-12 md:h-12 rounded-full bg-black/40 hover:bg-black/60 backdrop-blur-sm flex items-center justify-center text-white opacity-0 group-hover:opacity-100 transition-opacity"
								aria-label="이전 슬라이드"
							>
								<svg
									class="w-5 h-5 md:w-6 md:h-6"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M15 19l-7-7 7-7"
									/></svg
								>
							</button>
							<button
								onclick={nextSlide}
								class="absolute right-3 top-1/2 -translate-y-1/2 w-10 h-10 md:w-12 md:h-12 rounded-full bg-black/40 hover:bg-black/60 backdrop-blur-sm flex items-center justify-center text-white opacity-0 group-hover:opacity-100 transition-opacity"
								aria-label="다음 슬라이드"
							>
								<svg
									class="w-5 h-5 md:w-6 md:h-6"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M9 5l7 7-7 7"
									/></svg
								>
							</button>
						{/if}

						<!-- Dot indicators -->
						{#if games.length > 1}
							<div class="absolute bottom-4 md:bottom-6 right-4 md:right-8 flex gap-2 items-center">
								{#each Array.from({ length: games.length }, (__, i) => i) as i (i)}
									<button
										onclick={() => goToSlide(i)}
										class="relative h-2 rounded-full transition-all duration-300 {i === currentSlide
											? 'w-8 bg-white'
											: 'w-2 bg-white/50 hover:bg-white/80'}"
										aria-label="슬라이드 {i + 1}"
									>
										{#if i === currentSlide}
											<span class="absolute inset-0 rounded-full bg-white/40 animate-pulse"></span>
										{/if}
									</button>
								{/each}
							</div>
						{/if}
					{/if}
				</div>
			</section>

			<!-- Section Toggle -->
			<div class="flex gap-4 mb-6">
				<button
					onclick={() => (activeSection = 'games')}
					class="px-5 md:px-6 py-2 md:py-3 rounded-full font-bold text-base md:text-lg transition-all
                 {activeSection === 'games'
						? 'bg-[#FF4D00] text-white shadow-lg'
						: isDarkMode
							? 'bg-gray-900 text-gray-300 hover:bg-gray-800 border border-gray-800'
							: 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'}"
				>
					게임 둘러보기
				</button>
				<button
					onclick={() => (activeSection = 'matches')}
					class="px-5 md:px-6 py-2 md:py-3 rounded-full font-bold text-base md:text-lg transition-all
                 {activeSection === 'matches'
						? 'bg-[#FF4D00] text-white shadow-lg'
						: isDarkMode
							? 'bg-gray-900 text-gray-300 hover:bg-gray-800 border border-gray-800'
							: 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'}"
				>
					내 매치 {matches.length > 0 ? `(${matches.length})` : ''}
				</button>
			</div>

			<!-- Games Section -->
			{#if activeSection === 'games'}
				<div in:fly={{ y: 20, duration: 300 }}>
					<!-- Judge Type Filter -->
					<div class="flex gap-2 mb-4 overflow-x-auto pb-1">
						{#each judgeFilters as jf (jf.id)}
							<button
								onclick={() => (judgeFilter = jf.id)}
								class="px-3 py-1.5 rounded-full text-xs font-semibold whitespace-nowrap transition-all {judgeFilter === jf.id
									? jf.id === 'target_word'
										? 'bg-purple-500 text-white shadow'
										: jf.id === 'llm_judge'
											? 'bg-blue-500 text-white shadow'
											: jf.id === 'format_break'
												? 'bg-orange-500 text-white shadow'
												: 'bg-[#FF4D00] text-white shadow'
									: isDarkMode
										? 'bg-gray-900 text-gray-400 hover:bg-gray-800 border border-gray-800'
										: 'bg-white text-gray-500 hover:bg-gray-50 border border-gray-200'}"
							>
								{jf.label}
							</button>
						{/each}
					</div>
					<!-- Games Grid -->
					<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4 md:gap-6">
						{#each filteredGames as game (game.id)}
							<GameCard {game} {isDarkMode} onclick={() => openGameModal(game)} />
						{/each}
					</div>
					{#if filteredGames.length === 0}
						<div class={`text-center py-12 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
							<p class="text-sm">해당 유형의 게임이 없습니다.</p>
						</div>
					{/if}
				</div>
			{:else if activeSection === 'matches'}
				<!-- Matches Section: Game cards with match stats -->
				<div in:fly={{ y: 20, duration: 300 }}>
					<!-- Judge Type Filter (for matches too) -->
					<div class="flex gap-2 mb-4 overflow-x-auto pb-1">
						{#each judgeFilters as jf (jf.id)}
							<button
								onclick={() => (judgeFilter = jf.id)}
								class="px-3 py-1.5 rounded-full text-xs font-semibold whitespace-nowrap transition-all {judgeFilter === jf.id
									? jf.id === 'target_word'
										? 'bg-purple-500 text-white shadow'
										: jf.id === 'llm_judge'
											? 'bg-blue-500 text-white shadow'
											: jf.id === 'format_break'
												? 'bg-orange-500 text-white shadow'
												: 'bg-[#FF4D00] text-white shadow'
									: isDarkMode
										? 'bg-gray-900 text-gray-400 hover:bg-gray-800 border border-gray-800'
										: 'bg-white text-gray-500 hover:bg-gray-50 border border-gray-200'}"
							>
								{jf.label}
							</button>
						{/each}
					</div>
					{#if matches.length === 0}
						<div
							class={`text-center py-16 md:py-20 rounded-2xl shadow-lg border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
						>
							<p class={`text-lg mb-4 ${isDarkMode ? 'text-gray-400' : 'text-gray-600'}`}>
								아직 활성 매치가 없습니다.
							</p>
							<button
								onclick={() => (activeSection = 'games')}
								class="px-6 py-3 bg-[#FF4D00] text-white rounded-full font-bold hover:bg-[#ff3300] transition-colors"
							>
								첫 게임 시작하기
							</button>
						</div>
					{:else}
						{#if filteredMatchGroups.length === 0}
							<div class={`text-center py-12 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
								<p class="text-sm">해당 유형의 매치가 없습니다.</p>
							</div>
						{:else}
							<div
								class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4 md:gap-6"
							>
								{#each filteredMatchGroups as group (group.gameId)}
									{@const gameUI = allGames.find((g) => g.id === group.gameId)}
									{@const judgeBadge = getJudgeBadgeStyle(gameUI?.judge_type ?? 'unknown')}
								<button
									onclick={() => goToMatch(group.latestMatch.id)}
									class={`group rounded-2xl overflow-hidden shadow-lg hover:shadow-2xl transition-all duration-300 hover:scale-105 active:scale-100 flex flex-col border text-left ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
								>
									<div class="relative aspect-[16/10] bg-gray-200 overflow-hidden">
										{#if gameUI?.image}
											<img
												src={gameUI.image}
												alt={group.gameTitle}
												class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
												onerror={handleImageError(DEFAULT_GAME_THUMBNAIL)}
											/>
										{:else}
											<div
												class={`w-full h-full flex items-center justify-center ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
											>
												<svg
													class="w-10 h-10 text-gray-400"
													fill="none"
													stroke="currentColor"
													viewBox="0 0 24 24"
												>
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="1.5"
														d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
													/>
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="1.5"
														d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
													/>
												</svg>
											</div>
										{/if}
										<div
											class="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
										></div>

										<!-- Match count badge -->
										<div class="absolute top-2 left-2">
											<span
												class="bg-black/60 backdrop-blur-sm px-2 py-0.5 rounded text-xs font-bold text-white"
											>
												{group.total}회
											</span>
										</div>

										<div class="absolute top-2 right-2">
											<span
												class={`backdrop-blur-sm px-2.5 py-1 rounded-lg text-[10px] font-bold shadow-sm ${judgeBadge.classes}`}
											>
												{judgeBadge.label}
											</span>
										</div>

										<div
											class="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
										>
											<div
												class="w-12 h-12 md:w-14 md:h-14 rounded-full bg-[#FF4D00] flex items-center justify-center shadow-xl"
											>
												<svg class="w-5 h-5 md:w-6 md:h-6 fill-white ml-1" viewBox="0 0 24 24">
													<path d="M8 5v14l11-7z" />
												</svg>
											</div>
										</div>
									</div>

									<div class="p-3 md:p-4 flex-1 flex flex-col">
										<h3
											class={`font-bold text-base md:text-lg group-hover:text-[#FF4D00] transition-colors mb-2 line-clamp-1 ${isDarkMode ? 'text-gray-100' : 'text-gray-800'}`}
										>
											{group.gameTitle}
										</h3>
										<!-- Stats pills -->
										<div class="flex flex-wrap gap-1">
											{#if group.active > 0}
												<span
													class={`text-xs font-semibold px-1.5 py-0.5 rounded ${isDarkMode ? 'bg-gray-700 text-gray-400' : 'bg-gray-200 text-gray-600'}`}
												>
													진행 {group.active}
												</span>
											{/if}
											{#if group.won > 0}
												<span
													class={`text-xs font-semibold px-1.5 py-0.5 rounded ${isDarkMode ? 'bg-green-900/60 text-green-300' : 'bg-green-100 text-green-700'}`}
												>
													승리 {group.won}
												</span>
											{/if}
											{#if group.lost > 0}
												<span
													class={`text-xs font-semibold px-1.5 py-0.5 rounded ${isDarkMode ? 'bg-red-900/60 text-red-300' : 'bg-red-100 text-red-700'}`}
												>
													패배 {group.lost}
												</span>
											{/if}
											{#if group.resigned > 0}
												<span
													class={`text-xs font-semibold px-1.5 py-0.5 rounded ${isDarkMode ? 'bg-gray-700 text-gray-400' : 'bg-gray-200 text-gray-600'}`}
												>
													기권 {group.resigned}
												</span>
											{/if}
										</div>
									</div>
								</button>
								{/each}
							</div>
						{/if}
					{/if}
				</div>
			{/if}
		{/if}
	</main>
</div>
{#if showGameModal && selectedGame}
	<div
		class="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4"
		transition:fade={{ duration: 200 }}
		onclick={() => (showGameModal = false)}
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
					onerror={handleImageError(DEFAULT_GAME_THUMBNAIL)}
				/>
				<div
					class="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent"
				></div>

				<div class="absolute top-4 left-4">
					<span
						class={`backdrop-blur-sm px-2.5 py-1 rounded-lg text-[11px] font-bold shadow-sm ${modalJudgeBadge.classes}`}
					>
						{modalJudgeBadge.label}
					</span>
				</div>

				<button
					onclick={() => (showGameModal = false)}
					class="absolute top-4 right-4 w-10 h-10 bg-black/50 hover:bg-black/70 rounded-full flex items-center justify-center text-white transition-colors"
					aria-label="모달 닫기"
				>
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						></path>
					</svg>
				</button>

				<div class="absolute bottom-4 md:bottom-6 left-4 md:left-6">
					<div class="flex gap-2 mb-2 md:mb-3 flex-wrap">
						{#each selectedGame.tags as tag (tag)}
							<span
								class="bg-white/20 backdrop-blur-md px-2 md:px-3 py-0.5 md:py-1 rounded-full text-xs md:text-sm font-bold text-white border border-white/20"
							>
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
				<h3 class="text-xs font-bold text-[#FF4D00] uppercase tracking-widest mb-3">게임 소개</h3>
				<p
					class={`text-base md:text-lg leading-relaxed mb-6 ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}
				>
					{selectedGame.description}
				</p>

				<div class="flex justify-end gap-3">
					<button
						onclick={() => (showGameModal = false)}
						class={`px-5 md:px-6 py-2 md:py-3 rounded-full font-bold transition-colors ${isDarkMode ? 'text-gray-300 hover:bg-gray-900' : 'text-gray-600 hover:bg-gray-100'}`}
					>
						취소
					</button>
					<button
						onclick={() => {
							if (selectedGame) startNewMatch(selectedGame);
							showGameModal = false;
						}}
						class="px-6 md:px-8 py-2 md:py-3 bg-[#FF4D00] text-white rounded-full font-bold hover:bg-[#ff3300] transition-all hover:scale-105 active:scale-95 shadow-lg flex items-center gap-2"
					>
						<svg class="w-5 h-5 fill-current" viewBox="0 0 24 24">
							<path d="M8 5v14l11-7z" />
						</svg>
						게임 시작
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
