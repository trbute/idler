package main

import (
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

type Page int

const (
	Login Page = iota
	Signup
)

type newPageMsg struct {
	page Page
}

type baseModel struct {
	currentPage Page
	height      int
	width       int
	login       *loginModel
	signup      *signupModel
	renderer    *lg.Renderer
}

func (m baseModel) Init() tea.Cmd {
	return nil
}

func (m baseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case newPageMsg:
		m.currentPage = msg.page
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
	default:
		s = "Invalid View"
	}

	return m.renderer.NewStyle().Width(m.width).
		Height(m.height).
		Align(lg.Center).
		AlignVertical(lg.Center).
		Render(s)
}
