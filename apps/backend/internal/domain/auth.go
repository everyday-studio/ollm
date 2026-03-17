package domain

import (
	"context"
	"time"
)

type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Tag   string `json:"tag"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	Tag                    string    `json:"tag"`
	Email                  string    `json:"email"`
	AccessToken            string    `json:"access_token"`
	RefreshToken           string    `json:"-"` // Not included in JSON response
	RefreshTokenExpiration time.Time `json:"-"` // Not included in JSON response
}

// GoogleLoginRequest is the request body for the Google social login endpoint.
// The frontend must obtain the ID token from Google Sign-In and send it here.
type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
}

type AuthRepository interface {
}

type AuthUsecase interface {
	SignUpUser(ctx context.Context, user *User) (*User, error)
	Login(ctx context.Context, email string, password string) (*LoginResponse, error)
	Logout(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	// LoginWithGoogle verifies a Google ID token and upserts the user,
	// returning a LoginResponse with JWT tokens on success.
	LoginWithGoogle(ctx context.Context, idToken string) (*LoginResponse, error)
	// GuestLogin creates a temporary guest user and returns JWT tokens.
	GuestLogin(ctx context.Context) (*LoginResponse, error)
}
