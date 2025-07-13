package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type uiModel struct {
	*sharedState
	input        textinput.Model
	viewport     viewport.Model
	vpContent    strings.Builder
	selectedChar string
	cursor       int
	wsConn       *websocket.Conn
	wsConnected  bool
}

type characterData struct {
	CharacterName string `json:"character_name"`
	ActionName    string `json:"action_name"`
	ActionTarget  string `json:"action_target"`
}

type senseAreaResponse struct {
	Characters    []characterData `json:"characters"`
	ResourceNodes []string        `json:"resource_nodes"`
}

type inventoryResponse struct {
	Items    map[string]int32 `json:"items"`
	Weight   int32            `json:"weight"`
	Capacity int32            `json:"capacity"`
}

type wsMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type chatMsgReceived struct {
	message string
	color   Color
}

type wsConnected struct{}

type wsError struct {
	err error
}

func InitUIModel(state *sharedState) *uiModel {
	m := uiModel{}
	m.sharedState = state
	cmd := textinput.New()
	cmd.Placeholder = "command"
	cmd.Width = 50
	cmd.Focus()
	m.input = cmd

	vp := viewport.New(m.width-2, m.height-5)
	m.viewport = vp
	m.vpContent.WriteString("Welcome!\n")
	m.viewport.SetContent(m.vpContent.String())

	return &m
}

func (m *uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if !m.wsConnected && m.userToken != "" {
		m.wsConnected = true
		return m.connectWebSocketCmd()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height - 5

		m.input.Width = msg.Width - 2
	case apiResMsg:
		output := m.colorStyle(msg.text, msg.color)
		m.vpContent.WriteString(output + "\n")
		m.viewport.SetContent(m.vpContent.String())
		m.viewport.GotoBottom()
	case chatMsgReceived:
		output := m.colorStyle(msg.message, msg.color)
		m.vpContent.WriteString(output + "\n")
		m.viewport.SetContent(m.vpContent.String())
		m.viewport.GotoBottom()
		return m.listenForMessagesCmd()
	case wsConnected:
		output := m.colorStyle("Connected to chat", Green)
		m.vpContent.WriteString(output + "\n")
		m.viewport.SetContent(m.vpContent.String())
		m.viewport.GotoBottom()
		return m.listenForMessagesCmd()
	case wsError:
		output := m.colorStyle("Chat connection error: "+msg.err.Error(), Red)
		m.vpContent.WriteString(output + "\n")
		m.viewport.SetContent(m.vpContent.String())
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.cursor = (m.cursor + 1) % 2
			if m.cursor == 1 {
				m.input.Blur()
			} else {
				m.input.Focus()
			}
		case "up", "down", "k", "j":
			if m.cursor == 1 {
				if msg.String() == "up" || msg.String() == "k" {
					m.viewport.ScrollUp(1)
				} else {
					m.viewport.ScrollDown(1)
				}
			}
		case "enter":
			if m.cursor == 0 {
				command := strings.Split(m.input.Value(), " ")
				m.input.SetValue("")
				var output string
				var outputColor Color
				switch command[0] {
				case "act":
					if m.selectedChar == "" {
						output = "No character selected. Use 'sel <character>' first"
						outputColor = Red
					} else {
						target := strings.ToUpper(strings.Join(command[1:], " "))
						return m.setAction(target)
					}
				case "sense":
					return m.getArea()
				case "inv":
					return m.getInventory()
				case "sel":
					if len(command) < 2 {
						output = "Usage: sel <character>"
						outputColor = Red
					} else {
						m.selectedChar = command[1]
						output = fmt.Sprintf("Selected %v", m.selectedChar)
						outputColor = Green
					}
				case "newchar":
					if len(command) > 2 {
						output = fmt.Sprintf("Too many arguments for \"%v\" command", command[0])
						outputColor = Red
					} else {
						return m.createCharacter(command[1])
					}
				case "say":
					if len(command) < 2 {
						output = "Usage: say <message>"
						outputColor = Red
					} else {
						message := strings.Join(command[1:], " ")
						return m.sendChatMessage(message)
					}
				case "idle":
					return m.setIdle()
				case "drop":
					if len(command) < 3 {
						output = "Usage: drop <item_name> <quantity>"
						outputColor = Red
					} else {
						quantity := command[len(command)-1]
						itemName := strings.Join(command[1:len(command)-1], " ")
						return m.dropItem(itemName, quantity)
					}
				case "echo":
					output = strings.Join(command[1:], " ")
					outputColor = Green
				default:
					output = fmt.Sprintf("Command \"%v\" not found", command[0])
					outputColor = Red
				}

				output = m.colorStyle(output, outputColor)

				m.vpContent.WriteString(output + "\n")
				m.viewport.SetContent(m.vpContent.String())
				m.viewport.GotoBottom()
			}
		}
	}

	m.input, cmd = m.input.Update(msg)

	return cmd
}

func (m *uiModel) View() string {
	vpColor := Blue
	cmdColor := Green
	if m.cursor == 1 {
		vpColor = Green
		cmdColor = Blue
	}

	viewportStyle := m.viewportStyle(m.viewport.View(), vpColor)
	inputStyle := m.inputStyle(m.input.View(), cmdColor)

	return m.uiView(viewportStyle, inputStyle)
}

