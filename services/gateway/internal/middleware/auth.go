package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kexi/telegram-bot-gateway/internal/pkg/apikey"
	"github.com/kexi/telegram-bot-gateway/internal/pkg/jwt"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
)

// AuthContext holds authentication information
type AuthContext struct {
	UserID    uint
	Username  string
	Roles     []string
	APIKeyID  *uint
	IsAPIKey  bool
}

// AuthMiddleware creates a middleware for JWT and API key authentication
func AuthMiddleware(jwtService *jwt.Service, apiKeyService *apikey.Service, apiKeyRepo repository.APIKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT authentication first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtService.ValidateToken(token)
			if err == nil {
				// Valid JWT token
				authCtx := &AuthContext{
					UserID:   claims.UserID,
					Username: claims.Username,
					Roles:    claims.Roles,
					IsAPIKey: false,
				}
				c.Set("auth", authCtx)
				c.Next()
				return
			}
		}

		// Try API key authentication
		// Support three methods (similar to Telegram Bot API):
		// 1. Header: X-API-Key
		// 2. Query parameter: ?api_key=xxx or ?token=xxx
		// 3. POST body: api_key or token field
		apiKeyValue := c.GetHeader("X-API-Key")
		if apiKeyValue == "" {
			// Try query parameter
			apiKeyValue = c.Query("api_key")
			if apiKeyValue == "" {
				apiKeyValue = c.Query("token")
			}
		}
		if apiKeyValue == "" && c.Request.Method == "POST" {
			// Try POST body (form-data or x-www-form-urlencoded)
			apiKeyValue = c.PostForm("api_key")
			if apiKeyValue == "" {
				apiKeyValue = c.PostForm("token")
			}
		}

		if apiKeyValue != "" {
			if !apiKeyService.IsValid(apiKeyValue) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key format"})
				c.Abort()
				return
			}

			// Look up API key in database
			ctx := context.Background()
			apiKey, err := apiKeyRepo.GetByKey(ctx, apiKeyValue)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
				c.Abort()
				return
			}

			// Verify key hash
			if !apiKeyService.Verify(apiKeyValue, apiKey.HashedKey) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
				c.Abort()
				return
			}

			// Check if active
			if !apiKey.IsActive {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is inactive"})
				c.Abort()
				return
			}

			// Check expiration
			if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "API key has expired"})
				c.Abort()
				return
			}

			// Update last used timestamp (async, don't block request)
			go apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID)

			// Set auth context
			authCtx := &AuthContext{
				APIKeyID: &apiKey.ID,
				IsAPIKey: true,
			}
			c.Set("auth", authCtx)
			c.Next()
			return
		}

		// No valid authentication found
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		c.Abort()
	}
}

// OptionalAuthMiddleware allows unauthenticated requests but sets auth context if credentials are provided
func OptionalAuthMiddleware(jwtService *jwt.Service, apiKeyService *apikey.Service, apiKeyRepo repository.APIKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT authentication
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtService.ValidateToken(token)
			if err == nil {
				authCtx := &AuthContext{
					UserID:   claims.UserID,
					Username: claims.Username,
					Roles:    claims.Roles,
					IsAPIKey: false,
				}
				c.Set("auth", authCtx)
			}
		}

		// Try API key authentication (header, query param, or POST body)
		apiKeyValue := c.GetHeader("X-API-Key")
		if apiKeyValue == "" {
			apiKeyValue = c.Query("api_key")
			if apiKeyValue == "" {
				apiKeyValue = c.Query("token")
			}
		}
		if apiKeyValue == "" && c.Request.Method == "POST" {
			apiKeyValue = c.PostForm("api_key")
			if apiKeyValue == "" {
				apiKeyValue = c.PostForm("token")
			}
		}

		if apiKeyValue != "" && apiKeyService.IsValid(apiKeyValue) {
			ctx := context.Background()
			apiKey, err := apiKeyRepo.GetByKey(ctx, apiKeyValue)
			if err == nil && apiKey.IsActive {
				if apiKeyService.Verify(apiKeyValue, apiKey.HashedKey) {
					authCtx := &AuthContext{
						APIKeyID: &apiKey.ID,
						IsAPIKey: true,
					}
					c.Set("auth", authCtx)
					go apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID)
				}
			}
		}

		c.Next()
	}
}

// GetAuthContext retrieves the authentication context from the request
func GetAuthContext(c *gin.Context) (*AuthContext, bool) {
	val, exists := c.Get("auth")
	if !exists {
		return nil, false
	}
	authCtx, ok := val.(*AuthContext)
	return authCtx, ok
}

// RequireAuth ensures an auth context exists
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := GetAuthContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
