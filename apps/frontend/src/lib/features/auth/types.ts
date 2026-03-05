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
}

// 닉네임 변경 요청
export interface UpdateNicknameRequest {
  name: string;
}

// API 요청 타입
export interface LoginRequest { email: string; password: string; }
export interface SignupRequest { email: string; password: string; }