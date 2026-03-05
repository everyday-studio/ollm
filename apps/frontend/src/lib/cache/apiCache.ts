// src/lib/cache/apiCache.ts
// Generic in-memory TTL cache with request deduplication

interface CacheEntry<T> {
  data: T;
  expiresAt: number;
}

const cache = new Map<string, CacheEntry<unknown>>();
const pendingRequests = new Map<string, Promise<unknown>>();

const DEFAULT_TTL = 5 * 60 * 1000; // 5 minutes

/**
 * Fetch data with in-memory TTL caching and request deduplication.
 *
 * - If cached and not expired → returns cached data instantly (no network).
 * - If multiple callers request the same key concurrently → only one network
 *   request is made; all callers await the same promise.
 * - On failure the pending entry is cleared so the next call retries.
 */
export async function cachedFetch<T>(
  key: string,
  fetcher: () => Promise<T>,
  ttl = DEFAULT_TTL,
): Promise<T> {
  // 1. Cache hit
  const cached = cache.get(key);
  if (cached && Date.now() < cached.expiresAt) {
    return cached.data as T;
  }

  // 2. Deduplicate concurrent in-flight requests
  if (pendingRequests.has(key)) {
    return pendingRequests.get(key)! as Promise<T>;
  }

  // 3. Fetch, cache, return
  const promise = fetcher()
    .then((data) => {
      cache.set(key, { data, expiresAt: Date.now() + ttl });
      pendingRequests.delete(key);
      return data;
    })
    .catch((err) => {
      pendingRequests.delete(key);
      throw err;
    });

  pendingRequests.set(key, promise);
  return promise;
}

/** Remove one or all cache entries. */
export function invalidateCache(key?: string) {
  if (key) {
    cache.delete(key);
  } else {
    cache.clear();
  }
}
