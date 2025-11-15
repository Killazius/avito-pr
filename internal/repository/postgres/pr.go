package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	query := `
        INSERT INTO pull_requests (id, name, author_id, status) 
        VALUES ($1, $2, $3, $4)`

	conn := r.getConn(ctx)
	_, err := conn.Exec(ctx, query,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}
	return nil
}
func (r *Repository) PRExists(ctx context.Context, prID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id = $1)`

	conn := r.getConn(ctx)
	var exists bool
	err := conn.QueryRow(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check PR existence: %w", err)
	}

	return exists, nil
}

func (r *Repository) GetPRByID(ctx context.Context, prID string) (*models.PullRequest, error) {
	query := `
        SELECT id, name, author_id, status, created_at, merged_at 
        FROM pull_requests 
        WHERE id = $1`

	conn := r.getConn(ctx)
	var pr models.PullRequest
	var mergedAt *time.Time

	err := conn.QueryRow(ctx, query, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&mergedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	pr.MergedAt = mergedAt

	reviewers, err := r.GetPRReviewers(ctx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *Repository) AssignReviewer(ctx context.Context, prID, userID string) error {
	query := `INSERT INTO pr_reviewers (pr_id, user_id, assigned_at) VALUES ($1, $2, $3)`

	conn := r.getConn(ctx)
	_, err := conn.Exec(ctx, query, prID, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to assign reviewer: %w", err)
	}
	return nil
}

func (r *Repository) GetPRReviewers(ctx context.Context, prID string) ([]string, error) {
	query := `SELECT user_id FROM pr_reviewers WHERE pr_id = $1 ORDER BY assigned_at`

	conn := r.getConn(ctx)
	rows, err := conn.Query(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewers = append(reviewers, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviewers: %w", err)
	}

	return reviewers, nil
}

func (r *Repository) MarkPRAsMerged(ctx context.Context, prID string, status models.PRStatus, now time.Time) error {
	query := `UPDATE pull_requests SET status = $1, merged_at = $2 WHERE id = $3`
	conn := r.getConn(ctx)
	result, err := conn.Exec(ctx, query, status, now, prID)
	if err != nil {
		return fmt.Errorf("failed to mark PR as merged: %w", err)
	}

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("PR already merged")
	}

	return nil

}
func (r *Repository) RemoveReviewer(ctx context.Context, prID, userID string) error {
	query := `DELETE FROM pr_reviewers WHERE pr_id = $1 AND user_id = $2`

	conn := r.getConn(ctx)
	result, err := conn.Exec(ctx, query, prID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove reviewer: %w", err)
	}

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("reviewer not found for this PR")
	}

	return nil
}
