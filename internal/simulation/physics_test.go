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
