package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"
)

// Config represents the application configuration
type Config struct {
	Server          ServerConfig          `json:"server"`
	Database        DatabaseConfig        `json:"database"`
	Redis           RedisConfig           `json:"redis"`
	Auth            AuthConfig            `json:"auth"`
	Telegram        TelegramConfig        `json:"telegram"`
	WebhookDelivery WebhookDeliveryConfig `json:"webhook_delivery"`
	RateLimit       RateLimitConfig       `json:"rate_limit"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Mode string           `json:"mode"` // "debug", "release", "test"
	HTTP HTTPServerConfig `json:"http"`
	GRPC GRPCServerConfig `json:"grpc"`
}

// HTTPServerConfig holds HTTP server settings
type HTTPServerConfig struct {
	Address      string        `json:"address"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// GRPCServerConfig holds gRPC server settings
type GRPCServerConfig struct {
	Address string `json:"address"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Driver          string `json:"driver"` // "mysql", "postgres"
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Name            string `json:"name"`
	User            string `json:"user"`
	Password        string `json:"password"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime string `json:"conn_max_lifetime"`
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	JWT    JWTConfig    `json:"jwt"`
	APIKey APIKeyConfig `json:"api_key"`
}

// JWTConfig holds JWT token settings
type JWTConfig struct {
	Secret            string        `json:"secret"`
	AccessTokenTTL    time.Duration `json:"access_token_ttl"`
	RefreshTokenTTL   time.Duration `json:"refresh_token_ttl"`
	Issuer            string        `json:"issuer"`
	RefreshThreshold  time.Duration `json:"refresh_threshold"` // Auto-refresh if token expires within this duration
}

// APIKeyConfig holds API key settings
type APIKeyConfig struct {
	Prefix string `json:"prefix"` // e.g., "tgw_"
	Length int    `json:"length"` // Length of random part (default: 32)
}

// TelegramConfig holds Telegram API settings
type TelegramConfig struct {
	WebhookBaseURL string        `json:"webhook_base_url"` // e.g., "https://your-domain.com"
	Timeout        time.Duration `json:"timeout"`
}

// WebhookDeliveryConfig holds webhook worker settings
type WebhookDeliveryConfig struct {
	WorkerCount int           `json:"worker_count"`
	MaxRetries  int           `json:"max_retries"`
	Timeout     time.Duration `json:"timeout"`
	QueueName   string        `json:"queue_name"`
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	Burst             int           `json:"burst"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
}

// Load loads configuration from a JSON file
// Environment variables in the format ${VAR_NAME} are expanded
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := expandEnvVars(string(data))

	var cfg Config
	if err := json.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	cfg.setDefaults()

	// Validate
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// expandEnvVars replaces ${VAR_NAME} with environment variable values
func expandEnvVars(s string) string {
	re := regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[2 : len(match)-1] // Extract variable name
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match // Keep original if not found
	})
}

// setDefaults sets default values for optional fields
func (c *Config) setDefaults() {
	if c.Server.Mode == "" {
		c.Server.Mode = "release"
	}

	if c.Server.HTTP.Address == "" {
		c.Server.HTTP.Address = ":8080"
	}
	if c.Server.HTTP.ReadTimeout == 0 {
		c.Server.HTTP.ReadTimeout = 30 * time.Second
	}
	if c.Server.HTTP.WriteTimeout == 0 {
		c.Server.HTTP.WriteTimeout = 30 * time.Second
	}
	if c.Server.HTTP.IdleTimeout == 0 {
		c.Server.HTTP.IdleTimeout = 120 * time.Second
	}

	if c.Server.GRPC.Address == "" {
		c.Server.GRPC.Address = ":9090"
	}

	if c.Database.Driver == "" {
		c.Database.Driver = "mysql"
	}
	if c.Database.Port == 0 {
		c.Database.Port = 3306
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Database.ConnMaxLifetime == "" {
		c.Database.ConnMaxLifetime = "5m"
	}

	if c.Redis.DB == 0 {
		c.Redis.DB = 0
	}

	if c.Auth.JWT.AccessTokenTTL == 0 {
		c.Auth.JWT.AccessTokenTTL = 15 * time.Minute
	}
	if c.Auth.JWT.RefreshTokenTTL == 0 {
		c.Auth.JWT.RefreshTokenTTL = 7 * 24 * time.Hour
	}
	if c.Auth.JWT.Issuer == "" {
		c.Auth.JWT.Issuer = "telegram-bot-gateway"
	}
	if c.Auth.JWT.RefreshThreshold == 0 {
		c.Auth.JWT.RefreshThreshold = 5 * time.Minute
	}

	if c.Auth.APIKey.Prefix == "" {
		c.Auth.APIKey.Prefix = "tgw_"
	}
	if c.Auth.APIKey.Length == 0 {
		c.Auth.APIKey.Length = 32
	}

	if c.Telegram.Timeout == 0 {
		c.Telegram.Timeout = 30 * time.Second
	}

	if c.WebhookDelivery.WorkerCount == 0 {
		c.WebhookDelivery.WorkerCount = 10
	}
	if c.WebhookDelivery.MaxRetries == 0 {
		c.WebhookDelivery.MaxRetries = 5
	}
	if c.WebhookDelivery.Timeout == 0 {
		c.WebhookDelivery.Timeout = 30 * time.Second
	}
	if c.WebhookDelivery.QueueName == "" {
		c.WebhookDelivery.QueueName = "webhook_deliveries"
	}

	if c.RateLimit.RequestsPerSecond == 0 {
		c.RateLimit.RequestsPerSecond = 100
	}
	if c.RateLimit.Burst == 0 {
		c.RateLimit.Burst = 200
	}
	if c.RateLimit.CleanupInterval == 0 {
		c.RateLimit.CleanupInterval = 1 * time.Minute
	}
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Redis.Address == "" {
		return fmt.Errorf("redis address is required")
	}

	if c.Auth.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(c.Auth.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}

	if c.Telegram.WebhookBaseURL == "" {
		return fmt.Errorf("telegram webhook base URL is required")
	}

	return nil
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
			c.User, c.Password, c.Host, c.Port, c.Name)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Host, c.Port, c.User, c.Password, c.Name)
	default:
		return ""
	}
}
