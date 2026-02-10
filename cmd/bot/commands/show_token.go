package commands

import (
	"fmt"
	"strconv"
)

// ShowToken displays the decrypted bot token
func ShowToken(args []string) {
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

	// Get token
	ctx, cancel := getContext()
	defer cancel()

	token, err := botService.GetBotToken(ctx, uint(id))
	if err != nil {
		fatal("Failed to get bot token: %v", err)
	}

	fmt.Printf("\nBot Token (ID: %d):\n", id)
	fmt.Printf("  %s\n\n", token)
	fmt.Println("⚠️  Keep this token secure! Anyone with this token can control your bot.")
}
