package bluetooth

// FTMS UUIDs
const (
	FTMSServiceUUID                = "00001826-0000-1000-8000-00805f9b34fb"
	IndoorBikeDataUUID             = "00002ad2-0000-1000-8000-00805f9b34fb"
	FitnessMachineControlPointUUID = "00002ad9-0000-1000-8000-00805f9b34fb"
)

// FTMSManager implements Manager using real Bluetooth
// Note: This is a placeholder - full implementation requires tinygo bluetooth
type FTMSManager struct {
	// Will contain actual BLE adapter, device, characteristics
	connected bool
	dataCh    chan TrainerData
	shiftCh   chan ShiftEvent
}

// NewFTMSManager creates a new FTMS Bluetooth manager
func NewFTMSManager() *FTMSManager {
	return &FTMSManager{
		dataCh:  make(chan TrainerData, 10),
		shiftCh: make(chan ShiftEvent, 10),
	}
}

func (m *FTMSManager) Connect() error {
	// TODO: Implement actual BLE scanning and connection
	// Using tinygo.org/x/bluetooth:
	// 1. Enable adapter
	// 2. Scan for FTMS service
	// 3. Connect to device
	// 4. Discover services/characteristics
	// 5. Subscribe to Indoor Bike Data notifications
	m.connected = true
	return nil
}

func (m *FTMSManager) Disconnect() {
	m.connected = false
	close(m.dataCh)
	close(m.shiftCh)
}

func (m *FTMSManager) IsConnected() bool {
	return m.connected
}

func (m *FTMSManager) DataChannel() <-chan TrainerData {
	return m.dataCh
}

func (m *FTMSManager) ShiftChannel() <-chan ShiftEvent {
	return m.shiftCh
}

func (m *FTMSManager) SetResistance(level float64) error {
	// TODO: Write to Fitness Machine Control Point
	// Opcode 0x04 = Set Target Resistance Level
	return nil
}

func (m *FTMSManager) SetTargetPower(watts float64) error {
	// TODO: Write to Fitness Machine Control Point
	// Opcode 0x05 = Set Target Power
	return nil
}
