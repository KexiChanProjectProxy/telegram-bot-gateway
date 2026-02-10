package commands

import (
	"fmt"
	"strconv"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
)

const grantChatUsage = `Grant chat permissions to an API key

Usage:
  apikey grant-chat <apikey-id> <bot-id> <telegram-chat-id> [--read] [--send] [--manage]

Arguments:
  <apikey-id>               API key ID
  <bot-id>                  Bot ID
  <telegram-chat-id>        Telegram chat ID (the actual Telegram chat identifier)

Options:
  --read                    Grant read permission
  --send                    Grant send permission
  --manage                  Grant manage permission
  --help, -h                Show this help message

Examples:
  apikey grant-chat 1 1 1878878763 --read --send
  apikey grant-chat 1 1 -1001234567890 --read
  apikey grant-chat 1 2 987654321 --read --send --manage

Note: If no permissions are specified, no permissions will be granted.
      You must explicitly specify at least one permission flag.
      The telegram-chat-id is the actual chat ID from Telegram, not the internal database ID.
`

func GrantChatPermission(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) < 3 {
		fmt.Println(grantChatUsage)
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

	canRead := hasFlag(args, "--read")
	canSend := hasFlag(args, "--send")
	canManage := hasFlag(args, "--manage")

	if !canRead && !canSend && !canManage {
		fatal("No permissions specified. Use --read, --send, or --manage")
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
		fatal("Chat not found for bot %d and Telegram chat ID %d. Make sure the bot has received at least one message from this chat.", botID, telegramChatID)
	}

	// Check if permission already exists
	existingPerm, err := chatPermRepo.GetByAPIKeyAndChat(ctx, uint(apiKeyID), chat.ID)
	if err == nil {
		// Update existing permission
		existingPerm.CanRead = existingPerm.CanRead || canRead
		existingPerm.CanSend = existingPerm.CanSend || canSend
		existingPerm.CanManage = existingPerm.CanManage || canManage

		if err := chatPermRepo.Update(ctx, existingPerm); err != nil {
			fatal("Failed to update chat permission: %v", err)
		}
		success("Updated chat permission for API key %d on chat %d (Telegram ID: %d)", apiKeyID, chat.ID, telegramChatID)
	} else {
		// Create new permission
		akID := uint(apiKeyID)
		perm := &domain.ChatPermission{
			ChatID:    chat.ID,
			APIKeyID:  &akID,
			CanRead:   canRead,
			CanSend:   canSend,
			CanManage: canManage,
		}

		if err := chatPermRepo.Create(ctx, perm); err != nil {
			fatal("Failed to grant chat permission: %v", err)
		}
		success("Granted chat permission for API key %d on chat %d (Telegram ID: %d)", apiKeyID, chat.ID, telegramChatID)
	}

	var perms []string
	if canRead {
		perms = append(perms, "read")
	}
	if canSend {
		perms = append(perms, "send")
	}
	if canManage {
		perms = append(perms, "manage")
	}
	info("Permissions: %v", perms)
}
