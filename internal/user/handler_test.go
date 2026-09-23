package user

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/middleware/header"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/response"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/user/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockService is a testify mock implementing the Service interface, so
// Handler tests never touch the real service, repo, or a database.
type mockService struct {
	mock.Mock
}

func (m *mockService) Create(ctx context.Context, input *dto.CreateUserInput) (*dto.UserResponseData, error) {
	args := m.Called(ctx, input)
	out, _ := args.Get(0).(*dto.UserResponseData)
	return out, args.Error(1)
}

func (m *mockService) Update(ctx context.Context, id string, input *dto.UpdateUserInput) (*dto.UserResponseData, error) {
	args := m.Called(ctx, id, input)
	out, _ := args.Get(0).(*dto.UserResponseData)
	return out, args.Error(1)
}

func (m *mockService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockService) GetByID(ctx context.Context, id string) (*dto.UserResponseData, error) {
	args := m.Called(ctx, id)
	out, _ := args.Get(0).(*dto.UserResponseData)
	return out, args.Error(1)
}

func (m *mockService) GetByEmail(ctx context.Context, email string) (*dto.UserResponseData, error) {
	args := m.Called(ctx, email)
	out, _ := args.Get(0).(*dto.UserResponseData)
	return out, args.Error(1)
}

func (m *mockService) GetMany(ctx context.Context) ([]dto.UserResponseData, error) {
	args := m.Called(ctx)
	out, _ := args.Get(0).([]dto.UserResponseData)
	return out, args.Error(1)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// serve wraps the handler with the same request-ID middleware application.go
// applies, since GetRequestIDFromContext panics if the value is missing.
func serve(t *testing.T, h http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	header.SetRequestID(h).ServeHTTP(rec, req)
	return rec
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) response.ErrorResponse {
	t.Helper()
	var body response.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return body
}

func decodeSuccess(t *testing.T, rec *httptest.ResponseRecorder) response.JSONResponse {
	t.Helper()
	var body response.JSONResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return body
}

// NOTE: the "invalid type" cases below assume CreateUserInput/UpdateUserInput
// use JSON tags "name"/"email"/"password" (matching the exported Go field
// names lowercased). I haven't seen the dto file — adjust the request bodies
// below if your tags differ.

func TestHandler_CreateUser(t *testing.T) {
	validData := &dto.UserResponseData{ID: uuid.New(), Name: "Subham", Email: "subham@example.com"}

	tests := []struct {
		name       string
		body       string
		mockSetup  func(m *mockService)
		wantStatus int
		check      func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:       "malformed json returns 422",
			body:       `{"name":`,
			wantStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeError(t, rec)
				assert.False(t, body.Success)
				assert.Equal(t, response.CodeUnprocessableEntity, body.Error.Code)
			},
		},
		{
			name:       "wrong json field type returns 422 with field details",
			body:       `{"name": 123, "email": "e@example.com", "password": "pw"}`,
			wantStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeError(t, rec)
				assert.Equal(t, response.CodeUnprocessableEntity, body.Error.Code)
				assert.NotNil(t, body.Error.Details)
			},
		},
		{
			name: "service validation error maps to its own code",
			body: `{"name": "Subham", "email": "bad-email", "password": "pw"}`,
			mockSetup: func(m *mockService) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, &errorsx.ValidationError{
					Code:    response.CodeBadRequest,
					Message: "invalid email",
					Details: []errorsx.ValidationErrorDetails{{Field: "email", Message: "invalid format"}},
				})
			},
			wantStatus: http.StatusBadRequest,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeError(t, rec)
				assert.Equal(t, response.CodeBadRequest, body.Error.Code)
			},
		},
		{
			name: "duplicate email returns 409 conflict",
			body: `{"name": "Subham", "email": "dup@example.com", "password": "pw"}`,
			mockSetup: func(m *mockService) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, errorsx.ErrEmailAlreadyExists)
			},
			wantStatus: http.StatusConflict,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeError(t, rec)
				assert.Equal(t, response.CodeConflict, body.Error.Code)
				assert.Equal(t, errorsx.ErrEmailAlreadyExists.Error(), body.Error.Message)
			},
		},
		{
			name: "duplicate key returns 409 with duplicate key code",
			body: `{"name": "Subham", "email": "dup@example.com", "password": "pw"}`,
			mockSetup: func(m *mockService) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, errorsx.ErrDuplicateKey)
			},
			wantStatus: http.StatusConflict,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeError(t, rec)
				assert.Equal(t, response.CodeDuplicateKeyError, body.Error.Code)
			},
		},
		{
			name: "unmapped service error returns 500",
			body: `{"name": "Subham", "email": "e@example.com", "password": "pw"}`,
			mockSetup: func(m *mockService) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeError(t, rec)
				assert.Equal(t, response.CodeInternalServerError, body.Error.Code)
			},
		},
		{
			name: "success returns 201 with created user",
			body: `{"name": "Subham", "email": "subham@example.com", "password": "pw"}`,
			mockSetup: func(m *mockService) {
				m.On("Create", mock.Anything, mock.Anything).Return(validData, nil)
			},
			wantStatus: http.StatusCreated,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeSuccess(t, rec)
				assert.True(t, body.Success)
				assert.Equal(t, "user created successfully", body.Message)
				assert.NotNil(t, body.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockService)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}
			h := NewHandler(svc, testLogger())

			req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.body))
			rec := serve(t, h.CreateUser, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.check != nil {
				tt.check(t, rec)
			}
			svc.AssertExpectations(t)
		})
	}
}

