package main

import (
	"fmt"
	"os"

	"github.com/kexi/telegram-bot-gateway/cmd/bot/commands"
)

const usage = `Bot Management CLI

Usage:
  bot <command> [arguments] [flags]

Commands:
  create              Create a new bot and register webhook with Telegram
  list, ls            List all bots
  get, show <id>      Get bot details by ID
  update <id>         Update bot information
  delete, rm <id>     Delete a bot and deregister webhook
  show-token <id>     Display decrypted bot token

Flags for 'create':
  --username <name>        Bot username (required)
  --token <token>          Bot token from @BotFather (required)
  --display-name <name>    Display name (optional)
  --description <text>     Description (optional)

Flags for 'update':
  --display-name <name>    New display name
  --description <text>     New description
  --active <true|false>    Set active status

Flags for 'delete':
  --force                  Required to confirm deletion

Environment:
  CONFIG_PATH              Path to config file (default: configs/config.json)

Examples:
  bot create --username my_bot --token "123456:ABC-DEF" --display-name "My Bot"
  bot list
  bot get 1
  bot update 1 --display-name "New Name" --active true
  bot show-token 1
  bot delete 1 --force
`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "create":
		commands.Create(args)
	case "list", "ls":
		commands.List(args)
	case "get", "show":
		commands.Get(args)
	case "update":
		commands.Update(args)
	case "delete", "rm":
		commands.Delete(args)
	case "show-token":
		commands.ShowToken(args)
	case "help", "-h", "--help":
		fmt.Println(usage)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		fmt.Println(usage)
		os.Exit(1)
	}
}
