package domain

import (
	"time"

	"github.com/google/uuid"
)

type TenantStatus string

const (
	TenantStatusActive   TenantStatus = "active"
	TenantStatusInactive TenantStatus = "inactive"
)

type Tenant struct {
	ID           uuid.UUID
	Name         string
	UniqueCode   string
	CustomDomain string
	Status       TenantStatus
	CreatedAt    time.Time
}

func NewTenant(name, uniqueCode string) *Tenant {
	return &Tenant{
		ID:         uuid.New(),
		Name:       name,
		UniqueCode: uniqueCode,
		Status:     TenantStatusActive,
		CreatedAt:  time.Now(),
	}
}