func TestHandler_UpdateUser(t *testing.T) {
	id := uuid.New().String()
	validData := &dto.UserResponseData{ID: uuid.MustParse(id), Name: "New Name", Email: "new@example.com"}

	tests := []struct {
		name       string
		pathID     string // "" means no id path value at all
		body       string
		mockSetup  func(m *mockService)
		wantStatus int
		check      func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:       "missing id path value returns 400 without calling service",
			pathID:     "",
			body:       `{"name": "New Name", "email": "new@example.com"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed json returns 422",
			pathID:     id,
			body:       `{"name":`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "wrong json field type returns 422 with field details",
			pathID:     id,
			body:       `{"name": 123, "email": "e@example.com"}`,
			wantStatus: http.StatusUnprocessableEntity,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeError(t, rec)
				assert.NotNil(t, body.Error.Details)
			},
		},
		{
			name:   "invalid id from service returns 400",
			pathID: "not-a-uuid",
			body:   `{"name": "New Name", "email": "new@example.com"}`,
			mockSetup: func(m *mockService) {
				m.On("Update", mock.Anything, "not-a-uuid", mock.Anything).
					Return(nil, errorsx.ErrInvalidID)
			},
			wantStatus: http.StatusBadRequest,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeError(t, rec)
				assert.Equal(t, errorsx.ErrInvalidID.Error(), body.Error.Message)
			},
		},
		{
			name:   "duplicate email returns 409 conflict",
			pathID: id,
			body:   `{"name": "New Name", "email": "dup@example.com"}`,
			mockSetup: func(m *mockService) {
				m.On("Update", mock.Anything, id, mock.Anything).
					Return(nil, errorsx.ErrEmailAlreadyExists)
			},
			wantStatus: http.StatusConflict,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeError(t, rec)
				assert.Equal(t, response.CodeConflict, body.Error.Code)
			},
		},
		{
			name:   "duplicate key returns 409",
			pathID: id,
			body:   `{"name": "New Name", "email": "dup@example.com"}`,
			mockSetup: func(m *mockService) {
				m.On("Update", mock.Anything, id, mock.Anything).
					Return(nil, errorsx.ErrDuplicateKey)
			},
			wantStatus: http.StatusConflict,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeError(t, rec)
				assert.Equal(t, response.CodeDuplicateKeyError, body.Error.Code)
			},
		},
		{
			name:   "unmapped service error returns 500",
			pathID: id,
			body:   `{"name": "New Name", "email": "new@example.com"}`,
			mockSetup: func(m *mockService) {
				m.On("Update", mock.Anything, id, mock.Anything).
					Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "success returns 200 with updated user",
			pathID: id,
			body:   `{"name": "New Name", "email": "new@example.com"}`,
			mockSetup: func(m *mockService) {
				m.On("Update", mock.Anything, id, mock.Anything).Return(validData, nil)
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeSuccess(t, rec)
				assert.Equal(t, "user updated successfully", body.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockService)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}
			h := NewHandler(svc, testLogger())

			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.pathID, strings.NewReader(tt.body))
			if tt.pathID != "" {
				req.SetPathValue("id", tt.pathID)
			}
			rec := serve(t, h.UpdateUser, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.check != nil {
				tt.check(t, rec)
			}
			svc.AssertExpectations(t)
		})
	}
}

func TestHandler_DeleteUser(t *testing.T) {
	id := uuid.New().String()

	tests := []struct {
		name       string
		pathID     string
		mockSetup  func(m *mockService)
		wantStatus int
	}{
		{
			name:       "missing id path value returns 400 without calling service",
			pathID:     "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid id returns 400",
			pathID: "not-a-uuid",
			mockSetup: func(m *mockService) {
				m.On("Delete", mock.Anything, "not-a-uuid").Return(errorsx.ErrInvalidID)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "unmapped service error returns 500",
			pathID: id,
			mockSetup: func(m *mockService) {
				m.On("Delete", mock.Anything, id).Return(errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "success returns 200",
			pathID: id,
			mockSetup: func(m *mockService) {
				m.On("Delete", mock.Anything, id).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockService)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}
			h := NewHandler(svc, testLogger())

			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.pathID, nil)
			if tt.pathID != "" {
				req.SetPathValue("id", tt.pathID)
			}
			rec := serve(t, h.DeleteUser, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetUserByID(t *testing.T) {
	id := uuid.New().String()
	validData := &dto.UserResponseData{ID: uuid.MustParse(id), Name: "Found", Email: "found@example.com"}

	tests := []struct {
		name       string
		pathID     string
		mockSetup  func(m *mockService)
		wantStatus int
		check      func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:       "missing id path value returns 400 without calling service",
			pathID:     "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid id returns 400",
			pathID: "not-a-uuid",
			mockSetup: func(m *mockService) {
				m.On("GetByID", mock.Anything, "not-a-uuid").Return(nil, errorsx.ErrInvalidID)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "not found returns 404",
			pathID: id,
			mockSetup: func(m *mockService) {
				m.On("GetByID", mock.Anything, id).Return(nil, errorsx.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "unmapped service error returns 500",
			pathID: id,
			mockSetup: func(m *mockService) {
				m.On("GetByID", mock.Anything, id).Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "success returns 200 with user",
			pathID: id,
			mockSetup: func(m *mockService) {
				m.On("GetByID", mock.Anything, id).Return(validData, nil)
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeSuccess(t, rec)
				assert.NotNil(t, body.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockService)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}
			h := NewHandler(svc, testLogger())

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.pathID, nil)
			if tt.pathID != "" {
				req.SetPathValue("id", tt.pathID)
			}
			rec := serve(t, h.GetUserByID, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.check != nil {
				tt.check(t, rec)
			}
			svc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetUsers(t *testing.T) {
	tests := []struct {
		name       string
		mockSetup  func(m *mockService)
		wantStatus int
		check      func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "service error returns 500",
			mockSetup: func(m *mockService) {
				m.On("GetMany", mock.Anything).Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success returns 200 with list",
			mockSetup: func(m *mockService) {
				m.On("GetMany", mock.Anything).Return([]dto.UserResponseData{
					{ID: uuid.New(), Name: "A"},
					{ID: uuid.New(), Name: "B"},
				}, nil)
			},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				body := decodeSuccess(t, rec)
				list, ok := body.Data.([]any)
				require.True(t, ok)
				assert.Len(t, list, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockService)
			tt.mockSetup(svc)
			h := NewHandler(svc, testLogger())

			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			rec := serve(t, h.GetUsers, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.check != nil {
				tt.check(t, rec)
			}
			svc.AssertExpectations(t)
		})
	}
}
