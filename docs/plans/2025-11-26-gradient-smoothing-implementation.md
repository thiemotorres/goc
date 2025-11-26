# Gradient Smoothing Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Eliminate jerky resistance changes from GPS noise using exponential moving average gradient smoothing.

**Architecture:** Add EMA smoothing to Engine.Update(), making smoothed gradient drive resistance calculation instead of raw gradient. Configurable via bike.gradient_smoothing parameter (default 0.85).

**Tech Stack:** Go, standard library, existing test framework

---

## Context

**Problem:** GPS elevation data has ±5-10m accuracy, causing resistance changes every 1-2 seconds on "flat" terrain due to point-to-point gradient noise.

**Solution:** Apply exponential moving average (EMA) smoothing: `smoothed = α × previous + (1 - α) × new` where α = 0.85 default.

**Design Document:** `docs/plans/2025-11-26-gradient-smoothing-design.md`

---

## Task 1: Add config field for gradient smoothing

**Files:**
- Modify: `internal/config/config.go`

### Step 1: Write failing test for config default

Add to `internal/config/config_test.go`:

```go
func TestLoadConfig_GradientSmoothingDefault(t *testing.T) {
	dir := t.TempDir()

	cfg, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("LoadOrCreate failed: %v", err)
	}

	// Should have default gradient smoothing value
	if cfg.Bike.GradientSmoothing != 0.85 {
		t.Errorf("GradientSmoothing = %.2f, want 0.85", cfg.Bike.GradientSmoothing)
	}
}
```

### Step 2: Run test to verify it fails

```bash
cd /home/thiemo/src/tries/2025-11-22-new-cctui/.worktrees/gradient-smoothing
go test ./internal/config -run TestLoadConfig_GradientSmoothingDefault -v
```

**Expected output:**
```
cfg.Bike.GradientSmoothing undefined (type BikeConfig has no field or method GradientSmoothing)
FAIL
```

### Step 3: Add GradientSmoothing field to BikeConfig

Modify `internal/config/config.go` (around line 46, in BikeConfig struct):

```go
type BikeConfig struct {
	Chainrings         []int   `mapstructure:"chainrings"`
	Cassette           []int   `mapstructure:"cassette"`
	WheelCircumference float64 `mapstructure:"wheel_circumference"`
	RiderWeight        float64 `mapstructure:"rider_weight"`
	ResistanceScaling  float64 `mapstructure:"resistance_scaling"`
	GradientSmoothing  float64 `mapstructure:"gradient_smoothing"` // NEW
}
```

### Step 4: Set default value

Modify `internal/config/config.go` in `setDefaults()` function (around line 100):

```go
func setDefaults(v *viper.Viper) {
	// ... existing defaults ...
	v.SetDefault("bike.resistance_scaling", 0.2)
	v.SetDefault("bike.gradient_smoothing", 0.85) // NEW
}
```

### Step 5: Add to Save function

Modify `internal/config/config.go` in `Save()` function (around line 145):

```go
v.Set("bike.resistance_scaling", cfg.Bike.ResistanceScaling)
v.Set("bike.gradient_smoothing", cfg.Bike.GradientSmoothing) // NEW
```

### Step 6: Run test to verify it passes

```bash
go test ./internal/config -run TestLoadConfig_GradientSmoothingDefault -v
```

**Expected output:**
```
PASS: TestLoadConfig_GradientSmoothingDefault (0.00s)
ok
```

