<script lang="ts">
  import { fade, scale } from 'svelte/transition';
  import { goto, invalidateAll } from '$app/navigation';
  import { page } from '$app/stores';

  import { authApi } from '$lib/features/auth/api';
  import { authStore } from '$lib/features/auth/model';

  let { children } = $props();

  let showLogoutConfirm = $state(false);

  let currentUserEmail = $derived($authStore?.user?.email ?? 'Guest');
  let currentUserInitial = $derived(($authStore?.user?.email && $authStore.user.email[0]) ? $authStore.user.email[0].toUpperCase() : 'U');
  let currentPath = $derived($page.url.pathname);

  async function handleLogout() {
    try {
      await authApi.logout();
    } catch (e: any) {
      const status = e?.response?.status;
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

      try {
        await invalidateAll();
      } catch (e) {
        console.warn('invalidateAll failed', e);
      }

      await goto('/login');
    }
  }
</script>

<svelte:head>
  <link rel="preload" as="image" href="/logo.png" />
</svelte:head>

<header class="fixed top-0 left-0 right-0 bg-white border-b border-gray-200 z-50 h-24">
  <div class="max-w-6xl mx-auto px-6 h-full flex items-center justify-between">
    <div class="flex flex-wrap items-center gap-6">
      <a href="/lobby" class="flex items-center gap-3 hover:opacity-90 transition">
        <img
          src="/logo.png"
          alt="Ollm Logo"
          width="140"
          height="56"
          class="h-14 w-auto object-contain"
        />
      </a>

      <nav class="flex items-center gap-0 text-sm font-semibold">
        <a
          href="/lobby"
          class={`px-3 py-2 border-b-2 transition-colors ${currentPath.startsWith('/lobby') ? 'text-[#FF4D00] border-[#FF4D00]' : 'text-gray-500 border-transparent hover:text-gray-700 hover:border-gray-300'}`}
        >
          Home
        </a>
        <span class="mx-2 text-gray-300">|</span>
        <a
          href="/mypage"
          class={`px-3 py-2 border-b-2 transition-colors ${currentPath.startsWith('/mypage') ? 'text-[#FF4D00] border-[#FF4D00]' : 'text-gray-500 border-transparent hover:text-gray-700 hover:border-gray-300'}`}
        >
          My Page
        </a>
        <span class="mx-2 text-gray-300">|</span>
        <a
          href="/leaderboard"
          class={`px-3 py-2 border-b-2 transition-colors ${currentPath.startsWith('/leaderboard') ? 'text-[#FF4D00] border-[#FF4D00]' : 'text-gray-500 border-transparent hover:text-gray-700 hover:border-gray-300'}`}
        >
          Leaderboard
        </a>
      </nav>
    </div>

    <div class="flex items-center gap-4">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-full bg-gradient-to-tr from-green-400 to-green-600 flex items-center justify-center text-white font-bold shadow-sm">
          {currentUserInitial}
        </div>

        <div class="flex flex-col">
          <span class="text-sm font-bold text-gray-800 leading-tight">
            {$authStore.user?.name || 'Player'}
          </span>
          <span class="text-[10px] text-gray-500 font-mono">
            {currentUserEmail}
          </span>
        </div>
      </div>

      <button
        type="button"
        onclick={() => showLogoutConfirm = true}
        class="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors cursor-pointer"
        title="Logout"
        aria-label="Logout"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
        </svg>
      </button>
    </div>
  </div>
</header>

<main class="pt-24 pb-12">
  {@render children()}
</main>

{#if showLogoutConfirm}
  <div
    class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 cursor-pointer"
    transition:fade={{ duration: 200 }}
    onclick={() => showLogoutConfirm = false}
    onkeydown={(e) => e.key === 'Escape' && (showLogoutConfirm = false)}
    role="button"
    tabindex="0"
  >
    <div
      class="bg-white w-full max-w-md rounded-2xl shadow-2xl overflow-hidden relative cursor-default z-60"
      transition:scale={{ duration: 200, start: 0.95 }}
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog"
      tabindex="-1"
    >
      <button
        onclick={() => showLogoutConfirm = false}
        class="absolute top-4 right-4 text-gray-400 hover:text-gray-600 transition-colors cursor-pointer"
        aria-label="모달 닫기"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
      </button>

      <div class="p-8">
        <div class="text-center mb-6">
          <h2 class="text-2xl font-bold text-gray-900">로그아웃</h2>
          <p class="text-gray-500 text-sm mt-1">계정에서 로그아웃하시겠습니까?</p>
        </div>

        <div class="text-sm text-gray-600 mb-6 text-center">
          로그아웃 하시려는 계정: <span class="font-mono">{currentUserEmail}</span>
        </div>

        <div class="flex justify-end gap-3">
          <button class="px-4 py-2 rounded bg-gray-100 text-gray-700" onclick={() => showLogoutConfirm = false}>취소</button>
          <button class="px-4 py-2 rounded bg-red-500 text-white" onclick={async () => { await handleLogout(); showLogoutConfirm = false; }}>로그아웃</button>
        </div>
      </div>
    </div>
  </div>
{/if}