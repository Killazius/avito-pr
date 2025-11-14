package postgres

import (
	"context"
	"fmt"

	"github.com/Killazius/avito-pr/internal/models"
)

func (r *Repository) CreateTeam(ctx context.Context, teamName string) error {
	query := `INSERT INTO teams (name, created_at) VALUES ($1, NOW())`

	conn := r.getConn(ctx)
	_, err := conn.Exec(ctx, query, teamName)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}
	return nil
}

func (r *Repository) TeamExists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)`

	conn := r.getConn(ctx)
	var exists bool
	err := conn.QueryRow(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}

	return exists, nil
}

func (r *Repository) GetTeamWithMembers(ctx context.Context, name string) (*models.Team, error) {
	conn := r.db

	var teamExists bool
	err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)", name).
		Scan(&teamExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}
	if !teamExists {
		return nil, nil
	}

	rows, err := conn.Query(ctx,
		`SELECT user_id, username, is_active 
         FROM users 
         WHERE team_id = $1 
         ORDER BY username`, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team members: %w", err)
	}

	return &models.Team{
		TeamName: name,
		Members:  members,
	}, nil
}
