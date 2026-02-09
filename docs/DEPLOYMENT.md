# Deployment Guide

This guide covers deploying the Telegram Bot Gateway in different environments.

## Prerequisites

- Docker and Docker Compose (for containerized deployment)
- Go 1.21+ (for local development)
- MySQL 8.0+
- Redis 7+
- Protocol Buffers compiler (protoc) for gRPC

## Quick Start (Docker Compose)

The fastest way to get started is using Docker Compose:

```bash
# 1. Clone the repository
git clone <repository-url>
cd telegram-bot-gateway

# 2. Copy and configure environment variables
cp .env.example .env
# Edit .env with your configuration

# 3. Start all services
docker-compose up -d

# 4. Check service health
curl http://localhost:8080/health

# 5. View logs
docker-compose logs -f gateway
```

The gateway will be available at:
- HTTP API: http://localhost:8080
- gRPC: localhost:9090

## Local Development

### 1. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
make install-tools

# Generate Protocol Buffer code
make proto
```

### 2. Configure Database

```bash
# Start MySQL and Redis with Docker
docker-compose up -d mysql redis

# Run migrations
make migrate
```

### 3. Create Configuration

Create `configs/config.json` from `configs/config.example.json` and update values:

```json
{
  "database": {
    "host": "localhost",
    "password": "your_password"
  },
  "redis": {
    "address": "localhost:6379"
  },
  "auth": {
    "jwt": {
      "secret": "your-secret-key-min-32-characters"
    }
  }
}
```

### 4. Run the Gateway

```bash
# Run directly
make run

# Or with hot reload (requires air)
make dev

# Or build and run binary
make build
./bin/gateway
```

## Production Deployment

### Docker Compose (Production)

For production deployment with Docker Compose:

```bash
# Use the production compose file
cd deployments
docker-compose -f docker-compose.prod.yml up -d

# Set environment variables
export DB_PASSWORD="your_secure_password"
export REDIS_PASSWORD="your_redis_password"
export JWT_SECRET="your-jwt-secret-32-chars-minimum"
export WEBHOOK_BASE_URL="https://your-domain.com"

# Start services
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes Deployment

#### 1. Build and Push Docker Image

```bash
# Build the image
docker build -t your-registry/telegram-bot-gateway:v1.0.0 .

# Push to registry
docker push your-registry/telegram-bot-gateway:v1.0.0
```

#### 2. Update Kubernetes Secrets

Edit `deployments/kubernetes.yaml` and update the Secret values with base64-encoded credentials:

```bash
# Encode your values
echo -n "your_password" | base64
echo -n "your_jwt_secret" | base64
echo -n "https://your-domain.com" | base64
```

#### 3. Deploy to Kubernetes

```bash
# Apply the configuration
kubectl apply -f deployments/kubernetes.yaml

# Check deployment status
kubectl get pods -n telegram-bot-gateway
kubectl get svc -n telegram-bot-gateway

# View logs
kubectl logs -n telegram-bot-gateway -l app=gateway -f

# Check gateway health
kubectl port-forward -n telegram-bot-gateway svc/gateway-http 8080:80
curl http://localhost:8080/health
```

#### 4. Configure Ingress (Optional)

Create an Ingress resource for external access:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gateway-ingress
  namespace: telegram-bot-gateway
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - api.your-domain.com
    secretName: gateway-tls
  rules:
  - host: api.your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: gateway-http
            port:
              number: 80
```

## Configuration

### Environment Variables

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

### Configuration File

The gateway uses JSON configuration files in the `configs/` directory. Environment variables take precedence and can be embedded using `${VAR_NAME}` syntax.

## Initial Setup

### 1. Create Admin User

```bash
# Using the createuser tool
go run cmd/createuser/main.go --username admin --email admin@example.com --password changeme

# Or using Docker
docker-compose exec gateway /app/createuser --username admin --email admin@example.com --password changeme
```

### 2. Register a Telegram Bot

```bash
# 1. Get your bot token from @BotFather on Telegram
# 2. Login to get JWT token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"changeme"}'

# 3. Register the bot
curl -X POST http://localhost:8080/api/v1/bots \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_bot",
    "token": "123456:ABC-DEF...",
    "display_name": "My Bot",
    "description": "My Telegram Bot"
  }'
```

### 3. Configure Telegram Webhook

```bash
# The gateway will provide a webhook URL like:
# https://your-domain.com/telegram/webhook/YOUR_BOT_TOKEN

