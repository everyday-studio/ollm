package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/middleware"
)

// MatchHandler handles HTTP requests for match resources
type MatchHandler struct {
	matchUseCase domain.MatchUseCase
}

// NewMatchHandler creates a new match handler and registers routes
func NewMatchHandler(e *echo.Echo, matchUseCase domain.MatchUseCase) *MatchHandler {
	handler := &MatchHandler{
		matchUseCase: matchUseCase,
	}

	// Public routes
	group := e.Group("/matches", middleware.AllowRoles(domain.RolePublic))
	group.POST("", handler.Create)
	group.GET("/:id", handler.GetByID)
	group.GET("/user/:userId", handler.GetByUserID)

	// Admin-only routes
	adminGroup := e.Group("/matches", middleware.AllowRoles(domain.RoleAdmin))
	adminGroup.DELETE("/:id", handler.Delete)

	return handler
}

// Create handles POST /matches - creates a new match
func (h *MatchHandler) Create(c echo.Context) error {
	req := new(domain.CreateMatchRequest)
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	match, err := h.matchUseCase.Create(ctx, req)
	if err == nil {
		return c.JSON(http.StatusCreated, match)
	}

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(err))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

// GetByID handles GET /matches/:id - retrieves a single match
func (h *MatchHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	match, err := h.matchUseCase.GetByID(ctx, id)
	if err == nil {
		return c.JSON(http.StatusOK, match)
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(err))
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(err))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

// GetByUserID handles GET /matches/user/:userId - retrieves all matches for a user
func (h *MatchHandler) GetByUserID(c echo.Context) error {
	userID := c.Param("userId")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	matches, err := h.matchUseCase.GetByUserID(ctx, userID)
	if err == nil {
		return c.JSON(http.StatusOK, matches)
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(err))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

// Delete handles DELETE /matches/:id - deletes a match (admin only)
func (h *MatchHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	err := h.matchUseCase.Delete(ctx, id)
	if err == nil {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "match deleted successfully",
		})
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(err))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}
