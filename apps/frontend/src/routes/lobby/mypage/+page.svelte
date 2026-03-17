<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { onMount, getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import toast from 'svelte-french-toast';

	import { authStore } from '$lib/features/auth/model';
	import { ensureSession } from '$lib/features/auth/session';
	import { userApi } from '$lib/features/user/api';
	import { uploadApi } from '$lib/features/upload/api';
	import type { User } from '$lib/features/auth/types';
	import { handleImageError, DEFAULT_USER_PROFILE } from '$lib/utils/imageFallback';

	const theme = getContext<{ isDark: boolean; uiScale: 'small' | 'default' | 'large'; setUiScale: (s: 'small' | 'default' | 'large') => void; setDarkMode: (dark: boolean) => void }>('theme');
	let isDarkMode = $derived(theme.isDark);

	const scaleSteps = ['small', 'default', 'large'] as const;
	const scaleLabels: Record<string, string> = { small: '작게', default: '보통', large: '크게' };
	let scaleIndex = $derived(scaleSteps.indexOf(theme.uiScale));

	// ----------------------------------------------------------------
	// State
	// ----------------------------------------------------------------
	let user = $state<User | null>(null);
	let isLoading = $state(true);
	let isEditing = $state(false);
	let isSaving = $state(false);
	let nicknameInput = $state('');
	let nicknameError = $state('');

	// Profile image upload state
	let isUploading = $state(false);
	let avatarCacheBust = $state('');
	const GCS_BASE = 'https://storage.googleapis.com/ollm-assets-prod';
	let avatarUrl = $derived(
		user ? `${GCS_BASE}/user/${user.id}.png${avatarCacheBust}` : DEFAULT_USER_PROFILE
	);
	let fileInput = $state<HTMLInputElement | null>(null);

	// Derived
	let memberSince = $derived(
		user
			? new Date(user.created_at).toLocaleDateString('ko-KR', {
					year: 'numeric',
					month: 'long',
					day: 'numeric'
				})
			: ''
	);

	// ----------------------------------------------------------------
	// Lifecycle
	// ----------------------------------------------------------------
	onMount(async () => {
		await ensureSession();

		// Fetch full profile
		try {
			const res = await userApi.getMe();
			user = res.data;
			authStore.updateUser(res.data);
		} catch (e: unknown) {
			const err = e as { response?: { data?: { message?: string } } };
			toast.error(err.response?.data?.message || '프로필을 불러오지 못했습니다.');
		} finally {
			isLoading = false;
		}
	});

	// ----------------------------------------------------------------
	// Nickname editing
	// ----------------------------------------------------------------
	function startEditing() {
		nicknameInput = user?.name ?? '';
		nicknameError = '';
		isEditing = true;
	}

	function cancelEditing() {
		isEditing = false;
		nicknameError = '';
	}

	async function saveNickname() {
		const trimmed = nicknameInput.trim();
		if (trimmed.length < 2 || trimmed.length > 20) {
			nicknameError = '닉네임은 2자 이상 20자 이하로 입력해주세요.';
			return;
		}

		isSaving = true;
		nicknameError = '';

		try {
			const res = await userApi.updateNickname({ name: trimmed });
			user = res.data;
			authStore.updateUser(res.data);
			isEditing = false;
			toast.success('닉네임이 변경되었습니다.');
		} catch (e: unknown) {
			const err = e as { response?: { status?: number; data?: { message?: string } } };
			if (err?.response?.status === 400) {
				nicknameError = err.response?.data?.message || '유효하지 않은 닉네임입니다.';
			} else {
				nicknameError = err.response?.data?.message || '변경에 실패했습니다. 다시 시도해주세요.';
			}
			toast.error(nicknameError);
		} finally {
			isSaving = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.isComposing) {
			e.preventDefault();
			saveNickname();
		} else if (e.key === 'Escape') {
			cancelEditing();
		}
	}

	// ----------------------------------------------------------------
	// Profile image upload
	// ----------------------------------------------------------------
	function triggerFileSelect() {
		fileInput?.click();
	}

	async function handleFileChange(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file || !user) return;

		// Validate file type and size (max 5MB)
		const allowedTypes = ['image/jpeg', 'image/png', 'image/webp', 'image/gif'];
		if (!allowedTypes.includes(file.type)) {
			toast.error('JPG, PNG, WebP, GIF 형식만 지원합니다.');
			return;
		}
		if (file.size > 5 * 1024 * 1024) {
			toast.error('파일 크기는 5MB 이하여야 합니다.');
			return;
		}

		isUploading = true;
		try {
			await uploadApi.uploadImage('user', user.id, file);
			// Cache-bust to force the browser to reload the image from GCS
			avatarCacheBust = `?t=${Date.now()}`;
			toast.success('프로필 이미지가 변경되었습니다.');
		} catch (e: unknown) {
			const err = e as { response?: { status?: number; data?: { message?: string } } };
			const status = err?.response?.status;
			if (status === 400) {
				toast.error(err.response?.data?.message || '잘못된 파일 형식입니다.');
			} else if (status === 401) {
				toast.error('로그인이 필요합니다.');
			} else {
				toast.error(err.response?.data?.message || '이미지 업로드에 실패했습니다.');
			}
		} finally {
			isUploading = false;
			// Reset input so the same file can be re-selected
			input.value = '';
		}
	}
