<script>
  import { user } from '$lib/stores';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  
  onMount(() => {
    // 유저 정보가 없으면 로그인 화면으로 튕겨내기
    if (!$user) {
      alert('로그인이 필요합니다.');
      goto('/'); 
    }
  });
</script>

<div class="p-8">
  <h1 class="text-2xl font-bold">게임 로비</h1>
  <p>환영합니다, {$user?.nickname}님!</p>
  </div>