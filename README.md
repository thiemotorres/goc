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
