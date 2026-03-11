// src/lib/features/game/api.ts
import client from '$lib/api/client';
import type { GameDTO, MatchDTO, LeaderboardEntry, PaginatedResponse } from './types';

export const gameApi = {
  // GET /games - Fetch paginated games
  getGames: () => client.get<PaginatedResponse<GameDTO>>('/api/games'),

  // GET /games/:id - Fetch a single game by ID
  getGameById: (id: string) => client.get<GameDTO>(`/api/games/${id}`),

  // POST /matches - Create a new match
  createMatch: (gameId: string) => client.post<MatchDTO>('/api/matches', { game_id: gameId }),

  // GET /matches/me - Fetch my matches
  getMyMatches: () => client.get<MatchDTO[]>('/api/matches/me'),

  // GET /matches/me?game_id= - Fetch my matches filtered by game
  getMyMatchesByGame: (gameId: string) => client.get<MatchDTO[]>(`/api/matches/me?game_id=${gameId}`),

  // GET /matches/:id - Fetch a single match by ID
  getMatchById: (id: string) => client.get<MatchDTO>(`/api/matches/${id}`),

  // POST /matches/:id/resign - Resign from a match
  resignMatch: (id: string) => client.post<{ message: string }>(`/api/matches/${id}/resign`),

  // GET /games/:id/leaderboard - Fetch leaderboard for a game
  getLeaderboard: (gameId: string) => client.get<{ data: LeaderboardEntry[] }>(`/api/games/${gameId}/leaderboard`)
};