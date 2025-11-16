package service

import (
	"context"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type StatsService struct {
	tr   *manager.Manager
	repo StatsRepository
}

func NewStatsService(repo StatsRepository, tr *manager.Manager) *StatsService {
	return &StatsService{
		tr:   tr,
		repo: repo,
	}
}

func (s *StatsService) GetStats(ctx context.Context) (*models.Stats, error) {
	var stats *models.Stats

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		teamsCount, err := s.repo.GetTeamsCount(ctx)
		if err != nil {
			return err
		}

		usersStats, err := s.repo.GetUsersStats(ctx)
		if err != nil {
			return err
		}

		prStats, err := s.repo.GetPRStats(ctx)
		if err != nil {
			return err
		}

		assignmentsCount, err := s.repo.GetAssignmentsCount(ctx)
		if err != nil {
			return err
		}

		stats = &models.Stats{
			TotalTeams:       teamsCount,
			TotalUsers:       usersStats.TotalUsers,
			ActiveUsers:      usersStats.ActiveUsers,
			TotalPRs:         prStats.TotalPRs,
			OpenPRs:          prStats.OpenPRs,
			MergedPRs:        prStats.MergedPRs,
			TotalAssignments: assignmentsCount,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return stats, nil
}
