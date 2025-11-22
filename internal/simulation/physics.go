package simulation

import "math"

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
