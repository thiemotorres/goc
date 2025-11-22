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