### Step 7: Commit

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: add gradient_smoothing config field with 0.85 default"
```

---

## Task 2: Add smoothing fields to Engine

**Files:**
- Modify: `internal/simulation/simulation.go`
- Modify: `internal/simulation/simulation_test.go`

### Step 1: Write failing test for smoothing initialization

Add to `internal/simulation/simulation_test.go`:

```go
func TestEngine_GradientSmoothingInitialization(t *testing.T) {
	tests := []struct {
		name              string
		configSmoothing   float64
		expectedSmoothing float64
	}{
		{
			name:              "default smoothing",
			configSmoothing:   0.0,
			expectedSmoothing: 0.85,
		},
		{
			name:              "custom smoothing",
			configSmoothing:   0.7,
			expectedSmoothing: 0.7,
		},
		{
			name:              "disabled smoothing",
			configSmoothing:   0.0,
			expectedSmoothing: 0.85, // Falls back to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := EngineConfig{
				Chainrings:         []int{50},
				Cassette:           []int{11, 13, 15, 17},
				WheelCircumference: 2.105,
				RiderWeight:        75.0,
				GradientSmoothing:  tt.configSmoothing,
			}

			engine := NewEngine(cfg)

			if engine.smoothingFactor != tt.expectedSmoothing {
				t.Errorf("smoothingFactor = %.2f, want %.2f",
					engine.smoothingFactor, tt.expectedSmoothing)
			}

			if engine.smoothedGradient != 0.0 {
				t.Errorf("smoothedGradient = %.2f, want 0.0", engine.smoothedGradient)
			}
		})
	}
}
```

### Step 2: Run test to verify it fails

```bash
go test ./internal/simulation -run TestEngine_GradientSmoothingInitialization -v
```

**Expected output:**
```
engine.smoothingFactor undefined
FAIL
```

### Step 3: Add fields to EngineConfig

Modify `internal/simulation/simulation.go` in `EngineConfig` struct (around line 30):

```go
type EngineConfig struct {
	Chainrings         []int
	Cassette           []int
	WheelCircumference float64
	RiderWeight        float64
	ResistanceScaling  float64
	GradientSmoothing  float64 // NEW
}
```

### Step 4: Add fields to Engine struct

Modify `internal/simulation/simulation.go` in `Engine` struct (around line 48):

```go
type Engine struct {
	config           EngineConfig
	gears            *GearSystem
	mode             Mode
	targetPower      float64
	manualResistance float64
	distance         float64
	elapsedTime      float64
	smoothedGradient float64 // NEW: EMA-smoothed gradient
	smoothingFactor  float64 // NEW: alpha value for EMA
}
```

### Step 5: Initialize in NewEngine

Modify `internal/simulation/simulation.go` in `NewEngine()` function (around line 60):

```go
func NewEngine(cfg EngineConfig) *Engine {
	// Set default smoothing if not specified
	smoothing := cfg.GradientSmoothing
	if smoothing == 0 {
		smoothing = 0.85
	}

	return &Engine{
		config:           cfg,
		gears:            NewGearSystem(cfg.Chainrings, cfg.Cassette),
		mode:             ModeSIM,
		manualResistance: 20,
		smoothingFactor:  smoothing,     // NEW
		smoothedGradient: 0.0,           // NEW: initialize at flat
	}
}
```

### Step 6: Run test to verify it passes

```bash
go test ./internal/simulation -run TestEngine_GradientSmoothingInitialization -v
```

**Expected output:**
```
PASS: TestEngine_GradientSmoothingInitialization (0.00s)
ok
```

### Step 7: Commit

```bash
git add internal/simulation/simulation.go internal/simulation/simulation_test.go
git commit -m "feat: add gradient smoothing fields to Engine"
```

---

## Task 3: Apply EMA smoothing in Engine.Update

**Files:**
- Modify: `internal/simulation/simulation.go`
- Modify: `internal/simulation/simulation_test.go`

### Step 1: Write test for EMA smoothing behavior

Add to `internal/simulation/simulation_test.go`:

```go
func TestEngine_GradientSmoothing_EMA(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50},
		Cassette:           []int{11, 13, 15, 17, 19, 21},
		WheelCircumference: 2.105,
		RiderWeight:        75.0,
		GradientSmoothing:  0.8, // 80% history, 20% new
	}

	engine := NewEngine(cfg)
	engine.SetMode(ModeSIM)

	// Feed noisy gradients: 0%, 5%, -2%, 6%, 1%, 4%
	noisyGradients := []float64{0, 5, -2, 6, 1, 4}
	var smoothedValues []float64

	for _, gradient := range noisyGradients {
		state := engine.Update(80, 200, gradient)
		smoothedValues = append(smoothedValues, state.Gradient)
	}

	// Verify smoothing characteristics:
	// 1. Should not jump directly to input values
	if smoothedValues[1] >= 4.0 {
		t.Errorf("After gradient jump to 5%%, smoothed = %.2f, should be < 4.0 (gradual increase)",
			smoothedValues[1])
	}

	// 2. Should converge toward steady values
	if smoothedValues[len(smoothedValues)-1] < 2.0 || smoothedValues[len(smoothedValues)-1] > 4.0 {
		t.Errorf("Final smoothed gradient = %.2f, expected convergence to 2-4%%",
			smoothedValues[len(smoothedValues)-1])
	}

	// 3. Should be monotonically increasing when inputs are mostly positive
	increasing := true
	for i := 1; i < len(smoothedValues)-1; i++ {
		if smoothedValues[i] < smoothedValues[i-1]-0.1 { // Allow small decreases
			increasing = false
		}
	}
	if !increasing {
		t.Errorf("Smoothed values should generally increase with positive gradients, got: %v",
			smoothedValues)
	}
}
```

### Step 2: Write test for smoothing disabled

Add to `internal/simulation/simulation_test.go`:

```go
func TestEngine_GradientSmoothing_Disabled(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50},
		Cassette:           []int{11, 13, 15, 17},
		WheelCircumference: 2.105,
		RiderWeight:        75.0,
		GradientSmoothing:  0.0, // Should use 0.85 default, not disable
	}

	engine := NewEngine(cfg)

	// With default 0.85 smoothing, should NOT pass through directly
	state1 := engine.Update(80, 200, 5.0)
	if state1.Gradient >= 4.0 {
		t.Errorf("With default smoothing, gradient = %.2f after one update, should be < 4.0",
			state1.Gradient)
	}
}
```

### Step 3: Run tests to verify they fail

```bash
go test ./internal/simulation -run "TestEngine_GradientSmoothing" -v
```

**Expected output:**
```
FAIL: TestEngine_GradientSmoothing_EMA
FAIL: TestEngine_GradientSmoothing_Disabled
(tests fail because smoothing not implemented yet)
```

### Step 4: Implement EMA smoothing in Engine.Update

Modify `internal/simulation/simulation.go` in `Engine.Update()` method (around line 73):

```go
func (e *Engine) Update(cadence, power, gradient float64) State {
	// Apply exponential moving average to gradient
	e.smoothedGradient = e.smoothingFactor*e.smoothedGradient +
		(1-e.smoothingFactor)*gradient

	speed := CalculateSpeed(cadence, e.gears.Ratio(), e.config.WheelCircumference)

	var resistance float64
	switch e.mode {
	case ModeSIM:
		scaling := e.config.ResistanceScaling
		if scaling == 0 {
			scaling = 0.2
		}
		// Use smoothed gradient instead of raw gradient
		resistance = CalculateResistance(speed, e.smoothedGradient, e.config.RiderWeight, e.gears.Ratio(), scaling)
	case ModeERG:
		resistance = 0
	case ModeFREE:
		baseResistance := e.manualResistance
		const referenceGearRatio = 2.5
		resistance = baseResistance * (e.gears.Ratio() / referenceGearRatio)
		if resistance < 0 {
			resistance = 0
		}
		if resistance > 100 {
			resistance = 100
		}
	}

	return State{
		Cadence:     cadence,
		Power:       power,
		Speed:       speed,
		Resistance:  resistance,
		Gradient:    e.smoothedGradient, // Return smoothed gradient
		GearString:  e.gears.String(),
		GearRatio:   e.gears.Ratio(),
		Mode:        e.mode,
		TargetPower: e.targetPower,
		Distance:    e.distance,
		ElapsedTime: e.elapsedTime,
	}
}
```

### Step 5: Run tests to verify they pass

```bash
go test ./internal/simulation -run "TestEngine_GradientSmoothing" -v
```

**Expected output:**
```
PASS: TestEngine_GradientSmoothing_EMA (0.00s)
PASS: TestEngine_GradientSmoothing_Disabled (0.00s)
ok
```

### Step 6: Run all simulation tests

```bash
go test ./internal/simulation -v
```

**Expected:** All tests pass (existing + new)

### Step 7: Commit

```bash
git add internal/simulation/simulation.go internal/simulation/simulation_test.go
git commit -m "feat: apply EMA smoothing to gradient in Engine.Update"
```

---

## Task 4: Pass config value through session layer

**Files:**
- Modify: `internal/tui/session.go`

### Step 1: Update session to pass GradientSmoothing

Modify `internal/tui/session.go` in `NewRideSession()` function (around line 76):

```go
func NewRideSession(cfg *config.Config, rideType RideType, route *RouteInfo, mock bool) (*RideSession, error) {
	// Create simulation engine
	engine := simulation.NewEngine(simulation.EngineConfig{
		Chainrings:         cfg.Bike.Chainrings,
		Cassette:           cfg.Bike.Cassette,
		WheelCircumference: cfg.Bike.WheelCircumference,
		RiderWeight:        cfg.Bike.RiderWeight,
		ResistanceScaling:  cfg.Bike.ResistanceScaling,
		GradientSmoothing:  cfg.Bike.GradientSmoothing, // NEW
	})
	// ... rest of function
}
```

### Step 2: Verify build succeeds

```bash
go build ./...
```

**Expected:** Clean build, no errors

### Step 3: Run full test suite

```bash
go test ./...
```

**Expected:** All tests pass

### Step 4: Commit

```bash
git add internal/tui/session.go
git commit -m "feat: pass GradientSmoothing config to Engine"
```

---

## Task 5: Add comprehensive integration test

**Files:**
- Modify: `internal/simulation/simulation_test.go`

### Step 1: Write integration test for realistic gradient smoothing

Add to `internal/simulation/simulation_test.go`:

```go
func TestEngine_GradientSmoothing_RealisticScenario(t *testing.T) {
	// Simulate flat terrain with GPS noise, then a real climb
	cfg := EngineConfig{
		Chainrings:         []int{50, 34},
		Cassette:           []int{11, 13, 15, 17, 19, 21, 24, 28},
		WheelCircumference: 2.105,
		RiderWeight:        75.0,
		ResistanceScaling:  0.2,
		GradientSmoothing:  0.85,
	}

	engine := NewEngine(cfg)
	engine.SetMode(ModeSIM)

	// Phase 1: "Flat" terrain with GPS noise (±2% fluctuations)
	flatGradients := []float64{0, 1, -1, 2, -0.5, 1.5, -1, 0.5, -0.5, 1}
	var flatResistances []float64

	for _, g := range flatGradients {
		state := engine.Update(80, 200, g)
		flatResistances = append(flatResistances, state.Resistance)
	}

	// Verify: resistance should be relatively stable (not jumping every update)
	maxChange := 0.0
	for i := 1; i < len(flatResistances); i++ {
		change := flatResistances[i] - flatResistances[i-1]
		if change < 0 {
			change = -change
		}
		if change > maxChange {
			maxChange = change
		}
	}

	if maxChange > 2.0 {
		t.Errorf("On flat terrain with noise, max resistance change = %.2f, want < 2.0 (smooth)",
			maxChange)
	}

	// Phase 2: Real 5% climb (sustained gradient)
	resistanceBeforeClimb := flatResistances[len(flatResistances)-1]
	var climbResistance float64

	for i := 0; i < 20; i++ {
		state := engine.Update(80, 200, 5.0) // Sustained 5% climb
		climbResistance = state.Resistance
	}

	// Verify: resistance should increase significantly after sustained climb
	if climbResistance <= resistanceBeforeClimb+5.0 {
		t.Errorf("After 5%% climb, resistance = %.2f, before = %.2f, increase should be > 5.0",
			climbResistance, resistanceBeforeClimb)
	}

	t.Logf("Flat resistance range: %.2f-%.2f", flatResistances[0], flatResistances[len(flatResistances)-1])
	t.Logf("After climb: %.2f (increase: %.2f)", climbResistance, climbResistance-resistanceBeforeClimb)
}
```

### Step 2: Run test to verify it passes

```bash
go test ./internal/simulation -run TestEngine_GradientSmoothing_RealisticScenario -v
```

**Expected output:**
```
PASS: TestEngine_GradientSmoothing_RealisticScenario (0.00s)
ok
```

### Step 3: Commit

```bash
git add internal/simulation/simulation_test.go
git commit -m "test: add realistic gradient smoothing integration test"
```

---

## Task 6: Update documentation

**Files:**
- Modify: `README.md`

### Step 1: Add gradient_smoothing to README configuration section

Modify `README.md` after the `resistance_scaling` documentation (around line 45):

```markdown
#### gradient_smoothing

