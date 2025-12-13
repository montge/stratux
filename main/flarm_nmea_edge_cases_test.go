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

	// Initialize traffic mutex if not already initialized
	if trafficMutex == nil {
		trafficMutex = &sync.Mutex{}
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

// TestParseFlarmPFLAU_1090ESPriority tests that PFLAU ignores traffic when 1090ES has recent data
func TestParseFlarmPFLAU_1090ESPriority(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTrueCourse = 90.0
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)

	// Add existing traffic with 1090ES source and recent Age
	address := uint32(0xABCDEF)
	existingTraffic := TrafficInfo{
		Icao_addr:   address,
		Last_source: TRAFFIC_SOURCE_1090ES,
		Age:         2.0, // Recently seen
		Tail:        "TEST",
	}
	// Store with both possible keys
	traffic[address] = existingTraffic
	trafficMutex.Unlock()

	// Parse PFLAU message for same aircraft - should be ignored due to 1090ES priority
	// $PFLAU,<RX>,<TX>,<GPS>,<Power>,<AlarmLevel>,<RelativeBearing>,<AlarmType>,<RelativeVertical>,<RelativeDistance>,<ID>
	pflauMsg := []string{"$PFLAU", "1", "1", "2", "1", "0", "45", "0", "100", "500", "ABCDEF"}
	parseFlarmPFLAU(pflauMsg)

	// Verify traffic wasn't modified (still has 1090ES source)
	trafficMutex.Lock()
	if ti, ok := traffic[address]; ok {
		if ti.Last_source != TRAFFIC_SOURCE_1090ES {
			t.Error("Traffic source should still be 1090ES after PFLAU when 1090ES is recent")
		}
	}
	trafficMutex.Unlock()

	t.Log("PFLAU 1090ES priority test complete")
}

// TestParseFlarmPFLAA_1090ESPriority tests that PFLAA ignores traffic when 1090ES has recent data
func TestParseFlarmPFLAA_1090ESPriority(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS for distance calculation
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)

	// Add existing traffic with 1090ES source and recent Age
	address := uint32(0x123456)
	key := uint32(0)<<24 | address // ICAO address type
	existingTraffic := TrafficInfo{
		Icao_addr:   address,
		Last_source: TRAFFIC_SOURCE_1090ES,
		Age:         1.0, // Very recently seen
		Tail:        "TEST",
		Alt:         5000,
	}
	traffic[key] = existingTraffic
	trafficMutex.Unlock()

	// Parse PFLAA message for same aircraft - should be ignored due to 1090ES priority
	// $PFLAA,<AlarmLevel>,<RelativeNorth>,<RelativeEast>,<RelativeVertical>,<IDType>,<ID>,<Track>,<TurnRate>,<GroundSpeed>,<ClimbRate>,<AcftType>
	pflaaMsg := []string{"$PFLAA", "0", "1000", "500", "100", "1", "123456", "90", "0", "50", "2", "1"}
	parseFlarmPFLAA(pflaaMsg)

	// Verify traffic wasn't modified (should still have 1090ES source and original altitude)
	trafficMutex.Lock()
	if ti, ok := traffic[key]; ok {
		if ti.Last_source != TRAFFIC_SOURCE_1090ES {
			t.Error("Traffic source should still be 1090ES after PFLAA when 1090ES is recent")
		}
		if ti.Alt != 5000 {
			t.Errorf("Traffic altitude should still be 5000 after PFLAA when 1090ES is recent, got %d", ti.Alt)
		}
	}
	trafficMutex.Unlock()

	t.Log("PFLAA 1090ES priority test complete")
}

// TestParseFlarmPFLAU_OldTraffic tests that PFLAU updates traffic when 1090ES data is old
func TestParseFlarmPFLAU_OldTraffic(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTrueCourse = 90.0
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)

	// Add existing traffic with 1090ES source but OLD Age
	address := uint32(0xFEDCBA)
	existingTraffic := TrafficInfo{
		Icao_addr:   address,
		Last_source: TRAFFIC_SOURCE_1090ES,
		Age:         10.0, // Old - should be updated by FLARM
		Tail:        "OLD",
	}
	traffic[address] = existingTraffic
	trafficMutex.Unlock()

	// Parse PFLAU message - should update because 1090ES data is old
	pflauMsg := []string{"$PFLAU", "1", "1", "2", "1", "0", "45", "0", "100", "500", "FEDCBA"}
	parseFlarmPFLAU(pflauMsg)

	// Verify traffic WAS modified (now has OGN source)
	trafficMutex.Lock()
	// Check with non-ICAO key since we can't know idType from PFLAU
	key := uint32(1)<<24 | address
	if ti, ok := traffic[key]; ok {
		if ti.Last_source != TRAFFIC_SOURCE_OGN {
			t.Errorf("Traffic source should be OGN after PFLAU when 1090ES is old, got %d", ti.Last_source)
		}
	}
	trafficMutex.Unlock()

	t.Log("PFLAU old traffic update test complete")
}

