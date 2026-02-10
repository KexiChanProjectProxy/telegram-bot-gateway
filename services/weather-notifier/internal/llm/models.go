package llm

// Message represents a single message in the chat conversation
type Message struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"` // The message content
}

// ChatRequest represents the request payload for OpenAI-compatible chat completions
type ChatRequest struct {
	Model       string    `json:"model"`                 // The model ID to use
	Messages    []Message `json:"messages"`              // Array of conversation messages
	MaxTokens   int       `json:"max_tokens,omitempty"`  // Maximum tokens to generate
	Temperature float64   `json:"temperature,omitempty"` // Sampling temperature (0-2)
}

// Choice represents a single completion choice in the response
type Choice struct {
	Message Message `json:"message"` // The generated message
	Index   int     `json:"index"`   // The choice index
}

// ChatResponse represents the response from OpenAI-compatible chat completions
type ChatResponse struct {
	Choices []Choice `json:"choices"` // Array of completion choices
	ID      string   `json:"id"`      // Response ID
	Model   string   `json:"model"`   // Model used
}
