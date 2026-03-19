// src/lib/features/auth/types.ts

// 사용자 정보
export interface User {
  id: string;
  email: string;
  name: string;
  tag: string;
  role: string;
  created_at: string;
  updated_at: string;
}

// 로그인/리프레시 응답 (백엔드는 플랫 구조로 내려줌)
export interface AuthResponse {
  id: string;
  name: string;
  tag: string;
  email: string;
  access_token: string;
  refresh_token?: string; // 게스트 로그인 응답에만 포함됨 (localStorage 저장용)
}

// API 요청 타입
export interface LoginRequest { email: string; password: string; }
export interface SignupRequest { email: string; password: string; }
export interface GoogleLoginRequest { id_token: string; }