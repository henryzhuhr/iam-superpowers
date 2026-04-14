package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/henryzhuhr/iam-superpowers/internal/common/config"
)

type Claims struct {
	UserID   string   `json:"sub"`
	TenantID string   `json:"tid"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type Service struct {
	secret      []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func New(cfg config.JWTConfig) *Service {
	return &Service{
		secret:     []byte(cfg.Secret),
		accessTTL:  time.Duration(cfg.AccessTokenTTL) * time.Second,
		refreshTTL: time.Duration(cfg.RefreshTokenTTL) * time.Second,
	}
}

func (s *Service) GenerateTokenPair(userID, tenantID string, roles []string) (*TokenPair, error) {
	jti := uuid.New().String()

	claims := &Claims{
		UserID:   userID,
		TenantID: tenantID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessStr, err := accessToken.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken := uuid.New().String()

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessTTL.Seconds()),
	}, nil
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (s *Service) AccessTTL() int {
	return int(s.accessTTL.Seconds())
}
