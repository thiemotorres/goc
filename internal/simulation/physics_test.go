package simulation

import (
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
