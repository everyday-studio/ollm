package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
)

func TestGameUseCase_Create(t *testing.T) {
	tests := []struct {
		name       string
		req        *domain.CreateGameRequest
		mockReturn *domain.Game
		mockError  error
		want       *domain.Game
		wantErr    bool
	}{
		{
			name: "Create game successfully",
			req: &domain.CreateGameRequest{
				Title:       "Adventure Quest",
				Description: "A text-based adventure game",
				AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
			},
			mockReturn: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Adventure Quest",
				Description: "A text-based adventure game",
				AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
				Status:      domain.GameStatusActive,
				IsPublic:    true,
			},
			mockError: nil,
			want: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Adventure Quest",
				Description: "A text-based adventure game",
				AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
				Status:      domain.GameStatusActive,
				IsPublic:    true,
			},
			wantErr: false,
		},
		{
			name: "Fail to create game due to repository error",
			req: &domain.CreateGameRequest{
				Title:       "Adventure Quest",
				Description: "A text-based adventure game",
				AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
			},
			mockReturn: nil,
			mockError:  domain.ErrInternal,
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.GameRepository)
			// Use mock.Anything for the game argument because UseCase constructs it internally
			mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Game")).Return(tt.mockReturn, tt.mockError)

			uc := NewGameUseCase(mockRepo)
			ctx := context.Background()
			result, err := uc.Create(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameUseCase_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		inputID    string
		mockReturn *domain.Game
		mockError  error
		want       *domain.Game
		wantErr    error
	}{
		{
			name:    "Get game by ID successfully",
			inputID: "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
			mockReturn: &domain.Game{
				ID:    "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title: "Adventure Quest",
			},
			mockError: nil,
			want: &domain.Game{
				ID:    "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title: "Adventure Quest",
			},
			wantErr: nil,
		},
		{
			name:       "Fail to find game",
			inputID:    "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			want:       nil,
			wantErr:    domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.GameRepository)
			mockRepo.On("GetByID", mock.Anything, tt.inputID).Return(tt.mockReturn, tt.mockError)

			uc := NewGameUseCase(mockRepo)
			ctx := context.Background()
			result, err := uc.GetByID(ctx, tt.inputID)

			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantErr, err)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameUseCase_GetPaginated(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn []domain.Game
		mockCount  int
		mockError  error
		want       *domain.PaginatedData[domain.Game]
		wantErr    error
	}{
		{
			name: "Get paginated games successfully",
			mockReturn: []domain.Game{
				{ID: "01HQZYX3VQJQZ3Z0Z1Z2GAME01", Title: "Game 1"},
				{ID: "01HQZYX3VQJQZ3Z0Z1Z2GAME02", Title: "Game 2"},
			},
			mockCount: 2,
			mockError: nil,
			want: &domain.PaginatedData[domain.Game]{
				Data: []domain.Game{
					{ID: "01HQZYX3VQJQZ3Z0Z1Z2GAME01", Title: "Game 1"},
					{ID: "01HQZYX3VQJQZ3Z0Z1Z2GAME02", Title: "Game 2"},
				},
				Total:      2,
				Page:       1,
				Limit:      10,
				TotalPages: 1,
			},
			wantErr: nil,
		},
		{
			name:       "Fail to get games",
			mockReturn: nil,
			mockCount:  0,
			mockError:  domain.ErrNotFound,
			want:       nil,
			wantErr:    domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.GameRepository)
			mockRepo.On("CountAll", mock.Anything).Return(tt.mockCount, tt.mockError)
			if tt.mockError == nil {
				mockRepo.On("GetPaginated", mock.Anything, 1, 10).Return(tt.mockReturn, nil)
			}

			uc := NewGameUseCase(mockRepo)
			ctx := context.Background()
			result, err := uc.GetPaginated(ctx, 1, 10)

			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantErr, err)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameUseCase_Update(t *testing.T) {
	newTitle := "Updated Title"
	newDescription := "Updated description"
	newIsPublic := true

	tests := []struct {
		name          string
		inputID       string
		req           *domain.UpdateGameRequest
		mockGetReturn *domain.Game
		mockGetError  error
		mockUpdReturn *domain.Game
		mockUpdError  error
		want          *domain.Game
		wantErr       bool
		checkErr      error
	}{
		{
			name:    "Update game successfully",
			inputID: "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
			req: &domain.UpdateGameRequest{
				Title:       &newTitle,
				Description: &newDescription,
				IsPublic:    &newIsPublic,
			},
			mockGetReturn: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Original Title",
				Description: "Original description",
				AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
				Status:      domain.GameStatusActive,
				IsPublic:    false,
			},
			mockGetError: nil,
			mockUpdReturn: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Updated Title",
				Description: "Updated description",
				AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
				Status:      domain.GameStatusActive,
				IsPublic:    true,
			},
			mockUpdError: nil,
			want: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Updated Title",
				Description: "Updated description",
				AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
				Status:      domain.GameStatusActive,
				IsPublic:    true,
			},
			wantErr:  false,
			checkErr: nil,
		},
		{
			name:    "Partial update - title only",
			inputID: "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
			req: &domain.UpdateGameRequest{
				Title: &newTitle,
			},
			mockGetReturn: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Original Title",
				Description: "Original description",
				Status:      domain.GameStatusActive,
			},
			mockGetError: nil,
			mockUpdReturn: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Updated Title",
				Description: "Original description",
				Status:      domain.GameStatusActive,
			},
			mockUpdError: nil,
			want: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Updated Title",
				Description: "Original description",
				Status:      domain.GameStatusActive,
			},
			wantErr:  false,
			checkErr: nil,
		},
		{
			name:          "Fail to update non-existent game",
			inputID:       "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			req:           &domain.UpdateGameRequest{Title: &newTitle},
			mockGetReturn: nil,
			mockGetError:  domain.ErrNotFound,
			want:          nil,
			wantErr:       true,
			checkErr:      domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.GameRepository)
			mockRepo.On("GetByID", mock.Anything, tt.inputID).Return(tt.mockGetReturn, tt.mockGetError)
			if tt.mockGetError == nil {
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Game")).Return(tt.mockUpdReturn, tt.mockUpdError)
			}

			uc := NewGameUseCase(mockRepo)
			ctx := context.Background()
			result, err := uc.Update(ctx, tt.inputID, tt.req)

			assert.Equal(t, tt.want, result)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkErr != nil {
					assert.ErrorIs(t, err, tt.checkErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGameUseCase_Delete(t *testing.T) {
	tests := []struct {
		name      string
		inputID   string
		mockError error
		wantErr   error
	}{
		{
			name:      "Delete game successfully",
			inputID:   "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
			mockError: nil,
			wantErr:   nil,
		},
		{
			name:      "Fail to delete non-existent game",
			inputID:   "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			mockError: domain.ErrNotFound,
			wantErr:   domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.GameRepository)
			mockRepo.On("Delete", mock.Anything, tt.inputID).Return(tt.mockError)

			uc := NewGameUseCase(mockRepo)
			ctx := context.Background()
			err := uc.Delete(ctx, tt.inputID)

			assert.Equal(t, tt.wantErr, err)

			mockRepo.AssertExpectations(t)
		})
	}
}
