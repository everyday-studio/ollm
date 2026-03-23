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

	"github.com/everyday-studio/ollm/internal/config"
	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
)

// --- GetMe ---

func TestUserHandler_GetMe(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		mockReturn *domain.User
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:   "Get my profile successfully",
			userID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			mockReturn: &domain.User{
				ID:    "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				Name:  "John",
				Tag:   "john123",
				Email: "john@example.com",
				Role:  domain.RoleUser,
			},
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"id":"01HQZYX3VQJQZ3Z0Z1Z2ZUSER1","name":"John","tag":"john123","email":"john@example.com","role":"User","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:       "Fail due to missing user_id in context (unauthorized)",
			userID:     "", // not set in context
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusUnauthorized,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrUnauthorized.Error()),
		},
		{
			name:       "Fail due to user not found",
			userID:     "01HQZYX3VQJQZ3Z0Z1Z2ZUSER9",
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail due to internal error",
			userID:     "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			mockReturn: nil,
			mockError:  domain.ErrInternal,
			wantStatus: http.StatusInternalServerError,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInternal.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Only set user_id when it is not an unauthorized test case
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			mockUseCase := new(mocks.UserUseCase)
			mockUseCase.On("GetByID", mock.Anything, tt.userID).Return(tt.mockReturn, tt.mockError).Maybe()

			h := NewUserHandler(e, mockUseCase, &config.Config{})
			err := h.GetMe(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

// --- UpdateMe ---

func TestUserHandler_UpdateMe(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		body       string
		mockReturn *domain.User
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:   "Update nickname successfully",
			userID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			body:   `{"name":"NewNick"}`,
			mockReturn: &domain.User{
				ID:    "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				Name:  "NewNick",
				Tag:   "john123",
				Email: "john@example.com",
				Role:  domain.RoleUser,
			},
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"id":"01HQZYX3VQJQZ3Z0Z1Z2ZUSER1","name":"NewNick","tag":"john123","email":"john@example.com","role":"User","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:       "Fail due to missing user_id in context (unauthorized)",
			userID:     "", // not set in context
			body:       `{"name":"NewNick"}`,
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusUnauthorized,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrUnauthorized.Error()),
		},
		{
			name:       "Fail due to invalid JSON body",
			userID:     "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			body:       `invalid json`,
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
		{
			name:       "Fail due to nickname validation error (too short/long)",
			userID:     "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			body:       `{"name":"X"}`,
			mockReturn: nil,
			mockError:  domain.ErrInvalidInput,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
		{
			name:       "Fail due to user not found",
			userID:     "01HQZYX3VQJQZ3Z0Z1Z2ZUSER9",
			body:       `{"name":"NewNick"}`,
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail due to internal error",
			userID:     "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			body:       `{"name":"NewNick"}`,
			mockReturn: nil,
			mockError:  domain.ErrInternal,
			wantStatus: http.StatusInternalServerError,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInternal.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/api/users/me", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Only set user_id when it is not an unauthorized test case
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			mockUseCase := new(mocks.UserUseCase)
			// Only register mock if we expect the usecase to be called
			// (i.e., not the unauthorized or invalid JSON cases)
			if tt.userID != "" && tt.body != `invalid json` {
				mockUseCase.On("UpdateNickname", mock.Anything, tt.userID, mock.AnythingOfType("string")).Return(tt.mockReturn, tt.mockError).Maybe()
			}

			h := NewUserHandler(e, mockUseCase, &config.Config{})
			err := h.UpdateMe(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

// --- Withdraw ---

func TestUserHandler_Withdraw(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Withdraw user successfully",
			userID:     "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			mockError:  nil,
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
		{
			name:       "Fail due to missing user_id in context (unauthorized)",
			userID:     "", // not set in context
			mockError:  nil,
			wantStatus: http.StatusUnauthorized,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrUnauthorized.Error()),
		},
		{
			name:       "Fail due to user not found",
			userID:     "01HQZYX3VQJQZ3Z0Z1Z2ZUSER9",
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail due to internal error",
			userID:     "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
			mockError:  domain.ErrInternal,
			wantStatus: http.StatusInternalServerError,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInternal.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/api/users/me", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Only set user_id when it is not an unauthorized test case
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			mockUseCase := new(mocks.UserUseCase)
			if tt.userID != "" {
				mockUseCase.On("Delete", mock.Anything, tt.userID).Return(tt.mockError).Maybe()
			}

			h := NewUserHandler(e, mockUseCase, &config.Config{})
			err := h.Withdraw(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, rec.Body.String())
			} else {
				assert.Empty(t, rec.Body.String())
			}

			mockUseCase.AssertExpectations(t)
		})
	}
}
