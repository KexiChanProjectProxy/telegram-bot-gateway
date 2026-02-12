# Authentication

The Telegram Bot Gateway provides three flexible authentication methods for different use cases, from interactive user sessions to automated service integrations.

## Overview

| Method | Header/Parameter | Security | Best For |
|--------|-----------------|----------|----------|
| **JWT Bearer Token** | `Authorization: Bearer xxx` | High | User sessions with login/logout |
| **API Key Header** | `X-API-Key: tgw_xxx` | High | Production APIs, service-to-service |
| **API Key Query/Body** | `?api_key=tgw_xxx` or form field | Medium | Development, webhooks, simple tools |

## JWT Authentication

JWT (JSON Web Token) authentication is designed for interactive user sessions. It provides secure access with automatic token expiration and refresh capabilities.

### Login Flow

1. User authenticates with username and password
2. Server returns an access token (short-lived) and refresh token (long-lived)
3. Client includes access token in subsequent requests
4. When access token expires, use refresh token to obtain a new access token

### Obtaining Tokens

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "bearer",
  "expires_in": 3600
}
```

### Using JWT Tokens

Include the access token in the Authorization header with the "Bearer" prefix:

```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  http://localhost:8080/api/v1/bots
```

### Refreshing Tokens

When the access token expires, use the refresh token to obtain a new one:

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

## API Key Authentication

API keys provide static credentials for machine-to-machine communication and automated services. They support granular permissions and multiple delivery methods.

### Creating API Keys

API keys are created using the CLI tool. REST API endpoints for key creation are disabled for security.

```bash
# Build the CLI tool
cd services/gateway
go build -o bin/apikey cmd/apikey/main.go

# Create a new API key
./bin/apikey create --name "Production Service" \
  --rate-limit 5000 \
  --expires 1y
```

Output:
```
API key created successfully!

API Key ID: 1
API Key:    tgw_1234567890abcdef1234567890abcdef

IMPORTANT: Save this key now! It cannot be retrieved later.

Details:
  Name:        Production Service
  Rate Limit:  5000 requests/hour
  Expires:     2027-02-12 10:30:00
```

### Using API Keys

API keys can be delivered in three ways, providing flexibility for different client types and use cases.

#### Method 1: Header (Recommended for Production)

The most secure method. Include the API key in the X-API-Key header:

```bash
curl -H "X-API-Key: tgw_1234567890abcdef" \
  http://localhost:8080/api/v1/bots
```

#### Method 2: Query Parameter

Convenient for GET requests, webhooks, and simple integrations. Use either `api_key` or `token` parameter:

```bash
# Using api_key parameter
curl "http://localhost:8080/api/v1/bots?api_key=tgw_1234567890abcdef"

# Using token parameter (alias)
curl "http://localhost:8080/api/v1/bots?token=tgw_1234567890abcdef"

# Works with other parameters
curl "http://localhost:8080/api/v1/chats/1/messages?token=tgw_xxx&limit=50"
```

#### Method 3: POST Body

Useful for form submissions and POST requests. Supports both form-data and URL-encoded formats:

```bash
# Form-data / URL-encoded
curl -X POST http://localhost:8080/api/v1/chats/1/messages \
  -d "api_key=tgw_1234567890abcdef" \
  -d "text=Hello World"

# JSON body with query parameter
curl -X POST "http://localhost:8080/api/v1/chats/1/messages?api_key=tgw_xxx" \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello World"}'
```

### Managing API Keys

List all API keys:
```bash
./bin/apikey list
```

View details and permissions:
```bash
./bin/apikey get 1
./bin/apikey show-permissions 1
```

Deactivate an API key (reversible):
```bash
./bin/apikey revoke 1
```

Permanently delete an API key:
```bash
./bin/apikey delete 1
```

## Permission Model

The gateway implements granular access control for both users and API keys.

### Chat Permissions

Control which chats can be accessed and what actions are allowed:

- **Read**: View chat details and message history
- **Send**: Send messages to the chat
- **Manage**: Administrative operations (updating chat settings)

Grant chat permissions to an API key:
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
./bin/apikey revoke-chat 1 5
```

### Bot Restrictions

Control which bots an API key can use for sending messages.

By default, API keys can use all bots. Once you grant bot permission, the API key becomes restricted to only the explicitly allowed bots:

```bash
# Allow API key 1 to use bot 2
./bin/apikey grant-bot 1 2

# Allow API key 1 to also use bot 3
./bin/apikey grant-bot 1 3

# Now API key 1 can ONLY use bots 2 and 3
```

Revoke bot permission:
```bash
./bin/apikey revoke-bot 1 2
```

### Feedback Permissions

Control which chats can push messages back to the API key holder (for webhooks and callbacks).

By default, API keys can receive feedback from all chats. Once you grant feedback permission, the API key becomes restricted to only receive from explicitly allowed chats:

```bash
# Allow API key 1 to receive messages from chat 5
./bin/apikey grant-feedback 1 5

# Now API key 1 can ONLY receive feedback from chat 5
```

Revoke feedback permission:
```bash
./bin/apikey revoke-feedback 1 5
```

## Authentication Priority

When multiple credentials are provided, the gateway checks them in this order:

1. JWT Bearer token (Authorization header)
2. X-API-Key header
3. Query parameter (`?api_key=xxx` or `?token=xxx`)
4. POST body field (`api_key` or `token`)

The first valid credential found is used. This allows fallback mechanisms but ensures JWT authentication takes precedence.

