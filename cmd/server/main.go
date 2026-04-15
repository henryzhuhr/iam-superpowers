package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/henryzhuhr/iam-superpowers/internal/admin/handler"
	authhandler "github.com/henryzhuhr/iam-superpowers/internal/auth/handler"
	authrepo "github.com/henryzhuhr/iam-superpowers/internal/auth/repository"
	authservice "github.com/henryzhuhr/iam-superpowers/internal/auth/service"
	audithandler "github.com/henryzhuhr/iam-superpowers/internal/audit/handler"
	auditrepo "github.com/henryzhuhr/iam-superpowers/internal/audit/repository"
	auditservice "github.com/henryzhuhr/iam-superpowers/internal/audit/service"
	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
	"github.com/henryzhuhr/iam-superpowers/internal/common/database"
	"github.com/henryzhuhr/iam-superpowers/internal/common/email"
	"github.com/henryzhuhr/iam-superpowers/internal/common/jwt"
	"github.com/henryzhuhr/iam-superpowers/internal/common/middleware"
	redisclient "github.com/henryzhuhr/iam-superpowers/internal/common/redis"
	rolerepo "github.com/henryzhuhr/iam-superpowers/internal/role/repository"
	roleservice "github.com/henryzhuhr/iam-superpowers/internal/role/service"
	tenanthandler "github.com/henryzhuhr/iam-superpowers/internal/tenant/handler"
	tenantrepo "github.com/henryzhuhr/iam-superpowers/internal/tenant/repository"
	tenantservice "github.com/henryzhuhr/iam-superpowers/internal/tenant/service"
	userservice "github.com/henryzhuhr/iam-superpowers/internal/user/service"
	userhandler "github.com/henryzhuhr/iam-superpowers/internal/user/handler"
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
	redis, err := redisclient.New(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redis.Close()

	// Initialize shared services
	jwtSvc := jwt.New(cfg.JWT)
	emailSvc := email.New(cfg.SMTP)

	// Auth module
	userRepo := authrepo.NewUserRepository(db)
	authSvc := authservice.NewAuthService(userRepo, jwtSvc, emailSvc, redis.RDB())
	authH := authhandler.NewAuthHandler(authSvc)

	// User module
	userSvc := userservice.NewUserService(userRepo)
	userH := userhandler.NewUserHandler(userSvc)

	// Tenant module
	tenantRepo := tenantrepo.NewTenantRepository(db)
	tenantSvc := tenantservice.NewTenantService(tenantRepo)
	tenantH := tenanthandler.NewTenantHandler(tenantSvc)

	// Role module
	roleRepo := rolerepo.NewRoleRepository(db)
	roleSvc := roleservice.NewRoleService(roleRepo)

	// Audit module
	auditRepo := auditrepo.NewAuditRepository(db)
	auditSvc := auditservice.NewAuditService(auditRepo)
	auditH := audithandler.NewAuditHandler(auditSvc)

	// Admin module
	adminH := handler.NewAdminHandler(userRepo, roleSvc, tenantSvc)

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
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.POST("/refresh", authH.Refresh)
		auth.POST("/verify-email", authH.VerifyEmail)
	}

	// Authenticated routes
	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuth(jwtSvc))
	{
		protected.POST("/logout", authH.Logout)
		protected.GET("/users/me", userH.GetProfile)
		protected.PUT("/users/me", userH.UpdateProfile)
		protected.PUT("/users/me/password", userH.ChangePassword)

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", adminH.ListUsers)
			admin.GET("/users/:id", adminH.GetUser)
			admin.PUT("/users/:id", adminH.UpdateUser)
			admin.PUT("/users/:id/status", adminH.UpdateUserStatus)
			admin.POST("/users/:id/reset-password", adminH.ResetUserPassword)
			admin.GET("/users/:id/roles", adminH.GetUserRoles)
			admin.PUT("/users/:id/roles", adminH.AssignUserRole)

			admin.GET("/tenants", adminH.ListTenants)
			admin.POST("/tenants", adminH.CreateTenant)
			admin.GET("/tenants/:id", tenantH.GetTenant)

			admin.GET("/roles", adminH.ListRoles)
			admin.POST("/roles", adminH.CreateRole)

			admin.GET("/audit-logs", auditH.ListAuditLogs)
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
