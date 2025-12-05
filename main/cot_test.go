/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	cot_test.go: Unit tests for cot-in.go COT (Cursor on Target) message processing

	Tests cover:
	- processCotMessage: XML parsing
	- processCotMessage: Traffic creation and update
	- processCotMessage: Altitude conversion with/without baro
	- processCotMessage: Speed conversion
	- processCotMessage: Error handling for invalid XML
*/

package main

import (
	"sync"
	"testing"
	"time"
)

// resetCOTState clears global state for COT testing
func resetCOTState() {
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

	// Reset traffic tracking
	trafficMutex = &sync.Mutex{}
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)

	// Reset GPS and baro state
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 0
	mySituation.GPSLongitude = 0
	mySituation.GPSAltitudeMSL = 0
	mySituation.GPSFixQuality = 0
	mySituation.GPSLastFixLocalTime = time.Time{}
	mySituation.muGPS.Unlock()

	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 0
	mySituation.BaroLastMeasurementTime = time.Time{}
	mySituation.muBaro.Unlock()

	globalStatus.GPS_connected = false
}

// TestProcessCotMessage_ValidMessage tests processing a valid COT event
func TestProcessCotMessage_ValidMessage(t *testing.T) {
	resetCOTState()

	// Valid COT event XML
	cotMsg := `<event version="2.0" uid="DRONE-123" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail>
			<track speed="10.0" course="45.0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg)

	// Verify traffic was created
	trafficMutex.Lock()
	count := len(traffic)
	trafficMutex.Unlock()

	if count != 1 {
		t.Fatalf("Expected 1 traffic target from COT, got %d", count)
	}

	// Find the traffic entry
	trafficMutex.Lock()
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	// Verify position
	if ti.Lat != 47.5 {
		t.Errorf("Expected Lat=47.5, got %f", ti.Lat)
	}
	if ti.Lng != -122.3 {
		t.Errorf("Expected Lng=-122.3, got %f", ti.Lng)
	}

	// Verify position_valid
	if !ti.Position_valid {
		t.Error("Expected Position_valid=true")
	}

	// Verify UID stored
	if ti.Reg != "DRONE-123" {
		t.Errorf("Expected Reg='DRONE-123', got '%s'", ti.Reg)
	}
	if ti.Tail != "DRONE-123" {
		t.Errorf("Expected Tail='DRONE-123', got '%s'", ti.Tail)
	}

	// Verify track
	if ti.Track != 45.0 {
		t.Errorf("Expected Track=45.0, got %f", ti.Track)
	}

	// Verify speed conversion (10 m/s * 1.94384449 = ~19.4 kts)
	expectedSpeed := uint16(19)
	if ti.Speed < expectedSpeed-1 || ti.Speed > expectedSpeed+1 {
		t.Errorf("Expected Speed~%d, got %d", expectedSpeed, ti.Speed)
	}

	t.Logf("COT traffic: Lat=%f, Lng=%f, Track=%f, Speed=%d, Alt=%d",
		ti.Lat, ti.Lng, ti.Track, ti.Speed, ti.Alt)
}

// TestProcessCotMessage_InvalidXML tests handling of invalid XML
func TestProcessCotMessage_InvalidXML(t *testing.T) {
	resetCOTState()

	invalidMessages := []string{
		"",                              // Empty
		"not xml",                       // Plain text
		"<event>",                       // Incomplete
		"<event><point/></event>",       // Missing required attributes
		"<event version='2.0'></event>", // Missing point
	}

	for i, msg := range invalidMessages {
		processCotMessage(msg)
		t.Logf("Processed invalid COT message %d", i+1)
	}

	// Should not create any traffic
	trafficMutex.Lock()
	count := len(traffic)
	trafficMutex.Unlock()

	if count != 0 {
		t.Errorf("Expected 0 traffic from invalid XML, got %d", count)
	}
}

// TestProcessCotMessage_ZeroPosition tests filtering of zero lat/lon
func TestProcessCotMessage_ZeroPosition(t *testing.T) {
	resetCOTState()

	// COT event with zero coordinates (should be filtered)
	cotMsg := `<event version="2.0" uid="TEST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="0" lon="0" hae="100"/>
		<detail>
			<track speed="10.0" course="45.0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg)

	// Should not create traffic (0,0 is filtered)
	trafficMutex.Lock()
	count := len(traffic)
	trafficMutex.Unlock()

	if count != 0 {
		t.Errorf("Expected 0 traffic from zero position, got %d", count)
	}

	t.Log("Zero position correctly filtered")
}

