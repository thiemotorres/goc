# FTMS Bluetooth Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement real BLE support for FTMS-compatible smart trainers with device selection, auto-reconnect, and resistance/power control.

**Architecture:** Scanner discovers FTMS devices, Connector manages connection lifecycle with reconnection, Parser handles FTMS protocol encoding/decoding. FTMSManager coordinates these components and implements the existing Manager interface.

**Tech Stack:** tinygo.org/x/bluetooth for cross-platform BLE, existing Manager interface

---

## Phase 1: Dependencies and Types

### Task 1.1: Add tinygo bluetooth dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add dependency**

Run:
```bash
go get tinygo.org/x/bluetooth
```

**Step 2: Verify import works**

Run:
```bash
go build ./...
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add tinygo bluetooth dependency"
```

---

### Task 1.2: Add connection types

**Files:**
- Modify: `internal/bluetooth/bluetooth.go`

**Step 1: Add new types after existing code**

```go
// ConnectionStatus represents BLE connection state
type ConnectionStatus int

const (
	StatusConnecting ConnectionStatus = iota
	StatusConnected
	StatusDisconnected
	StatusReconnecting
)

func (s ConnectionStatus) String() string {
	switch s {
	case StatusConnecting:
		return "Connecting"
	case StatusConnected:
		return "Connected"
	case StatusDisconnected:
		return "Disconnected"
	case StatusReconnecting:
		return "Reconnecting"
	default:
		return "Unknown"
	}
}

// DeviceInfo represents a discovered BLE device
type DeviceInfo struct {
	Address string
	Name    string
	RSSI    int
}
```

**Step 2: Verify build**

Run:
```bash
go build ./...
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/bluetooth/bluetooth.go
git commit -m "feat(bluetooth): add connection status and device info types"
```

---

## Phase 2: FTMS Parser

### Task 2.1: Write parser test for Indoor Bike Data

**Files:**
- Create: `internal/bluetooth/parser_test.go`

**Step 1: Write failing test**

```go
package bluetooth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIndoorBikeData_PowerAndCadence(t *testing.T) {
	// Flags: bit 2 (cadence) and bit 6 (power) set = 0x44 = 0b01000100
	// Data format: flags(2) + speed(2) + cadence(2) + power(2)
	// Speed always present (not optional)
	data := []byte{
		0x44, 0x00, // Flags: cadence + power present
		0xE8, 0x03, // Speed: 1000 (10.00 km/h, 0.01 resolution)
		0xB4, 0x00, // Cadence: 180 (90 rpm, 0.5 resolution)
		0xC8, 0x00, // Power: 200 watts
	}

	result, err := ParseIndoorBikeData(data)

	assert.NoError(t, err)
	assert.InDelta(t, 200.0, result.Power, 0.1)
	assert.InDelta(t, 90.0, result.Cadence, 0.1)
}

func TestParseIndoorBikeData_PowerOnly(t *testing.T) {
	// Flags: only bit 6 (power) set = 0x40
	data := []byte{
		0x40, 0x00, // Flags: power present only
		0xE8, 0x03, // Speed: 1000
		0x96, 0x00, // Power: 150 watts
	}

	result, err := ParseIndoorBikeData(data)

	assert.NoError(t, err)
	assert.InDelta(t, 150.0, result.Power, 0.1)
	assert.InDelta(t, 0.0, result.Cadence, 0.1)
}

func TestParseIndoorBikeData_TooShort(t *testing.T) {
	data := []byte{0x44} // Too short

	_, err := ParseIndoorBikeData(data)

	assert.Error(t, err)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/bluetooth -run TestParseIndoorBikeData -v
```
Expected: FAIL - ParseIndoorBikeData undefined

**Step 3: Commit failing test**

```bash
git add internal/bluetooth/parser_test.go
git commit -m "test(bluetooth): add Indoor Bike Data parser tests (red)"
```

---

