package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
)

func TestMatchUseCase_Create(t *testing.T) {
	tests := []struct {
		name         string
		req          *domain.CreateMatchRequest
		mockGameRet  *domain.Game
		mockGameErr  error
		mockCountRet int
		mockCountErr error
		mockMatchRet *domain.Match
		mockMatchErr error
		want         *domain.Match
		wantErr      bool
	}{
		{
			name: "Create match successfully",
			req: &domain.CreateMatchRequest{
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				GameID: "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
			},
			mockGameRet: &domain.Game{
				ID:       "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
				MaxTurns: 10,
			},
			mockGameErr: nil,
			mockCountRet: 0,
			mockCountErr: nil,
			mockMatchRet: &domain.Match{
				ID:          "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID:      "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				GameID:      "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
				Status:      domain.MatchStatusActive,
				MaxTurns:    10,
				TotalTokens: 0,
				TurnCount:   0,
			},
			mockMatchErr: nil,
			want: &domain.Match{
				ID:          "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID:      "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				GameID:      "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
				Status:      domain.MatchStatusActive,
				MaxTurns:    10,
				TotalTokens: 0,
				TurnCount:   0,
			},
			wantErr: false,
		},
		{
			name: "Fail to create match due to game not found",
			req: &domain.CreateMatchRequest{
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				GameID: "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
			},
			mockGameRet:  nil,
			mockGameErr:  domain.ErrNotFound,
			mockCountRet: 0,
			mockCountErr: nil,
			mockMatchRet: nil,
			mockMatchErr: nil,
			want:         nil,
			wantErr:      true,
		},
		{
			name: "Fail to create match due to repository error",
			req: &domain.CreateMatchRequest{
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				GameID: "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
			},
			mockGameRet: &domain.Game{
				ID:       "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
				MaxTurns: 10,
			},
			mockGameErr:  nil,
			mockCountRet: 0,
			mockCountErr: nil,
			mockMatchRet: nil,
			mockMatchErr: domain.ErrInternal,
			want:         nil,
			wantErr:      true,
		},
		{
			name: "Fail to create match due to too many active matches",
			req: &domain.CreateMatchRequest{
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				GameID: "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
			},
			mockGameRet: &domain.Game{
				ID:       "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
				MaxTurns: 10,
			},
			mockGameErr:  nil,
			mockCountRet: 5,
			mockCountErr: nil,
			mockMatchRet: nil,
			mockMatchErr: nil,
			want:         nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGameRepo := new(mocks.GameRepository)
			mockMatchRepo := new(mocks.MatchRepository)

			mockGameRepo.On("GetByID", mock.Anything, tt.req.GameID).Return(tt.mockGameRet, tt.mockGameErr)

			if tt.mockGameErr == nil {
				mockMatchRepo.On("CountByUserIDGameIDAndStatus", mock.Anything, tt.req.UserID, tt.req.GameID, domain.MatchStatusActive).Return(tt.mockCountRet, tt.mockCountErr)
				if tt.mockCountErr == nil && tt.mockCountRet < 5 {
					mockMatchRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Match")).Return(tt.mockMatchRet, tt.mockMatchErr)
				}
			}

			uc := NewMatchUseCase(mockMatchRepo, mockGameRepo)
			ctx := context.Background()
			result, err := uc.Create(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockGameRepo.AssertExpectations(t)
			if tt.mockGameErr == nil {
				mockMatchRepo.AssertExpectations(t)
			}
		})
	}
}

func TestMatchUseCase_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		matchID    string
		userID     string
		mockReturn *domain.Match
		mockError  error
		want       *domain.Match
		wantErr    error
	}{
		{
			name:    "Get match by ID successfully",
			matchID: "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			userID:  "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			mockReturn: &domain.Match{
				ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				GameID: "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
			},
			mockError: nil,
			want: &domain.Match{
				ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				GameID: "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
			},
			wantErr: nil,
		},
		{
			name:    "Fail to get match due to forbidden access",
			matchID: "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			userID:  "01HQZYX3VQJQZ3Z0Z1Z2ZOTHER2", // Different user
			mockReturn: &domain.Match{
				ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			},
			mockError: nil,
			want:      nil,
			wantErr:   domain.ErrForbidden,
		},
		{
			name:       "Fail to get non-existent match",
			matchID:    "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			userID:     "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			want:       nil,
			wantErr:    domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMatchRepo := new(mocks.MatchRepository)
			mockGameRepo := new(mocks.GameRepository)

			mockMatchRepo.On("GetByID", mock.Anything, tt.matchID).Return(tt.mockReturn, tt.mockError)

			uc := NewMatchUseCase(mockMatchRepo, mockGameRepo)
			ctx := context.Background()
			result, err := uc.GetByID(ctx, tt.matchID, tt.userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, result)

			mockMatchRepo.AssertExpectations(t)
		})
	}
}

