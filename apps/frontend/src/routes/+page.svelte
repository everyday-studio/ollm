<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  onMount(() => {
    // 접속하자마자 로그인 페이지로 이동
    goto('/login');
  });
</script>

<div class="flex items-center justify-center h-screen">
    <p>Loading...</p>
</div>