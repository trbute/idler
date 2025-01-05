package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type loginModel struct {
	*sharedState
	fields       []textinput.Model
	submit       string
	signup       string
	subText      string
	subTextColor Color
	cursor       int
}

func InitLoginModel(state *sharedState) *loginModel {
	m := loginModel{}

	email := textinput.New()
	email.Prompt = ""
	email.Placeholder = "email"
	email.Focus()
	m.fields = append(m.fields, email)

	password := textinput.New()
	password.Prompt = ""
	password.Placeholder = "password"
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	m.fields = append(m.fields, password)

	m.submit = "submit"
	m.signup = "signup"
	m.sharedState = state

	return &m
}

func (m *loginModel) Init() tea.Cmd {
	return nil
}

func (m *loginModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case apiResMsg:
		m.subTextColor = msg.color
		m.subText = msg.text
		if m.subTextColor == Green {
			m.currentPage = UI
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "enter":
			if m.cursor == len(m.fields)+1 && msg.String() == "enter" {
				m.currentPage = Signup
			} else if m.cursor == len(m.fields) && msg.String() == "enter" {
				return m.loginUser()
			} else {
				m.cursor++
				if m.cursor < len(m.fields) {
					m.fields[m.cursor-1].Blur()
					m.fields[m.cursor].Focus()
				} else if m.cursor == len(m.fields) {
					m.fields[m.cursor-1].Blur()
				} else if m.cursor > len(m.fields)+1 {
					m.cursor = 0
					m.fields[m.cursor].Focus()
				}
			}
		}
	}

	cmd := m.updateFields(msg)

	return cmd
}

func (m *loginModel) updateFields(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.fields))
	for i := range m.fields {
		m.fields[i], cmds[i] = m.fields[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *loginModel) View() string {
	var lines []string
	lines = append(lines, m.colorStyle("login", Magenta))

	getColor := func(pos int) Color {
		if m.cursor == pos {
			return Green
		}
		return Blue
	}

	for i, field := range m.fields {
		lines = append(lines, m.colorStyle(field.View(), getColor(i)))
	}

	lines = append(lines, m.colorStyle(m.submit, getColor(len(m.fields))))
	lines = append(lines, m.colorStyle(m.signup, getColor(len(m.fields)+1)))
	lines = append(lines, m.colorStyle(m.subText, m.subTextColor))

	return m.centerStyle(m.borderStyle(strings.Join(lines, "\n"), Magenta))
}

func (m *loginModel) loginUser() tea.Cmd {
	return func() tea.Msg {
		data := map[string]string{
			"email":    m.fields[0].Value(),
			"password": m.fields[1].Value(),
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"POST",
			"http://127.0.0.1:8080/api/login",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")

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
		if res.StatusCode == 200 {
			resColor = Green
			bodyStr = "Login Successful"
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
