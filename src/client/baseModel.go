package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Page int

const (
	Login Page = iota
	Signup
	UI
)

type apiResMsg struct {
	color Color
	text  string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type sharedState struct {
	*style
	currentPage Page
}

type baseModel struct {
	*sharedState
	login  *loginModel
	signup *signupModel
	ui     *uiModel
}

func (m baseModel) Init() tea.Cmd {
	return nil
}

func (m baseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	switch m.currentPage {
	case Login:
		cmd = m.login.Update(msg)
	case Signup:
		cmd = m.signup.Update(msg)
	case UI:
		cmd = m.ui.Update(msg)
	}

	return m, cmd
}

func (m baseModel) View() string {
	var s string
	switch m.currentPage {
	case Login:
		s = m.login.View()
	case Signup:
		s = m.signup.View()
	case UI:
		s = m.ui.View()
	default:
		s = "Invalid View"
	}

	return s
}
