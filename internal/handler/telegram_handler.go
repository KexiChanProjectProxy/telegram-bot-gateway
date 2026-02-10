package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kexi/telegram-bot-gateway/internal/pubsub"
	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// TelegramHandler handles Telegram webhook endpoints
type TelegramHandler struct {
	botService     *service.BotService
	chatService    *service.ChatService
	messageService *service.MessageService
	messageBroker  *pubsub.MessageBroker
}

// NewTelegramHandler creates a new Telegram handler
func NewTelegramHandler(
	botService *service.BotService,
	chatService *service.ChatService,
	messageService *service.MessageService,
	messageBroker *pubsub.MessageBroker,
) *TelegramHandler {
	return &TelegramHandler{
		botService:     botService,
		chatService:    chatService,
		messageService: messageService,
		messageBroker:  messageBroker,
	}
}

// TelegramUpdate represents a Telegram update
type TelegramUpdate struct {
	UpdateID      int64            `json:"update_id"`
	Message       *TelegramMessage `json:"message,omitempty"`
	EditedMessage *TelegramMessage `json:"edited_message,omitempty"`
	ChannelPost   *TelegramMessage `json:"channel_post,omitempty"`
	CallbackQuery *CallbackQuery   `json:"callback_query,omitempty"`
}

// TelegramMessage represents a Telegram message
type TelegramMessage struct {
	MessageID int64        `json:"message_id"`
	From      *TelegramUser `json:"from,omitempty"`
	Chat      *TelegramChat `json:"chat"`
	Date      int64        `json:"date"`
	Text      string       `json:"text,omitempty"`
	ReplyToMessage *TelegramMessage `json:"reply_to_message,omitempty"`

	// Additional message types
	Photo    []PhotoSize  `json:"photo,omitempty"`
	Video    *Video       `json:"video,omitempty"`
	Document *Document    `json:"document,omitempty"`
	Audio    *Audio       `json:"audio,omitempty"`
	Voice    *Voice       `json:"voice,omitempty"`
	Sticker  *Sticker     `json:"sticker,omitempty"`
}

// TelegramUser represents a Telegram user
type TelegramUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// TelegramChat represents a Telegram chat
type TelegramChat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"` // "private", "group", "supergroup", "channel"
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// CallbackQuery represents a callback query
type CallbackQuery struct {
	ID      string           `json:"id"`
	From    *TelegramUser    `json:"from"`
	Message *TelegramMessage `json:"message,omitempty"`
	Data    string           `json:"data,omitempty"`
}

// Media types
type PhotoSize struct {
	FileID   string `json:"file_id"`
	FileSize int    `json:"file_size,omitempty"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type Video struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Duration int    `json:"duration"`
}

type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name,omitempty"`
}

type Audio struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
}

type Voice struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
}

type Sticker struct {
	FileID string `json:"file_id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// ReceiveUpdate handles incoming Telegram updates
// @Summary Receive Telegram update
// @Description Webhook endpoint for Telegram bot updates
// @Tags telegram
// @Accept json
// @Produce json
// @Param webhook_secret path string true "Webhook secret"
// @Param update body TelegramUpdate true "Telegram update"
// @Success 200 {object} SuccessResponse
// @Router /telegram/webhook/{webhook_secret} [post]
func (h *TelegramHandler) ReceiveUpdate(c *gin.Context) {
	webhookSecret := c.Param("webhook_secret")
	if webhookSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook secret is required"})
		return
	}

	var update TelegramUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update format"})
		return
	}

	// Get bot from database by webhook secret
	bot, err := h.botService.GetBotByWebhookSecret(c.Request.Context(), webhookSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bot not found"})
		return
	}

	// Process the update
	if err := h.processUpdate(c.Request.Context(), bot.ID, &update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// processUpdate processes a Telegram update
func (h *TelegramHandler) processUpdate(ctx context.Context, botID uint, update *TelegramUpdate) error {
	// Determine which message to process
	var msg *TelegramMessage
	var messageType string

	switch {
	case update.Message != nil:
		msg = update.Message
		messageType = "new_message"
	case update.EditedMessage != nil:
		msg = update.EditedMessage
		messageType = "edited_message"
	case update.ChannelPost != nil:
		msg = update.ChannelPost
		messageType = "channel_post"
	default:
		// No message to process (could be callback query, etc.)
		return nil
	}

	// Create or update chat
	chatReq := &service.CreateChatRequest{
		BotID:      botID,
		TelegramID: msg.Chat.ID,
		Type:       msg.Chat.Type,
		Title:      msg.Chat.Title,
		Username:   msg.Chat.Username,
		FirstName:  msg.Chat.FirstName,
		LastName:   msg.Chat.LastName,
	}

	chat, err := h.chatService.CreateOrUpdateChat(ctx, chatReq)
	if err != nil {
		return fmt.Errorf("failed to create/update chat: %w", err)
	}

	// Determine message content type
	msgType := "text"
	text := msg.Text
	if len(msg.Photo) > 0 {
		msgType = "photo"
	} else if msg.Video != nil {
		msgType = "video"
	} else if msg.Document != nil {
		msgType = "document"
	} else if msg.Audio != nil {
		msgType = "audio"
	} else if msg.Voice != nil {
		msgType = "voice"
	} else if msg.Sticker != nil {
		msgType = "sticker"
	}

	// Extract from user info
	var fromUserID *int64
	var fromUsername, fromFirstName, fromLastName string
	if msg.From != nil {
		fromUserID = &msg.From.ID
		fromUsername = msg.From.Username
		fromFirstName = msg.From.FirstName
		fromLastName = msg.From.LastName
	}

	// Get reply-to message ID if present
	var replyToMessageID *int64
	if msg.ReplyToMessage != nil {
		replyToMessageID = &msg.ReplyToMessage.MessageID
	}

	// Serialize full message to JSON
	rawData, _ := json.Marshal(msg)

	// Store message
	messageReq := &service.CreateMessageRequest{
		ChatID:           chat.ID,
		TelegramID:       msg.MessageID,
		FromUserID:       fromUserID,
		FromUsername:     fromUsername,
		FromFirstName:    fromFirstName,
		FromLastName:     fromLastName,
		Direction:        "incoming",
		MessageType:      msgType,
		Text:             text,
		RawData:          string(rawData),
		ReplyToMessageID: replyToMessageID,
		SentAt:           time.Unix(msg.Date, 0),
	}

	storedMsg, err := h.messageService.StoreMessage(ctx, messageReq)
	if err != nil {
		return fmt.Errorf("failed to store message: %w", err)
	}

	// Publish to message broker for real-time distribution
	event := &pubsub.MessageEvent{
		Type:         messageType,
		ChatID:       chat.ID,
		MessageID:    storedMsg.ID,
		TelegramID:   msg.MessageID,
		BotID:        botID,
		Direction:    "incoming",
		Text:         text,
		FromUsername: fromUsername,
		Timestamp:    time.Unix(msg.Date, 0),
		Payload: map[string]interface{}{
			"message_type": msgType,
			"from_user_id": fromUserID,
		},
	}

	if err := h.messageBroker.PublishMessage(ctx, event); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to publish message event: %v\n", err)
	}

	// TODO: Queue webhook deliveries for this message
	// For each active webhook matching this chat/scope, create a delivery job

	return nil
}
