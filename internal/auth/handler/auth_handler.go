package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/service"
	stderrors "github.com/henryzhuhr/iam-superpowers/internal/common/errors"
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
		stderrors.RespondError(c, stderrors.NewValidationError(err.Error()))
		return
	}

	// TODO: tenant lookup by tenant_code - need tenant service
	stderrors.RespondError(c, stderrors.NewInternalError("tenant service not yet implemented"))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		stderrors.RespondError(c, stderrors.NewValidationError(err.Error()))
		return
	}

	// TODO: tenant lookup by tenant_code - need tenant service
	tokens, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password, uuid.Nil)
	if err != nil {
		var apiErr *stderrors.APIError
		if errors.As(err, &apiErr) {
			stderrors.RespondError(c, apiErr)
			return
		}
		stderrors.RespondError(c, stderrors.NewInternalError("login failed"))
		return
	}

	stderrors.Respond(c, 200, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		stderrors.RespondError(c, stderrors.NewValidationError(err.Error()))
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)
	tokens, err := h.authSvc.Refresh(c.Request.Context(), userID, req.RefreshToken)
	if err != nil {
		var apiErr *stderrors.APIError
		if errors.As(err, &apiErr) {
			stderrors.RespondError(c, apiErr)
			return
		}
		stderrors.RespondError(c, stderrors.NewInternalError("refresh failed"))
		return
	}

	stderrors.Respond(c, 200, gin.H{
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
	stderrors.Respond(c, 200, gin.H{"message": "logged out successfully"})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		stderrors.RespondError(c, stderrors.NewValidationError(err.Error()))
		return
	}

	err := h.authSvc.VerifyEmail(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		var apiErr *stderrors.APIError
		if errors.As(err, &apiErr) {
			stderrors.RespondError(c, apiErr)
			return
		}
		stderrors.RespondError(c, stderrors.NewInternalError("verification failed"))
		return
	}

	stderrors.Respond(c, 200, gin.H{"message": "email verified successfully"})
}
