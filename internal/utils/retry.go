package utils

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries     int           // Maximum number of retry attempts
	InitialBackoff time.Duration // Initial backoff duration
	MaxBackoff     time.Duration // Maximum backoff duration
	Multiplier     float64       // Backoff multiplier for exponential backoff
	OnRetry        func(attempt int, err error) // Optional callback on retry
}

// DefaultRetryConfig returns a default retry configuration
// - MaxRetries: 3
// - InitialBackoff: 1 second
// - MaxBackoff: 30 seconds
// - Multiplier: 2.0 (exponential)
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// RetryFunc is a function type that can be retried
type RetryFunc func(ctx context.Context) error

// Retry executes the provided function with exponential backoff retry logic
// It respects context cancellation and returns the last error if all retries fail
func Retry(ctx context.Context, cfg RetryConfig, fn RetryFunc) error {
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context cancellation before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry if this was the last attempt
		if attempt == cfg.MaxRetries {
			break
		}

		// Call retry callback if provided
		if cfg.OnRetry != nil {
			cfg.OnRetry(attempt+1, err)
		}

		// Calculate backoff duration
		backoff := calculateBackoff(attempt, cfg)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// Continue to next retry
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}

// RetryWithResult executes a function that returns a result and an error, with retry logic
type RetryFuncWithResult[T any] func(ctx context.Context) (T, error)

// RetryWithResult executes the provided function with exponential backoff retry logic
// It returns the result and error from the last attempt
func RetryWithResult[T any](ctx context.Context, cfg RetryConfig, fn RetryFuncWithResult[T]) (T, error) {
	var lastErr error
	var result T

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context cancellation before attempting
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// Execute the function
		res, err := fn(ctx)
		if err == nil {
			return res, nil // Success
		}

		result = res
		lastErr = err

		// Don't retry if this was the last attempt
		if attempt == cfg.MaxRetries {
			break
		}

		// Call retry callback if provided
		if cfg.OnRetry != nil {
			cfg.OnRetry(attempt+1, err)
		}

		// Calculate backoff duration
		backoff := calculateBackoff(attempt, cfg)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(backoff):
			// Continue to next retry
		}
	}

	return result, fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}

// calculateBackoff calculates the backoff duration for a given attempt
// Uses exponential backoff: initialBackoff * (multiplier ^ attempt)
// Capped at maxBackoff
func calculateBackoff(attempt int, cfg RetryConfig) time.Duration {
	// Calculate exponential backoff
	backoff := float64(cfg.InitialBackoff) * math.Pow(cfg.Multiplier, float64(attempt))

	// Cap at max backoff
	if backoff > float64(cfg.MaxBackoff) {
		backoff = float64(cfg.MaxBackoff)
	}

	return time.Duration(backoff)
}

// IsRetryable checks if an error is retryable
// Override this function to customize retry logic for specific error types
type IsRetryableFunc func(error) bool

// RetryableConfig extends RetryConfig with custom retry logic
type RetryableConfig struct {
	RetryConfig
	IsRetryable IsRetryableFunc // Function to determine if error is retryable
}

// RetryIf executes the function with retry logic, but only retries if the error is retryable
func RetryIf(ctx context.Context, cfg RetryableConfig, fn RetryFunc) error {
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context cancellation before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if cfg.IsRetryable != nil && !cfg.IsRetryable(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't retry if this was the last attempt
		if attempt == cfg.MaxRetries {
			break
		}

		// Call retry callback if provided
		if cfg.OnRetry != nil {
			cfg.OnRetry(attempt+1, err)
		}

		// Calculate backoff duration
		backoff := calculateBackoff(attempt, cfg.RetryConfig)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// Continue to next retry
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}

// SimpleRetry is a convenience function for simple retry scenarios
// Uses default configuration with 3 retries and exponential backoff
func SimpleRetry(ctx context.Context, fn RetryFunc) error {
	return Retry(ctx, DefaultRetryConfig(), fn)
}

// SimpleRetryWithResult is a convenience function for simple retry scenarios with result
func SimpleRetryWithResult[T any](ctx context.Context, fn RetryFuncWithResult[T]) (T, error) {
	return RetryWithResult(ctx, DefaultRetryConfig(), fn)
}
