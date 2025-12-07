<script lang="ts">
  import { user } from '$lib/stores';
  import { fade, scale } from 'svelte/transition'; // 팝업 애니메이션용
  import { goto } from '$app/navigation'; // 👈 [추가] 라우팅 함수 가져오기
  import { toast } from 'svelte-french-toast';

  // 상태 변수들
  let showRegisterModal = false; // 회원가입 팝업 표시 여부
  let isLoading = false;
  
  // 로그인용 입력값
  let loginEmail = '';
  let loginPassword = '';

  // 회원가입용 입력값
  let regEmail = '';
  let regNickname = '';
  let regPassword = '';

  let errorMessage = '';

  // 모달 열기/닫기
  const openModal = () => {
    showRegisterModal = true;
    errorMessage = '';
    // 입력값 초기화
    regEmail = '';
    regNickname = '';
    regPassword = '';
  };

  const closeModal = () => {
    showRegisterModal = false;
    errorMessage = '';
  };

  // [MOCK] 로그인 처리
  const handleLogin = async () => {
    if (!loginEmail || !loginPassword) {
      errorMessage = "이메일과 비밀번호를 입력해주세요.";
      return;
    }
    
    isLoading = true;
    errorMessage = '';

    setTimeout(() => {
      isLoading = false;
      user.set({ email: loginEmail, nickname: 'Player1' });
      toast.success(`로그인 성공!`, {
        duration: 3000,       // 3초 뒤 사라짐
        position: 'top-center', // 화면 상단 중앙에 표시
        icon: '✅',           // 아이콘 커스텀 (선택사항)
      });
      // alert("로그인되었습니다."); // 👈 알림 대신 이동
      goto('/lobby'); // 👈 [수정] 게임 선택 화면으로 이동
    }, 1000);
  };

  // [회원가입 처리]
  const handleRegister = async () => {
    // 유효성 검사 실패 시 에러 토스트 (선택사항)
    if (!regEmail || !regNickname || !regPassword) {
      toast.error("모든 정보를 입력해주세요.", {
        position: "top-center"
      });
      return;
    }

    isLoading = true;

    setTimeout(() => {
      isLoading = false;
      
      // 2. alert 대신 toast.success 사용
      toast.success(`환영합니다, ${regNickname}님!\n로그인을 진행해주세요.`, {
        duration: 3000,       // 3초 뒤 사라짐
        position: 'top-center', // 화면 상단 중앙에 표시
        icon: '👏',           // 아이콘 커스텀 (선택사항)
      });

      // 사용자 편의: 이메일 자동 입력
      loginEmail = regEmail;
      loginPassword = '';
      
      closeModal(); // 회원가입 모달 닫기 -> 로그인 화면이 보임
    }, 1000);
  };

  // [MOCK] 구글 로그인 처리
  const handleGoogleLogin = () => {
    isLoading = true;
    setTimeout(() => {
      isLoading = false;
      user.set({ email: "google@example.com", nickname: "GoogleUser" });
        toast.success(`구글 로그인 성공!`, {
        duration: 3000,       // 3초 뒤 사라짐
        position: 'top-center', // 화면 상단 중앙에 표시
        icon: '✅',           // 아이콘 커스텀 (선택사항)
      });
      goto('/lobby'); // 👈 게임 선택 화면으로 이동
    }, 1000);
  };
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50 text-gray-900 font-sans p-4">
  
  <div class="w-full max-w-md bg-white rounded-2xl shadow-xl border border-gray-100 overflow-hidden relative p-8 md:p-10">
      
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-900 tracking-tight">LLM GAMES</h1>
        <p class="text-gray-500 mt-2 text-sm">프롬프트 인젝션 플레이그라운드</p>
      </div>

      <form on:submit|preventDefault={handleLogin} class="space-y-5">
        <div class="space-y-1.5">
            <label for="login-email" class="block text-sm font-medium text-gray-700">이메일</label>
            <input 
                type="email" 
                id="login-email"
                bind:value={loginEmail}
                placeholder="name@example.com"
                class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-gray-400"
            />
        </div>
        <div class="space-y-1.5">
            <label for="login-password" class="block text-sm font-medium text-gray-700">비밀번호</label>
            <input 
                type="password" 
                id="login-password"
                bind:value={loginPassword}
                placeholder="••••••••"
                class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-gray-400"
            />
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
  
  <div class="absolute bottom-4 text-center text-xs text-gray-400">
    &copy; 2025 LLM GAMES. All rights reserved.
  </div>

  {#if showRegisterModal}
    <div 
        class="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 flex items-center justify-center p-4 cursor-pointer"
        transition:fade={{ duration: 200 }}
        on:click={closeModal}
    >
        <div 
            class="bg-white w-full max-w-md rounded-2xl shadow-2xl overflow-hidden relative cursor-default"
            transition:scale={{ duration: 200, start: 0.95 }}
            on:click|stopPropagation
        >
            <button 
                on:click={closeModal}
                class="absolute top-4 right-4 text-gray-400 hover:text-gray-600 transition-colors cursor-pointer"
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
                            placeholder="name@example.com"
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        />
                    </div>

                    <div class="space-y-1.5">
                        <label for="reg-nickname" class="block text-sm font-medium text-gray-700">닉네임</label>
                        <input 
                            type="text" 
                            id="reg-nickname"
                            bind:value={regNickname}
                            placeholder="게임에서 사용할 이름"
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        />
                    </div>

                    <div class="space-y-1.5">
                        <label for="reg-password" class="block text-sm font-medium text-gray-700">비밀번호</label>
                        <input 
                            type="password" 
                            id="reg-password"
                            bind:value={regPassword}
                            placeholder="••••••••"
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        />
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