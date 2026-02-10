package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
)

// UserRepository defines operations for user management
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, offset, limit int) ([]domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uint) error
	WithRoles(ctx context.Context, userID uint) (*domain.User, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, id).Error
}

func (r *userRepository) WithRoles(ctx context.Context, userID uint) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Preload("Roles.Permissions").First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// BotRepository defines operations for bot management
type BotRepository interface {
	Create(ctx context.Context, bot *domain.Bot) error
	GetByID(ctx context.Context, id uint) (*domain.Bot, error)
	GetByUsername(ctx context.Context, username string) (*domain.Bot, error)
	GetByToken(ctx context.Context, token string) (*domain.Bot, error)
	GetByWebhookSecret(ctx context.Context, secret string) (*domain.Bot, error)
	List(ctx context.Context, offset, limit int) ([]domain.Bot, error)
	Update(ctx context.Context, bot *domain.Bot) error
	Delete(ctx context.Context, id uint) error
}

type botRepository struct {
	db *gorm.DB
}

// NewBotRepository creates a new bot repository
func NewBotRepository(db *gorm.DB) BotRepository {
	return &botRepository{db: db}
}

func (r *botRepository) Create(ctx context.Context, bot *domain.Bot) error {
	return r.db.WithContext(ctx).Create(bot).Error
}

func (r *botRepository) GetByID(ctx context.Context, id uint) (*domain.Bot, error) {
	var bot domain.Bot
	err := r.db.WithContext(ctx).First(&bot, id).Error
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *botRepository) GetByUsername(ctx context.Context, username string) (*domain.Bot, error) {
	var bot domain.Bot
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&bot).Error
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *botRepository) GetByToken(ctx context.Context, token string) (*domain.Bot, error) {
	var bot domain.Bot
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&bot).Error
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *botRepository) GetByWebhookSecret(ctx context.Context, secret string) (*domain.Bot, error) {
	var bot domain.Bot
	err := r.db.WithContext(ctx).Where("webhook_secret = ? AND is_active = true", secret).First(&bot).Error
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *botRepository) List(ctx context.Context, offset, limit int) ([]domain.Bot, error) {
	var bots []domain.Bot
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&bots).Error
	return bots, err
}

func (r *botRepository) Update(ctx context.Context, bot *domain.Bot) error {
	return r.db.WithContext(ctx).Save(bot).Error
}

func (r *botRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Bot{}, id).Error
}

// ChatRepository defines operations for chat management
type ChatRepository interface {
	Create(ctx context.Context, chat *domain.Chat) error
	GetByID(ctx context.Context, id uint) (*domain.Chat, error)
	GetByBotAndTelegramID(ctx context.Context, botID uint, telegramID int64) (*domain.Chat, error)
	List(ctx context.Context, offset, limit int) ([]domain.Chat, error)
	ListByBot(ctx context.Context, botID uint, offset, limit int) ([]domain.Chat, error)
	Update(ctx context.Context, chat *domain.Chat) error
	Delete(ctx context.Context, id uint) error
}

type chatRepository struct {
	db *gorm.DB
}

// NewChatRepository creates a new chat repository
func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Create(ctx context.Context, chat *domain.Chat) error {
	return r.db.WithContext(ctx).Create(chat).Error
}

func (r *chatRepository) GetByID(ctx context.Context, id uint) (*domain.Chat, error) {
	var chat domain.Chat
	err := r.db.WithContext(ctx).Preload("Bot").First(&chat, id).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *chatRepository) GetByBotAndTelegramID(ctx context.Context, botID uint, telegramID int64) (*domain.Chat, error) {
	var chat domain.Chat
	err := r.db.WithContext(ctx).
		Where("bot_id = ? AND telegram_id = ?", botID, telegramID).
		First(&chat).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *chatRepository) List(ctx context.Context, offset, limit int) ([]domain.Chat, error) {
	var chats []domain.Chat
	err := r.db.WithContext(ctx).Preload("Bot").Offset(offset).Limit(limit).Find(&chats).Error
	return chats, err
}

