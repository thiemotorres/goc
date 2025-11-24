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

	// Verify that the output contains elevation line characters
	// (the implementation uses colored rendering with "─" for the elevation line)
	if !strings.Contains(output, "─") && !strings.Contains(output, "▓") {
		t.Error("Expected output to contain elevation line or fill characters")
	}

	// Note: ANSI codes may not appear in test environment where lipgloss
	// detects no TTY, but the gradient color logic is tested separately
	// in TestGradientColor
}
