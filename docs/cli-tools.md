# CLI Tools Reference

This document covers all command-line tools available in the Telegram Bot Gateway project.

## Overview

The gateway provides several CLI tools for administration and management tasks. These tools follow a security-first design: sensitive operations that involve credentials (bot tokens, API keys) are CLI-only and never exposed via HTTP endpoints. This ensures that secrets never transit the network, reducing attack surface and requiring server access for privileged operations.

### Security Rationale

Moving sensitive management operations to CLI provides several security benefits:

1. **No Network Exposure**: Tokens and credentials are processed locally on the server
2. **Privilege Separation**: Administrative operations require shell access
3. **Reduced Attack Surface**: Fewer HTTP endpoints that could be exploited
4. **Better Audit Trail**: CLI operations are easier to log and monitor
5. **MITM Protection**: Credentials never transmitted over network connections

### Configuration

All CLI tools use the same configuration file as the gateway server. By default, they look for `configs/config.json` in the project root. You can override this path using the `CONFIG_PATH` environment variable:

```bash
export CONFIG_PATH=/path/to/config.json
./bin/apikey list
```

This applies to all CLI tools: `bot`, `apikey`, `createuser`, and `migrate`.

## Building CLI Tools

Build all CLI tools using the Makefile:

```bash
make build
```

Or build individual tools:

```bash
# Build all tools
go build -o bin/gateway cmd/gateway/main.go
go build -o bin/bot cmd/bot/main.go
go build -o bin/apikey cmd/apikey/main.go
go build -o bin/createuser cmd/createuser/main.go
go build -o bin/migrate cmd/migrate/main.go

# Build specific tool
go build -o bin/apikey cmd/apikey/main.go
```

## gateway - HTTP/gRPC Server

The main gateway server that handles HTTP and gRPC requests.

### Usage

```bash
./bin/gateway
```

### Features

- HTTP REST API server
- gRPC server (can run on shared port with HTTP or separate port)
- WebSocket support for real-time message streaming
- Redis-backed rate limiting and pub/sub
- Graceful shutdown on SIGINT/SIGTERM

### Configuration

The gateway reads from `configs/config.json` and supports:

- Shared port mode (HTTP and gRPC on same port using cmux)
- Separate ports for HTTP and gRPC
- Configurable timeouts, rate limits, and worker counts
- WebSocket and webhook delivery settings

### Server Modes

The gateway can run in two modes:

**Shared Port Mode** (default):
```json
{
  "server": {
    "use_shared_port": true,
    "address": ":8080"
  }
}
```

Both HTTP and gRPC run on the same port, multiplexed using cmux.

**Separate Ports Mode**:
```json
{
  "server": {
    "use_shared_port": false,
    "http": {
      "address": ":8080"
    },
    "grpc": {
      "address": ":9090"
    }
  }
}
```

HTTP and gRPC run on separate ports.

### Health Check

The gateway exposes a health check endpoint:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": 1707742800,
  "websocket_clients": 5
}
```

### Graceful Shutdown

Send SIGINT (Ctrl+C) or SIGTERM to trigger graceful shutdown:

1. Stops accepting new connections
2. Cancels background workers (WebSocket hub, webhook workers)
3. Waits up to 30 seconds for in-flight requests to complete
4. Closes database and Redis connections

## bot - Bot Management

Manage Telegram bots and webhook registration. Bot tokens grant full control over Telegram bots, so creation and deletion are CLI-only for security.

### Security Features

- Bot tokens never transmitted over network
- Automatic webhook registration with cryptographically random secrets
- Webhook URLs use unguessable 64-character random paths
- Automatic webhook deregistration on bot deletion

### Commands

#### create

Create a new bot and automatically register webhook with Telegram.

```bash
./bin/bot create --username <username> --token <token> [options]
```

Options:
- `--username <name>` (required): Bot username (without @)
- `--token <token>` (required): Bot token from @BotFather
- `--display-name <name>`: Human-readable display name
- `--description <text>`: Bot description

Example:
```bash
./bin/bot create \
  --username my_bot \
  --token "123456:ABC-DEF1234567890abcdefghijklmnopqrstuvwxyz" \
  --display-name "My Bot" \
  --description "Production notification bot"
