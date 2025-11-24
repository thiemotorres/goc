package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RideType represents the type of ride to start
type RideType int

const (
	RideFree RideType = iota
	RideERG
	RideRoute
)

// StartRideMenu is the start ride submenu
type StartRideMenu struct {
	items    []string
	selected int
}

func NewStartRideMenu() *StartRideMenu {
	return &StartRideMenu{
		items: []string{
			"Free Ride (no target)",
			"ERG Mode (fixed power)",
			"Ride a Route",
			"← Back",
		},
		selected: 0,
	}
}

func (m *StartRideMenu) MoveUp() {
	if m.selected > 0 {
		m.selected--
	}
}

func (m *StartRideMenu) MoveDown() {
	if m.selected < len(m.items)-1 {
		m.selected++
	}
}

func (m *StartRideMenu) Selected() int {
	return m.selected
}

func (m *StartRideMenu) View() string {
	var b strings.Builder

	title := titleStyle.Render("Start Ride")
	b.WriteString(title)
	b.WriteString("\n\n")

	for i, item := range m.items {
		cursor := "  "
		style := normalStyle
		if i == m.selected {
			cursor = "> "
			style = selectedStyle
		}
		b.WriteString(cursor + style.Render(item) + "\n")
	}

	help := helpStyle.Render("\n↑/↓: navigate • enter: select • esc: back")
	b.WriteString(help)

	return centerView(menuStyle.Render(b.String()))
}

func centerView(content string) string {
	return lipgloss.Place(80, 24, lipgloss.Center, lipgloss.Center, content)
}
