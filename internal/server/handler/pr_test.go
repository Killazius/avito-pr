package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/Killazius/avito-pr/internal/service"
	"github.com/Killazius/avito-pr/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestPRHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockPRService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			requestBody: CreatePRRequest{
				PullRequestID:   "pr-1001",
				PullRequestName: "Add search",
				AuthorID:        "u1",
			},
			mockSetup: func(m *mocks.MockPRService) {
				now := time.Now()
				m.On("CreatePR", mock.Anything, "pr-1001", "Add search", "u1").Return(&models.PullRequest{
					PullRequestID:     "pr-1001",
					PullRequestName:   "Add search",
					AuthorID:          "u1",
					Status:            models.PRStatusOpen,
					AssignedReviewers: []string{"u2"},
					CreatedAt:         &now,
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "pr")
				pr := response["pr"].(map[string]interface{})
				assert.Equal(t, "pr-1001", pr["pull_request_id"])
				assert.Equal(t, "Add search", pr["pull_request_name"])
				assert.Equal(t, "u1", pr["author_id"])
				assert.Equal(t, models.PRStatusOpen, pr["status"])
			},
		},
		{
			name: "invalid request - missing pull_request_id",
			requestBody: map[string]interface{}{
				"pull_request_name": "Add search",
				"author_id":         "u1",
			},
			mockSetup:      func(m *mocks.MockPRService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorBadRequest, response.Error.Code)
				assert.Equal(t, "Invalid request payload", response.Error.Message)
			},
		},
		{
			name: "PR already exists",
			requestBody: CreatePRRequest{
				PullRequestID:   "pr-1001",
				PullRequestName: "Add search",
				AuthorID:        "u1",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("CreatePR", mock.Anything, "pr-1001", "Add search", "u1").Return((*models.PullRequest)(nil), service.ErrPRExists)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorPRExists, response.Error.Code)
				assert.Equal(t, "PR already exists", response.Error.Message)
			},
		},
		{
			name: "author not found",
			requestBody: CreatePRRequest{
				PullRequestID:   "pr-1001",
				PullRequestName: "Add search",
				AuthorID:        "u999",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("CreatePR", mock.Anything, "pr-1001", "Add search", "u999").Return((*models.PullRequest)(nil), service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorNotFound, response.Error.Code)
				assert.Equal(t, "Author not found", response.Error.Message)
			},
		},
		{
			name: "team not found",
			requestBody: CreatePRRequest{
				PullRequestID:   "pr-1001",
				PullRequestName: "Add search",
				AuthorID:        "u1",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("CreatePR", mock.Anything, "pr-1001", "Add search", "u1").Return((*models.PullRequest)(nil), service.ErrTeamNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorNotFound, response.Error.Code)
				assert.Equal(t, "Team not found", response.Error.Message)
			},
		},
		{
			name: "author is not active",
			requestBody: CreatePRRequest{
				PullRequestID:   "pr-1001",
				PullRequestName: "Add search",
				AuthorID:        "u1",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("CreatePR", mock.Anything, "pr-1001", "Add search", "u1").Return((*models.PullRequest)(nil), service.ErrUserInactive)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorBadRequest, response.Error.Code)
				assert.Equal(t, "Author is not active", response.Error.Message)
			},
		},
		{
			name: "internal server error",
			requestBody: CreatePRRequest{
				PullRequestID:   "pr-1001",
				PullRequestName: "Add search",
				AuthorID:        "u1",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("CreatePR", mock.Anything, "pr-1001", "Add search", "u1").Return((*models.PullRequest)(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorInternal, response.Error.Code)
				assert.Equal(t, "Internal server error", response.Error.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockPRService)
			tt.mockSetup(mockService)

			handler := NewPRHandler(mockService)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				logger := zap.NewNop()
				c.Set("logger", logger)
				c.Next()
			})
			handler.RegisterRoutes(router)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestPRHandler_Merge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockPRService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			requestBody: MergePRRequest{
				PullRequestID: "pr-1001",
			},
			mockSetup: func(m *mocks.MockPRService) {
				now := time.Now()
				m.On("MergePR", mock.Anything, "pr-1001").Return(&models.PullRequest{
					PullRequestID:     "pr-1001",
					PullRequestName:   "Add search",
					AuthorID:          "u1",
					Status:            models.PRStatusMerged,
					AssignedReviewers: []string{"u2"},
					MergedAt:          &now,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "pr")
				pr := response["pr"].(map[string]interface{})
				assert.Equal(t, "pr-1001", pr["pull_request_id"])
				assert.Equal(t, models.PRStatusMerged, pr["status"])
			},
		},
		{
			name: "invalid request - missing pull_request_id",
			requestBody: map[string]interface{}{
				"some_field": "value",
			},
			mockSetup:      func(m *mocks.MockPRService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorBadRequest, response.Error.Code)
				assert.Equal(t, "Invalid request payload", response.Error.Message)
			},
		},
		{
			name: "PR not found",
			requestBody: MergePRRequest{
				PullRequestID: "pr-9999",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("MergePR", mock.Anything, "pr-9999").Return((*models.PullRequest)(nil), service.ErrPRNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorNotFound, response.Error.Code)
				assert.Equal(t, "resource not found", response.Error.Message)
			},
		},
		{
			name: "internal server error",
			requestBody: MergePRRequest{
				PullRequestID: "pr-1001",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("MergePR", mock.Anything, "pr-1001").Return((*models.PullRequest)(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorInternal, response.Error.Code)
				assert.Equal(t, "Internal server error", response.Error.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockPRService)
			tt.mockSetup(mockService)

			handler := NewPRHandler(mockService)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				logger := zap.NewNop()
				c.Set("logger", logger)
				c.Next()
			})
			handler.RegisterRoutes(router)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestPRHandler_Reassign(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockPRService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			requestBody: ReassignPRRequest{
				PullRequestID: "pr-1001",
				OldReviewerID: "u2",
			},
			mockSetup: func(m *mocks.MockPRService) {
				now := time.Now()
				m.On("ReassignReviewer", mock.Anything, "pr-1001", "u2").Return(&models.PullRequest{
					PullRequestID:     "pr-1001",
					PullRequestName:   "Add search",
					AuthorID:          "u1",
					Status:            models.PRStatusOpen,
					AssignedReviewers: []string{"u3"},
					CreatedAt:         &now,
				}, "u3", nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "pr")
				assert.Contains(t, response, "replaced_by")
				assert.Equal(t, "u3", response["replaced_by"])
				pr := response["pr"].(map[string]interface{})
				assert.Equal(t, "pr-1001", pr["pull_request_id"])
			},
		},
		{
			name: "invalid request - missing pull_request_id",
			requestBody: map[string]interface{}{
				"old_reviewer_id": "u2",
			},
			mockSetup:      func(m *mocks.MockPRService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorBadRequest, response.Error.Code)
				assert.Equal(t, "Invalid request payload", response.Error.Message)
			},
		},
		{
			name: "invalid request - missing old_reviewer_id",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr-1001",
			},
			mockSetup:      func(m *mocks.MockPRService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorBadRequest, response.Error.Code)
				assert.Equal(t, "Invalid request payload", response.Error.Message)
			},
		},
		{
			name: "PR not found",
			requestBody: ReassignPRRequest{
				PullRequestID: "pr-9999",
				OldReviewerID: "u2",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("ReassignReviewer", mock.Anything, "pr-9999", "u2").Return((*models.PullRequest)(nil), "", service.ErrPRNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorNotFound, response.Error.Code)
				assert.Equal(t, "PR not found", response.Error.Message)
			},
		},
		{
			name: "user not found",
			requestBody: ReassignPRRequest{
				PullRequestID: "pr-1001",
				OldReviewerID: "u999",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("ReassignReviewer", mock.Anything, "pr-1001", "u999").Return((*models.PullRequest)(nil), "", service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorNotFound, response.Error.Code)
				assert.Equal(t, "User not found", response.Error.Message)
			},
		},
		{
			name: "PR already merged",
			requestBody: ReassignPRRequest{
				PullRequestID: "pr-1001",
				OldReviewerID: "u2",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("ReassignReviewer", mock.Anything, "pr-1001", "u2").Return((*models.PullRequest)(nil), "", service.ErrPRMerged)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorPRMerged, response.Error.Code)
				assert.Equal(t, "cannot reassign on merged PR", response.Error.Message)
			},
		},
		{
			name: "reviewer not assigned to PR",
			requestBody: ReassignPRRequest{
				PullRequestID: "pr-1001",
				OldReviewerID: "u5",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("ReassignReviewer", mock.Anything, "pr-1001", "u5").Return((*models.PullRequest)(nil), "", service.ErrNotAssigned)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorNotAssigned, response.Error.Code)
				assert.Equal(t, "reviewer is not assigned to this PR", response.Error.Message)
			},
		},
		{
			name: "no candidate for replacement",
			requestBody: ReassignPRRequest{
				PullRequestID: "pr-1001",
				OldReviewerID: "u2",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("ReassignReviewer", mock.Anything, "pr-1001", "u2").Return((*models.PullRequest)(nil), "", service.ErrNoCandidate)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorNoCandidate, response.Error.Code)
				assert.Equal(t, "no active replacement candidate in team", response.Error.Message)
			},
		},
		{
			name: "internal server error",
			requestBody: ReassignPRRequest{
				PullRequestID: "pr-1001",
				OldReviewerID: "u2",
			},
			mockSetup: func(m *mocks.MockPRService) {
				m.On("ReassignReviewer", mock.Anything, "pr-1001", "u2").Return((*models.PullRequest)(nil), "", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.ErrorInternal, response.Error.Code)
				assert.Equal(t, "Internal server error", response.Error.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockPRService)
			tt.mockSetup(mockService)

			handler := NewPRHandler(mockService)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				logger := zap.NewNop()
				c.Set("logger", logger)
				c.Next()
			})
			handler.RegisterRoutes(router)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			mockService.AssertExpectations(t)
		})
	}
}
