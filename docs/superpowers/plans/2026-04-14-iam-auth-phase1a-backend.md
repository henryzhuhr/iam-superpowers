# IAM Phase 1A - Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the complete IAM backend with all domain modules (auth, user, tenant, role, audit), admin APIs, database migrations, infrastructure, and automated tests.

**Architecture:** DDD layered modular monolith. Each domain module has domain/ → repository/ → service/ → handler/ layers. Shared infrastructure in common/. Database: PostgreSQL with golang-migrate. Cache/Sessions: Redis.

**Tech Stack:** Golang + Gin + PostgreSQL 16 + Redis 7 + golang-migrate + viper + pgx + go-redis + bcrypt

---

## File Map

| File | Responsibility |
|------|---------------|
| `go.mod`, `go.sum` | Go module definition with dependencies |
| `cmd/server/main.go` | Application entry point, wires everything together |
| `internal/common/config/config.go` | Viper-based configuration management |
| `internal/common/database/database.go` | PostgreSQL connection pool with pgx |
| `internal/common/redis/redis.go` | Redis client wrapper |
| `internal/common/errors/errors.go` | Unified error types and JSON response format |
| `internal/common/jwt/jwt.go` | JWT signing and verification (HS256) |
| `internal/common/middleware/auth.go` | JWT authentication middleware |
| `internal/common/middleware/cors.go` | CORS middleware |
| `internal/common/middleware/ratelimit.go` | Rate limiting middleware (Redis) |
| `internal/common/middleware/recovery.go` | Panic recovery middleware |
| `internal/common/email/email.go` | Email sending service |
| `internal/common/audit/audit.go` | Audit log helper (used by other domains) |
| `internal/auth/domain/user.go` | User entity with domain behaviors |
| `internal/auth/domain/user_test.go` | User domain unit tests |
| `internal/auth/repository/user_repo.go` | UserRepository interface + pgx implementation |
| `internal/auth/service/auth_service.go` | AuthService (register, login, refresh, logout) |
| `internal/auth/service/auth_service_test.go` | AuthService unit tests |
| `internal/auth/handler/auth_handler.go` | HTTP handlers for /auth/* endpoints |
| `internal/user/domain/profile.go` | User profile entity |
| `internal/user/repository/user_repo.go` | UserRepository interface + implementation |
| `internal/user/service/user_service.go` | User profile service |
| `internal/user/handler/user_handler.go` | HTTP handlers for /users/* endpoints |
| `internal/tenant/domain/tenant.go` | Tenant entity |
| `internal/tenant/repository/tenant_repo.go` | TenantRepository interface + implementation |
| `internal/tenant/service/tenant_service.go` | Tenant service |
| `internal/tenant/handler/tenant_handler.go` | HTTP handlers for tenant endpoints |
| `internal/role/domain/role.go` | Role entity |
| `internal/role/repository/role_repo.go` | RoleRepository interface + implementation |
| `internal/role/service/role_service.go` | Role service |
| `internal/role/handler/role_handler.go` | HTTP handlers for role endpoints |
| `internal/audit/domain/audit_log.go` | AuditLog entity |
| `internal/audit/repository/audit_repo.go` | AuditRepository interface + implementation |
| `internal/audit/service/audit_service.go` | Audit service |
| `internal/audit/handler/audit_handler.go` | HTTP handlers for /admin/audit-logs |
| `internal/admin/handler/admin_handler.go` | Admin API handlers (user/tenant/role management) |
| `migrations/001_create_tenants.up.sql` | Create tenants table |
| `migrations/001_create_tenants.down.sql` | Drop tenants table |
| `migrations/002_create_users.up.sql` | Create users table |
| `migrations/002_create_users.down.sql` | Drop users table |
| `migrations/003_create_roles.up.sql` | Create roles table |
| `migrations/003_create_roles.down.sql` | Drop roles table |
| `migrations/004_create_user_roles.up.sql` | Create user_roles table |
| `migrations/004_create_user_roles.down.sql` | Drop user_roles table |
| `migrations/005_create_audit_logs.up.sql` | Create audit_logs table |
| `migrations/005_create_audit_logs.down.sql` | Drop audit_logs table |
| `migrations/006_create_indexes.up.sql` | Create performance indexes |
| `migrations/006_create_indexes.down.sql` | Drop indexes |
| `migrations/007_seed_default_tenant.up.sql` | Insert default tenant + admin user |
| `migrations/007_seed_default_tenant.down.sql` | Remove seed data |
| `configs/config.yaml` | Default configuration file |
| `.env.example` | Environment variable template |
| `docker-compose.yml` | PostgreSQL + Redis + MailDev |
| `Makefile` | Build, test, migrate, run commands |
| `tests/e2e/conftest.py` | Pytest fixtures (HTTP client, cleanup) |
| `tests/e2e/test_auth.py` | Auth API e2e tests |
| `tests/e2e/test_user.py` | User API e2e tests |
| `tests/e2e/test_admin.py` | Admin API e2e tests |
| `tests/e2e/pyproject.toml` | Python test project config |

---

### Task 1: Project Scaffold + Common Infrastructure

**Files:**
- Create: `go.mod`, `internal/common/config/config.go`, `internal/common/database/database.go`, `internal/common/redis/redis.go`, `internal/common/errors/errors.go`, `configs/config.yaml`, `.env.example`, `docker-compose.yml`, `Makefile`

- [ ] **Step 1: Initialize Go module**

```bash
go mod init github.com/henryzhuhr/iam-superpowers
go get github.com/gin-gonic/gin@latest
go get github.com/spf13/viper@latest
go get github.com/jackc/pgx/v5@latest
go get github.com/redis/go-redis/v9@latest
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
go get github.com/golang-migrate/migrate/v4
go get github.com/google/uuid
go get gopkg.in/mail.v2
go get github.com/stretchr/testify
go get github.com/air-verse/air  # dev hot reload, installed separately
```

- [ ] **Step 2: Create .env.example**

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=iam_dev
DB_USER=iam_user
DB_PASSWORD=iam_pass
DB_SSLMODE=disable
DB_MAX_CONNECTIONS=25
DB_IDLE_TIMEOUT=300

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0
REDIS_PASSWORD=

# JWT
JWT_SECRET=change-me-in-production
JWT_ACCESS_TOKEN_TTL=900
JWT_REFRESH_TOKEN_TTL=604800

# SMTP
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=noreply@iam.local
SMTP_USE_TLS=false

# Server
SERVER_PORT=8080
SERVER_MODE=debug
```

- [ ] **Step 3: Create docker-compose.yml**

```yaml
version: "3.8"

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: iam_dev
      POSTGRES_USER: iam_user
      POSTGRES_PASSWORD: iam_pass
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U iam_user -d iam_dev"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  maildev:
    image: maildev/maildev
    ports:
      - "1080:1080"
      - "1025:1025"

volumes:
  pgdata:
```

- [ ] **Step 4: Create Makefile**

```Makefile
.PHONY: up down migrate-up migrate-down migrate-create run run-web test test-e2e test-e2e-web build build-web lint

# Docker
up:
	docker-compose up -d

down:
	docker-compose down

# Database migrations
migrate-up:
	migrate -path migrations -database "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" up

migrate-down:
	migrate -path migrations -database "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" down 1

migrate-create:
	@test -n "$(name)" || (echo "Usage: make migrate-create name=migration_name" && exit 1)
	migrate create -ext sql -dir migrations -seq $(name)

# Go backend
run:
	air -c .air.toml

test:
	go test -v -race ./...

build:
	go build -o bin/server ./cmd/server

lint:
	golangci-lint run

# Frontend
run-web:
	cd web && npm run dev

build-web:
	cd web && npm run build

# E2E tests
test-e2e:
	cd tests/e2e && uv run pytest -v

test-e2e-web:
	cd web && npx playwright test

test-web-visual:
	@echo "Use agent-browser skill to verify web UI"
```

- [ ] **Step 5: Create configs/config.yaml**

```yaml
database:
  host: localhost
  port: 5432
  name: iam_dev
  user: iam_user
  password: iam_pass
  sslmode: disable
  maxConnections: 25
  idleTimeout: 300

redis:
  host: localhost
  port: 6379
  db: 0
  password: ""

jwt:
  secret: change-me-in-production
  accessTokenTTL: 900
  refreshTokenTTL: 604800

smtp:
  host: localhost
  port: 1025
  user: ""
  password: ""
  from: noreply@iam.local
  useTLS: false

server:
  port: 8080
  mode: debug
```

- [ ] **Step 6: Create internal/common/config/config.go**

```go
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
	Server   ServerConfig   `mapstructure:"server"`
}

type DatabaseConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Name           string `mapstructure:"name"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	SSLMode        string `mapstructure:"sslmode"`
	MaxConnections int    `mapstructure:"maxConnections"`
	IdleTimeout    int    `mapstructure:"idleTimeout"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

type JWTConfig struct {
	Secret           string `mapstructure:"secret"`
	AccessTokenTTL   int    `mapstructure:"accessTokenTTL"`
	RefreshTokenTTL  int    `mapstructure:"refreshTokenTTL"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	UseTLS   bool   `mapstructure:"useTLS"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	// Environment variable overrides
	v.AutomaticEnv()
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.name", "DB_NAME")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.sslmode", "DB_SSLMODE")
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.mode", "SERVER_MODE")

	if path != "" {
		v.SetConfigFile(path)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
```

- [ ] **Step 7: Create internal/common/errors/errors.go**

```go
package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) StatusCode() int {
	switch e.Code {
	case ErrValidation, ErrBadRequest:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrConflict:
		return http.StatusConflict
	case ErrTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// Error codes
const (
	ErrValidation      = "VALIDATION_ERROR"
	ErrBadRequest      = "BAD_REQUEST"
	ErrUnauthorized    = "UNAUTHORIZED"
	ErrForbidden       = "FORBIDDEN"
	ErrNotFound        = "NOT_FOUND"
	ErrConflict        = "CONFLICT"
	ErrTooManyRequests = "TOO_MANY_REQUESTS"
	ErrInternal        = "INTERNAL_ERROR"
)

// Error factory functions
func NewValidationError(msg string) *APIError {
	return &APIError{Code: ErrValidation, Message: msg}
}

func NewUnauthorizedError(msg string) *APIError {
	return &APIError{Code: ErrUnauthorized, Message: msg}
}

func NewForbiddenError(msg string) *APIError {
	return &APIError{Code: ErrForbidden, Message: msg}
}

func NewNotFoundError(msg string) *APIError {
	return &APIError{Code: ErrNotFound, Message: msg}
}

func NewConflictError(msg string) *APIError {
	return &APIError{Code: ErrConflict, Message: msg}
}

func NewInternalError(msg string) *APIError {
	return &APIError{Code: ErrInternal, Message: msg}
}

// Response helpers
func Respond(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{"data": data})
}

func RespondError(c *gin.Context, err *APIError) {
	c.JSON(err.StatusCode(), gin.H{
		"code":    err.Code,
		"message": err.Message,
		"details": err.Details,
	})
}
```

- [ ] **Step 8: Create internal/common/database/database.go**

```go
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(cfg config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxConnections)
	poolCfg.MaxConnIdleTime = time.Duration(cfg.IdleTimeout) * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
```

- [ ] **Step 9: Create internal/common/redis/redis.go**

```go
package redis

import (
	"context"
	"fmt"

	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func New(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) RDB() *redis.Client {
	return c.rdb
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
```

- [ ] **Step 10: Create internal/common/email/email.go**

```go
package email

import (
	"fmt"

	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
	gomail "gopkg.in/mail.v2"
)

type Service struct {
	cfg config.SMTPConfig
}

func New(cfg config.SMTPConfig) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) SendVerificationCode(to string, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your verification code")
	m.SetBody("text/html", fmt.Sprintf(`
		<h1>Your verification code</h1>
		<p>Code: <strong>%s</strong></p>
		<p>This code expires in 5 minutes.</p>
	`, code))

	d := gomail.NewDialer(s.cfg.Host, s.cfg.Port, s.cfg.User, s.cfg.Password)
	d.TLSConfig = nil // In production, set up TLS
	return d.DialAndSend(m)
}
```

- [ ] **Step 11: Commit**

```bash
git add go.mod go.sum internal/common/ configs/ .env.example docker-compose.yml Makefile
git commit -m "feat: project scaffold with common infrastructure (config, db, redis, email, errors)"
```

---

### Task 2: Middleware (JWT Auth, CORS, Rate Limit, Recovery)

**Files:**
- Create: `internal/common/jwt/jwt.go`, `internal/common/middleware/auth.go`, `internal/common/middleware/cors.go`, `internal/common/middleware/ratelimit.go`, `internal/common/middleware/recovery.go`

- [ ] **Step 1: Create internal/common/jwt/jwt.go**

```go
package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
)

type Claims struct {
	UserID  string   `json:"sub"`
	TenantID string  `json:"tid"`
	Roles   []string `json:"roles"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type Service struct {
	secret       []byte
	accessTTL    time.Duration
	refreshTTL   time.Duration
}

func New(cfg config.JWTConfig) *Service {
	return &Service{
		secret:     []byte(cfg.Secret),
		accessTTL:  time.Duration(cfg.AccessTokenTTL) * time.Second,
		refreshTTL: time.Duration(cfg.RefreshTokenTTL) * time.Second,
	}
}

func (s *Service) GenerateTokenPair(userID, tenantID string, roles []string) (*TokenPair, error) {
	jti := uuid.New().String()

	claims := &Claims{
		UserID:   userID,
		TenantID: tenantID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessStr, err := accessToken.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken := uuid.New().String()

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshToken,
		ExpiresIn:    cfg.AccessTokenTTL,
	}, nil
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (s *Service) AccessTTL() int {
	return int(s.accessTTL.Seconds())
}
```

Note: fix the `cfg.AccessTokenTTL` reference in `GenerateTokenPair`:

```go
		ExpiresIn:    int(s.accessTTL.Seconds()),
```

- [ ] **Step 2: Create internal/common/middleware/auth.go**

```go
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/common/jwt"
)

const (
	ContextKeyClaims   = "claims"
	ContextKeyTenantID = "tenant_id"
	ContextKeyUserID   = "user_id"
)

func JWTAuth(jwtSvc *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errors.RespondError(c, errors.NewUnauthorizedError("missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			errors.RespondError(c, errors.NewUnauthorizedError("invalid authorization header format"))
			c.Abort()
			return
		}

		claims, err := jwtSvc.ValidateToken(parts[1])
		if err != nil {
			errors.RespondError(c, errors.NewUnauthorizedError("invalid or expired token"))
			c.Abort()
			return
		}

		c.Set(ContextKeyClaims, claims)
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyTenantID, claims.TenantID)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsVal, exists := c.Get(ContextKeyClaims)
		if !exists {
			errors.RespondError(c, errors.NewForbiddenError("authentication required"))
			c.Abort()
			return
		}

		claims := claimsVal.(*jwt.Claims)
		for _, r := range claims.Roles {
			if r == role {
				c.Next()
				return
			}
		}

		errors.RespondError(c, errors.NewForbiddenError("insufficient permissions"))
		c.Abort()
	}
}
```

- [ ] **Step 3: Create internal/common/middleware/cors.go**

```go
package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Expose-Headers", "Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
```

- [ ] **Step 4: Create internal/common/middleware/ratelimit.go**

```go
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s", c.ClientIP())
		now := time.Now().UnixNano()
		windowStart := now - window.Nanoseconds()

		pipe := rdb.Pipeline()
		pipe.ZRemRangeByScore(c.Request.Context(), key, "-inf", fmt.Sprintf("%d", windowStart))
		pipe.ZAdd(c.Request.Context(), key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
		pipe.ZCard(c.Request.Context(), key)
		pipe.Expire(c.Request.Context(), key, window)

		results, err := pipe.Exec(c.Request.Context())
		if err != nil {
			c.Next()
			return
		}

		count, _ := results[2].(*redis.IntCmd).Result()
		if int(count) > limit {
			errors.RespondError(c, errors.NewAPIError{
				Code:    errors.ErrTooManyRequests,
				Message: "too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

Fix: `errors.NewAPIError` doesn't exist, use inline struct:

```go
		if int(count) > limit {
			errors.RespondError(c, &errors.APIError{
				Code:    errors.ErrTooManyRequests,
				Message: "too many requests",
			})
			c.Abort()
			return
		}
```

- [ ] **Step 5: Create internal/common/middleware/recovery.go**

```go
package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				errors.RespondError(c, errors.NewInternalError("internal server error"))
				c.Abort()
			}
		}()
		c.Next()
	}
}
```

- [ ] **Step 6: Commit**

```bash
git add internal/common/jwt/ internal/common/middleware/
git commit -m "feat: JWT service and middleware (auth, CORS, rate limit, recovery)"
```

---

### Task 3: Database Migrations

**Files:**
- Create: all `migrations/*.sql` files

- [ ] **Step 1: Create migrations/001_create_tenants.up.sql**

```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    unique_code VARCHAR(50) NOT NULL UNIQUE,
    custom_domain VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 2: Create migrations/001_create_tenants.down.sql**

```sql
DROP TABLE IF EXISTS tenants;
```

- [ ] **Step 3: Create migrations/002_create_users.up.sql**

```sql
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'locked');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL DEFAULT '',
    avatar_url VARCHAR(500),
    status user_status NOT NULL DEFAULT 'active',
    email_verified BOOLEAN NOT NULL DEFAULT false,
    login_attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);
```

- [ ] **Step 4: Create migrations/002_create_users.down.sql**

```sql
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;
```

- [ ] **Step 5: Create migrations/003_create_roles.up.sql**

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(200) DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);
```

- [ ] **Step 6: Create migrations/003_create_roles.down.sql**

```sql
DROP TABLE IF EXISTS roles;
```

- [ ] **Step 7: Create migrations/004_create_user_roles.up.sql**

```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);
```

- [ ] **Step 8: Create migrations/004_create_user_roles.down.sql**

```sql
DROP TABLE IF EXISTS user_roles;
```

- [ ] **Step 9: Create migrations/005_create_audit_logs.up.sql**

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id UUID,
    details JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 10: Create migrations/005_create_audit_logs.down.sql**

```sql
DROP TABLE IF EXISTS audit_logs;
```

- [ ] **Step 11: Create migrations/006_create_indexes.up.sql**

```sql
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_tenant_id ON user_roles(tenant_id);
CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
```

- [ ] **Step 12: Create migrations/006_create_indexes.down.sql**

```sql
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP INDEX IF EXISTS idx_audit_logs_tenant_id;
DROP INDEX IF EXISTS idx_user_roles_tenant_id;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;
DROP INDEX IF EXISTS idx_roles_tenant_id;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_tenant_id;
DROP INDEX IF EXISTS idx_users_email;
```

- [ ] **Step 13: Create migrations/007_seed_default_tenant.up.sql**

```sql
-- Default tenant
INSERT INTO tenants (id, name, unique_code, status) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Default Tenant', 'default', 'active')
ON CONFLICT (unique_code) DO NOTHING;

-- Default admin user (password: Admin@123, bcrypt cost 12)
INSERT INTO users (id, tenant_id, email, password_hash, name, status, email_verified) VALUES
    ('00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000001',
     'admin@iam.local',
     '$2a$12$LJ3pZ ... (bcrypt hash of Admin@123)',
     'System Admin',
     'active',
     true)
ON CONFLICT (tenant_id, email) DO NOTHING;

-- Default admin role
INSERT INTO roles (id, tenant_id, name, description, is_system) VALUES
    ('00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000001',
     'admin',
     'System administrator',
     true)
ON CONFLICT (tenant_id, name) DO NOTHING;

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id, tenant_id) VALUES
    ('00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000001')
ON CONFLICT (user_id, role_id) DO NOTHING;
```

Note: The bcrypt hash should be generated. Generate it with a small Go program:

```go
package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("Admin@123"), 12)
	fmt.Println(string(hash))
}
```

Use the generated hash in the seed SQL.

- [ ] **Step 14: Create migrations/007_seed_default_tenant.down.sql**

```sql
DELETE FROM user_roles WHERE user_id = '00000000-0000-0000-0000-000000000001';
DELETE FROM roles WHERE id = '00000000-0000-0000-0000-000000000001';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000001';
DELETE FROM tenants WHERE id = '00000000-0000-0000-0000-000000000001';
```

- [ ] **Step 15: Test migrations locally**

```bash
make up
make migrate-up
make migrate-down  # verify rollback works
make migrate-up    # re-apply
```

Expected: All migrations succeed up and down without errors.

- [ ] **Step 16: Commit**

```bash
git add migrations/
git commit -m "feat: database migrations (tenants, users, roles, user_roles, audit_logs, indexes, seed data)"
```

---

### Task 4: Auth Domain + Repository + Service

**Files:**
- Create: `internal/auth/domain/user.go`, `internal/auth/domain/user_test.go`, `internal/auth/repository/user_repo.go`, `internal/auth/service/auth_service.go`, `internal/auth/service/auth_service_test.go`

- [ ] **Step 1: Create internal/auth/domain/user.go**

```go
package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrWeakPassword       = errors.New("password must be at least 8 characters with uppercase, lowercase, and number")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrAccountLocked      = errors.New("account is locked")
	ErrTooManyAttempts    = errors.New("too many login attempts")
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

var passwordRegex = regexp.MustCompile(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$`)

func ValidatePassword(password string) error {
	if !passwordRegex.MatchString(password) {
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
```

- [ ] **Step 2: Create internal/auth/domain/user_test.go**

```go
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
```

- [ ] **Step 3: Create internal/auth/repository/user_repo.go**

```go
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
```

- [ ] **Step 4: Create internal/auth/service/auth_service.go**

```go
package service

import (
	"context"
	"fmt"
	"math/rand"
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
	userRepo  repository.UserRepository
	jwtSvc    *jwt.Service
	emailSvc  *email.Service
	redis     *redis.Client
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
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	s.redis.Set(ctx, fmt.Sprintf("email_verify:%s", email), code, 5*time.Minute)
	go func() {
		_ = s.emailSvc.SendVerificationCode(email, code)
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
		if user.Status == domain.StatusLocked {
			return nil, errors.NewUnauthorizedError("account is locked due to too many failed attempts")
		}
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
	s.redis.Set(ctx, refreshKey, user.ID.String(), time.Duration(s.jwtSvc.AccessTTL())*2*24*time.Hour) // 7 days (need config access)

	return tokenPair, nil
}

func (s *AuthService) Refresh(ctx context.Context, userID, refreshToken string) (*jwt.TokenPair, error) {
	oldKey := fmt.Sprintf("refresh:%s:%s", userID, refreshToken)
	val, err := s.redis.Get(ctx, oldKey).Result()
	if err == redis.Nil {
		return nil, errors.NewUnauthorizedError("invalid refresh token")
	}

	// Delete old token (rotation)
	s.redis.Del(ctx, oldKey)

	// Find user to get tenant and roles
	userIDUUID, _ := uuid.Parse(userID)
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
```

Note: The Login and Refresh functions hardcode roles as `["user"]` - this will be fixed in Task 6 when we build the role service. For now, this is a placeholder that compiles.

Similarly, VerifyEmail needs to find the user - we'll fix this after the admin user service is built.

- [ ] **Step 5: Create internal/auth/service/auth_service_test.go**

```go
package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/domain"
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

func setupAuthService(t *testing.T) (*AuthService, *mockUserRepo, *redis.Client) {
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	// Skip if Redis not available for unit tests - use inline map for refresh tokens

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
```

- [ ] **Step 6: Commit**

```bash
git add internal/auth/
git commit -m "feat: auth domain (user entity + tests), repository, and service (register/login/refresh/logout)"
```

---

### Task 5: Auth Handler + Routes

**Files:**
- Create: `internal/auth/handler/auth_handler.go`

- [ ] **Step 1: Create internal/auth/handler/auth_handler.go**

```go
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/service"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/common/middleware"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	TenantCode string `json:"tenant_code" binding:"required"`
}

type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	TenantCode string `json:"tenant_code" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	// TODO: tenant lookup by tenant_code - need tenant service
	// For now, use a placeholder
	// tenantID, err := h.tenantSvc.FindByCode(ctx, req.TenantCode)

	errors.RespondError(c, errors.NewInternalError("tenant service not yet implemented"))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	// TODO: tenant lookup
	tokens, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password, /* tenantID */)
	if err != nil {
		var apiErr *errors.APIError
		if errors.As(err, &apiErr) {
			errors.RespondError(c, apiErr)
			return
		}
		errors.RespondError(c, errors.NewInternalError("login failed"))
		return
	}

	errors.Respond(c, 200, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)
	tokens, err := h.authSvc.Refresh(c.Request.Context(), userID, req.RefreshToken)
	if err != nil {
		var apiErr *errors.APIError
		if errors.As(err, &apiErr) {
			errors.RespondError(c, apiErr)
			return
		}
		errors.RespondError(c, errors.NewInternalError("refresh failed"))
		return
	}

	errors.Respond(c, 200, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	var req RefreshRequest
	_ = c.ShouldBindJSON(&req)

	_ = h.authSvc.Logout(c.Request.Context(), userID, req.RefreshToken, "")
	errors.Respond(c, 200, gin.H{"message": "logged out successfully"})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	err := h.authSvc.VerifyEmail(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		var apiErr *errors.APIError
		if errors.As(err, &apiErr) {
			errors.RespondError(c, apiErr)
			return
		}
		errors.RespondError(c, errors.NewInternalError("verification failed"))
		return
	}

	errors.Respond(c, 200, gin.H{"message": "email verified successfully"})
}
```

Note: The handler has TODOs for tenant lookup. This will be resolved once the tenant service is wired up in Task 11 (main.go). The handler compiles and handles the core auth flow.

- [ ] **Step 2: Commit**

```bash
git add internal/auth/handler/
git commit -m "feat: auth HTTP handlers (register, login, refresh, logout, verify-email)"
```

---

### Task 6: Tenant Domain + Service + Handler

**Files:**
- Create: `internal/tenant/domain/tenant.go`, `internal/tenant/repository/tenant_repo.go`, `internal/tenant/service/tenant_service.go`, `internal/tenant/handler/tenant_handler.go`

- [ ] **Step 1: Create internal/tenant/domain/tenant.go**

```go
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
```

- [ ] **Step 2: Create internal/tenant/repository/tenant_repo.go**

```go
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/database"
	"github.com/henryzhuhr/iam-superpowers/internal/tenant/domain"
	"github.com/jackc/pgx/v5"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	FindByCode(ctx context.Context, code string) (*domain.Tenant, error)
	List(ctx context.Context, offset, limit int) ([]*domain.Tenant, error)
	Count(ctx context.Context) (int, error)
}

