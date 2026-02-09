# API Authentication Quick Reference

## Supported Authentication Methods

The Telegram Bot Gateway supports **flexible authentication** similar to Telegram Bot API.

### Method 1: JWT Bearer Token (User Sessions)

**Best for**: Interactive user sessions with login/logout

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/bots
```

---

### Method 2: API Key Header (Standard)

**Best for**: Service-to-service communication

```bash
curl -H "X-API-Key: tgw_1234567890abcdef" \
  http://localhost:8080/api/v1/bots
```

---

### Method 3a: API Key Query Parameter (Telegram Bot Style)

**Best for**: GET requests, simple integrations, Telegram-like usage

```bash
# Using 'api_key' parameter
curl "http://localhost:8080/api/v1/bots?api_key=tgw_1234567890abcdef"

# OR using 'token' parameter (alias)
curl "http://localhost:8080/api/v1/bots?token=tgw_1234567890abcdef"

# Works with other parameters
curl "http://localhost:8080/api/v1/chats/1/messages?token=tgw_xxx&limit=50"
```

---

### Method 3b: API Key POST Body (Telegram Bot Style)

**Best for**: POST requests, form submissions

#### Form-Data / URL-Encoded

```bash
# Using 'api_key' parameter
curl -X POST http://localhost:8080/api/v1/chats/1/messages \
  -d "api_key=tgw_1234567890abcdef" \
  -d "text=Hello World"

# OR using 'token' parameter (alias)
curl -X POST http://localhost:8080/api/v1/chats/1/messages \
  -d "token=tgw_1234567890abcdef" \
  -d "text=Hello World"
```

#### JSON with Query Parameter

```bash
curl -X POST "http://localhost:8080/api/v1/chats/1/messages?api_key=tgw_xxx" \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello World"}'
```

---

## Authentication Priority

The gateway checks credentials in this order:

1. ✅ **JWT Bearer token** (Authorization header)
2. ✅ **X-API-Key header**
3. ✅ **Query parameter** (`?api_key=xxx` or `?token=xxx`)
4. ✅ **POST body field** (`api_key` or `token`)

**First valid credential found is used.**

---

## Examples by Language

### cURL

```bash
# Header method
curl -H "X-API-Key: tgw_xxx" http://localhost:8080/api/v1/bots

# Query method (GET)
curl "http://localhost:8080/api/v1/bots?api_key=tgw_xxx"

# POST body method
curl -X POST http://localhost:8080/api/v1/chats/1/messages \
  -d "token=tgw_xxx" \
  -d "text=Hello"
```

### Python

```python
import requests

API_KEY = "tgw_1234567890abcdef"

# Method 1: Header
response = requests.get(
    "http://localhost:8080/api/v1/bots",
    headers={"X-API-Key": API_KEY}
)

# Method 2: Query parameter
response = requests.get(
    "http://localhost:8080/api/v1/bots",
    params={"api_key": API_KEY}
)

# Method 3: POST body
response = requests.post(
    "http://localhost:8080/api/v1/chats/1/messages",
    data={
        "token": API_KEY,
        "text": "Hello"
    }
)
```

### JavaScript/Node.js

```javascript
const API_KEY = 'tgw_1234567890abcdef';

// Method 1: Header
const response1 = await fetch('http://localhost:8080/api/v1/bots', {
  headers: { 'X-API-Key': API_KEY }
});

// Method 2: Query parameter
const response2 = await fetch(
  `http://localhost:8080/api/v1/bots?api_key=${API_KEY}`
);

// Method 3: POST body (form data)
const formData = new FormData();
formData.append('token', API_KEY);
formData.append('text', 'Hello');

const response3 = await fetch('http://localhost:8080/api/v1/chats/1/messages', {
  method: 'POST',
  body: formData
});
```

### PHP

```php
<?php
$apiKey = 'tgw_1234567890abcdef';

// Method 1: Header
$ch = curl_init('http://localhost:8080/api/v1/bots');
curl_setopt($ch, CURLOPT_HTTPHEADER, ["X-API-Key: $apiKey"]);
$response = curl_exec($ch);

// Method 2: Query parameter
$url = "http://localhost:8080/api/v1/bots?api_key=" . urlencode($apiKey);
$response = file_get_contents($url);

// Method 3: POST body
$data = [
    'token' => $apiKey,
    'text' => 'Hello'
];
$ch = curl_init('http://localhost:8080/api/v1/chats/1/messages');
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($data));
$response = curl_exec($ch);
?>
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
    // Method 1: Header
    req1, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/bots", nil)
    req1.Header.Set("X-API-Key", apiKey)
    resp1, _ := http.DefaultClient.Do(req1)

    // Method 2: Query parameter
    u, _ := url.Parse("http://localhost:8080/api/v1/bots")
    q := u.Query()
    q.Set("api_key", apiKey)
    u.RawQuery = q.Encode()
    resp2, _ := http.Get(u.String())

    // Method 3: POST body
    form := url.Values{}
    form.Set("token", apiKey)
    form.Set("text", "Hello")
    resp3, _ := http.Post(
        "http://localhost:8080/api/v1/chats/1/messages",
        "application/x-www-form-urlencoded",
        strings.NewReader(form.Encode()),
    )
}
```

### Ruby

```ruby
require 'net/http'
require 'uri'

