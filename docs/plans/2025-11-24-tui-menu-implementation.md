# TUI Menu System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace CLI-only interface with full Bubble Tea TUI menu system, including migrating the ride screen from termdash.

**Architecture:** Single Bubble Tea application with multiple screens (menu, ride, history). Screen transitions handled via model state. Config extended to support routes folder setting.

**Tech Stack:** Bubble Tea, Bubbles (list, textinput), ntcharts, Lip Gloss

---

## Task 1: Add Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add Bubble Tea ecosystem**

Run:
```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/NimbleMarkets/ntcharts@latest
```

**Step 2: Verify dependencies installed**

Run: `go mod tidy`
Expected: Clean exit, go.sum updated

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add Bubble Tea, Bubbles, Lip Gloss, ntcharts dependencies"
```

---

## Task 2: Add Routes Folder Config

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

**Step 1: Write failing test**

Add to `internal/config/config_test.go`:

```go
func TestRoutesFolder_Default(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "goc", "routes")
	if cfg.Routes.Folder != expected {
		t.Errorf("got %q, want %q", cfg.Routes.Folder, expected)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/... -run TestRoutesFolder -v`
Expected: FAIL - cfg.Routes undefined

**Step 3: Add RoutesConfig struct and field**

In `internal/config/config.go`, add after BluetoothConfig:

```go
// RoutesConfig holds route file settings
type RoutesConfig struct {
	Folder string `mapstructure:"folder"`
}
```

Add to Config struct:

```go
type Config struct {
	Trainer   TrainerConfig   `mapstructure:"trainer"`
	Shifter   ShifterConfig   `mapstructure:"shifter"`
	Bike      BikeConfig      `mapstructure:"bike"`
	Bluetooth BluetoothConfig `mapstructure:"bluetooth"`
	Routes    RoutesConfig    `mapstructure:"routes"`
	Display   DisplayConfig   `mapstructure:"display"`
	Controls  ControlsConfig  `mapstructure:"controls"`
}
```

**Step 4: Add default in setDefaults**

```go
// Routes defaults
home, _ := os.UserHomeDir()
v.SetDefault("routes.folder", filepath.Join(home, ".config", "goc", "routes"))
```

**Step 5: Add to Save function**

Add in Save():
```go
v.Set("routes.folder", cfg.Routes.Folder)
```

**Step 6: Run test to verify it passes**

Run: `go test ./internal/config/... -run TestRoutesFolder -v`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/config/
git commit -m "feat(config): add routes folder configuration"
```

---

## Task 3: Create TUI Package Structure

**Files:**
- Create: `internal/tui/app.go`
- Create: `internal/tui/styles.go`

**Step 1: Create styles.go with Lip Gloss styles**

Create `internal/tui/styles.go`:

```go
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
```

**Step 2: Create app.go with basic app model**

Create `internal/tui/app.go`:

```go
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
	screen       Screen
	prevScreen   Screen
	width        int
	height       int
	quitting     bool

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
```

**Step 3: Verify it compiles**

Run: `go build ./internal/tui/...`
Expected: Build error - MainMenu not defined (expected, we'll add it next)

**Step 4: Commit partial progress**

```bash
git add internal/tui/styles.go internal/tui/app.go
git commit -m "feat(tui): add Bubble Tea app structure and styles"
```

---

## Task 4: Create Main Menu Component

**Files:**
- Create: `internal/tui/menu.go`

**Step 1: Create menu.go**

```go
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
```

**Step 2: Verify it compiles**

Run: `go build ./internal/tui/...`
Expected: Success

**Step 3: Commit**

```bash
git add internal/tui/menu.go
git commit -m "feat(tui): add main menu component"
```

---

## Task 5: Wire Up Entry Point

**Files:**
- Modify: `main.go`

**Step 1: Update main.go to launch TUI when no args**

Replace the beginning of main() to check for no args:

```go
func main() {
	if len(os.Args) < 2 {
		// No args - launch TUI
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Rest of existing switch statement...
```

Add import:
```go
"github.com/thiemotorres/goc/internal/tui"
```

**Step 2: Test manually**

Run: `go build -o goc && ./goc`
Expected: TUI menu appears with 5 options

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: launch TUI menu when no args provided"
```

---

## Task 6: Create Start Ride Submenu

**Files:**
- Create: `internal/tui/startride.go`
- Modify: `internal/tui/app.go`

**Step 1: Create startride.go**

```go
package tui

import "strings"

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
```

**Step 2: Add to app.go - add field and update handler**

Add to App struct:
```go
startRideMenu *StartRideMenu
```

Add to NewApp():
```go
startRideMenu: NewStartRideMenu(),
```

Add case in Update() switch:
```go
case ScreenStartRide:
	return a.updateStartRide(msg)
```

Add case in View() switch:
```go
case ScreenStartRide:
	return a.startRideMenu.View()
```

Add update handler:
```go
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
```

**Step 3: Test manually**

Run: `go build -o goc && ./goc`
Expected: Can navigate to Start Ride submenu and back

**Step 4: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): add Start Ride submenu"
```

---

## Task 7: Create Route Browser

**Files:**
- Create: `internal/tui/routes.go`
- Modify: `internal/tui/app.go`

**Step 1: Create routes.go**

```go
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thiemotorres/goc/internal/gpx"
)

// RouteInfo holds summary info for a route
type RouteInfo struct {
	Path     string
	Name     string
	Distance float64 // meters
	Ascent   float64 // meters
	AvgGrade float64 // percent
}

// RoutesBrowser displays available GPX routes
type RoutesBrowser struct {
	routes   []RouteInfo
	selected int
	folder   string
	err      error
}

func NewRoutesBrowser(folder string) *RoutesBrowser {
	rb := &RoutesBrowser{folder: folder}
	rb.loadRoutes()
	return rb
}

func (rb *RoutesBrowser) loadRoutes() {
	rb.routes = nil
	rb.err = nil

	// Create folder if it doesn't exist
	if err := os.MkdirAll(rb.folder, 0755); err != nil {
		rb.err = err
		return
	}

	entries, err := os.ReadDir(rb.folder)
	if err != nil {
		rb.err = err
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".gpx") {
			continue
		}

		path := filepath.Join(rb.folder, entry.Name())
		route, err := gpx.Load(path)
		if err != nil {
			continue // Skip invalid files
		}

		name := route.Name
		if name == "" {
			name = strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		}

		var avgGrade float64
		if route.TotalDistance > 0 {
			avgGrade = (route.TotalAscent / route.TotalDistance) * 100
		}

		rb.routes = append(rb.routes, RouteInfo{
			Path:     path,
			Name:     name,
			Distance: route.TotalDistance,
			Ascent:   route.TotalAscent,
			AvgGrade: avgGrade,
		})
	}
}

func (rb *RoutesBrowser) MoveUp() {
	if rb.selected > 0 {
		rb.selected--
	}
}

func (rb *RoutesBrowser) MoveDown() {
	max := len(rb.routes) // includes Back option
	if rb.selected < max {
		rb.selected++
	}
}

func (rb *RoutesBrowser) Selected() int {
	return rb.selected
}

func (rb *RoutesBrowser) SelectedRoute() *RouteInfo {
	if rb.selected < len(rb.routes) {
		return &rb.routes[rb.selected]
	}
	return nil
}

func (rb *RoutesBrowser) View() string {
	var b strings.Builder

	title := titleStyle.Render("Browse Routes")
	b.WriteString(title)
	b.WriteString("\n\n")

	if rb.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", rb.err))
	} else if len(rb.routes) == 0 {
		b.WriteString(fmt.Sprintf("No routes found in:\n%s\n\n", rb.folder))
		b.WriteString("Add .gpx files to this folder.\n")
	} else {
		for i, route := range rb.routes {
			cursor := "  "
			style := normalStyle
			if i == rb.selected {
				cursor = "> "
				style = selectedStyle
			}
			line := fmt.Sprintf("%-20s %6.1f km  %5.0fm ↑  %4.1f%%",
				truncate(route.Name, 20),
				route.Distance/1000,
				route.Ascent,
				route.AvgGrade,
			)
			b.WriteString(cursor + style.Render(line) + "\n")
		}
	}

	// Back option
	cursor := "  "
	style := normalStyle
	if rb.selected == len(rb.routes) {
		cursor = "> "
		style = selectedStyle
	}
	b.WriteString("\n" + cursor + style.Render("← Back") + "\n")

	help := helpStyle.Render("\n↑/↓: navigate • enter: select • esc: back")
	b.WriteString(help)

	return centerView(menuStyle.Render(b.String()))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
```

**Step 2: Update app.go with routes browser**

Add to App struct:
```go
routesBrowser *RoutesBrowser
config        *config.Config
```

Update NewApp to accept config:
```go
func NewApp(cfg *config.Config) *App {
	return &App{
		screen:        ScreenMainMenu,
		mainMenu:      NewMainMenu(),
		startRideMenu: NewStartRideMenu(),
		routesBrowser: NewRoutesBrowser(cfg.Routes.Folder),
		config:        cfg,
	}
}
```

Add import for config package.

Add case in Update():
```go
case ScreenBrowseRoutes:
	return a.updateBrowseRoutes(msg)
```

Add case in View():
```go
case ScreenBrowseRoutes:
	return a.routesBrowser.View()
```

Add handler:
```go
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
				// TODO: Show route preview
				a.screen = ScreenRoutePreview
			} else {
				// Back selected
				a.screen = ScreenStartRide
			}
		}
	}
	return a, nil
}
```

Update Run() to load config:
```go
func Run() error {
	cfg, err := config.Load(config.DefaultConfigDir())
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	p := tea.NewProgram(NewApp(cfg), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
```

**Step 3: Test manually**

Run: `go build -o goc && ./goc`
Expected: Can navigate to Browse Routes, see empty folder message or routes if present

**Step 4: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): add route browser with GPX scanning"
```

---

## Task 8: Create Route Preview

**Files:**
- Create: `internal/tui/preview.go`
- Modify: `internal/tui/app.go`

**Step 1: Create preview.go with sparkline**

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/thiemotorres/goc/internal/gpx"
)

// RoutePreview shows route details before starting
type RoutePreview struct {
	route    *gpx.Route
	info     *RouteInfo
	selected int // 0 = Start, 1 = Back
}

func NewRoutePreview(info *RouteInfo) *RoutePreview {
	route, _ := gpx.Load(info.Path)
	return &RoutePreview{
		route: route,
		info:  info,
	}
}

func (rp *RoutePreview) MoveLeft() {
	if rp.selected > 0 {
		rp.selected--
	}
}

func (rp *RoutePreview) MoveRight() {
	if rp.selected < 1 {
		rp.selected++
	}
}

func (rp *RoutePreview) Selected() int {
	return rp.selected
}

func (rp *RoutePreview) View() string {
	var b strings.Builder

	title := titleStyle.Render(rp.info.Name)
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 40))
	b.WriteString("\n\n")

	// Stats
	b.WriteString(fmt.Sprintf("Distance:    %.1f km\n", rp.info.Distance/1000))
	b.WriteString(fmt.Sprintf("Elevation:   %.0fm ↑\n", rp.info.Ascent))
	b.WriteString(fmt.Sprintf("Avg Grade:   %.1f%%\n", rp.info.AvgGrade))

	if rp.route != nil {
		maxGrade := rp.findMaxGrade()
		b.WriteString(fmt.Sprintf("Max Grade:   %.1f%%\n", maxGrade))
	}

	b.WriteString("\n")

	// Elevation sparkline
	if rp.route != nil {
		sparkline := rp.generateSparkline(40)
		b.WriteString(sparkline)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Buttons
	startStyle := normalStyle
	backStyle := normalStyle
	if rp.selected == 0 {
		startStyle = selectedStyle
	} else {
		backStyle = selectedStyle
	}

	b.WriteString("        ")
	b.WriteString(startStyle.Render("[Start]"))
	b.WriteString("  ")
	b.WriteString(backStyle.Render("[Back]"))
	b.WriteString("\n")

	help := helpStyle.Render("\n←/→: select • enter: confirm")
	b.WriteString(help)

	return centerView(menuStyle.Render(b.String()))
}

func (rp *RoutePreview) findMaxGrade() float64 {
	if rp.route == nil || len(rp.route.Points) < 2 {
		return 0
	}

	var maxGrade float64
	for i := 1; i < len(rp.route.Points); i++ {
		prev := rp.route.Points[i-1]
		curr := rp.route.Points[i]
		dist := curr.Distance - prev.Distance
		if dist > 0 {
			grade := ((curr.Elevation - prev.Elevation) / dist) * 100
			if grade > maxGrade {
				maxGrade = grade
			}
		}
	}
	return maxGrade
}

func (rp *RoutePreview) generateSparkline(width int) string {
	if rp.route == nil || len(rp.route.Points) == 0 {
		return ""
	}

	// Sample elevations
	elevations := make([]float64, width)
	for i := 0; i < width; i++ {
		dist := (float64(i) / float64(width-1)) * rp.route.TotalDistance
		elevations[i] = rp.route.ElevationAt(dist)
	}

	// Find min/max
	minEle, maxEle := elevations[0], elevations[0]
	for _, e := range elevations {
		if e < minEle {
			minEle = e
		}
		if e > maxEle {
			maxEle = e
		}
	}

	// Sparkline characters
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var sb strings.Builder
	eleRange := maxEle - minEle
	if eleRange == 0 {
		eleRange = 1
	}

	for _, e := range elevations {
		normalized := (e - minEle) / eleRange
		idx := int(normalized * float64(len(chars)-1))
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		sb.WriteRune(chars[idx])
	}

	return sb.String()
}
```

**Step 2: Update app.go**

Add to App struct:
```go
routePreview  *RoutePreview
selectedRoute *RouteInfo
```

Add case in Update():
```go
case ScreenRoutePreview:
	return a.updateRoutePreview(msg)
```

Add case in View():
```go
case ScreenRoutePreview:
	if a.routePreview != nil {
		return a.routePreview.View()
	}
	return "No route selected"
```

Update browse routes handler to create preview:
```go
case "enter":
	if route := a.routesBrowser.SelectedRoute(); route != nil {
		a.selectedRoute = route
		a.routePreview = NewRoutePreview(route)
		a.screen = ScreenRoutePreview
	} else {
		a.screen = ScreenStartRide
	}
```

Add handler:
```go
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
```

**Step 3: Test manually**

Run: `go build -o goc && ./goc`
Expected: Can select route and see preview with sparkline

**Step 4: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): add route preview with elevation sparkline"
```

---

## Task 9: Create Settings Menu

**Files:**
- Create: `internal/tui/settings.go`
- Modify: `internal/tui/app.go`

**Step 1: Create settings.go**

```go
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
```

**Step 2: Update app.go**

Add to App struct:
```go
settingsMenu    *SettingsMenu
trainerSettings *TrainerSettings
```

Add to NewApp():
```go
settingsMenu: NewSettingsMenu(cfg),
```

Add cases and handlers (similar pattern to previous menus).

**Step 3: Test and commit**

```bash
git add internal/tui/
git commit -m "feat(tui): add settings menu with trainer and routes options"
```

---

## Task 10: Create History View

**Files:**
- Create: `internal/tui/history.go`
- Modify: `internal/tui/app.go`

**Step 1: Create history.go**

```go
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/thiemotorres/goc/internal/data"
)

