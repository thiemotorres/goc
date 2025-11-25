package tui

import (
	"fmt"
	"strings"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
	"github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiemotorres/goc/internal/gpx"
)

// Gradient color styles for elevation profile
var (
	gradientFlat  = lipgloss.NewStyle().Background(lipgloss.Color("34"))   // Green: 0-3%
	gradientMod   = lipgloss.NewStyle().Background(lipgloss.Color("226"))  // Yellow: 3-6%
	gradientHard  = lipgloss.NewStyle().Background(lipgloss.Color("214"))  // Orange: 6-10%
	gradientSteep = lipgloss.NewStyle().Background(lipgloss.Color("196"))  // Red: >10%
	gradientDesc  = lipgloss.NewStyle().Background(lipgloss.Color("240"))  // Gray: descent
)

// gradientColorStyle returns lipgloss style for given gradient percentage
func gradientColorStyle(gradient float64) lipgloss.Style {
	if gradient < 0 {
		return gradientDesc
	} else if gradient < 3.0 {
		return gradientFlat
	} else if gradient < 6.0 {
		return gradientMod
	} else if gradient < 10.0 {
		return gradientHard
	}
	return gradientSteep
}

// slopeCharacter returns appropriate Unicode character for elevation slope
func slopeCharacter(prevY, currY int) string {
	if currY == prevY {
		return "─" // Flat
	} else if currY < prevY {
		return "╱" // Going up (Y decreases as elevation increases)
	}
	return "╲" // Going down
}

// RouteViewMode represents the current view mode
type RouteViewMode int

const (
	RouteViewMinimap RouteViewMode = iota
	RouteViewElevation
)

// RouteView displays route information with minimap or elevation profile
type RouteView struct {
	route        *gpx.Route
	routeInfo    *RouteInfo
	distance     float64 // current position in meters
	gradient     float64 // current gradient
	viewMode     RouteViewMode
	autoSwitched bool

	// Charts (ntcharts-based)
	minimapChart   linechart.Model
	elevationChart timeserieslinechart.Model

	// Dimensions
	width  int
	height int

	// Auto-switch state
	climbTime float64 // time spent in climb mode
}

// calculateMinimapBounds calculates lat/lon bounds with padding
func calculateMinimapBounds(points []gpx.Point) (minLat, maxLat, minLon, maxLon float64) {
	if len(points) == 0 {
		return 0, 1, 0, 1
	}

	minLat, maxLat = points[0].Lat, points[0].Lat
	minLon, maxLon = points[0].Lon, points[0].Lon

	for _, p := range points {
		if p.Lat < minLat {
			minLat = p.Lat
		}
		if p.Lat > maxLat {
			maxLat = p.Lat
		}
		if p.Lon < minLon {
			minLon = p.Lon
		}
		if p.Lon > maxLon {
			maxLon = p.Lon
		}
	}

	// Add 10% padding
	latRange := maxLat - minLat
	lonRange := maxLon - minLon
	if latRange == 0 {
		latRange = 0.001
	}
	if lonRange == 0 {
		lonRange = 0.001
	}

	padding := 0.1
	minLat -= latRange * padding
	maxLat += latRange * padding
	minLon -= lonRange * padding
	maxLon += lonRange * padding

	return minLat, maxLat, minLon, maxLon
}

// createMinimapChart creates and populates the minimap chart
func createMinimapChart(route *gpx.Route, width, height int) linechart.Model {
	minLat, maxLat, minLon, maxLon := calculateMinimapBounds(route.Points)

	// Styles
	axisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	routeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	// Create chart
	chart := linechart.New(
		width, height,
		minLon, maxLon,
		minLat, maxLat,
		linechart.WithStyles(axisStyle, labelStyle, routeStyle),
	)

	// Draw all route points as braille dots
	for _, pt := range route.Points {
		point := canvas.Float64Point{X: pt.Lon, Y: pt.Lat}
		chart.DrawBrailleLine(point, point)
	}

	chart.DrawXYAxisAndLabel()
	return chart
}

// NewRouteView creates a new route view
func NewRouteView(routeInfo *RouteInfo, route *gpx.Route, width, height int) *RouteView {
	rv := &RouteView{
		route:     route,
		routeInfo: routeInfo,
		viewMode:  RouteViewMinimap,
		width:     width,
		height:    height,
	}

	if route != nil && len(route.Points) > 0 {
		rv.minimapChart = createMinimapChart(route, width, height)
	}

	return rv
}

