package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/Killazius/avito-pr/internal/models"
	userservice "github.com/Killazius/avito-pr/internal/service/user"
	"github.com/gin-gonic/gin"
)

type Service interface {
	UpdateUserStatus(ctx context.Context, userID string, isActive bool) (*models.User, error)
	GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequestShort, error)
}

type Handler struct {
	s Service
}

func NewHandler(s Service) *Handler {
	return &Handler{s: s}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	usersGroup := r.Group("/users")
	{
		usersGroup.POST("/setIsActive", h.SetUserStatus)
		usersGroup.GET("/getReview", h.GetReview)
	}
}

type SetUserStatusRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}

func (h *Handler) SetUserStatus(c *gin.Context) {
	var req SetUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	user, err := h.s.UpdateUserStatus(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, userservice.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

type GetReviewsRequest struct {
	UserID string `form:"user_id" binding:"required"`
}

func (h *Handler) GetReview(c *gin.Context) {
	var req GetReviewsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	reviews, err := h.s.GetUserReviews(c.Request.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, userservice.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user reviews"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":       req.UserID,
		"pull_requests": reviews,
	})

}