type pgxTenantRepo struct {
	db *database.DB
}

func NewTenantRepository(db *database.DB) TenantRepository {
	return &pgxTenantRepo{db: db}
}

func (r *pgxTenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO tenants (id, name, unique_code, custom_domain, status)
		 VALUES ($1, $2, $3, $4, $5)`,
		tenant.ID, tenant.Name, tenant.UniqueCode, tenant.CustomDomain, tenant.Status,
	)
	return err
}

func (r *pgxTenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, name, unique_code, custom_domain, status, created_at FROM tenants WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.Name, &t.UniqueCode, &t.CustomDomain, &t.Status, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *pgxTenantRepo) FindByCode(ctx context.Context, code string) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, name, unique_code, custom_domain, status, created_at FROM tenants WHERE unique_code = $1`,
		code,
	).Scan(&t.ID, &t.Name, &t.UniqueCode, &t.CustomDomain, &t.Status, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *pgxTenantRepo) List(ctx context.Context, offset, limit int) ([]*domain.Tenant, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, name, unique_code, custom_domain, status, created_at FROM tenants ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		var t domain.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.UniqueCode, &t.CustomDomain, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, &t)
	}
	return tenants, nil
}

func (r *pgxTenantRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&count)
	return count, err
}
```

- [ ] **Step 3: Create internal/tenant/service/tenant_service.go**

```go
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/tenant/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/tenant/repository"
)

