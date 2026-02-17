<script>
  let { children } = $props();
</script>

<svelte:head>
  <link rel="preload" as="image" href="/logo.png" />
</svelte:head>

<header class="fixed top-0 left-0 right-0 bg-white border-b border-gray-200 z-50 h-24">
  <div class="max-w-6xl mx-auto px-6 h-full flex items-center justify-between">
    
    <button type="button" class="flex items-center gap-3 hover:opacity-90 transition" onclick={() => location.reload()}>
      <img 
        src="/logo.png" 
        alt="Ollm Logo"
        width="140" 
        height="56"
        class="h-14 w-auto object-contain"
      /> 
    </button>

  </div>
</header>

<main class="pt-24 pb-12">
  {@render children()}
</main>