package service

import (
	"context"
	"fmt"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
	"github.com/kexi/telegram-bot-gateway/internal/pkg/apikey"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
)

// APIKeyService handles API key operations
type APIKeyService struct {
	apiKeyRepo    repository.APIKeyRepository
	apiKeyPkg     *apikey.Service
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(
	apiKeyRepo repository.APIKeyRepository,
	apiKeyPkg *apikey.Service,
) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
		apiKeyPkg:  apiKeyPkg,
	}
}

// APIKeyDTO represents an API key data transfer object
type APIKeyDTO struct {
	ID          uint   `json:"id"`
	Key         string `json:"key"` // Only shown once during creation
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Scopes      string `json:"scopes,omitempty"`
	RateLimit   int    `json:"rate_limit"`
	IsActive    bool   `json:"is_active"`
}

// CreateAPIKeyRequest represents an API key creation request
type CreateAPIKeyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Scopes      string `json:"scopes"`
	RateLimit   int    `json:"rate_limit"`
}

// CreateAPIKey generates a new API key
func (s *APIKeyService) CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*APIKeyDTO, error) {
	// Generate API key
	key, hashedKey, err := s.apiKeyPkg.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	rateLimit := req.RateLimit
	if rateLimit <= 0 {
		rateLimit = 1000 // Default: 1000 requests per hour
	}

	apiKey := &domain.APIKey{
		Key:         key,
		HashedKey:   hashedKey,
		Name:        req.Name,
		Description: req.Description,
		Scopes:      req.Scopes,
		RateLimit:   rateLimit,
		IsActive:    true,
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return &APIKeyDTO{
		ID:          apiKey.ID,
		Key:         key, // Return plaintext key (only time it's shown)
		Name:        apiKey.Name,
		Description: apiKey.Description,
		Scopes:      apiKey.Scopes,
		RateLimit:   apiKey.RateLimit,
		IsActive:    apiKey.IsActive,
	}, nil
}

// GetAPIKey retrieves an API key by ID (without the actual key)
func (s *APIKeyService) GetAPIKey(ctx context.Context, id uint) (*APIKeyDTO, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("API key not found: %w", err)
	}

	return &APIKeyDTO{
		ID:          apiKey.ID,
		Key:         maskKey(apiKey.Key), // Masked
		Name:        apiKey.Name,
		Description: apiKey.Description,
		Scopes:      apiKey.Scopes,
		RateLimit:   apiKey.RateLimit,
		IsActive:    apiKey.IsActive,
	}, nil
}

// ListAPIKeys retrieves all API keys
func (s *APIKeyService) ListAPIKeys(ctx context.Context, offset, limit int) ([]APIKeyDTO, error) {
	apiKeys, err := s.apiKeyRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	result := make([]APIKeyDTO, len(apiKeys))
	for i, key := range apiKeys {
		result[i] = APIKeyDTO{
			ID:          key.ID,
			Key:         maskKey(key.Key),
			Name:        key.Name,
			Description: key.Description,
			Scopes:      key.Scopes,
			RateLimit:   key.RateLimit,
			IsActive:    key.IsActive,
		}
	}

	return result, nil
}

// RevokeAPIKey deactivates an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id uint) error {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("API key not found: %w", err)
	}

	apiKey.IsActive = false
	return s.apiKeyRepo.Update(ctx, apiKey)
}

// DeleteAPIKey permanently deletes an API key
func (s *APIKeyService) DeleteAPIKey(ctx context.Context, id uint) error {
	return s.apiKeyRepo.Delete(ctx, id)
}

// maskKey masks an API key for display (show only first 8 chars)
func maskKey(key string) string {
	if len(key) <= 12 {
		return key[:4] + "****"
	}
	return key[:12] + "..." + key[len(key)-4:]
}
