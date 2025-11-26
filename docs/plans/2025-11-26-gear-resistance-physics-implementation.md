# Gear-Based Resistance Physics Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make gear shifting affect felt resistance at pedals using force-based physics model.

**Architecture:** Calculate wheel forces (air drag, rolling, gradient) in Newtons, scale by gear ratio to get pedal force, map to 0-100 resistance level. Replaces current wheel-only resistance model.

**Tech Stack:** Go, standard library math package, existing test framework

---

## Context

**Problem:** Gear shifting changes displayed gear and speed slightly, but doesn't affect felt resistance at pedals.

**Root Cause:** Current `CalculateResistance()` in `internal/simulation/physics.go` only considers wheel forces (air, rolling, gradient), ignoring gear ratio's mechanical disadvantage.

**Solution:** Force-based model where `F_pedal = F_wheel × gear_ratio`, properly reflecting that harder gears require more pedal force.

**Design Document:** `docs/plans/2025-11-26-gear-resistance-physics-design.md`

---

## Task 1: Add force calculation helpers

**Files:**
- Modify: `internal/simulation/physics.go`
- Create: `internal/simulation/physics_test.go` (if doesn't exist) or modify existing

### Step 1: Write failing test for CalculateWheelForce

Add to `internal/simulation/physics_test.go`:

```go
func TestCalculateWheelForce(t *testing.T) {
	tests := []struct {
		name            string
		speedKmh        float64
		gradientPercent float64
		weightKg        float64
		wantMin         float64
		wantMax         float64
	}{
		{
			name:            "flat road at 25 km/h",
			speedKmh:        25.0,
			gradientPercent: 0.0,
			weightKg:        75.0,
			wantMin:         40.0, // ~42N expected (rolling + air)
			wantMax:         45.0,
		},
		{
			name:            "5% climb at 15 km/h",
			speedKmh:        15.0,
			gradientPercent: 5.0,
			weightKg:        75.0,
			wantMin:         45.0, // rolling + air + gradient (~50N)
			wantMax:         55.0,
		},
		{
			name:            "zero speed flat",
			speedKmh:        0.0,
			gradientPercent: 0.0,
			weightKg:        75.0,
			wantMin:         40.0, // only rolling resistance
			wantMax:         45.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			force := CalculateWheelForce(tt.speedKmh, tt.gradientPercent, tt.weightKg)
			if force < tt.wantMin || force > tt.wantMax {
				t.Errorf("CalculateWheelForce() = %.2f, want between %.2f and %.2f",
					force, tt.wantMin, tt.wantMax)
			}
		})
	}
}
```

### Step 2: Run test to verify it fails

```bash
cd /home/thiemo/src/tries/2025-11-22-new-cctui
go test ./internal/simulation -run TestCalculateWheelForce -v
```

**Expected output:**
```
undefined: CalculateWheelForce
FAIL
```

### Step 3: Implement CalculateWheelForce

Add to `internal/simulation/physics.go` (after existing functions):

```go
// CalculateWheelForce computes total resistance force at the wheel in Newtons
// speedKmh: speed in km/h
// gradientPercent: gradient in percent (positive = uphill)
// weightKg: rider weight in kg
func CalculateWheelForce(speedKmh, gradientPercent, weightKg float64) float64 {
	// Convert speed to m/s
	speedMs := speedKmh / 3.6

	// Air drag: F = 0.5 × ρ × Cd × A × v²
	// ρ = 1.225 kg/m³ (air density at sea level)
	// Cd × A ≈ 0.3 (drag coefficient × frontal area for cycling)
	airDrag := 0.5 * 1.225 * 0.3 * speedMs * speedMs

	// Rolling resistance: F = Crr × m × g
	// Crr = 0.005 (rolling coefficient for road tires)
	// m = rider + bike mass (assume 10kg bike)
	// g = 9.81 m/s²
	totalMass := weightKg + 10.0
	rollingForce := 0.005 * totalMass * 9.81

	// Gradient resistance: F = m × g × sin(θ) ≈ m × g × (gradient/100)
	// Using small angle approximation: sin(θ) ≈ tan(θ) = gradient/100
	gradientForce := totalMass * 9.81 * (gradientPercent / 100.0)

	return airDrag + rollingForce + gradientForce
}
```

### Step 4: Run test to verify it passes

```bash
go test ./internal/simulation -run TestCalculateWheelForce -v
```

**Expected output:**
```
PASS: TestCalculateWheelForce (0.00s)
ok
```

### Step 5: Commit

```bash
git add internal/simulation/physics.go internal/simulation/physics_test.go
git commit -m "feat: add wheel force calculation in Newtons"
```

---

## Task 2: Add pedal force calculation

### Step 1: Write failing test for CalculatePedalForce

Add to `internal/simulation/physics_test.go`:

```go
func TestCalculatePedalForce(t *testing.T) {
	tests := []struct {
		name       string
		wheelForce float64
		gearRatio  float64
		want       float64
	}{
		{
			name:       "50N wheel force, 2.5 gear ratio",
			wheelForce: 50.0,
			gearRatio:  2.5,
			want:       125.0,
		},
		{
			name:       "100N wheel force, 3.0 gear ratio",
			wheelForce: 100.0,
			gearRatio:  3.0,
			want:       300.0,
		},
		{
			name:       "50N wheel force, 1.0 gear ratio",
			wheelForce: 50.0,
			gearRatio:  1.0,
			want:       50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculatePedalForce(tt.wheelForce, tt.gearRatio)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("CalculatePedalForce() = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}
```

### Step 2: Run test to verify it fails

```bash
go test ./internal/simulation -run TestCalculatePedalForce -v
```

**Expected output:**
```
undefined: CalculatePedalForce
FAIL
```

### Step 3: Implement CalculatePedalForce

Add to `internal/simulation/physics.go`:

```go
// CalculatePedalForce translates wheel force to pedal force using gear ratio
// wheelForce: resistance force at wheel in Newtons
// gearRatio: current gear ratio (chainring teeth / cog teeth)
// Returns: equivalent force required at pedals in Newtons
//
// Physics: Power is conserved through drivetrain
// P_pedal = P_wheel
// F_pedal × v_pedal = F_wheel × v_wheel
// Since v_wheel = v_pedal × gearRatio
// Therefore: F_pedal = F_wheel × gearRatio
func CalculatePedalForce(wheelForce, gearRatio float64) float64 {
	return wheelForce * gearRatio
}
```

### Step 4: Run test to verify it passes

```bash
go test ./internal/simulation -run TestCalculatePedalForce -v
```

**Expected output:**
```
PASS: TestCalculatePedalForce (0.00s)
ok
```

### Step 5: Commit

```bash
git add internal/simulation/physics.go internal/simulation/physics_test.go
git commit -m "feat: add pedal force calculation with gear ratio"
```

---

## Task 3: Add force-to-resistance mapping

### Step 1: Write failing test for MapForceToResistance

Add to `internal/simulation/physics_test.go`:

```go
func TestMapForceToResistance(t *testing.T) {
	tests := []struct {
		name          string
		pedalForce    float64
		scalingFactor float64
		want          float64
	}{
		{
			name:          "200N with 0.2 scaling",
			pedalForce:    200.0,
			scalingFactor: 0.2,
			want:          40.0,
		},
		{
			name:          "400N with 0.2 scaling",
			pedalForce:    400.0,
			scalingFactor: 0.2,
			want:          80.0,
		},
		{
			name:          "100N with 0.2 scaling",
			pedalForce:    100.0,
			scalingFactor: 0.2,
			want:          20.0,
		},
		{
			name:          "600N with 0.2 scaling (clamps to 100)",
			pedalForce:    600.0,
			scalingFactor: 0.2,
			want:          100.0,
		},
		{
			name:          "negative force (clamps to 0)",
			pedalForce:    -10.0,
			scalingFactor: 0.2,
			want:          0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapForceToResistance(tt.pedalForce, tt.scalingFactor)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("MapForceToResistance() = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}
```

### Step 2: Run test to verify it fails

```bash
go test ./internal/simulation -run TestMapForceToResistance -v
```

**Expected output:**
```
undefined: MapForceToResistance
FAIL
```

### Step 3: Implement MapForceToResistance

Add to `internal/simulation/physics.go`:

```go
// MapForceToResistance converts pedal force (Newtons) to trainer resistance level (0-100)
// pedalForce: force at pedals in Newtons
// scalingFactor: calibration factor (typical: 0.2)
// Returns: resistance level clamped to 0-100
//
// Typical pedal forces:
// - Light effort: 150-200N → 30-40 resistance
// - Moderate effort: 200-300N → 40-60 resistance
// - Hard effort: 300-400N → 60-80 resistance
func MapForceToResistance(pedalForce, scalingFactor float64) float64 {
	resistance := pedalForce * scalingFactor
	return math.Max(0, math.Min(100, resistance))
}
```

### Step 4: Run test to verify it passes

```bash
go test ./internal/simulation -run TestMapForceToResistance -v
```

**Expected output:**
```
PASS: TestMapForceToResistance (0.00s)
ok
```

### Step 5: Commit

```bash
git add internal/simulation/physics.go internal/simulation/physics_test.go
git commit -m "feat: add force-to-resistance mapping with clamping"
```

---

## Task 4: Update CalculateResistance to use force-based model

### Step 1: Write failing test for new CalculateResistance

Add to `internal/simulation/physics_test.go`:

```go
func TestCalculateResistance_WithGearRatio(t *testing.T) {
	tests := []struct {
		name            string
		speedKmh        float64
		gradientPercent float64
		weightKg        float64
		gearRatio       float64
		wantMin         float64
		wantMax         float64
	}{
		{
			name:            "flat road, easy gear (2.0)",
			speedKmh:        20.0,
			gradientPercent: 0.0,
			weightKg:        75.0,
			gearRatio:       2.0,
			wantMin:         15.0,
			wantMax:         20.0,
		},
		{
			name:            "flat road, hard gear (3.0)",
			speedKmh:        30.0,
			gradientPercent: 0.0,
			weightKg:        75.0,
			gearRatio:       3.0,
			wantMin:         25.0,
			wantMax:         35.0,
		},
		{
			name:            "5% climb, medium gear (2.5)",
			speedKmh:        15.0,
			gradientPercent: 5.0,
			weightKg:        75.0,
			gearRatio:       2.5,
			wantMin:         25.0,
			wantMax:         35.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateResistance(tt.speedKmh, tt.gradientPercent, tt.weightKg, tt.gearRatio)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculateResistance() = %.2f, want between %.2f and %.2f",
					got, tt.wantMin, tt.wantMax)
			}
			// Verify clamping
			if got < 0 || got > 100 {
				t.Errorf("CalculateResistance() = %.2f, must be between 0 and 100", got)
			}
		})
	}
}

func TestCalculateResistance_GearRatioEffect(t *testing.T) {
	// Same conditions, different gear ratios
	// Higher gear ratio should = higher resistance
	speedKmh := 25.0
	gradientPercent := 2.0
	weightKg := 75.0

	easyGear := CalculateResistance(speedKmh, gradientPercent, weightKg, 2.0)
	hardGear := CalculateResistance(speedKmh, gradientPercent, weightKg, 3.0)

	if hardGear <= easyGear {
		t.Errorf("Hard gear (3.0) resistance %.2f should be > easy gear (2.0) resistance %.2f",
			hardGear, easyGear)
	}

	// Should be roughly proportional (within 20% of expected ratio)
	expectedRatio := 3.0 / 2.0 // 1.5x
	actualRatio := hardGear / easyGear
	if actualRatio < expectedRatio*0.8 || actualRatio > expectedRatio*1.2 {
		t.Errorf("Resistance ratio %.2f not close to gear ratio %.2f", actualRatio, expectedRatio)
	}
}
```

### Step 2: Run test to verify it fails

```bash
go test ./internal/simulation -run TestCalculateResistance -v
```

**Expected output:**
```
too few arguments in call to CalculateResistance
FAIL
```

### Step 3: Update CalculateResistance implementation

Replace the existing `CalculateResistance` function in `internal/simulation/physics.go`:

```go
// CalculateResistance computes trainer resistance level (0-100) based on
// speed, gradient, rider weight, and gear ratio using force-based physics model
func CalculateResistance(speedKmh, gradientPercent, weightKg, gearRatio float64) float64 {
	// Default scaling factor (can be made configurable later)
	const scalingFactor = 0.2

	// Calculate total resistance force at wheel (Newtons)
	wheelForce := CalculateWheelForce(speedKmh, gradientPercent, weightKg)

	// Apply gear ratio mechanical disadvantage
	pedalForce := CalculatePedalForce(wheelForce, gearRatio)

	// Map to 0-100 resistance scale
	resistance := MapForceToResistance(pedalForce, scalingFactor)

	return resistance
}
```

### Step 4: Run test to verify it passes

```bash
go test ./internal/simulation -run TestCalculateResistance -v
```

**Expected output:**
```
PASS: TestCalculateResistance_WithGearRatio (0.00s)
PASS: TestCalculateResistance_GearRatioEffect (0.00s)
ok
```

### Step 5: Run all physics tests

```bash
go test ./internal/simulation/physics_test.go -v
```

**Expected:** All tests pass

### Step 6: Commit

```bash
git add internal/simulation/physics.go internal/simulation/physics_test.go
git commit -m "feat: update CalculateResistance to use force-based model with gear ratio"
```

---

## Task 5: Update simulation.go to pass gear ratio

### Step 1: Identify call site

Read `internal/simulation/simulation.go` to find where `CalculateResistance` is called.

**Location:** Line 79 in `Engine.Update()` method:
```go
resistance = CalculateResistance(speed, gradient, e.config.RiderWeight)
```

### Step 2: Write test for Engine.Update with gear effect

Check if `internal/simulation/simulation_test.go` exists. If not, create it.

Add test:

```go
func TestEngine_Update_GearAffectsResistance(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50, 34},
		Cassette:           []int{11, 13, 15, 17, 19, 21, 24, 28},
		WheelCircumference: 2.105,
		RiderWeight:        75.0,
	}
	engine := NewEngine(cfg)
	engine.SetMode(ModeSIM)

	cadence := 90.0
	power := 200.0
	gradient := 2.0

	// Get resistance in easy gear
	engine.gears.SetRear(5) // 21t cog, ratio ~2.38
	easyState := engine.Update(cadence, power, gradient)

	// Get resistance in hard gear
	engine.gears.SetRear(1) // 13t cog, ratio ~3.85
	hardState := engine.Update(cadence, power, gradient)

	// Hard gear should have higher resistance
	if hardState.Resistance <= easyState.Resistance {
		t.Errorf("Hard gear resistance %.2f should be > easy gear resistance %.2f",
			hardState.Resistance, easyState.Resistance)
	}

	// Hard gear should have higher speed at same cadence
	if hardState.Speed <= easyState.Speed {
		t.Errorf("Hard gear speed %.2f should be > easy gear speed %.2f",
			hardState.Speed, easyState.Speed)
	}
}
```

### Step 3: Run test to verify it fails

```bash
go test ./internal/simulation -run TestEngine_Update_GearAffectsResistance -v
```

**Expected output:**
```
too few arguments in call to CalculateResistance
FAIL
```

### Step 4: Update Engine.Update to pass gear ratio

Modify `internal/simulation/simulation.go` line 79:

```go
func (e *Engine) Update(cadence, power, gradient float64) State {
	speed := CalculateSpeed(cadence, e.gears.Ratio(), e.config.WheelCircumference)

	var resistance float64
	switch e.mode {
	case ModeSIM:
		resistance = CalculateResistance(speed, gradient, e.config.RiderWeight, e.gears.Ratio())
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
```

**Note:** Only ModeSIM uses the new calculation. ModeFREE keeps manual resistance, ModeERG keeps 0.

### Step 5: Run test to verify it passes

```bash
go test ./internal/simulation -run TestEngine_Update_GearAffectsResistance -v
```

**Expected output:**
```
PASS: TestEngine_Update_GearAffectsResistance (0.00s)
ok
```

### Step 6: Run all simulation tests

```bash
go test ./internal/simulation -v
```

**Expected:** All tests pass

### Step 7: Commit

```bash
git add internal/simulation/simulation.go internal/simulation/simulation_test.go
git commit -m "feat: pass gear ratio to resistance calculation in SIM mode"
```

---

## Task 6: Update FREE mode to use gear-based resistance

**Context:** FREE mode currently uses manual resistance directly. We should apply gear ratio to make gear shifting work in FREE mode too.

### Step 1: Write test for FREE mode gear effect

Add to `internal/simulation/simulation_test.go`:

```go
func TestEngine_Update_FreeMode_GearAffectsResistance(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50},
		Cassette:           []int{11, 13, 15, 17, 19, 21},
		WheelCircumference: 2.105,
		RiderWeight:        75.0,
	}
	engine := NewEngine(cfg)
	engine.SetMode(ModeFREE)
	engine.SetManualResistance(30.0) // Base resistance

	cadence := 80.0
	power := 150.0
	gradient := 0.0 // FREE mode ignores gradient

	// Easy gear
	engine.gears.SetRear(5) // 21t, ratio ~2.38
	easyState := engine.Update(cadence, power, gradient)

	// Hard gear
	engine.gears.SetRear(1) // 13t, ratio ~3.85
	hardState := engine.Update(cadence, power, gradient)

	// Hard gear should have higher resistance
	if hardState.Resistance <= easyState.Resistance {
		t.Errorf("FREE mode: hard gear resistance %.2f should be > easy gear %.2f",
			hardState.Resistance, easyState.Resistance)
	}
}
```

### Step 2: Run test to verify it fails

```bash
go test ./internal/simulation -run TestEngine_Update_FreeMode_GearAffectsResistance -v
```

**Expected output:**
```
FAIL: hard gear resistance 30.00 should be > easy gear 30.00
```

### Step 3: Update FREE mode to apply gear-based scaling

Modify `internal/simulation/simulation.go` in the `Engine.Update` method:

```go
func (e *Engine) Update(cadence, power, gradient float64) State {
	speed := CalculateSpeed(cadence, e.gears.Ratio(), e.config.WheelCircumference)

	var resistance float64
	switch e.mode {
	case ModeSIM:
		resistance = CalculateResistance(speed, gradient, e.config.RiderWeight, e.gears.Ratio())
	case ModeERG:
		resistance = 0 // ERG mode uses target power, not resistance
	case ModeFREE:
		// Apply gear ratio scaling to manual resistance
		// Treat manual resistance as a base wheel force equivalent
		// Scale by gear ratio to get pedal resistance
		baseResistance := e.manualResistance
		// Convert to approximate force, apply gear ratio, convert back
		// Simplified: directly scale by gear ratio relative to reference (2.5)
		const referenceGearRatio = 2.5
		resistance = baseResistance * (e.gears.Ratio() / referenceGearRatio)
		// Clamp to valid range
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
		Gradient:    gradient,
		GearString:  e.gears.String(),
		GearRatio:   e.gears.Ratio(),
		Mode:        e.mode,
		TargetPower: e.targetPower,
		Distance:    e.distance,
		ElapsedTime: e.elapsedTime,
	}
}
```

### Step 4: Run test to verify it passes

```bash
go test ./internal/simulation -run TestEngine_Update_FreeMode_GearAffectsResistance -v
```

**Expected output:**
```
PASS: TestEngine_Update_FreeMode_GearAffectsResistance (0.00s)
ok
```

### Step 5: Run all simulation tests

```bash
go test ./internal/simulation -v
```

**Expected:** All tests pass

### Step 6: Commit

```bash
git add internal/simulation/simulation.go internal/simulation/simulation_test.go
git commit -m "feat: apply gear ratio scaling in FREE mode"
```

---

## Task 7: Integration testing

### Step 1: Build the application

```bash
go build -o goc ./cmd
```

**Expected output:**
```
(success - binary created)
```

### Step 2: Run unit tests across all packages

```bash
go test ./... -v
```

**Expected:** All tests pass

### Step 3: Manual smoke test with mock trainer

```bash
./goc ride --mock
```

**Actions to test:**
1. Start ride in FREE mode
2. Press ↑ (shift up) - gear display should change (e.g., 50x17 → 50x15)
3. Observe resistance value in UI - should increase
4. Press ↓ (shift down) - gear should change (e.g., 50x15 → 50x17)
5. Observe resistance value - should decrease
6. Press q to quit

**Expected behavior:**
- Gear display updates immediately
- Resistance changes proportionally with gear changes
- Speed increases with harder gear (at same mock cadence)

### Step 4: Document smoke test results

Create `docs/testing/2025-11-26-gear-resistance-smoke-test.md`:

```markdown
# Gear Resistance Smoke Test

**Date:** 2025-11-26
**Tester:** [Your Name]
**Mode:** Mock trainer

## Test Results

### FREE Mode
- [ ] Shift up increases resistance
- [ ] Shift down decreases resistance
- [ ] Gear display updates correctly
- [ ] Speed changes with gear

### SIM Mode (if route available)
- [ ] Shift up increases resistance
- [ ] Shift down decreases resistance
- [ ] Gradient + gear both affect resistance
- [ ] Feels realistic on climbs

## Issues Found

[List any issues]

## Notes

[Any observations]
```

### Step 5: Commit

```bash
git add docs/testing/2025-11-26-gear-resistance-smoke-test.md
git commit -m "test: add gear resistance smoke test documentation"
```

---

## Task 8: Optional - Add configurable resistance scaling

**Context:** Allow users to tune the resistance scaling factor without code changes.

### Step 1: Add config field

Modify `internal/config/config.go` to add scaling factor to `BikeConfig`:

```go
type BikeConfig struct {
	Chainrings         []int   `json:"chainrings"`
	Cassette           []int   `json:"cassette"`
	WheelCircumference float64 `json:"wheel_circumference"`
	RiderWeight        float64 `json:"rider_weight"`
	ResistanceScaling  float64 `json:"resistance_scaling"` // New field
}
```

### Step 2: Set default value

In the `LoadOrCreate` function or wherever defaults are set, add:

```go
if cfg.Bike.ResistanceScaling == 0 {
	cfg.Bike.ResistanceScaling = 0.2 // Default scaling
}
```

### Step 3: Pass scaling to physics calculation

Modify `CalculateResistance` signature in `internal/simulation/physics.go`:

```go
func CalculateResistance(speedKmh, gradientPercent, weightKg, gearRatio, scalingFactor float64) float64 {
	// Remove the const, use parameter instead
	wheelForce := CalculateWheelForce(speedKmh, gradientPercent, weightKg)
	pedalForce := CalculatePedalForce(wheelForce, gearRatio)
	resistance := MapForceToResistance(pedalForce, scalingFactor)
	return resistance
}
```

### Step 4: Update Engine to use config scaling

Modify `internal/simulation/simulation.go`:

1. Add field to `EngineConfig`:
```go
type EngineConfig struct {
	Chainrings         []int
	Cassette           []int
	WheelCircumference float64
	RiderWeight        float64
	ResistanceScaling  float64 // New field
}
```

2. Update `Engine.Update` call:
```go
case ModeSIM:
	scaling := e.config.ResistanceScaling
	if scaling == 0 {
		scaling = 0.2 // Fallback default
	}
	resistance = CalculateResistance(speed, gradient, e.config.RiderWeight, e.gears.Ratio(), scaling)
```

### Step 5: Update session.go to pass config

Modify `internal/tui/session.go` in `NewRideSession`:

```go
engine := simulation.NewEngine(simulation.EngineConfig{
	Chainrings:         cfg.Bike.Chainrings,
	Cassette:           cfg.Bike.Cassette,
	WheelCircumference: cfg.Bike.WheelCircumference,
	RiderWeight:        cfg.Bike.RiderWeight,
	ResistanceScaling:  cfg.Bike.ResistanceScaling, // New field
})
```

### Step 6: Update tests to pass scaling factor

Update all calls to `CalculateResistance` in tests to include scaling parameter (use 0.2).

### Step 7: Test and commit

```bash
go test ./...
go build -o goc ./cmd
./goc ride --mock
```

**Test:** Verify gear shifting still works correctly.

```bash
git add internal/config/config.go internal/simulation/physics.go internal/simulation/simulation.go internal/tui/session.go
git commit -m "feat: make resistance scaling configurable"
```

---

## Task 9: Update documentation

### Step 1: Update README

Modify `README.md` to document the new config option:

Add to configuration section:

```markdown
#### resistance_scaling

**Type:** float
**Default:** 0.2
**Range:** 0.1 - 0.5

Controls how resistance force maps to trainer resistance level (0-100).

- Lower values (0.1): Lighter resistance feel
- Default (0.2): Balanced resistance
- Higher values (0.3-0.5): Heavier resistance feel

Adjust if gear shifting feels too easy or too hard.
```

### Step 2: Commit

```bash
git add README.md
git commit -m "docs: document resistance_scaling configuration option"
```

---

## Task 10: Final verification and cleanup

### Step 1: Run full test suite

```bash
go test ./... -v -race
```

**Expected:** All tests pass, no race conditions

### Step 2: Build release binary

```bash
go build -o goc ./cmd
```

**Expected:** Clean build

### Step 3: Review all commits

```bash
git log --oneline -10
```

**Verify:**
- Clear commit messages
- Logical progression
- No WIP or fixup commits

### Step 4: Create summary commit (optional)

If you want to document the full change:

```bash
git commit --allow-empty -m "feat: implement gear-based resistance physics

Summary of changes:
- Add force-based resistance calculation (wheel force → pedal force)
- Apply gear ratio mechanical disadvantage to resistance
- Update SIM and FREE modes to use new physics model
- Add configurable resistance scaling factor
- All tests passing

Fixes issue where gear shifting had no effect on felt resistance.
"
```

---

## Verification Checklist

After completing all tasks:

- [ ] All unit tests pass (`go test ./...`)
- [ ] Application builds cleanly
- [ ] Gear shifting changes displayed gear
- [ ] Gear shifting changes resistance (visible in UI)
- [ ] Higher gears = higher resistance
- [ ] Lower gears = lower resistance
- [ ] SIM mode: gradient + gear both affect resistance
- [ ] FREE mode: gear affects manual resistance
- [ ] ERG mode: unaffected (correct behavior)
- [ ] Config option documented
- [ ] Design document committed
- [ ] All code committed with clear messages

---

## Troubleshooting

### Issue: Resistance doesn't change enough

**Solution:** Adjust `resistance_scaling` in config:
- Increase to 0.3 or 0.4 for more noticeable effect
- Test and document recommended values

### Issue: Resistance changes too much (hits 100 quickly)

**Solution:** Decrease `resistance_scaling` to 0.15 or 0.1

### Issue: Tests fail with "unexpected resistance value"

**Solution:** Check test expectations - new physics model produces different values than old model. Update test ranges if needed.

### Issue: Build fails with "undefined: math"

**Solution:** Add `import "math"` to `internal/simulation/physics.go`

---

## Next Steps (Future Enhancements)

1. **Inertia simulation:** Heavier gears resist cadence changes
2. **Cross-chaining penalty:** Discourage extreme gear combinations (big-big, small-small)
3. **Gear shift smoothing:** Gradual resistance transition over 0.5-1s
4. **Auto-shifting suggestions:** Recommend gear changes based on cadence
5. **Power-based resistance in FREE mode:** Option to maintain target power instead of resistance

---

## Skills Referenced

- @superpowers:test-driven-development - Write test first, see it fail, implement, see it pass
- @superpowers:verification-before-completion - Run tests before claiming complete
- @superpowers:systematic-debugging - If issues arise, investigate root cause first
