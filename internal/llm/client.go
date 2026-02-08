package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with OpenAI-compatible LLM APIs
type Client struct {
	baseURL     string
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	httpClient  *http.Client
}

// NewClient creates a new LLM client
func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		baseURL:     baseURL,
		apiKey:      apiKey,
		model:       model,
		maxTokens:   500, // Default max tokens for weather advice
		temperature: 0.7, // Balanced creativity and consistency
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateAdvice generates weather advice based on the provided prompt
func (c *Client) GenerateAdvice(ctx context.Context, weatherPrompt string) (string, error) {
	// Build the request payload
	reqBody := ChatRequest{
		Model:       c.model,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
		Messages: []Message{
			{
				Role:    "system",
				Content: SystemPrompt,
			},
			{
				Role:    "user",
				Content: weatherPrompt,
			},
		},
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Validate response
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned in response")
	}

	// Extract the generated content
	content := chatResp.Choices[0].Message.Content
	if content == "" {
		return "", fmt.Errorf("empty content in response")
	}

	return content, nil
}
