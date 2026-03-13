package usecase

import (
	"context"
	"fmt"

	"github.com/everyday-studio/ollm/internal/domain"
)

type matchUseCase struct {
	matchRepo domain.MatchRepository
	gameRepo  domain.GameRepository
}

// NewMatchUseCase creates a new match use case
func NewMatchUseCase(matchRepo domain.MatchRepository, gameRepo domain.GameRepository) domain.MatchUseCase {
	return &matchUseCase{
		matchRepo: matchRepo,
		gameRepo:  gameRepo,
	}
}

// Create creates a new match with the provided request data
func (uc *matchUseCase) Create(ctx context.Context, req *domain.CreateMatchRequest) (*domain.Match, error) {
	// Get game to copy max_turns
	game, err := uc.gameRepo.GetByID(ctx, req.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game for match creation: %w", err)
	}

	// Restrict to max 5 active matches per game
	count, err := uc.matchRepo.CountByUserIDGameIDAndStatus(ctx, req.UserID, req.GameID, domain.MatchStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to count active matches: %w", err)
	}

	if count >= 5 {
		return nil, fmt.Errorf("%w: maximum number of active matches (5) for this game reached", domain.ErrConflict)
	}

	match := &domain.Match{
		UserID:      req.UserID,
		GameID:      req.GameID,
		Status:      domain.MatchStatusActive,
		MaxTurns:    game.MaxTurns,
		TotalTokens: 0,
		TurnCount:   0,
	}

	createdMatch, err := uc.matchRepo.Create(ctx, match)
	if err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	return createdMatch, nil
}

// GetByID retrieves a match by its ID and validates ownership
func (uc *matchUseCase) GetByID(ctx context.Context, id string, userID string) (*domain.Match, error) {
	match, err := uc.matchRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	if match.UserID != userID {
		return nil, domain.ErrForbidden
	}

	return match, nil
}

// GetByUserID retrieves all matches for a specific user
func (uc *matchUseCase) GetByUserID(ctx context.Context, userID string) ([]domain.Match, error) {
	return uc.matchRepo.GetByUserID(ctx, userID)
}

// GetByUserIDAndGameID retrieves all matches for a specific user and game
func (uc *matchUseCase) GetByUserIDAndGameID(ctx context.Context, userID string, gameID string) ([]domain.Match, error) {
	return uc.matchRepo.GetByUserIDAndGameID(ctx, userID, gameID)
}

// Resign allows a user to voluntarily forfeit a match
func (uc *matchUseCase) Resign(ctx context.Context, id string, userID string) error {
	match, err := uc.matchRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get match for resignation: %w", err)
	}

	if match.UserID != userID {
		return domain.ErrForbidden
	}

	if match.Status != domain.MatchStatusActive {
		return fmt.Errorf("%w: match is not active", domain.ErrConflict)
	}

	match.Status = domain.MatchStatusResigned
	_, err = uc.matchRepo.Update(ctx, match)
	if err != nil {
		return fmt.Errorf("failed to update match status to resigned: %w", err)
	}

	return nil
}

// Delete removes a match by its ID
func (uc *matchUseCase) Delete(ctx context.Context, id string) error {
	return uc.matchRepo.Delete(ctx, id)
}
