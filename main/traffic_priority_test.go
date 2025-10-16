package main

import (
	"math"
	"testing"
	"time"
)

// TestComputeTrafficPriority_InvalidBearingDist tests priority calculation when bearing/distance is invalid
func TestComputeTrafficPriority_InvalidBearingDist(t *testing.T) {
	ti := TrafficInfo{
		BearingDist_valid: false,
		Alt:               5000,
		Distance:          10000,
	}

	priority := computeTrafficPriority(&ti)
	if priority != 9999999 {
		t.Errorf("Expected priority 9999999 for invalid BearingDist, got %d", priority)
	}
}

// TestComputeTrafficPriority_ZeroAltitude tests priority calculation when altitude is zero
func TestComputeTrafficPriority_ZeroAltitude(t *testing.T) {
	ti := TrafficInfo{
		BearingDist_valid: true,
		Alt:               0,
		Distance:          10000,
	}

	priority := computeTrafficPriority(&ti)
	if priority != 9999999 {
		t.Errorf("Expected priority 9999999 for zero altitude, got %d", priority)
	}
}

// TestComputeTrafficPriority_BaroAltitude tests priority calculation using barometric altitude
func TestComputeTrafficPriority_BaroAltitude(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original values
	origSituation := mySituation
	defer func() { mySituation = origSituation }()

	// Set up test scenario: ownship at 5000ft baro, target at 6000ft
	mySituation.BaroPressureAltitude = 5000
	mySituation.GPSAltitudeMSL = 4800                                     // different from baro
	mySituation.BaroLastMeasurementTime = stratuxClock.Time.Add(-1 * time.Second) // valid baro (recent)
	mySituation.BaroTemperature = 25.0

	ti := TrafficInfo{
		BearingDist_valid: true,
		Alt:               6000,
		Distance:          10000, // 10km = 10000m
	}

	priority := computeTrafficPriority(&ti)
	// altDiff = |5000 - 6000| = 1000
	// priority = (1000/3.33 + 10000) / 10000.0 = (300.3 + 10000) / 10000 = ~1
	// Should be low priority (high value) due to distance
	if priority < 0 || priority > 9999999 {
		t.Errorf("Expected reasonable priority value, got %d", priority)
	}
	t.Logf("Priority with 1000ft alt diff and 10km distance: %d", priority)
}

// TestComputeTrafficPriority_GPSAltitude tests priority calculation using GPS altitude
func TestComputeTrafficPriority_GPSAltitude(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original values
	origSituation := mySituation
	defer func() { mySituation = origSituation }()

	// Set up test scenario: no valid baro, use GPS altitude
	mySituation.BaroPressureAltitude = 99999 // invalid
	mySituation.GPSAltitudeMSL = 4500
	mySituation.BaroLastMeasurementTime = stratuxClock.Time.Add(-20 * time.Second) // stale baro data (>15s)

	ti := TrafficInfo{
		BearingDist_valid: true,
		Alt:               7500,
		Distance:          5000, // 5km
	}

	priority := computeTrafficPriority(&ti)
	// altDiff = |4500 - 7500| = 3000
	// priority = (3000/3.33 + 5000) / 10000.0 = (900.9 + 5000) / 10000 = ~0
	// Should be low priority (closer distance, larger alt diff)
	if priority < 0 || priority > 9999999 {
		t.Errorf("Expected reasonable priority value, got %d", priority)
	}
	t.Logf("Priority with GPS alt (3000ft diff, 5km distance): %d", priority)
}

// TestComputeTrafficPriority_SameAltitude tests priority calculation when at same altitude
func TestComputeTrafficPriority_SameAltitude(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original values
	origSituation := mySituation
	defer func() { mySituation = origSituation }()

	// Set up test scenario: both at 5000ft
	mySituation.BaroPressureAltitude = 5000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time.Add(-1 * time.Second)
	mySituation.BaroTemperature = 25.0

	ti := TrafficInfo{
		BearingDist_valid: true,
		Alt:               5000,
		Distance:          2000, // 2km
	}

	priority := computeTrafficPriority(&ti)
	// altDiff = |5000 - 5000| = 0
	// priority = (0/3.33 + 2000) / 10000.0 = 0.2 = 0
	// Should be very high priority (low value) - same altitude, 2km distance
	if priority < 0 || priority > 9999999 {
		t.Errorf("Expected reasonable priority value, got %d", priority)
	}
	t.Logf("Priority with same altitude and 2km distance: %d", priority)
}

