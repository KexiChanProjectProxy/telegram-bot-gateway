# Configuration

The Telegram Bot Gateway uses a JSON configuration file with environment variable expansion for sensitive values. This guide covers all configuration options and best practices.

## Configuration File Location

The gateway reads configuration from:
- `services/gateway/configs/config.json` (main configuration file)
- Environment variables for sensitive values

## Configuration Format

The configuration file uses JSON format with `${VAR_NAME}` syntax for environment variable expansion:

```json
{
  "database": {
    "password": "${DB_PASSWORD}"
  }
}
```

At runtime, `${DB_PASSWORD}` is automatically replaced with the value from the environment variable.

## Environment Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_PASSWORD` | MySQL database password | `secure_password_123` |
| `JWT_SECRET` | JWT signing secret (minimum 32 characters) | `your-secret-key-min-32-characters-long` |
| `WEBHOOK_BASE_URL` | Base URL for Telegram webhooks | `https://api.yourdomain.com` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | MySQL host | `localhost` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_NAME` | Database name | `telegram_gateway` |
| `DB_USER` | Database user | `gateway` |
| `REDIS_ADDRESS` | Redis server address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `HTTP_PORT` | HTTP server port | `8080` |
| `GRPC_PORT` | gRPC server port | `9090` |
| `APP_ENV` | Environment mode | `dev` |

## Configuration Sections

### Server Configuration

Controls HTTP and gRPC server behavior.

```json
{
  "server": {
    "mode": "release",
    "http": {
      "address": ":8080",
      "read_timeout": "30s",
      "write_timeout": "30s",
      "idle_timeout": "120s"
    },
    "grpc": {
      "address": ":9090"
    }
  }
}
```

**Options:**
- `mode`: Server mode (`debug`, `release`, or `test`)
- `http.address`: HTTP server listen address
- `http.read_timeout`: Maximum duration for reading request
- `http.write_timeout`: Maximum duration for writing response
- `http.idle_timeout`: Maximum idle time for keep-alive connections
- `grpc.address`: gRPC server listen address

### Database Configuration

Controls MySQL connection and pooling settings.

```json
{
  "database": {
    "driver": "mysql",
    "host": "localhost",
    "port": 3306,
    "name": "telegram_gateway",
    "user": "gateway",
    "password": "${DB_PASSWORD}",
    "max_open_conns": 25,
    "max_idle_conns": 5,
    "conn_max_lifetime": "5m"
  }
}
```

**Options:**
- `driver`: Database driver (only `mysql` supported)
- `host`: Database server hostname
- `port`: Database server port
- `name`: Database name
- `user`: Database username
- `password`: Database password (use environment variable)
- `max_open_conns`: Maximum open connections (recommended: 25)
- `max_idle_conns`: Maximum idle connections (recommended: 5)
- `conn_max_lifetime`: Maximum connection lifetime

### Redis Configuration

Controls Redis connection for caching and pub/sub.

```json
{
  "redis": {
    "address": "localhost:6379",
    "password": "${REDIS_PASSWORD}",
    "db": 0
  }
}
```

**Options:**
- `address`: Redis server address (host:port)
- `password`: Redis password (empty if no auth)
- `db`: Redis database number (0-15)

### Authentication Configuration

Controls JWT and API key settings.

```json
{
  "auth": {
    "jwt": {
      "secret": "${JWT_SECRET}",
      "access_token_ttl": "15m",
      "refresh_token_ttl": "168h",
      "issuer": "telegram-bot-gateway",
      "refresh_threshold": "5m"
    },
    "api_key": {
      "prefix": "tgw_",
      "length": 32
    }
  }
}
```

**JWT Options:**
- `secret`: Signing secret (minimum 32 characters, use environment variable)
- `access_token_ttl`: Access token lifetime (e.g., `15m`, `1h`)
- `refresh_token_ttl`: Refresh token lifetime (e.g., `168h` = 7 days)
- `issuer`: JWT issuer identifier
- `refresh_threshold`: Time before expiry to allow refresh

**API Key Options:**
- `prefix`: API key prefix for identification
- `length`: Generated key length in characters

### Telegram Configuration

Controls Telegram Bot API integration.

```json
{
  "telegram": {
    "webhook_base_url": "${WEBHOOK_BASE_URL}",
    "timeout": "30s"
  }
}
```

**Options:**
- `webhook_base_url`: Base URL for webhook registration (must be HTTPS in production)
- `timeout`: Timeout for Telegram API requests

### Webhook Delivery Configuration

Controls webhook worker behavior for message delivery.

```json
{
  "webhook_delivery": {
    "worker_count": 10,
    "max_retries": 5,
    "timeout": "30s",
    "queue_name": "webhook_deliveries"
  }
}
```

**Options:**
- `worker_count`: Number of concurrent webhook workers
- `max_retries`: Maximum retry attempts for failed deliveries
- `timeout`: HTTP timeout for webhook requests
- `queue_name`: Redis queue name for deliveries

### Rate Limit Configuration

Controls rate limiting for API requests.

```json
{
  "rate_limit": {
    "requests_per_second": 100,
    "burst": 200,
    "cleanup_interval": "1m"
  }
}
```

**Options:**
- `requests_per_second`: Maximum sustained request rate
- `burst`: Maximum burst size for token bucket
- `cleanup_interval`: Interval for cleaning up expired limiters

## Example Configurations

### Development Configuration

For local development with debug logging:

```json
{
  "server": {
    "mode": "debug",
    "http": {
      "address": ":8080"
    },
    "grpc": {
      "address": ":9090"
    }
  },
  "database": {
    "driver": "mysql",
    "host": "localhost",
    "port": 3306,
    "name": "telegram_gateway",
    "user": "root",
    "password": "password"
  },
  "redis": {
    "address": "localhost:6379",
    "password": "",
    "db": 0
  },
  "auth": {
    "jwt": {
      "secret": "dev-secret-key-min-32-characters-change-in-production",
      "access_token_ttl": "15m",
      "refresh_token_ttl": "168h"
    },
    "api_key": {
      "prefix": "tgw_"
    }
  },
  "telegram": {
    "webhook_base_url": "https://dev.example.com"
  },
  "webhook_delivery": {
    "worker_count": 5,
    "max_retries": 3
  },
  "rate_limit": {
    "requests_per_second": 50,
    "burst": 100
  }
}
```

### Production Configuration

For production deployment with environment variables:

```json
{
  "server": {
    "mode": "release",
    "http": {
      "address": ":8080",
      "read_timeout": "30s",
      "write_timeout": "30s",
      "idle_timeout": "120s"
    },
    "grpc": {
      "address": ":9090"
    }
  },
  "database": {
    "driver": "mysql",
    "host": "${DB_HOST}",
    "port": 3306,
    "name": "${DB_NAME}",
    "user": "${DB_USER}",
    "password": "${DB_PASSWORD}",
    "max_open_conns": 25,
    "max_idle_conns": 5,
    "conn_max_lifetime": "5m"
  },
  "redis": {
    "address": "${REDIS_ADDRESS}",
    "password": "${REDIS_PASSWORD}",
    "db": 0
  },
  "auth": {
    "jwt": {
      "secret": "${JWT_SECRET}",
      "access_token_ttl": "15m",
      "refresh_token_ttl": "168h",
      "issuer": "telegram-bot-gateway",
      "refresh_threshold": "5m"
    },
    "api_key": {
      "prefix": "tgw_",
      "length": 32
    }
  },
  "telegram": {
    "webhook_base_url": "${WEBHOOK_BASE_URL}",
    "timeout": "30s"
  },
  "webhook_delivery": {
    "worker_count": 10,
    "max_retries": 5,
    "timeout": "30s",
    "queue_name": "webhook_deliveries"
  },
  "rate_limit": {
    "requests_per_second": 100,
    "burst": 200,
    "cleanup_interval": "1m"
  }
}
```

## Security Best Practices

### Secrets Management

1. Never commit secrets to version control
2. Use environment variables for all sensitive values
3. Generate strong JWT secrets (minimum 32 characters)
4. Use different secrets for development and production
5. Rotate secrets regularly

### JWT Configuration

1. Set appropriate token lifetimes:
   - Access tokens: 15-60 minutes
   - Refresh tokens: 7-30 days
2. Use minimum 32-character secret (64+ recommended)
3. Set issuer to your application domain
4. Configure refresh threshold to 1/3 of access token TTL

### Database Security

1. Use strong database passwords
2. Limit database user permissions to necessary operations only
3. Enable connection pooling to prevent exhaustion
4. Use SSL/TLS for database connections in production
5. Set appropriate connection limits based on load

### Redis Security

1. Enable Redis authentication in production
2. Use dedicated Redis instance (do not share with other apps)
3. Consider Redis Sentinel for high availability
4. Use separate databases for different purposes

## Performance Tuning

### Database Optimization

Adjust connection pool based on load:

```json
{
  "database": {
    "max_open_conns": 50,
    "max_idle_conns": 10,
    "conn_max_lifetime": "5m"
  }
}
```

Guidelines:
- `max_open_conns`: Set to 2-3x expected concurrent requests
- `max_idle_conns`: Set to 20-40% of max_open_conns
- `conn_max_lifetime`: Keep at 5m unless experiencing issues

### Rate Limiting

Adjust based on expected traffic:

```json
{
  "rate_limit": {
    "requests_per_second": 500,
    "burst": 1000
  }
}
```

Guidelines:
- Development: 50-100 req/s
- Production: 100-500 req/s
- High traffic: 500-1000 req/s
- Set burst to 2x requests_per_second

### Webhook Delivery

Scale workers based on webhook volume:

```json
{
  "webhook_delivery": {
    "worker_count": 20,
    "max_retries": 5,
    "timeout": "10s"
  }
}
```

Guidelines:
- Low volume (< 100 msg/min): 5-10 workers
- Medium volume (100-1000 msg/min): 10-20 workers
- High volume (> 1000 msg/min): 20-50 workers
- Reduce timeout if endpoints are fast and reliable

## Configuration Validation

The gateway validates configuration on startup and will exit with errors if:

1. Required environment variables are missing
2. JWT secret is less than 32 characters
3. Database or Redis connection fails
4. Configuration file is malformed JSON
5. Invalid duration formats (must use Go duration syntax: `15m`, `24h`)

## Troubleshooting

### Configuration Not Loading

Check that the configuration file exists:

```bash
ls -la services/gateway/configs/config.json
```

Verify environment variables are set:

```bash
env | grep -E 'DB_|REDIS_|JWT_|WEBHOOK_'
```

### Environment Variables Not Expanding

Ensure variables are exported before starting the gateway:

```bash
export DB_PASSWORD="your_password"
export JWT_SECRET="your-secret-key-min-32-characters"
./bin/gateway
```

### Connection Failures

Test database connectivity:

```bash
mysql -h localhost -u gateway -p telegram_gateway
```

Test Redis connectivity:

```bash
redis-cli -h localhost -p 6379 ping
```

Check firewall rules and network connectivity between services.

## Related Documentation

- [Getting Started Guide](getting-started.md) - Initial setup instructions
- [Deployment Guide](deployment.md) - Production deployment
- [Authentication Guide](authentication.md) - Auth configuration details
- [API Reference](api-reference.md) - API endpoints and usage
