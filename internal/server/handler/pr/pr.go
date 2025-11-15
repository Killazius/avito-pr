package pr

import (
	"context"
	"errors"
	"net/http"

	"github.com/Killazius/avito-pr/internal/models"
	prservice "github.com/Killazius/avito-pr/internal/service/pr"
	"github.com/gin-gonic/gin"
)

type Service interface {
	CreatePR(ctx context.Context, prID string, prName string, authorID string) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, pullRequestID, oldUserID string) (*models.PullRequest, string, error)
}

type Handler struct {
	s Service
}

func NewHandler(s Service) *Handler {
	return &Handler{s: s}
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}
	pr, err := h.s.CreatePR(c.Request.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		switch {
		case errors.Is(err, prservice.ErrPRExists):
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorPRExists,
					Message: "PR already exists",
				}})
		case errors.Is(err, prservice.ErrUserNotFound):
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "Author not found",
				}})
		case errors.Is(err, prservice.ErrTeamNotFound):
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "Team not found",
				}})
		default:
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

func (h *Handler) Merge(c *gin.Context) {
	var req MergePRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}
	pr, err := h.s.MergePR(c.Request.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, prservice.ErrPRNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "resource not found",
				}})
			return
		}
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

func (h *Handler) Reassign(c *gin.Context) {
	var req ReassignPRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	pr, newReviewerID, err := h.s.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		switch {
		case errors.Is(err, prservice.ErrPRNotFound):
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "PR not found",
				},
			})
		case errors.Is(err, prservice.ErrUserNotFound):
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "User not found",
				},
			})
		case errors.Is(err, prservice.ErrPRMerged):
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorPRMerged,
					Message: "cannot reassign on merged PR",
				},
			})
		case errors.Is(err, prservice.ErrNotAssigned):
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotAssigned,
					Message: "reviewer is not assigned to this PR",
				},
			})
		case errors.Is(err, prservice.ErrNoCandidate):
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNoCandidate,
					Message: "no active replacement candidate in team",
				},
			})
		default:
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
