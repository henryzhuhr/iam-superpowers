package handler

import (
	stderrors "errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apierrors "github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/audit/service"
)

type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant_id"))
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

	logs, count, err := h.svc.ListLogs(c.Request.Context(), tenantID, startTime, endTime, offset, limit)
	if err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
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
