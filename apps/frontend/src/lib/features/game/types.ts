// src/lib/features/game/types.ts

// [Backend DTO] Data structure exactly as defined in API_SPECIFICATION.md
export interface GameDTO {
  id: string; // ULID
  title: string;
  description: string;
  author_id: string;
  status: 'active' | 'inactive';
  is_public: boolean;
  created_at: string;
  updated_at: string;
}

export interface MatchDTO {
  id: string; // ULID
  user_id: string;
  game_id: string;
  status: 'active' | 'won' | 'lost' | 'resigned' | 'expired' | 'error';
  total_tokens: number;
  turn_count: number;
  created_at: string;
  updated_at: string;
}

// [Frontend UI Model] Extended structure for UI rendering
export interface GameUI extends GameDTO {
  subtitle: string; // e.g., "Lv.1 Basic Injection"
  image: string;    // Local or remote image URL
  tags: string[];   // e.g., ["Logic", "Basic"]
}

export interface MatchUI extends MatchDTO {
  gameTitle: string;    // Mapped from GameDTO
  displayTime: string;  // Formatted time (e.g., "2 hours ago")
  lastMessage: string;  // Placeholder or fetched separately
}