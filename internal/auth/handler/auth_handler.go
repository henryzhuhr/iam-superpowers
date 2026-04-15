package handler

import (
	stderrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/henryzhuhr/iam-superpowers/internal/auth/service"
	apierrors "github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/henryzhuhr/iam-superpowers/internal/common/middleware"
	tenantrepo "github.com/henryzhuhr/iam-superpowers/internal/tenant/repository"
)

type AuthHandler struct {
	authSvc    *service.AuthService
	tenantRepo tenantrepo.TenantRepository
}

func NewAuthHandler(authSvc *service.AuthService, tenantRepo tenantrepo.TenantRepository) *AuthHandler {
	return &AuthHandler{
		authSvc:    authSvc,
		tenantRepo: tenantRepo,
	}
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

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	tenant, err := h.tenantRepo.FindByCode(c.Request.Context(), req.TenantCode)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to find tenant"))
		return
	}
	if tenant == nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant code"))
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), tenant.ID, req.Email, req.Password)
	if err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("registration failed"))
		return
	}

	apierrors.Respond(c, http.StatusCreated, gin.H{
		"user_id": user.ID,
		"email":   user.Email,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	tenant, err := h.tenantRepo.FindByCode(c.Request.Context(), req.TenantCode)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewInternalError("failed to find tenant"))
		return
	}
	if tenant == nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid tenant code"))
		return
	}

	tokens, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password, tenant.ID)
	if err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("login failed"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)
	tokens, err := h.authSvc.Refresh(c.Request.Context(), userID, req.RefreshToken)
	if err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("refresh failed"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.RefreshToken = ""
	}

	_ = h.authSvc.Logout(c.Request.Context(), userID, req.RefreshToken, "")
	apierrors.Respond(c, http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	err := h.authSvc.VerifyEmail(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("verification failed"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "email verified successfully"})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	if err := h.authSvc.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("forgot password failed"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "password reset email sent if account exists"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	if err := h.authSvc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		var apiErr *apierrors.APIError
		if stderrors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("reset password failed"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "password reset successfully"})
}
