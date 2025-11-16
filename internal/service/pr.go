package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/Killazius/avito-pr/internal/repository/postgres"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type PRService struct {
	trManager *manager.Manager
	teamRepo  TeamRepository
	userRepo  UserRepository
	prRepo    PRRepository
}

func NewPRService(teamRepository TeamRepository, userRepository UserRepository, pRRepository PRRepository, trManager *manager.Manager) *PRService {
	return &PRService{
		trManager: trManager,
		teamRepo:  teamRepository,
		userRepo:  userRepository,
		prRepo:    pRRepository,
	}
}

func (s *PRService) CreatePR(ctx context.Context, prID, prName, authorID string) (*models.PullRequest, error) {
	if prID == "" || prName == "" || authorID == "" {
		return nil, errors.New("prID, prName and authorID are required")
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
			if errors.Is(err, postgres.ErrUserNotFound) {
				return ErrUserNotFound
			}

			return fmt.Errorf("failed to get author: %w", err)
		}
		if user == nil {
			return ErrUserNotFound
		}
		if !user.IsActive {
			return ErrUserInactive
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
			return fmt.Errorf("failed to create PR: %w", errCreate)
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
	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		if errors.Is(err, postgres.ErrPRNotFound) {
			return nil, ErrPRNotFound
		}

		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	return pr, nil
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	if prID == "" {
		return nil, errors.New("prID is required")
	}
	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		pr, err := s.prRepo.GetPRByID(ctx, prID)
		if err != nil {
			if errors.Is(err, postgres.ErrPRNotFound) {
				return ErrPRNotFound
			}

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

func (s *PRService) ReassignReviewer(ctx context.Context, pullRequestID, oldUserID string) (*models.PullRequest, string, error) {
	if pullRequestID == "" || oldUserID == "" {
		return nil, "", errors.New("pullRequestID and oldUserID are required")
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
			if errors.Is(err, postgres.ErrUserNotFound) {
				return ErrUserNotFound
			}

			return fmt.Errorf("failed to get PR author: %w", err)
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
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(availableCandidates))))
		if err != nil {
			return fmt.Errorf("failed to generate random index: %w", err)
		}

		selectedIndex := int(n.Int64())
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

func (s *PRService) autoAssignReviewers(ctx context.Context, author *models.User) ([]string, error) {
	candidates, err := s.userRepo.GetActiveTeamMembers(ctx, author.TeamName, author.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	if len(candidates) == 0 {
		return nil, ErrNoReviewers
	}
	for i := len(candidates) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, err
		}
		j := int(n.Int64())
		candidates[i], candidates[j] = candidates[j], candidates[i]
	}

	maxReviewers := min(2, len(candidates))
	reviewerIDs := make([]string, 0, maxReviewers)

	for i := 0; i < maxReviewers; i++ {
		reviewerIDs = append(reviewerIDs, candidates[i].UserID)
	}

	return reviewerIDs, nil
}
