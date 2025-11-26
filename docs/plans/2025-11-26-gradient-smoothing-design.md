# Gradient Smoothing Design

**Date:** 2025-11-26
**Status:** Approved
**Problem:** Resistance changes every 1-2 seconds due to GPS elevation noise, making rides feel jerky and unrealistic

## Problem Statement

GPS elevation data has ±5-10m accuracy, causing point-to-point gradient calculations to fluctuate wildly:
- Point 1: 50m elevation
- Point 2: 55m (+5m noise) → +50% gradient over 10m
- Point 3: 48m (-7m noise) → -70% gradient

On "flat" routes, this creates resistance changes every 1-2 seconds, which is:
- Unrealistic (real cycling has momentum/inertia)
- Distracting (constant trainer adjustment)
- Unusable (can't maintain steady effort)

## Design Goal

Create **gradual, momentum-based resistance changes** that:
- Eliminate 1-2 second jitter from GPS noise
- Mimic real outdoor cycling (momentum carries you into climbs)
- Feel natural and realistic
- Allow tuning via configuration

## Solution: Exponential Moving Average (EMA)

### Approach

Apply EMA smoothing to gradient values in the simulation engine:

```
smoothed_gradient = α × previous_gradient + (1 - α) × new_gradient
```

Where:
- α (alpha) = smoothing factor (0.0 - 1.0)
- α = 0.85 (default) → 85% history, 15% new value
- Higher α = smoother but slower response
- Lower α = faster response but less smoothing

### Why EMA?

**Advantages:**
- ✅ Simple (1 line of code)
- ✅ Natural lag mimics momentum/inertia
- ✅ No lookback window or buffer needed
- ✅ Low memory usage (single float64)
- ✅ Proven effective in other domains (signal processing, finance)

**Trade-offs:**
- ⚠️ Lag is time-based, not distance-based (may feel slightly different at different speeds)
- ⚠️ Requires tuning to find right α value

**Alternatives considered:**
- Distance-based rolling average (more complex, needs circular buffer)
- Minimum threshold filtering (doesn't provide smooth transitions)
- Combined approach (overengineered for initial implementation)

## Implementation Details

### Location

Apply smoothing in `internal/simulation/simulation.go` at the `Engine` level:

**Why here?**
- GPX data stays pure/unmodified (useful for other features)
- Smoothing is physics/simulation concern (logical grouping)
- Easy to test in isolation
- State persists across updates

### Engine Changes

```go
type Engine struct {
    // ... existing fields ...
    smoothedGradient float64  // EMA-smoothed gradient value
    smoothingFactor  float64  // alpha (0.85 default)
}

func (e *Engine) Update(cadence, power, gradient float64) State {
    // Apply EMA smoothing
    e.smoothedGradient = e.smoothingFactor * e.smoothedGradient +
                        (1 - e.smoothingFactor) * gradient

    // Use smoothed gradient for resistance calculation
    resistance = CalculateResistance(speed, e.smoothedGradient, weight, gearRatio, scaling)

    // Store both raw and smoothed in state for display/debug
    return State{
        // ...
        Gradient: e.smoothedGradient,  // Use smoothed value
        // ...
    }
}
```

### Configuration

Add optional parameter to config:

```toml
[bike]
gradient_smoothing = 0.85  # 0.0 = disabled, 0.95 = very smooth
```

**Default value:** 0.85 (provides 20-30 second lag, mimics momentum)

**Value guidelines:**
- 0.0 = No smoothing (instant response, noisy)
- 0.5 = Responsive (5-8 second lag)
- 0.7 = Balanced (10-15 second lag)
- 0.85 = Smooth (20-30 second lag) [DEFAULT]
- 0.9+ = Very smooth (30+ second lag, may miss short climbs)

### EngineConfig Update

```go
type EngineConfig struct {
    // ... existing fields ...
    GradientSmoothing float64  // NEW
}

func NewEngine(cfg EngineConfig) *Engine {
    smoothing := cfg.GradientSmoothing
    if smoothing == 0 {
        smoothing = 0.85  // Default
    }

    return &Engine{
        // ...
        smoothingFactor:  smoothing,
        smoothedGradient: 0.0,  // Initialize at flat
    }
}
```

## Edge Cases

### 1. Ride Start
- `smoothedGradient` initializes to 0.0
- First few seconds ramp up from 0% to actual gradient
- **Acceptable:** Mimics getting up to speed

### 2. Paused Rides
- Keep current `smoothedGradient` value
- Don't reset on pause/resume
- **Result:** Smooth continuation when resuming

### 3. Route Change/Restart
- Reset `smoothedGradient` to 0.0
- Fresh start for new route
- **Implementation:** Add `Reset()` method or handle in session layer

### 4. FREE Mode (No Route)
- Gradient always 0, no smoothing needed
- No code changes required
- **Result:** Works transparently

### 5. ERG Mode
- Ignores gradient anyway
- Smoothing has no effect
- **Result:** No impact

### 6. Smoothing Disabled (α = 0)
- Formula becomes: `smoothed = 0 × prev + 1 × new = new`
- Passes gradient through unchanged
- **Result:** Backward compatible with old behavior

## Testing Strategy

### Unit Tests

```go
// Test 1: Noisy gradient smoothing
func TestEngine_GradientSmoothing_NoisyInput() {
    engine := NewEngine(EngineConfig{GradientSmoothing: 0.8})

    gradients := []float64{0, 5, -2, 6, 1, 4, -1, 5}

    for _, g := range gradients {
        state := engine.Update(80, 200, g)
        // Verify: state.Gradient changes gradually, not jumping
    }

    // Final smoothed gradient should be ~3% (not jumping to 5%)
}

// Test 2: Climb detection
func TestEngine_GradientSmoothing_ClimbResponse() {
    engine := NewEngine(EngineConfig{GradientSmoothing: 0.85})

    // Flat then sudden 3% climb
    for i := 0; i < 5; i++ {
        engine.Update(80, 200, 0.0)  // Flat
    }

    resistanceBefore := engine.Update(80, 200, 0.0).Resistance

    for i := 0; i < 10; i++ {
        engine.Update(80, 200, 3.0)  // Climb
    }

    resistanceAfter := engine.Update(80, 200, 3.0).Resistance

    // Verify: resistance increased gradually (not instantly)
    assert.True(resistanceAfter > resistanceBefore)
}

// Test 3: Smoothing disabled
func TestEngine_GradientSmoothing_Disabled() {
    engine := NewEngine(EngineConfig{GradientSmoothing: 0.0})

    state1 := engine.Update(80, 200, 5.0)
    assert.Equal(5.0, state1.Gradient)  // Instant response

    state2 := engine.Update(80, 200, -2.0)
    assert.Equal(-2.0, state2.Gradient)  // No smoothing
}

// Test 4: Default configuration
func TestEngine_GradientSmoothing_DefaultValue() {
    engine := NewEngine(EngineConfig{})  // No smoothing specified

    // Should use 0.85 default
    assert.Equal(0.85, engine.smoothingFactor)
}
```

### Integration Testing

Manual testing procedure in `docs/testing/2025-11-26-gradient-smoothing-smoke-test.md`:
1. Load GPX route known to be "flat" but noisy
2. Ride for 2-3 minutes
3. Observe resistance changes
4. **Expected:** Smooth, gradual changes (not every 1-2 seconds)

### Validation Metrics

Success criteria:
- ✅ Resistance changes < once per 5 seconds on flat terrain
- ✅ Climbs still detected (resistance increases on hills)
- ✅ Configurable (users can tune α value)
- ✅ No performance impact (< 1μs per update)
- ✅ Backward compatible (existing routes work)

## Configuration Documentation

Update README.md with new option:

```markdown
#### gradient_smoothing

**Type:** float
**Default:** 0.85
**Range:** 0.0 - 0.95

Controls gradient smoothing using exponential moving average.
- 0.0: No smoothing (instant response, may be jerky)
- 0.85: Default (smooth, natural feel with ~20-30 second lag)
- 0.95: Very smooth (minimal jitter, ~30+ second lag)

Adjust if gradient changes feel too sudden or too slow to respond.
```

## Backward Compatibility

- ✅ Config file optional (uses 0.85 default)
- ✅ Existing routes work without modification
- ✅ Can disable by setting to 0.0
- ✅ No changes to GPX parsing or data storage
- ✅ No breaking changes to API

## Performance Impact

- **Memory:** +16 bytes per Engine instance (2 float64 fields)
- **CPU:** 1 multiplication + 1 addition per update (~1ns overhead)
- **I/O:** None
- **Assessment:** Negligible impact

## Future Enhancements

If EMA smoothing proves insufficient:

1. **Distance-based averaging** - Average gradient over last 50-100m
2. **Minimum threshold** - Ignore changes < 0.5%
3. **Adaptive smoothing** - Adjust α based on speed/terrain
4. **Pre-smoothing GPX** - Smooth on import (one-time cost)
5. **Configurable per route** - Different smoothing for different terrain types

These are not implemented initially (YAGNI principle).

## Implementation Tasks

1. Add `GradientSmoothing` field to `EngineConfig`
2. Add `smoothedGradient` and `smoothingFactor` to `Engine` struct
3. Initialize with 0.85 default in `NewEngine()`
4. Apply EMA formula in `Engine.Update()`
5. Update `EngineConfig` creation in session layer to pass config value
6. Add config field to `config.go` with default
7. Write unit tests for smoothing behavior
8. Document in README.md
9. Create smoke test procedure
10. Manual validation with real GPX route

## References

- Exponential Moving Average: https://en.wikipedia.org/wiki/Moving_average#Exponential_moving_average
- Zwift gradient smoothing discussion: https://zwiftinsider.com/gradient-changes/
- FTMS specification (no smoothing requirement)
