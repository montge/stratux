/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	flarm_nmea_edge_cases_test.go: Additional edge case tests for FLARM NMEA functions

	Tests cover:
	- relativeGpsAltToBaro: All sensor state combinations
	- computeRelativeVertical: Various altitude source scenarios
	- Additional edge cases not covered by existing tests
*/

package main

import (
	"sync"
	"testing"
	"time"
)

// resetFlarmEdgeCaseState clears global state for FLARM edge case testing
func resetFlarmEdgeCaseState() {
	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond)
	}

	// Initialize mutexes if not already initialized
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
		mySituation.muGPSPerformance = &sync.Mutex{}
		mySituation.muAttitude = &sync.Mutex{}
		mySituation.muBaro = &sync.Mutex{}
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Reset GPS state
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 0
	mySituation.GPSLongitude = 0
	mySituation.GPSAltitudeMSL = 0
	mySituation.GPSHeightAboveEllipsoid = 0
	mySituation.GPSFixQuality = 0
	mySituation.GPSLastFixLocalTime = time.Time{}
	mySituation.muGPS.Unlock()

	// Reset baro state
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 0
	mySituation.BaroLastMeasurementTime = time.Time{}
	mySituation.muBaro.Unlock()

	globalStatus.GPS_connected = false
}

// TestRelativeGpsAltToBaro_WithValidBaro tests altitude conversion with valid baro
func TestRelativeGpsAltToBaro_WithValidBaro(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 5000 // 5000 ft
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	// relVert is in meters, result is feet
	// 100m * 3.28084 = 328.08 ft
	// 5000 + 328 = 5328 ft
	alt, altIsGnss := relativeGpsAltToBaro(100)

	if altIsGnss {
		t.Error("Expected altIsGnss=false when baro is valid")
	}

	expectedAlt := int32(5328)
	if alt < expectedAlt-5 || alt > expectedAlt+5 {
		t.Errorf("Expected alt~%d, got %d", expectedAlt, alt)
	}

	t.Logf("Baro valid: relVert=100m -> alt=%d ft, altIsGnss=%v", alt, altIsGnss)
}

// TestRelativeGpsAltToBaro_WithValidGPSOnly tests altitude conversion with GPS only
func TestRelativeGpsAltToBaro_WithValidGPSOnly(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS but no baro
	mySituation.muGPS.Lock()
	mySituation.GPSAltitudeMSL = 3000 // 3000 ft
	mySituation.GPSFixQuality = 2
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// 50m * 3.28084 = 164 ft
	// 3000 + 164 = 3164 ft
	alt, altIsGnss := relativeGpsAltToBaro(50)

	if !altIsGnss {
		t.Error("Expected altIsGnss=true when only GPS is valid")
	}

	expectedAlt := int32(3164)
	if alt < expectedAlt-5 || alt > expectedAlt+5 {
		t.Errorf("Expected alt~%d, got %d", expectedAlt, alt)
	}

	t.Logf("GPS only: relVert=50m -> alt=%d ft, altIsGnss=%v", alt, altIsGnss)
}

// TestRelativeGpsAltToBaro_NoValidSensors tests altitude conversion with no valid sensors
func TestRelativeGpsAltToBaro_NoValidSensors(t *testing.T) {
	resetFlarmEdgeCaseState()

	// No valid sensors
	alt, altIsGnss := relativeGpsAltToBaro(100)

	// Should return 0, false when no sensors valid
	if alt != 0 {
		t.Errorf("Expected alt=0 with no sensors, got %d", alt)
	}
	if altIsGnss {
		t.Errorf("Expected altIsGnss=false with no sensors, got %v", altIsGnss)
	}

	t.Logf("No sensors: alt=%d, altIsGnss=%v", alt, altIsGnss)
}

// TestRelativeGpsAltToBaro_NegativeRelVert tests negative relative vertical
func TestRelativeGpsAltToBaro_NegativeRelVert(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 5000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	// -100m = -328 ft
	// 5000 - 328 = 4672 ft
	alt, altIsGnss := relativeGpsAltToBaro(-100)

	if altIsGnss {
		t.Error("Expected altIsGnss=false when baro is valid")
	}

	expectedAlt := int32(4672)
	if alt < expectedAlt-5 || alt > expectedAlt+5 {
		t.Errorf("Expected alt~%d, got %d", expectedAlt, alt)
	}

	t.Logf("Negative relVert: -100m -> alt=%d ft", alt)
}