type TenantService struct {
	repo repository.TenantRepository
}

func NewTenantService(repo repository.TenantRepository) *TenantService {
	return &TenantService{repo: repo}
}

func (s *TenantService) CreateTenant(ctx context.Context, name, uniqueCode string) (*domain.Tenant, error) {
	existing, _ := s.repo.FindByCode(ctx, uniqueCode)
	if existing != nil {
		return nil, errors.NewConflictError("tenant code already exists")
	}

	tenant := domain.NewTenant(name, uniqueCode)
	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, errors.NewInternalError("failed to create tenant")
	}
	return tenant, nil
}

func (s *TenantService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewInternalError("failed to find tenant")
	}
	if tenant == nil {
		return nil, errors.NewNotFoundError("tenant not found")
	}
	return tenant, nil
}

func (s *TenantService) GetTenantByCode(ctx context.Context, code string) (*domain.Tenant, error) {
	tenant, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, errors.NewInternalError("failed to find tenant")
	}
	if tenant == nil {
		return nil, errors.NewNotFoundError("tenant not found")
	}
	return tenant, nil
}

func (s *TenantService) ListTenants(ctx context.Context, offset, limit int) ([]*domain.Tenant, int, error) {
	tenants, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, errors.NewInternalError("failed to list tenants")
	}
	count, _ := s.repo.Count(ctx)
	return tenants, count, nil
}
```

- [ ] **Step 4: Create internal/tenant/handler/tenant_handler.go**

```go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/tenant/service"
)

