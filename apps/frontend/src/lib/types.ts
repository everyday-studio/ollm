// src/lib/types.ts

// 1. 공통 응답 포맷 (만약 백엔드가 공통 래퍼를 쓴다면)
export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}

// 2. 사용자 (User) - ULID 사용
export interface User {
  id: string; // ULID
  email: string;
  username: string;
  role: 'ADMIN' | 'USER';
  created_at: string; // ISO 8601
}

// 3. 게임 (Game)
export interface Game {
  id: string; // ULID
  title: string;
  description: string;
  genre: string;
  image_url?: string;
  created_at: string;
}

// 4. 인증 응답 (로그인 시 받는 데이터)
export interface AuthResponse {
  access_token: string;
  user: User;
}

// 5. 에러 응답
export interface ApiError {
  message: string;
  code?: string;
}