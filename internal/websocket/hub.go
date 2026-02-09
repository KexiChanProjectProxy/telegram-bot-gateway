package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/kexi/telegram-bot-gateway/internal/pubsub"
)

// Hub manages WebSocket connections
type Hub struct {
	clients       map[*Client]bool
	clientsMu     sync.RWMutex
	register      chan *Client
	unregister    chan *Client
	messageBroker *pubsub.MessageBroker
}

// NewHub creates a new WebSocket hub
func NewHub(messageBroker *pubsub.MessageBroker) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		messageBroker: messageBroker,
	}
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	log.Println("WebSocket hub starting")

	for {
		select {
		case <-ctx.Done():
			log.Println("WebSocket hub shutting down")
			h.closeAllClients()
			return

		case client := <-h.register:
			h.clientsMu.Lock()
			h.clients[client] = true
			h.clientsMu.Unlock()
			log.Printf("WebSocket client registered: %s (total: %d)", client.id, len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.clientsMu.Lock()
				delete(h.clients, client)
				h.clientsMu.Unlock()
				close(client.send)
				log.Printf("WebSocket client unregistered: %s (total: %d)", client.id, len(h.clients))
			}
		}
	}
}

// RegisterClient registers a new client
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// BroadcastToChat broadcasts a message to all clients subscribed to a chat
func (h *Hub) BroadcastToChat(chatID uint, message []byte) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for client := range h.clients {
		if client.IsSubscribedToChat(chatID) {
			select {
			case client.send <- message:
			default:
				// Client's send channel is full, skip
				log.Printf("Skipping slow client %s", client.id)
			}
		}
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	return len(h.clients)
}

// closeAllClients closes all client connections
func (h *Hub) closeAllClients() {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	for client := range h.clients {
		close(client.send)
	}
	h.clients = make(map[*Client]bool)
}

// Client represents a WebSocket client
type Client struct {
	id            string
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	subscriptions map[uint]bool // chat IDs
	subMu         sync.RWMutex
	userID        *uint
	apiKeyID      *uint
}

// NewClient creates a new WebSocket client
func NewClient(id string, hub *Hub, conn *websocket.Conn, userID *uint, apiKeyID *uint) *Client {
	return &Client{
		id:            id,
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[uint]bool),
		userID:        userID,
		apiKeyID:      apiKeyID,
	}
}

// IsSubscribedToChat checks if client is subscribed to a chat
func (c *Client) IsSubscribedToChat(chatID uint) bool {
	c.subMu.RLock()
	defer c.subMu.RUnlock()
	return c.subscriptions[chatID]
}

// SubscribeToChat subscribes the client to a chat
func (c *Client) SubscribeToChat(chatID uint) {
	c.subMu.Lock()
	c.subscriptions[chatID] = true
	c.subMu.Unlock()
	log.Printf("Client %s subscribed to chat %d", c.id, chatID)
}

// UnsubscribeFromChat unsubscribes the client from a chat
func (c *Client) UnsubscribeFromChat(chatID uint) {
	c.subMu.Lock()
	delete(c.subscriptions, chatID)
	c.subMu.Unlock()
	log.Printf("Client %s unsubscribed from chat %d", c.id, chatID)
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse client message
		c.handleClientMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage handles incoming messages from the client
func (c *Client) handleClientMessage(message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse client message: %v", err)
		c.sendError("Invalid message format")
		return
	}

	switch msg.Action {
	case "subscribe":
		if msg.ChatID > 0 {
			// TODO: Check ACL permissions before subscribing
			c.SubscribeToChat(msg.ChatID)
			c.sendAck("subscribed", msg.ChatID)
		}

	case "unsubscribe":
		if msg.ChatID > 0 {
			c.UnsubscribeFromChat(msg.ChatID)
			c.sendAck("unsubscribed", msg.ChatID)
		}

	case "ping":
		c.sendAck("pong", 0)

	default:
		c.sendError("Unknown action")
	}
}

// sendAck sends an acknowledgment message to the client
func (c *Client) sendAck(action string, chatID uint) {
	response := map[string]interface{}{
		"type":    "ack",
		"action":  action,
		"chat_id": chatID,
	}
	data, _ := json.Marshal(response)
	c.send <- data
}

// sendError sends an error message to the client
func (c *Client) sendError(errorMsg string) {
	response := map[string]interface{}{
		"type":  "error",
		"error": errorMsg,
	}
	data, _ := json.Marshal(response)
	c.send <- data
}

// ClientMessage represents a message from the client
type ClientMessage struct {
	Action string `json:"action"` // "subscribe", "unsubscribe", "ping"
	ChatID uint   `json:"chat_id,omitempty"`
}
