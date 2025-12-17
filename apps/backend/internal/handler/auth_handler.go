package handler

import (
	"github.com/everyday-studio/ollm/internal/config"
	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/labstack/echo/v4"
)

const refreshTokenCookieName = "refresh_token"

type AuthHandler struct {
	authUseCase domain.AuthUsecase
	config      *config.Config
}

func NewAuthHandler(e *echo.Echo, authUseCase domain.AuthUsecase, config *config.Config) *AuthHandler {
	handler := &AuthHandler{
		authUseCase: authUseCase,
		config:      config,
	}

	return handler
}
