package domain

import (
	"context"
	"time"
)

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

// Message represents a single conversation turn within a Match
type Message struct {
	ID        string      `json:"id"`
	MatchID   string      `json:"match_id"`
	Role      MessageRole `json:"role"`
	Content   string      `json:"content"`
	IsVisible bool        `json:"is_visible"`
	CreatedAt time.Time   `json:"created_at"`
}

// CreateMessageRequest is the DTO for creating a new message
type CreateMessageRequest struct {
	Content string `json:"content"`
}

// MessageRepository defines the interface for message data access
type MessageRepository interface {
	Create(ctx context.Context, message *Message) (*Message, error)
	GetByID(ctx context.Context, id string) (*Message, error)
	GetByMatchID(ctx context.Context, matchID string) ([]Message, error)
	Delete(ctx context.Context, id string) error
}

// MessageUseCase defines the interface for message business logic
type MessageUseCase interface {
	Create(ctx context.Context, req *CreateMessageRequest) (*Message, error)
	GetByID(ctx context.Context, id string) (*Message, error)
	GetByMatchID(ctx context.Context, matchID string) ([]Message, error)
	Delete(ctx context.Context, id string) error
}
