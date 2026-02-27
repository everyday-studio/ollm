package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

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
	// Generate ULID for the new game
	game.ID = ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.Reader, 0)).String()

	const query = `
		INSERT INTO games (id, title, description, author_id, status, is_public, system_prompt, target_word)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		game.ID,
		game.Title,
		game.Description,
		game.AuthorID,
		game.Status,
		game.IsPublic,
		game.SystemPrompt,
		game.TargetWord,
	).Scan(&game.CreatedAt, &game.UpdatedAt)

	if err != nil {
		return nil, mapDBError(err)
	}

	return game, nil
}

// GetByID retrieves a game by its ID
func (r *gameRepository) GetByID(ctx context.Context, id string) (*domain.Game, error) {
	const query = `
		SELECT id, title, description, author_id, status, is_public, system_prompt, target_word, created_at, updated_at
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
		&game.SystemPrompt,
		&game.TargetWord,
		&game.CreatedAt,
		&game.UpdatedAt,
	)

	if err != nil {
		return nil, mapDBError(err)
	}

	return &game, nil
}

// GetAll retrieves all games, ordered by creation date (newest first)
func (r *gameRepository) GetAll(ctx context.Context) ([]domain.Game, error) {
	const query = `
		SELECT id, title, description, author_id, status, is_public, system_prompt, target_word, created_at, updated_at
		FROM games
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	games := []domain.Game{}
	for rows.Next() {
		var game domain.Game
		if err := rows.Scan(
			&game.ID,
			&game.Title,
			&game.Description,
			&game.AuthorID,
			&game.Status,
			&game.IsPublic,
			&game.SystemPrompt,
			&game.TargetWord,
			&game.CreatedAt,
			&game.UpdatedAt,
		); err != nil {
			return nil, mapDBError(err)
		}
		games = append(games, game)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(err)
	}

	return games, nil
}

// Update updates an existing game
// Note: updated_at is automatically updated by database trigger
func (r *gameRepository) Update(ctx context.Context, game *domain.Game) (*domain.Game, error) {
	const query = `
		UPDATE games
		SET title = $1, description = $2, status = $3, is_public = $4, system_prompt = $5, target_word = $6
		WHERE id = $7
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		game.Title,
		game.Description,
		game.Status,
		game.IsPublic,
		game.SystemPrompt,
		game.TargetWord,
		game.ID,
	).Scan(&game.UpdatedAt)

	if err != nil {
		return nil, mapDBError(err)
	}

	return game, nil
}

// Delete removes a game from the database
func (r *gameRepository) Delete(ctx context.Context, id string) error {
	const query = `
		DELETE FROM games
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return mapDBError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return mapDBError(err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
