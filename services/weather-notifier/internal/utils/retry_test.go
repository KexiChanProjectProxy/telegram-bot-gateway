package utils

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry_Success(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
	}

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := Retry(ctx, cfg, fn)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
	}

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		return errors.New("persistent error")
	}

	err := Retry(ctx, cfg, fn)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if attempts != 3 { // MaxRetries + 1 (initial attempt)
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		Multiplier:     2.0,
	}

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		if attempts == 2 {
			cancel() // Cancel context after second attempt
		}
		return errors.New("error")
	}

	err := Retry(ctx, cfg, fn)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

func TestRetryWithResult_Success(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultRetryConfig()
	cfg.InitialBackoff = 10 * time.Millisecond

	attempts := 0
	fn := func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 2 {
			return "", errors.New("temporary error")
		}
		return "success", nil
	}

	result, err := RetryWithResult(ctx, cfg, fn)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected result 'success', got '%s'", result)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryIf_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	cfg := RetryableConfig{
		RetryConfig: RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			Multiplier:     2.0,
		},
		IsRetryable: func(err error) bool {
			return err.Error() != "non-retryable"
		},
	}

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		return errors.New("non-retryable")
	}

	err := RetryIf(ctx, cfg, fn)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt (non-retryable), got %d", attempts)
	}
}

func TestCalculateBackoff(t *testing.T) {
	cfg := RetryConfig{
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     10 * time.Second,
		Multiplier:     2.0,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 10 * time.Second}, // capped at MaxBackoff
		{5, 10 * time.Second}, // capped at MaxBackoff
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			backoff := calculateBackoff(tt.attempt, cfg)
			if backoff != tt.expected {
				t.Errorf("Attempt %d: expected backoff %v, got %v", tt.attempt, tt.expected, backoff)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", cfg.MaxRetries)
	}

	if cfg.InitialBackoff != 1*time.Second {
		t.Errorf("Expected InitialBackoff 1s, got %v", cfg.InitialBackoff)
	}

	if cfg.MaxBackoff != 30*time.Second {
		t.Errorf("Expected MaxBackoff 30s, got %v", cfg.MaxBackoff)
	}

	if cfg.Multiplier != 2.0 {
		t.Errorf("Expected Multiplier 2.0, got %f", cfg.Multiplier)
	}
}

func TestSimpleRetry(t *testing.T) {
	ctx := context.Background()

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := SimpleRetry(ctx, fn)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetry_OnRetryCallback(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
	}

	retryCallbacks := 0
	cfg.OnRetry = func(attempt int, err error) {
		retryCallbacks++
	}

	attempts := 0
	fn := func(ctx context.Context) error {
		attempts++
		return errors.New("error")
	}

	Retry(ctx, cfg, fn)

	// Should have 2 retry callbacks (for attempts 2 and 3)
	if retryCallbacks != 2 {
		t.Errorf("Expected 2 retry callbacks, got %d", retryCallbacks)
	}
}