// TestProcessCotMessage_AltitudeWithBaro tests altitude conversion with valid baro
func TestProcessCotMessage_AltitudeWithBaro(t *testing.T) {
	resetCOTState()

	// Setup valid GPS and baro
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = -122.3
	mySituation.GPSAltitudeMSL = 1000 // 1000 ft MSL
	mySituation.GPSFixQuality = 2
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 1050 // 1050 ft pressure altitude
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	// COT event with 500m HAE (height above ellipsoid)
	// 500m * 3.28084 = 1640.42 ft
	cotMsg := `<event version="2.0" uid="BARO-TEST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="500"/>
		<detail>
			<track speed="0" course="0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg)

	trafficMutex.Lock()
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	// With valid baro: alt = hae_ft - GPS_MSL + baro_alt
	// = 1640.42 - 1000 + 1050 = 1690.42 ft
	if ti.AltIsGNSS {
		t.Error("Expected AltIsGNSS=false when baro is valid")
	}

	t.Logf("Altitude with baro: %d ft, AltIsGNSS=%v", ti.Alt, ti.AltIsGNSS)
}

// TestProcessCotMessage_AltitudeWithGPSOnly tests altitude conversion with GPS only
func TestProcessCotMessage_AltitudeWithGPSOnly(t *testing.T) {
	resetCOTState()

	// Setup valid GPS but NO baro
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = -122.3
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 2
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Baro invalid (old timestamp)
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 0
	mySituation.BaroLastMeasurementTime = time.Time{} // Zero time = invalid
	mySituation.muBaro.Unlock()

	// COT event with 300m HAE
	cotMsg := `<event version="2.0" uid="GPS-ONLY" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="300"/>
		<detail>
			<track speed="0" course="0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg)

	trafficMutex.Lock()
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	// With GPS only (no baro): alt = hae_ft
	// 300m * 3.28084 = 984.25 ft
	if !ti.AltIsGNSS {
		t.Error("Expected AltIsGNSS=true when only GPS is valid")
	}

	expectedAlt := int32(984)
	if ti.Alt < expectedAlt-10 || ti.Alt > expectedAlt+10 {
		t.Errorf("Expected Alt~%d, got %d", expectedAlt, ti.Alt)
	}

	t.Logf("Altitude with GPS only: %d ft, AltIsGNSS=%v", ti.Alt, ti.AltIsGNSS)
}

// TestProcessCotMessage_AltitudeNoValidSensors tests altitude when neither GPS nor baro valid
func TestProcessCotMessage_AltitudeNoValidSensors(t *testing.T) {
	resetCOTState()

	// No valid GPS or baro
	globalStatus.GPS_connected = false

	cotMsg := `<event version="2.0" uid="NO-SENSORS" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail>
			<track speed="0" course="0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg)

	trafficMutex.Lock()
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	// With no valid sensors: alt should still be computed from HAE
	// 100m * 3.28084 = 328.08 ft
	expectedAlt := int32(328)
	if ti.Alt < expectedAlt-10 || ti.Alt > expectedAlt+10 {
		t.Errorf("Expected Alt~%d, got %d", expectedAlt, ti.Alt)
	}

	t.Logf("Altitude no sensors: %d ft, AltIsGNSS=%v", ti.Alt, ti.AltIsGNSS)
}

// TestProcessCotMessage_UpdateExisting tests updating an existing COT target
func TestProcessCotMessage_UpdateExisting(t *testing.T) {
	resetCOTState()

	// First message
	cotMsg1 := `<event version="2.0" uid="UPDATE-TEST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail>
			<track speed="10.0" course="45.0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg1)

	// Verify first message created traffic
	trafficMutex.Lock()
	count1 := len(traffic)
	trafficMutex.Unlock()

	if count1 != 1 {
		t.Fatalf("Expected 1 traffic after first message, got %d", count1)
	}

	// Second message with same UID but different position
	cotMsg2 := `<event version="2.0" uid="UPDATE-TEST" type="a-f-G" time="2024-01-01T12:01:00Z" start="2024-01-01T12:01:00Z" stale="2024-01-01T12:06:00Z">
		<point lat="47.6" lon="-122.4" hae="200"/>
		<detail>
			<track speed="20.0" course="90.0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg2)

	// Should still have only 1 target (updated, not new)
	trafficMutex.Lock()
	count2 := len(traffic)
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	if count2 != 1 {
		t.Errorf("Expected 1 traffic after update, got %d", count2)
	}

	// Verify position was updated
	if ti.Lat != 47.6 {
		t.Errorf("Expected updated Lat=47.6, got %f", ti.Lat)
	}
	if ti.Track != 90.0 {
		t.Errorf("Expected updated Track=90.0, got %f", ti.Track)
	}

	t.Logf("Updated traffic: Lat=%f, Lng=%f, Track=%f", ti.Lat, ti.Lng, ti.Track)
}

// TestProcessCotMessage_ZeroSpeed tests handling of zero speed
func TestProcessCotMessage_ZeroSpeed(t *testing.T) {
	resetCOTState()

	// Stationary target
	cotMsg := `<event version="2.0" uid="STATIONARY" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail>
			<track speed="0" course="0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg)

	trafficMutex.Lock()
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	if ti.Speed != 0 {
		t.Errorf("Expected Speed=0, got %d", ti.Speed)
	}

	if ti.Speed_valid {
		t.Error("Expected Speed_valid=false for zero speed")
	}

	t.Logf("Zero speed: Speed=%d, Speed_valid=%v", ti.Speed, ti.Speed_valid)
}

