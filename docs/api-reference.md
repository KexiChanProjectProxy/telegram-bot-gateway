# API Reference

This document provides a complete reference for the Telegram Bot Gateway REST API, WebSocket API, and webhook system.

## Base URL

```
http://your-domain.com/api/v1
```

All API endpoints use JSON for request and response bodies unless otherwise specified.

## Authentication

The gateway supports three authentication methods:

### JWT Bearer Token

Use for interactive user sessions with login/logout support.

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Example:
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/bots
```

### API Key - Header Method

Use for machine-to-machine communication and service integrations.

```
X-API-Key: tgw_1234567890abcdef
```

Example:
```bash
curl -H "X-API-Key: tgw_1234567890abcdef" \
  http://localhost:8080/api/v1/bots
```

### API Key - Telegram Bot Style

Supports API keys via query parameters or POST body, similar to Telegram Bot API format.

Query parameter:
```bash
# Using 'api_key' parameter
curl "http://localhost:8080/api/v1/bots?api_key=tgw_1234567890abcdef"

# OR using 'token' parameter (alias)
curl "http://localhost:8080/api/v1/bots?token=tgw_1234567890abcdef"
```

POST body:
```bash
# Form-data with 'api_key'
curl -X POST "http://localhost:8080/api/v1/chats/1/messages" \
  -d "api_key=tgw_1234567890abcdef" \
  -d "text=Hello World"

# JSON with query parameter
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
| Access Token (JWT) | 15 minutes | Yes (via refresh) |
| Refresh Token (JWT) | 7 days | No |
| API Key | Configurable | N/A |

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

## HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required or failed |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource already exists |
| 422 | Unprocessable Entity | Validation error |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service temporarily unavailable |

## Rate Limiting

Every response includes rate limit headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1644408600
```

Default limits:

| Endpoint Type | Limit | Window |
|---------------|-------|--------|
| Public (login, refresh) | 10 requests | Per IP per minute |
| Authenticated | 100 requests | Per user/API key per minute |
| Global (optional) | 1000 requests | Per minute |

Rate limit exceeded response:
```json
{
  "error": "Rate limit exceeded. Try again in 30 seconds.",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 30
}
```

## API Endpoints

### Health Check

#### GET /health

Health check endpoint (no authentication required).

Response (200):
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": 1644408600,
  "websocket_clients": 42
}
```

Example:
```bash
curl http://localhost:8080/health
```

### Authentication Endpoints

#### POST /api/v1/auth/login

Login and get JWT tokens.

Request:
```json
{
  "username": "admin",
  "password": "password123"
}
```

Response (200):
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

Example:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}'
```

#### POST /api/v1/auth/refresh

Refresh access token using refresh token.

Request:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Response (200):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

Example:
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

#### POST /api/v1/auth/logout

Logout and revoke refresh token.

Request:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Response (200):
```json
{
  "message": "Logged out successfully"
}
```

Example:
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

### Bot Management Endpoints

Note: Bot creation and deletion are CLI-only operations. Use the `telegram-bot-gateway bot` commands.

#### GET /api/v1/bots

List all registered bots.

Authentication: Required

Query Parameters:
- `offset` (optional, default: 0) - Pagination offset
- `limit` (optional, default: 20, max: 100) - Results per page

Response (200):
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

Examples:
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

#### GET /api/v1/bots/:id

Get details of a specific bot.

Authentication: Required

Response (200):
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

Examples:
```bash
# Using JWT
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/bots/1

# Using API Key (Query)
curl "http://localhost:8080/api/v1/bots/1?api_key=tgw_1234567890abcdef"
```

### Chat Management Endpoints

#### GET /api/v1/chats

List all accessible chats.

Authentication: Required

Query Parameters:
- `bot_id` (optional) - Filter by bot ID
- `offset` (optional, default: 0)
- `limit` (optional, default: 20, max: 100)

Response (200):
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

Examples:
```bash
# List all chats
curl "http://localhost:8080/api/v1/chats?api_key=tgw_xxx"

