# Telegram Bot Monorepo

A monorepo containing multiple services for Telegram bot infrastructure.

## Services

### Gateway (`services/gateway/`)
Main Telegram bot gateway service handling webhooks, message routing, gRPC/HTTP APIs, authentication, and multi-bot support.

**Tech Stack**: Go, MySQL, Redis, gRPC, HTTP

### Weather Notifier (`services/weather-notifier/`)
Weather notification service that fetches weather data, sends scheduled notifications via the gateway, and uses LLM for personalized advice.

**Tech Stack**: Go, Caiyun Weather API, OpenAI-compatible LLM

## Directory Structure

```
├── services/
│   ├── gateway/          # Gateway service
│   └── weather-notifier/ # Weather service
├── shared/
│   └── proto/           # Shared protobuf definitions
├── deployments/         # Docker configs
└── docs/               # Documentation
```

## Quick Start

```bash
# Build gateway
cd services/gateway && make build

# Build weather notifier
cd services/weather-notifier && make build

# Run with Docker
cd deployments && docker-compose up
```

## Documentation

- [Getting Started](docs/GETTING_STARTED.md)
- [API Documentation](docs/API.md)
- [gRPC Documentation](docs/GRPC.md)
- [Gateway README](docs/README.md)
- [Weather Notifier README](services/weather-notifier/README.md)
