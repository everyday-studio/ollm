// src/lib/stores/auth.ts
import { writable } from 'svelte/store';
import type { User } from '../types';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

const initialState: AuthState = {
  user: null,
  accessToken: null, // 액세스 토큰은 메모리에만 저장 (보안)
  isAuthenticated: false,
  isLoading: true,
};

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);

  return {
    subscribe,
    // 로그인 성공 시 호출
    loginSuccess: (token: string, user: User) => {
      update(state => ({
        ...state,
        user,
        accessToken: token,
        isAuthenticated: true,
        isLoading: false
      }));
    },
    // 로그아웃 시 호출
    logout: () => {
      set({ ...initialState, isLoading: false });
    },
    // 토큰 갱신 시 호출
    updateToken: (token: string) => {
      update(state => ({ ...state, accessToken: token }));
    },
    // 로딩 종료
    finishLoading: () => {
        update(state => ({ ...state, isLoading: false }));
    }
  };
}

export const authStore = createAuthStore();