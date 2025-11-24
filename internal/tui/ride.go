package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
	"github.com/thiemotorres/goc/internal/gpx"
)

// RideScreen is the active ride display
type RideScreen struct {
	width  int
	height int

	// Route
	route     *RouteInfo
	routeView *RouteView

	// Charts
	powerChart   streamlinechart.Model
	cadenceChart streamlinechart.Model
	speedChart   streamlinechart.Model
	maxPoints    int

	// Current values
	power   float64
	cadence float64
	speed   float64

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

func NewRideScreen(route *RouteInfo) *RideScreen {
	// Create charts with appropriate dimensions
	// Width and height will be adjusted in View() based on terminal size
	powerChart := streamlinechart.New(60, 15)
	cadenceChart := streamlinechart.New(60, 15)
	speedChart := streamlinechart.New(60, 15)

	// Load GPX route if provided
	var routeView *RouteView
	if route != nil {
		gpxRoute, err := gpx.Load(route.Path)
		if err == nil {
			routeView = NewRouteView(route, gpxRoute, 60, 15)
		}
	}

	return &RideScreen{
		route:        route,
		routeView:    routeView,
		powerChart:   powerChart,
		cadenceChart: cadenceChart,
		speedChart:   speedChart,
		maxPoints:    300, // ~5 minutes of data at 1 update/sec
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
		case "tab":
			if rs.routeView != nil {
				rs.routeView.ToggleMode()
			}
		case "q":
			if rs.onQuit != nil {
				rs.onQuit()
			}
		}
	case tea.WindowSizeMsg:
		rs.width = msg.Width
		rs.height = msg.Height
		// Resize route view if it exists
		if rs.routeView != nil {
			leftWidth := int(float64(rs.width) * 0.4)
			routeHeight := int(float64(rs.height-4) * 0.6)
			rs.routeView.Resize(leftWidth-8, routeHeight-8)
		}
	}
	return nil
}

func (rs *RideScreen) UpdateMetrics(power, cadence, speed float64) {
	rs.power = power
	rs.cadence = cadence
	rs.speed = speed

	// Push new data to charts
	rs.powerChart.Push(power)
	rs.cadenceChart.Push(cadence)
	rs.speedChart.Push(speed)
}

func (rs *RideScreen) UpdateStats(elapsed time.Duration, distance, avgPower, avgCadence, avgSpeed, elevation float64) {
	rs.elapsed = elapsed
	rs.distance = distance
	rs.avgPower = avgPower
	rs.avgCadence = avgCadence
	rs.avgSpeed = avgSpeed
	rs.elevation = elevation

	// Update route view with current position
	if rs.routeView != nil {
		rs.routeView.Update(distance, rs.gradient)
	}
}

func (rs *RideScreen) UpdateStatus(gear string, gradient float64, mode string, paused bool) {
	rs.gear = gear
	rs.gradient = gradient
	rs.mode = mode
	rs.paused = paused
}

func (rs *RideScreen) View() string {
	if rs.width == 0 || rs.height == 0 {
		return "Initializing..."
	}

	// Calculate layout dimensions
	leftWidth := int(float64(rs.width) * 0.4)
	rightWidth := rs.width - leftWidth - 2 // Account for borders

	// Title
	title := "goc - Indoor Cycling Trainer"
	if rs.paused {
		title += " [PAUSED]"
	}

	// Build left column (Route + Stats)
	leftColumn := rs.buildLeftColumn(leftWidth, rs.height-4)

	// Build right column (Charts + Status)
	rightColumn := rs.buildRightColumn(rightWidth, rs.height-4)

	// Join columns
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	// Add title and render
	var b strings.Builder
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(content)

	return b.String()
}

