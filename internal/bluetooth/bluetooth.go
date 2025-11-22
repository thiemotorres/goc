package bluetooth

// Manager handles BLE connections to trainer and shifter
type Manager struct{}

// NewManager creates a new Bluetooth manager
func NewManager() *Manager {
	return &Manager{}
}