// TestParseFlarmPFLAA_OldTraffic tests that PFLAA updates traffic when 1090ES data is old
func TestParseFlarmPFLAA_OldTraffic(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS for distance calculation
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)

	// Add existing traffic with 1090ES source but OLD Age
	address := uint32(0x654321)
	key := uint32(0)<<24 | address // ICAO address type
	existingTraffic := TrafficInfo{
		Icao_addr:   address,
		Last_source: TRAFFIC_SOURCE_1090ES,
		Age:         8.0, // Old - should be updated by FLARM
		Tail:        "OLD",
		Alt:         3000,
	}
	traffic[key] = existingTraffic
	trafficMutex.Unlock()

	// Parse PFLAA message - should update because 1090ES data is old
	pflaaMsg := []string{"$PFLAA", "0", "1000", "500", "100", "1", "654321", "90", "0", "50", "2", "1"}
	parseFlarmPFLAA(pflaaMsg)

	// Verify traffic WAS modified (now has OGN source)
	trafficMutex.Lock()
	if ti, ok := traffic[key]; ok {
		if ti.Last_source != TRAFFIC_SOURCE_OGN {
			t.Errorf("Traffic source should be OGN after PFLAA when 1090ES is old, got %d", ti.Last_source)
		}
	}
	trafficMutex.Unlock()

	t.Log("PFLAA old traffic update test complete")
}

// TestParseFlarmPFLAU_NegativeBearing tests negative bearing correction
func TestParseFlarmPFLAU_NegativeBearing(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS with heading that will create negative bearing
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTrueCourse = 10.0 // Low heading
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)
	trafficMutex.Unlock()

	// Parse PFLAU message with negative relative bearing
	// GPSTrueCourse (10) + RelativeBearing (-45) = -35, should wrap to 325
	pflauMsg := []string{"$PFLAU", "1", "1", "2", "1", "0", "-45", "0", "100", "500", "ABCD12"}
	parseFlarmPFLAU(pflauMsg)

	trafficMutex.Lock()
	found := false
	for _, ti := range traffic {
		if ti.Bearing >= 0 && ti.Bearing < 360 {
			found = true
			t.Logf("Bearing calculated: %.1f", ti.Bearing)
		}
	}
	trafficMutex.Unlock()

	if !found {
		t.Error("Expected to find traffic with valid bearing")
	}

	t.Log("PFLAU negative bearing test complete")
}

// TestParseFlarmPFLAA_NonICAOType tests PFLAA with non-ICAO ID type
func TestParseFlarmPFLAA_NonICAOType(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)
	trafficMutex.Unlock()

	// Parse PFLAA with non-ICAO ID type (2 = Flarm ID, should map to idType 1)
	pflaaMsg := []string{"$PFLAA", "0", "500", "300", "50", "2", "DEAD00", "180", "0", "40", "1", "8"}
	parseFlarmPFLAA(pflaaMsg)

	trafficMutex.Lock()
	// Check for non-ICAO key (idType=1)
	address := uint32(0xDEAD00)
	key := uint32(1)<<24 | address
	ti, ok := traffic[key]
	trafficMutex.Unlock()

	if !ok {
		t.Error("Expected to find traffic with non-ICAO ID type")
	} else {
		if ti.Addr_type != 1 {
			t.Errorf("Expected Addr_type 1 (non-ICAO), got %d", ti.Addr_type)
		}
		t.Logf("Traffic registered with Addr_type=%d", ti.Addr_type)
	}

	t.Log("PFLAA non-ICAO ID type test complete")
}

// TestParseFlarmPFLAU_ShortMessage tests PFLAU with fewer than 11 fields
func TestParseFlarmPFLAU_ShortMessage(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTrueCourse = 90.0
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)
	initialCount := len(traffic)
	trafficMutex.Unlock()

	// Test with only 10 fields (missing ID) - should be ignored
	pflauMsg := []string{"$PFLAU", "1", "1", "2", "1", "0", "45", "0", "100", "500"}
	parseFlarmPFLAU(pflauMsg)

	// Verify no traffic was added
	trafficMutex.Lock()
	if len(traffic) != initialCount {
		t.Error("Traffic should not be added for PFLAU with fewer than 11 fields")
	}
	trafficMutex.Unlock()

	t.Log("PFLAU short message test complete")
}

