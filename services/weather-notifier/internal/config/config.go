package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Config represents the application configuration structure
type Config struct {
	Telegram  TelegramConfig  `json:"telegram"`
	Caiyun    CaiyunConfig    `json:"caiyun"`
	LLM       LLMConfig       `json:"llm"`
	Schedule  ScheduleConfig  `json:"schedule"`
	Detection DetectionConfig `json:"detection"`
	Logging   LoggingConfig   `json:"logging"`
	Chats     []ChatConfig    `json:"chats"`
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	APIKey string `json:"api_key"`
	APIURL string `json:"api_url"`
}

// CaiyunConfig holds Caiyun Weather API configuration
type CaiyunConfig struct {
	APIToken string `json:"api_token"`
	Timeout  int    `json:"timeout"` // in seconds
}

// LLMConfig holds LLM API configuration
type LLMConfig struct {
	Provider    string  `json:"provider"` // "openai", "anthropic", etc.
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	BaseURL     string  `json:"base_url"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Timeout     int     `json:"timeout"` // in seconds
}

// ScheduleConfig holds scheduling configuration
type ScheduleConfig struct {
	Timezone     string `json:"timezone"`      // e.g., "Asia/Shanghai"
	MorningTime  string `json:"morning_time"`  // e.g., "08:00:00" (HH:MM:SS format)
	EveningTime  string `json:"evening_time"`  // e.g., "23:30:00" (HH:MM:SS format)
	PollInterval string `json:"poll_interval"` // e.g., "15m" (duration format)
}

// DetectionConfig holds weather detection thresholds
type DetectionConfig struct {
	TemperatureDelta float64 `json:"temperature_delta"` // in Celsius
	WindSpeedDelta   float64 `json:"wind_speed_delta"`  // in m/s
	VisibilityDelta  float64 `json:"visibility_delta"`  // in km
	AQICNDelta       int     `json:"aqi_cn_delta"`      // AQI CN delta
	AQIUSADelta      int     `json:"aqi_usa_delta"`     // AQI USA delta
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level       string `json:"level"`        // "debug", "info", "warn", "error"
	PrettyPrint bool   `json:"pretty_print"` // Enable pretty console output
}

// ChatConfig represents a single chat subscription
type ChatConfig struct {
	ChatID    int64            `json:"chat_id"`
	Name      string           `json:"name"`
	Locations []LocationConfig `json:"locations"`
	LLM       *LLMOverride     `json:"llm,omitempty"` // Optional per-chat LLM overrides
}

// LocationConfig represents a location to track
type LocationConfig struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// LLMOverride holds optional per-chat LLM configuration overrides
// Uses pointers to distinguish "not set" from "set to zero"
type LLMOverride struct {
	Provider    *string  `json:"provider,omitempty"`
	APIKey      *string  `json:"api_key,omitempty"`
	Model       *string  `json:"model,omitempty"`
	BaseURL     *string  `json:"base_url,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
}

// ResolveLLM merges per-chat overrides with global LLM config
func (c *ChatConfig) ResolveLLM(global LLMConfig) LLMConfig {
	resolved := global // Start with global config

	if c.LLM != nil {
		if c.LLM.Provider != nil {
			resolved.Provider = *c.LLM.Provider
		}
		if c.LLM.APIKey != nil {
			resolved.APIKey = *c.LLM.APIKey
		}
		if c.LLM.Model != nil {
			resolved.Model = *c.LLM.Model
		}
		if c.LLM.BaseURL != nil {
			resolved.BaseURL = *c.LLM.BaseURL
		}
		if c.LLM.MaxTokens != nil {
			resolved.MaxTokens = *c.LLM.MaxTokens
		}
		if c.LLM.Temperature != nil {
			resolved.Temperature = *c.LLM.Temperature
		}
	}

	return resolved
}

