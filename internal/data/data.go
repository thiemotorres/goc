package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// RideSummary is a lightweight ride listing
type RideSummary struct {
	ID        string
	StartTime time.Time
	Duration  time.Duration
	Distance  float64
	AvgPower  float64
	GPXName   string
}

// Store handles ride persistence
type Store struct {
	db      *sql.DB
	dataDir string
}

// NewStore creates a new data store
func NewStore(dataDir string) (*Store, error) {
	// Create directories
	ridesDir := filepath.Join(dataDir, "rides")
	if err := os.MkdirAll(ridesDir, 0755); err != nil {
		return nil, fmt.Errorf("create rides dir: %w", err)
	}

	// Open SQLite database
	dbPath := filepath.Join(dataDir, "history.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("create tables: %w", err)
	}

	return &Store{
		db:      db,
		dataDir: dataDir,
	}, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS rides (
			id TEXT PRIMARY KEY,
			start_time DATETIME,
			end_time DATETIME,
			duration_seconds INTEGER,
			distance_meters REAL,
			avg_power REAL,
			max_power REAL,
			avg_cadence REAL,
			avg_speed REAL,
			total_ascent REAL,
			gpx_name TEXT,
			metadata TEXT
		)
	`)
	return err
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveRide saves a ride to disk and database
func (s *Store) SaveRide(ride *Ride) error {
	stats := ride.Stats()

	// Save data file (JSON for MVP)
	fitPath := s.GetFITPath(ride.ID)
	if err := ExportFIT(ride, fitPath); err != nil {
		return fmt.Errorf("export data: %w", err)
	}

	// Save JSON metadata (for full reload if needed)
	jsonPath := filepath.Join(s.dataDir, "rides", ride.ID+".json")
	jsonData, err := json.Marshal(ride)
	if err != nil {
		return fmt.Errorf("marshal ride: %w", err)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	// Insert into database
	_, err = s.db.Exec(`
		INSERT INTO rides (id, start_time, end_time, duration_seconds, distance_meters,
			avg_power, max_power, avg_cadence, avg_speed, total_ascent, gpx_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		ride.ID,
		ride.StartTime,
		ride.EndTime,
		int(stats.Duration.Seconds()),
		stats.Distance,
		stats.AvgPower,
		stats.MaxPower,
		stats.AvgCadence,
		stats.AvgSpeed,
		stats.TotalAscent,
		ride.GPXName,
	)

	return err
}

// ListRides returns all rides ordered by date descending
func (s *Store) ListRides() ([]RideSummary, error) {
	rows, err := s.db.Query(`
		SELECT id, start_time, duration_seconds, distance_meters, avg_power, gpx_name
		FROM rides
		ORDER BY start_time DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []RideSummary
	for rows.Next() {
		var r RideSummary
		var durationSec int
		var gpxName sql.NullString

		if err := rows.Scan(&r.ID, &r.StartTime, &durationSec, &r.Distance, &r.AvgPower, &gpxName); err != nil {
			return nil, err
		}

		r.Duration = time.Duration(durationSec) * time.Second
		if gpxName.Valid {
			r.GPXName = gpxName.String
		}

		rides = append(rides, r)
	}

	return rides, rows.Err()
}

// GetFITPath returns the path to a ride's data file
func (s *Store) GetFITPath(rideID string) string {
	return filepath.Join(s.dataDir, "rides", rideID+".fit")
}

// DefaultDataDir returns the default data directory
func DefaultDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "goc")
}
