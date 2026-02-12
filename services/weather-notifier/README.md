# Weather Notifier Service

An intelligent weather notification service that monitors weather conditions and sends personalized alerts via Telegram. The service uses LLM to generate natural, context-aware weather notifications based on real-time data from the Caiyun Weather API, and integrates with the Telegram Bot Gateway for message delivery.

## Features

- Real-time weather monitoring from Caiyun Weather API
- LLM-powered natural language notifications
- Multi-chat support with independent configurations
- Multi-location monitoring per chat
- Scheduled weather checks with configurable times
- Delta-based weather change detection
- Per-chat LLM configuration overrides
- API key authentication with the gateway
- Automatic token refresh
- Structured logging with configurable levels
- Docker support for easy deployment

## Prerequisites

- Go 1.22 or higher
- Caiyun Weather API token from https://www.caiyunapp.com/api/
- LLM API access (OpenAI or compatible endpoint)
- Running instance of Telegram Bot Gateway
- API key from the gateway (format: tgw_xxx)
- Docker (optional, for containerized deployment)

## Quick Start

### 1. Get an API Key from Gateway

Before configuring the service, create an API key in the gateway:

1. Access the gateway web frontend (e.g., http://localhost:8080)
2. Log in with your user credentials
3. Navigate to API Keys management
4. Create a new API key
5. Copy the generated key (format: tgw_xxx)

### 2. Configure the Service

Copy the example configuration:

```bash
cp config.json.example config.json
```

Edit config.json with your settings:

```json
{
  "telegram": {
    "api_key": "tgw_1234567890abcdef",
    "api_url": "http://localhost:8080"
  },
  "caiyun": {
    "api_token": "${CAIYUN_API_TOKEN}"
  },
  "llm": {
    "api_key": "${LLM_API_KEY}",
    "model": "gpt-4",
    "max_tokens": 500,
    "temperature": 0.7
  },
  "schedule": {
    "timezone": "Asia/Shanghai",
    "morning_time": "08:00:00",
    "evening_time": "23:30:00",
    "poll_interval": "15m"
  },
  "chats": [
    {
      "chat_id": 123456,
      "name": "Personal",
      "locations": [
        {
          "name": "Home",
          "latitude": 39.9042,
          "longitude": 116.4074
        }
      ]
    }
  ]
}
```

### 3. Set Environment Variables

```bash
export CAIYUN_API_TOKEN="your-caiyun-token"
export LLM_API_KEY="your-openai-key"
```

### 4. Build and Run

```bash
# Build
make build

# Run
./bin/weather-notifier --config config.json
```

Or with Docker Compose:

```bash
docker-compose up -d
```

## Configuration

### Telegram Configuration

The service uses API key authentication to communicate with the gateway:

```json
{
  "telegram": {
    "api_key": "tgw_1234567890abcdef",
    "api_url": "http://gateway:8080"
  }
}
```

- `api_key`: API key from gateway (required, format: tgw_xxx)
- `api_url`: Gateway base URL (required)

### Caiyun Weather API

```json
{
  "caiyun": {
    "api_token": "your-token",
    "timeout": 30
  }
}
```

- `api_token`: Caiyun Weather API token (required)
- `timeout`: HTTP request timeout in seconds (default: 30)

### LLM Configuration

Global LLM settings:

```json
{
  "llm": {
    "provider": "openai",
    "api_key": "your-key",
    "model": "gpt-4",
    "base_url": "",
    "timeout": 60,
    "max_tokens": 500,
    "temperature": 0.7
  }
}
```

- `provider`: LLM provider (currently supports: openai)
- `api_key`: API key for the LLM service (required)
- `model`: Model name (required)
- `base_url`: Custom API endpoint (optional, for compatible APIs)
- `timeout`: Request timeout in seconds (default: 60)
- `max_tokens`: Maximum tokens in response (default: 500)
- `temperature`: Sampling temperature (default: 0.7)

### Schedule Configuration

The service supports time-based scheduling:

```json
{
  "schedule": {
    "timezone": "Asia/Shanghai",
    "morning_time": "08:00:00",
    "evening_time": "23:30:00",
    "poll_interval": "15m"
  }
}
```

- `timezone`: Timezone for scheduling (required)
- `morning_time`: Morning report time in HH:MM:SS format (required)
- `evening_time`: Evening report time in HH:MM:SS format (required)
- `poll_interval`: Polling interval as Go duration (e.g., 15m, 1h) (required)

On startup, the service sends an initial weather update immediately.

### Chat Configuration

Configure one or more chats to receive weather notifications:

```json
{
  "chats": [
    {
      "chat_id": 123456,
      "name": "Alice",
      "locations": [
        {
          "name": "Home",
          "latitude": 39.9042,
          "longitude": 116.4074
        },
        {
          "name": "Office",
          "latitude": 39.9289,
          "longitude": 116.3883
        }
      ],
      "llm": {
        "model": "gpt-4o",
        "temperature": 0.5
      }
    }
  ]
}
```

- `chat_id`: Telegram chat ID (required, must be unique)
- `name`: Descriptive name for the chat (required)
- `locations`: Array of locations to monitor (required, at least one)
  - `name`: Location name (required, unique within chat)
  - `latitude`: Latitude coordinate (required, -90 to 90)
  - `longitude`: Longitude coordinate (required, -180 to 180)
- `llm`: Optional LLM overrides for this chat (inherits global settings)

### Detection Thresholds

Configure delta-based thresholds for weather change alerts:

```json
{
  "detection": {
    "temperature_delta": 5.0,
    "wind_speed_delta": 10.0,
    "visibility_delta": 2.0,
    "aqi_cn_delta": 50.0,
    "aqi_usa_delta": 25.0
  }
}
```

Thresholds trigger alerts when values change by the specified amount since the last check.

### Logging Configuration

```json
{
  "logging": {
    "level": "info",
    "pretty_print": true
  }
}
```

- `level`: Log level (debug, info, warn, error) (default: info)
- `pretty_print`: Enable colored console output (default: true)

### Environment Variable Expansion

All configuration values support environment variable substitution using `${VAR_NAME}` syntax:

```json
{
  "telegram": {
    "api_key": "${TELEGRAM_API_KEY}"
  },
  "caiyun": {
    "api_token": "${CAIYUN_API_TOKEN}"
  }
}
```

For literal dollar signs, use `$$`.

## Gateway Integration

### Authentication Flow

The service uses API key authentication with the gateway:

1. API key is included in the Authorization header: `X-API-Key: tgw_xxx`
2. Gateway validates the API key
3. Service sends messages via `/api/v1/messages/send` endpoint

No JWT token management is required - API keys provide stateless authentication.

### Required Gateway Permissions

Ensure your API key has the following permissions:

- Message sending access to configured chat IDs
- If the gateway has chat-level ACL enabled, verify `can_send` permission for target chats

### Deployment Scenarios

#### Local Development

```json
{
  "telegram": {
    "api_url": "http://localhost:8080"
  }
}
```

#### Docker Compose (Same Network)

```yaml
services:
  weather-notifier:
    environment:
      - TELEGRAM_API_URL=http://gateway:8080
    networks:
      - gateway-network

  gateway:
    networks:
      - gateway-network
```

#### Separate Deployments

```json
{
  "telegram": {
    "api_url": "https://gateway.example.com"
  }
}
```

## Docker Deployment

### Using Docker Compose

1. Create a `.env` file with credentials:

```bash
cp .env.example .env
# Edit .env with your tokens
```

2. Start the service:

```bash
docker-compose up -d
```

3. View logs:

```bash
docker-compose logs -f weather-notifier
```

4. Stop the service:

```bash
docker-compose down
```

### Using Docker Directly

1. Build the image:

```bash
docker build -t weather-notifier:latest .
```

2. Run the container:

```bash
docker run -d \
  --name weather-notifier \
  -v $(pwd)/config.json:/app/config.json \
  -v $(pwd)/data:/app/data \
  -e TELEGRAM_API_KEY="tgw_xxx" \
  -e CAIYUN_API_TOKEN="your-token" \
  -e LLM_API_KEY="your-key" \
  weather-notifier:latest
```

### Docker Image Details

The Docker setup uses multi-stage builds:

- Build stage: Go 1.22 to compile the binary
- Runtime stage: Minimal Alpine Linux image
- Approximate image size: 20MB
- Health check: Verifies process is running

### Data Persistence

Weather state is stored in per-location files: `weather_state_{chatID}_{locationName}.json`

Examples:

- `weather_state_123456_Home.json`
- `weather_state_123456_Office.json`

Mount the `./data` directory as a volume to persist state across container restarts.

## Development

### Project Structure

```
services/weather-notifier/
├── cmd/
│   └── weather-notifier/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # JSON configuration management
│   ├── utils/                # Logging and utilities
│   ├── weather/              # Caiyun Weather API client
│   ├── llm/                  # LLM API client
│   ├── telegram/             # Gateway API client
│   ├── detector/             # Weather change detection
│   ├── notification/         # Notification system
│   └── app/                  # Application orchestration
├── pkg/
│   └── models/               # Shared types and models
├── config.json.example       # Example configuration
├── Dockerfile                # Docker build configuration
├── docker-compose.yml        # Docker Compose setup
├── Makefile                  # Build and development tasks
└── README.md                 # This file
```

### Building

```bash
# Build the binary
make build

# Build with version info
make build VERSION=v1.0.0

# Clean build artifacts
make clean
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/weather/...
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run security check
make security
```

## Troubleshooting

### Authentication Fails

Error: `API key authentication failed`

Solutions:

1. Verify API key format is correct (tgw_xxx)
2. Check if API key is active in gateway
3. Ensure API key has message sending permissions
4. Verify gateway URL is reachable:

```bash
curl http://gateway:8080/health
```

### Message Send Failures

Error: `send message failed with status 403`

Solutions:

1. Chat ACL: API key may not have `can_send` permission
2. Bot not in chat: Ensure bot is added to the target chat
3. Chat not registered: Gateway may not know about the chat yet

### Weather Data Not Fetching

Solutions:

1. Verify Caiyun API token is valid
2. Check latitude/longitude are correct (must be within valid ranges)
3. Ensure API timeout is sufficient
4. Check logs for API errors

### LLM Not Generating Notifications

Solutions:

1. Verify LLM API key is valid
2. Check model name is correct
3. Ensure base URL is set if using custom endpoint
4. Increase timeout if requests are timing out

### Schedule Not Working

Solutions:

1. Verify time format is HH:MM:SS (24-hour)
2. Check timezone setting matches your location
3. Ensure poll_interval uses valid Go duration format (e.g., 15m, 1h)
4. Check logs for schedule errors

### Container Health Check Failing

```bash
# Check container status
docker ps | grep weather-notifier

# View health check status
docker inspect --format='{{json .State.Health}}' weather-notifier

# Check logs
docker-compose logs weather-notifier
```

## Gateway Compatibility

The weather-notifier service is fully compatible with the Telegram Bot Gateway. It uses:

- API key authentication for stateless, secure access
- Standard message sending endpoints that remain stable
- No dependency on user authentication or API key management endpoints

The gateway's CLI-based API key management does not affect the weather service operation. API keys are created once via the web frontend or CLI and then used by the service.

## Configuration Migration

If migrating from an older YAML-based configuration:

### Key Changes

1. Configuration format: YAML to JSON
2. Authentication: JWT-based user login to API key authentication
3. Locations: Moved from global to per-chat configuration
4. Schedule: Cron expressions to time-of-day format
5. State files: Single file to per-location files

### Migration Steps

1. Create API key in gateway (see Quick Start section)
2. Copy `config.json.example` to `config.json`
3. Migrate settings from old `config.yaml`:
   - Replace `bot_token` and `password` with `api_key`
   - Move global location to `chats[].locations[]`
   - Convert cron schedules to time format
4. Update environment variables if using `${VAR}` syntax
5. Test the new configuration:

```bash
./weather-notifier --config config.json
```

### Schedule Format Conversion

Old cron format:

```yaml
schedule:
  check_weather_cron: "0 7,12,18 * * *"
```

New time format:

```json
{
  "schedule": {
    "morning_time": "07:00:00",
    "evening_time": "18:00:00",
    "poll_interval": "5h"
  }
}
```

Old state files (`weather_state.json`) are not automatically migrated. Remove them after verifying the new system works.

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome. Please:

1. Fork the repository at https://github.com/KexiChanProjectProxy/telegram-bot-gateway
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## Support

For issues and questions, visit the repository:

https://github.com/KexiChanProjectProxy/telegram-bot-gateway

## Acknowledgments

- Caiyun Weather API for weather data: https://www.caiyunapp.com/api/
- Telegram Bot API for messaging platform: https://core.telegram.org/bots/api
- OpenAI for LLM capabilities: https://openai.com/