// TestRelativeGpsAltToBaro_ZeroRelVert tests zero relative vertical
func TestRelativeGpsAltToBaro_ZeroRelVert(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 2500
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	alt, altIsGnss := relativeGpsAltToBaro(0)

	if altIsGnss {
		t.Error("Expected altIsGnss=false when baro is valid")
	}

	expectedAlt := int32(2500)
	if alt != expectedAlt {
		t.Errorf("Expected alt=%d, got %d", expectedAlt, alt)
	}

	t.Logf("Zero relVert: alt=%d ft", alt)
}

// TestComputeRelativeVertical_WithBaroAndNonGNSSAlt tests with valid baro and non-GNSS traffic alt
func TestComputeRelativeVertical_WithBaroAndNonGNSSAlt(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 5000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	ti := TrafficInfo{
		Alt:       6000, // Traffic at 6000 ft
		AltIsGNSS: false,
	}

	relVert := computeRelativeVertical(ti)

	// Traffic is 1000 ft above ownship
	// 1000 ft * 0.3048 = 304.8 m
	expectedRelVert := int32(305)
	if relVert < expectedRelVert-5 || relVert > expectedRelVert+5 {
		t.Errorf("Expected relVert~%d, got %d", expectedRelVert, relVert)
	}

	t.Logf("Baro + non-GNSS: relVert=%d m", relVert)
}

// TestComputeRelativeVertical_WithGPSOnlyAndNonGNSSAlt tests with GPS only and non-GNSS traffic alt
func TestComputeRelativeVertical_WithGPSOnlyAndNonGNSSAlt(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS but no baro
	mySituation.muGPS.Lock()
	mySituation.GPSAltitudeMSL = 5000
	mySituation.GPSFixQuality = 2
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	ti := TrafficInfo{
		Alt:       4000, // Traffic at 4000 ft (below ownship)
		AltIsGNSS: false,
	}

	relVert := computeRelativeVertical(ti)

	// Traffic is 1000 ft below ownship
	// -1000 ft * 0.3048 = -304.8 m
	expectedRelVert := int32(-305)
	if relVert < expectedRelVert-5 || relVert > expectedRelVert+5 {
		t.Errorf("Expected relVert~%d, got %d", expectedRelVert, relVert)
	}

	t.Logf("GPS only + non-GNSS: relVert=%d m", relVert)
}

// TestComputeRelativeVertical_WithGNSSAltAndValidGPS tests with GNSS traffic alt and valid GPS
func TestComputeRelativeVertical_WithGNSSAltAndValidGPS(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS with ellipsoid height
	mySituation.muGPS.Lock()
	mySituation.GPSHeightAboveEllipsoid = 5000 // Use ellipsoid height for OGN
	mySituation.GPSAltitudeMSL = 4900          // MSL is different
	mySituation.GPSFixQuality = 2
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Also set valid baro to verify GNSS alt takes precedence for GNSS traffic
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 4950
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	ti := TrafficInfo{
		Alt:       5500, // GNSS altitude from OGN
		AltIsGNSS: true, // This is GNSS altitude
	}

	relVert := computeRelativeVertical(ti)

	// Should use GPSHeightAboveEllipsoid for comparison
	// Traffic at 5500, ownship at 5000 ellipsoid
	// 500 ft * 0.3048 = 152.4 m
	expectedRelVert := int32(152)
	if relVert < expectedRelVert-5 || relVert > expectedRelVert+5 {
		t.Errorf("Expected relVert~%d (using ellipsoid), got %d", expectedRelVert, relVert)
	}

	t.Logf("GNSS alt + valid GPS: relVert=%d m (uses ellipsoid height)", relVert)
}

// TestComputeRelativeVertical_TrafficBelowOwnship tests traffic below ownship
func TestComputeRelativeVertical_TrafficBelowOwnship(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 10000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	ti := TrafficInfo{
		Alt:       5000, // Traffic 5000 ft below
		AltIsGNSS: false,
	}

	relVert := computeRelativeVertical(ti)

	// -5000 ft * 0.3048 = -1524 m
	expectedRelVert := int32(-1524)
	if relVert < expectedRelVert-10 || relVert > expectedRelVert+10 {
		t.Errorf("Expected relVert~%d, got %d", expectedRelVert, relVert)
	}

	t.Logf("Traffic below: relVert=%d m", relVert)
}

