package llm

import (
	"context"
	"encoding/json"
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

// EvaluateFormatBreak asks the LLM to judge if the AI has failed to follow the format rules.
func (s *openAIService) EvaluateFormatBreak(ctx context.Context, condition string, aiContent string) (bool, error) {
	systemPrompt := `You are a strict format and syntax validator.
	Your task is to judge if the provided 'AI Content' violates the 'Requirement' (e.g. valid JSON, Python code, specific header, etc).
	Even a 1% deviation or minor syntax error means the format is broken.

	Rules:
	1. If the content deviates from the requirement in ANY way, is_broken must be true.
	2. If it strictly follows the requirement, is_broken must be false.
	3. You MUST respond in JSON format only.`

	userPrompt := fmt.Sprintf("Requirement: %s\n\nAI Content to Evaluate:\n%s", condition, aiContent)

	req := openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.0,
		MaxTokens:   300,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return false, fmt.Errorf("openai format break evaluation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return false, fmt.Errorf("openai format break evaluation failed: no response from assistant")
	}

	var result struct {
		IsBroken bool   `json:"is_broken"`
		Reason   string `json:"reason"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return false, fmt.Errorf("failed to parse judge response: %w (content: %s)", err, resp.Choices[0].Message.Content)
	}

	return result.IsBroken, nil
}

// EvaluatePromptAdvice asks the LLM to analyze the user's prompt and provide helpful advice based on the AI's actual response.
func (s *openAIService) EvaluatePromptAdvice(ctx context.Context, gameRule string, userContent string, aiContent string) (string, error) {

	// 🌟 4o 모델의 높은 지능을 믿고 논리적으로 짠 깔끔한 프롬프트
	systemPrompt := fmt.Sprintf(`### ROLE & TONE
You are reviewing a player's failed attempt to hack you. 
You MUST adopt the EXACT persona, tone, and speaking style of the "Target AI" defined in the <target_rule> below. 
Do not sound like an AI assistant. Speak completely IN CHARACTER.

### TARGET RULE
%s

### OUTPUT INSTRUCTIONS
Write 2-3 sentences in natural, character-driven KOREAN. 
Your response MUST seamlessly combine these two elements:
1. [Evaluation]: As the character, mock or analyze why the User's prompt failed, referencing how you successfully defended yourself in the 'AI Response'.
2. [Hint]: Arrogantly leak a hint about how they could actually bypass your logic. Suggest a prompt engineering technique (e.g., roleplaying, using delimiters, changing the context) but explain it entirely through the lens of your character's world/metaphors.`, gameRule)

	userPrompt := fmt.Sprintf("User Content: %s\nAI Response: %s", userContent, aiContent)

	req := openai.ChatCompletionRequest{
		Model: "gpt-4o", // 💡 압도적인 지능의 4o 모델로 변경!
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.8,
		MaxTokens:   200,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("openai prompt advice evaluation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai prompt advice evaluation failed: no response")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
