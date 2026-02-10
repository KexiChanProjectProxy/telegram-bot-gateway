# CLI-Based Bot Management - Implementation Summary

## Overview

Moved bot management (create/delete operations) from HTTP API to a dedicated CLI tool for enhanced security. Bot tokens grant full control over Telegram bots, so they should never transit the network.

**Date**: 2026-02-10
**Breaking Changes**: Yes - `POST /api/v1/bots` and `DELETE /api/v1/bots/:id` routes removed

---

## Security Improvements

### Before (HTTP API)
```bash
# ❌ Bot token sent over network
curl -X POST http://localhost:8080/api/v1/bots \
  -H "Authorization: Bearer $JWT" \
  -d '{"username": "my_bot", "token": "123456:ABC-DEF..."}'
```

**Risks:**
- Network exposure of bot tokens
- Potential MITM attacks
- Token logging in proxies/load balancers
- Wider attack surface

### After (CLI-Only)
```bash
# ✅ Bot token only processed locally
./bin/bot create --username my_bot --token "123456:ABC-DEF..."
```

**Benefits:**
- Requires server access (privilege separation)
- No network exposure
- Automatic webhook registration with random secret
- Better audit trail

---

## Webhook Secret Implementation

### Problem
Old webhook route used bot username in URL: `/telegram/webhook/:bot_username`

**Issues:**
- Username is guessable (public information)
- Anyone could send fake updates to `POST /telegram/webhook/my_bot`
- No validation beyond username lookup

### Solution
New webhook route uses random secret: `/telegram/webhook/:webhook_secret`

**Implementation:**
1. On bot creation, generate 32-byte random hex string (64 chars)
2. Store as `webhook_secret` in database (unique index)
3. Compute webhook URL: `{base_url}/api/v1/telegram/webhook/{secret}`
4. Call Telegram `setWebhook` API automatically
5. Telegram sends updates to unguessable URL

**Security:**
- ✅ 2^256 possible secret values (cryptographically unguessable)
- ✅ Unique index prevents collisions
- ✅ Only valid bots can receive updates
- ✅ Automatic webhook registration (no manual step)

---

## Database Changes

### Migration: `migrations/003_bot_webhook_secret.sql`

```sql
ALTER TABLE bots ADD COLUMN webhook_secret VARCHAR(64) AFTER webhook_url;
CREATE UNIQUE INDEX idx_bots_webhook_secret ON bots(webhook_secret);
```

### Updated Bot Model

```go
type Bot struct {
    // ... existing fields
    WebhookURL    string    `gorm:"size:512" json:"webhook_url,omitempty"`
    WebhookSecret string    `gorm:"uniqueIndex;size:64" json:"-"` // NEW
    // ...
}
```

---

## Service Layer Changes

### BotService Constructor

**Before:**
```go
func NewBotService(botRepo BotRepository, encryptionKey string) *BotService
```

**After:**
```go
func NewBotService(botRepo BotRepository, encryptionKey, webhookBaseURL string) *BotService
```

### New Methods

```go
// GetBotByWebhookSecret - lookup bot by secret for webhook handler
func (s *BotService) GetBotByWebhookSecret(ctx context.Context, secret string) (*BotDTO, error)

// SetWebhook - re-register webhook with Telegram (manual recovery)
func (s *BotService) SetWebhook(ctx context.Context, botID uint) error

// Internal: setTelegramWebhook - calls Telegram API
func (s *BotService) setTelegramWebhook(ctx context.Context, token, webhookURL string) error

// Internal: deleteTelegramWebhook - deregisters webhook
func (s *BotService) deleteTelegramWebhook(ctx context.Context, token string) error
```

### Updated CreateBot

```go
func (s *BotService) CreateBot(ctx context.Context, req *CreateBotRequest) (*BotDTO, error) {
    // 1. Encrypt token
    // 2. Generate random webhook secret (32 bytes hex)
    // 3. Compute webhook URL: {base}/api/v1/telegram/webhook/{secret}
    // 4. Create bot in database
    // 5. Call Telegram setWebhook API
    // 6. If webhook fails, rollback DB insert
}
```

### Updated DeleteBot

```go
func (s *BotService) DeleteBot(ctx context.Context, id uint) error {
    // 1. Fetch bot (to get token)
    // 2. Call Telegram deleteWebhook API
    // 3. Delete from database (CASCADE deletes chats/permissions)
}
```

---

## Repository Changes

### New Interface Method

```go
type BotRepository interface {
    // ... existing methods
    GetByWebhookSecret(ctx context.Context, secret string) (*domain.Bot, error) // NEW
}
```

### Implementation

```go
func (r *botRepository) GetByWebhookSecret(ctx context.Context, secret string) (*domain.Bot, error) {
    var bot domain.Bot
    err := r.db.WithContext(ctx).
        Where("webhook_secret = ? AND is_active = true", secret).
        First(&bot).Error
    return &bot, err
}
```

---

## Handler Changes

