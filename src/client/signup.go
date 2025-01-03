package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

type signupModel struct {
	fields   []textinput.Model
	submit   string
	signup   string
	cursor   int
	renderer *lg.Renderer
}

func InitSignupModel(renderer *lg.Renderer) *signupModel {
	m := signupModel{}

	user := textinput.New()
	user.Placeholder = "user"
	user.Focus()
	m.fields = append(m.fields, user)

	password := textinput.New()
	password.Placeholder = "password"
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	m.fields = append(m.fields, password)

	passwordConfirm := textinput.New()
	passwordConfirm.Placeholder = "confirm"
	passwordConfirm.EchoMode = textinput.EchoPassword
	passwordConfirm.EchoCharacter = '•'
	m.fields = append(m.fields, passwordConfirm)

	m.submit = "[ submit ]"
	m.signup = "[ login ]"
	m.renderer = renderer
	return &m
}

func (m *signupModel) Init() tea.Cmd {
	return nil
}

func (m *signupModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "enter":
			if m.cursor == len(m.fields)+1 && msg.String() == "enter" {
				return func() tea.Msg {
					return newPageMsg{page: Login}
				}
			} else {
				m.cursor++
				if m.cursor == len(m.fields) {
					m.fields[m.cursor-1].Blur()
					m.submit = "[[submit]]"
				} else if m.cursor == len(m.fields)+1 {
					m.submit = "[submit]"
					m.signup = "[[login]]"
				} else if m.cursor > len(m.fields)+1 {
					m.signup = "[login]"
					m.cursor = 0
					m.fields[m.cursor].Focus()
				} else {
					m.fields[m.cursor-1].Blur()
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
	s := "signup\n"
	for _, field := range m.fields {
		s += fmt.Sprintf("%s\n", field.View())
	}
	s += fmt.Sprintf("%s\n", m.submit)
	s += fmt.Sprintf("%s\n", m.signup)

	return m.renderer.NewStyle().Foreground(lg.Color("10")).Render(s)
}
