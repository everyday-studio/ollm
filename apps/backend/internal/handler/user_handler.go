package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/middleware"
)

type UserHandler struct {
	userUseCase domain.UserUseCase
}

func NewUserHandler(e *echo.Echo, userUseCase domain.UserUseCase) *UserHandler {
	handler := &UserHandler{userUseCase: userUseCase}

	userGroup := e.Group("/users", middleware.AllowRoles(domain.RoleUser))
	userGroup.GET("/me", handler.GetMe)
	userGroup.PUT("/me", handler.UpdateMe)

	adminGroup := e.Group("/users", middleware.AllowRoles(domain.RoleAdmin))
	adminGroup.GET("", handler.GetAll)
	adminGroup.GET("/:id", handler.GetByID)

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

func (h *UserHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	user, err := h.userUseCase.GetByID(ctx, id)
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

func (h *UserHandler) GetAll(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 10
	}

	ctx := c.Request().Context()
	paginatedData, err := h.userUseCase.GetPaginated(ctx, page, limit)
	if err == nil {
		return c.JSON(http.StatusOK, paginatedData)
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}