# Filter by bot
curl "http://localhost:8080/api/v1/chats?api_key=tgw_xxx&bot_id=1"
```

#### GET /api/v1/chats/:id

Get chat details.

Authentication: Required

Response (200):
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

Example:
```bash
curl "http://localhost:8080/api/v1/chats/1?token=tgw_xxx"
```

#### GET /api/v1/chats/:id/messages

Get messages from a chat. Requires `can_read` permission for the chat.

Authentication: Required
Authorization: Chat-level ACL (can_read)

Query Parameters:
- `cursor` (optional) - Unix timestamp for pagination (get messages before this time)
- `limit` (optional, default: 50, max: 100)

Response (200):
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

Examples:
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

#### POST /api/v1/chats/:id/messages

Send a message to a chat. Requires `can_send` permission for the chat.

Authentication: Required
Authorization: Chat-level ACL (can_send)

Request:
```json
{
  "text": "Hello from the gateway!",
  "reply_to_message_id": 1001
}
```

Response (201):
```json
{
  "success": true,
  "message": "Message queued for delivery",
  "message_id": 2,
  "queued_at": 1644408600
}
```

Examples:
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

### Webhook Management Endpoints

#### POST /api/v1/webhooks

Register a new webhook URL.

Authentication: Required

Request:
```json
{
  "url": "https://my-app.com/telegram/updates",
  "chat_ids": [1, 2, 3],
  "event_types": ["new_message", "edited_message"],
  "scope": "chat",
  "is_active": true
}
```

Response (201):
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

Note: All webhook deliveries include an `X-Webhook-Signature` header with HMAC-SHA256 signature. Verify using the `secret`.

#### GET /api/v1/webhooks

List registered webhooks.

Authentication: Required

Response (200):
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

#### GET /api/v1/webhooks/:id

Get webhook details.

Authentication: Required

Response (200):
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

#### PUT /api/v1/webhooks/:id

Update webhook configuration.

Authentication: Required

Request:
```json
{
  "url": "https://my-app.com/new-endpoint",
  "chat_ids": [1, 2, 3, 4],
  "event_types": ["new_message"],
  "is_active": true
}
```

Response (200):
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

#### DELETE /api/v1/webhooks/:id

Delete a webhook.

Authentication: Required

Response (200):
```json
{
  "message": "Webhook deleted successfully"
}
```

### Webhook Delivery Format

When events occur, the gateway delivers webhooks to registered URLs via HTTP POST with the following format:

Headers:
```
Content-Type: application/json
X-Webhook-Signature: sha256=abc123...
X-Webhook-Event: new_message
X-Webhook-Delivery: uuid-1234
```

Payload:
```json
{
  "event": "new_message",
  "chat_id": 1,
  "message": {
    "id": 123,
    "telegram_id": 1001,
    "bot_id": 1,
    "direction": "incoming",
    "text": "Hello from Telegram!",
    "from_username": "john_doe",
    "from_first_name": "John",
    "message_type": "text",
    "timestamp": "2026-02-09T12:00:00Z"
  }
}
```

### Webhook HMAC Verification

Verify webhook authenticity using the HMAC-SHA256 signature:

```python
import hmac
import hashlib

def verify_webhook(secret, signature, body):
    expected = hmac.new(
        secret.encode(),
        body.encode(),
        hashlib.sha256
    ).hexdigest()

    provided = signature.replace('sha256=', '')
    return hmac.compare_digest(expected, provided)
```

```go
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "strings"
)

func verifyWebhook(secret, signature string, body []byte) bool {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(body)
    expected := hex.EncodeToString(h.Sum(nil))

    provided := strings.TrimPrefix(signature, "sha256=")
    return hmac.Equal([]byte(expected), []byte(provided))
}
```

## WebSocket API

### Connection

Endpoint: `ws://your-domain.com/api/v1/ws`

Authentication: Required (JWT or API Key via query parameter)

```javascript
// Using JWT
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?token=' + jwtToken);

// Using API Key
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?api_key=tgw_xxx');
```

### Client Messages

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

Response:
```json
{
  "action": "pong",
  "timestamp": 1644408600
}
```

### Server Messages

New message event:
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

## Telegram Webhook Receiver

### Receiving Updates from Telegram

Endpoint: `POST /api/v1/telegram/webhook/:webhook_secret`

No authentication required - validated by webhook secret in URL.

Telegram sends updates to a URL with a cryptographically random secret (64 characters) that is generated when creating a bot via CLI:
```
https://your-domain.com/api/v1/telegram/webhook/a1b2c3d4e5f6789012345678901234567890123456789012345678901234
```

