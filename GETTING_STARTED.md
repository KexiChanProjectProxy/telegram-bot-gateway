# Getting Started Guide

## Prerequisites

- Go 1.21 or higher
- MySQL 8.0+ or MariaDB 10.5+
- Redis 6.0+
- Docker and Docker Compose (optional, for easy setup)

## Quick Start with Docker

The fastest way to get started is using Docker Compose:

```bash
# 1. Clone the repository
git clone https://github.com/kexi/telegram-bot-gateway.git
cd telegram-bot-gateway

# 2. Copy environment variables
cp .env.example .env

# 3. Edit .env and set your secrets
nano .env  # or use your preferred editor

# 4. Start all services
docker-compose up -d

# 5. Check service health
curl http://localhost:8080/health
```

The gateway will be available at:
- **HTTP API**: http://localhost:8080
- **gRPC**: localhost:9090 (when implemented)

## Manual Setup

### Step 1: Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools (optional)
make install-tools
```

### Step 2: Configure Database

```bash
# Start MySQL (if not using Docker)
# Then create the database
mysql -u root -p

CREATE DATABASE telegram_gateway CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'gateway'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON telegram_gateway.* TO 'gateway'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### Step 3: Configure Redis

```bash
# Start Redis (if not using Docker)
redis-server

# Or install and start as a service
sudo systemctl start redis
```

### Step 4: Set Environment Variables

```bash
# Copy the example file
cp .env.example .env

# Edit with your values
nano .env
```

Required environment variables:
- `DB_PASSWORD`: Your MySQL password
- `JWT_SECRET`: A secure random string (min 32 chars)
- `WEBHOOK_BASE_URL`: Your public domain for Telegram webhooks

### Step 5: Run Migrations

```bash
# Run database migrations
make migrate

# Or manually:
go run cmd/migrate/main.go up
```

### Step 6: Create Admin User

```bash
# Set environment variables
export DEFAULT_ADMIN_USER=admin
export DEFAULT_ADMIN_PASSWORD=your_secure_password

# Start the gateway (it will create the admin user)
go run cmd/gateway/main.go
```

Or create manually via SQL:

```sql
-- First, hash your password with bcrypt (cost 10)
-- Then insert the user
INSERT INTO users (username, password, is_active)
VALUES ('admin', '$2a$10$...your_bcrypt_hash...', true);

-- Assign admin role
INSERT INTO user_roles (user_id, role_id)
VALUES (1, 1);  -- Assuming user ID 1 and admin role ID 1
```

### Step 7: Start the Gateway

```bash
# Using make
make run

# Or directly with go
go run cmd/gateway/main.go

# Or build and run
make build
./bin/gateway
```

## First Steps

### 1. Login and Get Access Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your_secure_password"
  }'
```

Response:
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

Save the `access_token` for subsequent requests.

### 2. Register a Telegram Bot

```bash
# Get your bot token from @BotFather on Telegram
# Then register it:

curl -X POST http://localhost:8080/api/v1/bots \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "my_awesome_bot",
    "token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
    "display_name": "My Awesome Bot",
    "description": "A bot for testing"
  }'
```

Response:
```json
{
  "id": 1,
  "username": "my_awesome_bot",
  "display_name": "My Awesome Bot",
  "description": "A bot for testing",
  "is_active": true
}
```

### 3. Create an API Key

```bash
curl -X POST http://localhost:8080/api/v1/apikeys \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Integration Key",
    "description": "For external service integration",
    "rate_limit": 5000
  }'
```

Response:
```json
{
  "id": 1,
  "key": "tgw_abc123def456...",
  "name": "Integration Key",
  "description": "For external service integration",
  "rate_limit": 5000,
  "is_active": true
}
```

**Important**: Save the `key` value - it's only shown once!

### 4. List Your Chats

```bash
curl -X GET http://localhost:8080/api/v1/chats \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 5. Get Messages from a Chat

```bash
# First, ensure you have can_read permission for the chat
# Then retrieve messages:

curl -X GET "http://localhost:8080/api/v1/chats/1/messages?limit=50" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 6. Set Up a Webhook

```bash
curl -X POST http://localhost:8080/api/v1/webhooks \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/telegram",
    "scope": "chat",
    "chat_id": 1,
    "events": "[\"message\", \"edited_message\"]"
  }'
```

Response includes a `secret` for HMAC signature verification.

## Using API Keys Instead of JWT

For machine-to-machine communication, use API keys:

```bash
curl -X GET http://localhost:8080/api/v1/chats \
  -H "X-API-Key: tgw_abc123def456..."
```

## Chat-Level Permissions

To grant permissions to a user or API key for a specific chat:

```sql
-- Grant read and send permissions to a user
INSERT INTO chat_permissions (chat_id, user_id, can_read, can_send, can_manage)
VALUES (1, 2, true, true, false);

-- Grant permissions to an API key
INSERT INTO chat_permissions (chat_id, api_key_id, can_read, can_send, can_manage)
VALUES (1, 1, true, true, false);
```

Permissions are automatically cached in Redis for 5 minutes.

## Development

### Running with Hot Reload

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with auto-reload
make dev
```

### Running Tests

```bash
# Run unit tests
make test

# Run integration tests (requires Docker)
make test-integration
```

### Generating API Documentation

The API is documented with Swagger-style comments. To generate docs:

```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/gateway/main.go
```

## Production Deployment

### Using Docker

```bash
# Build production image
make docker-build

# Run with docker-compose
docker-compose -f docker-compose.prod.yml up -d
```

### Security Checklist

Before deploying to production:

- [ ] Change all default passwords
- [ ] Set a strong JWT_SECRET (32+ characters)
- [ ] Use HTTPS for all external endpoints
- [ ] Configure firewall rules
- [ ] Set up rate limiting
- [ ] Enable database backups
- [ ] Configure log aggregation
- [ ] Set up monitoring and alerts
- [ ] Review and lock down API key permissions
- [ ] Configure proper CORS settings

### Environment Variables for Production

```bash
# Production .env
DB_PASSWORD=very_secure_password_here
REDIS_PASSWORD=very_secure_redis_password
JWT_SECRET=very_long_random_string_at_least_32_characters_abcdef123456
WEBHOOK_BASE_URL=https://your-production-domain.com
```

## Troubleshooting

### Database Connection Issues

```bash
# Test database connection
mysql -h localhost -u gateway -p telegram_gateway

# Check if migrations ran
mysql -u gateway -p -e "USE telegram_gateway; SHOW TABLES;"
```

### Redis Connection Issues

```bash
# Test Redis connection
redis-cli ping

# Check Redis is listening
sudo netstat -tlnp | grep redis
```

### Application Logs

```bash
# View logs with Docker Compose
docker-compose logs -f gateway

# View logs from binary
./bin/gateway 2>&1 | tee gateway.log
```

## Next Steps

- Set up Telegram webhooks for your bots
- Configure chat-level permissions
- Integrate WebSocket for real-time updates
- Set up gRPC streaming
- Configure webhook delivery workers

## Support

For issues and questions:
- GitHub Issues: https://github.com/kexi/telegram-bot-gateway/issues
- Documentation: See README.md and IMPLEMENTATION_STATUS.md
