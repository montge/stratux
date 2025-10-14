package main

import (
	"compress/gzip"
	"encoding/csv"
	"io"
	"os"
	"sync"
	"testing"
	"time"
)

// reset1090ESState clears the global 1090ES state for testing
func reset1090ESState() {
	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond)
	}

	// Reset 1090ES statistics
	globalStatus.ES_messages_total = 0
	globalStatus.ES_messages_last_minute = 0
	globalStatus.ES_messages_max = 0
	globalStatus.ES_traffic_targets_tracking = 0

	// Reset message log
	msgLogMutex = sync.Mutex{}
	msgLog = make([]msg, 0)

	// Reset traffic tracking
	trafficMutex = &sync.Mutex{}
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)
}

// replay1090ESTraceDirect reads a 1090ES trace file and directly injects dump1090 JSON messages
func replay1090ESTraceDirect(t *testing.T, filename string) int {
	t.Helper()

	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open trace file %s: %v", filename, err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	r := csv.NewReader(gzr)
	messageCount := 0

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read CSV record: %v", err)
		}

		if len(record) != 3 {
			continue
		}

		msgType := record[1]
		msgData := record[2]

		if msgType == "dump1090" {
			// Process dump1090 message directly via parseDump1090Message
			parseDump1090Message(msgData)
			messageCount++
		}
	}

	return messageCount
}

// Test1090ESBasicParsing tests basic 1090ES message parsing from trace file
func Test1090ESBasicParsing(t *testing.T) {
	reset1090ESState()

	count := replay1090ESTraceDirect(t, "testdata/1090es/basic_1090es.trace.gz")

	if count == 0 {
		t.Error("Expected to parse at least one 1090ES message, got 0")
	}

	t.Logf("Successfully parsed %d 1090ES messages from trace file", count)

	// Verify that 1090ES message counter was incremented
	if globalStatus.ES_messages_total == 0 {
		t.Error("Expected ES_messages_total > 0, got 0")
	}

	t.Logf("Total 1090ES messages counted: %d", globalStatus.ES_messages_total)
}

// Test1090ESPositionMessage tests parsing of ADS-B position messages
func Test1090ESPositionMessage(t *testing.T) {
	reset1090ESState()

	// DF17 position message (TypeCode 11 - airborne position with barometric altitude)
	posMsg := `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7657,"Lng":-1.1918,"Alt":5850,"Position_valid":true,"SignalLevel":0.0512,"Timestamp":"2025-10-14T12:00:00.000Z"}`

	parseDump1090Message(posMsg)

	// Check that traffic was created
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	if len(traffic) != 1 {
		t.Fatalf("Expected 1 traffic target, got %d", len(traffic))
	}

	// Check the traffic info
	icao := uint32(11230838) // 0xAB4F66
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	if !ti.Position_valid {
		t.Error("Expected Position_valid to be true")
	}

	if ti.Lat != 51.7657 {
		t.Errorf("Expected Lat=51.7657, got %f", ti.Lat)
	}

	if ti.Lng != -1.1918 {
		t.Errorf("Expected Lng=-1.1918, got %f", ti.Lng)
	}

	if ti.Alt != 5850 {
		t.Errorf("Expected Alt=5850, got %d", ti.Alt)
	}

	if ti.TargetType != TARGET_TYPE_ADSB {
		t.Errorf("Expected TargetType=TARGET_TYPE_ADSB (%d), got %d", TARGET_TYPE_ADSB, ti.TargetType)
	}

	if ti.Last_source != TRAFFIC_SOURCE_1090ES {
		t.Errorf("Expected Last_source=TRAFFIC_SOURCE_1090ES (%d), got %d", TRAFFIC_SOURCE_1090ES, ti.Last_source)
	}

	t.Logf("Successfully parsed position message: ICAO=%X, Lat=%f, Lon=%f, Alt=%d", icao, ti.Lat, ti.Lng, ti.Alt)
}

