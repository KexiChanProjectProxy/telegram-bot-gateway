package commands

import (
	"fmt"

	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// Create creates a new bot
func Create(args []string) {
	// Parse flags
	username := getFlagValue(args, "--username")
	token := getFlagValue(args, "--token")
	displayName := getFlagValue(args, "--display-name")
	description := getFlagValue(args, "--description")

	if username == "" {
		fatal("--username is required")
	}
	if token == "" {
		fatal("--token is required")
	}

	// Initialize database and service
	db, err := initDB()
	if err != nil {
		fatal("Failed to initialize database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	botService, err := initBotService(db)
	if err != nil {
		fatal("Failed to initialize bot service: %v", err)
	}

	// Create bot
	ctx, cancel := getContext()
	defer cancel()

	req := &service.CreateBotRequest{
		Username:    username,
		Token:       token,
		DisplayName: displayName,
		Description: description,
	}

	info("Creating bot and registering webhook with Telegram...")
	bot, err := botService.CreateBot(ctx, req)
	if err != nil {
		fatal("Failed to create bot: %v", err)
	}

	success("Bot created successfully")
	fmt.Printf("\nBot Details:\n")
	fmt.Printf("  ID:           %d\n", bot.ID)
	fmt.Printf("  Username:     %s\n", bot.Username)
	fmt.Printf("  Display Name: %s\n", bot.DisplayName)
	fmt.Printf("  Webhook URL:  %s\n", bot.WebhookURL)
	fmt.Printf("\nThe webhook has been registered with Telegram.\n")
}
