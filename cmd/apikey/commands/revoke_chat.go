package commands

import (
	"fmt"
	"strconv"
)

const revokeChatUsage = `Revoke chat permissions from an API key

Usage:
  apikey revoke-chat <apikey-id> <chat-id>

Arguments:
  <apikey-id>               API key ID
  <chat-id>                 Chat ID

Options:
  --help, -h                Show this help message

Examples:
  apikey revoke-chat 1 5

Note: This removes all chat permissions for the specified chat
`

func RevokeChatPermission(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) < 2 {
		fmt.Println(revokeChatUsage)
		return
	}

	apiKeyID, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		fatal("Invalid API key ID: %s", args[0])
	}

	chatID, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		fatal("Invalid chat ID: %s", args[1])
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

	_, chatPermRepo, _, _, _, _ := initRepositories(db)

	ctx, cancel := getContext()
	defer cancel()

	// Get existing permission
	perm, err := chatPermRepo.GetByAPIKeyAndChat(ctx, uint(apiKeyID), uint(chatID))
	if err != nil {
		fatal("Chat permission not found")
	}

	// Delete permission
	if err := chatPermRepo.Delete(ctx, perm.ID); err != nil {
		fatal("Failed to revoke chat permission: %v", err)
	}

	success("Revoked chat permission for API key %d on chat %d", apiKeyID, chatID)
}
