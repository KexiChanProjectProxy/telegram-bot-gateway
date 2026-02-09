# gRPC API Guide

## Overview

The Telegram Bot Gateway provides a gRPC API for high-performance streaming of messages. This is ideal for applications that need real-time message delivery with lower latency than WebSocket.

## Protocol Buffer Definition

The API is defined in `api/proto/gateway.proto` using Protocol Buffers v3.

## Services

### MessageService

Handles message streaming and retrieval.

#### Methods

1. **StreamMessages** - Stream messages from multiple chats
2. **StreamChatMessages** - Stream messages from a single chat
3. **SendMessage** - Send a message to a chat
4. **GetMessages** - Retrieve historical messages

### ChatService

Handles chat management.

#### Methods

1. **ListChats** - List accessible chats
2. **GetChat** - Get chat details

### BotService

Handles bot management.

#### Methods

1. **ListBots** - List registered bots
2. **GetBot** - Get bot details
3. **CreateBot** - Register a new bot
4. **DeleteBot** - Delete a bot

## Authentication

All gRPC requests require authentication via metadata:

```
authorization: Bearer YOUR_JWT_TOKEN
```

The server validates the JWT and extracts user information.

## Generating Client Code

### Go Client

```bash
# Install protoc compiler and Go plugins
brew install protobuf  # macOS
# or
apt-get install protobuf-compiler  # Linux

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
./scripts/generate-proto.sh
```

### Python Client

```bash
pip install grpcio grpcio-tools

python -m grpc_tools.protoc \
  -I. \
  --python_out=. \
  --grpc_python_out=. \
  api/proto/gateway.proto
```

### JavaScript/TypeScript Client

```bash
npm install @grpc/grpc-js @grpc/proto-loader

# Or use grpc-tools for code generation
npm install -g grpc-tools
grpc_tools_node_protoc \
  --js_out=import_style=commonjs,binary:. \
  --grpc_out=grpc_js:. \
  api/proto/gateway.proto
```

## Usage Examples

### Go Client

```go
package main

import (
    "context"
    "io"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"

    pb "github.com/kexi/telegram-bot-gateway/api/proto"
)

func main() {
    // Connect to server
    conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewMessageServiceClient(conn)

    // Add authentication
    token := "your_jwt_token"
    ctx := metadata.AppendToOutgoingContext(
        context.Background(),
        "authorization", "Bearer "+token,
    )

    // Stream messages
    stream, err := client.StreamMessages(ctx, &pb.StreamMessagesRequest{
        ChatIds: []uint64{1, 2, 3},
    })
    if err != nil {
        log.Fatal(err)
    }

    // Receive messages
    for {
        event, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }

        log.Printf("New message in chat %d: %s", event.ChatId, event.Text)
    }
}
```

### Python Client

```python
import grpc
from api.proto import gateway_pb2
from api.proto import gateway_pb2_grpc

def run():
    # Connect to server
    channel = grpc.insecure_channel('localhost:9090')
    stub = gateway_pb2_grpc.MessageServiceStub(channel)

    # Add authentication
    token = 'your_jwt_token'
    metadata = [('authorization', f'Bearer {token}')]

    # Stream messages
    request = gateway_pb2.StreamMessagesRequest(chat_ids=[1, 2, 3])

    try:
        for event in stub.StreamMessages(request, metadata=metadata):
            print(f"New message in chat {event.chat_id}: {event.text}")
    except grpc.RpcError as e:
        print(f"Error: {e.code()}: {e.details()}")

if __name__ == '__main__':
    run()
```

### JavaScript/Node.js Client

```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

// Load proto file
const packageDefinition = protoLoader.loadSync(
  'api/proto/gateway.proto',
  {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
  }
);

const gateway = grpc.loadPackageDefinition(packageDefinition).gateway;

// Connect to server
const client = new gateway.MessageService(
  'localhost:9090',
  grpc.credentials.createInsecure()
);

// Add authentication
const token = 'your_jwt_token';
const metadata = new grpc.Metadata();
metadata.add('authorization', `Bearer ${token}`);

// Stream messages
const call = client.StreamMessages(
  { chat_ids: [1, 2, 3] },
  metadata
);

call.on('data', (event) => {
  console.log(`New message in chat ${event.chat_id}: ${event.text}`);
});

call.on('error', (error) => {
  console.error('Error:', error);
});

call.on('end', () => {
  console.log('Stream ended');
});
```

