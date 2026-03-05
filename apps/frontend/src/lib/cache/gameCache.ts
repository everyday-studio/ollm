// src/lib/cache/gameCache.ts
// Convenience wrappers around cachedFetch for game-related data.

import { cachedFetch, invalidateCache } from './apiCache';
import { gameApi } from '$lib/features/game/api';
import { loadMockGames } from '$lib/features/game/mockData';
import type { GameDTO, MatchDTO } from '$lib/features/game/types';

const GAMES_TTL = 10 * 60 * 1000; // 10 min — games rarely change
const MATCHES_TTL = 2 * 60 * 1000; // 2 min — matches are more dynamic

/** Cached GET /games (falls back to mock on error). */
export function getCachedGames(): Promise<GameDTO[]> {
  return cachedFetch<GameDTO[]>(
    'games',
    async () => {
      try {
        const res = await gameApi.getGames();
        const games = res.data?.data;
        if (Array.isArray(games) && games.length > 0) return games;
        return await loadMockGames();
      } catch {
        return await loadMockGames();
      }
    },
    GAMES_TTL,
  );
}

/** Cached GET /matches/me. */
export function getCachedMyMatches(): Promise<MatchDTO[]> {
  return cachedFetch<MatchDTO[]>(
    'my-matches',
    async () => {
      try {
        const res = await gameApi.getMyMatches();
        return res.data ?? [];
      } catch {
        return [];
      }
    },
    MATCHES_TTL,
  );
}

/** Invalidate matches after creating/resigning — next fetch will hit network. */
export function invalidateMatchesCache() {
  invalidateCache('my-matches');
}

/** Invalidate all game-related caches. */
export function invalidateAllGameCaches() {
  invalidateCache('games');
  invalidateCache('my-matches');
}
