# Route Visualization with ntcharts

**Date:** 2025-11-25
**Status:** Approved

## Problem

The current custom minimap and elevation profile visualizations in the ride screen are too crude and may confuse users. They use basic ASCII blocks and custom rendering logic that lacks the polish of the ntcharts streamline charts used elsewhere in the UI.

## Solution

Replace custom rendering with ntcharts-based visualizations for both minimap and elevation profile modes, creating a consistent, professional look across all ride screen charts.

## Design

### Overall Approach

RouteView maintains two modes (minimap/elevation) but both use ntcharts:

1. **Minimap mode:** `linechart.Model`
   - Bounding box from GPX lat/lon (10% padding for margins)
   - All route points rendered as braille dots via `DrawBrailleLine(pt, pt)`
   - Current position as bright colored rune via `DrawRuneWithStyle(pos, '●', style)`

2. **Elevation mode:** `timeserieschart.Model`
   - X-axis: distance (0 to total route distance in km)
   - Y-axis: elevation (meters)
   - Line chart showing elevation profile
   - Current position as highlighted point via `DrawRuneWithStyle(pos, '●', style)`

**Key benefits:**
- Consistent professional look across all visualizations
- ntcharts handles scaling, axes, coordinate transforms
- Simpler code - removes custom grid/rendering logic
- Braille dots = 4x resolution improvement for minimap

### Minimap Implementation

**Data preparation:**
- Extract all GPX points' lat/lon coordinates
- Calculate bounding box: `minLat, maxLat, minLon, maxLon`
- Add 10% padding to bounds (prevent route from touching edges)
- Account for aspect ratio in bounds calculation (terminal chars are ~2:1)

**Chart creation:**
```go
minimapChart := linechart.New(
    width, height,
    minLon, maxLon,
    minLat, maxLat,
    linechart.WithStyles(axisStyle, labelStyle, routeStyle),
)
```

**Rendering route:**
- Loop through all GPX points
- For each point: `minimapChart.DrawBrailleLine(point, point)` (same point twice = single braille dot)
- Braille gives 2×4 sub-pixels per character = ~8x higher resolution than current `█` blocks

**Rendering current position:**
- Calculate current lat/lon from `rv.distance` using existing `rv.route.PositionAt(distance)`
- Draw highlighted marker: `minimapChart.DrawRuneWithStyle(currentPos, '●', positionStyle)`
- Position style: bright color (red/magenta), bold

### Elevation Profile Implementation

**Data preparation:**
- Sample elevation points across the route
- Create arrays of distance and elevation values
- Find min/max elevation for Y-axis bounds (add small padding for readability)

**Chart creation:**
```go
elevationChart := timeserieschart.New(
    width, height,
    timeserieschart.WithStyles(axisStyle, labelStyle, lineStyle),
)
```

**Rendering elevation line:**
- Push elevation data points to the chart
- Time Series Chart handles line drawing automatically
- X-axis shows distance progression (0 to total distance)
- Y-axis shows elevation range

**Rendering current position:**
- Calculate current elevation: `rv.route.ElevationAt(rv.distance)`
- Create point at (current distance, current elevation)
- Draw highlighted marker: `elevationChart.DrawRuneWithStyle(currentPos, '●', positionStyle)`
- Same bright style as minimap for consistency

**Gradient coloring removed:**
- No background colors
- Clean line chart only
- Line shape communicates terrain

### Chart Lifecycle & State Management

**RouteView struct changes:**
```go
type RouteView struct {
    route        *gpx.Route
    routeInfo    *RouteInfo
    distance     float64
    gradient     float64
    viewMode     RouteViewMode
    autoSwitched bool

    // Replace custom rendering with ntcharts models
    minimapChart    linechart.Model      // new
    elevationChart  timeserieschart.Model // new

    width  int
    height int
    climbTime float64
}
```

**Initialization (NewRouteView):**
- Create both charts immediately with initial dimensions
- Pre-populate minimap with all route braille dots (static, doesn't change)
- Pre-populate elevation chart with elevation line (static, doesn't change)
- Both charts ready to render, just need position marker updates

**Update cycle (rv.Update called ~10x/sec):**
- Charts already have route data drawn
- Only redraw position markers when distance changes
- Clear old position marker, draw new one at updated location
- Keep existing auto-switch logic (works well based on recent commits)

**Resize handling (rv.Resize):**
- Both charts need to be recreated with new dimensions
- Redraw all route points/elevation data
- Redraw position marker

### View Rendering & Styling

**View() method changes:**
- Remove `drawMinimap()` and `drawElevationProfile()` functions entirely
- Replace with simpler logic:
  ```go
  if rv.viewMode == RouteViewMinimap {
      visualization = rv.minimapChart.View()
  } else {
      visualization = rv.elevationChart.View()
  }
  ```
- Keep existing header (route name, mode indicator, progress info)
- Keep existing footer (Tab to toggle hint)

**Styling consistency:**
- **Axis/labels:** Subtle gray (color "240") to match current UI
- **Route line/dots:** White or light gray (color "255")
- **Position marker:** Bright red/magenta (color "196" or "212") with bold - matches current position style
- All styles use lipgloss, consistent with rest of TUI

**Chart axes configuration:**
- Minimap: Could hide axes/labels (just show route shape) OR show lat/lon
- Elevation: Show distance on X-axis (km), elevation on Y-axis (m)
- Both: Minimal axis labels to maximize chart space

**Error handling:**
- Keep fallback text if route/GPX fails to load
- Handle empty points array gracefully
- Validate bounds are non-zero before creating charts

## Code Changes

### Files to modify:
- `internal/tui/routeview.go` - Main implementation changes

### Files to remove:
- Custom rendering functions: `drawMinimap()`, `drawElevationProfile()`
- Helper functions: `drawLine()`, `slopeCharacter()`, `gradientColorStyle()`
- Gradient color styles constants

### Dependencies:
- Already using `github.com/NimbleMarkets/ntcharts/linechart/streamlinechart`
- Add: `github.com/NimbleMarkets/ntcharts/linechart`
- Add: `github.com/NimbleMarkets/ntcharts/timeserieschart`

## Testing

- Visual inspection of minimap with braille route rendering
- Visual inspection of elevation profile with time series chart
- Position marker accuracy in both modes
- Resize handling (terminal window changes)
- Auto-switch behavior between modes
- Manual toggle between modes (Tab key)
- Error cases (missing route, empty GPX, invalid bounds)

## Non-Goals

- Changing auto-switch logic or timing
- Adding zoom/pan functionality
- Interactive chart features (mouse support)
- Multiple route comparison
- Gradient background coloring (explicitly removed for cleaner look)
