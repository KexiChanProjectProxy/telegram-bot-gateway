# Deployment Guide

This guide covers deploying the Telegram Bot Gateway in production and development environments.

## Prerequisites

- Docker 20.10+ and Docker Compose 2.0+
- Go 1.21+ (for source builds)
- MySQL 8.0+
- Redis 7.0+
- Protocol Buffers compiler (protoc) for gRPC development

## Docker Compose Deployment

### Development Environment

Development deployment includes hot-reload and debug configurations:

```bash
# Clone repository
git clone <repository-url>
cd telegram-bot-gateway

# Copy environment template
cp .env.example .env

# Configure environment variables
# Edit .env with your settings

# Start all services
docker-compose up -d

# Verify health
curl http://localhost:8080/health

# View logs
docker-compose logs -f gateway
```

Services will be available at:
- HTTP API: http://localhost:8080
- gRPC: localhost:9090
- MySQL: localhost:3306
- Redis: localhost:6379

### Production Environment

Production deployment with optimized configurations:

```bash
cd deployments

# Set production environment variables
export DB_PASSWORD="<secure-database-password>"
export REDIS_PASSWORD="<secure-redis-password>"
export JWT_SECRET="<32-character-minimum-secret>"
export WEBHOOK_BASE_URL="https://your-domain.com"

# Start production stack
docker-compose -f docker-compose.prod.yml up -d

# Verify deployment
docker-compose -f docker-compose.prod.yml ps
curl https://your-domain.com/health
```

Production compose file includes:
- Resource limits and reservations
- Restart policies
- Health checks
- Volume persistence
- Network isolation
- Security options

### Test Environment

Isolated testing environment:

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
make test-integration

# Cleanup
docker-compose -f docker-compose.test.yml down -v
```

## Kubernetes Deployment

### Building Container Images

Build and publish gateway image:

```bash
# Build image
docker build -t your-registry.com/telegram-bot-gateway:v1.0.0 .

# Tag for latest
docker tag your-registry.com/telegram-bot-gateway:v1.0.0 \
  your-registry.com/telegram-bot-gateway:latest

# Push to registry
docker push your-registry.com/telegram-bot-gateway:v1.0.0
docker push your-registry.com/telegram-bot-gateway:latest
```

### Configuring Secrets

Create Kubernetes secrets with base64-encoded values:

```bash
# Encode credentials
echo -n "your_db_password" | base64
echo -n "your_jwt_secret_32_chars_minimum" | base64
echo -n "https://api.your-domain.com" | base64

# Update deployments/kubernetes.yaml Secret section
# Replace placeholder values with encoded credentials
```

### Deploying to Cluster

Deploy gateway and dependencies:

```bash
# Create namespace
kubectl create namespace telegram-bot-gateway

# Apply configurations
kubectl apply -f deployments/kubernetes.yaml

# Verify deployment
kubectl get pods -n telegram-bot-gateway
kubectl get svc -n telegram-bot-gateway

# Check logs
kubectl logs -n telegram-bot-gateway -l app=gateway -f

# Port forward for testing
kubectl port-forward -n telegram-bot-gateway svc/gateway-http 8080:80
curl http://localhost:8080/health
```

### Configuring Ingress

Production ingress with TLS termination:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gateway-ingress
  namespace: telegram-bot-gateway
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  ingressClassName: nginx
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

Apply ingress configuration:

```bash
kubectl apply -f ingress.yaml
kubectl get ingress -n telegram-bot-gateway
```

## Building from Source

### Local Development Build

Build gateway locally:

```bash
# Install dependencies
go mod download

# Install development tools
make install-tools

# Generate Protocol Buffer code
make proto

# Build binary
make build

# Binary output: bin/gateway
./bin/gateway --help
```

### Database Setup

Configure and migrate database:

```bash
# Start MySQL and Redis
docker-compose up -d mysql redis

# Wait for services
sleep 10

# Run migrations
make migrate

