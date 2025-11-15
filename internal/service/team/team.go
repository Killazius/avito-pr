package team

import (
	"context"
	"errors"
	"fmt"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/Killazius/avito-pr/internal/repository/postgres"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, teamName string) error
	TeamExists(ctx context.Context, name string) (bool, error)
	GetTeamWithMembers(ctx context.Context, name string) (*models.Team, error)
}

type UserRepository interface {
	CreateOrUpdateUser(ctx context.Context, user models.TeamMember, teamID string) error
	UpdateIsActiveStatus(ctx context.Context, userID string, isActive bool) error
}
type Service struct {
	teamRepo  TeamRepository
	userRepo  UserRepository
	trManager *manager.Manager
}

func NewService(teamRepo TeamRepository, userRepo UserRepository, trManager *manager.Manager) *Service {
	return &Service{
		teamRepo:  teamRepo,
		userRepo:  userRepo,
		trManager: trManager,
	}
}

var ErrTeamExists = fmt.Errorf("team already exists")
var ErrTeamNotFound = fmt.Errorf("team not found")

func (s *Service) CreateTeamWithMembers(ctx context.Context, teamName string, members []models.TeamMember) (*models.Team, error) {
	if teamName == "" {
		return nil, fmt.Errorf("team name cannot be empty")
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("team must have at least one member")
	}

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		exists, err := s.teamRepo.TeamExists(ctx, teamName)
		if err != nil {
			return fmt.Errorf("failed to check team existence: %w", err)
		}
		if exists {
			return ErrTeamExists
		}

		if err := s.teamRepo.CreateTeam(ctx, teamName); err != nil {
			return fmt.Errorf("failed to create team: %w", err)
		}

		for i, member := range members {
			if member.UserID == "" || member.Username == "" {
				return fmt.Errorf("member %d has empty user_id or username", i)
			}

			if err := s.userRepo.CreateOrUpdateUser(ctx, member, teamName); err != nil {
				return fmt.Errorf("failed to create user %s: %w", member.UserID, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.teamRepo.GetTeamWithMembers(ctx, teamName)
}

func (s *Service) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	if teamName == "" {
		return nil, fmt.Errorf("team name cannot be empty")
	}
	team, err := s.teamRepo.GetTeamWithMembers(ctx, teamName)
	if err != nil {
		if errors.Is(err, postgres.ErrTeamNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team with members: %w", err)
	}
	return team, nil
}
