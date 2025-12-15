package usecase

import (
	"context"
	"crypto/rsa"

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
