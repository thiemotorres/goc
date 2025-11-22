# goc - Indoor Cycling Trainer TUI

A terminal-based indoor cycling application for the Elite Rivo trainer with Swift Click V2 virtual shifting.

## Overview

**goc** (cog backwards, "go" prefix) is a TUI application for indoor cycling training that provides:
- Real-time metrics display (power, cadence, speed)
- GPX route simulation with gradient-based resistance
- Virtual shifting with realistic gear ratios
- ERG mode for fixed power training
- Local ride history with FIT export

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                      goc                            │
├─────────────┬─────────────┬─────────────────────────┤
│ Bluetooth   │ Simulation  │ TUI                     │
│ Manager     │ Engine      │ Renderer                │
├─────────────┼─────────────┼─────────────────────────┤
│ - FTMS      │ - Gradient  │ - termdash layout       │
│ - Swift     │ - Gears     │ - Rolling graphs        │
│   Click V2  │ - ERG mode  │ - Route view            │
│ - Reconnect │ - Physics   │ - Stats panel           │
└─────────────┴─────────────┴─────────────────────────┘
        │             │               │
        └─────────────┴───────────────┘
                      │
              ┌───────┴───────┐
              │ Data Layer    │
              │ - Ride history│
              │ - FIT export  │
              │ - Config      │
              └───────────────┘
```

**Data flow:** Bluetooth Manager receives cadence/power from trainer → Simulation Engine calculates speed, applies gradient/gear resistance → sends resistance commands back to trainer → TUI Renderer displays everything.

## TUI Layout

```
┌─────────────────────────────────┬──────────────────────────────────┐
│         ROUTE VIEW              │         ROLLING GRAPHS           │
│  ┌───────────────────────────┐  │  ┌────────────────────────────┐  │
│  │  Default: Top-down map    │  │  │ Power (W)            250W  │  │
│  │  with position marker     │  │  │ ▁▃▅▆▇█▇▆▅▃▁▃▅▆███▆▅▃▁▂▃  │  │
│  │                           │  │  ├────────────────────────────┤  │
│  │  Auto-switch to elevation │  │  │ Cadence (rpm)         92   │  │
│  │  profile on climb approach│  │  │ ▅▅▅▆▆▆▅▅▅▆▆▆▅▅▅▆▆▆▅▅▅▆▆  │  │
│  │                           │  │  ├────────────────────────────┤  │
│  │  [Tab] to toggle manually │  │  │ Speed (km/h)         32.5  │  │
│  └───────────────────────────┘  │  │ ▃▄▅▅▆▆▆▇▇▇▆▆▅▅▄▄▅▆▇▇▆▅▄  │  │
├─────────────────────────────────┤  └────────────────────────────┘  │
│         RIDE STATS              │                                  │
│  Time:      01:23:45            │  Current Gear: 50x17 (2.94)      │
│  Distance:  42.3 km             │  Gradient: +4.2%                 │
│  Avg Power: 185W                │  Mode: [SIM] / ERG / FREE        │
│  Avg Cad:   88 rpm              │                                  │
│  Elevation: +523m               │  [↑↓] Shift  [←→] Resistance     │
└─────────────────────────────────┴──────────────────────────────────┘
```

**Panels:**
- **Route view** (left upper): Top-down map or elevation profile, Tab to toggle, auto-switches on climbs (>3% grade OR >30m gain in 500m)
- **Ride stats** (left lower): Running averages, elapsed time, distance, elevation gain
- **Rolling graphs** (right): Last 5-10 minutes of power/cadence/speed with current values
- **Status bar**: Current gear ratio, gradient, mode indicator, control hints

## Bluetooth & Device Handling

### Devices
- **Elite Rivo (FTMS)** - Fitness Machine Service for power/cadence data + resistance control
- **Swift Click V2** - Separate BLE device for shift button events

### Connection Flow
```
Startup
   │
   ├─→ Scan for FTMS devices
   │      └─→ Filter by saved device ID (if configured)
   │          └─→ Or present selection list
   │
   ├─→ Scan for Swift Click V2
   │      └─→ Same logic: saved ID or selection
   │
   └─→ Subscribe to characteristics:
          - FTMS: Indoor Bike Data (0x2AD2) → power, cadence
          - FTMS: Training Status, Fitness Machine Control Point
          - Swift Click: Button events
