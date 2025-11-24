package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/thiemotorres/goc/internal/data"
)

// HistoryView shows past rides
type HistoryView struct {
	rides    []data.RideSummary
	selected int
	err      error
}

func NewHistoryView() *HistoryView {
	hv := &HistoryView{}
	hv.loadRides()
	return hv
}

func (hv *HistoryView) loadRides() {
	store, err := data.NewStore(data.DefaultDataDir())
	if err != nil {
		hv.err = err
		return
	}
	defer store.Close()

	hv.rides, hv.err = store.ListRides()
}

func (hv *HistoryView) MoveUp() {
	if hv.selected > 0 {
		hv.selected--
	}
}

func (hv *HistoryView) MoveDown() {
	max := len(hv.rides)
	if hv.selected < max {
		hv.selected++
	}
}

func (hv *HistoryView) Selected() int {
	return hv.selected
}

func (hv *HistoryView) SelectedRide() *data.RideSummary {
	if hv.selected < len(hv.rides) {
		return &hv.rides[hv.selected]
	}
	return nil
}

func (hv *HistoryView) View() string {
	var b strings.Builder

	title := titleStyle.Render("Ride History")
	b.WriteString(title)
	b.WriteString("\n\n")

	if hv.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", hv.err))
	} else if len(hv.rides) == 0 {
		b.WriteString("No rides recorded yet.\n")
	} else {
		for i, ride := range hv.rides {
			cursor := "  "
			style := normalStyle
			if i == hv.selected {
				cursor = "> "
				style = selectedStyle
			}

			date := ride.StartTime.Format("Jan 02")
			duration := formatDuration(ride.Duration)
			name := ride.GPXName
			if name == "" {
				name = "Free Ride"
			}

			line := fmt.Sprintf("%-6s  %-16s  %8s  %4.0fW avg",
				date,
				truncate(name, 16),
				duration,
				ride.AvgPower,
			)
			b.WriteString(cursor + style.Render(line) + "\n")
		}
	}

	// Back option
	cursor := "  "
	style := normalStyle
	if hv.selected == len(hv.rides) {
		cursor = "> "
		style = selectedStyle
	}
	b.WriteString("\n" + cursor + style.Render("← Back") + "\n")

	help := helpStyle.Render("\n↑/↓: navigate • enter: view • esc: back")
	b.WriteString(help)

	return centerView(menuStyle.Render(b.String()))
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
