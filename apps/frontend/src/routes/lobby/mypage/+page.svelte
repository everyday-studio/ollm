<script lang="ts">
  import { fade, fly } from 'svelte/transition';
  import { onMount, getContext } from 'svelte';
  import { goto } from '$app/navigation';
  import toast from 'svelte-french-toast';

  import { authApi } from '$lib/features/auth/api';
  import { authStore } from '$lib/features/auth/model';
  import { ensureSession } from '$lib/features/auth/session';
  import type { User } from '$lib/features/auth/types';

  const theme = getContext<{ isDark: boolean }>('theme');
  let isDarkMode = $derived(theme.isDark);

  // ----------------------------------------------------------------
  // State
  // ----------------------------------------------------------------
  let user = $state<User | null>(null);
  let isLoading = $state(true);
  let isEditing = $state(false);
  let isSaving = $state(false);
  let nicknameInput = $state('');
  let nicknameError = $state('');

  // Derived
  let memberSince = $derived(
    user ? new Date(user.created_at).toLocaleDateString('ko-KR', {
      year: 'numeric', month: 'long', day: 'numeric'
    }) : ''
  );

  // ----------------------------------------------------------------
  // Lifecycle
  // ----------------------------------------------------------------
  onMount(async () => {
    await ensureSession();

    // Fetch full profile
    try {
      const res = await authApi.getMe();
      user = res.data;
      authStore.updateUser(res.data);
    } catch {
      toast.error('프로필을 불러오지 못했습니다.');
    } finally {
      isLoading = false;
    }
  });

  // ----------------------------------------------------------------
  // Nickname editing
  // ----------------------------------------------------------------
  function startEditing() {
    nicknameInput = user?.name ?? '';
    nicknameError = '';
    isEditing = true;
  }

  function cancelEditing() {
    isEditing = false;
    nicknameError = '';
  }

  async function saveNickname() {
    const trimmed = nicknameInput.trim();
    if (trimmed.length < 2 || trimmed.length > 20) {
      nicknameError = '닉네임은 2자 이상 20자 이하로 입력해주세요.';
      return;
    }

    isSaving = true;
    nicknameError = '';

    try {
      const res = await authApi.updateNickname({ name: trimmed });
      user = res.data;
      authStore.updateUser(res.data);
      isEditing = false;
      toast.success('닉네임이 변경되었습니다.');
    } catch (e: any) {
      if (e?.response?.status === 400) {
        nicknameError = '유효하지 않은 닉네임입니다.';
      } else {
        nicknameError = '변경에 실패했습니다. 다시 시도해주세요.';
      }
    } finally {
      isSaving = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.isComposing) {
      e.preventDefault();
      saveNickname();
    } else if (e.key === 'Escape') {
      cancelEditing();
    }
  }
</script>

