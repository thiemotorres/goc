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