```

Output:
```
✓ Bot created successfully

Bot Details:
  ID:           1
  Username:     my_bot
  Display Name: My Bot
  Webhook URL:  https://example.com/api/v1/telegram/webhook/a1b2c3d4e5f6...

The webhook has been registered with Telegram.
```

Important: The full webhook URL is displayed only once. If webhook registration fails, the bot record is automatically rolled back.

#### list

List all bots.

```bash
./bin/bot list
```

Example output:
```
ID | Username         | Display Name     | Active | Webhook URL
---|------------------|------------------|--------|------------------
1  | my_bot           | My Bot           | Yes    | https://example.com/api/v1/telegram/webhook/a1b2c3d4...
2  | alerts_bot       | Alert Bot        | Yes    | https://example.com/api/v1/telegram/webhook/f7e8d9c0...
```

#### get

Get detailed information about a specific bot.

```bash
./bin/bot get <id>
```

Example:
```bash
./bin/bot get 1
```

#### update

Update bot metadata (display name, description, active status).

```bash
./bin/bot update <id> [options]
```

Options:
- `--display-name <name>`: New display name
- `--description <text>`: New description
- `--active <true|false>`: Set active status

Example:
```bash
./bin/bot update 1 --display-name "Production Bot" --active true
```

Note: This does not update the bot token. To change tokens, delete and recreate the bot.

#### delete

Delete a bot and deregister webhook from Telegram.

```bash
./bin/bot delete <id> --force
```

The `--force` flag is required to confirm deletion. Without it, the command displays a warning:

```bash
./bin/bot delete 1

# Output:
WARNING: Deleting a bot will:
  - Delete the bot record
  - CASCADE delete all associated chats
  - CASCADE delete all associated permissions
  - Deregister the webhook from Telegram

To proceed, add the --force flag:
  bot delete 1 --force
```

With `--force`:
```bash
./bin/bot delete 1 --force

# Output:
Deleting bot and deregistering webhook from Telegram...
✓ Bot deleted successfully
```

#### show-token

Display the decrypted bot token (requires server access).

```bash
./bin/bot show-token <id>
```

Example:
```bash
./bin/bot show-token 1

# Output:
Bot Token (ID: 1):
  123456:ABC-DEF1234567890abcdefghijklmnopqrstuvwxyz

⚠️  Keep this token secure! Anyone with this token can control your bot.
```

Use this when you need to retrieve a token for manual operations or troubleshooting.

## apikey - API Key Management

Manage API keys with granular permissions. API key management is CLI-only for security reasons.

### Permission Model

The API key system supports three types of permissions:

1. **Chat Permissions**: Control which chats an API key can read/send/manage (default: no access)
2. **Bot Restrictions**: Limit which bots the API key can use (default: all bots allowed)
3. **Feedback Control**: Restrict which chats can push messages to the API key holder (default: all chats allowed)

### Permission Logic

**Chat Permissions** (explicit grant, default deny):
- No chat permissions = No access to any chats
- Has chat permissions = Can only access explicitly granted chats

**Bot Restrictions** (opt-in restriction, default allow):
- No bot permissions = Can use all bots
- Has bot permissions = Can ONLY use explicitly allowed bots

**Feedback Permissions** (opt-in restriction, default allow):
- No feedback permissions = Can receive from all chats
- Has feedback permissions = Can ONLY receive from explicitly allowed chats

### Commands

#### create

Create a new API key.

```bash
./bin/apikey create --name <name> [options]
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

Important: The full API key (with prefix `tgw_...`) is displayed only once during creation. Save it securely.

#### list

List all API keys.

```bash
./bin/apikey list [--format table|json]
```

Examples:
```bash
./bin/apikey list
./bin/apikey list --format json
```

#### get

Get detailed information about a specific API key.

```bash
./bin/apikey get <id>
```

Example:
```bash
./bin/apikey get 1
```

#### revoke

