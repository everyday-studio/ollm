package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/everyday-studio/ollm/internal/domain"
)

type messageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB) domain.MessageRepository {
	return &messageRepository{
		db: db,
	}
}

// Create inserts a new message into the database
func (r *messageRepository) Create(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	// Generate secure ULID for the new message
	message.ID = ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.Reader, 0)).String()

	const query = `
        INSERT INTO messages (id, match_id, role, content, is_visible)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING created_at
    `

	err := r.db.QueryRowContext(
		ctx,
		query,
		message.ID,
		message.MatchID,
		message.Role,
		message.Content,
		message.IsVisible,
	).Scan(&message.CreatedAt)

	if err != nil {
		return nil, mapDBError(err)
	}

	return message, nil
}

// GetByID retrieves a single message by its ID
func (r *messageRepository) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	const query = `
        SELECT id, match_id, role, content, is_visible, created_at
        FROM messages
        WHERE id = $1
    `

	var msg domain.Message
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID,
		&msg.MatchID,
		&msg.Role,
		&msg.Content,
		&msg.IsVisible,
		&msg.CreatedAt,
	)

	if err != nil {
		return nil, mapDBError(err)
	}

	return &msg, nil
}

// GetByMatchID retrieves all messages for a specific match
func (r *messageRepository) GetByMatchID(ctx context.Context, matchID string) ([]domain.Message, error) {
	const query = `
        SELECT id, match_id, role, content, is_visible, created_at
        FROM messages
        WHERE match_id = $1
        ORDER BY created_at ASC
    `

	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	messages := []domain.Message{}

	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.MatchID,
			&msg.Role,
			&msg.Content,
			&msg.IsVisible,
			&msg.CreatedAt,
		); err != nil {
			return nil, mapDBError(err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(err)
	}

	return messages, nil
}

// Delete removes a message from the database
func (r *messageRepository) Delete(ctx context.Context, id string) error {
	const query = `
        DELETE FROM messages
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
