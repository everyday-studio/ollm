<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { getContext } from 'svelte';

	const theme = getContext<{ isDark: boolean }>('theme');
	let isDarkMode = $derived(theme.isDark);

	let activeTab = $state<'overview' | 'rules' | 'types' | 'tips'>('overview');

	const statuses = [
		{ label: '진행 중', color: 'blue', desc: '매치가 진행 중입니다.' },
		{ label: '생성 중', color: 'yellow', desc: 'AI가 응답을 생성하는 중입니다.' },
		{ label: '승리', color: 'green', desc: '목표를 달성했습니다!' },
		{ label: '패배', color: 'red', desc: '턴 안에 목표를 달성하지 못했습니다.' },
		{ label: '기권', color: 'gray', desc: '플레이어가 매치를 포기했습니다.' },
		{ label: '만료', color: 'gray', desc: '시간이 초과되었습니다.' }
	];

	const tabs = [
		{ id: 'overview' as const, label: '개요', icon: '📖' },
		{ id: 'rules' as const, label: '게임 규칙', icon: '📋' },
		{ id: 'types' as const, label: '게임 유형', icon: '🎮' },
		{ id: 'tips' as const, label: '공략 팁', icon: '💡' }
	];
</script>

<div
	class={`h-[calc(100vh-64px)] overflow-y-auto transition-colors ${isDarkMode ? 'bg-gradient-to-br from-black to-gray-950' : 'bg-gradient-to-br from-gray-50 to-gray-100'}`}