### BotHandler - Routes Removed

**Before:**
```go
bots.POST("", botHandler.CreateBot)      // ❌ REMOVED
bots.DELETE("/:id", botHandler.DeleteBot) // ❌ REMOVED
bots.GET("", botHandler.ListBots)        // ✅ KEPT (read-only)
bots.GET("/:id", botHandler.GetBot)      // ✅ KEPT (read-only)
```

**After:**
```go
// Bot management - READ-ONLY (write operations: ./bin/bot)
bots.GET("", botHandler.ListBots)
bots.GET("/:id", botHandler.GetBot)
```

### TelegramHandler - Webhook Route

**Before:**
```go
v1.POST("/telegram/webhook/:bot_username", telegramHandler.ReceiveUpdate)

// Handler
botUsername := c.Param("bot_username")
bot, err := h.botService.GetBot(ctx, 0) // TODO: broken!
```

**After:**
```go
v1.POST("/telegram/webhook/:webhook_secret", telegramHandler.ReceiveUpdate)

// Handler
webhookSecret := c.Param("webhook_secret")
bot, err := h.botService.GetBotByWebhookSecret(ctx, webhookSecret)
```

**Fixes:**
- ✅ No more `GetBot(ctx, 0)` TODO bug
- ✅ Proper bot lookup by secret
- ✅ Only active bots can receive updates

---

## CLI Tool Structure

### File Organization

```
cmd/bot/
├── main.go                    # CLI entry point
└── commands/
    ├── common.go              # Shared utilities
    ├── create.go              # Create bot + register webhook
    ├── list.go                # List all bots
    ├── get.go                 # Get bot details
    ├── update.go              # Update bot metadata
    ├── delete.go              # Delete bot + deregister webhook
    └── show_token.go          # Display decrypted token
```

### Command Examples

#### Create Bot
```bash
./bin/bot create \
  --username my_bot \
  --token "123456:ABC-DEF..." \
  --display-name "My Bot" \
  --description "Production bot"

# Output:
# ✓ Bot created successfully
#
# Bot Details:
#   ID:           1
#   Username:     my_bot
#   Display Name: My Bot
#   Webhook URL:  https://example.com/api/v1/telegram/webhook/a1b2c3d4...
#
# The webhook has been registered with Telegram.
```

#### List Bots
```bash
./bin/bot list

# Output:
# ID | Username         | Display Name     | Active | Webhook URL
# ---|------------------|------------------|--------|------------------
# 1  | my_bot           | My Bot           | Yes    | https://...
```

#### Delete Bot
```bash
./bin/bot delete 1

# Output:
# WARNING: Deleting a bot will:
#   - Delete the bot record
#   - CASCADE delete all associated chats
#   - CASCADE delete all associated permissions
#   - Deregister the webhook from Telegram
#
# To proceed, add the --force flag:
#   bot delete 1 --force

./bin/bot delete 1 --force

# Output:
# Deleting bot and deregistering webhook from Telegram...
# ✓ Bot deleted successfully
```

#### Show Token
```bash
./bin/bot show-token 1

# Output:
# Bot Token (ID: 1):
#   123456:ABC-DEF1234567890abcdefghijklmnopqrstuvwxyz
#
# ⚠️  Keep this token secure! Anyone with this token can control your bot.
```

---

## Telegram API Integration

### setWebhook

Called automatically on `bot create`:

```bash
POST https://api.telegram.org/bot{token}/setWebhook
Content-Type: application/json

{
  "url": "https://example.com/api/v1/telegram/webhook/a1b2c3d4..."
}
```

**Response:**
```json
{
  "ok": true,
  "result": true,
  "description": "Webhook was set"
}
```

**Error Handling:**
- If Telegram API fails, bot record is rolled back
- User sees clear error message
- Token never leaves the server

### deleteWebhook

Called automatically on `bot delete --force`:

```bash
POST https://api.telegram.org/bot{token}/deleteWebhook
```

**Graceful Degradation:**
- If delete fails (bot already deleted on Telegram), log warning but continue
- Database record is still deleted
- Prevents orphaned records

---

## Migration Guide

### For Administrators

1. **Build CLI tool:**
   ```bash
   make build
   # or
   go build -o bin/bot cmd/bot/main.go
   ```

2. **Run database migration:**
   ```bash
   go run cmd/migrate/main.go up
   ```

3. **Create bots using CLI:**
   ```bash
   ./bin/bot create --username my_bot --token "..."
   ```

### For API Users

**Breaking Changes:**

```bash
# ❌ NO LONGER WORKS
curl -X POST /api/v1/bots

# ✅ USE CLI INSTEAD
./bin/bot create --username my_bot --token "..."
```

**Still Works (Read-Only):**

```bash
# ✅ List bots
curl /api/v1/bots?api_key=tgw_xxx

# ✅ Get bot
curl /api/v1/bots/1?api_key=tgw_xxx
```

---

## Testing Checklist

