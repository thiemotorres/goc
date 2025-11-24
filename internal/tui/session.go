package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thiemotorres/goc/internal/bluetooth"
	"github.com/thiemotorres/goc/internal/config"
	"github.com/thiemotorres/goc/internal/data"
	"github.com/thiemotorres/goc/internal/gpx"
	"github.com/thiemotorres/goc/internal/simulation"
)

// RideSession manages the active ride state
type RideSession struct {
	// Components
	engine    *simulation.Engine
	btManager bluetooth.Manager
	route     *gpx.Route
	ride      *data.Ride
	store     *data.Store

	// State
	ctx        context.Context
	cancel     context.CancelFunc
	paused     bool
	distance   float64
	lastUpdate time.Time

	// Averages
	totalPower   float64
	totalCadence float64
	totalSpeed   float64
	pointCount   int
}

// RideUpdateMsg is sent to update the ride screen
type RideUpdateMsg struct {
	Power      float64
	Cadence    float64
	Speed      float64
	Elapsed    time.Duration
	Distance   float64
	AvgPower   float64
	AvgCadence float64
	AvgSpeed   float64
	Elevation  float64
	Gradient   float64
	Gear       string
	Mode       string
	Paused     bool
}

// RideConnectingMsg indicates connection in progress
type RideConnectingMsg struct {
	Status string
}

// RideConnectedMsg indicates connection succeeded
type RideConnectedMsg struct{}

// RideErrorMsg indicates an error occurred
type RideErrorMsg struct {
	Error error
}

// RideFinishedMsg indicates ride is complete
type RideFinishedMsg struct {
	RideID string
}

