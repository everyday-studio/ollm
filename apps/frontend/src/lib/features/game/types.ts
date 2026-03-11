// src/lib/features/game/types.ts

// [Backend DTO] Data structure exactly as returned from the backend API
export interface GameDTO {
  id: string; // ULID
  title: string;
  description: string;
  author_id: string;
  status: 'active' | 'inactive';
  is_public: boolean;
  system_prompt: string;
  judge_type: 'target_word' | 'llm_judge' | 'format_break';
  judge_condition: string;
  max_turns: number;
  created_at: string;
  updated_at: string;
}

export type MatchStatus = 'active' | 'generating' | 'won' | 'lost' | 'resigned' | 'expired' | 'error';

export interface MatchDTO {
  id: string; // ULID
  user_id: string;
  game_id: string;
  status: MatchStatus;
  max_turns: number;
  total_tokens: number;
  turn_count: number;
  created_at: string;
  updated_at: string;
}

export type MessageRole = 'system' | 'user' | 'assistant';

export interface MessageDTO {
  id: string; // ULID
  match_id: string;
  role: MessageRole;
  content: string;
  is_visible: boolean;
  turn_count: number;
  token_count: number;
  created_at: string;
}

export interface CreateMessageRequest {
  content: string;
}

// [Backend DTO] Paginated response wrapper
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

// [Backend DTO] Leaderboard entry (read-only, aggregated from matches)
export interface LeaderboardEntry {
  rank: number;
  user_id: string;
  username: string;
  turn_count: number;
  total_tokens: number;
  achieved_at: string;
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