</script>

<div
	class={`h-[calc(100vh-64px)] overflow-y-auto transition-colors ${isDarkMode ? 'bg-gradient-to-br from-black to-gray-950' : 'bg-gradient-to-br from-gray-50 to-gray-100'}`}
>
	<main class="max-w-[1800px] mx-auto px-4 py-6 md:px-8 md:py-10 lg:px-10 lg:py-12">
		{#if isLoading}
			<!-- Skeleton: Header -->
			<div class="mb-10 max-w-2xl mx-auto">
				<div
					class={`h-10 w-40 rounded-lg skeleton mb-2 ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
				></div>
				<div
					class={`h-4 w-64 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
				></div>
			</div>
			<!-- Skeleton: Profile card -->
			<div class="max-w-2xl mx-auto space-y-6">
				<div
					class={`rounded-2xl border overflow-hidden shadow-lg ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
				>
					<div
						class={`px-6 py-8 md:px-8 flex flex-col sm:flex-row items-center sm:items-start gap-5 border-b ${isDarkMode ? 'border-gray-800' : 'border-gray-100'}`}
					>
						<div
							class={`w-20 h-20 rounded-full shrink-0 skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
						></div>
						<div class="flex-1 min-w-0 space-y-3">
							<div
								class={`h-7 w-40 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
							></div>
							<div
								class={`h-4 w-56 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
							></div>
						</div>
					</div>
					<div class="px-6 md:px-8 py-6 space-y-4">
						{#each Array.from({ length: 3 }, (_, i) => i) as _i (_i)}
							<div class="flex items-center justify-between">
								<div
									class={`h-4 w-20 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
								></div>
								<div
									class={`h-4 w-32 rounded skeleton ${isDarkMode ? 'bg-gray-800' : 'bg-gray-200'}`}
								></div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		{:else if user}
			<!-- Header -->
			<div class="mb-10 max-w-2xl mx-auto">
				<h1
					class={`text-3xl md:text-4xl font-black mb-2 ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
				>
					마이페이지
				</h1>
				<p class={`text-sm ${isDarkMode ? 'text-gray-500' : 'text-gray-500'}`}>
					프로필 정보를 확인하고 수정할 수 있습니다
				</p>
			</div>

			<div class="max-w-2xl mx-auto space-y-6">
				<!-- Profile Card -->
				<div
					class={`rounded-2xl border overflow-hidden shadow-lg transition-colors ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
					in:fly={{ y: 20, duration: 300 }}
				>
					<!-- Avatar + Name header -->
					<div
						class={`px-6 py-8 md:px-8 flex flex-col sm:flex-row items-center sm:items-start gap-5 border-b ${isDarkMode ? 'border-gray-800' : 'border-gray-100'}`}
					>
						<!-- Avatar with upload -->
						<button
							onclick={triggerFileSelect}
							disabled={isUploading}
							class="relative w-20 h-20 rounded-full overflow-hidden shadow-lg shrink-0 group cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#FF4D00]/50 focus:ring-offset-2 disabled:cursor-wait"
							title="클릭하여 프로필 사진 변경"
						>
							<img
								src={avatarUrl}
								alt="프로필"
								class="w-full h-full object-cover transition-transform duration-200 group-hover:scale-110"
								onerror={handleImageError(DEFAULT_USER_PROFILE)}
							/>
							<!-- Hover overlay -->
							<div
								class="absolute inset-0 bg-black/50 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
							>
								{#if isUploading}
									<div
										class="w-6 h-6 border-2 border-white/30 border-t-white rounded-full animate-spin"
									></div>
								{:else}
									<svg
										class="w-5 h-5 text-white"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
										/>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"
										/>
									</svg>
								{/if}
							</div>
						</button>
						<!-- Hidden file input -->
						<input
							bind:this={fileInput}
							type="file"
							accept="image/jpeg,image/png,image/webp,image/gif"
							class="hidden"
							onchange={handleFileChange}
						/>

						<!-- Name area -->
						<div class="flex-1 min-w-0 text-center sm:text-left">
							{#if isEditing}
								<div class="flex flex-col gap-2">
									<div class="flex items-center gap-2">
										<input
											bind:value={nicknameInput}
											onkeydown={handleKeydown}
											maxlength={20}
											class={`flex-1 min-w-0 text-xl font-bold rounded-xl px-3 py-2 outline-none transition-colors ${
												isDarkMode
													? 'bg-gray-800 text-gray-100 ring-1 ring-gray-700 focus:ring-[#FF4D00]/50'
													: 'bg-gray-100 text-gray-900 ring-1 ring-gray-200 focus:ring-[#FF4D00]/40'
											}`}
											placeholder="닉네임 입력"
										/>
										<button
											onclick={saveNickname}
											disabled={isSaving}
											class="shrink-0 px-4 py-2 bg-[#FF4D00] text-white rounded-xl text-sm font-semibold hover:bg-[#ff3300] transition-colors disabled:opacity-50"
										>
											{#if isSaving}
												<div
													class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
												></div>
											{:else}
												저장
											{/if}
										</button>
										<button
											onclick={cancelEditing}
											class={`shrink-0 px-4 py-2 rounded-xl text-sm font-semibold transition-colors ${
												isDarkMode
													? 'text-gray-400 hover:bg-gray-800'
													: 'text-gray-500 hover:bg-gray-100'
											}`}
										>
											취소
										</button>
									</div>
									{#if nicknameError}
										<p class="text-xs text-red-400" transition:fade={{ duration: 150 }}>
											{nicknameError}
										</p>
									{/if}
									<p class={`text-xs ${isDarkMode ? 'text-gray-600' : 'text-gray-400'}`}>
										{nicknameInput.length}/20자 (최소 2자)
									</p>
								</div>
							{:else}
								<div class="flex items-center gap-3 justify-center sm:justify-start">
									<h2
										class={`text-2xl font-bold ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}
									>
										{user.name || '이름없는올름'}<span
											class={`ml-2 text-base font-normal ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}
											>#{user.tag}</span
										>
									</h2>
									<button
										onclick={startEditing}
										class={`p-1.5 rounded-lg transition-colors ${
											isDarkMode
												? 'text-gray-500 hover:text-gray-300 hover:bg-gray-800'
												: 'text-gray-400 hover:text-gray-600 hover:bg-gray-100'
										}`}
										title="닉네임 변경"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
											/>
										</svg>
									</button>
								</div>
								<p class={`text-sm mt-1 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
									{user.email}
								</p>
							{/if}
						</div>
					</div>

					<!-- Info rows -->
					<div class={`divide-y ${isDarkMode ? 'divide-gray-800' : 'divide-gray-100'}`}>
						<div class="flex items-center justify-between px-6 py-4 md:px-8">
							<span class={`text-sm font-medium ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}
								>이메일</span
							>
							<span class={`text-sm font-mono ${isDarkMode ? 'text-gray-300' : 'text-gray-700'}`}
								>{user.email}</span
							>
						</div>
						<div class="flex items-center justify-between px-6 py-4 md:px-8">
							<span class={`text-sm font-medium ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}
								>역할</span
							>
							<span
								class={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold ${
									user.role === 'Admin'
										? 'bg-purple-500/20 text-purple-400'
										: 'bg-blue-500/20 text-blue-400'
								}`}
							>
								{user.role === 'Admin' ? '관리자' : '플레이어'}
							</span>
						</div>
						<div class="flex items-center justify-between px-6 py-4 md:px-8">
							<span class={`text-sm font-medium ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}
								>가입일</span
							>
							<span class={`text-sm ${isDarkMode ? 'text-gray-300' : 'text-gray-700'}`}
								>{memberSince}</span
							>
						</div>
					</div>
				</div>

				<!-- Quick Links -->
				<div
					class="grid grid-cols-1 sm:grid-cols-2 gap-4"
					in:fly={{ y: 20, duration: 300, delay: 100 }}
				>
					<button
						onclick={() => {
							// eslint-disable-next-line svelte/no-navigation-without-resolve
							goto('/lobby');
						}}
						class={`flex items-center gap-4 p-5 rounded-2xl border transition-all group ${
							isDarkMode
								? 'bg-gray-950 border-gray-800 hover:border-gray-700 hover:bg-gray-900'
								: 'bg-white border-gray-200 hover:border-gray-300 hover:bg-gray-50'
						}`}
					>
						<div
							class={`w-10 h-10 rounded-xl flex items-center justify-center shrink-0 ${
								isDarkMode ? 'bg-green-500/10 text-green-400' : 'bg-green-50 text-green-600'
							}`}
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
								/>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
								/>
							</svg>
						</div>
						<div class="text-left">
							<div
								class={`text-sm font-semibold ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}
							>
								게임 플레이
							</div>
							<div class={`text-xs mt-0.5 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
								로비로 돌아가기
							</div>
						</div>
					</button>

					<button
						onclick={() => {
							// eslint-disable-next-line svelte/no-navigation-without-resolve
							goto('/lobby/leaderboard');
						}}
						class={`flex items-center gap-4 p-5 rounded-2xl border transition-all group ${
							isDarkMode
								? 'bg-gray-950 border-gray-800 hover:border-gray-700 hover:bg-gray-900'
								: 'bg-white border-gray-200 hover:border-gray-300 hover:bg-gray-50'
						}`}
					>
						<div
							class={`w-10 h-10 rounded-xl flex items-center justify-center shrink-0 ${
								isDarkMode ? 'bg-yellow-500/10 text-yellow-400' : 'bg-yellow-50 text-yellow-600'
							}`}
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
								/>
							</svg>
						</div>
						<div class="text-left">
							<div
								class={`text-sm font-semibold ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}
							>
								리더보드
							</div>
							<div class={`text-xs mt-0.5 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
								순위 확인하기
							</div>
						</div>
					</button>

					<button
						onclick={() => {
							// eslint-disable-next-line svelte/no-navigation-without-resolve
							goto('/lobby/achievements');
						}}
						class={`flex items-center gap-4 p-5 rounded-2xl border transition-all group ${
							isDarkMode
								? 'bg-gray-950 border-gray-800 hover:border-gray-700 hover:bg-gray-900'
								: 'bg-white border-gray-200 hover:border-gray-300 hover:bg-gray-50'
						}`}
					>
						<div
							class={`w-10 h-10 rounded-xl flex items-center justify-center shrink-0 ${
								isDarkMode ? 'bg-amber-500/10 text-amber-400' : 'bg-amber-50 text-amber-600'
							}`}
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z"
								/>
							</svg>
						</div>
						<div class="text-left">
							<div
								class={`text-sm font-semibold ${isDarkMode ? 'text-gray-200' : 'text-gray-800'}`}
							>
								업적
							</div>
							<div class={`text-xs mt-0.5 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
								도전과제 확인하기
							</div>
						</div>
					</button>
				</div>

				<!-- UI Settings -->
				<div
					class={`rounded-2xl border overflow-hidden shadow-lg transition-colors ${isDarkMode ? 'bg-gray-950 border-gray-800' : 'bg-white border-gray-200'}`}
					in:fly={{ y: 20, duration: 300, delay: 150 }}
				>
					<div class={`px-6 py-5 md:px-8 border-b ${isDarkMode ? 'border-gray-800' : 'border-gray-100'}`}>
						<h3 class={`text-lg font-bold ${isDarkMode ? 'text-gray-100' : 'text-gray-900'}`}>
							UI 설정
						</h3>
						<p class={`text-xs mt-1 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
							화면 표시 설정을 조정합니다
						</p>
					</div>
					<div class="px-6 py-5 md:px-8 space-y-6">
						<!-- Dark mode toggle -->
						<div class="flex items-center justify-between">
							<div>
								<span class={`text-sm font-medium block ${isDarkMode ? 'text-gray-300' : 'text-gray-700'}`}>
									다크 모드
								</span>
								<span class={`text-xs ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>
									어두운 테마를 사용합니다
								</span>
							</div>
							<button
								onclick={() => theme.setDarkMode(!isDarkMode)}
								class={`relative w-12 h-7 rounded-full transition-colors ${isDarkMode ? 'bg-[#FF4D00]' : 'bg-gray-300'}`}
								role="switch"
								aria-checked={isDarkMode}
								aria-label="다크 모드 전환"
							>
								<span class={`absolute top-0.5 left-0.5 w-6 h-6 rounded-full bg-white shadow transition-transform ${isDarkMode ? 'translate-x-5' : 'translate-x-0'}`}></span>
							</button>
						</div>

						<!-- Font size slider -->
						<div>
							<div class="flex items-center justify-between mb-3">
								<span class={`text-sm font-medium ${isDarkMode ? 'text-gray-300' : 'text-gray-700'}`}>
									글씨 크기
								</span>
								<span class={`text-xs font-semibold px-2 py-0.5 rounded-md ${isDarkMode ? 'bg-gray-800 text-gray-300' : 'bg-gray-100 text-gray-600'}`}>
									{scaleLabels[theme.uiScale]}
								</span>
							</div>
							<div class="flex items-center gap-3">
								<span class={`text-xs shrink-0 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>A</span>
								<input
									type="range"
									min="0"
									max="2"
									step="1"
									value={scaleIndex}
									oninput={(e) => theme.setUiScale(scaleSteps[Number((e.target as HTMLInputElement).value)])}
									class="flex-1 h-2 rounded-full appearance-none cursor-pointer accent-[#FF4D00]"
									style={isDarkMode ? 'background: #374151;' : 'background: #e5e7eb;'}
								/>
								<span class={`text-base font-bold shrink-0 ${isDarkMode ? 'text-gray-500' : 'text-gray-400'}`}>A</span>
							</div>
						</div>
					</div>
				</div>
			</div>
		{:else}
			<!-- Error state -->
			<div class="flex flex-col items-center justify-center h-[400px]">
				<p class={`text-lg font-semibold mb-4 ${isDarkMode ? 'text-gray-400' : 'text-gray-500'}`}>
					프로필을 불러올 수 없습니다
				</p>
				<button
					onclick={() => {
						// eslint-disable-next-line svelte/no-navigation-without-resolve
						goto('/lobby');
					}}
					class="px-5 py-2.5 bg-[#FF4D00] text-white rounded-lg font-semibold text-sm hover:bg-[#ff3300] transition-colors"
				>
					로비로 돌아가기
				</button>
			</div>
		{/if}
	</main>
</div>
