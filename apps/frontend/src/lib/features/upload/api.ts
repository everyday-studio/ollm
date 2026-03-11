// src/lib/features/upload/api.ts
import client from '$lib/api/client';
import type { UploadType, UploadResponse } from './types';

export const uploadApi = {
    /**
     * Upload an image via multipart/form-data.
     * POST /upload/image
     */
    uploadImage: (type: UploadType, refId: string, file: File) => {
        const formData = new FormData();
        formData.append('type', type);
        formData.append('ref_id', refId);
        formData.append('file', file);

        return client.post<UploadResponse>('/api/upload/image', formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });
    }
};
