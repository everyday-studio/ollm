import type { GameDTO } from './types';

export async function loadMockGames(): Promise<GameDTO[]> {
  const response = await fetch('/data/games.json');
  if (!response.ok) {
    throw new Error('Failed to load mock games');
  }

  const data = (await response.json()) as { games: GameDTO[] };
  return data.games;
}