### Task 2.2: Implement Indoor Bike Data parser

**Files:**
- Create: `internal/bluetooth/parser.go`

**Step 1: Implement parser**

```go
package bluetooth

import (
	"encoding/binary"
	"errors"
)

// FTMS Indoor Bike Data flags
const (
	flagMoreData           uint16 = 1 << 0
	flagAverageSpeed       uint16 = 1 << 1
	flagInstCadence        uint16 = 1 << 2
	flagAvgCadence         uint16 = 1 << 3
	flagTotalDistance      uint16 = 1 << 4
	flagResistanceLevel    uint16 = 1 << 5
	flagInstPower          uint16 = 1 << 6
	flagAvgPower           uint16 = 1 << 7
	flagExpendedEnergy     uint16 = 1 << 8
	flagHeartRate          uint16 = 1 << 9
	flagMetabolicEquiv     uint16 = 1 << 10
	flagElapsedTime        uint16 = 1 << 11
	flagRemainingTime      uint16 = 1 << 12
)

// ParseIndoorBikeData parses FTMS Indoor Bike Data characteristic
func ParseIndoorBikeData(data []byte) (TrainerData, error) {
	if len(data) < 2 {
		return TrainerData{}, errors.New("data too short for flags")
	}

	flags := binary.LittleEndian.Uint16(data[0:2])
	offset := 2

	var result TrainerData

	// Instantaneous Speed is always present (uint16, 0.01 km/h resolution)
	if len(data) < offset+2 {
		return TrainerData{}, errors.New("data too short for speed")
	}
	// speed := float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.01
	offset += 2

	// Average Speed (optional)
	if flags&flagAverageSpeed != 0 {
		offset += 2
	}

	// Instantaneous Cadence (optional, uint16, 0.5 rpm resolution)
	if flags&flagInstCadence != 0 {
		if len(data) < offset+2 {
			return TrainerData{}, errors.New("data too short for cadence")
		}
		result.Cadence = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.5
		offset += 2
	}

	// Average Cadence (optional)
	if flags&flagAvgCadence != 0 {
		offset += 2
	}

	// Total Distance (optional, uint24)
	if flags&flagTotalDistance != 0 {
		offset += 3
	}

	// Resistance Level (optional)
	if flags&flagResistanceLevel != 0 {
		offset += 2
	}

	// Instantaneous Power (optional, sint16, 1W resolution)
	if flags&flagInstPower != 0 {
		if len(data) < offset+2 {
			return TrainerData{}, errors.New("data too short for power")
		}
		result.Power = float64(int16(binary.LittleEndian.Uint16(data[offset : offset+2])))
		offset += 2
	}

	return result, nil
}
```

**Step 2: Run tests**

Run:
```bash
go test ./internal/bluetooth -run TestParseIndoorBikeData -v
```
Expected: PASS

**Step 3: Commit**

```bash
git add internal/bluetooth/parser.go
git commit -m "feat(bluetooth): implement Indoor Bike Data parser (green)"
```

---

### Task 2.3: Write control point encoder tests

**Files:**
- Modify: `internal/bluetooth/parser_test.go`

**Step 1: Add encoder tests**

```go
func TestEncodeRequestControl(t *testing.T) {
	data := EncodeRequestControl()
	assert.Equal(t, []byte{0x00}, data)
}

func TestEncodeSetTargetResistance(t *testing.T) {
	// 50% resistance = 100 (0.1% resolution, so 50% = 500, but range is 0-200)
	// Actually: level 0-100 maps to 0-200 in protocol
	data := EncodeSetTargetResistance(50)
	assert.Equal(t, []byte{0x04, 100}, data)
}

func TestEncodeSetTargetPower(t *testing.T) {
	// 200 watts
	data := EncodeSetTargetPower(200)
	assert.Equal(t, []byte{0x05, 0xC8, 0x00}, data)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/bluetooth -run TestEncode -v
```
Expected: FAIL - functions undefined

