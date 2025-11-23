package bluetooth

// TrainerData represents data received from trainer
type TrainerData struct {
	Power   float64
	Cadence float64
}

// ShiftEvent represents a shift button press
type ShiftEvent int

const (
	ShiftUp ShiftEvent = iota
	ShiftDown
)

// Manager defines the interface for Bluetooth communication
type Manager interface {
	// Connect initiates connection to trainer and shifter
	Connect() error

	// Disconnect closes all connections
	Disconnect()

	// IsConnected returns true if trainer is connected
	IsConnected() bool

	// DataChannel returns channel for trainer data updates
	DataChannel() <-chan TrainerData

	// ShiftChannel returns channel for shift events
	ShiftChannel() <-chan ShiftEvent

	// SetResistance sets trainer resistance (0-100)
	SetResistance(level float64) error

	// SetTargetPower sets ERG mode target power
	SetTargetPower(watts float64) error
}

// ConnectionStatus represents BLE connection state
type ConnectionStatus int

const (
	StatusConnecting ConnectionStatus = iota
	StatusConnected
	StatusDisconnected
	StatusReconnecting
)

func (s ConnectionStatus) String() string {
	switch s {
	case StatusConnecting:
		return "Connecting"
	case StatusConnected:
		return "Connected"
	case StatusDisconnected:
		return "Disconnected"
	case StatusReconnecting:
		return "Reconnecting"
	default:
		return "Unknown"
	}
}

// DeviceInfo represents a discovered BLE device
type DeviceInfo struct {
	Address string
	Name    string
	RSSI    int
}
