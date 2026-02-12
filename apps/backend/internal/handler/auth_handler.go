package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/everyday-studio/ollm/internal/config"
	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/kit/security"
	"github.com/everyday-studio/ollm/internal/middleware"
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

	group := e.Group("/auth", middleware.AllowRoles(domain.RolePublic))
	group.POST("/signup", handler.SignUpUser)
	group.POST("/login", handler.Login)

	return handler
}

func (h *AuthHandler) SignUpUser(c echo.Context) error {
	req := new(domain.SignUpRequest)
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	user := &domain.User{
		Name:     "TESTUSER", //TODOs
		Email:    req.Email,
		Password: req.Password,
		Role:     domain.RoleUser,
	}

	ctx := c.Request().Context()
	createdUser, err := h.authUseCase.SignUpUser(ctx, user)
	if err == nil {
		return c.JSON(http.StatusCreated, domain.SignUpResponse{
			ID:    createdUser.ID,
			Name:  createdUser.Name,
			Email: createdUser.Email,
		})
	}

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(err))
	case errors.Is(err, domain.ErrAlreadyExists):
		return c.JSON(http.StatusConflict, ErrResponse(err))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

func (h *AuthHandler) Login(c echo.Context) error {
	req := new(domain.LoginRequest)
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	loginResponse, err := h.authUseCase.Login(ctx, req.Email, req.Password)
	if err == nil {
		cookie := h.createRefreshTokenCookie(
			loginResponse.RefreshToken,
			loginResponse.RefreshTokenExpiration,
		)
		c.SetCookie(cookie)

		return c.JSON(http.StatusOK, loginResponse)
	}

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusUnauthorized, ErrResponse(errors.New("invalid email or password")))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

func (h *AuthHandler) createRefreshTokenCookie(tokenValue string, expiration time.Time) *http.Cookie {
	cookieConfig := h.config.Secure.JWT.Cookie

	sameSite := http.SameSiteLaxMode
	switch strings.ToLower(cookieConfig.SameSite) {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}

	cookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    tokenValue,
		Path:     "/",
		Domain:   cookieConfig.Domain,
		Expires:  expiration,
		Secure:   cookieConfig.Secure,
		HttpOnly: cookieConfig.HTTPOnly,
		SameSite: sameSite,
	}

	return cookie
}

// CreateLogoutCookie creates a cookie that expires the refresh token
func (h *AuthHandler) CreateLogoutCookie() *http.Cookie {
	// 빈 값과 과거 만료일로 쿠키를 생성하여 쿠키를 삭제하는 효과를 냅니다
	return h.createRefreshTokenCookie("", time.Now().Add(-1*time.Hour))
}

func (h *AuthHandler) Logout(c echo.Context) error {
	// Get user ID from context (set by JWT middleware)
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}

	ctx := c.Request().Context()
	// 로그아웃 처리 - 현재는 간단하게 구현
	err := h.authUseCase.Logout(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}

	// 리프레시 토큰 쿠키 만료시키기
	c.SetCookie(h.CreateLogoutCookie())

	// 클라이언트에 응답 - 프론트엔드에서 액세스 토큰을 삭제해야 함을 알림
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully logged out. Please remove the access token from your client storage.",
		"status":  "success",
	})
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	// Get refresh token from cookie
	cookie, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrResponse(errors.New("refresh token not found")))
	}

	ctx := c.Request().Context()
	// Call usecase to refresh token
	loginResponse, err := h.authUseCase.RefreshToken(ctx, cookie.Value)
	if err != nil {
		switch {
		case errors.Is(err, security.ErrInvalidToken), errors.Is(err, security.ErrExpiredToken):
			return c.JSON(http.StatusUnauthorized, ErrResponse(errors.New("invalid or expired refresh token")))
		case errors.Is(err, domain.ErrUnauthorized):
			return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
		default:
			return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
		}
	}

	// Set new refresh token as HTTP-only cookie
	newCookie := h.createRefreshTokenCookie(loginResponse.RefreshToken, loginResponse.RefreshTokenExpiration)
	c.SetCookie(newCookie)

	// Return new access token
	return c.JSON(http.StatusOK, loginResponse)
}
