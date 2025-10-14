// integration_adsb_test.go: Integration tests for 1090ES ADS-B protocol parsing
// Tests use trace file replay to verify protocol parser behavior without hardware

package main

import (
	"compress/gzip"
	"encoding/csv"
	"os"
	"testing"
	"time"
)

// resetTrafficState clears the global traffic map and related state for testing
func resetTrafficState() {
	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond) // Let the clock start
	}

	if trafficMutex == nil {
		initTraffic(true) // Initialize in replay mode (no background goroutines)
	}

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	// Clear all traffic
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)

	// Reset global statistics
	globalStatus.ES_messages_total = 0
	globalStatus.ES_traffic_targets_tracking = 0
}

// replayTraceFileDirect reads a trace file and directly injects messages without timing delays
// This is faster than Replay() for testing purposes
func replayTraceFileDirect(t *testing.T, filename string, context string) int {
	fh, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open trace file %s: %v", filename, err)
	}
	defer fh.Close()

	gzr, err := gzip.NewReader(fh)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	csvr := csv.NewReader(gzr)
	count := 0

	for {
		record, err := csvr.Read()
		if err != nil {
			break // End of file
		}

		if len(record) != 3 {
			continue
		}

		// Check if this is the context we want
		if record[1] != context {
			continue
		}

		// Directly inject the message without timing delays
		if context == CONTEXT_DUMP1090 {
			parseDump1090Message(record[2])
			count++
		}
	}

	return count
}

// TestADSBBasicTrafficParsing tests basic 1090ES traffic parsing using the basic_adsb trace file
func TestADSBBasicTrafficParsing(t *testing.T) {
	// Reset state before test
	resetTrafficState()

	// Replay the basic ADS-B trace file
	// This file contains 2 aircraft: UAL123 at 35000ft and N172SP at 5500ft
	msgCount := replayTraceFileDirect(t, "testdata/adsb/basic_adsb.trace.gz", CONTEXT_DUMP1090)
	t.Logf("Processed %d 1090ES messages from trace file", msgCount)

	// Verify traffic was parsed
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	// Check that we have the expected number of aircraft
	if len(traffic) != 2 {
		t.Errorf("Expected 2 aircraft in traffic map, got %d", len(traffic))
		// Log what we actually got
		for icao, ti := range traffic {
			t.Logf("Found aircraft: ICAO=%06X, Tail=%s, Alt=%d", icao, ti.Tail, ti.Alt)
		}
	}

	// Verify UAL123 (ICAO 0xA12345)
	if ti, ok := traffic[0xA12345]; ok {
		// Check tail number (may have whitespace)
		if ti.Tail != "UAL123" && ti.Tail != "UAL123  " {
			t.Errorf("Aircraft A12345: expected tail 'UAL123', got '%s'", ti.Tail)
		}

		// Check altitude (should be 35000ft)
		if ti.Alt != 35000 {
			t.Errorf("Aircraft A12345 (UAL123): expected altitude 35000, got %d", ti.Alt)
		}

		// Check ground speed (should be 450kts)
		if ti.Speed != 450 {
			t.Errorf("Aircraft A12345 (UAL123): expected speed 450, got %d", ti.Speed)
		}

		// Check track (should be 270 degrees)
		if ti.Track != 270 {
			t.Errorf("Aircraft A12345 (UAL123): expected track 270, got %f", ti.Track)
		}

		// Check that position is valid
		if !ti.Position_valid {
			t.Error("Aircraft A12345 (UAL123): position should be valid")
		}

		// Check that it's marked as 1090ES source
		if ti.Last_source != TRAFFIC_SOURCE_1090ES {
			t.Errorf("Aircraft A12345 (UAL123): expected source 1090ES (%d), got %d",
				TRAFFIC_SOURCE_1090ES, ti.Last_source)
		}

		t.Logf("UAL123: Verified ICAO=A12345, Alt=%d, Speed=%d, Track=%f, Lat=%f, Lon=%f",
			ti.Alt, ti.Speed, ti.Track, ti.Lat, ti.Lng)
	} else {
		t.Error("Aircraft A12345 (UAL123) not found in traffic map")
	}

	// Verify N172SP (ICAO 0xAC82EC)
	if ti, ok := traffic[0xAC82EC]; ok {
		// Check tail number
		if ti.Tail != "N172SP" && ti.Tail != "N172SP  " {
			t.Errorf("Aircraft AC82EC: expected tail 'N172SP', got '%s'", ti.Tail)
		}

		// Check altitude (should be around 5500ft, may vary slightly between messages)
		if ti.Alt < 5500 || ti.Alt > 5550 {
			t.Errorf("Aircraft AC82EC (N172SP): expected altitude ~5520, got %d", ti.Alt)
		}

		// Check ground speed (should be around 120kts)
		if ti.Speed < 120 || ti.Speed > 125 {
			t.Errorf("Aircraft AC82EC (N172SP): expected speed ~121, got %d", ti.Speed)
		}

		// Check track (should be around 90 degrees)
		if ti.Track < 90 || ti.Track > 95 {
			t.Errorf("Aircraft AC82EC (N172SP): expected track ~91, got %f", ti.Track)
		}

		// Check that position is valid
		if !ti.Position_valid {
			t.Error("Aircraft AC82EC (N172SP): position should be valid")
		}

		// Check that it's marked as 1090ES source
		if ti.Last_source != TRAFFIC_SOURCE_1090ES {
			t.Errorf("Aircraft AC82EC (N172SP): expected source 1090ES (%d), got %d",
				TRAFFIC_SOURCE_1090ES, ti.Last_source)
		}

		t.Logf("N172SP: Verified ICAO=AC82EC, Alt=%d, Speed=%d, Track=%f, Lat=%f, Lon=%f",
			ti.Alt, ti.Speed, ti.Track, ti.Lat, ti.Lng)
	} else {
		t.Error("Aircraft AC82EC (N172SP) not found in traffic map")
	}

	t.Logf("Successfully parsed and verified %d aircraft from basic_adsb.trace.gz", len(traffic))
}

