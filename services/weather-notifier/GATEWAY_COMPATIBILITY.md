# Gateway API Compatibility Guide

## Weather Notifier Service ↔ Telegram Bot Gateway

This document explains how the Weather Notifier Service integrates with the Telegram Bot Gateway after the CLI-based API key management update.

---

## ✅ Full Compatibility Confirmed

The Weather Notifier Service is **fully compatible** with the updated Telegram Bot Gateway (main branch).

### Why It Works

The weather service uses **JWT authentication** for API access, which remains unchanged:

```go
// internal/telegram/auth.go
// Uses these endpoints (STILL AVAILABLE):
POST /api/v1/auth/login     ✅ Still works
POST /api/v1/auth/refresh   ✅ Still works

// internal/telegram/client.go
// Uses these endpoints (STILL AVAILABLE):
POST /api/v1/messages/send  ✅ Still works
```

### What Changed in Gateway (Doesn't Affect Weather Service)

The gateway disabled **API key management endpoints**:

```
❌ POST   /api/v1/apikeys          # Now CLI-only
❌ GET    /api/v1/apikeys          # Now CLI-only
❌ GET    /api/v1/apikeys/:id      # Now CLI-only
❌ POST   /api/v1/apikeys/:id/revoke  # Now CLI-only
❌ DELETE /api/v1/apikeys/:id      # Now CLI-only
```

**Impact**: None - Weather service doesn't use these endpoints.

---

## Authentication Flow

The weather service authenticates with the gateway using:

### 1. Bot Credentials
Configured in `config.yaml`:
```yaml
telegram:
  bot_token: "YOUR_TELEGRAM_BOT_TOKEN"    # From @BotFather
  password: "YOUR_SECURE_PASSWORD"         # Gateway user password
  api_url: "http://gateway:8080"           # Gateway URL
```

### 2. JWT Token Management
The `TokenManager` handles authentication automatically:

1. **Login** → Receives access token + refresh token
2. **Auto-refresh** → Refreshes token before expiration (5 min buffer)
3. **Resilience** → Falls back to login if refresh fails

### 3. API Requests
All API requests use the JWT Bearer token:

```go
// Automatic in client.go
req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
```

---

## Setup with Updated Gateway

### Prerequisites

1. **Gateway is running** with user account created
2. **Bot is registered** in the gateway
3. **Bot credentials** are configured

### Configuration

#### Option 1: Using Existing User Account

If you already have a user account in the gateway:

```yaml
telegram:
  bot_token: "123456:ABC-DEF..."  # Your bot token
  password: "your_user_password"   # Your gateway user password
  api_url: "http://gateway:8080"
```

#### Option 2: Create Dedicated Bot User

For better security, create a dedicated user for the weather bot:

**On Gateway Server:**
```bash
# Create user (manual SQL or admin panel)
# Then configure weather service with those credentials
```

**In Weather Service:**
```yaml
telegram:
  bot_token: "123456:ABC-DEF..."
  password: "weather_bot_password"
  api_url: "http://gateway:8080"
```

---

## API Permissions

### Required Gateway Permissions

The weather service needs these permissions on the gateway:

1. **Authentication** (`/api/v1/auth/*`) ✅ Public endpoints
2. **Send Messages** (`/api/v1/messages/send`) ✅ With valid JWT

### Chat-Level Permissions

If the gateway has chat-level ACL enabled, ensure the bot user has:
- `can_send` permission on target chats

**Check permissions:**
```bash
# On gateway server (if using CLI-managed permissions)
./bin/apikey show-permissions <user-id>
```

---

## Deployment Scenarios

### Scenario 1: Local Development

```yaml
# docker-compose.yml or config.yaml
telegram:
  api_url: "http://localhost:8080"
```

### Scenario 2: Docker Compose (Same Network)

```yaml
# docker-compose.yml
services:
  weather-notifier:
    environment:
      - WNB_TELEGRAM_API_URL=http://gateway:8080
    networks:
      - gateway-network

  gateway:
    networks:
      - gateway-network
```

### Scenario 3: Separate Deployments

```yaml
# Weather service config
telegram:
  api_url: "https://gateway.example.com"
```

---

## Troubleshooting

### Authentication Fails

**Error**: `login failed with status 401`

**Solutions**:
1. Verify bot_token and password are correct
2. Check if user account is active in gateway
3. Ensure API URL is reachable

```bash
# Test gateway reachability
curl http://gateway:8080/health
```

### Token Refresh Issues

**Error**: `refresh failed, attempting login`

**This is normal** - The service automatically falls back to login if refresh fails.

If it persists:
1. Check gateway logs for refresh endpoint errors
2. Verify clock sync between services
3. Check if refresh token TTL is too short

### Message Send Failures

**Error**: `send message failed with status 403`

**Solutions**:
1. **Chat ACL**: User may not have `can_send` permission
2. **Bot not in chat**: Ensure bot is added to the target chat
3. **Chat not registered**: Gateway may not know about the chat yet

---

## Migration Notes

### From Old Gateway (Pre-CLI) → New Gateway

**No code changes needed!**

The weather service continues to work as-is because:
- ✅ JWT authentication unchanged
- ✅ Message sending endpoints unchanged
- ✅ Auth flow unchanged

Only **API key management** was moved to CLI (doesn't affect weather service).

### If You Were Using API Keys (Hypothetical)

If the weather service hypothetically used API keys instead of JWT:

**Before** (REST API):
```bash
curl -X POST http://gateway/api/v1/apikeys \
  -H "Authorization: Bearer $JWT" \
  -d '{"name":"weather-bot"}'
```

**After** (CLI):
```bash
./bin/apikey create --name "weather-bot"
```

But since weather service uses JWT, **no changes needed**.

---

## Best Practices

### 1. Use Environment Variables

```bash
export WNB_TELEGRAM_BOT_TOKEN="..."
export WNB_TELEGRAM_PASSWORD="..."
export WNB_TELEGRAM_API_URL="http://gateway:8080"
```

### 2. Secure Credentials

- Never commit passwords to git
- Use Docker secrets or Kubernetes secrets for production
- Rotate passwords periodically

### 3. Monitor Authentication

Check logs for authentication issues:

```bash
docker-compose logs weather-notifier | grep token_manager
```

### 4. Handle Gateway Downtime

The service handles temporary gateway downtime gracefully:
- Failed requests trigger retry
- Token refresh failures fall back to login
- Logs errors for monitoring

---

## Summary

| Feature | Status | Notes |
|---------|--------|-------|
| **JWT Authentication** | ✅ Fully supported | No changes |
| **Message Sending** | ✅ Fully supported | No changes |
| **Auto Token Refresh** | ✅ Fully supported | No changes |
| **API Key Usage** | ❌ Not used | N/A for weather service |
| **Chat ACL** | ✅ Compatible | Ensure permissions if enabled |

**Conclusion**: Weather Notifier Service requires **zero changes** to work with the updated gateway.

---

## Related Documentation

- [Gateway API Documentation](https://github.com/KexiChanProjectProxy/telegram-bot-gateway/blob/main/docs/API_COMPLETE.md)
- [Gateway Authentication Guide](https://github.com/KexiChanProjectProxy/telegram-bot-gateway/blob/main/docs/AUTHENTICATION.md)
- [CLI-Based API Key Management](https://github.com/KexiChanProjectProxy/telegram-bot-gateway/blob/main/MIGRATION_APIKEY.md)

---

**Last Updated**: 2026-02-09
**Gateway Version**: v1.1.0 (CLI-based API key management)
**Weather Service Version**: v1.0.0
