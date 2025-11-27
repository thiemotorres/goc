# Gradient Smoothing Smoke Test

**Date:** 2025-11-26
**Feature:** Gradient smoothing using exponential moving average (EMA)
**Purpose:** Validate gradient smoothing behavior with real GPX routes and trainer

## Test Objectives

Verify that:
1. Resistance changes are smooth and gradual (not jerky)
2. GPS elevation noise doesn't cause frequent resistance fluctuations
3. Real climbs are still detected and feel realistic
4. Configuration changes (gradient_smoothing value) affect behavior as expected

## Prerequisites

- Compiled build of the application with gradient smoothing feature
- Bluetooth trainer connected
- GPX route file with known characteristics:
  - **Flat route:** Known to be flat terrain but may have GPS noise
  - **Hilly route:** Route with at least one sustained 2-3% climb
- Power meter or ability to provide consistent pedaling effort

## Test Procedure

### Test 1: Flat Terrain with Default Smoothing (α=0.85)

**Setup:**
1. Load a GPX route known to be flat (e.g., user's daily flat route)
2. Ensure config has default `gradient_smoothing = 0.85` (or not set)
3. Start ride in SIM mode

**Execution:**
1. Pedal at consistent power (~150-200W) and cadence (~80 RPM)
2. Ride for 2-3 minutes on the "flat" section
3. Observe resistance changes in the UI

**Expected Results:**
- ✅ Resistance changes **less than once every 5 seconds** on average
- ✅ Resistance feels **smooth and gradual** (not sudden jumps)
- ✅ No jarring or jerky resistance changes
- ✅ Gradient display updates gradually, not jumping wildly

**Failure Indicators:**
- ❌ Resistance changes every 1-2 seconds
- ❌ Sudden, jarring resistance changes
- ❌ Gradient display jumping between values like -5%, +3%, -2%

### Test 2: Sustained Climb Detection

**Setup:**
1. Load a GPX route with a known sustained climb (2-5% gradient, 500m+ length)
2. Use default `gradient_smoothing = 0.85`
3. Start ride in SIM mode

**Execution:**
1. Pedal at consistent cadence (~80 RPM) through flat section
2. Continue into the sustained climb
3. Observe resistance increase
4. Continue through the climb

**Expected Results:**
- ✅ Resistance **gradually increases** as you enter the climb
- ✅ Increased resistance is **sustained** during the climb (not dropping back down)
- ✅ Resistance increase feels **realistic** (takes 20-30 seconds to stabilize)
- ✅ Gradient display gradually rises to match actual climb gradient

**Failure Indicators:**
- ❌ Resistance doesn't increase on the climb
- ❌ Resistance increases but immediately drops back down
- ❌ No noticeable difference between flat and climbing sections

### Test 3: Configuration Tuning (Optional)

**Setup:**
1. Edit config file to set `gradient_smoothing = 0.5` (more responsive)
2. Reload application
3. Ride same flat route from Test 1

**Execution:**
1. Pedal at consistent effort on flat section
2. Compare feel to Test 1 (α=0.85)

**Expected Results:**
- ✅ Resistance changes **more frequently** than default (α=0.85)
- ✅ Resistance is **more responsive** to gradient changes
- ✅ May feel **slightly less smooth** than default

**Then:**
1. Edit config to set `gradient_smoothing = 0.95` (very smooth)
2. Reload and ride same flat route

**Expected Results:**
- ✅ Resistance is **very smooth**, almost no changes on flat
- ✅ May feel **slower to respond** to real climbs
- ✅ Very gradual transitions

### Test 4: Smoothing Disabled (α=0.0)

**Setup:**
1. Edit config to set `gradient_smoothing = 0.0`
2. Reload application
3. Ride flat route from Test 1

**Execution:**
1. Pedal on flat section with GPS noise
2. Observe resistance behavior

**Expected Results:**
- ✅ Resistance changes **frequently** (every 1-2 seconds)
- ✅ May feel **jerky or unrealistic**
- ✅ Gradient display jumps around
- ✅ Behavior matches **old (pre-smoothing) behavior**

**Purpose:** Verify that α=0.0 disables smoothing (backward compatibility)

## Success Criteria

The feature passes smoke testing if:
- ✅ Test 1 passes: Resistance changes < once per 5 seconds on flat terrain
- ✅ Test 2 passes: Climbs are detected and resistance increases appropriately
- ✅ Test 3 passes: Configuration changes affect smoothing behavior as expected
- ✅ Test 4 passes: Smoothing can be disabled (α=0.0 restores old behavior)
- ✅ No crashes or errors during testing
- ✅ Overall ride feel is **realistic and natural**

## Failure Cases

If any test fails, investigate:
1. Check config value is being read correctly (print/log cfg.Bike.GradientSmoothing)
2. Verify Engine.smoothingFactor is initialized correctly
3. Verify EMA formula is being applied in Update()
4. Check that smoothed gradient (not raw) is used for resistance calculation
5. Review unit and integration test results

## Notes

- Smoke testing requires subjective assessment of "feel" - use best judgment
- GPS noise varies by device and conditions - some routes may be noisier than others
- Trainer response time (~300-1000ms) is independent of gradient smoothing
- Test with trainer you'll actually use (different trainers have different characteristics)

## Test Environment

- **Tester:** [Name]
- **Date:** [Date]
- **Build:** [Git commit hash or version]
- **Trainer:** [Trainer model]
- **GPX Routes Used:** [Route names/descriptions]
- **Config Values Tested:** [e.g., 0.0, 0.5, 0.85, 0.95]

## Test Results

[Record results here after testing]

### Test 1: Flat Terrain
- **Status:** [ ] Pass / [ ] Fail
- **Notes:**

### Test 2: Sustained Climb
- **Status:** [ ] Pass / [ ] Fail
- **Notes:**

### Test 3: Configuration Tuning
- **Status:** [ ] Pass / [ ] Fail
- **Notes:**

### Test 4: Smoothing Disabled
- **Status:** [ ] Pass / [ ] Fail
- **Notes:**

### Overall Assessment
- **Status:** [ ] Pass / [ ] Fail
- **Comments:**

## Automated Test Coverage

The following aspects are covered by automated unit tests:
- EMA formula correctness (0.85 * old + 0.15 * new)
- Smoothing factor initialization from config
- Smoothed gradient clamping to [-20%, 20%]
- First update initialization (smoothed = raw on first call)
- Integration with resistance calculation
- Configuration value validation (0.0-0.95 range)

All automated tests passing as of 2025-11-26.
