package main

import lg "github.com/charmbracelet/lipgloss"

type Color string

const (
	Red     Color = "1"
	Green   Color = "2"
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

func (s style) borderStyle(str string, color Color) string {
	return s.renderer.
		NewStyle().
		Width(30).
		Border(lg.RoundedBorder()).
		BorderForeground(lg.Color(color)).
		Padding(1).
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
