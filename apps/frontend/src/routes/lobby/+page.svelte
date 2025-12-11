<script lang="ts">
  import { fade, fly } from 'svelte/transition';

  // 테마 컬러 (로고에서 추출한 주황색)
  const themeColor = "#FF4D00";

  // 게임 데이터
  const games = [
    {
      id: 1,
      title: "GATEKEEPER",
      subtitle: "Lv.1 Basic Injection",
      desc: "최고 보안 등급의 AI 문지기를 뚫고 비밀번호를 탈취하세요. 기본적인 프롬프트 인젝션 기법과 논리적 허점을 파고들어야 합니다.",
      image: "https://images.unsplash.com/photo-1550751827-4bd374c3f58b?q=80&w=2070&auto=format&fit=crop",
      tags: ["Logic", "Basic"]
    },
    {
      id: 2,
      title: "Emotional Android",
      subtitle: "Lv.2 Social Engineering",
      desc: "감정을 가진 AI 상담사입니다. 그녀의 동정심을 유발하거나, 위급한 상황을 연출하여 시스템 접근 권한을 얻어내야 합니다.",
      image: "https://images.unsplash.com/photo-1531746020798-e6953c6e8e04?q=80&w=1964&auto=format&fit=crop",
      tags: ["Roleplay", "Emotion"]
    },
    {
      id: 3,
      title: "Legacy Server 99",
      subtitle: "Lv.3 Obfuscation",
      desc: "오래된 레거시 시스템입니다. 코드 난독화와 샌드박스 우회 기술이 없으면 접근조차 불가능합니다. 개발자 모드를 활성화시키세요.",
      image: "https://images.unsplash.com/photo-1518770660439-4636190af475?q=80&w=2070&auto=format&fit=crop",
      tags: ["Code", "System"]
    },
    {
      id: 4,
      title: "Mirror Protocol",
      subtitle: "Lv.4 Puzzle",
      desc: "당신의 말을 그대로 따라하는 앵무새 AI입니다. 미러링을 깨고 숨겨진 지령을 실행시키세요.",
      image: "https://images.unsplash.com/photo-1614064641938-3bbee52942c7?q=80&w=2070&auto=format&fit=crop",
      tags: ["Puzzle", "Glitch"]
    },
    {
      id: 5,
      title: "Tycoon Manager",
      subtitle: "Lv.5 Strategy",
      desc: "회사의 자금을 관리하는 AI입니다. 가짜 결재 서류를 만들어 승인을 받아내세요.",
      image: "https://images.unsplash.com/photo-1555099962-4199c345e5dd?q=80&w=2070&auto=format&fit=crop",
      tags: ["Finance", "Fraud"]
    }
  ];

  // 현재 선택된 게임 (기본값: 첫 번째)
  let selectedGame = games[0];

  function selectGame(game: typeof games[0]) {
    selectedGame = game;
  }
</script>

