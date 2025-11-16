package handler

import (
	"context"
	"net/http"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/Killazius/avito-pr/internal/server"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StatsService interface {
	GetStats(ctx context.Context) (*models.Stats, error)
}
type StatsHandler struct {
	s StatsService
}

func NewStatsHandler(s StatsService) *StatsHandler {
	return &StatsHandler{s: s}
}

func (h *StatsHandler) RegisterRoutes(r gin.IRouter) {
	r.GET("/stats", h.GetStats)
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	log := server.GetLoggerFromContext(c)
	stats, err := h.s.GetStats(c.Request.Context())
	if err != nil {
		log.Warn("Failed to get stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorInternal,
				Message: "Internal server error",
			}})
		return
	}
	c.JSON(http.StatusOK, stats)
}