func (rs *RideScreen) buildLeftColumn(width, height int) string {
	// Route view (top 60% of left column)
	routeHeight := int(float64(height) * 0.6)
	routeView := rs.buildRouteView(width-4, routeHeight-4)
	routePanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(width - 4).
		Height(routeHeight - 2).
		Render("┤ Route ├\n" + routeView)

	// Stats view (bottom 40% of left column)
	statsHeight := height - routeHeight
	statsView := rs.buildStatsView(width-4, statsHeight-4)
	statsPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(width - 4).
		Height(statsHeight - 2).
		Render("┤ Stats ├\n" + statsView)

	return lipgloss.JoinVertical(lipgloss.Left, routePanel, statsPanel)
}

func (rs *RideScreen) buildRightColumn(width, height int) string {
	chartHeight := int(float64(height) * 0.25)

	// Update chart dimensions
	rs.powerChart.Resize(width-8, chartHeight-4)
	rs.cadenceChart.Resize(width-8, chartHeight-4)
	rs.speedChart.Resize(width-8, chartHeight-4)

	// Draw charts
	rs.powerChart.Draw()
	rs.cadenceChart.Draw()
	rs.speedChart.Draw()

	// Power chart
	powerPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(1).
		Width(width - 4).
		Height(chartHeight - 2).
		Render(fmt.Sprintf("┤ Power: %.0f W ├\n%s", rs.power, rs.powerChart.View()))

	// Cadence chart
	cadencePanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("45")).
		Padding(1).
		Width(width - 4).
		Height(chartHeight - 2).
		Render(fmt.Sprintf("┤ Cadence: %.0f rpm ├\n%s", rs.cadence, rs.cadenceChart.View()))

	// Speed chart
	speedPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(1).
		Width(width - 4).
		Height(chartHeight - 2).
		Render(fmt.Sprintf("┤ Speed: %.1f km/h ├\n%s", rs.speed, rs.speedChart.View()))

	// Status panel
	statusView := rs.buildStatusView(width-4, chartHeight-4)
	statusPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(width - 4).
		Height(chartHeight - 2).
		Render("┤ Status ├\n" + statusView)

	return lipgloss.JoinVertical(lipgloss.Left,
		powerPanel,
		cadencePanel,
		speedPanel,
		statusPanel,
	)
}

func (rs *RideScreen) buildRouteView(width, height int) string {
	if rs.routeView != nil {
		return rs.routeView.View()
	} else if rs.route != nil {
		// Fallback if routeView failed to load
		return fmt.Sprintf("Route: %s\n\nDistance: %.1f km\nElevation: +%.0fm\n\n[Route failed to load]",
			rs.route.Name,
			rs.route.Distance/1000,
			rs.route.Ascent)
	}
	return "Free Ride\n\nNo route selected"
}

func (rs *RideScreen) buildStatsView(width, height int) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Time:      %s\n", formatDuration(rs.elapsed)))
	b.WriteString(fmt.Sprintf("Distance:  %.2f km\n", rs.distance/1000))
	b.WriteString(fmt.Sprintf("Elevation: +%.0f m\n\n", rs.elevation))
	b.WriteString(fmt.Sprintf("Avg Power:   %.0f W\n", rs.avgPower))
	b.WriteString(fmt.Sprintf("Avg Cadence: %.0f rpm\n", rs.avgCadence))
	b.WriteString(fmt.Sprintf("Avg Speed:   %.1f km/h\n", rs.avgSpeed))

	return b.String()
}

func (rs *RideScreen) buildStatusView(width, height int) string {
	var b strings.Builder

	gearStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))

	b.WriteString(fmt.Sprintf("Gear:     %s\n", gearStyle.Render(rs.gear)))
	b.WriteString(fmt.Sprintf("Gradient: %+.1f%%\n", rs.gradient))
	b.WriteString(fmt.Sprintf("Mode:     %s\n\n", rs.mode))
	b.WriteString(helpStyle.Render("[↑↓] Shift  [←→] Resistance  [Space] Pause  [q] Quit"))

	return b.String()
}
