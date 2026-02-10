package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
	"github.com/kexi/telegram-bot-gateway/internal/pkg/jwt"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtService       *jwt.Service
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtService *jwt.Service,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         *UserDTO `json:"user"`
}

// UserDTO represents user data transfer object
type UserDTO struct {
	ID       uint     `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email,omitempty"`
	IsActive bool     `json:"is_active"`
	Roles    []string `json:"roles"`
}

// Login authenticates a user and returns JWT tokens
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Get user roles
	userWithRoles, err := s.userRepo.WithRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user roles: %w", err)
	}

	// Extract role names
	roleNames := make([]string, len(userWithRoles.Roles))
	for i, role := range userWithRoles.Roles {
		roleNames[i] = role.Name
	}

	// Generate tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Username, roleNames)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token
	hashedRefreshToken := hashToken(tokenPair.RefreshToken)
	refreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     hashedRefreshToken,
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenTTL()),
	}
	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User: &UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsActive: user.IsActive,
			Roles:    roleNames,
		},
	}, nil
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh generates a new access token using a refresh token
func (s *AuthService) Refresh(ctx context.Context, req *RefreshRequest) (*LoginResponse, error) {
	// Hash the refresh token to look it up
	hashedToken := hashToken(req.RefreshToken)

	// Get refresh token from database
	refreshToken, err := s.refreshTokenRepo.GetByToken(ctx, hashedToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user with roles
	user, err := s.userRepo.WithRoles(ctx, refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("account is inactive")
	}

	// Extract role names
	roleNames := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleNames[i] = role.Name
	}

	// Generate new tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Username, roleNames)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Revoke old refresh token
	if err := s.refreshTokenRepo.Revoke(ctx, hashedToken); err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Store new refresh token
	hashedNewToken := hashToken(tokenPair.RefreshToken)
	newRefreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     hashedNewToken,
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenTTL()),
	}
	if err := s.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User: &UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsActive: user.IsActive,
			Roles:    roleNames,
		},
	}, nil
}

// Logout revokes a user's refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	hashedToken := hashToken(refreshToken)
	return s.refreshTokenRepo.Revoke(ctx, hashedToken)
}

// LogoutAll revokes all of a user's refresh tokens
func (s *AuthService) LogoutAll(ctx context.Context, userID uint) error {
	return s.refreshTokenRepo.RevokeAllByUser(ctx, userID)
}

// CreateUser creates a new user account
func (s *AuthService) CreateUser(ctx context.Context, username, email, password string) (*domain.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// hashToken creates a SHA-256 hash of a token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