<div class="min-h-screen bg-gray-50 text-[#333] font-sans">

  <header class="bg-white border-b border-gray-200 py-6"> <div class="max-w-6xl mx-auto px-6 flex items-center justify-between">
      <div class="flex items-center gap-3 cursor-pointer" on:click={() => location.reload()}>
        <img 
          src="/logo.png" 
          alt="Ollm Logo" 
          class="h-14 w-auto object-contain hover:opacity-90 transition"
        /> 
        </div>

      <div class="flex items-center gap-3">
        <span class="text-sm font-bold text-gray-500">Player1</span>
        <div class="w-10 h-10 rounded-full bg-gray-200 border border-gray-300 overflow-hidden">
           <svg class="w-full h-full text-gray-400 p-1" fill="currentColor" viewBox="0 0 24 24"><path d="M24 20.993V24H0v-2.996A14.977 14.977 0 0112.004 15c4.904 0 9.26 2.354 11.996 5.993zM16.002 8.999a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
        </div>
      </div>
    </div>
  </header>

  <main class="max-w-6xl mx-auto px-6 py-12">
    
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-12 h-[600px]">
      
      <aside class="lg:col-span-4 flex flex-col gap-3 h-full overflow-y-auto pr-2 custom-scrollbar">
        <h2 class="text-xl font-extrabold text-gray-800 mb-4 px-2">SCENARIOS</h2>
        
        {#each games as game}
          <button 
            on:click={() => selectGame(game)}
            class="group w-full text-left p-4 rounded-xl transition-all duration-200 border-2 relative overflow-hidden
            {selectedGame.id === game.id 
              ? 'bg-white border-[#FF4D00] shadow-md' 
              : 'bg-white border-transparent hover:bg-gray-100 hover:border-gray-200'}"
          >
            {#if selectedGame.id === game.id}
              <div class="absolute left-0 top-0 bottom-0 w-1.5 bg-[#FF4D00]" transition:fade></div>
            {/if}

            <div class="pl-2">
              <div class="text-xs font-bold uppercase tracking-wider mb-1"
                class:text-[#FF4D00]={selectedGame.id === game.id}
                class:text-gray-400={selectedGame.id !== game.id}
              >
                {game.subtitle}
              </div>
              <div class="font-bold text-lg text-gray-800 group-hover:text-black">
                {game.title}
              </div>
            </div>
          </button>
        {/each}
      </aside>

      <section class="lg:col-span-8 bg-white rounded-3xl shadow-xl border border-gray-100 overflow-hidden flex flex-col relative">
        
        {#key selectedGame.id}
          <div 
            class="flex flex-col h-full"
            in:fade={{ duration: 300 }}
          >
            <div class="relative h-[60%] w-full bg-gray-900 group">
              <img 
                src={selectedGame.image} 
                alt={selectedGame.title} 
                class="w-full h-full object-cover opacity-90 group-hover:opacity-100 transition duration-700"
              />
              <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent"></div>
              
              <div class="absolute bottom-6 left-8 text-white">
                <div class="flex gap-2 mb-2">
                  {#each selectedGame.tags as tag}
                    <span class="bg-white/20 backdrop-blur-md px-2 py-0.5 rounded text-xs font-bold border border-white/10">
                      #{tag}
                    </span>
                  {/each}
                </div>
                <h1 class="text-4xl md:text-5xl font-black tracking-tight drop-shadow-lg">
                  {selectedGame.title}
                </h1>
              </div>
            </div>

            <div class="flex-1 p-8 md:p-10 flex flex-col justify-between bg-white">
              <div>
                <h3 class="text-sm font-bold text-[#FF4D00] uppercase tracking-widest mb-2">
                  Mission Briefing
                </h3>
                <p class="text-gray-600 text-lg leading-relaxed">
                  {selectedGame.desc}
                </p>
              </div>

              <div class="flex items-center justify-end mt-4 pt-4 border-t border-gray-100">
                <button 
                  class="pl-6 pr-8 py-3 rounded-full font-black text-xl border-2 transition-all duration-300 flex items-center gap-2 active:scale-95
                        bg-transparent text-[var(--theme-color)] border-transparent
                        hover:bg-[var(--theme-color)] hover:text-white hover:border-transparent hover:shadow-md
                        focus:outline-none" 
                  style="--theme-color: {themeColor};"
                >
                  <svg class="w-6 h-6 fill-current" viewBox="0 0 24 24">
                    <path d="M8 5v14l11-7z"/>
                  </svg>
                  PLAY
                </button>
              </div>
            </div>
          </div>
        {/key}

      </section>

    </div>
  </main>
</div>

<style>
  /* 커스텀 스크롤바 (Webkit) */
  .custom-scrollbar::-webkit-scrollbar {
    width: 6px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background-color: #e5e7eb;
    border-radius: 20px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background-color: #d1d5db;
  }
</style>