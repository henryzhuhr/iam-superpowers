package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/database"
	"github.com/henryzhuhr/iam-superpowers/internal/tenant/domain"
	"github.com/jackc/pgx/v5"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	FindByCode(ctx context.Context, code string) (*domain.Tenant, error)
	List(ctx context.Context, offset, limit int) ([]*domain.Tenant, error)
	Count(ctx context.Context) (int, error)
}

type pgxTenantRepo struct {
	db *database.DB
}

func NewTenantRepository(db *database.DB) TenantRepository {
	return &pgxTenantRepo{db: db}
}

func (r *pgxTenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO tenants (id, name, unique_code, custom_domain, status)
		 VALUES ($1, $2, $3, $4, $5)`,
		tenant.ID, tenant.Name, tenant.UniqueCode, tenant.CustomDomain, tenant.Status,
	)
	return err
}

func (r *pgxTenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, name, unique_code, custom_domain, status, created_at FROM tenants WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.Name, &t.UniqueCode, &t.CustomDomain, &t.Status, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *pgxTenantRepo) FindByCode(ctx context.Context, code string) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, name, unique_code, custom_domain, status, created_at FROM tenants WHERE unique_code = $1`,
		code,
	).Scan(&t.ID, &t.Name, &t.UniqueCode, &t.CustomDomain, &t.Status, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *pgxTenantRepo) List(ctx context.Context, offset, limit int) ([]*domain.Tenant, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, name, unique_code, custom_domain, status, created_at FROM tenants ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		var t domain.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.UniqueCode, &t.CustomDomain, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, &t)
	}
	return tenants, nil
}

func (r *pgxTenantRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&count)
	return count, err
}

// Suppress unused import error
var _ = fmt.Sprintf
