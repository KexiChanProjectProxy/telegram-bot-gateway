package config

import (
	"os"
	"testing"
)

func TestLoadWithEnvVars(t *testing.T) {
	// Note: This test requires actual environment variables to be set
	// because Viper's environment variable binding happens at initialization time.
	// For CI/CD, set these env vars before running tests:
	// WNB_TELEGRAM_BOT_TOKEN, WNB_TELEGRAM_PASSWORD, WNB_CAIYUN_API_TOKEN,
	// WNB_LLM_API_KEY, WNB_LLM_MODEL

	// Check if required env vars are set
	requiredEnvVars := []string{
		"WNB_TELEGRAM_BOT_TOKEN",
		"WNB_TELEGRAM_PASSWORD",
		"WNB_CAIYUN_API_TOKEN",
		"WNB_LLM_API_KEY",
	}

	allSet := true
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			allSet = false
			break
		}
	}

	if !allSet {
		t.Skip("Skipping test: required environment variables not set")
	}

	// Test loading config
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values are loaded from env vars
	if cfg.Telegram.BotToken == "" {
		t.Error("Expected BotToken to be loaded from env var")
	}

	if cfg.Telegram.Password == "" {
		t.Error("Expected Password to be loaded from env var")
	}

	if cfg.Caiyun.APIToken == "" {
		t.Error("Expected APIToken to be loaded from env var")
	}

	if cfg.LLM.APIKey == "" {
		t.Error("Expected APIKey to be loaded from env var")
	}

	// Verify defaults
	if cfg.Caiyun.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", cfg.Caiyun.Timeout)
	}

	if cfg.Detection.RainProbabilityThreshold != 0.3 {
		t.Errorf("Expected default rain threshold 0.3, got %f", cfg.Detection.RainProbabilityThreshold)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		expectError bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: "token",
					Password: "password",
				},
				Caiyun: CaiyunConfig{
					APIToken:  "token",
					Latitude:  39.9042,
					Longitude: 116.4074,
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Schedule: ScheduleConfig{
					CheckWeatherCron: "0 7 * * *",
				},
				Detection: DetectionConfig{
					RainProbabilityThreshold: 0.5,
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			expectError: false,
		},
		{
			name: "missing bot token",
			cfg: &Config{
				Telegram: TelegramConfig{
					Password: "password",
				},
				Caiyun: CaiyunConfig{
					APIToken: "token",
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Schedule: ScheduleConfig{
					CheckWeatherCron: "0 7 * * *",
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			expectError: true,
		},
		{
			name: "invalid latitude",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: "token",
					Password: "password",
				},
				Caiyun: CaiyunConfig{
					APIToken:  "token",
					Latitude:  200.0,
					Longitude: 116.4074,
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Schedule: ScheduleConfig{
					CheckWeatherCron: "0 7 * * *",
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			expectError: true,
		},
		{
			name: "invalid log level",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: "token",
					Password: "password",
				},
				Caiyun: CaiyunConfig{
					APIToken:  "token",
					Latitude:  39.9042,
					Longitude: 116.4074,
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Schedule: ScheduleConfig{
					CheckWeatherCron: "0 7 * * *",
				},
				Logging: LoggingConfig{
					Level: "invalid",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.cfg)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
