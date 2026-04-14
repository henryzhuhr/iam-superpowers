package jwt_test

import (
	"strings"
	"testing"

	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
	jwtlib "github.com/henryzhuhr/iam-superpowers/internal/common/jwt"
)

func testConfig() config.JWTConfig {
	return config.JWTConfig{
		Secret:          "test-secret-key-for-testing",
		AccessTokenTTL:  3600,
		RefreshTokenTTL: 86400,
	}
}

func TestGenerateTokenPair(t *testing.T) {
	svc := jwtlib.New(testConfig())

	t.Run("generates valid token pair", func(t *testing.T) {
		pair, err := svc.GenerateTokenPair("user-1", "tenant-1", []string{"admin", "user"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pair.AccessToken == "" {
			t.Error("access token should not be empty")
		}
		if pair.RefreshToken == "" {
			t.Error("refresh token should not be empty")
		}
		if pair.ExpiresIn != 3600 {
			t.Errorf("expected ExpiresIn 3600, got %d", pair.ExpiresIn)
		}

		// Access token should be a valid JWT with 3 parts
		parts := strings.Split(pair.AccessToken, ".")
		if len(parts) != 3 {
			t.Errorf("expected JWT with 3 parts, got %d", len(parts))
		}
	})

	t.Run("tokens are unique per call", func(t *testing.T) {
		pair1, _ := svc.GenerateTokenPair("user-1", "tenant-1", nil)
		pair2, _ := svc.GenerateTokenPair("user-1", "tenant-1", nil)

		if pair1.AccessToken == pair2.AccessToken {
			t.Error("access tokens should be unique")
		}
		if pair1.RefreshToken == pair2.RefreshToken {
			t.Error("refresh tokens should be unique")
		}
	})
}

func TestValidateToken(t *testing.T) {
	svc := jwtlib.New(testConfig())

	t.Run("validates a valid token", func(t *testing.T) {
		pair, _ := svc.GenerateTokenPair("user-123", "tenant-456", []string{"admin"})

		claims, err := svc.ValidateToken(pair.AccessToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if claims.UserID != "user-123" {
			t.Errorf("expected UserID 'user-123', got '%s'", claims.UserID)
		}
		if claims.TenantID != "tenant-456" {
			t.Errorf("expected TenantID 'tenant-456', got '%s'", claims.TenantID)
		}
		if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
			t.Errorf("expected roles ['admin'], got %v", claims.Roles)
		}
	})

	t.Run("rejects malformed token", func(t *testing.T) {
		_, err := svc.ValidateToken("not-a-valid-token")
		if err == nil {
			t.Fatal("expected error for malformed token")
		}
		if !strings.Contains(err.Error(), "invalid token") {
			t.Errorf("expected 'invalid token' error, got: %v", err)
		}
	})

	t.Run("rejects token signed with different secret", func(t *testing.T) {
		// Generate with one service
		svc1 := jwtlib.New(config.JWTConfig{
			Secret:         "secret-one",
			AccessTokenTTL: 3600,
		})
		pair, _ := svc1.GenerateTokenPair("user-1", "tenant-1", nil)

		// Validate with a different secret
		svc2 := jwtlib.New(config.JWTConfig{
			Secret:         "secret-two",
			AccessTokenTTL: 3600,
		})
		_, err := svc2.ValidateToken(pair.AccessToken)
		if err == nil {
			t.Fatal("expected error for wrong secret")
		}
	})

	t.Run("rejects empty token", func(t *testing.T) {
		_, err := svc.ValidateToken("")
		if err == nil {
			t.Fatal("expected error for empty token")
		}
	})
}

func TestAccessTTL(t *testing.T) {
	svc := jwtlib.New(testConfig())

	ttl := svc.AccessTTL()
	if ttl != 3600 {
		t.Errorf("expected AccessTTL 3600, got %d", ttl)
	}
}

func TestAccessTTL_CustomValue(t *testing.T) {
	svc := jwtlib.New(config.JWTConfig{
		Secret:         "test",
		AccessTokenTTL: 1800,
	})

	ttl := svc.AccessTTL()
	if ttl != 1800 {
		t.Errorf("expected AccessTTL 1800, got %d", ttl)
	}
}