API_KEY = 'tgw_1234567890abcdef'

# Method 1: Header
uri = URI('http://localhost:8080/api/v1/bots')
req = Net::HTTP::Get.new(uri)
req['X-API-Key'] = API_KEY
response = Net::HTTP.start(uri.hostname, uri.port) { |http| http.request(req) }

# Method 2: Query parameter
uri = URI('http://localhost:8080/api/v1/bots')
uri.query = URI.encode_www_form({ api_key: API_KEY })
response = Net::HTTP.get_response(uri)

# Method 3: POST body
uri = URI('http://localhost:8080/api/v1/chats/1/messages')
response = Net::HTTP.post_form(uri, {
  token: API_KEY,
  text: 'Hello'
})
```

---

## Why Telegram Bot API Style?

Supporting API keys via query parameters and POST body (like Telegram Bot API) provides:

✅ **Simplicity** - Easy to use in browsers, webhooks, and simple scripts
✅ **Compatibility** - Familiar to Telegram bot developers
✅ **Flexibility** - Works with clients that don't support custom headers
✅ **Convenience** - No need to set up authentication headers for simple requests

**Example**: Simple webhook callback
```
https://my-gateway.com/api/v1/chats/1/messages?token=tgw_xxx&text=Alert
```

---

## Security Considerations

### ✅ Recommended Practices

1. **Use HTTPS in production** - Always encrypt API keys in transit
2. **Store API keys securely** - Use environment variables or secret managers
3. **Set expiration dates** - Rotate keys periodically
4. **Use scoped keys** - Grant minimum required permissions
5. **Prefer headers in production** - More secure than query parameters

### ⚠️ Query Parameter Security

While convenient, query parameters have security considerations:

- **Logged in server logs** - Query strings may be logged
- **Visible in browser history** - If used in browser
- **Visible in referrer headers** - May leak to third-party sites

**Recommendation**:
- Use **query parameters** for development and simple integrations
- Use **headers** for production and sensitive operations

---

## Creating API Keys

```bash
# 1. Login to get JWT token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# Response:
# {
#   "access_token": "eyJhbGc...",
#   "refresh_token": "eyJhbGc..."
# }

# 2. Create API key (requires JWT)
curl -X POST http://localhost:8080/api/v1/apikeys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Key",
    "scopes": ["bots:read", "chats:read", "messages:send"],
    "expires_in_days": 365
  }'

# Response:
# {
#   "id": 1,
#   "key": "tgw_1234567890abcdef1234567890abcdef",
#   "name": "Production Key",
#   ...
# }

# 3. Use the API key (save it - shown only once!)
API_KEY="tgw_1234567890abcdef1234567890abcdef"

# 4. Test the API key
curl "http://localhost:8080/api/v1/bots?api_key=$API_KEY"
```

---

## Troubleshooting

### Error: "Authentication required"

**Cause**: No valid credentials provided

**Solution**: Ensure you're sending the API key via header, query, or POST body

```bash
# ❌ Wrong
curl http://localhost:8080/api/v1/bots

# ✅ Correct (any of these work)
curl -H "X-API-Key: tgw_xxx" http://localhost:8080/api/v1/bots
curl "http://localhost:8080/api/v1/bots?api_key=tgw_xxx"
curl -X POST http://localhost:8080/api/v1/bots -d "token=tgw_xxx"
```

### Error: "Invalid API key format"

**Cause**: API key doesn't start with `tgw_` prefix

**Solution**: Use the full API key including prefix

```bash
# ❌ Wrong
curl "http://localhost:8080/api/v1/bots?api_key=1234567890"

# ✅ Correct
curl "http://localhost:8080/api/v1/bots?api_key=tgw_1234567890abcdef"
```

### Error: "Invalid API key"

**Cause**: API key not found in database or incorrect

**Solution**: Verify the API key is correct and active

```bash
# List your API keys (requires JWT)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/apikeys
```

### Error: "API key is inactive"

**Cause**: API key was revoked

**Solution**: Create a new API key or reactivate the old one

### Error: "API key has expired"

**Cause**: API key passed expiration date

**Solution**: Create a new API key with a new expiration date

---

## Quick Reference

| Method | Usage | Security | Best For |
|--------|-------|----------|----------|
| **JWT Bearer** | `Authorization: Bearer xxx` | ⭐⭐⭐⭐⭐ | User sessions |
| **Header** | `X-API-Key: tgw_xxx` | ⭐⭐⭐⭐⭐ | Production APIs |
| **Query** | `?api_key=tgw_xxx` | ⭐⭐⭐ | Development, simple tools |
| **POST Body** | `api_key=tgw_xxx` | ⭐⭐⭐⭐ | Forms, webhooks |

---

**Last Updated**: February 9, 2026
**API Version**: 1.0.0
