package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/middleware"
)

// GameHandler handles HTTP requests for game resources
type GameHandler struct {
	gameUseCase domain.GameUseCase
}

// NewGameHandler creates a new game handler and registers routes
func NewGameHandler(e *echo.Echo, gameUseCase domain.GameUseCase) *GameHandler {
	handler := &GameHandler{
		gameUseCase: gameUseCase,
	}

	// Public routes
	publicGroup := e.Group("/api/games", middleware.AllowRoles(domain.RolePublic))
	publicGroup.GET("", handler.GetAll)
	publicGroup.GET("/:id", handler.GetByID)

	// Admin routes
	adminGroup := e.Group("/api/games", middleware.AllowRoles(domain.RoleAdmin))
	adminGroup.POST("", handler.Create)
	adminGroup.PUT("/:id", handler.Update)

	return handler
}

// Create handles POST /games - creates a new game
func (h *GameHandler) Create(c echo.Context) error {
	req := new(domain.CreateGameRequest)
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}
	req.AuthorID = userID

	ctx := c.Request().Context()
	game, err := h.gameUseCase.Create(ctx, req)
	if err == nil {
		return c.JSON(http.StatusCreated, game)
	}

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(err))
	case errors.Is(err, domain.ErrConflict):
		return c.JSON(http.StatusConflict, ErrResponse(err))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

// GetByID handles GET /games/:id - retrieves a single game
func (h *GameHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	game, err := h.gameUseCase.GetByID(ctx, id)
	if err == nil {
		// Hide sensitive information
		game.SystemPrompt = ""
		game.JudgeCondition = ""
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
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 10
	}

	sortParam := c.QueryParam("sort")
	var sortBy domain.GameSortBy
	switch sortParam {
	case string(domain.GameSortByName):
		sortBy = domain.GameSortByName
	case string(domain.GameSortByPopular):
		sortBy = domain.GameSortByPopular
	default:
		sortBy = domain.GameSortByRecent
	}

	ctx := c.Request().Context()
	isPublic := true
	filter := &domain.GameFilter{
		IsPublic: &isPublic,
		SortBy:   sortBy,
	}

	paginatedData, err := h.gameUseCase.GetPaginated(ctx, page, limit, filter)
	if err == nil {
		// Hide sensitive information
		for i := range paginatedData.Data {
			paginatedData.Data[i].SystemPrompt = ""
			paginatedData.Data[i].JudgeCondition = ""
		}
		return c.JSON(http.StatusOK, paginatedData)
	}

	return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
}

// Update handles PUT /games/:id - updates an existing game
func (h *GameHandler) Update(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
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
