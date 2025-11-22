package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Use temp dir with no config file
	tmpDir := t.TempDir()

	cfg, err := Load(tmpDir)
	require.NoError(t, err)

	// Check defaults
	assert.Equal(t, []int{50, 34}, cfg.Bike.Chainrings)
	assert.Equal(t, []int{11, 12, 13, 14, 15, 17, 19, 21, 24, 28}, cfg.Bike.Cassette)
	assert.Equal(t, 2.1, cfg.Bike.WheelCircumference)
	assert.Equal(t, 75.0, cfg.Bike.RiderWeight)
	assert.Equal(t, 5, cfg.Display.GraphWindowMinutes)
	assert.Equal(t, 3.0, cfg.Display.ClimbGradientThreshold)
}
