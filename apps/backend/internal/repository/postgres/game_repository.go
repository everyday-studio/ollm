package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"strconv"
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
		INSERT INTO games (id, title, description, author_id, status, is_public, system_prompt, first_message, judge_type, judge_condition, max_turns)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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
		game.FirstMessage,
		game.JudgeType,
		game.JudgeCondition,
		game.MaxTurns,
	).Scan(&game.CreatedAt, &game.UpdatedAt)

	if err != nil {
		return nil, mapDBError(err)
	}

	return game, nil
}

// GetByID retrieves a game by its ID
func (r *gameRepository) GetByID(ctx context.Context, id string) (*domain.Game, error) {
	const query = `
		SELECT id, title, description, author_id, status, is_public, system_prompt, first_message, judge_type, judge_condition, max_turns, created_at, updated_at
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
		&game.FirstMessage,
		&game.JudgeType,
		&game.JudgeCondition,
		&game.MaxTurns,
		&game.CreatedAt,
		&game.UpdatedAt,
	)

	if err != nil {
		return nil, mapDBError(err)
	}

	return &game, nil
}

// CountAll returns the total number of games with optional filtering
func (r *gameRepository) CountAll(ctx context.Context, filter *domain.GameFilter) (int, error) {
	query := `SELECT COUNT(*) FROM games`
	args := []interface{}{}

	if filter != nil && filter.IsPublic != nil {
		query += ` WHERE is_public = $1`
		args = append(args, *filter.IsPublic)
	}

	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, mapDBError(err)
	}
	return count, nil
}

// GetPaginated retrieves a paginated list of games with optional filtering, ordered by creation date (newest first)
func (r *gameRepository) GetPaginated(ctx context.Context, page, limit int, filter *domain.GameFilter) ([]domain.Game, error) {
	offset := (page - 1) * limit
	query := `
		SELECT id, title, description, author_id, status, is_public, system_prompt, first_message, judge_type, judge_condition, max_turns, created_at, updated_at
		FROM games
	`
	args := []interface{}{}
	argIdx := 1

	if filter != nil && filter.IsPublic != nil {
		query += ` WHERE is_public = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.IsPublic)
		argIdx++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
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
			&game.FirstMessage,
			&game.JudgeType,
			&game.JudgeCondition,
			&game.MaxTurns,
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
		SET title = $1, description = $2, status = $3, is_public = $4, system_prompt = $5, first_message = $6, judge_type = $7, judge_condition = $8, max_turns = $9
		WHERE id = $10
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
		game.FirstMessage,
		game.JudgeType,
		game.JudgeCondition,
		game.MaxTurns,
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
