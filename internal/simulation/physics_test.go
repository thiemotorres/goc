package simulation

import (
	"math"
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

func TestCalculateResistance_Flat(t *testing.T) {
	// Flat ground, 30 km/h, 75kg rider, medium gear ratio
	resistance := CalculateResistance(30, 0, 75, 2.5)
	// Should be moderate resistance from air/rolling
	assert.Greater(t, resistance, 0.0)
	assert.Less(t, resistance, 50.0) // FTMS resistance is 0-100 scale
}

func TestCalculateResistance_Climb(t *testing.T) {
	// 5% climb should increase resistance significantly
	resistanceFlat := CalculateResistance(20, 0, 75, 2.5)
	resistanceClimb := CalculateResistance(20, 5, 75, 2.5)

	assert.Greater(t, resistanceClimb, resistanceFlat)
}

func TestCalculateResistance_Descent(t *testing.T) {
	// Descent should reduce resistance
	resistanceFlat := CalculateResistance(30, 0, 75, 2.5)
	resistanceDescent := CalculateResistance(30, -5, 75, 2.5)

	assert.Less(t, resistanceDescent, resistanceFlat)
}

func TestCalculateResistance_Clamped(t *testing.T) {
	// Extreme values should be clamped to 0-100
	resistanceSteep := CalculateResistance(5, 20, 100, 2.5)
	assert.LessOrEqual(t, resistanceSteep, 100.0)

	resistanceDownhill := CalculateResistance(50, -15, 75, 2.5)
	assert.GreaterOrEqual(t, resistanceDownhill, 0.0)
}

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
			wantMin:         12.0, // rolling + air (~13N)
			wantMax:         14.0,
		},
		{
			name:            "5% climb at 15 km/h",
			speedKmh:        15.0,
			gradientPercent: 5.0,
			weightKg:        75.0,
			wantMin:         48.0, // rolling + air + gradient (~49N)
			wantMax:         50.0,
		},
		{
			name:            "zero speed flat",
			speedKmh:        0.0,
			gradientPercent: 0.0,
			weightKg:        75.0,
			wantMin:         4.0, // only rolling resistance (~4.17N)
			wantMax:         4.5,
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
			wantMin:         3.0,
			wantMax:         5.0,
		},
		{
			name:            "flat road, hard gear (3.0)",
			speedKmh:        30.0,
			gradientPercent: 0.0,
			weightKg:        75.0,
			gearRatio:       3.0,
			wantMin:         9.0,
			wantMax:         11.0,
		},
		{
			name:            "5% climb, medium gear (2.5)",
			speedKmh:        15.0,
			gradientPercent: 5.0,
			weightKg:        75.0,
			gearRatio:       2.5,
			wantMin:         23.0,
			wantMax:         26.0,
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