# Verify migration
docker-compose exec mysql mysql -u gateway -p telegram_gateway -e "SHOW TABLES;"
```

### Configuration File

Create production configuration at `configs/config.json`:

```json
{
  "database": {
    "host": "${DB_HOST}",
    "port": 3306,
    "name": "telegram_gateway",
    "user": "gateway",
    "password": "${DB_PASSWORD}",
    "max_open_conns": 100,
    "max_idle_conns": 25,
    "conn_max_lifetime": 300
  },
  "redis": {
    "address": "${REDIS_ADDRESS}",
    "password": "${REDIS_PASSWORD}",
    "db": 0,
    "pool_size": 50
  },
  "auth": {
    "jwt": {
      "secret": "${JWT_SECRET}",
      "access_token_ttl": 3600,
      "refresh_token_ttl": 604800
    }
  },
  "webhook": {
    "base_url": "${WEBHOOK_BASE_URL}",
    "timeout": 30
  },
  "http": {
    "port": 8080,
    "read_timeout": 15,
    "write_timeout": 15
  },
  "grpc": {
    "port": 9090
  }
}
```

Environment variables (syntax `${VAR_NAME}`) are expanded at runtime.

### Running the Gateway

Execute gateway service:

```bash
# Run directly
make run

# With hot reload (requires air)
make dev

# Run built binary
./bin/gateway

# With custom config
./bin/gateway --config configs/config.production.json
```

## Initial Configuration

### Creating Admin User

Create administrative account:

```bash
# Using createuser utility
go run cmd/createuser/main.go \
  --username admin \
  --email admin@example.com \
  --password <secure-password>

# In Docker environment
docker-compose exec gateway /app/createuser \
  --username admin \
  --email admin@example.com \
  --password <secure-password>

# In Kubernetes
kubectl exec -n telegram-bot-gateway deployment/gateway -- \
  /app/createuser --username admin --email admin@example.com --password <secure-password>
```

### Registering Telegram Bot

Register bot with gateway:

```bash
# Obtain bot token from @BotFather on Telegram

# Authenticate to get JWT token
TOKEN=$(curl -s -X POST https://api.your-domain.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<password>"}' | jq -r '.access_token')

# Register bot
curl -X POST https://api.your-domain.com/api/v1/bots \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_bot",
    "token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
    "display_name": "Production Bot",
    "description": "Bot description"
  }'
```

### Configuring Webhook

Set Telegram webhook to gateway:

```bash
# Gateway webhook URL format:
# https://api.your-domain.com/telegram/webhook/<BOT_TOKEN>

# Configure webhook via Telegram API
curl -X POST "https://api.telegram.org/bot<BOT_TOKEN>/setWebhook" \
  -d "url=https://api.your-domain.com/telegram/webhook/<BOT_TOKEN>" \
  -d "max_connections=40" \
  -d "drop_pending_updates=false"

# Verify webhook configuration
curl "https://api.telegram.org/bot<BOT_TOKEN>/getWebhookInfo"
```

## Monitoring and Observability

### Health Checks

Gateway exposes health endpoints:

```bash
# Basic health check
curl https://api.your-domain.com/health

# Response: {"status": "healthy", "timestamp": "..."}

# Detailed health with authentication
curl -H "Authorization: Bearer $TOKEN" \
  https://api.your-domain.com/health/detailed
```

### Metrics Endpoint

Access system metrics:

```bash
# Requires authentication
curl -H "Authorization: Bearer $TOKEN" \
  https://api.your-domain.com/metrics

# Metrics include:
# - Request rates and latencies
# - Active connections
# - Database pool stats
# - Redis connection stats
# - Message processing rates
# - Error rates
```

### Log Management

Access and filter logs:

```bash
# Docker Compose logs
docker-compose logs -f gateway
docker-compose logs --tail=100 gateway

# Filter by level
docker-compose logs gateway | grep ERROR

# Kubernetes logs
kubectl logs -n telegram-bot-gateway -l app=gateway -f
kubectl logs -n telegram-bot-gateway deployment/gateway --tail=100