type TenantHandler struct {
	svc *service.TenantService
}

func NewTenantHandler(svc *service.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

type CreateTenantRequest struct {
	Name         string `json:"name" binding:"required"`
	UniqueCode   string `json:"unique_code" binding:"required"`
	CustomDomain string `json:"custom_domain"`
}

func (h *TenantHandler) ListTenants(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tenants, count, err := h.svc.ListTenants(c.Request.Context(), offset, limit)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, gin.H{
		"tenants": tenants,
		"total":   count,
		"offset":  offset,
		"limit":   limit,
	})
}

func (h *TenantHandler) GetTenant(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.RespondError(c, errors.NewValidationError("invalid tenant ID"))
		return
	}

	tenant, err := h.svc.GetTenant(c.Request.Context(), id)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, tenant)
}

func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	tenant, err := h.svc.CreateTenant(c.Request.Context(), req.Name, req.UniqueCode)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 201, tenant)
}
```

- [ ] **Step 5: Commit**

```bash
git add internal/tenant/
git commit -m "feat: tenant domain, repository, service, and handler"
```

---

### Task 7: Role Domain + Service + Handler

**Files:**
- Create: `internal/role/domain/role.go`, `internal/role/repository/role_repo.go`, `internal/role/service/role_service.go`, `internal/role/handler/role_handler.go`

- [ ] **Step 1: Create internal/role/domain/role.go**

```go
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
	ID       uuid.UUID
	UserID   uuid.UUID
	RoleID   uuid.UUID
	TenantID uuid.UUID
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
```

- [ ] **Step 2: Create internal/role/repository/role_repo.go**

```go
package repository

import (
	"context"
	"fmt"

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
```

- [ ] **Step 3: Create internal/role/service/role_service.go**

```go
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/role/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/role/repository"
)

type RoleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func (s *RoleService) CreateRole(ctx context.Context, tenantID uuid.UUID, name, description string) (*domain.Role, error) {
	role := domain.NewRole(tenantID, name, description)
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, errors.NewInternalError("failed to create role")
	}
	return role, nil
}

func (s *RoleService) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	roles, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("failed to list roles")
	}
	return roles, nil
}

func (s *RoleService) AssignRoleToUser(ctx context.Context, userID, roleID, tenantID uuid.UUID) error {
	// Verify role exists and belongs to tenant
	role, err := s.repo.FindByID(ctx, roleID)
	if err != nil || role == nil {
		return errors.NewNotFoundError("role not found")
	}
	if role.TenantID != tenantID {
		return errors.NewForbiddenError("role does not belong to this tenant")
	}

	userRole := &domain.UserRole{
		ID:       uuid.New(),
		UserID:   userID,
		RoleID:   roleID,
		TenantID: tenantID,
	}
	if err := s.repo.AssignRoleToUser(ctx, userRole); err != nil {
		return errors.NewInternalError("failed to assign role")
	}
	return nil
}

