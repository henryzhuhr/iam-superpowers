package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/repository"
	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
	"github.com/henryzhuhr/iam-superpowers/internal/common/email"
	"github.com/henryzhuhr/iam-superpowers/internal/common/jwt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock user repository
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) FindByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) ListUsers(ctx context.Context, filter repository.ListUsersFilter) ([]*domain.User, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.User), args.Int(1), args.Error(2)
}

func setupAuthService(t *testing.T) (*AuthService, *mockUserRepo, *redis.Client) {
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	jwtSvc := jwt.New(config.JWTConfig{
		Secret:          "test-secret",
		AccessTokenTTL:  900,
		RefreshTokenTTL: 604800,
	})
	emailSvc := email.New(config.SMTPConfig{Host: "localhost", Port: 1025})

	repo := &mockUserRepo{}
	svc := NewAuthService(repo, jwtSvc, emailSvc, redisClient)
	return svc, repo, redisClient
}

func TestAuthService_Register_Success(t *testing.T) {
	svc, repo, _ := setupAuthService(t)
	tenantID := uuid.New()

	repo.On("FindByEmailAndTenant", mock.Anything, "test@example.com", tenantID).Return((*domain.User)(nil), nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := svc.Register(context.Background(), tenantID, "test@example.com", "Password1")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	repo.AssertExpectations(t)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	svc, repo, _ := setupAuthService(t)
	tenantID := uuid.New()
	existingUser, _ := domain.NewUser(tenantID, "test@example.com", "Password1")

	repo.On("FindByEmailAndTenant", mock.Anything, "test@example.com", tenantID).Return(existingUser, nil)

	user, err := svc.Register(context.Background(), tenantID, "test@example.com", "Password1")
	assert.Error(t, err)
	assert.Nil(t, user)
	repo.AssertExpectations(t)
}

func TestAuthService_Register_WeakPassword(t *testing.T) {
	svc, repo, _ := setupAuthService(t)
	tenantID := uuid.New()

	repo.On("FindByEmailAndTenant", mock.Anything, "test@example.com", tenantID).Return((*domain.User)(nil), nil)

	user, err := svc.Register(context.Background(), tenantID, "test@example.com", "weak")
	assert.Error(t, err)
	assert.Nil(t, user)
	repo.AssertExpectations(t)
}
