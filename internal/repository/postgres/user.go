package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) CreateOrUpdateUser(ctx context.Context, user models.TeamMember, teamID string) error {
	query := `
        INSERT INTO users (user_id, username, team_id, is_active, created_at, updated_at) 
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        ON CONFLICT (user_id) 
        DO UPDATE SET 
            username = EXCLUDED.username,
            team_id = EXCLUDED.team_id,
            is_active = EXCLUDED.is_active,
            updated_at = NOW()`

	conn := r.getConn(ctx)
	_, err := conn.Exec(ctx, query, user.UserID, user.Username, teamID, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create or update user: %w", err)
	}
	return nil
}

var ErrUserNotFound = fmt.Errorf("user not found")

func (r *Repository) UpdateIsActiveStatus(ctx context.Context, userID string, isActive bool) error {
	query := `UPDATE users SET is_active = $1, updated_at = NOW() WHERE user_id = $2`

	conn := r.getConn(ctx)
	result, err := conn.Exec(ctx, query, isActive, userID)
	if err != nil {
		return fmt.Errorf("failed to update is_active status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	query := `SELECT user_id, username, team_id, is_active FROM users WHERE user_id = $1`
	conn := r.getConn(ctx)
	var user models.User
	err := conn.QueryRow(ctx, query, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

func (r *Repository) GetActiveTeamMembers(ctx context.Context, teamID, authorID string, excludeIDs ...string) ([]*models.User, error) {
	query := `
        SELECT user_id, username, team_id, is_active
        FROM users 
        WHERE team_id = $1 AND is_active = true`

	args := []interface{}{teamID}

	if authorID != "" {
		excludeIDs = append(excludeIDs, authorID)
	}

	if len(excludeIDs) > 0 {
		placeholders := make([]string, len(excludeIDs))
		for i := range excludeIDs {
			placeholders[i] = fmt.Sprintf("$%d", len(args)+1)
			args = append(args, excludeIDs[i])
		}
		query += fmt.Sprintf(" AND user_id NOT IN (%s)", strings.Join(placeholders, ","))
	}

	query += " ORDER BY user_id"

	conn := r.getConn(ctx)
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get active team members: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}
