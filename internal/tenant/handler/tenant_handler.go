package handler

import (
	stderrors "errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apierrors "github.com/henryzhuhr/iam-superpowers/internal/common/errors"
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
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to list tenants"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{
		"tenants": tenants,
		"total":   count,
		"offset":  offset,
		"limit":   limit,
	})
}

func (h *TenantHandler) GetTenant(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant ID"))
		return
	}

	tenant, err := h.svc.GetTenant(c.Request.Context(), id)
	if err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to get tenant"))
		return
	}

	apierrors.Respond(c, http.StatusOK, tenant)
}

func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	tenant, err := h.svc.CreateTenant(c.Request.Context(), req.Name, req.UniqueCode)
	if err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to create tenant"))
		return
	}

	apierrors.Respond(c, http.StatusCreated, tenant)
}
