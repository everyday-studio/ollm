package domain

import (
	"context"
	"time"
)

// Game represents a text-based game in the platform
type Game struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AuthorID    string    `json:"author_id"`
	Status      string    `json:"status"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateGameRequest is the DTO for creating a new game
type CreateGameRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorID    string `json:"author_id"`
}

// UpdateGameRequest is the DTO for updating an existing game
// All fields are optional (pointers indicate optional fields)
type UpdateGameRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	IsPublic    *bool   `json:"is_public"`
}

// GameRepository defines the interface for game data access
type GameRepository interface {
	Create(ctx context.Context, game *Game) (*Game, error)
	GetByID(ctx context.Context, id string) (*Game, error)
	GetAll(ctx context.Context) ([]Game, error)
	Update(ctx context.Context, game *Game) (*Game, error)
	Delete(ctx context.Context, id string) error
}

// GameUseCase defines the interface for game business logic
type GameUseCase interface {
	Create(ctx context.Context, req *CreateGameRequest) (*Game, error)
	GetByID(ctx context.Context, id string) (*Game, error)
	GetAll(ctx context.Context) ([]Game, error)
	Update(ctx context.Context, id string, req *UpdateGameRequest) (*Game, error)
	Delete(ctx context.Context, id string) error
}
