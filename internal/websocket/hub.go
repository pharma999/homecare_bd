package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"context"
)

// MessageType identifies the kind of real-time event sent to clients.
type MessageType string

const (
	MsgEmergencyTriggered MessageType = "emergency_triggered"
	MsgEmergencyStatus    MessageType = "emergency_status"
	MsgAmbulanceLocation  MessageType = "ambulance_location"
	MsgPing               MessageType = "ping"
)

// Message is the envelope sent over every WebSocket connection.
type Message struct {
	Type    MessageType     `json:"type"`
	Room    string          `json:"room"`
	Payload json.RawMessage `json:"payload"`
}

// client wraps a single WebSocket connection.
type client struct {
	conn  *websocket.Conn
	rooms map[string]bool
	mu    sync.Mutex
}

func (cl *client) send(ctx context.Context, msg Message) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	if err := wsjson.Write(ctx, cl.conn, msg); err != nil {
		log.Printf("[WS] send error: %v", err)
	}
}

// Hub manages all active WebSocket connections and room subscriptions.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

var Global = &Hub{
	clients: make(map[*client]struct{}),
}

// Register adds a new connection to the hub.
func (h *Hub) Register(conn *websocket.Conn) *client {
	cl := &client{conn: conn, rooms: make(map[string]bool)}
	h.mu.Lock()
	h.clients[cl] = struct{}{}
	h.mu.Unlock()
	return cl
}

// Unregister removes a connection from the hub.
func (h *Hub) Unregister(cl *client) {
	h.mu.Lock()
	delete(h.clients, cl)
	h.mu.Unlock()
}

// Subscribe joins a client to a room (e.g. "emergency:abc123", "admin:emergencies").
func (h *Hub) Subscribe(cl *client, room string) {
	cl.mu.Lock()
	cl.rooms[room] = true
	cl.mu.Unlock()
}

// Broadcast sends a message to all clients subscribed to the given room.
func (h *Hub) Broadcast(room string, msgType MessageType, payload interface{}) {
	raw, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[WS] marshal error: %v", err)
		return
	}
	msg := Message{Type: msgType, Room: room, Payload: raw}
	ctx := context.Background()

	h.mu.RLock()
	defer h.mu.RUnlock()
	for cl := range h.clients {
		cl.mu.Lock()
		inRoom := cl.rooms[room]
		cl.mu.Unlock()
		if inRoom {
			go cl.send(ctx, msg)
		}
	}
}