# Follow logs from all replicas
kubectl logs -n telegram-bot-gateway -l app=gateway -f --all-containers
```

Production log configuration should use structured JSON logging with appropriate levels.

### Database Monitoring

Monitor database health and performance:

```bash
# Connect to database
docker-compose exec mysql mysql -u gateway -p telegram_gateway

# Check message volume
SELECT COUNT(*) as total_messages FROM messages;

# Active bots
SELECT id, username, display_name, is_active FROM bots WHERE is_active = 1;

# Recent message activity
SELECT DATE(sent_at) as date, COUNT(*) as count
FROM messages
WHERE sent_at > DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY DATE(sent_at);

# Connection pool status
SHOW STATUS LIKE 'Threads_connected';
SHOW STATUS LIKE 'Max_used_connections';
```

## Scaling Strategies

### Horizontal Scaling

Gateway is stateless and horizontally scalable:

```bash
# Docker Compose scaling
docker-compose up -d --scale gateway=3

# Kubernetes manual scaling
kubectl scale deployment gateway -n telegram-bot-gateway --replicas=5

# Kubernetes autoscaling (HPA)
kubectl autoscale deployment gateway -n telegram-bot-gateway \
  --cpu-percent=70 \
  --min=3 \
  --max=10
```

Load balancing is handled by Docker Compose networking or Kubernetes Services.

### Database Scaling

For high-traffic deployments:

1. **MySQL Replication**: Configure master-slave replication for read scaling
2. **Read Replicas**: Route read-only queries to replicas
3. **Connection Pooling**: Tune max_open_conns and max_idle_conns
4. **Sharding**: Partition data by bot_id for write scaling

Example replication setup:

```sql
-- Master configuration
[mysqld]
server-id = 1
log-bin = mysql-bin
binlog-format = ROW

-- Slave configuration
[mysqld]
server-id = 2
relay-log = relay-bin
read-only = 1
```

### Redis Scaling

For large deployments:

1. **Redis Cluster**: Horizontal partitioning across nodes
2. **Redis Sentinel**: High availability with automatic failover
3. **Separate Instances**: Dedicated instances for pub/sub vs. caching
4. **Memory Optimization**: Configure maxmemory and eviction policies

Example Redis Cluster configuration:

```bash
# Create 6-node cluster (3 masters, 3 replicas)
redis-cli --cluster create \
  192.168.1.1:7000 192.168.1.2:7000 192.168.1.3:7000 \
  192.168.1.4:7000 192.168.1.5:7000 192.168.1.6:7000 \
  --cluster-replicas 1
```

## Backup and Recovery

### Database Backup

Regular backup procedures:

```bash
# Full database dump
docker-compose exec mysql mysqldump \
  -u gateway -p \
  --single-transaction \
  --routines \
  --triggers \
  telegram_gateway > backup-$(date +%Y%m%d).sql

# Compressed backup
docker-compose exec mysql mysqldump \
  -u gateway -p \
  --single-transaction \
  telegram_gateway | gzip > backup-$(date +%Y%m%d).sql.gz

# Kubernetes backup
kubectl exec -n telegram-bot-gateway deployment/mysql -- \
  mysqldump -u gateway -p$DB_PASSWORD telegram_gateway > backup.sql
```

### Database Restore

Restore from backup:

```bash
# Restore from dump
docker-compose exec -T mysql mysql -u gateway -p telegram_gateway < backup-20260212.sql

# Restore from compressed backup
gunzip < backup-20260212.sql.gz | \
  docker-compose exec -T mysql mysql -u gateway -p telegram_gateway

# Kubernetes restore
kubectl exec -i -n telegram-bot-gateway deployment/mysql -- \
  mysql -u gateway -p$DB_PASSWORD telegram_gateway < backup.sql
```

### Redis Backup

Persist Redis data:

```bash
# Trigger save
docker-compose exec redis redis-cli SAVE

# Background save
docker-compose exec redis redis-cli BGSAVE

# Copy RDB file
docker cp telegram-bot-gateway-redis-1:/data/dump.rdb ./redis-backup-$(date +%Y%m%d).rdb