// HistoryView shows past rides
type HistoryView struct {
	rides    []data.RideSummary
	selected int
	err      error
}

func NewHistoryView() *HistoryView {
	hv := &HistoryView{}
	hv.loadRides()
	return hv
}

func (hv *HistoryView) loadRides() {
	store, err := data.NewStore(data.DefaultDataDir())
	if err != nil {
		hv.err = err
		return
	}
	defer store.Close()

	hv.rides, hv.err = store.ListRides()
}

func (hv *HistoryView) MoveUp() {
	if hv.selected > 0 {
		hv.selected--
	}
}

func (hv *HistoryView) MoveDown() {
	max := len(hv.rides)
	if hv.selected < max {
		hv.selected++
	}
}

func (hv *HistoryView) Selected() int {
	return hv.selected
}

func (hv *HistoryView) SelectedRide() *data.RideSummary {
	if hv.selected < len(hv.rides) {
		return &hv.rides[hv.selected]
	}
	return nil
}

func (hv *HistoryView) View() string {
	var b strings.Builder

	title := titleStyle.Render("Ride History")
	b.WriteString(title)
	b.WriteString("\n\n")

	if hv.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", hv.err))
	} else if len(hv.rides) == 0 {
		b.WriteString("No rides recorded yet.\n")
	} else {
		for i, ride := range hv.rides {
			cursor := "  "
			style := normalStyle
			if i == hv.selected {
				cursor = "> "
				style = selectedStyle
			}

			date := ride.StartTime.Format("Jan 02")
			duration := formatDuration(ride.Duration)
			name := ride.GPXName
			if name == "" {
				name = "Free Ride"
			}

			line := fmt.Sprintf("%-6s  %-16s  %8s  %4.0fW avg",
				date,
				truncate(name, 16),
				duration,
				ride.AvgPower,
			)
			b.WriteString(cursor + style.Render(line) + "\n")
		}
	}

	// Back option
	cursor := "  "
	style := normalStyle
	if hv.selected == len(hv.rides) {
		cursor = "> "
		style = selectedStyle
	}
	b.WriteString("\n" + cursor + style.Render("← Back") + "\n")

	help := helpStyle.Render("\n↑/↓: navigate • enter: view • esc: back")
	b.WriteString(help)

	return centerView(menuStyle.Render(b.String()))
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
```

**Step 2: Update app.go with history handling**

**Step 3: Test and commit**

```bash
git add internal/tui/
git commit -m "feat(tui): add ride history view"
```

---

## Task 11: Remove Old TUI (termdash)

**Files:**
- Delete: `internal/tui/tui.go` (old termdash implementation)
- Modify: `cmd/ride.go`

**Step 1: Rename old tui.go**

```bash
mv internal/tui/tui.go internal/tui/tui_old.go
```

**Step 2: Update cmd/ride.go to not use old TUI**

Remove TUI creation and use from ride.go temporarily - we'll add back the Bubble Tea ride screen next.

For now, make it run headless or print to console.

**Step 3: Commit**

```bash
git add internal/tui/ cmd/ride.go
git commit -m "refactor: remove termdash TUI in preparation for Bubble Tea ride screen"
```

---

## Task 12: Create Bubble Tea Ride Screen

**Files:**
- Create: `internal/tui/ride.go`

This is the largest task - creating the ride screen with ntcharts.

**Step 1: Create ride.go with basic structure**

```go
package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/NimbleMarkets/ntcharts/linechart"

	"github.com/thiemotorres/goc/internal/bluetooth"
	"github.com/thiemotorres/goc/internal/simulation"
)

