package data

import (
	"encoding/json"
	"os"
)

// ExportFIT writes ride data to a file
// Note: For MVP, this exports as JSON. FIT binary format can be added later.
func ExportFIT(ride *Ride, path string) error {
	// For MVP, export as JSON which is human-readable and importable
	// FIT binary encoding can be added with a proper encoder library

	export := struct {
		ID        string      `json:"id"`
		StartTime string      `json:"start_time"`
		EndTime   string      `json:"end_time"`
		GPXName   string      `json:"gpx_name,omitempty"`
		Stats     RideStats   `json:"stats"`
		Points    []RidePoint `json:"points"`
	}{
		ID:        ride.ID,
		StartTime: ride.StartTime.Format("2006-01-02T15:04:05Z"),
		EndTime:   ride.EndTime.Format("2006-01-02T15:04:05Z"),
		GPXName:   ride.GPXName,
		Stats:     ride.Stats(),
		Points:    ride.Points,
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
