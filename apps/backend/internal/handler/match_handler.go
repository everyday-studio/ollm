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

	// User routes
	userGroup := e.Group("/matches", middleware.AllowRoles(domain.RoleUser))
	userGroup.POST("", handler.Create)
	userGroup.GET("/me", handler.GetMyMatches)
	userGroup.GET("/:id", handler.GetByID)
	userGroup.POST("/:id/resign", handler.Resign)

	// Admin routes
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

	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}
	req.UserID = userID

	ctx := c.Request().Context()
	match, err := h.matchUseCase.Create(ctx, req)
	if err == nil {
		return c.JSON(http.StatusCreated, match)
	}

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
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

	// Extract user ID from JWT token
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}

	ctx := c.Request().Context()
	match, err := h.matchUseCase.GetByID(ctx, id, userID)
	if err == nil {
		return c.JSON(http.StatusOK, match)
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
	case errors.Is(err, domain.ErrForbidden):
		return c.JSON(http.StatusForbidden, ErrResponse(domain.ErrForbidden))
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

// GetMyMatches handles GET /matches/me - retrieves all matches for the authenticated user
// Supports optional query parameter: game_id (filters matches by specific game)
func (h *MatchHandler) GetMyMatches(c echo.Context) error {
	// Extract user ID from JWT token
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}

	ctx := c.Request().Context()
	gameID := c.QueryParam("game_id")

	var matches []domain.Match
	var err error

	// If game_id is provided, filter by game_id as well
	if gameID != "" {
		matches, err = h.matchUseCase.GetByUserIDAndGameID(ctx, userID, gameID)
	} else {
		matches, err = h.matchUseCase.GetByUserID(ctx, userID)
	}

	if err == nil {
		return c.JSON(http.StatusOK, matches)
	}

	return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
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
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

// Resign handles POST /matches/:id/resign - allows user to forfeit a match
func (h *MatchHandler) Resign(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}

	ctx := c.Request().Context()
	err := h.matchUseCase.Resign(ctx, id, userID)
	if err == nil {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "successfully resigned from match",
		})
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
	case errors.Is(err, domain.ErrForbidden):
		return c.JSON(http.StatusForbidden, ErrResponse(domain.ErrForbidden))
	case errors.Is(err, domain.ErrConflict):
		return c.JSON(http.StatusConflict, ErrResponse(domain.ErrConflict))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}
