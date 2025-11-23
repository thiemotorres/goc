package bluetooth

import (
	"errors"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

// FTMS UUIDs
const (
	FTMSServiceUUID                = "00001826-0000-1000-8000-00805f9b34fb"
	IndoorBikeDataUUID             = "00002ad2-0000-1000-8000-00805f9b34fb"
	FitnessMachineControlPointUUID = "00002ad9-0000-1000-8000-00805f9b34fb"
	FitnessMachineStatusUUID       = "00002ada-0000-1000-8000-00805f9b34fb"
)

// FTMSManagerConfig configures the FTMS manager
type FTMSManagerConfig struct {
	OnStatusChange    func(ConnectionStatus)
	OnDeviceSelection func([]DeviceInfo) int // returns selected index, -1 to cancel
	SavedAddress      string
	OnSaveDevice      func(address string) // called after successful connection
}

// FTMSManager implements Manager using real Bluetooth
type FTMSManager struct {
	config FTMSManagerConfig

	mu             sync.Mutex
	connected      bool
	status         ConnectionStatus
	device         bluetooth.Device
	controlPoint   bluetooth.DeviceCharacteristic
	deviceAddress  string

	dataCh  chan TrainerData
	shiftCh chan ShiftEvent
	stopCh  chan struct{}
}

// NewFTMSManager creates a new FTMS Bluetooth manager
func NewFTMSManager() *FTMSManager {
	return NewFTMSManagerWithConfig(FTMSManagerConfig{})
}

// NewFTMSManagerWithConfig creates a new FTMS manager with config
func NewFTMSManagerWithConfig(config FTMSManagerConfig) *FTMSManager {
	return &FTMSManager{
		config:  config,
		dataCh:  make(chan TrainerData, 10),
		shiftCh: make(chan ShiftEvent, 10),
		stopCh:  make(chan struct{}),
	}
}

func (m *FTMSManager) setStatus(s ConnectionStatus) {
	m.mu.Lock()
	m.status = s
	m.mu.Unlock()

	if m.config.OnStatusChange != nil {
		m.config.OnStatusChange(s)
	}
}

func (m *FTMSManager) Connect() error {
	m.setStatus(StatusConnecting)

	if err := adapter.Enable(); err != nil {
		return errors.New("failed to enable Bluetooth: " + err.Error())
	}

	var targetAddress string

	// Try saved address first
	if m.config.SavedAddress != "" {
		targetAddress = m.config.SavedAddress
	} else {
		// Scan for devices
		scanner := NewScanner()
		devices, err := scanner.Scan(10 * time.Second)
		if err != nil {
			return err
		}

		if len(devices) == 0 {
			return errors.New("no FTMS trainers found")
		}

		// Let user select
		selectedIdx := 0
		if m.config.OnDeviceSelection != nil {
			selectedIdx = m.config.OnDeviceSelection(devices)
			if selectedIdx < 0 || selectedIdx >= len(devices) {
				return errors.New("device selection cancelled")
			}
		}

		targetAddress = devices[selectedIdx].Address
	}

	// Connect to device
	var addr bluetooth.Address
	addr.Set(targetAddress)

	device, err := adapter.Connect(addr, bluetooth.ConnectionParams{})
	if err != nil {
		// If saved address failed, clear and retry with scan
		if m.config.SavedAddress != "" {
			m.config.SavedAddress = ""
			return m.Connect()
		}
		return errors.New("failed to connect: " + err.Error())
	}

	m.device = device
	m.deviceAddress = targetAddress

	// Discover services
	services, err := device.DiscoverServices([]bluetooth.UUID{
		bluetooth.NewUUID(mustParseUUID(FTMSServiceUUID)),
	})
	if err != nil || len(services) == 0 {
		device.Disconnect()
		return errors.New("FTMS service not found")
	}

	ftmsService := services[0]

	// Discover characteristics
	chars, err := ftmsService.DiscoverCharacteristics([]bluetooth.UUID{
		bluetooth.NewUUID(mustParseUUID(IndoorBikeDataUUID)),
		bluetooth.NewUUID(mustParseUUID(FitnessMachineControlPointUUID)),
	})
	if err != nil {
		device.Disconnect()
		return errors.New("failed to discover characteristics: " + err.Error())
	}

	var indoorBikeData, controlPoint bluetooth.DeviceCharacteristic
	for _, c := range chars {
		uuid := c.UUID().String()
		if uuid == IndoorBikeDataUUID {
			indoorBikeData = c
		} else if uuid == FitnessMachineControlPointUUID {
			controlPoint = c
		}
	}

	m.controlPoint = controlPoint

	// Subscribe to Indoor Bike Data notifications
	err = indoorBikeData.EnableNotifications(func(buf []byte) {
		data, err := ParseIndoorBikeData(buf)
		if err != nil {
			return
		}
		select {
		case m.dataCh <- data:
		default:
			// Channel full, drop
		}
	})
	if err != nil {
		device.Disconnect()
		return errors.New("failed to enable notifications: " + err.Error())
	}

	// Request control
	_, err = controlPoint.WriteWithoutResponse(EncodeRequestControl())
	if err != nil {
		// Non-fatal, some trainers don't require this
	}

	m.mu.Lock()
	m.connected = true
	m.mu.Unlock()

	m.setStatus(StatusConnected)

	// Save device for next time
	if m.config.OnSaveDevice != nil {
		m.config.OnSaveDevice(targetAddress)
	}

	// Start disconnect monitor
	go m.monitorConnection()

	return nil
}

func (m *FTMSManager) monitorConnection() {
	// tinygo bluetooth doesn't have disconnect callbacks yet
	// Poll connection status
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			// Check if we can still communicate
			// If not, trigger reconnection
		}
	}
}

func (m *FTMSManager) Disconnect() {
	m.mu.Lock()
	wasConnected := m.connected
	m.connected = false
	m.mu.Unlock()

	if wasConnected {
		close(m.stopCh)
		m.device.Disconnect()
	}

	m.setStatus(StatusDisconnected)
}

func (m *FTMSManager) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

func (m *FTMSManager) DataChannel() <-chan TrainerData {
	return m.dataCh
}

func (m *FTMSManager) ShiftChannel() <-chan ShiftEvent {
	return m.shiftCh
}

func (m *FTMSManager) SetResistance(level float64) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}
	_, err := m.controlPoint.WriteWithoutResponse(EncodeSetTargetResistance(level))
	return err
}

func (m *FTMSManager) SetTargetPower(watts float64) error {
	if !m.IsConnected() {
		return errors.New("not connected")
	}
	_, err := m.controlPoint.WriteWithoutResponse(EncodeSetTargetPower(watts))
	return err
}

func mustParseUUID(s string) [16]byte {
	uuid, err := bluetooth.ParseUUID(s)
	if err != nil {
		panic("invalid UUID: " + s)
	}
	return uuid.Bytes()
}
