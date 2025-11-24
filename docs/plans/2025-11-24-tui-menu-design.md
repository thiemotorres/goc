# TUI Menu System Design

## Overview

Replace the current CLI-only interface with a full Bubble Tea TUI while preserving CLI flags for scripting. The ride screen will also migrate from termdash to Bubble Tea for consistent UX.

## Entry Point

- `goc` (no args) → opens TUI menu
- `goc ride -gpx file.gpx` → still works for scripting
- `goc history` → still works

## Menu Structure

### Main Menu

```
goc - Indoor Cycling Trainer

> Start Ride
  Browse Routes
  Ride History
  Settings
  Quit
```

### Start Ride Submenu

```
Start Ride

> Free Ride (no target)
  ERG Mode (fixed power)
  Ride a Route
  ← Back
```

- User selects ride type first
- Trainer connection happens when actually starting the ride
- ERG Mode prompts for target watts before starting

### Browse Routes

Scans a configurable folder (default: `~/.config/goc/routes/`) for GPX files.

```
Routes

> Alpe d'Huez         13.8 km   1,071m ↑   7.9% avg
  Col du Galibier     34.8 km   2,120m ↑   6.1% avg
  Recovery Spin        8.2 km      42m ↑   0.5% avg
  ← Back
```

Selecting a route shows a preview screen:

```
Alpe d'Huez
────────────────────────────────
Distance:    13.8 km
Elevation:   1,071m ↑
Avg Grade:   7.9%
Max Grade:   13.1%

▁▂▃▄▅▆▆▇▇█████▇▇▆▅▄

        [Start]  [Back]
```

- Elevation profile shown as ASCII sparkline
- Confirm starts trainer connection flow

### Ride History

```
Ride History

> Nov 24  Alpe d'Huez      45:32   187W avg   1,071m ↑
  Nov 22  Free Ride        30:15   165W avg       -
  Nov 20  ERG 150W         60:00   150W avg       -
  ← Back
```

- View only for initial version
- Selecting shows detailed stats

### Settings

```
Settings

> Trainer Connection
  Routes Folder
  ← Back
```

**Trainer Connection:**
- Shows currently saved trainer (name + address)
- Options: Forget Device, Scan for New

**Routes Folder:**
- Shows current path
- Text input to change

## Configuration Changes

Add to config:

```go
type Config struct {
    TrainerAddress string // existing
    RoutesFolder   string // new - default: ~/.config/goc/routes/
}
```

## Technology

- **Bubble Tea** - TUI framework (replaces termdash)
- **ntcharts** - Line charts for ride screen (github.com/NimbleMarkets/ntcharts)
- **Lip Gloss** - Styling
- **Bubbles** - Standard components (list, textinput, etc.)

## Ride Screen Migration

Current termdash layout preserved in Bubble Tea:
- Power/Cadence/Speed line charts (via ntcharts)
- Route info panel
- Stats panel
- Status bar with controls

Keyboard controls unchanged:
- ↑↓ Shift gears
- ←→ Adjust resistance
- Space: Pause
- q: Quit (returns to menu instead of exiting)

## Navigation Flow

```
Main Menu
├── Start Ride
│   ├── Free Ride → [Connect] → Ride Screen → [End] → Main Menu
│   ├── ERG Mode → [Set Watts] → [Connect] → Ride Screen → Main Menu
│   └── Ride a Route → Browse Routes
├── Browse Routes
│   └── [Select] → Preview → [Start] → [Connect] → Ride Screen → Main Menu
├── Ride History
│   └── [Select] → Ride Details → [Back] → Ride History
├── Settings
│   ├── Trainer Connection
│   └── Routes Folder
└── Quit → Exit
```

## Out of Scope (Future)

- Ride history: Export FIT, Delete, Ride Again actions
- Route preview: Detailed multi-line elevation graph
- Settings: Auto-connect, connection timeout preferences
- Workout builder / structured workouts
