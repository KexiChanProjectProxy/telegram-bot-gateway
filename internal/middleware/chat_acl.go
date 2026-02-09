package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
)

// Permission types
const (
	PermissionRead   = "read"
	PermissionSend   = "send"
	PermissionManage = "manage"
)

// ChatACLMiddleware checks granular chat-level permissions
func ChatACLMiddleware(permission string, chatPermRepo repository.ChatPermissionRepository, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get auth context
		authCtx, exists := GetAuthContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Get chat ID from URL parameter
		chatIDStr := c.Param("id")
		if chatIDStr == "" {
			chatIDStr = c.Param("chat_id")
		}
		if chatIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Chat ID required"})
			c.Abort()
			return
		}

		chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
			c.Abort()
			return
		}

		// Check permission (with Redis caching)
		allowed, err := checkChatPermission(c.Request.Context(), authCtx, uint(chatID), permission, chatPermRepo, redisClient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions for this chat"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkChatPermission checks if the authenticated user/API key has the required permission
// Uses Redis cache with 5-minute TTL
func checkChatPermission(ctx context.Context, authCtx *AuthContext, chatID uint, permission string, chatPermRepo repository.ChatPermissionRepository, redisClient *redis.Client) (bool, error) {
	// Build cache key
	var cacheKey string
	if authCtx.IsAPIKey {
		cacheKey = fmt.Sprintf("chat_perm:apikey:%d:chat:%d:%s", *authCtx.APIKeyID, chatID, permission)
	} else {
		cacheKey = fmt.Sprintf("chat_perm:user:%d:chat:%d:%s", authCtx.UserID, chatID, permission)
	}

	// Try cache first
	if redisClient != nil {
		cached, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			return cached == "1", nil
		}
	}

	// Not in cache, query database
	var chatPerm interface{}
	var err error

	if authCtx.IsAPIKey {
		chatPerm, err = chatPermRepo.GetByAPIKeyAndChat(ctx, *authCtx.APIKeyID, chatID)
	} else {
		chatPerm, err = chatPermRepo.GetByUserAndChat(ctx, authCtx.UserID, chatID)
	}

	// Permission not found = not allowed
	if err != nil {
		// Cache negative result for 1 minute to prevent repeated DB queries
		if redisClient != nil {
			redisClient.Set(ctx, cacheKey, "0", 1*time.Minute)
		}
		return false, nil
	}

	// Extract permission based on type
	var allowed bool
	if perm, ok := chatPerm.(*domain.ChatPermission); ok {
		switch permission {
		case PermissionRead:
			allowed = perm.CanRead
		case PermissionSend:
			allowed = perm.CanSend
		case PermissionManage:
			allowed = perm.CanManage
		default:
			allowed = false
		}
	}

	// Cache result for 5 minutes
	if redisClient != nil {
		cacheValue := "0"
		if allowed {
			cacheValue = "1"
		}
		redisClient.Set(ctx, cacheKey, cacheValue, 5*time.Minute)
	}

	return allowed, nil
}
