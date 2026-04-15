package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/common/database"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

type pgxUserRepo struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &pgxUserRepo{db: db}
}

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
		`SELECT id, tenant_id, email, password_hash, name, avatar_url, status,
		        email_verified, login_attempts, created_at, updated_at
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

func (r *pgxUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, tenant_id, email, password_hash, name, avatar_url, status,
		        email_verified, login_attempts, created_at, updated_at
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
