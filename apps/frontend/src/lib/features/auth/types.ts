// src/lib/features/auth/types.ts

// 사용자 정보
export interface User {
  id: string;
  email: string;
  username: string; // 백엔드 명세에 'name'이면 name으로 수정
  role: 'ADMIN' | 'USER';
  created_at: string;
}

// 로그인 응답 (백엔드 명세 기준)
export interface AuthResponse {
  access_token: string;
  user: User; // 만약 백엔드가 user 객체를 같이 준다면
  // 만약 토큰만 준다면 user는 빼야 합니다.
}

// API 요청 타입
export interface LoginRequest { email: string; password: string; }
export interface SignupRequest { email: string; password: string; }