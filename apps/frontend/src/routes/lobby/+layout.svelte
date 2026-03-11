<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { goto, invalidateAll, onNavigate } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount, setContext } from 'svelte';

	import { authApi } from '$lib/features/auth/api';
	import { authStore } from '$lib/features/auth/model';
	import { ensureSession, resetSession } from '$lib/features/auth/session';
	import { invalidateCache } from '$lib/cache/apiCache';

	let { children } = $props();

	let showLogoutConfirm = $state(false);
	let isDarkMode = $state(true);

	setContext('theme', {
		get isDark() {
			return isDarkMode;
		}
	});

	let currentUserEmail = $derived($authStore?.user?.email ?? 'Guest');
	//let currentUserInitial = $derived(($authStore?.user?.email && $authStore.user.email[0]) ? $authStore.user.email[0].toUpperCase() : 'U');
	let currentPath = $derived($page.url.pathname);

	onMount(async () => {
		const savedTheme = localStorage.getItem('theme');
		isDarkMode = savedTheme !== 'light';

		// Restore session (deduplicated — safe if child pages also call ensureSession)
		const isValidSession = await ensureSession();
		if (!isValidSession) {
			console.warn('Session is invalid or expired, redirecting to login...');
			authStore.logout();
			goto('/login?clear=true');
		}
	});

	// Sync dark mode state → html element class so CSS can target it
	$effect(() => {
		if (isDarkMode) {
			document.documentElement.classList.add('dark');
		} else {
			document.documentElement.classList.remove('dark');
		}
	});

	function toggleTheme() {
		isDarkMode = !isDarkMode;
		localStorage.setItem('theme', isDarkMode ? 'dark' : 'light');
	}

	// ----------------------------------------------------------------
	// View Transitions API — smooth cross-fade between page navigations
	// ----------------------------------------------------------------
	onNavigate((navigation) => {
		if (!document.startViewTransition) return;

		return new Promise((resolve) => {
			document.startViewTransition(async () => {
				resolve();
				await navigation.complete;
			});
		});
	});

	async function handleLogout() {
		try {
			await authApi.logout();
		} catch (e: unknown) {
			const status = (e as { response?: { status?: number } })?.response?.status;
			if (status === 401 || status === 403) {
				try {
					const refreshRes = await authApi.refresh();
					if (refreshRes?.data?.access_token) {
						authStore.updateToken(refreshRes.data.access_token);
					}
					await authApi.logout();
				} catch (refreshErr) {
					console.warn('Refresh or retry logout failed', refreshErr);
				}
			} else {
				console.warn('Logout request failed', e);
			}
		} finally {
			authStore.logout();
			resetSession();
			invalidateCache();

			try {
				await invalidateAll();
			} catch (e) {
				console.warn('invalidateAll failed', e);
			}

			// eslint-disable-next-line svelte/no-navigation-without-resolve
			await goto('/login?clear=true');
		}
	}
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -->
<svelte:head>
	<link rel="preload" as="image" href="/logo.png" />
</svelte:head>

