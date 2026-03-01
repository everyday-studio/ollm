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
		INSERT INTO matches (id, user_id, game_id, status, max_turns, total_tokens, turn_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		match.ID,
		match.UserID,
		match.GameID,
		match.Status,
		match.MaxTurns,
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
		SELECT id, user_id, game_id, status, max_turns, total_tokens, turn_count, created_at, updated_at
		FROM matches
		WHERE id = $1
	`

	var match domain.Match
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID,
		&match.UserID,
		&match.GameID,
		&match.Status,
		&match.MaxTurns,
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
		SELECT id, user_id, game_id, status, max_turns, total_tokens, turn_count, created_at, updated_at
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
			&match.MaxTurns,
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
		SELECT id, user_id, game_id, status, max_turns, total_tokens, turn_count, created_at, updated_at
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
			&match.MaxTurns,
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

// Update updates an existing match
func (r *matchRepository) Update(ctx context.Context, match *domain.Match) (*domain.Match, error) {
	const query = `
		UPDATE matches
		SET status = $1, max_turns = $2, total_tokens = $3, turn_count = $4
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		match.Status,
		match.MaxTurns,
		match.TotalTokens,
		match.TurnCount,
		match.ID,
	).Scan(&match.UpdatedAt)

	if err != nil {
		return nil, mapDBError(err)
	}

	return match, nil
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

// GetLeaderboard retrieves the top scores for a specific game
func (r *matchRepository) GetLeaderboard(ctx context.Context, gameID string, limit int) ([]domain.LeaderboardEntry, error) {
	const query = `
		WITH RankedMatches AS (
			SELECT 
				m.user_id,
				m.turn_count,
				m.total_tokens,
				m.updated_at,
				ROW_NUMBER() OVER(
					PARTITION BY m.user_id 
					ORDER BY m.turn_count ASC, m.total_tokens ASC, m.updated_at ASC
				) as rn
			FROM matches m
			WHERE m.game_id = $1 AND m.status = 'won'
		)
		SELECT 
			r.user_id,
			u.name as username,
			r.turn_count,
			r.total_tokens,
			r.updated_at
		FROM RankedMatches r
		JOIN users u ON r.user_id = u.id
		WHERE r.rn = 1
		ORDER BY r.turn_count ASC, r.total_tokens ASC, r.updated_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, gameID, limit)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	leaderboard := make([]domain.LeaderboardEntry, 0, limit)
	for rows.Next() {
		var entry domain.LeaderboardEntry
		if err := rows.Scan(
			&entry.UserID,
			&entry.Username,
			&entry.TurnCount,
			&entry.TotalTokens,
			&entry.AchievedAt,
		); err != nil {
			return nil, mapDBError(err)
		}
		leaderboard = append(leaderboard, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(err)
	}

	return leaderboard, nil
}
