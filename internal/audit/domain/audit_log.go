package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	UserID     *uuid.UUID
	Action     string
	TargetType string
	TargetID   *uuid.UUID
	Details    map[string]interface{}
	IPAddress  string
	CreatedAt  time.Time
}
