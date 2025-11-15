package pr

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type TeamRepository interface {
	TeamExists(ctx context.Context, name string) (bool, error)
}

type UserRepository interface {
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GetActiveTeamMembers(ctx context.Context, teamID, authorID string, excludeIDs ...string) ([]*models.User, error)
}

type PRRepository interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) error
	PRExists(ctx context.Context, prID string) (bool, error)
	GetPRByID(ctx context.Context, prID string) (*models.PullRequest, error)
	AssignReviewer(ctx context.Context, prID, userID string) error
	MarkPRAsMerged(ctx context.Context, prID string, status models.PRStatus, now time.Time) error
	RemoveReviewer(ctx context.Context, prID, userID string) error
}
type Service struct {
	trManager *manager.Manager
	teamRepo  TeamRepository
	userRepo  UserRepository
	prRepo    PRRepository
}

func NewService(TeamRepository TeamRepository, UserRepository UserRepository, PRRepository PRRepository, trManager *manager.Manager) *Service {
	return &Service{
		trManager: trManager,
		teamRepo:  TeamRepository,
		userRepo:  UserRepository,
		prRepo:    PRRepository,
	}
}

var (
	ErrPRExists     = errors.New("PR already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrTeamNotFound = errors.New("team not found")
	ErrNoReviewers  = errors.New("no available reviewers in team")
	ErrPRNotFound   = errors.New("PR not found")
	ErrPRMerged     = errors.New("cannot reassign reviewer for merged PR")
	ErrNotAssigned  = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate  = errors.New("no active replacement candidate in team")
)

func (s *Service) CreatePR(ctx context.Context, prID string, prName string, authorID string) (*models.PullRequest, error) {
	if prID == "" || prName == "" || authorID == "" {
		return nil, fmt.Errorf("prID, prName and authorID are required")
	}
	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		exists, err := s.prRepo.PRExists(ctx, prID)
		if err != nil {
			return fmt.Errorf("failed to check PR existence: %w", err)
		}
		if exists {
			return ErrPRExists
		}
		user, err := s.userRepo.GetUserByID(ctx, authorID)
		if err != nil {
			return fmt.Errorf("failed to get author: %w", err)
		}
		if user == nil {
			return ErrUserNotFound
		}
		if user.IsActive == false {
			return fmt.Errorf("author user is not active")
		}
		teamExists, err := s.teamRepo.TeamExists(ctx, user.TeamName)
		if err != nil {
			return fmt.Errorf("failed to check team existence: %w", err)
		}
		if !teamExists {
			return ErrTeamNotFound
		}

		pr := &models.PullRequest{
			PullRequestID:   prID,
			PullRequestName: prName,
			AuthorID:        authorID,
			Status:          models.PRStatusOpen,
		}
		if errCreate := s.prRepo.CreatePR(ctx, pr); errCreate != nil {
			return fmt.Errorf("failed to create PR: %w", err)
		}
		reviewers, err := s.autoAssignReviewers(ctx, user)
		if err != nil && !errors.Is(err, ErrNoReviewers) {
			return err
		}
		for _, reviewerID := range reviewers {
			if err := s.prRepo.AssignReviewer(ctx, prID, reviewerID); err != nil {
				return fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.prRepo.GetPRByID(ctx, prID)
}

func (s *Service) autoAssignReviewers(ctx context.Context, author *models.User) ([]string, error) {
	candidates, err := s.userRepo.GetActiveTeamMembers(ctx, author.TeamName, author.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	if len(candidates) == 0 {
		return nil, ErrNoReviewers
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	maxReviewers := min(2, len(candidates))
	reviewerIDs := make([]string, 0, maxReviewers)

	for i := 0; i < maxReviewers; i++ {
		reviewerIDs = append(reviewerIDs, candidates[i].UserID)
	}

	return reviewerIDs, nil
}

func (s *Service) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	if prID == "" {
		return nil, fmt.Errorf("prID is required")
	}
	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		pr, err := s.prRepo.GetPRByID(ctx, prID)
		if err != nil {
			return fmt.Errorf("failed to get PR: %w", err)
		}
		if pr == nil {
			return ErrPRNotFound
		}
		if pr.Status == models.PRStatusMerged {
			return nil
		}
		err = s.prRepo.MarkPRAsMerged(ctx, prID, models.PRStatusMerged, time.Now())
		if err != nil {
			return fmt.Errorf("failed to mark PR as merged: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR after merge: %w", err)
	}
	return pr, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, pullRequestID, oldUserID string) (*models.PullRequest, string, error) {
	if pullRequestID == "" || oldUserID == "" {
		return nil, "", fmt.Errorf("pullRequestID and oldUserID are required")
	}
	var newReviewerID string
	var resultPR *models.PullRequest
	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		pr, err := s.prRepo.GetPRByID(ctx, pullRequestID)
		if err != nil {
			return fmt.Errorf("failed to get PR: %w", err)
		}
		if pr == nil {
			return ErrPRNotFound
		}
		if pr.Status == models.PRStatusMerged {
			return ErrPRMerged
		}
		isAssigned := false
		for _, reviewer := range pr.AssignedReviewers {
			if reviewer == oldUserID {
				isAssigned = true
				break
			}
		}
		if !isAssigned {
			return ErrNotAssigned
		}

		author, err := s.userRepo.GetUserByID(ctx, pr.AuthorID)
		if err != nil {
			return fmt.Errorf("failed to get PR author: %w", err)
		}
		if author == nil {
			return ErrUserNotFound
		}
		candidates, err := s.userRepo.GetActiveTeamMembers(ctx, author.TeamName, oldUserID, author.UserID)
		if err != nil {
			return fmt.Errorf("failed to get team members: %w", err)
		}
		if len(candidates) == 0 {
			return ErrNoReviewers
		}
		var availableCandidates []*models.User
		for _, candidate := range candidates {
			isAlreadyReviewer := false
			for _, reviewer := range pr.AssignedReviewers {
				if candidate.UserID == reviewer {
					isAlreadyReviewer = true
					break
				}
			}
			if !isAlreadyReviewer {
				availableCandidates = append(availableCandidates, candidate)
			}
		}
		if len(availableCandidates) == 0 {
			return ErrNoCandidate
		}
		rand.New(rand.NewSource(time.Now().UnixNano()))
		selectedIndex := rand.Intn(len(availableCandidates))
		newReviewerID = availableCandidates[selectedIndex].UserID
		if err = s.prRepo.RemoveReviewer(ctx, pullRequestID, oldUserID); err != nil {
			return fmt.Errorf("failed to remove old reviewer: %w", err)
		}
		if err = s.prRepo.AssignReviewer(ctx, pullRequestID, newReviewerID); err != nil {
			return fmt.Errorf("failed to assign new reviewer: %w", err)
		}
		updatedPR, err := s.prRepo.GetPRByID(ctx, pullRequestID)
		if err != nil {
			return fmt.Errorf("failed to get updated PR: %w", err)
		}
		resultPR = updatedPR
		return nil
	})
	if err != nil {
		return nil, "", err
	}

	return resultPR, newReviewerID, nil
}
