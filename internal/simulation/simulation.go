package simulation

// Mode represents the training mode
type Mode int

const (
	ModeSIM  Mode = iota // GPX simulation
	ModeERG              // Fixed power
	ModeFREE             // Manual resistance
)

func (m Mode) String() string {
	switch m {
	case ModeSIM:
		return "SIM"
	case ModeERG:
		return "ERG"
	case ModeFREE:
		return "FREE"
	default:
		return "UNKNOWN"
	}
}

// EngineConfig holds simulation parameters
type EngineConfig struct {
	Chainrings         []int
	Cassette           []int
	WheelCircumference float64
	RiderWeight        float64
	ResistanceScaling  float64
}

// State represents current simulation state
type State struct {
	Cadence      float64
	Power        float64
	Speed        float64
	Resistance   float64
	Gradient     float64
	GearString   string
	GearRatio    float64
	Mode         Mode
	TargetPower  float64 // For ERG mode
	Distance     float64 // Cumulative meters
	ElapsedTime  float64 // Seconds
}

// Engine handles physics calculations
type Engine struct {
	config           EngineConfig
	gears            *GearSystem
	mode             Mode
	targetPower      float64
	manualResistance float64
	distance         float64
	elapsedTime      float64
}

// NewEngine creates a new simulation engine
func NewEngine(cfg EngineConfig) *Engine {
	return &Engine{
		config:           cfg,
		gears:            NewGearSystem(cfg.Chainrings, cfg.Cassette),
		mode:             ModeSIM,
		manualResistance: 20, // Default for FREE mode
	}
}

// Update calculates new state based on inputs
// cadence: RPM from trainer
// power: Watts from trainer
// gradient: current gradient in percent (from GPX)
func (e *Engine) Update(cadence, power, gradient float64) State {
	speed := CalculateSpeed(cadence, e.gears.Ratio(), e.config.WheelCircumference)

	var resistance float64
	switch e.mode {
	case ModeSIM:
		scaling := e.config.ResistanceScaling
		if scaling == 0 {
			scaling = 0.2 // Fallback default
		}
		resistance = CalculateResistance(speed, gradient, e.config.RiderWeight, e.gears.Ratio(), scaling)
	case ModeERG:
		resistance = 0 // ERG mode uses target power, not resistance
	case ModeFREE:
		// Apply gear ratio scaling to manual resistance
		// Treat manual resistance as a base wheel force equivalent
		// Scale by gear ratio to get pedal resistance
		baseResistance := e.manualResistance
		// Convert to approximate force, apply gear ratio, convert back
		// Simplified: directly scale by gear ratio relative to reference (2.5)
		const referenceGearRatio = 2.5
		resistance = baseResistance * (e.gears.Ratio() / referenceGearRatio)
		// Clamp to valid range
		if resistance < 0 {
			resistance = 0
		}
		if resistance > 100 {
			resistance = 100
		}
	}

	return State{
		Cadence:     cadence,
		Power:       power,
		Speed:       speed,
		Resistance:  resistance,
		Gradient:    gradient,
		GearString:  e.gears.String(),
		GearRatio:   e.gears.Ratio(),
		Mode:        e.mode,
		TargetPower: e.targetPower,
		Distance:    e.distance,
		ElapsedTime: e.elapsedTime,
	}
}

// Tick advances time and distance
func (e *Engine) Tick(deltaSeconds float64, speedKmh float64) {
	e.elapsedTime += deltaSeconds
	e.distance += (speedKmh / 3.6) * deltaSeconds // km/h to m/s
}

// Mode returns current training mode
func (e *Engine) Mode() Mode {
	return e.mode
}

// SetMode changes training mode
func (e *Engine) SetMode(m Mode) {
	e.mode = m
}

// SetTargetPower sets ERG mode target
func (e *Engine) SetTargetPower(watts float64) {
	e.targetPower = watts
}

// SetManualResistance sets FREE mode resistance
func (e *Engine) SetManualResistance(level float64) {
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}
	e.manualResistance = level
}

// AdjustManualResistance changes FREE mode resistance by delta
func (e *Engine) AdjustManualResistance(delta float64) {
	e.SetManualResistance(e.manualResistance + delta)
}

// ShiftUp shifts to harder gear
func (e *Engine) ShiftUp() {
	e.gears.ShiftUp()
}

// ShiftDown shifts to easier gear
func (e *Engine) ShiftDown() {
	e.gears.ShiftDown()
}

// GearRatio returns current gear ratio
func (e *Engine) GearRatio() float64 {
	return e.gears.Ratio()
}

// GearString returns current gear as string
func (e *Engine) GearString() string {
	return e.gears.String()
}

// Reset clears distance and time
func (e *Engine) Reset() {
	e.distance = 0
	e.elapsedTime = 0
}
