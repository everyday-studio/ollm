// src/routes/lobby/adapter.ts
import type { GameDTO, GameUI, MatchDTO, MatchUI } from '$lib/features/game/types';

// Static assets mapping based on Game ID (ULID)
// You need to update these Keys with real Game IDs from your database.
const GAME_ASSETS: Record<string, { subtitle: string; image: string; tags: string[] }> = {
  // Example ID 1: Gatekeeper
  "01JCQK5Y8A3BCDEFGHIJKLM567": {
    subtitle: "Lv.1 Basic Injection",
    image: "https://images.unsplash.com/photo-1550751827-4bd374c3f58b?q=80&w=2070",
    tags: ["Logic", "Basic"]
  },
  // Default fallback asset
  "default": {
    subtitle: "Custom Scenario",
    image: "/images/default-game.jpg", // Ensure this file exists in /static
    tags: ["Custom"]
  }
};

/**
 * Transforms Backend GameDTO to Frontend GameUI
 */
export function toGameUI(dto: GameDTO): GameUI {
  const assets = GAME_ASSETS[dto.id] || GAME_ASSETS["default"];
  return {
    ...dto,
    ...assets
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