package usecase

import (
	"context"
	"crypto/rsa"
	"errors"
	"time"

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

func NewAuthUseCase(authRepo domain.AuthRepository, userRepo domain.UserRepository, config *config.Config) (domain.AuthUsecase, error) {
	privateKey, err := security.ParseRSAPrivateKeyFromBase64(config.Secure.JWT.PrivateKey)
	if err != nil {
		return nil, err
	}

	publicKey, err := security.ParseRSAPublicKeyFromBase64(config.Secure.JWT.PublicKey)
	if err != nil {
		return nil, err
	}

	return &authUseCase{
		authRepo:   authRepo,
		userRepo:   userRepo,
		config:     config,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
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

func (uc *authUseCase) Logout(ctx context.Context, userID int64) error {
	return nil
}

func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.LoginResponse, error) {
	return nil, nil
}

func (uc *authUseCase) generateTokens(user *domain.User) (*domain.LoginResponse, error) {
	accessTokenExpiration := time.Duration(uc.config.Secure.JWT.AccessExpirationMin) * time.Minute
	accessToken, err := security.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role,
		uc.privateKey,
		accessTokenExpiration,
	)
	if err != nil {
		return nil, err
	}

	refreshTokenExpiration := time.Duration(uc.config.Secure.JWT.RefreshExpirationDay) * 24 * time.Hour
	refreshToken, err := security.GenerateRefreshToken(
		user.ID,
		user.Email,
		user.Role,
		uc.privateKey,
		refreshTokenExpiration,
	)
	if err != nil {
		return nil, err
	}

	response := &domain.LoginResponse{
		ID:                     user.ID,
		Email:                  user.Email,
		AccessToken:            accessToken,
		RefreshToken:           refreshToken,
		RefreshTokenExpiration: time.Now().Add(refreshTokenExpiration),
	}
	return response, nil
}
