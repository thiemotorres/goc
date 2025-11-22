# goc Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a TUI indoor cycling trainer app that connects to Elite Rivo (FTMS) and Swift Click V2 via Bluetooth, displays real-time metrics, simulates GPX routes, and exports rides as FIT files.

**Architecture:** Component-based Go application with clear separation: Bluetooth manager handles device communication, Simulation engine computes physics/resistance, TUI renderer displays via termdash, Data layer handles persistence. Components communicate via channels.

**Tech Stack:** Go 1.21+, termdash (TUI), tinygo.org/x/bluetooth (BLE), github.com/tormoder/fit (FIT export), github.com/tkrajina/gpxgo (GPX), github.com/spf13/viper (config), modernc.org/sqlite (DB)

---

## Phase 1: Project Setup

### Task 1.1: Initialize Go Module

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `README.md`

**Step 1: Initialize module**

Run:
```bash
go mod init github.com/thiemo/goc
```
Expected: `go.mod` created

**Step 2: Create minimal main.go**

```go
package main

import "fmt"

func main() {
	fmt.Println("goc - indoor cycling trainer")
}
```

**Step 3: Verify it runs**

Run: `go run main.go`
Expected: `goc - indoor cycling trainer`

**Step 4: Create README**

```markdown
# goc

Terminal-based indoor cycling trainer for Elite Rivo with Swift Click V2 virtual shifting.

## Features
- Real-time power, cadence, speed graphs
- GPX route simulation with gradient-based resistance
- Virtual gear shifting
- ERG mode
- FIT file export

## Usage
\`\`\`bash
goc ride                          # Free ride
goc ride --gpx route.gpx          # GPX simulation
goc ride --erg 200                # ERG mode at 200W
goc history                       # View past rides
\`\`\`
```

**Step 5: Commit**

```bash
git add go.mod main.go README.md
git commit -m "feat: initialize goc project"
```

---

### Task 1.2: Add Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add all dependencies**

Run:
```bash
go get tinygo.org/x/bluetooth
go get github.com/mum4k/termdash
go get github.com/tormoder/fit
go get github.com/tkrajina/gpxgo/gpx
go get github.com/spf13/viper
go get modernc.org/sqlite
go get github.com/stretchr/testify
```

**Step 2: Tidy**

Run: `go mod tidy`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat: add project dependencies"
```

---

### Task 1.3: Create Directory Structure

**Files:**
- Create: `internal/bluetooth/bluetooth.go`
- Create: `internal/simulation/simulation.go`
- Create: `internal/tui/tui.go`
- Create: `internal/gpx/gpx.go`
- Create: `internal/data/data.go`
- Create: `internal/config/config.go`
- Create: `cmd/ride.go`
- Create: `cmd/history.go`

**Step 1: Create directory structure with placeholder files**

```bash
mkdir -p internal/bluetooth internal/simulation internal/tui internal/gpx internal/data internal/config cmd
```

**Step 2: Create bluetooth.go**

```go
package bluetooth

// Manager handles BLE connections to trainer and shifter
type Manager struct{}

// NewManager creates a new Bluetooth manager
func NewManager() *Manager {
	return &Manager{}
}
```

**Step 3: Create simulation.go**

```go
package simulation

// Engine handles physics calculations and resistance
type Engine struct{}

// NewEngine creates a new simulation engine
func NewEngine() *Engine {
	return &Engine{}
}
```

**Step 4: Create tui.go**

```go
package tui

// Renderer handles the terminal UI
type Renderer struct{}

// NewRenderer creates a new TUI renderer
func NewRenderer() *Renderer {
	return &Renderer{}
}
```

**Step 5: Create gpx.go**

```go
package gpx

// Route represents a loaded GPX route
type Route struct{}

// Load parses a GPX file
func Load(path string) (*Route, error) {
	return &Route{}, nil
}
```

**Step 6: Create data.go**

```go
package data

// Store handles ride persistence
type Store struct{}

// NewStore creates a new data store
func NewStore() *Store {
	return &Store{}
}
```

**Step 7: Create config.go**

```go
package config

// Config holds application configuration
type Config struct{}

// Load reads config from file
func Load() (*Config, error) {
	return &Config{}, nil
}
```

**Step 8: Create cmd/ride.go**

```go
package cmd

// Ride starts a cycling session
func Ride() error {
	return nil
}
```

**Step 9: Create cmd/history.go**

```go
package cmd

// History shows past rides
func History() error {
	return nil
}
```

**Step 10: Verify build**

Run: `go build ./...`
Expected: No errors

**Step 11: Commit**

```bash
git add internal/ cmd/
git commit -m "feat: create project directory structure"
```

---

## Phase 2: Configuration

### Task 2.1: Config Types

**Files:**
- Modify: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write the failing test**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Use temp dir with no config file
	tmpDir := t.TempDir()

	cfg, err := Load(tmpDir)
	require.NoError(t, err)

	// Check defaults
	assert.Equal(t, []int{50, 34}, cfg.Bike.Chainrings)
	assert.Equal(t, []int{11, 12, 13, 14, 15, 17, 19, 21, 24, 28}, cfg.Bike.Cassette)
	assert.Equal(t, 2.1, cfg.Bike.WheelCircumference)
	assert.Equal(t, 75.0, cfg.Bike.RiderWeight)
	assert.Equal(t, 5, cfg.Display.GraphWindowMinutes)
	assert.Equal(t, 3.0, cfg.Display.ClimbGradientThreshold)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/... -v`
Expected: FAIL

**Step 3: Write implementation**

```go
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	Trainer  TrainerConfig  `mapstructure:"trainer"`
	Shifter  ShifterConfig  `mapstructure:"shifter"`
	Bike     BikeConfig     `mapstructure:"bike"`
	Display  DisplayConfig  `mapstructure:"display"`
	Controls ControlsConfig `mapstructure:"controls"`
}

type TrainerConfig struct {
	DeviceID string `mapstructure:"device_id"`
}

type ShifterConfig struct {
	DeviceID string `mapstructure:"device_id"`
}

type BikeConfig struct {
	Preset             string  `mapstructure:"preset"`
	Chainrings         []int   `mapstructure:"chainrings"`
	Cassette           []int   `mapstructure:"cassette"`
	WheelCircumference float64 `mapstructure:"wheel_circumference"`
	RiderWeight        float64 `mapstructure:"rider_weight"`
}

type DisplayConfig struct {
	GraphWindowMinutes      int     `mapstructure:"graph_window_minutes"`
	ClimbGradientThreshold  float64 `mapstructure:"climb_gradient_threshold"`
	ClimbElevationThreshold float64 `mapstructure:"climb_elevation_threshold"`
}

type ControlsConfig struct {
	ShiftUp        string `mapstructure:"shift_up"`
	ShiftDown      string `mapstructure:"shift_down"`
	ResistanceUp   string `mapstructure:"resistance_up"`
	ResistanceDown string `mapstructure:"resistance_down"`
	Pause          string `mapstructure:"pause"`
	ToggleView     string `mapstructure:"toggle_view"`
}

// Load reads config from file with defaults
func Load(configDir string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(configDir)

	// Set defaults
	setDefaults(v)

	// Try to read config file (ignore if not found)
	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Bike defaults
	v.SetDefault("bike.preset", "road-2x11")
	v.SetDefault("bike.chainrings", []int{50, 34})
	v.SetDefault("bike.cassette", []int{11, 12, 13, 14, 15, 17, 19, 21, 24, 28})
	v.SetDefault("bike.wheel_circumference", 2.1)
	v.SetDefault("bike.rider_weight", 75.0)

	// Display defaults
	v.SetDefault("display.graph_window_minutes", 5)
	v.SetDefault("display.climb_gradient_threshold", 3.0)
	v.SetDefault("display.climb_elevation_threshold", 30.0)

	// Controls defaults
	v.SetDefault("controls.shift_up", "Up")
	v.SetDefault("controls.shift_down", "Down")
	v.SetDefault("controls.resistance_up", "Right")
	v.SetDefault("controls.resistance_down", "Left")
	v.SetDefault("controls.pause", "Space")
	v.SetDefault("controls.toggle_view", "Tab")
}

// DefaultConfigDir returns the default config directory
func DefaultConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "goc")
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config loading with defaults"
```

---

### Task 2.2: Config Save

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

**Step 1: Write the failing test**

Add to `config_test.go`:

```go
func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		Trainer: TrainerConfig{DeviceID: "AA:BB:CC:DD:EE:FF"},
		Bike: BikeConfig{
			Chainrings:         []int{52, 36},
			Cassette:           []int{11, 13, 15, 17, 19, 21, 23, 25},
			WheelCircumference: 2.1,
			RiderWeight:        80.0,
		},
	}

	err := Save(cfg, tmpDir)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filepath.Join(tmpDir, "config.toml"))
	require.NoError(t, err)

	// Load it back
	loaded, err := Load(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", loaded.Trainer.DeviceID)
	assert.Equal(t, []int{52, 36}, loaded.Bike.Chainrings)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/... -v`
Expected: FAIL - Save undefined

**Step 3: Write implementation**

Add to `config.go`:

```go
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Save writes config to file
func Save(cfg *Config, configDir string) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	v := viper.New()
	v.SetConfigType("toml")

	v.Set("trainer.device_id", cfg.Trainer.DeviceID)
	v.Set("shifter.device_id", cfg.Shifter.DeviceID)
	v.Set("bike.preset", cfg.Bike.Preset)
	v.Set("bike.chainrings", cfg.Bike.Chainrings)
	v.Set("bike.cassette", cfg.Bike.Cassette)
	v.Set("bike.wheel_circumference", cfg.Bike.WheelCircumference)
	v.Set("bike.rider_weight", cfg.Bike.RiderWeight)
	v.Set("display.graph_window_minutes", cfg.Display.GraphWindowMinutes)
	v.Set("display.climb_gradient_threshold", cfg.Display.ClimbGradientThreshold)
	v.Set("display.climb_elevation_threshold", cfg.Display.ClimbElevationThreshold)
	v.Set("controls.shift_up", cfg.Controls.ShiftUp)
	v.Set("controls.shift_down", cfg.Controls.ShiftDown)
	v.Set("controls.resistance_up", cfg.Controls.ResistanceUp)
	v.Set("controls.resistance_down", cfg.Controls.ResistanceDown)
	v.Set("controls.pause", cfg.Controls.Pause)
	v.Set("controls.toggle_view", cfg.Controls.ToggleView)

	configPath := filepath.Join(configDir, "config.toml")
	return v.WriteConfigAs(configPath)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config save functionality"
```

---

## Phase 3: Simulation Engine

### Task 3.1: Gear System

**Files:**
- Modify: `internal/simulation/simulation.go`
- Create: `internal/simulation/gears.go`
- Create: `internal/simulation/gears_test.go`

**Step 1: Write the failing test**

Create `internal/simulation/gears_test.go`:

```go
package simulation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGearSystem_Ratio(t *testing.T) {
	gs := NewGearSystem([]int{50, 34}, []int{11, 13, 15, 17, 19, 21, 24, 28})

	// 50/11 = 4.545...
	gs.SetFront(0)
	gs.SetRear(0)
	assert.InDelta(t, 4.545, gs.Ratio(), 0.01)

	// 34/28 = 1.214...
	gs.SetFront(1)
	gs.SetRear(7)
	assert.InDelta(t, 1.214, gs.Ratio(), 0.01)
}

func TestGearSystem_ShiftUp(t *testing.T) {
	gs := NewGearSystem([]int{50, 34}, []int{11, 13, 15, 17, 19})
	gs.SetFront(0)
	gs.SetRear(2) // 50/15

	gs.ShiftUp()
	assert.Equal(t, 1, gs.RearIndex()) // 50/13 - harder gear

	// At limit
	gs.SetRear(0)
	gs.ShiftUp()
	assert.Equal(t, 0, gs.RearIndex()) // stays at 0
}

func TestGearSystem_ShiftDown(t *testing.T) {
	gs := NewGearSystem([]int{50, 34}, []int{11, 13, 15, 17, 19})
	gs.SetFront(0)
	gs.SetRear(2) // 50/15

	gs.ShiftDown()
	assert.Equal(t, 3, gs.RearIndex()) // 50/17 - easier gear
}

func TestGearSystem_String(t *testing.T) {
	gs := NewGearSystem([]int{50, 34}, []int{11, 13, 15, 17})
	gs.SetFront(0)
	gs.SetRear(2)

	assert.Equal(t, "50x15", gs.String())
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/simulation/... -v`
Expected: FAIL

**Step 3: Write implementation**

Create `internal/simulation/gears.go`:

```go
package simulation

import "fmt"

// GearSystem manages virtual drivetrain
type GearSystem struct {
	chainrings []int
	cassette   []int
	frontIndex int
	rearIndex  int
}

// NewGearSystem creates a gear system with given chainrings and cassette
func NewGearSystem(chainrings, cassette []int) *GearSystem {
	return &GearSystem{
		chainrings: chainrings,
		cassette:   cassette,
		frontIndex: 0,
		rearIndex:  len(cassette) / 2, // Start in middle
	}
}

// Ratio returns current gear ratio (chainring / cog)
func (g *GearSystem) Ratio() float64 {
	return float64(g.chainrings[g.frontIndex]) / float64(g.cassette[g.rearIndex])
}

// SetFront sets the front chainring index
func (g *GearSystem) SetFront(index int) {
	if index >= 0 && index < len(g.chainrings) {
		g.frontIndex = index
	}
}

// SetRear sets the rear cassette index
func (g *GearSystem) SetRear(index int) {
	if index >= 0 && index < len(g.cassette) {
		g.rearIndex = index
	}
}

// ShiftUp shifts to a harder gear (smaller cog)
func (g *GearSystem) ShiftUp() {
	if g.rearIndex > 0 {
		g.rearIndex--
	}
}

// ShiftDown shifts to an easier gear (larger cog)
func (g *GearSystem) ShiftDown() {
	if g.rearIndex < len(g.cassette)-1 {
		g.rearIndex++
	}
}

// FrontIndex returns current front chainring index
func (g *GearSystem) FrontIndex() int {
	return g.frontIndex
}

// RearIndex returns current rear cassette index
func (g *GearSystem) RearIndex() int {
	return g.rearIndex
}

// String returns human-readable gear (e.g., "50x17")
func (g *GearSystem) String() string {
	return fmt.Sprintf("%dx%d", g.chainrings[g.frontIndex], g.cassette[g.rearIndex])
}

// Chainring returns current chainring teeth
func (g *GearSystem) Chainring() int {
	return g.chainrings[g.frontIndex]
}

// Cog returns current cassette cog teeth
func (g *GearSystem) Cog() int {
	return g.cassette[g.rearIndex]
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/simulation/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/simulation/
git commit -m "feat: add gear system with shifting"
```

---

### Task 3.2: Speed Calculation

**Files:**
- Create: `internal/simulation/physics.go`
- Create: `internal/simulation/physics_test.go`

**Step 1: Write the failing test**

Create `internal/simulation/physics_test.go`:

```go
package simulation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateSpeed(t *testing.T) {
	// 90 RPM, gear ratio 2.94, wheel 2.1m
	// speed = 90 * 2.94 * 2.1 * 60 / 1000 = 33.34 km/h
	speed := CalculateSpeed(90, 2.94, 2.1)
	assert.InDelta(t, 33.34, speed, 0.1)
}

func TestCalculateSpeed_Zero(t *testing.T) {
	speed := CalculateSpeed(0, 2.94, 2.1)
	assert.Equal(t, 0.0, speed)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/simulation/... -v`
Expected: FAIL

**Step 3: Write implementation**

Create `internal/simulation/physics.go`:

```go
package simulation

// CalculateSpeed computes speed in km/h from cadence, gear ratio, and wheel circumference
// cadence: RPM
// gearRatio: chainring/cog
// wheelCircumference: meters
func CalculateSpeed(cadence, gearRatio, wheelCircumference float64) float64 {
	if cadence <= 0 {
		return 0
	}
	// distance per minute = cadence * gearRatio * wheelCircumference (meters)
	// speed km/h = distance per minute * 60 / 1000
	return cadence * gearRatio * wheelCircumference * 60 / 1000
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/simulation/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/simulation/
git commit -m "feat: add speed calculation"
```

---

### Task 3.3: Resistance Calculation

**Files:**
- Modify: `internal/simulation/physics.go`
- Modify: `internal/simulation/physics_test.go`

**Step 1: Write the failing test**

Add to `physics_test.go`:

```go
func TestCalculateResistance_Flat(t *testing.T) {
	// Flat ground, 30 km/h, 75kg rider
	resistance := CalculateResistance(30, 0, 75)
	// Should be moderate resistance from air/rolling
	assert.Greater(t, resistance, 0.0)
	assert.Less(t, resistance, 50.0) // FTMS resistance is 0-100 scale
}

func TestCalculateResistance_Climb(t *testing.T) {
	// 5% climb should increase resistance significantly
	resistanceFlat := CalculateResistance(20, 0, 75)
	resistanceClimb := CalculateResistance(20, 5, 75)

	assert.Greater(t, resistanceClimb, resistanceFlat)
}

func TestCalculateResistance_Descent(t *testing.T) {
	// Descent should reduce resistance
	resistanceFlat := CalculateResistance(30, 0, 75)
	resistanceDescent := CalculateResistance(30, -5, 75)

	assert.Less(t, resistanceDescent, resistanceFlat)
}

func TestCalculateResistance_Clamped(t *testing.T) {
	// Extreme values should be clamped to 0-100
	resistanceSteep := CalculateResistance(5, 20, 100)
	assert.LessOrEqual(t, resistanceSteep, 100.0)

	resistanceDownhill := CalculateResistance(50, -15, 75)
	assert.GreaterOrEqual(t, resistanceDownhill, 0.0)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/simulation/... -v`
Expected: FAIL

**Step 3: Write implementation**

Add to `physics.go`:

```go
import "math"

// CalculateResistance computes trainer resistance level (0-100) based on
// speed (km/h), gradient (%), and rider weight (kg)
func CalculateResistance(speedKmh, gradientPercent, weightKg float64) float64 {
	// Base resistance from rolling resistance and air drag
	// Simplified model: quadratic with speed
	airResistance := 0.005 * speedKmh * speedKmh // increases with speed squared
	rollingResistance := 2.0                      // constant base

	// Gradient contribution
	// At 10% grade, adds significant resistance
	// gravity component: weight * sin(angle) ≈ weight * gradient/100 for small angles
	gravityFactor := 0.5 // scaling factor to map to 0-100 range
	gradientResistance := weightKg * (gradientPercent / 100) * gravityFactor

	totalResistance := airResistance + rollingResistance + gradientResistance

	// Clamp to 0-100 range (FTMS resistance level)
	return math.Max(0, math.Min(100, totalResistance))
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/simulation/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/simulation/
git commit -m "feat: add resistance calculation with gradient"
```

---

### Task 3.4: Simulation Engine Integration

**Files:**
- Modify: `internal/simulation/simulation.go`
- Create: `internal/simulation/simulation_test.go`

**Step 1: Write the failing test**

Create `internal/simulation/simulation_test.go`:

```go
package simulation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_Update(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50, 34},
		Cassette:           []int{11, 13, 15, 17, 19, 21, 24, 28},
		WheelCircumference: 2.1,
		RiderWeight:        75,
	}

	engine := NewEngine(cfg)

	// Simulate pedaling at 90 RPM with 200W
	state := engine.Update(90, 200, 0) // 0% gradient

	assert.Greater(t, state.Speed, 0.0)
	assert.Equal(t, 90.0, state.Cadence)
	assert.Equal(t, 200.0, state.Power)
	assert.Greater(t, state.Resistance, 0.0)
}

func TestEngine_Modes(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50, 34},
		Cassette:           []int{11, 13, 15, 17, 19, 21, 24, 28},
		WheelCircumference: 2.1,
		RiderWeight:        75,
	}

	engine := NewEngine(cfg)

	assert.Equal(t, ModeSIM, engine.Mode())

	engine.SetMode(ModeERG)
	assert.Equal(t, ModeERG, engine.Mode())

	engine.SetTargetPower(250)
	state := engine.Update(90, 200, 0)
	assert.Equal(t, 250.0, state.TargetPower)
}

func TestEngine_Shifting(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50, 34},
		Cassette:           []int{11, 13, 15, 17, 19},
		WheelCircumference: 2.1,
		RiderWeight:        75,
	}

	engine := NewEngine(cfg)
	initialRatio := engine.GearRatio()

	engine.ShiftUp()
	assert.Greater(t, engine.GearRatio(), initialRatio)

	engine.ShiftDown()
	assert.InDelta(t, initialRatio, engine.GearRatio(), 0.01)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/simulation/... -v`
Expected: FAIL

**Step 3: Write implementation**

Replace `internal/simulation/simulation.go`:

```go
package simulation

// Mode represents the training mode
type Mode int

const (
	ModeSIM  Mode = iota // GPX simulation
	ModeERG              // Fixed power
	ModeFREE             // Manual resistance
)

func (m Mode) String() string {
	switch m {
	case ModeSIM:
		return "SIM"
	case ModeERG:
		return "ERG"
	case ModeFREE:
		return "FREE"
	default:
		return "UNKNOWN"
	}
}

// EngineConfig holds simulation parameters
type EngineConfig struct {
	Chainrings         []int
	Cassette           []int
	WheelCircumference float64
	RiderWeight        float64
}

// State represents current simulation state
type State struct {
	Cadence      float64
	Power        float64
	Speed        float64
	Resistance   float64
	Gradient     float64
	GearString   string
	GearRatio    float64
	Mode         Mode
	TargetPower  float64 // For ERG mode
	Distance     float64 // Cumulative meters
	ElapsedTime  float64 // Seconds
}

// Engine handles physics calculations
type Engine struct {
	config          EngineConfig
	gears           *GearSystem
	mode            Mode
	targetPower     float64
	manualResistance float64
	distance        float64
	elapsedTime     float64
}

// NewEngine creates a new simulation engine
func NewEngine(cfg EngineConfig) *Engine {
	return &Engine{
		config:           cfg,
		gears:            NewGearSystem(cfg.Chainrings, cfg.Cassette),
		mode:             ModeSIM,
		manualResistance: 20, // Default for FREE mode
	}
}

// Update calculates new state based on inputs
// cadence: RPM from trainer
// power: Watts from trainer
// gradient: current gradient in percent (from GPX)
func (e *Engine) Update(cadence, power, gradient float64) State {
	speed := CalculateSpeed(cadence, e.gears.Ratio(), e.config.WheelCircumference)

	var resistance float64
	switch e.mode {
	case ModeSIM:
		resistance = CalculateResistance(speed, gradient, e.config.RiderWeight)
	case ModeERG:
		resistance = 0 // ERG mode uses target power, not resistance
	case ModeFREE:
		resistance = e.manualResistance
	}

	return State{
		Cadence:     cadence,
		Power:       power,
		Speed:       speed,
		Resistance:  resistance,
		Gradient:    gradient,
		GearString:  e.gears.String(),
		GearRatio:   e.gears.Ratio(),
		Mode:        e.mode,
		TargetPower: e.targetPower,
		Distance:    e.distance,
		ElapsedTime: e.elapsedTime,
	}
}

// Tick advances time and distance
func (e *Engine) Tick(deltaSeconds float64, speedKmh float64) {
	e.elapsedTime += deltaSeconds
	e.distance += (speedKmh / 3.6) * deltaSeconds // km/h to m/s
}

// Mode returns current training mode
func (e *Engine) Mode() Mode {
	return e.mode
}

// SetMode changes training mode
func (e *Engine) SetMode(m Mode) {
	e.mode = m
}

// SetTargetPower sets ERG mode target
func (e *Engine) SetTargetPower(watts float64) {
	e.targetPower = watts
}

// SetManualResistance sets FREE mode resistance
func (e *Engine) SetManualResistance(level float64) {
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}
	e.manualResistance = level
}

// AdjustManualResistance changes FREE mode resistance by delta
func (e *Engine) AdjustManualResistance(delta float64) {
	e.SetManualResistance(e.manualResistance + delta)
}

// ShiftUp shifts to harder gear
func (e *Engine) ShiftUp() {
	e.gears.ShiftUp()
}

// ShiftDown shifts to easier gear
func (e *Engine) ShiftDown() {
	e.gears.ShiftDown()
}

// GearRatio returns current gear ratio
func (e *Engine) GearRatio() float64 {
	return e.gears.Ratio()
}

// GearString returns current gear as string
func (e *Engine) GearString() string {
	return e.gears.String()
}

// Reset clears distance and time
func (e *Engine) Reset() {
	e.distance = 0
	e.elapsedTime = 0
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/simulation/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/simulation/
git commit -m "feat: integrate simulation engine with modes"
```

---

## Phase 4: GPX Handling

### Task 4.1: GPX Loading

**Files:**
- Modify: `internal/gpx/gpx.go`
- Create: `internal/gpx/gpx_test.go`
- Create: `testdata/simple.gpx`

**Step 1: Create test GPX file**

Create `testdata/simple.gpx`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="test">
  <trk>
    <name>Test Route</name>
    <trkseg>
      <trkpt lat="45.0" lon="7.0"><ele>100</ele></trkpt>
      <trkpt lat="45.001" lon="7.0"><ele>110</ele></trkpt>
      <trkpt lat="45.002" lon="7.0"><ele>115</ele></trkpt>
      <trkpt lat="45.003" lon="7.0"><ele>100</ele></trkpt>
    </trkseg>
  </trk>
</gpx>
```

**Step 2: Write the failing test**

Create `internal/gpx/gpx_test.go`:

```go
package gpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	route, err := Load("../../testdata/simple.gpx")
	require.NoError(t, err)

	assert.Equal(t, "Test Route", route.Name)
	assert.Equal(t, 4, len(route.Points))
	assert.Greater(t, route.TotalDistance, 0.0)
}

func TestRoute_GradientAt(t *testing.T) {
	route, err := Load("../../testdata/simple.gpx")
	require.NoError(t, err)

	// First segment goes uphill (100 -> 110m)
	gradient := route.GradientAt(50) // 50m into ride
	assert.Greater(t, gradient, 0.0)

	// Last segment goes downhill (115 -> 100m)
	gradient = route.GradientAt(route.TotalDistance - 10)
	assert.Less(t, gradient, 0.0)
}

func TestRoute_ElevationAt(t *testing.T) {
	route, err := Load("../../testdata/simple.gpx")
	require.NoError(t, err)

	// Start elevation
	ele := route.ElevationAt(0)
	assert.InDelta(t, 100, ele, 1)
}
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/gpx/... -v`
Expected: FAIL

**Step 4: Write implementation**

Replace `internal/gpx/gpx.go`:

```go
package gpx

import (
	"math"

	"github.com/tkrajina/gpxgo/gpx"
)

// Point represents a track point with distance
type Point struct {
	Lat       float64
	Lon       float64
	Elevation float64
	Distance  float64 // Cumulative distance from start in meters
}

// Route represents a loaded GPX route
type Route struct {
	Name          string
	Points        []Point
	TotalDistance float64
	TotalAscent   float64
	TotalDescent  float64
}