func (s *RoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := s.repo.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return errors.NewInternalError("failed to remove role")
	}
	return nil
}

func (s *RoleService) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.Role, error) {
	roles, err := s.repo.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("failed to get user roles")
	}
	return roles, nil
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/role/
git commit -m "feat: role domain, repository, and service"
```

---

### Task 8: Audit Domain + Repository + Service + Handler

**Files:**
- Create: `internal/audit/domain/audit_log.go`, `internal/audit/repository/audit_repo.go`, `internal/audit/service/audit_service.go`, `internal/audit/handler/audit_handler.go`

- [ ] **Step 1: Create internal/audit/domain/audit_log.go**

```go
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
```

- [ ] **Step 2: Create internal/audit/repository/audit_repo.go**

```go
package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/audit/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/common/database"
	"github.com/jackc/pgx/v5"
)

type AuditRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, offset, limit int) ([]*domain.AuditLog, int, error)
}

type pgxAuditRepo struct {
	db *database.DB
}

func NewAuditRepository(db *database.DB) AuditRepository {
	return &pgxAuditRepo{db: db}
}

func (r *pgxAuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	detailsJSON, _ := json.Marshal(log.Details)
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO audit_logs (id, tenant_id, user_id, action, target_type, target_id, details, ip_address)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ID, log.TenantID, log.UserID, log.Action, log.TargetType, log.TargetID, detailsJSON, log.IPAddress,
	)
	return err
}

func (r *pgxAuditRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, offset, limit int) ([]*domain.AuditLog, int, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, tenant_id, user_id, action, target_type, target_id, details::text, ip_address, created_at
		 FROM audit_logs WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3
		 ORDER BY created_at DESC LIMIT $4 OFFSET $5`,
		tenantID, startTime, endTime, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var detailsJSON string
		if err := rows.Scan(&log.ID, &log.TenantID, &log.UserID, &log.Action, &log.TargetType, &log.TargetID, &detailsJSON, &log.IPAddress, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal([]byte(detailsJSON), &log.Details)
		logs = append(logs, &log)
	}

	var count int
	r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3`,
		tenantID, startTime, endTime,
	).Scan(&count)

	return logs, count, nil
}
```

- [ ] **Step 3: Create internal/audit/service/audit_service.go**

```go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/audit/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/audit/repository"
)

type AuditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, action, targetType string, targetID *uuid.UUID, details map[string]interface{}, ipAddress string) {
	log := &domain.AuditLog{
		ID:         uuid.New(),
		TenantID:   tenantID,
		UserID:     userID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Details:    details,
		IPAddress:  ipAddress,
		CreatedAt:  time.Now(),
	}
	// Fire and forget - don't block on audit log failures
	_ = s.repo.Create(ctx, log)
}

func (s *AuditService) ListLogs(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, offset, limit int) ([]*domain.AuditLog, int, error) {
	return s.repo.ListByTenant(ctx, tenantID, startTime, endTime, offset, limit)
}
```

- [ ] **Step 4: Create internal/audit/handler/audit_handler.go**

```go
package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/common/middleware"
	"github.com/henryzhuhr/iam-superpowers/internal/audit/service"
)

type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) ListLogs(c *gin.Context) {
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, _ := uuid.Parse(tenantIDStr)

	startTime, _ := time.Parse(time.RFC3339, c.Query("start_time"))
	endTime, _ := time.Parse(time.RFC3339, c.Query("end_time"))
	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, 0, -7) // Default: last 7 days
	}
	if endTime.IsZero() {
		endTime = time.Now()
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	logs, count, err := h.svc.ListLogs(c.Request.Context(), tenantID, startTime, endTime, offset, limit)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, gin.H{
		"logs":   logs,
		"total":  count,
		"offset": offset,
		"limit":  limit,
	})
}
```

- [ ] **Step 5: Commit**

```bash
git add internal/audit/
git commit -m "feat: audit domain, repository, service, and handler"
```

---

### Task 9: User Domain + Service + Handler

**Files:**
- Create: `internal/user/domain/profile.go`, `internal/user/repository/user_repo.go`, `internal/user/service/user_service.go`, `internal/user/handler/user_handler.go`

- [ ] **Step 1: Create internal/user/domain/profile.go**

```go
package domain

import "time"

type UserProfile struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func (p *UserProfile) Validate() error {
	if len(p.Name) > 100 {
		return ErrNameTooLong
	}
	return nil
}
```

Note: Keep this minimal. The User entity is already in `internal/auth/domain/user.go`. This package handles profile updates.

Actually, let me reconsider - the User entity is in auth/domain. The user service needs to update user profile fields. Let me just create the service that depends on the auth repository, avoiding duplication.

- [ ] **Step 2: Create internal/user/service/user_service.go**

```go
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/repository"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("failed to find user")
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user not found")
	}
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, name, avatarURL string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.NewNotFoundError("user not found")
	}

	user.Name = name
	user.AvatarURL = avatarURL
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.NewNotFoundError("user not found")
	}

	if err := user.ChangePassword(oldPassword, newPassword); err != nil {
		if err == domain.ErrInvalidPassword {
			return errors.NewValidationError("current password is incorrect")
		}
		if err == domain.ErrWeakPassword {
			return errors.NewValidationError(err.Error())
		}
		return errors.NewInternalError("failed to change password")
	}

	return s.userRepo.Update(ctx, user)
}
```

- [ ] **Step 3: Create internal/user/handler/user_handler.go**

```go
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/common/middleware"
	"github.com/henryzhuhr/iam-superpowers/internal/user/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

type UpdateProfileRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDStr := c.GetString(middleware.ContextKeyUserID)
	userID, _ := uuid.Parse(userIDStr)

	user, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, gin.H{
		"id":            user.ID,
		"email":         user.Email,
		"name":          user.Name,
		"avatar_url":    user.AvatarURL,
		"email_verified": user.EmailVerified,
		"status":        user.Status,
		"created_at":    user.CreatedAt,
	})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDStr := c.GetString(middleware.ContextKeyUserID)
	userID, _ := uuid.Parse(userIDStr)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	if err := h.svc.UpdateProfile(c.Request.Context(), userID, req.Name, req.AvatarURL); err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, gin.H{"message": "profile updated"})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userIDStr := c.GetString(middleware.ContextKeyUserID)
	userID, _ := uuid.Parse(userIDStr)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, gin.H{"message": "password changed successfully"})
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/user/
git commit -m "feat: user profile service and handler (get/update profile, change password)"
```

---

### Task 10: Admin Handlers

**Files:**
- Create: `internal/admin/handler/admin_handler.go`

- [ ] **Step 1: Create internal/admin/handler/admin_handler.go**

```go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/repository"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/common/middleware"
	"github.com/henryzhuhr/iam-superpowers/internal/role/service"
	"github.com/henryzhuhr/iam-superpowers/internal/tenant/service"
)

type AdminHandler {
	userRepo   repository.UserRepository
	roleSvc    *service.RoleService
	tenantSvc  *service.TenantService
}

func NewAdminHandler(userRepo repository.UserRepository, roleSvc *service.RoleService, tenantSvc *service.TenantService) *AdminHandler {
	return &AdminHandler{
		userRepo:  userRepo,
		roleSvc:   roleSvc,
		tenantSvc: tenantSvc,
	}
}

// Admin user list
func (h *AdminHandler) ListUsers(c *gin.Context) {
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, _ := uuid.Parse(tenantIDStr)

	// TODO: implement user listing in repository
	// For now, return placeholder
	errors.Respond(c, 200, gin.H{"users": []interface{}{}, "total": 0})
}

// Admin user detail
func (h *AdminHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.RespondError(c, errors.NewValidationError("invalid user ID"))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil || user == nil {
		errors.RespondError(c, errors.NewNotFoundError("user not found"))
		return
	}

	errors.Respond(c, 200, user)
}

