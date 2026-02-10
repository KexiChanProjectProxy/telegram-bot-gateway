package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// MessageBroker handles pub/sub operations for real-time message distribution
type MessageBroker struct {
	client *redis.Client
}

// NewMessageBroker creates a new message broker
func NewMessageBroker(client *redis.Client) *MessageBroker {
	return &MessageBroker{
		client: client,
	}
}

// MessageEvent represents a message event to be published
type MessageEvent struct {
	Type           string                 `json:"type"` // "new_message", "edited_message", "deleted_message"
	ChatID         uint                   `json:"chat_id"`
	MessageID      uint                   `json:"message_id"`
	TelegramID     int64                  `json:"telegram_id"`
	BotID          uint                   `json:"bot_id"`
	Direction      string                 `json:"direction"` // "incoming", "outgoing"
	Text           string                 `json:"text,omitempty"`
	FromUsername   string                 `json:"from_username,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Payload        map[string]interface{} `json:"payload,omitempty"` // Full message data
}

// PublishMessage publishes a message event to the appropriate channels
func (b *MessageBroker) PublishMessage(ctx context.Context, event *MessageEvent) error {
	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to multiple channels for different subscribers
	channels := []string{
		fmt.Sprintf("chat:%d", event.ChatID),           // Chat-specific channel
		fmt.Sprintf("bot:%d", event.BotID),             // Bot-specific channel
		"messages:all",                                  // Global channel
	}

	// Publish to all channels
	for _, channel := range channels {
		if err := b.client.Publish(ctx, channel, data).Err(); err != nil {
			log.Printf("Failed to publish to channel %s: %v", channel, err)
			// Don't return error, continue publishing to other channels
		}
	}

	return nil
}

// Subscribe subscribes to a specific channel and returns a channel for receiving messages
func (b *MessageBroker) Subscribe(ctx context.Context, channels ...string) (<-chan *MessageEvent, error) {
	pubsub := b.client.Subscribe(ctx, channels...)

	// Wait for confirmation that subscription is created
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	// Create output channel
	eventChan := make(chan *MessageEvent, 100) // Buffered to prevent blocking

	// Start goroutine to receive messages
	go func() {
		defer close(eventChan)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				var event MessageEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					continue
				}

				// Try to send to output channel, skip if full
				select {
				case eventChan <- &event:
				default:
					log.Printf("Event channel full, dropping message for chat %d", event.ChatID)
				}
			}
		}
	}()

	return eventChan, nil
}

// SubscribeToChat subscribes to messages for a specific chat
func (b *MessageBroker) SubscribeToChat(ctx context.Context, chatID uint) (<-chan *MessageEvent, error) {
	channel := fmt.Sprintf("chat:%d", chatID)
	return b.Subscribe(ctx, channel)
}

// SubscribeToBot subscribes to messages for a specific bot
func (b *MessageBroker) SubscribeToBot(ctx context.Context, botID uint) (<-chan *MessageEvent, error) {
	channel := fmt.Sprintf("bot:%d", botID)
	return b.Subscribe(ctx, channel)
}

// SubscribeToAll subscribes to all messages
func (b *MessageBroker) SubscribeToAll(ctx context.Context) (<-chan *MessageEvent, error) {
	return b.Subscribe(ctx, "messages:all")
}

// SubscribeToMultipleChats subscribes to multiple chat channels
func (b *MessageBroker) SubscribeToMultipleChats(ctx context.Context, chatIDs []uint) (<-chan *MessageEvent, error) {
	channels := make([]string, len(chatIDs))
	for i, chatID := range chatIDs {
		channels[i] = fmt.Sprintf("chat:%d", chatID)
	}
	return b.Subscribe(ctx, channels...)
}

// QueueWebhookDelivery adds a webhook delivery job to the queue
func (b *MessageBroker) QueueWebhookDelivery(ctx context.Context, deliveryID uint) error {
	// Add to Redis list for webhook workers to process
	return b.client.RPush(ctx, "webhook_deliveries", deliveryID).Err()
}

// DequeueWebhookDelivery retrieves a webhook delivery job from the queue (blocking)
func (b *MessageBroker) DequeueWebhookDelivery(ctx context.Context, timeout time.Duration) (uint, error) {
	result, err := b.client.BLPop(ctx, timeout, "webhook_deliveries").Result()
	if err != nil {
		return 0, err
	}

	if len(result) < 2 {
		return 0, fmt.Errorf("invalid result from BLPop")
	}

	var deliveryID uint
	if _, err := fmt.Sscanf(result[1], "%d", &deliveryID); err != nil {
		return 0, fmt.Errorf("failed to parse delivery ID: %w", err)
	}

	return deliveryID, nil
}

// GetPendingWebhookDeliveryCount returns the number of pending webhook deliveries
func (b *MessageBroker) GetPendingWebhookDeliveryCount(ctx context.Context) (int64, error) {
	return b.client.LLen(ctx, "webhook_deliveries").Result()
}

// PublishWebhookDeliveryResult publishes the result of a webhook delivery
func (b *MessageBroker) PublishWebhookDeliveryResult(ctx context.Context, deliveryID uint, success bool, errorMsg string) error {
	result := map[string]interface{}{
		"delivery_id": deliveryID,
		"success":     success,
		"error":       errorMsg,
		"timestamp":   time.Now().Unix(),
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return b.client.Publish(ctx, "webhook_delivery_results", data).Err()
}
