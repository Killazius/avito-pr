package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Killazius/avito-pr/internal/models"
	"github.com/Killazius/avito-pr/internal/repository/postgres"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type UserService struct {
	userRepo  UserRepository
	trManager *manager.Manager
}

func NewUserService(userRepo UserRepository, trManager *manager.Manager) *UserService {
	return &UserService{
		userRepo:  userRepo,
		trManager: trManager,
	}
}

func (s *UserService) UpdateUserStatus(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
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

func (s *UserService) GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequestShort, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
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
