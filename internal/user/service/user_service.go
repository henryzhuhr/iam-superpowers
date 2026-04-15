package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/repository"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("failed to find user")
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user not found")
	}
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, name, avatarURL string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.NewNotFoundError("user not found")
	}

	user.Name = name
	user.AvatarURL = avatarURL
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.NewNotFoundError("user not found")
	}

	if err := user.ChangePassword(oldPassword, newPassword); err != nil {
		if err == domain.ErrInvalidPassword {
			return errors.NewValidationError("current password is incorrect")
		}
		if err == domain.ErrWeakPassword {
			return errors.NewValidationError(err.Error())
		}
		return errors.NewInternalError("failed to change password")
	}

	return s.userRepo.Update(ctx, user)
}
