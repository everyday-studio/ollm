<script lang="ts">
  import { fade, scale } from 'svelte/transition';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { toast } from 'svelte-french-toast';

  import { authApi } from '$lib/features/auth/api';
  import { authStore } from '$lib/features/auth/model';
  import type { AuthResponse } from '$lib/features/auth/types';

  // 상태 변수들
  let showRegisterModal = false;
  let isLoading = false;
  let googleBtnContainer: HTMLDivElement;
  
  // ✅ [추가] 비밀번호 보임/숨김 상태 관리
  let showLoginPassword = false;
  let showRegPassword = false;

  // 로그인용 입력값
  let loginEmail = '';
  let loginPassword = '';

  // 회원가입용 입력값
  let regEmail = '';
  let regPassword = '';

  let errorMessage = '';

  // Guide tooltip / modal
  let showGuide = false;
  let guideStep = 0;

  const guideSteps = [
    {
      icon: '🎮',
      title: '게임 선택',
      desc:
        'Ollm 에 오신 걸 환영합니다! 먼저 플레이하고 싶은 게임을 골라보세요. 각 게임 카드에서 규칙과 컨셉을 확인할 수 있습니다.'
    },
    {
      icon: '💬',
      title: 'AI와 대화',
      desc:
        '매치에서 AI와 번갈아 대화를 주고받습니다. AI는 게임의 상대방이자 심판입니다. 대화 내용이 게임의 승패를 결정하니, 창의적이고 전략적으로 접근해 보세요!'
    },
    {
      icon: '⚖️',
      title: '심판 판정',
      desc:
        '심판은 대화 내용을 바탕으로 승패를 판단합니다. 판정 로직은 게임마다 다르니 규칙을 잘 확인하세요. 승리 조건을 달성하면 매치에서 승리하게 됩니다.'
    },
    {
      icon: '🏆',
      title: '리더보드',
      desc:
        '턴 수와 사용한 토큰 수가 적을수록 상위에 오릅니다. 자신의 기록을 확인하고 다른 플레이어와 경쟁해 보세요 — 번뜩이는 아이디어로 순위 상승이 가능합니다.'
    }
  ];

  function handleToggleGuide() {
    showGuide = !showGuide;
    guideStep = 0;
  }

  // Open guide automatically on client mount (only once per browser via localStorage)
  onMount(() => {
    const key = 'ollm_guide_shown_v1';
    try {
      const shown = localStorage.getItem(key);
      if (!shown) {
        showGuide = true;
        guideStep = 0;
        localStorage.setItem(key, '1');
      }
    } catch (e) {
      // If localStorage is unavailable for any reason, fall back to showing the guide.
      showGuide = true;
      guideStep = 0;
    }

    initGoogleButton();
  });

  // 구글 로그인 콜백
  async function handleGoogleCallback(response: { credential: string }) {
    isLoading = true;
    errorMessage = '';
    try {
      const res = await authApi.googleLogin({ id_token: response.credential });
      const { access_token, id, name, tag, email } = res.data as AuthResponse;
      authStore.loginSuccess(access_token, { id, name, tag, email, role: '', created_at: '', updated_at: '' });
      toast.success('Google 로그인 성공!', { duration: 3000, position: 'top-center', icon: '✅' });
      // eslint-disable-next-line svelte/no-navigation-without-resolve
      await goto('/lobby');
    } catch (err: unknown) {
      const axiosErr = err as { response?: { status?: number } };
      errorMessage = axiosErr.response?.status === 401
        ? 'Google 인증에 실패했습니다. 다시 시도해주세요.'
        : '서버에 문제가 발생했습니다. 잠시 후 다시 시도해주세요.';
      toast.error(errorMessage, { position: 'top-center' });
    } finally {
      isLoading = false;
    }
  }

  // GIS 라이브러리 초기화 및 숨겨진 버튼 렌더링
  // renderButton()은 클릭 시 팝업 다이얼로그를 열어줌 (prompt()의 One Tap과 달리 억제되지 않음)
  function initGoogleButton() {
    if (!GOOGLE_CLIENT_ID) return;

    const tryInit = () => {
      type GIS = { accounts: { id: { initialize: (c: object) => void; renderButton: (el: HTMLElement, o: object) => void; } } };
      const g = (window as unknown as { google?: GIS }).google;
      if (!g?.accounts?.id) {
        setTimeout(tryInit, 200);
        return;
      }
      g.accounts.id.initialize({
        client_id: GOOGLE_CLIENT_ID,
        callback: handleGoogleCallback,
      });
      g.accounts.id.renderButton(googleBtnContainer, { type: 'standard', size: 'large', width: 1 });
    };
    tryInit();
  }

  function handleNextStep() {
    if (guideStep < guideSteps.length - 1) {
      guideStep++;
    } else {
      showGuide = false;
      guideStep = 0;
    }
  }

  function handlePrevStep() {
    if (guideStep > 0) guideStep--;
  }

  // 모달 열기/닫기
  const openModal = () => {
    showRegisterModal = true;
    errorMessage = '';
    regEmail = '';
    regPassword = '';
    showRegPassword = false; // 모달 열 때 비밀번호 숨김 초기화
  };

  const closeModal = () => {
    showRegisterModal = false;
    errorMessage = '';
  };

  // 로그인 처리
  const handleLogin = async () => {
    if (!loginEmail || !loginPassword) {
      errorMessage = "이메일과 비밀번호를 입력해주세요.";
      return;
    }
    
    isLoading = true;
    errorMessage = '';
    let loggedIn = false;

    try {
      const res = await authApi.login({ 
        email: loginEmail, 
        password: loginPassword 
      });

      const { access_token, id, name, tag, email } = res.data as AuthResponse;
      authStore.loginSuccess(access_token, { id, name, tag, email, role: '', created_at: '', updated_at: '' });

      toast.success(`로그인 성공!`, {
        duration: 3000,
        position: 'top-center',
        icon: '✅',
      });

      loggedIn = true;

    } catch (err: unknown) {
      const axiosErr = err as { response?: { status?: number } };
      if (axiosErr.response) {
        const status = axiosErr.response.status;
        if (status === 400 || status === 401 || status === 404) {
          errorMessage = "아이디 또는 비밀번호가 일치하지 않습니다.";
        } else {
          errorMessage = "서버에 문제가 발생했습니다. 잠시 후 다시 시도해주세요.";
        }
      } else {
        errorMessage = "서버와 통신할 수 없습니다.";
      }
      toast.error(errorMessage, { position: 'top-center' });
    } finally {
      isLoading = false;
    }

    // eslint-disable-next-line svelte/no-navigation-without-resolve
    if (loggedIn) await goto('/lobby');
  };

  // 회원가입 처리
  const handleRegister = async () => {
    if (!regEmail || !regPassword) {
      toast.error("이메일과 비밀번호는 필수입니다.", {
        position: "top-center"
      });
      return;
    }

    isLoading = true;
    errorMessage = ''; 

    try {
      await authApi.signup({
        email: regEmail,
        password: regPassword,
      });

      toast.success(`가입 완료! 로그인을 진행해주세요.`, {
        duration: 3000,
        position: 'top-center',
        icon: '👏',
      });

      loginEmail = regEmail;
      loginPassword = '';
      
      closeModal(); 

    } catch (err: unknown) {
      const axiosErr = err as { response?: { status?: number; data?: { error?: string } } };
      if (axiosErr.response) {
        switch (axiosErr.response.status) {
          case 409:
            errorMessage = "이미 존재하는 이메일입니다.";
            break;
          case 400:
            errorMessage = "입력한 형식이 올바르지 않습니다.";
            break;
          case 500:
            errorMessage = "서버에 문제가 발생했습니다. 잠시 후 다시 시도해주세요.";
            break;
          default:
            errorMessage = axiosErr.response.data?.error ?? "회원가입에 실패했습니다.";
        }
      } else {
        errorMessage = "서버와 통신할 수 없습니다.";
      }
      toast.error(errorMessage, { position: 'top-center' });
    } finally {
      isLoading = false;
    }
  };

  // 구글 로그인
  const GOOGLE_CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID || '';

  const handleGoogleLogin = () => {
    if (!GOOGLE_CLIENT_ID) {
      toast.error('Google 로그인이 설정되지 않았습니다.', { position: 'top-center' });
      return;
    }
    if (isLoading) return;
    // 숨겨진 Google 공식 버튼을 programmatic 클릭 → 팝업 다이얼로그 오픈
    googleBtnContainer?.querySelector<HTMLElement>('div[role="button"]')?.click();
  };
