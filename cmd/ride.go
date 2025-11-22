package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thiemotorres/goc/internal/bluetooth"
	"github.com/thiemotorres/goc/internal/config"
	"github.com/thiemotorres/goc/internal/data"
	"github.com/thiemotorres/goc/internal/gpx"
	"github.com/thiemotorres/goc/internal/simulation"
	"github.com/thiemotorres/goc/internal/tui"
)

// RideOptions configures a ride session
type RideOptions struct {
	GPXPath  string
	ERGWatts int
	Mock     bool // Use mock Bluetooth for development
}

// Ride starts a cycling session
func Ride(opts RideOptions) error {
	// Load config
	cfg, err := config.Load(config.DefaultConfigDir())
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Create simulation engine
	engine := simulation.NewEngine(simulation.EngineConfig{
		Chainrings:         cfg.Bike.Chainrings,
		Cassette:           cfg.Bike.Cassette,
		WheelCircumference: cfg.Bike.WheelCircumference,
		RiderWeight:        cfg.Bike.RiderWeight,
	})

	// Set mode
	if opts.ERGWatts > 0 {
		engine.SetMode(simulation.ModeERG)
		engine.SetTargetPower(float64(opts.ERGWatts))
	} else if opts.GPXPath == "" {
		engine.SetMode(simulation.ModeFREE)
	}

	// Load GPX if provided
	var route *gpx.Route
	if opts.GPXPath != "" {
		route, err = gpx.Load(opts.GPXPath)
		if err != nil {
			return fmt.Errorf("load GPX: %w", err)
		}
		fmt.Printf("Loaded route: %s (%.1f km)\n", route.Name, route.TotalDistance/1000)
	}

	// Create Bluetooth manager
	var btManager bluetooth.Manager
	if opts.Mock {
		btManager = bluetooth.NewMockManager()
	} else {
		btManager = bluetooth.NewFTMSManager()
	}

	// Connect to trainer
	fmt.Println("Connecting to trainer...")
	if err := btManager.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer btManager.Disconnect()
	fmt.Println("Connected!")

	// Create data store
	store, err := data.NewStore(data.DefaultDataDir())
	if err != nil {
		return fmt.Errorf("create store: %w", err)
	}
	defer store.Close()

	// Create ride recording
	ride := data.NewRide()
	if route != nil {
		ride.GPXName = route.Name
	}

	// Create TUI
	renderer, err := tui.NewRenderer()
	if err != nil {
		return fmt.Errorf("create TUI: %w", err)
	}
	defer renderer.Close()

	// Context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// State
	var (
		paused       bool
		currentDist  float64
		lastUpdate   = time.Now()
		totalPower   float64
		totalCadence float64
		totalSpeed   float64
		pointCount   int
	)

	// Set up callbacks
	renderer.SetCallbacks(
		func() { engine.ShiftUp() },                  // Shift up
		func() { engine.ShiftDown() },                // Shift down
		func() { engine.AdjustManualResistance(5) },  // Resistance up
		func() { engine.AdjustManualResistance(-5) }, // Resistance down
		func() { // Pause toggle
			paused = !paused
			if paused {
				ride.Pause()
			} else {
				ride.Resume()
			}
		},
		func() { /* Toggle view */ },
		func() { cancel() }, // Quit
	)

	// Main loop goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case trainerData := <-btManager.DataChannel():
				now := time.Now()
				dt := now.Sub(lastUpdate).Seconds()
				lastUpdate = now

				// Get gradient from route
				var gradient float64
				if route != nil {
					gradient = route.GradientAt(currentDist)
				}

				// Update simulation
				state := engine.Update(trainerData.Cadence, trainerData.Power, gradient)

				// Update position
				if !paused {
					currentDist += (state.Speed / 3.6) * dt
					engine.Tick(dt, state.Speed)
				}

				// Record point
				var lat, lon, ele float64
				if route != nil {
					lat, lon = route.PositionAt(currentDist)
					ele = route.ElevationAt(currentDist)
				}

				ride.AddPoint(data.RidePoint{
					Timestamp:  now,
					Power:      state.Power,
					Cadence:    state.Cadence,
					Speed:      state.Speed,
					Latitude:   lat,
					Longitude:  lon,
					Elevation:  ele,
					Distance:   currentDist,
					Gradient:   gradient,
					GearString: state.GearString,
				})

				// Update averages
				if !paused {
					totalPower += state.Power
					totalCadence += state.Cadence
					totalSpeed += state.Speed
					pointCount++
				}

				// Update TUI
				renderer.UpdateMetrics(state.Power, state.Cadence, state.Speed)

				elapsed := time.Since(ride.StartTime)
				var avgPower, avgCadence, avgSpeed float64
				if pointCount > 0 {
					avgPower = totalPower / float64(pointCount)
					avgCadence = totalCadence / float64(pointCount)
					avgSpeed = totalSpeed / float64(pointCount)
				}

				renderer.UpdateStats(
					formatDuration(elapsed),
					currentDist,
					avgPower,
					avgCadence,
					avgSpeed,
					ele,
				)

				renderer.UpdateStatus(
					state.GearString,
					gradient,
					state.Mode.String(),
					paused,
				)

				// Send resistance to trainer
				if state.Mode == simulation.ModeSIM || state.Mode == simulation.ModeFREE {
					btManager.SetResistance(state.Resistance)
				} else if state.Mode == simulation.ModeERG {
					btManager.SetTargetPower(state.TargetPower)
				}

			case event := <-btManager.ShiftChannel():
				switch event {
				case bluetooth.ShiftUp:
					engine.ShiftUp()
				case bluetooth.ShiftDown:
					engine.ShiftDown()
				}
			}
		}
	}()

	// Run TUI
	if err := renderer.Run(ctx); err != nil && ctx.Err() == nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Save ride
	ride.Finish()
	if len(ride.Points) > 0 {
		fmt.Println("\nSaving ride...")
		if err := store.SaveRide(ride); err != nil {
			return fmt.Errorf("save ride: %w", err)
		}
		fmt.Printf("Ride saved: %s\n", store.GetFITPath(ride.ID))
	}

	return nil
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
