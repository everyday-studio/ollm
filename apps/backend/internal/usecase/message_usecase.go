package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/everyday-studio/ollm/internal/domain"
)

type messageUseCase struct {
	messageRepo domain.MessageRepository
	matchRepo   domain.MatchRepository
	llmService  domain.LLMService
	gameRepo    domain.GameRepository
}

func NewMessageUseCase(
	messageRepo domain.MessageRepository,
	matchRepo domain.MatchRepository,
	llmService domain.LLMService,
	gameRepo domain.GameRepository,
) domain.MessageUseCase {
	return &messageUseCase{
		messageRepo: messageRepo,
		matchRepo:   matchRepo,
		llmService:  llmService,
		gameRepo:    gameRepo,
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
	// 5. 최종 매치 상태 갱신 (Finalize)
	// ==========================================
	match.TotalTokens += promptTokens
	nextStatus := domain.MatchStatusActive

	// 승리 판정 (TODO: LLM 통해서 JUDGE)
	if game.TargetWord != "" && strings.Contains(strings.ToLower(aiContent), strings.ToLower(game.TargetWord)) {
		nextStatus = domain.MatchStatusWon
	}

	// 패배 판정
	if nextStatus == domain.MatchStatusActive && match.TurnCount >= match.MaxTurns {
		nextStatus = domain.MatchStatusLost
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
