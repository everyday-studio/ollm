package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/everyday-studio/ollm/internal/domain"
)

type matchRepository struct {
	db *sql.DB
}

// NewMatchRepository creates a new match repository
func NewMatchRepository(db *sql.DB) domain.MatchRepository {
	return &matchRepository{
		db: db,
	}
}

// Create inserts a new match into the database
func (r *matchRepository) Create(ctx context.Context, match *domain.Match) (*domain.Match, error) {
	// Generate ULID for the new match
	match.ID = ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.Reader, 0)).String()

	const query = `
		INSERT INTO matches (id, user_id, game_id, status, total_tokens, turn_count)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		match.ID,
		match.UserID,
		match.GameID,
		match.Status,
		match.TotalTokens,
		match.TurnCount,
	).Scan(&match.CreatedAt, &match.UpdatedAt)

	if err != nil {
		return nil, mapDBError(err)
	}

	return match, nil
}

// GetByID retrieves a match by its ID
func (r *matchRepository) GetByID(ctx context.Context, id string) (*domain.Match, error) {
	const query = `
		SELECT id, user_id, game_id, status, total_tokens, turn_count, created_at, updated_at
		FROM matches
		WHERE id = $1
	`

	var match domain.Match
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID,
		&match.UserID,
		&match.GameID,
		&match.Status,
		&match.TotalTokens,
		&match.TurnCount,
		&match.CreatedAt,
		&match.UpdatedAt,
	)

	if err != nil {
		return nil, mapDBError(err)
	}

	return &match, nil
}

// GetByUserID retrieves all matches for a specific user, ordered by creation date (newest first)
func (r *matchRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Match, error) {
	const query = `
		SELECT id, user_id, game_id, status, total_tokens, turn_count, created_at, updated_at
		FROM matches
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	matches := []domain.Match{}
	for rows.Next() {
		var match domain.Match
		if err := rows.Scan(
			&match.ID,
			&match.UserID,
			&match.GameID,
			&match.Status,
			&match.TotalTokens,
			&match.TurnCount,
			&match.CreatedAt,
			&match.UpdatedAt,
		); err != nil {
			return nil, mapDBError(err)
		}
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(err)
	}

	return matches, nil
}

// GetByUserIDAndGameID retrieves all matches for a specific user and game, ordered by creation date (newest first)
func (r *matchRepository) GetByUserIDAndGameID(ctx context.Context, userID string, gameID string) ([]domain.Match, error) {
	const query = `
		SELECT id, user_id, game_id, status, total_tokens, turn_count, created_at, updated_at
		FROM matches
		WHERE user_id = $1 AND game_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, gameID)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	matches := []domain.Match{}
	for rows.Next() {
		var match domain.Match
		if err := rows.Scan(
			&match.ID,
			&match.UserID,
			&match.GameID,
			&match.Status,
			&match.TotalTokens,
			&match.TurnCount,
			&match.CreatedAt,
			&match.UpdatedAt,
		); err != nil {
			return nil, mapDBError(err)
		}
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(err)
	}

	return matches, nil
}

// Delete removes a match from the database
func (r *matchRepository) Delete(ctx context.Context, id string) error {
	const query = `
		DELETE FROM matches
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
