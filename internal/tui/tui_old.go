package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
)

// Renderer handles the terminal UI
type Renderer struct {
	terminal  *tcell.Terminal
	container *container.Container

	// Widgets
	powerChart   *linechart.LineChart
	cadenceChart *linechart.LineChart
	speedChart   *linechart.LineChart
	routeView    *text.Text
	statsView    *text.Text
	statusView   *text.Text

	// Data buffers for charts
	powerData   []float64
	cadenceData []float64
	speedData   []float64
	maxPoints   int

	// Callbacks
	onShiftUp    func()
	onShiftDown  func()
	onResUp      func()
	onResDown    func()
	onPause      func()
	onToggleView func()
	onQuit       func()
}

// NewRenderer creates a new TUI renderer
func NewRenderer() (*Renderer, error) {
	t, err := tcell.New()
	if err != nil {
		return nil, fmt.Errorf("create terminal: %w", err)
	}

	r := &Renderer{
		terminal:  t,
		maxPoints: 300, // ~5 minutes at 1 update/sec
	}

	if err := r.createWidgets(); err != nil {
		t.Close()
		return nil, err
	}

	if err := r.createLayout(); err != nil {
		t.Close()
		return nil, err
	}

	return r, nil
}

func (r *Renderer) createWidgets() error {
	var err error

	// Power chart
	r.powerChart, err = linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
	)
	if err != nil {
		return fmt.Errorf("create power chart: %w", err)
	}

	// Cadence chart
	r.cadenceChart, err = linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
	)
	if err != nil {
		return fmt.Errorf("create cadence chart: %w", err)
	}

	// Speed chart
	r.speedChart, err = linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
	)
	if err != nil {
		return fmt.Errorf("create speed chart: %w", err)
	}

	// Route view
	r.routeView, err = text.New(text.WrapAtRunes())
	if err != nil {
		return fmt.Errorf("create route view: %w", err)
	}

	// Stats view
	r.statsView, err = text.New(text.WrapAtRunes())
	if err != nil {
		return fmt.Errorf("create stats view: %w", err)
	}

	// Status view
	r.statusView, err = text.New(text.WrapAtRunes())
	if err != nil {
		return fmt.Errorf("create status view: %w", err)
	}

	return nil
}

func (r *Renderer) createLayout() error {
	var err error

	r.container, err = container.New(
		r.terminal,
		container.Border(linestyle.Light),
		container.BorderTitle(" goc - Indoor Cycling Trainer "),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle(" Route "),
						container.PlaceWidget(r.routeView),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle(" Stats "),
						container.PlaceWidget(r.statsView),
					),
					container.SplitPercent(60),
				),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.BorderTitle(" Power (W) "),
								container.PlaceWidget(r.powerChart),
							),
							container.Bottom(
								container.Border(linestyle.Light),
								container.BorderTitle(" Cadence (rpm) "),
								container.PlaceWidget(r.cadenceChart),
							),
							container.SplitPercent(50),
						),
					),
					container.Bottom(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.BorderTitle(" Speed (km/h) "),
								container.PlaceWidget(r.speedChart),
							),
							container.Bottom(
								container.Border(linestyle.Light),
								container.BorderTitle(" Status "),
								container.PlaceWidget(r.statusView),
							),
							container.SplitPercent(70),
						),
					),
					container.SplitPercent(66),
				),
			),
			container.SplitPercent(40),
		),
	)

	return err
}

// SetCallbacks sets keyboard event callbacks
func (r *Renderer) SetCallbacks(
	onShiftUp, onShiftDown, onResUp, onResDown, onPause, onToggleView, onQuit func(),
) {
	r.onShiftUp = onShiftUp
	r.onShiftDown = onShiftDown
	r.onResUp = onResUp
	r.onResDown = onResDown
	r.onPause = onPause
	r.onToggleView = onToggleView
	r.onQuit = onQuit
}

// Run starts the TUI event loop
func (r *Renderer) Run(ctx context.Context) error {
	keyHandler := func(k *terminalapi.Keyboard) {
		switch k.Key {
		case keyboard.KeyArrowUp:
			if r.onShiftUp != nil {
				r.onShiftUp()
			}
		case keyboard.KeyArrowDown:
			if r.onShiftDown != nil {
				r.onShiftDown()
			}
		case keyboard.KeyArrowRight:
			if r.onResUp != nil {
				r.onResUp()
			}
		case keyboard.KeyArrowLeft:
			if r.onResDown != nil {
				r.onResDown()
			}
		case keyboard.KeySpace:
			if r.onPause != nil {
				r.onPause()
			}
		case keyboard.KeyTab:
			if r.onToggleView != nil {
				r.onToggleView()
			}
		case 'q', 'Q':
			if r.onQuit != nil {
				r.onQuit()
			}
		}
	}

	return termdash.Run(ctx, r.terminal, r.container,
		termdash.KeyboardSubscriber(keyHandler),
		termdash.RedrawInterval(100*time.Millisecond),
	)
}

// Close cleans up terminal
func (r *Renderer) Close() {
	r.terminal.Close()
}

// UpdateMetrics updates the chart data
func (r *Renderer) UpdateMetrics(power, cadence, speed float64) {
	r.powerData = append(r.powerData, power)
	r.cadenceData = append(r.cadenceData, cadence)
	r.speedData = append(r.speedData, speed)

	// Trim to max points
	if len(r.powerData) > r.maxPoints {
		r.powerData = r.powerData[1:]
		r.cadenceData = r.cadenceData[1:]
		r.speedData = r.speedData[1:]
	}

	// Update charts
	r.powerChart.Series("power", r.powerData,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorYellow)))
	r.cadenceChart.Series("cadence", r.cadenceData,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorCyan)))
	r.speedChart.Series("speed", r.speedData,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorGreen)))
}

// UpdateStats updates the stats panel
func (r *Renderer) UpdateStats(elapsed string, distance, avgPower, avgCadence, avgSpeed, elevation float64) {
	r.statsView.Reset()
	r.statsView.Write(fmt.Sprintf(
		"Time:      %s\n"+
			"Distance:  %.1f km\n"+
			"Avg Power: %.0f W\n"+
			"Avg Cad:   %.0f rpm\n"+
			"Avg Speed: %.1f km/h\n"+
			"Elevation: +%.0f m",
		elapsed, distance/1000, avgPower, avgCadence, avgSpeed, elevation,
	))
}

// UpdateStatus updates the status bar
func (r *Renderer) UpdateStatus(gear string, gradient float64, mode string, paused bool) {
	r.statusView.Reset()
	pauseStr := ""
	if paused {
		pauseStr = " [PAUSED]"
	}
	r.statusView.Write(fmt.Sprintf(
		"Gear: %s  Gradient: %+.1f%%  Mode: %s%s\n"+
			"[↑↓] Shift  [←→] Resistance  [Space] Pause  [q] Quit",
		gear, gradient, mode, pauseStr,
	))
}

// UpdateRoute updates the route view
func (r *Renderer) UpdateRoute(content string) {
	r.routeView.Reset()
	r.routeView.Write(content)
}
