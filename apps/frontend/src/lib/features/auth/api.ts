// src/lib/features/auth/api.ts
import client from '$lib/api/client';
import type { LoginRequest, SignupRequest, AuthResponse } from './types';

export const authApi = {
    login: (data: LoginRequest) => client.post<AuthResponse>('/auth/login', data),
    signup: (data: SignupRequest) => client.post<AuthResponse>('/auth/signup', data),
    getMe: () => client.get<AuthResponse>('/users/me'),
    logout: () => client.post('/auth/logout'),
    // Attempt to refresh access token using HTTP-only refresh cookie
    refresh: () => client.post<AuthResponse>('/auth/refresh')
};