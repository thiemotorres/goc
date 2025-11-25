package tui

import (
	"fmt"
	"strings"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiemotorres/goc/internal/gpx"
)

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
	elevationChart linechart.Model

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

// createElevationChart creates and populates the elevation profile chart
func createElevationChart(route *gpx.Route, routeInfo *RouteInfo, width, height int) linechart.Model {
	// Find min/max elevation for Y axis
	minEle, maxEle := route.Points[0].Elevation, route.Points[0].Elevation
	for _, p := range route.Points {
		if p.Elevation < minEle {
			minEle = p.Elevation
		}
		if p.Elevation > maxEle {
			maxEle = p.Elevation
		}
	}

	// Add some padding to Y axis
	eleRange := maxEle - minEle
	if eleRange == 0 {
		eleRange = 10
	}
	padding := eleRange * 0.1
	minEle -= padding
	maxEle += padding

	// Styles
	axisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	// Create chart with distance on X axis and elevation on Y axis
	chart := linechart.New(
		width, height,
		0, routeInfo.Distance,
		minEle, maxEle,
		linechart.WithStyles(axisStyle, labelStyle, lineStyle),
	)

	// Draw elevation profile as connected line segments
	if routeInfo.Distance > 0 {
		var prevPoint canvas.Float64Point
		for i := 0; i < width; i++ {
			distance := float64(i) / float64(width-1) * routeInfo.Distance
			elevation := route.ElevationAt(distance)
			point := canvas.Float64Point{X: distance, Y: elevation}

			if i > 0 {
				chart.DrawBrailleLine(prevPoint, point)
			}
			prevPoint = point
		}
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
		rv.elevationChart = createElevationChart(route, routeInfo, width, height)
	}

	return rv
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

// drawMinimapPosition draws current position marker on minimap
func (rv *RouteView) drawMinimapPosition() {
	if rv.distance > 0 && rv.distance < rv.routeInfo.Distance {
		lat, lon := rv.route.PositionAt(rv.distance)
		point := canvas.Float64Point{X: lon, Y: lat}
		posStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		rv.minimapChart.DrawRuneWithStyle(point, '●', posStyle)
	}
}

// drawElevationPosition draws current position marker on elevation chart
func (rv *RouteView) drawElevationPosition() {
	if rv.distance > 0 && rv.distance < rv.routeInfo.Distance {
		elevation := rv.route.ElevationAt(rv.distance)
		point := canvas.Float64Point{X: rv.distance, Y: elevation}
		posStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		rv.elevationChart.DrawRuneWithStyle(point, '●', posStyle)
	}
}

// Resize resizes the view and recreates charts
func (rv *RouteView) Resize(width, height int) {
	rv.width = width
	rv.height = height

	// Recreate charts with new dimensions
	if rv.route != nil && len(rv.route.Points) > 0 {
		rv.minimapChart = createMinimapChart(rv.route, width, height)
		rv.elevationChart = createElevationChart(rv.route, rv.routeInfo, width, height)
	}
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

	// Render appropriate chart with position marker
	if rv.viewMode == RouteViewMinimap {
		rv.drawMinimapPosition()
		b.WriteString(rv.minimapChart.View())
	} else {
		rv.drawElevationPosition()
		b.WriteString(rv.elevationChart.View())
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(Tab to toggle view)"))

	return b.String()
}
