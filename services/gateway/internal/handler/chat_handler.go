package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// ChatHandler handles chat endpoints
type ChatHandler struct {
	chatService    *service.ChatService
	messageService *service.MessageService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService *service.ChatService, messageService *service.MessageService) *ChatHandler {
	return &ChatHandler{
		chatService:    chatService,
		messageService: messageService,
	}
}

// ListChats handles listing chats
// @Summary List chats
// @Description Get list of accessible chats
// @Tags chats
// @Produce json
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Success 200 {array} service.ChatDTO
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/chats [get]
func (h *ChatHandler) ListChats(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	chats, err := h.chatService.ListChats(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chats)
}

// GetChat handles getting a specific chat
// @Summary Get chat
// @Description Get chat by ID
// @Tags chats
// @Produce json
// @Param id path int true "Chat ID"
// @Success 200 {object} service.ChatDTO
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/chats/{id} [get]
func (h *ChatHandler) GetChat(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	chat, err := h.chatService.GetChat(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat)
}

// GetMessages handles listing messages in a chat
// @Summary Get chat messages
// @Description Get messages for a chat with cursor-based pagination
// @Tags chats
// @Produce json
// @Param id path int true "Chat ID"
// @Param cursor query string false "Cursor (RFC3339 timestamp)"
// @Param limit query int false "Limit"
// @Success 200 {array} service.MessageDTO
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/chats/{id}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	chatID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	var cursor *time.Time
	cursorStr := c.Query("cursor")
	if cursorStr != "" {
		t, err := time.Parse(time.RFC3339, cursorStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cursor format"})
			return
		}
		cursor = &t
	}

	messages, err := h.messageService.ListMessages(c.Request.Context(), uint(chatID), cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// SendMessage handles sending a message to a chat
// @Summary Send message
// @Description Send a message to a chat (requires can_send permission)
// @Tags chats
// @Accept json
// @Produce json
// @Param id path int true "Chat ID"
// @Param request body SendMessageRequest true "Message content"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/chats/{id}/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	chatID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Integrate with Telegram Bot API to actually send the message
	// For now, just acknowledge with Telegram Bot API-compatible response
	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"message_id": time.Now().Unix(), // Temporary placeholder
		"chat_id":    chatID,
		"text":       req.Text,
		"queued_at":  time.Now(),
	})
}

// SendMessageRequest represents a message send request
type SendMessageRequest struct {
	Text             string `json:"text" binding:"required"`
	ReplyToMessageID *int64 `json:"reply_to_message_id,omitempty"`
}
