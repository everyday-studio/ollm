package postgres

import (
	"context"
	"testing"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestSaveUser(t *testing.T) {
	repo := NewUserRepository(testDB)
	cleanDB(t, "users")
	ctx := context.Background()

	t.Run("Save user successfully", func(t *testing.T) {
		user := &domain.User{Name: "Jane Doe", Email: "jane@exmaple.com"}
		savedUser, err := repo.Save(ctx, user)

		assert.NoError(t, err)
		assert.NotZero(t, savedUser.ID)
		assert.Equal(t, user.Name, savedUser.Name)
		assert.Equal(t, user.Email, savedUser.Email)
	})

	t.Run("Fails to save user due to existing email", func(t *testing.T) {
		user1 := &domain.User{Name: "Jane Doe", Email: "jane@example.com"}
		_, err := repo.Save(ctx, user1)
		assert.NoError(t, err)

		user2 := &domain.User{Name: "Jane Smith", Email: "jane@example.com"}
		_, err = repo.Save(ctx, user2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAlreadyExists)
	})
}

func TestGetByID(t *testing.T) {
	repo := NewUserRepository(testDB)
	cleanDB(t, "users")
	ctx := context.Background()

	t.Run("Get users by id successfully", func(t *testing.T) {
		user := &domain.User{Name: "Alice", Email: "alice@example.com"}
		savedUser, _ := repo.Save(ctx, user)

		fetchedUser, err := repo.GetByID(ctx, savedUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, savedUser.ID, fetchedUser.ID)
		assert.Equal(t, savedUser.Name, fetchedUser.Name)
		assert.Equal(t, savedUser.Email, fetchedUser.Email)
	})

	t.Run("Fails to get users", func(t *testing.T) {
		fetchedUser, err := repo.GetByID(ctx, 9999)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, fetchedUser)
	})
}

func TestGetAllUsers(t *testing.T) {
	repo := NewUserRepository(testDB)
	cleanDB(t, "users")
	ctx := context.Background()

	t.Run("Get all users successfully", func(t *testing.T) {
		user1 := &domain.User{Name: "User1", Email: "user1@example.com"}
		user2 := &domain.User{Name: "User2", Email: "user2@example.com"}
		repo.Save(ctx, user1)
		repo.Save(ctx, user2)

		users, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Contains(t, []string{user1.Name, user2.Name}, users[0].Name)
		assert.Contains(t, []string{user1.Email, user2.Email}, users[0].Email)
		assert.Contains(t, []string{user1.Name, user2.Name}, users[1].Name)
		assert.Contains(t, []string{user1.Email, user2.Email}, users[1].Email)
	})

	t.Run("Return empty array successfully", func(t *testing.T) {
		cleanDB(t, "users") // 데이터 초기화
		users, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 0)
	})
}
