// place files you want to import through the `$lib` alias in this folder.
// src/lib/stores/index.ts
import { authStore } from './auth';

// 기존 user라는 이름으로 authStore를 내보내줌 (호환성 유지)
export const user = authStore;