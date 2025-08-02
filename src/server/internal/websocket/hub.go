package websocket

import (
	"log"
	"time"

	"github.com/google/uuid"
)

type Hub struct {
	clients          map[string]*Client
	userConnections  map[uuid.UUID][]*ClientInfo
	broadcast        chan *Message
	register         chan *Client
	unregister       chan *Client
	maxConnections   int
}

type Message struct {
	Type   string                 `json:"type"`
	UserID uuid.UUID              `json:"user_id,omitempty"`
	To     string                 `json:"to,omitempty"` // "all" or specific user ID
	Data   map[string]interface{} `json:"data"`
}

type ClientInfo struct {
	client    *Client
	timestamp time.Time
}

func NewHub() *Hub {
	return &Hub{
		clients:         make(map[string]*Client),
		userConnections: make(map[uuid.UUID][]*ClientInfo),
		broadcast:       make(chan *Message, 256),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		maxConnections:  5,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.enforceConnectionLimit(client)
			h.clients[client.tokenID] = client
			
			clientInfo := &ClientInfo{
				client:    client,
				timestamp: time.Now(),
			}
			h.userConnections[client.userID] = append(h.userConnections[client.userID], clientInfo)

		case client := <-h.unregister:
			if _, ok := h.clients[client.tokenID]; ok {
				delete(h.clients, client.tokenID)
				h.removeUserConnection(client)
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
		select {
		case client.send <- &Message{
			Type: "error",
			Data: map[string]interface{}{"message": "Session expired. Please reconnect."},
		}:
		default:
		}
		
		h.unregister <- client
	}
}

func (h *Hub) enforceConnectionLimit(newClient *Client) {
	userConnections := h.userConnections[newClient.userID]
	
	if len(userConnections) >= h.maxConnections {
		oldestIndex := 0
		oldestTime := userConnections[0].timestamp
		
		for i, conn := range userConnections {
			if conn.timestamp.Before(oldestTime) {
				oldestIndex = i
				oldestTime = conn.timestamp
			}
		}
		
		oldestClient := userConnections[oldestIndex].client
		log.Printf("User %s exceeded connection limit (%d), disconnecting oldest connection", 
			newClient.userID.String(), h.maxConnections)
		
		select {
		case oldestClient.send <- &Message{
			Type: "error",
			Data: map[string]interface{}{"message": "Connection limit exceeded. Disconnecting oldest session."},
		}:
		default:
		}
		
		h.unregister <- oldestClient
	}
}

func (h *Hub) removeUserConnection(client *Client) {
	userConnections := h.userConnections[client.userID]
	
	for i, conn := range userConnections {
		if conn.client == client {
			h.userConnections[client.userID] = append(userConnections[:i], userConnections[i+1:]...)
			break
		}
	}
	
	if len(h.userConnections[client.userID]) == 0 {
		delete(h.userConnections, client.userID)
	}
}
