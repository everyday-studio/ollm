package usecase

import (
	"context"
	"fmt"

	"github.com/everyday-studio/ollm/internal/domain"
)

type matchUseCase struct {
	matchRepo domain.MatchRepository
}

// NewMatchUseCase creates a new match use case
func NewMatchUseCase(matchRepo domain.MatchRepository) domain.MatchUseCase {
	return &matchUseCase{
		matchRepo: matchRepo,
	}
}

// Create creates a new match with the provided request data
func (uc *matchUseCase) Create(ctx context.Context, req *domain.CreateMatchRequest) (*domain.Match, error) {
	match := &domain.Match{
		UserID:      req.UserID,
		GameID:      req.GameID,
		Status:      domain.MatchStatusActive,
		TotalTokens: 0,
		TurnCount:   0,
	}

	createdMatch, err := uc.matchRepo.Create(ctx, match)
	if err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	return createdMatch, nil
}

// GetByID retrieves a match by its ID
func (uc *matchUseCase) GetByID(ctx context.Context, id string) (*domain.Match, error) {
	return uc.matchRepo.GetByID(ctx, id)
}

// GetByUserID retrieves all matches for a specific user
func (uc *matchUseCase) GetByUserID(ctx context.Context, userID string) ([]domain.Match, error) {
	return uc.matchRepo.GetByUserID(ctx, userID)
}

// Delete removes a match by its ID
func (uc *matchUseCase) Delete(ctx context.Context, id string) error {
	return uc.matchRepo.Delete(ctx, id)
}
