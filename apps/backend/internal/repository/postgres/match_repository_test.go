package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/stretchr/testify/assert"
)

// createTestGame inserts a test game for foreign key reference (matches.game_id -> games.id)
func createTestGame(t *testing.T, author *domain.User) *domain.Game {
	t.Helper()
	gameRepo := NewGameRepository(testDB)
	game := &domain.Game{
		Title:       "Test Game for Match",
		Description: "A game used for match tests",
		AuthorID:    author.ID,
		Status:      domain.GameStatusActive,
		IsPublic:    true,
	}
	savedGame, err := gameRepo.Create(context.Background(), game)
	assert.NoError(t, err)
	return savedGame
}

func TestMatchRepository_Create(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewMatchRepository(testDB)
	user := createTestUser(t)
	game := createTestGame(t, user)

	t.Run("Create match successfully", func(t *testing.T) {
		match := &domain.Match{
			UserID:      user.ID,
			GameID:      game.ID,
			Status:      domain.MatchStatusActive,
			MaxTurns:    10,
			TotalTokens: 0,
			TurnCount:   0,
		}

		createdMatch, err := repo.Create(ctx, match)

		assert.NoError(t, err)
		assert.NotEmpty(t, createdMatch.ID)
		assert.Equal(t, user.ID, createdMatch.UserID)
		assert.Equal(t, game.ID, createdMatch.GameID)
		assert.Equal(t, domain.MatchStatusActive, createdMatch.Status)
		assert.Equal(t, 10, createdMatch.MaxTurns)
		assert.Equal(t, 0, createdMatch.TotalTokens)
		assert.Equal(t, 0, createdMatch.TurnCount)
		assert.NotZero(t, createdMatch.CreatedAt)
		assert.NotZero(t, createdMatch.UpdatedAt)
	})

	t.Run("Fail to create match with non-existent user_id", func(t *testing.T) {
		match := &domain.Match{
			UserID: "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			GameID: game.ID,
			Status: domain.MatchStatusActive,
		}

		createdMatch, err := repo.Create(ctx, match)

		assert.Error(t, err)
		assert.Nil(t, createdMatch)
	})

	t.Run("Fail to create match with non-existent game_id", func(t *testing.T) {
		match := &domain.Match{
			UserID: user.ID,
			GameID: "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			Status: domain.MatchStatusActive,
		}

		createdMatch, err := repo.Create(ctx, match)

		assert.Error(t, err)
		assert.Nil(t, createdMatch)
	})
}

func TestMatchRepository_GetByID(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewMatchRepository(testDB)
	user := createTestUser(t)
	game := createTestGame(t, user)

	t.Run("Get match by ID successfully", func(t *testing.T) {
		match := &domain.Match{
			UserID: user.ID,
			GameID: game.ID,
			Status: domain.MatchStatusActive,
		}
		createdMatch, _ := repo.Create(ctx, match)

		fetchedMatch, err := repo.GetByID(ctx, createdMatch.ID)

		assert.NoError(t, err)
		assert.Equal(t, createdMatch.ID, fetchedMatch.ID)
		assert.Equal(t, createdMatch.UserID, fetchedMatch.UserID)
		assert.Equal(t, createdMatch.GameID, fetchedMatch.GameID)
	})

	t.Run("Fail to get match with non-existent ID", func(t *testing.T) {
		fetchedMatch, err := repo.GetByID(ctx, "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST")

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, fetchedMatch)
	})
}