// Load parses a GPX file
func Load(path string) (*Route, error) {
	gpxFile, err := gpx.ParseFile(path)
	if err != nil {
		return nil, err
	}

	route := &Route{}

	// Get track name
	if len(gpxFile.Tracks) > 0 {
		route.Name = gpxFile.Tracks[0].Name
	}

	// Collect all points
	var cumDistance float64
	var prevPoint *gpx.GPXPoint

	for _, track := range gpxFile.Tracks {
		for _, segment := range track.Segments {
			for i, pt := range segment.Points {
				if i > 0 && prevPoint != nil {
					dist := haversineDistance(
						prevPoint.Latitude, prevPoint.Longitude,
						pt.Latitude, pt.Longitude,
					)
					cumDistance += dist

					eleDiff := pt.Elevation.Value() - prevPoint.Elevation.Value()
					if eleDiff > 0 {
						route.TotalAscent += eleDiff
					} else {
						route.TotalDescent += -eleDiff
					}
				}

				route.Points = append(route.Points, Point{
					Lat:       pt.Latitude,
					Lon:       pt.Longitude,
					Elevation: pt.Elevation.Value(),
					Distance:  cumDistance,
				})

				prevPoint = &pt
			}
		}
	}

	route.TotalDistance = cumDistance
	return route, nil
}

// GradientAt returns gradient (%) at given distance
func (r *Route) GradientAt(distance float64) float64 {
	if len(r.Points) < 2 {
		return 0
	}

	// Find segment containing this distance
	for i := 1; i < len(r.Points); i++ {
		if r.Points[i].Distance >= distance {
			prev := r.Points[i-1]
			curr := r.Points[i]

			segmentDist := curr.Distance - prev.Distance
			if segmentDist == 0 {
				return 0
			}

			elevationChange := curr.Elevation - prev.Elevation
			return (elevationChange / segmentDist) * 100
		}
	}

	// Past end, return last segment gradient
	if len(r.Points) >= 2 {
		prev := r.Points[len(r.Points)-2]
		curr := r.Points[len(r.Points)-1]
		segmentDist := curr.Distance - prev.Distance
		if segmentDist > 0 {
			return ((curr.Elevation - prev.Elevation) / segmentDist) * 100
		}
	}

	return 0
}

// ElevationAt returns elevation at given distance
func (r *Route) ElevationAt(distance float64) float64 {
	if len(r.Points) == 0 {
		return 0
	}

	if distance <= 0 {
		return r.Points[0].Elevation
	}

	for i := 1; i < len(r.Points); i++ {
		if r.Points[i].Distance >= distance {
			prev := r.Points[i-1]
			curr := r.Points[i]

			// Interpolate
			segmentDist := curr.Distance - prev.Distance
			if segmentDist == 0 {
				return curr.Elevation
			}

			ratio := (distance - prev.Distance) / segmentDist
			return prev.Elevation + ratio*(curr.Elevation-prev.Elevation)
		}
	}

	return r.Points[len(r.Points)-1].Elevation
}

// PositionAt returns lat/lon at given distance
func (r *Route) PositionAt(distance float64) (lat, lon float64) {
	if len(r.Points) == 0 {
		return 0, 0
	}

	if distance <= 0 {
		return r.Points[0].Lat, r.Points[0].Lon
	}

	for i := 1; i < len(r.Points); i++ {
		if r.Points[i].Distance >= distance {
			prev := r.Points[i-1]
			curr := r.Points[i]

			segmentDist := curr.Distance - prev.Distance
			if segmentDist == 0 {
				return curr.Lat, curr.Lon
			}

			ratio := (distance - prev.Distance) / segmentDist
			return prev.Lat + ratio*(curr.Lat-prev.Lat),
				prev.Lon + ratio*(curr.Lon-prev.Lon)
		}
	}

	last := r.Points[len(r.Points)-1]
	return last.Lat, last.Lon
}

// haversineDistance calculates distance between two points in meters
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
```

**Step 5: Create testdata directory**

Run: `mkdir -p testdata`

**Step 6: Run test to verify it passes**

Run: `go test ./internal/gpx/... -v`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/gpx/ testdata/
git commit -m "feat: add GPX loading and gradient calculation"
```

---

### Task 4.2: Climb Detection

**Files:**
- Modify: `internal/gpx/gpx.go`
- Modify: `internal/gpx/gpx_test.go`

**Step 1: Write the failing test**

Add to `gpx_test.go`:

```go
func TestRoute_DetectClimbs(t *testing.T) {
	route, err := Load("../../testdata/simple.gpx")
	require.NoError(t, err)

	climbs := route.DetectClimbs(3.0, 5) // 3% threshold, 5m elevation threshold

	// Our test route has an uphill section
	assert.GreaterOrEqual(t, len(climbs), 1)

	if len(climbs) > 0 {
		assert.Greater(t, climbs[0].StartDistance, 0.0)
		assert.Greater(t, climbs[0].AverageGradient, 0.0)
	}
}

func TestRoute_IsClimbApproaching(t *testing.T) {
	route, err := Load("../../testdata/simple.gpx")
	require.NoError(t, err)

	// At start, should detect upcoming climb
	approaching, climb := route.IsClimbApproaching(0, 500, 3.0, 5)
	// May or may not be approaching depending on route
	_ = approaching
	_ = climb
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/gpx/... -v`
Expected: FAIL

**Step 3: Write implementation**

Add to `gpx.go`:

```go
// Climb represents a detected climb segment
type Climb struct {
	StartDistance   float64
	EndDistance     float64
	StartElevation  float64
	EndElevation    float64
	AverageGradient float64
	MaxGradient     float64
}

// DetectClimbs finds significant climbs in the route
// gradientThreshold: minimum average gradient (%)
// elevationThreshold: minimum elevation gain (meters)
func (r *Route) DetectClimbs(gradientThreshold, elevationThreshold float64) []Climb {
	if len(r.Points) < 2 {
		return nil
	}

	var climbs []Climb
	var currentClimb *Climb

	for i := 1; i < len(r.Points); i++ {
		prev := r.Points[i-1]
		curr := r.Points[i]

		segmentDist := curr.Distance - prev.Distance
		if segmentDist == 0 {
			continue
		}

		gradient := ((curr.Elevation - prev.Elevation) / segmentDist) * 100

		if gradient >= gradientThreshold {
			if currentClimb == nil {
				currentClimb = &Climb{
					StartDistance:  prev.Distance,
					StartElevation: prev.Elevation,
					MaxGradient:    gradient,
				}
			}
			if gradient > currentClimb.MaxGradient {
				currentClimb.MaxGradient = gradient
			}
			currentClimb.EndDistance = curr.Distance
			currentClimb.EndElevation = curr.Elevation
		} else if currentClimb != nil {
			// End of climb
			elevGain := currentClimb.EndElevation - currentClimb.StartElevation
			if elevGain >= elevationThreshold {
				dist := currentClimb.EndDistance - currentClimb.StartDistance
				currentClimb.AverageGradient = (elevGain / dist) * 100
				climbs = append(climbs, *currentClimb)
			}
			currentClimb = nil
		}
	}

	// Check if route ends in a climb
	if currentClimb != nil {
		elevGain := currentClimb.EndElevation - currentClimb.StartElevation
		if elevGain >= elevationThreshold {
			dist := currentClimb.EndDistance - currentClimb.StartDistance
			currentClimb.AverageGradient = (elevGain / dist) * 100
			climbs = append(climbs, *currentClimb)
		}
	}

	return climbs
}

// IsClimbApproaching checks if a climb starts within lookAhead meters
func (r *Route) IsClimbApproaching(currentDistance, lookAhead, gradientThreshold, elevationThreshold float64) (bool, *Climb) {
	climbs := r.DetectClimbs(gradientThreshold, elevationThreshold)

	for _, climb := range climbs {
		if climb.StartDistance > currentDistance &&
		   climb.StartDistance <= currentDistance+lookAhead {
			return true, &climb
		}
	}

	return false, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/gpx/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/gpx/
git commit -m "feat: add climb detection for auto view toggle"
```

---

## Phase 5: Data Layer

### Task 5.1: Ride Recording

**Files:**
- Modify: `internal/data/data.go`
- Create: `internal/data/ride.go`
- Create: `internal/data/ride_test.go`

**Step 1: Write the failing test**

Create `internal/data/ride_test.go`:

```go
package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRide_AddPoint(t *testing.T) {
	ride := NewRide()

	ride.AddPoint(RidePoint{
		Timestamp: time.Now(),
		Power:     200,
		Cadence:   90,
		Speed:     30.5,
	})

	assert.Equal(t, 1, len(ride.Points))
	assert.Equal(t, 200.0, ride.Points[0].Power)
}

func TestRide_Stats(t *testing.T) {
	ride := NewRide()
	now := time.Now()

	ride.AddPoint(RidePoint{Timestamp: now, Power: 200, Cadence: 90, Speed: 30})
	ride.AddPoint(RidePoint{Timestamp: now.Add(time.Second), Power: 250, Cadence: 95, Speed: 32})
	ride.AddPoint(RidePoint{Timestamp: now.Add(2 * time.Second), Power: 150, Cadence: 85, Speed: 28})

	stats := ride.Stats()

	assert.Equal(t, 200.0, stats.AvgPower)
	assert.Equal(t, 90.0, stats.AvgCadence)
	assert.Equal(t, 30.0, stats.AvgSpeed)
	assert.Equal(t, 250.0, stats.MaxPower)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/data/... -v`
