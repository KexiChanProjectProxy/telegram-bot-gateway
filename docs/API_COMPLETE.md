# Telegram Bot Gateway - Complete API Documentation

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Response Format](#response-format)
4. [Error Handling](#error-handling)
5. [Rate Limiting](#rate-limiting)
6. [API Endpoints](#api-endpoints)
7. [WebSocket API](#websocket-api)
8. [Telegram Webhook](#telegram-webhook)
9. [Examples](#examples)

---

## Overview

**Base URL**: `http://your-domain.com/api/v1`

The Telegram Bot Gateway provides a RESTful API for managing Telegram bots, chats, messages, webhooks, and API keys. All requests and responses use JSON format unless otherwise specified.

### API Versions

- **v1** (Current): `/api/v1/*`

---

## Authentication

The gateway supports **three authentication methods**:

### 1. JWT Bearer Token (Recommended for User Accounts)

Use for interactive user sessions with login/logout support.

**Header**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Example**:
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/bots
```

### 2. API Key - Header Method

Use for machine-to-machine communication and service integrations.

**Header**:
```
X-API-Key: tgw_1234567890abcdef
```

**Example**:
```bash
curl -H "X-API-Key: tgw_1234567890abcdef" \
  http://localhost:8080/api/v1/bots
```

### 3. API Key - Telegram Bot Style (Query/POST)

**Similar to Telegram Bot API format** - supports API keys via query parameters or POST body.

#### 3a. Query Parameter (GET requests)

```bash
# Using 'api_key' parameter
curl "http://localhost:8080/api/v1/bots?api_key=tgw_1234567890abcdef"

# OR using 'token' parameter (alias)
curl "http://localhost:8080/api/v1/bots?token=tgw_1234567890abcdef"
```

#### 3b. POST Body (POST requests)

**Form-data or x-www-form-urlencoded**:
```bash
# Using 'api_key' parameter
curl -X POST "http://localhost:8080/api/v1/chats/1/messages" \
  -d "api_key=tgw_1234567890abcdef" \
  -d "text=Hello World"

# OR using 'token' parameter (alias)
curl -X POST "http://localhost:8080/api/v1/chats/1/messages" \
  -d "token=tgw_1234567890abcdef" \
  -d "text=Hello World"
```

**JSON with query parameter**:
```bash
curl -X POST "http://localhost:8080/api/v1/chats/1/messages?api_key=tgw_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello World"}'
```

### Authentication Priority

The gateway checks authentication in this order:
1. JWT Bearer token (if present)
2. X-API-Key header (if present)
3. `api_key` or `token` query parameter (if present)
4. `api_key` or `token` POST body field (if present and POST request)

### Token Lifetime

| Token Type | Lifetime | Renewable |
|------------|----------|-----------|
| **Access Token (JWT)** | 15 minutes | Yes (via refresh) |
| **Refresh Token (JWT)** | 7 days | No |
| **API Key** | Configurable | N/A |

---

## Response Format

### Success Response

```json
{
  "id": 1,
  "username": "my_bot",
  "display_name": "My Bot",
  "is_active": true,
  "created_at": "2026-02-09T10:30:00Z"
}
```

### Error Response

```json
{
  "error": "Bot not found",
  "code": "NOT_FOUND",
  "details": {
    "bot_id": 123
  }
}
```

### Pagination Response

```json
{
  "data": [...],
  "pagination": {
    "offset": 0,
    "limit": 20,
    "total": 150,
    "has_more": true
  }
}
```

---

## Error Handling

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| **200** | OK | Request successful |
| **201** | Created | Resource created successfully |
| **400** | Bad Request | Invalid request parameters |
| **401** | Unauthorized | Authentication required or failed |
| **403** | Forbidden | Insufficient permissions |
| **404** | Not Found | Resource not found |
| **409** | Conflict | Resource already exists |
| **422** | Unprocessable Entity | Validation error |
| **429** | Too Many Requests | Rate limit exceeded |
| **500** | Internal Server Error | Server error |
| **503** | Service Unavailable | Service temporarily unavailable |

### Error Codes

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

Common error codes:
- `INVALID_REQUEST` - Malformed request
- `AUTHENTICATION_REQUIRED` - No credentials provided
- `INVALID_CREDENTIALS` - Invalid token or API key
- `PERMISSION_DENIED` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `ALREADY_EXISTS` - Resource conflict
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `INTERNAL_ERROR` - Server error

---

## Rate Limiting

### Rate Limit Headers

Every response includes rate limit headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1644408600
```

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Maximum requests per window |
| `X-RateLimit-Remaining` | Remaining requests in current window |
| `X-RateLimit-Reset` | Unix timestamp when limit resets |

### Default Limits

| Endpoint Type | Limit | Window |
|---------------|-------|--------|
| **Public** (login, refresh) | 10 requests | Per IP per minute |
| **Authenticated** | 100 requests | Per user/API key per minute |
| **Global** (optional) | 1000 requests | Per minute |

### Rate Limit Exceeded Response

```json
{
  "error": "Rate limit exceeded. Try again in 30 seconds.",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 30
}
```

---

## API Endpoints

### 1. Authentication Endpoints

#### POST /api/v1/auth/login

Login and get JWT tokens.

**Request**:
```json
{
  "username": "admin",
  "password": "password123"
}
```

**Response** (200):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "is_active": true,
    "created_at": "2026-02-09T10:00:00Z"
  }
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}'
```

---

#### POST /api/v1/auth/refresh

Refresh access token using refresh token.

**Request**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response** (200):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

---

#### POST /api/v1/auth/logout

Logout and revoke refresh token.

**Request**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response** (200):
```json
{
  "message": "Logged out successfully"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

---

### 2. Bot Management Endpoints

#### POST /api/v1/bots

Create/register a new Telegram bot.

**Authentication**: Required (JWT or API Key)

**Request**:
```json
{
  "username": "my_bot",
  "token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
  "display_name": "My Awesome Bot",
  "description": "A bot that does amazing things"
}
```

**Response** (201):
```json
{
  "id": 1,
  "username": "my_bot",
  "display_name": "My Awesome Bot",
  "description": "A bot that does amazing things",
  "is_active": true,
  "webhook_url": "https://your-domain.com/telegram/webhook/123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
  "created_at": "2026-02-09T10:30:00Z",
  "updated_at": "2026-02-09T10:30:00Z"
}
```

**Examples**:
```bash
# Using JWT
curl -X POST http://localhost:8080/api/v1/bots \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "my_bot",
    "token": "123456:ABC-DEF...",
    "display_name": "My Bot"
  }'

# Using API Key (Header)
curl -X POST http://localhost:8080/api/v1/bots \
  -H "X-API-Key: tgw_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "my_bot",
    "token": "123456:ABC-DEF...",
    "display_name": "My Bot"
  }'

# Using API Key (Query Parameter)
curl -X POST "http://localhost:8080/api/v1/bots?api_key=tgw_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "my_bot",
    "token": "123456:ABC-DEF...",
    "display_name": "My Bot"
  }'
```

---

#### GET /api/v1/bots

List all registered bots.

**Authentication**: Required

**Query Parameters**:
- `offset` (optional, default: 0) - Pagination offset
- `limit` (optional, default: 20, max: 100) - Results per page

**Response** (200):
```json
{
  "bots": [
    {
      "id": 1,
      "username": "my_bot",
      "display_name": "My Bot",
      "is_active": true,
      "created_at": "2026-02-09T10:30:00Z"
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 20
}
```

**Examples**:
```bash
# Using JWT
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/bots

# Using API Key (Header)
curl -H "X-API-Key: tgw_1234567890abcdef" \
  http://localhost:8080/api/v1/bots

# Using API Key (Query)
curl "http://localhost:8080/api/v1/bots?api_key=tgw_1234567890abcdef"

# With pagination
curl "http://localhost:8080/api/v1/bots?api_key=tgw_xxx&offset=20&limit=50"
```

---

#### GET /api/v1/bots/:id

Get details of a specific bot.

**Authentication**: Required

**Response** (200):
```json
{
  "id": 1,
  "username": "my_bot",
  "display_name": "My Bot",
  "description": "A bot that does amazing things",
  "is_active": true,
  "webhook_url": "https://your-domain.com/telegram/webhook/123456:ABC-DEF...",
  "created_at": "2026-02-09T10:30:00Z",
  "updated_at": "2026-02-09T10:30:00Z"
}
```

**Examples**:
```bash
# Using JWT
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/bots/1

# Using API Key (Query)
curl "http://localhost:8080/api/v1/bots/1?api_key=tgw_1234567890abcdef"
```

---

#### DELETE /api/v1/bots/:id

Delete a bot.

**Authentication**: Required

**Response** (200):
```json
{
  "message": "Bot deleted successfully"
}
```

**Example**:
```bash
curl -X DELETE "http://localhost:8080/api/v1/bots/1?api_key=tgw_xxx"
```

---

### 3. Chat Management Endpoints

#### GET /api/v1/chats

List all accessible chats.

**Authentication**: Required

**Query Parameters**:
- `bot_id` (optional) - Filter by bot ID
- `offset` (optional, default: 0)
- `limit` (optional, default: 20, max: 100)

**Response** (200):
```json
{
  "chats": [
    {
      "id": 1,
      "bot_id": 1,
      "telegram_id": 123456789,
      "type": "private",
      "username": "john_doe",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2026-02-09T11:00:00Z"
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 20
}
```

**Examples**:
```bash
# List all chats
curl "http://localhost:8080/api/v1/chats?api_key=tgw_xxx"

# Filter by bot
curl "http://localhost:8080/api/v1/chats?api_key=tgw_xxx&bot_id=1"
```

---

#### GET /api/v1/chats/:id

Get chat details.

**Authentication**: Required

**Response** (200):
```json
{
  "id": 1,
  "bot_id": 1,
  "telegram_id": 123456789,
  "type": "private",
  "username": "john_doe",
  "first_name": "John",
  "last_name": "Doe",
  "is_active": true,
  "message_count": 42,
  "created_at": "2026-02-09T11:00:00Z",
  "updated_at": "2026-02-09T12:00:00Z"
}
```

**Example**:
```bash
curl "http://localhost:8080/api/v1/chats/1?token=tgw_xxx"
```

---

#### GET /api/v1/chats/:id/messages

Get messages from a chat. **Requires `can_read` permission** for the chat.

**Authentication**: Required
**Authorization**: Chat-level ACL (can_read)

**Query Parameters**:
- `cursor` (optional) - Unix timestamp for pagination (get messages before this time)
- `limit` (optional, default: 50, max: 100)

**Response** (200):
```json
{
  "messages": [
    {
      "id": 1,
      "chat_id": 1,
      "telegram_id": 1001,
      "from_user_id": 123456789,
      "from_username": "john_doe",
      "from_first_name": "John",
      "direction": "incoming",
      "message_type": "text",
      "text": "Hello, bot!",
      "sent_at": "2026-02-09T12:00:00Z",
      "created_at": "2026-02-09T12:00:01Z"
    }
  ],
  "has_more": false,
  "next_cursor": null
}
```

**Examples**:
```bash
# Get latest messages
curl "http://localhost:8080/api/v1/chats/1/messages?api_key=tgw_xxx"

# Paginated (get older messages)
curl "http://localhost:8080/api/v1/chats/1/messages?api_key=tgw_xxx&cursor=1644408000&limit=100"

# Using POST body for API key
curl -X GET "http://localhost:8080/api/v1/chats/1/messages" \
  -d "api_key=tgw_xxx" \
  -d "limit=50"
```

---

#### POST /api/v1/chats/:id/messages

Send a message to a chat. **Requires `can_send` permission** for the chat.

**Authentication**: Required
**Authorization**: Chat-level ACL (can_send)

**Request**:
```json
{
  "text": "Hello from the gateway!",
  "reply_to_message_id": 1001
}
```

**Response** (201):
```json
{
  "success": true,
  "message": "Message queued for delivery",
  "message_id": 2,
  "queued_at": 1644408600
}
```

**Examples**:
```bash
# Send message (JSON)
curl -X POST "http://localhost:8080/api/v1/chats/1/messages?api_key=tgw_xxx" \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello from gateway!"}'

# Send message (Form-data with API key in body)
curl -X POST "http://localhost:8080/api/v1/chats/1/messages" \
  -d "api_key=tgw_xxx" \
  -d "text=Hello from gateway!"

# Reply to message
curl -X POST "http://localhost:8080/api/v1/chats/1/messages" \
  -d "token=tgw_xxx" \
  -d "text=Reply message" \
  -d "reply_to_message_id=1001"
```

---

### 4. API Key Management Endpoints

#### POST /api/v1/apikeys

Create a new API key.

**Authentication**: Required (JWT only - not API key)

**Request**:
```json
{
  "name": "Production API Key",
  "scopes": ["bots:read", "chats:read", "messages:send"],
  "expires_in_days": 365
}
```

**Response** (201):
```json
{
  "id": 1,
  "key": "tgw_1234567890abcdef1234567890abcdef",
  "name": "Production API Key",
  "scopes": ["bots:read", "chats:read", "messages:send"],
  "is_active": true,
  "expires_at": "2027-02-09T10:00:00Z",
  "created_at": "2026-02-09T10:00:00Z"
}
```

**⚠️ Important**: The `key` field is only shown once during creation. Store it securely!

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/apikeys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Key",
    "scopes": ["bots:read", "messages:send"],
    "expires_in_days": 365
  }'
```

---

#### GET /api/v1/apikeys

List all API keys for the authenticated user.

**Authentication**: Required (JWT only)

**Response** (200):
```json
{
  "api_keys": [
    {
      "id": 1,
      "name": "Production API Key",
      "scopes": ["bots:read", "chats:read"],
      "is_active": true,
      "last_used_at": "2026-02-09T12:30:00Z",
      "expires_at": "2027-02-09T10:00:00Z",
      "created_at": "2026-02-09T10:00:00Z"
    }
  ],
  "total": 1
}
```

**Note**: The actual key value is never returned after creation.

---

#### GET /api/v1/apikeys/:id

Get API key details.

**Authentication**: Required (JWT only)

**Response** (200):
```json
{
  "id": 1,
  "name": "Production API Key",
  "scopes": ["bots:read", "chats:read", "messages:send"],
  "is_active": true,
  "last_used_at": "2026-02-09T12:30:00Z",
  "usage_count": 1542,
  "expires_at": "2027-02-09T10:00:00Z",
  "created_at": "2026-02-09T10:00:00Z"
}
```

---

#### POST /api/v1/apikeys/:id/revoke

Revoke an API key (deactivate).

**Authentication**: Required (JWT only)

**Response** (200):
```json
{
  "message": "API key revoked successfully"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/apikeys/1/revoke \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

#### DELETE /api/v1/apikeys/:id

Permanently delete an API key.

**Authentication**: Required (JWT only)

**Response** (200):
```json
{
  "message": "API key deleted successfully"
}
```

---

### 5. Webhook Management Endpoints

#### POST /api/v1/webhooks

Register a new webhook URL.

**Authentication**: Required

**Request**:
```json
{
  "url": "https://my-app.com/telegram/updates",
  "chat_ids": [1, 2, 3],
  "event_types": ["new_message", "edited_message"],
  "scope": "chat",
  "is_active": true
}
```

**Response** (201):
```json
{
  "id": 1,
  "url": "https://my-app.com/telegram/updates",
  "secret": "wh_secret_abc123",
  "chat_ids": [1, 2, 3],
  "event_types": ["new_message", "edited_message"],
  "scope": "chat",
  "is_active": true,
  "created_at": "2026-02-09T10:00:00Z"
}
```

**Webhook Payload Signing**: All webhook deliveries include an `X-Webhook-Signature` header with HMAC-SHA256 signature. Verify using the `secret`.

---

#### GET /api/v1/webhooks

List registered webhooks.

**Authentication**: Required

**Response** (200):
```json
{
  "webhooks": [
    {
      "id": 1,
      "url": "https://my-app.com/telegram/updates",
      "chat_ids": [1, 2, 3],
      "event_types": ["new_message"],
      "scope": "chat",
      "is_active": true,
      "delivery_count": 1542,
      "last_delivery_at": "2026-02-09T12:30:00Z",
      "created_at": "2026-02-09T10:00:00Z"
    }
  ],
  "total": 1
}
```

---

#### GET /api/v1/webhooks/:id

Get webhook details.

**Authentication**: Required

**Response** (200):
```json
{
  "id": 1,
  "url": "https://my-app.com/telegram/updates",
  "secret": "wh_secret_abc123",
  "chat_ids": [1, 2, 3],
  "event_types": ["new_message", "edited_message"],
  "scope": "chat",
  "is_active": true,
  "delivery_count": 1542,
  "failure_count": 3,
  "last_delivery_at": "2026-02-09T12:30:00Z",
  "last_failure_at": "2026-02-08T08:00:00Z",
  "created_at": "2026-02-09T10:00:00Z",
  "updated_at": "2026-02-09T11:00:00Z"
}
```

---

#### PUT /api/v1/webhooks/:id

Update webhook configuration.

**Authentication**: Required

**Request**:
```json
{
  "url": "https://my-app.com/new-endpoint",
  "chat_ids": [1, 2, 3, 4],
  "event_types": ["new_message"],
  "is_active": true
}
```

**Response** (200):
```json
{
  "id": 1,
  "url": "https://my-app.com/new-endpoint",
  "chat_ids": [1, 2, 3, 4],
  "event_types": ["new_message"],
  "is_active": true,
  "updated_at": "2026-02-09T13:00:00Z"
}
```

---

#### DELETE /api/v1/webhooks/:id

Delete a webhook.

**Authentication**: Required

**Response** (200):
```json
{
  "message": "Webhook deleted successfully"
}
```

---

### 6. System Endpoints

#### GET /health

Health check endpoint (no authentication required).

**Response** (200):
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": 1644408600,
  "websocket_clients": 42
}
```

**Example**:
```bash
curl http://localhost:8080/health
```

---

#### GET /metrics

System metrics (authentication required).

**Authentication**: Required

**Response** (200):
```json
{
  "uptime_seconds": 86400,
  "goroutines": 42,
  "memory": {
    "alloc_mb": 25,
    "total_alloc_mb": 150,
    "sys_mb": 45
  },
  "database": {
    "open_connections": 5,
    "in_use": 2,
    "idle": 3
  },
  "redis": {
    "connected": true,
    "used_memory_mb": 12
  },
  "websocket": {
    "active_clients": 42,
    "total_messages": 15420
  },
  "webhooks": {
    "active_workers": 10,
    "queued_deliveries": 5
  }
}
```

**Example**:
```bash
curl -H "X-API-Key: tgw_xxx" http://localhost:8080/metrics
```

---

## WebSocket API

### Connection

**Endpoint**: `ws://your-domain.com/api/v1/ws`

**Authentication**: Required (JWT or API Key via query parameter)

```javascript
// Using JWT
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?token=' + jwtToken);

// Using API Key
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?api_key=tgw_xxx');
```

### Messages

#### Subscribe to Chat

```json
{
  "action": "subscribe",
  "chat_id": 1
}
```

#### Unsubscribe from Chat

```json
{
  "action": "unsubscribe",
  "chat_id": 1
}
```

#### Ping (Keep-Alive)

```json
{
  "action": "ping"
}
```

**Response**:
```json
{
  "action": "pong",
  "timestamp": 1644408600
}
```

### Receiving Messages

```json
{
  "type": "new_message",
  "chat_id": 1,
  "message_id": 123,
  "telegram_id": 1001,
  "bot_id": 1,
  "direction": "incoming",
  "text": "Hello from Telegram!",
  "from_username": "john_doe",
  "from_first_name": "John",
  "message_type": "text",
  "timestamp": "2026-02-09T12:00:00Z",
  "metadata": {}
}
```

---

## Telegram Webhook

### Receiving Updates from Telegram

**Endpoint**: `POST /telegram/webhook/:bot_token`

**No authentication required** - validated by bot token in URL.

Telegram will send updates to:
```
https://your-domain.com/telegram/webhook/123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
```

### Setting Webhook with Telegram

```bash
curl -X POST "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook" \
  -d "url=https://your-domain.com/telegram/webhook/<YOUR_BOT_TOKEN>"
```

### Verifying Webhook

```bash
curl "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getWebhookInfo"
```

---

## Examples

### Complete Workflow Example

```bash
# 1. Login and get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r '.access_token')

# 2. Create API key
API_KEY=$(curl -s -X POST http://localhost:8080/api/v1/apikeys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Key",
    "scopes": ["bots:read", "chats:read", "messages:send"],
    "expires_in_days": 365
  }' | jq -r '.key')

echo "API Key: $API_KEY"

# 3. Register a bot (using API key)
curl -X POST "http://localhost:8080/api/v1/bots?api_key=$API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "my_bot",
    "token": "123456:ABC-DEF...",
    "display_name": "My Bot"
  }'

# 4. List chats (using API key in header)
curl -H "X-API-Key: $API_KEY" \
  http://localhost:8080/api/v1/chats

# 5. Get messages (using token query param - Telegram style)
curl "http://localhost:8080/api/v1/chats/1/messages?token=$API_KEY&limit=50"

# 6. Send message (using API key in POST body - Telegram style)
curl -X POST http://localhost:8080/api/v1/chats/1/messages \
  -d "api_key=$API_KEY" \
  -d "text=Hello from the gateway!"

# 7. Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
```

### Python Example

```python
import requests

BASE_URL = "http://localhost:8080/api/v1"

# Login
response = requests.post(f"{BASE_URL}/auth/login", json={
    "username": "admin",
    "password": "password123"
})
tokens = response.json()
jwt_token = tokens["access_token"]

# Create API key
response = requests.post(
    f"{BASE_URL}/apikeys",
    headers={"Authorization": f"Bearer {jwt_token}"},
    json={
        "name": "Python Client",
        "scopes": ["bots:read", "messages:send"],
        "expires_in_days": 30
    }
)
api_key = response.json()["key"]

# List bots (using API key in header)
response = requests.get(
    f"{BASE_URL}/bots",
    headers={"X-API-Key": api_key}
)
bots = response.json()

# Get messages (using query parameter - Telegram style)
response = requests.get(
    f"{BASE_URL}/chats/1/messages",
    params={"api_key": api_key, "limit": 50}
)
messages = response.json()

# Send message (using POST body - Telegram style)
response = requests.post(
    f"{BASE_URL}/chats/1/messages",
    data={
        "token": api_key,
        "text": "Hello from Python!"
    }
)
result = response.json()
print(result)
```

### JavaScript Example

```javascript
const BASE_URL = 'http://localhost:8080/api/v1';

// Login
const loginResponse = await fetch(`${BASE_URL}/auth/login`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'admin',
    password: 'password123'
  })
});
const { access_token } = await loginResponse.json();

// Create API key
const apiKeyResponse = await fetch(`${BASE_URL}/apikeys`, {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${access_token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    name: 'JS Client',
    scopes: ['bots:read', 'messages:send'],
    expires_in_days: 30
  })
});
const { key: apiKey } = await apiKeyResponse.json();

// List bots (using query parameter)
const botsResponse = await fetch(
  `${BASE_URL}/bots?api_key=${apiKey}`
);
const bots = await botsResponse.json();

// Send message (using form data)
const formData = new FormData();
formData.append('token', apiKey);
formData.append('text', 'Hello from JavaScript!');

const sendResponse = await fetch(`${BASE_URL}/chats/1/messages`, {
  method: 'POST',
  body: formData
});
const result = await sendResponse.json();
console.log(result);
```

---

## Best Practices

### Security

1. **Use HTTPS in production** - Always use TLS/SSL
2. **Store API keys securely** - Never commit to version control
3. **Rotate API keys periodically** - Set expiration dates
4. **Use scoped API keys** - Grant minimum required permissions
5. **Validate webhook signatures** - Always verify HMAC signatures
6. **Rate limit your requests** - Respect rate limits

### Performance

1. **Use API keys for service-to-service** - Lower overhead than JWT
2. **Enable gzip compression** - Reduce bandwidth
3. **Implement pagination** - Don't fetch all data at once
4. **Use WebSocket for real-time** - More efficient than polling
5. **Cache frequently accessed data** - Reduce API calls

### Error Handling

1. **Check HTTP status codes** - Don't assume 200 OK
2. **Parse error responses** - Get detailed error information
3. **Implement retry logic** - With exponential backoff
4. **Handle rate limits** - Check X-RateLimit headers
5. **Log errors** - For debugging and monitoring

---

## Support

For issues, questions, or feature requests:
- Check this documentation
- Review error messages and logs
- Contact support team
- Report bugs on GitHub

---

**API Version**: 1.0.0
**Last Updated**: February 9, 2026