func TestMatchRepository_GetByUserID(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewMatchRepository(testDB)
	user := createTestUser(t)
	game1 := createTestGame(t, user)
	game2 := createTestGame(t, user)

	t.Run("Get all matches by user ID successfully", func(t *testing.T) {
		match1 := &domain.Match{UserID: user.ID, GameID: game1.ID, Status: domain.MatchStatusActive}
		match2 := &domain.Match{UserID: user.ID, GameID: game2.ID, Status: domain.MatchStatusWon}
		repo.Create(ctx, match1)
		time.Sleep(1 * time.Millisecond) // ensure ordering
		repo.Create(ctx, match2)

		matches, err := repo.GetByUserID(ctx, user.ID)

		assert.NoError(t, err)
		assert.Len(t, matches, 2)
		// Assuming order by created_at DESC
		assert.Equal(t, domain.MatchStatusWon, matches[0].Status)
		assert.Equal(t, domain.MatchStatusActive, matches[1].Status)
	})

	t.Run("Return empty slice when no matches exist for user", func(t *testing.T) {
		cleanDB(t, "matches", "games", "users")
		user2 := createTestUser(t)

		matches, err := repo.GetByUserID(ctx, user2.ID)

		assert.NoError(t, err)
		assert.Len(t, matches, 0)
	})
}

func TestMatchRepository_GetByUserIDAndGameID(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewMatchRepository(testDB)
	user := createTestUser(t)
	game1 := createTestGame(t, user)
	game2 := createTestGame(t, user)

	t.Run("Get matches by user ID and game ID successfully", func(t *testing.T) {
		match1 := &domain.Match{UserID: user.ID, GameID: game1.ID, Status: domain.MatchStatusActive}
		match2 := &domain.Match{UserID: user.ID, GameID: game1.ID, Status: domain.MatchStatusWon}
		match3 := &domain.Match{UserID: user.ID, GameID: game2.ID, Status: domain.MatchStatusActive}
		repo.Create(ctx, match1)
		repo.Create(ctx, match2)
		repo.Create(ctx, match3)

		matches, err := repo.GetByUserIDAndGameID(ctx, user.ID, game1.ID)

		assert.NoError(t, err)
		assert.Len(t, matches, 2)

		for _, m := range matches {
			assert.Equal(t, game1.ID, m.GameID)
			assert.Equal(t, user.ID, m.UserID)
		}
	})
}

func TestMatchRepository_Update(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewMatchRepository(testDB)
	user := createTestUser(t)
	game := createTestGame(t, user)

	t.Run("Update match successfully", func(t *testing.T) {
		match := &domain.Match{
			UserID:    user.ID,
			GameID:    game.ID,
			Status:    domain.MatchStatusActive,
			TurnCount: 1,
		}
		createdMatch, _ := repo.Create(ctx, match)

		// Modify fields
		createdMatch.Status = domain.MatchStatusWon
		createdMatch.TurnCount = 5
		createdMatch.TotalTokens = 150

		updatedMatch, err := repo.Update(ctx, createdMatch)

		assert.NoError(t, err)
		assert.Equal(t, domain.MatchStatusWon, updatedMatch.Status)
		assert.Equal(t, 5, updatedMatch.TurnCount)
		assert.Equal(t, 150, updatedMatch.TotalTokens)
	})

	t.Run("Fail to update non-existent match", func(t *testing.T) {
		match := &domain.Match{
			ID:     "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			Status: domain.MatchStatusResigned,
		}

		updatedMatch, err := repo.Update(ctx, match)

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, updatedMatch)
	})
}

func TestMatchRepository_Delete(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewMatchRepository(testDB)
	user := createTestUser(t)
	game := createTestGame(t, user)

	t.Run("Delete match successfully", func(t *testing.T) {
		match := &domain.Match{
			UserID: user.ID,
			GameID: game.ID,
			Status: domain.MatchStatusActive,
		}
		createdMatch, _ := repo.Create(ctx, match)

		err := repo.Delete(ctx, createdMatch.ID)
		assert.NoError(t, err)

		// Verify deletion
		fetchedMatch, err := repo.GetByID(ctx, createdMatch.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, fetchedMatch)
	})

	t.Run("Fail to delete non-existent match", func(t *testing.T) {
		err := repo.Delete(ctx, "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST")

		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
