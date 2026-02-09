package commands

import (
	"fmt"
	"strconv"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
)

const grantFeedbackUsage = `Grant feedback permission to an API key

Usage:
  apikey grant-feedback <apikey-id> <chat-id>

Arguments:
  <apikey-id>               API key ID
  <chat-id>                 Chat ID that can send feedback

Options:
  --help, -h                Show this help message

Examples:
  apikey grant-feedback 1 5

Note: Once ANY feedback permission is granted, the API key can ONLY receive feedback from explicitly allowed chats.
      If no feedback permissions exist, all chats can send feedback (default behavior).
`

func GrantFeedbackPermission(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) < 2 {
		fmt.Println(grantFeedbackUsage)
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

	// Create feedback permission
	perm := &domain.APIKeyFeedbackPermission{
		APIKeyID:           uint(apiKeyID),
		ChatID:             uint(chatID),
		CanReceiveFeedback: true,
	}

	if err := feedbackPermRepo.Create(ctx, perm); err != nil {
		fatal("Failed to grant feedback permission: %v", err)
	}

	success("Granted feedback permission: API key %d can now receive messages from chat %d", apiKeyID, chatID)
	info("")
	info("⚠️  Note: This API key is now restricted to receiving feedback from explicitly allowed chats only.")
}
