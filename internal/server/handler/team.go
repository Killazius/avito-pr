package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/Killazius/avito-pr/internal/server"
	teamservice "github.com/Killazius/avito-pr/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TeamService interface {
	CreateTeamWithMembers(ctx context.Context, teamName string, members []models.TeamMember) (*models.Team, error)
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
}

type TeamHandler struct {
	s TeamService
}

func NewTeamHandler(s TeamService) *TeamHandler {
	return &TeamHandler{
		s: s,
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
	log := server.GetLoggerFromContext(c)
	var req CreateTeamRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid request payload for Create Team",
			zap.Error(err),
			zap.String("method", "CreateTeam"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	log.Info("creating team",
		zap.String("team_name", req.TeamName),
		zap.Int("members_count", len(req.Members)))

	team, err := h.s.CreateTeamWithMembers(c.Request.Context(), req.TeamName, req.Members)
	if err != nil {
		if errors.Is(err, teamservice.ErrTeamExists) {
			log.Warn("team already exists",
				zap.String("team_name", req.TeamName),
				zap.Error(err))
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorTeamExists,
					Message: "team_name already exists",
				}})
			return
		}
		log.Error("internal server error while creating team",
			zap.String("team_name", req.TeamName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorInternal,
				Message: "Internal server error",
			}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"team": team,
	})
}

type GetTeamRequest struct {
	TeamName string `form:"team_name" binding:"required"`
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	log := server.GetLoggerFromContext(c)
	var req GetTeamRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Warn("invalid request payload for Get Team",
			zap.Error(err),
			zap.String("method", "GetTeam"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}

	log.Info("getting team",
		zap.String("team_name", req.TeamName))

	team, err := h.s.GetTeam(c.Request.Context(), req.TeamName)
	if err != nil {
		if errors.Is(err, teamservice.ErrTeamNotFound) {
			log.Warn("team not found",
				zap.String("team_name", req.TeamName),
				zap.Error(err))
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorNotFound,
					Message: "resource not found",
				}})
			return
		}
		log.Error("internal server error while getting team",
			zap.String("team_name", req.TeamName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorInternal,
				Message: "Internal server error",
			}})
		return
	}

	c.JSON(http.StatusOK, team)
}
