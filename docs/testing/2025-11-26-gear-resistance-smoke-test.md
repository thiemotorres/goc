# Gear Resistance Smoke Test

**Date:** 2025-11-26
**Tester:** [Manual testing required]
**Mode:** Mock trainer

## Test Results

### FREE Mode
- [ ] Shift up increases resistance
- [ ] Shift down decreases resistance
- [ ] Gear display updates correctly
- [ ] Speed changes with gear

### SIM Mode (if route available)
- [ ] Shift up increases resistance
- [ ] Shift down decreases resistance
- [ ] Gradient + gear both affect resistance
- [ ] Feels realistic on climbs

## Manual Test Procedure

### Prerequisites
1. Build application: `go build -o goc ./cmd`
2. Run mock trainer: `./goc ride --mock`

### Test Steps for FREE Mode

1. Start ride in FREE mode
2. Press up arrow key to shift up (e.g., 50x17 → 50x15)
   - Expected: Gear display should change
   - Expected: Resistance value should increase
   - Expected: Speed should increase (at same mock cadence)
3. Press down arrow key to shift down (e.g., 50x15 → 50x17)
   - Expected: Gear display should change
   - Expected: Resistance value should decrease
   - Expected: Speed should decrease (at same mock cadence)
4. Repeat shifts several times to verify consistency
5. Press q to quit

### Test Steps for SIM Mode (with route)

1. Load a route with varied gradient
2. Switch to SIM mode
3. On flat section:
   - Shift up → resistance should increase
   - Shift down → resistance should decrease
4. On uphill section:
   - Observe higher base resistance due to gradient
   - Shift up → resistance should increase further
   - Shift down → resistance should decrease but stay higher than flat
5. On downhill section:
   - Observe lower or minimal resistance
   - Gear changes should still affect resistance proportionally

### Expected Behavior

- Gear display updates immediately when shifting
- Resistance changes proportionally with gear ratio changes
- Higher gear ratio (harder gear) = higher resistance
- Lower gear ratio (easier gear) = lower resistance
- Speed increases with harder gear at same cadence
- In SIM mode, both gradient and gear affect resistance
- Resistance values stay within 0-100 range (clamped)

## Issues Found

[List any issues discovered during manual testing]

## Notes

[Any observations or recommendations]

## Automated Test Coverage

The following aspects are covered by automated unit tests:
- Wheel force calculation (air drag, rolling resistance, gradient)
- Pedal force calculation with gear ratio
- Force-to-resistance mapping with clamping
- CalculateResistance integration with gear ratio
- Engine.Update gear ratio effects in SIM mode
- Engine.Update gear ratio effects in FREE mode
- Gear ratio proportionality verification

All automated tests passing as of 2025-11-26.