# Set it with Telegram API:
curl -X POST "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook" \
  -d "url=https://your-domain.com/telegram/webhook/<YOUR_BOT_TOKEN>"
```

## Monitoring

### Health Checks

```bash
# Check gateway health
curl http://localhost:8080/health

# Check system metrics
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/metrics
```

### Logs

```bash
# Docker Compose logs
docker-compose logs -f gateway

# Kubernetes logs
kubectl logs -n telegram-bot-gateway -l app=gateway -f --tail=100
```

### Database Monitoring

```bash
# Connect to MySQL
docker-compose exec mysql mysql -u gateway -p telegram_gateway

# Check message counts
SELECT COUNT(*) FROM messages;

# Check active bots
SELECT * FROM bots WHERE is_active = 1;
```

## Scaling

### Horizontal Scaling

The gateway is stateless and can be scaled horizontally:

```bash
# Docker Compose
docker-compose up -d --scale gateway=3

# Kubernetes (automatic with HPA)
kubectl scale deployment gateway -n telegram-bot-gateway --replicas=5
```

### Database Scaling

For high traffic:
1. Use MySQL replication (master-slave)
2. Implement read replicas
3. Consider sharding by bot_id

### Redis Scaling

For large deployments:
1. Use Redis Cluster for horizontal scaling
2. Implement Redis Sentinel for high availability
3. Separate pub/sub and caching Redis instances

## Backup and Recovery

### Database Backup

```bash
# Backup
docker-compose exec mysql mysqldump -u gateway -p telegram_gateway > backup.sql

# Restore
docker-compose exec -T mysql mysql -u gateway -p telegram_gateway < backup.sql
```

### Redis Backup

```bash
# Backup (creates dump.rdb)
docker-compose exec redis redis-cli SAVE

# Copy backup
docker cp tgw-redis:/data/dump.rdb ./redis-backup.rdb
```

## Troubleshooting

### Gateway Won't Start

1. Check logs: `docker-compose logs gateway`
2. Verify database connection: `docker-compose exec mysql mysql -u gateway -p`
3. Check configuration: Ensure all required environment variables are set
4. Verify migrations: Check if migrations ran successfully

### Messages Not Received

1. Check Telegram webhook status:
   ```bash
   curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo
   ```
2. Verify bot is active in database
3. Check webhook delivery logs
4. Verify network connectivity

### High Memory Usage

1. Check Redis memory: `docker stats tgw-redis`
2. Review message retention policies
3. Consider implementing message archival
4. Tune Redis maxmemory settings

### Rate Limiting Issues

1. Check rate limit settings in config.json
2. Review rate limit headers in API responses
3. Consider increasing limits for production
4. Implement per-tenant rate limits

## Security Checklist

- [ ] Change default passwords and secrets
- [ ] Use strong JWT secret (32+ characters)
- [ ] Enable TLS/HTTPS in production
- [ ] Configure firewall rules
- [ ] Regular security updates
- [ ] Implement backup strategy
- [ ] Monitor failed authentication attempts
- [ ] Use secret management (e.g., HashiCorp Vault)
- [ ] Regular vulnerability scans
- [ ] Audit logs enabled

## Performance Tuning

### Database Optimization

```sql
-- Add indexes for common queries
CREATE INDEX idx_messages_chat_sent ON messages(chat_id, sent_at DESC);
CREATE INDEX idx_messages_telegram ON messages(telegram_id);

-- Optimize table
OPTIMIZE TABLE messages;
```

### Redis Optimization

```bash
# Set appropriate maxmemory
redis-cli CONFIG SET maxmemory 512mb

# Use appropriate eviction policy
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

### Go Runtime

```bash
# Set GOMAXPROCS for optimal CPU usage
export GOMAXPROCS=4

# Adjust garbage collection
export GOGC=100
```

## Maintenance

### Regular Tasks

- Weekly database backups
- Monthly security updates
- Quarterly performance reviews
- Log rotation and cleanup
- Certificate renewals (if using TLS)

### Upgrading

```bash
# 1. Backup database and config
./scripts/backup.sh

# 2. Pull new version
git pull origin main

# 3. Rebuild
docker-compose build gateway

# 4. Run migrations
make migrate

# 5. Restart services
docker-compose up -d gateway
```

## Support

For issues and questions:
- Check documentation in `docs/`
- Review logs for error messages
- Check GitHub issues
- Contact support team
