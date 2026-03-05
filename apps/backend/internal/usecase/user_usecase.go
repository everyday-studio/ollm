package usecase

import (
	"context"
	"unicode/utf8"

	"github.com/everyday-studio/ollm/internal/domain"
)

type userUseCase struct {
	userRepo domain.UserRepository
}

func NewUserUseCase(userRepo domain.UserRepository) domain.UserUseCase {
	return &userUseCase{userRepo: userRepo}
}

func (uc *userUseCase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *userUseCase) GetAll(ctx context.Context) ([]domain.User, error) {
	return uc.userRepo.GetAll(ctx)
}

func (uc *userUseCase) UpdateNickname(ctx context.Context, id string, name string) (*domain.User, error) {
	nameLen := utf8.RuneCountInString(name)
	if nameLen < 2 || nameLen > 20 {
		return nil, domain.ErrInvalidInput
	}

	err := uc.userRepo.UpdateNickname(ctx, id, name)
	if err != nil {
		return nil, err
	}

	return uc.userRepo.GetByID(ctx, id)
}
