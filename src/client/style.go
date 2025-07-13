package main

import lg "github.com/charmbracelet/lipgloss"

type Color string

const (
	Red     Color = "1"
	Green   Color = "2"
	Yellow  Color = "3"
	Blue    Color = "4"
	Magenta Color = "5"
	Cyan    Color = "6"
)

type style struct {
	renderer *lg.Renderer
	height   int
	width    int
}

func (s style) centerStyle(str string) string {
	return s.renderer.
		NewStyle().
		Width(s.width).
		Height(s.height).
		Align(lg.Center).
		AlignVertical(lg.Center).
		Render(str)
}

func (s style) borderStyle(str string, color Color, width int) string {
	return s.renderer.
		NewStyle().
		Width(width).
		Border(lg.RoundedBorder()).
		BorderForeground(lg.Color(color)).
		Align(lg.Center).
		AlignVertical(lg.Center).
		Render(str)
}

func (s style) colorStyle(str string, color Color) string {
	return s.renderer.
		NewStyle().
		Foreground(lg.Color(color)).
		Render(str)
}

func (s style) viewportStyle(str string, borderColor Color) string {
	return s.renderer.NewStyle().
		Height(s.height - 5).
		Width(s.width - 2).
		Border(lg.RoundedBorder()).
		BorderForeground(lg.Color(borderColor)).Render(str)
}

func (s style) inputStyle(str string, borderColor Color) string {
	return s.renderer.NewStyle().
		Width(s.width - 2).
		Border(lg.RoundedBorder()).
		BorderForeground(lg.Color(borderColor)).Render(str)
}

func (s style) uiView(viewport string, input string) string {
	return lg.JoinVertical(
		lg.Left,
		viewport,
		input,
	)
}