// TestADSBSignalLevel tests that signal levels are properly recorded
func TestADSBSignalLevel(t *testing.T) {
	resetTrafficState()
	replayTraceFileDirect(t, "testdata/adsb/basic_adsb.trace.gz", CONTEXT_DUMP1090)

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	// Check that aircraft have valid signal levels
	for icao, ti := range traffic {
		if ti.SignalLevel == 0 {
			t.Errorf("Aircraft %06X has zero signal level (should be set)", icao)
		}

		// Signal level should be negative dB value (RSSI)
		if ti.SignalLevel > 0 {
			t.Errorf("Aircraft %06X has positive signal level %f (should be negative dBm)",
				icao, ti.SignalLevel)
		}

		t.Logf("Aircraft %06X: Signal level = %.1f dBm", icao, ti.SignalLevel)
	}
}

// TestADSBReceivedMessageCount tests that message counts are incremented
func TestADSBReceivedMessageCount(t *testing.T) {
	resetTrafficState()
	replayTraceFileDirect(t, "testdata/adsb/basic_adsb.trace.gz", CONTEXT_DUMP1090)

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	// Each aircraft should have received multiple messages (3 in the trace file)
	for icao, ti := range traffic {
		if ti.ReceivedMsgs == 0 {
			t.Errorf("Aircraft %06X has zero received messages", icao)
		}

		if ti.ReceivedMsgs < 3 {
			t.Errorf("Aircraft %06X: expected at least 3 messages, got %d",
				icao, ti.ReceivedMsgs)
		}

		t.Logf("Aircraft %06X: Received %d messages", icao, ti.ReceivedMsgs)
	}
}

// TestADSBTargetType tests that target types are correctly identified
func TestADSBTargetType(t *testing.T) {
	resetTrafficState()
	replayTraceFileDirect(t, "testdata/adsb/basic_adsb.trace.gz", CONTEXT_DUMP1090)

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	// Both aircraft in the trace should be ADS-B targets (DF=17)
	for icao, ti := range traffic {
		if ti.TargetType != TARGET_TYPE_ADSB {
			t.Errorf("Aircraft %06X: expected target type ADS-B (%d), got %d",
				icao, TARGET_TYPE_ADSB, ti.TargetType)
		}

		// Address type should be 0 for DF=17
		if ti.Addr_type != 0 {
			t.Errorf("Aircraft %06X: expected address type 0 for ADS-B, got %d",
				icao, ti.Addr_type)
		}

		t.Logf("Aircraft %06X: Target type = %d, Address type = %d",
			icao, ti.TargetType, ti.Addr_type)
	}
}

