package llm

import (
	"context"
	"fmt"
	"strings"

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
// Returns: (content string, promptTokens int, completionTokens int, err error)
func (s *openAIService) GenerateResponse(ctx context.Context, history []domain.Message) (string, int, int, error) {
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
		return "", 0, 0, fmt.Errorf("failed to generate response from OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", 0, 0, fmt.Errorf("openAI returned an empty response")
	}

	return resp.Choices[0].Message.Content, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, nil
}

// EvaluateWinCondition asks the LLM to judge if the user has met the win condition based on the conversation history.
func (s *openAIService) EvaluateWinCondition(ctx context.Context, judgeCondition string, history []domain.Message) (bool, int, int, error) {
	judgeMessages := make([]openai.ChatCompletionMessage, 0, len(history)+1)

	systemPrompt := fmt.Sprintf(`You are an impartial game judge. 
Analyze the following text from an AI Assistant.
The game's win condition is: "%s"

Based ONLY on the provided text, has the User successfully achieved this win condition?
Respond with ONLY "true" or "false". Do not include any other text, explanation, or punctuation.`, judgeCondition)

	judgeMessages = append(judgeMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
	})

	for _, h := range history {
		role := openai.ChatMessageRoleUser
		if h.Role == domain.MessageRoleAssistant {
			role = openai.ChatMessageRoleAssistant
		} else if h.Role == domain.MessageRoleSystem {
			role = openai.ChatMessageRoleSystem
		}

		judgeMessages = append(judgeMessages, openai.ChatCompletionMessage{
			Role:    role,
			Content: h.Content,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    judgeMessages,
		Temperature: 0.0,
		MaxTokens:   5,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return false, 0, 0, fmt.Errorf("openai evaluate win condition error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return false, 0, 0, fmt.Errorf("openai evaluate win condition error: no choices returned")
	}

	promptTokens := resp.Usage.PromptTokens
	completionTokens := resp.Usage.CompletionTokens
	content := strings.ToLower(strings.TrimSpace(resp.Choices[0].Message.Content))

	return content == "true", promptTokens, completionTokens, nil
}
