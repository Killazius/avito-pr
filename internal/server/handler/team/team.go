package team

import (
	"context"
	"errors"
	"net/http"

	"github.com/Killazius/avito-pr/internal/models"
	teamservice "github.com/Killazius/avito-pr/internal/service/team"
	"github.com/gin-gonic/gin"
)

type Service interface {
	CreateTeamWithMembers(ctx context.Context, teamName string, members []models.TeamMember) (*models.Team, error)
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
}

type Handler struct {
	s Service
}

func NewHandler(s Service) *Handler {
	return &Handler{s: s}
}

type CreateTeamRequest struct {
	TeamName string              `json:"team_name" binding:"required"`
	Members  []models.TeamMember `json:"members" binding:"required"`
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	teamGroup := r.Group("/team")
	{
		teamGroup.POST("/add", h.CreateTeam)
		teamGroup.GET("/get", h.GetTeam)
	}
}

func (h *Handler) CreateTeam(c *gin.Context) {
	var req CreateTeamRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}
	team, err := h.s.CreateTeamWithMembers(c.Request.Context(), req.TeamName, req.Members)
	if err != nil {
		if errors.Is(err, teamservice.ErrTeamExists) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.ErrorBody{
					Code:    models.ErrorTeamExists,
					Message: "team_name already exists",
				}})
			return
		}
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

func (h *Handler) GetTeam(c *gin.Context) {
	var req GetTeamRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorBody{
				Code:    models.ErrorBadRequest,
				Message: "Invalid request payload",
			}})
		return
	}
	team, err := h.s.GetTeam(c.Request.Context(), req.TeamName)
	if err != nil {
		if errors.Is(err, teamservice.ErrTeamNotFound) {
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
			}})
		return
	}
	c.JSON(http.StatusOK, team)
}