func (r *chatRepository) ListByBot(ctx context.Context, botID uint, offset, limit int) ([]domain.Chat, error) {
	var chats []domain.Chat
	err := r.db.WithContext(ctx).
		Where("bot_id = ?", botID).
		Offset(offset).Limit(limit).
		Find(&chats).Error
	return chats, err
}

func (r *chatRepository) Update(ctx context.Context, chat *domain.Chat) error {
	return r.db.WithContext(ctx).Save(chat).Error
}

func (r *chatRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Chat{}, id).Error
}

// MessageRepository defines operations for message management
type MessageRepository interface {
	Create(ctx context.Context, message *domain.Message) error
	GetByID(ctx context.Context, id uint) (*domain.Message, error)
	ListByChat(ctx context.Context, chatID uint, cursor *time.Time, limit int) ([]domain.Message, error)
	ListByReplyTo(ctx context.Context, replyToMessageID int64, offset, limit int) ([]domain.Message, error)
	Delete(ctx context.Context, id uint) error
	DeleteOlderThan(ctx context.Context, cutoff time.Time) error
}

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *domain.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *messageRepository) GetByID(ctx context.Context, id uint) (*domain.Message, error) {
	var message domain.Message
	err := r.db.WithContext(ctx).Preload("Chat").First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// ListByChat returns messages with cursor-based pagination
func (r *messageRepository) ListByChat(ctx context.Context, chatID uint, cursor *time.Time, limit int) ([]domain.Message, error) {
	var messages []domain.Message
	query := r.db.WithContext(ctx).Where("chat_id = ?", chatID).Order("sent_at DESC")

	if cursor != nil {
		query = query.Where("sent_at < ?", cursor)
	}

	err := query.Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *messageRepository) ListByReplyTo(ctx context.Context, replyToMessageID int64, offset, limit int) ([]domain.Message, error) {
	var messages []domain.Message
	err := r.db.WithContext(ctx).
		Where("reply_to_message_id = ?", replyToMessageID).
		Order("sent_at DESC").
		Offset(offset).Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Message{}, id).Error
}

func (r *messageRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&domain.Message{}).Error
}

// ChatPermissionRepository defines operations for chat permissions
type ChatPermissionRepository interface {
	Create(ctx context.Context, permission *domain.ChatPermission) error
	GetByID(ctx context.Context, id uint) (*domain.ChatPermission, error)
	GetByUserAndChat(ctx context.Context, userID, chatID uint) (*domain.ChatPermission, error)
	GetByAPIKeyAndChat(ctx context.Context, apiKeyID, chatID uint) (*domain.ChatPermission, error)
	ListByChat(ctx context.Context, chatID uint) ([]domain.ChatPermission, error)
	ListByUser(ctx context.Context, userID uint) ([]domain.ChatPermission, error)
	ListByAPIKey(ctx context.Context, apiKeyID uint) ([]domain.ChatPermission, error)
	Update(ctx context.Context, permission *domain.ChatPermission) error
	Delete(ctx context.Context, id uint) error
}

type chatPermissionRepository struct {
	db *gorm.DB
}

// NewChatPermissionRepository creates a new chat permission repository
func NewChatPermissionRepository(db *gorm.DB) ChatPermissionRepository {
	return &chatPermissionRepository{db: db}
}

func (r *chatPermissionRepository) Create(ctx context.Context, permission *domain.ChatPermission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *chatPermissionRepository) GetByID(ctx context.Context, id uint) (*domain.ChatPermission, error) {
	var permission domain.ChatPermission
	err := r.db.WithContext(ctx).First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *chatPermissionRepository) GetByUserAndChat(ctx context.Context, userID, chatID uint) (*domain.ChatPermission, error) {
	var permission domain.ChatPermission
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND chat_id = ?", userID, chatID).
		First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *chatPermissionRepository) GetByAPIKeyAndChat(ctx context.Context, apiKeyID, chatID uint) (*domain.ChatPermission, error) {
	var permission domain.ChatPermission
	err := r.db.WithContext(ctx).
		Where("api_key_id = ? AND chat_id = ?", apiKeyID, chatID).
		First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *chatPermissionRepository) ListByChat(ctx context.Context, chatID uint) ([]domain.ChatPermission, error) {
	var permissions []domain.ChatPermission
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("APIKey").
		Where("chat_id = ?", chatID).
		Find(&permissions).Error
	return permissions, err
}

