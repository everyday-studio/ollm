package domain

import "context"

// LLMService defines the interface for communicating with external AI models.
type LLMService interface {
	// GenerateResponse takes the conversation history and returns the AI's generated text, prompt tokens, and completion tokens.
	GenerateResponse(ctx context.Context, history []Message) (string, int, int, error)

	// EvaluateWinCondition asks the LLM to judge if the user has met the win condition based on the conversation history.
	// It returns true if the condition is met, false otherwise, along with token usage and any error.
	EvaluateWinCondition(ctx context.Context, judgeCondition string, history []Message) (bool, int, int, error)
}
