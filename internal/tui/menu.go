package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MainMenu is the main menu model
type MainMenu struct {
	items    []string
	selected int
}

// NewMainMenu creates a new main menu
func NewMainMenu() *MainMenu {
	return &MainMenu{
		items: []string{
			"Start Ride",
			"Browse Routes",
			"Ride History",
			"Settings",
			"Quit",
		},
		selected: 0,
	}
}

func (m *MainMenu) MoveUp() {
	if m.selected > 0 {
		m.selected--
	}
}

func (m *MainMenu) MoveDown() {
	if m.selected < len(m.items)-1 {
		m.selected++
	}
}

func (m *MainMenu) Selected() int {
	return m.selected
}

func (m *MainMenu) View() string {
	var b strings.Builder

	title := titleStyle.Render("goc - Indoor Cycling Trainer")
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

	help := helpStyle.Render("\n↑/↓: navigate • enter: select • q: quit")
	b.WriteString(help)

	return lipgloss.Place(
		80, 24,
		lipgloss.Center, lipgloss.Center,
		menuStyle.Render(b.String()),
	)
}