func (r *chatPermissionRepository) ListByUser(ctx context.Context, userID uint) ([]domain.ChatPermission, error) {
	var permissions []domain.ChatPermission
	err := r.db.WithContext(ctx).
		Preload("Chat").
		Where("user_id = ?", userID).
		Find(&permissions).Error
	return permissions, err
}

func (r *chatPermissionRepository) Update(ctx context.Context, permission *domain.ChatPermission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *chatPermissionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.ChatPermission{}, id).Error
}

func (r *chatPermissionRepository) ListByAPIKey(ctx context.Context, apiKeyID uint) ([]domain.ChatPermission, error) {
	var permissions []domain.ChatPermission
	err := r.db.WithContext(ctx).
		Preload("Chat").
		Where("api_key_id = ?", apiKeyID).
		Find(&permissions).Error
	return permissions, err
}

// APIKeyRepository defines operations for API key management
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *domain.APIKey) error
	GetByID(ctx context.Context, id uint) (*domain.APIKey, error)
	GetByKey(ctx context.Context, key string) (*domain.APIKey, error)
	List(ctx context.Context, offset, limit int) ([]domain.APIKey, error)
	Update(ctx context.Context, apiKey *domain.APIKey) error
	Delete(ctx context.Context, id uint) error
	UpdateLastUsed(ctx context.Context, id uint) error
}

type apiKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *domain.APIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id uint) (*domain.APIKey, error) {
	var apiKey domain.APIKey
	err := r.db.WithContext(ctx).First(&apiKey, id).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) GetByKey(ctx context.Context, key string) (*domain.APIKey, error) {
	var apiKey domain.APIKey
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) List(ctx context.Context, offset, limit int) ([]domain.APIKey, error) {
	var apiKeys []domain.APIKey
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&apiKeys).Error
	return apiKeys, err
}

func (r *apiKeyRepository) Update(ctx context.Context, apiKey *domain.APIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

func (r *apiKeyRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.APIKey{}, id).Error
}

func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.APIKey{}).Where("id = ?", id).Update("last_used_at", now).Error
}

// WebhookRepository defines operations for webhook management
type WebhookRepository interface {
	Create(ctx context.Context, webhook *domain.Webhook) error
	GetByID(ctx context.Context, id uint) (*domain.Webhook, error)
	ListByChat(ctx context.Context, chatID uint) ([]domain.Webhook, error)
	ListActive(ctx context.Context) ([]domain.Webhook, error)
	Update(ctx context.Context, webhook *domain.Webhook) error
	Delete(ctx context.Context, id uint) error
}

type webhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) Create(ctx context.Context, webhook *domain.Webhook) error {
	return r.db.WithContext(ctx).Create(webhook).Error
}

func (r *webhookRepository) GetByID(ctx context.Context, id uint) (*domain.Webhook, error) {
	var webhook domain.Webhook
	err := r.db.WithContext(ctx).First(&webhook, id).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (r *webhookRepository) ListByChat(ctx context.Context, chatID uint) ([]domain.Webhook, error) {
	var webhooks []domain.Webhook
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Find(&webhooks).Error
	return webhooks, err
}

func (r *webhookRepository) ListActive(ctx context.Context) ([]domain.Webhook, error) {
	var webhooks []domain.Webhook
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&webhooks).Error
	return webhooks, err
}

func (r *webhookRepository) Update(ctx context.Context, webhook *domain.Webhook) error {
	return r.db.WithContext(ctx).Save(webhook).Error
}

