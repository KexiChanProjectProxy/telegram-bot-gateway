package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/kexi/telegram-bot-gateway/api/proto"
	"github.com/kexi/telegram-bot-gateway/internal/pubsub"
	"github.com/kexi/telegram-bot-gateway/internal/service"
)

// MessageServiceServer implements the gRPC MessageService
type MessageServiceServer struct {
	pb.UnimplementedMessageServiceServer
	messageService *service.MessageService
	chatService    *service.ChatService
	messageBroker  *pubsub.MessageBroker
}

// NewMessageServiceServer creates a new message service server
func NewMessageServiceServer(
	messageService *service.MessageService,
	chatService *service.ChatService,
	messageBroker *pubsub.MessageBroker,
) *MessageServiceServer {
	return &MessageServiceServer{
		messageService: messageService,
		chatService:    chatService,
		messageBroker:  messageBroker,
	}
}

// StreamMessages streams messages for specified chats
func (s *MessageServiceServer) StreamMessages(req *pb.StreamMessagesRequest, stream pb.MessageService_StreamMessagesServer) error {
	ctx := stream.Context()

	// Validate authentication from metadata
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, "authentication required")
	}

	log.Printf("gRPC: User %d streaming messages for %d chats", userID, len(req.ChatIds))

	// Convert chat IDs
	chatIDs := make([]uint, len(req.ChatIds))
	for i, id := range req.ChatIds {
		chatIDs[i] = uint(id)
	}

	// Subscribe to message broker
	eventChan, err := s.messageBroker.SubscribeToMultipleChats(ctx, chatIDs)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}

	// Stream events to client
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-eventChan:
			if !ok {
				return nil
			}

			// Convert to protobuf message
			pbEvent := &pb.MessageEvent{
				Type:          event.Type,
				ChatId:        uint64(event.ChatID),
				MessageId:     uint64(event.MessageID),
				TelegramId:    event.TelegramID,
				BotId:         uint64(event.BotID),
				Direction:     event.Direction,
				Text:          event.Text,
				FromUsername:  event.FromUsername,
				MessageType:   event.Payload["message_type"].(string),
				Timestamp:     event.Timestamp.Unix(),
			}

			// Send to client
			if err := stream.Send(pbEvent); err != nil {
				return err
			}
		}
	}
}

// StreamChatMessages streams messages for a single chat
func (s *MessageServiceServer) StreamChatMessages(req *pb.StreamChatMessagesRequest, stream pb.MessageService_StreamChatMessagesServer) error {
	ctx := stream.Context()

	// Validate authentication
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, "authentication required")
	}

	// TODO: Check ACL permissions for this chat

	log.Printf("gRPC: User %d streaming chat %d", userID, req.ChatId)

	// Subscribe to chat
	eventChan, err := s.messageBroker.SubscribeToChat(ctx, uint(req.ChatId))
	if err != nil {
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}

	// Stream events
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-eventChan:
			if !ok {
				return nil
			}

			pbEvent := &pb.MessageEvent{
				Type:         event.Type,
				ChatId:       uint64(event.ChatID),
				MessageId:    uint64(event.MessageID),
				TelegramId:   event.TelegramID,
				BotId:        uint64(event.BotID),
				Direction:    event.Direction,
				Text:         event.Text,
				FromUsername: event.FromUsername,
				Timestamp:    event.Timestamp.Unix(),
			}

			if err := stream.Send(pbEvent); err != nil {
				return err
			}
		}
	}
}

// SendMessage sends a message to a chat
func (s *MessageServiceServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// Validate authentication
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// TODO: Check ACL permissions
	// TODO: Integrate with Telegram Bot API to actually send message

	return &pb.SendMessageResponse{
		Success: true,
		Message: "Message queued for delivery",
		QueuedAt: time.Now().Unix(),
	}, nil
}

// GetMessages retrieves historical messages
func (s *MessageServiceServer) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	// Validate authentication
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Convert cursor
	var cursor *time.Time
	if req.Cursor != 0 {
		t := time.Unix(req.Cursor, 0)
		cursor = &t
	}

	limit := int(req.Limit)
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Get messages
	messages, err := s.messageService.ListMessages(ctx, uint(req.ChatId), cursor, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get messages: %v", err)
	}

	// Convert to protobuf
	pbMessages := make([]*pb.Message, len(messages))
	for i, msg := range messages {
		pbMessages[i] = &pb.Message{
			Id:             uint64(msg.ID),
			ChatId:         uint64(msg.ChatID),
			TelegramId:     msg.TelegramID,
			FromUsername:   msg.FromUsername,
			FromFirstName:  msg.FromFirstName,
			FromLastName:   msg.FromLastName,
			Direction:      msg.Direction,
			MessageType:    msg.MessageType,
			Text:           msg.Text,
			SentAt:         msg.SentAt.Unix(),
			CreatedAt:      msg.CreatedAt.Unix(),
		}
		if msg.FromUserID != nil {
			pbMessages[i].FromUserId = *msg.FromUserID
		}
		if msg.ReplyToMessageID != nil {
			pbMessages[i].ReplyToMessageId = *msg.ReplyToMessageID
		}
	}

	// Determine if there are more messages
	hasMore := len(messages) == limit

	var nextCursor int64
	if hasMore && len(messages) > 0 {
		nextCursor = messages[len(messages)-1].SentAt.Unix()
	}

	return &pb.GetMessagesResponse{
		Messages:   pbMessages,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

// getUserIDFromContext extracts user ID from gRPC metadata
func getUserIDFromContext(ctx context.Context) (uint, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("no metadata")
	}

	userIDs := md.Get("user-id")
	if len(userIDs) == 0 {
		return 0, fmt.Errorf("no user-id in metadata")
	}

	var userID uint
	if _, err := fmt.Sscanf(userIDs[0], "%d", &userID); err != nil {
		return 0, fmt.Errorf("invalid user-id: %w", err)
	}

	return userID, nil
}