<header
	class={`fixed top-0 left-0 right-0 border-b z-50 h-16 shadow-lg transition-colors ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
>
	<div class="w-full px-6 h-full flex items-center justify-between">
		<div class="flex items-center gap-6">
			<a href="/lobby" class="flex items-center hover:opacity-80 transition">
				<img
					src="/logo.png"
					alt="Ollm Logo"
					width="80"
					height="32"
					class="h-8 w-auto object-contain"
				/>
			</a>

			<nav class="flex items-stretch gap-0 text-sm font-semibold h-16">
				<a
					href="/lobby"
					class={`px-4 h-full flex items-center border-b-4 transition-colors ${currentPath === '/lobby' || currentPath.startsWith('/lobby/match') ? 'text-[#FF4D00] border-[#FF4D00] font-bold' : isDarkMode ? 'text-gray-400 border-transparent hover:text-gray-200 hover:border-gray-600' : 'text-gray-500 border-transparent hover:text-gray-700 hover:border-gray-300'}`}
				>
					메인
				</a>
				<span class={`mx-2 self-center ${isDarkMode ? 'text-gray-700' : 'text-gray-300'}`}>|</span>
				<a
					href="/lobby/mypage"
					class={`px-4 h-full flex items-center border-b-4 transition-colors ${currentPath.startsWith('/lobby/mypage') ? 'text-[#FF4D00] border-[#FF4D00] font-bold' : isDarkMode ? 'text-gray-400 border-transparent hover:text-gray-200 hover:border-gray-600' : 'text-gray-500 border-transparent hover:text-gray-700 hover:border-gray-300'}`}
				>
					마이페이지
				</a>
				<span class={`mx-2 self-center ${isDarkMode ? 'text-gray-700' : 'text-gray-300'}`}>|</span>
				<a
					href="/lobby/leaderboard"
					class={`px-4 h-full flex items-center border-b-4 transition-colors ${currentPath.startsWith('/lobby/leaderboard') ? 'text-[#FF4D00] border-[#FF4D00] font-bold' : isDarkMode ? 'text-gray-400 border-transparent hover:text-gray-200 hover:border-gray-600' : 'text-gray-500 border-transparent hover:text-gray-700 hover:border-gray-300'}`}
				>
					리더보드
				</a>
			</nav>
		</div>

		<div class="flex items-center gap-3">
			<!-- Theme Toggle Button -->
			<button
				type="button"
				onclick={toggleTheme}
				class={`p-2 rounded-lg transition-colors ${isDarkMode ? 'text-gray-400 hover:text-yellow-400 hover:bg-gray-900' : 'text-gray-600 hover:text-blue-600 hover:bg-gray-100'}`}
				title={isDarkMode ? '라이트 모드로 전환' : '다크 모드로 전환'}
				aria-label="테마 전환"
			>
				{#if isDarkMode}
					<!-- Sun Icon (for switching to light mode) -->
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
						/>
					</svg>
				{:else}
					<!-- Moon Icon (for switching to dark mode) -->
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
						/>
					</svg>
				{/if}
			</button>

			<div class="relative group">
				<a
					href="/lobby/mypage"
					class={`flex items-center gap-2 rounded-lg px-2 py-1.5 transition-colors cursor-pointer ${isDarkMode ? 'group-hover:bg-gray-900' : 'group-hover:bg-gray-100'}`}
				>
					<img
						src={$authStore.user?.id
							? `https://storage.googleapis.com/ollm-assets-prod/user/${$authStore.user.id}.png`
							: 'https://storage.googleapis.com/ollm-assets-prod/user_profile_default.png'}
						alt="프로필"
						class="w-8 h-8 rounded-full object-cover shadow-sm"
						onerror={(e) => {
							const el = e.currentTarget as HTMLImageElement;
							el.onerror = null;
							el.src = 'https://storage.googleapis.com/ollm-assets-prod/user_profile_default.png';
						}}
					/>

					<div class="flex flex-col">
						<span
							class={`text-xs font-semibold leading-tight ${isDarkMode ? 'text-gray-100' : 'text-gray-800'}`}
						>
							{$authStore.user?.name || '플레이어'}<span
								class={`ml-1 font-normal ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}
								>#{$authStore.user?.tag ?? ''}</span
							>
						</span>
						<span class={`text-[10px] ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>
							{currentUserEmail}
						</span>
					</div>
				</a>

				<div
					class={`absolute right-0 top-full mt-0 w-56 rounded-2xl border shadow-xl opacity-0 translate-y-1 pointer-events-none transition-all duration-150 group-hover:opacity-100 group-hover:translate-y-0 group-hover:pointer-events-auto ${isDarkMode ? 'border-gray-800 bg-gray-950' : 'border-gray-200 bg-white'}`}
				>
					<div class={`px-4 py-3 border-b ${isDarkMode ? 'border-gray-800' : 'border-gray-200'}`}>
						<div class={`text-sm font-semibold ${isDarkMode ? 'text-gray-100' : 'text-gray-800'}`}>
							{$authStore.user?.name || '플레이어'}<span
								class={`ml-1 font-normal text-xs ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}
								>#{$authStore.user?.tag ?? ''}</span
							>
						</div>
						<div class={`text-xs font-mono ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>
							{currentUserEmail}
						</div>
					</div>
					<div class="py-2">
						<a
							href="/lobby/mypage"
							class={`flex items-center px-4 py-2 text-sm transition-colors ${isDarkMode ? 'text-gray-300 hover:bg-gray-900 hover:text-white' : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'}`}
							>마이페이지</a
						>
						<a
							href="/lobby/leaderboard"
							class={`flex items-center px-4 py-2 text-sm transition-colors ${isDarkMode ? 'text-gray-300 hover:bg-gray-900 hover:text-white' : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'}`}
							>리더보드</a
						>
						{#if $authStore.user?.role === 'Admin'}
							<a
								href="http://localhost:8080/admin"
								target="_blank"
								rel="noopener noreferrer"
								class={`flex items-center px-4 py-2 text-sm transition-colors ${isDarkMode ? 'text-purple-400 hover:bg-gray-900 hover:text-purple-300' : 'text-purple-600 hover:bg-gray-50 hover:text-purple-700'}`}
								>관리자 페이지</a
							>
						{/if}
					</div>
				</div>
			</div>

			<button
				type="button"
				onclick={() => (showLogoutConfirm = true)}
				class={`p-2 rounded-lg transition-colors cursor-pointer ${isDarkMode ? 'text-gray-400 hover:text-red-400 hover:bg-gray-900' : 'text-gray-600 hover:text-red-600 hover:bg-gray-100'}`}
				title="로그아웃"
				aria-label="로그아웃"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
					/>
				</svg>
			</button>
		</div>
	</div>
</header>

<main class="pt-16">
	{@render children()}
</main>

{#if showLogoutConfirm}
	<div
		class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 cursor-pointer"
		transition:fade={{ duration: 200 }}
		onclick={() => (showLogoutConfirm = false)}
		onkeydown={(e) => e.key === 'Escape' && (showLogoutConfirm = false)}
		role="button"
		tabindex="0"
	>
		<div
			class={`w-full max-w-md rounded-2xl shadow-2xl overflow-hidden relative cursor-default z-60 border transition-colors ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
			transition:scale={{ duration: 200, start: 0.95 }}
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
			role="dialog"
			tabindex="-1"
		>
			<button
				onclick={() => (showLogoutConfirm = false)}
				class={`absolute top-4 right-4 transition-colors cursor-pointer ${isDarkMode ? 'text-gray-500 hover:text-gray-300' : 'text-gray-400 hover:text-gray-600'}`}
				aria-label="모달 닫기"
			>
				<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M6 18L18 6M6 6l12 12"
					></path></svg
				>
			</button>

			<div class="p-8">
				<div class="text-center mb-6">
					<h2 class={`text-2xl font-bold ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>
						로그아웃
					</h2>
					<p class={`text-sm mt-1 ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>
						계정에서 로그아웃하시겠습니까?
					</p>
				</div>

				<div class={`text-sm mb-6 text-center ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}>
					로그아웃 하시려는 계정: <span class="font-mono">{currentUserEmail}</span>
				</div>

				<div class="flex justify-end gap-3">
					<button
						class={`px-4 py-2 rounded transition-colors ${isDarkMode ? 'bg-gray-800 text-gray-200 hover:bg-gray-700' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}`}
						onclick={() => (showLogoutConfirm = false)}>취소</button
					>
					<button
						class={`px-4 py-2 rounded transition-colors ${isDarkMode ? 'bg-red-600 text-white hover:bg-red-700' : 'bg-red-500 text-white hover:bg-red-600'}`}
						onclick={async () => {
							await handleLogout();
							showLogoutConfirm = false;
						}}>로그아웃</button
					>
				</div>
			</div>
		</div>
	</div>
{/if}