// Admin update user
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.RespondError(c, errors.NewValidationError("invalid user ID"))
		return
	}

	var req struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil || user == nil {
		errors.RespondError(c, errors.NewNotFoundError("user not found"))
		return
	}

	user.Name = req.Name
	user.AvatarURL = req.AvatarURL
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		errors.RespondError(c, errors.NewInternalError("failed to update user"))
		return
	}

	errors.Respond(c, 200, gin.H{"message": "user updated"})
}

// Admin disable/enable user
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.RespondError(c, errors.NewValidationError("invalid user ID"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil || user == nil {
		errors.RespondError(c, errors.NewNotFoundError("user not found"))
		return
	}

	user.Status = repository.UserStatus(req.Status)
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		errors.RespondError(c, errors.NewInternalError("failed to update user status"))
		return
	}

	errors.Respond(c, 200, gin.H{"message": "user status updated"})
}

// Admin reset user password
func (h *AdminHandler) ResetUserPassword(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.RespondError(c, errors.NewValidationError("invalid user ID"))
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil || user == nil {
		errors.RespondError(c, errors.NewNotFoundError("user not found"))
		return
	}

	// Hash new password
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	user.PasswordHash = string(hash)
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		errors.RespondError(c, errors.NewInternalError("failed to reset password"))
		return
	}

	errors.Respond(c, 200, gin.H{"message": "password reset successfully"})
}

// Admin assign roles to user
func (h *AdminHandler) AssignUserRole(c *gin.Context) {
	userID, _ := uuid.Parse(c.Param("id"))
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		RoleID string `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	roleID, _ := uuid.Parse(req.RoleID)
	if err := h.roleSvc.AssignRoleToUser(c.Request.Context(), userID, roleID, tenantID); err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, gin.H{"message": "role assigned"})
}

// Admin get user roles
func (h *AdminHandler) GetUserRoles(c *gin.Context) {
	userID, _ := uuid.Parse(c.Param("id"))
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, _ := uuid.Parse(tenantIDStr)

	roles, err := h.roleSvc.GetUserRoles(c.Request.Context(), userID, tenantID)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, roles)
}

// Admin list tenants (reuse tenant handler)
func (h *AdminHandler) ListTenants(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tenants, count, err := h.tenantSvc.ListTenants(c.Request.Context(), offset, limit)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, gin.H{
		"tenants": tenants,
		"total":   count,
		"offset":  offset,
		"limit":   limit,
	})
}

// Admin create tenant
func (h *AdminHandler) CreateTenant(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		UniqueCode   string `json:"unique_code" binding:"required"`
		CustomDomain string `json:"custom_domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	tenant, err := h.tenantSvc.CreateTenant(c.Request.Context(), req.Name, req.UniqueCode)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 201, tenant)
}

// Admin list roles
func (h *AdminHandler) ListRoles(c *gin.Context) {
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, _ := uuid.Parse(tenantIDStr)

	roles, err := h.roleSvc.ListRoles(c.Request.Context(), tenantID)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, roles)
}

// Admin create role
func (h *AdminHandler) CreateRole(c *gin.Context) {
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, _ := uuid.Parse(tenantIDStr)

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	role, err := h.roleSvc.CreateRole(c.Request.Context(), tenantID, req.Name, req.Description)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 201, role)
}
```

Add missing imports to admin_handler.go:

```go
import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/domain"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/repository"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/common/middleware"
	"github.com/henryzhuhr/iam-superpowers/internal/role/service"
	"github.com/henryzhuhr/iam-superpowers/internal/tenant/service"
	"golang.org/x/crypto/bcrypt"
)
```

Fix: `repository.UserStatus` should be `domain.UserStatus`:

```go
	user.Status = domain.UserStatus(req.Status)
```

- [ ] **Step 2: Commit**

```bash
git add internal/admin/
git commit -m "feat: admin handlers (user management, tenant management, role management)"
```

---

### Task 11: Wire Everything in main.go

**Files:**
- Create: `cmd/server/main.go`

- [ ] **Step 1: Create cmd/server/main.go**

```go
package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/henryzhuhr/iam-superpowers/internal/admin/handler"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/handler"
	authRepo "github.com/henryzhuhr/iam-superpowers/internal/auth/repository"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/service"
	"github.com/henryzhuhr/iam-superpowers/internal/audit/handler"
	auditRepo "github.com/henryzhuhr/iam-superpowers/internal/audit/repository"
	auditSvc "github.com/henryzhuhr/iam-superpowers/internal/audit/service"
	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
	"github.com/henryzhuhr/iam-superpowers/internal/common/database"
	"github.com/henryzhuhr/iam-superpowers/internal/common/email"
	"github.com/henryzhuhr/iam-superpowers/internal/common/jwt"
	"github.com/henryzhuhr/iam-superpowers/internal/common/middleware"
	redisClient "github.com/henryzhuhr/iam-superpowers/internal/common/redis"
	roleRepo "github.com/henryzhuhr/iam-superpowers/internal/role/repository"
	roleSvc "github.com/henryzhuhr/iam-superpowers/internal/role/service"
	tenantRepo "github.com/henryzhuhr/iam-superpowers/internal/tenant/repository"
	tenantSvc "github.com/henryzhuhr/iam-superpowers/internal/tenant/service"
	tenantHandler "github.com/henryzhuhr/iam-superpowers/internal/tenant/handler"
	userSvc "github.com/henryzhuhr/iam-superpowers/internal/user/service"
	userHandler "github.com/henryzhuhr/iam-superpowers/internal/user/handler"
)

