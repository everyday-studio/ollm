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

func (uc *userUseCase) CountAll(ctx context.Context) (int, error) {
	return uc.userRepo.CountAll(ctx)
}

func (uc *userUseCase) GetPaginated(ctx context.Context, page, limit int) (*domain.PaginatedData[domain.User], error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	total, err := uc.userRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}

	users, err := uc.userRepo.GetPaginated(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit

	return &domain.PaginatedData[domain.User]{
		Data:       users,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
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