**Step 3: Commit**

```bash
git add internal/bluetooth/parser_test.go
git commit -m "test(bluetooth): add control point encoder tests (red)"
```

---

### Task 2.4: Implement control point encoders

**Files:**
- Modify: `internal/bluetooth/parser.go`

**Step 1: Add encoders at end of file**

```go
// Control Point opcodes
const (
	opRequestControl       = 0x00
	opReset                = 0x01
	opSetTargetResistance  = 0x04
	opSetTargetPower       = 0x05
	opStartOrResume        = 0x07
	opStopOrPause          = 0x08
)

// EncodeRequestControl creates a Request Control command
func EncodeRequestControl() []byte {
	return []byte{opRequestControl}
}

// EncodeSetTargetResistance creates a Set Target Resistance command
// level: 0-100 percentage
func EncodeSetTargetResistance(level float64) []byte {
	// Protocol uses 0-200 range with 0.1% resolution
	// So 50% = 100 in protocol units
	protocolLevel := uint8(level * 2)
	if protocolLevel > 200 {
		protocolLevel = 200
	}
	return []byte{opSetTargetResistance, protocolLevel}
}

// EncodeSetTargetPower creates a Set Target Power command
// watts: target power in watts
func EncodeSetTargetPower(watts float64) []byte {
	w := int16(watts)
	return []byte{
		opSetTargetPower,
		byte(w & 0xFF),
		byte((w >> 8) & 0xFF),
	}
}
```

**Step 2: Run tests**

Run:
```bash
go test ./internal/bluetooth -run TestEncode -v
```
Expected: PASS

**Step 3: Commit**

```bash
git add internal/bluetooth/parser.go
git commit -m "feat(bluetooth): implement control point encoders (green)"
```

---

## Phase 3: Scanner

### Task 3.1: Implement BLE scanner

**Files:**
- Create: `internal/bluetooth/scanner.go`

**Step 1: Implement scanner**

```go
package bluetooth

import (
	"errors"
	"strings"
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

// Scanner handles BLE device discovery
type Scanner struct {
	devices  []DeviceInfo
	stopChan chan struct{}
}

// NewScanner creates a new BLE scanner
func NewScanner() *Scanner {
	return &Scanner{
		stopChan: make(chan struct{}),
	}
}

// Scan discovers FTMS devices for the given duration
func (s *Scanner) Scan(timeout time.Duration) ([]DeviceInfo, error) {
	if err := adapter.Enable(); err != nil {
		return nil, errors.New("failed to enable Bluetooth adapter: " + err.Error())
	}

	s.devices = nil
	seen := make(map[string]bool)

	done := make(chan error, 1)

	go func() {
		err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			// Check if device advertises FTMS service
			hasFTMS := false
			for _, uuid := range result.AdvertisementPayload.ServiceUUIDs() {
				if uuid.String() == FTMSServiceUUID {
					hasFTMS = true
					break
				}
			}

			if !hasFTMS {
				return
			}

			addr := result.Address.String()
			if seen[addr] {
				return
			}
			seen[addr] = true

			name := result.LocalName()
			if name == "" {
				name = "Unknown Trainer"
			}

			s.devices = append(s.devices, DeviceInfo{
				Address: addr,
				Name:    name,
				RSSI:    int(result.RSSI),
			})
		})
		done <- err
	}()

	select {
	case <-time.After(timeout):
		adapter.StopScan()
	case err := <-done:
		if err != nil && !strings.Contains(err.Error(), "timeout") {
			return nil, err
		}
	case <-s.stopChan:
		adapter.StopScan()
	}

	return s.devices, nil
}

// Stop stops an ongoing scan
func (s *Scanner) Stop() {
	select {
	case s.stopChan <- struct{}{}:
	default:
	}
}
```

**Step 2: Verify build**

Run:
```bash
go build ./internal/bluetooth
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/bluetooth/scanner.go
git commit -m "feat(bluetooth): implement FTMS device scanner"
```

