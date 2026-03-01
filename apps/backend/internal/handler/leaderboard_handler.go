package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/everyday-studio/ollm/internal/domain"
)

type LeaderboardHandler struct {
	usecase domain.LeaderboardUseCase
}

// NewLeaderboardHandler creates a new leaderboard handler
func NewLeaderboardHandler(e *echo.Echo, usecase domain.LeaderboardUseCase) *LeaderboardHandler {
	handler := &LeaderboardHandler{
		usecase: usecase,
	}

	e.GET("/games/:id/leaderboard", handler.GetLeaderboard)

	return handler
}

// GetLeaderboard handles the request to fetch the leaderboard for a game
func (h *LeaderboardHandler) GetLeaderboard(c echo.Context) error {
	gameID := c.Param("id")
	if gameID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "game ID is required")
	}

	ctx := c.Request().Context()

	// Requesting top 10 as per requirements
	limit := 10

	leaderboard, err := h.usecase.GetLeaderboard(ctx, gameID, limit)
	if err == nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data": leaderboard,
		})
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}
