package domain

import (
	"time"
)

// User represents a gateway user
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null;size:100" json:"username"`
	Email     string    `gorm:"uniqueIndex;size:255" json:"email,omitempty"`
	Password  string    `gorm:"not null;size:255" json:"-"` // bcrypt hash
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Roles          []Role          `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	RefreshTokens  []RefreshToken  `gorm:"foreignKey:UserID" json:"-"`
	ChatPermissions []ChatPermission `gorm:"foreignKey:UserID" json:"-"`
}

// Role represents a user role for RBAC
type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null;size:50" json:"name"`
	Description string    `gorm:"size:255" json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	Users       []User       `gorm:"many2many:user_roles;" json:"-"`
}

// Permission represents a granular permission
type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null;size:100" json:"name"` // e.g., "messages:send", "chats:read"
	Description string    `gorm:"size:255" json:"description,omitempty"`
	Resource    string    `gorm:"not null;size:50" json:"resource"` // e.g., "messages", "chats", "bots"
	Action      string    `gorm:"not null;size:50" json:"action"`   // e.g., "read", "create", "update", "delete"
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	Roles []Role `gorm:"many2many:role_permissions;" json:"-"`
}

// Bot represents a registered Telegram bot
type Bot struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"uniqueIndex;not null;size:100" json:"username"`
	Token       string    `gorm:"uniqueIndex;not null;size:255" json:"-"` // Encrypted
	DisplayName string    `gorm:"size:255" json:"display_name,omitempty"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	WebhookURL  string    `gorm:"size:512" json:"webhook_url,omitempty"` // Set when registered with Telegram
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Chats []Chat `gorm:"foreignKey:BotID" json:"-"`
}

// Chat represents a Telegram chat associated with a bot
type Chat struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	BotID      uint      `gorm:"not null;index:idx_bot_telegram_chat" json:"bot_id"`
	TelegramID int64     `gorm:"not null;index:idx_bot_telegram_chat" json:"telegram_id"` // Telegram's chat ID
	Type       string    `gorm:"not null;size:50" json:"type"` // "private", "group", "supergroup", "channel"
	Title      string    `gorm:"size:255" json:"title,omitempty"`
	Username   string    `gorm:"size:100" json:"username,omitempty"`
	FirstName  string    `gorm:"size:255" json:"first_name,omitempty"`
	LastName   string    `gorm:"size:255" json:"last_name,omitempty"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	Bot              Bot              `gorm:"foreignKey:BotID" json:"bot,omitempty"`
	Messages         []Message        `gorm:"foreignKey:ChatID" json:"-"`
	ChatPermissions  []ChatPermission `gorm:"foreignKey:ChatID" json:"-"`
}

// ChatPermission represents granular chat-level access control
type ChatPermission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChatID    uint      `gorm:"not null;index:idx_chat_subject" json:"chat_id"`
	UserID    *uint     `gorm:"index:idx_chat_subject" json:"user_id,omitempty"`    // NULL if permission is for API key
	APIKeyID  *uint     `gorm:"index:idx_chat_subject" json:"api_key_id,omitempty"` // NULL if permission is for user
	CanRead   bool      `gorm:"default:false" json:"can_read"`
	CanSend   bool      `gorm:"default:false" json:"can_send"`
	CanManage bool      `gorm:"default:false" json:"can_manage"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Chat   Chat    `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	APIKey *APIKey `gorm:"foreignKey:APIKeyID" json:"api_key,omitempty"`
}

// APIKey represents a static API key for machine-to-machine auth
type APIKey struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Key         string    `gorm:"uniqueIndex;not null;size:100" json:"key"` // Visible part (e.g., "tgw_abc123...")
	HashedKey   string    `gorm:"uniqueIndex;not null;size:255" json:"-"`   // argon2id hash
	Name        string    `gorm:"not null;size:100" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Scopes      string    `gorm:"type:text" json:"scopes,omitempty"` // JSON array of scopes
	RateLimit   int       `gorm:"default:1000" json:"rate_limit"`    // Requests per hour
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	ChatPermissions     []ChatPermission             `gorm:"foreignKey:APIKeyID" json:"-"`
	BotPermissions      []APIKeyBotPermission        `gorm:"foreignKey:APIKeyID" json:"-"`
	FeedbackPermissions []APIKeyFeedbackPermission   `gorm:"foreignKey:APIKeyID" json:"-"`
}

