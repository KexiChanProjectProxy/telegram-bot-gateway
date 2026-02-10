# Multi-stage build for smaller image
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o gateway cmd/gateway/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gateway .
COPY --from=builder /build/configs ./configs
COPY --from=builder /build/migrations ./migrations

# Expose ports
EXPOSE 8080 9090

# Run the gateway
CMD ["./gateway"]
