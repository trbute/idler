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

type signupModel struct {
	*sharedState
	fields       []textinput.Model
	submit       string
	login        string
	subText      string
	subTextColor Color
	cursor       int
}

func InitSignupModel(state *sharedState) *signupModel {
	m := signupModel{}

	m.sharedState = state

	surname := textinput.New()
	surname.Placeholder = "surname"
	surname.Width = 30
	surname.Focus()
	m.fields = append(m.fields, surname)

	email := textinput.New()
	email.Placeholder = "email"
	email.Width = 30
	m.fields = append(m.fields, email)

	password := textinput.New()
	password.Placeholder = "password"
	password.Width = 30
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	m.fields = append(m.fields, password)

	passwordConfirm := textinput.New()
	passwordConfirm.Placeholder = "confirm password"
	passwordConfirm.Width = 30
	passwordConfirm.EchoMode = textinput.EchoPassword
	passwordConfirm.EchoCharacter = '•'
	m.fields = append(m.fields, passwordConfirm)

	m.submit = "submit"
	m.login = "login"

	return &m
}

func (m *signupModel) Init() tea.Cmd {
	return nil
}

func (m *signupModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case apiResMsg:
		m.subTextColor = msg.color
		m.subText = msg.text
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "enter":
			if m.cursor == len(m.fields)+1 && msg.String() == "enter" {
				m.currentPage = Login
			} else if m.cursor == len(m.fields) && msg.String() == "enter" {
				if m.fields[2].Value() != m.fields[3].Value() {
					m.subText = "Passwords do not match"
					m.subTextColor = Red
				} else {
					return m.createUser()
				}
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

func (m *signupModel) updateFields(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.fields))
	for i := range m.fields {
		m.fields[i], cmds[i] = m.fields[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *signupModel) View() string {
	var lines []string
	lines = append(lines, m.colorStyle("signup", Magenta))

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
	lines = append(lines, m.colorStyle(m.login, getColor(len(m.fields)+1)))
	lines = append(lines, m.colorStyle(m.subText, m.subTextColor))

	return m.centerStyle(m.borderStyle(strings.Join(lines, "\n"), Magenta, 30))
}

func (m *signupModel) createUser() tea.Cmd {
	return func() tea.Msg {
		data := map[string]string{
			"surname":  m.fields[0].Value(),
			"email":    m.fields[1].Value(),
			"password": m.fields[2].Value(),
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"POST",
			m.apiUrl+"/users",
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

		var bodyStr string
		var resColor Color
		if res.StatusCode == 201 {
			resColor = Green
			bodyStr = "User Creation Successful"
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
