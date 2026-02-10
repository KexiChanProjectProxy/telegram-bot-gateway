package commands

import (
	"fmt"
	"strconv"
)

const revokeFeedbackUsage = `Revoke feedback permission from an API key

Usage:
  apikey revoke-feedback <apikey-id> <chat-id>

Arguments:
  <apikey-id>               API key ID
  <chat-id>                 Chat ID

Options:
  --help, -h                Show this help message

Examples:
  apikey revoke-feedback 1 5

Note: If this was the last feedback permission, the API key will again be able to receive feedback from ALL chats.
`

func RevokeFeedbackPermission(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) < 2 {
		fmt.Println(revokeFeedbackUsage)
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

	_, _, _, feedbackPermRepo, _, _ := initRepositories(db)

	ctx, cancel := getContext()
	defer cancel()

	// Delete feedback permission
	if err := feedbackPermRepo.Delete(ctx, uint(apiKeyID), uint(chatID)); err != nil {
		fatal("Failed to revoke feedback permission: %v", err)
	}

	success("Revoked feedback permission: API key %d can no longer receive messages from chat %d", apiKeyID, chatID)
}
