package bluetooth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockManager(t *testing.T) {
	mgr := NewMockManager()

	// Connect should succeed
	err := mgr.Connect()
	require.NoError(t, err)
	assert.True(t, mgr.IsConnected())

	// Should receive data
	dataCh := mgr.DataChannel()

	// Wait for some data
	select {
	case data := <-dataCh:
		assert.Greater(t, data.Power, 0.0)
		assert.Greater(t, data.Cadence, 0.0)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for data")
	}

	// Disconnect
	mgr.Disconnect()
	assert.False(t, mgr.IsConnected())
}

func TestMockManager_SetResistance(t *testing.T) {
	mgr := NewMockManager()
	err := mgr.Connect()
	require.NoError(t, err)

	err = mgr.SetResistance(50)
	require.NoError(t, err)

	err = mgr.SetTargetPower(200)
	require.NoError(t, err)

	mgr.Disconnect()
}
