# goc

Terminal-based indoor cycling trainer for FTMS-compatible trainers with virtual shifting support.

> **Note:** This project is a test case for the [superpowers](https://github.com/obra/superpowers) skills framework and was mostly (if not 100%) built using [Claude Code](https://claude.com/claude-code).

## Features
- Real-time power, cadence, speed graphs
- GPX route simulation with gradient-based resistance
- Virtual gear shifting
- ERG mode
- FIT file export

## Usage
```bash
goc ride                          # Free ride
goc ride --gpx route.gpx          # GPX simulation
goc ride --erg 200                # ERG mode at 200W
goc history                       # View past rides
```

## Configuration

Configuration is stored in `~/.config/goc/config.toml`. The file is created with defaults on first run.

### resistance_scaling

**Type:** float
**Default:** 0.2
**Range:** 0.1 - 0.5

Controls how resistance force maps to trainer resistance level (0-100).

- Lower values (0.1): Lighter resistance feel
- Default (0.2): Balanced resistance
- Higher values (0.3-0.5): Heavier resistance feel

Adjust if gear shifting feels too easy or too hard.

**Example:**
```toml
[bike]
resistance_scaling = 0.2
```

### gradient_smoothing

**Type:** float
**Default:** 0.85
**Range:** 0.0 - 0.95

Controls gradient smoothing using exponential moving average (EMA).

- `0.0`: No smoothing (instant response, may feel jerky due to GPS noise)
- `0.85`: Default (smooth, natural feel with ~20-30 second lag, mimics momentum)
- `0.95`: Very smooth (minimal jitter, ~30+ second lag)

Adjust this value if gradient changes feel too sudden or too slow to respond. Higher values provide smoother resistance changes but slower response to real climbs. Lower values are more responsive but may feel jerky on routes with GPS elevation noise.

**Example:**
```toml
[bike]
gradient_smoothing = 0.85
```
