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

func TestMessageHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		pathParam  string
		body       string
		mockReturn *domain.Message
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:      "Create message successfully",
			pathParam: "01HQZYX3VQJQZ3Z0ZMATCH1",
			body:      `{"content":"Hello"}`,
			mockReturn: &domain.Message{
				ID:         "01HQZYX3VQJQZ3Z0ZMSG1",
				MatchID:    "01HQZYX3VQJQZ3Z0ZMATCH1",
				Role:       domain.MessageRoleAssistant,
				Content:    "Hi",
				IsVisible:  true,
				TurnCount:  1,
				TokenCount: 15,
			},
			mockError:  nil,
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"01HQZYX3VQJQZ3Z0ZMSG1","match_id":"01HQZYX3VQJQZ3Z0ZMATCH1","role":"assistant","content":"Hi","is_visible":true,"turn_count":1,"token_count":15,"created_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:       "Fail due to empty match param",
			pathParam:  "",
			body:       `{"content":"Hello"}`,
			mockReturn: nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
		{
			name:       "Fail due to invalid body",
			pathParam:  "01HQZYX3VQJQZ3Z0ZMATCH1",
			body:       `invalid json`,
			mockReturn: nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
		{
			name:       "Fail due to empty content",
			pathParam:  "01HQZYX3VQJQZ3Z0ZMATCH1",
			body:       `{"content":""}`,
			mockReturn: nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
		{
			name:       "Fail due to forbidden access",
			pathParam:  "01HQZYX3VQJQZ3Z0ZMATCH1",
			body:       `{"content":"Hello"}`,
			mockError:  domain.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrForbidden.Error()),
		},
		{
			name:       "Fail due to conflict status",
			pathParam:  "01HQZYX3VQJQZ3Z0ZMATCH1",
			body:       `{"content":"Hello"}`,
			mockError:  domain.ErrConflict,
			wantStatus: http.StatusConflict,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrConflict.Error()),
		},
		{
			name:       "Fail due to LLM error",
			pathParam:  "01HQZYX3VQJQZ3Z0ZMATCH1",
			body:       `{"content":"Hello"}`,
			mockError:  domain.ErrInternal,
			wantStatus: http.StatusInternalServerError,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInternal.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/matches/"+tt.pathParam+"/messages", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "01HQZYX3VQJQZ3Z0ZUSER1")
			c.SetParamNames("match_id")
			c.SetParamValues(tt.pathParam)

			mockUC := new(mocks.MessageUseCase)
			if tt.pathParam != "" && tt.name != "Fail due to invalid body" && tt.name != "Fail due to empty content" {
				mockUC.On("Create", mock.Anything, tt.pathParam, "01HQZYX3VQJQZ3Z0ZUSER1", mock.AnythingOfType("*domain.CreateMessageRequest")).Return(tt.mockReturn, tt.mockError)
			}

			h := NewMessageHandler(e, mockUC)
			err := h.Create(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUC.AssertExpectations(t)
		})
	}
}

// --- GetHistory ---

func TestMessageHandler_GetHistory(t *testing.T) {
	tests := []struct {
		name       string
		pathParam  string
		mockReturn []domain.Message
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:      "Get messages successfully",
			pathParam: "01HQZYX3VQJQZ3Z0ZMATCH1",
			mockReturn: []domain.Message{
				{
					ID:      "01HQZYX3VQJQZ3Z0ZMSG1",
					MatchID: "01HQZYX3VQJQZ3Z0ZMATCH1",
					Role:    domain.MessageRoleUser,
					Content: "Hello",
				},
				{
					ID:      "01HQZYX3VQJQZ3Z0ZMSG2",
					MatchID: "01HQZYX3VQJQZ3Z0ZMATCH1",
					Role:    domain.MessageRoleAssistant,
					Content: "Hi",
				},
			},
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `[{"id":"01HQZYX3VQJQZ3Z0ZMSG1","match_id":"01HQZYX3VQJQZ3Z0ZMATCH1","role":"user","content":"Hello","is_visible":false,"turn_count":0,"token_count":0,"created_at":"0001-01-01T00:00:00Z"},{"id":"01HQZYX3VQJQZ3Z0ZMSG2","match_id":"01HQZYX3VQJQZ3Z0ZMATCH1","role":"assistant","content":"Hi","is_visible":false,"turn_count":0,"token_count":0,"created_at":"0001-01-01T00:00:00Z"}]`,
		},
		{
			name:       "Fail due to empty param",
			pathParam:  "",
			mockReturn: nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
		{
			name:       "Fail due to forbidden access",
			pathParam:  "01HQZYX3VQJQZ3Z0ZMATCH1",
			mockError:  domain.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrForbidden.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/matches/"+tt.pathParam+"/messages", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "01HQZYX3VQJQZ3Z0ZUSER1")
			c.SetParamNames("match_id")
			c.SetParamValues(tt.pathParam)

			mockUC := new(mocks.MessageUseCase)
			if tt.pathParam != "" {
				mockUC.On("GetByMatchID", mock.Anything, tt.pathParam, "01HQZYX3VQJQZ3Z0ZUSER1").Return(tt.mockReturn, tt.mockError)
			}

			h := NewMessageHandler(e, mockUC)
			err := h.GetHistory(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUC.AssertExpectations(t)
		})
	}
}