// APIKeyBotPermission restricts which bots an API key can use
type APIKeyBotPermission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	APIKeyID  uint      `gorm:"not null;uniqueIndex:idx_apikey_bot" json:"api_key_id"`
	BotID     uint      `gorm:"not null;uniqueIndex:idx_apikey_bot" json:"bot_id"`
	CanSend   bool      `gorm:"default:true" json:"can_send"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	APIKey APIKey `gorm:"foreignKey:APIKeyID" json:"-"`
	Bot    Bot    `gorm:"foreignKey:BotID" json:"-"`
}

// APIKeyFeedbackPermission controls which chats can push messages
type APIKeyFeedbackPermission struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	APIKeyID           uint      `gorm:"not null;uniqueIndex:idx_apikey_feedback_chat" json:"api_key_id"`
	ChatID             uint      `gorm:"not null;uniqueIndex:idx_apikey_feedback_chat" json:"chat_id"`
	CanReceiveFeedback bool      `gorm:"default:true" json:"can_receive_feedback"`
	CreatedAt          time.Time `json:"created_at"`

	// Relationships
	APIKey APIKey `gorm:"foreignKey:APIKeyID" json:"-"`
	Chat   Chat   `gorm:"foreignKey:ChatID" json:"-"`
}

// Message represents a Telegram message
type Message struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ChatID          uint      `gorm:"not null;index:idx_chat_messages" json:"chat_id"`
	TelegramID      int64     `gorm:"not null;index" json:"telegram_id"` // Telegram's message ID
	FromUserID      *int64    `json:"from_user_id,omitempty"`
	FromUsername    string    `gorm:"size:100" json:"from_username,omitempty"`
	FromFirstName   string    `gorm:"size:255" json:"from_first_name,omitempty"`
	FromLastName    string    `gorm:"size:255" json:"from_last_name,omitempty"`
	Direction       string    `gorm:"not null;size:20;index:idx_chat_messages" json:"direction"` // "incoming", "outgoing"
	MessageType     string    `gorm:"not null;size:50" json:"message_type"` // "text", "photo", "video", etc.
	Text            string    `gorm:"type:text" json:"text,omitempty"`
	RawData         string    `gorm:"type:longtext" json:"-"` // Full Telegram message JSON
	ReplyToMessageID *int64   `gorm:"index" json:"reply_to_message_id,omitempty"`
	SentAt          time.Time `gorm:"not null;index:idx_chat_messages" json:"sent_at"`
	CreatedAt       time.Time `json:"created_at"`

	// Relationships
	Chat Chat `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
}

// Webhook represents a registered webhook endpoint
type Webhook struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	URL         string    `gorm:"not null;size:512" json:"url"`
	Secret      string    `gorm:"not null;size:255" json:"-"` // For HMAC signing
	Scope       string    `gorm:"not null;size:20" json:"scope"` // "chat" or "reply"
	ChatID      *uint     `gorm:"index" json:"chat_id,omitempty"` // NULL for global webhooks
	ReplyToMessageID *int64 `json:"reply_to_message_id,omitempty"` // For "reply" scope
	Events      string    `gorm:"type:text" json:"events,omitempty"` // JSON array of event types
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Chat       *Chat              `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
	Deliveries []WebhookDelivery  `gorm:"foreignKey:WebhookID" json:"-"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	WebhookID    uint       `gorm:"not null;index:idx_webhook_deliveries" json:"webhook_id"`
	MessageID    uint       `gorm:"not null;index" json:"message_id"`
	Status       string     `gorm:"not null;size:20;index:idx_webhook_deliveries" json:"status"` // "pending", "delivered", "failed"
	AttemptCount int        `gorm:"default:0" json:"attempt_count"`
	LastError    string     `gorm:"type:text" json:"last_error,omitempty"`
	NextRetryAt  *time.Time `gorm:"index:idx_webhook_deliveries" json:"next_retry_at,omitempty"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relationships
	Webhook Webhook `gorm:"foreignKey:WebhookID" json:"webhook,omitempty"`
	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
}

// RefreshToken represents a JWT refresh token
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null;size:255" json:"-"` // Hashed
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	RevokedAt *time.Time `gorm:"index" json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName methods to ensure correct table names
func (User) TableName() string                       { return "users" }
func (Role) TableName() string                       { return "roles" }
func (Permission) TableName() string                 { return "permissions" }
func (Bot) TableName() string                        { return "bots" }
func (Chat) TableName() string                       { return "chats" }
func (ChatPermission) TableName() string             { return "chat_permissions" }
func (APIKey) TableName() string                     { return "api_keys" }
func (APIKeyBotPermission) TableName() string        { return "api_key_bot_permissions" }
func (APIKeyFeedbackPermission) TableName() string   { return "api_key_feedback_permissions" }
func (Message) TableName() string                    { return "messages" }
func (Webhook) TableName() string                    { return "webhooks" }
func (WebhookDelivery) TableName() string            { return "webhook_deliveries" }
func (RefreshToken) TableName() string               { return "refresh_tokens" }
