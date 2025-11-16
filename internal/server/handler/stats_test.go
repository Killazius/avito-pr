package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/Killazius/avito-pr/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestStatsHandler_GetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockStatsService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			mockSetup: func(m *mocks.MockStatsService) {
				m.On("GetStats", mock.Anything).Return(&models.Stats{
					TotalTeams:       5,
					TotalUsers:       20,
					ActiveUsers:      15,
					TotalPRs:         100,
					OpenPRs:          30,
					MergedPRs:        70,
					TotalAssignments: 150,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"total_teams":5,"total_users":20,"active_users":15,"total_prs":100,"open_prs":30,"merged_prs":70,"total_assignments":150}`,
		},
		{
			name: "error from service",
			mockSetup: func(m *mocks.MockStatsService) {
				m.On("GetStats", mock.Anything).Return((*models.Stats)(nil), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":{"code":"INTERNAL","message":"Internal server error"}}`,
		},
		{
			name: "empty stats",
			mockSetup: func(m *mocks.MockStatsService) {
				m.On("GetStats", mock.Anything).Return(&models.Stats{
					TotalTeams:       0,
					TotalUsers:       0,
					ActiveUsers:      0,
					TotalPRs:         0,
					OpenPRs:          0,
					MergedPRs:        0,
					TotalAssignments: 0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"total_teams":0,"total_users":0,"active_users":0,"total_prs":0,"open_prs":0,"merged_prs":0,"total_assignments":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockStatsService)
			tt.mockSetup(mockService)

			handler := NewStatsHandler(mockService)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				logger := zap.NewNop()
				c.Set("logger", logger)
				c.Next()
			})
			handler.RegisterRoutes(router)

			req := httptest.NewRequest(http.MethodGet, "/stats", http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockService.AssertExpectations(t)
		})
	}
}

func TestNewStatsHandler(t *testing.T) {
	mockService := new(mocks.MockStatsService)
	handler := NewStatsHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.s)
}

func TestStatsHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(mocks.MockStatsService)
	handler := NewStatsHandler(mockService)

	router := gin.New()
	handler.RegisterRoutes(router)

	routes := router.Routes()
	assert.Len(t, routes, 1)
	assert.Equal(t, "GET", routes[0].Method)
	assert.Equal(t, "/stats", routes[0].Path)
}
