package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
)

// --- Create ---

func TestMatchHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockReturn *domain.Match
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name: "Create match successfully",
			body: `{"game_id":"01HQZYX3VQJQZ3Z0Z1Z2ZGAME1"}`,
			mockReturn: &domain.Match{
				ID:          "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID:      "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				GameID:      "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
				Status:      domain.MatchStatusActive,
				MaxTurns:    10,
				TotalTokens: 0,
				TurnCount:   0,
				CreatedAt:   time.Time{}, // zero value for json comparison
				UpdatedAt:   time.Time{},
			},
			mockError:  nil,
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"01HQZYX3VQJQZ3Z0Z1ZMATCH01","user_id":"01HQZYX3VQJQZ3Z0Z1Z2ZUSER1","game_id":"01HQZYX3VQJQZ3Z0Z1Z2ZGAME1","status":"active","max_turns":10,"total_tokens":0,"turn_count":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:       "Fail due to invalid JSON body",
			body:       `invalid json`,
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
		{
			name:       "Fail due to game not found",
			body:       `{"game_id":"01HQZYX3VQJQZ3Z0Z1Z2NONEXIST"}`,
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail due to internal error",
			body:       `{"game_id":"01HQZYX3VQJQZ3Z0Z1Z2ZGAME1"}`,
			mockReturn: nil,
			mockError:  domain.ErrInternal,
			wantStatus: http.StatusInternalServerError,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInternal.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/matches", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1")

			mockUseCase := new(mocks.MatchUseCase)
			if tt.name != "Fail due to invalid JSON body" {
				mockUseCase.On("Create", mock.Anything, mock.AnythingOfType("*domain.CreateMatchRequest")).Return(tt.mockReturn, tt.mockError).Maybe()
			}

			h := NewMatchHandler(e, mockUseCase)
			err := h.Create(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

// --- GetByID ---

func TestMatchHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		pathParam  string
		mockReturn *domain.Match
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:      "Get match by ID successfully",
			pathParam: "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			mockReturn: &domain.Match{
				ID:          "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
				UserID:      "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
				GameID:      "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
				Status:      domain.MatchStatusActive,
				MaxTurns:    10,
				TotalTokens: 0,
				TurnCount:   0,
			},
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"id":"01HQZYX3VQJQZ3Z0Z1ZMATCH01","user_id":"01HQZYX3VQJQZ3Z0Z1Z2ZUSER1","game_id":"01HQZYX3VQJQZ3Z0Z1Z2ZGAME1","status":"active","max_turns":10,"total_tokens":0,"turn_count":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:       "Fail due to not found",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail due to forbidden (not owner)",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			mockReturn: nil,
			mockError:  domain.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrForbidden.Error()),
		},
		{
			name:       "Fail due to empty id",
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
			req := httptest.NewRequest(http.MethodGet, "/matches/"+tt.pathParam, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1")
			c.SetParamNames("id")
			c.SetParamValues(tt.pathParam)

			mockUseCase := new(mocks.MatchUseCase)
			mockUseCase.On("GetByID", mock.Anything, tt.pathParam, mock.Anything).Return(tt.mockReturn, tt.mockError).Maybe()

			h := NewMatchHandler(e, mockUseCase)
			err := h.GetByID(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

// --- GetMyMatches ---

func TestMatchHandler_GetMyMatches(t *testing.T) {
	tests := []struct {
		name         string
		queryGameID  string
		mockReturn   []domain.Match
		mockError    error
		wantStatus   int
		wantBody     string
		expectGameID bool
	}{
		{
			name:        "Get all matches successfully",
			queryGameID: "",
			mockReturn: []domain.Match{
				{
					ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
					UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
					GameID: "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
					Status: domain.MatchStatusActive,
				},
				{
					ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH02",
					UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
					GameID: "01HQZYX3VQJQZ3Z0Z1Z2ZGAME2",
					Status: domain.MatchStatusWon,
				},
			},
			mockError:    nil,
			wantStatus:   http.StatusOK,
			wantBody:     `[{"id":"01HQZYX3VQJQZ3Z0Z1ZMATCH01","user_id":"01HQZYX3VQJQZ3Z0Z1Z2ZUSER1","game_id":"01HQZYX3VQJQZ3Z0Z1Z2ZGAME1","status":"active","max_turns":0,"total_tokens":0,"turn_count":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"},{"id":"01HQZYX3VQJQZ3Z0Z1ZMATCH02","user_id":"01HQZYX3VQJQZ3Z0Z1Z2ZUSER1","game_id":"01HQZYX3VQJQZ3Z0Z1Z2ZGAME2","status":"won","max_turns":0,"total_tokens":0,"turn_count":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}]`,
			expectGameID: false,
		},
		{
			name:        "Get matches filtered by game ID successfully",
			queryGameID: "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
			mockReturn: []domain.Match{
				{
					ID:     "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
					UserID: "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1",
					GameID: "01HQZYX3VQJQZ3Z0Z1Z2ZGAME1",
					Status: domain.MatchStatusActive,
				},
			},
			mockError:    nil,
			wantStatus:   http.StatusOK,
			wantBody:     `[{"id":"01HQZYX3VQJQZ3Z0Z1ZMATCH01","user_id":"01HQZYX3VQJQZ3Z0Z1Z2ZUSER1","game_id":"01HQZYX3VQJQZ3Z0Z1Z2ZGAME1","status":"active","max_turns":0,"total_tokens":0,"turn_count":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}]`,
			expectGameID: true,
		},
		{
			name:         "Fail to get matches due to internal error",
			queryGameID:  "",
			mockReturn:   nil,
			mockError:    domain.ErrInternal,
			wantStatus:   http.StatusInternalServerError,
			wantBody:     fmt.Sprintf(`{"error":"%s"}`, domain.ErrInternal.Error()),
			expectGameID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			url := "/matches/me"
			if tt.queryGameID != "" {
				url += "?game_id=" + tt.queryGameID
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1")

			mockUseCase := new(mocks.MatchUseCase)
			if tt.expectGameID {
				mockUseCase.On("GetByUserIDAndGameID", mock.Anything, "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1", tt.queryGameID).Return(tt.mockReturn, tt.mockError)
			} else {
				mockUseCase.On("GetByUserID", mock.Anything, "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1").Return(tt.mockReturn, tt.mockError)
			}

			h := NewMatchHandler(e, mockUseCase)
			err := h.GetMyMatches(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

// --- Delete ---

func TestMatchHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		pathParam  string
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Delete match successfully",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"match deleted successfully"}`,
		},
		{
			name:       "Fail due to not found",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail due to empty path param",
			pathParam:  "",
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/matches/"+tt.pathParam, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1") // user_id is mostly irrelevant as it's an admin route, but for sanity
			c.SetParamNames("id")
			c.SetParamValues(tt.pathParam)

			mockUseCase := new(mocks.MatchUseCase)
			mockUseCase.On("Delete", mock.Anything, tt.pathParam).Return(tt.mockError).Maybe()

			h := NewMatchHandler(e, mockUseCase)
			err := h.Delete(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

// --- Resign ---

func TestMatchHandler_Resign(t *testing.T) {
	tests := []struct {
		name       string
		pathParam  string
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Resign match successfully",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"successfully resigned from match"}`,
		},
		{
			name:       "Fail to resign empty match id",
			pathParam:  "",
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
		{
			name:       "Fail to resign non-existent match",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1Z2NONEXIST",
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail to resign due to forbidden access (not owner)",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			mockError:  domain.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrForbidden.Error()),
		},
		{
			name:       "Fail to resign game already finished",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1ZMATCH01",
			mockError:  domain.ErrConflict,
			wantStatus: http.StatusConflict,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrConflict.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/matches/"+tt.pathParam+"/resign", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1")
			c.SetParamNames("id")
			c.SetParamValues(tt.pathParam)

			mockUseCase := new(mocks.MatchUseCase)
			mockUseCase.On("Resign", mock.Anything, tt.pathParam, "01HQZYX3VQJQZ3Z0Z1Z2ZUSER1").Return(tt.mockError).Maybe()

			h := NewMatchHandler(e, mockUseCase)
			err := h.Resign(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}
