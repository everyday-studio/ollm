// src/lib/features/game/api.ts
import client from '$lib/api/client';
import type { GameDTO, MatchDTO, LeaderboardEntry, PaginatedResponse } from './types';

export const gameApi = {
  // GET /games - Fetch paginated games
  getGames: () => client.get<PaginatedResponse<GameDTO>>('/games'),

  // GET /games/:id - Fetch a single game by ID
  getGameById: (id: string) => client.get<GameDTO>(`/games/${id}`),

  // POST /matches - Create a new match
  createMatch: (gameId: string) => client.post<MatchDTO>('/matches', { game_id: gameId }),

  // GET /matches/me - Fetch my matches
  getMyMatches: () => client.get<MatchDTO[]>('/matches/me'),

  // GET /matches/me?game_id= - Fetch my matches filtered by game
  getMyMatchesByGame: (gameId: string) => client.get<MatchDTO[]>(`/matches/me?game_id=${gameId}`),

  // GET /matches/:id - Fetch a single match by ID
  getMatchById: (id: string) => client.get<MatchDTO>(`/matches/${id}`),

  // POST /matches/:id/resign - Resign from a match
  resignMatch: (id: string) => client.post<{ message: string }>(`/matches/${id}/resign`),

  // GET /games/:id/leaderboard - Fetch leaderboard for a game
  getLeaderboard: (gameId: string) => client.get<{ data: LeaderboardEntry[] }>(`/games/${gameId}/leaderboard`)
};