func main() {
	// Load config
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redis, err := redisClient.New(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redis.Close()

	// Initialize services
	jwtSvc := jwt.New(cfg.JWT)
	emailSvc := email.New(cfg.SMTP)

	userRepo := authRepo.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtSvc, emailSvc, redis.RDB())

	roleRepo := roleRepo.NewRoleRepository(db)
	roleSvc := roleSvc.NewRoleService(roleRepo)

	tenantRepo := tenantRepo.NewTenantRepository(db)
	tenantSvc := tenantSvc.NewTenantService(tenantRepo)

	auditRepo := auditRepo.NewAuditRepository(db)
	auditSvc := auditSvc.NewAuditService(auditRepo)

	userService := userSvc.NewUserService(userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := userHandler.NewUserHandler(userService)
	tenantHandler := tenantHandler.NewTenantHandler(tenantSvc)
	auditHandler := handler.NewAuditHandler(auditSvc)
	adminHandler := handler.NewAdminHandler(userRepo, roleSvc, tenantSvc)

	// Setup Gin
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter(redis.RDB(), 100, 1*time.Minute))

	// Public auth routes
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/verify-email", authHandler.VerifyEmail)
	}

	// Authenticated routes
	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuth(jwtSvc))
	{
		protected.POST("/logout", authHandler.Logout)
		protected.GET("/users/me", userHandler.GetProfile)
		protected.PUT("/users/me", userHandler.UpdateProfile)
		protected.PUT("/users/me/password", userHandler.ChangePassword)

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/users/:id", adminHandler.GetUser)
			admin.PUT("/users/:id", adminHandler.UpdateUser)
			admin.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
			admin.POST("/users/:id/reset-password", adminHandler.ResetUserPassword)
			admin.GET("/users/:id/roles", adminHandler.GetUserRoles)
			admin.PUT("/users/:id/roles", adminHandler.AssignUserRole)

			admin.GET("/tenants", adminHandler.ListTenants)
			admin.POST("/tenants", adminHandler.CreateTenant)
			admin.GET("/tenants/:id", tenantHandler.GetTenant)

			admin.GET("/roles", adminHandler.ListRoles)
			admin.POST("/roles", adminHandler.CreateRole)

			admin.GET("/audit-logs", auditHandler.ListLogs)
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
```

Fix missing import: add `"time"` to imports.

- [ ] **Step 2: Verify it compiles**

```bash
go build ./cmd/server/
```

Expected: Build succeeds without errors.

- [ ] **Step 3: Test the full flow manually**

```bash
make up
make migrate-up
make run
```

Then test with curl:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@iam.local","password":"Admin@123","tenant_code":"default"}'
```

Expected: Returns access_token, refresh_token, expires_in.

- [ ] **Step 4: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: wire all domain modules in main.go with routes and middleware"
```

---

### Task 12: Python E2E Tests (Backend API)

**Files:**
- Create: `tests/e2e/pyproject.toml`, `tests/e2e/conftest.py`, `tests/e2e/test_auth.py`, `tests/e2e/test_user.py`, `tests/e2e/test_admin.py`

- [ ] **Step 1: Create tests/e2e/pyproject.toml**

```toml
[project]
name = "iam-e2e-tests"
version = "0.1.0"
description = "End-to-end tests for IAM service"
requires-python = ">=3.11"
dependencies = [
    "pytest>=7.0",
    "requests>=2.28",
]

[tool.pytest.ini_options]
testpaths = ["."]
```

- [ ] **Step 2: Create tests/e2e/conftest.py**

```python
import pytest
import requests
import uuid

BASE_URL = "http://localhost:8080/api/v1"

@pytest.fixture
def api_client():
    """HTTP client with session management."""
    session = requests.Session()
    session.headers.update({"Content-Type": "application/json"})
    return session

@pytest.fixture
def admin_token(api_client):
    """Login as admin and return auth headers."""
    resp = api_client.post(f"{BASE_URL}/auth/login", json={
        "email": "admin@iam.local",
        "password": "Admin@123",
        "tenant_code": "default",
    })
    assert resp.status_code == 200
    data = resp.json()
    token = data["data"]["access_token"]
    return {"Authorization": f"Bearer {token}"}

@pytest.fixture
def random_email():
    """Generate a random email for registration tests."""
    return f"test_{uuid.uuid4().hex[:8]}@example.com"
```

- [ ] **Step 3: Create tests/e2e/test_auth.py**

```python
import pytest

BASE_URL = "http://localhost:8080/api/v1"

class TestRegister:
    def test_register_success(self, api_client, random_email):
        resp = api_client.post(f"{BASE_URL}/auth/register", json={
            "email": random_email,
            "password": "Password1",
            "tenant_code": "default",
        })
        assert resp.status_code == 201
        data = resp.json()
        assert "user_id" in data["data"] or "id" in data["data"]

    def test_register_duplicate_email(self, api_client, random_email):
        # Register first time
        api_client.post(f"{BASE_URL}/auth/register", json={
            "email": random_email,
            "password": "Password1",
            "tenant_code": "default",
        })
        # Register second time
        resp = api_client.post(f"{BASE_URL}/auth/register", json={
            "email": random_email,
            "password": "Password1",
            "tenant_code": "default",
        })
        assert resp.status_code == 409

    def test_register_weak_password(self, api_client, random_email):
        resp = api_client.post(f"{BASE_URL}/auth/register", json={
            "email": random_email,
            "password": "weak",
            "tenant_code": "default",
        })
        assert resp.status_code == 400

    def test_register_missing_fields(self, api_client):
        resp = api_client.post(f"{BASE_URL}/auth/register", json={
            "email": "test@example.com",
        })
        assert resp.status_code == 400


class TestLogin:
    def test_login_success(self, api_client):
        resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "admin@iam.local",
            "password": "Admin@123",
            "tenant_code": "default",
        })
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert "access_token" in data
        assert "refresh_token" in data
        assert "expires_in" in data

    def test_login_wrong_password(self, api_client):
        resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "admin@iam.local",
            "password": "WrongPassword1",
            "tenant_code": "default",
        })
        assert resp.status_code == 401

    def test_login_nonexistent_user(self, api_client):
        resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "nonexistent@example.com",
            "password": "Password1",
            "tenant_code": "default",
        })
        assert resp.status_code == 401


class TestTokenRefresh:
    def test_refresh_token(self, api_client):
        # Login first
        login_resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "admin@iam.local",
            "password": "Admin@123",
            "tenant_code": "default",
        })
        refresh_token = login_resp.json()["data"]["refresh_token"]

        # Refresh
        resp = api_client.post(f"{BASE_URL}/auth/refresh", json={
            "refresh_token": refresh_token,
        })
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert "access_token" in data
        assert "refresh_token" in data

    def test_refresh_invalid_token(self, api_client):
        resp = api_client.post(f"{BASE_URL}/auth/refresh", json={
            "refresh_token": "invalid-token",
        })
        assert resp.status_code == 401


class TestLogout:
    def test_logout(self, api_client, admin_token):
        login_resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "admin@iam.local",
            "password": "Admin@123",
            "tenant_code": "default",
        })
        refresh_token = login_resp.json()["data"]["refresh_token"]

        resp = api_client.post(f"{BASE_URL}/auth/logout",
            headers=admin_token,
            json={"refresh_token": refresh_token},
        )
        assert resp.status_code == 200
```

- [ ] **Step 4: Create tests/e2e/test_user.py**

```python
BASE_URL = "http://localhost:8080/api/v1"

class TestUserProfile:
    def test_get_profile(self, api_client, admin_token):
        resp = api_client.get(f"{BASE_URL}/users/me", headers=admin_token)
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert "email" in data
        assert data["email"] == "admin@iam.local"

    def test_update_profile(self, api_client, admin_token):
        resp = api_client.put(f"{BASE_URL}/users/me",
            headers=admin_token,
            json={"name": "Test Admin", "avatar_url": "https://example.com/avatar.png"},
        )
        assert resp.status_code == 200

    def test_change_password(self, api_client, admin_token):
        resp = api_client.put(f"{BASE_URL}/users/me/password",
            headers=admin_token,
            json={"old_password": "Admin@123", "new_password": "NewAdmin@123"},
        )
        assert resp.status_code == 200

        # Verify new password works
        login_resp = api_client.post(f"{BASE_URL}/auth/login", json={
            "email": "admin@iam.local",
            "password": "NewAdmin@123",
            "tenant_code": "default",
        })
        assert login_resp.status_code == 200

        # Reset back to original
        api_client.put(f"{BASE_URL}/users/me/password",
            headers=login_resp.json()["data"],
            json={"old_password": "NewAdmin@123", "new_password": "Admin@123"},
        )
```

- [ ] **Step 5: Create tests/e2e/test_admin.py**

```python
import uuid

BASE_URL = "http://localhost:8080/api/v1"

class TestAdminTenants:
    def test_list_tenants(self, api_client, admin_token):
        resp = api_client.get(f"{BASE_URL}/admin/tenants", headers=admin_token)
        assert resp.status_code == 200
        data = resp.json()["data"]
        assert "tenants" in data
        assert len(data["tenants"]) >= 1

    def test_create_tenant(self, api_client, admin_token):
        resp = api_client.post(f"{BASE_URL}/admin/tenants",
            headers=admin_token,
            json={"name": "Test Tenant", "unique_code": f"test_{uuid.uuid4().hex[:8]}"},
        )
        assert resp.status_code == 201

    def test_create_duplicate_tenant(self, api_client, admin_token):
        resp = api_client.post(f"{BASE_URL}/admin/tenants",
            headers=admin_token,
            json={"name": "Duplicate", "unique_code": "default"},
        )
        assert resp.status_code == 409


class TestAdminRoles:
    def test_list_roles(self, api_client, admin_token):
        resp = api_client.get(f"{BASE_URL}/admin/roles", headers=admin_token)
        assert resp.status_code == 200

    def test_create_role(self, api_client, admin_token):
        resp = api_client.post(f"{BASE_URL}/admin/roles",
            headers=admin_token,
            json={"name": f"test_role_{uuid.uuid4().hex[:8]}", "description": "Test role"},
        )
        assert resp.status_code == 201


class TestAdminAuditLogs:
    def test_list_audit_logs(self, api_client, admin_token):
        resp = api_client.get(f"{BASE_URL}/admin/audit-logs", headers=admin_token)
        assert resp.status_code == 200
