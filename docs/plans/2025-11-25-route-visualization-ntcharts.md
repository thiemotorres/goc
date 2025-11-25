# Route Visualization with ntcharts Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace custom minimap and elevation profile rendering with ntcharts-based visualizations for improved visual quality and code simplicity.

**Architecture:** RouteView maintains two chart models (linechart.Model for minimap, timeserieschart.Model for elevation). Charts are initialized once with all route data, then only position markers are updated during ride. Custom grid rendering and gradient coloring logic removed entirely.

**Tech Stack:** ntcharts (linechart, timeserieschart, canvas), lipgloss for styling, existing gpx.Route data structures.

---

## Task 1: Add imports and update struct

**Files:**
- Modify: `internal/tui/routeview.go:1-68`

**Step 1: Add ntcharts imports**

Add these imports after existing imports in `internal/tui/routeview.go`:

```go
"github.com/NimbleMarkets/ntcharts/canvas"
"github.com/NimbleMarkets/ntcharts/linechart"
"github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
```

**Step 2: Update RouteView struct**

Replace the RouteView struct (lines 52-67) with:

```go
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
```

**Step 3: Verify it compiles**

Run: `go build ./...`
Expected: Compilation errors about uninitialized chart fields (we'll fix in next task)

**Step 4: Commit**

```bash
git add internal/tui/routeview.go
git commit -m "refactor: add ntcharts imports and update RouteView struct"
```

---

## Task 2: Implement minimap chart initialization

**Files:**
- Modify: `internal/tui/routeview.go:69-78` (NewRouteView function)

**Step 1: Create helper function for minimap bounds**

Add this function before `NewRouteView`:

```go
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
```

**Step 2: Create helper function to initialize minimap**

Add this function before `NewRouteView`:

```go
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
```

**Step 3: Update NewRouteView to use minimap helper**

Replace `NewRouteView` function (lines 69-78) with:

```go
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
```

**Step 4: Verify it compiles**

Run: `go build ./...`
Expected: Success (minimap chart initialized, elevation still missing)

**Step 5: Commit**

```bash
git add internal/tui/routeview.go
git commit -m "feat: implement minimap chart initialization with braille dots"
```

---

## Task 3: Implement elevation chart initialization

**Files:**
- Modify: `internal/tui/routeview.go` (NewRouteView and add helper)

**Step 1: Create helper function to initialize elevation chart**

Add this function before `NewRouteView`:

```go
// createElevationChart creates and populates the elevation profile chart
func createElevationChart(route *gpx.Route, routeInfo *RouteInfo, width, height int) timeserieslinechart.Model {
	// Styles
	axisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	// Create chart
	chart := timeserieslinechart.New(
		width, height,
		timeserieslinechart.WithStyles(axisStyle, labelStyle, lineStyle),
	)

	// Sample elevation points across route width
	if routeInfo.Distance > 0 {
		for i := 0; i < width; i++ {
			distance := float64(i) / float64(width-1) * routeInfo.Distance
			elevation := route.ElevationAt(distance)
			chart.Push(timeserieslinechart.TimePoint{
				Time:  distance,
				Value: elevation,
			})
		}
	}

	chart.DrawXYAxisAndLabel()
	return chart
}
```

**Step 2: Update NewRouteView to initialize elevation chart**

Update `NewRouteView` to call elevation helper:

```go
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
```

**Step 3: Verify it compiles**

Run: `go build ./...`
Expected: Success

**Step 4: Commit**

```bash
git add internal/tui/routeview.go
git commit -m "feat: implement elevation chart initialization with time series"
```

---

## Task 4: Update View() to use ntcharts

**Files:**
- Modify: `internal/tui/routeview.go:339-388` (View function)

**Step 1: Create helper to draw position marker on minimap**

Add this function before `View()`:

```go
// drawMinimapPosition draws current position marker on minimap
func (rv *RouteView) drawMinimapPosition() {
	if rv.distance > 0 && rv.distance < rv.routeInfo.Distance {
		lat, lon := rv.route.PositionAt(rv.distance)
		point := canvas.Float64Point{X: lon, Y: lat}
		posStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		rv.minimapChart.DrawRuneWithStyle(point, '●', posStyle)
	}
}
```

**Step 2: Create helper to draw position marker on elevation**

Add this function before `View()`:

```go
// drawElevationPosition draws current position marker on elevation chart
func (rv *RouteView) drawElevationPosition() {
	if rv.distance > 0 && rv.distance < rv.routeInfo.Distance {
		elevation := rv.route.ElevationAt(rv.distance)
		point := canvas.Float64Point{X: rv.distance, Y: elevation}
		posStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		rv.elevationChart.DrawRuneWithStyle(point, '●', posStyle)
	}
}
```

**Step 3: Replace View() method**

Replace the entire `View()` method (lines 339-388) with:

```go
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
```

**Step 4: Verify it compiles**

Run: `go build ./...`
Expected: Success

**Step 5: Commit**

```bash
git add internal/tui/routeview.go
git commit -m "feat: update View() to use ntcharts-based rendering"
```

---

## Task 5: Update Resize to recreate charts

**Files:**
- Modify: `internal/tui/routeview.go:333-337` (Resize function)

**Step 1: Replace Resize() method**

Replace the existing `Resize()` method (lines 333-337) with:

```go
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
```

**Step 2: Verify it compiles**

Run: `go build ./...`
Expected: Success

**Step 3: Commit**

```bash
git add internal/tui/routeview.go
git commit -m "feat: update Resize to recreate charts with new dimensions"
```

---

## Task 6: Remove old custom rendering code

**Files:**
- Modify: `internal/tui/routeview.go`

**Step 1: Remove gradient color constants and functions**

Delete lines 11-32 (gradient color styles and gradientColorStyle function):
- `gradientFlat`, `gradientMod`, `gradientHard`, `gradientSteep`, `gradientDesc`
- `gradientColorStyle()` function

**Step 2: Remove slopeCharacter function**

Delete lines 34-42 (slopeCharacter function)

**Step 3: Remove drawMinimap function**

Delete the entire `drawMinimap()` method (lines 80-168)

**Step 4: Remove drawElevationProfile function**

Delete the entire `drawElevationProfile()` method (lines 170-266)

**Step 5: Remove drawLine helper function**

Delete the entire `drawLine()` function (lines 390-426)

**Step 6: Remove abs helper function**

Delete the entire `abs()` function (lines 428-434)

**Step 7: Verify it compiles**

Run: `go build ./...`
Expected: Success

**Step 8: Commit**

```bash
git add internal/tui/routeview.go
git commit -m "refactor: remove custom grid rendering and gradient coloring code"
```

---

## Task 7: Update tests for ntcharts-based rendering

**Files:**
- Modify: `internal/tui/routeview_test.go`

**Step 1: Remove obsolete tests**

Delete these test functions (they test removed code):
- `TestGradientColor` (lines 12-38)
- `TestDrawElevationProfileWithColors` (lines 40-76)
- `TestSlopeCharacter` (lines 78-98)
- `TestPositionMarkerInElevationProfile` (lines 100-146)
- `TestDrawLine` (lines 148-185)
- `TestDrawMinimapConnected` (lines 187-220)

**Step 2: Add test for minimap chart creation**

Add this test at the end of the file:

```go
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
```

**Step 3: Add test for elevation chart creation**

Add this test:

```go
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
```

**Step 4: Add test for chart resize**

Add this test:

```go
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
```

**Step 5: Run tests**

Run: `go test ./internal/tui -v -run TestRouteView`
Expected: All RouteView tests pass

**Step 6: Commit**

```bash
git add internal/tui/routeview_test.go
git commit -m "test: update tests for ntcharts-based rendering"
```

---

## Task 8: Run full test suite and verify

**Files:**
- N/A (verification step)

**Step 1: Run all tests**

Run: `go test ./...`
Expected: All tests pass

**Step 2: Build binary**

Run: `go build -o goc .`
Expected: Success

**Step 3: Verify integration test still passes**

Run: `go test ./internal/tui -v -run TestRouteViewIntegration`
Expected: Integration test passes

**Step 4: Manual verification checklist**

If you have a test route GPX file, run the application and verify:
- [ ] Minimap displays with braille dots showing route shape
- [ ] Position marker (●) appears on minimap
- [ ] Elevation profile displays clean line chart
- [ ] Position marker (●) appears on elevation profile
- [ ] Tab key toggles between modes
- [ ] Auto-switch works during climbs (if applicable)
- [ ] Resize works (change terminal size)

**Step 5: Final commit**

```bash
git add -A
git commit -m "feat: complete route visualization ntcharts migration

Replace custom minimap and elevation rendering with ntcharts-based
visualizations. Minimap uses linechart with braille dots for higher
resolution. Elevation uses time series chart for cleaner display.
Removes gradient coloring and custom grid rendering logic."
```

---

## Testing Strategy

**Unit tests:**
- Chart creation and initialization
- Chart resize behavior
- View rendering with different modes
- Auto-switch timing (existing test kept)

**Integration tests:**
- Existing `TestRouteViewIntegration` verifies end-to-end behavior

**Manual testing:**
- Visual quality of minimap with real GPX data
- Visual quality of elevation profile
- Position marker accuracy
- Mode switching (manual and auto)
- Terminal resize handling

## Rollback Plan

If issues are discovered:
1. This work is in isolated worktree on feature branch
2. Main branch unchanged - can abandon worktree
3. All changes in single PR - easy to revert if merged

## Dependencies

- ntcharts v0.3.1 already in go.mod
- No new external dependencies required

## Non-Goals

- Changing auto-switch logic (keep existing behavior)
- Adding zoom/pan (future enhancement)
- Mouse interaction (future enhancement)
- Multiple route comparison (out of scope)
