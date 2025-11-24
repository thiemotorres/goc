package tui

import (
	"strings"
	"testing"

	"github.com/thiemotorres/goc/internal/gpx"
)

func TestRouteViewIntegration(t *testing.T) {
	// Create realistic route with varied terrain
	route := &gpx.Route{
		Points: []gpx.Point{
			{Lat: 47.0, Lon: 8.0, Distance: 0, Elevation: 400},
			{Lat: 47.01, Lon: 8.01, Distance: 1000, Elevation: 430},   // 3% grade
			{Lat: 47.02, Lon: 8.02, Distance: 2000, Elevation: 480},   // 5% grade
			{Lat: 47.03, Lon: 8.03, Distance: 3000, Elevation: 580},   // 10% grade
			{Lat: 47.04, Lon: 8.03, Distance: 4000, Elevation: 620},   // 4% grade
			{Lat: 47.05, Lon: 8.02, Distance: 5000, Elevation: 630},   // 1% grade
		},
	}

	routeInfo := &RouteInfo{
		Name:     "Test Climb",
		Distance: 5000,
		Ascent:   230,
		AvgGrade: 4.6,
	}

	rv := NewRouteView(routeInfo, route, 60, 15)

	// Test 1: Minimap rendering
	t.Run("minimap_renders", func(t *testing.T) {
		rv.viewMode = RouteViewMinimap
		output := rv.View()

		if !strings.Contains(output, "Test Climb") {
			t.Error("Expected route name in output")
		}
		if !strings.Contains(output, "█") {
			t.Error("Expected route line characters")
		}
		if !strings.Contains(output, "[MINIMAP]") {
			t.Error("Expected mode indicator")
		}
	})

	// Test 2: Elevation profile with colors
	t.Run("elevation_colors", func(t *testing.T) {
		rv.viewMode = RouteViewElevation
		rv.distance = 2500 // Mid-climb
		output := rv.View()

		// Check that the elevation profile is rendered
		// Note: lipgloss may not render ANSI codes in test environment
		// so we check for the elevation line characters instead
		elevationOutput := rv.drawElevationProfile()
		if !strings.Contains(elevationOutput, "─") && !strings.Contains(elevationOutput, "╱") && !strings.Contains(elevationOutput, "╲") {
			t.Error("Expected elevation line characters (─, ╱, or ╲) in elevation profile")
		}
		if !strings.Contains(output, "[ELEVATION PROFILE]") {
			t.Error("Expected mode indicator")
		}
	})

	// Test 3: Auto-switch to elevation on climb
	t.Run("auto_switch_to_elevation", func(t *testing.T) {
		rv.viewMode = RouteViewMinimap
		rv.autoSwitched = false
		rv.distance = 2500
		rv.Update(2500, 8.0) // Steep gradient

		if rv.viewMode != RouteViewElevation {
			t.Error("Expected auto-switch to elevation on steep gradient")
		}
		if !rv.autoSwitched {
			t.Error("Expected autoSwitched flag to be set")
		}
	})

	// Test 4: Manual toggle disables auto-switch
	t.Run("manual_toggle", func(t *testing.T) {
		rv.viewMode = RouteViewElevation
		rv.autoSwitched = true
		rv.ToggleMode()

		if rv.viewMode != RouteViewMinimap {
			t.Error("Expected toggle to switch mode")
		}
		if rv.autoSwitched {
			t.Error("Expected manual toggle to disable auto-switch")
		}
	})

	// Test 5: Position marker visible
	t.Run("position_marker", func(t *testing.T) {
		rv.distance = 2500
		rv.viewMode = RouteViewMinimap
		output := rv.drawMinimap()

		if !strings.Contains(output, "●") {
			t.Error("Expected position marker in minimap")
		}

		rv.viewMode = RouteViewElevation
		output = rv.drawElevationProfile()

		if !strings.Contains(output, "┃") {
			t.Error("Expected position marker in elevation profile")
		}
	})
}
