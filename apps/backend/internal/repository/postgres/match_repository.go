package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
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
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	match.ID = ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()

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
		return nil, fmt.Errorf("failed to create match: %w", err)
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
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get match by ID: %w", err)
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
		return nil, fmt.Errorf("failed to get matches by user ID: %w", err)
	}
	defer rows.Close()

	var matches []domain.Match
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
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over matches: %w", err)
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
		return fmt.Errorf("failed to delete match: %w", err)
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
