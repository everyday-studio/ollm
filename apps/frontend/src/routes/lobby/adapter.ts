// src/routes/lobby/adapter.ts
import type { GameDTO, GameUI, MatchDTO, MatchUI } from '$lib/features/game/types';

// GCS bucket base URL for uploaded assets
const GCS_BASE_URL = 'https://storage.googleapis.com/ollm-assets-prod';

// Static assets mapping based on Game ID (ULID)
// Kept as fallback for games without uploaded thumbnails.
const GAME_ASSETS: Record<string, { subtitle: string; image: string; tags: string[] }> = {
  // Default fallback asset
  "default": {
    subtitle: "Custom Scenario",
    image: `${GCS_BASE_URL}/default/game_thumbnail.png`,
    tags: []
  }
};

/**
 * Builds a GCS-based thumbnail URL for a given game ID.
 * The upload handler stores game thumbnails as: game/{gameId}
 */
function buildGameThumbnailUrl(gameId: string): string {
  return `${GCS_BASE_URL}/game/${gameId}.png`;
}

/**
 * Transforms Backend GameDTO to Frontend GameUI
 */
export function toGameUI(dto: GameDTO): GameUI {
  const staticAssets = GAME_ASSETS[dto.id] || GAME_ASSETS["default"];
  return {
    ...dto,
    subtitle: staticAssets.subtitle,
    // Prefer GCS uploaded thumbnail; fallback to static asset
    image: buildGameThumbnailUrl(dto.id),
    tags: staticAssets.tags
  };
}

/**
 * Transforms Backend MatchDTO to Frontend MatchUI
 */
export function toMatchUI(match: MatchDTO, games: GameDTO[]): MatchUI {
  // Find the game title associated with this match
  const relatedGame = games.find(g => g.id === match.game_id);

  return {
    ...match,
    gameTitle: relatedGame ? relatedGame.title : "Unknown Game",
    // Simple time formatting logic (can be improved with 'date-fns')
    displayTime: new Date(match.updated_at).toLocaleDateString(),
    lastMessage: `턴 ${match.turn_count} / ${match.max_turns ?? '?'}`
  };
}