package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Killazius/avito-pr/internal/models"
	teamservice "github.com/Killazius/avito-pr/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TeamService interface {
	CreateTeamWithMembers(ctx context.Context, teamName string, members []models.TeamMember) (*models.Team, error)
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
}

type TeamHandler struct {
	s   TeamService
	log *zap.Logger
}

func NewTeamHandler(s TeamService, log *zap.Logger) *TeamHandler {
	return &TeamHandler{
		s:   s,
		log: log,
	}
}

type CreateTeamRequest struct {
	TeamName string              `json:"team_name" binding:"required"`
	Members  []models.TeamMember `json:"members" binding:"required"`
}

func (h *TeamHandler) RegisterRoutes(r gin.IRouter) {
	teamGroup := r.Group("/team")
	{
		teamGroup.POST("/add", h.CreateTeam)
		teamGroup.GET("/get", h.GetTeam)
	}
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req CreateTeamRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid request payload for Create Team",
			zap.Error(err),
			zap.String("method", "CreateTeam"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	h.log.Info("creating team",
		zap.String("team_name", req.TeamName),
		zap.Int("members_count", len(req.Members)))

	team, err := h.s.CreateTeamWithMembers(c.Request.Context(), req.TeamName, req.Members)
	if err != nil {
		if errors.Is(err, teamservice.ErrTeamExists) {
			h.log.Warn("team already exists",
				zap.String("team_name", req.TeamName),
				zap.Error(err))
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorTeamExists,
					Message: "team_name already exists",
				}})
			return
		}
		h.log.Error("internal server error while creating team",
			zap.String("team_name", req.TeamName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorInternal,
				Message: "Internal server error",
			}})
		return
	}

	h.log.Info("team created successfully",
		zap.String("team_name", req.TeamName))
	c.JSON(http.StatusCreated, gin.H{
		"team": team,
	})
}

type GetTeamRequest struct {
	TeamName string `form:"team_name" binding:"required"`
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	var req GetTeamRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Warn("invalid request payload for Get Team",
			zap.Error(err),
			zap.String("method", "GetTeam"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	h.log.Info("getting team",
		zap.String("team_name", req.TeamName))

	team, err := h.s.GetTeam(c.Request.Context(), req.TeamName)
	if err != nil {
		if errors.Is(err, teamservice.ErrTeamNotFound) {
			h.log.Warn("team not found",
				zap.String("team_name", req.TeamName),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "resource not found",
				}})
			return
		}
		h.log.Error("internal server error while getting team",
			zap.String("team_name", req.TeamName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorInternal,
				Message: "Internal server error",
			}})
		return
	}

	h.log.Info("team retrieved successfully",
		zap.String("team_name", req.TeamName))
	c.JSON(http.StatusOK, team)
}
