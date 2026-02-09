package commands

import (
	"fmt"
	"strconv"

	"github.com/kexi/telegram-bot-gateway/internal/domain"
)

const grantChatUsage = `Grant chat permissions to an API key

Usage:
  apikey grant-chat <apikey-id> <chat-id> [--read] [--send] [--manage]

Arguments:
  <apikey-id>               API key ID
  <chat-id>                 Chat ID

Options:
  --read                    Grant read permission
  --send                    Grant send permission
  --manage                  Grant manage permission
  --help, -h                Show this help message

Examples:
  apikey grant-chat 1 5 --read --send
  apikey grant-chat 1 8 --read
  apikey grant-chat 1 10 --read --send --manage

Note: If no permissions are specified, no permissions will be granted.
      You must explicitly specify at least one permission flag.
`

func GrantChatPermission(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) < 2 {
		fmt.Println(grantChatUsage)
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

	_, chatPermRepo, _, _, _, _ := initRepositories(db)

	ctx, cancel := getContext()
	defer cancel()

	// Check if permission already exists
	existingPerm, err := chatPermRepo.GetByAPIKeyAndChat(ctx, uint(apiKeyID), uint(chatID))
	if err == nil {
		// Update existing permission
		existingPerm.CanRead = existingPerm.CanRead || canRead
		existingPerm.CanSend = existingPerm.CanSend || canSend
		existingPerm.CanManage = existingPerm.CanManage || canManage

		if err := chatPermRepo.Update(ctx, existingPerm); err != nil {
			fatal("Failed to update chat permission: %v", err)
		}
		success("Updated chat permission for API key %d on chat %d", apiKeyID, chatID)
	} else {
		// Create new permission
		akID := uint(apiKeyID)
		perm := &domain.ChatPermission{
			ChatID:    uint(chatID),
			APIKeyID:  &akID,
			CanRead:   canRead,
			CanSend:   canSend,
			CanManage: canManage,
		}

		if err := chatPermRepo.Create(ctx, perm); err != nil {
			fatal("Failed to grant chat permission: %v", err)
		}
		success("Granted chat permission for API key %d on chat %d", apiKeyID, chatID)
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
