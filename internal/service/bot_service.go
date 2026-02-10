package service

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
)

// BotService handles bot management operations
type BotService struct {
	botRepo        repository.BotRepository
	encryptionKey  []byte
	webhookBaseURL string
	httpClient     *http.Client
}

// NewBotService creates a new bot service
func NewBotService(botRepo repository.BotRepository, encryptionKey, webhookBaseURL string) *BotService {
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
		botRepo:        botRepo,
		encryptionKey:  key,
		webhookBaseURL: webhookBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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

	// Generate random webhook secret (32 bytes = 64 hex chars)
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate webhook secret: %w", err)
	}
	webhookSecret := hex.EncodeToString(secretBytes)

	// Compute webhook URL
	webhookURL := fmt.Sprintf("%s/api/v1/telegram/webhook/%s", s.webhookBaseURL, webhookSecret)

	bot := &domain.Bot{
		Username:      req.Username,
		Token:         encryptedToken,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		IsActive:      true,
		WebhookURL:    webhookURL,
		WebhookSecret: webhookSecret,
	}

	// Create bot in database first
	if err := s.botRepo.Create(ctx, bot); err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	// Register webhook with Telegram
	if err := s.setTelegramWebhook(ctx, req.Token, webhookURL); err != nil {
		// Rollback: delete the bot record
		_ = s.botRepo.Delete(ctx, bot.ID)
		return nil, fmt.Errorf("failed to set Telegram webhook: %w", err)
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
	// Get bot to retrieve token for Telegram API call
	bot, err := s.botRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("bot not found: %w", err)
	}

	// Decrypt token
	token, err := s.decryptToken(bot.Token)
	if err != nil {
		return fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Delete webhook from Telegram
	if err := s.deleteTelegramWebhook(ctx, token); err != nil {
		// Log warning but don't fail - bot might already be deleted on Telegram side
		fmt.Printf("Warning: failed to delete Telegram webhook: %v\n", err)
	}

	// Delete from database
	return s.botRepo.Delete(ctx, id)
}

// GetBotByWebhookSecret retrieves a bot by webhook secret
func (s *BotService) GetBotByWebhookSecret(ctx context.Context, secret string) (*BotDTO, error) {
	bot, err := s.botRepo.GetByWebhookSecret(ctx, secret)
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

// SetWebhook re-registers the webhook with Telegram
func (s *BotService) SetWebhook(ctx context.Context, botID uint) error {
	bot, err := s.botRepo.GetByID(ctx, botID)
	if err != nil {
		return fmt.Errorf("bot not found: %w", err)
	}

	token, err := s.decryptToken(bot.Token)
	if err != nil {
		return fmt.Errorf("failed to decrypt token: %w", err)
	}

	return s.setTelegramWebhook(ctx, token, bot.WebhookURL)
}

// setTelegramWebhook calls Telegram API to set webhook
func (s *BotService) setTelegramWebhook(ctx context.Context, token, webhookURL string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", token)

	payload := map[string]string{
		"url": webhookURL,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Telegram API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Ok {
		return fmt.Errorf("Telegram API error: %s", result.Description)
	}

	return nil
}

// deleteTelegramWebhook calls Telegram API to delete webhook
func (s *BotService) deleteTelegramWebhook(ctx context.Context, token string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook", token)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Telegram API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
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
