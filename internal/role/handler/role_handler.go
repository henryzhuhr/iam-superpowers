package handler

import (
	stderrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apierrors "github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/role/service"
)

type RoleHandler struct {
	svc *service.RoleService
}

func NewRoleHandler(svc *service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

func (h *RoleHandler) ListRoles(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant_id"))
		return
	}

	roles, err := h.svc.ListRoles(c.Request.Context(), tenantID)
	if err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to list roles"))
		return
	}

	apierrors.Respond(c, http.StatusOK, roles)
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant_id"))
		return
	}

	role, err := h.svc.CreateRole(c.Request.Context(), tenantID, req.Name, req.Description)
	if err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to create role"))
		return
	}

	apierrors.Respond(c, http.StatusCreated, role)
}

func (h *RoleHandler) AssignRole(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user_id"))
		return
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid role ID"))
		return
	}

	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant_id"))
		return
	}

	if err := h.svc.AssignRoleToUser(c.Request.Context(), userID, roleID, tenantID); err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to assign role"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "role assigned successfully"})
}

func (h *RoleHandler) RemoveRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid role ID"))
		return
	}

	userID, err := uuid.Parse(c.Query("user_id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user_id"))
		return
	}

	if err := h.svc.RemoveRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to remove role"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "role removed successfully"})
}

func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user ID"))
		return
	}

	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant_id"))
		return
	}

	roles, err := h.svc.GetUserRoles(c.Request.Context(), userID, tenantID)
	if err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to get user roles"))
		return
	}

	apierrors.Respond(c, http.StatusOK, roles)
}
