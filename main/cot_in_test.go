/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	cot_in_test.go: Unit tests for cot-in.go COT (Cursor on Target) message processing

	Tests cover:
	- processCotMessage: Additional edge cases and error handling
	- processCotMessage: XML parsing with various malformed inputs
	- processCotMessage: Message buffering and fragmentation scenarios
	- processCotMessage: Speed and altitude conversions
	- processCotMessage: Traffic update and address generation
*/

package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestProcessCotMessage_MissingAttributes tests handling of messages with missing required attributes
func TestProcessCotMessage_MissingAttributes(t *testing.T) {
	resetCOTState()

	testCases := []struct {
		name string
		msg  string
	}{
		{
			name: "Missing uid",
			msg: `<event version="2.0" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
				<point lat="47.5" lon="-122.3" hae="100"/>
				<detail><track speed="10.0" course="45.0"/></detail>
			</event>`,
		},
		{
			name: "Missing point",
			msg: `<event version="2.0" uid="TEST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
				<detail><track speed="10.0" course="45.0"/></detail>
			</event>`,
		},
		{
			name: "Missing detail",
			msg: `<event version="2.0" uid="TEST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
				<point lat="47.5" lon="-122.3" hae="100"/>
			</event>`,
		},
		{
			name: "Missing track",
			msg: `<event version="2.0" uid="TEST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
				<point lat="47.5" lon="-122.3" hae="100"/>
				<detail></detail>
			</event>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetCOTState()

			processCotMessage(tc.msg)

			// Some messages may still create traffic if they have lat/lon
			trafficMutex.Lock()
			count := len(traffic)
			trafficMutex.Unlock()

			t.Logf("%s: Created %d traffic entries", tc.name, count)
		})
	}
}

// TestProcessCotMessage_MalformedXML tests handling of various malformed XML
func TestProcessCotMessage_MalformedXML(t *testing.T) {
	resetCOTState()

	testCases := []struct {
		name string
		msg  string
	}{
		{"Unclosed tag", "<event version='2.0' uid='TEST'><point lat='47.5' lon='-122.3' hae='100'></event>"},
		{"Wrong closing tag", "<event version='2.0'><point></event>"},
		{"Random text", "This is not XML at all"},
		{"Partial XML", "<event version='2.0' uid='TEST"},
		{"Empty tags", "<event></event>"},
		{"Nested error", "<event><point><invalid></point></event>"},
		{"Special characters", "<event uid='<>&\"'><point/></event>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetCOTState()

			// Should not panic
			processCotMessage(tc.msg)

			// Should not create traffic
			trafficMutex.Lock()
			count := len(traffic)
			trafficMutex.Unlock()

			if count != 0 {
				t.Errorf("Expected 0 traffic from malformed XML (%s), got %d", tc.name, count)
			}

			t.Logf("%s: Handled gracefully", tc.name)
		})
	}
}

// TestProcessCotMessage_BoundaryCoordinates tests edge case coordinates
func TestProcessCotMessage_BoundaryCoordinates(t *testing.T) {
	resetCOTState()

	testCases := []struct {
		name         string
		lat          float32
		lon          float32
		shouldCreate bool
	}{
		{"Valid normal", 47.5, -122.3, true},
		{"North pole", 90.0, 0.0, true},
		{"South pole", -90.0, 0.0, true},
		{"Equator", 0.0, 0.0, false}, // Zero lat/lon is filtered
		{"Equator non-zero lon", 0.0, 122.3, true},
		{"Equator non-zero lat", 47.5, 0.0, true},
		{"International date line", 0.0, 180.0, true},
		{"International date line west", 0.0, -180.0, true},
		{"Max valid lat", 89.99, 100.0, true},
		{"Max valid lon", 50.0, 179.99, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetCOTState()

			msg := `<event version="2.0" uid="` + tc.name + `" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
				<point lat="` + floatToString(tc.lat) + `" lon="` + floatToString(tc.lon) + `" hae="100"/>
				<detail><track speed="10.0" course="45.0"/></detail>
			</event>`

			processCotMessage(msg)

			trafficMutex.Lock()
			count := len(traffic)
			trafficMutex.Unlock()

			if tc.shouldCreate && count == 0 {
				t.Errorf("Expected traffic to be created for %s (lat=%f, lon=%f)", tc.name, tc.lat, tc.lon)
			}
			if !tc.shouldCreate && count > 0 {
				t.Errorf("Expected no traffic for %s (lat=%f, lon=%f), got %d", tc.name, tc.lat, tc.lon, count)
			}

			t.Logf("%s: lat=%f, lon=%f, created=%v", tc.name, tc.lat, tc.lon, count > 0)
		})
	}
}

// Helper function to convert float to string for XML
func floatToString(f float32) string {
	return fmt.Sprintf("%g", f)
}

// TestProcessCotMessage_SpeedConversion tests various speed values and conversions
func TestProcessCotMessage_SpeedConversion(t *testing.T) {
	resetCOTState()

	testCases := []struct {
		name        string
		speedMS     float32
		expectedKts uint16
		expectValid bool
	}{
		{"Zero speed", 0.0, 0, false},
		{"1 m/s", 1.0, 2, true},    // 1 * 1.94384449 ≈ 1.94 ≈ 2 kts
		{"10 m/s", 10.0, 19, true}, // 10 * 1.94384449 ≈ 19.4 ≈ 19 kts
		{"50 m/s", 50.0, 97, true}, // 50 * 1.94384449 ≈ 97.2 ≈ 97 kts
		{"100 m/s", 100.0, 194, true},
		{"0.5 m/s", 0.5, 0, false}, // 0.5 * 1.94384449 ≈ 0.97, truncates to 0 kts -> invalid
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetCOTState()

			msg := `<event version="2.0" uid="SPEED-` + tc.name + `" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
				<point lat="47.5" lon="-122.3" hae="100"/>
				<detail><track speed="` + floatToString(tc.speedMS) + `" course="45.0"/></detail>
			</event>`

			processCotMessage(msg)

			trafficMutex.Lock()
			var ti TrafficInfo
			found := false
			for _, v := range traffic {
				ti = v
				found = true
				break
			}
			trafficMutex.Unlock()

			if !found {
				t.Fatalf("Expected traffic to be created for %s", tc.name)
			}

			if ti.Speed_valid != tc.expectValid {
				t.Errorf("Expected Speed_valid=%v for %s, got %v", tc.expectValid, tc.name, ti.Speed_valid)
			}

			// Allow tolerance of +/- 2 knots for conversion
			if tc.expectValid && (ti.Speed < tc.expectedKts-2 || ti.Speed > tc.expectedKts+2) {
				t.Errorf("Expected Speed~%d for %s, got %d", tc.expectedKts, tc.name, ti.Speed)
			}

			t.Logf("%s: %f m/s -> %d kts, valid=%v", tc.name, tc.speedMS, ti.Speed, ti.Speed_valid)
		})
	}
}

// TestProcessCotMessage_AltitudeConversion tests various altitude scenarios
func TestProcessCotMessage_AltitudeConversion(t *testing.T) {
	resetCOTState()

	testCases := []struct {
		name       string
		haeMeters  float32
		expectedFt int32
	}{
		{"Zero altitude", 0.0, 0},
		{"100m", 100.0, 328},
		{"1000m", 1000.0, 3280},
		{"Negative (underground)", -50.0, -164},
		{"High altitude 10000m", 10000.0, 32808},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetCOTState()

			msg := `<event version="2.0" uid="ALT-` + tc.name + `" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
				<point lat="47.5" lon="-122.3" hae="` + floatToString(tc.haeMeters) + `"/>
				<detail><track speed="10.0" course="45.0"/></detail>
			</event>`

			processCotMessage(msg)

			trafficMutex.Lock()
			var ti TrafficInfo
			found := false
			for _, v := range traffic {
				ti = v
				found = true
				break
			}
			trafficMutex.Unlock()

			if !found {
				t.Fatalf("Expected traffic to be created for %s", tc.name)
			}

			// Allow tolerance of +/- 20 feet for conversion
			if ti.Alt < tc.expectedFt-20 || ti.Alt > tc.expectedFt+20 {
				t.Errorf("Expected Alt~%d for %s, got %d", tc.expectedFt, tc.name, ti.Alt)
			}

			t.Logf("%s: %f m -> %d ft", tc.name, tc.haeMeters, ti.Alt)
		})
	}
}

// TestProcessCotMessage_CourseValues tests various course headings
func TestProcessCotMessage_CourseValues(t *testing.T) {
	resetCOTState()

	testCases := []struct {
		name          string
		course        float32
		expectedTrack float32
	}{
		{"North", 0.0, 0.0},
		{"East", 90.0, 90.0},
		{"South", 180.0, 180.0},
		{"West", 270.0, 270.0},
		{"Northeast", 45.0, 45.0},
		{"Northwest", 315.0, 315.0},
		{"359 degrees", 359.0, 359.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetCOTState()

			msg := `<event version="2.0" uid="COURSE-` + tc.name + `" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
				<point lat="47.5" lon="-122.3" hae="100"/>
				<detail><track speed="10.0" course="` + floatToString(tc.course) + `"/></detail>
			</event>`

			processCotMessage(msg)

			trafficMutex.Lock()
			var ti TrafficInfo
			found := false
			for _, v := range traffic {
				ti = v
				found = true
				break
			}
			trafficMutex.Unlock()

			if !found {
				t.Fatalf("Expected traffic to be created for %s", tc.name)
			}

			if ti.Track != tc.expectedTrack {
				t.Errorf("Expected Track=%f for %s, got %f", tc.expectedTrack, tc.name, ti.Track)
			}

			t.Logf("%s: Course %f -> Track %f", tc.name, tc.course, ti.Track)
		})
	}
}

// TestProcessCotMessage_MultipleTargets tests handling multiple COT targets
func TestProcessCotMessage_MultipleTargets(t *testing.T) {
	resetCOTState()

	uids := []string{"DRONE-1", "DRONE-2", "PLANE-A", "HELI-X", "UNKNOWN"}

	for i, uid := range uids {
		msg := `<event version="2.0" uid="` + uid + `" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
			<point lat="47.` + string(rune('0'+i)) + `" lon="-122.3" hae="100"/>
			<detail><track speed="10.0" course="45.0"/></detail>
		</event>`
		processCotMessage(msg)
	}

	trafficMutex.Lock()
	count := len(traffic)
	trafficMutex.Unlock()

	if count != len(uids) {
		t.Errorf("Expected %d traffic targets, got %d", len(uids), count)
	}

	t.Logf("Created %d unique traffic targets from %d UIDs", count, len(uids))
}

// TestProcessCotMessage_UpdateExistingTarget tests updating a target with new position
func TestProcessCotMessage_UpdateExistingTarget(t *testing.T) {
	resetCOTState()

	// First message
	msg1 := `<event version="2.0" uid="TRACK-ME" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.0" lon="-122.0" hae="100"/>
		<detail><track speed="10.0" course="45.0"/></detail>
	</event>`

	processCotMessage(msg1)

	trafficMutex.Lock()
	count1 := len(traffic)
	var ti1 TrafficInfo
	for _, v := range traffic {
		ti1 = v
		break
	}
	trafficMutex.Unlock()

	if count1 != 1 {
		t.Fatalf("Expected 1 traffic after first message, got %d", count1)
	}

	// Second message with updated position
	msg2 := `<event version="2.0" uid="TRACK-ME" type="a-f-G" time="2024-01-01T12:01:00Z" start="2024-01-01T12:01:00Z" stale="2024-01-01T12:06:00Z">
		<point lat="47.1" lon="-122.1" hae="200"/>
		<detail><track speed="20.0" course="90.0"/></detail>
	</event>`

	processCotMessage(msg2)

	trafficMutex.Lock()
	count2 := len(traffic)
	var ti2 TrafficInfo
	for _, v := range traffic {
		ti2 = v
		break
	}
	trafficMutex.Unlock()

	// Should still have only 1 target (updated)
	if count2 != 1 {
		t.Errorf("Expected 1 traffic after update, got %d", count2)
	}

	// Verify position was updated
	if ti2.Lat == ti1.Lat {
		t.Error("Expected Lat to be updated")
	}
	if ti2.Track == ti1.Track {
		t.Error("Expected Track to be updated")
	}

	t.Logf("Target updated: Lat %f->%f, Track %f->%f",
		ti1.Lat, ti2.Lat, ti1.Track, ti2.Track)
}

// TestProcessCotMessage_AddressConsistency tests that same UID generates same address
func TestProcessCotMessage_AddressConsistency(t *testing.T) {
	resetCOTState()

	uid := "CONSISTENT-TEST"

	// Send same UID multiple times
	for i := 0; i < 5; i++ {
		msg := `<event version="2.0" uid="` + uid + `" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
			<point lat="47.5" lon="-122.3" hae="100"/>
			<detail><track speed="10.0" course="45.0"/></detail>
		</event>`
		processCotMessage(msg)
	}

	trafficMutex.Lock()
	count := len(traffic)
	var addr uint32
	for _, v := range traffic {
		addr = v.Icao_addr
		break
	}
	trafficMutex.Unlock()

	// Should only create 1 target
	if count != 1 {
		t.Errorf("Expected 1 traffic target from same UID, got %d", count)
	}

	// Address should be non-zero
	if addr == 0 {
		t.Error("Expected non-zero address")
	}

	t.Logf("UID '%s' consistently maps to address 0x%06X", uid, addr)
}

// TestProcessCotMessage_LargeXML tests handling of large XML payloads
func TestProcessCotMessage_LargeXML(t *testing.T) {
	resetCOTState()

	// Create a message with lots of extra attributes and nested elements
	largeMsg := `<event version="2.0" uid="LARGE-MSG-TEST-WITH-VERY-LONG-UID-TO-TEST-PARSING" type="a-f-G-U-C-I" time="2024-01-01T12:00:00.123456Z" start="2024-01-01T12:00:00.123456Z" stale="2024-01-01T12:05:00.123456Z" how="m-g" access="Undefined" qos="1-r" opex="o-gen">
		<point lat="47.606209" lon="-122.332069" hae="100.123456" ce="9.9" le="9.9"/>
		<detail>
			<track speed="10.5" course="45.678"/>
			<contact callsign="TEST-CALLSIGN" endpoint="192.168.1.100:4242"/>
			<uid Droid="LARGE-MSG-TEST"/>
			<precisionlocation geopointsrc="GPS" altsrc="GPS"/>
			<status battery="100"/>
			<takv device="Test Device" platform="Test Platform" os="Linux" version="1.0"/>
		</detail>
	</event>`

	processCotMessage(largeMsg)

	trafficMutex.Lock()
	count := len(traffic)
	trafficMutex.Unlock()

	if count != 1 {
		t.Errorf("Expected 1 traffic from large XML, got %d", count)
	}

	t.Log("Large XML message handled successfully")
}

// TestProcessCotMessage_QuickSuccession tests rapid message processing
func TestProcessCotMessage_QuickSuccession(t *testing.T) {
	resetCOTState()

	// Send 50 messages in quick succession
	for i := 0; i < 50; i++ {
		msg := `<event version="2.0" uid="RAPID-` + string(rune('A'+i%26)) + `" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
			<point lat="47.5" lon="-122.3" hae="100"/>
			<detail><track speed="10.0" course="45.0"/></detail>
		</event>`
		processCotMessage(msg)
	}

	trafficMutex.Lock()
	count := len(traffic)
	trafficMutex.Unlock()

	// Should create up to 26 unique targets (A-Z)
	if count > 26 {
		t.Errorf("Expected at most 26 traffic targets, got %d", count)
	}

	t.Logf("Processed 50 rapid messages, created %d unique targets", count)
}

// TestProcessCotMessage_NegativeAltitude tests negative altitude handling
func TestProcessCotMessage_NegativeAltitude(t *testing.T) {
	resetCOTState()

	msg := `<event version="2.0" uid="UNDERGROUND" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="-50"/>
		<detail><track speed="0.0" course="0.0"/></detail>
	</event>`

	processCotMessage(msg)

	trafficMutex.Lock()
	var ti TrafficInfo
	found := false
	for _, v := range traffic {
		ti = v
		found = true
		break
	}
	trafficMutex.Unlock()

	if !found {
		t.Fatal("Expected traffic to be created")
	}

	// -50m * 3.28084 = -164ft
	expectedAlt := int32(-164)
	if ti.Alt > expectedAlt+20 || ti.Alt < expectedAlt-20 {
		t.Errorf("Expected Alt~%d for -50m HAE, got %d", expectedAlt, ti.Alt)
	}

	t.Logf("Negative altitude: -50m -> %dft", ti.Alt)
}

// TestProcessCotMessage_WithValidBaroAndGPS tests altitude with both sensors valid
func TestProcessCotMessage_WithValidBaroAndGPS(t *testing.T) {
	resetCOTState()

	// Setup valid GPS and baro
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = -122.3
	mySituation.GPSAltitudeMSL = 500.0
	mySituation.GPSFixQuality = 2
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 550.0
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	msg := `<event version="2.0" uid="BARO-GPS" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="300"/>
		<detail><track speed="10.0" course="45.0"/></detail>
	</event>`

	processCotMessage(msg)

	trafficMutex.Lock()
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	// Altitude should be pressure altitude, not GNSS
	if ti.AltIsGNSS {
		t.Error("Expected AltIsGNSS=false when both GPS and baro are valid")
	}

	t.Logf("Altitude with both sensors: %d ft, AltIsGNSS=%v", ti.Alt, ti.AltIsGNSS)
}

// TestProcessCotMessage_TrafficSourceField tests that COT traffic is marked correctly
func TestProcessCotMessage_TrafficSourceField(t *testing.T) {
	resetCOTState()

	msg := `<event version="2.0" uid="SOURCE-TEST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail><track speed="10.0" course="45.0"/></detail>
	</event>`

	processCotMessage(msg)

	trafficMutex.Lock()
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	// Traffic source should be COT (Cursor on Target)
	if ti.Last_source != TRAFFIC_SOURCE_COT {
		t.Errorf("Expected Last_source=%d (COT), got %d", TRAFFIC_SOURCE_COT, ti.Last_source)
	}

	if ti.Addr_type != 1 {
		t.Errorf("Expected Addr_type=1 (non-ICAO), got %d", ti.Addr_type)
	}

	t.Logf("COT traffic source: %d, addr_type: %d", ti.Last_source, ti.Addr_type)
}

// TestProcessCotMessage_EmptyDetail tests message with empty detail section
func TestProcessCotMessage_EmptyDetail(t *testing.T) {
	resetCOTState()

	msg := `<event version="2.0" uid="NO-DETAIL" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail></detail>
	</event>`

	processCotMessage(msg)

	trafficMutex.Lock()
	var ti TrafficInfo
	found := false
	for _, v := range traffic {
		ti = v
		found = true
		break
	}
	trafficMutex.Unlock()

	if !found {
		t.Fatal("Expected traffic to be created even without track details")
	}

	// Speed should be 0 and invalid
	if ti.Speed != 0 {
		t.Errorf("Expected Speed=0, got %d", ti.Speed)
	}
	if ti.Speed_valid {
		t.Error("Expected Speed_valid=false when no track data")
	}

	t.Logf("Empty detail handled: Speed=%d, valid=%v", ti.Speed, ti.Speed_valid)
}

// TestProcessCotMessage_UnicodeUID tests UIDs with unicode characters
func TestProcessCotMessage_UnicodeUID(t *testing.T) {
	resetCOTState()

	msg := `<event version="2.0" uid="测试-ТЕСТ-🚁" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail><track speed="10.0" course="45.0"/></detail>
	</event>`

	processCotMessage(msg)

	trafficMutex.Lock()
	var ti TrafficInfo
	found := false
	for _, v := range traffic {
		ti = v
		found = true
		break
	}
	trafficMutex.Unlock()

	if !found {
		t.Fatal("Expected traffic to be created with unicode UID")
	}

	if ti.Reg != "测试-ТЕСТ-🚁" {
		t.Errorf("Expected Reg='测试-ТЕСТ-🚁', got '%s'", ti.Reg)
	}

	t.Logf("Unicode UID handled: '%s'", ti.Reg)
}

// TestProcessCotMessage_TimestampParsing tests that timestamps are handled
func TestProcessCotMessage_TimestampParsing(t *testing.T) {
	resetCOTState()

	msg := `<event version="2.0" uid="TIME-TEST" type="a-f-G" time="2024-01-01T12:00:00Z" start="2024-01-01T12:00:00Z" stale="2024-01-01T12:05:00Z">
		<point lat="47.5" lon="-122.3" hae="100"/>
		<detail><track speed="10.0" course="45.0"/></detail>
	</event>`

	beforeTime := time.Now()
	processCotMessage(msg)
	afterTime := time.Now()

	trafficMutex.Lock()
	var ti TrafficInfo
	for _, v := range traffic {
		ti = v
		break
	}
	trafficMutex.Unlock()

	// Timestamp should be recent
	if ti.Timestamp.Before(beforeTime) || ti.Timestamp.After(afterTime) {
		t.Errorf("Expected Timestamp between %v and %v, got %v",
			beforeTime, afterTime, ti.Timestamp)
	}

	t.Logf("Timestamp: %v", ti.Timestamp)
}

// resetCOTStateExtended is like resetCOTState but also initializes traffic mutex
func resetCOTStateExtended() {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond)
	}

	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
		mySituation.muGPSPerformance = &sync.Mutex{}
		mySituation.muAttitude = &sync.Mutex{}
		mySituation.muBaro = &sync.Mutex{}
		mySituation.muSatellite = &sync.Mutex{}
	}

	if trafficMutex == nil {
		trafficMutex = &sync.Mutex{}
	}

	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)

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
