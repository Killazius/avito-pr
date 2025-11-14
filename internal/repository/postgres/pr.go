package postgres

import (
	"context"
	"fmt"

	"github.com/Killazius/avito-pr/internal/models"
)

func (r *Repository) GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequestShort, error) {
	query := `
        SELECT 
            p.id as pull_request_id,
            p.name as pull_request_name, 
            p.author_id,
            p.status,
            p.created_at,
            p.merged_at
        FROM pull_requests p
        INNER JOIN pr_reviewers pr ON p.id = pr.pr_id
        WHERE pr.user_id = $1
        ORDER BY p.created_at DESC`

	conn := r.getConn(ctx)
	rows, err := conn.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reviews: %w", err)
	}
	defer rows.Close()

	var pullRequests []*models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		var createdAt, mergedAt *string

		err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
			&createdAt,
			&mergedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pull request: %w", err)
		}

		pullRequests = append(pullRequests, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pull requests: %w", err)
	}

	return pullRequests, nil
}
