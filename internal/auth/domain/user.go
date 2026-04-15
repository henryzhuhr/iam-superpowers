package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrWeakPassword    = errors.New("password must be at least 8 characters with uppercase, lowercase, and number")
	ErrInvalidPassword = errors.New("invalid password")
	ErrAccountLocked   = errors.New("account is locked")
	ErrTooManyAttempts = errors.New("too many login attempts")
)

type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusLocked   UserStatus = "locked"
)

type User struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Email         string
	PasswordHash  string
	Name          string
	AvatarURL     string
	Status        UserStatus
	EmailVerified bool
	LoginAttempts int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return ErrWeakPassword
	}
	return nil
}

func NewUser(tenantID uuid.UUID, email, password string) (*User, error) {
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: string(hash),
		Name:         "",
		Status:       StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (u *User) VerifyPassword(password string) error {
	if u.Status == StatusLocked {
		return ErrAccountLocked
	}
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

func (u *User) RecordFailedLogin() {
	u.LoginAttempts++
	if u.LoginAttempts >= 5 {
		u.Status = StatusLocked
	}
}

func (u *User) RecordSuccessfulLogin() {
	u.LoginAttempts = 0
	u.Status = StatusActive
}

func (u *User) ChangePassword(oldPassword, newPassword string) error {
	if err := u.VerifyPassword(oldPassword); err != nil {
		return err
	}
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	u.UpdatedAt = time.Now()
	return nil
}
