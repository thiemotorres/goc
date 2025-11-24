package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thiemotorres/goc/internal/config"
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
	mainMenu        *MainMenu
	startRideMenu   *StartRideMenu
	routesBrowser   *RoutesBrowser
	routePreview    *RoutePreview
	selectedRoute   *RouteInfo
	settingsMenu    *SettingsMenu
	trainerSettings *TrainerSettings
	historyView     *HistoryView

	// Config
	config *config.Config
}

// NewApp creates a new application
func NewApp(cfg *config.Config) *App {
	return &App{
		screen:        ScreenMainMenu,
		mainMenu:      NewMainMenu(),
		startRideMenu: NewStartRideMenu(),
		routesBrowser: NewRoutesBrowser(cfg.Routes.Folder),
		settingsMenu:  NewSettingsMenu(cfg),
		config:        cfg,
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
	case ScreenBrowseRoutes:
		return a.updateBrowseRoutes(msg)
	case ScreenRoutePreview:
		return a.updateRoutePreview(msg)
	case ScreenSettings:
		return a.updateSettings(msg)
	case ScreenTrainerSettings:
		return a.updateTrainerSettings(msg)
	case ScreenHistory:
		return a.updateHistory(msg)
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
	case ScreenBrowseRoutes:
		return a.routesBrowser.View()
	case ScreenRoutePreview:
		if a.routePreview != nil {
			return a.routePreview.View()
		}
		return "No route selected"
	case ScreenSettings:
		return a.settingsMenu.View()
	case ScreenTrainerSettings:
		if a.trainerSettings != nil {
			return a.trainerSettings.View()
		}
		return "Settings not loaded"
	case ScreenHistory:
		if a.historyView != nil {
			return a.historyView.View()
		}
		return "History not loaded"
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
				a.historyView = NewHistoryView()
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

func (a *App) updateBrowseRoutes(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.screen = ScreenStartRide
		case "up", "k":
			a.routesBrowser.MoveUp()
		case "down", "j":
			a.routesBrowser.MoveDown()
		case "enter":
			if route := a.routesBrowser.SelectedRoute(); route != nil {
				a.selectedRoute = route
				a.routePreview = NewRoutePreview(route)
				a.screen = ScreenRoutePreview
			} else {
				// Back selected
				a.screen = ScreenStartRide
			}
		}
	}
	return a, nil
}

func (a *App) updateRoutePreview(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.screen = ScreenBrowseRoutes
		case "left", "h":
			a.routePreview.MoveLeft()
		case "right", "l":
			a.routePreview.MoveRight()
		case "enter":
			if a.routePreview.Selected() == 0 {
				// Start ride with route
				// TODO: Connect and start
			} else {
				a.screen = ScreenBrowseRoutes
			}
		}
	}
	return a, nil
}

func (a *App) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.screen = ScreenMainMenu
		case "up", "k":
			a.settingsMenu.MoveUp()
		case "down", "j":
			a.settingsMenu.MoveDown()
		case "enter":
			switch a.settingsMenu.Selected() {
			case 0: // Trainer Connection
				a.trainerSettings = NewTrainerSettings(a.config.Bluetooth.TrainerAddress)
				a.screen = ScreenTrainerSettings
			case 1: // Routes Folder
				// TODO: Allow editing routes folder
			case 2: // Back
				a.screen = ScreenMainMenu
			}
		}
	}
	return a, nil
}

func (a *App) updateTrainerSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.screen = ScreenSettings
		case "up", "k":
			a.trainerSettings.MoveUp()
		case "down", "j":
			a.trainerSettings.MoveDown()
		case "enter":
			switch a.trainerSettings.Selected() {
			case 0: // Scan for Trainers
				// TODO: Start Bluetooth scan
			case 1: // Forget Saved Trainer
				a.config.Bluetooth.TrainerAddress = ""
				a.trainerSettings.address = ""
				// TODO: Save config
			case 2: // Back
				a.screen = ScreenSettings
			}
		}
	}
	return a, nil
}

func (a *App) updateHistory(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.screen = ScreenMainMenu
		case "up", "k":
			a.historyView.MoveUp()
		case "down", "j":
			a.historyView.MoveDown()
		case "enter":
			if ride := a.historyView.SelectedRide(); ride != nil {
				// TODO: Show ride detail
			} else {
				// Back selected
				a.screen = ScreenMainMenu
			}
		}
	}
	return a, nil
}

// Run starts the TUI application
func Run() error {
	cfg, err := config.Load(config.DefaultConfigDir())
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	p := tea.NewProgram(NewApp(cfg), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
