package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
)

// BotService handles bot management operations
type BotService struct {
	botRepo        repository.BotRepository
	encryptionKey  []byte
}

// NewBotService creates a new bot service
func NewBotService(botRepo repository.BotRepository, encryptionKey string) *BotService {
	// Ensure key is 32 bytes for AES-256
	key := []byte(encryptionKey)
	if len(key) < 32 {
		// Pad with zeros if too short (in production, use proper key derivation)
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		key = key[:32]
	}

	return &BotService{
		botRepo:       botRepo,
		encryptionKey: key,
	}
}

// CreateBotRequest represents a bot creation request
type CreateBotRequest struct {
	Username    string `json:"username" binding:"required"`
	Token       string `json:"token" binding:"required"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// BotDTO represents bot data transfer object
type BotDTO struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"is_active"`
	WebhookURL  string `json:"webhook_url,omitempty"`
}

// CreateBot registers a new Telegram bot
func (s *BotService) CreateBot(ctx context.Context, req *CreateBotRequest) (*BotDTO, error) {
	// Encrypt the bot token
	encryptedToken, err := s.encryptToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt token: %w", err)
	}

	bot := &domain.Bot{
		Username:    req.Username,
		Token:       encryptedToken,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsActive:    true,
	}

	if err := s.botRepo.Create(ctx, bot); err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &BotDTO{
		ID:          bot.ID,
		Username:    bot.Username,
		DisplayName: bot.DisplayName,
		Description: bot.Description,
		IsActive:    bot.IsActive,
	}, nil
}

// GetBot retrieves a bot by ID
func (s *BotService) GetBot(ctx context.Context, id uint) (*BotDTO, error) {
	bot, err := s.botRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bot not found: %w", err)
	}

	return &BotDTO{
		ID:          bot.ID,
		Username:    bot.Username,
		DisplayName: bot.DisplayName,
		Description: bot.Description,
		IsActive:    bot.IsActive,
		WebhookURL:  bot.WebhookURL,
	}, nil
}

// GetBotToken retrieves and decrypts a bot's token
func (s *BotService) GetBotToken(ctx context.Context, botID uint) (string, error) {
	bot, err := s.botRepo.GetByID(ctx, botID)
	if err != nil {
		return "", fmt.Errorf("bot not found: %w", err)
	}

	token, err := s.decryptToken(bot.Token)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	return token, nil
}

// ValidateBotToken checks if a token belongs to a registered bot
func (s *BotService) ValidateBotToken(ctx context.Context, token string) (*domain.Bot, error) {
	// This is tricky because tokens are encrypted in DB
	// For webhook validation, we'd typically extract bot username from the webhook URL
	// and validate against that bot's decrypted token
	// For now, this is a placeholder that would need refinement
	return nil, fmt.Errorf("not implemented - use ValidateBotByUsername")
}

// ValidateBotByUsername validates a bot token against a known bot username
func (s *BotService) ValidateBotByUsername(ctx context.Context, username, token string) (*domain.Bot, error) {
	bot, err := s.botRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("bot not found: %w", err)
	}

	decryptedToken, err := s.decryptToken(bot.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	if decryptedToken != token {
		return nil, fmt.Errorf("invalid token")
	}

	return bot, nil
}

// ListBots returns a list of all bots
func (s *BotService) ListBots(ctx context.Context, offset, limit int) ([]BotDTO, error) {
	bots, err := s.botRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list bots: %w", err)
	}

	result := make([]BotDTO, len(bots))
	for i, bot := range bots {
		result[i] = BotDTO{
			ID:          bot.ID,
			Username:    bot.Username,
			DisplayName: bot.DisplayName,
			Description: bot.Description,
			IsActive:    bot.IsActive,
			WebhookURL:  bot.WebhookURL,
		}
	}

	return result, nil
}

// UpdateBot updates bot information
func (s *BotService) UpdateBot(ctx context.Context, id uint, displayName, description string, isActive bool) error {
	bot, err := s.botRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("bot not found: %w", err)
	}

	if displayName != "" {
		bot.DisplayName = displayName
	}
	if description != "" {
		bot.Description = description
	}
	bot.IsActive = isActive

	return s.botRepo.Update(ctx, bot)
}

// DeleteBot deletes a bot
func (s *BotService) DeleteBot(ctx context.Context, id uint) error {
	return s.botRepo.Delete(ctx, id)
}

// encryptToken encrypts a bot token using AES-256-GCM
func (s *BotService) encryptToken(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptToken decrypts a bot token
func (s *BotService) decryptToken(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
