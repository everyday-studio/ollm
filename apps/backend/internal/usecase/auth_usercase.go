package usecase

import (
	"context"
	"crypto/rsa"
	"errors"

	"github.com/everyday-studio/ollm/internal/config"
	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/kit/security"
)

type authUseCase struct {
	authRepo   domain.AuthRepository
	userRepo   domain.UserRepository
	config     *config.Config
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func (uc *authUseCase) SignUpUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	hashedPassword, err := security.GeneratePasswordHash(user.Password, nil)
	if err != nil {
		return nil, err
	}
	user.Password = hashedPassword

	return uc.userRepo.Save(ctx, user)
}

func (uc *authUseCase) Login(ctx context.Context, email string, password string) (*domain.LoginResponse, error) {
	user, err := uc.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	match, err := security.ComparePasswordHash(password, user.Password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, errors.New("invalid credentials")
	}

	return uc.generateTokens(user)
}

func (uc *authUseCase) generateTokens(user *domain.User) (*domain.LoginResponse, error) {
	response := &domain.LoginResponse{}
	return response, nil
}
