// src/lib/features/auth/api.ts
import client from '$lib/api/client';

// 로그인/회원가입 요청 데이터 타입 정의 (필요하면 types.ts로 분리 가능)
import type { LoginRequest, SignupRequest, AuthResponse } from './types';

export const authApi = {
    // 로그인
    login: (data: LoginRequest) => client.post('/auth/login', data),
    
    // 회원가입
    signup: (data: SignupRequest) => client.post('/auth/signup', data),
    
    // 내 정보 조회
    getMe: () => client.get('/users/me'),
    
    // 로그아웃
    logout: () => client.post('/auth/logout')
};