**Type:** float
**Default:** 0.85
**Range:** 0.0 - 0.95

Controls gradient smoothing using exponential moving average to eliminate GPS noise.

- 0.0: No smoothing (instant response, may be jerky with noisy GPS data)
- 0.85: Default (smooth, natural feel with ~20-30 second lag)
- 0.95: Very smooth (minimal jitter, ~30+ second lag)

Adjust if gradient changes feel too sudden or too slow to respond. Higher values provide smoother resistance but slower response to real climbs.
```

### Step 2: Verify documentation

```bash
cat README.md | grep -A 10 "gradient_smoothing"
```

**Expected:** Documentation displays correctly

### Step 3: Commit

```bash
git add README.md
git commit -m "docs: document gradient_smoothing configuration option"
```

---

## Task 7: Create smoke test documentation

**Files:**
- Create: `docs/testing/2025-11-26-gradient-smoothing-smoke-test.md`

### Step 1: Create smoke test document

```bash
cat > docs/testing/2025-11-26-gradient-smoothing-smoke-test.md << 'EOF'
# Gradient Smoothing Smoke Test

**Date:** 2025-11-26
**Tester:** [Name]
**Feature:** EMA-based gradient smoothing

## Test Procedure

### Prerequisites
- Build latest version with gradient smoothing
- Have GPX route with known "flat" sections (e.g., daily commute route)
- Mock or real FTMS trainer

