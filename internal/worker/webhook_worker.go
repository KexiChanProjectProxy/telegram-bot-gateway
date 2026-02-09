package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
	"github.com/kexi/telegram-bot-gateway/internal/pubsub"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// WebhookWorker processes webhook deliveries
type WebhookWorker struct {
	workerID          int
	messageBroker     *pubsub.MessageBroker
	webhookService    *service.WebhookService
	messageService    *service.MessageService
	deliveryRepo      repository.WebhookDeliveryRepository
	circuitBreakers   map[string]*CircuitBreaker
	circuitBreakersMu sync.RWMutex
	httpClient        *http.Client
	maxRetries        int
}

// NewWebhookWorker creates a new webhook worker
func NewWebhookWorker(
	workerID int,
	messageBroker *pubsub.MessageBroker,
	webhookService *service.WebhookService,
	messageService *service.MessageService,
	deliveryRepo repository.WebhookDeliveryRepository,
	maxRetries int,
) *WebhookWorker {
	return &WebhookWorker{
		workerID:        workerID,
		messageBroker:   messageBroker,
		webhookService:  webhookService,
		messageService:  messageService,
		deliveryRepo:    deliveryRepo,
		circuitBreakers: make(map[string]*CircuitBreaker),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: maxRetries,
	}
}

// Start starts the webhook worker
func (w *WebhookWorker) Start(ctx context.Context) {
	log.Printf("Worker #%d: Starting", w.workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker #%d: Shutting down", w.workerID)
			return
		default:
			// Dequeue delivery job (blocking with 5 second timeout)
			deliveryID, err := w.messageBroker.DequeueWebhookDelivery(ctx, 5*time.Second)
			if err != nil {
				// Timeout or error, continue
				continue
			}

			// Process the delivery
			if err := w.processDelivery(ctx, deliveryID); err != nil {
				log.Printf("Worker #%d: Failed to process delivery %d: %v", w.workerID, deliveryID, err)
			}
		}
	}
}

// processDelivery processes a single webhook delivery
func (w *WebhookWorker) processDelivery(ctx context.Context, deliveryID uint) error {
	// Get delivery details
	delivery, err := w.deliveryRepo.GetByID(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	// Check if we should retry
	if delivery.AttemptCount >= w.maxRetries {
		delivery.Status = "failed"
		w.deliveryRepo.Update(ctx, delivery)
		w.messageBroker.PublishWebhookDeliveryResult(ctx, deliveryID, false, "Max retries exceeded")
		return fmt.Errorf("max retries exceeded")
	}

	// Get circuit breaker for this URL
	cb := w.getCircuitBreaker(delivery.Webhook.URL)

	// Check if circuit is open
	if !cb.CanAttempt() {
		// Circuit is open, requeue for later
		nextRetry := time.Now().Add(w.getRetryDelay(delivery.AttemptCount))
		delivery.NextRetryAt = &nextRetry
		w.deliveryRepo.Update(ctx, delivery)
		w.messageBroker.QueueWebhookDelivery(ctx, deliveryID)
		return fmt.Errorf("circuit breaker open for %s", delivery.Webhook.URL)
	}

	// Attempt delivery
	success, err := w.attemptDelivery(ctx, delivery)

	// Update circuit breaker
	if success {
		cb.RecordSuccess()
	} else {
		cb.RecordFailure()
	}

	// Update delivery status
	delivery.AttemptCount++
	if success {
		delivery.Status = "delivered"
		now := time.Now()
		delivery.DeliveredAt = &now
		w.deliveryRepo.Update(ctx, delivery)
		w.messageBroker.PublishWebhookDeliveryResult(ctx, deliveryID, true, "")
		log.Printf("Worker #%d: Successfully delivered webhook %d", w.workerID, deliveryID)
	} else {
		delivery.Status = "pending"
		if err != nil {
			delivery.LastError = err.Error()
		}

		// Calculate next retry time with exponential backoff
		nextRetry := time.Now().Add(w.getRetryDelay(delivery.AttemptCount))
		delivery.NextRetryAt = &nextRetry

		w.deliveryRepo.Update(ctx, delivery)

		// Requeue for retry
		w.messageBroker.QueueWebhookDelivery(ctx, deliveryID)

		log.Printf("Worker #%d: Failed delivery %d, retry at %s", w.workerID, deliveryID, nextRetry.Format(time.RFC3339))
	}

	return nil
}

// attemptDelivery attempts to deliver a webhook
func (w *WebhookWorker) attemptDelivery(ctx context.Context, delivery *domain.WebhookDelivery) (bool, error) {
	// Get message details
	message, err := w.messageService.GetMessage(ctx, delivery.MessageID)
	if err != nil {
		return false, fmt.Errorf("failed to get message: %w", err)
	}

	// Build payload
	payload := map[string]interface{}{
		"event":      "message",
		"message_id": message.ID,
		"chat_id":    message.ChatID,
		"telegram_id": message.TelegramID,
		"text":       message.Text,
		"from_username": message.FromUsername,
		"from_first_name": message.FromFirstName,
		"direction":  message.Direction,
		"message_type": message.MessageType,
		"sent_at":    message.SentAt,
		"timestamp":  time.Now().Unix(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", delivery.Webhook.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TelegramBotGateway/1.0")

	// Sign payload with HMAC
	signature := w.webhookService.SignPayload(delivery.Webhook.Secret, payloadBytes)
	req.Header.Set("X-Webhook-Signature", signature)

	// Send request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (limit to 1MB)
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, nil
	}

	return false, fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
}

// getRetryDelay calculates the retry delay with exponential backoff
func (w *WebhookWorker) getRetryDelay(attemptCount int) time.Duration {
	// Exponential backoff: 1s, 10s, 1m, 5m, 30m
	delays := []time.Duration{
		1 * time.Second,
		10 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
		30 * time.Minute,
	}

	if attemptCount >= len(delays) {
		return delays[len(delays)-1]
	}

	return delays[attemptCount]
}

// getCircuitBreaker gets or creates a circuit breaker for a URL
func (w *WebhookWorker) getCircuitBreaker(url string) *CircuitBreaker {
	w.circuitBreakersMu.RLock()
	cb, exists := w.circuitBreakers[url]
	w.circuitBreakersMu.RUnlock()

	if exists {
		return cb
	}

	w.circuitBreakersMu.Lock()
	defer w.circuitBreakersMu.Unlock()

	// Check again after acquiring write lock
	if cb, exists := w.circuitBreakers[url]; exists {
		return cb
	}

	// Create new circuit breaker
	cb = NewCircuitBreaker(5, 1*time.Minute) // 5 failures, 1 minute timeout
	w.circuitBreakers[url] = cb

	return cb
}

// CircuitBreaker implements a simple circuit breaker pattern
type CircuitBreaker struct {
	failureThreshold int
	timeout          time.Duration
	failures         int
	lastFailureTime  time.Time
	state            string // "closed", "open", "half-open"
	mu               sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		timeout:          timeout,
		state:            "closed",
	}
}

// CanAttempt checks if a request can be attempted
func (cb *CircuitBreaker) CanAttempt() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == "closed" {
		return true
	}

	if cb.state == "open" {
		// Check if timeout has elapsed
		if time.Since(cb.lastFailureTime) > cb.timeout {
			return true // Allow one attempt (half-open)
		}
		return false
	}

	// Half-open state
	return true
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = "closed"
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.failures >= cb.failureThreshold {
		cb.state = "open"
	} else if cb.state == "half-open" {
		cb.state = "open"
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
