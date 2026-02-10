package commands

import (
	"strconv"
)

// Update updates a bot
func Update(args []string) {
	if len(args) < 1 {
		fatal("Bot ID is required")
	}

	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		fatal("Invalid bot ID: %v", err)
	}

	// Parse flags
	displayName := getFlagValue(args, "--display-name")
	description := getFlagValue(args, "--description")
	activeStr := getFlagValue(args, "--active")

	if displayName == "" && description == "" && activeStr == "" {
		fatal("At least one of --display-name, --description, or --active must be provided")
	}

	active := true
	if activeStr != "" {
		if activeStr == "false" || activeStr == "0" {
			active = false
		} else if activeStr != "true" && activeStr != "1" {
			fatal("Invalid value for --active (use true/false or 1/0)")
		}
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

	// Update bot
	ctx, cancel := getContext()
	defer cancel()

	err = botService.UpdateBot(ctx, uint(id), displayName, description, active)
	if err != nil {
		fatal("Failed to update bot: %v", err)
	}

	success("Bot updated successfully")
}
