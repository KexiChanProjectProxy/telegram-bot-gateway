#!/bin/bash

# Generate Protocol Buffer code
# This script generates Go code from .proto files

set -e

echo "Generating Protocol Buffer code..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed"
    echo "Install with: brew install protobuf (macOS) or apt-get install protobuf-compiler (Linux)"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Generate code
export PATH="$HOME/go/bin:$PATH"
protoc \
    --proto_path=api/proto \
    --go_out=api/proto \
    --go_opt=paths=source_relative \
    --go-grpc_out=api/proto \
    --go-grpc_opt=paths=source_relative \
    api/proto/*.proto

echo "✓ Protocol Buffer code generated successfully"
echo "  Generated files:"
echo "    - api/proto/gateway.pb.go"
echo "    - api/proto/gateway_grpc.pb.go"
