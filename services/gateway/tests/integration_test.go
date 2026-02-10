package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kexi/telegram-bot-gateway/internal/config"
	"github.com/kexi/telegram-bot-gateway/internal/pkg/jwt"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// TestAuthFlow tests the complete authentication flow
func TestAuthFlow(t *testing.T) {
	// Setup
	cfg := getTestConfig(t)
	db := setupTestDB(t, cfg)
	defer cleanupTestDB(t, db)

	// Initialize services
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := jwt.NewService(
		cfg.Auth.JWT.Secret,
		cfg.Auth.JWT.Issuer,
		cfg.Auth.JWT.AccessTokenTTL,
		cfg.Auth.JWT.RefreshTokenTTL,
		cfg.Auth.JWT.RefreshThreshold,
	)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtService)

	ctx := context.Background()

	// Test user creation
	t.Run("CreateUser", func(t *testing.T) {
		user, err := authService.CreateUser(ctx, "testuser", "test@example.com", "password123")
		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.True(t, user.IsActive)
	})

	// Test login
	t.Run("Login", func(t *testing.T) {
		loginReq := &service.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}

		resp, err := authService.Login(ctx, loginReq)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Equal(t, "testuser", resp.User.Username)

		// Validate access token
		claims, err := jwtService.ValidateToken(resp.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, "testuser", claims.Username)
	})

	// Test login with wrong password
	t.Run("LoginWrongPassword", func(t *testing.T) {
		loginReq := &service.LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}

		_, err := authService.Login(ctx, loginReq)
		assert.Error(t, err)
	})

	// Test token refresh
	t.Run("RefreshToken", func(t *testing.T) {
		// First login
		loginReq := &service.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}
		loginResp, err := authService.Login(ctx, loginReq)
		require.NoError(t, err)

		// Refresh
		refreshReq := &service.RefreshRequest{
			RefreshToken: loginResp.RefreshToken,
		}
		refreshResp, err := authService.Refresh(ctx, refreshReq)
		require.NoError(t, err)
		assert.NotEmpty(t, refreshResp.AccessToken)
		assert.NotEqual(t, loginResp.AccessToken, refreshResp.AccessToken)
	})

	// Test logout
	t.Run("Logout", func(t *testing.T) {
		// Login
		loginReq := &service.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}
		loginResp, err := authService.Login(ctx, loginReq)
		require.NoError(t, err)

		// Logout
		err = authService.Logout(ctx, loginResp.RefreshToken)
		require.NoError(t, err)

		// Try to use revoked token
		refreshReq := &service.RefreshRequest{
			RefreshToken: loginResp.RefreshToken,
		}
		_, err = authService.Refresh(ctx, refreshReq)
		assert.Error(t, err)
	})
}

// TestBotManagement tests bot CRUD operations
func TestBotManagement(t *testing.T) {
	cfg := getTestConfig(t)
	db := setupTestDB(t, cfg)
	defer cleanupTestDB(t, db)

	botRepo := repository.NewBotRepository(db)
	botService := service.NewBotService(botRepo, cfg.Auth.JWT.Secret)

	ctx := context.Background()

	var botID uint

	t.Run("CreateBot", func(t *testing.T) {
		req := &service.CreateBotRequest{
			Username:    "test_bot",
			Token:       "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
			DisplayName: "Test Bot",
			Description: "A bot for testing",
		}

		bot, err := botService.CreateBot(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test_bot", bot.Username)
		assert.Equal(t, "Test Bot", bot.DisplayName)
		assert.True(t, bot.IsActive)

		botID = bot.ID
	})

	t.Run("GetBot", func(t *testing.T) {
		bot, err := botService.GetBot(ctx, botID)
		require.NoError(t, err)
		assert.Equal(t, "test_bot", bot.Username)
	})

	t.Run("ListBots", func(t *testing.T) {
		bots, err := botService.ListBots(ctx, 0, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(bots), 1)
	})

	t.Run("UpdateBot", func(t *testing.T) {
		err := botService.UpdateBot(ctx, botID, "Updated Bot", "Updated description", true)
		require.NoError(t, err)

		bot, err := botService.GetBot(ctx, botID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Bot", bot.DisplayName)
	})

	t.Run("DeleteBot", func(t *testing.T) {
		err := botService.DeleteBot(ctx, botID)
		require.NoError(t, err)

		_, err = botService.GetBot(ctx, botID)
		assert.Error(t, err)
	})
}

// TestMessageFlow tests message storage and retrieval
func TestMessageFlow(t *testing.T) {
	cfg := getTestConfig(t)
	db := setupTestDB(t, cfg)
	defer cleanupTestDB(t, db)

	botRepo := repository.NewBotRepository(db)
	chatRepo := repository.NewChatRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	botService := service.NewBotService(botRepo, cfg.Auth.JWT.Secret)
	chatService := service.NewChatService(chatRepo, botRepo)
	messageService := service.NewMessageService(messageRepo, chatRepo)

	ctx := context.Background()

	// Create bot
	botReq := &service.CreateBotRequest{
		Username: "msg_test_bot",
		Token:    "123456:TOKEN",
	}
	bot, err := botService.CreateBot(ctx, botReq)
	require.NoError(t, err)

	// Create chat
	chatReq := &service.CreateChatRequest{
		BotID:      bot.ID,
		TelegramID: 123456789,
		Type:       "private",
		Username:   "testuser",
		FirstName:  "Test",
	}
	chat, err := chatService.CreateOrUpdateChat(ctx, chatReq)
	require.NoError(t, err)

	// Store message
	t.Run("StoreMessage", func(t *testing.T) {
		msgReq := &service.CreateMessageRequest{
			ChatID:       chat.ID,
			TelegramID:   1,
			FromUsername: "testuser",
			Direction:    "incoming",
			MessageType:  "text",
			Text:         "Hello, World!",
			SentAt:       time.Now(),
		}

		msg, err := messageService.StoreMessage(ctx, msgReq)
		require.NoError(t, err)
		assert.Equal(t, "Hello, World!", msg.Text)
	})

	// List messages
	t.Run("ListMessages", func(t *testing.T) {
		messages, err := messageService.ListMessages(ctx, chat.ID, nil, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(messages), 1)
	})
}

// Helper functions

func getTestConfig(t *testing.T) *config.Config {
	return &config.Config{
		Database: config.DatabaseConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     3306,
			Name:     "telegram_gateway_test",
			User:     "root",
			Password: "password",
		},
		Auth: config.AuthConfig{
			JWT: config.JWTConfig{
				Secret:          "test-secret-key-min-32-chars-long",
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 7 * 24 * time.Hour,
				Issuer:          "test",
			},
		},
	}
}

func setupTestDB(t *testing.T, cfg *config.Config) interface{} {
	db, err := repository.NewDatabase(&cfg.Database, "test")
	require.NoError(t, err)

	// Run migrations
	// Note: In real tests, you'd run migrations here

	return db
}

func cleanupTestDB(t *testing.T, db interface{}) {
	// Cleanup tables
	// Note: In real tests, you'd clean up test data here
}
