# Configuration Migration: YAML to JSON + API Key Authentication

## Overview

The weather-notifier service configuration has been migrated from YAML (with Viper) to JSON (with native Go support). Additionally, **authentication has been changed from JWT-based user authentication to API key authentication**.

This enables:

1. **Correct authentication method**: Uses API keys (tgw_xxx) instead of incorrectly attempting user login
2. **Multi-chat support**: Configure multiple chat subscriptions in a single instance
3. **Multi-location per chat**: Each chat can monitor multiple locations
4. **Per-chat LLM overrides**: Override global LLM settings for specific chats
5. **Simpler dependency tree**: Removed Viper and all transitive dependencies
6. **No JWT token management**: Simpler, stateless authentication

## Breaking Changes

### Authentication Method Changed! 🚨

**Old (INCORRECT):**
```yaml
telegram:
  bot_token: "${TELEGRAM_BOT_TOKEN}"
  password: "${TELEGRAM_PASSWORD}"
```

This was trying to use the **user authentication** endpoint (for humans logging into the web frontend), which was incorrect.

**New (CORRECT):**
```json
{
  "telegram": {
    "api_key": "tgw_1234567890abcdef",
    "api_url": "http://localhost:8080"
  }
}
```

Now uses **API key authentication** (for services/bots), which is the correct method.

### How to Get an API Key

1. Log into the gateway web frontend
2. Navigate to API Keys section
3. Create a new API key (format: `tgw_xxx`)
4. Use this API key in your weather-notifier config

### Configuration Format

**Old (config.yaml)**:
```yaml
telegram:
  bot_token: "${TELEGRAM_BOT_TOKEN}"
caiyun:
  api_token: "${CAIYUN_API_TOKEN}"
  latitude: 39.9042
  longitude: 116.4074
llm:
  api_key: "${LLM_API_KEY}"
  model: "gpt-4"
```

**New (config.json)**:
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
  "chats": [
    {
      "chat_id": 123456,
      "name": "My Chat",
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

### Key Differences

1. **Authentication method changed**:
   - Old: `bot_token` + `password` (attempting user login - INCORRECT)
   - New: `api_key` (using service authentication - CORRECT)
2. **Locations moved from global to per-chat**: Each chat can have multiple locations
2. **New required fields**:
   - `telegram.api_key` (replaces bot_token + password)
   - `chats` array (at least one chat required)
   - `chat_id` and `locations` for each chat
   - `schedule.morning_cron`, `schedule.evening_cron`, `schedule.poll_cron` (explicit cron expressions)
3. **LLM configuration**:
   - Added `max_tokens` and `temperature` to global LLM config
   - Per-chat overrides supported via optional `llm` object
4. **Detection thresholds**: Changed from `rain_probability_threshold` / `temperature_high_threshold` / etc. to delta-based thresholds
   - `temperature_delta`, `wind_speed_delta`, `visibility_delta`, `aqi_cn_delta`, `aqi_usa_delta`

### State Files

**Old**: Single `weather_state.json` file

**New**: Per-location state files: `weather_state_{chatID}_{locationName}.json`

Example:
- `weather_state_123456_Home.json`
- `weather_state_123456_Office.json`

**Note**: Old `weather_state.json` files are not automatically migrated or deleted. You can safely remove them after verifying the new system works.

## Migration Guide

### 0. Create an API Key (REQUIRED FIRST STEP)

Before updating the config, you **must** create an API key:

1. Access the gateway web frontend (usually http://localhost:8080)
2. Log in with your user credentials
3. Navigate to **API Keys** management
4. Click **Create New API Key**
5. Copy the generated key (format: `tgw_xxx`)
6. Store it securely - you'll use it in step 1 below

### 1. Convert config.yaml to config.json

Use the provided `config.json.example` as a template:

```bash
cp config.json.example config.json
```

Edit `config.json` and migrate your settings:

1. Copy over basic settings
   - **telegram.api_key**: Use the API key you created in step 0 (NOT bot_token/password!)
   - **telegram.api_url**: Your gateway URL (e.g., http://localhost:8080)
   - caiyun, llm settings
2. Create a `chats` array
3. For each chat, add locations that were previously in `caiyun.latitude/longitude`
4. Optionally add per-chat LLM overrides

### 2. Environment Variables

Environment variables are still supported using `${VAR_NAME}` syntax:

```json
{
  "telegram": {
    "api_key": "${TELEGRAM_API_KEY}",
    "api_url": "${TELEGRAM_API_URL}"
  }
}
```

**Set the environment variable:**
```bash
export TELEGRAM_API_KEY="tgw_1234567890abcdef"
export TELEGRAM_API_URL="http://localhost:8080"
```

**Important**: If you need a literal `$` in a value, use `$$` (standard `os.ExpandEnv` behavior).

### 3. Update Cron Schedules

The old `schedule.check_weather_cron` has been split into three separate schedules:

- `schedule.morning_cron`: Morning notification (default: `0 8 * * *`)
- `schedule.evening_cron`: Evening notification (default: `30 23 * * *`)
- `schedule.poll_cron`: Weather change polling (default: `*/15 * * * *`)

### 4. Per-Chat LLM Overrides

You can override LLM settings per chat:

```json
{
  "llm": {
    "model": "gpt-4",
    "temperature": 0.7
  },
  "chats": [
    {
      "chat_id": 123456,
      "name": "Alice",
      "locations": [...],
      "llm": {
        "model": "gpt-4o",
        "temperature": 0.5
      }
    }
  ]
}
```

Only specified fields are overridden; others inherit from global config.

## Example Configurations

### Single Chat, Single Location (minimal migration)

```json
{
  "telegram": {
    "bot_token": "${TELEGRAM_BOT_TOKEN}",
    "password": "${TELEGRAM_PASSWORD}"
  },
  "caiyun": {
    "api_token": "${CAIYUN_API_TOKEN}"
  },
  "llm": {
    "api_key": "${LLM_API_KEY}",
    "model": "gpt-4"
  },
  "chats": [
    {
      "chat_id": 123456,
      "name": "Personal",
      "locations": [
        {
          "name": "Beijing",
          "latitude": 39.9042,
          "longitude": 116.4074
        }
      ]
    }
  ]
}
```

### Multiple Chats with Location Tracking

```json
{
  "chats": [
    {
      "chat_id": 111111,
      "name": "Alice",
      "locations": [
        {"name": "Home", "latitude": 39.9289, "longitude": 116.3883},
        {"name": "Office", "latitude": 39.9042, "longitude": 116.4074}
      ],
      "llm": {
        "model": "gpt-4o",
        "temperature": 0.5
      }
    },
    {
      "chat_id": 222222,
      "name": "Bob",
      "locations": [
        {"name": "Shanghai", "latitude": 31.2304, "longitude": 121.4737}
      ]
    }
  ]
}
```

## Validation

The config loader validates:

- Required fields (bot_token, password, api_token, llm config, chats)
- At least one chat configured
- Each chat has at least one location
- No duplicate chat IDs
- No duplicate location names within a chat
- Latitude/longitude ranges (-90 to 90, -180 to 180)

Run the service to validate your config:

```bash
./weather-notifier --config config.json
```

## Troubleshooting

### "at least one chat must be configured"

Add a `chats` array with at least one chat entry.

### "duplicate chat_id X found"

Each chat must have a unique `chat_id`.

### "duplicate location name 'X' in chat"

Location names must be unique within each chat (different chats can reuse names).

### State file not found / initializing

This is normal on first run. State files will be created automatically.
