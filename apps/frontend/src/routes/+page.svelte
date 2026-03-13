<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  onMount(() => {
    // 서버 로드가 실패한 경우의 클라이언트 폴백
    // 로비 레이아웃 서버 가드가 비로그인 유저를 /login으로 리다이렉트
    goto('/lobby');
  });
</script>

<div class="flex items-center justify-center h-screen">
    <p>Loading...</p>
</div>