// RideScreen is the active ride display
type RideScreen struct {
	width  int
	height int

	// Charts
	powerChart   *linechart.Model
	cadenceChart *linechart.Model
	speedChart   *linechart.Model

	// Data
	powerData   []float64
	cadenceData []float64
	speedData   []float64
	maxPoints   int

	// State
	elapsed    time.Duration
	distance   float64
	avgPower   float64
	avgCadence float64
	avgSpeed   float64
	elevation  float64
	gradient   float64
	gear       string
	mode       string
	paused     bool

	// Callbacks
	onShiftUp   func()
	onShiftDown func()
	onResUp     func()
	onResDown   func()
	onPause     func()
	onQuit      func()
}

func NewRideScreen() *RideScreen {
	// Create charts
	powerChart := linechart.New(40, 8)
	cadenceChart := linechart.New(40, 8)
	speedChart := linechart.New(40, 8)

	return &RideScreen{
		powerChart:   &powerChart,
		cadenceChart: &cadenceChart,
		speedChart:   &speedChart,
		maxPoints:    300,
	}
}

func (rs *RideScreen) SetCallbacks(shiftUp, shiftDown, resUp, resDown, pause, quit func()) {
	rs.onShiftUp = shiftUp
	rs.onShiftDown = shiftDown
	rs.onResUp = resUp
	rs.onResDown = resDown
	rs.onPause = pause
	rs.onQuit = quit
}

