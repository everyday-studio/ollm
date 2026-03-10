package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/middleware"
)

// UploadHandler handles HTTP requests for uploads
type UploadHandler struct {
	uploadUseCase domain.UploadUseCase
}

// NewUploadHandler creates a new upload handler and registers routes
func NewUploadHandler(e *echo.Echo, uploadUseCase domain.UploadUseCase) *UploadHandler {
	handler := &UploadHandler{
		uploadUseCase: uploadUseCase,
	}

	// Since uploading requires tracking user ownership, RoleUser is the minimum.
	userGroup := e.Group("/upload", middleware.AllowRoles(domain.RoleUser))
	userGroup.POST("/image", handler.UploadImage)

	return handler
}

// UploadImage handles POST /upload/image (multipart/form-data)
func (h *UploadHandler) UploadImage(c echo.Context) error {
	// 1. Get user logic
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	}

	// 2. Parse form
	uploadType := c.FormValue("type")
	refID := c.FormValue("ref_id")
	if uploadType == "" || refID == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(errors.New("type and ref_id form fields are required")))
	}

	// 3. Read File
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(errors.New("failed to read file from request")))
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
	defer src.Close()

	// 4. Construct Request
	req := &domain.UploadImageRequest{
		Type:        domain.UploadType(uploadType),
		RefID:       refID,
		File:        src,
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		FileSize:    fileHeader.Size,
		UpdaterID:   userID,
	}

	ctx := c.Request().Context()
	resp, err := h.uploadUseCase.UploadImage(ctx, req)
	if err == nil {
		return c.JSON(http.StatusOK, resp)
	}

	// 5. Error handling mapping
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	case errors.Is(err, domain.ErrUnauthorized):
		return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
	case errors.Is(err, domain.ErrForbidden):
		return c.JSON(http.StatusForbidden, ErrResponse(domain.ErrForbidden))
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
	default:
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}
}
