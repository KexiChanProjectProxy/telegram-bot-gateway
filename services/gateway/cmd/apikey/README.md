# API Key Management CLI

Command-line tool for managing API keys with granular permissions in the Telegram Bot Gateway.

For complete documentation, security rationale, and detailed examples, see [docs/cli-tools.md](../../../../docs/cli-tools.md).

## Quick Start

Build the tool:
```bash
go build -o bin/apikey cmd/apikey/main.go
```

Set config path (optional):
```bash
export CONFIG_PATH=/path/to/config.json
```

## Command Reference

| Command | Usage | Description |
|---------|-------|-------------|
| **create** | `./bin/apikey create --name <name> [options]` | Create new API key |
| **list** | `./bin/apikey list [--format table\|json]` | List all API keys |
| **get** | `./bin/apikey get <id>` | Get API key details |
| **revoke** | `./bin/apikey revoke <id>` | Deactivate API key |
| **delete** | `./bin/apikey delete <id>` | Delete API key permanently |
| **show-permissions** | `./bin/apikey show-permissions <id>` | Display all permissions |

### Create Options

| Flag | Description | Default |
|------|-------------|---------|
| `--name <name>` | Name for the API key (required) | - |
| `--description <desc>` | Description | - |
| `--rate-limit <n>` | Requests per hour | 1000 |
| `--expires <duration>` | Expiration time (e.g., `1y`, `30d`, `24h`) | - |

### Permission Commands

| Command | Usage | Description |
|---------|-------|-------------|
| **grant-chat** | `./bin/apikey grant-chat <key-id> <chat-id> [--read] [--send] [--manage]` | Grant chat access |
| **revoke-chat** | `./bin/apikey revoke-chat <key-id> <chat-id>` | Revoke chat access |
| **grant-bot** | `./bin/apikey grant-bot <key-id> <bot-id>` | Allow bot usage |
| **revoke-bot** | `./bin/apikey revoke-bot <key-id> <bot-id>` | Revoke bot usage |
| **grant-feedback** | `./bin/apikey grant-feedback <key-id> <chat-id>` | Allow feedback from chat |
| **revoke-feedback** | `./bin/apikey revoke-feedback <key-id> <chat-id>` | Revoke feedback permission |

## Permission Model Summary

- **Chat Permissions**: Default deny (must grant explicitly)
- **Bot Restrictions**: Default allow all (grant to restrict)
- **Feedback Permissions**: Default allow all (grant to restrict)

## Example

```bash
# Create API key
./bin/apikey create --name "Production" --rate-limit 5000 --expires 1y

# Grant permissions
./bin/apikey grant-chat 1 5 --read --send
./bin/apikey grant-bot 1 2

# View all permissions
./bin/apikey show-permissions 1
```

For complete documentation including workflows, security best practices, and troubleshooting, see [docs/cli-tools.md](../../../../docs/cli-tools.md).
