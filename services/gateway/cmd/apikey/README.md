# API Key Management CLI

A command-line tool for managing API keys with granular permissions in the Telegram Bot Gateway.

## Overview

The API key management system supports three types of granular permissions:

1. **Chat Permissions**: Control which chats an API key can read/send/manage
2. **Bot Restrictions**: Limit which bot(s) the API key can use for sending messages
3. **Feedback Control**: Restrict which chats can push messages back to the API key holder

## Installation

Build the CLI tool:

```bash
go build -o bin/apikey cmd/apikey/main.go
```

## Configuration

The CLI tool uses the same configuration file as the gateway (`configs/config.json` by default). You can override this with the `CONFIG_PATH` environment variable:

```bash
export CONFIG_PATH=/path/to/config.json
./bin/apikey list
```

## Commands

### Create an API Key

```bash
./bin/apikey create --name "Production Service" [options]
```

Options:
- `--name <name>` (required): Name for the API key
- `--description <desc>`: Description
- `--rate-limit <n>`: Requests per hour (default: 1000)
- `--expires <duration>`: Expiration time (e.g., `1y`, `30d`, `24h`)

Example:
```bash
./bin/apikey create --name "Production Service" --rate-limit 5000 --expires 1y
```

⚠️ **Important**: The full API key is displayed only once during creation. Save it securely!

### List API Keys

```bash
./bin/apikey list [--format table|json]
```

Example:
```bash
./bin/apikey list
./bin/apikey list --format json
```

### Get API Key Details

```bash
./bin/apikey get <id>
```

Example:
```bash
./bin/apikey get 1
```

### Revoke an API Key

Deactivates an API key without deleting it:

```bash
./bin/apikey revoke <id>
```

Example:
```bash
./bin/apikey revoke 1
```

### Delete an API Key

Permanently deletes an API key and all associated permissions:

```bash
./bin/apikey delete <id>
```

Example:
```bash
./bin/apikey delete 1
```

⚠️ **Warning**: This action cannot be undone.

## Permission Management

### Chat Permissions

Grant an API key access to specific chats:

```bash
./bin/apikey grant-chat <apikey-id> <chat-id> [--read] [--send] [--manage]
```

Flags:
- `--read`: Allow reading messages from the chat
- `--send`: Allow sending messages to the chat
- `--manage`: Allow managing the chat (admin operations)

Examples:
```bash
# Read and send access
./bin/apikey grant-chat 1 5 --read --send

# Read-only access
./bin/apikey grant-chat 1 8 --read

# Full access
./bin/apikey grant-chat 1 10 --read --send --manage
```

Revoke chat permissions:
```bash
./bin/apikey revoke-chat <apikey-id> <chat-id>
```

### Bot Restrictions

By default, API keys can use **all bots**. Once you grant bot permission, the API key becomes **restricted** to only the explicitly allowed bots.

Allow an API key to use a specific bot:

```bash
./bin/apikey grant-bot <apikey-id> <bot-id>
```

Example:
```bash
# Allow API key 1 to use bot 2
./bin/apikey grant-bot 1 2

# Allow API key 1 to also use bot 3
./bin/apikey grant-bot 1 3

# Now API key 1 can ONLY use bots 2 and 3
```

Revoke bot permission:
```bash
./bin/apikey revoke-bot <apikey-id> <bot-id>
```

### Feedback Permissions

By default, API keys can receive feedback messages from **all chats**. Once you grant feedback permission, the API key becomes **restricted** to only receive from explicitly allowed chats.

Allow an API key to receive feedback from a chat:

```bash
./bin/apikey grant-feedback <apikey-id> <chat-id>
```

Example:
```bash
# Allow API key 1 to receive messages from chat 5
./bin/apikey grant-feedback 1 5

# Now API key 1 can ONLY receive feedback from chat 5
```

Revoke feedback permission:
```bash
./bin/apikey revoke-feedback <apikey-id> <chat-id>
```

### View All Permissions

Display all permissions for an API key:

```bash
./bin/apikey show-permissions <apikey-id>
```

Example output:
```
Permissions for API Key: Production Service (tgw_abc123...)
============================================================

Chat Permissions:
  Chat ID | Read  | Send  | Manage | Chat Info
  --------|-------|-------|--------|------------------
  5       | Yes   | Yes   | No     | Development Group
  8       | Yes   | No    | No     | Notifications

Bot Restrictions (can ONLY use these bots):
  Bot ID | Username
  -------|------------------
  2      | mybot_prod

Feedback Permissions: None (can receive from ALL chats)
```

## Permission Logic

### Chat Permissions
- **No chat permissions** = No access to any chats
- **Has chat permissions** = Can only access explicitly granted chats

### Bot Restrictions
- **No bot permissions** = Can use **ALL** bots (default)
- **Has bot permissions** = Can **ONLY** use explicitly allowed bots

### Feedback Permissions
- **No feedback permissions** = Can receive from **ALL** chats (default)
- **Has feedback permissions** = Can **ONLY** receive from explicitly allowed chats

## Common Workflows

### Create API key for external service

```bash
# 1. Create the API key
./bin/apikey create --name "External Service" --rate-limit 5000 --expires 1y

# 2. Grant access to specific chats
./bin/apikey grant-chat 1 5 --read --send
./bin/apikey grant-chat 1 8 --read

# 3. Restrict to specific bot
./bin/apikey grant-bot 1 2

# 4. Verify permissions
./bin/apikey show-permissions 1
```

### Create read-only monitoring key

```bash
# 1. Create the API key
./bin/apikey create --name "Monitoring" --rate-limit 10000

# 2. Grant read-only access to multiple chats
./bin/apikey grant-chat 1 5 --read
./bin/apikey grant-chat 1 8 --read
./bin/apikey grant-chat 1 10 --read

# 3. No bot restrictions needed (monitoring doesn't send)
```

### Create webhook receiver key

```bash
# 1. Create the API key
./bin/apikey create --name "Webhook Receiver" --rate-limit 1000

# 2. Grant send permission for notifications
./bin/apikey grant-chat 1 10 --send

# 3. Allow feedback from specific groups
./bin/apikey grant-feedback 1 5
./bin/apikey grant-feedback 1 8
```

## Security Best Practices

1. **Principle of Least Privilege**: Only grant the minimum permissions needed
2. **Use Expiration Dates**: Set expiration dates for temporary keys
3. **Regular Audits**: Periodically review API keys with `./bin/apikey list`
4. **Revoke Unused Keys**: Delete or revoke keys that are no longer needed
5. **Bot Restrictions**: Limit API keys to specific bots in production
6. **Rate Limiting**: Set appropriate rate limits based on expected usage

## Migration from REST API

If you were previously using the REST API for API key management, note that:

1. **REST endpoints are disabled** in the gateway
2. **All management is CLI-only** for security
3. **Existing API keys continue to work** without changes
4. **New granular permissions** are additive - existing keys have default behavior

## Troubleshooting

### "Failed to connect to database"
- Check that `configs/config.json` exists and has correct database credentials
- Ensure the database is running and accessible

### "API key not authorized for this bot"
- Check bot restrictions with `./bin/apikey show-permissions <id>`
- Grant bot permission with `./bin/apikey grant-bot <apikey-id> <bot-id>`

### "Insufficient permissions for this chat"
- Check chat permissions with `./bin/apikey show-permissions <id>`
- Grant chat permission with `./bin/apikey grant-chat <apikey-id> <chat-id> --read --send`

## Examples

See the [examples directory](../../examples/apikey/) for complete usage examples.