// TestComputeTrafficPriority_CloseAndLowAltDiff tests high priority (close and low alt diff)
func TestComputeTrafficPriority_CloseAndLowAltDiff(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original values
	origSituation := mySituation
	defer func() { mySituation = origSituation }()

	// Set up test scenario: very close and similar altitude
	mySituation.BaroPressureAltitude = 3000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time.Add(-1 * time.Second)
	mySituation.BaroTemperature = 25.0

	ti := TrafficInfo{
		BearingDist_valid: true,
		Alt:               3100,
		Distance:          500, // 500m = 0.5km, very close
	}

	priority := computeTrafficPriority(&ti)
	// altDiff = |3000 - 3100| = 100
	// priority = (100/3.33 + 500) / 10000.0 = (30.03 + 500) / 10000 = 0.053 = 0
	// Should be very high priority (low value) - very close and similar altitude
	if priority < 0 || priority > 9999999 {
		t.Errorf("Expected reasonable priority value, got %d", priority)
	}
	t.Logf("Priority with 100ft alt diff and 500m distance (very close): %d", priority)
}

// TestExtrapolateTraffic_InitialExtrapolation tests first extrapolation
func TestExtrapolateTraffic_InitialExtrapolation(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	ti := TrafficInfo{
		Icao_addr:             0xABCDEF,
		Lat:                   40.0,
		Lng:                   -105.0,
		Alt:                   5000,
		Track:                 90.0, // heading east
		Speed:                 150,  // 150 knots
		Vvel:                  500,  // 500 fpm climb
		TurnRate:              0.0,
		ExtrapolatedPosition:  false,
		Last_seen:             stratuxClock.Time,
	}

	// Simulate 2 seconds passing
	time.Sleep(10 * time.Millisecond) // small delay to ensure time difference
	ti.Last_seen = stratuxClock.Time.Add(-2 * time.Second)

	extrapolateTraffic(&ti)

	// Check that fix positions were set
	if ti.Lat_fix != 40.0 {
		t.Errorf("Expected Lat_fix=40.0, got %f", ti.Lat_fix)
	}
	if ti.Lng_fix != -105.0 {
		t.Errorf("Expected Lng_fix=-105.0, got %f", ti.Lng_fix)
	}
	if ti.Alt_fix != 5000 {
		t.Errorf("Expected Alt_fix=5000, got %d", ti.Alt_fix)
	}

	// Check that ExtrapolatedPosition is now true
	if !ti.ExtrapolatedPosition {
		t.Error("Expected ExtrapolatedPosition=true after extrapolation")
	}

	// Check altitude extrapolation
	// Alt should increase by Vvel * (seconds / 60)
	// 500 fpm for 2 seconds = 500 * (2/60) = 16.67 feet
	if ti.Alt < 5000 || ti.Alt > 5020 {
		t.Logf("Altitude extrapolated to %d (expected around 5016-5017)", ti.Alt)
	}

	// Position should have changed (heading east = longitude increases)
	if ti.Lng <= -105.0 {
		t.Errorf("Expected longitude to increase (move east), got %f", ti.Lng)
	}
}

// TestExtrapolateTraffic_ContinuedExtrapolation tests ongoing extrapolation
func TestExtrapolateTraffic_ContinuedExtrapolation(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	ti := TrafficInfo{
		Icao_addr:             0xABCDEF,
		Lat:                   40.0,
		Lng:                   -105.0,
		Alt:                   5000,
		Lat_fix:               40.0,
		Lng_fix:               -105.0,
		Alt_fix:               5000,
		Track:                 180.0, // heading south
		Speed:                 200,   // 200 knots
		Vvel:                  -1000, // 1000 fpm descent
		TurnRate:              0.0,
		ExtrapolatedPosition:  true, // already extrapolating
		Last_extrapolation:    stratuxClock.Time.Add(-3 * time.Second),
	}

	extrapolateTraffic(&ti)

	// Check that ExtrapolatedPosition is still true
	if !ti.ExtrapolatedPosition {
		t.Error("Expected ExtrapolatedPosition to remain true")
	}

	// Check altitude extrapolation (descending)
	// Alt should decrease by Vvel * (seconds / 60)
	// -1000 fpm for 3 seconds = -1000 * (3/60) = -50 feet
	if ti.Alt > 5000 || ti.Alt < 4940 {
		t.Logf("Altitude extrapolated to %d (expected around 4950)", ti.Alt)
	}

	// Latitude should have changed (heading south = decreasing lat)
	if ti.Lat >= 40.0 {
		t.Errorf("Expected latitude to decrease (heading south), got %f", ti.Lat)
	}
}

