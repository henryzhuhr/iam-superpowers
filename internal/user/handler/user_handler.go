package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apierrors "github.com/henryzhuhr/iam-superpowers/internal/common/errors"
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
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user ID"))
		return
	}

	user, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to get profile"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{
		"id":             user.ID,
		"email":          user.Email,
		"name":           user.Name,
		"avatar_url":     user.AvatarURL,
		"email_verified": user.EmailVerified,
		"status":         user.Status,
		"created_at":     user.CreatedAt,
	})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDStr := c.GetString(middleware.ContextKeyUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user ID"))
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	if err := h.svc.UpdateProfile(c.Request.Context(), userID, req.Name, req.AvatarURL); err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to update profile"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "profile updated"})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userIDStr := c.GetString(middleware.ContextKeyUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError("invalid user ID"))
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.RespondError(c, apierrors.NewValidationError(err.Error()))
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			apierrors.RespondError(c, apiErr)
			return
		}
		apierrors.RespondError(c, apierrors.NewInternalError("failed to change password"))
		return
	}

	apierrors.Respond(c, http.StatusOK, gin.H{"message": "password changed successfully"})
}
