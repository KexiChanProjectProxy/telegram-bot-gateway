package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/kexi/telegram-bot-gateway/api/proto"
	"github.com/kexi/telegram-bot-gateway/internal/pkg/jwt"
	"github.com/kexi/telegram-bot-gateway/internal/pubsub"
	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// Server wraps the gRPC server
type Server struct {
	grpcServer     *grpc.Server
	messageService *MessageServiceServer
	address        string
}

// NewServer creates a new gRPC server
func NewServer(
	address string,
	jwtService *jwt.Service,
	messageService *service.MessageService,
	chatService *service.ChatService,
	messageBroker *pubsub.MessageBroker,
) *Server {
	// Create interceptor for authentication
	authInterceptor := NewAuthInterceptor(jwtService)

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)

	// Create service server
	msgServiceServer := NewMessageServiceServer(messageService, chatService, messageBroker)

	// Register services
	pb.RegisterMessageServiceServer(grpcServer, msgServiceServer)

	return &Server{
		grpcServer:     grpcServer,
		messageService: msgServiceServer,
		address:        address,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	log.Printf("gRPC server listening on %s", s.address)

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	log.Println("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
}

// AuthInterceptor provides authentication for gRPC requests
type AuthInterceptor struct {
	jwtService *jwt.Service
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(jwtService *jwt.Service) *AuthInterceptor {
	return &AuthInterceptor{
		jwtService: jwtService,
	}
}

// Unary returns a unary server interceptor for authentication
func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract and validate token
		newCtx, err := a.authenticate(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// Stream returns a stream server interceptor for authentication
func (a *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Extract and validate token
		newCtx, err := a.authenticate(stream.Context())
		if err != nil {
			return err
		}

		wrapped := &wrappedStream{
			ServerStream: stream,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

// authenticate validates the JWT token and adds user info to context
func (a *AuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata provided")
	}

	// Get authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, fmt.Errorf("no authorization header")
	}

	token := authHeaders[0]
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Validate token
	claims, err := a.jwtService.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Add user info to metadata
	newMD := metadata.Pairs(
		"user-id", fmt.Sprintf("%d", claims.UserID),
		"username", claims.Username,
	)
	newCtx := metadata.NewIncomingContext(ctx, metadata.Join(md, newMD))

	return newCtx, nil
}

// wrappedStream wraps a grpc.ServerStream with a custom context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
