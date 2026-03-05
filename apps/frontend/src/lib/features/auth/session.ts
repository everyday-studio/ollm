// src/lib/features/auth/session.ts
// Deduplicated session restore — ensures auth/refresh is called at most once.

import { authApi } from './api';
import { authStore } from './model';

let refreshPromise: Promise<void> | null = null;

/**
 * Ensures the access token is available by calling POST /auth/refresh once.
 * Multiple callers (layout + page) will all await the same promise so only
 * one network request is made per app lifecycle.
 *
 * Safe to call from any component's onMount — first caller triggers the
 * actual refresh, subsequent callers piggyback on the same promise.
 */
export function ensureSession(): Promise<void> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    try {
      const res = await authApi.refresh();
      if (res?.data?.access_token) {
        authStore.updateToken(res.data.access_token);
      }
    } catch {
      // Ignore — the Axios interceptor will handle 401 on subsequent calls
    }
  })();

  return refreshPromise;
}

/** Reset state on logout so next login gets a fresh refresh. */
export function resetSession() {
  refreshPromise = null;
}
