package pr

import (
	"context"
	"errors"
	"net/http"

	"github.com/Killazius/avito-pr/internal/models"
	prservice "github.com/Killazius/avito-pr/internal/service/pr"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Service interface {
	CreatePR(ctx context.Context, prID string, prName string, authorID string) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, pullRequestID, oldUserID string) (*models.PullRequest, string, error)
}

type Handler struct {
	s   Service
	log *zap.Logger
}

func NewHandler(s Service, log *zap.Logger) *Handler {
	return &Handler{
		s:   s,
		log: log,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
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

func (h *Handler) Create(c *gin.Context) {
	var req CreatePRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid request payload for Create PR",
			zap.Error(err),
			zap.String("method", "Create"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	h.log.Info("creating PR",
		zap.String("pull_request_id", req.PullRequestID),
		zap.String("pull_request_name", req.PullRequestName),
		zap.String("author_id", req.AuthorID))

	pr, err := h.s.CreatePR(c.Request.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		switch {
		case errors.Is(err, prservice.ErrPRExists):
			h.log.Warn("PR already exists",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorPRExists,
					Message: "PR already exists",
				}})
		case errors.Is(err, prservice.ErrUserNotFound):
			h.log.Warn("author not found",
				zap.String("author_id", req.AuthorID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "Author not found",
				}})
		case errors.Is(err, prservice.ErrTeamNotFound):
			h.log.Warn("team not found",
				zap.String("author_id", req.AuthorID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "Team not found",
				}})
		case errors.Is(err, prservice.ErrUserInactive):
			h.log.Warn("author is not active",
				zap.String("author_id", req.AuthorID),
				zap.Error(err))
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorBadRequest,
					Message: "Author is not active",
				}})
		default:
			h.log.Error("internal server error while creating PR",
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

	h.log.Info("PR created successfully",
		zap.String("pull_request_id", req.PullRequestID))
	c.JSON(http.StatusCreated, gin.H{"pr": pr})
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

func (h *Handler) Merge(c *gin.Context) {
	var req MergePRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid request payload for Merge PR",
			zap.Error(err),
			zap.String("method", "Merge"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	h.log.Info("merging PR",
		zap.String("pull_request_id", req.PullRequestID))

	pr, err := h.s.MergePR(c.Request.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, prservice.ErrPRNotFound) {
			h.log.Warn("PR not found",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "resource not found",
				}})
			return
		}
		h.log.Error("internal server error while merging PR",
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

	h.log.Info("PR merged successfully",
		zap.String("pull_request_id", req.PullRequestID))
	c.JSON(http.StatusOK, gin.H{"pr": pr})
}

type ReassignPRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldReviewerID string `json:"old_reviewer_id" binding:"required"`
}

func (h *Handler) Reassign(c *gin.Context) {
	var req ReassignPRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid request payload for Reassign PR",
			zap.Error(err),
			zap.String("method", "Reassign"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	h.log.Info("reassigning reviewer",
		zap.String("pull_request_id", req.PullRequestID),
		zap.String("old_reviewer_id", req.OldReviewerID))

	pr, newReviewerID, err := h.s.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		switch {
		case errors.Is(err, prservice.ErrPRNotFound):
			h.log.Warn("PR not found",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "PR not found",
				},
			})
		case errors.Is(err, prservice.ErrUserNotFound):
			h.log.Warn("user not found",
				zap.String("old_reviewer_id", req.OldReviewerID),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "User not found",
				},
			})
		case errors.Is(err, prservice.ErrPRMerged):
			h.log.Warn("cannot reassign on merged PR",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorPRMerged,
					Message: "cannot reassign on merged PR",
				},
			})
		case errors.Is(err, prservice.ErrNotAssigned):
			h.log.Warn("reviewer is not assigned to this PR",
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
			h.log.Warn("no active replacement candidate in team",
				zap.String("pull_request_id", req.PullRequestID),
				zap.Error(err))
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNoCandidate,
					Message: "no active replacement candidate in team",
				},
			})
		default:
			h.log.Error("internal server error while reassigning reviewer",
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

	h.log.Info("reviewer reassigned successfully",
		zap.String("pull_request_id", req.PullRequestID),
		zap.String("old_reviewer_id", req.OldReviewerID),
		zap.String("new_reviewer_id", newReviewerID))
	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}