func (rs *RideScreen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if rs.onShiftUp != nil {
				rs.onShiftUp()
			}
		case "down", "j":
			if rs.onShiftDown != nil {
				rs.onShiftDown()
			}
		case "right", "l":
			if rs.onResUp != nil {
				rs.onResUp()
			}
		case "left", "h":
			if rs.onResDown != nil {
				rs.onResDown()
			}
		case " ":
			if rs.onPause != nil {
				rs.onPause()
			}
		case "q":
			if rs.onQuit != nil {
				rs.onQuit()
			}
		}
	}
	return nil
}

func (rs *RideScreen) UpdateMetrics(power, cadence, speed float64) {
	rs.powerData = append(rs.powerData, power)
	rs.cadenceData = append(rs.cadenceData, cadence)
	rs.speedData = append(rs.speedData, speed)

	if len(rs.powerData) > rs.maxPoints {
		rs.powerData = rs.powerData[1:]
		rs.cadenceData = rs.cadenceData[1:]
		rs.speedData = rs.speedData[1:]
	}

	// Update charts
	rs.powerChart.SetData(rs.powerData)
	rs.cadenceChart.SetData(rs.cadenceData)
	rs.speedChart.SetData(rs.speedData)
}

func (rs *RideScreen) UpdateStats(elapsed time.Duration, distance, avgPower, avgCadence, avgSpeed, elevation float64) {
	rs.elapsed = elapsed
	rs.distance = distance
	rs.avgPower = avgPower
	rs.avgCadence = avgCadence
	rs.avgSpeed = avgSpeed
	rs.elevation = elevation
}

