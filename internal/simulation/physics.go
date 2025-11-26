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