# Restore (stop Redis, replace dump.rdb, restart)
docker-compose stop redis
docker cp redis-backup-20260212.rdb telegram-bot-gateway-redis-1:/data/dump.rdb
docker-compose start redis
```

Configure automatic snapshots in redis.conf:

```
save 900 1      # Save after 900s if at least 1 key changed
save 300 10     # Save after 300s if at least 10 keys changed
save 60 10000   # Save after 60s if at least 10000 keys changed
```

## Security Hardening

### Production Security Checklist

Critical security requirements:

- Change all default passwords and secrets before deployment
- Use JWT secret minimum 32 characters, cryptographically random
- Enable TLS/HTTPS for all external endpoints
- Configure firewall rules to restrict access to database and Redis
- Apply security updates monthly or on critical CVE disclosure
- Implement automated backup strategy with offsite storage
- Monitor and alert on failed authentication attempts
- Use dedicated secret management solution (HashiCorp Vault, AWS Secrets Manager)
- Schedule quarterly vulnerability scans
- Enable audit logging for sensitive operations
- Restrict container privileges and capabilities
- Use non-root users in containers
- Implement network policies in Kubernetes
- Enable Pod Security Standards in Kubernetes

### TLS Configuration

Enable HTTPS in production:

```yaml
# Nginx reverse proxy configuration
server {
    listen 443 ssl http2;
    server_name api.your-domain.com;

    ssl_certificate /etc/nginx/ssl/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://gateway:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Network Security

Restrict network access:

```bash
# Docker network isolation
docker network create --internal backend
docker network connect backend mysql
docker network connect backend redis
docker network connect backend gateway

# Kubernetes NetworkPolicy
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: gateway-network-policy
  namespace: telegram-bot-gateway
spec:
  podSelector:
    matchLabels:
      app: gateway
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: mysql
    ports:
    - protocol: TCP
      port: 3306
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

## Troubleshooting

### Gateway Fails to Start

Diagnostic steps:

1. Check logs for error messages:
   ```bash
   docker-compose logs gateway
   kubectl logs -n telegram-bot-gateway deployment/gateway
   ```

2. Verify database connectivity:
   ```bash
   docker-compose exec mysql mysql -u gateway -p -e "SELECT 1;"
   ```

3. Validate environment variables:
   ```bash
   docker-compose exec gateway env | grep -E "(DB_|REDIS_|JWT_)"
   ```

4. Check migration status:
   ```bash
   docker-compose exec mysql mysql -u gateway -p telegram_gateway -e "SELECT * FROM schema_migrations;"
   ```

5. Verify port availability:
   ```bash
   netstat -ln | grep -E "(8080|9090)"
   ```

### Messages Not Received

Debug webhook delivery:

1. Verify webhook configuration:
   ```bash
   curl "https://api.telegram.org/bot<TOKEN>/getWebhookInfo"
   ```

2. Check bot active status:
   ```bash
   docker-compose exec mysql mysql -u gateway -p telegram_gateway \
     -e "SELECT id, username, is_active FROM bots WHERE token = '<TOKEN>';"
   ```

3. Test webhook endpoint:
   ```bash
   curl -X POST https://api.your-domain.com/telegram/webhook/<TOKEN> \
     -H "Content-Type: application/json" \
     -d '{"update_id": 1, "message": {"message_id": 1, "text": "test"}}'
   ```

4. Review webhook delivery logs in gateway:
   ```bash
   docker-compose logs gateway | grep webhook
   ```

5. Verify network path:
   ```bash
   curl -I https://api.your-domain.com/health
   traceroute api.your-domain.com
   ```

### High Memory Usage

Diagnose memory issues:

1. Check container memory:
   ```bash
   docker stats
   kubectl top pods -n telegram-bot-gateway
   ```

2. Analyze Redis memory:
   ```bash
   docker-compose exec redis redis-cli INFO memory
   docker-compose exec redis redis-cli MEMORY STATS
   ```

3. Review message retention:
   ```bash
   docker-compose exec mysql mysql -u gateway -p telegram_gateway \
     -e "SELECT COUNT(*), MIN(sent_at), MAX(sent_at) FROM messages;"
   ```

4. Implement message archival:
   ```sql
   DELETE FROM messages WHERE sent_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
   ```

5. Configure Redis maxmemory:
   ```bash
   docker-compose exec redis redis-cli CONFIG SET maxmemory 1gb
   docker-compose exec redis redis-cli CONFIG SET maxmemory-policy allkeys-lru
   ```

### Rate Limiting Issues

Handle rate limit errors:

1. Check current rate limit configuration in config.json

2. Review rate limit headers in responses:
   ```bash
   curl -i https://api.your-domain.com/api/v1/bots
   # Check X-RateLimit-* headers
   ```

3. Adjust limits for production traffic

4. Implement per-tenant rate limits if needed

5. Monitor rate limit metrics

## Performance Optimization

### Database Tuning

Optimize query performance:

```sql
-- Create indexes for common queries
CREATE INDEX idx_messages_chat_sent ON messages(chat_id, sent_at DESC);
CREATE INDEX idx_messages_telegram ON messages(telegram_id);
CREATE INDEX idx_messages_bot_id ON messages(bot_id);
CREATE INDEX idx_bots_username ON bots(username);

-- Analyze tables
ANALYZE TABLE messages;
ANALYZE TABLE bots;

-- Optimize tables
OPTIMIZE TABLE messages;
```

InnoDB configuration:

```ini
[mysqld]
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT
```

### Redis Optimization

Configure Redis for performance:

```bash
# Set memory limit
redis-cli CONFIG SET maxmemory 512mb

# Configure eviction policy
redis-cli CONFIG SET maxmemory-policy allkeys-lru

# Disable persistence for cache-only workloads
redis-cli CONFIG SET save ""

# Enable key expiration
redis-cli CONFIG SET maxmemory-samples 5
```

### Go Runtime Tuning

Optimize Go runtime:

```bash
# Set GOMAXPROCS to CPU count
export GOMAXPROCS=4

# Adjust garbage collection target
export GOGC=100

# Enable CPU profiling
./bin/gateway --cpuprofile=cpu.prof

# Enable memory profiling
./bin/gateway --memprofile=mem.prof
```

## Maintenance Procedures

### Regular Maintenance Tasks

Schedule and execute:

- **Daily**: Log rotation, health checks
- **Weekly**: Database backups, metrics review
- **Monthly**: Security updates, dependency updates
- **Quarterly**: Performance review, capacity planning, vulnerability scans
- **Ongoing**: Certificate renewal monitoring, disk space monitoring

### Upgrading Gateway

Safe upgrade procedure:

```bash
# 1. Backup database and configuration
docker-compose exec mysql mysqldump -u gateway -p telegram_gateway > backup-pre-upgrade.sql
cp .env .env.backup

# 2. Pull latest version
git fetch origin
git checkout v1.1.0  # or desired version

# 3. Review CHANGELOG.md for breaking changes

# 4. Rebuild containers
docker-compose build gateway

# 5. Run database migrations
make migrate

# 6. Restart gateway with downtime
docker-compose stop gateway
docker-compose up -d gateway

# 7. Verify deployment
curl http://localhost:8080/health
docker-compose logs gateway

# 8. Monitor for errors
docker-compose logs -f gateway
```

Zero-downtime upgrade in Kubernetes:

```bash
# Update image tag
kubectl set image deployment/gateway -n telegram-bot-gateway \
  gateway=your-registry.com/telegram-bot-gateway:v1.1.0

# Monitor rollout
kubectl rollout status deployment/gateway -n telegram-bot-gateway

# Rollback if needed
kubectl rollout undo deployment/gateway -n telegram-bot-gateway
```

## Support and Resources

For deployment issues:

- Review logs for detailed error messages
- Consult documentation in docs/ directory
- Check GitHub issues for known problems
- Review Telegram Bot API documentation
- Contact support team with logs and configuration details

Production deployment assistance:

- Provide full logs from affected components
- Include environment configuration (sanitized)
- Describe expected vs. actual behavior
- Note any recent changes to configuration or infrastructure
