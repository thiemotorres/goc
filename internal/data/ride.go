package data

import (
	"time"
)

// RidePoint represents a single data point during ride
type RidePoint struct {
	Timestamp  time.Time
	Power      float64
	Cadence    float64
	Speed      float64
	Latitude   float64
	Longitude  float64
	Elevation  float64
	Distance   float64
	HeartRate  int // Optional, if HR monitor connected
	Gradient   float64
	GearString string
}

// RideStats contains computed statistics
type RideStats struct {
	Duration    time.Duration
	Distance    float64 // meters
	AvgPower    float64
	MaxPower    float64
	AvgCadence  float64
	AvgSpeed    float64
	MaxSpeed    float64
	TotalAscent float64
}

// Ride represents a single cycling session
type Ride struct {
	ID        string
	StartTime time.Time
	EndTime   time.Time
	Name      string
	Points    []RidePoint
	GPXName   string // Source GPX file name, if any
	Paused    bool
}

// NewRide creates a new ride recording
func NewRide() *Ride {
	return &Ride{
		ID:        time.Now().Format("2006-01-02-150405"),
		StartTime: time.Now(),
		Points:    make([]RidePoint, 0),
	}
}

// AddPoint records a data point
func (r *Ride) AddPoint(p RidePoint) {
	if !r.Paused {
		r.Points = append(r.Points, p)
	}
}

// Pause pauses recording
func (r *Ride) Pause() {
	r.Paused = true
}

// Resume resumes recording
func (r *Ride) Resume() {
	r.Paused = false
}

// Finish marks ride as complete
func (r *Ride) Finish() {
	r.EndTime = time.Now()
}

// Stats computes ride statistics
func (r *Ride) Stats() RideStats {
	if len(r.Points) == 0 {
		return RideStats{}
	}

	var totalPower, totalCadence, totalSpeed float64
	var maxPower, maxSpeed float64
	var totalAscent float64
	var prevElevation float64

	for i, p := range r.Points {
		totalPower += p.Power
		totalCadence += p.Cadence
		totalSpeed += p.Speed

		if p.Power > maxPower {
			maxPower = p.Power
		}
		if p.Speed > maxSpeed {
			maxSpeed = p.Speed
		}

		if i > 0 && p.Elevation > prevElevation {
			totalAscent += p.Elevation - prevElevation
		}
		prevElevation = p.Elevation
	}

	n := float64(len(r.Points))

	var duration time.Duration
	if !r.EndTime.IsZero() {
		duration = r.EndTime.Sub(r.StartTime)
	} else if len(r.Points) > 0 {
		duration = r.Points[len(r.Points)-1].Timestamp.Sub(r.StartTime)
	}

	var distance float64
	if len(r.Points) > 0 {
		distance = r.Points[len(r.Points)-1].Distance
	}

	return RideStats{
		Duration:    duration,
		Distance:    distance,
		AvgPower:    totalPower / n,
		MaxPower:    maxPower,
		AvgCadence:  totalCadence / n,
		AvgSpeed:    totalSpeed / n,
		MaxSpeed:    maxSpeed,
		TotalAscent: totalAscent,
	}
}
