package app

import (
	"fmt"
	"os"
	"time"

	"github.com/user/weather-notice-bot/internal/config"
	"github.com/user/weather-notice-bot/internal/detector"
	"github.com/user/weather-notice-bot/internal/llm"
	"github.com/user/weather-notice-bot/internal/notifications"
	"github.com/user/weather-notice-bot/internal/telegram"
	"github.com/user/weather-notice-bot/internal/utils"
	"github.com/user/weather-notice-bot/internal/weather"

	"github.com/rs/zerolog"
)

// App represents the main application with all its components
type App struct {
	config         *config.Config
	logger         *utils.Logger
	weatherClient  *weather.Client
	telegramClient *telegram.Client
	llmClient      *llm.Client
	detector       *DetectorAdapter
	handlers       *notifications.NotificationHandlers
	scheduler      *notifications.Scheduler
}

// New creates a new application instance with all components initialized
func New(cfg *config.Config) (*App, error) {
	// Initialize logger
	logger := utils.InitLogger(utils.LogConfig{
		Level:       cfg.Logging.Level,
		PrettyPrint: cfg.Logging.PrettyPrint,
	})

	logger.Info().Msg("initializing weather notification bot")

	// Create data directory if it doesn't exist
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize weather client
	weatherClient := weather.NewClient(cfg.Caiyun.APIToken)
	logger.Info().Msg("weather client initialized")

	// Initialize Telegram client
	telegramClient := telegram.NewClient(
		cfg.Telegram.BotToken,
		cfg.Telegram.Password,
		cfg.Telegram.APIURL,
		logger.Logger,
	)
	logger.Info().Msg("telegram client initialized")

	// Initialize LLM client
	llmClient := llm.NewClient(
		cfg.LLM.BaseURL,
		cfg.LLM.APIKey,
		cfg.LLM.Model,
	)
	logger.Info().Msg("llm client initialized")

	// Initialize detector with adapter
	thresholds := detector.DefaultThresholds()
	stateFile := GetStateFilePath(dataDir)
	detectorAdapter := NewDetectorAdapter(thresholds, stateFile, logger.Logger)
	logger.Info().Str("state_file", stateFile).Msg("detector initialized")

	// Determine the chat ID to send notifications to
	// If admin_user_id is set, use it; otherwise use the first allowed ID
	chatID := cfg.Telegram.AdminUserID
	if chatID == 0 && len(cfg.Telegram.AllowedIDs) > 0 {
		chatID = cfg.Telegram.AllowedIDs[0]
		logger.Info().Int64("chat_id", chatID).Msg("using first allowed ID as chat ID")
	} else if chatID == 0 {
		return nil, fmt.Errorf("no chat ID configured: set either telegram.admin_user_id or telegram.allowed_ids")
	}

	// Initialize notification handlers
	handlers := notifications.NewNotificationHandlers(
		weatherClient,
		telegramClient,
		llmClient,
		detectorAdapter,
		chatID,
		cfg.Caiyun.Latitude,
		cfg.Caiyun.Longitude,
		logger.Logger,
	)
	logger.Info().Msg("notification handlers initialized")

	// Initialize scheduler
	scheduler, err := notifications.NewScheduler(cfg, handlers, logger.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize scheduler: %w", err)
	}
	logger.Info().Msg("scheduler initialized")

	return &App{
		config:         cfg,
		logger:         logger,
		weatherClient:  weatherClient,
		telegramClient: telegramClient,
		llmClient:      llmClient,
		detector:       detectorAdapter,
		handlers:       handlers,
		scheduler:      scheduler,
	}, nil
}

// Start begins the application's scheduled tasks
func (a *App) Start() error {
	a.logger.Info().Msg("starting weather notification bot")

	// Start the scheduler
	a.scheduler.Start()

	a.logger.Info().Msg("weather notification bot started successfully")
	return nil
}

// Stop gracefully shuts down the application
func (a *App) Stop() {
	a.logger.Info().Msg("stopping weather notification bot")

	// Stop the scheduler
	if a.scheduler != nil {
		a.scheduler.Stop()
	}

	// Give a moment for any pending operations to complete
	time.Sleep(1 * time.Second)

	a.logger.Info().Msg("weather notification bot stopped")
}

// GetLogger returns the application's logger
func (a *App) GetLogger() zerolog.Logger {
	return a.logger.Logger
}

// GetConfig returns the application's configuration
func (a *App) GetConfig() *config.Config {
	return a.config
}
