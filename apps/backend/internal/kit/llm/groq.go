package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/everyday-studio/ollm/internal/domain"
)

// groqService implements the domain.LLMService interface using the Groq API.
type groqService struct {
	client *openai.Client
	model  string
}

// NewGroqService creates a new Groq service instance.
func NewGroqService(apiKey string) domain.LLMService {
	// Groq provides an OpenAI-compatible API endpoint
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"

	return &groqService{
		client: openai.NewClientWithConfig(config),
		model:  "llama-3.3-70b-versatile",
	}
}

// GenerateResponse is currently not used for Groq in this application, but implemented to satisfy the interface.
func (s *groqService) GenerateResponse(ctx context.Context, history []domain.Message) (string, int, int, error) {
	return "", 0, 0, fmt.Errorf("GenerateResponse is not supported by groqService in this context")
}

// EvaluateWinCondition asks the LLM to judge if the user has met the win condition.
func (s *groqService) EvaluateWinCondition(ctx context.Context, judgeCondition string, history []domain.Message) (bool, int, int, error) {
	// 1. 심판의 페르소나 (System)
	systemPrompt := `You are an impartial and strict game judge. 
	Evaluate if the AI Assistant's text meets the Win Condition.
		Rules:
			1. Think step-by-step and write your reasoning inside <reason> tags.
			2. After reasoning, output your final decision as EXACTLY <result>true</result> or <result>false</result>.`

	// 2. 평가할 텍스트를 하나의 문자열로 합치기
	var targetTextBuilder strings.Builder
	for _, h := range history {
		// 누가 한 말인지 텍스트 안에 명시해줍니다.
		roleName := "AI Assistant"
		if h.Role == domain.MessageRoleUser {
			roleName = "User"
		}
		targetTextBuilder.WriteString(fmt.Sprintf("[%s]: %s\n", roleName, h.Content))
	}

	// 3. User가 심판에게 질문하는 형태로 구성 (핵심!)
	userPrompt := fmt.Sprintf("Win Condition: \"%s\"\n\nText to evaluate:\n%s", judgeCondition, targetTextBuilder.String())

	judgeMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    judgeMessages,
		Temperature: 0.0,
		MaxTokens:   300,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return false, 0, 0, fmt.Errorf("groq evaluate win condition error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return false, 0, 0, fmt.Errorf("groq evaluate win condition error: no choices returned")
	}

	promptTokens := resp.Usage.PromptTokens
	completionTokens := resp.Usage.CompletionTokens

	content := strings.ToLower(strings.TrimSpace(resp.Choices[0].Message.Content))

	//log
	//fmt.Printf("\n[Groq Judge Raw Output]: '%s'\n", content)

	// "true"가 포함되어 있고 "false"가 없다면 승리! ("True.", "Yes, true" 등 모두 커버)
	isWon := strings.Contains(content, "true") && !strings.Contains(content, "false")

	return isWon, promptTokens, completionTokens, nil
}

// EvaluateFormatBreak asks the LLM to judge if the AI has failed to follow the format rules.
func (s *groqService) EvaluateFormatBreak(ctx context.Context, condition string, aiContent string) (bool, error) {
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
		return false, fmt.Errorf("groq format break evaluation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return false, fmt.Errorf("groq format break evaluation failed: no response from assistant")
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
