package postgres

import (
	"context"

	"github.com/Killazius/avito-pr/internal/models"
)

func (r *Repository) GetTeamsCount(ctx context.Context) (int, error) {
	conn := r.getConn(ctx)
	query := `SELECT COUNT(*) FROM teams`

	var count int

	err := conn.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetUsersStats(ctx context.Context) (*models.UsersStats, error) {
	conn := r.getConn(ctx)
	query := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_users
		FROM users
	`

	var totalUsers, activeUsers int

	err := conn.QueryRow(ctx, query).Scan(&totalUsers, &activeUsers)
	if err != nil {
		return nil, err
	}

	return &models.UsersStats{
		TotalUsers:  totalUsers,
		ActiveUsers: activeUsers,
	}, nil
}

func (r *Repository) GetPRStats(ctx context.Context) (*models.PRStats, error) {
	conn := r.getConn(ctx)
	query := `
		SELECT 
			COUNT(*) as total_prs,
			COUNT(CASE WHEN status = 'OPEN' THEN 1 END) as open_prs,
			COUNT(CASE WHEN status = 'MERGED' THEN 1 END) as merged_prs
		FROM pull_requests
	`

	var totalPRs, openPRs, mergedPRs int

	err := conn.QueryRow(ctx, query).Scan(&totalPRs, &openPRs, &mergedPRs)
	if err != nil {
		return nil, err
	}

	return &models.PRStats{
		TotalPRs:  totalPRs,
		OpenPRs:   openPRs,
		MergedPRs: mergedPRs,
	}, nil
}

func (r *Repository) GetAssignmentsCount(ctx context.Context) (int, error) {
	conn := r.getConn(ctx)
	query := `SELECT COUNT(*) FROM pr_reviewers`

	var count int

	err := conn.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
