package service

import (
	"context"
	"fmt"
	"time"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
)

// MessageService handles message operations
type MessageService struct {
	messageRepo repository.MessageRepository
	chatRepo    repository.ChatRepository
}

// NewMessageService creates a new message service
func NewMessageService(
	messageRepo repository.MessageRepository,
	chatRepo repository.ChatRepository,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
	}
}

// MessageDTO represents a message data transfer object
type MessageDTO struct {
	ID             uint      `json:"id"`
	ChatID         uint      `json:"chat_id"`
	TelegramID     int64     `json:"telegram_id"`
	FromUserID     *int64    `json:"from_user_id,omitempty"`
	FromUsername   string    `json:"from_username,omitempty"`
	FromFirstName  string    `json:"from_first_name,omitempty"`
	FromLastName   string    `json:"from_last_name,omitempty"`
	Direction      string    `json:"direction"`
	MessageType    string    `json:"message_type"`
	Text           string    `json:"text,omitempty"`
	ReplyToMessageID *int64  `json:"reply_to_message_id,omitempty"`
	SentAt         time.Time `json:"sent_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateMessageRequest represents a message creation request
type CreateMessageRequest struct {
	ChatID           uint      `json:"chat_id"`
	TelegramID       int64     `json:"telegram_id"`
	FromUserID       *int64    `json:"from_user_id,omitempty"`
	FromUsername     string    `json:"from_username,omitempty"`
	FromFirstName    string    `json:"from_first_name,omitempty"`
	FromLastName     string    `json:"from_last_name,omitempty"`
	Direction        string    `json:"direction"`
	MessageType      string    `json:"message_type"`
	Text             string    `json:"text,omitempty"`
	RawData          string    `json:"raw_data,omitempty"`
	ReplyToMessageID *int64    `json:"reply_to_message_id,omitempty"`
	SentAt           time.Time `json:"sent_at"`
}

// StoreMessage stores a new message
func (s *MessageService) StoreMessage(ctx context.Context, req *CreateMessageRequest) (*MessageDTO, error) {
	// Verify chat exists
	_, err := s.chatRepo.GetByID(ctx, req.ChatID)
	if err != nil {
		return nil, fmt.Errorf("chat not found: %w", err)
	}

	message := &domain.Message{
		ChatID:           req.ChatID,
		TelegramID:       req.TelegramID,
		FromUserID:       req.FromUserID,
		FromUsername:     req.FromUsername,
		FromFirstName:    req.FromFirstName,
		FromLastName:     req.FromLastName,
		Direction:        req.Direction,
		MessageType:      req.MessageType,
		Text:             req.Text,
		RawData:          req.RawData,
		ReplyToMessageID: req.ReplyToMessageID,
		SentAt:           req.SentAt,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to store message: %w", err)
	}

	return &MessageDTO{
		ID:               message.ID,
		ChatID:           message.ChatID,
		TelegramID:       message.TelegramID,
		FromUserID:       message.FromUserID,
		FromUsername:     message.FromUsername,
		FromFirstName:    message.FromFirstName,
		FromLastName:     message.FromLastName,
		Direction:        message.Direction,
		MessageType:      message.MessageType,
		Text:             message.Text,
		ReplyToMessageID: message.ReplyToMessageID,
		SentAt:           message.SentAt,
		CreatedAt:        message.CreatedAt,
	}, nil
}

// GetMessage retrieves a message by ID
func (s *MessageService) GetMessage(ctx context.Context, id uint) (*MessageDTO, error) {
	message, err := s.messageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}

	return &MessageDTO{
		ID:               message.ID,
		ChatID:           message.ChatID,
		TelegramID:       message.TelegramID,
		FromUserID:       message.FromUserID,
		FromUsername:     message.FromUsername,
		FromFirstName:    message.FromFirstName,
		FromLastName:     message.FromLastName,
		Direction:        message.Direction,
		MessageType:      message.MessageType,
		Text:             message.Text,
		ReplyToMessageID: message.ReplyToMessageID,
		SentAt:           message.SentAt,
		CreatedAt:        message.CreatedAt,
	}, nil
}

// ListMessages retrieves messages for a chat with cursor-based pagination
func (s *MessageService) ListMessages(ctx context.Context, chatID uint, cursor *time.Time, limit int) ([]MessageDTO, error) {
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	messages, err := s.messageRepo.ListByChat(ctx, chatID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	result := make([]MessageDTO, len(messages))
	for i, msg := range messages {
		result[i] = MessageDTO{
			ID:               msg.ID,
			ChatID:           msg.ChatID,
			TelegramID:       msg.TelegramID,
			FromUserID:       msg.FromUserID,
			FromUsername:     msg.FromUsername,
			FromFirstName:    msg.FromFirstName,
			FromLastName:     msg.FromLastName,
			Direction:        msg.Direction,
			MessageType:      msg.MessageType,
			Text:             msg.Text,
			ReplyToMessageID: msg.ReplyToMessageID,
			SentAt:           msg.SentAt,
			CreatedAt:        msg.CreatedAt,
		}
	}

	return result, nil
}

// ListMessagesByReply retrieves messages that are replies to a specific message
func (s *MessageService) ListMessagesByReply(ctx context.Context, replyToMessageID int64, offset, limit int) ([]MessageDTO, error) {
	messages, err := s.messageRepo.ListByReplyTo(ctx, replyToMessageID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list reply messages: %w", err)
	}

	result := make([]MessageDTO, len(messages))
	for i, msg := range messages {
		result[i] = MessageDTO{
			ID:               msg.ID,
			ChatID:           msg.ChatID,
			TelegramID:       msg.TelegramID,
			FromUserID:       msg.FromUserID,
			FromUsername:     msg.FromUsername,
			FromFirstName:    msg.FromFirstName,
			FromLastName:     msg.FromLastName,
			Direction:        msg.Direction,
			MessageType:      msg.MessageType,
			Text:             msg.Text,
			ReplyToMessageID: msg.ReplyToMessageID,
			SentAt:           msg.SentAt,
			CreatedAt:        msg.CreatedAt,
		}
	}

	return result, nil
}

// CleanupOldMessages deletes messages older than the specified cutoff date
func (s *MessageService) CleanupOldMessages(ctx context.Context, cutoff time.Time) error {
	return s.messageRepo.DeleteOlderThan(ctx, cutoff)
}
