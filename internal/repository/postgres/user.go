package postgres

import (
	"context"
	"errors"
	"fmt"

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
