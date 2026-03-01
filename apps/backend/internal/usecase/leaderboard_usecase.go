package usecase

import (
	"context"

	"github.com/everyday-studio/ollm/internal/domain"
)

type leaderboardUseCase struct {
	repo domain.LeaderboardRepository
}

// NewLeaderboardUseCase creates a new leaderboard usecase
func NewLeaderboardUseCase(repo domain.LeaderboardRepository) domain.LeaderboardUseCase {
	return &leaderboardUseCase{
		repo: repo,
	}
}

// GetLeaderboard returns the leaderboard for a specific game, ranked and formatted
func (uc *leaderboardUseCase) GetLeaderboard(ctx context.Context, gameID string, limit int) ([]domain.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	entries, err := uc.repo.GetLeaderboard(ctx, gameID, limit)
	if err != nil {
		return nil, err
	}

	// Assign ranks
	for i := range entries {
		entries[i].Rank = i + 1
	}

	// Make sure we don't return nil for an empty leaderboard, return empty slice instead
	if entries == nil {
		return []domain.LeaderboardEntry{}, nil
	}

	return entries, nil
}
