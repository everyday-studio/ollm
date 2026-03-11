// src/lib/features/upload/types.ts

// Upload type determines the category of the uploaded image
export type UploadType = 'game' | 'user';

// Response from POST /upload/image
export interface UploadResponse {
    url: string;
    updated_at: string;
}
