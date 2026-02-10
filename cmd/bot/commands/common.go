package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/kexi/telegram-bot-gateway/internal/config"
	"github.com/kexi/telegram-bot-gateway/internal/repository"
	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// initDB initializes the database connection
func initDB() (*gorm.DB, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.json"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := repository.NewDatabase(&cfg.Database, "release")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// initBotService initializes the bot service
func initBotService(db *gorm.DB) (*service.BotService, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.json"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	botRepo := repository.NewBotRepository(db)
	botService := service.NewBotService(botRepo, cfg.Auth.JWT.Secret, cfg.Telegram.WebhookBaseURL)

	return botService, nil
}

// getContext returns a context with timeout
func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// fatal logs an error and exits
func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

// success prints a success message
func success(format string, args ...interface{}) {
	fmt.Printf("✓ "+format+"\n", args...)
}

// info prints an info message
func info(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// printBotTable prints a formatted table of bots
func printBotTable(bots []service.BotDTO) {
	if len(bots) == 0 {
		info("No bots found")
		return
	}

	fmt.Println("\nID | Username         | Display Name     | Active | Webhook URL")
	fmt.Println("---|------------------|------------------|--------|--------------------------------------------------")

	for _, b := range bots {
		username := b.Username
		if len(username) > 16 {
			username = username[:13] + "..."
		}

		displayName := b.DisplayName
		if displayName == "" {
			displayName = "-"
		}
		if len(displayName) > 16 {
			displayName = displayName[:13] + "..."
		}

		active := "Yes"
		if !b.IsActive {
			active = "No"
		}

		webhookURL := b.WebhookURL
		if len(webhookURL) > 50 {
			webhookURL = webhookURL[:47] + "..."
		}

		fmt.Printf("%-2d | %-16s | %-16s | %-6s | %s\n",
			b.ID, username, displayName, active, webhookURL)
	}
	fmt.Println()
}

// hasFlag checks if a flag is present in args
func hasFlag(args []string, flag ...string) bool {
	for _, arg := range args {
		for _, f := range flag {
			if arg == f {
				return true
			}
		}
	}
	return false
}

// getFlagValue gets the value of a flag
func getFlagValue(args []string, flag ...string) string {
	for i, arg := range args {
		for _, f := range flag {
			if arg == f && i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return ""
}