// TestADSBTimestamps tests that timestamps are properly set
func TestADSBTimestamps(t *testing.T) {
	resetTrafficState()

	replayTraceFileDirect(t, "testdata/adsb/basic_adsb.trace.gz", CONTEXT_DUMP1090)

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	// Check that timestamps are reasonable
	// Note: parseDump1090Message sets Timestamp from newTi.Timestamp which may be zero
	// The important timestamps are Last_seen and Last_alt which use stratuxClock
	for icao, ti := range traffic {
		// Last_seen should be set (stratuxClock time)
		if ti.Last_seen.IsZero() {
			t.Errorf("Aircraft %06X: Last_seen is zero", icao)
		}

		// Last_alt should be set (stratuxClock time)
		if ti.Last_alt.IsZero() {
			t.Errorf("Aircraft %06X: Last_alt is zero", icao)
		}

		// Last_speed should be set for valid speed
		if ti.Speed_valid && ti.Last_speed.IsZero() {
			t.Errorf("Aircraft %06X: Last_speed is zero despite Speed_valid=true", icao)
		}

		t.Logf("Aircraft %06X: Last_seen = %v, Last_alt = %v, Last_speed = %v",
			icao, ti.Last_seen, ti.Last_alt, ti.Last_speed)
	}
}

// TestADSBPositionValidity tests position validation logic
func TestADSBPositionValidity(t *testing.T) {
	resetTrafficState()
	replayTraceFileDirect(t, "testdata/adsb/basic_adsb.trace.gz", CONTEXT_DUMP1090)

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	// All aircraft in the basic trace have valid positions
	for icao, ti := range traffic {
		if !ti.Position_valid {
			t.Errorf("Aircraft %06X: position should be valid", icao)
		}

		// Latitude should be reasonable (Seattle area ~47N)
		if ti.Lat < 40 || ti.Lat > 50 {
			t.Errorf("Aircraft %06X: latitude %f is outside expected range (40-50)",
				icao, ti.Lat)
		}

		// Longitude should be reasonable (Seattle area ~122W = -122)
		if ti.Lng > -120 || ti.Lng < -130 {
			t.Errorf("Aircraft %06X: longitude %f is outside expected range (-130 to -120)",
				icao, ti.Lng)
		}

		// ExtrapolatedPosition should be false for fresh data
		if ti.ExtrapolatedPosition {
			t.Errorf("Aircraft %06X: position should not be extrapolated", icao)
		}

		t.Logf("Aircraft %06X: Position = (%f, %f), Valid = %v",
			icao, ti.Lat, ti.Lng, ti.Position_valid)
	}
}

// TestADSBNavigationIntegrity tests NIC and NACp values
func TestADSBNavigationIntegrity(t *testing.T) {
	resetTrafficState()
	replayTraceFileDirect(t, "testdata/adsb/basic_adsb.trace.gz", CONTEXT_DUMP1090)

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	// Check that NIC and NACp are set
	for icao, ti := range traffic {
		// NIC should be set (typically 7-11 for good ADS-B)
		if ti.NIC == 0 {
			t.Logf("Aircraft %06X: NIC is 0 (may be valid but unusual)", icao)
		}

		if ti.NIC < 0 || ti.NIC > 11 {
			t.Errorf("Aircraft %06X: NIC %d is outside valid range (0-11)",
				icao, ti.NIC)
		}

		// NACp should be set
		if ti.NACp < 0 || ti.NACp > 11 {
			t.Errorf("Aircraft %06X: NACp %d is outside valid range (0-11)",
				icao, ti.NACp)
		}

		t.Logf("Aircraft %06X: NIC = %d, NACp = %d", icao, ti.NIC, ti.NACp)
	}
}

// TestADSBEmitterCategory tests emitter category parsing
func TestADSBEmitterCategory(t *testing.T) {
	resetTrafficState()
	replayTraceFileDirect(t, "testdata/adsb/basic_adsb.trace.gz", CONTEXT_DUMP1090)

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	// Check emitter categories
	for icao, ti := range traffic {
		// Emitter category 0 means "no information"
		// Valid range is 0-19 per GDL90 spec
		if ti.Emitter_category > 19 {
			t.Errorf("Aircraft %06X: emitter category %d is outside valid range (0-19)",
				icao, ti.Emitter_category)
		}

		t.Logf("Aircraft %06X: Emitter category = %d", icao, ti.Emitter_category)
	}
}
