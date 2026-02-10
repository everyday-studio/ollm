package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
)

// --- Create ---

func TestGameHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockReturn *domain.Game
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name: "Create game successfully",
			body: `{"title":"Adventure Quest","description":"A text-based adventure","author_id":"01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5"}`,
			mockReturn: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Adventure Quest",
				Description: "A text-based adventure",
				AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
				Status:      "active",
				IsPublic:    true,
			},
			mockError:  nil,
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"01HQZYX3VQJQZ3Z0Z1Z2GAME01","title":"Adventure Quest","description":"A text-based adventure","author_id":"01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5","status":"active","is_public":true,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:       "Fail to create game due to invalid input",
			body:       `invalid json`,
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
		{
			name:       "Fail to create game due to internal error",
			body:       `{"title":"Adventure Quest","description":"A text-based adventure","author_id":"01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5"}`,
			mockReturn: nil,
			mockError:  domain.ErrInternal,
			wantStatus: http.StatusInternalServerError,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInternal.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/games", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockUseCase := new(mocks.GameUseCase)
			if tt.name != "Fail to create game due to invalid input" {
				mockUseCase.On("Create", mock.Anything, mock.AnythingOfType("*domain.CreateGameRequest")).Return(tt.mockReturn, tt.mockError).Maybe()
			}
			handler := NewGameHandler(e, mockUseCase)

			err := handler.Create(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

// --- GetByID ---

func TestGameHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		pathParam  string
		mockReturn *domain.Game
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:      "Get game by ID successfully",
			pathParam: "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
			mockReturn: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Adventure Quest",
				Description: "A text-based adventure",
				AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
				Status:      "active",
				IsPublic:    true,
			},
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"id":"01HQZYX3VQJQZ3Z0Z1Z2GAME01","title":"Adventure Quest","description":"A text-based adventure","author_id":"01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5","status":"active","is_public":true,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:       "Fail to find game",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail with empty ID",
			pathParam:  "",
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/games/"+tt.pathParam, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.pathParam)

			mockUseCase := new(mocks.GameUseCase)
			mockUseCase.On("GetByID", mock.Anything, tt.pathParam).Return(tt.mockReturn, tt.mockError).Maybe()
			handler := NewGameHandler(e, mockUseCase)

			err := handler.GetByID(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

// --- GetAll ---

func TestGameHandler_GetAll(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn []domain.Game
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name: "Get all games successfully",
			mockReturn: []domain.Game{
				{
					ID:       "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
					Title:    "Game 1",
					Status:   "active",
					IsPublic: true,
				},
				{
					ID:       "01HQZYX3VQJQZ3Z0Z1Z2GAME02",
					Title:    "Game 2",
					Status:   "active",
					IsPublic: false,
				},
			},
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `[{"id":"01HQZYX3VQJQZ3Z0Z1Z2GAME01","title":"Game 1","description":"","author_id":"","status":"active","is_public":true,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"},{"id":"01HQZYX3VQJQZ3Z0Z1Z2GAME02","title":"Game 2","description":"","author_id":"","status":"active","is_public":false,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}]`,
		},
		{
			name:       "Fail to find any games",
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/games", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockUseCase := new(mocks.GameUseCase)
			mockUseCase.On("GetAll", mock.Anything).Return(tt.mockReturn, tt.mockError)
			handler := NewGameHandler(e, mockUseCase)

			err := handler.GetAll(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

// --- Update ---

func TestGameHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		pathParam  string
		body       string
		mockReturn *domain.Game
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:      "Update game successfully",
			pathParam: "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
			body:      `{"title":"Updated Title"}`,
			mockReturn: &domain.Game{
				ID:          "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
				Title:       "Updated Title",
				Description: "Original description",
				AuthorID:    "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
				Status:      "active",
				IsPublic:    true,
			},
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"id":"01HQZYX3VQJQZ3Z0Z1Z2GAME01","title":"Updated Title","description":"Original description","author_id":"01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5","status":"active","is_public":true,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:       "Fail to update non-existent game",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			body:       `{"title":"Updated Title"}`,
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail with empty ID",
			pathParam:  "",
			body:       `{"title":"Updated Title"}`,
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
		{
			name:       "Fail with invalid JSON body",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
			body:       `invalid json`,
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/games/"+tt.pathParam, strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.pathParam)

			mockUseCase := new(mocks.GameUseCase)
			mockUseCase.On("Update", mock.Anything, tt.pathParam, mock.AnythingOfType("*domain.UpdateGameRequest")).Return(tt.mockReturn, tt.mockError).Maybe()
			handler := NewGameHandler(e, mockUseCase)

			err := handler.Update(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

// --- Delete ---

func TestGameHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		pathParam  string
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Delete game successfully",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1Z2GAME01",
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"game deleted successfully"}`,
		},
		{
			name:       "Fail to delete non-existent game",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail with empty ID",
			pathParam:  "",
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/games/"+tt.pathParam, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.pathParam)

			mockUseCase := new(mocks.GameUseCase)
			mockUseCase.On("Delete", mock.Anything, tt.pathParam).Return(tt.mockError).Maybe()
			handler := NewGameHandler(e, mockUseCase)

			err := handler.Delete(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}