Expected: FAIL

**Step 3: Write implementation**

Create `internal/data/ride.go`:

```go
package data

import (
	"time"
)

// RidePoint represents a single data point during ride
type RidePoint struct {
	Timestamp   time.Time
	Power       float64
	Cadence     float64
	Speed       float64
	Latitude    float64
	Longitude   float64
	Elevation   float64
	Distance    float64
	HeartRate   int // Optional, if HR monitor connected
	Gradient    float64
	GearString  string
}

// RideStats contains computed statistics
type RideStats struct {
	Duration    time.Duration
	Distance    float64 // meters
	AvgPower    float64
	MaxPower    float64
	AvgCadence  float64
	AvgSpeed    float64
	MaxSpeed    float64
	TotalAscent float64
}

// Ride represents a single cycling session
type Ride struct {
	ID        string
	StartTime time.Time
	EndTime   time.Time
	Name      string
	Points    []RidePoint
	GPXName   string // Source GPX file name, if any
	Paused    bool
}

// NewRide creates a new ride recording
func NewRide() *Ride {
	return &Ride{
		ID:        time.Now().Format("2006-01-02-150405"),
		StartTime: time.Now(),
		Points:    make([]RidePoint, 0),
	}
}

// AddPoint records a data point
func (r *Ride) AddPoint(p RidePoint) {
	if !r.Paused {
		r.Points = append(r.Points, p)
	}
}

// Pause pauses recording
func (r *Ride) Pause() {
	r.Paused = true
}

// Resume resumes recording
func (r *Ride) Resume() {
	r.Paused = false
}

// Finish marks ride as complete
func (r *Ride) Finish() {
	r.EndTime = time.Now()
}

// Stats computes ride statistics
func (r *Ride) Stats() RideStats {
	if len(r.Points) == 0 {
		return RideStats{}
	}

	var totalPower, totalCadence, totalSpeed float64
	var maxPower, maxSpeed float64
	var totalAscent float64
	var prevElevation float64

	for i, p := range r.Points {
		totalPower += p.Power
		totalCadence += p.Cadence
		totalSpeed += p.Speed

		if p.Power > maxPower {
			maxPower = p.Power
		}
		if p.Speed > maxSpeed {
			maxSpeed = p.Speed
		}

		if i > 0 && p.Elevation > prevElevation {
			totalAscent += p.Elevation - prevElevation
		}
		prevElevation = p.Elevation
	}

	n := float64(len(r.Points))

	var duration time.Duration
	if !r.EndTime.IsZero() {
		duration = r.EndTime.Sub(r.StartTime)
	} else if len(r.Points) > 0 {
		duration = r.Points[len(r.Points)-1].Timestamp.Sub(r.StartTime)
	}

	var distance float64
	if len(r.Points) > 0 {
		distance = r.Points[len(r.Points)-1].Distance
	}

	return RideStats{
		Duration:    duration,
		Distance:    distance,
		AvgPower:    totalPower / n,
		MaxPower:    maxPower,
		AvgCadence:  totalCadence / n,
		AvgSpeed:    totalSpeed / n,
		MaxSpeed:    maxSpeed,
		TotalAscent: totalAscent,
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/data/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/data/
git commit -m "feat: add ride recording with stats"
```

---

### Task 5.2: FIT Export

**Files:**
- Create: `internal/data/fit.go`
- Create: `internal/data/fit_test.go`

**Step 1: Write the failing test**

Create `internal/data/fit_test.go`:

```go
package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportFIT(t *testing.T) {
	ride := NewRide()
	now := time.Now()

	// Add some points
	for i := 0; i < 10; i++ {
		ride.AddPoint(RidePoint{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Power:     200 + float64(i*10),
			Cadence:   90,
			Speed:     30,
			Latitude:  45.0 + float64(i)*0.001,
			Longitude: 7.0,
			Elevation: 100 + float64(i),
			Distance:  float64(i * 100),
		})
	}
	ride.Finish()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.fit")

	err := ExportFIT(ride, path)
	require.NoError(t, err)

	// Verify file exists
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/data/... -v`
Expected: FAIL

**Step 3: Write implementation**

Create `internal/data/fit.go`:

```go
package data

import (
	"os"
	"time"

	"github.com/tormoder/fit"
)

// ExportFIT writes ride data to a FIT file
func ExportFIT(ride *Ride, path string) error {
	// Create FIT file
	fitFile := new(fit.File)
	fitFile.Type = fit.FileTypeActivity

	// File ID message
	fitFile.FileId = fit.FileIdMsg{
		Type:         fit.FileTypeActivity,
		Manufacturer: fit.ManufacturerDevelopment,
		Product:      1,
		SerialNumber: 1,
		TimeCreated:  ride.StartTime,
	}

	// Create activity
	activity, err := fit.NewActivity()
	if err != nil {
		return err
	}

	// Add records
	for _, p := range ride.Points {
		record := fit.RecordMsg{
			Timestamp:       p.Timestamp,
			Power:           uint16(p.Power),
			Cadence:         uint8(p.Cadence),
			Speed:           uint16(p.Speed * 1000 / 3.6), // m/s * 1000
			Distance:        uint32(p.Distance * 100),     // meters * 100
			Altitude:        uint16((p.Elevation + 500) * 5), // offset + scale per FIT spec
			PositionLat:     degreesToSemicircles(p.Latitude),
			PositionLong:    degreesToSemicircles(p.Longitude),
		}
		activity.Records = append(activity.Records, &record)
	}

	// Session summary
	stats := ride.Stats()
	session := fit.SessionMsg{
		Timestamp:       ride.EndTime,
		StartTime:       ride.StartTime,
		TotalElapsedTime: uint32(stats.Duration.Seconds() * 1000),
		TotalTimerTime:   uint32(stats.Duration.Seconds() * 1000),
		TotalDistance:    uint32(stats.Distance * 100),
		AvgPower:         uint16(stats.AvgPower),
		MaxPower:         uint16(stats.MaxPower),
		AvgCadence:       uint8(stats.AvgCadence),
		AvgSpeed:         uint16(stats.AvgSpeed * 1000 / 3.6),
		MaxSpeed:         uint16(stats.MaxSpeed * 1000 / 3.6),
		TotalAscent:      uint16(stats.TotalAscent),
		Sport:            fit.SportCycling,
		SubSport:         fit.SubSportIndoorCycling,
	}
	activity.Sessions = append(activity.Sessions, &session)

	// Activity summary
	activityMsg := fit.ActivityMsg{
		Timestamp:      ride.EndTime,
		TotalTimerTime: uint32(stats.Duration.Seconds() * 1000),
		NumSessions:    1,
		Type:           fit.ActivityTypeManual,
		Event:          fit.EventActivity,
		EventType:      fit.EventTypeStop,
	}
	activity.Activity = &activityMsg

	fitFile.Activity = activity

	// Write to file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return fit.Encode(f, fitFile, fit.LittleEndian)
}

// degreesToSemicircles converts decimal degrees to FIT semicircles
func degreesToSemicircles(degrees float64) int32 {
	return int32(degrees * (2147483648.0 / 180.0))
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/data/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/data/
git commit -m "feat: add FIT file export"
```

---

### Task 5.3: Ride History Store

**Files:**
- Modify: `internal/data/data.go`
- Create: `internal/data/store_test.go`

**Step 1: Write the failing test**

Create `internal/data/store_test.go`:

```go
package data

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	// Create a ride
	ride := NewRide()
	now := time.Now()
	ride.AddPoint(RidePoint{Timestamp: now, Power: 200, Cadence: 90, Speed: 30})
	ride.AddPoint(RidePoint{Timestamp: now.Add(time.Second), Power: 250, Cadence: 95, Speed: 32})
	ride.Finish()

	// Save it
	err = store.SaveRide(ride)
	require.NoError(t, err)

	// List rides
	rides, err := store.ListRides()
	require.NoError(t, err)
	assert.Equal(t, 1, len(rides))
	assert.Equal(t, ride.ID, rides[0].ID)
}

func TestStore_GetRide(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ride := NewRide()
	ride.AddPoint(RidePoint{Timestamp: time.Now(), Power: 200, Cadence: 90, Speed: 30})
	ride.Finish()

	err = store.SaveRide(ride)
	require.NoError(t, err)

	// Get FIT path
	fitPath := store.GetFITPath(ride.ID)
	assert.True(t, filepath.IsAbs(fitPath))
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/data/... -v`
Expected: FAIL

**Step 3: Write implementation**

Replace `internal/data/data.go`:

```go
package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// RideSummary is a lightweight ride listing
type RideSummary struct {
	ID          string
	StartTime   time.Time
	Duration    time.Duration
	Distance    float64
	AvgPower    float64
	GPXName     string
}

// Store handles ride persistence
type Store struct {
	db      *sql.DB
	dataDir string
}

// NewStore creates a new data store
func NewStore(dataDir string) (*Store, error) {
	// Create directories
	ridesDir := filepath.Join(dataDir, "rides")
	if err := os.MkdirAll(ridesDir, 0755); err != nil {
		return nil, fmt.Errorf("create rides dir: %w", err)
	}

	// Open SQLite database
	dbPath := filepath.Join(dataDir, "history.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("create tables: %w", err)
	}

	return &Store{
		db:      db,
		dataDir: dataDir,
	}, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS rides (
			id TEXT PRIMARY KEY,
			start_time DATETIME,
			end_time DATETIME,
			duration_seconds INTEGER,
			distance_meters REAL,
			avg_power REAL,
			max_power REAL,
			avg_cadence REAL,
			avg_speed REAL,
			total_ascent REAL,
			gpx_name TEXT,
			metadata TEXT
		)
	`)
	return err
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveRide saves a ride to disk and database
func (s *Store) SaveRide(ride *Ride) error {
	stats := ride.Stats()

	// Save FIT file
	fitPath := s.GetFITPath(ride.ID)
	if err := ExportFIT(ride, fitPath); err != nil {
		return fmt.Errorf("export FIT: %w", err)
	}

	// Save JSON metadata (for full reload if needed)
	jsonPath := filepath.Join(s.dataDir, "rides", ride.ID+".json")
	jsonData, err := json.Marshal(ride)
	if err != nil {
		return fmt.Errorf("marshal ride: %w", err)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	// Insert into database
	_, err = s.db.Exec(`
		INSERT INTO rides (id, start_time, end_time, duration_seconds, distance_meters,
			avg_power, max_power, avg_cadence, avg_speed, total_ascent, gpx_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		ride.ID,
		ride.StartTime,
		ride.EndTime,
		int(stats.Duration.Seconds()),
		stats.Distance,
		stats.AvgPower,
		stats.MaxPower,
		stats.AvgCadence,
		stats.AvgSpeed,
		stats.TotalAscent,
		ride.GPXName,
	)

	return err
}

// ListRides returns all rides ordered by date descending
func (s *Store) ListRides() ([]RideSummary, error) {
	rows, err := s.db.Query(`
		SELECT id, start_time, duration_seconds, distance_meters, avg_power, gpx_name
		FROM rides
		ORDER BY start_time DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []RideSummary
	for rows.Next() {
		var r RideSummary
		var durationSec int
		var gpxName sql.NullString

		if err := rows.Scan(&r.ID, &r.StartTime, &durationSec, &r.Distance, &r.AvgPower, &gpxName); err != nil {
			return nil, err
		}

		r.Duration = time.Duration(durationSec) * time.Second
		if gpxName.Valid {
			r.GPXName = gpxName.String
		}

		rides = append(rides, r)
	}

	return rides, rows.Err()
}

// GetFITPath returns the path to a ride's FIT file
func (s *Store) GetFITPath(rideID string) string {
	return filepath.Join(s.dataDir, "rides", rideID+".fit")
}

// DefaultDataDir returns the default data directory
func DefaultDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "goc")
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/data/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/data/
git commit -m "feat: add ride history store with SQLite"
```

---

## Phase 6: Bluetooth (Stub)

Note: Full Bluetooth implementation requires actual hardware for testing. This phase creates a working interface with mock implementation for development.

### Task 6.1: Bluetooth Manager Interface

**Files:**
- Modify: `internal/bluetooth/bluetooth.go`
- Create: `internal/bluetooth/ftms.go`
- Create: `internal/bluetooth/mock.go`
- Create: `internal/bluetooth/bluetooth_test.go`

**Step 1: Write the failing test**

Create `internal/bluetooth/bluetooth_test.go`:

```go
package bluetooth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockManager(t *testing.T) {
	mgr := NewMockManager()

	// Connect should succeed
	err := mgr.Connect()
	require.NoError(t, err)
	assert.True(t, mgr.IsConnected())

	// Should receive data
	dataCh := mgr.DataChannel()

	// Wait for some data
	select {
	case data := <-dataCh:
		assert.Greater(t, data.Power, 0.0)
		assert.Greater(t, data.Cadence, 0.0)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for data")
	}

	// Disconnect
	mgr.Disconnect()
	assert.False(t, mgr.IsConnected())
}

func TestMockManager_SetResistance(t *testing.T) {
	mgr := NewMockManager()
	err := mgr.Connect()
	require.NoError(t, err)

	err = mgr.SetResistance(50)
	require.NoError(t, err)

	err = mgr.SetTargetPower(200)
	require.NoError(t, err)

	mgr.Disconnect()
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/bluetooth/... -v`
Expected: FAIL

**Step 3: Write implementation**

Replace `internal/bluetooth/bluetooth.go`:

```go
package bluetooth

// TrainerData represents data received from trainer
type TrainerData struct {
	Power   float64
	Cadence float64
}

// ShiftEvent represents a shift button press
type ShiftEvent int

const (
	ShiftUp ShiftEvent = iota
	ShiftDown
)

// Manager defines the interface for Bluetooth communication
type Manager interface {
	// Connect initiates connection to trainer and shifter
	Connect() error

	// Disconnect closes all connections
	Disconnect()

	// IsConnected returns true if trainer is connected
	IsConnected() bool

	// DataChannel returns channel for trainer data updates
	DataChannel() <-chan TrainerData

	// ShiftChannel returns channel for shift events
	ShiftChannel() <-chan ShiftEvent

	// SetResistance sets trainer resistance (0-100)
	SetResistance(level float64) error

	// SetTargetPower sets ERG mode target power
	SetTargetPower(watts float64) error
}
```

Create `internal/bluetooth/ftms.go`:

```go
package bluetooth

// FTMS UUIDs
const (
	FTMSServiceUUID              = "00001826-0000-1000-8000-00805f9b34fb"
	IndoorBikeDataUUID           = "00002ad2-0000-1000-8000-00805f9b34fb"
	FitnessMachineControlPointUUID = "00002ad9-0000-1000-8000-00805f9b34fb"
)

// FTMSManager implements Manager using real Bluetooth
// Note: This is a placeholder - full implementation requires tinygo bluetooth
type FTMSManager struct {
	// Will contain actual BLE adapter, device, characteristics
	connected bool
	dataCh    chan TrainerData
	shiftCh   chan ShiftEvent
}

// NewFTMSManager creates a new FTMS Bluetooth manager
func NewFTMSManager() *FTMSManager {
	return &FTMSManager{
		dataCh:  make(chan TrainerData, 10),
		shiftCh: make(chan ShiftEvent, 10),
	}
}

func (m *FTMSManager) Connect() error {
	// TODO: Implement actual BLE scanning and connection
	// Using tinygo.org/x/bluetooth:
	// 1. Enable adapter
	// 2. Scan for FTMS service
	// 3. Connect to device
	// 4. Discover services/characteristics
	// 5. Subscribe to Indoor Bike Data notifications
	m.connected = true
	return nil
}

func (m *FTMSManager) Disconnect() {
	m.connected = false
	close(m.dataCh)
	close(m.shiftCh)
}

func (m *FTMSManager) IsConnected() bool {
	return m.connected
}

func (m *FTMSManager) DataChannel() <-chan TrainerData {
	return m.dataCh
}

func (m *FTMSManager) ShiftChannel() <-chan ShiftEvent {
	return m.shiftCh
}

func (m *FTMSManager) SetResistance(level float64) error {
	// TODO: Write to Fitness Machine Control Point
	// Opcode 0x04 = Set Target Resistance Level
	return nil
}

func (m *FTMSManager) SetTargetPower(watts float64) error {
	// TODO: Write to Fitness Machine Control Point
	// Opcode 0x05 = Set Target Power
	return nil
}
```

Create `internal/bluetooth/mock.go`:

```go
package bluetooth

import (
	"math/rand"
	"time"
)

// MockManager simulates Bluetooth for development
type MockManager struct {
	connected  bool
	dataCh     chan TrainerData
	shiftCh    chan ShiftEvent
	stopCh     chan struct{}
	resistance float64
	targetPower float64
}

// NewMockManager creates a mock Bluetooth manager
func NewMockManager() *MockManager {
	return &MockManager{
		dataCh:     make(chan TrainerData, 10),
		shiftCh:    make(chan ShiftEvent, 10),
		stopCh:     make(chan struct{}),
		resistance: 20,
	}
}

func (m *MockManager) Connect() error {
	m.connected = true
	go m.generateData()
	return nil
}

func (m *MockManager) Disconnect() {
	if m.connected {
		close(m.stopCh)
		m.connected = false
	}
}

func (m *MockManager) IsConnected() bool {
	return m.connected
}

func (m *MockManager) DataChannel() <-chan TrainerData {
	return m.dataCh
}

func (m *MockManager) ShiftChannel() <-chan ShiftEvent {
	return m.shiftCh
}

func (m *MockManager) SetResistance(level float64) error {
	m.resistance = level
	return nil
}

func (m *MockManager) SetTargetPower(watts float64) error {
	m.targetPower = watts
	return nil
}

// SimulateShift simulates a shift button press (for testing)
func (m *MockManager) SimulateShift(event ShiftEvent) {
	if m.connected {
		m.shiftCh <- event
	}
}

func (m *MockManager) generateData() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	basePower := 150.0
	baseCadence := 85.0

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			// Generate realistic-ish data with some variation
			power := basePower + (m.resistance-20)*2 + (rand.Float64()-0.5)*20
			cadence := baseCadence + (rand.Float64()-0.5)*10

			if m.targetPower > 0 {
				// ERG mode: power tends toward target
				power = m.targetPower + (rand.Float64()-0.5)*10
			}

			select {
			case m.dataCh <- TrainerData{Power: power, Cadence: cadence}:
			default:
				// Channel full, skip
			}
		}
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/bluetooth/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/bluetooth/
git commit -m "feat: add Bluetooth manager with mock implementation"
```

---

## Phase 7: TUI

### Task 7.1: Basic TUI Layout

**Files:**
- Modify: `internal/tui/tui.go`
- Create: `internal/tui/layout.go`

**Step 1: Write basic layout implementation**

Replace `internal/tui/tui.go`:

```go
package tui

import (
	"context"
	"fmt"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
)

// Renderer handles the terminal UI
type Renderer struct {
	terminal   *tcell.Terminal
	container  *container.Container

	// Widgets
	powerChart   *linechart.LineChart
	cadenceChart *linechart.LineChart
	speedChart   *linechart.LineChart
	routeView    *text.Text
	statsView    *text.Text
	statusView   *text.Text

	// Data buffers for charts
	powerData   []float64
	cadenceData []float64
	speedData   []float64
	maxPoints   int

	// Callbacks
	onShiftUp   func()
	onShiftDown func()
	onResUp     func()
	onResDown   func()
	onPause     func()
	onToggleView func()
	onQuit      func()
}

// NewRenderer creates a new TUI renderer
func NewRenderer() (*Renderer, error) {
	t, err := tcell.New()
	if err != nil {
		return nil, fmt.Errorf("create terminal: %w", err)
	}

	r := &Renderer{
		terminal:  t,
		maxPoints: 300, // ~5 minutes at 1 update/sec
	}

	if err := r.createWidgets(); err != nil {
		t.Close()
		return nil, err
	}

	if err := r.createLayout(); err != nil {
		t.Close()
		return nil, err
	}

	return r, nil
}

func (r *Renderer) createWidgets() error {
	var err error

	// Power chart
	r.powerChart, err = linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
	)
	if err != nil {
		return fmt.Errorf("create power chart: %w", err)
	}

	// Cadence chart
	r.cadenceChart, err = linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
	)
	if err != nil {
		return fmt.Errorf("create cadence chart: %w", err)
	}

	// Speed chart
	r.speedChart, err = linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
	)
	if err != nil {
		return fmt.Errorf("create speed chart: %w", err)
	}

	// Route view
	r.routeView, err = text.New(text.WrapAtRunes())
	if err != nil {
		return fmt.Errorf("create route view: %w", err)
	}

	// Stats view
	r.statsView, err = text.New(text.WrapAtRunes())
	if err != nil {
		return fmt.Errorf("create stats view: %w", err)
	}

	// Status view
	r.statusView, err = text.New(text.WrapAtRunes())
	if err != nil {
		return fmt.Errorf("create status view: %w", err)
	}

	return nil
}