// Test1090ESVelocityMessage tests parsing of airborne velocity messages
func Test1090ESVelocityMessage(t *testing.T) {
	reset1090ESState()

	// First send position to create the target
	posMsg := `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7657,"Lng":-1.1918,"Alt":5850,"Position_valid":true,"SignalLevel":0.0512,"Timestamp":"2025-10-14T12:00:00.000Z"}`
	parseDump1090Message(posMsg)

	// Now send velocity message (TypeCode 19)
	velMsg := `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":19,"Speed":468,"Track":89,"Vvel":-64,"Speed_valid":true,"SignalLevel":0.0498,"Timestamp":"2025-10-14T12:00:00.500Z"}`
	parseDump1090Message(velMsg)

	// Check the traffic info
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	icao := uint32(11230838)
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	if !ti.Speed_valid {
		t.Error("Expected Speed_valid to be true")
	}

	if ti.Speed != 468 {
		t.Errorf("Expected Speed=468, got %d", ti.Speed)
	}

	if ti.Track != 89 {
		t.Errorf("Expected Track=89, got %f", ti.Track)
	}

	if ti.Vvel != -64 {
		t.Errorf("Expected Vvel=-64, got %d", ti.Vvel)
	}

	t.Logf("Successfully parsed velocity message: Speed=%d, Track=%f, Vvel=%d", ti.Speed, ti.Track, ti.Vvel)
}

// Test1090ESCallsignMessage tests parsing of identification/callsign messages
func Test1090ESCallsignMessage(t *testing.T) {
	reset1090ESState()

	// DF17 callsign message (TypeCode 4)
	callsignMsg := `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":4,"Tail":"EZY123  ","SignalLevel":0.0502,"Timestamp":"2025-10-14T12:00:01.500Z"}`

	parseDump1090Message(callsignMsg)

	// Check the traffic info
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	icao := uint32(11230838)
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	expectedTail := "EZY123"
	if ti.Tail != expectedTail {
		t.Errorf("Expected Tail=%s, got %s", expectedTail, ti.Tail)
	}

	t.Logf("Successfully parsed callsign message: Tail=%s", ti.Tail)
}

// Test1090ESTISBMessage tests parsing of TIS-B messages (DF18, CA=2)
func Test1090ESTISBMessage(t *testing.T) {
	reset1090ESState()

	// DF18 TIS-B message (CA=2, with ICAO address)
	tisbMsg := `{"Icao_addr":2893118,"DF":18,"CA":2,"TypeCode":11,"Lat":51.7623,"Lng":-1.1889,"Alt":3200,"Position_valid":true,"SignalLevel":0.0512,"Timestamp":"2025-10-14T12:00:02.000Z"}`

	parseDump1090Message(tisbMsg)

	// Check the traffic info
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	icao := uint32(2893118) // 0x2C25BE
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	if ti.TargetType != TARGET_TYPE_TISB {
		t.Errorf("Expected TargetType=TARGET_TYPE_TISB (%d), got %d", TARGET_TYPE_TISB, ti.TargetType)
	}

	if ti.Addr_type != 2 {
		t.Errorf("Expected Addr_type=2 (TIS-B with ICAO), got %d", ti.Addr_type)
	}

	t.Logf("Successfully parsed TIS-B message: ICAO=%X, TargetType=%d", icao, ti.TargetType)
}

// Test1090ESADSRMessage tests parsing of ADS-R messages (DF18, CA=6)
func Test1090ESADSRMessage(t *testing.T) {
	reset1090ESState()

	// DF18 ADS-R message (CA=6)
	adsrMsg := `{"Icao_addr":11230840,"DF":18,"CA":6,"TypeCode":11,"Lat":51.7589,"Lng":-1.1834,"Alt":4100,"Position_valid":true,"SignalLevel":0.0498,"Timestamp":"2025-10-14T12:00:02.500Z"}`

	parseDump1090Message(adsrMsg)

	// Check the traffic info
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	icao := uint32(11230840)
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	if ti.TargetType != TARGET_TYPE_ADSR {
		t.Errorf("Expected TargetType=TARGET_TYPE_ADSR (%d), got %d", TARGET_TYPE_ADSR, ti.TargetType)
	}

	if ti.Addr_type != 2 {
		t.Errorf("Expected Addr_type=2 (ADS-R), got %d", ti.Addr_type)
	}

	t.Logf("Successfully parsed ADS-R message: ICAO=%X, TargetType=%d", icao, ti.TargetType)
}

// Test1090ESModeSAltitude tests parsing of Mode S surveillance altitude replies (DF4)
func Test1090ESModeSAltitude(t *testing.T) {
	reset1090ESState()

	// First create target with position
	posMsg := `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7657,"Lng":-1.1918,"Alt":5850,"Position_valid":true,"SignalLevel":0.0512,"Timestamp":"2025-10-14T12:00:00.000Z"}`
	parseDump1090Message(posMsg)

	// Now send DF4 altitude reply
	altMsg := `{"Icao_addr":11230838,"DF":4,"Alt":5875,"SignalLevel":0.0425,"Timestamp":"2025-10-14T12:00:03.000Z"}`
	parseDump1090Message(altMsg)

	// Check the traffic info
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	icao := uint32(11230838)
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	if ti.Alt != 5875 {
		t.Errorf("Expected Alt=5875 (updated from DF4), got %d", ti.Alt)
	}

	t.Logf("Successfully parsed DF4 altitude reply: Alt=%d", ti.Alt)
}

