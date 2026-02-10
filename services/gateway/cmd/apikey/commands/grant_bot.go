package commands

import (
	"fmt"
	"strconv"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
)

const grantBotUsage = `Grant bot usage permission to an API key

Usage:
  apikey grant-bot <apikey-id> <bot-id>

Arguments:
  <apikey-id>               API key ID
  <bot-id>                  Bot ID

Options:
  --help, -h                Show this help message

Examples:
  apikey grant-bot 1 2

Note: Once ANY bot permission is granted, the API key can ONLY use explicitly allowed bots.
      If no bot permissions exist, all bots are allowed (default behavior).
`

func GrantBotPermission(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) < 2 {
		fmt.Println(grantBotUsage)
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

	// Create bot permission
	perm := &domain.APIKeyBotPermission{
		APIKeyID: uint(apiKeyID),
		BotID:    uint(botID),
		CanSend:  true,
	}

	if err := botPermRepo.Create(ctx, perm); err != nil {
		fatal("Failed to grant bot permission: %v", err)
	}

	success("Granted bot permission: API key %d can now use bot %d", apiKeyID, botID)
	info("")
	info("⚠️  Note: This API key is now restricted to explicitly allowed bots only.")
}
