package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/common/database"
	"github.com/jackc/pgx/v5"
)

type ListUsersFilter struct {
	TenantID uuid.UUID
	Email    string
	Status   string
	Offset   int
	Limit    int
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	ListUsers(ctx context.Context, filter ListUsersFilter) ([]*domain.User, int, error)
	Update(ctx context.Context, user *domain.User) error
}

type pgxUserRepo struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &pgxUserRepo{db: db}
}

const userColumns = "id, tenant_id, email, password_hash, name, avatar_url, status, email_verified, login_attempts, created_at, updated_at"

func (r *pgxUserRepo) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO users (id, tenant_id, email, password_hash, name, status, email_verified)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.TenantID, user.Email, user.PasswordHash, user.Name,
		user.Status, user.EmailVerified,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *pgxUserRepo) FindByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.db.Pool.QueryRow(ctx,
		`SELECT `+userColumns+`
		 FROM users WHERE email = $1 AND tenant_id = $2`,
		email, tenantID,
	).Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name,
		&u.AvatarURL, &u.Status, &u.EmailVerified, &u.LoginAttempts, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &u, nil
}

func (r *pgxUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.Pool.QueryRow(ctx,
		`SELECT `+userColumns+`
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name,
		&u.AvatarURL, &u.Status, &u.EmailVerified, &u.LoginAttempts, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &u, nil
}

func (r *pgxUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.db.Pool.QueryRow(ctx,
		`SELECT `+userColumns+`
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name,
		&u.AvatarURL, &u.Status, &u.EmailVerified, &u.LoginAttempts, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &u, nil
}

func (r *pgxUserRepo) ListUsers(ctx context.Context, filter ListUsersFilter) ([]*domain.User, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filter.TenantID != uuid.Nil {
		where = append(where, fmt.Sprintf("tenant_id = $%d", argIdx))
		args = append(args, filter.TenantID)
		argIdx++
	}
	if filter.Email != "" {
		where = append(where, fmt.Sprintf("email ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Email+"%")
		argIdx++
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var count int
	err := r.db.Pool.QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause),
		args...,
	).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Query
	query := fmt.Sprintf(
		"SELECT %s FROM users WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		userColumns, whereClause, argIdx, argIdx+1,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Status, &u.EmailVerified, &u.LoginAttempts, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &u)
	}
	return users, count, nil
}

func (r *pgxUserRepo) Update(ctx context.Context, user *domain.User) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE users SET password_hash = $2, name = $3, avatar_url = $4,
			                status = $5, email_verified = $6, login_attempts = $7, updated_at = NOW()
			 WHERE id = $1`,
		user.ID, user.PasswordHash, user.Name, user.AvatarURL,
		user.Status, user.EmailVerified, user.LoginAttempts,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
