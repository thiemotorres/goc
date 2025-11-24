package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
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
