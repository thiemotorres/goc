package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("212")
	secondaryColor = lipgloss.Color("241")
	accentColor    = lipgloss.Color("229")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	menuStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(1, 2)

	selectedStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	helpStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			MarginTop(1)
)