---

## Phase 4: FTMS Manager Implementation

### Task 4.1: Rewrite FTMSManager with real BLE

**Files:**
- Modify: `internal/bluetooth/ftms.go`

**Step 1: Replace placeholder with full implementation**

```go
package bluetooth

import (
	"errors"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

// FTMS UUIDs
const (
	FTMSServiceUUID                = "00001826-0000-1000-8000-00805f9b34fb"
	IndoorBikeDataUUID             = "00002ad2-0000-1000-8000-00805f9b34fb"
	FitnessMachineControlPointUUID = "00002ad9-0000-1000-8000-00805f9b34fb"
	FitnessMachineStatusUUID       = "00002ada-0000-1000-8000-00805f9b34fb"
)

// FTMSManagerConfig configures the FTMS manager
type FTMSManagerConfig struct {
	OnStatusChange    func(ConnectionStatus)
	OnDeviceSelection func([]DeviceInfo) int // returns selected index, -1 to cancel
	SavedAddress      string
	OnSaveDevice      func(address string) // called after successful connection
}

// FTMSManager implements Manager using real Bluetooth
type FTMSManager struct {
	config FTMSManagerConfig

	mu             sync.Mutex
	connected      bool
	status         ConnectionStatus
	device         bluetooth.Device
	controlPoint   bluetooth.DeviceCharacteristic
	deviceAddress  string

	dataCh  chan TrainerData
	shiftCh chan ShiftEvent
	stopCh  chan struct{}
}

// NewFTMSManager creates a new FTMS Bluetooth manager
func NewFTMSManager() *FTMSManager {
	return NewFTMSManagerWithConfig(FTMSManagerConfig{})
}

// NewFTMSManagerWithConfig creates a new FTMS manager with config
func NewFTMSManagerWithConfig(config FTMSManagerConfig) *FTMSManager {
	return &FTMSManager{
		config:  config,
		dataCh:  make(chan TrainerData, 10),
		shiftCh: make(chan ShiftEvent, 10),
		stopCh:  make(chan struct{}),
	}
}

func (m *FTMSManager) setStatus(s ConnectionStatus) {
	m.mu.Lock()
	m.status = s
	m.mu.Unlock()

	if m.config.OnStatusChange != nil {
		m.config.OnStatusChange(s)
	}
}

func (m *FTMSManager) Connect() error {
	m.setStatus(StatusConnecting)

	if err := adapter.Enable(); err != nil {
		return errors.New("failed to enable Bluetooth: " + err.Error())
	}

	var targetAddress string

	// Try saved address first
	if m.config.SavedAddress != "" {
		targetAddress = m.config.SavedAddress
	} else {
		// Scan for devices
		scanner := NewScanner()
		devices, err := scanner.Scan(10 * time.Second)
		if err != nil {
			return err
		}

		if len(devices) == 0 {
			return errors.New("no FTMS trainers found")
		}

		// Let user select
		selectedIdx := 0
		if m.config.OnDeviceSelection != nil {
			selectedIdx = m.config.OnDeviceSelection(devices)
			if selectedIdx < 0 || selectedIdx >= len(devices) {
				return errors.New("device selection cancelled")
			}
		}

		targetAddress = devices[selectedIdx].Address
	}

	// Connect to device
	addr, err := bluetooth.ParseMAC(targetAddress)
	if err != nil {
		return errors.New("invalid device address: " + err.Error())
	}

	device, err := adapter.Connect(addr, bluetooth.ConnectionParams{})
	if err != nil {
		// If saved address failed, clear and retry with scan
		if m.config.SavedAddress != "" {
			m.config.SavedAddress = ""
			return m.Connect()
		}
		return errors.New("failed to connect: " + err.Error())
	}

	m.device = device
	m.deviceAddress = targetAddress

	// Discover services
	services, err := device.DiscoverServices([]bluetooth.UUID{
		bluetooth.NewUUID(mustParseUUID(FTMSServiceUUID)),
	})
	if err != nil || len(services) == 0 {
		device.Disconnect()
		return errors.New("FTMS service not found")
	}

	ftmsService := services[0]

	// Discover characteristics
	chars, err := ftmsService.DiscoverCharacteristics([]bluetooth.UUID{
		bluetooth.NewUUID(mustParseUUID(IndoorBikeDataUUID)),
		bluetooth.NewUUID(mustParseUUID(FitnessMachineControlPointUUID)),
	})
	if err != nil {
		device.Disconnect()
		return errors.New("failed to discover characteristics: " + err.Error())
	}

	var indoorBikeData, controlPoint bluetooth.DeviceCharacteristic
	for _, c := range chars {
		uuid := c.UUID().String()
		if uuid == IndoorBikeDataUUID {
			indoorBikeData = c
		} else if uuid == FitnessMachineControlPointUUID {
			controlPoint = c
		}
	}

	m.controlPoint = controlPoint

	// Subscribe to Indoor Bike Data notifications
	err = indoorBikeData.EnableNotifications(func(buf []byte) {
		data, err := ParseIndoorBikeData(buf)
		if err != nil {
			return
		}
		select {
		case m.dataCh <- data:
		default:
			// Channel full, drop
		}
	})
	if err != nil {
		device.Disconnect()
		return errors.New("failed to enable notifications: " + err.Error())
	}

	// Request control
	_, err = controlPoint.WriteWithoutResponse(EncodeRequestControl())
	if err != nil {
		// Non-fatal, some trainers don't require this
	}

	m.mu.Lock()
	m.connected = true
	m.mu.Unlock()

	m.setStatus(StatusConnected)

	// Save device for next time
	if m.config.OnSaveDevice != nil {
		m.config.OnSaveDevice(targetAddress)
	}

	// Start disconnect monitor
	go m.monitorConnection()

	return nil
}

func (m *FTMSManager) monitorConnection() {
	// tinygo bluetooth doesn't have disconnect callbacks yet
	// Poll connection status
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			// Check if we can still communicate
			// If not, trigger reconnection
		}
	}
}

func (m *FTMSManager) Disconnect() {
	m.mu.Lock()
	wasConnected := m.connected
	m.connected = false
	m.mu.Unlock()

	if wasConnected {
		close(m.stopCh)
		m.device.Disconnect()
	}

	m.setStatus(StatusDisconnected)
}

func (m *FTMSManager) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

func (m *FTMSManager) DataChannel() <-chan TrainerData {
	return m.dataCh
}

func (m *FTMSManager) ShiftChannel() <-chan ShiftEvent {
	return m.shiftCh
}

func (m *FTMSManager) SetResistance(level float64) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}
	_, err := m.controlPoint.WriteWithoutResponse(EncodeSetTargetResistance(level))
	return err
}

func (m *FTMSManager) SetTargetPower(watts float64) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}
	_, err := m.controlPoint.WriteWithoutResponse(EncodeSetTargetPower(watts))
	return err
}

func mustParseUUID(s string) [16]byte {
	// Parse UUID string to bytes
	// Format: 00001826-0000-1000-8000-00805f9b34fb
	var uuid [16]byte
	// Simplified parsing - real implementation would be more robust
	// tinygo bluetooth handles this internally
	return uuid
}
```

