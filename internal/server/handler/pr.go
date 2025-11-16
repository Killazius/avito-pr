package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/Killazius/avito-pr/internal/server"
	prservice "github.com/Killazius/avito-pr/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PRService interface {
	CreatePR(ctx context.Context, prID string, prName string, authorID string) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, pullRequestID, oldUserID string) (*models.PullRequest, string, error)
}

type PRHandler struct {
	s PRService
}

func NewPRHandler(s PRService) *PRHandler {
	return &PRHandler{
		s: s,
	}
}

func (h *PRHandler) RegisterRoutes(r gin.IRouter) {
	prGroup := r.Group("/pullRequest")
	{
		prGroup.POST("/create", h.Create)
		prGroup.POST("/merge", h.Merge)
		prGroup.POST("/reassign", h.Reassign)
	}
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

func (h *PRHandler) Create(c *gin.Context) {
	log := server.GetLoggerFromContext(c)
	var req CreatePRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid request payload for Create PR",
			zap.Error(err),
			zap.String("method", "Create"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	log.Info("creating PR",
		zap.String("pull_request_id", req.PullRequestID),
		zap.String("pull_request_name", req.PullRequestName),
		zap.String("author_id", req.AuthorID))

	pr, err := h.s.CreatePR(c.Request.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		switch {
		case errors.Is(err, prservice.ErrPRExists):
			log.Warn("PR already exists",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorPRExists,
					Message: "PR already exists",
				}})
		case errors.Is(err, prservice.ErrUserNotFound):
			log.Warn("author not found",
				zap.String("author_id", req.AuthorID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "Author not found",
				}})
		case errors.Is(err, prservice.ErrTeamNotFound):
			log.Warn("team not found",
				zap.String("author_id", req.AuthorID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "Team not found",
				}})
		case errors.Is(err, prservice.ErrUserInactive):
			log.Warn("author is not active",
				zap.String("author_id", req.AuthorID),
				zap.Error(err))
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorBadRequest,
					Message: "Author is not active",
				}})
		default:
			log.Error("internal server error while creating PR",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorInternal,
					Message: "Internal server error",
				},
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pr": pr})
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

func (h *PRHandler) Merge(c *gin.Context) {
	log := server.GetLoggerFromContext(c)
	var req MergePRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid request payload for Merge PR",
			zap.Error(err),
			zap.String("method", "Merge"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	log.Info("merging PR",
		zap.String("pull_request_id", req.PullRequestID))

	pr, err := h.s.MergePR(c.Request.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, prservice.ErrPRNotFound) {
			log.Warn("PR not found",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "resource not found",
				}})
			return
		}
		log.Error("internal server error while merging PR",
			zap.String("pull_request_id", req.PullRequestID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorInternal,
				Message: "Internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": pr})
}

type ReassignPRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldReviewerID string `json:"old_reviewer_id" binding:"required"`
}

func (h *PRHandler) Reassign(c *gin.Context) {
	log := server.GetLoggerFromContext(c)
	var req ReassignPRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid request payload for Reassign PR",
			zap.Error(err),
			zap.String("method", "Reassign"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	log.Info("reassigning reviewer",
		zap.String("pull_request_id", req.PullRequestID),
		zap.String("old_reviewer_id", req.OldReviewerID))

	pr, newReviewerID, err := h.s.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		switch {
		case errors.Is(err, prservice.ErrPRNotFound):
			log.Warn("PR not found",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "PR not found",
				},
			})
		case errors.Is(err, prservice.ErrUserNotFound):
			log.Warn("user not found",
				zap.String("old_reviewer_id", req.OldReviewerID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "User not found",
				},
			})
		case errors.Is(err, prservice.ErrPRMerged):
			log.Warn("cannot reassign on merged PR",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorPRMerged,
					Message: "cannot reassign on merged PR",
				},
			})
		case errors.Is(err, prservice.ErrNotAssigned):
			log.Warn("reviewer is not assigned to this PR",
				zap.String("pull_request_id", req.PullRequestID),
				zap.String("old_reviewer_id", req.OldReviewerID),
				zap.Error(err))
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotAssigned,
					Message: "reviewer is not assigned to this PR",
				},
			})
		case errors.Is(err, prservice.ErrNoCandidate):
			log.Warn("no active replacement candidate in team",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNoCandidate,
					Message: "no active replacement candidate in team",
				},
			})
		default:
			log.Error("internal server error while reassigning reviewer",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorInternal,
					Message: "Internal server error",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}
