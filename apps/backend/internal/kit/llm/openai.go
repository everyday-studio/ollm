package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"github.com/everyday-studio/ollm/internal/domain"
)

// openAIService implements the domain.LLMService interface using the OpenAI API.
type openAIService struct {
	client *openai.Client
	model  string // e.g., gpt-4o-mini, gpt-3.5-turbo
}

// NewOpenAIService creates a new OpenAI service instance.
func NewOpenAIService(apiKey string) domain.LLMService {
	return &openAIService{
		client: openai.NewClient(apiKey),
		model:  openai.GPT4oMini,
	}
}

// GenerateResponse calls the OpenAI Chat Completions API with the provided history.
func (s *openAIService) GenerateResponse(ctx context.Context, history []domain.Message) (string, error) {
	openaiMessages := make([]openai.ChatCompletionMessage, 0, len(history))

	for _, msg := range history {
		var role string
		switch msg.Role {
		case domain.MessageRoleSystem:
			role = openai.ChatMessageRoleSystem
		case domain.MessageRoleUser:
			role = openai.ChatMessageRoleUser
		case domain.MessageRoleAssistant:
			role = openai.ChatMessageRoleAssistant
		default:
			role = openai.ChatMessageRoleUser // default
		}

		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:    s.model,
		Messages: openaiMessages,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate response from OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openAI returned an empty response")
	}

	return resp.Choices[0].Message.Content, nil
}
