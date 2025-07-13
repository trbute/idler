package websocket

import (
	"github.com/google/uuid"
)

type Hub struct {
	clients    map[string]*Client // Use tokenID as key instead of userID
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
		clients:    make(map[string]*Client),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.tokenID] = client

		case client := <-h.unregister:
			if _, ok := h.clients[client.tokenID]; ok {
				delete(h.clients, client.tokenID)
				close(client.send)
			}

		case message := <-h.broadcast:
			if message.To == "all" {
				for _, client := range h.clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client.tokenID)
					}
				}
			} else if message.To != "" {
				if targetID, err := uuid.Parse(message.To); err == nil {
					for _, client := range h.clients {
						if client.userID == targetID {
							select {
							case client.send <- message:
							default:
								close(client.send)
								delete(h.clients, client.tokenID)
							}
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

func (h *Hub) DisconnectClientByToken(tokenID string) {
	if client, ok := h.clients[tokenID]; ok {
		// Send session expired message
		select {
		case client.send <- &Message{
			Type: "error",
			Data: map[string]interface{}{"message": "Session expired. Please reconnect."},
		}:
		default:
		}
		
		// Trigger unregister
		h.unregister <- client
	}
}