func (r *webhookRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Webhook{}, id).Error
}

// WebhookDeliveryRepository defines operations for webhook delivery tracking
type WebhookDeliveryRepository interface {
	Create(ctx context.Context, delivery *domain.WebhookDelivery) error
	GetByID(ctx context.Context, id uint) (*domain.WebhookDelivery, error)
	GetPendingRetries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error)
	Update(ctx context.Context, delivery *domain.WebhookDelivery) error
	DeleteOlderThan(ctx context.Context, cutoff time.Time) error
}

type webhookDeliveryRepository struct {
	db *gorm.DB
}

// NewWebhookDeliveryRepository creates a new webhook delivery repository
func NewWebhookDeliveryRepository(db *gorm.DB) WebhookDeliveryRepository {
	return &webhookDeliveryRepository{db: db}
}

func (r *webhookDeliveryRepository) Create(ctx context.Context, delivery *domain.WebhookDelivery) error {
	return r.db.WithContext(ctx).Create(delivery).Error
}

func (r *webhookDeliveryRepository) GetByID(ctx context.Context, id uint) (*domain.WebhookDelivery, error) {
	var delivery domain.WebhookDelivery
	err := r.db.WithContext(ctx).Preload("Webhook").Preload("Message").First(&delivery, id).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *webhookDeliveryRepository) GetPendingRetries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error) {
	var deliveries []domain.WebhookDelivery
	now := time.Now()
	err := r.db.WithContext(ctx).
		Preload("Webhook").
		Preload("Message").
		Where("status = ? AND next_retry_at <= ?", "pending", now).
		Order("next_retry_at ASC").
		Limit(limit).
		Find(&deliveries).Error
	return deliveries, err
}

func (r *webhookDeliveryRepository) Update(ctx context.Context, delivery *domain.WebhookDelivery) error {
	return r.db.WithContext(ctx).Save(delivery).Error
}

func (r *webhookDeliveryRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&domain.WebhookDelivery{}).Error
}

// RefreshTokenRepository defines operations for refresh token management
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	GetActiveByUser(ctx context.Context, userID uint) ([]domain.RefreshToken, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllByUser(ctx context.Context, userID uint) error
	DeleteExpired(ctx context.Context) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("token = ? AND revoked_at IS NULL AND expires_at > ?", token, time.Now()).
		First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) GetActiveByUser(ctx context.Context, userID uint) ([]domain.RefreshToken, error) {
	var tokens []domain.RefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Find(&tokens).Error
	return tokens, err
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, token string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked_at", now).Error
}

func (r *refreshTokenRepository) RevokeAllByUser(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? OR revoked_at IS NOT NULL", time.Now().AddDate(0, 0, -30)).
		Delete(&domain.RefreshToken{}).Error
}

// APIKeyBotPermissionRepository defines operations for API key bot permissions
type APIKeyBotPermissionRepository interface {
	Create(ctx context.Context, perm *domain.APIKeyBotPermission) error
	ListByAPIKey(ctx context.Context, apiKeyID uint) ([]domain.APIKeyBotPermission, error)
	Delete(ctx context.Context, apiKeyID, botID uint) error
	HasBotAccess(ctx context.Context, apiKeyID, botID uint) (bool, error)
}

type apiKeyBotPermissionRepository struct {
	db *gorm.DB
}

// NewAPIKeyBotPermissionRepository creates a new API key bot permission repository
func NewAPIKeyBotPermissionRepository(db *gorm.DB) APIKeyBotPermissionRepository {
	return &apiKeyBotPermissionRepository{db: db}
}

func (r *apiKeyBotPermissionRepository) Create(ctx context.Context, perm *domain.APIKeyBotPermission) error {
	return r.db.WithContext(ctx).Create(perm).Error
}

