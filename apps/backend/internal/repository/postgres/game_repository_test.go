package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/stretchr/testify/assert"
)

// createTestUser inserts a test user for foreign key reference (games.author_id -> users.id)
func createTestUser(t *testing.T) *domain.User {
	t.Helper()
	userRepo := NewUserRepository(testDB)
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%10000)
	user := &domain.User{Name: "TestAuthor", Tag: uniqueSuffix, Email: fmt.Sprintf("author%s@example.com", uniqueSuffix), Password: "testpassword"}
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
			Title:        "Adventure Quest",
			Description:  "A text-based adventure game",
			AuthorID:     author.ID,
			Status:       domain.GameStatusActive,
			IsPublic:     false,
			FirstMessage: "Welcome to the adventure!",
		}

		createdGame, err := repo.Create(ctx, game)

		assert.NoError(t, err)
		assert.NotEmpty(t, createdGame.ID)
		assert.Equal(t, "Adventure Quest", createdGame.Title)
		assert.Equal(t, "A text-based adventure game", createdGame.Description)
		assert.Equal(t, author.ID, createdGame.AuthorID)
		assert.Equal(t, "Welcome to the adventure!", createdGame.FirstMessage)
		assert.Equal(t, domain.GameStatusActive, createdGame.Status)
		assert.False(t, createdGame.IsPublic)
		assert.Equal(t, 0, createdGame.PlayCount)
		assert.NotZero(t, createdGame.CreatedAt)
		assert.NotZero(t, createdGame.UpdatedAt)
	})

	t.Run("Create game with explicit status and is_public", func(t *testing.T) {
		game := &domain.Game{
			Title:       "Mystery Dungeon",
			Description: "A mystery game",
			AuthorID:    author.ID,
			Status:      domain.GameStatusInactive,
			IsPublic:    true,
		}

		createdGame, err := repo.Create(ctx, game)

		assert.NoError(t, err)
		assert.Equal(t, domain.GameStatusInactive, createdGame.Status)
		assert.True(t, createdGame.IsPublic)
	})

	t.Run("Fail to create game with non-existent author_id", func(t *testing.T) {
		game := &domain.Game{
			Title:       "Ghost Game",
			Description: "Author does not exist",
			AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			Status:      domain.GameStatusActive,
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
			Status:      domain.GameStatusActive,
		}
		createdGame, _ := repo.Create(ctx, game)

		fetchedGame, err := repo.GetByID(ctx, createdGame.ID)

		assert.NoError(t, err)
		assert.Equal(t, createdGame.ID, fetchedGame.ID)
		assert.Equal(t, createdGame.Title, fetchedGame.Title)
		assert.Equal(t, createdGame.Description, fetchedGame.Description)
		assert.Equal(t, createdGame.AuthorID, fetchedGame.AuthorID)
		assert.Equal(t, 0, fetchedGame.PlayCount)
	})

	t.Run("Fail to get game with non-existent ID", func(t *testing.T) {
		fetchedGame, err := repo.GetByID(ctx, "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST")

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, fetchedGame)
	})
}

func TestGameRepository_CountAllAndGetPaginated(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewGameRepository(testDB)
	author := createTestUser(t)

	t.Run("Get paginated games and count successfully", func(t *testing.T) {
		game1 := &domain.Game{Title: "Game 1", Description: "First game", AuthorID: author.ID, Status: domain.GameStatusActive, IsPublic: true}
		game2 := &domain.Game{Title: "Game 2", Description: "Second game", AuthorID: author.ID, Status: domain.GameStatusActive, IsPublic: false}
		repo.Create(ctx, game1)
		repo.Create(ctx, game2)

		// 1. All games
		total, err := repo.CountAll(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)

		games, err := repo.GetPaginated(ctx, 1, 10, nil)

		assert.NoError(t, err)
		assert.Len(t, games, 2)
		// Default sort is updated_at DESC, so Game 2 (created later) comes first
		assert.Equal(t, "Game 2", games[0].Title)
		assert.Equal(t, "Game 1", games[1].Title)

		// 2. Public games only
		isPublic := true
		publicFilter := &domain.GameFilter{IsPublic: &isPublic}
		totalPublic, err := repo.CountAll(ctx, publicFilter)
		assert.NoError(t, err)
		assert.Equal(t, 1, totalPublic)

		publicGames, err := repo.GetPaginated(ctx, 1, 10, publicFilter)
		assert.NoError(t, err)
		assert.Len(t, publicGames, 1)
		assert.Equal(t, "Game 1", publicGames[0].Title)

		// 3. Private games only
		isNotPublic := false
		privateFilter := &domain.GameFilter{IsPublic: &isNotPublic}
		totalPrivate, err := repo.CountAll(ctx, privateFilter)
		assert.NoError(t, err)
		assert.Equal(t, 1, totalPrivate)

		privateGames, err := repo.GetPaginated(ctx, 1, 10, privateFilter)
		assert.NoError(t, err)
		assert.Len(t, privateGames, 1)
		assert.Equal(t, "Game 2", privateGames[0].Title)
	})

	t.Run("Return empty slice when no games exist", func(t *testing.T) {
		cleanDB(t, "matches", "games", "users")

		total, err := repo.CountAll(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, total)

		games, err := repo.GetPaginated(ctx, 1, 10, nil)

		assert.NoError(t, err)
		assert.Len(t, games, 0)
	})
}

