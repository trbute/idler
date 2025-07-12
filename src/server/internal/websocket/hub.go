package websocket

import (
	"log"
	
	"github.com/google/uuid"
)

type Hub struct {
	clients    map[uuid.UUID]*Client
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
}

type Message struct {
	Type   string                 `json:"type"`
	UserID uuid.UUID              `json:"user_id,omitempty"`
	To     string                 `json:"to,omitempty"` // "all" or specific user ID
	Data   map[string]interface{} `json:"data"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userID] = client
			log.Printf("WebSocket client registered: %s (total clients: %d)", client.userID, len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				log.Printf("WebSocket client unregistered: %s (total clients: %d)", client.userID, len(h.clients))
			}

		case message := <-h.broadcast:
			log.Printf("Broadcasting message type '%s' to '%s', %d clients connected", message.Type, message.To, len(h.clients))
			if message.To == "all" {
				for _, client := range h.clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client.userID)
					}
				}
			} else if message.To != "" {
				if targetID, err := uuid.Parse(message.To); err == nil {
					if client, ok := h.clients[targetID]; ok {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(h.clients, client.userID)
						}
					}
				}
			}
		}
	}
}

func (h *Hub) SendToUser(userID uuid.UUID, msgType string, data map[string]interface{}) {
	message := &Message{
		Type: msgType,
		To:   userID.String(),
		Data: data,
	}
	h.broadcast <- message
}

func (h *Hub) SendToAll(msgType string, data map[string]interface{}) {
	message := &Message{
		Type: msgType,
		To:   "all",
		Data: data,
	}
	h.broadcast <- message
}

// Server-only methods for sending secure message types
func (h *Hub) SendNotificationToUser(userID uuid.UUID, message string, severity string) {
	data := map[string]interface{}{
		"message":  message,
		"severity": severity,
	}
	h.SendToUser(userID, "notification", data)
}

func (h *Hub) SendSystemMessage(message string) {
	data := map[string]interface{}{
		"message": message,
	}
	h.SendToAll("system", data)
}