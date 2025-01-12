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
	Items map[string]int32 `json:"items"`
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
		m.viewport.GotoBottom()
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
				case "wc":
					return m.setAction(
						"WOODCUTTING",
						strings.ToUpper(strings.Join(command[1:], " ")),
					)
				case "sense":
					return m.getArea()
				case "inv":
					return m.getInventory()
				case "sel":
					m.selectedChar = command[1]
					output = fmt.Sprintf("Selected %v", m.selectedChar)
					outputColor = Green
				case "newchar":
					if len(command) > 2 {
						output = fmt.Sprintf("Too many arguments for \"%v\" command", command[0])
						outputColor = Red
					} else {
						return m.createCharacter(command[1])
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

func (m *uiModel) setAction(actionName string, target string) tea.Cmd {
	return func() tea.Msg {
		data := map[string]string{
			"character_name": m.selectedChar,
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

		bodyStr := ""
		resColor := Red
		if res.StatusCode == 201 {
			caser := cases.Title(language.English)
			resColor = Green
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
			fmt.Sprintf("http://127.0.0.1:8080/api/sense/area/%v", m.selectedChar),
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
					bodyStr += fmt.Sprint("Characters\n")
					for _, value := range res.Characters {
						bodyStr += fmt.Sprintf(
							"\t%v is %v at %v\n",
							value.CharacterName,
							caser.String(value.ActionName),
							caser.String(value.ActionTarget),
						)
					}
				}

				if len(res.ResourceNodes) > 0 {
					bodyStr += fmt.Sprint("Resources\n")
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
			fmt.Sprintf("http://127.0.0.1:8080/api/inventory/%v", m.selectedChar),
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
					bodyStr += fmt.Sprint("Inventory\n")
					for name, quantity := range res.Items {
						bodyStr += fmt.Sprintf(
							"\t%v: %v\n",
							caser.String(name),
							quantity,
						)
					}
				}
			}
		} else {
			resColor = Red
			bodyStr = fmt.Sprintf("Inventory get failed for %v", m.selectedChar)
		}

		return apiResMsg{resColor, bodyStr}
	}
}
