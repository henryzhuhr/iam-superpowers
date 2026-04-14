package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	er "github.com/henryzhuhr/iam-superpowers/internal/common/errors"
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
			er.RespondError(c, er.NewUnauthorizedError("missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			er.RespondError(c, er.NewUnauthorizedError("invalid authorization header format"))
			c.Abort()
			return
		}

		claims, err := jwtSvc.ValidateToken(parts[1])
		if err != nil {
			er.RespondError(c, er.NewUnauthorizedError("invalid or expired token"))
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
			er.RespondError(c, er.NewForbiddenError("authentication required"))
			c.Abort()
			return
		}

		claims, ok := claimsVal.(*jwt.Claims)
		if !ok {
			er.RespondError(c, er.NewInternalError("invalid claims type"))
			c.Abort()
			return
		}

		for _, r := range claims.Roles {
			if r == role {
				c.Next()
				return
			}
		}

		er.RespondError(c, er.NewForbiddenError("insufficient permissions"))
		c.Abort()
	}
}
