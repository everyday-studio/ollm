<script lang="ts">
	import { fade, fly } from 'svelte/transition';

	let {
		show,
		title,
		description,
		confirmText = '확인',
		cancelText = '취소',
		isDarkMode = false,
		onconfirm,
		oncancel
	}: {
		show: boolean;
		title: string;
		description: string;
		confirmText?: string;
		cancelText?: string;
		isDarkMode?: boolean;
		onconfirm: () => void;
		oncancel: () => void;
	} = $props();

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') oncancel();
	}
</script>

{#if show}
	<div
		class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4"
		transition:fade={{ duration: 200 }}
		onclick={oncancel}
		onkeydown={handleKeydown}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
	>
		<div
			class={`w-full max-w-md rounded-2xl shadow-2xl z-60 p-8 ${isDarkMode ? 'bg-gray-950 border border-gray-800' : 'bg-white'}`}
			transition:fly={{ y: 30, duration: 200 }}
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
			role="presentation"
		>
			<h2 class={`text-2xl font-bold ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>
				{title}
			</h2>
			<p class={`text-sm mt-2 ${isDarkMode ? 'text-gray-400' : 'text-gray-600'}`}>{description}</p>
			<div class="flex justify-end gap-3 mt-6">
				<button
					onclick={oncancel}
					class={`px-5 py-2.5 rounded-full font-semibold transition-colors ${isDarkMode ? 'text-gray-300 hover:bg-gray-800' : 'text-gray-600 hover:bg-gray-100'}`}
				>
					{cancelText}
				</button>
				<button
					onclick={onconfirm}
					class="px-6 py-2.5 bg-[#FF4D00] text-white rounded-full font-semibold hover:bg-[#ff3300] transition-all hover:scale-105 active:scale-95 shadow-lg"
				>
					{confirmText}
				</button>
			</div>
		</div>
	</div>
{/if}
