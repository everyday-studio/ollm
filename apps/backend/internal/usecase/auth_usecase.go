package usecase

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/idtoken"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"

	"github.com/everyday-studio/ollm/internal/config"
	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/kit/nickname"
	"github.com/everyday-studio/ollm/internal/kit/security"
	"github.com/everyday-studio/ollm/internal/kit/tag"
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
	if user.Name == "" || user.Name == "TESTUSER" {
		generatedName, err := nickname.Generate()
		if err != nil {
			user.Name = nickname.GenerateFallback()
		} else {
			user.Name = generatedName
		}
	}

	hashedPassword, err := security.GeneratePasswordHash(user.Password, nil)
	if err != nil {
		return nil, err
	}
	user.Password = hashedPassword

	for i := 0; i < 5; i++ {
		generatedTag, err := tag.Generate()
		if err != nil {
			return nil, err
		}
		user.Tag = generatedTag

		savedUser, err := uc.userRepo.Save(ctx, user)
		if err == nil {
			return savedUser, nil
		}

		if errors.Is(err, domain.ErrConflict) && strings.Contains(err.Error(), "duplicate tag") {
			continue // Retry on tag collision
		}
		return nil, err
	}

	return nil, domain.ErrConflict // Max retries exceeded
}

func (uc *authUseCase) Login(ctx context.Context, email string, password string) (*domain.LoginResponse, error) {
	user, err := uc.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	match, err := security.ComparePasswordHash(password, user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to compare password hash: %w", err)
	}
	if !match {
		return nil, domain.ErrUnauthorized
	}

	return uc.generateTokens(user)
}

func (uc *authUseCase) Logout(ctx context.Context, userID string) error {
	return nil
}

// LoginWithGoogle verifies the Google ID token provided by the frontend,
// then upserts the user in the database and issues JWT tokens.
func (uc *authUseCase) LoginWithGoogle(ctx context.Context, idToken string) (*domain.LoginResponse, error) {
	// Validate the Google ID token using Google's public key infrastructure.
	payload, err := idtoken.Validate(ctx, idToken, uc.config.GCP.GoogleClientID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Extract required claims from the verified payload.
	sub, _ := payload.Claims["sub"].(string) // Unique Google user ID
	email, _ := payload.Claims["email"].(string)
	googleName, _ := payload.Claims["name"].(string)

	if sub == "" || email == "" {
		return nil, domain.ErrUnauthorized
	}

	// Generate a fallback display name if Google name is empty.
	if googleName == "" {
		googleName, err = nickname.Generate()
		if err != nil {
			googleName = nickname.GenerateFallback()
		}
	}

	user := &domain.User{
		Name:     googleName,
		Email:    email,
		GoogleID: sub,
		Role:     domain.RoleUser,
	}

	// Upsert: create if new, return existing if already registered.
	savedUser, err := uc.userRepo.UpsertGoogleUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert google user: %w", err)
	}

	return uc.generateTokens(savedUser)
}

func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.LoginResponse, error) {
	claims, err := security.ValidateRefreshToken(refreshToken, uc.publicKey)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	response, err := uc.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (uc *authUseCase) GuestLogin(ctx context.Context) (*domain.LoginResponse, error) {
	// Generate a unique identifier for the guest
	guestID := ulid.Make().String()
	email := fmt.Sprintf("guest_%s@ollm.xyz", guestID)

	// Generate a random dummy password and hash it
	dummyPassword := uuid.New().String()
	hashedPassword, err := security.GeneratePasswordHash(dummyPassword, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate dummy password hash: %w", err)
	}

	// Generate random nickname
	name, err := nickname.Generate()
	if err != nil {
		name = nickname.GenerateFallback()
	}

	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     domain.RoleUser,
	}

	// Retries for tag generation logic (similar to SignUpUser)
	var savedUser *domain.User
	for i := 0; i < 5; i++ {
		generatedTag, err := tag.Generate()
		if err != nil {
			return nil, err
		}
		user.Tag = generatedTag

		savedUser, err = uc.userRepo.Save(ctx, user)
		if err == nil {
			break
		}

		if errors.Is(err, domain.ErrConflict) && strings.Contains(err.Error(), "duplicate tag") {
			continue
		}
		return nil, err
	}

	if savedUser == nil {
		return nil, domain.ErrConflict
	}

	return uc.generateTokens(savedUser)
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
		Name:                   user.Name,
		Tag:                    user.Tag,
		Email:                  user.Email,
		AccessToken:            accessToken,
		RefreshToken:           refreshToken,
		RefreshTokenExpiration: time.Now().Add(refreshTokenExpiration),
	}
	return response, nil
}
