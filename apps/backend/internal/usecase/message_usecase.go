package usecase

import (
	"context"
	"fmt"

	"github.com/everyday-studio/ollm/internal/domain"
)

type messageUseCase struct {
	messageRepo domain.MessageRepository
	matchRepo   domain.MatchRepository
	llmService  domain.LLMService
	gameRepo    domain.GameRepository
}

// NewMessageUseCase creates a new message use case
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

// Create handles the core game turn: User Input -> Save -> Fetch History -> LLM -> Save AI Output
func (uc *messageUseCase) Create(ctx context.Context, matchID string, userID string, req *domain.CreateMessageRequest) (*domain.Message, error) {
	// 1. 권한 검증: 이 매치(게임)가 존재하는지, 그리고 현재 요청한 유저의 게임이 맞는지 확인
	match, err := uc.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match for authorization: %w", err)
	}
	if match.UserID != userID {
		return nil, domain.ErrForbidden // 내 게임이 아니면 접근 차단!
	}

	// 2. 유저의 메시지를 DB에 먼저 저장 (Role: User 강제 고정)
	userMsg := &domain.Message{
		MatchID:   matchID,
		Role:      domain.MessageRoleUser,
		Content:   req.Content,
		IsVisible: true,
	}
	if _, err := uc.messageRepo.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// 3. AI에게 문맥을 전달하기 위해, 방금 저장한 메시지를 포함한 전체 대화 내역(History)을 가져옵니다.
	// (Repository에서 ORDER BY created_at ASC 로 정렬해두었기 때문에 완벽합니다)
	history, err := uc.messageRepo.GetByMatchID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match history: %w", err)
	}

	// 4. 게임의 시스템 프롬프트를 조회하여 대화 내역 맨 앞에 덧붙입니다.
	game, err := uc.gameRepo.GetByID(ctx, match.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game for system prompt: %w", err)
	}

	fullHistory := []domain.Message{
		{
			Role:    domain.MessageRoleSystem,
			Content: game.SystemPrompt,
		},
	}
	fullHistory = append(fullHistory, history...)

	// 5. 외부 LLM(OpenAI) 호출 (동기식 대기)
	aiContent, err := uc.llmService.GenerateResponse(ctx, fullHistory)
	if err != nil {
		return nil, fmt.Errorf("llm failed to generate response: %w", err)
	}

	// TODO: 게임의 TargetWord를 DB에서 조회하여, AI의 응답(aiContent)에 해당 단어가 포함되어 있다면 매치를 승리(MatchStatusWon) 상태로 변경하는 로직 추가

	// 6. 무사히 도착한 AI의 답변을 DB에 저장
	aiMsg := &domain.Message{
		MatchID:   matchID,
		Role:      domain.MessageRoleAssistant,
		Content:   aiContent,
		IsVisible: true,
	}
	savedAIMsg, err := uc.messageRepo.Create(ctx, aiMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to save ai message: %w", err)
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
