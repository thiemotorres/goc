package bluetooth

import (
	"errors"
	"strings"
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

// Scanner handles BLE device discovery
type Scanner struct {
	devices  []DeviceInfo
	stopChan chan struct{}
}

// NewScanner creates a new BLE scanner
func NewScanner() *Scanner {
	return &Scanner{
		stopChan: make(chan struct{}),
	}
}

// Scan discovers FTMS devices for the given duration
func (s *Scanner) Scan(timeout time.Duration) ([]DeviceInfo, error) {
	if err := adapter.Enable(); err != nil {
		return nil, errors.New("failed to enable Bluetooth adapter: " + err.Error())
	}

	s.devices = nil
	seen := make(map[string]bool)

	done := make(chan error, 1)

	go func() {
		err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			// Check if device advertises FTMS service
			hasFTMS := false
			for _, uuid := range result.AdvertisementPayload.ServiceUUIDs() {
				if uuid.String() == FTMSServiceUUID {
					hasFTMS = true
					break
				}
			}

			if !hasFTMS {
				return
			}

			addr := result.Address.String()
			if seen[addr] {
				return
			}
			seen[addr] = true

			name := result.LocalName()
			if name == "" {
				name = "Unknown Trainer"
			}

			s.devices = append(s.devices, DeviceInfo{
				Address: addr,
				Name:    name,
				RSSI:    int(result.RSSI),
			})
		})
		done <- err
	}()

	select {
	case <-time.After(timeout):
		adapter.StopScan()
	case err := <-done:
		if err != nil && !strings.Contains(err.Error(), "timeout") {
			return nil, err
		}
	case <-s.stopChan:
		adapter.StopScan()
	}

	return s.devices, nil
}

// Stop stops an ongoing scan
func (s *Scanner) Stop() {
	select {
	case s.stopChan <- struct{}{}:
	default:
	}
}
