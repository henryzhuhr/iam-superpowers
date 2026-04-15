package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/database"
	"github.com/henryzhuhr/iam-superpowers/internal/role/domain"
	"github.com/jackc/pgx/v5"
)

type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error)
	AssignRoleToUser(ctx context.Context, userRole *domain.UserRole) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.Role, error)
}

type pgxRoleRepo struct {
	db *database.DB
}

func NewRoleRepository(db *database.DB) RoleRepository {
	return &pgxRoleRepo{db: db}
}

func (r *pgxRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO roles (id, tenant_id, name, description, is_system)
		 VALUES ($1, $2, $3, $4, $5)`,
		role.ID, role.TenantID, role.Name, role.Description, role.IsSystem,
	)
	return err
}

func (r *pgxRoleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, description, is_system, created_at FROM roles WHERE id = $1`,
		id,
	).Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &role, err
}

func (r *pgxRoleRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, tenant_id, name, description, is_system, created_at FROM roles WHERE tenant_id = $1`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

func (r *pgxRoleRepo) AssignRoleToUser(ctx context.Context, userRole *domain.UserRole) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO user_roles (id, user_id, role_id, tenant_id) VALUES ($1, $2, $3, $4)`,
		userRole.ID, userRole.UserID, userRole.RoleID, userRole.TenantID,
	)
	return err
}

func (r *pgxRoleRepo) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`, userID, roleID)
	return err
}

func (r *pgxRoleRepo) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.Role, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT r.id, r.tenant_id, r.name, r.description, r.is_system, r.created_at
		 FROM roles r JOIN user_roles ur ON r.id = ur.role_id
		 WHERE ur.user_id = $1 AND r.tenant_id = $2`,
		userID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}
