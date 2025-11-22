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
