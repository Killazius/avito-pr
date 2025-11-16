package service

import (
	"context"
	"errors"
	"time"

	"github.com/Killazius/avito-pr/internal/models"
)

var (
	ErrPRExists     = errors.New("PR already exists")
	ErrNoReviewers  = errors.New("no available reviewers in team")
	ErrPRNotFound   = errors.New("PR not found")
	ErrPRMerged     = errors.New("cannot reassign reviewer for merged PR")
	ErrNotAssigned  = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate  = errors.New("no active replacement candidate in team")
	ErrUserInactive = errors.New("user is not active")
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found")
	ErrUserNotFound = errors.New("user not found")
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, teamName string) error
	TeamExists(ctx context.Context, name string) (bool, error)
	GetTeamWithMembers(ctx context.Context, name string) (*models.Team, error)
}

type UserRepository interface {
	CreateOrUpdateUser(ctx context.Context, user models.TeamMember, teamID string) error
	UpdateIsActiveStatus(ctx context.Context, userID string, isActive bool) error
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GetActiveTeamMembers(ctx context.Context, teamID, authorID string, excludeIDs ...string) ([]*models.User, error)
	GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequestShort, error)
}

type PRRepository interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) error
	PRExists(ctx context.Context, prID string) (bool, error)
	GetPRByID(ctx context.Context, prID string) (*models.PullRequest, error)
	AssignReviewer(ctx context.Context, prID, userID string) error
	MarkPRAsMerged(ctx context.Context, prID string, status models.PRStatus, now time.Time) error
	RemoveReviewer(ctx context.Context, prID, userID string) error
}

type StatsRepository interface {
	GetTeamsCount(ctx context.Context) (int, error)
	GetUsersStats(ctx context.Context) (*models.UsersStats, error)
	GetPRStats(ctx context.Context) (*models.PRStats, error)
	GetAssignmentsCount(ctx context.Context) (int, error)
}
