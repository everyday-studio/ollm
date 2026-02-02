package usecase

import (
	"context"
	"fmt"

	"github.com/everyday-studio/ollm/internal/domain"
)

type gameUseCase struct {
	gameRepo domain.GameRepository
}

// NewGameUseCase creates a new game use case
func NewGameUseCase(gameRepo domain.GameRepository) domain.GameUseCase {
	return &gameUseCase{
		gameRepo: gameRepo,
	}
}

// Create creates a new game with the provided request data
func (uc *gameUseCase) Create(ctx context.Context, req *domain.CreateGameRequest) (*domain.Game, error) {
	game := &domain.Game{
		Title:       req.Title,
		Description: req.Description,
		AuthorID:    req.AuthorID,
		Status:      "active",
		IsPublic:    true,
	}

	createdGame, err := uc.gameRepo.Create(ctx, game)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return createdGame, nil
}

// GetByID retrieves a game by its ID
func (uc *gameUseCase) GetByID(ctx context.Context, id string) (*domain.Game, error) {
	return uc.gameRepo.GetByID(ctx, id)
}

// GetAll retrieves all games
func (uc *gameUseCase) GetAll(ctx context.Context) ([]domain.Game, error) {
	return uc.gameRepo.GetAll(ctx)
}

// Update updates an existing game
func (uc *gameUseCase) Update(ctx context.Context, id string, req *domain.UpdateGameRequest) (*domain.Game, error) {
	// Get existing game
	existingGame, err := uc.gameRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if req.Title != nil {
		existingGame.Title = *req.Title
	}

	if req.Description != nil {
		existingGame.Description = *req.Description
	}

	if req.Status != nil {
		existingGame.Status = *req.Status
	}

	if req.IsPublic != nil {
		existingGame.IsPublic = *req.IsPublic
	}

	// Save updated game
	return uc.gameRepo.Update(ctx, existingGame)
}

// Delete removes a game by its ID
func (uc *gameUseCase) Delete(ctx context.Context, id string) error {
	return uc.gameRepo.Delete(ctx, id)
}
