package bluetooth

import (
	"math/rand"
	"time"
)

// MockManager simulates Bluetooth for development
type MockManager struct {
	connected   bool
	dataCh      chan TrainerData
	shiftCh     chan ShiftEvent
	stopCh      chan struct{}
	resistance  float64
	targetPower float64
}

// NewMockManager creates a mock Bluetooth manager
func NewMockManager() *MockManager {
	return &MockManager{
		dataCh:     make(chan TrainerData, 10),
		shiftCh:    make(chan ShiftEvent, 10),
		stopCh:     make(chan struct{}),
		resistance: 20,
	}
}

func (m *MockManager) Connect() error {
	m.connected = true
	go m.generateData()
	return nil
}

func (m *MockManager) Disconnect() {
	if m.connected {
		close(m.stopCh)
		m.connected = false
	}
}

func (m *MockManager) IsConnected() bool {
	return m.connected
}

func (m *MockManager) DataChannel() <-chan TrainerData {
	return m.dataCh
}

func (m *MockManager) ShiftChannel() <-chan ShiftEvent {
	return m.shiftCh
}

func (m *MockManager) SetResistance(level float64) error {
	m.resistance = level
	return nil
}

func (m *MockManager) SetTargetPower(watts float64) error {
	m.targetPower = watts
	return nil
}

// SimulateShift simulates a shift button press (for testing)
func (m *MockManager) SimulateShift(event ShiftEvent) {
	if m.connected {
		m.shiftCh <- event
	}
}

func (m *MockManager) generateData() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	basePower := 150.0
	baseCadence := 85.0

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			// Generate realistic-ish data with some variation
			power := basePower + (m.resistance-20)*2 + (rand.Float64()-0.5)*20
			cadence := baseCadence + (rand.Float64()-0.5)*10

			if m.targetPower > 0 {
				// ERG mode: power tends toward target
				power = m.targetPower + (rand.Float64()-0.5)*10
			}

			select {
			case m.dataCh <- TrainerData{Power: power, Cadence: cadence}:
			default:
				// Channel full, skip
			}
		}
	}
}
