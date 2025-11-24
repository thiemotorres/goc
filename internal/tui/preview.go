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
	if rp.route != nil {
		b.WriteString(fmt.Sprintf("Elevation:   %.0fm ↑  %.0fm ↓\n", rp.info.Ascent, rp.route.TotalDescent))
	} else {
		b.WriteString(fmt.Sprintf("Elevation:   %.0fm ↑\n", rp.info.Ascent))
	}
	b.WriteString(fmt.Sprintf("Avg Grade:   %.1f%%\n", rp.info.AvgGrade))

	if rp.route != nil {
		maxGrade := rp.findMaxGrade()
		minEle, maxEle := rp.findElevationRange()
		b.WriteString(fmt.Sprintf("Max Grade:   %.1f%%\n", maxGrade))
		b.WriteString(fmt.Sprintf("Elev Range:  %.0fm - %.0fm\n", minEle, maxEle))
	}

	b.WriteString("\n")

	// Elevation profile
	if rp.route != nil {
		b.WriteString("Elevation Profile:\n")
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

func (rp *RoutePreview) findElevationRange() (min, max float64) {
	if rp.route == nil || len(rp.route.Points) == 0 {
		return 0, 0
	}

	min = rp.route.Points[0].Elevation
	max = rp.route.Points[0].Elevation

	for _, pt := range rp.route.Points {
		if pt.Elevation < min {
			min = pt.Elevation
		}
		if pt.Elevation > max {
			max = pt.Elevation
		}
	}
	return min, max
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
