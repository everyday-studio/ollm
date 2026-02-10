package domain

import (
	"context"
	"time"
)

// MatchStatus represents the current state of a match
type MatchStatus string

const (
	MatchStatusActive   MatchStatus = "active"
	MatchStatusWon      MatchStatus = "won"
	MatchStatusLost     MatchStatus = "lost"
	MatchStatusResigned MatchStatus = "resigned"
	MatchStatusExpired  MatchStatus = "expired"
	MatchStatusError    MatchStatus = "error"
)

// Match represents an individual play record of a game
type Match struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	GameID      string      `json:"game_id"`
	Status      MatchStatus `json:"status"`
	TotalTokens int         `json:"total_tokens"`
	TurnCount   int         `json:"turn_count"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// CreateMatchRequest is the DTO for creating a new match
type CreateMatchRequest struct {
	UserID string `json:"user_id"`
	GameID string `json:"game_id"`
}

// MatchRepository defines the interface for match data access
type MatchRepository interface {
	Create(ctx context.Context, match *Match) (*Match, error)
	GetByID(ctx context.Context, id string) (*Match, error)
	GetByUserID(ctx context.Context, userID string) ([]Match, error)
	Delete(ctx context.Context, id string) error
}

// MatchUseCase defines the interface for match business logic
type MatchUseCase interface {
	Create(ctx context.Context, req *CreateMatchRequest) (*Match, error)
	GetByID(ctx context.Context, id string) (*Match, error)
	GetByUserID(ctx context.Context, userID string) ([]Match, error)
	Delete(ctx context.Context, id string) error
}
