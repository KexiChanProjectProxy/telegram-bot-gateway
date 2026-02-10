# API Documentation

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

The API supports two authentication methods:

1. **JWT Bearer Token**
   ```
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

2. **API Key**
   ```
   X-API-Key: tgw_abc123def456...
   ```

## Rate Limiting

All endpoints are rate-limited:
- **Default**: 100 requests per second per user/API key
- **Auth endpoints**: 10 requests per second per IP

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when the limit resets
- `Retry-After`: Seconds to wait if rate limited

## Endpoints

### Health & Metrics

#### GET /health
Health check endpoint (no auth required)

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": 1707440000,
  "websocket_clients": 5
}
```

#### GET /metrics
System metrics (requires auth)

**Response:**
```json
{
  "timestamp": 1707440000,
  "uptime": 3600.5,
  "system": {
    "goroutines": 42,
    "cpu_cores": 8,
    "memory": {
      "alloc_mb": 25,
      "heap_alloc_mb": 22
    }
  },
  "websocket": {
    "connected_clients": 5
  },
  "webhooks": {
    "pending_deliveries": 3
  }
}
```

---

### Authentication

#### POST /auth/login
Authenticate user and get JWT tokens

**Request:**
```json
{
  "username": "admin",
  "password": "your_password"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "random_base64_string",
  "expires_in": 900,
  "user": {
    "id": 1,
    "username": "admin",
    "is_active": true,
    "roles": ["admin"]
  }
}
```

#### POST /auth/refresh
Refresh access token

**Request:**
```json
{
  "refresh_token": "your_refresh_token"
}
```

**Response:** Same as login

#### POST /auth/logout
Revoke refresh token

**Request:**
```json
{
  "refresh_token": "your_refresh_token"
}
```

**Response:**
```json
{
  "message": "Logged out successfully"
}
```

---

### Bots

#### POST /bots
Register a new Telegram bot

**Auth:** Required
**Permission:** bots:create

**Request:**
```json
{
  "username": "my_awesome_bot",
  "token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
  "display_name": "My Awesome Bot",
  "description": "A bot for testing"
}
```

**Response:**
```json
{
  "id": 1,
  "username": "my_awesome_bot",
  "display_name": "My Awesome Bot",
  "description": "A bot for testing",
  "is_active": true
}
```

#### GET /bots
List all bots

**Auth:** Required
**Query Parameters:**
- `offset` (int): Pagination offset (default: 0)
- `limit` (int): Results per page (default: 50)

**Response:**
```json
[
  {
    "id": 1,
    "username": "my_bot",
    "display_name": "My Bot",
    "is_active": true
  }
]
```

#### GET /bots/:id
Get bot by ID

**Auth:** Required

**Response:**
```json
{
  "id": 1,
  "username": "my_bot",
  "display_name": "My Bot",
  "is_active": true,
  "webhook_url": "https://api.telegram.org/bot123456:ABC/setWebhook"
}
```

#### DELETE /bots/:id
Delete a bot

**Auth:** Required
**Permission:** bots:delete

**Response:**
```json
{
  "message": "Bot deleted successfully"
}
```

---

### Chats

#### GET /chats
List accessible chats

**Auth:** Required
**Query Parameters:**
- `offset` (int): Pagination offset
- `limit` (int): Results per page

**Response:**
```json
[
  {
    "id": 1,
    "bot_id": 1,
    "telegram_id": 123456789,
    "type": "private",
    "username": "john_doe",
    "first_name": "John",
    "is_active": true
  }
]
```

#### GET /chats/:id
Get chat by ID

**Auth:** Required

**Response:**
```json
{
  "id": 1,
  "bot_id": 1,
  "telegram_id": 123456789,
  "type": "private",
  "username": "john_doe",
  "first_name": "John",
  "is_active": true
}
```

#### GET /chats/:id/messages
Get messages from a chat