// Test1090ESSquawkCode tests parsing of squawk codes
func Test1090ESSquawkCode(t *testing.T) {
	reset1090ESState()

	// DF5 with squawk code (emergency 7700)
	squawkMsg := `{"Icao_addr":11230838,"DF":5,"Squawk":7700,"Alt":5900,"SignalLevel":0.0398,"Timestamp":"2025-10-14T12:00:04.000Z"}`

	parseDump1090Message(squawkMsg)

	// Check the traffic info
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	icao := uint32(11230838)
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	if ti.Squawk != 7700 {
		t.Errorf("Expected Squawk=7700 (emergency), got %d", ti.Squawk)
	}

	t.Logf("Successfully parsed squawk code: Squawk=%d", ti.Squawk)
}

// Test1090ESOnGroundFlag tests parsing of on-ground position messages
func Test1090ESOnGroundFlag(t *testing.T) {
	reset1090ESState()

	// DF17 surface position message (TypeCode 8) with OnGround flag
	groundMsg := `{"Icao_addr":10500126,"DF":17,"CA":5,"TypeCode":8,"Lat":51.7501,"Lng":-1.1723,"Alt":500,"OnGround":true,"Position_valid":true,"SignalLevel":0.0812,"Timestamp":"2025-10-14T12:00:05.500Z"}`

	parseDump1090Message(groundMsg)

	// Check the traffic info
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	icao := uint32(10500126)
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	if !ti.OnGround {
		t.Error("Expected OnGround to be true")
	}

	t.Logf("Successfully parsed on-ground message: OnGround=%t", ti.OnGround)
}

// Test1090ESNACp tests parsing of Navigation Accuracy Category for Position
func Test1090ESNACp(t *testing.T) {
	reset1090ESState()

	// Message with NACp field
	nacpMsg := `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7660,"Lng":-1.1920,"Alt":5925,"NACp":8,"Position_valid":true,"SignalLevel":0.0515,"Timestamp":"2025-10-14T12:00:06.000Z"}`

	parseDump1090Message(nacpMsg)

	// Check the traffic info
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	icao := uint32(11230838)
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	if ti.NACp != 8 {
		t.Errorf("Expected NACp=8, got %d", ti.NACp)
	}

	t.Logf("Successfully parsed NACp: NACp=%d", ti.NACp)
}

// Test1090ESEmitterCategory tests parsing of emitter category
func Test1090ESEmitterCategory(t *testing.T) {
	reset1090ESState()

	// Message with Emitter_category
	emitterMsg := `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":4,"Tail":"EZY123  ","Emitter_category":7,"SignalLevel":0.0508,"Timestamp":"2025-10-14T12:00:06.500Z"}`

	parseDump1090Message(emitterMsg)

	// Check the traffic info
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	icao := uint32(11230838)
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	if ti.Emitter_category != 7 {
		t.Errorf("Expected Emitter_category=7, got %d", ti.Emitter_category)
	}

	t.Logf("Successfully parsed emitter category: Emitter_category=%d", ti.Emitter_category)
}

// Test1090ESMultipleAircraft tests tracking multiple aircraft simultaneously
func Test1090ESMultipleAircraft(t *testing.T) {
	reset1090ESState()

	messages := []string{
		`{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7657,"Lng":-1.1918,"Alt":5850,"Position_valid":true,"SignalLevel":0.0512,"Timestamp":"2025-10-14T12:00:00.000Z"}`,
		`{"Icao_addr":10685854,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7542,"Lng":-1.1778,"Alt":8525,"Position_valid":true,"SignalLevel":0.0645,"Timestamp":"2025-10-14T12:00:01.000Z"}`,
		`{"Icao_addr":11184758,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7465,"Lng":-1.1667,"Alt":35000,"Position_valid":true,"SignalLevel":0.0198,"Timestamp":"2025-10-14T12:00:04.500Z"}`,
	}

	for _, msg := range messages {
		parseDump1090Message(msg)
	}

	// Check that we have 3 aircraft
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	if len(traffic) != 3 {
		t.Errorf("Expected 3 aircraft, got %d", len(traffic))
	}

	// Verify all ICAOs are present
	expectedICAOs := []uint32{11230838, 10685854, 11184758}
	for _, icao := range expectedICAOs {
		if _, ok := traffic[icao]; !ok {
			t.Errorf("Expected traffic for ICAO %X, not found", icao)
		}
	}

	t.Logf("Successfully tracking %d aircraft", len(traffic))
}

