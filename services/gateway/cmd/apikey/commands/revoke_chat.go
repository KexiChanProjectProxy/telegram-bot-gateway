package commands

import (
	"fmt"
	"strconv"
)

const revokeChatUsage = `Revoke chat permissions from an API key

Usage:
  apikey revoke-chat <apikey-id> <bot-id> <telegram-chat-id>

Arguments:
  <apikey-id>               API key ID
  <bot-id>                  Bot ID
  <telegram-chat-id>        Telegram chat ID (the actual Telegram chat identifier)

Options:
  --help, -h                Show this help message

Examples:
  apikey revoke-chat 1 1 1878878763

Note: This removes all chat permissions for the specified chat
`

func RevokeChatPermission(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) < 3 {
		fmt.Println(revokeChatUsage)
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

	telegramChatID, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		fatal("Invalid Telegram chat ID: %s", args[2])
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

	_, chatPermRepo, _, _, chatRepo, _ := initRepositories(db)

	ctx, cancel := getContext()
	defer cancel()

	// Look up the chat by bot_id and telegram_chat_id
	chat, err := chatRepo.GetByBotAndTelegramID(ctx, uint(botID), telegramChatID)
	if err != nil {
		fatal("Chat not found for bot %d and Telegram chat ID %d", botID, telegramChatID)
	}

	// Get existing permission
	perm, err := chatPermRepo.GetByAPIKeyAndChat(ctx, uint(apiKeyID), chat.ID)
	if err != nil {
		fatal("Chat permission not found")
	}

	// Delete permission
	if err := chatPermRepo.Delete(ctx, perm.ID); err != nil {
		fatal("Failed to revoke chat permission: %v", err)
	}

	success("Revoked chat permission for API key %d on chat %d (Telegram ID: %d)", apiKeyID, chat.ID, telegramChatID)
}
