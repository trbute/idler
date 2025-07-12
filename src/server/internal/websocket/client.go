package websocket

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		
		if origin == "" {
			return true
		}
		
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "http://localhost:23234,http://127.0.0.1:23234"
		}
		
		origins := strings.Split(allowedOrigins, ",")
		for _, allowedOrigin := range origins {
			if strings.TrimSpace(allowedOrigin) == origin {
				return true
			}
		}
		
		log.Printf("WebSocket connection denied for origin: %s", origin)
		return false
	},
}

type Provider interface {
	GetSurnameById(ctx context.Context, userID uuid.UUID) (string, error)
	ValidateCharacterOwnership(ctx context.Context, characterName string, userID uuid.UUID) (bool, error)
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan *Message
	userID   uuid.UUID
	provider Provider
}

func ServeWS(hub *Hub, provider Provider, w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan *Message, 256),
		userID:   userID,
		provider: provider,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		message.UserID = c.userID

		switch message.Type {
		case "chat":
			if message.Data == nil {
				message.Data = make(map[string]interface{})
			}
			message.Data["user_id"] = c.userID.String()
			
			surname, err := c.provider.GetSurnameById(context.Background(), c.userID)
			if err != nil {
				log.Printf("Failed to get surname for user %s: %v", c.userID, err)
				surname = "Unknown User"
			}
			message.Data["surname"] = surname
			
			if characterNameInterface, exists := message.Data["character_name"]; exists {
				characterName, ok := characterNameInterface.(string)
				if ok && characterName != "" {
					isValid, err := c.provider.ValidateCharacterOwnership(context.Background(), characterName, c.userID)
					if err != nil {
						log.Printf("Error validating character %s for user %s: %v", characterName, c.userID, err)
						c.send <- &Message{
							Type: "error",
							Data: map[string]interface{}{"message": "Character not found"},
						}
						continue
					}
					
					if !isValid {
						log.Printf("User %s attempted to use character %s they don't own", c.userID, characterName)
						c.send <- &Message{
							Type: "error",
							Data: map[string]interface{}{"message": "You don't own that character"},
						}
						continue
					}
					
					message.Data["character_name"] = characterName
				}
			}
			
			message.To = "all"
			
			c.hub.broadcast <- &message
		case "ping":
			c.send <- &Message{
				Type: "pong",
				Data: map[string]interface{}{"timestamp": time.Now().Unix()},
			}
		case "notification", "system":
			log.Printf("Client %s attempted to send restricted message type: %s", c.userID, message.Type)
		default:
			log.Printf("Client %s sent unknown message type: %s", c.userID, message.Type)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}