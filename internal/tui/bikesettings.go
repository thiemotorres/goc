package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/thiemotorres/goc/internal/config"
)

// BikeSettings shows bike configuration options
type BikeSettings struct {
	items       []string
	selected    int
	config      *config.Config
	editing     bool
	editField   int
	editBuffer  string
}

func NewBikeSettings(cfg *config.Config) *BikeSettings {
	return &BikeSettings{
		items: []string{
			"Chainrings",
			"Cassette",
			"Wheel Circumference",
			"Rider Weight",
			"← Back",
		},
		config: cfg,
	}
}

func (m *BikeSettings) MoveUp() {
	if !m.editing && m.selected > 0 {
		m.selected--
	}
}

func (m *BikeSettings) MoveDown() {
	if !m.editing && m.selected < len(m.items)-1 {
		m.selected++
	}
}

func (m *BikeSettings) Selected() int {
	return m.selected
}

func (m *BikeSettings) IsEditing() bool {
	return m.editing
}

func (m *BikeSettings) StartEdit() {
	m.editing = true
	m.editField = m.selected

	// Pre-fill with current value
	switch m.selected {
	case 0: // Chainrings
		m.editBuffer = intsToString(m.config.Bike.Chainrings)
	case 1: // Cassette
		m.editBuffer = intsToString(m.config.Bike.Cassette)
	case 2: // Wheel Circumference
		m.editBuffer = fmt.Sprintf("%.3f", m.config.Bike.WheelCircumference)
	case 3: // Rider Weight
		m.editBuffer = fmt.Sprintf("%.1f", m.config.Bike.RiderWeight)
	}
}

func (m *BikeSettings) CancelEdit() {
	m.editing = false
	m.editBuffer = ""
}

func (m *BikeSettings) HandleKey(key string) bool {
	if !m.editing {
		return false
	}

	switch key {
	case "backspace":
		if len(m.editBuffer) > 0 {
			m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
		}
	case "enter":
		m.applyEdit()
		m.editing = false
		return true
	case "esc":
		m.CancelEdit()
		return true
	default:
		// Only accept valid characters
		if len(key) == 1 && (key[0] >= '0' && key[0] <= '9' || key[0] == '.' || key[0] == ',' || key[0] == ' ') {
			m.editBuffer += key
		}
	}
	return true
}

func (m *BikeSettings) applyEdit() {
	switch m.editField {
	case 0: // Chainrings
		if ints := parseInts(m.editBuffer); len(ints) > 0 {
			m.config.Bike.Chainrings = ints
		}
	case 1: // Cassette
		if ints := parseInts(m.editBuffer); len(ints) > 0 {
			m.config.Bike.Cassette = ints
		}
	case 2: // Wheel Circumference
		if f, err := strconv.ParseFloat(strings.TrimSpace(m.editBuffer), 64); err == nil && f > 0 {
			m.config.Bike.WheelCircumference = f
		}
	case 3: // Rider Weight
		if f, err := strconv.ParseFloat(strings.TrimSpace(m.editBuffer), 64); err == nil && f > 0 {
			m.config.Bike.RiderWeight = f
		}
	}
}

func (m *BikeSettings) View() string {
	var b strings.Builder

	title := titleStyle.Render("Bike Settings")
	b.WriteString(title)
	b.WriteString("\n\n")

	for i, item := range m.items {
		cursor := "  "
		style := normalStyle
		if i == m.selected {
			cursor = "> "
			style = selectedStyle
		}

		// Add current value
		var value string
		switch i {
		case 0: // Chainrings
			value = fmt.Sprintf(" [%s]", intsToString(m.config.Bike.Chainrings))
		case 1: // Cassette
			value = fmt.Sprintf(" [%s]", intsToString(m.config.Bike.Cassette))
		case 2: // Wheel Circumference
			value = fmt.Sprintf(" (%.3fm)", m.config.Bike.WheelCircumference)
		case 3: // Rider Weight
			value = fmt.Sprintf(" (%.1f kg)", m.config.Bike.RiderWeight)
		}

		line := item + value

		// Show edit mode
		if m.editing && i == m.editField {
			line = item + ": " + m.editBuffer + "█"
		}

		b.WriteString(cursor + style.Render(line) + "\n")
	}

	var help string
	if m.editing {
		help = helpStyle.Render("\nenter: save • esc: cancel")
	} else {
		help = helpStyle.Render("\n↑/↓: navigate • enter: edit • esc: back")
	}
	b.WriteString(help)

	return centerView(menuStyle.Render(b.String()))
}

func intsToString(ints []int) string {
	strs := make([]string, len(ints))
	for i, v := range ints {
		strs[i] = strconv.Itoa(v)
	}
	return strings.Join(strs, ", ")
}

func parseInts(s string) []int {
	// Accept comma or space separated
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)

	var result []int
	for _, p := range parts {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			result = append(result, v)
		}
	}
	return result
}
