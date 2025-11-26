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