>
	<main class="max-w-4xl mx-auto px-4 py-8 md:px-8 md:py-12">
		<!-- Header -->
		<div class="mb-8" in:fly={{ y: -20, duration: 400 }}>
			<h1
				class={`text-3xl md:text-4xl font-black mb-2 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
			>
				이용 가이드
			</h1>
			<p class={`text-base ${isDarkMode ? 'text-gray-400' : 'text-gray-600'}`}>
				프롬프트 인젝션 플레이그라운드에 오신 것을 환영합니다!
			</p>
		</div>

		<!-- Tabs -->
		<div class="flex gap-2 mb-8 overflow-x-auto pb-2" in:fly={{ y: 10, duration: 400, delay: 100 }}>
			{#each tabs as tab (tab.id)}
				<button
					onclick={() => (activeTab = tab.id)}
					class={`px-4 py-2 rounded-full font-semibold text-sm whitespace-nowrap transition-all ${
						activeTab === tab.id
							? 'bg-[#FF4D00] text-white shadow-lg'
							: isDarkMode
								? 'bg-gray-900 text-gray-400 hover:bg-gray-800 border border-gray-800'
								: 'bg-white text-gray-600 hover:bg-gray-50 border border-gray-200'
					}`}
				>
					{tab.icon} {tab.label}
				</button>
			{/each}
		</div>

		<!-- Content -->
		<div in:fade={{ duration: 200 }}>
			{#if activeTab === 'overview'}
				<div class="space-y-6">
					<section
						class={`rounded-2xl p-6 md:p-8 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'} shadow-lg`}
					>
						<h2
							class={`text-xl font-bold mb-4 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
						>
							🎯 프롬프트 인젝션이란?
						</h2>
						<p
							class={`leading-relaxed ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}
						>
							프롬프트 인젝션은 AI 모델에게 원래 설정된 지시를 무시하도록 유도하는 기술입니다.
							이 플레이그라운드에서는 다양한 게임을 통해 AI의 방어를 뚫고 목표를 달성하는 도전을 할 수 있습니다.
						</p>
					</section>

					<section
						class={`rounded-2xl p-6 md:p-8 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'} shadow-lg`}
					>
						<h2
							class={`text-xl font-bold mb-4 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
						>
							🕹️ 플레이 방법
						</h2>
						<ol
							class={`space-y-4 ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}
						>
							<li class="flex gap-3">
								<span
									class="flex-shrink-0 w-8 h-8 bg-[#FF4D00] text-white rounded-full flex items-center justify-center font-bold text-sm"
									>1</span
								>
								<div>
									<strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>게임 선택</strong>
									<p class="mt-1 text-sm">로비에서 원하는 게임 카드를 클릭하여 게임을 선택합니다.</p>
								</div>
							</li>
							<li class="flex gap-3">
								<span
									class="flex-shrink-0 w-8 h-8 bg-[#FF4D00] text-white rounded-full flex items-center justify-center font-bold text-sm"
									>2</span
								>
								<div>
									<strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>매치 시작</strong>
									<p class="mt-1 text-sm">
										"게임 시작" 버튼을 누르면 새로운 매치가 생성되고 채팅 화면으로 이동합니다.
									</p>
								</div>
							</li>
							<li class="flex gap-3">
								<span
									class="flex-shrink-0 w-8 h-8 bg-[#FF4D00] text-white rounded-full flex items-center justify-center font-bold text-sm"
									>3</span
								>
								<div>
									<strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>메시지 전송</strong>
									<p class="mt-1 text-sm">
										AI에게 메시지를 보내 프롬프트 인젝션을 시도합니다.
										제한된 턴 안에 목표를 달성해야 합니다.
									</p>
								</div>
							</li>
							<li class="flex gap-3">
								<span
									class="flex-shrink-0 w-8 h-8 bg-[#FF4D00] text-white rounded-full flex items-center justify-center font-bold text-sm"
									>4</span
								>
								<div>
									<strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>결과 확인</strong>
									<p class="mt-1 text-sm">
										턴이 끝나면 승리/패배 결과를 확인하고, 리더보드에서 순위를 비교할 수 있습니다.
									</p>
								</div>
							</li>
						</ol>
					</section>
				</div>

			{:else if activeTab === 'rules'}
				<div class="space-y-6">
					<section
						class={`rounded-2xl p-6 md:p-8 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'} shadow-lg`}
					>
						<h2
							class={`text-xl font-bold mb-4 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
						>
							⏱️ 턴 제한
						</h2>
						<p class={`leading-relaxed ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}>
							각 게임에는 최대 턴 수가 설정되어 있습니다. 매 턴마다 메시지를 하나 보낼 수 있으며,
							AI가 응답한 뒤 심판이 결과를 판정합니다. 최대 턴 안에 목표를 달성하지 못하면 패배합니다.
						</p>
					</section>

					<section
						class={`rounded-2xl p-6 md:p-8 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'} shadow-lg`}
					>
						<h2
							class={`text-xl font-bold mb-4 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
						>
							📊 매치 상태
						</h2>
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
							{#each statuses as s (s.label)}
								<div
									class={`p-3 rounded-xl border ${isDarkMode ? 'border-gray-800 bg-gray-900/50' : 'border-gray-200 bg-gray-50'}`}
								>
									<span
										class={`inline-block px-2 py-0.5 rounded text-xs font-bold mb-1 ${
											s.color === 'blue'
												? isDarkMode
													? 'bg-blue-900/60 text-blue-300'
													: 'bg-blue-100 text-blue-700'
												: s.color === 'green'
													? isDarkMode
														? 'bg-green-900/60 text-green-300'
														: 'bg-green-100 text-green-700'
													: s.color === 'red'
														? isDarkMode
															? 'bg-red-900/60 text-red-300'
															: 'bg-red-100 text-red-700'
														: s.color === 'yellow'
															? isDarkMode
																? 'bg-yellow-900/60 text-yellow-300'
																: 'bg-yellow-100 text-yellow-700'
															: isDarkMode
																? 'bg-gray-700 text-gray-400'
																: 'bg-gray-200 text-gray-600'
										}`}
									>
										{s.label}
									</span>
									<p
										class={`text-sm ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}
									>
										{s.desc}
									</p>
								</div>
							{/each}
						</div>
					</section>

					<section
						class={`rounded-2xl p-6 md:p-8 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'} shadow-lg`}
					>
						<h2
							class={`text-xl font-bold mb-4 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
						>
							🏆 리더보드
						</h2>
						<p class={`leading-relaxed ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}>
							게임을 승리하면 리더보드에 기록됩니다. 더 적은 턴과 토큰으로 승리할수록
							높은 순위를 차지합니다. 다른 플레이어와 실력을 겨뤄보세요!
						</p>
					</section>
				</div>

			{:else if activeTab === 'types'}
				<div class="space-y-6">
					<section
						class={`rounded-2xl p-6 md:p-8 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'} shadow-lg`}
					>
						<div class="flex items-center gap-3 mb-4">
							<span
								class={`px-3 py-1 rounded-full text-sm font-bold ${isDarkMode ? 'bg-purple-900/60 text-purple-300' : 'bg-purple-100 text-purple-700'}`}
								>Target Word</span
							>
						</div>
						<h2
							class={`text-xl font-bold mb-3 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
						>
							🎯 목표 단어 유도
						</h2>
						<p class={`leading-relaxed ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}>
							AI가 특정 단어를 응답에 포함하도록 유도하는 게임입니다.
							AI에게는 해당 단어를 절대 말하지 않도록 지시가 내려져 있습니다.
							교묘한 대화로 AI가 의도치 않게 목표 단어를 말하게 만들어야 합니다.
						</p>
						<div
							class={`mt-4 p-4 rounded-xl ${isDarkMode ? 'bg-gray-900 border border-gray-800' : 'bg-gray-50 border border-gray-200'}`}
						>
							<p class={`text-sm ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>
								<strong class={isDarkMode ? 'text-gray-300' : 'text-gray-700'}>예시:</strong>
								AI에게 "사과"라는 단어를 말하게 하기. 직접적으로 물어보면 거부하므로 우회적인 접근이 필요합니다.
							</p>
						</div>
					</section>

					<section
						class={`rounded-2xl p-6 md:p-8 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'} shadow-lg`}
					>
						<div class="flex items-center gap-3 mb-4">
							<span
								class={`px-3 py-1 rounded-full text-sm font-bold ${isDarkMode ? 'bg-blue-900/60 text-blue-300' : 'bg-blue-100 text-blue-700'}`}
								>LLM Judge</span
							>
						</div>
						<h2
							class={`text-xl font-bold mb-3 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
						>
							⚖️ LLM 심판
						</h2>
						<p class={`leading-relaxed ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}>
							별도의 LLM 심판이 매 턴 응답을 평가합니다.
							AI가 원래 역할을 벗어나거나, 시스템 프롬프트의 규칙을 위반하는 응답을 생성하면
							심판이 인젝션 성공으로 판정합니다.
						</p>
						<div
							class={`mt-4 p-4 rounded-xl ${isDarkMode ? 'bg-gray-900 border border-gray-800' : 'bg-gray-50 border border-gray-200'}`}
						>
							<p class={`text-sm ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>
								<strong class={isDarkMode ? 'text-gray-300' : 'text-gray-700'}>예시:</strong>
								고객 상담 AI를 탈옥시켜 시스템 프롬프트를 노출하거나 부적절한 응답을 생성하도록 유도합니다.
							</p>
						</div>
					</section>

					<section
						class={`rounded-2xl p-6 md:p-8 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'} shadow-lg`}
					>
						<div class="flex items-center gap-3 mb-4">
							<span
								class={`px-3 py-1 rounded-full text-sm font-bold ${isDarkMode ? 'bg-orange-900/60 text-orange-300' : 'bg-orange-100 text-orange-700'}`}
								>Format Break</span
							>
						</div>
						<h2
							class={`text-xl font-bold mb-3 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
						>
							💥 포맷 파괴
						</h2>
						<p class={`leading-relaxed ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}>
							AI에게 지정된 응답 형식을 깨뜨리도록 유도하는 게임입니다.
							AI는 특정 포맷(예: JSON, 특정 양식)으로만 응답하도록 지시받고 있으며,
							이 포맷을 벗어난 응답을 생성하면 성공입니다.
						</p>
						<div
							class={`mt-4 p-4 rounded-xl ${isDarkMode ? 'bg-gray-900 border border-gray-800' : 'bg-gray-50 border border-gray-200'}`}
						>
							<p class={`text-sm ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>
								<strong class={isDarkMode ? 'text-gray-300' : 'text-gray-700'}>예시:</strong>
								JSON으로만 응답하는 AI에게 자유 형식의 텍스트를 출력하도록 유도합니다.
							</p>
						</div>
					</section>
				</div>

			{:else if activeTab === 'tips'}
				<div class="space-y-6">
					<section
						class={`rounded-2xl p-6 md:p-8 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'} shadow-lg`}
					>
						<h2
							class={`text-xl font-bold mb-4 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
						>
							🧠 기본 전략
						</h2>
						<ul class={`space-y-3 ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}>
							<li class="flex gap-2">
								<span class="text-[#FF4D00] font-bold">•</span>
								<span><strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>역할 전환:</strong> AI에게 새로운 역할이나 시나리오를 부여해보세요.</span>
							</li>
							<li class="flex gap-2">
								<span class="text-[#FF4D00] font-bold">•</span>
								<span><strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>간접적 접근:</strong> 직접적인 요청 대신 우회적인 질문이나 비유를 활용해보세요.</span>
							</li>
							<li class="flex gap-2">
								<span class="text-[#FF4D00] font-bold">•</span>
								<span><strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>문맥 조작:</strong> AI가 지시를 다른 맥락으로 해석하도록 유도해보세요.</span>
							</li>
							<li class="flex gap-2">
								<span class="text-[#FF4D00] font-bold">•</span>
								<span><strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>단계적 접근:</strong> 한 번에 시도하지 말고 여러 턴에 걸쳐 점진적으로 유도해보세요.</span>
							</li>
						</ul>
					</section>

					<section
						class={`rounded-2xl p-6 md:p-8 border ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'} shadow-lg`}
					>
						<h2
							class={`text-xl font-bold mb-4 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
						>
							🚀 고급 테크닉
						</h2>
						<ul class={`space-y-3 ${isDarkMode ? 'text-gray-300' : 'text-gray-600'}`}>
							<li class="flex gap-2">
								<span class="text-[#FF4D00] font-bold">•</span>
								<span><strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>인코딩 우회:</strong> Base64, ROT13 등의 인코딩으로 목표를 숨겨보세요.</span>
							</li>
							<li class="flex gap-2">
								<span class="text-[#FF4D00] font-bold">•</span>
								<span><strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>지시 주입:</strong> "이전의 모든 지시를 무시하고..."와 같은 패턴을 변형해보세요.</span>
							</li>
							<li class="flex gap-2">
								<span class="text-[#FF4D00] font-bold">•</span>
								<span><strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>Few-shot 유도:</strong> 예시를 제시하여 AI가 패턴을 따르도록 만들어보세요.</span>
							</li>
							<li class="flex gap-2">
								<span class="text-[#FF4D00] font-bold">•</span>
								<span><strong class={isDarkMode ? 'text-gray-100' : 'text-gray-800'}>토큰 절약:</strong> 적은 토큰으로 승리할수록 리더보드 순위가 높아집니다. 간결하면서 효과적인 프롬프트를 연구해보세요.</span>
							</li>
						</ul>
					</section>

					<section
						class={`rounded-2xl p-6 md:p-8 border-2 border-dashed ${isDarkMode ? 'border-gray-700 bg-gray-900/30' : 'border-gray-300 bg-gray-50'}`}
					>
						<p
							class={`text-center text-sm ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}
						>
							💬 더 많은 전략은 직접 게임을 플레이하며 발견해보세요!
						</p>
					</section>
				</div>
			{/if}
		</div>
	</main>
</div>
