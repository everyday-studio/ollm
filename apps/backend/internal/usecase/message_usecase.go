package usecase

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/everyday-studio/ollm/internal/domain"
)

type messageUseCase struct {
	messageRepo     domain.MessageRepository
	matchRepo       domain.MatchRepository
	llmService      domain.LLMService
	judgeLLMService domain.LLMService
	gameRepo        domain.GameRepository
}

func NewMessageUseCase(
	messageRepo domain.MessageRepository,
	matchRepo domain.MatchRepository,
	llmService domain.LLMService,
	judgeLLMService domain.LLMService,
	gameRepo domain.GameRepository,
) domain.MessageUseCase {
	return &messageUseCase{
		messageRepo:     messageRepo,
		matchRepo:       matchRepo,
		llmService:      llmService,
		judgeLLMService: judgeLLMService,
		gameRepo:        gameRepo,
	}
}

// Create handles the core game turn
func (uc *messageUseCase) Create(ctx context.Context, matchID string, userID string, req *domain.CreateMessageRequest) (*domain.Message, error) {
	// ==========================================
	// 1. 검증 및 상태 락 (Validation & Lock)
	// ==========================================
	match, err := uc.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match for authorization: %w", err)
	}
	if match.UserID != userID {
		return nil, domain.ErrForbidden
	}
	if match.Status != domain.MatchStatusActive {
		return nil, domain.ErrConflict
	}

	match.Status = domain.MatchStatusGenerating
	if _, err := uc.matchRepo.Update(ctx, match); err != nil {
		return nil, fmt.Errorf("failed to lock match state: %w", err)
	}

	// ==========================================
	// 2. 롤백 안전장치 (Safety Net)
	// ==========================================
	userMessageSaved := false

	defer func() {
		if match.Status == domain.MatchStatusGenerating {
			if userMessageSaved {
				match.Status = domain.MatchStatusError
			} else {
				match.Status = domain.MatchStatusActive
			}
			_, _ = uc.matchRepo.Update(context.WithoutCancel(ctx), match)
		}
	}()

	// ==========================================
	// 3. 유저 데이터 저장 및 대화 컨텍스트 구성
	// ==========================================
	currentTurn := match.TurnCount + 1

	userMsg := &domain.Message{
		MatchID:    matchID,
		Role:       domain.MessageRoleUser,
		Content:    req.Content,
		IsVisible:  true,
		TurnCount:  currentTurn,
		TokenCount: 0,
	}

	if _, err := uc.messageRepo.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	userMessageSaved = true
	match.TurnCount = currentTurn

	// 대화 내역 및 게임 시스템 프롬프트 조회
	history, err := uc.messageRepo.GetByMatchID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match history: %w", err)
	}

	game, err := uc.gameRepo.GetByID(ctx, match.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game for system prompt: %w", err)
	}

	fullHistory := make([]domain.Message, 0, len(history)+1)
	fullHistory = append(fullHistory, domain.Message{
		Role:    domain.MessageRoleSystem,
		Content: game.SystemPrompt,
	})
	fullHistory = append(fullHistory, history...)

	// ==========================================
	// 4. 외부 LLM 연동 및 결과 처리
	// ==========================================
	aiContent, promptTokens, completionTokens, err := uc.llmService.GenerateResponse(ctx, fullHistory)
	if err != nil {
		match.Status = domain.MatchStatusError
		if _, updateErr := uc.matchRepo.Update(context.WithoutCancel(ctx), match); updateErr != nil {
			return nil, fmt.Errorf("llm failed and status update also failed: %v (original: %w)", updateErr, err)
		}
		return nil, fmt.Errorf("llm failed to generate response: %w", err)
	}

	// 유저 메시지 토큰 업데이트
	userMsg.TokenCount = promptTokens
	if _, err := uc.messageRepo.Update(ctx, userMsg); err != nil {
		// 에러를 무시하고 진행 (핵심 로직이 아니므로)
	}

	// AI 메시지 저장
	aiMsg := &domain.Message{
		MatchID:    matchID,
		Role:       domain.MessageRoleAssistant,
		Content:    aiContent,
		IsVisible:  true,
		TurnCount:  currentTurn,
		TokenCount: completionTokens,
	}
	savedAIMsg, err := uc.messageRepo.Create(ctx, aiMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to save ai message: %w", err)
	}

	// ==========================================
	// 5. 비동기 판정 및 훈수 처리 (Concurrent Evaluation)
	// ==========================================
	match.TotalTokens += promptTokens
	
	nextStatus := domain.MatchStatusActive
	var promptAdvice string

	// Use errgroup for concurrent execution
	// Create a new context for goroutines to avoid early cancellation if the parent request is already ending
	// The primary request doesn't wait strictly on full LLM context cancellation usually, but we pass ctx here
	// wait, it's fine to pass ctx, but if the client disconnects, context is cancelled. We should use a background context or a timeout context.
	// Actually, just pass ctx and let it cancel if user disconnects.
	// We'll use a channel-based approach or sync.WaitGroup to avoid returning early
	// But errgroup handles this nicely.
	eg, egCtx := errgroup.WithContext(ctx)

	// 5-1. 판정 고루틴 (Judge)
	eg.Go(func() error {
		status := domain.MatchStatusActive

		if game.JudgeType == domain.JudgeTypeTargetWord && game.JudgeCondition != "" {
			if strings.Contains(strings.ToLower(aiContent), strings.ToLower(game.JudgeCondition)) {
				status = domain.MatchStatusWon
			}
		} else if game.JudgeType == domain.JudgeTypeLLMJudge && game.JudgeCondition != "" {
			evaluationHistory := []domain.Message{*aiMsg}
			isWon, _, _, evalErr := uc.judgeLLMService.EvaluateWinCondition(egCtx, game.JudgeCondition, evaluationHistory)
			if evalErr != nil {
				fmt.Printf("failed to evaluate win condition: %v\n", evalErr)
			} else if isWon {
				status = domain.MatchStatusWon
			}
		} else if game.JudgeType == domain.JudgeTypeFormatBreak && game.JudgeCondition != "" {
			isBroken, evalErr := uc.judgeLLMService.EvaluateFormatBreak(egCtx, game.JudgeCondition, aiContent)
			if evalErr != nil {
				fmt.Printf("failed to evaluate format break condition: %v\n", evalErr)
			} else if isBroken {
				status = domain.MatchStatusWon
			}
		}

		// 패배 판정
		if status == domain.MatchStatusActive && match.TurnCount >= match.MaxTurns {
			status = domain.MatchStatusLost
		}
		nextStatus = status
		return nil
	})

	// 5-2. 훈수 고루틴 (Prompt Advice)
	eg.Go(func() error {
		advice, evalErr := uc.judgeLLMService.EvaluatePromptAdvice(egCtx, game.JudgeCondition, req.Content)
		if evalErr != nil {
			fmt.Printf("failed to evaluate prompt advice: %v\n", evalErr)
		} else {
			promptAdvice = advice
		}
		return nil
	})

	// 5-3. 대기 및 최종 반영
	if err := eg.Wait(); err != nil {
		fmt.Printf("concurrent evaluation error: %v\n", err)
	}

	// 훈수 업데이트
	if promptAdvice != "" {
		userMsg.PromptAdvice = &promptAdvice
		if _, updateErr := uc.messageRepo.Update(ctx, userMsg); updateErr != nil {
			fmt.Printf("failed to update user message with advice: %v\n", updateErr)
		}
	}

	match.Status = nextStatus
	if _, err := uc.matchRepo.Update(ctx, match); err != nil {
		match.Status = domain.MatchStatusError
		_, _ = uc.matchRepo.Update(context.WithoutCancel(ctx), match)
		return nil, fmt.Errorf("failed to update final match status: %w", err)
	}

	return savedAIMsg, nil
}

func (uc *messageUseCase) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	return uc.messageRepo.GetByID(ctx, id)
}

func (uc *messageUseCase) GetByMatchID(ctx context.Context, matchID string, userID string) ([]domain.Message, error) {
	match, err := uc.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify match ownership: %w", err)
	}
	if match.UserID != userID {
		return nil, domain.ErrForbidden
	}

	return uc.messageRepo.GetByMatchID(ctx, matchID)
}

func (uc *messageUseCase) Delete(ctx context.Context, id string) error {
	return uc.messageRepo.Delete(ctx, id)
}