</script>

<div class="min-h-screen flex flex-col bg-gray-50 text-gray-900 font-sans">
  
  <div class="flex-1 flex items-center justify-center p-4">
  <div class="w-full max-w-md bg-white rounded-2xl shadow-xl border border-gray-100 overflow-hidden relative p-8 md:p-10">
      
      <div class="text-center mb-8">
        <img 
          src="/logo.png" 
          alt="LLM GAMES Logo" 
          class="mx-auto h-16 w-auto mb-4" 
        />
        <div class="relative inline-flex items-center justify-center gap-1.5 mt-2">
          <p class="text-gray-500 text-sm">LLM 플레이그라운드</p>
          <button
            type="button"
            on:click={handleToggleGuide}
            class="w-5 h-5 rounded-full bg-blue-100 hover:bg-blue-200 text-blue-600 text-xs font-bold flex items-center justify-center transition-all hover:scale-110 cursor-pointer"
            aria-label="이용 가이드 보기"
          >
            ?
          </button>

          <!-- guide tooltip removed; guide will show as modal on page load -->
        </div>
      </div>

      <form on:submit|preventDefault={handleLogin} class="space-y-5">
        <div class="space-y-1.5">
            <label for="login-email" class="block text-sm font-medium text-gray-700">이메일</label>
            <input 
                type="email" 
                id="login-email"
                bind:value={loginEmail}
                placeholder="이메일 주소"
                class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-gray-400"
            />
        </div>
        
        <div class="space-y-1.5">
            <label for="login-password" class="block text-sm font-medium text-gray-700">비밀번호</label>
            <div class="relative">
                <input 
                    type={showLoginPassword ? 'text' : 'password'} 
                    id="login-password"
                    bind:value={loginPassword}
                    placeholder="••••••••"
                    class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-gray-400 pr-10"
                />
                <button 
                    type="button"
                    class="absolute inset-y-0 right-0 flex items-center px-3 text-gray-400 hover:text-gray-600 focus:outline-none cursor-pointer"
                    on:click={() => showLoginPassword = !showLoginPassword}
                >
                    {#if showLoginPassword}
                        <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                        </svg>
                    {:else}
                        <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.542-7a10.059 10.059 0 013.945-5.301m1.966-1.967a9.96 9.96 0 013.631-.732c4.478 0 8.268 2.943 9.542 7a10.059 10.059 0 01-3.23 5.485m-1.55 1.55l-9.32-9.32" />
                        </svg>
                    {/if}
                </button>
            </div>
        </div>

        {#if errorMessage && !showRegisterModal}
            <div class="text-red-500 text-sm font-medium text-center bg-red-50 py-2 rounded animate-pulse">
                {errorMessage}
            </div>
        {/if}

        <button 
            type="submit" 
            disabled={isLoading}
            class="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-3.5 rounded-lg shadow-md hover:shadow-lg transition-all transform active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer flex items-center justify-center"
        >
            {#if isLoading}
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                처리 중...
            {:else}
                로그인
            {/if}
        </button>
      </form>

      <button 
        type="button"
        on:click={openModal}
        class="w-full mt-3 bg-gray-100 hover:bg-gray-200 text-gray-700 font-bold py-3.5 rounded-lg transition-all transform active:scale-[0.98] cursor-pointer"
      >
        새 계정 만들기
      </button>

      <div class="relative my-8">
        <div class="absolute inset-0 flex items-center">
          <div class="w-full border-t border-gray-200"></div>
        </div>
        <div class="relative flex justify-center text-sm">
          <span class="px-2 bg-white text-gray-500">또는</span>
        </div>
      </div>

      <!-- GIS renderButton() 숨겨진 컨테이너 (팝업 트리거용) -->
      <div bind:this={googleBtnContainer} class="absolute w-px h-px overflow-hidden opacity-0" aria-hidden="true"></div>

      <button 
        on:click={handleGoogleLogin}
        disabled={isLoading}
        class="w-full flex items-center justify-center bg-white border border-gray-300 hover:bg-gray-50 text-gray-700 font-semibold py-3 rounded-lg transition-colors disabled:opacity-50 cursor-pointer"
      >
        <svg class="w-5 h-5 mr-3" viewBox="0 0 24 24">
            <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
            <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
            <path fill="#FBBC05" d="M5.84 14.11c-.22-.66-.35-1.36-.35-2.11s.13-1.45.35-2.11V7.05H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.95l2.66-2.84z" />
            <path fill="#EA4335" d="M12 4.63c1.61 0 3.02.56 4.13 1.62L19.16 3.16C17.27 1.4 14.82 0 12 0 7.7 0 3.99 2.47 2.18 7.05l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
        </svg>
        Google 계정으로 계속하기
      </button>
  </div>
  </div>
  
  <div class="hidden sm:flex flex-col items-center gap-1.5 text-xs text-gray-400 py-4">
    <p>&copy; 2026 <a href="https://everydaystudio.xyz" target="_blank" rel="noopener noreferrer" class="text-gray-500 hover:text-blue-500 font-semibold transition-colors">everydaystudio</a> &middot; ollm</p>
    <a href="mailto:everydaystudio365@gmail.com" class="text-[10px] text-gray-400 hover:text-gray-600 transition-colors">everydaystudio365@gmail.com</a>
    <a
      href="https://everydaystudio.xyz"
      target="_blank"
      rel="noopener noreferrer"
      class="mt-1 inline-flex items-center gap-1 px-3 py-1 rounded-full border border-gray-200 text-[11px] text-gray-500 hover:text-blue-600 hover:border-blue-300 hover:bg-blue-50 transition-all"
    >
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" /></svg>
      everydaystudio.xyz
    </a>
  </div>

  {#if showGuide}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-6">
      <div
        class="fixed inset-0 bg-black/50"
        transition:fade
        role="button"
        tabindex="0"
        on:click={() => { showGuide = false; guideStep = 0; }}
        on:keydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') { showGuide = false; guideStep = 0 } }}
      ></div>

      <div class="relative bg-white rounded-xl shadow-2xl w-full max-w-2xl z-60 overflow-hidden" transition:scale={{ duration: 150, start: 0.95 }} role="dialog" aria-modal="true">
        <div class="h-1 bg-gray-100">
          <div class="h-full bg-blue-500" style="width: {((guideStep + 1) / guideSteps.length) * 100}%"></div>
        </div>

        <div class="p-6">
          <button
            type="button"
            aria-label="가이드 닫기"
            on:click={() => { showGuide = false; guideStep = 0; }}
            class="absolute top-3 right-3 text-gray-400 hover:text-gray-600 p-2 rounded-full focus:outline-none"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
          </button>
          <div class="text-4xl mb-4">{guideSteps[guideStep].icon}</div>
          <h3 class="text-lg font-bold text-gray-900 mb-2">
            <span class="text-blue-500">{guideStep + 1}/{guideSteps.length}</span>
            <span class="ml-2">{guideSteps[guideStep].title}</span>
          </h3>
          <p class="text-sm text-gray-600">{guideSteps[guideStep].desc}</p>
        </div>

        <div class="flex border-t border-gray-100">
          {#if guideStep > 0}
            <button type="button" on:click={handlePrevStep} class="flex-1 py-3 text-sm font-semibold text-gray-500 hover:bg-gray-50">이전</button>
          {/if}
          <button type="button" on:click={handleNextStep} class="flex-1 py-3 text-sm font-semibold text-blue-600 hover:bg-blue-50">
            {guideStep < guideSteps.length -1 ? '다음' : '닫기'}
          </button>
        </div>
      </div>
    </div>
  {/if}

  {#if showRegisterModal}
    <div 
        class="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 flex items-center justify-center p-4 cursor-pointer"
        transition:fade={{ duration: 200 }}
        on:click={closeModal}
        on:keydown={(e) => e.key === 'Escape' && closeModal()}
        role="button"
        tabindex="0"
    >
        <div 
            class="bg-white w-full max-w-md rounded-2xl shadow-2xl overflow-hidden relative cursor-default"
            transition:scale={{ duration: 200, start: 0.95 }}
            on:click|stopPropagation
            on:keydown|stopPropagation
            role="dialog"
            tabindex="-1"
        >
            <button 
                on:click={closeModal}
                class="absolute top-4 right-4 text-gray-400 hover:text-gray-600 transition-colors cursor-pointer"
                aria-label="모달 닫기"
            >
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
            </button>

            <div class="p-8">
                <div class="text-center mb-6">
                    <h2 class="text-2xl font-bold text-gray-900">회원가입</h2>
                    <p class="text-gray-500 text-sm mt-1">LLM GAMES의 새로운 사용자가 되어보세요.</p>
                </div>

                <form on:submit|preventDefault={handleRegister} class="space-y-4">
                    <div class="space-y-1.5">
                        <label for="reg-email" class="block text-sm font-medium text-gray-700">이메일</label>
                        <input 
                            type="email" 
                            id="reg-email"
                            bind:value={regEmail}
                            placeholder="이메일 주소"
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        />
                    </div>

                    <div class="space-y-1.5">
                        <label for="reg-password" class="block text-sm font-medium text-gray-700">비밀번호</label>
                        <div class="relative">
                            <input 
                                type={showRegPassword ? 'text' : 'password'} 
                                id="reg-password"
                                bind:value={regPassword}
                                placeholder="••••••••"
                                class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent pr-10"
                            />
                            <button 
                                type="button"
                                class="absolute inset-y-0 right-0 flex items-center px-3 text-gray-400 hover:text-gray-600 focus:outline-none cursor-pointer"
                                on:click={() => showRegPassword = !showRegPassword}
                            >
                                {#if showRegPassword}
                                    <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                    </svg>
                                {:else}
                                    <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.542-7a10.059 10.059 0 013.945-5.301m1.966-1.967a9.96 9.96 0 013.631-.732c4.478 0 8.268 2.943 9.542 7a10.059 10.059 0 01-3.23 5.485m-1.55 1.55l-9.32-9.32" />
                                    </svg>
                                {/if}
                            </button>
                        </div>
                    </div>

                    {#if errorMessage}
                        <div class="text-red-500 text-sm font-medium text-center bg-red-50 py-2 rounded">
                            {errorMessage}
                        </div>
                    {/if}

                    <div class="pt-2">
                        <button 
                            type="submit" 
                            disabled={isLoading}
                            class="w-full bg-green-600 hover:bg-green-700 text-white font-bold py-3.5 rounded-lg shadow-md hover:shadow-lg transition-all transform active:scale-[0.98] disabled:opacity-50 cursor-pointer flex items-center justify-center"
                        >
                            {#if isLoading}
                                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                                가입 중...
                            {:else}
                                가입하기
                            {/if}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
  {/if}

</div>