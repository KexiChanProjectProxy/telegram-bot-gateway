# Quick Deployment Guide

## Quick Start with Docker Compose

1. **Clone and configure**:
```bash
git clone https://github.com/user/weather-notice-bot.git
cd weather-notice-bot
cp .env.example .env
```

2. **Edit `.env` file with your credentials**:
```bash
nano .env  # or use your favorite editor
```

3. **Start the service**:
```bash
docker-compose up -d
```

4. **View logs**:
```bash
docker-compose logs -f weather-notifier
```

## Environment Variables Quick Reference

### Required Variables
- `WNB_TELEGRAM_BOT_TOKEN` - Your Telegram bot token from @BotFather
- `WNB_TELEGRAM_PASSWORD` - Password for bot authentication
- `WNB_TELEGRAM_ADMIN_USER_ID` - Your Telegram user ID
- `WNB_CAIYUN_API_TOKEN` - Caiyun Weather API token
- `WNB_LLM_API_KEY` - OpenAI or compatible LLM API key

### Optional Variables (have defaults)
- `WNB_CAIYUN_LATITUDE` - Default: 39.9042 (Beijing)
- `WNB_CAIYUN_LONGITUDE` - Default: 116.4074 (Beijing)
- `WNB_SCHEDULE_CHECK_WEATHER_CRON` - Default: "0 7,12,18 * * *"
- `WNB_SCHEDULE_TIMEZONE` - Default: "Asia/Shanghai"
- `WNB_DETECTION_RAIN_PROBABILITY_THRESHOLD` - Default: 0.3
- `WNB_LOGGING_LEVEL` - Default: "info"

## Troubleshooting

### Check if container is running
```bash
docker ps | grep weather-notice-bot
```

### View container logs
```bash
docker-compose logs weather-notifier
```

### Restart the service
```bash
docker-compose restart
```

### Stop the service
```bash
docker-compose down
```

### Rebuild after code changes
```bash
docker-compose up -d --build
```

## Health Check

The container includes a health check that verifies the process is running:
```bash
docker inspect --format='{{json .State.Health}}' weather-notice-bot
```

## Data Persistence

Weather data and state are stored in the `./data` directory, which is mounted as a volume. This ensures data persists across container restarts.

## Security Notes

1. **Never commit `.env` file** - It contains sensitive credentials
2. **Use strong passwords** - For Telegram authentication
3. **Restrict user access** - Use `allowed_ids` in config to whitelist users
4. **Keep tokens secure** - Store API tokens in environment variables, not in code

## Updating

To update to the latest version:
```bash
git pull
docker-compose up -d --build
```

## Monitoring

Check service status:
```bash
# Container status
docker-compose ps

# Resource usage
docker stats weather-notice-bot

# Recent logs
docker-compose logs --tail=100 weather-notifier
```

For full documentation, see [README.md](README.md)
