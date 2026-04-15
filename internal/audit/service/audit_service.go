package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/audit/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/audit/repository"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
)

type AuditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) CreateLog(ctx context.Context, log *domain.AuditLog) error {
	if err := s.repo.Create(ctx, log); err != nil {
		return errors.NewInternalError("failed to create audit log")
	}
	return nil
}

func (s *AuditService) ListLogs(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, offset, limit int) ([]*domain.AuditLog, int, error) {
	logs, count, err := s.repo.ListByTenant(ctx, tenantID, startTime, endTime, offset, limit)
	if err != nil {
		return nil, 0, errors.NewInternalError("failed to list audit logs")
	}
	return logs, count, nil
}
