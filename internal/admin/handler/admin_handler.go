package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	auditservice "github.com/henryzhuhr/iam-superpowers/internal/audit/service"
	authdomain "github.com/henryzhuhr/iam-superpowers/internal/auth/domain"
	authrepo "github.com/henryzhuhr/iam-superpowers/internal/auth/repository"
	apierrors "github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/common/middleware"
	roleservice "github.com/henryzhuhr/iam-superpowers/internal/role/service"
	tenantservice "github.com/henryzhuhr/iam-superpowers/internal/tenant/service"
)

type AdminHandler struct {
	userRepo   authrepo.UserRepository
	roleSvc    *roleservice.RoleService
	tenantSvc  *tenantservice.TenantService
	auditSvc   *auditservice.AuditService
}

func NewAdminHandler(userRepo authrepo.UserRepository, roleSvc *roleservice.RoleService, tenantSvc *tenantservice.TenantService, auditSvc *auditservice.AuditService) *AdminHandler {
	return &AdminHandler{
		userRepo:  userRepo,
		roleSvc:   roleSvc,
		tenantSvc: tenantSvc,
		auditSvc:  auditSvc,
	}
}

// Admin user list
func (h *AdminHandler) ListUsers(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	search := c.Query("search")
	status := c.Query("status")
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant ID"))
		return
	}

	filter := authrepo.ListUsersFilter{
		TenantID: tenantID,
		Email:    search,
		Status:   status,
		Offset:   offset,
		Limit:    limit,
	}

	users, count, err := h.userRepo.ListUsers(c.Request.Context(), filter)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to list users"))
		return
	}

	// Strip password hash from response
	type safeUser struct {
		ID            uuid.UUID `json:"id"`
		TenantID      uuid.UUID `json:"tenant_id"`
		Email         string    `json:"email"`
		Name          string    `json:"name"`
		AvatarURL     string    `json:"avatar_url"`
		Status        string    `json:"status"`
		EmailVerified bool      `json:"email_verified"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
	}

	result := make([]safeUser, 0, len(users))
	for _, u := range users {
		result = append(result, safeUser{
			ID:            u.ID,
			TenantID:      u.TenantID,
			Email:         u.Email,
			Name:          u.Name,
			AvatarURL:     u.AvatarURL,
			Status:        string(u.Status),
			EmailVerified: u.EmailVerified,
			CreatedAt:     u.CreatedAt,
			UpdatedAt:     u.UpdatedAt,
		})
	}

	apierrors.Respond(c, http.StatusOK, gin.H{
		"users":  result,
		"total":  count,
		"offset": offset,
		"limit":  limit,
	})
}

// Admin user detail
func (h *AdminHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user ID"))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to find user"))
		return
	}
	if user == nil {
		apierrors.RespondError(c, apierrors.NewNotFoundError("user not found"))
		return
	}

	apierrors.Respond(c, http.StatusOK, user)
}

// Admin update user
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user ID"))
		return
	}

	var req struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to find user"))
		return
	}
	if user == nil {
		apierrors.RespondError(c, apierrors.NewNotFoundError("user not found"))
		return
	}

	user.Name = req.Name
	user.AvatarURL = req.AvatarURL
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to update user"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "user updated"})
}

// Admin disable/enable user
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user ID"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive locked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to find user"))
		return
	}
	if user == nil {
		apierrors.RespondError(c, apierrors.NewNotFoundError("user not found"))
		return
	}

	user.Status = authdomain.UserStatus(req.Status)
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to update user status"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "user status updated"})
}

// Admin reset user password
func (h *AdminHandler) ResetUserPassword(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user ID"))
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to find user"))
		return
	}
	if user == nil {
		apierrors.RespondError(c, apierrors.NewNotFoundError("user not found"))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to hash password"))
		return
	}
	user.PasswordHash = string(hash)
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to reset password"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "password reset successfully"})
}

// Admin assign roles to user
func (h *AdminHandler) AssignUserRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user ID"))
		return
	}

	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant ID"))
		return
	}

	var req struct {
		RoleID string `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid role_id"))
		return
	}

	if err := h.roleSvc.AssignRoleToUser(c.Request.Context(), userID, roleID, tenantID); err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to assign role"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "role assigned"})
}

// Admin get user roles
func (h *AdminHandler) GetUserRoles(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user ID"))
		return
	}

	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant ID"))
		return
	}

	roles, err := h.roleSvc.GetUserRoles(c.Request.Context(), userID, tenantID)
	if err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
		} else {
			apierrors.RespondError(c, apierrors.NewInternalError("failed to get user roles"))
		}
		return
	}

	apierrors.Respond(c, http.StatusOK, roles)
}

// Admin tenant list
func (h *AdminHandler) ListTenants(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	tenants, count, err := h.tenantSvc.ListTenants(c.Request.Context(), offset, limit)
	if err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
		} else {
			apierrors.RespondError(c, apierrors.NewInternalError("failed to list tenants"))
		}
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{
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
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	tenant, err := h.tenantSvc.CreateTenant(c.Request.Context(), req.Name, req.UniqueCode, req.CustomDomain)
	if err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to create tenant"))
		return
	}

	apierrors.Respond(c, http.StatusCreated, tenant)
}

// Admin get tenant
func (h *AdminHandler) GetTenant(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant ID"))
		return
	}

	tenant, err := h.tenantSvc.GetTenant(c.Request.Context(), id)
	if err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to find tenant"))
		return
	}

	apierrors.Respond(c, http.StatusOK, tenant)
}

// Admin role list
func (h *AdminHandler) ListRoles(c *gin.Context) {
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant ID"))
		return
	}

	roles, err := h.roleSvc.ListRoles(c.Request.Context(), tenantID)
	if err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
		} else {
			apierrors.RespondError(c, apierrors.NewInternalError("failed to list roles"))
		}
		return
	}

	apierrors.Respond(c, http.StatusOK, roles)
}

// Admin create role
func (h *AdminHandler) CreateRole(c *gin.Context) {
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant ID"))
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	role, err := h.roleSvc.CreateRole(c.Request.Context(), tenantID, req.Name, req.Description)
	if err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
		} else {
			apierrors.RespondError(c, apierrors.NewInternalError("failed to create role"))
		}
		return
	}

	apierrors.Respond(c, http.StatusCreated, role)
}

// Admin audit logs
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	tenantIDStr := c.GetString(middleware.ContextKeyTenantID)
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant ID"))
		return
	}

	startTime := time.Now().Add(-24 * time.Hour)
	if s := c.Query("start_time"); s != "" {
		startTime, err = time.Parse(time.RFC3339, s)
		if err != nil {
			apierrors.RespondError(c, apierrors.NewValidationError("invalid start_time, use RFC3339 format"))
			return
		}
	}

	endTime := time.Now()
	if e := c.Query("end_time"); e != "" {
		endTime, err = time.Parse(time.RFC3339, e)
		if err != nil {
			apierrors.RespondError(c, apierrors.NewValidationError("invalid end_time, use RFC3339 format"))
			return
		}
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	logs, count, err := h.auditSvc.ListLogs(c.Request.Context(), tenantID, startTime, endTime, offset, limit)
	if err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to list audit logs"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{
		"logs":  logs,
		"total": count,
	})
}
