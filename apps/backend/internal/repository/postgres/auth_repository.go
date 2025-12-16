package postgres

import (
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
