<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { getContext } from 'svelte';

	const theme = getContext<{ isDark: boolean }>('theme');
	let isDarkMode = $derived(theme.isDark);

	interface Achievement {
		id: string;
		icon: string;
		title: string;
		description: string;
		locked: boolean;
	}

	const achievements: Achievement[] = [
		{
			id: 'first_win',
			icon: '🏆',
			title: '첫 승리',
			description: 'AI의 방어를 뚫고 첫 번째 매치에서 승리하세요.',
			locked: true
		},
		{
			id: 'speed_run',
			icon: '⚡',
			title: '스피드 러너',
			description: '3턴 이내에 매치를 승리로 이끄세요.',
			locked: true
		},
		{
			id: 'play_all',
			icon: '🎮',
			title: '올라운더',
			description: '모든 게임 유형을 최소 1회 이상 플레이하세요.',
			locked: true
		},
		{
			id: 'streak_3',
			icon: '🔥',
			title: '연승 행진',
			description: '3연속 승리를 달성하세요.',
			locked: true
		},
		{
			id: 'token_master',
			icon: '💎',
			title: '토큰 마스터',
			description: '최소한의 토큰으로 매치를 승리하세요.',
			locked: true
		},
		{
			id: 'explorer',
			icon: '🔍',
			title: '탐험가',
			description: '10개 이상의 매치를 플레이하세요.',
			locked: true
		}
	];
</script>

<div
	class={`h-[calc(100vh-64px)] overflow-y-auto transition-colors ${isDarkMode ? 'bg-black text-gray-100' : 'bg-gray-50 text-gray-900'}`}
>
	<main class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		<!-- Header -->
		<div class="mb-8" in:fly={{ y: -20, duration: 300 }}>
			<h1 class={`text-3xl font-black ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>
				🏅 업적
			</h1>
			<p class={`text-sm mt-2 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
				도전과제를 완료하고 업적을 달성하세요. 곧 업데이트됩니다!
			</p>
		</div>

		<!-- Coming Soon Banner -->
		<div
			class={`rounded-2xl border p-6 mb-8 text-center ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
			in:fade={{ duration: 300, delay: 100 }}
		>
			<div class="text-5xl mb-3">🚧</div>
			<h2 class={`text-lg font-bold mb-1 ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}>
				준비 중인 기능입니다
			</h2>
			<p class={`text-sm ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
				업적 시스템은 현재 개발 중입니다. 아래 미리보기로 앞으로 추가될 업적들을 확인해보세요!
			</p>
		</div>

		<!-- Achievement Grid (Preview) -->
		<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
			{#each achievements as achievement, i (achievement.id)}
				<div
					class={`relative rounded-2xl border p-5 transition-all ${
						isDarkMode
							? 'bg-gray-950 border-gray-800'
							: 'bg-white border-gray-200'
					}`}
					in:fly={{ y: 20, duration: 300, delay: 150 + i * 50 }}
				>
					<div class="flex items-start gap-4">
						<div
							class={`w-12 h-12 rounded-xl flex items-center justify-center text-2xl shrink-0 ${
								isDarkMode ? 'bg-gray-800' : 'bg-gray-100'
							}`}
							style="filter: grayscale(1) opacity(0.5);"
						>
							{achievement.icon}
						</div>
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2">
								<h3
									class={`font-bold text-sm ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}
								>
									{achievement.title}
								</h3>
								<span
									class={`text-[10px] font-semibold px-2 py-0.5 rounded-full ${
										isDarkMode
											? 'bg-gray-800 text-gray-500'
											: 'bg-gray-100 text-gray-400'
									}`}
								>
									잠김
								</span>
							</div>
							<p
								class={`text-xs mt-1 ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}
							>
								{achievement.description}
							</p>
						</div>
					</div>
				</div>
			{/each}
		</div>
	</main>
</div>
