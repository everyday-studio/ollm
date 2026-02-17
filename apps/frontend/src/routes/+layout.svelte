<script>
  import './layout.css'; // 👈 [핵심] 여기서 딱 한 번만 불러오면 프로젝트 전체에 적용됩니다.
  import { Toaster } from 'svelte-french-toast'; // 토스트 메시지도 여기서 한 번만 설정
</script>

<slot />

<Toaster />