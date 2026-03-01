package usecase

import (
	"context"
	"testing"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
)

func TestLeaderboardUseCase_GetLeaderboard(t *testing.T) {
	t.Run("Get leaderboard successfully", func(t *testing.T) {
		mockRepo := new(mocks.LeaderboardRepository)
		uc := NewLeaderboardUseCase(mockRepo)

		gameID := "test_game_id"
		ctx := context.Background()

		expectedEntries := []domain.LeaderboardEntry{
			{Rank: 1, UserID: "user_1", Username: "User 1", TurnCount: 3},
			{Rank: 2, UserID: "user_2", Username: "User 2", TurnCount: 5},
		}

		mockRepo.On("GetLeaderboard", ctx, gameID, 10).Return(expectedEntries, nil)

		entries, err := uc.GetLeaderboard(ctx, gameID, 10)

		assert.NoError(t, err)
		assert.Len(t, entries, 2)
		assert.Equal(t, expectedEntries, entries)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Return empty slice if nil returned from repo", func(t *testing.T) {
		mockRepo := new(mocks.LeaderboardRepository)
		uc := NewLeaderboardUseCase(mockRepo)

		gameID := "test_game_id"
		ctx := context.Background()

		mockRepo.On("GetLeaderboard", ctx, gameID, 10).Return(nil, nil)

		entries, err := uc.GetLeaderboard(ctx, gameID, 10)

		assert.NoError(t, err)
		assert.NotNil(t, entries)
		assert.Len(t, entries, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Default limit is 10 if limit <= 0", func(t *testing.T) {
		mockRepo := new(mocks.LeaderboardRepository)
		uc := NewLeaderboardUseCase(mockRepo)

		gameID := "test_game_id"
		ctx := context.Background()

		mockRepo.On("GetLeaderboard", ctx, gameID, 10).Return([]domain.LeaderboardEntry{}, nil) // EXPECT 10 limit here

		entries, err := uc.GetLeaderboard(ctx, gameID, -5)

		assert.NoError(t, err)
		assert.NotNil(t, entries)
		assert.Len(t, entries, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Return repo error", func(t *testing.T) {
		mockRepo := new(mocks.LeaderboardRepository)
		uc := NewLeaderboardUseCase(mockRepo)

		gameID := "test_game_id"
		ctx := context.Background()

		mockRepo.On("GetLeaderboard", ctx, gameID, 10).Return(nil, domain.ErrInternal)

		entries, err := uc.GetLeaderboard(ctx, gameID, 10)

		assert.ErrorIs(t, err, domain.ErrInternal)
		assert.Nil(t, entries)
		mockRepo.AssertExpectations(t)
	})
}