func (m *uiModel) createCharacter(charName string) tea.Cmd {
	return func() tea.Msg {
		data := map[string]string{
			"name": charName,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"POST",
			m.apiUrl+"/characters",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		var bodyStr string
		var resColor Color
		if res.StatusCode == 201 {
			resColor = Green
			bodyStr = "Character Creation Successful"
		} else {
			resColor = Red
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = "Failed to parse error response"
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}

func (m *uiModel) setAction(target string) tea.Cmd {
	return func() tea.Msg {
		data := map[string]string{
			"character_name": m.selectedChar,
			"target":         target,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"PUT",
			m.apiUrl+"/characters",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		bodyStr := ""
		resColor := Red
		if res.StatusCode == 201 {
			caser := cases.Title(language.English)
			resColor = Green

			var response map[string]interface{}
			json.Unmarshal(body, &response)
			actionName := response["action_name"].(string)

			bodyStr = fmt.Sprintf(
				"%v started %v on %v",
				caser.String(m.selectedChar),
				caser.String(actionName),
				caser.String(target),
			)

		} else {
			resColor = Red
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = "Failed to parse error response"
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}

func (m *uiModel) getArea() tea.Cmd {
	return func() tea.Msg {
		bodyStr := ""
		if m.selectedChar == "" {
			return apiResMsg{Red, "No character selected"}
		}

		req, err := http.NewRequest(
			"GET",
			m.apiUrl+fmt.Sprintf("/sense/area/%v", m.selectedChar),
			nil,
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		resColor := Red
		if res.StatusCode == 200 {
			resColor = Green
			caser := cases.Title(language.English)
			var res senseAreaResponse
			if err := json.Unmarshal(body, &res); err != nil {
				bodyStr = err.Error()
			} else {
				bodyStr = "\n"
				if len(res.Characters) > 0 {
					bodyStr += "Characters\n"
					for _, value := range res.Characters {
						if value.ActionName == "IDLE" || value.ActionTarget == "" {
							bodyStr += fmt.Sprintf(
								"\t%v is idle\n",
								value.CharacterName,
							)
						} else {
							bodyStr += fmt.Sprintf(
								"\t%v is %v at %v\n",
								value.CharacterName,
								caser.String(value.ActionName),
								caser.String(value.ActionTarget),
							)
						}
					}
				}

				if len(res.ResourceNodes) > 0 {
					bodyStr += "Resources\n"
					for _, value := range res.ResourceNodes {
						resource := caser.String(value)
						bodyStr += fmt.Sprintf("\t%v\n", resource)
					}
				}
			}
		} else {
			resColor = Red
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = err.Error()
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}

func (m *uiModel) getInventory() tea.Cmd {
	return func() tea.Msg {
		bodyStr := ""
		if m.selectedChar == "" {
			return apiResMsg{Red, "No character selected"}
		}

		req, err := http.NewRequest(
			"GET",
			m.apiUrl+fmt.Sprintf("/inventory/%v", m.selectedChar),
			nil,
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		resColor := Red
		if res.StatusCode == 200 {
			resColor = Green
			caser := cases.Title(language.English)
			var res inventoryResponse
			if err := json.Unmarshal(body, &res); err != nil {
				bodyStr = err.Error()
			} else {
				bodyStr = "\n"
				if len(res.Items) > 0 {
					bodyStr += "Inventory\n"
					for name, quantity := range res.Items {
						bodyStr += fmt.Sprintf(
							"\t%v: %v\n",
							caser.String(name),
							quantity,
						)
					}
				}
				bodyStr += fmt.Sprintf("\nWeight: %d/%d", res.Weight, res.Capacity)
			}
		} else {
			resColor = Red
			bodyStr = fmt.Sprintf("Inventory get failed for %v", m.selectedChar)
		}

		return apiResMsg{resColor, bodyStr}
	}
}

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

func (m *uiModel) setIdle() tea.Cmd {
	return func() tea.Msg {
		if m.selectedChar == "" {
			return apiResMsg{Red, "No character selected. Use 'sel <character>' first"}
		}

		data := map[string]string{
			"character_name": m.selectedChar,
			"target":         "IDLE",
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"PUT",
			m.apiUrl+"/characters",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		bodyStr := ""
		resColor := Red
		if res.StatusCode == 201 {
			resColor = Green
			bodyStr = fmt.Sprintf("%v is now idle", m.selectedChar)
		} else {
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = "Failed to parse error response"
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}

func (m *uiModel) dropItem(itemName, quantityStr string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedChar == "" {
			return apiResMsg{Red, "No character selected. Use 'sel <character>' first"}
		}

		quantity, err := strconv.Atoi(quantityStr)
		if err != nil || quantity <= 0 {
			return apiResMsg{Red, "Invalid quantity. Must be a positive number"}
		}

		data := map[string]interface{}{
			"character_name": m.selectedChar,
			"item_name":      itemName,
			"quantity":       quantity,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"POST",
			m.apiUrl+"/inventory/drop",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		bodyStr := ""
		resColor := Red
		if res.StatusCode == 200 {
			resColor = Green
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err == nil {
				if message, ok := response["message"].(string); ok {
					bodyStr = message
				} else {
					bodyStr = "Item dropped successfully"
				}
			} else {
				bodyStr = "Item dropped successfully"
			}
		} else {
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = "Failed to parse error response"
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}
