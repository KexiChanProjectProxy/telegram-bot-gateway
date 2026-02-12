# gRPC API

The Telegram Bot Gateway provides a high-performance gRPC API for real-time message streaming and bot management. This guide covers service definitions, client implementation, authentication, and best practices.

## When to Use gRPC

### gRPC vs REST vs WebSocket

| Feature | gRPC | WebSocket | REST |
|---------|------|-----------|------|
| Protocol | HTTP/2 | HTTP/1.1 Upgrade | HTTP/1.1 |
| Latency | Very Low | Low | Medium |
| Throughput | Very High | High | Medium |
| Binary Protocol | Yes (Protobuf) | Optional | No (JSON) |
| Streaming | Bidirectional | Bidirectional | Unidirectional |
| Browser Support | Limited | Excellent | Excellent |
| Code Generation | Yes | No | Optional (OpenAPI) |
| Type Safety | Strong | None | Optional |

**Use gRPC when:**
- You need the lowest latency for message delivery
- Strong typing and code generation are important
- Implementing server-to-server communication
- High throughput is critical
- You're building backend services or integrations

**Use WebSocket when:**
- Browser clients are required
- Simple integration is preferred
- JSON is the standard format in your stack

**Use REST when:**
- You need simple request/response patterns
- HTTP caching is important
- Debugging and testing simplicity is a priority

## Services

The gateway provides three gRPC services defined in `shared/proto/api/api/proto/gateway.proto`.

### MessageService

Handles real-time message streaming, historical message retrieval, and message sending.

**Methods:**

- `StreamMessages(StreamMessagesRequest) returns (stream MessageEvent)` - Subscribe to messages from multiple chats simultaneously
- `StreamChatMessages(StreamChatMessagesRequest) returns (stream MessageEvent)` - Subscribe to messages from a single chat
- `SendMessage(SendMessageRequest) returns (SendMessageResponse)` - Send a message to a chat
- `GetMessages(GetMessagesRequest) returns (GetMessagesResponse)` - Retrieve historical messages with pagination

### ChatService

Manages chat information and access.

**Methods:**

- `ListChats(ListChatsRequest) returns (ListChatsResponse)` - List all accessible chats with pagination
- `GetChat(GetChatRequest) returns (Chat)` - Get details for a specific chat

### BotService

Manages bot registration and configuration.

**Methods:**

- `ListBots(ListBotsRequest) returns (ListBotsResponse)` - List registered bots
- `GetBot(GetBotRequest) returns (Bot)` - Get bot details
- `CreateBot(CreateBotRequest) returns (Bot)` - Register a new bot
- `DeleteBot(DeleteBotRequest) returns (DeleteBotResponse)` - Remove a bot

## Authentication

All gRPC requests require JWT authentication via metadata headers.

**Authorization Header:**
```
authorization: Bearer YOUR_JWT_TOKEN
```

The server validates the JWT token using an authentication interceptor and extracts user information. Both unary and streaming RPCs require valid authentication.

**How it works:**
1. Client includes JWT token in gRPC metadata
2. Server interceptor validates token before processing request
3. User ID and username are added to request context
4. Service methods use context to enforce permissions

## Proto File Location

Protocol Buffer definitions are located at:
```
shared/proto/api/api/proto/gateway.proto
```

This file defines all services, messages, and data structures for the gRPC API.

## Client Code Generation

### Go Client

Install prerequisites:
```bash
# Install protoc compiler
brew install protobuf  # macOS
# or
apt-get install protobuf-compiler  # Linux

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Generate Go code:
```bash
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  shared/proto/api/api/proto/gateway.proto
```

### Python Client

Install prerequisites:
```bash
pip install grpcio grpcio-tools
```

Generate Python code:
```bash
python -m grpc_tools.protoc \
  -I. \
  --python_out=. \
  --grpc_python_out=. \
  shared/proto/api/api/proto/gateway.proto
```

### JavaScript/TypeScript Client

Install prerequisites:
```bash
npm install @grpc/grpc-js @grpc/proto-loader
```

For dynamic loading (no code generation):
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const packageDefinition = protoLoader.loadSync(
  'shared/proto/api/api/proto/gateway.proto',
  {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
  }
);
```

For static code generation:
```bash
npm install -g grpc-tools
grpc_tools_node_protoc \
  --js_out=import_style=commonjs,binary:. \
  --grpc_out=grpc_js:. \
  shared/proto/api/api/proto/gateway.proto
```

## Streaming Examples

### Go: Stream Messages from Multiple Chats

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

    // Stream messages from multiple chats
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

        log.Printf("Event: %s, Chat: %d, Text: %s",
            event.Type, event.ChatId, event.Text)
    }
}
```

### Python: Stream Messages with Error Handling

```python
import grpc
from api.proto import gateway_pb2
from api.proto import gateway_pb2_grpc