Deactivate an API key without deleting it.

```bash
./bin/apikey revoke <id>
```

Example:
```bash
./bin/apikey revoke 1
```

Revoked keys can no longer authenticate but remain in the database for audit purposes.

#### delete

Permanently delete an API key and all associated permissions.

```bash
./bin/apikey delete <id>
```

Example:
```bash
./bin/apikey delete 1
```

Warning: This action cannot be undone. All permissions are CASCADE deleted.

### Permission Management

#### grant-chat

Grant an API key access to a specific chat.

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

#### revoke-chat

Revoke all permissions for a specific chat.

```bash
./bin/apikey revoke-chat <apikey-id> <chat-id>
```

Example:
```bash
./bin/apikey revoke-chat 1 5
```

#### grant-bot

Allow an API key to use a specific bot. Once you grant any bot permission, the API key becomes restricted to only the explicitly allowed bots.

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

#### revoke-bot

Revoke permission to use a specific bot.

```bash
./bin/apikey revoke-bot <apikey-id> <bot-id>
```

Example:
```bash
./bin/apikey revoke-bot 1 2
```

#### grant-feedback

Allow an API key to receive feedback from a specific chat. Once you grant any feedback permission, the API key becomes restricted to only receive from explicitly allowed chats.

```bash
./bin/apikey grant-feedback <apikey-id> <chat-id>
```

Example:
```bash
# Allow API key 1 to receive messages from chat 5
./bin/apikey grant-feedback 1 5

# Now API key 1 can ONLY receive feedback from chat 5
```

#### revoke-feedback

Revoke feedback permission for a specific chat.

```bash
./bin/apikey revoke-feedback <apikey-id> <chat-id>
```

Example:
```bash
./bin/apikey revoke-feedback 1 5
```

#### show-permissions

Display all permissions for an API key.

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

### Common Workflows

#### Create API key for external service

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

#### Create read-only monitoring key

```bash
# 1. Create the API key
./bin/apikey create --name "Monitoring" --rate-limit 10000

# 2. Grant read-only access to multiple chats
./bin/apikey grant-chat 1 5 --read
./bin/apikey grant-chat 1 8 --read
./bin/apikey grant-chat 1 10 --read

# 3. No bot restrictions needed (monitoring doesn't send)
```

#### Create webhook receiver key

```bash
# 1. Create the API key
./bin/apikey create --name "Webhook Receiver" --rate-limit 1000

# 2. Grant send permission for notifications
./bin/apikey grant-chat 1 10 --send

# 3. Allow feedback from specific groups
./bin/apikey grant-feedback 1 5
./bin/apikey grant-feedback 1 8
```

### Security Best Practices

1. **Principle of Least Privilege**: Only grant the minimum permissions needed
2. **Use Expiration Dates**: Set expiration dates for temporary keys
3. **Regular Audits**: Periodically review API keys with `./bin/apikey list`
4. **Revoke Unused Keys**: Delete or revoke keys that are no longer needed
5. **Bot Restrictions**: Limit API keys to specific bots in production
6. **Rate Limiting**: Set appropriate rate limits based on expected usage

## createuser - User Account Creation

Create user accounts for JWT-based authentication.

### Usage

```bash
./bin/createuser --username <username> --password <password> [--email <email>]
```

Options:
- `--username <name>`: Username (default: "admin")
- `--password <pass>`: Password (required)
- `--email <email>`: Email address (optional)

Example:
```bash
./bin/createuser --username admin --password "SecurePassword123" --email admin@example.com
```

Output:
```
✓ Created user: admin (ID: 1)
✓ Now assign admin role manually:
   INSERT INTO user_roles (user_id, role_id) VALUES (1, 1);
```

### Role Assignment

After creating a user, you need to manually assign roles via SQL:

```sql
-- Assign admin role (role_id 1) to user
INSERT INTO user_roles (user_id, role_id) VALUES (1, 1);
```

This is intentional to prevent accidental privilege escalation.

## migrate - Database Migrations

Run database migrations to initialize or upgrade the schema.

### Usage

