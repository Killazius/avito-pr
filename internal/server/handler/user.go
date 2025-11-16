package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Killazius/avito-pr/internal/models"
	userservice "github.com/Killazius/avito-pr/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserService interface {
	UpdateUserStatus(ctx context.Context, userID string, isActive bool) (*models.User, error)
	GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequestShort, error)
}

type UserHandler struct {
	s   UserService
	log *zap.Logger
}

func NewHandler(s UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{
		s:   s,
		log: log,
	}
}

func (h *UserHandler) RegisterRoutes(r gin.IRouter) {
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

func (h *UserHandler) SetUserStatus(c *gin.Context) {
	var req SetUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid request payload for Set User Status",
			zap.Error(err),
			zap.String("method", "SetUserStatus"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	h.log.Info("updating user status",
		zap.String("user_id", req.UserID),
		zap.Bool("is_active", req.IsActive))

	user, err := h.s.UpdateUserStatus(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, userservice.ErrUserNotFound) {
			h.log.Warn("user not found",
				zap.String("user_id", req.UserID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "resource not found",
				}})
			return
		}
		h.log.Error("internal server error while updating user status",
			zap.String("user_id", req.UserID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorInternal,
				Message: "Internal server error",
			}})
		return
	}

	h.log.Info("user status updated successfully",
		zap.String("user_id", req.UserID),
		zap.Bool("is_active", req.IsActive))
	c.JSON(http.StatusOK, gin.H{"user": user})
}

type GetReviewsRequest struct {
	UserID string `form:"user_id" binding:"required"`
}

func (h *UserHandler) GetReview(c *gin.Context) {
	var req GetReviewsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Warn("invalid request payload for Get Review",
			zap.Error(err),
			zap.String("method", "GetReview"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	h.log.Info("getting user reviews",
		zap.String("user_id", req.UserID))

	reviews, err := h.s.GetUserReviews(c.Request.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, userservice.ErrUserNotFound) {
			h.log.Warn("user not found",
				zap.String("user_id", req.UserID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "resource not found",
				}})
			return
		}
		h.log.Error("internal server error while getting user reviews",
			zap.String("user_id", req.UserID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorInternal,
				Message: "Internal server error",
			}})
		return
	}

	h.log.Info("user reviews retrieved successfully",
		zap.String("user_id", req.UserID),
		zap.Int("reviews_count", len(reviews)))
	c.JSON(http.StatusOK, gin.H{
		"user_id":       req.UserID,
		"pull_requests": reviews,
	})
}
