# FTMS Bluetooth Implementation Design

## Overview

Implement real Bluetooth Low Energy (BLE) support for FTMS-compatible smart trainers, replacing the current mock implementation.

## Requirements

- Cross-platform support (Linux, macOS, Windows)
- Interactive device selection with config-based reconnection
- Auto-reconnect on connection loss with pause/notify UX
- Power and cadence data (MVP)
- Trainer only (shifter support deferred)

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Scanner   │────▶│  Connector  │────▶│   Parser    │
│             │     │             │     │             │
│ - Discover  │     │ - Connect   │     │ - Indoor    │
│ - Filter    │     │ - Subscribe │     │   Bike Data │
│ - List      │     │ - Reconnect │     │ - Control   │
└─────────────┘     └─────────────┘     └─────────────┘
```

**Scanner**: Discovers BLE devices advertising FTMS service, returns list for selection.

**Connector**: Manages connection lifecycle - initial connect, characteristic discovery, subscriptions, and automatic reconnection on disconnect.

**Parser**: Decodes FTMS Indoor Bike Data notifications (power, cadence) and encodes Control Point commands (set resistance, set target power).

## Library Choice

Using `tinygo.org/x/bluetooth` for cross-platform BLE support.

## Connection Flow

1. Check config for saved trainer address
2. If saved device exists, try to connect (5s timeout)
3. If no saved device or connection fails:
   - Scan for FTMS devices (10s timeout)
   - Display selection menu via callback to TUI
   - Connect to selected device
   - Save address to config for next time
4. Discover FTMS service and characteristics
5. Subscribe to Indoor Bike Data notifications

## FTMS Protocol

### Indoor Bike Data (0x2AD2) - Notifications

```
Byte 0-1: Flags (little-endian)
  Bit 2: Instantaneous Cadence present
  Bit 6: Instantaneous Power present

Fields (variable based on flags):
  - Instantaneous Cadence: uint16, 0.5 rpm resolution
  - Instantaneous Power: sint16, 1 watt resolution
```

### Fitness Machine Control Point (0x2AD9) - Commands

```
Request Control:       [0x00]
Set Target Resistance: [0x04, level_uint8]        (0-200, 0.1% resolution)
Set Target Power:      [0x05, power_lo, power_hi] (sint16, 1W resolution)
```

## Reconnection Handling

On disconnect:
1. Set state to "disconnected"
2. Notify TUI via status callback
3. Attempt reconnection every 2 seconds
4. After 30 seconds, give up and return error

## API Changes

```go
type ConnectionStatus int
const (
    StatusConnecting ConnectionStatus = iota
    StatusConnected
    StatusDisconnected
    StatusReconnecting
)

type DeviceInfo struct {
    Address string
    Name    string
    RSSI    int
}

type FTMSManagerConfig struct {
    OnStatusChange   func(ConnectionStatus)
    OnDeviceSelection func([]DeviceInfo) int  // returns selected index
    SavedAddress     string                   // from config
}
```

## Deferred

- Swift Click V2 / shifter support (separate BLE profile)
- Additional FTMS data fields (speed, heart rate, etc.)
- Multiple simultaneous device connections
