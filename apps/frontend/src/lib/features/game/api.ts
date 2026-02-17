// src/lib/features/game/api.ts
import client from '$lib/api/client';
import type { GameDTO, MatchDTO } from './types';

export const gameApi = {
  // GET /games - Fetch all public games
  getGames: () => client.get<GameDTO[]>('/games'),

  // GET /matches/me - Fetch my matches
  getMyMatches: () => client.get<MatchDTO[]>('/matches/me'),

  // POST /matches - Create a new match
  createMatch: (gameId: string) => client.post<MatchDTO>('/matches', { game_id: gameId })
};