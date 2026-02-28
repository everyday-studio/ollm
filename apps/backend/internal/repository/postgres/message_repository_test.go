package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/stretchr/testify/assert"
)

// createTestMatch inserts a test match for foreign key reference (messages.match_id -> matches.id)
func createTestMatch(t *testing.T, user *domain.User, game *domain.Game) *domain.Match {
	t.Helper()
	matchRepo := NewMatchRepository(testDB)
	match := &domain.Match{
		UserID:      user.ID,
		GameID:      game.ID,
		Status:      domain.MatchStatusActive,
		MaxTurns:    10,
		TurnCount:   0,
		TotalTokens: 0,
	}
	savedMatch, err := matchRepo.Create(context.Background(), match)
	assert.NoError(t, err)
	return savedMatch
}

func TestMessageRepository_Create(t *testing.T) {
	cleanDB(t, "messages", "matches", "games", "users")
	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	user := createTestUser(t)
	game := createTestGame(t, user)
	match := createTestMatch(t, user, game)

	t.Run("Create message successfully", func(t *testing.T) {
		msg := &domain.Message{
			MatchID:    match.ID,
			Role:       domain.MessageRoleUser,
			Content:    "Hello",
			IsVisible:  true,
			TurnCount:  1,
			TokenCount: 10,
		}

		createdMsg, err := repo.Create(ctx, msg)

		assert.NoError(t, err)
		assert.NotEmpty(t, createdMsg.ID)
		assert.Equal(t, match.ID, createdMsg.MatchID)
		assert.Equal(t, domain.MessageRoleUser, createdMsg.Role)
		assert.Equal(t, "Hello", createdMsg.Content)
		assert.True(t, createdMsg.IsVisible)
		assert.Equal(t, 1, createdMsg.TurnCount)
		assert.Equal(t, 10, createdMsg.TokenCount)
		assert.NotZero(t, createdMsg.CreatedAt)
	})

	t.Run("Fail to create message with non-existent match_id", func(t *testing.T) {
		msg := &domain.Message{
			MatchID: "01HQZYX3VQJQZ3Z0Z1NONEXIST",
			Role:    domain.MessageRoleUser,
			Content: "Hello again",
		}

		createdMsg, err := repo.Create(ctx, msg)

		assert.Error(t, err)
		assert.Nil(t, createdMsg)
	})
}

func TestMessageRepository_GetByID(t *testing.T) {
	cleanDB(t, "messages", "matches", "games", "users")
	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	user := createTestUser(t)
	game := createTestGame(t, user)
	match := createTestMatch(t, user, game)

	msg := &domain.Message{
		MatchID: match.ID,
		Role:    domain.MessageRoleSystem,
		Content: "You are a helpful assistant",
	}
	createdMsg, _ := repo.Create(ctx, msg)

	t.Run("Get message by ID successfully", func(t *testing.T) {
		fetchedMsg, err := repo.GetByID(ctx, createdMsg.ID)

		assert.NoError(t, err)
		assert.Equal(t, createdMsg.ID, fetchedMsg.ID)
		assert.Equal(t, match.ID, fetchedMsg.MatchID)
		assert.Equal(t, domain.MessageRoleSystem, fetchedMsg.Role)
	})

	t.Run("Fail to get message with non-existent ID", func(t *testing.T) {
		fetchedMsg, err := repo.GetByID(ctx, "01HQZYX3VQJQZ3Z0Z1ZNONEXIST")

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, fetchedMsg)
	})
}

func TestMessageRepository_GetByMatchID(t *testing.T) {
	cleanDB(t, "messages", "matches", "games", "users")
	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	user := createTestUser(t)
	game := createTestGame(t, user)
	match1 := createTestMatch(t, user, game)
	match2 := createTestMatch(t, user, game)

	// Create messages for match1
	msg1 := &domain.Message{MatchID: match1.ID, Role: domain.MessageRoleUser, Content: "Msg 1"}
	msg2 := &domain.Message{MatchID: match1.ID, Role: domain.MessageRoleAssistant, Content: "Msg 2"}
	repo.Create(ctx, msg1)
	time.Sleep(1 * time.Millisecond) // ensure ordering
	repo.Create(ctx, msg2)

	// Create message for match2
	msg3 := &domain.Message{MatchID: match2.ID, Role: domain.MessageRoleUser, Content: "Msg 3"}
	repo.Create(ctx, msg3)

	t.Run("Get all messages by match ID successfully", func(t *testing.T) {
		messages, err := repo.GetByMatchID(ctx, match1.ID)

		assert.NoError(t, err)
		assert.Len(t, messages, 2)
		// Ordered by created_at ASC
		assert.Equal(t, "Msg 1", messages[0].Content)
		assert.Equal(t, "Msg 2", messages[1].Content)
	})

	t.Run("Return empty slice when no messages exist for match", func(t *testing.T) {
		match3 := createTestMatch(t, user, game)
		messages, err := repo.GetByMatchID(ctx, match3.ID)

		assert.NoError(t, err)
		assert.Len(t, messages, 0)
	})
}

func TestMessageRepository_Update(t *testing.T) {
	cleanDB(t, "messages", "matches", "games", "users")
	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	user := createTestUser(t)
	game := createTestGame(t, user)
	match := createTestMatch(t, user, game)

	msg := &domain.Message{
		MatchID: match.ID,
		Role:    domain.MessageRoleUser,
		Content: "Original content",
	}
	createdMsg, _ := repo.Create(ctx, msg)

	t.Run("Update message successfully", func(t *testing.T) {
		createdMsg.Content = "Updated content"
		createdMsg.IsVisible = false
		createdMsg.TurnCount = 5
		createdMsg.TokenCount = 100

		updatedMsg, err := repo.Update(ctx, createdMsg)

		assert.NoError(t, err)
		assert.Equal(t, "Updated content", updatedMsg.Content)
		assert.False(t, updatedMsg.IsVisible)
		assert.Equal(t, 5, updatedMsg.TurnCount)
		assert.Equal(t, 100, updatedMsg.TokenCount)
	})
}

func TestMessageRepository_Delete(t *testing.T) {
	cleanDB(t, "messages", "matches", "games", "users")
	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	user := createTestUser(t)
	game := createTestGame(t, user)
	match := createTestMatch(t, user, game)

	msg := &domain.Message{
		MatchID: match.ID,
		Role:    domain.MessageRoleUser,
		Content: "To be deleted",
	}
	createdMsg, _ := repo.Create(ctx, msg)

	t.Run("Delete message successfully", func(t *testing.T) {
		err := repo.Delete(ctx, createdMsg.ID)
		assert.NoError(t, err)

		fetchedMsg, err := repo.GetByID(ctx, createdMsg.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, fetchedMsg)
	})

	t.Run("Fail to delete non-existent message", func(t *testing.T) {
		err := repo.Delete(ctx, "01HQZYX3VQJQZ3Z0Z1ZNONEXIST")
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
