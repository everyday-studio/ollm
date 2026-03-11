// src/lib/features/auth/session.ts
// Deduplicated session restore — ensures auth/refresh is called at most once.

import { authApi } from './api';
import { authStore } from './model';
import { userApi } from '$lib/features/user/api';

let refreshPromise: Promise<boolean> | null = null;

/**
 * Ensures the access token is available by calling POST /auth/refresh once.
 * Multiple callers (layout + page) will all await the same promise so only
 * one network request is made per app lifecycle.
 *
 * Safe to call from any component's onMount — first caller triggers the
 * actual refresh, subsequent callers piggyback on the same promise.
 */
export function ensureSession(): Promise<boolean> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    try {
      const res = await authApi.refresh();
      if (res?.data?.access_token) {
        authStore.updateToken(res.data.access_token);
        // Populate user from refresh response (id, name, tag, email)
        const { id, name, tag, email } = res.data;
        authStore.updateUser({ id, name, tag, email, role: '', created_at: '', updated_at: '' });
      } else {
        return false;
      }
      // Fetch full user info to fill in role etc.
      try {
        const meRes = await userApi.getMe();
        authStore.updateUser(meRes.data);
      } catch (meErr) {
        console.warn('[ensureSession] getMe failed:', meErr);
      }
      return true;
    } catch {
      // Ignore — the Axios interceptor will handle 401 on subsequent calls
      // Return false indicating no valid session was established
      return false;
    }
  })();

  return refreshPromise;
}

/** Reset state on logout so next login gets a fresh refresh. */
export function resetSession() {
  refreshPromise = null;
}