```

### Reconnection (on disconnect)
1. Pause ride recording immediately
2. Show "Reconnecting..." indicator in TUI
3. Attempt reconnect every 2 seconds, up to 30 attempts
4. On success: resume recording, show "Reconnected" briefly
5. On failure after 60s: prompt user to retry or save partial ride

### FTMS Control
- Set Target Resistance Level (for free ride / GPX simulation)
- Set Target Power (for ERG mode)

## Simulation Engine

### Gear System
```
Chainrings: [50, 34] (configurable)
Cassette:   [11,12,13,14,15,17,19,21,24,28] (configurable)
Presets:    road-2x11, gravel-1x12, custom

Current gear state: front=0|1, rear=0-9
Gear ratio = chainring[front] / cassette[rear]
Example: 50/17 = 2.94
```

### Speed Calculation
```
wheel_circumference = 2.1m (700x25c default, configurable)
speed = cadence × gear_ratio × wheel_circumference × 60 / 1000
      = rpm × ratio × 2.1 × 0.06 km/h
```

### Resistance Model (GPX mode)
```
base_resistance = f(speed)  // air resistance curve
gradient_factor = gradient% × weight × gravity_constant
gear_factor     = modifier based on gear ratio

target_resistance = base_resistance + gradient_factor + gear_factor
```

### ERG Mode
```
target_power = user_set_watts
// FTMS handles cadence-independent power targeting
// Just send Set Target Power command
```

### Modes
- **SIM** - GPX simulation, resistance from gradient + gears
- **ERG** - Fixed power target, trainer auto-adjusts
- **FREE** - Manual resistance control with arrow keys

## GPX Route Handling

### Loading a Route
```
goc ride --gpx ~/routes/alpe-dhuez.gpx
goc ride --gpx ~/routes/local-loop.gpx --start-km 5.2
```

### GPX Processing
1. Parse track points (lat, lon, elevation)
2. Calculate distances between points
3. Compute gradient for each segment
4. Build elevation profile for visualization
5. Detect climbs (for auto-view-switch triggers)

### During Ride
```
position = cumulative distance from speed integration
current_segment = find segment at position
current_gradient = segment.gradient
upcoming_climb = scan ahead for climb triggers

// Update every 100ms
resistance = calculate(gradient, gear, speed)
send_to_trainer(resistance)
```

### Map Rendering (top-down view)
- Scale route to fit panel
- Draw polyline of track
- Show position marker (moving dot)
- Highlight completed vs upcoming sections

### Elevation Profile Rendering
- X-axis: distance
- Y-axis: elevation
- Vertical line showing current position
- Color-code by gradient intensity

## Data & Persistence

### Config File (`~/.config/goc/config.toml`)
```toml
[trainer]
device_id = "XX:XX:XX:XX:XX:XX"  # saved after first pairing

[shifter]
device_id = "YY:YY:YY:YY:YY:YY"

[bike]
preset = "road-2x11"  # or "custom"
chainrings = [50, 34]
cassette = [11,12,13,14,15,17,19,21,24,28]
wheel_circumference = 2.1
rider_weight = 75  # kg, for gradient calc

[display]
graph_window_minutes = 5
climb_gradient_threshold = 3.0
climb_elevation_threshold = 30  # meters in 500m

[controls]
shift_up = "Up"
shift_down = "Down"
resistance_up = "Right"
resistance_down = "Left"
pause = "Space"
toggle_view = "Tab"
```

### Ride History (`~/.local/share/goc/`)
```
rides/
  2024-11-22-083012.fit   # FIT file for export
  2024-11-22-083012.json  # metadata for TUI history view
history.db                # SQLite: quick queries for ride list
```

### History View in TUI
- List past rides: date, duration, distance, avg power
- Select to view summary or locate FIT file for upload

## Controls

| Key | Action |
|-----|--------|
| ↑ / ↓ | Shift up / down |
| ← / → | Decrease / increase resistance (FREE mode) |
| Space | Pause / resume ride |
| Tab | Toggle route view (map / elevation) |
| q | Quit (prompts to save if ride in progress) |

## Tech Stack

| Component | Choice |
|-----------|--------|
| Language | Go |
| TUI Framework | termdash |
| Bluetooth | tinygo.org/x/bluetooth |
| FIT Export | github.com/tormoder/fit |
| GPX Parsing | github.com/tkrajina/gpxgo |
| Config | github.com/spf13/viper (TOML) |
| Local DB | SQLite via modernc.org/sqlite |

## Feature Summary

### MVP
- Real-time metrics: power, cadence, speed (rolling graphs)
- GPX simulation with gradient-based resistance
- Virtual shifting via Swift Click V2 with realistic gear ratios
- ERG mode for fixed power training
- Free ride with manual resistance
- Auto view toggle on climb approach
- Local ride history + FIT export
- TOML config with TUI editing
- Auto-reconnect on Bluetooth drop

### Future
- Structured workouts (load workout files with intervals/ramps)
