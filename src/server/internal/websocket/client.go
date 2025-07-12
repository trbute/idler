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
	maxMessageSize = 512 * 1024 // 512KB
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		
		// Allow connections without Origin header (non-browser clients like our Go client)
		if origin == "" {
			return true
		}
		
		// Get allowed origins from environment variable
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			// If no origins specified, allow localhost for development
			allowedOrigins = "http://localhost:23234,http://127.0.0.1:23234"
		}
		
		// Split and check each allowed origin
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

type SurnameProvider interface {
	GetSurnameById(ctx context.Context, userID uuid.UUID) (string, error)
}

type Client struct {
	hub             *Hub
	conn            *websocket.Conn
	send            chan *Message
	userID          uuid.UUID
	surnameProvider SurnameProvider
}

func ServeWS(hub *Hub, surnameProvider SurnameProvider, w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:             hub,
		conn:            conn,
		send:            make(chan *Message, 256),
		userID:          userID,
		surnameProvider: surnameProvider,
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

		// Handle different message types - only allow specific client types
		switch message.Type {
		case "chat":
			// Server controls the username - never trust client
			if message.Data == nil {
				message.Data = make(map[string]interface{})
			}
			message.Data["user_id"] = c.userID.String()
			
			// Get surname from database
			surname, err := c.surnameProvider.GetSurnameById(context.Background(), c.userID)
			if err != nil {
				log.Printf("Failed to get surname for user %s: %v", c.userID, err)
				surname = "Unknown User"
			}
			message.Data["surname"] = surname
			
			// Include character name from client message if provided
			if characterName, exists := message.Data["character_name"]; exists {
				message.Data["character_name"] = characterName
			}
			
			// Set To field to broadcast to all users
			message.To = "all"
			
			c.hub.broadcast <- &message
		case "ping":
			// Send pong back to this client
			c.send <- &Message{
				Type: "pong",
				Data: map[string]interface{}{"timestamp": time.Now().Unix()},
			}
		case "notification", "system":
			// SECURITY: Never allow clients to send notifications or system messages
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