# Getting Started

This guide will help you set up and run the Telegram Bot Gateway for the first time.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21 or higher** (for building from source)
- **MySQL 8.0+** or MariaDB 10.5+
- **Redis 6.0+** or higher
- **Docker and Docker Compose** (recommended for quick setup)

## Quick Start with Docker Compose

The fastest way to get started is using Docker Compose. This will run the gateway, MySQL, and Redis in containers.

```bash
# Clone the repository
git clone https://github.com/kexi/telegram-bot-gateway.git
cd telegram-bot-gateway

# Copy environment variables
cp .env.example .env

# Edit .env and set your secrets (see Environment Variables section)
nano .env

# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8080/health
```

The gateway will be available at:
- HTTP API: http://localhost:8080
- gRPC: localhost:9090

## Building from Source

If you prefer to build and run the gateway locally without Docker:

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools (optional)
make install-tools

# Generate Protocol Buffer code (if using gRPC)
make proto
```

### Configure Database

```bash
# Start MySQL
# Then create the database
mysql -u root -p

CREATE DATABASE telegram_gateway CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'gateway'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON telegram_gateway.* TO 'gateway'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### Configure Redis

```bash
# Start Redis
redis-server

# Or install and start as a service
sudo systemctl start redis
```

### Set Environment Variables

```bash
# Copy the example file
cp .env.example .env

# Edit with your values
nano .env
```

Required environment variables:
- `DB_PASSWORD`: Your MySQL password
- `JWT_SECRET`: A secure random string (minimum 32 characters)
- `WEBHOOK_BASE_URL`: Your public domain for Telegram webhooks

See the Environment Variables section below for the complete list.

### Run Database Migrations

```bash
# Run migrations
make migrate

# Or manually
go run cmd/migrate/main.go up
```

### Start the Gateway

```bash
# Using make
make run

# Or with hot reload (requires air)
make dev

# Or build and run binary
make build
./bin/gateway
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | MySQL host | localhost | Yes |
| `DB_PORT` | MySQL port | 3306 | No |
| `DB_NAME` | Database name | telegram_gateway | Yes |
| `DB_USER` | Database user | root | Yes |
| `DB_PASSWORD` | Database password | - | Yes |
| `REDIS_ADDRESS` | Redis address | localhost:6379 | Yes |
| `REDIS_PASSWORD` | Redis password | - | No |
| `JWT_SECRET` | JWT signing secret (min 32 chars) | - | Yes |
| `WEBHOOK_BASE_URL` | Base URL for webhooks | - | Yes |
| `HTTP_PORT` | HTTP server port | 8080 | No |
| `GRPC_PORT` | gRPC server port | 9090 | No |
| `APP_ENV` | Environment (dev/production) | dev | No |

## Initial Setup

### Create Admin User

```bash
# Using the createuser tool
go run cmd/createuser/main.go --username admin --email admin@example.com --password changeme

# Or using Docker
docker-compose exec gateway /app/createuser --username admin --email admin@example.com --password changeme

# Or set environment variables and let gateway create user on startup
export DEFAULT_ADMIN_USER=admin
export DEFAULT_ADMIN_PASSWORD=your_secure_password
```

### Login and Get Access Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "changeme"
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

### Register a Telegram Bot

First, create a bot with @BotFather on Telegram to get your bot token. Then register it with the gateway:

```bash
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

### Create an API Key

For machine-to-machine communication, create an API key:

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

**Important**: Save the `key` value - it is only shown once.

## Making Your First API Calls

### List Your Bots

```bash
curl -X GET http://localhost:8080/api/v1/bots \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### List Your Chats

```bash
curl -X GET http://localhost:8080/api/v1/chats \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Get Messages from a Chat

Ensure you have `can_read` permission for the chat, then retrieve messages:

```bash
curl -X GET "http://localhost:8080/api/v1/chats/1/messages?limit=50" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Set Up a Webhook

To receive real-time notifications when messages arrive:

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

The response includes a `secret` for HMAC signature verification.

### Configure Telegram Webhook

Set the Telegram webhook to point to your gateway:

```bash
curl -X POST "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook" \
  -d "url=https://your-domain.com/telegram/webhook/<YOUR_BOT_TOKEN>"

# Check webhook status
curl "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getWebhookInfo"
```

### Using API Keys

For machine-to-machine communication, use API keys instead of JWT tokens:

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

### Gateway Won't Start

1. Check logs: `docker-compose logs gateway`
2. Verify database connection: `docker-compose exec mysql mysql -u gateway -p`
3. Ensure all required environment variables are set
4. Verify migrations ran successfully

### Messages Not Received

1. Check Telegram webhook status:
   ```bash
   curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo
   ```
2. Verify bot is active in database
3. Check webhook delivery logs
4. Verify network connectivity and firewall rules

### Application Logs

```bash
# View logs with Docker Compose
docker-compose logs -f gateway

# View logs from binary
./bin/gateway 2>&1 | tee gateway.log
```

## Next Steps

- Read the [API Reference](./api-reference.md) for detailed endpoint documentation
- Review the [Authentication Guide](./authentication.md) for security configuration
- Check [Deployment Guide](./deployment.md) for production setup
- Explore [CLI Tools](./cli-tools.md) for bot and API key management
- Configure webhook delivery workers for real-time message processing
- Set up monitoring and alerting for production deployments
- Implement rate limiting and quota management for API consumers

## Support

For issues and questions:
- GitHub Issues: https://github.com/kexi/telegram-bot-gateway/issues
- Documentation: See README.md and other docs in the `docs/` directory