func TestMatchUseCase_Resign(t *testing.T) {
	tests := []struct {
		name         string
		matchID      string
		userID       string
		mockGetRet   *domain.Match
		mockGetErr   error
		mockUpdRet   *domain.Match
		mockUpdErr   error
		wantErr      bool
		checkErrType error
	}{
		{
			name:    "Resign match successfully",
			matchID: "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			userID:  "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			mockGetRet: &domain.Match{
				ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				Status: domain.MatchStatusActive,
			},
			mockGetErr: nil,
			mockUpdRet: &domain.Match{
				ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				Status: domain.MatchStatusResigned,
			},
			mockUpdErr:   nil,
			wantErr:      false,
			checkErrType: nil,
		},
		{
			name:         "Fail to resign non-existent match",
			matchID:      "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			userID:       "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			mockGetRet:   nil,
			mockGetErr:   domain.ErrNotFound,
			wantErr:      true,
			checkErrType: domain.ErrNotFound,
		},
		{
			name:    "Fail to resign due to forbidden access (not owner)",
			matchID: "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			userID:  "01HQZYX3VQJQZ3Z0Z1ZOTHER2", // Different user
			mockGetRet: &domain.Match{
				ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				Status: domain.MatchStatusActive,
			},
			mockGetErr:   nil,
			wantErr:      true,
			checkErrType: domain.ErrForbidden,
		},
		{
			name:    "Fail to resign match that is already won",
			matchID: "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			userID:  "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			mockGetRet: &domain.Match{
				ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				Status: domain.MatchStatusWon,
			},
			mockGetErr:   nil,
			wantErr:      true,
			checkErrType: domain.ErrConflict,
		},
		{
			name:    "Fail to resign due to update error",
			matchID: "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			userID:  "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			mockGetRet: &domain.Match{
				ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				Status: domain.MatchStatusActive,
			},
			mockGetErr:   nil,
			mockUpdRet:   nil,
			mockUpdErr:   domain.ErrInternal,
			wantErr:      true,
			checkErrType: domain.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMatchRepo := new(mocks.MatchRepository)
			mockGameRepo := new(mocks.GameRepository)

			mockMatchRepo.On("GetByID", mock.Anything, tt.matchID).Return(tt.mockGetRet, tt.mockGetErr)

			// If it fetches successfully, belongs to the user, and is active, Update is called
			if tt.mockGetErr == nil && tt.mockGetRet != nil && tt.mockGetRet.UserID == tt.userID && tt.mockGetRet.Status == domain.MatchStatusActive {
				mockMatchRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Match")).Return(tt.mockUpdRet, tt.mockUpdErr)
			}

			uc := NewMatchUseCase(mockMatchRepo, mockGameRepo)
			ctx := context.Background()
			err := uc.Resign(ctx, tt.matchID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkErrType != nil {
					assert.ErrorIs(t, err, tt.checkErrType)
				}
			} else {
				assert.NoError(t, err)
			}

			mockMatchRepo.AssertExpectations(t)
		})
	}
}

func TestMatchUseCase_Delete(t *testing.T) {
	tests := []struct {
		name      string
		matchID   string
		mockError error
		wantErr   error
	}{
		{
			name:      "Delete match successfully",
			matchID:   "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			mockError: nil,
			wantErr:   nil,
		},
		{
			name:      "Fail to delete non-existent match",
			matchID:   "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			mockError: domain.ErrNotFound,
			wantErr:   domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMatchRepo := new(mocks.MatchRepository)
			mockGameRepo := new(mocks.GameRepository)

			mockMatchRepo.On("Delete", mock.Anything, tt.matchID).Return(tt.mockError)

			uc := NewMatchUseCase(mockMatchRepo, mockGameRepo)
			ctx := context.Background()
			err := uc.Delete(ctx, tt.matchID)

			assert.Equal(t, tt.wantErr, err)

			mockMatchRepo.AssertExpectations(t)
		})
	}
}
