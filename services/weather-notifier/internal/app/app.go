package app

import (
	"fmt"
	"os"

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
	chatHandlers   []*notifications.ChatHandler
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
		cfg.Telegram.APIKey,
		cfg.Telegram.APIURL,
		logger.Logger,
	)
	logger.Info().Msg("telegram client initialized")

	// Initialize chat handlers for each configured chat
	var chatHandlers []*notifications.ChatHandler
	thresholds := detector.DefaultThresholds()

	for _, chatCfg := range cfg.Chats {
		// Resolve LLM config for this chat (global + per-chat overrides)
		llmConfig := chatCfg.ResolveLLM(cfg.LLM)

		// Create LLM client for this chat
		chatLLMClient := llm.NewClient(
			llmConfig.BaseURL,
			llmConfig.APIKey,
			llmConfig.Model,
			llmConfig.MaxTokens,
			llmConfig.Temperature,
		)

		// Initialize locations for this chat
		var locations []notifications.LocationContext
		for _, locCfg := range chatCfg.Locations {
			// Create per-location state file path
			stateFile := GetStateFilePath(dataDir, chatCfg.ChatID, locCfg.Name)

			// Create detector for this location
			detectorAdapter := NewDetectorAdapter(thresholds, stateFile, logger.Logger)

			locations = append(locations, notifications.LocationContext{
				Name:      locCfg.Name,
				Latitude:  locCfg.Latitude,
				Longitude: locCfg.Longitude,
				Detector:  detectorAdapter,
			})

			logger.Info().
				Int64("chat_id", chatCfg.ChatID).
				Str("location", locCfg.Name).
				Str("state_file", stateFile).
				Msg("location initialized")
		}

		// Create chat handler
		chatHandler := notifications.NewChatHandler(
			weatherClient,
			telegramClient,
			chatLLMClient,
			chatCfg.ChatID,
			chatCfg.Name,
			locations,
			logger.Logger,
		)
		chatHandlers = append(chatHandlers, chatHandler)

		logger.Info().
			Int64("chat_id", chatCfg.ChatID).
			Str("chat_name", chatCfg.Name).
			Int("locations", len(locations)).
			Str("llm_model", llmConfig.Model).
			Msg("chat handler initialized")
	}

	// Initialize scheduler
	scheduler, err := notifications.NewScheduler(cfg, chatHandlers, logger.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize scheduler: %w", err)
	}
	logger.Info().Msg("scheduler initialized")

	return &App{
		config:         cfg,
		logger:         logger,
		weatherClient:  weatherClient,
		telegramClient: telegramClient,
		chatHandlers:   chatHandlers,
		scheduler:      scheduler,
	}, nil
}

// Start begins the application's scheduled tasks
func (a *App) Start() error {
	a.logger.Info().Msg("starting weather notification bot")

	// Start the scheduler
	a.scheduler.Start()

	a.logger.Info().
		Int("chats", len(a.chatHandlers)).
		Msg("weather notification bot started successfully")
	return nil
}

// Stop gracefully shuts down the application
func (a *App) Stop() {
	a.logger.Info().Msg("stopping weather notification bot")

	// Stop the scheduler
	if a.scheduler != nil {
		a.scheduler.Stop()
	}

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