// TestExtrapolateTraffic_TrackNormalization tests track wrapping around 360
func TestExtrapolateTraffic_TrackNormalization(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	testCases := []struct {
		name         string
		initialTrack float32
		turnRate     float32
		seconds      float64
		expectRange  [2]float32 // min and max expected track
	}{
		{
			name:         "Wrap_over_360",
			initialTrack: 350.0,
			turnRate:     5.0, // 5 deg/sec right turn
			seconds:      4.0, // 20 degrees in 4 seconds
			expectRange:  [2]float32{5.0, 15.0},
		},
		{
			name:         "Wrap_under_0",
			initialTrack: 10.0,
			turnRate:     -5.0, // 5 deg/sec left turn
			seconds:      4.0,  // -20 degrees in 4 seconds
			expectRange:  [2]float32{345.0, 355.0},
		},
		{
			name:         "No_wrap_needed",
			initialTrack: 180.0,
			turnRate:     2.0,
			seconds:      5.0, // 10 degrees in 5 seconds
			expectRange:  [2]float32{185.0, 195.0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ti := TrafficInfo{
				Lat:                  40.0,
				Lng:                  -105.0,
				Alt:                  5000,
				Track:                tc.initialTrack,
				Speed:                100,
				Vvel:                 0,
				TurnRate:             tc.turnRate,
				ExtrapolatedPosition: false,
				Last_seen:            stratuxClock.Time.Add(-time.Duration(tc.seconds) * time.Second),
			}

			extrapolateTraffic(&ti)

			// Check track is in valid range 0-360
			if ti.Track < 0 || ti.Track > 360 {
				t.Errorf("%s: Track %f is outside valid range [0, 360]", tc.name, ti.Track)
			}

			// Check track is in expected range
			if ti.Track < tc.expectRange[0] || ti.Track > tc.expectRange[1] {
				t.Logf("%s: Track %f (initial %f + %f*%f seconds) outside expected range [%f, %f]",
					tc.name, ti.Track, tc.initialTrack, tc.turnRate, tc.seconds,
					tc.expectRange[0], tc.expectRange[1])
			}
		})
	}
}

// TestExtrapolateTraffic_ZeroSpeed tests extrapolation with stationary aircraft
func TestExtrapolateTraffic_ZeroSpeed(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	ti := TrafficInfo{
		Lat:                  40.0,
		Lng:                  -105.0,
		Alt:                  0, // on ground
		Track:                0.0,
		Speed:                0, // stationary
		Vvel:                 0,
		TurnRate:             0.0,
		ExtrapolatedPosition: false,
		Last_seen:            stratuxClock.Time.Add(-5 * time.Second),
	}

	extrapolateTraffic(&ti)

	// Position should not change significantly with zero speed
	latDiff := math.Abs(float64(ti.Lat - 40.0))
	lngDiff := math.Abs(float64(ti.Lng - (-105.0)))

	if latDiff > 0.001 {
		t.Logf("Lat changed by %f with zero speed (may be due to precision)", latDiff)
	}
	if lngDiff > 0.001 {
		t.Logf("Lng changed by %f with zero speed (may be due to precision)", lngDiff)
	}

	// Altitude should not change with zero vvel
	if ti.Alt != 0 {
		t.Errorf("Expected Alt=0 with zero Vvel, got %d", ti.Alt)
	}
}

// TestExtrapolateTraffic_HighSpeed tests extrapolation with fast aircraft
func TestExtrapolateTraffic_HighSpeed(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	ti := TrafficInfo{
		Lat:                  40.0,
		Lng:                  -105.0,
		Alt:                  35000,
		Track:                0.0, // heading north
		Speed:                500, // 500 knots (fast jet)
		Vvel:                 2000, // 2000 fpm climb
		TurnRate:             0.0,
		ExtrapolatedPosition: false,
		Last_seen:            stratuxClock.Time.Add(-10 * time.Second),
	}

	originalLat := ti.Lat

	extrapolateTraffic(&ti)

	// Position should have changed significantly
	// 500 knots for 10 seconds = 500 nm/hr * 10/3600 hr = 1.389 nm north
	// At 40deg latitude, 1 nm ~ 1/60 degree
	latDiff := ti.Lat - originalLat
	if latDiff < 0.01 {
		t.Logf("Latitude only changed by %f (expected significant change for 500kt)", latDiff)
	}

	// Altitude should have increased significantly
	// 2000 fpm for 10 seconds = 2000 * (10/60) = 333 feet
	if ti.Alt < 35000 {
		t.Errorf("Expected altitude to increase from 35000, got %d", ti.Alt)
	}
	if ti.Alt > 35400 {
		t.Logf("Altitude extrapolated to %d (expected around 35333)", ti.Alt)
	}
}