```bash
./bin/migrate <command>
```

Commands:
- `up`: Apply all pending migrations
- `down`: Rollback migrations

### Migration Files

Migrations are located in `migrations/` directory:

- `001_initial_schema.sql`: Base schema (users, bots, chats, messages, API keys)
- `003_bot_webhook_secret.sql`: Add webhook_secret column for secure webhook URLs

### Examples

Apply all migrations:
```bash
./bin/migrate up
```

Output:
```
Running migration: migrations/001_initial_schema.sql
Running migration: migrations/003_bot_webhook_secret.sql
Migration completed successfully
```

Rollback migrations:
```bash
./bin/migrate down
```

Output:
```
Rollback completed successfully
```

### Configuration

The migrate tool uses the same database configuration as the gateway. Set `CONFIG_PATH` to use a different config file:

```bash
CONFIG_PATH=/path/to/config.json ./bin/migrate up
```

### First-Time Setup

For a new installation:

```bash
# 1. Create database
mysql -u root -p -e "CREATE DATABASE telegram_gateway;"

# 2. Configure database credentials in configs/config.json

# 3. Run migrations
./bin/migrate up

# 4. Create admin user
./bin/createuser --username admin --password "YourSecurePassword"

# 5. Assign admin role
mysql -u root -p telegram_gateway -e "INSERT INTO user_roles (user_id, role_id) VALUES (1, 1);"
```

## Troubleshooting

### "Failed to connect to database"

Check that:
- `configs/config.json` exists and has correct database credentials
- Database server is running and accessible
- Database exists (create with `CREATE DATABASE telegram_gateway;`)

### "Failed to load config"

Ensure `CONFIG_PATH` points to a valid config file, or that `configs/config.json` exists in the current directory.

### "API key not authorized for this bot"

Check bot restrictions:
```bash
./bin/apikey show-permissions <id>
```

Grant bot permission:
```bash
./bin/apikey grant-bot <apikey-id> <bot-id>
```

### "Insufficient permissions for this chat"

Check chat permissions:
```bash
./bin/apikey show-permissions <id>
```

Grant chat permission:
```bash
./bin/apikey grant-chat <apikey-id> <chat-id> --read --send
```

### Migration fails with "table already exists"

This is normal if running migrations that were already applied. The migrate tool does not track migration state. To re-run migrations on an existing database, either:

1. Skip already-applied migrations (modify migration list in `cmd/migrate/main.go`)
2. Drop and recreate the database (WARNING: data loss)

## Automation and Scripting

All CLI tools are designed to be scriptable:

### Export API keys

```bash
./bin/apikey list --format json > apikeys.json
```

### Automated key creation

```bash
KEY_ID=$(./bin/apikey create --name "Auto-Created" | grep "API Key ID" | awk '{print $4}')
./bin/apikey grant-chat $KEY_ID 5 --read --send
```

### Audit all permissions

```bash
for id in $(./bin/apikey list --format json | jq '.[].id'); do
  ./bin/apikey show-permissions $id
done
```

### Remote execution via SSH

```bash
ssh user@server "cd /path/to/gateway && ./bin/apikey list"
```

### Configuration management integration

The CLI tools work well with Ansible, Terraform, or other configuration management tools:

```yaml
# Ansible example
- name: Create API key
  command: /opt/gateway/bin/apikey create --name "{{ service_name }}" --rate-limit 5000
  register: apikey_result

- name: Grant permissions
  command: /opt/gateway/bin/apikey grant-chat {{ apikey_id }} {{ chat_id }} --read --send
```

## Summary

The Telegram Bot Gateway CLI tools provide secure, server-side management of sensitive resources:

- **gateway**: Main HTTP/gRPC server
- **bot**: Create, manage, and configure Telegram bots
- **apikey**: Create and manage API keys with granular permissions
- **createuser**: Create user accounts for JWT authentication
- **migrate**: Run database migrations

All tools share the same configuration file and follow consistent patterns for flags and output. Sensitive operations (bot tokens, API keys) are CLI-only, ensuring credentials never transit the network.
