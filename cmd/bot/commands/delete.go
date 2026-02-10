package commands

import (
	"fmt"
	"strconv"
)

// Delete deletes a bot
func Delete(args []string) {
	if len(args) < 1 {
		fatal("Bot ID is required")
	}

	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		fatal("Invalid bot ID: %v", err)
	}

	// Check for --force flag
	if !hasFlag(args, "--force") {
		fmt.Println("\nWARNING: Deleting a bot will:")
		fmt.Println("  - Delete the bot record")
		fmt.Println("  - CASCADE delete all associated chats")
		fmt.Println("  - CASCADE delete all associated permissions")
		fmt.Println("  - Deregister the webhook from Telegram")
		fmt.Println("\nThis operation cannot be undone!")
		fmt.Println("\nTo proceed, add the --force flag:")
		fmt.Printf("  bot delete %d --force\n\n", id)
		return
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

	// Delete bot
	ctx, cancel := getContext()
	defer cancel()

	info("Deleting bot and deregistering webhook from Telegram...")
	err = botService.DeleteBot(ctx, uint(id))
	if err != nil {
		fatal("Failed to delete bot: %v", err)
	}

	success("Bot deleted successfully")
}
