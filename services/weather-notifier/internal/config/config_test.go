package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary test config file
	testConfig := `{
  "telegram": {
    "api_key": "tgw_test_key_123"
  },
  "caiyun": {
    "api_token": "test-caiyun-token"
  },
  "llm": {
    "api_key": "test-llm-key",
    "model": "gpt-4"
  },
  "schedule": {
    "timezone": "Asia/Shanghai",
    "morning_cron": "0 8 * * *",
    "evening_cron": "30 23 * * *",
    "poll_cron": "*/15 * * * *"
  },
  "detection": {
    "temperature_delta": 3.0,
    "wind_speed_delta": 5.0,
    "visibility_delta": 5.0,
    "aqi_cn_delta": 50,
    "aqi_usa_delta": 50
  },
  "logging": {
    "level": "info",
    "pretty_print": true
  },
  "chats": [
    {
      "chat_id": 123456,
      "name": "Test Chat",
      "locations": [
        {
          "name": "Beijing",
          "latitude": 39.9042,
          "longitude": 116.4074
        }
      ]
    }
  ]
}`

	tmpFile, err := os.CreateTemp("", "config-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testConfig); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Test loading config
	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if cfg.Telegram.APIKey != "tgw_test_key_123" {
		t.Errorf("Expected APIKey 'tgw_test_key_123', got '%s'", cfg.Telegram.APIKey)
	}

	if cfg.Caiyun.APIToken != "test-caiyun-token" {
		t.Errorf("Expected APIToken 'test-caiyun-token', got '%s'", cfg.Caiyun.APIToken)
	}

	if cfg.LLM.APIKey != "test-llm-key" {
		t.Errorf("Expected APIKey 'test-llm-key', got '%s'", cfg.LLM.APIKey)
	}

	if len(cfg.Chats) != 1 {
		t.Errorf("Expected 1 chat, got %d", len(cfg.Chats))
	}

	if cfg.Chats[0].ChatID != 123456 {
		t.Errorf("Expected chat ID 123456, got %d", cfg.Chats[0].ChatID)
	}

	if len(cfg.Chats[0].Locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(cfg.Chats[0].Locations))
	}

	// Verify defaults
	if cfg.Caiyun.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", cfg.Caiyun.Timeout)
	}

	if cfg.LLM.MaxTokens != 500 {
		t.Errorf("Expected default max tokens 500, got %d", cfg.LLM.MaxTokens)
	}

	if cfg.LLM.Temperature != 0.7 {
		t.Errorf("Expected default temperature 0.7, got %f", cfg.LLM.Temperature)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Create a temporary test config file with env var placeholders
	testConfig := `{
  "telegram": {
    "api_key": "${TEST_API_KEY}"
  },
  "caiyun": {
    "api_token": "${TEST_CAIYUN_TOKEN}"
  },
  "llm": {
    "api_key": "${TEST_LLM_KEY}",
    "model": "gpt-4"
  },
  "chats": [
    {
      "chat_id": 123456,
      "name": "Test",
      "locations": [
        {
          "name": "Beijing",
          "latitude": 39.9042,
          "longitude": 116.4074
        }
      ]
    }
  ]
}`

	// Set environment variables
	os.Setenv("TEST_API_KEY", "tgw_env_key")
	os.Setenv("TEST_CAIYUN_TOKEN", "env-caiyun-token")
	os.Setenv("TEST_LLM_KEY", "env-llm-key")
	defer func() {
		os.Unsetenv("TEST_API_KEY")
		os.Unsetenv("TEST_CAIYUN_TOKEN")
		os.Unsetenv("TEST_LLM_KEY")
	}()

	tmpFile, err := os.CreateTemp("", "config-env-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testConfig); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Test loading config
	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify env vars were expanded
	if cfg.Telegram.APIKey != "tgw_env_key" {
		t.Errorf("Expected APIKey 'tgw_env_key', got '%s'", cfg.Telegram.APIKey)
	}

	if cfg.Caiyun.APIToken != "env-caiyun-token" {
		t.Errorf("Expected APIToken 'env-caiyun-token', got '%s'", cfg.Caiyun.APIToken)
	}

	if cfg.LLM.APIKey != "env-llm-key" {
		t.Errorf("Expected APIKey 'env-llm-key', got '%s'", cfg.LLM.APIKey)
	}
}

func TestResolveLLM(t *testing.T) {
	global := LLMConfig{
		Provider:    "openai",
		APIKey:      "global-key",
		Model:       "gpt-4",
		BaseURL:     "https://api.openai.com/v1",
		MaxTokens:   500,
		Temperature: 0.7,
	}

	tests := []struct {
		name     string
		chat     ChatConfig
		expected LLMConfig
	}{
		{
			name: "no overrides",
			chat: ChatConfig{
				LLM: nil,
			},
			expected: global,
		},
		{
			name: "override model only",
			chat: ChatConfig{
				LLM: &LLMOverride{
					Model: strPtr("gpt-4o"),
				},
			},
			expected: LLMConfig{
				Provider:    "openai",
				APIKey:      "global-key",
				Model:       "gpt-4o",
				BaseURL:     "https://api.openai.com/v1",
				MaxTokens:   500,
				Temperature: 0.7,
			},
		},
		{
			name: "override temperature to 0",
			chat: ChatConfig{
				LLM: &LLMOverride{
					Temperature: float64Ptr(0.0),
				},
			},
			expected: LLMConfig{
				Provider:    "openai",
				APIKey:      "global-key",
				Model:       "gpt-4",
				BaseURL:     "https://api.openai.com/v1",
				MaxTokens:   500,
				Temperature: 0.0,
			},
		},
		{
			name: "override multiple fields",
			chat: ChatConfig{
				LLM: &LLMOverride{
					Model:       strPtr("gpt-4o"),
					MaxTokens:   intPtr(1000),
					Temperature: float64Ptr(0.5),
				},
			},
			expected: LLMConfig{
				Provider:    "openai",
				APIKey:      "global-key",
				Model:       "gpt-4o",
				BaseURL:     "https://api.openai.com/v1",
				MaxTokens:   1000,
				Temperature: 0.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := tt.chat.ResolveLLM(global)

			if resolved.Provider != tt.expected.Provider {
				t.Errorf("Provider: expected %s, got %s", tt.expected.Provider, resolved.Provider)
			}
			if resolved.APIKey != tt.expected.APIKey {
				t.Errorf("APIKey: expected %s, got %s", tt.expected.APIKey, resolved.APIKey)
			}
			if resolved.Model != tt.expected.Model {
				t.Errorf("Model: expected %s, got %s", tt.expected.Model, resolved.Model)
			}
			if resolved.BaseURL != tt.expected.BaseURL {
				t.Errorf("BaseURL: expected %s, got %s", tt.expected.BaseURL, resolved.BaseURL)
			}
			if resolved.MaxTokens != tt.expected.MaxTokens {
				t.Errorf("MaxTokens: expected %d, got %d", tt.expected.MaxTokens, resolved.MaxTokens)
			}
			if resolved.Temperature != tt.expected.Temperature {
				t.Errorf("Temperature: expected %f, got %f", tt.expected.Temperature, resolved.Temperature)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			cfg: &Config{
				Telegram: TelegramConfig{
					APIKey: "tgw_test_key",
				},
				Caiyun: CaiyunConfig{
					APIToken: "token",
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Schedule: ScheduleConfig{
					MorningCron: "0 8 * * *",
					EveningCron: "30 23 * * *",
					PollCron:    "*/15 * * * *",
				},
				Logging: LoggingConfig{
					Level: "info",
				},
				Chats: []ChatConfig{
					{
						ChatID: 123456,
						Name:   "Test",
						Locations: []LocationConfig{
							{
								Name:      "Beijing",
								Latitude:  39.9042,
								Longitude: 116.4074,
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing api key",
			cfg: &Config{
				Telegram: TelegramConfig{},
				Caiyun: CaiyunConfig{
					APIToken: "token",
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Chats: []ChatConfig{
					{
						ChatID: 123456,
						Locations: []LocationConfig{
							{Name: "Test", Latitude: 39.9, Longitude: 116.4},
						},
					},
				},
			},
			expectError: true,
		},
		{
			name: "no chats configured",
			cfg: &Config{
				Telegram: TelegramConfig{
					APIKey: "tgw_test_key",
				},
				Caiyun: CaiyunConfig{
					APIToken: "token",
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Chats: []ChatConfig{},
			},
			expectError: true,
		},
		{
			name: "duplicate chat IDs",
			cfg: &Config{
				Telegram: TelegramConfig{
					APIKey: "tgw_test_key",
				},
				Caiyun: CaiyunConfig{
					APIToken: "token",
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Chats: []ChatConfig{
					{
						ChatID: 123456,
						Locations: []LocationConfig{
							{Name: "Test", Latitude: 39.9, Longitude: 116.4},
						},
					},
					{
						ChatID: 123456,
						Locations: []LocationConfig{
							{Name: "Test2", Latitude: 31.2, Longitude: 121.5},
						},
					},
				},
			},
			expectError: true,
		},
		{
			name: "duplicate location names",
			cfg: &Config{
				Telegram: TelegramConfig{
					APIKey: "tgw_test_key",
				},
				Caiyun: CaiyunConfig{
					APIToken: "token",
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Chats: []ChatConfig{
					{
						ChatID: 123456,
						Locations: []LocationConfig{
							{Name: "Test", Latitude: 39.9, Longitude: 116.4},
							{Name: "Test", Latitude: 31.2, Longitude: 121.5},
						},
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid latitude",
			cfg: &Config{
				Telegram: TelegramConfig{
					APIKey: "tgw_test_key",
				},
				Caiyun: CaiyunConfig{
					APIToken: "token",
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Chats: []ChatConfig{
					{
						ChatID: 123456,
						Locations: []LocationConfig{
							{Name: "Test", Latitude: 200.0, Longitude: 116.4},
						},
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid log level",
			cfg: &Config{
				Telegram: TelegramConfig{
					APIKey: "tgw_test_key",
				},
				Caiyun: CaiyunConfig{
					APIToken: "token",
				},
				LLM: LLMConfig{
					APIKey: "key",
					Model:  "gpt-4",
				},
				Logging: LoggingConfig{
					Level: "invalid",
				},
				Chats: []ChatConfig{
					{
						ChatID: 123456,
						Locations: []LocationConfig{
							{Name: "Test", Latitude: 39.9, Longitude: 116.4},
						},
					},
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

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple", "Simple"},
		{"With Spaces", "With_Spaces"},
		{"with-dashes", "with-dashes"},
		{"with_underscores", "with_underscores"},
		{"Special!@#$%Chars", "Special_Chars"},
		{"北京", "unnamed"},
		{"Mixed123ABC", "Mixed123ABC"},
		{"   leading", "leading"},
		{"trailing   ", "trailing"},
		{"", "unnamed"},
		{"___", "unnamed"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFileName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper functions for pointer creation
func strPtr(s string) *string       { return &s }
func intPtr(i int) *int             { return &i }
func float64Ptr(f float64) *float64 { return &f }
