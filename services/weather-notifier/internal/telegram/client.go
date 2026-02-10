package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// Client is the Telegram Bot Gateway client
type Client struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
	logger     zerolog.Logger
}

// SendMessageRequest represents the message send request payload
type SendMessageRequest struct {
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// SendMessageResponse represents the message send response
type SendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
	MessageID   int64  `json:"message_id,omitempty"`
}

// NewClient creates a new Telegram Gateway client with API key authentication
func NewClient(apiKey, apiURL string, logger zerolog.Logger) *Client {
	return &Client{
		apiKey: apiKey,
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger.With().Str("component", "telegram_client").Logger(),
	}
}

// SendMessage sends a message to the specified chat using API key authentication
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string, parseMode string) error {
	// Prepare request body
	reqBody := SendMessageRequest{
		Text:      text,
		ParseMode: parseMode,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal send message request: %w", err)
	}

	// Create HTTP request - use chat-specific endpoint
	url := fmt.Sprintf("%s/api/v1/chats/%d/messages", c.apiURL, chatID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create send message request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send message request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read send message response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send message failed with status %d: %s", resp.StatusCode, string(body))
	}

	var msgResp SendMessageResponse
	if err := json.Unmarshal(body, &msgResp); err != nil {
		return fmt.Errorf("failed to unmarshal send message response: %w", err)
	}

	if !msgResp.OK {
		return fmt.Errorf("send message failed: %s", msgResp.Description)
	}

	c.logger.Info().
		Int64("chat_id", chatID).
		Int64("message_id", msgResp.MessageID).
		Str("parse_mode", parseMode).
		Msg("message sent successfully")

	return nil
}
