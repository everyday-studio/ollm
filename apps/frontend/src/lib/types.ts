// src/lib/types.ts

// 1. 공통 응답 포맷 (만약 백엔드가 공통 래퍼를 쓴다면)
export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
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

// 5. 에러 응답
export interface ApiError {
  message: string;
  code?: string;
}