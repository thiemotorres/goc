# Route Visualization Design

**Date:** 2025-11-24
**Status:** Approved

## Problem Statement

Current ASCII visualizations in the ride screen lack clarity and visual quality:

- **Minimap**: Disconnected dots ('·') make route shape unclear
- **Elevation profile**: Block-filled rendering ('█', '▓') obscures elevation detail
- **Position indicators**: Markers don't stand out sufficiently

Users need better terrain awareness (upcoming climbs, gradient changes) and general orientation (progress tracking) during rides.

## Design Goals

1. **Terrain awareness**: Instant understanding of upcoming gradient zones for pacing
2. **General orientation**: Clear sense of position and progress on route
3. **Visual clarity**: Smooth, readable lines instead of blocky ASCII
4. **Terminal compatibility**: Works across platforms using Unicode (graphics support as future enhancement)
5. **Simplicity**: Strava-style clean design without clutter

## Solution Overview

Upgrade to high-quality Unicode rendering:

- **Minimap**: Continuous route line with clear position marker
- **Elevation profile**: Gradient-color-coded zones with smooth profile line
- **Smart interaction**: Auto-switch to elevation during climbs, manual toggle available
- **Progressive enhancement**: Unicode baseline, graphics rendering as future add-on

## Elevation Profile Design

### Profile Line Rendering

**Current:** Filled area with '█' and '▓' blocks, obscures detail

**New approach:**
- Use Unicode box-drawing for smooth elevation line: `─`, `╱`, `╲`, `│`
- Calculate slope between sample points, choose appropriate character:
  - `│` for vertical changes >45°
  - `╱` or `╲` for moderate slopes
  - `─` for flat sections
- No fill below line - clean and uncluttered
- Alternative: Half-blocks `▀ ▄ █` with double-height rendering for smoother curves

### Gradient Zone Color Coding

Apply background colors based on gradient at each x-position:

- **Green** (0-3%): Flat/easy terrain
- **Yellow** (3-6%): Moderate climbing
- **Orange** (6-10%): Hard climbing
- **Red** (>10%): Very steep climbing

Colors create zones behind elevation line for instant terrain interpretation.

### Position Indicator

- Vertical line using `┃` (thick) or `║` (double)
- Bright contrasting color (white or bright cyan)
- Full height line with marker at elevation intersection
- Clearly stands out against colored zones

### Axes and Labels

- Min/max elevation on left axis
- Distance markers along bottom
- Minimal labels to maximize chart space

## Minimap Design

### Route Line Drawing

**Current:** Disconnected dots ('·')

**New approach:**
- Plot GPS points into 2D grid
- Connect points using line-drawing algorithm (Bresenheim)
- Options for rendering:
  - **Solid blocks**: `█` or `▓` for thick visible line
  - **Box-drawing**: `─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼` for connected paths
- Result: Continuous, visible route line

### Position Marker

- Distinct character: `●` (filled circle) or `◆` (diamond)
- Bright contrasting color (red or cyan)
- Single character, drawn on top of route line
- Immediately visible against route

### Styling

- Clean background, no decoration
- Route line: white or light gray
- Strava-style simplicity
- No grid lines or extra elements

### Aspect Ratio Handling

- Account for terminal character aspect ratio (~2:1 height:width)
- Scale longitude axis by ~2x to prevent distortion
- Ensure route fits available space without stretching

## Auto-Switching and Toggle Behavior

### Auto-Switch Triggers

**To elevation view:**
- Gradient >3% OR
- Climb approaching (>4% gradient, >50m elevation) within 500m

**To minimap view:**
- Gradient <1% for 30 seconds (reduced from 60s for quicker response)

### Manual Toggle

- Tab key toggles between modes
- Manual toggle disables auto-switching until next significant climb
- Prevents fighting user intent

### Visual Feedback

- Current mode in header: `[MINIMAP]` or `[ELEVATION PROFILE]`
- Auto-switched views show `(AUTO)` suffix
- Hint text at bottom: `(Tab to toggle view)`

### Default Behavior

- Start with minimap at ride beginning
- First significant climb triggers auto-switch

## Implementation Approach

### Unicode Character Sets

**Elevation profile:**
- Line drawing: `─ ╱ ╲ │ ╳`
- Position marker: `┃ ║ ▐ ▌`
- Alternative: `▀ ▄ █` half-blocks

**Minimap:**
- Solid blocks: `█ ▓ ▒ ░`
- Box drawing: `─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼`
- Position markers: `● ◆ ◉ ⬤`

### Color Rendering

- Use lipgloss for styling
- Define gradient palette (green/yellow/orange/red)
- Apply background colors to elevation zones
- Foreground colors for lines and markers
- Ensure contrast for readability

### Smoothing and Sampling

- **Elevation**: Even sampling across width, light smoothing to reduce noise
- **Minimap**: Plot all GPS points, apply line-drawing to connect
- Consider Douglas-Peucker algorithm for route simplification if needed

### Layout

- Keep existing panel structure and dimensions
- Elevation and minimap render to same space (toggle)
- Support various terminal sizes
- Minimum viable: ~40 cols x 10 rows

### Testing Considerations

- Various route profiles (flat, hilly, mountain)
- Different terminal themes and color support
- Large GPX file performance
- Auto-switching logic with gradient changes

## Future Enhancements

1. **Graphics support**: Kitty/Sixel protocols for terminals that support them
2. **Route simplification**: Optimize rendering for very long routes
3. **Zoom levels**: Allow zooming in/out on minimap
4. **Segment markers**: Show climb/descent segments on profile
5. **Gradient preview**: Numeric readout of upcoming gradient changes

## Success Criteria

- Route shape clearly visible on minimap
- Elevation changes easy to interpret at a glance
- Position indicator immediately obvious
- Gradient zones provide actionable terrain information
- Works reliably across common terminal emulators
- No performance degradation with typical GPX files (<1000 points)
