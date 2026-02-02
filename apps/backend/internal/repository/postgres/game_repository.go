package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/everyday-studio/ollm/internal/domain"
)

type gameRepository struct {
	db *sql.DB
}

// NewGameRepository creates a new game repository
func NewGameRepository(db *sql.DB) domain.GameRepository {
	return &gameRepository{
		db: db,
	}
}

// Create inserts a new game into the database
func (r *gameRepository) Create(ctx context.Context, game *domain.Game) (*domain.Game, error) {
	const query = `
		INSERT INTO games (title, description, author_id, status, is_public)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	// Set default values if not provided
	if game.Status == "" {
		game.Status = "active"
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		game.Title,
		game.Description,
		game.AuthorID,
		game.Status,
		game.IsPublic,
	).Scan(&game.ID, &game.CreatedAt, &game.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return game, nil
}

// GetByID retrieves a game by its ID
func (r *gameRepository) GetByID(ctx context.Context, id int64) (*domain.Game, error) {
	const query = `
		SELECT id, title, description, author_id, status, is_public, created_at, updated_at
		FROM games
		WHERE id = $1
	`

	var game domain.Game
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&game.ID,
		&game.Title,
		&game.Description,
		&game.AuthorID,
		&game.Status,
		&game.IsPublic,
		&game.CreatedAt,
		&game.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get game by ID: %w", err)
	}

	return &game, nil
}

// GetAll retrieves all games, ordered by creation date (newest first)
func (r *gameRepository) GetAll(ctx context.Context) ([]domain.Game, error) {
	const query = `
		SELECT id, title, description, author_id, status, is_public, created_at, updated_at
		FROM games
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all games: %w", err)
	}
	defer rows.Close()

	var games []domain.Game
	for rows.Next() {
		var game domain.Game
		if err := rows.Scan(
			&game.ID,
			&game.Title,
			&game.Description,
			&game.AuthorID,
			&game.Status,
			&game.IsPublic,
			&game.CreatedAt,
			&game.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, game)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over games: %w", err)
	}

	return games, nil
}

// Update updates an existing game
func (r *gameRepository) Update(ctx context.Context, game *domain.Game) (*domain.Game, error) {
	const query = `
		UPDATE games
		SET title = $1, description = $2, status = $3, is_public = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		game.Title,
		game.Description,
		game.Status,
		game.IsPublic,
		game.ID,
	).Scan(&game.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return game, nil
}

// Delete removes a game from the database
func (r *gameRepository) Delete(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM games
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