func (r *Renderer) createLayout() error {
	var err error

	r.container, err = container.New(
		r.terminal,
		container.Border(linestyle.Light),
		container.BorderTitle(" goc - Indoor Cycling Trainer "),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle(" Route "),
						container.PlaceWidget(r.routeView),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle(" Stats "),
						container.PlaceWidget(r.statsView),
					),
					container.SplitPercent(60),
				),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.BorderTitle(" Power (W) "),
								container.PlaceWidget(r.powerChart),
							),
							container.Bottom(
								container.Border(linestyle.Light),
								container.BorderTitle(" Cadence (rpm) "),
								container.PlaceWidget(r.cadenceChart),
							),
							container.SplitPercent(50),
						),
					),
					container.Bottom(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.BorderTitle(" Speed (km/h) "),
								container.PlaceWidget(r.speedChart),
							),
							container.Bottom(
								container.Border(linestyle.Light),
								container.BorderTitle(" Status "),
								container.PlaceWidget(r.statusView),
							),
							container.SplitPercent(70),
						),
					),
					container.SplitPercent(66),
				),
			),
			container.SplitPercent(40),
		),
	)

	return err
}

// SetCallbacks sets keyboard event callbacks
func (r *Renderer) SetCallbacks(
	onShiftUp, onShiftDown, onResUp, onResDown, onPause, onToggleView, onQuit func(),
) {
	r.onShiftUp = onShiftUp
	r.onShiftDown = onShiftDown
	r.onResUp = onResUp
	r.onResDown = onResDown
	r.onPause = onPause
	r.onToggleView = onToggleView
	r.onQuit = onQuit
}

// Run starts the TUI event loop
func (r *Renderer) Run(ctx context.Context) error {
	keyHandler := func(k *terminalapi.Keyboard) {
		switch k.Key {
		case keyboard.KeyArrowUp:
			if r.onShiftUp != nil {
				r.onShiftUp()
			}
		case keyboard.KeyArrowDown:
			if r.onShiftDown != nil {
				r.onShiftDown()
			}
		case keyboard.KeyArrowRight:
			if r.onResUp != nil {
				r.onResUp()
			}
		case keyboard.KeyArrowLeft:
			if r.onResDown != nil {
				r.onResDown()
			}
		case keyboard.KeySpace:
			if r.onPause != nil {
				r.onPause()
			}
		case keyboard.KeyTab:
			if r.onToggleView != nil {
				r.onToggleView()
			}
		case 'q', 'Q':
			if r.onQuit != nil {
				r.onQuit()
			}
		}
	}

	return termdash.Run(ctx, r.terminal, r.container,
		termdash.KeyboardSubscriber(keyHandler),
		termdash.RedrawInterval(100*time.Millisecond),
	)
}

// Close cleans up terminal
func (r *Renderer) Close() {
	r.terminal.Close()
}

// UpdateMetrics updates the chart data
func (r *Renderer) UpdateMetrics(power, cadence, speed float64) {
	r.powerData = append(r.powerData, power)
	r.cadenceData = append(r.cadenceData, cadence)
	r.speedData = append(r.speedData, speed)

	// Trim to max points
	if len(r.powerData) > r.maxPoints {
		r.powerData = r.powerData[1:]
		r.cadenceData = r.cadenceData[1:]
		r.speedData = r.speedData[1:]
	}

	// Update charts
	r.powerChart.Series("power", r.powerData,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorYellow)))
	r.cadenceChart.Series("cadence", r.cadenceData,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorCyan)))
	r.speedChart.Series("speed", r.speedData,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorGreen)))
}

// UpdateStats updates the stats panel
func (r *Renderer) UpdateStats(elapsed string, distance, avgPower, avgCadence, avgSpeed, elevation float64) {
	r.statsView.Reset()
	r.statsView.Write(fmt.Sprintf(
		"Time:      %s\n"+
		"Distance:  %.1f km\n"+
		"Avg Power: %.0f W\n"+
		"Avg Cad:   %.0f rpm\n"+
		"Avg Speed: %.1f km/h\n"+
		"Elevation: +%.0f m",
		elapsed, distance/1000, avgPower, avgCadence, avgSpeed, elevation,
	))
}

// UpdateStatus updates the status bar
func (r *Renderer) UpdateStatus(gear string, gradient float64, mode string, paused bool) {
	r.statusView.Reset()
	pauseStr := ""
	if paused {
		pauseStr = " [PAUSED]"
	}
	r.statusView.Write(fmt.Sprintf(
		"Gear: %s  Gradient: %+.1f%%  Mode: %s%s\n"+
		"[↑↓] Shift  [←→] Resistance  [Space] Pause  [q] Quit",
		gear, gradient, mode, pauseStr,
	))
}

