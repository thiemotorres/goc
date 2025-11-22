package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportFIT(t *testing.T) {
	ride := NewRide()
	now := time.Now()

	// Add some points
	for i := 0; i < 10; i++ {
		ride.AddPoint(RidePoint{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Power:     200 + float64(i*10),
			Cadence:   90,
			Speed:     30,
			Latitude:  45.0 + float64(i)*0.001,
			Longitude: 7.0,
			Elevation: 100 + float64(i),
			Distance:  float64(i * 100),
		})
	}
	ride.Finish()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.fit")

	err := ExportFIT(ride, path)
	require.NoError(t, err)

	// Verify file exists
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}
