package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thiemotorres/goc/internal/bluetooth"
	"github.com/thiemotorres/goc/internal/config"
)

// ScannerScreen handles Bluetooth device scanning
type ScannerScreen struct {
	devices  []bluetooth.DeviceInfo
	selected int
	scanning bool
	err      error
	config   *config.Config
}

// ScanStartMsg initiates scanning
type ScanStartMsg struct{}

// ScanResultMsg contains scan results
type ScanResultMsg struct {
	Devices []bluetooth.DeviceInfo
	Error   error
}

// DeviceSelectedMsg indicates a device was selected
type DeviceSelectedMsg struct {
	Address string
	Name    string
}

func NewScannerScreen(cfg *config.Config) *ScannerScreen {
	return &ScannerScreen{
		config:   cfg,
		scanning: true,
	}
}

func (s *ScannerScreen) StartScan() tea.Cmd {
	return func() tea.Msg {
		// Use the Scanner directly
		scanner := bluetooth.NewScanner()
		devices, err := scanner.Scan(10 * time.Second)
		if err != nil {
			return ScanResultMsg{Error: err}
		}

		return ScanResultMsg{Devices: devices}
	}
}

func (s *ScannerScreen) MoveUp() {
	if s.selected > 0 {
		s.selected--
	}
}

func (s *ScannerScreen) MoveDown() {
	max := len(s.devices) // Back option is at len(devices)
	if s.selected < max {
		s.selected++
	}
}

func (s *ScannerScreen) Selected() int {
	return s.selected
}

func (s *ScannerScreen) Update(msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case ScanResultMsg:
		s.scanning = false
		s.err = msg.Error
		s.devices = msg.Devices
		return false, nil
	case tea.KeyMsg:
		if s.scanning {
			return false, nil // Ignore keys while scanning
		}
	}
	return false, nil
}

func (s *ScannerScreen) SelectDevice() *bluetooth.DeviceInfo {
	if s.selected < len(s.devices) {
		return &s.devices[s.selected]
	}
	return nil
}

func (s *ScannerScreen) View() string {
	var b strings.Builder

	title := titleStyle.Render("Scan for Trainers")
	b.WriteString(title)
	b.WriteString("\n\n")

	if s.scanning {
		b.WriteString("Scanning for FTMS trainers...\n\n")
		b.WriteString("Please wait (up to 10 seconds)\n")
	} else if s.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n\n", s.err))
		b.WriteString("Press any key to go back.\n")
	} else if len(s.devices) == 0 {
		b.WriteString("No trainers found.\n\n")
		b.WriteString("Make sure your trainer is:\n")
		b.WriteString("  • Powered on\n")
		b.WriteString("  • In pairing mode\n")
		b.WriteString("  • Not connected to another device\n")
	} else {
		b.WriteString(fmt.Sprintf("Found %d trainer(s):\n\n", len(s.devices)))

		for i, device := range s.devices {
			cursor := "  "
			style := normalStyle
			if i == s.selected {
				cursor = "> "
				style = selectedStyle
			}
			rssi := ""
			if device.RSSI != 0 {
				rssi = fmt.Sprintf(" (%d dBm)", device.RSSI)
			}
			line := fmt.Sprintf("%s%s", device.Name, rssi)
			b.WriteString(cursor + style.Render(line) + "\n")
		}
	}

	// Back option (always shown when not scanning)
	if !s.scanning {
		cursor := "  "
		style := normalStyle
		if s.selected == len(s.devices) {
			cursor = "> "
			style = selectedStyle
		}
		b.WriteString("\n" + cursor + style.Render("← Back") + "\n")
	}

	if !s.scanning && len(s.devices) > 0 {
		help := helpStyle.Render("\n↑/↓: navigate • enter: select • esc: back")
		b.WriteString(help)
	} else if !s.scanning {
		help := helpStyle.Render("\nesc: back • r: retry scan")
		b.WriteString(help)
	}

	return centerView(menuStyle.Render(b.String()))
}