// UpdateRoute updates the route view
func (r *Renderer) UpdateRoute(content string) {
	r.routeView.Reset()
	r.routeView.Write(content)
}
```

**Step 2: Add missing import**

The code needs terminalapi import. Add to imports:

```go
import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
)
```

**Step 3: Verify build**

Run: `go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/tui/
git commit -m "feat: add TUI layout with charts and panels"
```

---

## Phase 8: Main Application

### Task 8.1: Ride Command

**Files:**
- Modify: `cmd/ride.go`
- Modify: `main.go`

**Step 1: Implement ride command**

Replace `cmd/ride.go`:

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thiemo/goc/internal/bluetooth"
	"github.com/thiemo/goc/internal/config"
	"github.com/thiemo/goc/internal/data"
	"github.com/thiemo/goc/internal/gpx"
	"github.com/thiemo/goc/internal/simulation"
	"github.com/thiemo/goc/internal/tui"
)

// RideOptions configures a ride session
type RideOptions struct {
	GPXPath     string
	ERGWatts    int
	Mock        bool // Use mock Bluetooth for development
}

// Ride starts a cycling session
func Ride(opts RideOptions) error {
	// Load config
	cfg, err := config.Load(config.DefaultConfigDir())
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Create simulation engine
	engine := simulation.NewEngine(simulation.EngineConfig{
		Chainrings:         cfg.Bike.Chainrings,
		Cassette:           cfg.Bike.Cassette,
		WheelCircumference: cfg.Bike.WheelCircumference,
		RiderWeight:        cfg.Bike.RiderWeight,
	})

	// Set mode
	if opts.ERGWatts > 0 {
		engine.SetMode(simulation.ModeERG)
		engine.SetTargetPower(float64(opts.ERGWatts))
	} else if opts.GPXPath == "" {
		engine.SetMode(simulation.ModeFREE)
	}

	// Load GPX if provided
	var route *gpx.Route
	if opts.GPXPath != "" {
		route, err = gpx.Load(opts.GPXPath)
		if err != nil {
			return fmt.Errorf("load GPX: %w", err)
		}
		fmt.Printf("Loaded route: %s (%.1f km)\n", route.Name, route.TotalDistance/1000)
	}

	// Create Bluetooth manager
	var btManager bluetooth.Manager
	if opts.Mock {
		btManager = bluetooth.NewMockManager()
	} else {
		btManager = bluetooth.NewFTMSManager()
	}

	// Connect to trainer
	fmt.Println("Connecting to trainer...")
	if err := btManager.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer btManager.Disconnect()
	fmt.Println("Connected!")

	// Create data store
	store, err := data.NewStore(data.DefaultDataDir())
	if err != nil {
		return fmt.Errorf("create store: %w", err)
	}
	defer store.Close()

	// Create ride recording
	ride := data.NewRide()
	if route != nil {
		ride.GPXName = route.Name
	}

	// Create TUI
	renderer, err := tui.NewRenderer()
	if err != nil {
		return fmt.Errorf("create TUI: %w", err)
	}
	defer renderer.Close()

	// Context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// State
	var (
		paused       bool
		currentDist  float64
		lastUpdate   = time.Now()
		totalPower   float64
		totalCadence float64
		totalSpeed   float64
		pointCount   int
	)

	// Set up callbacks
	renderer.SetCallbacks(
		func() { engine.ShiftUp() },             // Shift up
		func() { engine.ShiftDown() },           // Shift down
		func() { engine.AdjustManualResistance(5) },  // Resistance up
		func() { engine.AdjustManualResistance(-5) }, // Resistance down
		func() { // Pause toggle
			paused = !paused
			if paused {
				ride.Pause()
			} else {
				ride.Resume()
			}
		},
		func() { /* Toggle view */ },
		func() { cancel() }, // Quit
	)

	// Main loop goroutine
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case trainerData := <-btManager.DataChannel():
				now := time.Now()
				dt := now.Sub(lastUpdate).Seconds()
				lastUpdate = now

				// Get gradient from route
				var gradient float64
				if route != nil {
					gradient = route.GradientAt(currentDist)
				}

				// Update simulation
				state := engine.Update(trainerData.Cadence, trainerData.Power, gradient)

				// Update position
				if !paused {
					currentDist += (state.Speed / 3.6) * dt
					engine.Tick(dt, state.Speed)
				}

				// Record point
				var lat, lon, ele float64
				if route != nil {
					lat, lon = route.PositionAt(currentDist)
					ele = route.ElevationAt(currentDist)
				}

				ride.AddPoint(data.RidePoint{
					Timestamp:  now,
					Power:      state.Power,
					Cadence:    state.Cadence,
					Speed:      state.Speed,
					Latitude:   lat,
					Longitude:  lon,
					Elevation:  ele,
					Distance:   currentDist,
					Gradient:   gradient,
					GearString: state.GearString,
				})

				// Update averages
				if !paused {
					totalPower += state.Power
					totalCadence += state.Cadence
					totalSpeed += state.Speed
					pointCount++
				}

				// Update TUI
				renderer.UpdateMetrics(state.Power, state.Cadence, state.Speed)

				elapsed := time.Since(ride.StartTime)
				var avgPower, avgCadence, avgSpeed float64
				if pointCount > 0 {
					avgPower = totalPower / float64(pointCount)
					avgCadence = totalCadence / float64(pointCount)
					avgSpeed = totalSpeed / float64(pointCount)
				}

				renderer.UpdateStats(
					formatDuration(elapsed),
					currentDist,
					avgPower,
					avgCadence,
					avgSpeed,
					ele,
				)

				renderer.UpdateStatus(
					state.GearString,
					gradient,
					state.Mode.String(),
					paused,
				)

				// Send resistance to trainer
				if state.Mode == simulation.ModeSIM || state.Mode == simulation.ModeFREE {
					btManager.SetResistance(state.Resistance)
				} else if state.Mode == simulation.ModeERG {
					btManager.SetTargetPower(state.TargetPower)
				}

			case event := <-btManager.ShiftChannel():
				switch event {
				case bluetooth.ShiftUp:
					engine.ShiftUp()
				case bluetooth.ShiftDown:
					engine.ShiftDown()
				}
			}
		}
	}()

	// Run TUI
	if err := renderer.Run(ctx); err != nil && ctx.Err() == nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Save ride
	ride.Finish()
	if len(ride.Points) > 0 {
		fmt.Println("\nSaving ride...")
		if err := store.SaveRide(ride); err != nil {
			return fmt.Errorf("save ride: %w", err)
		}
		fmt.Printf("Ride saved: %s\n", store.GetFITPath(ride.ID))
	}

	return nil
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
```

**Step 2: Update main.go**

Replace `main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/thiemo/goc/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "ride":
		opts := parseRideArgs(os.Args[2:])
		if err := cmd.Ride(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "history":
		if err := cmd.History(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func parseRideArgs(args []string) cmd.RideOptions {
	opts := cmd.RideOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--gpx":
			if i+1 < len(args) {
				opts.GPXPath = args[i+1]
				i++
			}
		case "--erg":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &opts.ERGWatts)
				i++
			}
		case "--mock":
			opts.Mock = true
		}
	}

	return opts
}

func printUsage() {
	fmt.Println(`goc - Indoor Cycling Trainer TUI

Usage:
  goc ride [options]     Start a cycling session
  goc history            View past rides

Ride Options:
  --gpx <file>          Load GPX route for simulation
  --erg <watts>         ERG mode with target power
  --mock                Use mock Bluetooth (for development)

Controls:
  ↑/↓                   Shift gears
  ←/→                   Adjust resistance (FREE mode)
  Space                 Pause/resume
  Tab                   Toggle route view
  q                     Quit`)
}
```

**Step 3: Verify build**

Run: `go build`
Expected: No errors

**Step 4: Commit**

```bash
git add cmd/ main.go
git commit -m "feat: implement ride command with TUI"
```

---

### Task 8.2: History Command

**Files:**
- Modify: `cmd/history.go`

**Step 1: Implement history command**

Replace `cmd/history.go`:

```go
package cmd

import (
	"fmt"

	"github.com/thiemo/goc/internal/data"
)

// History shows past rides
func History() error {
	store, err := data.NewStore(data.DefaultDataDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer store.Close()

	rides, err := store.ListRides()
	if err != nil {
		return fmt.Errorf("list rides: %w", err)
	}

	if len(rides) == 0 {
		fmt.Println("No rides recorded yet.")
		return nil
	}

	fmt.Println("Past Rides:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("%-20s  %-10s  %-10s  %-10s  %s\n", "Date", "Duration", "Distance", "Avg Power", "Route")
	fmt.Println("─────────────────────────────────────────────────────────────")

	for _, r := range rides {
		routeName := r.GPXName
		if routeName == "" {
			routeName = "Free ride"
		}

		fmt.Printf("%-20s  %-10s  %-10.1f  %-10.0f  %s\n",
			r.StartTime.Format("2006-01-02 15:04"),
			formatDuration(r.Duration),
			r.Distance/1000,
			r.AvgPower,
			routeName,
		)
	}

	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("\nFIT files location: %s/rides/\n", data.DefaultDataDir())

	return nil
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
```

Add missing import:

```go
import (
	"fmt"
	"time"

	"github.com/thiemo/goc/internal/data"
)
```

**Step 2: Verify build**

Run: `go build`
Expected: No errors

**Step 3: Commit**

```bash
git add cmd/history.go
git commit -m "feat: implement history command"
```

---

## Phase 9: Final Integration

### Task 9.1: Run All Tests

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: All tests pass

**Step 2: Fix any failures**

If tests fail, debug and fix issues.

**Step 3: Commit any fixes**

```bash
git add .
git commit -m "fix: resolve test failures"
```

---

### Task 9.2: Test Manual Run

**Step 1: Build and run with mock**

Run:
```bash
go build
./goc ride --mock
```

Expected: TUI launches, shows mock data updating

**Step 2: Test controls**

- Press arrow keys to shift/adjust resistance
- Press space to pause/resume
- Press q to quit

**Step 3: Check ride was saved**

Run: `./goc history`
Expected: Shows the ride you just did

---

### Task 9.3: Final Commit

**Step 1: Clean up**

Run: `go mod tidy`

**Step 2: Final commit**

```bash
git add .
git commit -m "chore: clean up and finalize MVP"
```

---

## Summary

This implementation plan covers the MVP features:
1. **Config** - TOML config with defaults
2. **Simulation** - Gear system, speed/resistance calculation, modes
3. **GPX** - Route loading, gradient calculation, climb detection
4. **Data** - Ride recording, FIT export, SQLite history
5. **Bluetooth** - Interface + mock for development
6. **TUI** - termdash layout with charts
7. **Commands** - ride and history

**Not included in MVP** (future work):
- Real Bluetooth FTMS implementation (requires hardware testing)
- Structured workouts
- TUI config editing
- Top-down map rendering (text placeholder only)
- Elevation profile visualization (text placeholder only)

Each task is designed to be completable in 2-5 minutes with clear verification steps.