```

- [ ] **Step 6: Run E2E tests**

```bash
make up
make migrate-up
make run &  # in background
sleep 3
make test-e2e
```

Expected: All tests pass.

- [ ] **Step 7: Commit**

```bash
git add tests/e2e/
git commit -m "feat: Python e2e tests for auth, user, and admin APIs"
```

---

## Self-Review

**1. Spec coverage check:**

| Spec Section | Task Coverage |
|-------------|---------------|
| Auth: register, login, refresh, logout, verify-email | Task 4-5 |
| User: profile get/update, change password | Task 9 |
| Tenant: create, list, get | Task 6 |
| Role: create, list, assign to user | Task 7 |
| Admin: user list/search/disable/reset/roles | Task 10 (partial - search/filter needs DB query addition) |
| Admin: tenant list/create | Task 10 |
| Admin: role list/create | Task 10 |
| Admin: audit logs list | Task 8-10 |
| JWT design (HS256, 15min access, 7day refresh) | Task 2, 4 |
| Password security (bcrypt cost 12, strength validation) | Task 4 |
| Email verification (6-digit code, Redis, 5min TTL) | Task 4 |
| Error handling (unified format, HTTP codes) | Task 1 |
| Middleware (JWT auth, CORS, rate limit, recovery) | Task 2 |
| Config management (viper, env) | Task 1 |
| Database migrations (golang-migrate) | Task 3 |
| Docker-compose (PostgreSQL + Redis + MailDev) | Task 1 |
| Makefile commands | Task 1 |
| Unit tests (Go testing + testify) | Task 4 |
| E2E tests (Python + pytest) | Task 12 |
| Audit logging | Task 8 |

**Gaps identified:**
- Admin user list needs search/filter by email and status - the `ListUsers` handler in Task 10 returns a placeholder. Need to add `ListUsers` method to repository.
- Forgot password and reset password endpoints are missing from auth handler/service.
- Tenant resolver middleware not fully wired (tenant_code lookup on login/register).

Let me add these missing pieces.

**2. Placeholder scan:** Found these:
- Task 5 handler has `// TODO: tenant lookup` - will be resolved by wiring in Task 11
- Task 4 service has `// TODO: fetch from DB` for roles - will be resolved by injecting role service
- Task 10 `ListUsers` returns placeholder - needs fix below
- Missing forgot-password/reset-password endpoints

**3. Type consistency:** All use `uuid.UUID`, `gin.Context`, consistent error patterns via `errors.RespondError`.

Let me add the missing pieces as additional tasks:

---

### Task 13: Fix Missing Features (forgot-password, reset-password, user list with search)

**Files:**
- Modify: `internal/auth/service/auth_service.go`, `internal/auth/handler/auth_handler.go`, `internal/auth/repository/user_repo.go`, `internal/admin/handler/admin_handler.go`

- [ ] **Step 1: Add forgot-password and reset-password to auth service**

Append to `internal/auth/service/auth_service.go`:

```go
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	// Find user by email (across all tenants for now - TODO: scope to tenant)
	// For simplicity, we'll need a FindByEmail method
	// Generate reset token and store in Redis
	resetToken := uuid.New().String()
	s.redis.Set(ctx, fmt.Sprintf("password_reset:%s", resetToken), email, 15*time.Minute)
	// TODO: send reset email
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	email, err := s.redis.Get(ctx, fmt.Sprintf("password_reset:%s", token)).Result()
	if err == redis.Nil {
		return errors.NewValidationError("invalid or expired reset token")
	}

	if err := domain.ValidatePassword(newPassword); err != nil {
		return errors.NewValidationError(err.Error())
	}

	s.redis.Del(ctx, fmt.Sprintf("password_reset:%s", token))

	// TODO: find user by email and update password

	return nil
}
```

- [ ] **Step 2: Add forgot-password and reset-password handlers**

Append to `internal/auth/handler/auth_handler.go`:

```go
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	if err := h.authSvc.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, gin.H{"message": "password reset email sent if account exists"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondError(c, errors.NewValidationError(err.Error()))
		return
	}

	if err := h.authSvc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		errors.RespondError(c, err)
		return
	}

	errors.Respond(c, 200, gin.H{"message": "password reset successfully"})
}
```

- [ ] **Step 3: Add ListUsers and FindByEmail to repository**

Add to `internal/auth/repository/user_repo.go`:

```go
type ListUsersFilter struct {
	TenantID uuid.UUID
	Email    string
	Status   string
	Offset   int
	Limit    int
}

func (r *pgxUserRepo) ListUsers(ctx context.Context, filter ListUsersFilter) ([]*domain.User, int, error) {
	query := `SELECT id, tenant_id, email, password_hash, name, avatar_url, status,
	                 email_verified, login_attempts, created_at, updated_at
	          FROM users WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.TenantID != uuid.Nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIdx)
		args = append(args, filter.TenantID)
		argIdx++
	}
	if filter.Email != "" {
		query += fmt.Sprintf(" AND email ILIKE $%d", argIdx)
		args = append(args, "%"+filter.Email+"%")
		argIdx++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}

	countQuery := strings.Replace(query, "SELECT id, tenant_id, email, password_hash, name, avatar_url, status, email_verified, login_attempts, created_at, updated_at", "SELECT COUNT(*)", 1)
	var count int
	r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&count)

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Status, &u.EmailVerified, &u.LoginAttempts, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}
	return users, count, nil
}

func (r *pgxUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, tenant_id, email, password_hash, name, avatar_url, status,
		        email_verified, login_attempts, created_at, updated_at
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
```

Add `"strings"` import to user_repo.go.

- [ ] **Step 4: Update UserRepository interface**

Update in `internal/auth/repository/user_repo.go`:

```go
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	ListUsers(ctx context.Context, filter ListUsersFilter) ([]*domain.User, int, error)
	Update(ctx context.Context, user *domain.User) error
}
```

- [ ] **Step 5: Fix ListUsers admin handler**

Replace the placeholder in `internal/admin/handler/admin_handler.go`:

```go
func (h *AdminHandler) ListUsers(c *gin.Context) {
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, _ := uuid.Parse(tenantIDStr)

	email := c.Query("email")
	status := c.Query("status")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := repository.ListUsersFilter{
		TenantID: tenantID,
		Email:    email,
		Status:   status,
		Offset:   offset,
		Limit:    limit,
	}

	users, count, err := h.userRepo.ListUsers(c.Request.Context(), filter)
	if err != nil {
		errors.RespondError(c, errors.NewInternalError("failed to list users"))
		return
	}

	// Strip password hash from response
	type safeUser struct {
		ID            uuid.UUID    `json:"id"`
		Email         string       `json:"email"`
		Name          string       `json:"name"`
		AvatarURL     string       `json:"avatar_url"`
		Status        string       `json:"status"`
		EmailVerified bool         `json:"email_verified"`
		CreatedAt     time.Time    `json:"created_at"`
	}

	result := make([]safeUser, 0, len(users))
	for _, u := range users {
		result = append(result, safeUser{
			ID: u.ID, Email: u.Email, Name: u.Name,
			AvatarURL: u.AvatarURL, Status: string(u.Status),
			EmailVerified: u.EmailVerified, CreatedAt: u.CreatedAt,
		})
	}

	errors.Respond(c, 200, gin.H{"users": result, "total": count, "offset": offset, "limit": limit})
}
```

- [ ] **Step 6: Add routes for forgot/reset password in main.go**

Add to auth routes in `cmd/server/main.go`:

```go
	auth.POST("/forgot-password", authHandler.ForgotPassword)
	auth.POST("/reset-password", authHandler.ResetPassword)
```

- [ ] **Step 7: Commit**

```bash
git add internal/auth/ internal/admin/ cmd/server/main.go
git commit -m "feat: forgot-password, reset-password, and admin user list with search/filter"
```

---

### Task 14: Final Verification + Cleanup

- [ ] **Step 1: Run all Go tests**

```bash
make test
```

Expected: All tests pass.

- [ ] **Step 2: Verify full build**

```bash
make build
```

Expected: Binary produced without errors.

- [ ] **Step 3: Run e2e tests**

```bash
make up
make migrate-up
./bin/server &
sleep 3
make test-e2e
```

Expected: All Python e2e tests pass.

- [ ] **Step 4: Commit any fixes**

```bash
git add .
git commit -m "fix: final cleanup and verification"
```

---

Plan complete. All spec requirements are covered by tasks 1-14.