func TestGameRepository_GetPaginated_Sort(t *testing.T) {
	cleanDB(t, "matches", "games", "users")
	ctx := context.Background()
	repo := NewGameRepository(testDB)
	author := createTestUser(t)

	// Create games: Zebra (0 matches), Alpha (2 matches), Middle (1 match)
	// createTestGame (defined in match_repository_test.go) creates a fixed-title game,
	// so we create directly here to control the titles.
	newGame := func(title string) *domain.Game {
		g, err := repo.Create(ctx, &domain.Game{
			Title:    title,
			AuthorID: author.ID,
			Status:   domain.GameStatusActive,
			IsPublic: true,
		})
		assert.NoError(t, err)
		return g
	}

	zebra := newGame("Zebra Game")
	alpha := newGame("Alpha Game")
	middle := newGame("Middle Game")

	// Create matches via createTestMatch (defined in message_repository_test.go).
	// Each match INSERT triggers play_count +1 on the parent game.
	// Match creation order: alpha x2, middle x1
	// → updated_at order (DESC): middle > alpha > zebra
	createTestMatch(t, author, alpha)
	createTestMatch(t, author, alpha)
	createTestMatch(t, author, middle)

	t.Run("Sort by name ASC", func(t *testing.T) {
		filter := &domain.GameFilter{SortBy: domain.GameSortByName}
		games, err := repo.GetPaginated(ctx, 1, 10, filter)

		assert.NoError(t, err)
		assert.Len(t, games, 3)
		assert.Equal(t, "Alpha Game", games[0].Title)
		assert.Equal(t, "Middle Game", games[1].Title)
		assert.Equal(t, "Zebra Game", games[2].Title)
	})

	t.Run("Sort by popular DESC (play_count)", func(t *testing.T) {
		filter := &domain.GameFilter{SortBy: domain.GameSortByPopular}
		games, err := repo.GetPaginated(ctx, 1, 10, filter)

		assert.NoError(t, err)
		assert.Len(t, games, 3)
		assert.Equal(t, "Alpha Game", games[0].Title)
		assert.Equal(t, 2, games[0].PlayCount)
		assert.Equal(t, "Middle Game", games[1].Title)
		assert.Equal(t, 1, games[1].PlayCount)
		assert.Equal(t, "Zebra Game", games[2].Title)
		assert.Equal(t, 0, games[2].PlayCount)
	})

	t.Run("Sort by recent DESC (updated_at)", func(t *testing.T) {
		// updated_at: middle (T_last match) > alpha (T_second match) > zebra (T_created, no matches)
		filter := &domain.GameFilter{SortBy: domain.GameSortByRecent}
		games, err := repo.GetPaginated(ctx, 1, 10, filter)

		assert.NoError(t, err)
		assert.Len(t, games, 3)
		assert.Equal(t, "Middle Game", games[0].Title)
		assert.Equal(t, "Alpha Game", games[1].Title)
		assert.Equal(t, "Zebra Game", games[2].Title)
	})

	t.Run("play_count reflected in GetByID after matches created", func(t *testing.T) {
		fetchedAlpha, err := repo.GetByID(ctx, alpha.ID)
		assert.NoError(t, err)
		assert.Equal(t, 2, fetchedAlpha.PlayCount)

		fetchedMiddle, err := repo.GetByID(ctx, middle.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, fetchedMiddle.PlayCount)

		fetchedZebra, err := repo.GetByID(ctx, zebra.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, fetchedZebra.PlayCount)
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
			Status:      domain.GameStatusActive,
		}
		createdGame, _ := repo.Create(ctx, game)

		// Modify fields
		createdGame.Title = "Updated Title"
		createdGame.Description = "Updated description"
		createdGame.FirstMessage = "New greeting!"
		createdGame.IsPublic = true

		updatedGame, err := repo.Update(ctx, createdGame)

		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", updatedGame.Title)
		assert.Equal(t, "Updated description", updatedGame.Description)
		assert.Equal(t, "New greeting!", updatedGame.FirstMessage)
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
			Status:      domain.GameStatusActive,
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
