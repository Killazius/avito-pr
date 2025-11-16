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

func TestTeamHandler_CreateTeam(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockTeamService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			requestBody: CreateTeamRequest{
				TeamName: "backend",
				Members: []models.TeamMember{
					{UserID: "u1", Username: "Alice", IsActive: true},
					{UserID: "u2", Username: "Bob", IsActive: true},
				},
			},
			mockSetup: func(m *mocks.MockTeamService) {
				m.On("CreateTeamWithMembers", mock.Anything, "backend", []models.TeamMember{
					{UserID: "u1", Username: "Alice", IsActive: true},
					{UserID: "u2", Username: "Bob", IsActive: true},
				}).Return(&models.Team{
					TeamName: "backend",
					Members: []models.TeamMember{
						{UserID: "u1", Username: "Alice", IsActive: true},
						{UserID: "u2", Username: "Bob", IsActive: true},
					},
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "team")
				team, ok := response["team"].(map[string]interface{})
				assert.True(t, ok, "team should be a map")
				assert.Equal(t, "backend", team["team_name"])
				assert.NotNil(t, team["members"])
				members, ok := team["members"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, members, 2)
			},
		},
		{
			name: "invalid request - missing team_name",
			requestBody: map[string]interface{}{
				"members": []models.TeamMember{
					{UserID: "u1", Username: "Alice", IsActive: true},
				},
			},
			mockSetup:      func(_ *mocks.MockTeamService) {},
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
			name: "invalid request - missing members",
			requestBody: map[string]interface{}{
				"team_name": "backend",
			},
			mockSetup:      func(_ *mocks.MockTeamService) {},
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
			name: "team already exists",
			requestBody: CreateTeamRequest{
				TeamName: "payments",
				Members: []models.TeamMember{
					{UserID: "u1", Username: "Alice", IsActive: true},
				},
			},
			mockSetup: func(m *mocks.MockTeamService) {
				m.On("CreateTeamWithMembers", mock.Anything, "payments", []models.TeamMember{
					{UserID: "u1", Username: "Alice", IsActive: true},
				}).Return((*models.Team)(nil), service.ErrTeamExists)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var response models.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, models.ErrorTeamExists, response.Error.Code)
				assert.Equal(t, "team_name already exists", response.Error.Message)
			},
		},
		{
			name: "internal server error",
			requestBody: CreateTeamRequest{
				TeamName: "backend",
				Members: []models.TeamMember{
					{UserID: "u1", Username: "Alice", IsActive: true},
				},
			},
			mockSetup: func(m *mocks.MockTeamService) {
				m.On("CreateTeamWithMembers", mock.Anything, "backend", []models.TeamMember{
					{UserID: "u1", Username: "Alice", IsActive: true},
				}).Return((*models.Team)(nil), errors.New("database error"))
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
			mockSetup:      func(_ *mocks.MockTeamService) {},
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
			mockService := new(mocks.MockTeamService)
			tt.mockSetup(mockService)

			handler := NewTeamHandler(mockService)

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

			req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
		})
	}
}

func TestTeamHandler_GetTeam(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*mocks.MockTeamService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "success",
			queryParams: "team_name=backend",
			mockSetup: func(m *mocks.MockTeamService) {
				m.On("GetTeam", mock.Anything, "backend").Return(&models.Team{
					TeamName: "backend",
					Members: []models.TeamMember{
						{UserID: "u1", Username: "Alice", IsActive: true},
						{UserID: "u2", Username: "Bob", IsActive: false},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				var team models.Team
				err := json.Unmarshal(rec.Body.Bytes(), &team)
				require.NoError(t, err)
				assert.Equal(t, "backend", team.TeamName)
				assert.Len(t, team.Members, 2)
				assert.Equal(t, "u1", team.Members[0].UserID)
				assert.Equal(t, "Alice", team.Members[0].Username)
				assert.True(t, team.Members[0].IsActive)
				assert.Equal(t, "u2", team.Members[1].UserID)
				assert.Equal(t, "Bob", team.Members[1].Username)
				assert.False(t, team.Members[1].IsActive)
			},
		},
		{
			name:           "missing team_name parameter",
			queryParams:    "",
			mockSetup:      func(_ *mocks.MockTeamService) {},
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
			name:        "team not found",
			queryParams: "team_name=nonexistent",
			mockSetup: func(m *mocks.MockTeamService) {
				m.On("GetTeam", mock.Anything, "nonexistent").Return((*models.Team)(nil), service.ErrTeamNotFound)
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
			queryParams: "team_name=backend",
			mockSetup: func(m *mocks.MockTeamService) {
				m.On("GetTeam", mock.Anything, "backend").Return((*models.Team)(nil), errors.New("database error"))
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
			mockService := new(mocks.MockTeamService)
			tt.mockSetup(mockService)

			handler := NewTeamHandler(mockService)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				logger := zap.NewNop()
				c.Set("logger", logger)
				c.Next()
			})
			handler.RegisterRoutes(router)

			url := "/team/get"
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

func TestNewTeamHandler(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.MockTeamService)
	handler := NewTeamHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.s)
}
