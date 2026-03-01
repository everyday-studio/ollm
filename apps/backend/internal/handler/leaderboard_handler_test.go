package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLeaderboardHandler_GetLeaderboard(t *testing.T) {
	e := echo.New()

	t.Run("Get leaderboard successfully", func(t *testing.T) {
		mockUseCase := new(mocks.LeaderboardUseCase)
		handler := NewLeaderboardHandler(e, mockUseCase)

		gameID := "test-game-id"

		req := httptest.NewRequest(http.MethodGet, "/games/"+gameID+"/leaderboard", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/games/:id/leaderboard")
		c.SetParamNames("id")
		c.SetParamValues(gameID)

		expectedEntries := []domain.LeaderboardEntry{
			{Rank: 1, UserID: "user_1", Username: "User 1", TurnCount: 5},
			{Rank: 2, UserID: "user_2", Username: "User 2", TurnCount: 8},
		}

		mockUseCase.On("GetLeaderboard", req.Context(), gameID, 10).Return(expectedEntries, nil)

		err := handler.GetLeaderboard(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string][]domain.LeaderboardEntry
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp["data"], 2)
		assert.Equal(t, expectedEntries[0].UserID, resp["data"][0].UserID)

		mockUseCase.AssertExpectations(t)
	})

	t.Run("Return bad request if gameID is empty", func(t *testing.T) {
		mockUseCase := new(mocks.LeaderboardUseCase)
		handler := NewLeaderboardHandler(e, mockUseCase)

		req := httptest.NewRequest(http.MethodGet, "/games//leaderboard", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/games/:id/leaderboard")
		c.SetParamNames("id")
		c.SetParamValues("") // Empty game ID

		err := handler.GetLeaderboard(c)

		// The error returned from GetLeaderboard is evaluated by Echo error handler,
		// but since we are testing the handler method directly, we check the returned error
		var httpErr *echo.HTTPError
		assert.ErrorAs(t, err, &httpErr)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Equal(t, "game ID is required", httpErr.Message)
	})

	t.Run("Return not found from use case", func(t *testing.T) {
		mockUseCase := new(mocks.LeaderboardUseCase)
		handler := NewLeaderboardHandler(e, mockUseCase)

		gameID := "test-game-id"

		req := httptest.NewRequest(http.MethodGet, "/games/"+gameID+"/leaderboard", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/games/:id/leaderboard")
		c.SetParamNames("id")
		c.SetParamValues(gameID)

		mockUseCase.On("GetLeaderboard", req.Context(), gameID, 10).Return(nil, domain.ErrNotFound)

		err := handler.GetLeaderboard(c)

		// It returns c.JSON in the handler instead of returning an error
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		var resp ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, domain.ErrNotFound.Error(), resp.Error)
	})
}
