package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// TokenManager manages JWT tokens with auto-refresh capability
type TokenManager struct {
	botToken   string
	password   string
	apiURL     string
	httpClient *http.Client
	logger     zerolog.Logger

	mu           sync.RWMutex
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	BotToken string `json:"bot_token"`
	Password string `json:"password"`
}

// RefreshRequest represents the refresh request payload
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// NewTokenManager creates a new TokenManager instance
func NewTokenManager(botToken, password, apiURL string, logger zerolog.Logger) *TokenManager {
	return &TokenManager{
		botToken: botToken,
		password: password,
		apiURL:   apiURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.With().Str("component", "token_manager").Logger(),
	}
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (tm *TokenManager) GetAccessToken(ctx context.Context) (string, error) {
	tm.mu.RLock()
	// Check if we have a valid token (expires in more than 5 minutes)
	if tm.accessToken != "" && time.Until(tm.expiresAt) > 5*time.Minute {
		token := tm.accessToken
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	// Need to refresh or login
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring write lock
	if tm.accessToken != "" && time.Until(tm.expiresAt) > 5*time.Minute {
		return tm.accessToken, nil
	}

	// Try to refresh if we have a refresh token
	if tm.refreshToken != "" {
		tm.logger.Debug().Msg("attempting to refresh access token")
		if err := tm.refresh(ctx); err != nil {
			tm.logger.Warn().Err(err).Msg("token refresh failed, attempting login")
			// Refresh failed, try to login
			if loginErr := tm.login(ctx); loginErr != nil {
				return "", fmt.Errorf("failed to refresh and login: refresh=%w, login=%v", err, loginErr)
			}
		}
	} else {
		// No refresh token, perform login
		tm.logger.Debug().Msg("no refresh token available, performing login")
		if err := tm.login(ctx); err != nil {
			return "", fmt.Errorf("login failed: %w", err)
		}
	}

	return tm.accessToken, nil
}

// login performs the initial authentication
func (tm *TokenManager) login(ctx context.Context) error {
	reqBody := LoginRequest{
		BotToken: tm.botToken,
		Password: tm.password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/auth/login", tm.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	tm.accessToken = authResp.AccessToken
	tm.refreshToken = authResp.RefreshToken
	tm.expiresAt = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	tm.logger.Info().
		Time("expires_at", tm.expiresAt).
		Msg("successfully logged in and obtained tokens")

	return nil
}

// refresh refreshes the access token using the refresh token
func (tm *TokenManager) refresh(ctx context.Context) error {
	reqBody := RefreshRequest{
		RefreshToken: tm.refreshToken,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/auth/refresh", tm.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to unmarshal refresh response: %w", err)
	}

	tm.accessToken = authResp.AccessToken
	tm.refreshToken = authResp.RefreshToken
	tm.expiresAt = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	tm.logger.Info().
		Time("expires_at", tm.expiresAt).
		Msg("successfully refreshed access token")

	return nil
}