// TestProcessCotMessage_HighSpeed tests high speed conversion
func TestProcessCotMessage_HighSpeed(t *testing.T) {
	resetCOTState()

	// Fast target: 100 m/s = ~194 knots
	cotMsg := `<event version="2.0" uid="FAST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail>
			<track speed="100.0" course="270.0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg)

	trafficMutex.Lock()
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	// 100 m/s * 1.94384449 = 194.38 kts
	expectedSpeed := uint16(194)
	if ti.Speed < expectedSpeed-2 || ti.Speed > expectedSpeed+2 {
		t.Errorf("Expected Speed~%d, got %d", expectedSpeed, ti.Speed)
	}

	if !ti.Speed_valid {
		t.Error("Expected Speed_valid=true for non-zero speed")
	}

	t.Logf("High speed: Speed=%d kts, Speed_valid=%v", ti.Speed, ti.Speed_valid)
}

// TestProcessCotMessage_DEBUGMode tests processing with DEBUG enabled
func TestProcessCotMessage_DEBUGMode(t *testing.T) {
	resetCOTState()

	// Enable DEBUG mode
	origDebug := globalSettings.DEBUG
	globalSettings.DEBUG = true
	defer func() { globalSettings.DEBUG = origDebug }()

	cotMsg := `<event version="2.0" uid="DEBUG-TEST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail>
			<track speed="10.0" course="45.0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg)

	// Should still work with DEBUG enabled
	trafficMutex.Lock()
	count := len(traffic)
	trafficMutex.Unlock()

	if count != 1 {
		t.Errorf("Expected 1 traffic with DEBUG enabled, got %d", count)
	}

	t.Log("DEBUG mode processing test complete")
}

// TestProcessCotMessage_AddressGeneration tests unique address generation from UID
func TestProcessCotMessage_AddressGeneration(t *testing.T) {
	resetCOTState()

	// Process multiple COT events with different UIDs
	uids := []string{"DRONE-A", "DRONE-B", "DRONE-C", "PLANE-1", "HELI-X"}

	for _, uid := range uids {
		cotMsg := `<event version="2.0" uid="` + uid + `" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
			<point lat="47.5" lon="-122.3" hae="100"/>
			<detail>
				<track speed="10.0" course="45.0"/>
			</detail>
		</event>`
		processCotMessage(cotMsg)
	}

	// Should have 5 unique targets
	trafficMutex.Lock()
	count := len(traffic)
	trafficMutex.Unlock()

	if count != 5 {
		t.Errorf("Expected 5 traffic targets from different UIDs, got %d", count)
	}

	t.Logf("Address generation: %d unique targets from %d UIDs", count, len(uids))
}

// TestProcessCotMessage_NonICAOAddress tests that COT addresses are marked as non-ICAO
func TestProcessCotMessage_NonICAOAddress(t *testing.T) {
	resetCOTState()

	cotMsg := `<event version="2.0" uid="ADDR-TEST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail>
			<track speed="10.0" course="45.0"/>
		</detail>
	</event>`

	processCotMessage(cotMsg)

	trafficMutex.Lock()
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	// COT addresses should be non-ICAO (Addr_type = 1)
	if ti.Addr_type != 1 {
		t.Errorf("Expected Addr_type=1 (non-ICAO), got %d", ti.Addr_type)
	}

	t.Logf("Address type: %d (non-ICAO)", ti.Addr_type)
}