## Security Recommendations

### General Best Practices

1. **Use HTTPS in production** - Always encrypt credentials in transit
2. **Store credentials securely** - Use environment variables or secret managers, never hardcode
3. **Set expiration dates** - Rotate API keys periodically
4. **Implement least privilege** - Grant only the minimum permissions required
5. **Monitor usage** - Regularly audit active keys and revoke unused ones

### JWT-Specific

1. **Short access token lifetime** - Default 1 hour limits exposure if compromised
2. **Secure refresh token storage** - Store refresh tokens securely, never in localStorage
3. **Implement logout** - Properly invalidate refresh tokens on logout

### API Key-Specific

1. **Prefer headers in production** - More secure than query parameters
2. **Use granular permissions** - Restrict access to specific chats and bots
3. **Set rate limits** - Prevent abuse with appropriate rate limiting
4. **Rotate keys regularly** - Create new keys with expiration dates

### Query Parameter Considerations

While convenient, query parameters have security trade-offs:

- Logged in server access logs
- Visible in browser history
- May leak via referrer headers to third-party sites

Use query parameters for:
- Development and testing
- Simple integrations without custom header support
- Webhooks from trusted sources

Use headers for:
- Production environments
- Sensitive operations
- Public-facing APIs

## Code Examples

### cURL

```bash
# JWT authentication
curl -H "Authorization: Bearer eyJhbGc..." \
  http://localhost:8080/api/v1/bots

# API key header
curl -H "X-API-Key: tgw_xxx" \
  http://localhost:8080/api/v1/bots

# API key query parameter
curl "http://localhost:8080/api/v1/bots?api_key=tgw_xxx"

# API key POST body
curl -X POST http://localhost:8080/api/v1/chats/1/messages \
  -d "token=tgw_xxx" \
  -d "text=Hello"
```

### Python

```python
import requests

# JWT authentication
headers = {"Authorization": f"Bearer {access_token}"}
response = requests.get("http://localhost:8080/api/v1/bots", headers=headers)

# API key header
headers = {"X-API-Key": "tgw_1234567890abcdef"}
response = requests.get("http://localhost:8080/api/v1/bots", headers=headers)

# API key query parameter
params = {"api_key": "tgw_1234567890abcdef"}
response = requests.get("http://localhost:8080/api/v1/bots", params=params)

# API key POST body
data = {"token": "tgw_1234567890abcdef", "text": "Hello"}
response = requests.post("http://localhost:8080/api/v1/chats/1/messages", data=data)
```

### Go

```go
package main

import (
    "net/http"
    "net/url"
    "strings"
)

const apiKey = "tgw_1234567890abcdef"

func main() {
    // JWT authentication
    req1, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/bots", nil)
    req1.Header.Set("Authorization", "Bearer "+jwtToken)
    resp1, _ := http.DefaultClient.Do(req1)

    // API key header
    req2, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/bots", nil)
    req2.Header.Set("X-API-Key", apiKey)
    resp2, _ := http.DefaultClient.Do(req2)

    // API key query parameter
    u, _ := url.Parse("http://localhost:8080/api/v1/bots")
    q := u.Query()
    q.Set("api_key", apiKey)
    u.RawQuery = q.Encode()
    resp3, _ := http.Get(u.String())

    // API key POST body
    form := url.Values{}
    form.Set("token", apiKey)
    form.Set("text", "Hello")
    resp4, _ := http.Post(
        "http://localhost:8080/api/v1/chats/1/messages",
        "application/x-www-form-urlencoded",
        strings.NewReader(form.Encode()),
    )
}
```

## Troubleshooting

### Error: "Authentication required"

No valid credentials were provided.

Solution: Ensure you're sending credentials via Authorization header, X-API-Key header, query parameter, or POST body.

```bash
# Wrong - no credentials
curl http://localhost:8080/api/v1/bots

# Correct - API key in header
curl -H "X-API-Key: tgw_xxx" http://localhost:8080/api/v1/bots
```

### Error: "Invalid API key format"

The API key doesn't match the expected format (must start with `tgw_` prefix).

Solution: Use the full API key including the prefix.

```bash
# Wrong - missing prefix
curl "http://localhost:8080/api/v1/bots?api_key=1234567890"

# Correct - full key with prefix
curl "http://localhost:8080/api/v1/bots?api_key=tgw_1234567890abcdef"
```

### Error: "Invalid API key"

The API key was not found in the database or the hash doesn't match.

Solution: Verify the API key is correct and hasn't been deleted.

```bash
# List all API keys
./bin/apikey list

# Get specific key details
./bin/apikey get 1
```

### Error: "API key is inactive"

The API key has been revoked.

Solution: Create a new API key or contact an administrator to reactivate.

### Error: "API key has expired"

The API key passed its expiration date.

Solution: Create a new API key with a new expiration date.

### Error: "Insufficient permissions for this chat"

The API key doesn't have the required chat permissions.

Solution: Grant appropriate chat permissions.

```bash
# Check current permissions
./bin/apikey show-permissions 1

# Grant read and send access
./bin/apikey grant-chat 1 5 --read --send
```

### Error: "API key not authorized for this bot"

The API key has bot restrictions and the requested bot is not allowed.

Solution: Grant bot permission or remove restrictions.

```bash
# Check bot restrictions
./bin/apikey show-permissions 1

# Grant permission to use bot 2
./bin/apikey grant-bot 1 2
```
