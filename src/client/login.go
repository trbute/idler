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

type loginResult struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Email        string `json:"email"`
	Surname      string `json:"surname,omitempty"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type loginModel struct {
	*sharedState
	fields       []textinput.Model
	submit       string
	signup       string
	subText      string
	subTextColor Color
	cursor       int
}

func (m *loginModel) reset() {
	for i := range m.fields {
		m.fields[i].SetValue("")
	}
	
	m.cursor = 0
	m.subText = ""
	m.subTextColor = ""
	
	if len(m.fields) > 0 {
		m.fields[0].Focus()
		for i := 1; i < len(m.fields); i++ {
			m.fields[i].Blur()
		}
	}
}

func InitLoginModel(state *sharedState) *loginModel {
	m := loginModel{}

	email := textinput.New()
	email.Placeholder = "email"
	email.Width = 15
	email.Focus()
	m.fields = append(m.fields, email)

	password := textinput.New()
	password.Placeholder = "password"
	password.Width = 15
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
			// Clear any logout message when successfully logging in
			m.logoutMessage = ""
			m.logoutMsgColor = ""
			return func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
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
	if m.logoutMessage != "" {
		m.reset()
		m.subText = m.logoutMessage
		m.subTextColor = m.logoutMsgColor
		m.logoutMessage = ""
		m.logoutMsgColor = ""
	}
	
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

	return m.centerStyle(m.borderStyle(strings.Join(lines, "\n"), Magenta, 30))
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
			m.apiUrl+"/login",
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

		var resStr string
		resColor := Red
		
		if res.StatusCode == 200 {
			var loginRes loginResult
			err = json.Unmarshal(body, &loginRes)
			if err != nil {
				return apiResMsg{Red, "Failed to parse login response"}
			}
			
			m.userToken = loginRes.Token
			m.refreshToken = loginRes.RefreshToken
			m.surname = loginRes.Surname
			resColor = Green
			resStr = "Login Successful"
		} else {
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				resStr = string(body)
			} else {
				resStr = errResp.Error
			}
		}

		return apiResMsg{resColor, resStr}
	}
}
