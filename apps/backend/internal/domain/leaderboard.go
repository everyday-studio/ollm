package domain

import (
	"context"
	"time"
)

// LeaderboardEntry represents a single row in the leaderboard response.
// Note: This is a read-only DTO, not a database table entity.
type LeaderboardEntry struct {
	Rank        int       `json:"rank"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	TurnCount   int       `json:"turn_count"`
	TotalTokens int       `json:"total_tokens"`
	AchievedAt  time.Time `json:"achieved_at"`
}

// LeaderboardRepository defines the interface for leaderboard data access
type LeaderboardRepository interface {
	GetLeaderboard(ctx context.Context, gameID string, limit int) ([]LeaderboardEntry, error)
}

// LeaderboardUseCase defines the interface for leaderboard business logic
type LeaderboardUseCase interface {
	GetLeaderboard(ctx context.Context, gameID string, limit int) ([]LeaderboardEntry, error)
}
