package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/kexi/telegram-bot-gateway/internal/config"
	"github.com/kexi/telegram-bot-gateway/internal/handler"
	"github.com/kexi/telegram-bot-gateway/internal/middleware"
	"github.com/kexi/telegram-bot-gateway/internal/pkg/apikey"
	"github.com/kexi/telegram-bot-gateway/internal/pkg/jwt"
	"github.com/kexi/telegram-bot-gateway/internal/pubsub"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
	"github.com/kexi/telegram-bot-gateway/internal/service"
	"github.com/kexi/telegram-bot-gateway/internal/websocket"
	"github.com/kexi/telegram-bot-gateway/internal/worker"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database
	db, err := repository.NewDatabase(&cfg.Database, cfg.Server.Mode)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	log.Println("✓ Connected to database")

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	log.Println("✓ Connected to Redis")

	// Initialize services
	jwtService := jwt.NewService(
		cfg.Auth.JWT.Secret,
		cfg.Auth.JWT.Issuer,
		cfg.Auth.JWT.AccessTokenTTL,
		cfg.Auth.JWT.RefreshTokenTTL,
		cfg.Auth.JWT.RefreshThreshold,
	)

	apiKeyService := apikey.NewService(
		cfg.Auth.APIKey.Prefix,
		cfg.Auth.APIKey.Length,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	botRepo := repository.NewBotRepository(db)
	chatRepo := repository.NewChatRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	chatPermRepo := repository.NewChatPermissionRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)
	webhookDeliveryRepo := repository.NewWebhookDeliveryRepository(db)

	// Initialize message broker and real-time components
	messageBroker := pubsub.NewMessageBroker(redisClient)
	wsHub := websocket.NewHub(messageBroker)

	// Initialize business services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtService)
	botService := service.NewBotService(botRepo, cfg.Auth.JWT.Secret)
	chatService := service.NewChatService(chatRepo, botRepo)
	messageService := service.NewMessageService(messageRepo, chatRepo)
	// apiKeySvc removed - API key management moved to CLI tool
	webhookService := service.NewWebhookService(webhookRepo, chatRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	botHandler := handler.NewBotHandler(botService)
	chatHandler := handler.NewChatHandler(chatService, messageService)
	// apiKeyHandler removed - API key management moved to CLI tool
	webhookHandler := handler.NewWebhookHandler(webhookService)
	telegramHandler := handler.NewTelegramHandler(botService, chatService, messageService, messageBroker)
	wsHandler := handler.NewWebSocketHandler(wsHub)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(
		redisClient,
		cfg.RateLimit.RequestsPerSecond,
		cfg.RateLimit.Burst,
		cfg.RateLimit.CleanupInterval,
	)

	// Setup Gin router
	router := gin.Default()

	// Global rate limiting (optional - can be enabled for DDoS protection)
	// router.Use(middleware.GlobalRateLimitMiddleware(rateLimiter))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"websocket_clients": wsHub.GetClientCount(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public auth endpoints (with rate limiting)
		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimitMiddleware(rateLimiter))
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected endpoints (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService, apiKeyService, apiKeyRepo))
		protected.Use(middleware.PerUserRateLimitMiddleware(rateLimiter))
		{
			// Bot management
			bots := protected.Group("/bots")
			{
				bots.POST("", botHandler.CreateBot)
				bots.GET("", botHandler.ListBots)
				bots.GET("/:id", botHandler.GetBot)
				bots.DELETE("/:id", botHandler.DeleteBot)
			}

			// Chat management
			chats := protected.Group("/chats")
			{
				chats.GET("", chatHandler.ListChats)
				chats.GET("/:id", chatHandler.GetChat)

				// Message endpoints with ACL
				chats.GET("/:id/messages",
					middleware.ChatACLMiddleware(middleware.PermissionRead, chatPermRepo, redisClient),
					chatHandler.GetMessages,
				)
				chats.POST("/:id/messages",
					middleware.ChatACLMiddleware(middleware.PermissionSend, chatPermRepo, redisClient),
					chatHandler.SendMessage,
				)
			}

			// API key management - DISABLED (use CLI: ./bin/apikey)
			// apikeys := protected.Group("/apikeys")
			// {
			//     apikeys.POST("", apiKeyHandler.CreateAPIKey)
			//     apikeys.GET("", apiKeyHandler.ListAPIKeys)
			//     apikeys.GET("/:id", apiKeyHandler.GetAPIKey)
			//     apikeys.POST("/:id/revoke", apiKeyHandler.RevokeAPIKey)
			//     apikeys.DELETE("/:id", apiKeyHandler.DeleteAPIKey)
			// }

			// Webhook management
			webhooks := protected.Group("/webhooks")
			{
				webhooks.POST("", webhookHandler.CreateWebhook)
				webhooks.GET("", webhookHandler.ListWebhooks)
				webhooks.GET("/:id", webhookHandler.GetWebhook)
				webhooks.PUT("/:id", webhookHandler.UpdateWebhook)
				webhooks.DELETE("/:id", webhookHandler.DeleteWebhook)
			}

			// WebSocket endpoint
			protected.GET("/ws", wsHandler.HandleWebSocket)
		}

		// Telegram webhook receiver (no auth - validated by bot token in URL)
		v1.POST("/telegram/webhook/:bot_username", telegramHandler.ReceiveUpdate)
	}

	// Create context for background workers
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	// Start WebSocket hub
	go wsHub.Run(workerCtx)
	log.Println("✓ WebSocket hub started")

	// Start webhook workers
	for i := 0; i < cfg.WebhookDelivery.WorkerCount; i++ {
		webhookWorker := worker.NewWebhookWorker(
			i+1,
			messageBroker,
			webhookService,
			messageService,
			webhookDeliveryRepo,
			cfg.WebhookDelivery.MaxRetries,
		)
		go webhookWorker.Start(workerCtx)
	}
	log.Printf("✓ Started %d webhook workers", cfg.WebhookDelivery.WorkerCount)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         cfg.Server.HTTP.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.HTTP.ReadTimeout,
		WriteTimeout: cfg.Server.HTTP.WriteTimeout,
		IdleTimeout:  cfg.Server.HTTP.IdleTimeout,
	}

	// Start HTTP server in goroutine
	go func() {
		log.Printf("🚀 HTTP server starting on %s", cfg.Server.HTTP.Address)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	log.Println("✓ All services started successfully")
	log.Println("✓ Press Ctrl+C to shutdown")

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Stop workers first
	workerCancel()

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("✓ Servers stopped gracefully")
}

// initDefaultUser creates a default admin user if none exists
func initDefaultUser(authService *service.AuthService) {
	ctx := context.Background()

	// Check if we should create default user
	// This would typically check if any users exist
	// For now, we'll skip this in production

	username := os.Getenv("DEFAULT_ADMIN_USER")
	password := os.Getenv("DEFAULT_ADMIN_PASSWORD")

	if username != "" && password != "" {
		_, err := authService.CreateUser(ctx, username, "", password)
		if err != nil {
			log.Printf("Warning: Failed to create default user: %v", err)
		} else {
			log.Printf("✓ Created default admin user: %s", username)
		}
	}
}