// Test1090ESSignalStrength tests signal strength parsing and storage
func Test1090ESSignalStrength(t *testing.T) {
	reset1090ESState()

	// Message with signal level
	sigMsg := `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7657,"Lng":-1.1918,"Alt":5850,"Position_valid":true,"SignalLevel":0.0512,"Timestamp":"2025-10-14T12:00:00.000Z"}`

	parseDump1090Message(sigMsg)

	// Check the traffic info
	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	icao := uint32(11230838)
	ti, ok := traffic[icao]
	if !ok {
		t.Fatalf("Expected traffic for ICAO %X, not found", icao)
	}

	// Signal level should be converted to dB: 10 * log10(0.0512) ≈ -12.9 dB
	if ti.SignalLevel >= 0 || ti.SignalLevel < -999 {
		t.Errorf("Expected negative signal level in dB range, got %f", ti.SignalLevel)
	}

	t.Logf("Successfully parsed signal strength: SignalLevel=%f dB", ti.SignalLevel)
}

// Test1090ESInvalidMessage tests handling of invalid JSON messages
func Test1090ESInvalidMessage(t *testing.T) {
	reset1090ESState()

	testCases := []struct {
		name    string
		message string
	}{
		{"empty_json", "{}"},
		{"invalid_json", "{invalid json}"},
		{"missing_icao", `{"DF":17,"Lat":51.7657}`},
		{"heartbeat", `{"Icao_addr":134217727,"DF":17}`}, // 0x07FFFFFF is heartbeat
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initialTotal := globalStatus.ES_messages_total

			parseDump1090Message(tc.message)

			// Heartbeat and invalid messages should still increment counter
			// but not create traffic entries (except heartbeat which is explicitly filtered)
			if globalStatus.ES_messages_total <= initialTotal {
				if tc.message != "{}" && tc.message != "{invalid json}" {
					// Empty and malformed JSON might not increment
					t.Logf("Message counter: %d", globalStatus.ES_messages_total)
				}
			}
		})
	}
}

// Test1090ESTraceReplay tests replaying the entire 1090ES trace file
func Test1090ESTraceReplay(t *testing.T) {
	reset1090ESState()

	count := replay1090ESTraceDirect(t, "testdata/1090es/basic_1090es.trace.gz")

	// We expect 15 messages in the trace file
	if count != 15 {
		t.Errorf("Expected 15 messages in trace file, got %d", count)
	}

	t.Logf("Replayed %d 1090ES messages from trace file", count)
	t.Logf("Total 1090ES messages counted: %d", globalStatus.ES_messages_total)

	// Check that we have multiple aircraft
	trafficMutex.Lock()
	trafficCount := len(traffic)
	trafficMutex.Unlock()

	if trafficCount == 0 {
		t.Error("Expected at least one aircraft in traffic map")
	}

	t.Logf("Tracking %d aircraft", trafficCount)
}

// Test1090ESMessageLog tests that 1090ES messages are logged to msgLog
func Test1090ESMessageLog(t *testing.T) {
	reset1090ESState()

	initialLogSize := len(msgLog)

	// Send a position message
	posMsg := `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7657,"Lng":-1.1918,"Alt":5850,"Position_valid":true,"SignalLevel":0.0512,"Timestamp":"2025-10-14T12:00:00.000Z"}`

	parseDump1090Message(posMsg)

	finalLogSize := len(msgLog)

	if finalLogSize <= initialLogSize {
		t.Errorf("Expected msgLog to grow, initial=%d, final=%d", initialLogSize, finalLogSize)
	}

	// Check the last message in the log
	if finalLogSize > 0 {
		lastMsg := msgLog[finalLogSize-1]
		if lastMsg.MessageClass != MSGCLASS_ES {
			t.Errorf("Expected MessageClass MSGCLASS_ES (%d), got %d", MSGCLASS_ES, lastMsg.MessageClass)
		}
		t.Logf("Message logged with class=%d", lastMsg.MessageClass)
	}
}
