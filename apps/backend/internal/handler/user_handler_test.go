package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/domain/mocks"
)

func TestUserHandler_ErrResponse(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "InvalidInput",
			args: args{err: domain.ErrInvalidInput},
			want: map[string]string{"error": domain.ErrInvalidInput.Error()},
		},
		{
			name: "AlreadyExists",
			args: args{err: domain.ErrAlreadyExists},
			want: map[string]string{"error": domain.ErrAlreadyExists.Error()},
		},
		{
			name: "Internal",
			args: args{err: domain.ErrInternal},
			want: map[string]string{"error": domain.ErrInternal.Error()},
		},
		{
			name: "NilError",
			args: args{err: nil},
			want: map[string]string{"error": "unknown error"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrResponse(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ErrResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserHandler_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		pathParam      string
		mockReturn     interface{}
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Get users by id successfully",
			pathParam:      "1",
			mockReturn:     &domain.User{ID: 1, Name: "John", Email: "john@example.com", Role: domain.RoleUser},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"name":"John","email":"john@example.com","role":"User"}`,
		},
		{
			name:           "Fail to find user",
			pathParam:      "1",
			mockReturn:     nil,
			mockError:      domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:           "Fail to find user due to invalid id",
			pathParam:      "invalid",
			mockReturn:     nil,
			mockError:      domain.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.pathParam, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.pathParam)

			mockUseCase := new(mocks.UserUseCase)
			// Convert pathParam to int64 for mock expectation
			var expectedID int64
			if id, err := strconv.Atoi(tt.pathParam); err == nil {
				expectedID = int64(id)
			}
			mockUseCase.On("GetByID", mock.Anything, expectedID).Return(tt.mockReturn, tt.mockError).Maybe()
			handler := NewUserHandler(e, mockUseCase)

			err := handler.GetByID(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetAll(t *testing.T) {

	tests := []struct {
		name           string
		mockReturn     interface{}
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Get all users successfully",
			mockReturn: []domain.User{
				{ID: 1, Name: "John", Email: "john@example.com", Role: domain.RoleUser},
				{ID: 2, Name: "Jane", Email: "jane@example.com", Role: domain.RoleUser},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"name":"John","email":"john@example.com","role":"User"},{"id":2,"name":"Jane","email":"jane@example.com","role":"User"}]`,
		},
		{
			name:           "Fail to find any users",
			mockReturn:     nil,
			mockError:      domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockUseCase := new(mocks.UserUseCase)
			mockUseCase.On("GetAll", mock.Anything).Return(tt.mockReturn, tt.mockError)
			handler := NewUserHandler(e, mockUseCase)

			err := handler.GetAll(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}