The webhook URL is automatically registered with Telegram when you create a bot using the CLI tool:
```bash
./bin/bot create --username my_bot --token "123456:ABC-DEF..."
```

For security, the webhook secret is:
- Randomly generated (32 bytes, 64 hex characters)
- Unique per bot
- Unguessable (2^256 possible values)
- Automatically configured with Telegram's setWebhook API

### Manual Webhook Management

If you need to re-register a webhook manually, use the `setWebhook` method in the bot service (requires decrypted bot token).

### Verifying Webhook

Check webhook status via Telegram API:
```bash
curl "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getWebhookInfo"
```

## Code Examples

### Complete Workflow (Bash)

```bash
# 1. Login and get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r '.access_token')

# 2. List bots
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/bots

# 3. List chats
curl -H "X-API-Key: tgw_xxx" \
  http://localhost:8080/api/v1/chats

# 4. Get messages (using token query param)
curl "http://localhost:8080/api/v1/chats/1/messages?token=tgw_xxx&limit=50"

# 5. Send message (using API key in POST body)
curl -X POST http://localhost:8080/api/v1/chats/1/messages \
  -d "api_key=tgw_xxx" \
  -d "text=Hello from the gateway!"

# 6. Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
```

### Python Client

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

# List bots
response = requests.get(
    f"{BASE_URL}/bots",
    headers={"Authorization": f"Bearer {jwt_token}"}
)
bots = response.json()

# Get messages (using query parameter)
response = requests.get(
    f"{BASE_URL}/chats/1/messages",
    params={"api_key": "tgw_xxx", "limit": 50}
)
messages = response.json()

# Send message (using POST body)
response = requests.post(
    f"{BASE_URL}/chats/1/messages",
    data={
        "token": "tgw_xxx",
        "text": "Hello from Python!"
    }
)
result = response.json()
print(result)
```

### Go Client

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`
}

func main() {
    baseURL := "http://localhost:8080/api/v1"

    // Login
    loginReq := LoginRequest{
        Username: "admin",
        Password: "password123",
    }
    body, _ := json.Marshal(loginReq)

    resp, err := http.Post(
        baseURL+"/auth/login",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var loginResp LoginResponse
    json.NewDecoder(resp.Body).Decode(&loginResp)

    // List bots
    req, _ := http.NewRequest("GET", baseURL+"/bots", nil)
    req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)

    client := &http.Client{}
    resp, _ = client.Do(req)
    defer resp.Body.Close()

    // Send message (using API key in query)
    data := url.Values{}
    data.Set("api_key", "tgw_xxx")
    data.Set("text", "Hello from Go!")

    resp, _ = http.PostForm(
        baseURL+"/chats/1/messages",
        data,
    )
    defer resp.Body.Close()

    fmt.Println("Message sent successfully")
}
```

### JavaScript/Node.js Client

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

// List bots (using query parameter)
const botsResponse = await fetch(
  `${BASE_URL}/bots?api_key=tgw_xxx`
);
const bots = await botsResponse.json();

// Send message (using form data)
const formData = new FormData();
formData.append('token', 'tgw_xxx');
formData.append('text', 'Hello from JavaScript!');

const sendResponse = await fetch(`${BASE_URL}/chats/1/messages`, {
  method: 'POST',
  body: formData
});
const result = await sendResponse.json();
console.log(result);
```

## Best Practices

### Security

1. Use HTTPS in production - Always use TLS/SSL
2. Store API keys securely - Never commit to version control
3. Rotate API keys periodically - Set expiration dates
4. Use scoped API keys - Grant minimum required permissions
5. Validate webhook signatures - Always verify HMAC signatures
6. Rate limit your requests - Respect rate limits

### Performance

1. Use API keys for service-to-service - Lower overhead than JWT
2. Enable gzip compression - Reduce bandwidth
3. Implement pagination - Don't fetch all data at once
4. Use WebSocket for real-time - More efficient than polling
5. Cache frequently accessed data - Reduce API calls

### Error Handling

1. Check HTTP status codes - Don't assume 200 OK
2. Parse error responses - Get detailed error information
3. Implement retry logic - With exponential backoff
4. Handle rate limits - Check X-RateLimit headers
5. Log errors - For debugging and monitoring
