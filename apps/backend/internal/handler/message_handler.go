package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/middleware"
)

type MessageHandler struct {
	messageUC domain.MessageUseCase
}

// NewMessageHandler creates a new message handler and registers routes
func NewMessageHandler(e *echo.Echo, messageUC domain.MessageUseCase) *MessageHandler {
	handler := &MessageHandler{
		messageUC: messageUC,
	}

	// Group routes with JWT auth
	userGroup := e.Group("/matches/:match_id/messages", middleware.AllowRoles(domain.RoleUser))
	userGroup.POST("", handler.Create)
	userGroup.GET("", handler.GetHistory)

	return handler
}

// Create handles the endpoint to send a new message and receive the AI response
func (h *MessageHandler) Create(c echo.Context) error {
	matchID := c.Param("match_id")
	if matchID == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}

	var req domain.CreateMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	aiMsg, err := h.messageUC.Create(ctx, matchID, userID, &req)
	if err == nil {
		return c.JSON(http.StatusCreated, aiMsg)
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(err))
	case errors.Is(err, domain.ErrForbidden):
		return c.JSON(http.StatusForbidden, ErrResponse(err))
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(err))
	case errors.Is(err, domain.ErrConflict):
		return c.JSON(http.StatusConflict, ErrResponse(err))
	default:
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}

// GetHistory retrieves the conversation history of a match
func (h *MessageHandler) GetHistory(c echo.Context) error {
	matchID := c.Param("match_id")
	if matchID == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}

	ctx := c.Request().Context()
	messages, err := h.messageUC.GetByMatchID(ctx, matchID, userID)
	if err == nil {
		return c.JSON(http.StatusOK, messages)
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(err))
	case errors.Is(err, domain.ErrForbidden):
		return c.JSON(http.StatusForbidden, ErrResponse(err))
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(err))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}
