package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type uiModel struct {
	*sharedState
	input        textinput.Model
	viewport     viewport.Model
	vpContent    strings.Builder
	selectedChar string
	cursor       int
}

func InitUIModel(state *sharedState) *uiModel {
	m := uiModel{}
	m.sharedState = state
	cmd := textinput.New()
	cmd.Placeholder = "command"
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
		m.input.SetValue("")
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
					m.viewport.LineUp(1)
				} else {
					m.viewport.LineDown(1)
				}
			}
		case "enter":
			if m.cursor == 0 {
				command := strings.Split(m.input.Value(), " ")
				var output string
				var outputColor Color
				switch command[0] {
				case "chop":
					return m.setAction(m.selectedChar, "WOODCUTTING", strings.ToUpper(strings.Join(command[1:], " ")))
				case "select_char":
					m.selectedChar = command[1]
					output = fmt.Sprintf("selected character:\"%v\"", m.selectedChar)
					outputColor = Green
				case "create_char":
					if len(command) > 2 {
						output = fmt.Sprintf("too many arguments for \"%v\" command", command[0])
						outputColor = Red
					} else {
						return m.createCharacter(command[1])
					}
				case "echo":
					output = strings.Join(command[1:], " ")
					outputColor = Green
				default:
					output = fmt.Sprintf("command \"%v\" not found", command[0])
					outputColor = Red
				}

				output = m.colorStyle(output, outputColor)

				m.vpContent.WriteString(output + "\n")
				m.viewport.SetContent(m.vpContent.String())
				m.input.SetValue("")
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
			"http://127.0.0.1:8080/api/characters",
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

		bodyStr := string(body)
		resColor := Red
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

func (m *uiModel) setAction(charName string, actionName string, target string) tea.Cmd {
	return func() tea.Msg {
		data := map[string]string{
			"character_name": charName,
			"action_name":    actionName,
			"target":         target,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"PUT",
			"http://127.0.0.1:8080/api/characters",
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

		bodyStr := string(body)
		resColor := Red
		if res.StatusCode == 201 {
			resColor = Green
			bodyStr = "Started choppin wood"
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
