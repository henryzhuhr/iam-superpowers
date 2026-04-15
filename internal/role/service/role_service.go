package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/role/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/role/repository"
)

type RoleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func (s *RoleService) CreateRole(ctx context.Context, tenantID uuid.UUID, name, description string) (*domain.Role, error) {
	role := domain.NewRole(tenantID, name, description)
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, errors.NewInternalError("failed to create role")
	}
	return role, nil
}

func (s *RoleService) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	roles, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("failed to list roles")
	}
	return roles, nil
}

func (s *RoleService) AssignRoleToUser(ctx context.Context, userID, roleID, tenantID uuid.UUID) error {
	role, err := s.repo.FindByID(ctx, roleID)
	if err != nil || role == nil {
		return errors.NewNotFoundError("role not found")
	}
	if role.TenantID != tenantID {
		return errors.NewForbiddenError("role does not belong to this tenant")
	}

	userRole := &domain.UserRole{
		ID:       uuid.New(),
		UserID:   userID,
		RoleID:   roleID,
		TenantID: tenantID,
	}
	if err := s.repo.AssignRoleToUser(ctx, userRole); err != nil {
		return errors.NewInternalError("failed to assign role")
	}
	return nil
}

func (s *RoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := s.repo.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return errors.NewInternalError("failed to remove role")
	}
	return nil
}

func (s *RoleService) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.Role, error) {
	roles, err := s.repo.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("failed to get user roles")
	}
	return roles, nil
}
