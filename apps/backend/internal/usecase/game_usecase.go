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
	maxTurns := req.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 5 // Default to 5 turns if not specified
	}

	game := &domain.Game{
		Title:        req.Title,
		Description:  req.Description,
		AuthorID:     req.AuthorID,
		Status:       domain.GameStatusActive,
		IsPublic:     true,
		SystemPrompt: req.SystemPrompt,
		TargetWord:   req.TargetWord,
		MaxTurns:     maxTurns,
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

// CountAll returns the total number of games
func (uc *gameUseCase) CountAll(ctx context.Context) (int, error) {
	return uc.gameRepo.CountAll(ctx)
}

// GetPaginated retrieves a paginated list of games
func (uc *gameUseCase) GetPaginated(ctx context.Context, page, limit int) (*domain.PaginatedData[domain.Game], error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	total, err := uc.gameRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}

	games, err := uc.gameRepo.GetPaginated(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit

	return &domain.PaginatedData[domain.Game]{
		Data:       games,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// Update updates an existing game
func (uc *gameUseCase) Update(ctx context.Context, id string, req *domain.UpdateGameRequest) (*domain.Game, error) {
	// Get existing game
	existingGame, err := uc.gameRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get game by id: %w", err)
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

	if req.SystemPrompt != nil {
		existingGame.SystemPrompt = *req.SystemPrompt
	}

	if req.TargetWord != nil {
		existingGame.TargetWord = *req.TargetWord
	}

	if req.MaxTurns != nil {
		existingGame.MaxTurns = *req.MaxTurns
	}

	updatedGame, err := uc.gameRepo.Update(ctx, existingGame)
	if err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return updatedGame, nil
}

// Delete removes a game by its ID
func (uc *gameUseCase) Delete(ctx context.Context, id string) error {
	return uc.gameRepo.Delete(ctx, id)
}
