package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
)

// WebhookService handles webhook operations
type WebhookService struct {
	webhookRepo repository.WebhookRepository
	chatRepo    repository.ChatRepository
}

// NewWebhookService creates a new webhook service
func NewWebhookService(
	webhookRepo repository.WebhookRepository,
	chatRepo repository.ChatRepository,
) *WebhookService {
	return &WebhookService{
		webhookRepo: webhookRepo,
		chatRepo:    chatRepo,
	}
}

// WebhookDTO represents a webhook data transfer object
type WebhookDTO struct {
	ID               uint   `json:"id"`
	URL              string `json:"url"`
	Secret           string `json:"secret,omitempty"` // Only shown once during creation
	Scope            string `json:"scope"`
	ChatID           *uint  `json:"chat_id,omitempty"`
	ReplyToMessageID *int64 `json:"reply_to_message_id,omitempty"`
	Events           string `json:"events,omitempty"`
	IsActive         bool   `json:"is_active"`
}

// CreateWebhookRequest represents a webhook creation request
type CreateWebhookRequest struct {
	URL              string `json:"url" binding:"required"`
	Scope            string `json:"scope" binding:"required"` // "chat" or "reply"
	ChatID           *uint  `json:"chat_id"`
	ReplyToMessageID *int64 `json:"reply_to_message_id"`
	Events           string `json:"events"`
}

// CreateWebhook registers a new webhook
func (s *WebhookService) CreateWebhook(ctx context.Context, req *CreateWebhookRequest) (*WebhookDTO, error) {
	// Validate scope
	if req.Scope != "chat" && req.Scope != "reply" {
		return nil, fmt.Errorf("invalid scope: must be 'chat' or 'reply'")
	}

	// Validate chat-level webhook
	if req.Scope == "chat" && req.ChatID != nil {
		_, err := s.chatRepo.GetByID(ctx, *req.ChatID)
		if err != nil {
			return nil, fmt.Errorf("chat not found: %w", err)
		}
	}

	// Validate reply-level webhook
	if req.Scope == "reply" {
		if req.ChatID == nil || req.ReplyToMessageID == nil {
			return nil, fmt.Errorf("reply scope requires both chat_id and reply_to_message_id")
		}
		_, err := s.chatRepo.GetByID(ctx, *req.ChatID)
		if err != nil {
			return nil, fmt.Errorf("chat not found: %w", err)
		}
	}

	// Generate webhook secret for HMAC signing
	secret, err := generateSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	webhook := &domain.Webhook{
		URL:              req.URL,
		Secret:           secret,
		Scope:            req.Scope,
		ChatID:           req.ChatID,
		ReplyToMessageID: req.ReplyToMessageID,
		Events:           req.Events,
		IsActive:         true,
	}

	if err := s.webhookRepo.Create(ctx, webhook); err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return &WebhookDTO{
		ID:               webhook.ID,
		URL:              webhook.URL,
		Secret:           secret, // Return plaintext secret (only time it's shown)
		Scope:            webhook.Scope,
		ChatID:           webhook.ChatID,
		ReplyToMessageID: webhook.ReplyToMessageID,
		Events:           webhook.Events,
		IsActive:         webhook.IsActive,
	}, nil
}

// GetWebhook retrieves a webhook by ID
func (s *WebhookService) GetWebhook(ctx context.Context, id uint) (*WebhookDTO, error) {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}

	return &WebhookDTO{
		ID:               webhook.ID,
		URL:              webhook.URL,
		Scope:            webhook.Scope,
		ChatID:           webhook.ChatID,
		ReplyToMessageID: webhook.ReplyToMessageID,
		Events:           webhook.Events,
		IsActive:         webhook.IsActive,
	}, nil
}

// ListWebhooksByChat retrieves webhooks for a specific chat
func (s *WebhookService) ListWebhooksByChat(ctx context.Context, chatID uint) ([]WebhookDTO, error) {
	webhooks, err := s.webhookRepo.ListByChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	result := make([]WebhookDTO, len(webhooks))
	for i, wh := range webhooks {
		result[i] = WebhookDTO{
			ID:               wh.ID,
			URL:              wh.URL,
			Scope:            wh.Scope,
			ChatID:           wh.ChatID,
			ReplyToMessageID: wh.ReplyToMessageID,
			Events:           wh.Events,
			IsActive:         wh.IsActive,
		}
	}

	return result, nil
}

// ListActiveWebhooks retrieves all active webhooks
func (s *WebhookService) ListActiveWebhooks(ctx context.Context) ([]domain.Webhook, error) {
	return s.webhookRepo.ListActive(ctx)
}

// UpdateWebhook updates webhook settings
func (s *WebhookService) UpdateWebhook(ctx context.Context, id uint, url string, isActive bool) error {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("webhook not found: %w", err)
	}

	if url != "" {
		webhook.URL = url
	}
	webhook.IsActive = isActive

	return s.webhookRepo.Update(ctx, webhook)
}

// DeleteWebhook deletes a webhook
func (s *WebhookService) DeleteWebhook(ctx context.Context, id uint) error {
	return s.webhookRepo.Delete(ctx, id)
}

// SignPayload creates an HMAC-SHA256 signature for webhook payload
func (s *WebhookService) SignPayload(secret string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies webhook signature
func (s *WebhookService) VerifySignature(secret string, payload []byte, signature string) bool {
	expectedSignature := s.SignPayload(secret, payload)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// generateSecret generates a random secret for webhook signing
func generateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