// Load reads configuration from config.json and expands environment variables
func Load(configPath string) (*Config, error) {
	// Default config path
	if configPath == "" {
		configPath = "config.json"
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in the JSON content
	expanded := os.ExpandEnv(string(data))

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Set defaults
	setDefaults(&cfg)

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(cfg *Config) {
	// Telegram defaults
	if cfg.Telegram.APIURL == "" {
		cfg.Telegram.APIURL = "https://api.telegram.org"
	}

	// Caiyun defaults
	if cfg.Caiyun.Timeout == 0 {
		cfg.Caiyun.Timeout = 30
	}

	// LLM defaults
	if cfg.LLM.Provider == "" {
		cfg.LLM.Provider = "openai"
	}
	if cfg.LLM.Model == "" {
		cfg.LLM.Model = "gpt-4"
	}
	if cfg.LLM.MaxTokens == 0 {
		cfg.LLM.MaxTokens = 500
	}
	if cfg.LLM.Temperature == 0 {
		cfg.LLM.Temperature = 0.7
	}
	if cfg.LLM.Timeout == 0 {
		cfg.LLM.Timeout = 60
	}

	// Schedule defaults
	if cfg.Schedule.Timezone == "" {
		cfg.Schedule.Timezone = "Asia/Shanghai"
	}
	if cfg.Schedule.MorningTime == "" {
		cfg.Schedule.MorningTime = "08:00:00"
	}
	if cfg.Schedule.EveningTime == "" {
		cfg.Schedule.EveningTime = "23:30:00"
	}
	if cfg.Schedule.PollInterval == "" {
		cfg.Schedule.PollInterval = "15m"
	}

	// Detection defaults
	if cfg.Detection.TemperatureDelta == 0 {
		cfg.Detection.TemperatureDelta = 3.0
	}
	if cfg.Detection.WindSpeedDelta == 0 {
		cfg.Detection.WindSpeedDelta = 5.0
	}
	if cfg.Detection.VisibilityDelta == 0 {
		cfg.Detection.VisibilityDelta = 5.0
	}
	if cfg.Detection.AQICNDelta == 0 {
		cfg.Detection.AQICNDelta = 50
	}
	if cfg.Detection.AQIUSADelta == 0 {
		cfg.Detection.AQIUSADelta = 50
	}

	// Logging defaults
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	// PrettyPrint defaults to false (zero value)
}

// validate checks if the configuration is valid
func validate(cfg *Config) error {
	// Validate Telegram config
	if cfg.Telegram.APIKey == "" {
		return fmt.Errorf("telegram.api_key is required")
	}

	// Validate Caiyun config
	if cfg.Caiyun.APIToken == "" {
		return fmt.Errorf("caiyun.api_token is required")
	}

	// Validate LLM config
	if cfg.LLM.APIKey == "" {
		return fmt.Errorf("llm.api_key is required")
	}
	if cfg.LLM.Model == "" {
		return fmt.Errorf("llm.model is required")
	}

	// Validate Schedule config
	if cfg.Schedule.MorningTime == "" {
		return fmt.Errorf("schedule.morning_time is required")
	}
	if cfg.Schedule.EveningTime == "" {
		return fmt.Errorf("schedule.evening_time is required")
	}
	if cfg.Schedule.PollInterval == "" {
		return fmt.Errorf("schedule.poll_interval is required")
	}

	// Validate Logging config
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[cfg.Logging.Level] {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error")
	}

	// Validate Chats config
	if len(cfg.Chats) == 0 {
		return fmt.Errorf("at least one chat must be configured in chats array")
	}

	chatIDs := make(map[int64]bool)
	for i, chat := range cfg.Chats {
		// Validate chat ID
		if chat.ChatID == 0 {
			return fmt.Errorf("chats[%d].chat_id is required and must be non-zero", i)
		}

		// Check for duplicate chat IDs
		if chatIDs[chat.ChatID] {
			return fmt.Errorf("duplicate chat_id %d found in chats array", chat.ChatID)
		}
		chatIDs[chat.ChatID] = true

		// Validate locations
		if len(chat.Locations) == 0 {
			return fmt.Errorf("chats[%d] (chat_id=%d) must have at least one location", i, chat.ChatID)
		}

		locationNames := make(map[string]bool)
		for j, loc := range chat.Locations {
			// Validate location name
			if loc.Name == "" {
				return fmt.Errorf("chats[%d].locations[%d].name is required", i, j)
			}

			// Check for duplicate location names within chat
			if locationNames[loc.Name] {
				return fmt.Errorf("duplicate location name '%s' in chats[%d] (chat_id=%d)", loc.Name, i, chat.ChatID)
			}
			locationNames[loc.Name] = true

			// Validate latitude
			if loc.Latitude < -90 || loc.Latitude > 90 {
				return fmt.Errorf("chats[%d].locations[%d].latitude must be between -90 and 90", i, j)
			}

			// Validate longitude
			if loc.Longitude < -180 || loc.Longitude > 180 {
				return fmt.Errorf("chats[%d].locations[%d].longitude must be between -180 and 180", i, j)
			}
		}
	}

	return nil
}

// sanitizeFileName removes or replaces characters unsafe for filenames
func sanitizeFileName(name string) string {
	// Replace spaces and special characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	sanitized := reg.ReplaceAllString(name, "_")
	// Remove leading/trailing underscores
	sanitized = strings.Trim(sanitized, "_")
	// Ensure non-empty result
	if sanitized == "" {
		sanitized = "unnamed"
	}
	return sanitized
}

// SanitizeFileName is exported for use by other packages
func SanitizeFileName(name string) string {
	return sanitizeFileName(name)
}
