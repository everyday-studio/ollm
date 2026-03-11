// src/lib/features/user/api.ts
import client from '$lib/api/client';
import type { User } from '$lib/features/auth/types';
import type { UpdateNicknameRequest } from './types';

export const userApi = {
    // GET /users/me - Fetch current user profile
    getMe: () => client.get<User>('/api/users/me'),

    // PUT /users/me - Update current user nickname
    updateNickname: (data: UpdateNicknameRequest) => client.put<User>('/api/users/me', data)
};
