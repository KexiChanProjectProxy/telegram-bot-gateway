package commands

import (
	"fmt"
	"strconv"
)

// Get retrieves a bot by ID
func Get(args []string) {
	if len(args) < 1 {
		fatal("Bot ID is required")
	}

	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		fatal("Invalid bot ID: %v", err)
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

	// Get bot
	ctx, cancel := getContext()
	defer cancel()

	bot, err := botService.GetBot(ctx, uint(id))
	if err != nil {
		fatal("Failed to get bot: %v", err)
	}

	fmt.Printf("\nBot Details:\n")
	fmt.Printf("  ID:           %d\n", bot.ID)
	fmt.Printf("  Username:     %s\n", bot.Username)
	fmt.Printf("  Display Name: %s\n", bot.DisplayName)
	fmt.Printf("  Description:  %s\n", bot.Description)
	fmt.Printf("  Active:       %v\n", bot.IsActive)
	fmt.Printf("  Webhook URL:  %s\n", bot.WebhookURL)
	fmt.Println()
}
