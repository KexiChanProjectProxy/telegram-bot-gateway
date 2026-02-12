# Telegram Bot Gateway

A production-ready monorepo infrastructure for managing multiple Telegram bots with high-performance message distribution, granular access control, and flexible integration options. The gateway provides a unified interface between Telegram bots and downstream applications, enabling real-time message delivery via WebSocket, gRPC, or HTTP webhooks.

## Key Capabilities

- Multi-bot management with encrypted token storage and automatic webhook registration
- Real-time message distribution through WebSocket streams, gRPC bidirectional streaming, or HTTP webhooks with circuit breakers
- Chat-level access control with granular permissions (can_read, can_send, can_manage) and Redis-cached authorization
- Multiple authentication methods including JWT tokens and API keys compatible with Telegram Bot API query parameter style
- Production-grade reliability with rate limiting, connection pooling, graceful shutdown, and comprehensive health checks

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                   Telegram Bot Gateway                          │
├─────────────────────────────────────────────────────────────────┤
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌───────────┐ │
│  │ HTTP/REST  │  │  WebSocket │  │   gRPC     │  │ Telegram  │ │
│  │  (Gin)     │  │    Hub     │  │  Server    │  │  Webhook  │ │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘  └─────┬─────┘ │
│        │               │               │               │        │
│        └───────────────┴───────────────┴───────────────┘        │
│                            │                                    │
│                ┌───────────┴───────────┐                        │
│                │   Message Broker      │                        │
│                │   (Redis PubSub)      │                        │
│                └───────────┬───────────┘                        │
│                            │                                    │
│        ┌───────────────────┼───────────────────────┐            │
│        ▼                   ▼                       ▼            │
│  ┌───────────┐      ┌───────────┐          ┌───────────┐       │
│  │ WebSocket │      │   gRPC    │          │  Webhook  │       │
│  │  Clients  │      │  Streams  │          │  Workers  │       │
│  └───────────┘      └───────────┘          └───────────┘       │
└─────────────────────────────────────────────────────────────────┘
                            │
                ┌───────────┴───────────┐
                │    MySQL + Redis      │
                └───────────────────────┘

Technology Stack: Go 1.21+ | Gin | MySQL 8.0+ | Redis 7+ | gRPC | Docker
```

## Repository Structure

```
telegram-bot-gateway/
├── services/
│   ├── gateway/              # Main gateway service (Go)
│   │   ├── cmd/              # Entry points (main, CLI tools)
│   │   ├── internal/         # Core business logic
│   │   │   ├── api/          # HTTP/WebSocket handlers
│   │   │   ├── grpc/         # gRPC service implementation
│   │   │   ├── service/      # Business logic layer
│   │   │   ├── repository/   # Data access layer
│   │   │   └── middleware/   # Auth, rate limiting, etc.
│   │   ├── migrations/       # Database migrations
│   │   └── configs/          # Configuration files
│   └── weather-notifier/     # Weather notification service (Go)
│       ├── cmd/              # Service entry point
│       └── internal/         # Weather service logic
├── shared/
│   └── proto/               # Shared protobuf definitions
├── deployments/
│   ├── docker-compose.yml   # Development environment
│   ├── gateway/             # Gateway service configs
│   └── weather-notifier/    # Weather service configs
└── docs/                    # Comprehensive documentation
```

## Quick Start

```bash
# Clone and configure
git clone https://github.com/yourusername/telegram-bot-gateway.git
cd telegram-bot-gateway
cp .env.example .env

# Start all services with Docker Compose
docker-compose up -d

# Verify deployment
curl http://localhost:8080/health
```

The gateway will be available at http://localhost:8080 (HTTP API), localhost:9090 (gRPC), and http://localhost:8080/metrics (metrics endpoint).

## Documentation Index

| Document | Description |
|----------|-------------|
| [Getting Started Guide](docs/getting-started.md) | Complete setup, configuration, and first-run instructions |
| [API Reference](docs/api-reference.md) | REST API endpoints with request/response examples |
| [Authentication Guide](docs/authentication.md) | JWT, API key, and Telegram Bot API-style authentication methods |
| [CLI Tools](docs/cli-tools.md) | Command-line tools for bot and API key management |
| [Configuration](docs/configuration.md) | Complete configuration reference |
| [gRPC Documentation](docs/grpc.md) | gRPC API specification with streaming examples |
| [Deployment Guide](docs/deployment.md) | Production deployment with Docker Compose and Kubernetes |
| [Weather Notifier](services/weather-notifier/README.md) | Weather notification service documentation |

## License

MIT License - see LICENSE file for details.
