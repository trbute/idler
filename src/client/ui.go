package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
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
	m.vpContent.WriteString("Welcome!\nType '?' for help with commands.\n")
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
				case "?":
					if len(command) == 1 {
						return m.showHelp()
					} else {
						return m.showCommandHelp(command[1])
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