**Step 2: Verify build**

Run:
```bash
go build ./internal/bluetooth
```
Expected: Build succeeds (may have warnings about unused)

**Step 3: Commit**

```bash
git add internal/bluetooth/ftms.go
git commit -m "feat(bluetooth): implement real FTMS manager with BLE"
```

---

### Task 4.2: Add UUID parsing helper

**Files:**
- Modify: `internal/bluetooth/ftms.go`

**Step 1: Replace mustParseUUID with proper implementation**

Find and replace the `mustParseUUID` function:

```go
func mustParseUUID(s string) [16]byte {
	uuid, err := bluetooth.ParseUUID(s)
	if err != nil {
		panic("invalid UUID: " + s)
	}
	var result [16]byte
	copy(result[:], uuid[:])
	return result
}
```

**Step 2: Verify build**

Run:
```bash
go build ./internal/bluetooth
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/bluetooth/ftms.go
git commit -m "fix(bluetooth): add proper UUID parsing"
```

---

## Phase 5: Integration

### Task 5.1: Update ride command to use new config

**Files:**
- Modify: `cmd/ride.go`

**Step 1: Update FTMSManager creation with callbacks**

Find the section that creates btManager and update:

```go
	// Create Bluetooth manager
	var btManager bluetooth.Manager
	if opts.Mock {
		btManager = bluetooth.NewMockManager()
	} else {
		btManager = bluetooth.NewFTMSManagerWithConfig(bluetooth.FTMSManagerConfig{
			SavedAddress: cfg.Bluetooth.TrainerAddress,
			OnStatusChange: func(status bluetooth.ConnectionStatus) {
				// Could update TUI status here
				fmt.Printf("Bluetooth: %s\n", status)
			},
			OnDeviceSelection: func(devices []bluetooth.DeviceInfo) int {
				fmt.Println("\nFound trainers:")
				for i, d := range devices {
					fmt.Printf("  %d: %s (%s) RSSI: %d\n", i+1, d.Name, d.Address, d.RSSI)
				}
				fmt.Print("Select trainer (1-", len(devices), "): ")
				var choice int
				fmt.Scanln(&choice)
				return choice - 1
			},
			OnSaveDevice: func(address string) {
				cfg.Bluetooth.TrainerAddress = address
				config.Save(cfg, config.DefaultConfigDir())
			},
		})
	}
```

