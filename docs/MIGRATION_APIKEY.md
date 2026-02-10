# Migration Guide: API Key Management Changes

## What Changed

### Security Improvement: CLI-Only API Key Management

As of this version, API key management has been moved from REST API endpoints to a dedicated CLI tool for enhanced security.

### New Features

1. **CLI Tool for API Keys**: `./bin/apikey` command-line tool for all API key operations
2. **Granular Bot Permissions**: Restrict which bots an API key can use
3. **Feedback Control**: Control which chats can push messages to API key holders
4. **Enhanced Chat Permissions**: Existing chat-level permissions (read/send/manage)

## Breaking Changes

### REST API Endpoints Removed

The following REST API endpoints have been **disabled**:

```
POST   /api/v1/apikeys          # Create API key
GET    /api/v1/apikeys          # List API keys
GET    /api/v1/apikeys/:id      # Get API key
POST   /api/v1/apikeys/:id/revoke  # Revoke API key
DELETE /api/v1/apikeys/:id      # Delete API key
```

**Migration**: Use the CLI tool instead:

| Old REST API | New CLI Command |
|-------------|-----------------|
| `POST /api/v1/apikeys` | `./bin/apikey create --name "..."` |
| `GET /api/v1/apikeys` | `./bin/apikey list` |
| `GET /api/v1/apikeys/:id` | `./bin/apikey get <id>` |
| `POST /api/v1/apikeys/:id/revoke` | `./bin/apikey revoke <id>` |
| `DELETE /api/v1/apikeys/:id` | `./bin/apikey delete <id>` |

## Database Migration

Run the new migration to add permission tables:

```bash
# Apply migration
mysql -u username -p database_name < migrations/002_api_key_permissions.sql

# Rollback (if needed)
mysql -u username -p database_name < migrations/002_api_key_permissions_down.sql
```

### New Tables

- `api_key_bot_permissions`: Bot usage restrictions
- `api_key_feedback_permissions`: Feedback control

## CLI Tool Setup

### Build

```bash
go build -o bin/apikey cmd/apikey/main.go
```

### Basic Usage

```bash
# Create an API key
./bin/apikey create --name "Production" --rate-limit 5000 --expires 1y

# List all keys
./bin/apikey list

# Grant permissions
./bin/apikey grant-chat 1 5 --read --send
./bin/apikey grant-bot 1 2
./bin/apikey grant-feedback 1 5

# View all permissions
./bin/apikey show-permissions 1
```

See [cmd/apikey/README.md](cmd/apikey/README.md) for complete documentation.

## Permission Model

### Default Behavior (No Restrictions)

When an API key has no explicit permissions:

- **Chat access**: No access to any chats (must grant explicitly)
- **Bot usage**: Can use **all** bots
- **Feedback**: Can receive from **all** chats

### Restrictive Behavior (With Permissions)

Once you grant any permission of a type:

- **Bot permissions exist** → Can **ONLY** use explicitly allowed bots
- **Feedback permissions exist** → Can **ONLY** receive from explicitly allowed chats

### Example Scenarios

#### Scenario 1: Basic API Key
```bash
./bin/apikey create --name "Service A"
./bin/apikey grant-chat 1 5 --read --send
# Result: Can access chat 5, can use all bots, can receive from all chats
```

#### Scenario 2: Bot-Restricted Key
```bash
./bin/apikey create --name "Service B"
./bin/apikey grant-chat 1 5 --send
./bin/apikey grant-bot 1 2
# Result: Can send to chat 5 via bot 2 ONLY (no other bots allowed)
```

#### Scenario 3: Feedback-Restricted Key
```bash
./bin/apikey create --name "Service C"
./bin/apikey grant-chat 1 10 --send
./bin/apikey grant-feedback 1 5
# Result: Can send to chat 10, can receive feedback from chat 5 ONLY
```

## Updating Existing Deployments

### Step 1: Build New Binaries

```bash
go build -o bin/gateway cmd/gateway/main.go
go build -o bin/apikey cmd/apikey/main.go
```

### Step 2: Run Migration

```bash
mysql -u username -p database_name < migrations/002_api_key_permissions.sql
```

### Step 3: Deploy

```bash
# Stop gateway
systemctl stop telegram-gateway

# Replace binary
cp bin/gateway /usr/local/bin/telegram-gateway

# Start gateway
systemctl start telegram-gateway
```

### Step 4: Update Documentation

Update any internal documentation referencing the old REST API endpoints.

## Backward Compatibility

### Existing API Keys

- **Continue to work** without any changes
- **Chat permissions** remain intact
- **Default behavior** applies for bot and feedback permissions
- **No action required** unless you want to add restrictions

### Authentication

- API key authentication via `X-API-Key` header **unchanged**
- JWT authentication **unchanged**
- All other API endpoints **unchanged**

## Security Benefits

1. **Reduced Attack Surface**: API key management not exposed to network
2. **Privilege Separation**: Key management requires server access
3. **Audit Trail**: CLI operations easier to log and audit
4. **Granular Control**: Fine-grained bot and feedback restrictions
5. **Principle of Least Privilege**: Default-deny for chats, opt-in for restrictions

## FAQ

### Q: Can I still use existing API keys?
**A**: Yes, all existing API keys continue to work unchanged.

### Q: How do I create API keys now?
**A**: Use the CLI tool: `./bin/apikey create --name "..."`

### Q: What if I need remote API key management?
**A**: Use SSH or a configuration management tool (Ansible, Terraform, etc.) to run CLI commands remotely.

### Q: Can I automate API key creation?
**A**: Yes, the CLI tool can be scripted. Example:
```bash
KEY_ID=$(./bin/apikey create --name "Auto-Created" | grep "API Key ID" | awk '{print $4}')
./bin/apikey grant-chat $KEY_ID 5 --read --send
```

### Q: How do I list all permissions for auditing?
**A**: Use `./bin/apikey show-permissions <id>` for individual keys, or script it:
```bash
for id in $(./bin/apikey list --format json | jq '.[].id'); do
  ./bin/apikey show-permissions $id
done
```

### Q: Can I export permissions?
**A**: Yes, use JSON format:
```bash
./bin/apikey list --format json > apikeys.json
```

## Support

For issues or questions:
1. Check [cmd/apikey/README.md](cmd/apikey/README.md)
2. Review [examples/apikey/](examples/apikey/)
3. Open a GitHub issue

## Related Documentation

- [API Key CLI Reference](cmd/apikey/README.md)
- [Permission Model](docs/permissions.md)
- [Security Best Practices](docs/security.md)
