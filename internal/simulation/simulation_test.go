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
