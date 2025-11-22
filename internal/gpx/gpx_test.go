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