### Test 1: Baseline (Smoothing Enabled - Default)

**Config:** Default (gradient_smoothing = 0.85)

**Steps:**
1. Start ride with test GPX route
2. Ride through "flat" section for 2-3 minutes
3. Observe resistance value in UI

**Expected Behavior:**
- [ ] Resistance changes less than once per 5 seconds
- [ ] No rapid fluctuations (e.g., 30 → 45 → 32 → 50)
- [ ] Smooth, gradual transitions
- [ ] Still detects real climbs (resistance increases on hills)

**Actual Result:**
[Describe what happened]

---

### Test 2: Smoothing Disabled (Comparison)

**Config:** Set `gradient_smoothing = 0.0` in config.toml (forces use of 0.85 default)

Note: Cannot actually disable smoothing in current implementation (0.0 uses default). This test verifies default behavior.

**Steps:**
1. Verify default smoothing is applied
2. Ride same section as Test 1

**Expected Behavior:**
- [ ] Should behave same as Test 1 (default smoothing)

---

### Test 3: High Smoothing

**Config:** Set `gradient_smoothing = 0.95` in config.toml

**Steps:**
1. Start ride with test route
2. Ride through section with known climb
3. Observe how quickly resistance responds

**Expected Behavior:**
- [ ] Very smooth (almost no jitter)
- [ ] Climbs detected but with longer lag (~30-40 seconds)
- [ ] Resistance increases gradually, not suddenly

