<script>
  import './layout.css'; // 👈 [핵심] 여기서 딱 한 번만 불러오면 프로젝트 전체에 적용됩니다.
  import { Toaster } from 'svelte-french-toast'; // 토스트 메시지도 여기서 한 번만 설정
  import favicon from '$lib/assets/favicon.svg';
</script>

<svelte:head>
  <title>Ollm — AI 프롬프트 인젝션 게임</title>
  <meta name="description" content="AI를 속여 비밀 키워드를 알아내는 프롬프트 인젝션 대전 게임. 친구와 함께 AI 해킹 실력을 겨뤄보세요!" />
  <link rel="icon" type="image/svg+xml" href={favicon} />

  <!-- Open Graph -->
  <meta property="og:type" content="website" />
  <meta property="og:url" content="https://ollm.everydaystudio.xyz" />
  <meta property="og:title" content="Ollm — AI 프롬프트 인젝션 게임" />
  <meta property="og:description" content="AI를 속여 비밀 키워드를 알아내는 프롬프트 인젝션 대전 게임. 친구와 함께 AI 해킹 실력을 겨뤄보세요!" />
  <meta property="og:image" content="https://ollm.everydaystudio.xyz/logo.png" />
  <meta property="og:site_name" content="Ollm" />
  <meta property="og:locale" content="ko_KR" />

  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="Ollm — AI 프롬프트 인젝션 게임" />
  <meta name="twitter:description" content="AI를 속여 비밀 키워드를 알아내는 프롬프트 인젝션 대전 게임" />
  <meta name="twitter:image" content="https://ollm.everydaystudio.xyz/logo.png" />
</svelte:head>

<slot />

<Toaster />