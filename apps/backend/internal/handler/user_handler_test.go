package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
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
		name       string
		pathParam  string
		mockReturn interface{}
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Get users by id successfully",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5",
			mockReturn: &domain.User{ID: "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5", Name: "John", Email: "john@example.com", Role: domain.RoleUser},
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `{"id":"01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5","name":"John","email":"john@example.com","role":"User","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:       "Fail to find user",
			pathParam:  "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z6",
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
		},
		{
			name:       "Fail to find user due to invalid id",
			pathParam:  "",
			mockReturn: nil,
			mockError:  domain.ErrInvalidInput,
			wantStatus: http.StatusBadRequest,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrInvalidInput.Error()),
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
			// Use pathParam directly as string (ULID)
			mockUseCase.On("GetByID", mock.Anything, tt.pathParam).Return(tt.mockReturn, tt.mockError).Maybe()
			handler := NewUserHandler(e, mockUseCase)

			err := handler.GetByID(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetAll(t *testing.T) {

	tests := []struct {
		name       string
		mockReturn interface{}
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name: "Get all users successfully",
			mockReturn: []domain.User{
				{ID: "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5", Name: "John", Email: "john@example.com", Role: domain.RoleUser},
				{ID: "01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z6", Name: "Jane", Email: "jane@example.com", Role: domain.RoleUser},
			},
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantBody:   `[{"id":"01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z5","name":"John","email":"john@example.com","role":"User","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"},{"id":"01HQZYX3VQJQZ3Z0Z1Z2Z3Z4Z6","name":"Jane","email":"jane@example.com","role":"User","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}]`,
		},
		{
			name:       "Fail to find any users",
			mockReturn: nil,
			mockError:  domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   fmt.Sprintf(`{"error":"%s"}`, domain.ErrNotFound.Error()),
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
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}