## Message Format

### MessageEvent

Streamed message events have the following structure:

```protobuf
message MessageEvent {
  string type = 1;                    // "new_message", "edited_message"
  uint64 chat_id = 2;
  uint64 message_id = 3;
  int64 telegram_id = 4;
  uint64 bot_id = 5;
  string direction = 6;               // "incoming", "outgoing"
  string text = 7;
  string from_username = 8;
  string from_first_name = 9;
  string from_last_name = 10;
  string message_type = 11;           // "text", "photo", "video", etc.
  google.protobuf.Timestamp timestamp = 12;
  map<string, string> metadata = 13;
}
```

## Server Configuration

The gRPC server runs on port 9090 by default (configurable in `configs/config.json`):

```json
{
  "server": {
    "grpc": {
      "address": ":9090"
    }
  }
}
```

## Error Handling

gRPC uses standard status codes:

- `UNAUTHENTICATED (16)` - No or invalid authentication
- `PERMISSION_DENIED (7)` - Insufficient permissions
- `NOT_FOUND (5)` - Resource not found
- `INTERNAL (13)` - Internal server error
- `INVALID_ARGUMENT (3)` - Invalid request parameters

Example error handling:

```go
stream, err := client.StreamMessages(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        log.Printf("Error code: %s, message: %s", st.Code(), st.Message())
    }
}
```

## Performance Considerations

- **Connection Pooling**: Reuse gRPC connections across requests
- **Keep-Alive**: Configure keep-alive to maintain long-lived connections
- **Compression**: Enable gzip compression for large payloads
- **Deadlines**: Set appropriate deadlines/timeouts for requests

### Example with Keep-Alive (Go)

```go
conn, err := grpc.Dial(
    "localhost:9090",
    grpc.WithInsecure(),
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             3 * time.Second,
        PermitWithoutStream: true,
    }),
)
```

## Comparison: gRPC vs WebSocket vs REST

| Feature | gRPC | WebSocket | REST |
|---------|------|-----------|------|
| **Protocol** | HTTP/2 | HTTP/1.1 Upgrade | HTTP/1.1 |
| **Latency** | Very Low | Low | Medium |
| **Throughput** | Very High | High | Medium |
| **Binary** | Yes (Protobuf) | Optional | No (JSON) |
| **Streaming** | Bidirectional | Bidirectional | Unidirectional |
| **Browser Support** | Limited | Excellent | Excellent |
| **Code Generation** | Yes | No | Optional (OpenAPI) |
| **Type Safety** | Strong | None | Optional |

**Use gRPC when:**
- You need lowest latency
- Strong typing is important
- Server-to-server communication
- High throughput is critical

**Use WebSocket when:**
- Browser clients needed
- Simple integration required
- JSON is preferred

**Use REST when:**
- Request/response pattern
- Caching is important
- Simple debugging needed

## Testing

Test the gRPC server:

```bash
# Install grpcurl
brew install grpcurl  # macOS
# or
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:9090 list

# Call a method
grpcurl -plaintext \
  -H "authorization: Bearer YOUR_TOKEN" \
  -d '{"chat_ids": [1]}' \
  localhost:9090 \
  gateway.MessageService/StreamMessages
```

## Production Deployment

For production, use TLS:

```go
// Server
creds, err := credentials.NewServerTLSFromFile("cert.pem", "key.pem")
grpcServer := grpc.NewServer(grpc.Creds(creds))

// Client
creds, err := credentials.NewClientTLSFromFile("ca.pem", "")
conn, err := grpc.Dial("gateway.example.com:9090", grpc.WithTransportCredentials(creds))
```

## Troubleshooting

### Connection Refused

Ensure the gRPC server is running:
```bash
curl http://localhost:8080/health
```

### Authentication Failed

Check that your JWT token is valid:
```bash
# Decode JWT (using jwt.io or jwt-cli)
echo "YOUR_TOKEN" | jwt decode -
```

### Stream Timeout

Increase client timeout:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
```

## Additional Resources

- [gRPC Official Documentation](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)
