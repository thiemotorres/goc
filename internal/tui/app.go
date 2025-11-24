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
	mainMenu      *MainMenu
	startRideMenu *StartRideMenu
}

// NewApp creates a new application
func NewApp() *App {
	return &App{
		screen:        ScreenMainMenu,
		mainMenu:      NewMainMenu(),
		startRideMenu: NewStartRideMenu(),
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
	case ScreenStartRide:
		return a.updateStartRide(msg)
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
	case ScreenStartRide:
		return a.startRideMenu.View()
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

func (a *App) updateStartRide(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.screen = ScreenMainMenu
		case "up", "k":
			a.startRideMenu.MoveUp()
		case "down", "j":
			a.startRideMenu.MoveDown()
		case "enter":
			switch a.startRideMenu.Selected() {
			case 0: // Free Ride
				// TODO: Start free ride
			case 1: // ERG Mode
				// TODO: Show ERG watts input
			case 2: // Ride a Route
				a.screen = ScreenBrowseRoutes
			case 3: // Back
				a.screen = ScreenMainMenu
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
