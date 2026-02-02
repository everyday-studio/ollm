package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
)

func TestUserUsecase_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		inputID    string
		mockReturn *domain.User
		mockError  error
		want       *domain.User
		wantErr    error
	}{
		{
			name:       "Find user successfully",
			inputID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
			mockReturn: &domain.User{ID: "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5", Name: "John", Email: "john@example.com"},
			mockError:  nil,
			want:       &domain.User{ID: "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5", Name: "John", Email: "john@example.com"},
			wantErr:    nil,
		},
		{
			name:      "Fail to find user",
			inputID:   "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z6",
			mockError: domain.ErrNotFound,
			want:      nil,
			wantErr:   domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.UserRepository)
			mockRepo.On("GetByID", mock.Anything, tt.inputID).Return(tt.mockReturn, tt.mockError)

			uc := NewUserUseCase(mockRepo)
			ctx := context.Background()
			result, err := uc.GetByID(ctx, tt.inputID)

			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantErr, err)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserUsecase_GetAll(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn []domain.User
		mockError  error
		want       []domain.User
		wantErr    error
	}{
		{
			name: "Find user successfully",
			mockReturn: []domain.User{
				{ID: "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5", Name: "John", Email: "john@example.com"},
				{ID: "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z6", Name: "Jane", Email: "jane@example.com"},
			},
			mockError: nil,
			want: []domain.User{
				{ID: "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5", Name: "John", Email: "john@example.com"},
				{ID: "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z6", Name: "Jane", Email: "jane@example.com"},
			},
			wantErr: nil,
		},
		{
			name:      "Fail to find any users",
			mockError: domain.ErrNotFound,
			want:      nil,
			wantErr:   domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.UserRepository)
			mockRepo.On("GetAll", mock.Anything).Return(tt.mockReturn, tt.mockError)

			uc := NewUserUseCase(mockRepo)
			ctx := context.Background()
			result, err := uc.GetAll(ctx)

			assert.Equal(t, tt.want, result)
			assert.Equal(t, tt.wantErr, err)

			mockRepo.AssertExpectations(t)
		})
	}
}
