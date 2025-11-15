package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/Killazius/avito-pr/internal/repository/postgres"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

var ErrUserNotFound = fmt.Errorf("user not found")

type UserRepository interface {
	UpdateIsActiveStatus(ctx context.Context, userID string, isActive bool) error
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequestShort, error)
}
type Service struct {
	userRepo  UserRepository
	trManager *manager.Manager
}

func NewService(userRepo UserRepository, trManager *manager.Manager) *Service {
	return &Service{
		userRepo:  userRepo,
		trManager: trManager,
	}
}

func (s *Service) UpdateUserStatus(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		err := s.userRepo.UpdateIsActiveStatus(ctx, userID, isActive)
		if err != nil {
			if errors.Is(err, postgres.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("failed to update user status: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil

}
func (s *Service) GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequestShort, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	pullRequests, err := s.userRepo.GetUserReviews(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reviews: %w", err)
	}

	if pullRequests == nil {
		pullRequests = []*models.PullRequestShort{}
	}
	return pullRequests, nil
}
