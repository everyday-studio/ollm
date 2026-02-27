package domain

import "context"

// LLMService defines the interface for communicating with external AI models.
type LLMService interface {
	// GenerateResponse takes the conversation history and returns the AI's generated text.
	GenerateResponse(ctx context.Context, history []Message) (string, error)
}
