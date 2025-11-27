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
