<script lang="ts">
	import type { GameUI } from '$lib/features/game/types';

	let {
		game,
		isDarkMode,
		onclick
	}: {
		game: GameUI;
		isDarkMode: boolean;
		onclick: () => void;
	} = $props();

	const judgeBadge = $derived.by(() => {
		switch (game.judge_type) {
			case 'target_word':
				return {
					label: 'Target Word',
					classes: 'bg-purple-500/90 text-white border border-purple-300/40'
				};
			case 'llm_judge':
				return {
					label: 'LLM Judge',
					classes: 'bg-blue-500/90 text-white border border-blue-300/40'
				};
			case 'format_break':
				return {
					label: 'Format Break',
					classes: 'bg-orange-500/90 text-white border border-orange-300/40'
				};
			default:
				return {
					label: 'Unknown',
					classes: 'bg-gray-500/90 text-white border border-gray-300/40'
				};
		}
	});
</script>

<button
	{onclick}
	class={`group rounded-2xl overflow-hidden shadow-lg hover:shadow-2xl transition-all duration-300 hover:scale-105 active:scale-100 flex flex-col border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
>
	<div class="relative aspect-[16/10] bg-gray-200 overflow-hidden">
		<img
			src={game.image}
			alt={game.title}
			class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
			onerror={(e) => {
				const el = e.currentTarget as HTMLImageElement;
				el.onerror = null;
				el.src = 'https://storage.googleapis.com/ollm-assets-prod/default/game_thumbnail.png';
			}}
		/>
		<div
			class="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
		></div>

		<div class="absolute top-2 left-2 flex gap-1 flex-wrap">
			{#each game.tags as tag (tag)}
				<span class="bg-black/60 backdrop-blur-sm px-2 py-0.5 rounded text-xs font-bold text-white">
					{tag}
				</span>
			{/each}
		</div>

		<div class="absolute top-2 right-2">
			<span
				class="{judgeBadge.classes} backdrop-blur-sm px-2.5 py-1 rounded-lg text-[10px] font-bold shadow-sm"
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
			class={`font-bold text-base md:text-lg group-hover:text-[#FF4D00] transition-colors mb-1 line-clamp-1 ${isDarkMode ? 'text-gray-100' : 'text-gray-800'}`}
		>
			{game.title}
		</h3>
		<p class={`text-xs line-clamp-2 flex-1 ${isDarkMode ? 'text-gray-400' : 'text-gray-600'}`}>
			{game.description}
		</p>
	</div>
</button>