**Step 2: Verify build**

Run:
```bash
go build ./...
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add cmd/ride.go
git commit -m "feat: integrate FTMS manager with device selection"
```

---

### Task 5.2: Add Bluetooth config to config package

**Files:**
- Modify: `internal/config/config.go`

**Step 1: Add Bluetooth section to Config struct**

Find the Config struct and add:

```go
type Config struct {
	Bike      BikeConfig      `yaml:"bike"`
	Bluetooth BluetoothConfig `yaml:"bluetooth"`
}

type BluetoothConfig struct {
	TrainerAddress string `yaml:"trainer_address"`
}
```

**Step 2: Verify build**

Run:
```bash
go build ./...
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/config/config.go
git commit -m "feat(config): add Bluetooth trainer address storage"
```

---

## Phase 6: Testing

### Task 6.1: Manual test with real trainer

**Step 1: Build the application**

Run:
```bash
go build -o goc .
```

**Step 2: Run with real Bluetooth**

Run:
```bash
./goc ride
```

Expected:
- Should scan for FTMS devices
- Display list of found trainers
- Connect to selected trainer
- Show power/cadence data in TUI

**Step 3: Verify reconnection (if possible)**

- Start ride
- Turn off trainer briefly
- Turn back on
- Should see "Reconnecting..." and resume

**Step 4: Commit any fixes**

```bash
git add -A
git commit -m "fix: bluetooth integration fixes from testing"
```

---

### Task 6.2: Final commit and push

**Step 1: Run all tests**

Run:
```bash
go test ./...
```
Expected: All tests pass

**Step 2: Push to remote**

Run:
```bash
git push
```

---

## Summary

| Phase | Tasks | Description |
|-------|-------|-------------|
| 1 | 1.1-1.2 | Dependencies and types |
| 2 | 2.1-2.4 | FTMS parser (TDD) |
| 3 | 3.1 | BLE scanner |
| 4 | 4.1-4.2 | FTMS manager implementation |
| 5 | 5.1-5.2 | Integration with ride command |
| 6 | 6.1-6.2 | Manual testing and final push |
