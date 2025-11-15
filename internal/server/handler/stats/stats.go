package stats

import (
	"context"
	"net/http"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/gin-gonic/gin"
)

type Service interface {
	GetStats(ctx context.Context) (*models.Stats, error)
}
type Handler struct {
	s Service
}

func NewHandler(s Service) *Handler {
	return &Handler{s: s}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/stats", h.GetStats)
}

func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.s.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorInternal,
				Message: "Internal server error",
			}})
		return
	}
	c.JSON(http.StatusOK, stats)
}
