# Gear-Based Resistance Physics Design

**Date:** 2025-11-26
**Status:** Approved
**Problem:** Gear shifting has no effect on felt resistance at the pedals

## Problem Statement

Currently, when shifting gears (↑/↓ keys):
- Gear display updates correctly (e.g., "50x17" → "50x15")
- Speed calculation changes slightly
- BUT: Felt resistance at pedals doesn't change perceptibly

This occurs in both SIM and FREE modes.

## Root Cause

The resistance calculation in `internal/simulation/physics.go:20-36` only considers:
- Air resistance (0.005 × speed²)
- Rolling resistance (constant 2.0)
- Gradient resistance (weight × gradient)

These represent forces at the **wheel**, not at the **pedals**.

**Missing:** Gear ratio creates mechanical disadvantage. A higher gear ratio means more force needed at the pedals to achieve the same wheel force.

Currently:
```
resistance = airResistance + rollingResistance + gradientResistance
```

This ignores that shifting from 50x21 to 50x17 (~24% harder gear) should require ~24% more force at the pedals.

## Physics Solution

**Key relationship:**
- Power is conserved: P_pedal = P_wheel
- Power = Force × Velocity
- Therefore: F_pedal × v_pedal = F_wheel × v_wheel
- Since v_wheel = v_pedal × gear_ratio
- **Result: F_pedal = F_wheel × gear_ratio**

**Implementation:**
1. Calculate total resistance force at wheel (Newtons)
2. Scale by gear ratio to get pedal force
3. Map to trainer resistance level (0-100)

### Force Calculations

**Air Drag:**
```
F_air = 0.5 × ρ × Cd × A × v²
Where:
  ρ = 1.225 kg/m³ (air density)
  Cd × A = 0.3 (drag coefficient × frontal area for cycling)
  v = speed in m/s

F_air ≈ 0.184 × v²
```

**Rolling Resistance:**
```
F_roll = Crr × m × g
Where:
  Crr = 0.005 (rolling coefficient for road tires)
  m = rider weight + 10kg (bike weight)
  g = 9.81 m/s²

F_roll ≈ 0.049 × (weight + 10)
```

**Gradient Resistance:**
```
F_grade = m × g × sin(θ) ≈ m × g × (gradient/100)
Where:
  gradient is in percent

F_grade ≈ 0.0981 × (weight + 10) × gradient
```

**Total:**
```
wheelForce = F_air + F_roll + F_grade
pedalForce = wheelForce × gearRatio
resistance = pedalForce × scalingFactor
```

### Scaling Factor

Map force (Newtons) to resistance level (0-100):
- Typical pedal forces: 150-400N
- Scaling factor: ~0.2 (makes 200N → 40, 400N → 80)
- **Configurable** for user tuning

## Implementation Changes

### 1. `internal/simulation/physics.go`

**Current function signature:**
```go
func CalculateResistance(speedKmh, gradientPercent, weightKg float64) float64
```

**New structure:**
```go
// Helper: Calculate total force at wheel (Newtons)
func CalculateWheelForce(speedKmh, gradientPercent, weightKg float64) float64

// Helper: Apply gear mechanical disadvantage
func CalculatePedalForce(wheelForce, gearRatio float64) float64

// Helper: Map force to 0-100 resistance scale
func MapForceToResistance(pedalForce, scalingFactor float64) float64

// Main function: Calculate trainer resistance level
func CalculateResistance(speedKmh, gradientPercent, weightKg, gearRatio float64) float64
```

### 2. `internal/simulation/simulation.go`

**Update `Engine.Update()` method:**
- Pass `gearRatio` to `CalculateResistance()`
- Line 79: Add `e.gears.Ratio()` parameter

### 3. `internal/config/config.go` (Optional Enhancement)

Add tunable parameter:
```go
type BikeConfig struct {
    // ... existing fields ...
    ResistanceScaling float64 // Default: 0.2
}
```

Allows users to adjust resistance feel without code changes.

## Testing Strategy

1. **Unit tests:** Verify force calculations with known inputs
2. **Integration test:** Shift gears, verify resistance changes proportionally
3. **Manual testing:** Ride and validate feel
   - Start in middle gear
   - Shift up → should feel harder immediately
   - Shift down → should feel easier
   - Gradient changes should still work

## Expected Behavior After Fix

**Shifting to harder gear (e.g., 50x21 → 50x17):**
- Gear ratio increases: 2.38 → 2.94 (+24%)
- At same cadence and gradient:
  - Speed increases slightly
  - Resistance increases ~24%
  - Feels noticeably harder

**FREE mode:**
- Gear shifting provides virtual gearing feel
- Can maintain preferred cadence by shifting

**SIM mode:**
- Gradient + gear both affect resistance
- Encourages realistic gear management on climbs

## Migration Notes

- **Breaking change:** Resistance levels will be different after this change
- Existing rides may feel harder/easier depending on gear used
- Users may need to adjust expectations or tune scaling factor

## Future Enhancements

1. Add inertia simulation (heavier gears resist cadence changes)
2. Cross-chaining penalty (discourage extreme gear combinations)
3. Gear shift smoothing (gradual resistance transition)
