package commands

import (
	"fmt"
	"strconv"
)

const revokeBotUsage = `Revoke bot usage permission from an API key

Usage:
  apikey revoke-bot <apikey-id> <bot-id>

Arguments:
  <apikey-id>               API key ID
  <bot-id>                  Bot ID

Options:
  --help, -h                Show this help message

Examples:
  apikey revoke-bot 1 2

Note: If this was the last bot permission, the API key will again be able to use ALL bots.
`

func RevokeBotPermission(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) < 2 {
		fmt.Println(revokeBotUsage)
		return
	}

	apiKeyID, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		fatal("Invalid API key ID: %s", args[0])
	}

	botID, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		fatal("Invalid bot ID: %s", args[1])
	}

	// Initialize database
	db, err := initDB()
	if err != nil {
		fatal("Failed to initialize database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	_, _, botPermRepo, _, _, _ := initRepositories(db)

	ctx, cancel := getContext()
	defer cancel()

	// Delete bot permission
	if err := botPermRepo.Delete(ctx, uint(apiKeyID), uint(botID)); err != nil {
		fatal("Failed to revoke bot permission: %v", err)
	}

	success("Revoked bot permission: API key %d can no longer use bot %d", apiKeyID, botID)
}
