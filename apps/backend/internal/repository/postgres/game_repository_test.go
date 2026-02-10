package postgres

import (
	"context"
	"testing"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/stretchr/testify/assert"
)

// createTestUser inserts a test user for foreign key reference (games.author_id -> users.id)
func createTestUser(t *testing.T) *domain.User {
	t.Helper()
	userRepo := NewUserRepository(testDB)
	user := &domain.User{Name: "TestAuthor", Email: "author@example.com", Password: "testpassword"}
	savedUser, err := userRepo.Save(context.Background(), user)
	assert.NoError(t, err)
	return savedUser
}

func TestGameRepository_Create(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewGameRepository(testDB)
	author := createTestUser(t)

	t.Run("Create game successfully", func(t *testing.T) {
		game := &domain.Game{
			Title:       "Adventure Quest",
			Description: "A text-based adventure game",
			AuthorID:    author.ID,
		}

		createdGame, err := repo.Create(ctx, game)

		assert.NoError(t, err)
		assert.NotEmpty(t, createdGame.ID)
		assert.Equal(t, "Adventure Quest", createdGame.Title)
		assert.Equal(t, "A text-based adventure game", createdGame.Description)
		assert.Equal(t, author.ID, createdGame.AuthorID)
		assert.Equal(t, "active", createdGame.Status) // default status
		assert.False(t, createdGame.IsPublic)         // default is false from Go zero value
		assert.NotZero(t, createdGame.CreatedAt)
		assert.NotZero(t, createdGame.UpdatedAt)
	})

	t.Run("Create game with explicit status and is_public", func(t *testing.T) {
		game := &domain.Game{
			Title:       "Mystery Dungeon",
			Description: "A mystery game",
			AuthorID:    author.ID,
			Status:      "inactive",
			IsPublic:    true,
		}

		createdGame, err := repo.Create(ctx, game)

		assert.NoError(t, err)
		assert.Equal(t, "inactive", createdGame.Status)
		assert.True(t, createdGame.IsPublic)
	})

	t.Run("Fail to create game with non-existent author_id", func(t *testing.T) {
		game := &domain.Game{
			Title:       "Ghost Game",
			Description: "Author does not exist",
			AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
		}

		createdGame, err := repo.Create(ctx, game)

		assert.Error(t, err)
		assert.Nil(t, createdGame)
	})

	t.Run("Fail to create game with invalid status", func(t *testing.T) {
		game := &domain.Game{
			Title:    "Bad Status Game",
			AuthorID: author.ID,
			Status:   "invalid_status",
		}

		createdGame, err := repo.Create(ctx, game)

		assert.Error(t, err)
		assert.Nil(t, createdGame)
	})
}

func TestGameRepository_GetByID(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewGameRepository(testDB)
	author := createTestUser(t)

	t.Run("Get game by ID successfully", func(t *testing.T) {
		game := &domain.Game{
			Title:       "Adventure Quest",
			Description: "A text-based adventure game",
			AuthorID:    author.ID,
		}
		createdGame, _ := repo.Create(ctx, game)

		fetchedGame, err := repo.GetByID(ctx, createdGame.ID)

		assert.NoError(t, err)
		assert.Equal(t, createdGame.ID, fetchedGame.ID)
		assert.Equal(t, createdGame.Title, fetchedGame.Title)
		assert.Equal(t, createdGame.Description, fetchedGame.Description)
		assert.Equal(t, createdGame.AuthorID, fetchedGame.AuthorID)
	})

	t.Run("Fail to get game with non-existent ID", func(t *testing.T) {
		fetchedGame, err := repo.GetByID(ctx, "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST")

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, fetchedGame)
	})
}

func TestGameRepository_GetAll(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewGameRepository(testDB)
	author := createTestUser(t)

	t.Run("Get all games successfully", func(t *testing.T) {
		game1 := &domain.Game{Title: "Game 1", Description: "First game", AuthorID: author.ID}
		game2 := &domain.Game{Title: "Game 2", Description: "Second game", AuthorID: author.ID}
		repo.Create(ctx, game1)
		repo.Create(ctx, game2)

		games, err := repo.GetAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, games, 2)
		// Ordered by created_at DESC, so Game 2 comes first
		assert.Equal(t, "Game 2", games[0].Title)
		assert.Equal(t, "Game 1", games[1].Title)
	})

	t.Run("Return empty slice when no games exist", func(t *testing.T) {
		cleanDB(t, "matches", "games", "users")

		games, err := repo.GetAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, games, 0)
	})
}

func TestGameRepository_Update(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewGameRepository(testDB)
	author := createTestUser(t)

	t.Run("Update game successfully", func(t *testing.T) {
		game := &domain.Game{
			Title:       "Original Title",
			Description: "Original description",
			AuthorID:    author.ID,
		}
		createdGame, _ := repo.Create(ctx, game)

		// Modify fields
		createdGame.Title = "Updated Title"
		createdGame.Description = "Updated description"
		createdGame.IsPublic = true

		updatedGame, err := repo.Update(ctx, createdGame)

		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", updatedGame.Title)
		assert.Equal(t, "Updated description", updatedGame.Description)
		assert.True(t, updatedGame.IsPublic)
	})

	t.Run("Fail to update non-existent game", func(t *testing.T) {
		game := &domain.Game{
			ID:    "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			Title: "Ghost Game",
		}

		updatedGame, err := repo.Update(ctx, game)

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, updatedGame)
	})
}

func TestGameRepository_Delete(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewGameRepository(testDB)
	author := createTestUser(t)

	t.Run("Delete game successfully", func(t *testing.T) {
		game := &domain.Game{
			Title:       "To Be Deleted",
			Description: "This game will be deleted",
			AuthorID:    author.ID,
		}
		createdGame, _ := repo.Create(ctx, game)

		err := repo.Delete(ctx, createdGame.ID)
		assert.NoError(t, err)

		// Verify deletion
		fetchedGame, err := repo.GetByID(ctx, createdGame.ID)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, fetchedGame)
	})

	t.Run("Fail to delete non-existent game", func(t *testing.T) {
		err := repo.Delete(ctx, "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST")

		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
