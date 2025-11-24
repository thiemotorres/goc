package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RideScreen is the active ride display
type RideScreen struct {
	width  int
	height int

	// Data buffers for charts
	powerData   []float64
	cadenceData []float64
	speedData   []float64
	maxPoints   int

	// Current values
	power   float64
	cadence float64
	speed   float64

	// State
	elapsed    time.Duration
	distance   float64
	avgPower   float64
	avgCadence float64
	avgSpeed   float64
	elevation  float64
	gradient   float64
	gear       string
	mode       string
	paused     bool

	// Callbacks
	onShiftUp   func()
	onShiftDown func()
	onResUp     func()
	onResDown   func()
	onPause     func()
	onQuit      func()
}

func NewRideScreen() *RideScreen {
	return &RideScreen{
		maxPoints: 60, // ~1 minute of data for sparkline
	}
}

func (rs *RideScreen) SetCallbacks(shiftUp, shiftDown, resUp, resDown, pause, quit func()) {
	rs.onShiftUp = shiftUp
	rs.onShiftDown = shiftDown
	rs.onResUp = resUp
	rs.onResDown = resDown
	rs.onPause = pause
	rs.onQuit = quit
}

func (rs *RideScreen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if rs.onShiftUp != nil {
				rs.onShiftUp()
			}
		case "down", "j":
			if rs.onShiftDown != nil {
				rs.onShiftDown()
			}
		case "right", "l":
			if rs.onResUp != nil {
				rs.onResUp()
			}
		case "left", "h":
			if rs.onResDown != nil {
				rs.onResDown()
			}
		case " ":
			if rs.onPause != nil {
				rs.onPause()
			}
		case "q":
			if rs.onQuit != nil {
				rs.onQuit()
			}
		}
	case tea.WindowSizeMsg:
		rs.width = msg.Width
		rs.height = msg.Height
	}
	return nil
}

func (rs *RideScreen) UpdateMetrics(power, cadence, speed float64) {
	rs.power = power
	rs.cadence = cadence
	rs.speed = speed

	rs.powerData = append(rs.powerData, power)
	rs.cadenceData = append(rs.cadenceData, cadence)
	rs.speedData = append(rs.speedData, speed)

	if len(rs.powerData) > rs.maxPoints {
		rs.powerData = rs.powerData[1:]
		rs.cadenceData = rs.cadenceData[1:]
		rs.speedData = rs.speedData[1:]
	}
}

func (rs *RideScreen) UpdateStats(elapsed time.Duration, distance, avgPower, avgCadence, avgSpeed, elevation float64) {
	rs.elapsed = elapsed
	rs.distance = distance
	rs.avgPower = avgPower
	rs.avgCadence = avgCadence
	rs.avgSpeed = avgSpeed
	rs.elevation = elevation
}

func (rs *RideScreen) UpdateStatus(gear string, gradient float64, mode string, paused bool) {
	rs.gear = gear
	rs.gradient = gradient
	rs.mode = mode
	rs.paused = paused
}

func (rs *RideScreen) View() string {
	var b strings.Builder

	// Title
	title := "goc - Indoor Cycling Trainer"
	if rs.paused {
		title += " [PAUSED]"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Big metrics row
	powerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Width(20).
		Align(lipgloss.Center)

	cadenceStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("45")).
		Width(20).
		Align(lipgloss.Center)

	speedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("42")).
		Width(20).
		Align(lipgloss.Center)

	bigMetrics := lipgloss.JoinHorizontal(lipgloss.Center,
		powerStyle.Render(fmt.Sprintf("%.0f W", rs.power)),
		cadenceStyle.Render(fmt.Sprintf("%.0f rpm", rs.cadence)),
		speedStyle.Render(fmt.Sprintf("%.1f km/h", rs.speed)),
	)
	b.WriteString(bigMetrics)
	b.WriteString("\n")

	// Sparklines
	powerSparkline := rs.generateSparkline(rs.powerData, 20)
	cadenceSparkline := rs.generateSparkline(rs.cadenceData, 20)
	speedSparkline := rs.generateSparkline(rs.speedData, 20)

	sparklines := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(20).Align(lipgloss.Center).Foreground(lipgloss.Color("212")).Render(powerSparkline),
		lipgloss.NewStyle().Width(20).Align(lipgloss.Center).Foreground(lipgloss.Color("45")).Render(cadenceSparkline),
		lipgloss.NewStyle().Width(20).Align(lipgloss.Center).Foreground(lipgloss.Color("42")).Render(speedSparkline),
	)
	b.WriteString(sparklines)
	b.WriteString("\n\n")

	// Stats
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	stats := fmt.Sprintf(
		"Time: %s  |  Distance: %.1f km  |  Elevation: +%.0fm",
		formatDuration(rs.elapsed),
		rs.distance/1000,
		rs.elevation,
	)
	b.WriteString(statsStyle.Render(stats))
	b.WriteString("\n")

	avgs := fmt.Sprintf(
		"Avg Power: %.0fW  |  Avg Cadence: %.0f rpm  |  Avg Speed: %.1f km/h",
		rs.avgPower, rs.avgCadence, rs.avgSpeed,
	)
	b.WriteString(statsStyle.Render(avgs))
	b.WriteString("\n\n")

	// Status bar
	gearStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	status := fmt.Sprintf("Gear: %s  |  Gradient: %+.1f%%  |  Mode: %s",
		gearStyle.Render(rs.gear), rs.gradient, rs.mode)
	b.WriteString(status)
	b.WriteString("\n\n")

	// Controls help
	help := helpStyle.Render("[↑↓] Shift  [←→] Resistance  [Space] Pause  [q] Quit")
	b.WriteString(help)

	return b.String()
}

func (rs *RideScreen) generateSparkline(data []float64, width int) string {
	if len(data) == 0 {
		return strings.Repeat("▁", width)
	}

	// Sample data to fit width
	sampled := make([]float64, width)
	if len(data) <= width {
		// Pad with zeros at start
		offset := width - len(data)
		for i := 0; i < width; i++ {
			if i < offset {
				sampled[i] = 0
			} else {
				sampled[i] = data[i-offset]
			}
		}
	} else {
		// Sample evenly
		for i := 0; i < width; i++ {
			idx := int(float64(i) * float64(len(data)-1) / float64(width-1))
			sampled[i] = data[idx]
		}
	}

	// Find min/max
	minVal, maxVal := sampled[0], sampled[0]
	for _, v := range sampled {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// Sparkline characters
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var sb strings.Builder
	valRange := maxVal - minVal
	if valRange == 0 {
		valRange = 1
	}

	for _, v := range sampled {
		normalized := (v - minVal) / valRange
		idx := int(normalized * float64(len(chars)-1))
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		if idx < 0 {
			idx = 0
		}
		sb.WriteRune(chars[idx])
	}

	return sb.String()
}
