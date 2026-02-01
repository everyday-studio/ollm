package usecase

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/everyday-studio/ollm/internal/config"
	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
	"github.com/everyday-studio/ollm/internal/kit/security"
)

func generateTestRSAKeys() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyPEM)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyPEM)

	return privateKeyBase64, publicKeyBase64, nil
}

func TestAuthUsecase_SignUpUser(t *testing.T) {
	privateKeyPEM, publicKeyPEM, err := generateTestRSAKeys()
	if err != nil {
		t.Fatalf("Failed to generate test RSA keys: %v", err)
	}

	cfg := &config.Config{
		Secure: config.SecureConfig{
			JWT: config.JWTConfig{
				PrivateKey:           privateKeyPEM,
				PublicKey:            publicKeyPEM,
				AccessExpirationMin:  60,
				RefreshExpirationDay: 30,
				Cookie: config.CookieConfig{
					Secure:   false,
					HTTPOnly: true,
					SameSite: "Lax",
					Domain:   "localhost",
				},
			},
		},
	}

	tests := []struct {
		name       string
		input      *domain.User
		mockReturn *domain.User
		mockError  error
		want       *domain.User
		wantErr    error
	}{
		{
			name: "Success",
			input: &domain.User{
				Email:    "test@example.com",
				Password: "password123",
				Role:     domain.RoleUser,
			},
			mockReturn: &domain.User{
				ID:       1,
				Email:    "test@example.com",
				Password: "hashedpassword",
				Role:     domain.RoleUser,
			},
			mockError: nil,
			want: &domain.User{
				ID:       1,
				Email:    "test@example.com",
				Password: "hashedpassword",
				Role:     domain.RoleUser,
			},
			wantErr: nil,
		},
		{
			name: "Already Exists",
			input: &domain.User{
				Email:    "test@example.com",
				Password: "password123",
				Role:     domain.RoleUser,
			},
			mockReturn: nil,
			mockError:  domain.ErrAlreadyExists,
			want:       nil,
			wantErr:    domain.ErrAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := new(mocks.AuthRepository)
			mockUserRepo := new(mocks.UserRepository)

			mockUserRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.User")).Return(tt.mockReturn, tt.mockError)

			uc, err := NewAuthUseCase(mockAuthRepo, mockUserRepo, cfg)
			if err != nil {
				t.Fatalf("Failed to create auth usecase: %v", err)
			}

			ctx := context.Background()
			result, err := uc.SignUpUser(ctx, tt.input)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestLogin(t *testing.T) {
	privateKeyPEM, publickeyPEM, err := generateTestRSAKeys()
	if err != nil {
		t.Fatalf("Failed to generate test RSA keys: %v", err)
	}

	cfg := &config.Config{
		Secure: config.SecureConfig{
			JWT: config.JWTConfig{
				PrivateKey:           privateKeyPEM,
				PublicKey:            publickeyPEM,
				AccessExpirationMin:  15,
				RefreshExpirationDay: 7,
				Cookie: config.CookieConfig{
					Secure:   false,
					HTTPOnly: true,
					SameSite: "Lax",
					Domain:   "localhost",
				},
			},
		},
	}

	hashedPassword, err := security.GeneratePasswordHash("password", nil)
	if err != nil {
		t.Fatalf("Failed to generate hashed password: %v", err)
	}

	user := &domain.User{
		ID:       1,
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     domain.RoleUser,
	}

	tests := []struct {
		name       string
		email      string
		password   string
		mockReturn *domain.User
		mockError  error
		want       *domain.LoginResponse
		wantErr    error
	}{
		{
			name:       "Success",
			email:      "test@example.com",
			password:   "password",
			mockReturn: user,
			mockError:  nil,
			want: &domain.LoginResponse{
				ID:    1,
				Email: "test@example.com",
			},
			wantErr: nil,
		},
		{
			name:       "User Not Found",
			email:      "nonexistent@example.com",
			password:   "password",
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			want:       nil,
			wantErr:    domain.ErrNotFound,
		},
		{
			name:       "Invalid Password",
			email:      "test@example.com",
			password:   "wrongpassword",
			mockReturn: user,
			mockError:  nil,
			want:       nil,
			wantErr:    domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := new(mocks.AuthRepository)
			mockUserRepo := new(mocks.UserRepository)

			if tt.email != "" {
				mockUserRepo.On("GetUserByEmail", mock.Anything, tt.email).Return(tt.mockReturn, tt.mockError)
			}

			uc, err := NewAuthUseCase(mockAuthRepo, mockUserRepo, cfg)
			if err != nil {
				t.Fatalf("Failed to create auth usecase: %v", err)
			}

			ctx := context.Background()
			result, err := uc.Login(ctx, tt.email, tt.password)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}
