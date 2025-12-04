import { writable } from 'svelte/store';

// 유저 정보를 담는 저장소 (초기값: null = 비로그인)
export const user = writable<{ email: string; nickname: string } | null>(null);

// 로그인 상태인지 확인하는 파생 스토어
// export const isLoggedIn = derived(user, $user => !!$user);