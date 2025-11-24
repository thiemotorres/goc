package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConnectingScreen shows a spinner while connecting to the trainer
type ConnectingScreen struct {
	spinner spinner.Model
}

func NewConnectingScreen() *ConnectingScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return &ConnectingScreen{
		spinner: s,
	}
}

func (cs *ConnectingScreen) Init() tea.Cmd {
	return cs.spinner.Tick
}

func (cs *ConnectingScreen) Update(msg tea.Msg) (*ConnectingScreen, tea.Cmd) {
	var cmd tea.Cmd
	cs.spinner, cmd = cs.spinner.Update(msg)
	return cs, cmd
}

func (cs *ConnectingScreen) View() string {
	var b strings.Builder

	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("Connecting to Trainer"))
	b.WriteString("\n\n\n")

	connectingText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Render("Please wait while we connect to your trainer...")

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left,
		cs.spinner.View(),
		"  ",
		connectingText,
	))
	b.WriteString("\n\n")

	help := helpStyle.Render("esc: cancel")
	b.WriteString(help)

	return centerView(menuStyle.Render(b.String()))
}
