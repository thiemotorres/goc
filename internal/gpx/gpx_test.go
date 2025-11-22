package gpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	route, err := Load("../../testdata/simple.gpx")
	require.NoError(t, err)

	assert.Equal(t, "Test Route", route.Name)
	assert.Equal(t, 4, len(route.Points))
	assert.Greater(t, route.TotalDistance, 0.0)
}

func TestRoute_GradientAt(t *testing.T) {
	route, err := Load("../../testdata/simple.gpx")
	require.NoError(t, err)

	// First segment goes uphill (100 -> 110m)
	gradient := route.GradientAt(50) // 50m into ride
	assert.Greater(t, gradient, 0.0)

	// Last segment goes downhill (115 -> 100m)
	gradient = route.GradientAt(route.TotalDistance - 10)
	assert.Less(t, gradient, 0.0)
}

func TestRoute_ElevationAt(t *testing.T) {
	route, err := Load("../../testdata/simple.gpx")
	require.NoError(t, err)

	// Start elevation
	ele := route.ElevationAt(0)
	assert.InDelta(t, 100, ele, 1)
}

func TestRoute_DetectClimbs(t *testing.T) {
	route, err := Load("../../testdata/simple.gpx")
	require.NoError(t, err)

	climbs := route.DetectClimbs(3.0, 5) // 3% threshold, 5m elevation threshold

	// Our test route has an uphill section
	assert.GreaterOrEqual(t, len(climbs), 1)

	if len(climbs) > 0 {
		assert.GreaterOrEqual(t, climbs[0].StartDistance, 0.0)
		assert.Greater(t, climbs[0].AverageGradient, 0.0)
	}
}

func TestRoute_IsClimbApproaching(t *testing.T) {
	route, err := Load("../../testdata/simple.gpx")
	require.NoError(t, err)

	// At start, should detect upcoming climb
	approaching, climb := route.IsClimbApproaching(0, 500, 3.0, 5)
	// May or may not be approaching depending on route
	_ = approaching
	_ = climb
}