**Auth:** Required
**Permission:** can_read on chat
**Query Parameters:**
- `cursor` (RFC3339 timestamp): For pagination (messages older than this)
- `limit` (int): Results per page (default: 50, max: 100)

**Response:**
```json
[
  {
    "id": 1,
    "chat_id": 1,
    "telegram_id": 123,
    "from_username": "john_doe",
    "from_first_name": "John",
    "direction": "incoming",
    "message_type": "text",
    "text": "Hello, world!",
    "sent_at": "2024-02-09T12:00:00Z",
    "created_at": "2024-02-09T12:00:01Z"
  }
]
```

#### POST /chats/:id/messages
Send a message to a chat

**Auth:** Required
**Permission:** can_send on chat

**Request:**
```json
{
  "text": "Hello from the gateway!",
  "reply_to_message_id": 123
}
```

**Response:**
```json
{
  "message": "Message queued for delivery",
  "chat_id": 1,
  "text": "Hello from the gateway!",
  "queued_at": "2024-02-09T12:00:00Z"
}
```

---

### Webhooks

#### POST /webhooks
Register a new webhook

**Auth:** Required

**Request:**
```json
{
  "url": "https://your-app.com/webhooks/telegram",
  "scope": "chat",
  "chat_id": 1,
  "events": "[\"message\", \"edited_message\"]"
}
```

**Scopes:**
- `chat`: All messages in a chat
- `reply`: Only replies to a specific message (requires `reply_to_message_id`)

**Response:**
```json
{
  "id": 1,
  "url": "https://your-app.com/webhooks/telegram",
  "secret": "base64_encoded_secret",
  "scope": "chat",
  "chat_id": 1,
  "is_active": true
}
```

**Note:** Save the `secret` - it's only shown once and used for HMAC signature verification.

#### GET /webhooks
List webhooks for a chat

**Auth:** Required
**Query Parameters:**
- `chat_id` (required): Chat ID

**Response:**
```json
[
  {
    "id": 1,
    "url": "https://your-app.com/webhooks/telegram",
    "scope": "chat",
    "chat_id": 1,
    "is_active": true
  }
]
```

#### GET /webhooks/:id
Get webhook by ID

**Auth:** Required

#### PUT /webhooks/:id
Update webhook

**Auth:** Required

**Request:**
```json
{
  "url": "https://new-url.com/webhook",
  "is_active": true
}
```

#### DELETE /webhooks/:id
Delete webhook

**Auth:** Required

---

### API Keys

#### POST /apikeys
Create a new API key

**Auth:** Required
**Permission:** apikeys:create

**Request:**
```json
{
  "name": "Integration Key",
  "description": "For external service",
  "rate_limit": 5000
}
```

**Response:**
```json
{
  "id": 1,
  "key": "tgw_abc123def456...",
  "name": "Integration Key",
  "rate_limit": 5000,
  "is_active": true
}
```

**Note:** Save the `key` - it's only shown once!

#### GET /apikeys
List API keys

**Auth:** Required

**Response:**
```json
[
  {
    "id": 1,
    "key": "tgw_abc123...****",
    "name": "Integration Key",
    "rate_limit": 5000,
    "is_active": true
  }
]
```

#### POST /apikeys/:id/revoke
Revoke (deactivate) an API key

**Auth:** Required

#### DELETE /apikeys/:id
Permanently delete an API key

**Auth:** Required

---

### WebSocket

#### GET /ws
Upgrade to WebSocket connection

**Auth:** Required (via query parameter or header)
**Query Parameters:**
- `token` (optional): JWT token if not using Authorization header

**Protocol:**

Client → Server:
```json
{
  "action": "subscribe",
  "chat_id": 123
}
```

Server → Client (acknowledgment):
```json
{
  "type": "ack",
  "action": "subscribed",
  "chat_id": 123
}
```

Server → Client (new message):
```json
{
  "type": "message",
  "chat_id": 123,
  "message_id": 456,
  "telegram_id": 789,
  "text": "Hello!",
  "from_username": "john_doe",
  "timestamp": "2024-02-09T12:00:00Z"
}
```

