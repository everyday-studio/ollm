<script lang="ts">
  import { user } from '$lib/stores';

  // 상태 변수들
  let isRegisterMode = false; // false: 로그인, true: 회원가입
  let isLoading = false;
  let email = '';
  let password = '';
  let nickname = ''; 
  let errorMessage = '';

  // 탭 전환
  const toggleMode = (mode: boolean) => {
    isRegisterMode = mode;
    errorMessage = ''; 
  };

  // [MOCK] 자체 로그인/회원가입 처리
  const handleAuth = async () => {
    if (!email || !password) {
      errorMessage = "이메일과 비밀번호를 입력해주세요.";
      return;
    }

    if (isRegisterMode && !nickname) {
        errorMessage = "닉네임을 입력해주세요.";
        return;
    }

    isLoading = true;
    errorMessage = '';

    // 백엔드 통신 시뮬레이션 (1초)
    setTimeout(() => {
      isLoading = false;
      
      console.log(`[Simulation] ${isRegisterMode ? '회원가입' : '로그인'} 성공:`, email);
      
      // 스토어 업데이트
      user.set({ 
        email: email, 
        nickname: isRegisterMode ? nickname : 'Player1' 
      });

      alert(`환영합니다! ${isRegisterMode ? '회원가입' : '로그인'}이 완료되었습니다.`);
    }, 1000);
  };

  // [MOCK] 구글 로그인 처리
  const handleGoogleLogin = () => {
    isLoading = true;
    setTimeout(() => {
        isLoading = false;
        user.set({ email: "google@example.com", nickname: "GoogleUser" });
        alert("Google 계정으로 로그인되었습니다.");
    }, 1000);
  };
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50 text-gray-900 font-sans p-4">
  
  <div class="w-full max-w-md bg-white rounded-2xl shadow-xl overflow-hidden border border-gray-100">
    
    <div class="p-8 md:p-10">
      <div class="text-center mb-10">
        <h1 class="text-3xl font-bold text-gray-900 tracking-tight">LLM GAMES</h1>
        <p class="text-gray-500 mt-2 text-sm">프롬프트 인젝션 플레이그라운드</p>
      </div>

      <div class="flex mb-8 bg-gray-100 rounded-lg p-1">
        <button 
          class="flex-1 py-2.5 text-sm font-semibold rounded-md transition-all duration-200 cursor-pointer { !isRegisterMode ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700' }"
          on:click={() => toggleMode(false)}
        >
          로그인
        </button>
        <button 
          class="flex-1 py-2.5 text-sm font-semibold rounded-md transition-all duration-200 cursor-pointer { isRegisterMode ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700' }"
          on:click={() => toggleMode(true)}
        >
          회원가입
        </button>
      </div>

      <form on:submit|preventDefault={handleAuth} class="space-y-5">
        
        <div class="space-y-1.5">
          <label for="email" class="block text-sm font-medium text-gray-700">이메일</label>
          <input 
            type="email" 
            id="email"
            bind:value={email}
            placeholder="name@example.com"
            class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-gray-400"
          />
        </div>

        {#if isRegisterMode}
          <div class="space-y-1.5">
            <label for="nickname" class="block text-sm font-medium text-gray-700">닉네임</label>
            <input 
              type="text" 
              id="nickname"
              bind:value={nickname}
              placeholder="게임에서 사용할 이름"
              class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-gray-400"
            />
          </div>
        {/if}

        <div class="space-y-1.5">
          <label for="password" class="block text-sm font-medium text-gray-700">비밀번호</label>
          <input 
            type="password" 
            id="password"
            bind:value={password}
            placeholder="••••••••"
            class="w-full px-4 py-3 rounded-lg border border-gray-300 bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-gray-400"
          />
        </div>

        {#if errorMessage}
          <div class="text-red-500 text-sm font-medium text-center bg-red-50 py-2 rounded">
            {errorMessage}
          </div>
        {/if}

        <button 
          type="submit" 
          disabled={isLoading}
          class="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-3.5 rounded-lg shadow-md hover:shadow-lg transition-all transform active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer flex items-center justify-center"
        >
          {#if isLoading}
            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            처리 중...
          {:else}
            {isRegisterMode ? '회원가입 완료' : '로그인'}
          {/if}
        </button>
      </form>

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
  </div>
  
  <div class="absolute bottom-4 text-center text-xs text-gray-400">
    &copy; 2025 LLM GAMES. All rights reserved.
  </div>
</div>