package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/repository"
	"github.com/henryzhuhr/iam-superpowers/internal/common/email"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/common/jwt"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	userRepo repository.UserRepository
	jwtSvc   *jwt.Service
	emailSvc *email.Service
	redis    *redis.Client
}

func NewAuthService(userRepo repository.UserRepository, jwtSvc *jwt.Service, emailSvc *email.Service, redis *redis.Client) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtSvc:   jwtSvc,
		emailSvc: emailSvc,
		redis:    redis,
	}
}

func (s *AuthService) Register(ctx context.Context, tenantID uuid.UUID, email, password string) (*domain.User, error) {
	existing, err := s.userRepo.FindByEmailAndTenant(ctx, email, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("failed to check user existence")
	}
	if existing != nil {
		return nil, errors.NewConflictError("email already registered in this tenant")
	}

	user, err := domain.NewUser(tenantID, email, password)
	if err != nil {
		if err == domain.ErrWeakPassword {
			return nil, errors.NewValidationError(err.Error())
		}
		return nil, errors.NewInternalError("failed to create user")
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.NewInternalError("failed to save user")
	}

	// Send verification code
	code, err := generateVerificationCode()
	if err != nil {
		return nil, errors.NewInternalError("failed to generate verification code")
	}
	s.redis.Set(ctx, fmt.Sprintf("email_verify:%s", email), code, 5*time.Minute)
	go func() {
		if err := s.emailSvc.SendVerificationCode(email, code); err != nil {
			slog.Error("failed to send verification email", "email", email, "error", err)
		}
	}()

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string, tenantID uuid.UUID) (*jwt.TokenPair, error) {
	user, err := s.userRepo.FindByEmailAndTenant(ctx, email, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("failed to find user")
	}
	if user == nil {
		return nil, errors.NewUnauthorizedError("invalid email or password")
	}

	if err := user.VerifyPassword(password); err != nil {
		user.RecordFailedLogin()
		_ = s.userRepo.Update(ctx, user)
		return nil, errors.NewUnauthorizedError("invalid email or password")
	}

	user.RecordSuccessfulLogin()
	_ = s.userRepo.Update(ctx, user)

	// Get user roles
	roles := []string{"user"} // TODO: fetch from DB

	tokenPair, err := s.jwtSvc.GenerateTokenPair(user.ID.String(), user.TenantID.String(), roles)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate tokens")
	}

	// Store refresh token in Redis
	refreshKey := fmt.Sprintf("refresh:%s:%s", user.ID, tokenPair.RefreshToken)
	s.redis.Set(ctx, refreshKey, user.ID.String(), 7*24*time.Hour)

	return tokenPair, nil
}

func (s *AuthService) Refresh(ctx context.Context, userID, refreshToken string) (*jwt.TokenPair, error) {
	oldKey := fmt.Sprintf("refresh:%s:%s", userID, refreshToken)
	val, err := s.redis.Get(ctx, oldKey).Result()
	if err == redis.Nil {
		return nil, errors.NewUnauthorizedError("invalid refresh token")
	}
	if val == "" {
		return nil, errors.NewUnauthorizedError("invalid refresh token")
	}

	// Delete old token (rotation)
	s.redis.Del(ctx, oldKey)

	// Find user to get tenant and roles
	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.NewUnauthorizedError("invalid user ID")
	}
	user, err := s.userRepo.FindByID(ctx, userIDUUID)
	if err != nil || user == nil {
		return nil, errors.NewUnauthorizedError("user not found")
	}

	roles := []string{"user"} // TODO: fetch from DB
	tokenPair, err := s.jwtSvc.GenerateTokenPair(user.ID.String(), user.TenantID.String(), roles)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate tokens")
	}

	// Store new refresh token
	newKey := fmt.Sprintf("refresh:%s:%s", user.ID, tokenPair.RefreshToken)
	s.redis.Set(ctx, newKey, user.ID.String(), 7*24*time.Hour)

	return tokenPair, nil
}

func (s *AuthService) Logout(ctx context.Context, userID, refreshToken, jti string) error {
	// Delete refresh token
	oldKey := fmt.Sprintf("refresh:%s:%s", userID, refreshToken)
	s.redis.Del(ctx, oldKey)

	// Blacklist JWT
	if jti != "" {
		s.redis.Set(ctx, fmt.Sprintf("jwt_blacklist:%s", jti), "1", 15*time.Minute)
	}

	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, email, code string) error {
	stored, err := s.redis.Get(ctx, fmt.Sprintf("email_verify:%s", email)).Result()
	if err == redis.Nil {
		return errors.NewValidationError("verification code expired or not found")
	}
	if stored != code {
		return errors.NewValidationError("invalid verification code")
	}

	s.redis.Del(ctx, fmt.Sprintf("email_verify:%s", email))

	// Update user
	// TODO: need to find user by email and set email_verified = true

	return nil
}

// ForgotPassword generates a reset token for the given email and stores it in Redis.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return errors.NewInternalError("failed to find user")
	}
	if user == nil {
		// Return nil to avoid leaking email existence
		return nil
	}

	resetToken := uuid.New().String()
	s.redis.Set(ctx, fmt.Sprintf("password_reset:%s", resetToken), user.ID.String(), 15*time.Minute)

	// Send reset email (best-effort; email send errors are logged)
	go func() {
		if err := s.emailSvc.SendVerificationCode(email, resetToken); err != nil {
			slog.Error("failed to send password reset email", "email", email, "error", err)
		}
	}()

	return nil
}

// ResetPassword validates the reset token and updates the user's password.
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	userIDStr, err := s.redis.Get(ctx, fmt.Sprintf("password_reset:%s", token)).Result()
	if err == redis.Nil {
		return errors.NewValidationError("invalid or expired reset token")
	}
	if err != nil {
		return errors.NewInternalError("failed to validate reset token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.NewInternalError("invalid reset token")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.NewInternalError("failed to find user")
	}
	if user == nil {
		return errors.NewNotFoundError("user not found")
	}

	if err := user.ChangePassword("", newPassword); err != nil {
		if err == domain.ErrWeakPassword {
			return errors.NewValidationError(err.Error())
		}
		return errors.NewInternalError("failed to reset password")
	}

	s.redis.Del(ctx, fmt.Sprintf("password_reset:%s", token))

	return s.userRepo.Update(ctx, user)
}

// generateVerificationCode generates a cryptographically secure 6-digit code.
func generateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
