package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description string
	IsSystem    bool
	CreatedAt   time.Time
}

type UserRole struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	RoleID    uuid.UUID
	TenantID  uuid.UUID
	CreatedAt time.Time
}

func NewRole(tenantID uuid.UUID, name, description string) *Role {
	return &Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}
}
