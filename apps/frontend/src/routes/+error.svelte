<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	let isDarkMode = $state(true);

	onMount(() => {
		const saved = localStorage.getItem('theme');
		isDarkMode = saved !== 'light';
	});

	let status = $derived($page.status);
	let message = $derived($page.error?.message ?? '알 수 없는 오류가 발생했습니다.');

	let title = $derived.by(() => {
		switch (status) {
			case 404:
				return '열심히 찾아봤지만 여기엔 아무것도 없네요. 도롱뇽이 길을 잃었거나, 당신이 잘못된 지도를 보고 있는 것 같아요.';
			case 403:
				return '접근 권한이 없습니다';
			case 500:
				return '서버 오류가 발생했습니다';
			default:
				return '오류가 발생했습니다';
		}
	});

	let emoji = $derived.by(() => {
		switch (status) {
			case 404:
				return '🔍';
			case 403:
				return '🔒';
			case 500:
				return '⚙️';
			default:
				return '⚠️';
		}
	});
</script>

<div
	class={`min-h-screen flex items-center justify-center p-4 transition-colors ${isDarkMode ? 'bg-gradient-to-br from-black to-gray-950 text-white' : 'bg-gradient-to-br from-gray-50 to-gray-100 text-gray-900'}`}
>
	<div class="text-center max-w-md mx-auto">
		{#if status === 404}
			<div class="mb-6">
				<img src="/Gemini_Generated_Image_lvz09rlvz09rlvz0.svg" alt="404 Lost Dinosaur" class="w-auto h-auto mx-auto" />
			</div>
		{:else}
			<div class="text-7xl mb-6">{emoji}</div>
		{/if}

		<h1
			class={`text-8xl font-black mb-4 ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}
		>
			{status}
		</h1>

		<h2
			class={`text-xl font-bold mb-3 ${isDarkMode ? 'text-gray-300' : 'text-gray-700'}`}
		>
			{title}
		</h2>

		<p
			class={`text-sm mb-8 ${isDarkMode ? 'text-gray-500' : 'text-gray-500'}`}
		>
			{message}
		</p>

		<div class="flex flex-col sm:flex-row gap-3 justify-center">
			<a
				href="/lobby"
				class="px-6 py-3 bg-[#FF4D00] text-white rounded-full font-bold hover:bg-[#ff3300] transition-all hover:scale-105 active:scale-95 shadow-lg"
			>
				로비로 돌아가기
			</a>
			<button
				onclick={() => history.back()}
				class={`px-6 py-3 rounded-full font-bold transition-all hover:scale-105 active:scale-95 ${isDarkMode ? 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700' : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200 shadow'}`}
			>
				이전 페이지
			</button>
		</div>
	</div>
</div>
