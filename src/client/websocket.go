package main

import (
	"fmt"
	"net/url"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

func (m *uiModel) connectWebSocketCmd() tea.Cmd {
	return func() tea.Msg {
		if m.userToken == "" {
			return wsError{err: fmt.Errorf("no user token available")}
		}

		wsURL := m.wsUrl + "?token=" + url.QueryEscape(m.userToken)

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			return wsError{err: err}
		}

		m.wsConn = conn
		return wsConnected{}
	}
}

func (m *uiModel) listenForMessagesCmd() tea.Cmd {
	return func() tea.Msg {
		if m.wsConn == nil {
			return wsError{err: fmt.Errorf("no websocket connection")}
		}

		var msg wsMessage
		err := m.wsConn.ReadJSON(&msg)
		if err != nil {
			return wsError{err: err}
		}

		switch msg.Type {
		case "chat":
			if data, ok := msg.Data["message"].(string); ok {
				characterName := ""
				surname := ""

				if char, exists := msg.Data["character_name"].(string); exists {
					characterName = char
				}
				if sur, exists := msg.Data["surname"].(string); exists {
					surname = sur
				}

				var displayName string
				if characterName != "" && surname != "" {
					displayName = fmt.Sprintf("%s %s", characterName, surname)
				} else if surname != "" {
					displayName = surname
				} else {
					displayName = "Unknown User"
				}

				chatMsg := fmt.Sprintf("[%s]: %s", displayName, data)
				return chatMsgReceived{message: chatMsg, color: Blue}
			}
		case "notification":
			if data, ok := msg.Data["message"].(string); ok {
				notificationMsg := fmt.Sprintf("⚠ %s", data)
				return chatMsgReceived{message: notificationMsg, color: Magenta}
			}
		case "error":
			if data, ok := msg.Data["message"].(string); ok {
				errorMsg := fmt.Sprintf("Error: %s", data)
				
				// If it's a session expiration, close the connection to trigger reconnect logic
				if data == "Session expired. Please reconnect." {
					if m.wsConn != nil {
						m.wsConn.Close()
						m.wsConn = nil
					}
					return wsError{err: fmt.Errorf("session expired")}
				}
				
				return chatMsgReceived{message: errorMsg, color: Red}
			}
		}

		return m.listenForMessagesCmd()
	}
}

func (m *uiModel) sendChatMessage(message string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedChar == "" {
			return apiResMsg{Red, "No character selected. Use 'sel <character>' first"}
		}

		if m.wsConn == nil {
			return apiResMsg{Red, "Not connected to chat. Connection will retry automatically."}
		}

		chatMsg := wsMessage{
			Type: "chat",
			Data: map[string]interface{}{
				"message":        message,
				"character_name": m.selectedChar,
				"surname":        m.surname,
			},
		}

		err := m.wsConn.WriteJSON(chatMsg)
		if err != nil {
			return apiResMsg{Red, "Failed to send message: " + err.Error()}
		}

		return nil
	}
}