**Actions:**
- `subscribe`: Subscribe to chat messages
- `unsubscribe`: Unsubscribe from chat
- `ping`: Ping server (responds with `pong`)

---

### Telegram Webhook

#### POST /telegram/webhook/:bot_username
Receive Telegram updates (called by Telegram, not clients)

**Auth:** None (validated by bot token in URL path)

**Request:** Standard Telegram Update object

**Response:**
```json
{
  "ok": true
}
```

---

## Webhook Payload Format

When the gateway delivers a webhook, it sends:

**Headers:**
- `Content-Type: application/json`
- `User-Agent: TelegramBotGateway/1.0`
- `X-Webhook-Signature: base64_hmac_sha256_signature`

**Body:**
```json
{
  "event": "message",
  "message_id": 1,
  "chat_id": 1,
  "telegram_id": 123,
  "text": "Hello!",
  "from_username": "john_doe",
  "from_first_name": "John",
  "direction": "incoming",
  "message_type": "text",
  "sent_at": "2024-02-09T12:00:00Z",
  "timestamp": 1707480000
}
```

**Signature Verification:**
```python
import hmac
import hashlib
import base64

def verify_signature(secret, payload, signature):
    expected = base64.b64encode(
        hmac.new(
            secret.encode(),
            payload.encode(),
            hashlib.sha256
        ).digest()
    ).decode()
    return hmac.compare_digest(expected, signature)
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request format"
}
```

### 401 Unauthorized
```json
{
  "error": "Authentication required"
}
```

### 403 Forbidden
```json
{
  "error": "Insufficient permissions for this chat"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found"
}
```

### 429 Too Many Requests
```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Limit: 100 requests per second. Try again after 2024-02-09T12:01:00Z",
  "retry_after": 1707480060
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```

---

## Examples

### Complete Flow: Register Bot and Receive Messages

```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r '.access_token')

# 2. Register bot
BOT_ID=$(curl -s -X POST http://localhost:8080/api/v1/bots \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "my_bot",
    "token": "123456:ABC-DEF",
    "display_name": "My Bot"
  }' | jq -r '.id')

# 3. Set up Telegram webhook
curl -X POST "https://api.telegram.org/bot123456:ABC-DEF/setWebhook" \
  -d "url=https://your-domain.com/api/v1/telegram/webhook/my_bot"

# 4. List chats (after receiving some messages)
curl -s http://localhost:8080/api/v1/chats \
  -H "Authorization: Bearer $TOKEN" | jq

# 5. Get messages
curl -s "http://localhost:8080/api/v1/chats/1/messages?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq
```

### WebSocket Example (JavaScript)

```javascript
const token = 'your_jwt_token';
const ws = new WebSocket(`ws://localhost:8080/api/v1/ws?token=${token}`);

ws.onopen = () => {
  console.log('Connected');
  // Subscribe to chat
  ws.send(JSON.stringify({
    action: 'subscribe',
    chat_id: 1
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);

  if (data.type === 'message') {
    console.log(`New message in chat ${data.chat_id}: ${data.text}`);
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected');
};
```

### Webhook Receiver Example (Node.js/Express)

```javascript
const express = require('express');
const crypto = require('crypto');

const app = express();
app.use(express.json());

const WEBHOOK_SECRET = 'your_webhook_secret';

function verifySignature(payload, signature) {
  const expected = crypto
    .createHmac('sha256', WEBHOOK_SECRET)
    .update(JSON.stringify(payload))
    .digest('base64');
  return crypto.timingSafeEqual(
    Buffer.from(expected),
    Buffer.from(signature)
  );
}

app.post('/webhooks/telegram', (req, res) => {
  const signature = req.headers['x-webhook-signature'];

  if (!verifySignature(req.body, signature)) {
    return res.status(401).json({ error: 'Invalid signature' });
  }

  console.log('New message:', req.body);
  res.json({ ok: true });
});

app.listen(3000);
```
