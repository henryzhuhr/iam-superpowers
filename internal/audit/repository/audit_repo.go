package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/audit/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/common/database"
)

type AuditRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, offset, limit int) ([]*domain.AuditLog, int, error)
}

type pgxAuditRepo struct {
	db *database.DB
}

func NewAuditRepository(db *database.DB) AuditRepository {
	return &pgxAuditRepo{db: db}
}

func (r *pgxAuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	detailsJSON, _ := json.Marshal(log.Details)
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO audit_logs (id, tenant_id, user_id, action, target_type, target_id, details, ip_address)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ID, log.TenantID, log.UserID, log.Action, log.TargetType, log.TargetID, detailsJSON, log.IPAddress,
	)
	return err
}

func (r *pgxAuditRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, offset, limit int) ([]*domain.AuditLog, int, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, tenant_id, user_id, action, target_type, target_id, details::text, ip_address, created_at
		 FROM audit_logs WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3
		 ORDER BY created_at DESC LIMIT $4 OFFSET $5`,
		tenantID, startTime, endTime, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var detailsText string
		if err := rows.Scan(&l.ID, &l.TenantID, &l.UserID, &l.Action, &l.TargetType, &l.TargetID, &detailsText, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		if detailsText != "" {
			_ = json.Unmarshal([]byte(detailsText), &l.Details)
		}
		logs = append(logs, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var count int
	err = r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3`,
		tenantID, startTime, endTime,
	).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}