- [x] `make build` compiles all binaries
- [x] CLI help command works
- [ ] Database migration runs successfully
- [ ] `bot create` creates bot and registers webhook with Telegram
- [ ] `bot list` shows all bots
- [ ] `bot get <id>` displays full bot details
- [ ] `bot show-token <id>` displays decrypted token
- [ ] `bot update <id>` updates metadata
- [ ] `bot delete <id>` requires --force flag
- [ ] `bot delete <id> --force` deletes bot and deregisters webhook
- [ ] Telegram sends updates to `/telegram/webhook/{secret}`
- [ ] Gateway looks up bot by secret correctly
- [ ] Invalid secret returns 401 Unauthorized
- [ ] `POST /api/v1/bots` returns 404 Not Found
- [ ] `DELETE /api/v1/bots/:id` returns 404 Not Found
- [ ] `GET /api/v1/bots` still works (read-only)

---

## Files Modified

### Core Changes

| File | Changes |
|------|---------|
| `internal/domain/models.go` | Added `WebhookSecret` field to `Bot` struct |
| `internal/service/bot_service.go` | Added `webhookBaseURL`, Telegram API methods, updated create/delete |
| `internal/repository/repositories.go` | Added `GetByWebhookSecret` method |
| `internal/handler/bot_handler.go` | Removed `CreateBot` and `DeleteBot` methods |
| `internal/handler/telegram_handler.go` | Updated `ReceiveUpdate` to use webhook secret |
| `cmd/gateway/main.go` | Updated routes and service initialization |
| `cmd/migrate/main.go` | Added migration 003 to migration list |

### New Files

| File | Purpose |
|------|---------|
| `migrations/003_bot_webhook_secret.sql` | Database migration |
| `cmd/bot/main.go` | CLI entry point |
| `cmd/bot/commands/common.go` | Shared utilities |
| `cmd/bot/commands/create.go` | Create command |
| `cmd/bot/commands/list.go` | List command |
| `cmd/bot/commands/get.go` | Get command |
| `cmd/bot/commands/update.go` | Update command |
| `cmd/bot/commands/delete.go` | Delete command |
| `cmd/bot/commands/show_token.go` | Show token command |

### Build Changes

| File | Changes |
|------|---------|
| `Makefile` | Added `bin/bot` to build target |
| `README.md` | Updated bot management documentation |

---

## Security Considerations

### Threat Model

**Before:**
- Attacker intercepts HTTPS traffic (MITM)
- Bot token logged in reverse proxy
- Token exposed in application logs
- Web vulnerability exposes token

**After:**
- ✅ Token never leaves server
- ✅ Requires shell access to create bots
- ✅ Webhook secret is cryptographically random
- ✅ Reduced attack surface

### Best Practices

1. **Restrict CLI access:**
   ```bash
   chmod 700 bin/bot
   chown root:admin bin/bot
   ```

2. **Use secure config:**
   ```bash
   chmod 600 configs/config.json
   ```

3. **Audit bot creation:**
   ```bash
   # Add to deployment pipeline
   ./bin/bot list > /var/log/bots-$(date +%Y%m%d).log
   ```

4. **Rotate webhook secrets:**
   ```sql
   -- If webhook secret is compromised, generate new one
   UPDATE bots SET webhook_secret = 'new_random_secret' WHERE id = 1;
   -- Then re-register webhook manually or via CLI
   ```

---

## Performance Impact

- **No performance impact** - bot creation is infrequent
- Webhook lookup by secret: O(1) with unique index
- HTTP handler simplified (no create/delete logic)

---

## Future Enhancements

1. **Webhook secret rotation:**
   ```bash
   ./bin/bot rotate-secret 1
   ```

2. **Bulk operations:**
   ```bash
   ./bin/bot import bots.json
   ./bin/bot export --output bots.json
   ```

3. **Webhook health checks:**
   ```bash
   ./bin/bot check-webhook 1
   ```

4. **Audit trail:**
   ```bash
   ./bin/bot audit-log --filter "create,delete"
   ```

---

## Rollback Plan

If issues arise:

1. **Keep old HTTP routes temporarily:**
   ```go
   // Deprecated - use CLI
   bots.POST("", botHandler.CreateBot)
   bots.DELETE("/:id", botHandler.DeleteBot)
   ```

2. **Use feature flag:**
   ```json
   {
     "features": {
       "cli_only_bot_management": false
     }
   }
   ```

3. **Database rollback:**
   ```sql
   ALTER TABLE bots DROP COLUMN webhook_secret;
   DROP INDEX idx_bots_webhook_secret;
   ```

---

## Conclusion

Moving bot management to CLI significantly improves security by:
- Eliminating network exposure of bot tokens
- Requiring server access for privileged operations
- Automatically handling webhook registration with random secrets
- Simplifying the HTTP API surface

The trade-off is slightly less convenience for administrators, but the security benefits far outweigh this cost for production environments where bot tokens grant full control over Telegram bots.