**Actual Result:**
[Describe what happened]

---

### Test 4: Responsive Smoothing

**Config:** Set `gradient_smoothing = 0.5` in config.toml

**Steps:**
1. Start ride with test route
2. Ride through varied terrain

**Expected Behavior:**
- [ ] More responsive than default
- [ ] Still smoother than no smoothing
- [ ] Quick response to climbs (~5-10 seconds)

**Actual Result:**
[Describe what happened]

---

## Issues Found

[List any issues encountered]

## Recommendations

[Any suggestions for improvement]

## Sign-off

- [ ] Gradient smoothing eliminates 1-2 second resistance jitter
- [ ] Real climbs still detected with reasonable lag
- [ ] Configuration works as documented
- [ ] No regressions in other features

**Tester signature:** ________________
**Date:** ________________
EOF
```

### Step 2: Commit

```bash
git add docs/testing/2025-11-26-gradient-smoothing-smoke-test.md
git commit -m "docs: add gradient smoothing smoke test procedure"
```

---

## Task 8: Final verification

**Files:** N/A

### Step 1: Run full test suite

```bash
go test ./... -v
```

**Expected:** All tests pass (46 existing + new gradient smoothing tests)

### Step 2: Run with race detector

```bash
go test ./... -race
```

**Expected:** No race conditions detected

### Step 3: Build binary

```bash
go build -o goc ./cmd
```

**Expected:** Clean build

### Step 4: Verify all commits

```bash
git log --oneline origin/main..HEAD
```

**Expected commits:**
1. feat: add gradient_smoothing config field with 0.85 default
2. feat: add gradient smoothing fields to Engine
3. feat: apply EMA smoothing to gradient in Engine.Update
4. feat: pass GradientSmoothing config to Engine
5. test: add realistic gradient smoothing integration test
6. docs: document gradient_smoothing configuration option
7. docs: add gradient smoothing smoke test procedure

### Step 5: Verify working tree clean

```bash
git status
```

**Expected:** Nothing to commit, working tree clean

---

## Verification Checklist

After completing all tasks:

- [ ] Config field added with 0.85 default
- [ ] Engine fields added (smoothedGradient, smoothingFactor)
- [ ] EMA formula applied in Engine.Update()
- [ ] Session layer passes config value
- [ ] All unit tests pass
- [ ] Integration test passes
- [ ] No race conditions
- [ ] Documentation updated (README)
- [ ] Smoke test procedure created
- [ ] Clean git history with descriptive commits

---

## Manual Testing (After Implementation)

Follow smoke test procedure:
1. Use your daily route GPX (known to be "flat" but noisy)
2. Verify resistance doesn't change every 1-2 seconds
3. Test a route with real climbs to ensure they're still detected
4. Try different smoothing values (0.5, 0.85, 0.95) to feel the difference

---

## Troubleshooting

### Issue: Resistance still changing every 1-2 seconds

**Check:**
- Config loaded correctly (`gradient_smoothing` not 0.0 in config file)
- Engine.Update() using smoothed gradient, not raw gradient
- Smoothing factor applied (log `e.smoothedGradient` vs input gradient)

### Issue: Climbs not detected

**Check:**
- Smoothing factor not too high (> 0.95)
- Climbs long enough to overcome smoothing lag
- Resistance calculation still using smoothed gradient

### Issue: Tests fail

**Check:**
- Test expectations reasonable for smoothing behavior
- Floating point comparison tolerances appropriate
- Integration test uses enough updates to see smoothing effect
