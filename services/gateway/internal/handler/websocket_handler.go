package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/kexi/telegram-bot-gateway/internal/middleware"
	ws "github.com/kexi/telegram-bot-gateway/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking in production
		return true
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub *ws.Hub
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
// @Summary WebSocket connection
// @Description Upgrade to WebSocket for real-time message streaming
// @Tags websocket
// @Param token query string false "JWT token (alternative to Authorization header)"
// @Success 101 {string} string "Switching Protocols"
// @Router /api/v1/ws [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get auth context (should be set by auth middleware)
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		// Try to get token from query parameter as fallback
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}
		// TODO: Validate token and set auth context
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	// Create client ID
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())

	// Create client
	var userID, apiKeyID *uint
	if authCtx != nil {
		if authCtx.IsAPIKey {
			apiKeyID = authCtx.APIKeyID
		} else {
			userID = &authCtx.UserID
		}
	}

	client := ws.NewClient(clientID, h.hub, conn, userID, apiKeyID)

	// Register client with hub
	h.hub.RegisterClient(client)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}
