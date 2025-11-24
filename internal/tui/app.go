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
	ScreenScanner
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
	rideScreen      *RideScreen
	rideSession     *RideSession
	scannerScreen   *ScannerScreen
	connecting      bool
	connectStatus   string

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
			if a.rideSession != nil {
				return a, a.rideSession.Stop()
			}
			a.quitting = true
			return a, tea.Quit
		}

	case RideConnectingMsg:
		a.connectStatus = msg.Status
		return a, nil

	case RideConnectedMsg:
		a.connecting = false
		a.screen = ScreenRide
		// Start data loop
		return a, a.rideSession.StartDataLoop()

	case RideUpdateMsg:
		if a.rideScreen != nil {
			a.rideScreen.UpdateMetrics(msg.Power, msg.Cadence, msg.Speed)
			a.rideScreen.UpdateStats(msg.Elapsed, msg.Distance, msg.AvgPower, msg.AvgCadence, msg.AvgSpeed, msg.Elevation)
			a.rideScreen.UpdateStatus(msg.Gear, msg.Gradient, msg.Mode, msg.Paused)
		}
		// Continue data loop
		if a.rideSession != nil {
			return a, a.rideSession.StartDataLoop()
		}
		return a, nil

	case RideErrorMsg:
		a.connecting = false
		a.connectStatus = msg.Error.Error()
		// Return to menu after error
		a.screen = ScreenStartRide
		return a, nil

	case RideFinishedMsg:
		a.rideSession = nil
		a.rideScreen = nil
		a.screen = ScreenMainMenu
		return a, nil

	case ScanResultMsg:
		if a.scannerScreen != nil {
			a.scannerScreen.Update(msg)
		}
		return a, nil

	case DeviceSelectedMsg:
		// Save the selected device
		a.config.Bluetooth.TrainerAddress = msg.Address
		config.Save(a.config, config.DefaultConfigDir())
		// Update trainer settings display
		if a.trainerSettings != nil {
			a.trainerSettings.address = msg.Address
		}
		a.screen = ScreenTrainerSettings
		return a, nil
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
	case ScreenRide:
		return a.updateRide(msg)
	case ScreenScanner:
		return a.updateScanner(msg)
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
	case ScreenRide:
		if a.rideScreen != nil {
			return a.rideScreen.View()
		}
		return "Ride not started"
	case ScreenScanner:
		if a.scannerScreen != nil {
			return a.scannerScreen.View()
		}
		return "Scanner not loaded"
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
				return a, a.startRide(RideFree, nil)
			case 1: // ERG Mode
				// TODO: Show ERG watts input, for now start with 150W
				return a, a.startRide(RideERG, nil)
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
				return a, a.startRide(RideRoute, a.selectedRoute)
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
				a.scannerScreen = NewScannerScreen(a.config)
				a.screen = ScreenScanner
				return a, a.scannerScreen.StartScan()
			case 1: // Forget Saved Trainer
				a.config.Bluetooth.TrainerAddress = ""
				a.trainerSettings.address = ""
				config.Save(a.config, config.DefaultConfigDir())
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

func (a *App) updateRide(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.rideScreen != nil {
		return a, a.rideScreen.Update(msg)
	}
	return a, nil
}

func (a *App) updateScanner(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.scannerScreen == nil {
		return a, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if a.scannerScreen.scanning {
			return a, nil // Ignore keys while scanning
		}
		switch msg.String() {
		case "esc":
			a.screen = ScreenTrainerSettings
		case "up", "k":
			a.scannerScreen.MoveUp()
		case "down", "j":
			a.scannerScreen.MoveDown()
		case "r":
			// Retry scan
			a.scannerScreen = NewScannerScreen(a.config)
			return a, a.scannerScreen.StartScan()
		case "enter":
			if device := a.scannerScreen.SelectDevice(); device != nil {
				// Save selected device
				a.config.Bluetooth.TrainerAddress = device.Address
				config.Save(a.config, config.DefaultConfigDir())
				if a.trainerSettings != nil {
					a.trainerSettings.address = device.Address
				}
				a.screen = ScreenTrainerSettings
			} else {
				// Back selected
				a.screen = ScreenTrainerSettings
			}
		}
	}
	return a, nil
}

func (a *App) startRide(rideType RideType, route *RouteInfo) tea.Cmd {
	// Create ride session with real Bluetooth
	// Set mock=false to use actual trainer, mock=true for development testing
	session, err := NewRideSession(a.config, rideType, route, false)
	if err != nil {
		a.connectStatus = err.Error()
		return nil
	}

	a.rideSession = session
	a.rideScreen = NewRideScreen()

	// Set up callbacks
	a.rideScreen.SetCallbacks(
		func() { session.ShiftUp() },
		func() { session.ShiftDown() },
		func() { session.AdjustResistance(5) },
		func() { session.AdjustResistance(-5) },
		func() { session.TogglePause() },
		func() {
			// Stop ride and return to menu
			session.Stop()
			a.screen = ScreenMainMenu
			a.rideScreen = nil
			a.rideSession = nil
		},
	)

	a.connecting = true
	a.connectStatus = "Connecting to trainer..."

	// Start connection
	return session.Connect()
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
