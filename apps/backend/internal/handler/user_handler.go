package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/everyday-studio/ollm/internal/config"
	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/middleware"
)

type UserHandler struct {
	userUseCase domain.UserUseCase
	config      *config.Config
}

func NewUserHandler(e *echo.Echo, userUseCase domain.UserUseCase, cfg *config.Config) *UserHandler {
	handler := &UserHandler{
		userUseCase: userUseCase,
		config:      cfg,
	}

	userGroup := e.Group("/api/users", middleware.AllowRoles(domain.RoleUser))
	userGroup.GET("/me", handler.GetMe)
	userGroup.PUT("/me", handler.UpdateMe)
	userGroup.DELETE("/me", handler.Withdraw)

	return handler
}

func (h *UserHandler) GetMe(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}

	ctx := c.Request().Context()
	user, err := h.userUseCase.GetByID(ctx, userID)
	if err == nil {
		return c.JSON(http.StatusOK, user)
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

func (h *UserHandler) UpdateMe(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}

	req := new(domain.UpdateNicknameRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	user, err := h.userUseCase.UpdateNickname(ctx, userID, req.Name)
	if err == nil {
		return c.JSON(http.StatusOK, user)
	}

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(err))
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

func (h *UserHandler) Withdraw(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}

	ctx := c.Request().Context()
	err := h.userUseCase.Delete(ctx, userID)
	if err == nil {
		c.SetCookie(h.CreateLogoutCookie())
		return c.NoContent(http.StatusNoContent)
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

func (h *UserHandler) createRefreshTokenCookie(tokenValue string, expiration time.Time) *http.Cookie {
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
		Name:     "refresh_token",
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
func (h *UserHandler) CreateLogoutCookie() *http.Cookie {
	// 빈 값과 과거 만료일로 쿠키를 생성하여 쿠키를 삭제하는 효과를 냅니다
	return h.createRefreshTokenCookie("", time.Now().Add(-1*time.Hour))
}
