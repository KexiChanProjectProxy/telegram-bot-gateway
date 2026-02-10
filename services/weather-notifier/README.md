# Weather Notice Bot

An intelligent weather notification service that monitors weather conditions and sends personalized alerts via Telegram. The bot uses LLM (Large Language Model) to generate natural, context-aware weather notifications based on real-time data from the Caiyun Weather API.

## Features

- **Real-time Weather Monitoring**: Fetches current weather conditions and forecasts from Caiyun Weather API
- **Intelligent Notifications**: Uses LLM to generate natural, personalized weather alerts
- **Telegram Integration**: Delivers notifications via Telegram with password-protected access
- **Scheduled Checks**: Configurable cron-based weather monitoring (default: 7 AM, 12 PM, 6 PM)
- **Weather Change Detection**: Monitors for significant weather changes and sends alerts
- **Threshold-based Alerts**: Customizable thresholds for rain probability, temperature, and air quality
- **Secure Authentication**: Password-protected Telegram bot with user ID allowlisting
- **Flexible Configuration**: YAML-based config with environment variable overrides
- **Robust Logging**: Structured logging with configurable levels and pretty-print mode
- **Docker Support**: Easy deployment with Docker and Docker Compose

## Prerequisites

- **Go 1.22** or higher
- **Caiyun Weather API Token**: Sign up at [Caiyun Weather](https://www.caiyunapp.com/api/) to get an API token
- **LLM API Access**: OpenAI API key or compatible API endpoint
- **Telegram Bot Token**: Create a bot via [@BotFather](https://t.me/botfather) on Telegram
- **Docker** (optional): For containerized deployment

## Installation

### Option 1: Build from Source

1. Clone the repository:
```bash
git clone https://github.com/user/weather-notice-bot.git
cd weather-notice-bot
```

2. Install dependencies:
```bash
go mod download
```

3. Build the binary:
```bash
make build
```

The binary will be created at `bin/weather-notice-bot`.

### Option 2: Docker

See the [Deployment with Docker](#deployment-with-docker) section below.

## Configuration

The bot can be configured using a YAML configuration file and/or environment variables. Environment variables take precedence over the config file.

### Configuration File

Copy the example configuration:
```bash
cp config.yaml config.yaml.local
```

Edit `config.yaml.local` with your settings:

```yaml
# Telegram Bot Configuration
telegram:
  bot_token: "YOUR_TELEGRAM_BOT_TOKEN"
  password: "YOUR_SECURE_PASSWORD"
  admin_user_id: 123456789  # Your Telegram user ID
  allowed_ids: []           # Optional: List of allowed user IDs

# Caiyun Weather API Configuration
caiyun:
  api_token: "YOUR_CAIYUN_API_TOKEN"
  latitude: 39.9042   # Beijing example
  longitude: 116.4074
  timeout: 30         # Request timeout in seconds

# LLM Configuration
llm:
  provider: "openai"  # Currently supports: openai
  api_key: "YOUR_OPENAI_API_KEY"
  model: "gpt-4"
  base_url: ""        # Optional: Custom API endpoint
  timeout: 60         # Request timeout in seconds

# Schedule Configuration (cron format)
schedule:
  check_weather_cron: "0 7,12,18 * * *"  # 7 AM, 12 PM, 6 PM daily
  timezone: "Asia/Shanghai"

# Weather Detection Thresholds
detection:
  rain_probability_threshold: 0.3    # 30% or higher triggers alert
  temperature_high_threshold: 35.0   # Celsius
  temperature_low_threshold: 0.0     # Celsius
  aqi_threshold: 150                 # Air Quality Index

# Logging Configuration
logging:
  level: "info"         # debug, info, warn, error
  pretty_print: true    # Enable colored console output
```

### Environment Variables

All configuration values can be overridden with environment variables prefixed with `WNB_`:

```bash
# Telegram
export WNB_TELEGRAM_BOT_TOKEN="your-bot-token"
export WNB_TELEGRAM_PASSWORD="your-password"
export WNB_TELEGRAM_ADMIN_USER_ID="123456789"

# Caiyun Weather API
export WNB_CAIYUN_API_TOKEN="your-caiyun-token"
export WNB_CAIYUN_LATITUDE="39.9042"
export WNB_CAIYUN_LONGITUDE="116.4074"

# LLM
export WNB_LLM_PROVIDER="openai"
export WNB_LLM_API_KEY="your-openai-key"
export WNB_LLM_MODEL="gpt-4"

# Schedule
export WNB_SCHEDULE_CHECK_WEATHER_CRON="0 7,12,18 * * *"
export WNB_SCHEDULE_TIMEZONE="Asia/Shanghai"

# Detection
export WNB_DETECTION_RAIN_PROBABILITY_THRESHOLD="0.3"
export WNB_DETECTION_TEMPERATURE_HIGH_THRESHOLD="35.0"
export WNB_DETECTION_TEMPERATURE_LOW_THRESHOLD="0.0"
export WNB_DETECTION_AQI_THRESHOLD="150"

# Logging
export WNB_LOGGING_LEVEL="info"
export WNB_LOGGING_PRETTY_PRINT="true"
```

## Usage

### Running the Bot

#### From Binary
```bash
./bin/weather-notice-bot
```

Or specify a custom config file:
```bash
./bin/weather-notice-bot -config /path/to/config.yaml
```

#### Using Make
```bash
make run
```

### Telegram Commands

Once the bot is running, interact with it via Telegram:

1. **Start the bot**: Send `/start` to your bot
2. **Authenticate**: Send `/auth <password>` using the password from your config
3. **Get weather**: Send `/weather` to receive current weather conditions
4. **Get forecast**: Send `/forecast` to receive upcoming forecast
5. **Get status**: Send `/status` to check bot status
6. **Help**: Send `/help` for command list

### Scheduled Notifications

The bot automatically checks weather conditions according to the cron schedule and sends notifications when:
- Rain probability exceeds the threshold
- Temperature goes above/below configured thresholds
- Air Quality Index exceeds the threshold
- Significant weather changes are detected

## API Integration Details

### Caiyun Weather API

The bot uses Caiyun Weather API v2.6 to fetch:
- Real-time weather conditions (temperature, humidity, wind, AQI)
- Hourly forecasts for the next 48 hours
- Daily forecasts for the next 15 days

**API Endpoint**: `https://api.caiyunapp.com/v2.6/{token}/{longitude},{latitude}/weather`

### LLM Integration

The bot uses LLM to analyze weather data and generate natural language notifications. Supported providers:
- **OpenAI**: GPT-4, GPT-3.5-turbo
- **Compatible APIs**: Any OpenAI-compatible endpoint

The LLM receives structured weather data and thresholds, then generates:
- Context-aware weather summaries
- Personalized recommendations (e.g., "Bring an umbrella", "Dress warmly")
- Natural language alerts for threshold violations

### Telegram Bot API

The bot uses Telegram Bot API to:
- Receive user commands
- Send weather notifications
- Authenticate users via password
- Manage user sessions

## Development

### Project Structure

```
weather-notice-bot/
├── cmd/
│   └── weather-notice-bot/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── utils/                # Logging and utilities
│   ├── weather/              # Caiyun Weather API client
│   ├── llm/                  # LLM API client
│   ├── telegram/             # Telegram bot integration
│   ├── detector/             # Weather change detection
│   ├── notification/         # Notification system
│   └── app/                  # Application orchestration
├── pkg/
│   └── models/               # Shared types and models
├── config.yaml               # Example configuration
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

## Deployment with Docker

### Using Docker Compose (Recommended)

1. Create a `.env` file with your credentials:
```bash
cp .env.example .env
# Edit .env with your API tokens and passwords
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
docker build -t weather-notice-bot:latest .
```

2. Run the container:
```bash
docker run -d \
  --name weather-notifier \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/data:/app/data \
  -e WNB_TELEGRAM_BOT_TOKEN="your-token" \
  -e WNB_TELEGRAM_PASSWORD="your-password" \
  -e WNB_CAIYUN_API_TOKEN="your-token" \
  -e WNB_LLM_API_KEY="your-key" \
  weather-notice-bot:latest
```

### Environment Configuration for Docker

The Docker setup uses a multi-stage build for minimal image size:
- **Build stage**: Uses Go 1.22 to compile the binary
- **Runtime stage**: Uses minimal Alpine Linux image
- **Image size**: Approximately 20MB

Configuration can be provided via:
1. Volume-mounted `config.yaml` file
2. Environment variables in `docker-compose.yml` or `.env` file
3. Command-line environment variables with `docker run`

## Troubleshooting

### Bot doesn't respond to commands
- Verify the bot token is correct
- Check that you've started the bot with `/start`
- Ensure you're authenticated with `/auth <password>`
- Check logs for authentication errors

### Weather data not fetching
- Verify Caiyun API token is valid
- Check latitude/longitude are correct
- Ensure API timeout is sufficient
- Check logs for API errors

### LLM not generating notifications
- Verify LLM API key is valid
- Check the model name is correct
- Ensure base URL is set (if using custom endpoint)
- Increase timeout if requests are timing out

### Cron schedule not working
- Verify cron expression syntax
- Check timezone setting matches your location
- Ensure the bot is running continuously
- Check logs for schedule errors

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## Support

For issues and questions:
- GitHub Issues: https://github.com/user/weather-notice-bot/issues
- Documentation: https://github.com/user/weather-notice-bot/wiki

## Acknowledgments

- [Caiyun Weather API](https://www.caiyunapp.com/api/) for weather data
- [Telegram Bot API](https://core.telegram.org/bots/api) for messaging platform
- [OpenAI](https://openai.com/) for LLM capabilities
