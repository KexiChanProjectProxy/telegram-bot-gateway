# Multi-stage Dockerfile for Weather Notice Bot
# Build stage: Compile the Go binary
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 for static binary, GOOS=linux for Linux target
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o weather-notice-bot \
    ./cmd/weather-notice-bot

# Runtime stage: Minimal image with just the binary
FROM alpine:latest

# Install ca-certificates for HTTPS requests and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/weather-notice-bot .

# Copy config template (can be overridden with volume mount)
COPY config.yaml ./config.yaml.template

# Create data directory for state files
RUN mkdir -p /app/data && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose no ports (bot connects outbound only)

# Health check (optional - checks if process is running)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep -f weather-notice-bot || exit 1

# Set entrypoint
ENTRYPOINT ["/app/weather-notice-bot"]

# Default command (can be overridden)
CMD ["-config", "/app/config.yaml"]
