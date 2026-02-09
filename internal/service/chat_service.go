package service

import (
	"context"
	"fmt"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
)

// ChatService handles chat operations
type ChatService struct {
	chatRepo repository.ChatRepository
	botRepo  repository.BotRepository
}

// NewChatService creates a new chat service
func NewChatService(
	chatRepo repository.ChatRepository,
	botRepo repository.BotRepository,
) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		botRepo:  botRepo,
	}
}

// ChatDTO represents a chat data transfer object
type ChatDTO struct {
	ID         uint   `json:"id"`
	BotID      uint   `json:"bot_id"`
	TelegramID int64  `json:"telegram_id"`
	Type       string `json:"type"`
	Title      string `json:"title,omitempty"`
	Username   string `json:"username,omitempty"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	IsActive   bool   `json:"is_active"`
}

// CreateChatRequest represents a chat creation request
type CreateChatRequest struct {
	BotID      uint   `json:"bot_id"`
	TelegramID int64  `json:"telegram_id"`
	Type       string `json:"type"`
	Title      string `json:"title,omitempty"`
	Username   string `json:"username,omitempty"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
}

// CreateOrUpdateChat creates a new chat or updates if it exists
func (s *ChatService) CreateOrUpdateChat(ctx context.Context, req *CreateChatRequest) (*ChatDTO, error) {
	// Verify bot exists
	_, err := s.botRepo.GetByID(ctx, req.BotID)
	if err != nil {
		return nil, fmt.Errorf("bot not found: %w", err)
	}

	// Check if chat already exists
	existingChat, err := s.chatRepo.GetByBotAndTelegramID(ctx, req.BotID, req.TelegramID)
	if err == nil {
		// Update existing chat
		existingChat.Type = req.Type
		existingChat.Title = req.Title
		existingChat.Username = req.Username
		existingChat.FirstName = req.FirstName
		existingChat.LastName = req.LastName
		existingChat.IsActive = true

		if err := s.chatRepo.Update(ctx, existingChat); err != nil {
			return nil, fmt.Errorf("failed to update chat: %w", err)
		}

		return &ChatDTO{
			ID:         existingChat.ID,
			BotID:      existingChat.BotID,
			TelegramID: existingChat.TelegramID,
			Type:       existingChat.Type,
			Title:      existingChat.Title,
			Username:   existingChat.Username,
			FirstName:  existingChat.FirstName,
			LastName:   existingChat.LastName,
			IsActive:   existingChat.IsActive,
		}, nil
	}

	// Create new chat
	chat := &domain.Chat{
		BotID:      req.BotID,
		TelegramID: req.TelegramID,
		Type:       req.Type,
		Title:      req.Title,
		Username:   req.Username,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		IsActive:   true,
	}

	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	return &ChatDTO{
		ID:         chat.ID,
		BotID:      chat.BotID,
		TelegramID: chat.TelegramID,
		Type:       chat.Type,
		Title:      chat.Title,
		Username:   chat.Username,
		FirstName:  chat.FirstName,
		LastName:   chat.LastName,
		IsActive:   chat.IsActive,
	}, nil
}

// GetChat retrieves a chat by ID
func (s *ChatService) GetChat(ctx context.Context, id uint) (*ChatDTO, error) {
	chat, err := s.chatRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("chat not found: %w", err)
	}

	return &ChatDTO{
		ID:         chat.ID,
		BotID:      chat.BotID,
		TelegramID: chat.TelegramID,
		Type:       chat.Type,
		Title:      chat.Title,
		Username:   chat.Username,
		FirstName:  chat.FirstName,
		LastName:   chat.LastName,
		IsActive:   chat.IsActive,
	}, nil
}

// ListChats retrieves all chats
func (s *ChatService) ListChats(ctx context.Context, offset, limit int) ([]ChatDTO, error) {
	chats, err := s.chatRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}

	result := make([]ChatDTO, len(chats))
	for i, chat := range chats {
		result[i] = ChatDTO{
			ID:         chat.ID,
			BotID:      chat.BotID,
			TelegramID: chat.TelegramID,
			Type:       chat.Type,
			Title:      chat.Title,
			Username:   chat.Username,
			FirstName:  chat.FirstName,
			LastName:   chat.LastName,
			IsActive:   chat.IsActive,
		}
	}

	return result, nil
}

// ListChatsByBot retrieves chats for a specific bot
func (s *ChatService) ListChatsByBot(ctx context.Context, botID uint, offset, limit int) ([]ChatDTO, error) {
	chats, err := s.chatRepo.ListByBot(ctx, botID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}

	result := make([]ChatDTO, len(chats))
	for i, chat := range chats {
		result[i] = ChatDTO{
			ID:         chat.ID,
			BotID:      chat.BotID,
			TelegramID: chat.TelegramID,
			Type:       chat.Type,
			Title:      chat.Title,
			Username:   chat.Username,
			FirstName:  chat.FirstName,
			LastName:   chat.LastName,
			IsActive:   chat.IsActive,
		}
	}

	return result, nil
}

// UpdateChat updates chat information
func (s *ChatService) UpdateChat(ctx context.Context, id uint, isActive bool) error {
	chat, err := s.chatRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("chat not found: %w", err)
	}

	chat.IsActive = isActive
	return s.chatRepo.Update(ctx, chat)
}

// DeleteChat deletes a chat
func (s *ChatService) DeleteChat(ctx context.Context, id uint) error {
	return s.chatRepo.Delete(ctx, id)
}