func (rs *RideScreen) UpdateStatus(gear string, gradient float64, mode string, paused bool) {
	rs.gear = gear
	rs.gradient = gradient
	rs.mode = mode
	rs.paused = paused
}

func (rs *RideScreen) View() string {
	var b strings.Builder

	// Title
	title := "goc - Indoor Cycling Trainer"
	if rs.paused {
		title += " [PAUSED]"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Stats row
	stats := fmt.Sprintf(
		"Time: %s  Distance: %.1f km  Elevation: +%.0fm",
		formatDuration(rs.elapsed),
		rs.distance/1000,
		rs.elevation,
	)
	b.WriteString(stats)
	b.WriteString("\n")

	avgs := fmt.Sprintf(
		"Avg Power: %.0fW  Avg Cadence: %.0f rpm  Avg Speed: %.1f km/h",
		rs.avgPower, rs.avgCadence, rs.avgSpeed,
	)
	b.WriteString(avgs)
	b.WriteString("\n\n")

	// Charts (side by side if width allows)
	powerView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(0, 1).
		Render("Power (W)\n" + rs.powerChart.View())

	cadenceView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("45")).
		Padding(0, 1).
		Render("Cadence (rpm)\n" + rs.cadenceChart.View())

	speedView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(0, 1).
		Render("Speed (km/h)\n" + rs.speedChart.View())

	charts := lipgloss.JoinHorizontal(lipgloss.Top, powerView, cadenceView, speedView)
	b.WriteString(charts)
	b.WriteString("\n\n")

	// Status bar
	status := fmt.Sprintf("Gear: %s  Gradient: %+.1f%%  Mode: %s", rs.gear, rs.gradient, rs.mode)
	b.WriteString(status)
	b.WriteString("\n")

	// Controls help
	help := helpStyle.Render("[↑↓] Shift  [←→] Resistance  [Space] Pause  [q] Quit")
	b.WriteString(help)

	return b.String()
}
```

**Step 2: Integrate with app.go**

Add ride screen to App and handle transitions.

**Step 3: Test and commit**

```bash
git add internal/tui/
git commit -m "feat(tui): add Bubble Tea ride screen with ntcharts"
```

---

## Task 13: Wire Up Trainer Connection and Ride Flow

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `cmd/ride.go`

Connect the TUI to actually start rides:
1. When user confirms ride type, initiate Bluetooth connection
2. Show connecting status
3. On success, switch to ride screen
4. Handle ride completion, return to menu

This task involves integrating existing `bluetooth.Manager` and `simulation.Engine` code.

**Step 1: Add ride state to App**

**Step 2: Create connection flow**

**Step 3: Test end-to-end**

**Step 4: Commit**

```bash
git add internal/tui/ cmd/
git commit -m "feat(tui): wire up trainer connection and ride flow"
```

---

## Task 14: Final Cleanup and Testing

**Files:**
- Delete: `internal/tui/tui_old.go`
- Update: `go.mod` (remove termdash)

**Step 1: Remove termdash dependency**

```bash
go mod tidy
```

**Step 2: Test all flows manually**

- Launch TUI with no args
- Navigate all menus
- Start free ride
- Start ERG ride
- Browse and select route
- View history
- Change settings

**Step 3: Final commit**

```bash
git add -A
git commit -m "chore: remove termdash, complete TUI migration"
```

---

## Summary

| Task | Description |
|------|-------------|
| 1 | Add Bubble Tea dependencies |
| 2 | Add routes folder config |
| 3 | Create TUI package structure |
| 4 | Create main menu component |
| 5 | Wire up entry point |
| 6 | Create Start Ride submenu |
| 7 | Create route browser |
| 8 | Create route preview |
| 9 | Create settings menu |
| 10 | Create history view |
| 11 | Remove old termdash TUI |
| 12 | Create Bubble Tea ride screen |
| 13 | Wire up trainer connection |
| 14 | Final cleanup |
