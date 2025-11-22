package cmd

import (
	"fmt"
	"time"

	"github.com/thiemotorres/goc/internal/data"
)

// HistoryOptions configures history display
type HistoryOptions struct {
	Limit int
}

// History displays ride history
func History(opts HistoryOptions) error {
	store, err := data.NewStore(data.DefaultDataDir())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer store.Close()

	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}

	rides, err := store.ListRides()
	if err != nil {
		return fmt.Errorf("list rides: %w", err)
	}

	// Apply limit
	if len(rides) > limit {
		rides = rides[:limit]
	}

	if len(rides) == 0 {
		fmt.Println("No rides recorded yet.")
		return nil
	}

	fmt.Println("Recent Rides:")
	fmt.Println("─────────────────────────────────────────────────────────────────")
	fmt.Printf("%-20s  %-10s  %-10s  %-10s  %-12s\n",
		"Date", "Duration", "Distance", "Avg Power", "Route")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	for _, r := range rides {
		date := r.StartTime.Format("2006-01-02 15:04")
		duration := formatDurationShort(r.Duration)
		distance := fmt.Sprintf("%.1f km", r.Distance/1000)
		avgPower := fmt.Sprintf("%.0f W", r.AvgPower)
		route := r.GPXName
		if route == "" {
			route = "Free ride"
		}
		if len(route) > 12 {
			route = route[:9] + "..."
		}

		fmt.Printf("%-20s  %-10s  %-10s  %-10s  %-12s\n",
			date, duration, distance, avgPower, route)
	}

	return nil
}

func formatDurationShort(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%dm%02ds", m, s)
}
