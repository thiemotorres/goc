package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRide_AddPoint(t *testing.T) {
	ride := NewRide()

	ride.AddPoint(RidePoint{
		Timestamp: time.Now(),
		Power:     200,
		Cadence:   90,
		Speed:     30.5,
	})

	assert.Equal(t, 1, len(ride.Points))
	assert.Equal(t, 200.0, ride.Points[0].Power)
}

func TestRide_Stats(t *testing.T) {
	ride := NewRide()
	now := time.Now()

	ride.AddPoint(RidePoint{Timestamp: now, Power: 200, Cadence: 90, Speed: 30})
	ride.AddPoint(RidePoint{Timestamp: now.Add(time.Second), Power: 250, Cadence: 95, Speed: 32})
	ride.AddPoint(RidePoint{Timestamp: now.Add(2 * time.Second), Power: 150, Cadence: 85, Speed: 28})

	stats := ride.Stats()

	assert.Equal(t, 200.0, stats.AvgPower)
	assert.Equal(t, 90.0, stats.AvgCadence)
	assert.Equal(t, 30.0, stats.AvgSpeed)
	assert.Equal(t, 250.0, stats.MaxPower)
}
