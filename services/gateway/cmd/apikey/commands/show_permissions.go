package commands

import (
	"fmt"
	"strconv"
)

const showPermissionsUsage = `Show all permissions for an API key

Usage:
  apikey show-permissions <apikey-id>

Arguments:
  <apikey-id>               API key ID

Options:
  --help, -h                Show this help message

Examples:
  apikey show-permissions 1
`

func ShowPermissions(args []string) {
	if hasFlag(args, "--help", "-h") || len(args) == 0 {
		fmt.Println(showPermissionsUsage)
		return
	}

	apiKeyID, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		fatal("Invalid API key ID: %s", args[0])
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

	apiKeyRepo, chatPermRepo, botPermRepo, feedbackPermRepo, _, _ := initRepositories(db)

	ctx, cancel := getContext()
	defer cancel()

	// Get API key details
	apiKey, err := apiKeyRepo.GetByID(ctx, uint(apiKeyID))
	if err != nil {
		fatal("Failed to get API key: %v", err)
	}

	info("Permissions for API Key: %s (%s)", apiKey.Name, apiKey.Key)
	info("=" + repeatString("=", 60))
	info("")

	// Chat permissions
	chatPerms, err := chatPermRepo.ListByAPIKey(ctx, uint(apiKeyID))
	if err == nil && len(chatPerms) > 0 {
		info("Chat Permissions:")
		info("  Chat ID | Read  | Send  | Manage | Chat Info")
		info("  --------|-------|-------|--------|------------------")
		for _, perm := range chatPerms {
			read := boolToYesNo(perm.CanRead)
			send := boolToYesNo(perm.CanSend)
			manage := boolToYesNo(perm.CanManage)
			chatInfo := fmt.Sprintf("ID: %d", perm.ChatID)
			if perm.Chat.Title != "" {
				chatInfo = perm.Chat.Title
			}
			info("  %-7d | %-5s | %-5s | %-6s | %s", perm.ChatID, read, send, manage, chatInfo)
		}
		info("")
	} else {
		info("Chat Permissions: None (no access to any chats)")
		info("")
	}

	// Bot permissions
	botPerms, err := botPermRepo.ListByAPIKey(ctx, uint(apiKeyID))
	if err == nil && len(botPerms) > 0 {
		info("Bot Restrictions (can ONLY use these bots):")
		info("  Bot ID | Username")
		info("  -------|------------------")
		for _, perm := range botPerms {
			info("  %-6d | %s", perm.BotID, perm.Bot.Username)
		}
		info("")
	} else {
		info("Bot Restrictions: None (can use ALL bots)")
		info("")
	}

	// Feedback permissions
	feedbackPerms, err := feedbackPermRepo.ListByAPIKey(ctx, uint(apiKeyID))
	if err == nil && len(feedbackPerms) > 0 {
		info("Feedback Permissions (can ONLY receive from these chats):")
		info("  Chat ID | Chat Info")
		info("  --------|------------------")
		for _, perm := range feedbackPerms {
			chatInfo := fmt.Sprintf("ID: %d", perm.ChatID)
			if perm.Chat.Title != "" {
				chatInfo = perm.Chat.Title
			}
			info("  %-7d | %s", perm.ChatID, chatInfo)
		}
		info("")
	} else {
		info("Feedback Permissions: None (can receive from ALL chats)")
		info("")
	}

	info("Command examples:")
	info("  Grant chat access:     apikey grant-chat %d <chat-id> --read --send", apiKeyID)
	info("  Restrict to bot:       apikey grant-bot %d <bot-id>", apiKeyID)
	info("  Allow feedback:        apikey grant-feedback %d <chat-id>", apiKeyID)
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
