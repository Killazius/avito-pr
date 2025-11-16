package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/Killazius/avito-pr/internal/service"
	"github.com/Killazius/avito-pr/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUserHandler_SetUserStatus(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockUserService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - activate user",
			requestBody: SetUserStatusRequest{
				UserID:   "u1",
				IsActive: true,
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.On("UpdateUserStatus", mock.Anything, "u1", true).Return(&models.User{
					UserID:   "u1",
					Username: "Alice",
					TeamName: "backend",
					IsActive: true,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "user")
				user, ok := response["user"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "u1", user["user_id"])
				assert.Equal(t, "Alice", user["username"])
				assert.Equal(t, "backend", user["team_name"])
				assert.Equal(t, true, user["is_active"])
			},
		},
		{
			name: "success - deactivate user",
			requestBody: SetUserStatusRequest{
				UserID:   "u2",
				IsActive: false,
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.On("UpdateUserStatus", mock.Anything, "u2", false).Return(&models.User{
					UserID:   "u2",
					Username: "Bob",
					TeamName: "backend",
					IsActive: false,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "user")
				user, ok := response["user"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "u2", user["user_id"])
				assert.Equal(t, false, user["is_active"])
			},
		},
		{
			name: "invalid request - missing user_id",
			requestBody: map[string]interface{}{
				"is_active": true,
			},
			mockSetup:      func(_ *mocks.MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, models.ErrorBadRequest, response.Error.Code)
				assert.Equal(t, "Invalid request payload", response.Error.Message)
			},
		},
		{
			name: "user not found",
			requestBody: SetUserStatusRequest{
				UserID:   "u999",
				IsActive: true,
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.On("UpdateUserStatus", mock.Anything, "u999", true).Return((*models.User)(nil), service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, models.ErrorNotFound, response.Error.Code)
				assert.Equal(t, "resource not found", response.Error.Message)
			},
		},
		{
			name: "internal server error",
			requestBody: SetUserStatusRequest{
				UserID:   "u1",
				IsActive: true,
			},
			mockSetup: func(m *mocks.MockUserService) {
				m.On("UpdateUserStatus", mock.Anything, "u1", true).Return((*models.User)(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, models.ErrorInternal, response.Error.Code)
				assert.Equal(t, "Internal server error", response.Error.Message)
			},
		},
		{
			name:           "invalid json",
			requestBody:    "invalid json",
			mockSetup:      func(_ *mocks.MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, models.ErrorBadRequest, response.Error.Code)
				assert.Equal(t, "Invalid request payload", response.Error.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockService := new(mocks.MockUserService)
			tt.mockSetup(mockService)

			handler := NewHandler(mockService)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				logger := zap.NewNop()
				c.Set("logger", logger)
				c.Next()
			})
			handler.RegisterRoutes(router)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetReview(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*mocks.MockUserService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "success - with reviews",
			queryParams: "user_id=u2",
			mockSetup: func(m *mocks.MockUserService) {
				m.On("GetUserReviews", mock.Anything, "u2").Return([]*models.PullRequestShort{
					{
						PullRequestID:   "pr-1001",
						PullRequestName: "Add search",
						AuthorID:        "u1",
						Status:          models.PRStatusOpen,
					},
					{
						PullRequestID:   "pr-1002",
						PullRequestName: "Fix bug",
						AuthorID:        "u3",
						Status:          models.PRStatusMerged,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "u2", response["user_id"])
				assert.Contains(t, response, "pull_requests")
				prs, ok := response["pull_requests"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, prs, 2)

				pr1, ok := prs[0].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "pr-1001", pr1["pull_request_id"])
				assert.Equal(t, "Add search", pr1["pull_request_name"])
				assert.Equal(t, "u1", pr1["author_id"])
				assert.Equal(t, models.PRStatusOpen, pr1["status"])

				pr2, ok := prs[1].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "pr-1002", pr2["pull_request_id"])
				assert.Equal(t, "Fix bug", pr2["pull_request_name"])
				assert.Equal(t, "u3", pr2["author_id"])
				assert.Equal(t, models.PRStatusMerged, pr2["status"])
			},
		},
		{
			name:        "success - no reviews",
			queryParams: "user_id=u3",
			mockSetup: func(m *mocks.MockUserService) {
				m.On("GetUserReviews", mock.Anything, "u3").Return([]*models.PullRequestShort{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "u3", response["user_id"])
				assert.Contains(t, response, "pull_requests")
				prs, ok := response["pull_requests"].([]interface{})
				assert.True(t, ok)
				assert.Empty(t, prs)
			},
		},
		{
			name:           "missing user_id parameter",
			queryParams:    "",
			mockSetup:      func(_ *mocks.MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, models.ErrorBadRequest, response.Error.Code)
				assert.Equal(t, "Invalid request payload", response.Error.Message)
			},
		},
		{
			name:        "user not found",
			queryParams: "user_id=u999",
			mockSetup: func(m *mocks.MockUserService) {
				m.On("GetUserReviews", mock.Anything, "u999").Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, models.ErrorNotFound, response.Error.Code)
				assert.Equal(t, "resource not found", response.Error.Message)
			},
		},
		{
			name:        "internal server error",
			queryParams: "user_id=u1",
			mockSetup: func(m *mocks.MockUserService) {
				m.On("GetUserReviews", mock.Anything, "u1").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, models.ErrorInternal, response.Error.Code)
				assert.Equal(t, "Internal server error", response.Error.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockService := new(mocks.MockUserService)
			tt.mockSetup(mockService)

			handler := NewHandler(mockService)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				logger := zap.NewNop()
				c.Set("logger", logger)
				c.Next()
			})
			handler.RegisterRoutes(router)

			url := "/users/getReview"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req := httptest.NewRequest(http.MethodGet, url, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
		})
	}
}

func TestNewUserHandler(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.MockUserService)
	handler := NewHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.s)
}
