package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration structure
type Config struct {
	Telegram  TelegramConfig  `mapstructure:"telegram"`
	Caiyun    CaiyunConfig    `mapstructure:"caiyun"`
	LLM       LLMConfig       `mapstructure:"llm"`
	Schedule  ScheduleConfig  `mapstructure:"schedule"`
	Detection DetectionConfig `mapstructure:"detection"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken    string   `mapstructure:"bot_token"`
	Password    string   `mapstructure:"password"`
	APIURL      string   `mapstructure:"api_url"`
	AdminUserID int64    `mapstructure:"admin_user_id"`
	AllowedIDs  []int64  `mapstructure:"allowed_ids"`
}

// CaiyunConfig holds Caiyun Weather API configuration
type CaiyunConfig struct {
	APIToken  string  `mapstructure:"api_token"`
	Latitude  float64 `mapstructure:"latitude"`
	Longitude float64 `mapstructure:"longitude"`
	Timeout   int     `mapstructure:"timeout"` // in seconds
}

// LLMConfig holds LLM API configuration
type LLMConfig struct {
	Provider string `mapstructure:"provider"` // "openai", "anthropic", etc.
	APIKey   string `mapstructure:"api_key"`
	Model    string `mapstructure:"model"`
	BaseURL  string `mapstructure:"base_url"`
	Timeout  int    `mapstructure:"timeout"` // in seconds
}

// ScheduleConfig holds scheduling configuration
type ScheduleConfig struct {
	CheckWeatherCron string `mapstructure:"check_weather_cron"` // e.g., "0 7,12,18 * * *"
	Timezone         string `mapstructure:"timezone"`           // e.g., "Asia/Shanghai"
}

// DetectionConfig holds weather detection thresholds
type DetectionConfig struct {
	RainProbabilityThreshold float64 `mapstructure:"rain_probability_threshold"` // 0.0 - 1.0
	TemperatureHighThreshold float64 `mapstructure:"temperature_high_threshold"` // in Celsius
	TemperatureLowThreshold  float64 `mapstructure:"temperature_low_threshold"`  // in Celsius
	AQIThreshold             int     `mapstructure:"aqi_threshold"`              // Air Quality Index
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`       // "debug", "info", "warn", "error"
	PrettyPrint bool   `mapstructure:"pretty_print"` // Enable pretty console output
}

// Load reads configuration from config.yaml and environment variables
// Environment variables take precedence over config file values
// Environment variables should be prefixed with WNB_ (Weather Notice Bot)
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Configure Viper to read from config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Enable environment variable support
	v.SetEnvPrefix("WNB") // Prefix for environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file (ignore error if file doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, rely on defaults and env vars
	}

	// Expand environment variables in config values
	expandEnvVars(v)

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Telegram defaults
	v.SetDefault("telegram.api_url", "https://api.telegram.org")
	v.SetDefault("telegram.admin_user_id", 0)
	v.SetDefault("telegram.allowed_ids", []int64{})

	// Caiyun defaults
	v.SetDefault("caiyun.latitude", 39.9042)
	v.SetDefault("caiyun.longitude", 116.4074)
	v.SetDefault("caiyun.timeout", 30)

	// LLM defaults
	v.SetDefault("llm.provider", "openai")
	v.SetDefault("llm.model", "gpt-4")
	v.SetDefault("llm.timeout", 60)

	// Schedule defaults
	v.SetDefault("schedule.check_weather_cron", "0 7,12,18 * * *")
	v.SetDefault("schedule.timezone", "Asia/Shanghai")

	// Detection defaults
	v.SetDefault("detection.rain_probability_threshold", 0.3)
	v.SetDefault("detection.temperature_high_threshold", 35.0)
	v.SetDefault("detection.temperature_low_threshold", 0.0)
	v.SetDefault("detection.aqi_threshold", 150)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.pretty_print", true)
}

// expandEnvVars expands environment variables in string config values
func expandEnvVars(v *viper.Viper) {
	// Get all settings as a map
	settings := v.AllSettings()

	// Recursively expand env vars in string values
	expandMap(settings)

	// Merge back into viper
	for key, value := range flattenMap("", settings) {
		v.Set(key, value)
	}
}

// expandMap recursively expands environment variables in map values
func expandMap(m map[string]interface{}) {
	for key, value := range m {
		switch v := value.(type) {
		case string:
			m[key] = os.ExpandEnv(v)
		case map[string]interface{}:
			expandMap(v)
		case []interface{}:
			for i, item := range v {
				if str, ok := item.(string); ok {
					v[i] = os.ExpandEnv(str)
				}
			}
		}
	}
}

// flattenMap flattens a nested map into dot-notation keys
func flattenMap(prefix string, m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range m {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if nestedMap, ok := value.(map[string]interface{}); ok {
			for k, v := range flattenMap(fullKey, nestedMap) {
				result[k] = v
			}
		} else {
			result[fullKey] = value
		}
	}
	return result
}

// validate checks if the configuration is valid
func validate(cfg *Config) error {
	// Validate Telegram config
	if cfg.Telegram.BotToken == "" {
		return fmt.Errorf("telegram.bot_token is required")
	}
	if cfg.Telegram.Password == "" {
		return fmt.Errorf("telegram.password is required")
	}

	// Validate Caiyun config
	if cfg.Caiyun.APIToken == "" {
		return fmt.Errorf("caiyun.api_token is required")
	}
	if cfg.Caiyun.Latitude < -90 || cfg.Caiyun.Latitude > 90 {
		return fmt.Errorf("caiyun.latitude must be between -90 and 90")
	}
	if cfg.Caiyun.Longitude < -180 || cfg.Caiyun.Longitude > 180 {
		return fmt.Errorf("caiyun.longitude must be between -180 and 180")
	}

	// Validate LLM config
	if cfg.LLM.APIKey == "" {
		return fmt.Errorf("llm.api_key is required")
	}
	if cfg.LLM.Model == "" {
		return fmt.Errorf("llm.model is required")
	}

	// Validate Schedule config
	if cfg.Schedule.CheckWeatherCron == "" {
		return fmt.Errorf("schedule.check_weather_cron is required")
	}

	// Validate Detection config
	if cfg.Detection.RainProbabilityThreshold < 0 || cfg.Detection.RainProbabilityThreshold > 1 {
		return fmt.Errorf("detection.rain_probability_threshold must be between 0 and 1")
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

	return nil
}
