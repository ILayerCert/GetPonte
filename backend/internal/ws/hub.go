// Package ws implements a WebSocket hub for real-time room-based messaging.
package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"

	"github.com/interactionlayer/ponte/internal/auth"
	"github.com/interactionlayer/ponte/internal/chat"
	roomPkg "github.com/interactionlayer/ponte/internal/room"
)

// ----- Wire types (sent/received over WebSocket) -----

// InboundMessage is a message received from a WebSocket client.
type InboundMessage struct {
	Type    string `json:"type"`    // "chat_message", "typing"
	Content string `json:"content"` // message body (for chat_message)
}

// OutboundMessage is a message sent to WebSocket clients.
type OutboundMessage struct {
	Type      string    `json:"type"`                // "chat_message", "user_joined", "user_left", "typing"
	MessageID string    `json:"message_id,omitempty"` // set for persisted chat messages
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content,omitempty"`
	RoomID    string    `json:"room_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ----- Client -----

// Client represents a single WebSocket connection.
type Client struct {
	Conn     *websocket.Conn
	UserID   uuid.UUID
	Username string
	RoomID   uuid.UUID
}

// ----- Hub -----

// Hub manages all WebSocket connections, grouped by room.
type Hub struct {
	mu       sync.RWMutex
	rooms    map[uuid.UUID]map[*Client]struct{} // roomID → set of clients
	chatRepo *chat.Repository
	roomRepo *roomPkg.Repository
	secret   string
}

// NewHub creates a new WebSocket hub.
func NewHub(chatRepo *chat.Repository, roomRepo *roomPkg.Repository, jwtSecret string) *Hub {
	return &Hub{
		rooms:    make(map[uuid.UUID]map[*Client]struct{}),
		chatRepo: chatRepo,
		roomRepo: roomRepo,
		secret:   jwtSecret,
	}
}

// UpgradeMiddleware returns a Fiber handler that checks the WebSocket upgrade
// prerequisites (valid token + room membership) before allowing the upgrade.
func (h *Hub) UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		tokenStr := c.Query("token")
		roomIDStr := c.Query("room_id")
		if tokenStr == "" || roomIDStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "token and room_id query parameters are required",
			})
		}

		claims, err := auth.ValidateToken(h.secret, tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		roomID, err := uuid.Parse(roomIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room_id"})
		}

		// Verify membership.
		isMember, err := h.roomRepo.IsMember(c.Context(), roomID, claims.UserID)
		if err != nil || !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not a member of this room"})
		}

		// Store validated data in locals for the upgrade handler.
		c.Locals("ws_user_id", claims.UserID)
		c.Locals("ws_username", claims.Username)
		c.Locals("ws_room_id", roomID)

		return c.Next()
	}
}

// Handler returns the websocket.Handler that manages the client lifecycle.
func (h *Hub) Handler() fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		userID := conn.Locals("ws_user_id").(uuid.UUID)
		username := conn.Locals("ws_username").(string)
		roomID := conn.Locals("ws_room_id").(uuid.UUID)

		client := &Client{
			Conn:     conn,
			UserID:   userID,
			Username: username,
			RoomID:   roomID,
		}

		h.addClient(client)
		defer h.removeClient(client)

		// Notify room that a user joined.
		h.broadcast(roomID, OutboundMessage{
			Type:      "user_joined",
			UserID:    userID.String(),
			Username:  username,
			RoomID:    roomID.String(),
			Timestamp: time.Now().UTC(),
		}, nil)

		// Read loop.
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				// Connection closed or errored.
				break
			}

			var msg InboundMessage
			if err := json.Unmarshal(raw, &msg); err != nil {
				log.Printf("ws: invalid message from %s: %v", username, err)
				continue
			}

			h.handleMessage(client, msg)
		}
	})
}

// addClient registers a client in the room's connection set.
func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[c.RoomID]; !ok {
		h.rooms[c.RoomID] = make(map[*Client]struct{})
	}
	h.rooms[c.RoomID][c] = struct{}{}
	log.Printf("ws: %s joined room %s (%d connections)", c.Username, c.RoomID, len(h.rooms[c.RoomID]))
}

// removeClient unregisters a client and notifies the room.
func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	if clients, ok := h.rooms[c.RoomID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.rooms, c.RoomID)
		}
	}
	h.mu.Unlock()

	_ = c.Conn.Close()

	// Notify room that a user left.
	h.broadcast(c.RoomID, OutboundMessage{
		Type:      "user_left",
		UserID:    c.UserID.String(),
		Username:  c.Username,
		RoomID:    c.RoomID.String(),
		Timestamp: time.Now().UTC(),
	}, nil)

	log.Printf("ws: %s left room %s", c.Username, c.RoomID)
}

// handleMessage processes an inbound WebSocket message.
func (h *Hub) handleMessage(c *Client, msg InboundMessage) {
	switch msg.Type {
	case "chat_message":
		if msg.Content == "" {
			return
		}

		// Persist message to the database.
		saved, err := h.chatRepo.Create(context.Background(), c.RoomID, c.UserID, msg.Content, "text")
		if err != nil {
			log.Printf("ws: failed to persist message: %v", err)
			return
		}

		// Broadcast to the room (including the sender).
		h.broadcast(c.RoomID, OutboundMessage{
			Type:      "chat_message",
			MessageID: saved.ID.String(),
			UserID:    c.UserID.String(),
			Username:  c.Username,
			Content:   saved.Content,
			RoomID:    c.RoomID.String(),
			Timestamp: saved.CreatedAt,
		}, nil)

	case "typing":
		// Broadcast typing indicator to everyone except the sender.
		h.broadcast(c.RoomID, OutboundMessage{
			Type:      "typing",
			UserID:    c.UserID.String(),
			Username:  c.Username,
			RoomID:    c.RoomID.String(),
			Timestamp: time.Now().UTC(),
		}, c)

	default:
		log.Printf("ws: unknown message type %q from %s", msg.Type, c.Username)
	}
}

// broadcast sends an outbound message to all clients in a room.
// If exclude is non-nil, that client is skipped.
func (h *Hub) broadcast(roomID uuid.UUID, msg OutboundMessage, exclude *Client) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws: marshal error: %v", err)
		return
	}

	h.mu.RLock()
	clients := h.rooms[roomID]
	h.mu.RUnlock()

	for c := range clients {
		if c == exclude {
			continue
		}
		if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("ws: write error to %s: %v", c.Username, err)
		}
	}
}
