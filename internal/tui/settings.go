package tui

import (
	"fmt"
	"strings"

	"github.com/thiemotorres/goc/internal/config"
)

// SettingsMenu is the settings screen
type SettingsMenu struct {
	items    []string
	selected int
	config   *config.Config
}

func NewSettingsMenu(cfg *config.Config) *SettingsMenu {
	return &SettingsMenu{
		items: []string{
			"Trainer Connection",
			"Routes Folder",
			"← Back",
		},
		config: cfg,
	}
}

func (m *SettingsMenu) MoveUp() {
	if m.selected > 0 {
		m.selected--
	}
}

func (m *SettingsMenu) MoveDown() {
	if m.selected < len(m.items)-1 {
		m.selected++
	}
}

func (m *SettingsMenu) Selected() int {
	return m.selected
}

func (m *SettingsMenu) View() string {
	var b strings.Builder

	title := titleStyle.Render("Settings")
	b.WriteString(title)
	b.WriteString("\n\n")

	for i, item := range m.items {
		cursor := "  "
		style := normalStyle
		if i == m.selected {
			cursor = "> "
			style = selectedStyle
		}

		// Add current value for some items
		extra := ""
		switch i {
		case 0: // Trainer
			if m.config.Bluetooth.TrainerAddress != "" {
				extra = fmt.Sprintf(" (%s)", truncate(m.config.Bluetooth.TrainerAddress, 17))
			} else {
				extra = " (not set)"
			}
		case 1: // Routes
			extra = fmt.Sprintf("\n      %s", truncate(m.config.Routes.Folder, 40))
		}

		b.WriteString(cursor + style.Render(item+extra) + "\n")
	}

	help := helpStyle.Render("\n↑/↓: navigate • enter: select • esc: back")
	b.WriteString(help)

	return centerView(menuStyle.Render(b.String()))
}

// TrainerSettings shows trainer connection options
type TrainerSettings struct {
	items    []string
	selected int
	address  string
}

func NewTrainerSettings(address string) *TrainerSettings {
	return &TrainerSettings{
		items: []string{
			"Scan for Trainers",
			"Forget Saved Trainer",
			"← Back",
		},
		address: address,
	}
}

func (m *TrainerSettings) MoveUp() {
	if m.selected > 0 {
		m.selected--
	}
}

func (m *TrainerSettings) MoveDown() {
	if m.selected < len(m.items)-1 {
		m.selected++
	}
}

func (m *TrainerSettings) Selected() int {
	return m.selected
}

func (m *TrainerSettings) View() string {
	var b strings.Builder

	title := titleStyle.Render("Trainer Connection")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.address != "" {
		b.WriteString(fmt.Sprintf("Saved: %s\n\n", m.address))
	} else {
		b.WriteString("No trainer saved\n\n")
	}

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
