package data

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	// Create a ride
	ride := NewRide()
	now := time.Now()
	ride.AddPoint(RidePoint{Timestamp: now, Power: 200, Cadence: 90, Speed: 30})
	ride.AddPoint(RidePoint{Timestamp: now.Add(time.Second), Power: 250, Cadence: 95, Speed: 32})
	ride.Finish()

	// Save it
	err = store.SaveRide(ride)
	require.NoError(t, err)

	// List rides
	rides, err := store.ListRides()
	require.NoError(t, err)
	assert.Equal(t, 1, len(rides))
	assert.Equal(t, ride.ID, rides[0].ID)
}

func TestStore_GetRide(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ride := NewRide()
	ride.AddPoint(RidePoint{Timestamp: time.Now(), Power: 200, Cadence: 90, Speed: 30})
	ride.Finish()

	err = store.SaveRide(ride)
	require.NoError(t, err)

	// Get FIT path
	fitPath := store.GetFITPath(ride.ID)
	assert.True(t, filepath.IsAbs(fitPath))
}
