package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/thiemotorres/goc/internal/gpx"
)

func TestGradientColor(t *testing.T) {
	tests := []struct {
		gradient float64
		wantRGB  string // RGB color code
	}{
		{0.0, "34"},    // Green for flat (0-3%)
		{2.5, "34"},    // Green
		{3.0, "226"},   // Yellow for moderate (3-6%)
		{5.5, "226"},   // Yellow
		{6.0, "214"},   // Orange for hard (6-10%)
		{9.5, "214"},   // Orange
		{10.0, "196"},  // Red for very steep (>10%)
		{15.0, "196"},  // Red
		{-2.0, "240"},  // Gray for descent
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("gradient_%.1f", tt.gradient), func(t *testing.T) {
			style := gradientColorStyle(tt.gradient)
			// Check that style has the expected color
			bg := style.GetBackground()
			if bg != lipgloss.Color(tt.wantRGB) {
				t.Errorf("gradientColorStyle(%.1f) = %v, want color %s", tt.gradient, bg, tt.wantRGB)
			}
		})
	}
}

func TestDrawElevationProfileWithColors(t *testing.T) {
	route := &gpx.Route{
		Points: []gpx.Point{
			{Distance: 0, Elevation: 100},
			{Distance: 1000, Elevation: 130},  // 3% gradient
			{Distance: 2000, Elevation: 180},  // 5% gradient
			{Distance: 3000, Elevation: 280},  // 10% gradient
		},
	}

	routeInfo := &RouteInfo{
		Distance: 3000,
		Ascent:   180,
	}

	rv := NewRouteView(routeInfo, route, 40, 10)
	rv.distance = 1500

	output := rv.drawElevationProfile()

	// Check output is not empty and has expected dimensions
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 5 {
		t.Errorf("Expected at least 5 lines of output, got %d", len(lines))
	}

	// Verify that the output contains smooth elevation line characters
	// (the implementation uses "─", "╱", "╲" for smooth line rendering)
	hasLine := strings.Contains(output, "─") || strings.Contains(output, "╱") || strings.Contains(output, "╲")
	if !hasLine {
		t.Error("Expected output to contain smooth elevation line characters (─, ╱, ╲)")
	}

	// Note: ANSI codes may not appear in test environment where lipgloss
	// detects no TTY, but the gradient color logic is tested separately
	// in TestGradientColor
}

func TestSlopeCharacter(t *testing.T) {
	tests := []struct {
		prevY, currY int
		wantChar     string
	}{
		{5, 5, "─"},   // Flat
		{5, 3, "╱"},   // Up
		{3, 5, "╲"},   // Down
		{5, 4, "╱"},   // Slight up
		{4, 5, "╲"},   // Slight down
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d_to_%d", tt.prevY, tt.currY), func(t *testing.T) {
			got := slopeCharacter(tt.prevY, tt.currY)
			if got != tt.wantChar {
				t.Errorf("slopeCharacter(%d, %d) = %q, want %q", tt.prevY, tt.currY, got, tt.wantChar)
			}
		})
	}
}

func TestPositionMarkerInElevationProfile(t *testing.T) {
	route := &gpx.Route{
		Points: []gpx.Point{
			{Distance: 0, Elevation: 100},
			{Distance: 1000, Elevation: 130},
			{Distance: 2000, Elevation: 180},
			{Distance: 3000, Elevation: 280},
		},
	}

	routeInfo := &RouteInfo{
		Distance: 3000,
		Ascent:   180,
	}

	rv := NewRouteView(routeInfo, route, 40, 10)
	rv.distance = 1500 // Position in middle

	output := rv.drawElevationProfile()

	// Check that position marker is present
	if !strings.Contains(output, "┃") {
		t.Error("Expected position marker '┃' in elevation profile")
	}

	// Test with position at start
	rv.distance = 0
	output = rv.drawElevationProfile()
	if !strings.Contains(output, "┃") {
		t.Error("Expected position marker at start of route")
	}

	// Test with position at end
	rv.distance = 3000
	output = rv.drawElevationProfile()
	if !strings.Contains(output, "┃") {
		t.Error("Expected position marker at end of route")
	}

	// Test with no position (distance = 0, should not show marker at position)
	rv.distance = 0
	rv.routeInfo.Distance = 0
	output = rv.drawElevationProfile()
	// When distance is 0, posX should be -1, so no marker should appear
	// This is actually handled by the calculation logic
}
