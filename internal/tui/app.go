package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents the current screen
type Screen int

const (
	ScreenMainMenu Screen = iota
	ScreenStartRide
	ScreenBrowseRoutes
	ScreenRoutePreview
	ScreenHistory
	ScreenRideDetail
	ScreenSettings
	ScreenTrainerSettings
	ScreenRoutesSettings
	ScreenRide
)

// App is the main application model
type App struct {
	screen     Screen
	prevScreen Screen
	width      int
	height     int
	quitting   bool

	// Sub-models
	mainMenu *MainMenu
}

// NewApp creates a new application
func NewApp() *App {
	return &App{
		screen:   ScreenMainMenu,
		mainMenu: NewMainMenu(),
	}
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			a.quitting = true
			return a, tea.Quit
		}
	}

	// Delegate to current screen
	switch a.screen {
	case ScreenMainMenu:
		return a.updateMainMenu(msg)
	}

	return a, nil
}

func (a *App) View() string {
	if a.quitting {
		return ""
	}

	switch a.screen {
	case ScreenMainMenu:
		return a.mainMenu.View()
	default:
		return "Unknown screen"
	}
}

func (a *App) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			a.quitting = true
			return a, tea.Quit
		case "up", "k":
			a.mainMenu.MoveUp()
		case "down", "j":
			a.mainMenu.MoveDown()
		case "enter":
			switch a.mainMenu.Selected() {
			case 0: // Start Ride
				a.screen = ScreenStartRide
			case 1: // Browse Routes
				a.screen = ScreenBrowseRoutes
			case 2: // History
				a.screen = ScreenHistory
			case 3: // Settings
				a.screen = ScreenSettings
			case 4: // Quit
				a.quitting = true
				return a, tea.Quit
			}
		}
	}
	return a, nil
}

// Run starts the TUI application
func Run() error {
	p := tea.NewProgram(NewApp(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
