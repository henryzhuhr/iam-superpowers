package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
	"github.com/henryzhuhr/iam-superpowers/internal/common/jwt"
	"github.com/stretchr/testify/assert"
)

func setupTestContext(token string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	if token != "" {
		c.Request.Header.Set("Authorization", "Bearer "+token)
	}
	return c, w
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	jwtSvc := jwt.New(config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 900, RefreshTokenTTL: 604800})
	c, w := setupTestContext("")
	JWTAuth(jwtSvc)(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	jwtSvc := jwt.New(config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 900, RefreshTokenTTL: 604800})
	c, w := setupTestContext("invalid-token")
	JWTAuth(jwtSvc)(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_ValidToken(t *testing.T) {
	jwtSvc := jwt.New(config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 900, RefreshTokenTTL: 604800})
	tokens, _ := jwtSvc.GenerateTokenPair("user-1", "tenant-1", []string{"admin"})
	c, w := setupTestContext(tokens.AccessToken)
	JWTAuth(jwtSvc)(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user-1", c.GetString(ContextKeyUserID))
	assert.Equal(t, "tenant-1", c.GetString(ContextKeyTenantID))
}

func TestRequireRole_MatchingRole(t *testing.T) {
	jwtSvc := jwt.New(config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 900, RefreshTokenTTL: 604800})
	tokens, _ := jwtSvc.GenerateTokenPair("user-1", "tenant-1", []string{"admin"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(JWTAuth(jwtSvc))
	router.GET("/test", func(c *gin.Context) {
		RequireRole("admin")(c)
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_NonMatchingRole(t *testing.T) {
	jwtSvc := jwt.New(config.JWTConfig{Secret: "test-secret", AccessTokenTTL: 900, RefreshTokenTTL: 604800})
	tokens, _ := jwtSvc.GenerateTokenPair("user-1", "tenant-1", []string{"user"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(JWTAuth(jwtSvc))
	router.GET("/test", func(c *gin.Context) {
		RequireRole("admin")(c)
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_NoClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		RequireRole("admin")(c)
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
