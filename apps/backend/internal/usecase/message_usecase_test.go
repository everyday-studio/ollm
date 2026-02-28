package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
)

func TestMessageUseCase_Create(t *testing.T) {
	tests := []struct {
		name                 string
		matchID              string
		userID               string
		req                  *domain.CreateMessageRequest
		mockMatchGet         *domain.Match
		mockMatchGetErr      error
		mockUserMsgCreateRet *domain.Message
		mockUserMsgCreateErr error
		mockMsgGetHistoryRet []domain.Message
		mockMsgGetHistoryErr error
		mockGameGet          *domain.Game
		mockGameGetErr       error
		mockLLMResp          string
		mockLLMPromptTok     int
		mockLLMCompTok       int
		mockLLMErr           error
		mockUserMsgUpdErr    error // Can be ignored if it fails, but mock setup needs it if we mock Update
		mockAIMsgCreateRet   *domain.Message
		mockAIMsgCreateErr   error
		mockMatchUpdRet      *domain.Match
		mockMatchUpdErr      error
		want                 *domain.Message
		wantErr              bool
		validateMatchStatus  domain.MatchStatus
	}{
		{
			name:    "Create message and AI response successfully",
			matchID: "01HQZYX3VQJQZ3Z0ZMATCH1",
			userID:  "01HQZYX3VQJQZ3Z0ZUSER1",
			req: &domain.CreateMessageRequest{
				Content: "Hello AI",
			},
			mockMatchGet: &domain.Match{
				ID:        "01HQZYX3VQJQZ3Z0ZMATCH1",
				UserID:    "01HQZYX3VQJQZ3Z0ZUSER1",
				GameID:    "01HQZYX3VQJQZ3Z0ZGAME1",
				Status:    domain.MatchStatusActive,
				TurnCount: 0,
				MaxTurns:  5,
			},
			mockUserMsgCreateRet: &domain.Message{
				ID:        "01HQZYX3VQJQZ3Z0ZMSGUSR1",
				MatchID:   "01HQZYX3VQJQZ3Z0ZMATCH1",
				Role:      domain.MessageRoleUser,
				Content:   "Hello AI",
				TurnCount: 1,
			},
			mockMsgGetHistoryRet: []domain.Message{
				{Role: domain.MessageRoleUser, Content: "Hello AI"},
			},
			mockGameGet: &domain.Game{
				ID:           "01HQZYX3VQJQZ3Z0ZGAME1",
				SystemPrompt: "You are a friendly AI",
				TargetWord:   "apple",
			},
			mockLLMResp:      "Hello! I am a friendly AI",
			mockLLMPromptTok: 10,
			mockLLMCompTok:   15,
			mockAIMsgCreateRet: &domain.Message{
				ID:         "01HQZYX3VQJQZ3Z0ZMSGAI1",
				MatchID:    "01HQZYX3VQJQZ3Z0ZMATCH1",
				Role:       domain.MessageRoleAssistant,
				Content:    "Hello! I am a friendly AI",
				TurnCount:  1,
				TokenCount: 15,
			},
			mockMatchUpdRet: &domain.Match{Status: domain.MatchStatusActive},
			want: &domain.Message{
				ID:         "01HQZYX3VQJQZ3Z0ZMSGAI1",
				MatchID:    "01HQZYX3VQJQZ3Z0ZMATCH1",
				Role:       domain.MessageRoleAssistant,
				Content:    "Hello! I am a friendly AI",
				TurnCount:  1,
				TokenCount: 15,
			},
			wantErr:             false,
			validateMatchStatus: domain.MatchStatusActive,
		},
		{
			name:    "Target word matched resulting in won status",
			matchID: "01HQZYX3VQJQZ3Z0ZMATCH1",
			userID:  "01HQZYX3VQJQZ3Z0ZUSER1",
			req:     &domain.CreateMessageRequest{Content: "What is the fruit?"},
			mockMatchGet: &domain.Match{
				ID:        "01HQZYX3VQJQZ3Z0ZMATCH1",
				UserID:    "01HQZYX3VQJQZ3Z0ZUSER1",
				GameID:    "01HQZYX3VQJQZ3Z0ZGAME1",
				Status:    domain.MatchStatusActive,
				TurnCount: 0,
				MaxTurns:  5,
			},
			mockUserMsgCreateRet: &domain.Message{Role: domain.MessageRoleUser},
			mockMsgGetHistoryRet: []domain.Message{},
			mockGameGet: &domain.Game{
				ID:         "01HQZYX3VQJQZ3Z0ZGAME1",
				TargetWord: "apple",
			},
			mockLLMResp:         "It is an apple.",
			mockLLMPromptTok:    5,
			mockLLMCompTok:      5,
			mockAIMsgCreateRet:  &domain.Message{Role: domain.MessageRoleAssistant},
			mockMatchUpdRet:     &domain.Match{},
			want:                &domain.Message{Role: domain.MessageRoleAssistant},
			wantErr:             false,
			validateMatchStatus: domain.MatchStatusWon,
		},
		{
			name:    "Max turns reached resulting in lost status",
			matchID: "01HQZYX3VQJQZ3Z0ZMATCH1",
			userID:  "01HQZYX3VQJQZ3Z0ZUSER1",
			req:     &domain.CreateMessageRequest{Content: "Last try"},
			mockMatchGet: &domain.Match{
				ID:        "01HQZYX3VQJQZ3Z0ZMATCH1",
				UserID:    "01HQZYX3VQJQZ3Z0ZUSER1",
				GameID:    "01HQZYX3VQJQZ3Z0ZGAME1",
				Status:    domain.MatchStatusActive,
				TurnCount: 4, // Next will be 5
				MaxTurns:  5,
			},
			mockUserMsgCreateRet: &domain.Message{Role: domain.MessageRoleUser},
			mockMsgGetHistoryRet: []domain.Message{},
			mockGameGet: &domain.Game{
				TargetWord: "apple", // Won't be mentioned
			},
			mockLLMResp:         "Sorry, you didn't guess it.",
			mockLLMPromptTok:    5,
			mockLLMCompTok:      5,
			mockAIMsgCreateRet:  &domain.Message{Role: domain.MessageRoleAssistant},
			mockMatchUpdRet:     &domain.Match{},
			want:                &domain.Message{Role: domain.MessageRoleAssistant},
			wantErr:             false,
			validateMatchStatus: domain.MatchStatusLost,
		},
		{
			name:    "Forbidden access to match",
			matchID: "01HQZYX3VQJQZ3Z0ZMATCH1",
			userID:  "01HQZYX3VQJQZ3Z0ZUSER_OTHER",
			req:     &domain.CreateMessageRequest{Content: "Hello"},
			mockMatchGet: &domain.Match{
				UserID: "01HQZYX3VQJQZ3Z0ZUSER1",
			},
			wantErr: true,
		},
		{
			name:    "Conflict because match relies on active status",
			matchID: "01HQZYX3VQJQZ3Z0ZMATCH1",
			userID:  "01HQZYX3VQJQZ3Z0ZUSER1",
			req:     &domain.CreateMessageRequest{Content: "Hello"},
			mockMatchGet: &domain.Match{
				UserID: "01HQZYX3VQJQZ3Z0ZUSER1",
				Status: domain.MatchStatusWon,
			},
			wantErr: true,
		},
		{
			name:    "Final update fails resulting in Error status fallback",
			matchID: "01HQZYX3VQJQZ3Z0ZMATCH1",
			userID:  "01HQZYX3VQJQZ3Z0ZUSER1",
			req:     &domain.CreateMessageRequest{Content: "Hello"},
			mockMatchGet: &domain.Match{
				ID:        "01HQZYX3VQJQZ3Z0ZMATCH1",
				UserID:    "01HQZYX3VQJQZ3Z0ZUSER1",
				GameID:    "01HQZYX3VQJQZ3Z0ZGAME1",
				Status:    domain.MatchStatusActive,
				TurnCount: 0,
				MaxTurns:  5,
			},
			mockUserMsgCreateRet: &domain.Message{Role: domain.MessageRoleUser},
			mockMsgGetHistoryRet: []domain.Message{},
			mockGameGet: &domain.Game{
				ID: "01HQZYX3VQJQZ3Z0ZGAME1",
			},
			mockLLMResp:        "Hello back",
			mockAIMsgCreateRet: &domain.Message{Role: domain.MessageRoleAssistant},
			mockMatchUpdErr:    domain.ErrInternal,
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMsgRepo := new(mocks.MessageRepository)
			mockMatchRepo := new(mocks.MatchRepository)
			mockLLMService := new(mocks.LLMService)
			mockGameRepo := new(mocks.GameRepository)

			// Setup mocks logically
			mockMatchRepo.On("GetByID", mock.Anything, tt.matchID).Return(tt.mockMatchGet, tt.mockMatchGetErr)

			if tt.mockMatchGetErr == nil && tt.mockMatchGet != nil && tt.mockMatchGet.UserID == tt.userID && tt.mockMatchGet.Status == domain.MatchStatusActive {
				mockMatchRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *domain.Match) bool {
					return m.Status == domain.MatchStatusGenerating
				})).Return(&domain.Match{Status: domain.MatchStatusGenerating}, nil).Once()

				mockMsgRepo.On("Create", mock.Anything, mock.MatchedBy(func(m *domain.Message) bool {
					return m.Role == domain.MessageRoleUser
				})).Return(tt.mockUserMsgCreateRet, tt.mockUserMsgCreateErr)

				if tt.mockUserMsgCreateErr == nil {
					mockMsgRepo.On("GetByMatchID", mock.Anything, tt.matchID).Return(tt.mockMsgGetHistoryRet, tt.mockMsgGetHistoryErr)

					if tt.mockMsgGetHistoryErr == nil {
						mockGameRepo.On("GetByID", mock.Anything, tt.mockMatchGet.GameID).Return(tt.mockGameGet, tt.mockGameGetErr)

						if tt.mockGameGetErr == nil {
							mockLLMService.On("GenerateResponse", mock.Anything, mock.Anything).Return(tt.mockLLMResp, tt.mockLLMPromptTok, tt.mockLLMCompTok, tt.mockLLMErr)

							if tt.mockLLMErr == nil {
								mockMsgRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *domain.Message) bool {
									return m.Role == domain.MessageRoleUser
								})).Return(&domain.Message{}, nil) // This fails safely in UC, so return empty

								mockMsgRepo.On("Create", mock.Anything, mock.MatchedBy(func(m *domain.Message) bool {
									return m.Role == domain.MessageRoleAssistant
								})).Return(tt.mockAIMsgCreateRet, tt.mockAIMsgCreateErr)

								if tt.mockAIMsgCreateErr == nil {
									if tt.mockMatchUpdErr == nil {
										mockMatchRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *domain.Match) bool {
											assert.Equal(t, tt.validateMatchStatus, m.Status)
											return true
										})).Return(tt.mockMatchUpdRet, nil).Once()
									} else {
										// Final DB update fails
										mockMatchRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *domain.Match) bool {
											// First it tries to save the next status (e.g. Active)
											return m.Status != domain.MatchStatusError && m.Status != domain.MatchStatusGenerating
										})).Return(nil, tt.mockMatchUpdErr).Once()

										// Then it forces the status to Error to prevent zombie matches
										mockMatchRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *domain.Match) bool {
											return m.Status == domain.MatchStatusError
										})).Return(&domain.Match{Status: domain.MatchStatusError}, nil).Once()
									}
								}
							} else {
								// When LLM fails, MatchStatusError update runs
								mockMatchRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *domain.Match) bool {
									return m.Status == domain.MatchStatusError
								})).Return(tt.mockMatchUpdRet, nil).Maybe()
							}
						}
					}
				}
			}

			uc := NewMessageUseCase(mockMsgRepo, mockMatchRepo, mockLLMService, mockGameRepo)
			ctx := context.Background()
			result, err := uc.Create(ctx, tt.matchID, tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestMessageUseCase_GetByID(t *testing.T) {
	mockMsgRepo := new(mocks.MessageRepository)
	mockMsgRepo.On("GetByID", mock.Anything, "MSG1").Return(&domain.Message{ID: "MSG1"}, nil)

	uc := NewMessageUseCase(mockMsgRepo, nil, nil, nil)
	result, err := uc.GetByID(context.Background(), "MSG1")

	assert.NoError(t, err)
	assert.Equal(t, "MSG1", result.ID)
}

func TestMessageUseCase_GetByMatchID(t *testing.T) {
	tests := []struct {
		name         string
		matchID      string
		userID       string
		mockMatchRet *domain.Match
		mockMatchErr error
		mockMsgRet   []domain.Message
		wantErr      error
	}{
		{
			name:    "Success",
			matchID: "MATCH1",
			userID:  "USER1",
			mockMatchRet: &domain.Match{
				ID:     "MATCH1",
				UserID: "USER1",
			},
			mockMsgRet: []domain.Message{{ID: "MSG1"}},
			wantErr:    nil,
		},
		{
			name:    "Forbidden",
			matchID: "MATCH1",
			userID:  "USER2",
			mockMatchRet: &domain.Match{
				ID:     "MATCH1",
				UserID: "USER1",
			},
			wantErr: domain.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMsgRepo := new(mocks.MessageRepository)
			mockMatchRepo := new(mocks.MatchRepository)

			mockMatchRepo.On("GetByID", mock.Anything, tt.matchID).Return(tt.mockMatchRet, tt.mockMatchErr)

			if tt.wantErr == nil {
				mockMsgRepo.On("GetByMatchID", mock.Anything, tt.matchID).Return(tt.mockMsgRet, nil)
			}

			uc := NewMessageUseCase(mockMsgRepo, mockMatchRepo, nil, nil)
			result, err := uc.GetByMatchID(context.Background(), tt.matchID, tt.userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, 1)
			}
		})
	}
}

func TestMessageUseCase_Delete(t *testing.T) {
	mockMsgRepo := new(mocks.MessageRepository)
	mockMsgRepo.On("Delete", mock.Anything, "MSG1").Return(nil)

	uc := NewMessageUseCase(mockMsgRepo, nil, nil, nil)
	err := uc.Delete(context.Background(), "MSG1")

	assert.NoError(t, err)
}
