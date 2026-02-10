.PHONY: help build run test clean dev migrate lint proto

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building..."
	@go build -o bin/gateway cmd/gateway/main.go
	@go build -o bin/bot cmd/bot/main.go
	@go build -o bin/apikey cmd/apikey/main.go
	@go build -o bin/createuser cmd/createuser/main.go
	@echo "✓ All binaries built successfully"

run: ## Run the application
	@echo "Running..."
	@go run cmd/gateway/main.go

dev: ## Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
	@air

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@docker-compose -f deployments/docker-compose.test.yml up --abort-on-container-exit
	@docker-compose -f deployments/docker-compose.test.yml down

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ build/ dist/
	@go clean

migrate: ## Run database migrations
	@echo "Running migrations..."
	@go run cmd/migrate/main.go up

migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	@go run cmd/migrate/main.go down

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

proto: ## Generate code from proto files
	@echo "Generating protobuf code..."
	@./scripts/generate-proto.sh

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

install-tools: ## Install development tools
	@echo "Installing tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t telegram-bot-gateway:latest .

docker-up: ## Start services with docker-compose
	@echo "Starting services..."
	@docker-compose up -d

docker-down: ## Stop services
	@echo "Stopping services..."
	@docker-compose down

docker-logs: ## Show logs
	@docker-compose logs -f

.DEFAULT_GOAL := help
