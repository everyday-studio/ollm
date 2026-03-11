// src/lib/features/auth/api.ts
import client from '$lib/api/client';
import type { LoginRequest, SignupRequest, AuthResponse } from './types';

export const authApi = {
    login: (data: LoginRequest) => client.post<AuthResponse>('/api/auth/login', data),
    signup: (data: SignupRequest) => client.post<AuthResponse>('/api/auth/signup', data),
    logout: () => client.post('/api/auth/logout'),
    refresh: () => client.post<AuthResponse>('/api/auth/refresh')
};