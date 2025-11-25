package tui

import (
	"strings"
	"testing"

	"github.com/thiemotorres/goc/internal/gpx"
)

func TestMinimapChartCreation(t *testing.T) {
	route := &gpx.Route{
		Points: []gpx.Point{
			{Lat: 0, Lon: 0, Distance: 0},
			{Lat: 0.01, Lon: 0.01, Distance: 1000},
			{Lat: 0.02, Lon: 0.01, Distance: 2000},
		},
	}

	routeInfo := &RouteInfo{
		Distance: 2000,
	}

	rv := NewRouteView(routeInfo, route, 40, 10)

	// Verify minimap chart was created
	output := rv.View()
	if len(output) == 0 {
		t.Error("Expected non-empty view output")
	}

	// Verify it contains route name and mode indicator
	if !strings.Contains(output, "[MINIMAP]") {
		t.Error("Expected view to contain [MINIMAP] mode indicator")
	}
}

func TestElevationChartCreation(t *testing.T) {
	route := &gpx.Route{
		Points: []gpx.Point{
			{Distance: 0, Elevation: 100},
			{Distance: 1000, Elevation: 150},
			{Distance: 2000, Elevation: 200},
		},
	}

	routeInfo := &RouteInfo{
		Distance: 2000,
		Ascent:   100,
	}

	rv := NewRouteView(routeInfo, route, 40, 10)
	rv.viewMode = RouteViewElevation

	// Verify elevation chart was created
	output := rv.View()
	if len(output) == 0 {
		t.Error("Expected non-empty view output")
	}

	// Verify it contains elevation mode indicator
	if !strings.Contains(output, "[ELEVATION PROFILE]") {
		t.Error("Expected view to contain [ELEVATION PROFILE] mode indicator")
	}
}

func TestChartResize(t *testing.T) {
	route := &gpx.Route{
		Points: []gpx.Point{
			{Lat: 0, Lon: 0, Distance: 0, Elevation: 100},
			{Lat: 0.01, Lon: 0.01, Distance: 1000, Elevation: 150},
		},
	}

	routeInfo := &RouteInfo{
		Distance: 1000,
	}

	rv := NewRouteView(routeInfo, route, 40, 10)

	// Resize
	rv.Resize(80, 20)

	// Verify dimensions updated
	if rv.width != 80 || rv.height != 20 {
		t.Errorf("Expected dimensions 80x20, got %dx%d", rv.width, rv.height)
	}

	// Verify charts still work after resize
	output := rv.View()
	if len(output) == 0 {
		t.Error("Expected non-empty view output after resize")
	}
}

func TestAutoSwitchTiming(t *testing.T) {
	route := &gpx.Route{
		Points: []gpx.Point{
			{Distance: 0, Elevation: 100},
			{Distance: 1000, Elevation: 150},
		},
	}

	routeInfo := &RouteInfo{
		Distance: 1000,
	}

	rv := NewRouteView(routeInfo, route, 40, 10)
	rv.viewMode = RouteViewElevation
	rv.autoSwitched = true

	// Simulate 30 seconds of flat terrain (300 updates at 0.1s each)
	for i := 0; i < 300; i++ {
		rv.Update(500, 0.5) // Low gradient
	}

	// After 30s of flat, should auto-switch back to minimap
	if rv.viewMode != RouteViewMinimap {
		t.Error("Expected auto-switch to minimap after 30s of flat terrain")
	}
}