<div class={`h-[calc(100vh-64px)] overflow-y-auto transition-colors ${isDarkMode ? 'bg-gradient-to-br from-black to-gray-950' : 'bg-gradient-to-br from-gray-50 to-gray-100'}`}>
  <main class="max-w-[1800px] mx-auto px-4 py-6 md:px-8 md:py-10 lg:px-10 lg:py-12">

    {#if isLoading}
      <!-- Skeleton: Header -->
      <div class="mb-10">
        <div class={`h-10 w-40 rounded-lg skeleton mb-2 ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
        <div class={`h-4 w-64 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
      </div>
      <!-- Skeleton: Profile card -->
      <div class="max-w-2xl space-y-6">
        <div class={`rounded-2xl border overflow-hidden shadow-lg ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}>
          <div class={`px-6 py-8 md:px-8 flex flex-col sm:flex-row items-center sm:items-start gap-5 border-b ${isDarkMode ? 'border-gray-800' : 'border-gray-100'}`}>
            <div class={`w-20 h-20 rounded-full shrink-0 skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
            <div class="flex-1 min-w-0 space-y-3">
              <div class={`h-7 w-40 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
              <div class={`h-4 w-56 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
            </div>
          </div>
          <div class="px-6 md:px-8 py-6 space-y-4">
            {#each Array(3) as _}
              <div class="flex items-center justify-between">
                <div class={`h-4 w-20 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
                <div class={`h-4 w-32 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}></div>
              </div>
            {/each}
          </div>
        </div>
      </div>
    {:else if user}
      <!-- Header -->
      <div class="mb-10">
        <h1 class={`text-3xl md:text-4xl font-black mb-2 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>마이페이지</h1>
        <p class={`text-sm ${isDarkMode ? 'text-gray-500' : 'text-gray-500'}`}>프로필 정보를 확인하고 수정할 수 있습니다</p>
      </div>

      <div class="max-w-2xl space-y-6">
        <!-- Profile Card -->
        <div
          class={`rounded-2xl border overflow-hidden shadow-lg transition-colors ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
          in:fly={{ y: 20, duration: 300 }}
        >
          <!-- Avatar + Name header -->
          <div class={`px-6 py-8 md:px-8 flex flex-col sm:flex-row items-center sm:items-start gap-5 border-b ${isDarkMode ? 'border-gray-800' : 'border-gray-100'}`}>
            <!-- Avatar -->
            <img
              src="https://storage.googleapis.com/ollm-assets-prod/user_profile_default.png"
              alt="프로필"
              class="w-20 h-20 rounded-full object-cover shadow-lg shrink-0"
            />

            <!-- Name area -->
            <div class="flex-1 min-w-0 text-center sm:text-left">
              {#if isEditing}
                <div class="flex flex-col gap-2">
                  <div class="flex items-center gap-2">
                    <input
                      bind:value={nicknameInput}
                      onkeydown={handleKeydown}
                      maxlength={20}
                      class={`flex-1 min-w-0 text-xl font-bold rounded-xl px-3 py-2 outline-none transition-colors ${
                        isDarkMode
                          ? 'bg-gray-800 text-gray-100 ring-1 ring-gray-700 focus:ring-[#FF4D00]/50'
                          : 'bg-gray-100 text-gray-900 ring-1 ring-gray-200 focus:ring-[#FF4D00]/40'
                      }`}
                      placeholder="닉네임 입력"
                    />
                    <button
                      onclick={saveNickname}
                      disabled={isSaving}
                      class="shrink-0 px-4 py-2 bg-[#FF4D00] text-white rounded-xl text-sm font-semibold hover:bg-[#ff3300] transition-colors disabled:opacity-50"
                    >
                      {#if isSaving}
                        <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                      {:else}
                        저장
                      {/if}
                    </button>
                    <button
                      onclick={cancelEditing}
                      class={`shrink-0 px-4 py-2 rounded-xl text-sm font-semibold transition-colors ${
                        isDarkMode ? 'text-gray-400 hover:bg-gray-800' : 'text-gray-500 hover:bg-gray-100'
                      }`}
                    >
                      취소
                    </button>
                  </div>
                  {#if nicknameError}
                    <p class="text-xs text-red-400" transition:fade={{ duration: 150 }}>{nicknameError}</p>
                  {/if}
                  <p class={`text-xs ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}>
                    {nicknameInput.length}/20자 (최소 2자)
                  </p>
                </div>
              {:else}
                <div class="flex items-center gap-3 justify-center sm:justify-start">
                  <h2 class={`text-2xl font-bold ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>
                    {user.name || '이름없는올름'}<span class={`ml-2 text-base font-normal ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>#{user.tag}</span>
                  </h2>
                  <button
                    onclick={startEditing}
                    class={`p-1.5 rounded-lg transition-colors ${
                      isDarkMode ? 'text-gray-500 hover:text-gray-300 hover:bg-gray-800' : 'text-gray-400 hover:text-gray-600 hover:bg-gray-100'
                    }`}
                    title="닉네임 변경"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/>
                    </svg>
                  </button>
                </div>
                <p class={`text-sm mt-1 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>{user.email}</p>
              {/if}
            </div>
          </div>

          <!-- Info rows -->
          <div class={`divide-y ${isDarkMode ? 'divide-gray-800' : 'divide-gray-100'}`}>
            <div class="flex items-center justify-between px-6 py-4 md:px-8">
              <span class={`text-sm font-medium ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>이메일</span>
              <span class={`text-sm font-mono ${isDarkMode ? 'text-gray-300' : 'text-gray-700'}`}>{user.email}</span>
            </div>
            <div class="flex items-center justify-between px-6 py-4 md:px-8">
              <span class={`text-sm font-medium ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>역할</span>
              <span class={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold ${
                user.role === 'Admin'
                  ? 'bg-purple-500/20 text-purple-400'
                  : 'bg-blue-500/20 text-blue-400'
              }`}>
                {user.role === 'Admin' ? '관리자' : '플레이어'}
              </span>
            </div>
            <div class="flex items-center justify-between px-6 py-4 md:px-8">
              <span class={`text-sm font-medium ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>가입일</span>
              <span class={`text-sm ${isDarkMode ? 'text-gray-300' : 'text-gray-700'}`}>{memberSince}</span>
            </div>
          </div>
        </div>

        <!-- Quick Links -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4" in:fly={{ y: 20, duration: 300, delay: 100 }}>
          <button
            onclick={() => goto('/lobby')}
            class={`flex items-center gap-4 p-5 rounded-2xl border transition-all group ${
              isDarkMode
                ? 'bg-gray-950 border-gray-800 hover:border-gray-700 hover:bg-gray-900'
                : 'bg-white border-gray-200 hover:border-gray-300 hover:bg-gray-50'
            }`}
          >
            <div class={`w-10 h-10 rounded-xl flex items-center justify-center shrink-0 ${
              isDarkMode ? 'bg-green-500/10 text-green-400' : 'bg-green-50 text-green-600'
            }`}>
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
            <div class="text-left">
              <div class={`text-sm font-semibold ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}>게임 플레이</div>
              <div class={`text-xs mt-0.5 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>로비로 돌아가기</div>
            </div>
          </button>

          <button
            onclick={() => goto('/lobby/leaderboard')}
            class={`flex items-center gap-4 p-5 rounded-2xl border transition-all group ${
              isDarkMode
                ? 'bg-gray-950 border-gray-800 hover:border-gray-700 hover:bg-gray-900'
                : 'bg-white border-gray-200 hover:border-gray-300 hover:bg-gray-50'
            }`}
          >
            <div class={`w-10 h-10 rounded-xl flex items-center justify-center shrink-0 ${
              isDarkMode ? 'bg-yellow-500/10 text-yellow-400' : 'bg-yellow-50 text-yellow-600'
            }`}>
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
              </svg>
            </div>
            <div class="text-left">
              <div class={`text-sm font-semibold ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}>리더보드</div>
              <div class={`text-xs mt-0.5 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>순위 확인하기</div>
            </div>
          </button>
        </div>
      </div>
    {:else}
      <!-- Error state -->
      <div class="flex flex-col items-center justify-center h-[400px]">
        <p class={`text-lg font-semibold mb-4 ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>프로필을 불러올 수 없습니다</p>
        <button
          onclick={() => goto('/lobby')}
          class="px-5 py-2.5 bg-[#FF4D00] text-white rounded-lg font-semibold text-sm hover:bg-[#ff3300] transition-colors"
        >
          로비로 돌아가기
        </button>
      </div>
    {/if}
  </main>
</div>
