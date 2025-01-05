package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

type uiModel struct {
	*sharedState
	fields       []textinput.Model
	subText      string
	subTextColor Color
	cursor       int
}

func InitUIModel(state *sharedState) *uiModel {
	m := uiModel{}
	m.subText = "You did it"
	m.subTextColor = Magenta
	m.sharedState = state
	return &m
}

func (m *uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case apiResMsg:
		m.subTextColor = msg.color
		m.subText = msg.text
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "enter":
			m.cursor++
		}
	}

	cmd := m.updateFields(msg)

	return cmd
}

func (m *uiModel) updateFields(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.fields))
	for i := range m.fields {
		m.fields[i], cmds[i] = m.fields[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *uiModel) View() string {
	return m.renderer.NewStyle().Foreground(lg.Color(m.subTextColor)).Render(m.subText)
}