func (rv *RouteView) drawMinimap() string {
	points := rv.route.Points
	if len(points) == 0 {
		return "No route data"
	}

	// Find lat/lon bounds
	minLat, maxLat := points[0].Lat, points[0].Lat
	minLon, maxLon := points[0].Lon, points[0].Lon

	for _, p := range points {
		if p.Lat < minLat {
			minLat = p.Lat
		}
		if p.Lat > maxLat {
			maxLat = p.Lat
		}
		if p.Lon < minLon {
			minLon = p.Lon
		}
		if p.Lon > maxLon {
			maxLon = p.Lon
		}
	}

	latRange := maxLat - minLat
	lonRange := maxLon - minLon
	if latRange == 0 {
		latRange = 1
	}
	if lonRange == 0 {
		lonRange = 1
	}

	w, h := rv.width, rv.height
	if w <= 0 {
		w = 40
	}
	if h <= 0 {
		h = 10
	}

	// Account for terminal character aspect ratio (~2:1)
	aspectRatio := 2.0

	// Create grid for route line
	grid := make([][]bool, h)
	for i := range grid {
		grid[i] = make([]bool, w)
	}

	// Draw connected route path using line algorithm
	for i := 1; i < len(points); i++ {
		prevX := int((points[i-1].Lon - minLon) / lonRange * float64(w-1) / aspectRatio)
		prevY := int((maxLat - points[i-1].Lat) / latRange * float64(h-1))
		currX := int((points[i].Lon - minLon) / lonRange * float64(w-1) / aspectRatio)
		currY := int((maxLat - points[i].Lat) / latRange * float64(h-1))

		drawLine(grid, prevX, prevY, currX, currY)
	}

	// Calculate current position marker location
	posX, posY := -1, -1
	if rv.distance > 0 && rv.distance < rv.routeInfo.Distance {
		lat, lon := rv.route.PositionAt(rv.distance)
		posX = int((lon - minLon) / lonRange * float64(w-1) / aspectRatio)
		posY = int((maxLat - lat) / latRange * float64(h-1))
	}

	// Render grid to string
	routeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	posStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	var b strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x == posX && y == posY {
				b.WriteString(posStyle.Render("●"))
			} else if grid[y][x] {
				b.WriteString(routeStyle.Render("█"))
			} else {
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (rv *RouteView) drawElevationProfile() string {
	if rv.route == nil {
		return "No route data"
	}

	points := rv.route.Points
	if len(points) == 0 {
		return "No route data"
	}

	w, h := rv.width, rv.height
	if w <= 0 {
		w = 60
	}
	if h <= 0 {
		h = 15
	}

	// Sample elevations and gradients evenly across width
	sampledElevations := make([]float64, w)
	sampledGradients := make([]float64, w)
	for i := 0; i < w; i++ {
		distance := float64(i) / float64(w-1) * rv.routeInfo.Distance
		sampledElevations[i] = rv.route.ElevationAt(distance)
		sampledGradients[i] = rv.route.GradientAt(distance)
	}

	// Find min/max for scaling
	minEle, maxEle := sampledElevations[0], sampledElevations[0]
	for _, e := range sampledElevations {
		if e < minEle {
			minEle = e
		}
		if e > maxEle {
			maxEle = e
		}
	}

	eleRange := maxEle - minEle
	if eleRange == 0 {
		eleRange = 1
	}

	// Calculate position marker x coordinate
	posX := -1
	if rv.routeInfo.Distance > 0 {
		posX = int((rv.distance / rv.routeInfo.Distance) * float64(w-1))
		if posX < 0 || posX >= w {
			posX = -1
		}
	}

	// Marker style
	markerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true)

	// Pre-calculate elevation Y positions for smooth line rendering
	eleYPositions := make([]int, w)
	for x := 0; x < w; x++ {
		eleYPositions[x] = int((maxEle - sampledElevations[x]) / eleRange * float64(h-1))
	}

	// Build output line by line with colored backgrounds
	var b strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			eleY := eleYPositions[x]

			// Determine character
			var char string
			if x == posX {
				// Render position marker
				char = "┃"
				b.WriteString(markerStyle.Render(char))
			} else {
				if y == eleY {
					// Determine slope from previous point
					if x > 0 {
						prevEleY := eleYPositions[x-1]
						char = slopeCharacter(prevEleY, eleY)
					} else {
						char = "─"
					}
				} else {
					char = " " // Changed from "▓" - no fill
				}
				// Apply gradient color as background
				style := gradientColorStyle(sampledGradients[x])
				b.WriteString(style.Render(char))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// Update updates the route view with current position and gradient
func (rv *RouteView) Update(distance, gradient float64) {
	rv.distance = distance
	rv.gradient = gradient

	// Auto-switch logic
	if rv.shouldAutoSwitch() {
		rv.autoSwitchMode()
	}
}

func (rv *RouteView) shouldAutoSwitch() bool {
	// Switch to elevation when climbing
	if rv.viewMode == RouteViewMinimap {
		// Check if entering a climb (gradient > 3%)
		if rv.gradient > 3.0 {
			return true
		}

		// Check if climb is approaching (using built-in detection)
		if rv.route != nil && rv.distance < rv.routeInfo.Distance-500 {
			approaching, _ := rv.route.IsClimbApproaching(rv.distance, 500, 4.0, 50)
			if approaching {
				return true
			}
		}
	}

	// Switch back to minimap when climb is done
	if rv.viewMode == RouteViewElevation && rv.autoSwitched {
		// Check if gradient is low for sustained period
		if rv.gradient < 1.0 {
			rv.climbTime += 0.1 // Assuming ~10 updates per second
			if rv.climbTime > 30 { // 30 seconds of flat terrain
				rv.climbTime = 0
				return true
			}
		} else {
			rv.climbTime = 0
		}
	}

	return false
}

func (rv *RouteView) autoSwitchMode() {
	if rv.viewMode == RouteViewMinimap {
		rv.viewMode = RouteViewElevation
		rv.autoSwitched = true
	} else {
		rv.viewMode = RouteViewMinimap
		rv.autoSwitched = false
	}
}

// ToggleMode manually toggles between minimap and elevation profile
func (rv *RouteView) ToggleMode() {
	if rv.viewMode == RouteViewMinimap {
		rv.viewMode = RouteViewElevation
	} else {
		rv.viewMode = RouteViewMinimap
	}
	rv.autoSwitched = false // Manual toggle disables auto-switch
}

// Resize resizes the view
func (rv *RouteView) Resize(width, height int) {
	rv.width = width
	rv.height = height
}

// View renders the route view
func (rv *RouteView) View() string {
	if rv.route == nil {
		return "Free Ride\n\nNo route selected"
	}

	var b strings.Builder

	// Header with mode indicator
	modeIndicator := ""
	if rv.viewMode == RouteViewMinimap {
		modeIndicator = "[MINIMAP]"
	} else {
		modeIndicator = "[ELEVATION PROFILE]"
		if rv.autoSwitched {
			modeIndicator += " (AUTO)"
		}
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	b.WriteString(headerStyle.Render(rv.routeInfo.Name))
	b.WriteString("  ")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(modeIndicator))
	b.WriteString("\n\n")

	// Route info
	progress := 0.0
	if rv.routeInfo.Distance > 0 {
		progress = (rv.distance / rv.routeInfo.Distance) * 100
	}

	b.WriteString(fmt.Sprintf("Distance: %.1f / %.1f km (%.0f%%)\n",
		rv.distance/1000,
		rv.routeInfo.Distance/1000,
		progress))
	b.WriteString(fmt.Sprintf("Elevation: +%.0fm  Avg Grade: %.1f%%\n\n",
		rv.routeInfo.Ascent,
		rv.routeInfo.AvgGrade))

	// Render appropriate view
	if rv.viewMode == RouteViewMinimap {
		b.WriteString(rv.drawMinimap())
	} else {
		b.WriteString(rv.drawElevationProfile())
	}

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(Tab to toggle view)"))

	return b.String()
}

// drawLine uses Bresenham's algorithm to draw a line on a 2D grid
func drawLine(grid [][]bool, x0, y0, x1, y1 int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}

	err := dx - dy
	x, y := x0, y0

	for {
		// Mark point if in bounds
		if y >= 0 && y < len(grid) && x >= 0 && x < len(grid[0]) {
			grid[y][x] = true
		}

		if x == x1 && y == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// abs returns absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