def stream_messages():
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
            print(f"Event: {event.type}, Chat: {event.chat_id}, Text: {event.text}")
    except grpc.RpcError as e:
        print(f"Error: {e.code()}: {e.details()}")

if __name__ == '__main__':
    stream_messages()
```

### JavaScript: Stream Messages with Node.js

```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

// Load proto file
const packageDefinition = protoLoader.loadSync(
  'shared/proto/api/api/proto/gateway.proto',
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
  console.log(`Event: ${event.type}, Chat: ${event.chat_id}, Text: ${event.text}`);
});

call.on('error', (error) => {
  console.error('Error:', error);
});

call.on('end', () => {
  console.log('Stream ended');
});
```

### Go: Send Message

```go
func sendMessage(client pb.MessageServiceClient, token string) {
    ctx := metadata.AppendToOutgoingContext(
        context.Background(),
        "authorization", "Bearer "+token,
    )

    resp, err := client.SendMessage(ctx, &pb.SendMessageRequest{
        ChatId: 123,
        Text:   "Hello from gRPC!",
        ReplyToMessageId: 0,
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Message sent: ID=%d, Success=%v", resp.MessageId, resp.Success)
}
```

### Python: Retrieve Historical Messages

```python
def get_messages(stub, token, chat_id):
    metadata = [('authorization', f'Bearer {token}')]

    request = gateway_pb2.GetMessagesRequest(
        chat_id=chat_id,
        cursor=0,
        limit=50
    )

    response = stub.GetMessages(request, metadata=metadata)

    for msg in response.messages:
        print(f"Message {msg.id}: {msg.text}")

    if response.has_more:
        print(f"More messages available. Next cursor: {response.next_cursor}")
```

## Message Format

### MessageEvent

Real-time message events streamed from `StreamMessages` and `StreamChatMessages`:

```protobuf
message MessageEvent {
  string type = 1;                    // "new_message", "edited_message", "deleted_message"
  uint64 chat_id = 2;                 // Internal chat ID
  uint64 message_id = 3;              // Internal message ID
  int64 telegram_id = 4;              // Telegram message ID
  uint64 bot_id = 5;                  // Bot that received the message
  string direction = 6;               // "incoming" or "outgoing"
  string text = 7;                    // Message text
  string from_username = 8;           // Sender's username
  string from_first_name = 9;         // Sender's first name
  string from_last_name = 10;         // Sender's last name
  string message_type = 11;           // "text", "photo", "video", etc.
  int64 timestamp = 12;               // Unix timestamp in seconds
  map<string, string> metadata = 13;  // Additional metadata
}
```

### Message

Historical messages retrieved from `GetMessages`:

```protobuf
message Message {
  uint64 id = 1;                      // Internal message ID
  uint64 chat_id = 2;                 // Internal chat ID
  int64 telegram_id = 3;              // Telegram message ID
  int64 from_user_id = 4;             // Sender's Telegram user ID
  string from_username = 5;           // Sender's username
  string from_first_name = 6;         // Sender's first name
  string from_last_name = 7;          // Sender's last name
  string direction = 8;               // "incoming" or "outgoing"
  string message_type = 9;            // "text", "photo", "video", etc.
  string text = 10;                   // Message text
  int64 reply_to_message_id = 11;     // ID of message being replied to
  int64 sent_at = 12;                 // Unix timestamp in seconds
  int64 created_at = 13;              // Unix timestamp in seconds
}
```

## Error Handling

gRPC uses standard status codes for error reporting:

- `UNAUTHENTICATED (16)` - Missing or invalid authentication token
- `PERMISSION_DENIED (7)` - Insufficient permissions for the requested resource
- `NOT_FOUND (5)` - Requested resource does not exist
- `INTERNAL (13)` - Internal server error
- `INVALID_ARGUMENT (3)` - Invalid request parameters
- `UNAVAILABLE (14)` - Service temporarily unavailable

### Go Error Handling

```go
import (
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/codes"
)

stream, err := client.StreamMessages(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        switch st.Code() {
        case codes.Unauthenticated:
            log.Println("Authentication failed. Check your token.")
        case codes.PermissionDenied:
            log.Println("Access denied to requested resource.")
        case codes.NotFound:
            log.Println("Resource not found.")
        default:
            log.Printf("Error: %s - %s", st.Code(), st.Message())
        }
    }
    return
}
```

### Python Error Handling

```python
import grpc

try:
    response = stub.GetMessages(request, metadata=metadata)
except grpc.RpcError as e:
    if e.code() == grpc.StatusCode.UNAUTHENTICATED:
        print("Authentication failed. Check your token.")
    elif e.code() == grpc.StatusCode.PERMISSION_DENIED:
        print("Access denied to requested resource.")
    elif e.code() == grpc.StatusCode.NOT_FOUND:
        print("Resource not found.")
    else:
        print(f"Error: {e.code()} - {e.details()}")
```

## Performance Tips

### Connection Pooling

Reuse gRPC connections across multiple requests to avoid connection overhead:

```go
// Create once and reuse
conn, err := grpc.Dial("gateway.example.com:9090", opts...)
defer conn.Close()

client := pb.NewMessageServiceClient(conn)
// Use client for multiple requests
```

### Keep-Alive Configuration

Configure keep-alive to maintain long-lived connections:

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/keepalive"
    "time"
)

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

### Compression

Enable gzip compression for large payloads:

```go
import "google.golang.org/grpc"

conn, err := grpc.Dial(
    "localhost:9090",
    grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")),
)
```

### Deadlines and Timeouts

Set appropriate deadlines to prevent hanging requests:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

stream, err := client.StreamMessages(ctx, req)
```

### Connection Multiplexing

HTTP/2 multiplexing allows multiple concurrent streams over a single connection. Take advantage of this:

```go
// Single connection, multiple concurrent streams
go streamMessages(client, ctx1)
go streamMessages(client, ctx2)
go streamMessages(client, ctx3)
```

## Testing with grpcurl

grpcurl is a command-line tool for interacting with gRPC servers.

### Installation

```bash
# macOS
brew install grpcurl

# Go
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### List Services

```bash
grpcurl -plaintext localhost:9090 list
```

Output:
```
gateway.BotService
gateway.ChatService
gateway.MessageService
```

### List Methods

```bash
grpcurl -plaintext localhost:9090 list gateway.MessageService
```

Output:
```
gateway.MessageService.GetMessages
gateway.MessageService.SendMessage
gateway.MessageService.StreamChatMessages
gateway.MessageService.StreamMessages
```

### Call a Method

Stream messages with authentication:
```bash
grpcurl -plaintext \
  -H "authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"chat_ids": [1, 2, 3]}' \
  localhost:9090 \
  gateway.MessageService/StreamMessages
```

Send a message:
```bash
grpcurl -plaintext \
  -H "authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"chat_id": 1, "text": "Hello from grpcurl!"}' \
  localhost:9090 \
  gateway.MessageService/SendMessage
```

List chats:
```bash
grpcurl -plaintext \
  -H "authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"limit": 10, "offset": 0}' \
  localhost:9090 \
  gateway.ChatService/ListChats
```

## Server Configuration

The gRPC server runs on port 9090 by default. Configuration is in `services/gateway/configs/config.json`:

```json
{
  "server": {
    "grpc": {
      "address": ":9090"
    }
  }
}
```

To change the port:
```json
{
  "server": {
    "grpc": {
      "address": ":50051"
    }
  }
}
```

The server can also share a port with HTTP if `use_shared_port` is enabled.

## Production Deployment

### TLS Configuration

For production, always use TLS to encrypt gRPC traffic.

**Server (Go):**
```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

creds, err := credentials.NewServerTLSFromFile("cert.pem", "key.pem")
if err != nil {
    log.Fatal(err)
}

grpcServer := grpc.NewServer(grpc.Creds(creds))
```

**Client (Go):**
```go
creds, err := credentials.NewClientTLSFromFile("ca.pem", "")
if err != nil {
    log.Fatal(err)
}

conn, err := grpc.Dial(
    "gateway.example.com:9090",
    grpc.WithTransportCredentials(creds),
)
```

**Client (Python):**
```python
import grpc

with open('ca.pem', 'rb') as f:
    creds = grpc.ssl_channel_credentials(f.read())

channel = grpc.secure_channel('gateway.example.com:9090', creds)
```

### Load Balancing

Use DNS-based load balancing or a service mesh:

```go
// DNS load balancing
conn, err := grpc.Dial(
    "dns:///gateway.example.com:9090",
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
```

### Health Checks

Implement health checks for monitoring:

```bash
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
```

## Troubleshooting

### Connection Refused

Verify the gRPC server is running:
```bash
# Check if port is listening
netstat -an | grep 9090

# Check gateway health
curl http://localhost:8080/health
```

### Authentication Failed

Verify your JWT token is valid and not expired:
```bash
# Decode JWT to check expiration
echo "YOUR_TOKEN" | base64 -d | jq .
```

Ensure the authorization header is properly formatted:
```
authorization: Bearer YOUR_JWT_TOKEN
```

### Stream Timeout

Increase client timeout for long-lived streams:
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

stream, err := client.StreamMessages(ctx, req)
```

### Metadata Not Received

Ensure metadata is added to outgoing context:
```go
// Correct
ctx := metadata.AppendToOutgoingContext(context.Background(), "key", "value")

// Incorrect
ctx := metadata.NewIncomingContext(context.Background(), md)
```

### Proto Compilation Errors

Ensure you have the latest protoc compiler and plugins:
```bash
protoc --version
# Should be 3.0 or higher

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Additional Resources

- [gRPC Official Documentation](https://grpc.io/docs/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
- [gRPC Go Quick Start](https://grpc.io/docs/languages/go/quickstart/)
- [gRPC Python Quick Start](https://grpc.io/docs/languages/python/quickstart/)
- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)
