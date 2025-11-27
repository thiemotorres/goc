package simulation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngine_Update(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50, 34},
		Cassette:           []int{11, 13, 15, 17, 19, 21, 24, 28},
		WheelCircumference: 2.1,
		RiderWeight:        75,
		ResistanceScaling:  0.2,
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
		ResistanceScaling:  0.2,
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
		ResistanceScaling:  0.2,
	}

	engine := NewEngine(cfg)
	initialRatio := engine.GearRatio()

	engine.ShiftUp()
	assert.Greater(t, engine.GearRatio(), initialRatio)

	engine.ShiftDown()
	assert.InDelta(t, initialRatio, engine.GearRatio(), 0.01)
}

func TestEngine_Update_GearAffectsResistance(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50, 34},
		Cassette:           []int{11, 13, 15, 17, 19, 21, 24, 28},
		WheelCircumference: 2.105,
		RiderWeight:        75.0,
		ResistanceScaling:  0.2,
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

func TestEngine_Update_FreeMode_GearAffectsResistance(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50},
		Cassette:           []int{11, 13, 15, 17, 19, 21},
		WheelCircumference: 2.105,
		RiderWeight:        75.0,
		ResistanceScaling:  0.2,
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

func TestEngine_GradientSmoothing_EMA(t *testing.T) {
	cfg := EngineConfig{
		Chainrings:         []int{50},
		Cassette:           []int{11, 13, 15, 17, 19, 21},
		WheelCircumference: 2.105,
		RiderWeight:        75.0,
		ResistanceScaling:  0.2,
		GradientSmoothing:  0.85, // 85% history, 15% new
	}

	engine := NewEngine(cfg)
	engine.SetMode(ModeSIM)

	// Feed sequence of noisy gradients: [0, 5, -2, 6, 1, 4, -1, 5]
	noisyGradients := []float64{0, 5, -2, 6, 1, 4, -1, 5}
	var smoothedValues []float64

	for _, gradient := range noisyGradients {
		state := engine.Update(80, 200, gradient)
		smoothedValues = append(smoothedValues, state.Gradient)
	}

	// Verify smoothing characteristics:
	// 1. Should not jump directly to input values (gradual changes)
	if smoothedValues[1] >= 1.0 {
		t.Errorf("After gradient jump from 0%% to 5%%, smoothed = %.2f, should be < 1.0 (gradual increase)",
			smoothedValues[1])
	}

	// 2. Should be smoothed, not matching raw input
	// After 8 updates with varying inputs, smoothed should be around 2-3%, NOT 5%
	finalSmoothed := smoothedValues[len(smoothedValues)-1]
	if finalSmoothed < 1.0 || finalSmoothed > 4.0 {
		t.Errorf("Final smoothed gradient = %.2f, expected 1-4%% range (not matching last raw value of 5%%)",
			finalSmoothed)
	}

	// 3. Verify it's using EMA (not passing through raw values)
	for i, smoothed := range smoothedValues {
		if smoothed == noisyGradients[i] && i > 0 {
			t.Errorf("Smoothed gradient at index %d = %.2f matches raw input exactly, smoothing not applied",
				i, smoothed)
		}
	}

	t.Logf("Noisy gradients: %v", noisyGradients)
	t.Logf("Smoothed values: %v", smoothedValues)
}

func TestEngine_GradientSmoothing_Integration(t *testing.T) {
	engine := NewEngine(EngineConfig{
		Chainrings:         []int{50, 34},
		Cassette:           []int{11, 13, 15, 17, 19, 21, 24, 28},
		WheelCircumference: 2.105,
		RiderWeight:        75.0,
		ResistanceScaling:  0.2,
		GradientSmoothing:  0.85, // Default smoothing
	})
	engine.SetMode(ModeSIM)

	// Phase 1: Flat with noise
	flatNoisy := []float64{0, 2, -1, 1, -2, 0, 1, -1, 0, 2, -1, 1, 0, -2, 1, 0, -1, 2, 0, -1}
	resistanceStart := 0.0
	for i, g := range flatNoisy {
		state := engine.Update(80, 200, g)
		if i == 0 {
			resistanceStart = state.Resistance
		}
	}
	resistanceFlat := engine.Update(80, 200, 0).Resistance

	// Verify: resistance stayed relatively stable (< 10% change)
	change := (resistanceFlat - resistanceStart) / resistanceStart
	if change < 0 {
		change = -change
	}
	if change > 0.10 {
		t.Errorf("Flat terrain resistance changed too much: %.2f%%, expected < 10%%", change*100)
	}

	// Phase 2: Climb (sustained 3%)
	for i := 0; i < 20; i++ {
		engine.Update(80, 200, 3.0)
	}
	resistanceClimb := engine.Update(80, 200, 3.0).Resistance

	// Verify: resistance increased significantly (should be higher)
	if resistanceClimb <= resistanceFlat {
		t.Errorf("Climb resistance (%.2f) not higher than flat (%.2f)", resistanceClimb, resistanceFlat)
	}

	// Phase 3: Return to flat with noise
	for i := 0; i < 20; i++ {
		g := flatNoisy[i%len(flatNoisy)]
		engine.Update(80, 200, g)
	}
	resistanceEnd := engine.Update(80, 200, 0).Resistance

	// Verify: resistance decreased back toward flat level
	if resistanceEnd >= resistanceClimb {
		t.Errorf("After descent, resistance (%.2f) not lower than climb (%.2f)", resistanceEnd, resistanceClimb)
	}

	t.Logf("Resistance progression: flat=%.2f → climb=%.2f → end=%.2f", resistanceFlat, resistanceClimb, resistanceEnd)
}