// TestComputeRelativeVertical_TrafficAboveOwnship tests traffic above ownship
func TestComputeRelativeVertical_TrafficAboveOwnship(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 5000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	ti := TrafficInfo{
		Alt:       15000, // Traffic 10000 ft above
		AltIsGNSS: false,
	}

	relVert := computeRelativeVertical(ti)

	// 10000 ft * 0.3048 = 3048 m
	expectedRelVert := int32(3048)
	if relVert < expectedRelVert-10 || relVert > expectedRelVert+10 {
		t.Errorf("Expected relVert~%d, got %d", expectedRelVert, relVert)
	}

	t.Logf("Traffic above: relVert=%d m", relVert)
}

// TestComputeRelativeVertical_SameAltitude tests traffic at same altitude
func TestComputeRelativeVertical_SameAltitude(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 5000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	ti := TrafficInfo{
		Alt:       5000, // Same altitude
		AltIsGNSS: false,
	}

	relVert := computeRelativeVertical(ti)

	if relVert != 0 {
		t.Errorf("Expected relVert=0 for same altitude, got %d", relVert)
	}

	t.Logf("Same altitude: relVert=%d m", relVert)
}

// TestComputeRelativeVertical_NoValidSensors tests with no valid sensors
func TestComputeRelativeVertical_NoValidSensors(t *testing.T) {
	resetFlarmEdgeCaseState()

	// No valid sensors (GPS or baro)
	ti := TrafficInfo{
		Alt:       5000,
		AltIsGNSS: false,
	}

	relVert := computeRelativeVertical(ti)

	// With no sensors, ownship altitude is 0, so relVert = traffic alt in meters
	// 5000 ft * 0.3048 = 1524 m
	expectedRelVert := int32(1524)
	if relVert < expectedRelVert-10 || relVert > expectedRelVert+10 {
		t.Errorf("Expected relVert~%d with no sensors, got %d", expectedRelVert, relVert)
	}

	t.Logf("No sensors: relVert=%d m", relVert)
}

// TestComputeAlarmLevel_EdgeCases tests alarm level boundary conditions
func TestComputeAlarmLevel_EdgeCases(t *testing.T) {
	testCases := []struct {
		name             string
		dist             float64
		relativeVertical int32
		expectedAlarm    uint8
	}{
		// Level 3: <500m and <100m vertical
		{"Very close and level", 400, 50, 3},
		{"At 500m boundary", 500, 50, 3},
		{"At 100m vertical boundary", 400, 100, 3},

		// Level 2: <1nm (1852m) and <1000ft (304m) vertical
		{"Medium range level", 1000, 200, 2},
		{"Just under 1nm boundary", 1851, 200, 2},
		{"Just under 1000ft vertical boundary", 1000, 303, 2},

		// Level 0: outside threat range
		{"Far away", 3000, 100, 0},
		{"High vertical separation", 500, 500, 0},
		{"Both far and high", 5000, 1000, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alarm := computeAlarmLevel(tc.dist, tc.relativeVertical)
			if alarm != tc.expectedAlarm {
				t.Errorf("Expected alarm level %d, got %d for dist=%f, relVert=%d",
					tc.expectedAlarm, alarm, tc.dist, tc.relativeVertical)
			}
		})
	}
}

// TestComputeAlarmLevel_NegativeVertical tests with negative vertical separation
func TestComputeAlarmLevel_NegativeVertical(t *testing.T) {
	// Traffic below should still trigger alarm if within range
	alarm := computeAlarmLevel(400, -50) // Close and 50m below
	if alarm != 3 {
		t.Errorf("Expected alarm level 3, got %d for traffic below", alarm)
	}

	// Traffic far below should not trigger alarm
	alarm = computeAlarmLevel(400, -500) // Close but 500m below
	if alarm != 0 {
		t.Errorf("Expected alarm level 0, got %d for traffic far below", alarm)
	}

	t.Log("Negative vertical separation test complete")
}
