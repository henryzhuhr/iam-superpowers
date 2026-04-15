package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"weak: too short", "Ab1", true},
		{"weak: no uppercase", "abcdefgh1", true},
		{"weak: no lowercase", "ABCDEFGH1", true},
		{"weak: no number", "Abcdefgh", true},
		{"valid", "Password1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewUser(t *testing.T) {
	tenantID := uuid.New()
	user, err := NewUser(tenantID, "test@example.com", "Password1")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, StatusActive, user.Status)
	assert.NotEmpty(t, user.PasswordHash)
}

func TestNewUser_WeakPassword(t *testing.T) {
	tenantID := uuid.New()
	user, err := NewUser(tenantID, "test@example.com", "weak")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUser_VerifyPassword(t *testing.T) {
	tenantID := uuid.New()
	user, _ := NewUser(tenantID, "test@example.com", "Password1")

	assert.NoError(t, user.VerifyPassword("Password1"))
	assert.Error(t, user.VerifyPassword("WrongPassword"))
}

func TestUser_RecordFailedLogin_Lockout(t *testing.T) {
	tenantID := uuid.New()
	user, _ := NewUser(tenantID, "test@example.com", "Password1")

	for i := 0; i < 5; i++ {
		user.RecordFailedLogin()
	}
	assert.Equal(t, StatusLocked, user.Status)
}

func TestUser_ChangePassword(t *testing.T) {
	tenantID := uuid.New()
	user, _ := NewUser(tenantID, "test@example.com", "Password1")

	err := user.ChangePassword("Password1", "NewPassword1")
	assert.NoError(t, err)
	assert.NoError(t, user.VerifyPassword("NewPassword1"))
}
