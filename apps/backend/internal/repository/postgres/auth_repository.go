package postgres

import (
	"context"
	"database/sql"

	"github.com/everyday-studio/ollm/internal/domain"
)

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) domain.AuthRepository {
	return &authRepository{
		db: db,
	}
}

func (r *authRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	//TODO
	return user, nil
}
