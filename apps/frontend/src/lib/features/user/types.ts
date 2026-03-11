// src/lib/features/user/types.ts

// Re-export User type from auth for convenience
export type { User } from '$lib/features/auth/types';

// Nickname update request
export interface UpdateNicknameRequest {
    name: string;
}
