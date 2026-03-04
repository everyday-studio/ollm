// src/lib/features/auth/api.ts
import client from '$lib/api/client';
import type { LoginRequest, SignupRequest, AuthResponse, User, UpdateNicknameRequest } from './types';

export const authApi = {
    login: (data: LoginRequest) => client.post<AuthResponse>('/auth/login', data),
    signup: (data: SignupRequest) => client.post<AuthResponse>('/auth/signup', data),
    getMe: () => client.get<User>('/users/me'),
    updateNickname: (data: UpdateNicknameRequest) => client.put<User>('/users/me', data),
    logout: () => client.post('/auth/logout'),
    refresh: () => client.post<AuthResponse>('/auth/refresh')
};