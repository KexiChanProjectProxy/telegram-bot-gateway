package main

import (
	"fmt"
	"os"

	"github.com/kexi/telegram-bot-gateway/cmd/apikey/commands"
)

const usage = `API Key Management CLI

Usage:
  apikey <command> [arguments]

Commands:
  create                  Create a new API key
  list                    List all API keys
  get <id>                Get API key details
  revoke <id>             Revoke (deactivate) an API key
  delete <id>             Delete an API key

  grant-chat <apikey-id> <chat-id>     Grant chat permissions
  revoke-chat <apikey-id> <chat-id>    Revoke chat permissions

  grant-bot <apikey-id> <bot-id>       Allow API key to use a bot
  revoke-bot <apikey-id> <bot-id>      Disallow API key from using a bot

  grant-feedback <apikey-id> <chat-id> Allow feedback from chat
  revoke-feedback <apikey-id> <chat-id> Disallow feedback from chat

  show-permissions <apikey-id>         Show all permissions for an API key

Options:
  --help, -h              Show this help message

Examples:
  # Create an API key
  apikey create --name "Production Service" --rate-limit 5000 --expires 1y

  # List all API keys
  apikey list

  # Grant chat permissions
  apikey grant-chat 1 5 --read --send

  # Restrict to specific bot
  apikey grant-bot 1 2

  # Show all permissions
  apikey show-permissions 1
`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		commands.CreateAPIKey(os.Args[2:])
	case "list", "ls":
		commands.ListAPIKeys(os.Args[2:])
	case "get", "show":
		commands.GetAPIKey(os.Args[2:])
	case "revoke":
		commands.RevokeAPIKey(os.Args[2:])
	case "delete", "rm":
		commands.DeleteAPIKey(os.Args[2:])
	case "grant-chat":
		commands.GrantChatPermission(os.Args[2:])
	case "revoke-chat":
		commands.RevokeChatPermission(os.Args[2:])
	case "grant-bot":
		commands.GrantBotPermission(os.Args[2:])
	case "revoke-bot":
		commands.RevokeBotPermission(os.Args[2:])
	case "grant-feedback":
		commands.GrantFeedbackPermission(os.Args[2:])
	case "revoke-feedback":
		commands.RevokeFeedbackPermission(os.Args[2:])
	case "show-permissions", "permissions":
		commands.ShowPermissions(os.Args[2:])
	case "help", "--help", "-h":
		fmt.Println(usage)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		fmt.Println(usage)
		os.Exit(1)
	}
}
