package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims with custom fields
type Claims struct {
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// Service handles JWT token operations
type Service struct {
	secret            []byte
	issuer            string
	accessTokenTTL    time.Duration
	refreshTokenTTL   time.Duration
	refreshThreshold  time.Duration
}

// NewService creates a new JWT service
func NewService(secret string, issuer string, accessTokenTTL, refreshTokenTTL, refreshThreshold time.Duration) *Service {
	return &Service{
		secret:           []byte(secret),
		issuer:           issuer,
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		refreshThreshold: refreshThreshold,
	}
}

// GenerateAccessToken creates a new access token
func (s *Service) GenerateAccessToken(userID uint, username string, roles []string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken creates a new refresh token (random string)
func (s *Service) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateTokenPair creates both access and refresh tokens
func (s *Service) GenerateTokenPair(userID uint, username string, roles []string) (*TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(userID, username, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
	}, nil
}

// ValidateToken validates and parses a JWT token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ShouldRefresh checks if the token should be auto-refreshed
func (s *Service) ShouldRefresh(claims *Claims) bool {
	if claims.ExpiresAt == nil {
		return false
	}
	timeUntilExpiry := time.Until(claims.ExpiresAt.Time)
	return timeUntilExpiry < s.refreshThreshold
}

// GetRefreshTokenTTL returns the refresh token TTL
func (s *Service) GetRefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}

// GetAccessTokenTTL returns the access token TTL
func (s *Service) GetAccessTokenTTL() time.Duration {
	return s.accessTokenTTL
}
