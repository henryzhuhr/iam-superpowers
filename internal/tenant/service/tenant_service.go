package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/tenant/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/tenant/repository"
)

type TenantService struct {
	repo repository.TenantRepository
}

func NewTenantService(repo repository.TenantRepository) *TenantService {
	return &TenantService{repo: repo}
}

func (s *TenantService) CreateTenant(ctx context.Context, name, uniqueCode, customDomain string) (*domain.Tenant, error) {
	existing, err := s.repo.FindByCode(ctx, uniqueCode)
	if err != nil {
		return nil, errors.NewInternalError("failed to check tenant existence")
	}
	if existing != nil {
		return nil, errors.NewConflictError("tenant code already exists")
	}

	tenant := domain.NewTenant(name, uniqueCode)
	tenant.CustomDomain = customDomain
	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, errors.NewInternalError("failed to create tenant")
	}
	return tenant, nil
}

func (s *TenantService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewInternalError("failed to find tenant")
	}
	if tenant == nil {
		return nil, errors.NewNotFoundError("tenant not found")
	}
	return tenant, nil
}

func (s *TenantService) GetTenantByCode(ctx context.Context, code string) (*domain.Tenant, error) {
	tenant, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, errors.NewInternalError("failed to find tenant")
	}
	if tenant == nil {
		return nil, errors.NewNotFoundError("tenant not found")
	}
	return tenant, nil
}

func (s *TenantService) ListTenants(ctx context.Context, offset, limit int) ([]*domain.Tenant, int, error) {
	tenants, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, errors.NewInternalError("failed to list tenants")
	}
	count, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, errors.NewInternalError("failed to count tenants")
	}
	return tenants, count, nil
}