func (r *apiKeyBotPermissionRepository) ListByAPIKey(ctx context.Context, apiKeyID uint) ([]domain.APIKeyBotPermission, error) {
	var perms []domain.APIKeyBotPermission
	err := r.db.WithContext(ctx).
		Preload("Bot").
		Where("api_key_id = ?", apiKeyID).
		Find(&perms).Error
	return perms, err
}

func (r *apiKeyBotPermissionRepository) Delete(ctx context.Context, apiKeyID, botID uint) error {
	return r.db.WithContext(ctx).
		Where("api_key_id = ? AND bot_id = ?", apiKeyID, botID).
		Delete(&domain.APIKeyBotPermission{}).Error
}

func (r *apiKeyBotPermissionRepository) HasBotAccess(ctx context.Context, apiKeyID, botID uint) (bool, error) {
	// First check if there are ANY bot permissions for this API key
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.APIKeyBotPermission{}).
		Where("api_key_id = ?", apiKeyID).
		Count(&count).Error; err != nil {
		return false, err
	}

	// If no bot permissions exist, all bots are allowed (default behavior)
	if count == 0 {
		return true, nil
	}

	// If bot permissions exist, check if this specific bot is allowed
	var perm domain.APIKeyBotPermission
	err := r.db.WithContext(ctx).
		Where("api_key_id = ? AND bot_id = ? AND can_send = ?", apiKeyID, botID, true).
		First(&perm).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// APIKeyFeedbackPermissionRepository defines operations for API key feedback permissions
type APIKeyFeedbackPermissionRepository interface {
	Create(ctx context.Context, perm *domain.APIKeyFeedbackPermission) error
	ListByAPIKey(ctx context.Context, apiKeyID uint) ([]domain.APIKeyFeedbackPermission, error)
	Delete(ctx context.Context, apiKeyID, chatID uint) error
	CanReceiveFeedback(ctx context.Context, apiKeyID, chatID uint) (bool, error)
}

type apiKeyFeedbackPermissionRepository struct {
	db *gorm.DB
}

// NewAPIKeyFeedbackPermissionRepository creates a new API key feedback permission repository
func NewAPIKeyFeedbackPermissionRepository(db *gorm.DB) APIKeyFeedbackPermissionRepository {
	return &apiKeyFeedbackPermissionRepository{db: db}
}

func (r *apiKeyFeedbackPermissionRepository) Create(ctx context.Context, perm *domain.APIKeyFeedbackPermission) error {
	return r.db.WithContext(ctx).Create(perm).Error
}

func (r *apiKeyFeedbackPermissionRepository) ListByAPIKey(ctx context.Context, apiKeyID uint) ([]domain.APIKeyFeedbackPermission, error) {
	var perms []domain.APIKeyFeedbackPermission
	err := r.db.WithContext(ctx).
		Preload("Chat").
		Where("api_key_id = ?", apiKeyID).
		Find(&perms).Error
	return perms, err
}

func (r *apiKeyFeedbackPermissionRepository) Delete(ctx context.Context, apiKeyID, chatID uint) error {
	return r.db.WithContext(ctx).
		Where("api_key_id = ? AND chat_id = ?", apiKeyID, chatID).
		Delete(&domain.APIKeyFeedbackPermission{}).Error
}

func (r *apiKeyFeedbackPermissionRepository) CanReceiveFeedback(ctx context.Context, apiKeyID, chatID uint) (bool, error) {
	// First check if there are ANY feedback permissions for this API key
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.APIKeyFeedbackPermission{}).
		Where("api_key_id = ?", apiKeyID).
		Count(&count).Error; err != nil {
		return false, err
	}

	// If no feedback permissions exist, all chats can send feedback (default behavior)
	if count == 0 {
		return true, nil
	}

	// If feedback permissions exist, check if this specific chat is allowed
	var perm domain.APIKeyFeedbackPermission
	err := r.db.WithContext(ctx).
		Where("api_key_id = ? AND chat_id = ? AND can_receive_feedback = ?", apiKeyID, chatID, true).
		First(&perm).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