// TestParseFlarmPFLAU_EmptyFields tests PFLAU with empty required fields
func TestParseFlarmPFLAU_EmptyFields(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTrueCourse = 90.0
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)
	trafficMutex.Unlock()

	// Test with empty ID field (message[10])
	pflauMsg1 := []string{"$PFLAU", "1", "1", "2", "1", "0", "45", "0", "100", "500", ""}
	parseFlarmPFLAU(pflauMsg1)

	trafficMutex.Lock()
	if len(traffic) > 0 {
		t.Error("Traffic should not be added for PFLAU with empty ID field")
	}
	traffic = make(map[uint32]TrafficInfo)
	trafficMutex.Unlock()

	// Test with empty RelativeDistance field (message[9])
	pflauMsg2 := []string{"$PFLAU", "1", "1", "2", "1", "0", "45", "0", "100", "", "ABCDEF"}
	parseFlarmPFLAU(pflauMsg2)

	trafficMutex.Lock()
	if len(traffic) > 0 {
		t.Error("Traffic should not be added for PFLAU with empty RelativeDistance field")
	}
	traffic = make(map[uint32]TrafficInfo)
	trafficMutex.Unlock()

	// Test with empty RelativeVertical field (message[8])
	pflauMsg3 := []string{"$PFLAU", "1", "1", "2", "1", "0", "45", "0", "", "500", "ABCDEF"}
	parseFlarmPFLAU(pflauMsg3)

	trafficMutex.Lock()
	if len(traffic) > 0 {
		t.Error("Traffic should not be added for PFLAU with empty RelativeVertical field")
	}
	traffic = make(map[uint32]TrafficInfo)
	trafficMutex.Unlock()

	// Test with empty RelativeBearing field (message[6])
	pflauMsg4 := []string{"$PFLAU", "1", "1", "2", "1", "0", "", "0", "100", "500", "ABCDEF"}
	parseFlarmPFLAU(pflauMsg4)

	trafficMutex.Lock()
	if len(traffic) > 0 {
		t.Error("Traffic should not be added for PFLAU with empty RelativeBearing field")
	}
	trafficMutex.Unlock()

	t.Log("PFLAU empty fields test complete")
}

// TestParseFlarmPFLAU_NoGPS tests PFLAU processing when GPS is invalid
func TestParseFlarmPFLAU_NoGPS(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup INVALID GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 0
	mySituation.GPSLongitude = 0
	mySituation.GPSFixQuality = 0                 // No fix
	mySituation.GPSLastFixLocalTime = time.Time{} // Zero time
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = false

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)
	trafficMutex.Unlock()

	// Valid PFLAU message but no GPS - should return early
	pflauMsg := []string{"$PFLAU", "1", "1", "2", "1", "0", "45", "0", "100", "500", "ABCDEF"}
	parseFlarmPFLAU(pflauMsg)

	// Verify no traffic was added
	trafficMutex.Lock()
	if len(traffic) > 0 {
		t.Error("Traffic should not be added for PFLAU when GPS is invalid")
	}
	trafficMutex.Unlock()

	t.Log("PFLAU no GPS test complete")
}

// TestParseFlarmPFLAU_TailFromMessage tests PFLAU with tail in ID field
func TestParseFlarmPFLAU_TailFromMessage(t *testing.T) {
	resetFlarmEdgeCaseState()

	// Setup valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTrueCourse = 90.0
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)
	trafficMutex.Unlock()

	// Parse PFLAU message with tail in ID field (IDIDID!TAIL syntax)
	pflauMsg := []string{"$PFLAU", "1", "1", "2", "1", "0", "45", "0", "100", "500", "ABCDEF!N12345"}
	parseFlarmPFLAU(pflauMsg)

	// Verify tail was set
	trafficMutex.Lock()
	found := false
	for _, ti := range traffic {
		if ti.Tail == "N12345" {
			found = true
			t.Logf("Traffic tail set from message: %s", ti.Tail)
		}
	}
	trafficMutex.Unlock()

	if !found {
		t.Error("Expected to find traffic with tail from PFLAU message")
	}

	t.Log("PFLAU tail from message test complete")
}
