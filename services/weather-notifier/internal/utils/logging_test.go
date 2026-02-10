package utils

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestInitLogger(t *testing.T) {
	var buf bytes.Buffer

	cfg := LogConfig{
		Level:       "debug",
		PrettyPrint: false,
		Output:      &buf,
	}

	logger := InitLogger(cfg)
	if logger == nil {
		t.Fatal("Expected logger, got nil")
	}

	// Verify global log level is set
	if zerolog.GlobalLevel() != zerolog.DebugLevel {
		t.Errorf("Expected global level DEBUG, got %v", zerolog.GlobalLevel())
	}
}

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer

	cfg := LogConfig{
		Level:       "info",
		PrettyPrint: false,
		Output:      &buf,
	}

	logger := NewLogger(cfg)
	if logger == nil {
		t.Fatal("Expected logger, got nil")
	}

	// Test logging
	logger.Info().Msg("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected log output to contain 'test message', got: %s", output)
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"invalid", zerolog.InfoLevel}, // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := parseLogLevel(tt.input)
			if level != tt.expected {
				t.Errorf("Expected level %v, got %v", tt.expected, level)
			}
		})
	}
}

func TestWithComponent(t *testing.T) {
	var buf bytes.Buffer

	cfg := LogConfig{
		Level:       "info",
		PrettyPrint: false,
		Output:      &buf,
	}

	logger := NewLogger(cfg)
	componentLogger := logger.WithComponent("test-component")

	componentLogger.Info().Msg("test message")

	output := buf.String()
	if !strings.Contains(output, "test-component") {
		t.Errorf("Expected log output to contain 'test-component', got: %s", output)
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer

	cfg := LogConfig{
		Level:       "info",
		PrettyPrint: false,
		Output:      &buf,
	}

	logger := NewLogger(cfg)
	fieldLogger := logger.WithFields(map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})

	fieldLogger.Info().Msg("test message")

	output := buf.String()
	if !strings.Contains(output, "key1") || !strings.Contains(output, "value1") {
		t.Errorf("Expected log output to contain fields, got: %s", output)
	}
}