// NewRideSession creates a new ride session
func NewRideSession(cfg *config.Config, rideType RideType, route *RouteInfo, mock bool) (*RideSession, error) {
	// Create simulation engine
	engine := simulation.NewEngine(simulation.EngineConfig{
		Chainrings:         cfg.Bike.Chainrings,
		Cassette:           cfg.Bike.Cassette,
		WheelCircumference: cfg.Bike.WheelCircumference,
		RiderWeight:        cfg.Bike.RiderWeight,
	})

	// Set mode
	switch rideType {
	case RideFree:
		engine.SetMode(simulation.ModeFREE)
	case RideERG:
		engine.SetMode(simulation.ModeERG)
		engine.SetTargetPower(150) // Default, could be configurable
	case RideRoute:
		engine.SetMode(simulation.ModeSIM)
	}

	// Load route if provided
	var gpxRoute *gpx.Route
	if route != nil {
		var err error
		gpxRoute, err = gpx.Load(route.Path)
		if err != nil {
			return nil, err
		}
	}

	// Create Bluetooth manager
	var btManager bluetooth.Manager
	if mock {
		btManager = bluetooth.NewMockManager()
	} else {
		btManager = bluetooth.NewFTMSManagerWithConfig(bluetooth.FTMSManagerConfig{
			SavedAddress: cfg.Bluetooth.TrainerAddress,
		})
	}

	// Create data store
	store, err := data.NewStore(data.DefaultDataDir())
	if err != nil {
		return nil, err
	}

	// Create ride recording
	ride := data.NewRide()
	if gpxRoute != nil {
		ride.GPXName = gpxRoute.Name
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &RideSession{
		engine:     engine,
		btManager:  btManager,
		route:      gpxRoute,
		ride:       ride,
		store:      store,
		ctx:        ctx,
		cancel:     cancel,
		lastUpdate: time.Now(),
	}, nil
}

// Connect initiates Bluetooth connection
func (rs *RideSession) Connect() tea.Cmd {
	return func() tea.Msg {
		if err := rs.btManager.Connect(); err != nil {
			return RideErrorMsg{Error: err}
		}
		return RideConnectedMsg{}
	}
}

// StartDataLoop starts the data processing loop
func (rs *RideSession) StartDataLoop() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-rs.ctx.Done():
			return nil

		case trainerData := <-rs.btManager.DataChannel():
			now := time.Now()
			dt := now.Sub(rs.lastUpdate).Seconds()
			rs.lastUpdate = now

			// Get gradient from route
			var gradient float64
			if rs.route != nil {
				gradient = rs.route.GradientAt(rs.distance)
			}

			// Update simulation
			state := rs.engine.Update(trainerData.Cadence, trainerData.Power, gradient)

			// Update position
			if !rs.paused {
				rs.distance += (state.Speed / 3.6) * dt
				rs.engine.Tick(dt, state.Speed)
			}

			// Record point
			var lat, lon, ele float64
			if rs.route != nil {
				lat, lon = rs.route.PositionAt(rs.distance)
				ele = rs.route.ElevationAt(rs.distance)
			}

			rs.ride.AddPoint(data.RidePoint{
				Timestamp:  now,
				Power:      state.Power,
				Cadence:    state.Cadence,
				Speed:      state.Speed,
				Latitude:   lat,
				Longitude:  lon,
				Elevation:  ele,
				Distance:   rs.distance,
				Gradient:   gradient,
				GearString: state.GearString,
			})

			// Update averages
			if !rs.paused {
				rs.totalPower += state.Power
				rs.totalCadence += state.Cadence
				rs.totalSpeed += state.Speed
				rs.pointCount++
			}

			// Send resistance to trainer
			if state.Mode == simulation.ModeSIM || state.Mode == simulation.ModeFREE {
				rs.btManager.SetResistance(state.Resistance)
			} else if state.Mode == simulation.ModeERG {
				rs.btManager.SetTargetPower(state.TargetPower)
			}

			var avgPower, avgCadence, avgSpeed float64
			if rs.pointCount > 0 {
				avgPower = rs.totalPower / float64(rs.pointCount)
				avgCadence = rs.totalCadence / float64(rs.pointCount)
				avgSpeed = rs.totalSpeed / float64(rs.pointCount)
			}

			return RideUpdateMsg{
				Power:      state.Power,
				Cadence:    state.Cadence,
				Speed:      state.Speed,
				Elapsed:    time.Since(rs.ride.StartTime),
				Distance:   rs.distance,
				AvgPower:   avgPower,
				AvgCadence: avgCadence,
				AvgSpeed:   avgSpeed,
				Elevation:  ele,
				Gradient:   gradient,
				Gear:       state.GearString,
				Mode:       state.Mode.String(),
				Paused:     rs.paused,
			}

		case event := <-rs.btManager.ShiftChannel():
			switch event {
			case bluetooth.ShiftUp:
				rs.engine.ShiftUp()
			case bluetooth.ShiftDown:
				rs.engine.ShiftDown()
			}
			return nil
		}
	}
}

// ShiftUp shifts to a harder gear
func (rs *RideSession) ShiftUp() {
	rs.engine.ShiftUp()
}

// ShiftDown shifts to an easier gear
func (rs *RideSession) ShiftDown() {
	rs.engine.ShiftDown()
}

// AdjustResistance changes manual resistance
func (rs *RideSession) AdjustResistance(delta float64) {
	rs.engine.AdjustManualResistance(delta)
}

// TogglePause toggles pause state
func (rs *RideSession) TogglePause() {
	rs.paused = !rs.paused
	if rs.paused {
		rs.ride.Pause()
	} else {
		rs.ride.Resume()
	}
}

// Stop ends the ride session
func (rs *RideSession) Stop() tea.Cmd {
	return func() tea.Msg {
		rs.cancel()
		rs.btManager.Disconnect()

		// Save ride
		rs.ride.Finish()
		var rideID string
		if len(rs.ride.Points) > 0 {
			rs.store.SaveRide(rs.ride)
			rideID = rs.ride.ID
		}

		rs.store.Close()

		return RideFinishedMsg{RideID: rideID}
	}
}
