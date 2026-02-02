package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/middleware"
)

type GameHandler struct {
	gameUseCase domain.GameUseCase
}

// NewGameHandler creates a new game handler and registers routes
func NewGameHandler(e *echo.Echo, gameUseCase domain.GameUseCase) *GameHandler {
	handler := &GameHandler{
		gameUseCase: gameUseCase,
	}

	// All routes are public for now
	group := e.Group("/games", middleware.AllowRoles(domain.RolePublic))
	group.POST("", handler.Create)
	group.GET("", handler.GetAll)
	group.GET("/:id", handler.GetByID)
	group.PUT("/:id", handler.Update)
	group.DELETE("/:id", handler.Delete)

	return handler
}

// Create handles POST /games - creates a new game
func (h *GameHandler) Create(c echo.Context) error {
	req := new(domain.CreateGameRequest)
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	game, err := h.gameUseCase.Create(ctx, req)
	if err == nil {
		return c.JSON(http.StatusCreated, game)
	}

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(err))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

// GetByID handles GET /games/:id - retrieves a single game
func (h *GameHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	game, err := h.gameUseCase.GetByID(ctx, id)
	if err == nil {
		return c.JSON(http.StatusOK, game)
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

// GetAll handles GET /games - retrieves all games
func (h *GameHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	games, err := h.gameUseCase.GetAll(ctx)
	if err == nil {
		return c.JSON(http.StatusOK, games)
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(err))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

// Update handles PUT /games/:id - updates an existing game
func (h *GameHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	req := new(domain.UpdateGameRequest)
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	game, err := h.gameUseCase.Update(ctx, id, req)
	if err == nil {
		return c.JSON(http.StatusOK, game)
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

// Delete handles DELETE /games/:id - deletes a game
func (h *GameHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	err = h.gameUseCase.Delete(ctx, id)
	if err == nil {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "game deleted successfully",
		})
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
