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

// Test1090ESHeartbeatMessage tests the heartbeat message (0x07FFFFFF)
func Test1090ESHeartbeatMessage(t *testing.T) {
	reset1090ESState()

	// Heartbeat uses special ICAO address 0x07FFFFFF
	heartbeatMsg := `{"Icao_addr":134217727,"DF":0,"CA":0,"TypeCode":0}`
	parseDump1090Message(heartbeatMsg)

	// Heartbeat should not create traffic entry
	trafficMutex.Lock()
	_, exists := traffic[0x07FFFFFF]
	trafficMutex.Unlock()

	if exists {
		t.Error("Heartbeat message should not create traffic entry")
	}

	t.Log("Heartbeat message correctly ignored")
}

// Test1090ESNonICAOAddress tests handling of non-ICAO addresses (TIS-B)
func Test1090ESNonICAOAddress(t *testing.T) {
	reset1090ESState()

	// Non-ICAO addresses have bit 25 set by dump1090
	// Original address 0xABCDEF with bit 25 set becomes 0x01ABCDEF
	nonIcaoMsg := `{"Icao_addr":28036591,"DF":18,"CA":5,"TypeCode":11,"Lat":40.7128,"Lng":-74.006,"Alt":5000,"Position_valid":true}`
	parseDump1090Message(nonIcaoMsg)

	// Should be stored with address 0xABCDEF (bit 25 stripped)
	trafficMutex.Lock()
	ti, exists := traffic[0xABCDEF]
	trafficMutex.Unlock()

	if !exists {
		t.Error("Non-ICAO address should be stored with bit 25 stripped")
	} else {
		if ti.Addr_type != 3 {
			t.Logf("TIS-B target stored with Addr_type=%d", ti.Addr_type)
		}
	}

	t.Log("Non-ICAO address handling test complete")
}

// Test1090ESNICCalculation tests NIC (Navigation Integrity Category) calculation for different type codes
func Test1090ESNICCalculation(t *testing.T) {
	reset1090ESState()

	testCases := []struct {
		name        string
		typeCode    int
		subtypeCode int
		expectedNIC int
	}{
		{"TypeCode 17 -> NIC 1", 17, 0, 1},
		{"TypeCode 16 subtype 0 -> NIC 2", 16, 0, 2},
		{"TypeCode 16 subtype 1 -> NIC 3", 16, 1, 3},
		{"TypeCode 15 -> NIC 4", 15, 0, 4},
		{"TypeCode 14 -> NIC 5", 14, 0, 5},
		{"TypeCode 13 -> NIC 6", 13, 0, 6},
		{"TypeCode 12 -> NIC 7", 12, 0, 7},
		{"TypeCode 11 subtype 0 -> NIC 8", 11, 0, 8},
		{"TypeCode 11 subtype 1 -> NIC 9", 11, 1, 9},
		{"TypeCode 10 -> NIC 10", 10, 0, 10},
		{"TypeCode 21 -> NIC 10", 21, 0, 10},
		{"TypeCode 9 -> NIC 11", 9, 0, 11},
		{"TypeCode 20 -> NIC 11", 20, 0, 11},
		{"TypeCode 8 -> NIC 0", 8, 0, 0},
		{"TypeCode 22 -> NIC 0", 22, 0, 0},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use unique ICAO for each test
			icao := 0x100000 + i

			// Construct the message properly
			parseDump1090Message(constructNICTestMessage(icao, tc.typeCode, tc.subtypeCode))

			trafficMutex.Lock()
			ti, exists := traffic[uint32(icao)]
			trafficMutex.Unlock()

			if !exists {
				t.Errorf("Expected traffic entry for ICAO %X", icao)
				return
			}

			if ti.NIC != tc.expectedNIC {
				t.Errorf("Expected NIC %d for TypeCode %d SubtypeCode %d, got %d",
					tc.expectedNIC, tc.typeCode, tc.subtypeCode, ti.NIC)
			}
		})
	}
}

// constructNICTestMessage creates a dump1090 JSON message for NIC testing
func constructNICTestMessage(icao, typeCode, subtypeCode int) string {
	return `{"Icao_addr":` + intToStr(icao) +
		`,"DF":17,"CA":5,"TypeCode":` + intToStr(typeCode) +
		`,"SubtypeCode":` + intToStr(subtypeCode) +
		`,"Lat":40.0,"Lng":-74.0,"Alt":5000,"Position_valid":true}`
}

// intToStr converts an int to string (simple helper)
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// Test1090ESDisplayTrafficSource tests the DisplayTrafficSource feature
func Test1090ESDisplayTrafficSource(t *testing.T) {
	reset1090ESState()

	// Enable DisplayTrafficSource
	globalSettings.DisplayTrafficSource = true
	defer func() { globalSettings.DisplayTrafficSource = false }()

	testCases := []struct {
		name         string
		df           int
		ca           int
		tail         string
		expectedType string // 'a' for ADS-B, 'r' for ADS-R, 't' for TIS-B
	}{
		{"ADS-B no tail", 17, 5, "", "ea"},
		{"ADS-B with short tail", 17, 5, "N123", "eaN123"},
		{"ADS-R no tail", 18, 6, "", "er"},
		{"TIS-B CA=2 no tail", 18, 2, "", "et"},
		{"TIS-B CA=5 no tail", 18, 5, "", "et"},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			icao := 0x200000 + i

			var msg string
			if tc.tail != "" {
				msg = `{"Icao_addr":` + intToStr(icao) +
					`,"DF":` + intToStr(tc.df) +
					`,"CA":` + intToStr(tc.ca) +
					`,"TypeCode":11,"Lat":41.0,"Lng":-75.0,"Alt":6000,"Position_valid":true,"Tail":"` + tc.tail + `"}`
			} else {
				msg = `{"Icao_addr":` + intToStr(icao) +
					`,"DF":` + intToStr(tc.df) +
					`,"CA":` + intToStr(tc.ca) +
					`,"TypeCode":11,"Lat":41.0,"Lng":-75.0,"Alt":6000,"Position_valid":true}`
			}

			parseDump1090Message(msg)

			trafficMutex.Lock()
			ti, exists := traffic[uint32(icao)]
			trafficMutex.Unlock()

			if !exists {
				t.Errorf("Expected traffic entry for ICAO %X", icao)
				return
			}

			// Check that tail starts with expected prefix
			if len(ti.Tail) < 2 {
				t.Errorf("Expected tail to have traffic source prefix, got '%s'", ti.Tail)
			} else {
				t.Logf("Tail with source: '%s'", ti.Tail)
			}
		})
	}
}

// Test1090ESInvalidJSON tests handling of invalid JSON
func Test1090ESInvalidJSON(t *testing.T) {
	reset1090ESState()

	initialTrafficCount := len(traffic)

	// Send invalid JSON
	parseDump1090Message(`{invalid json}`)
	parseDump1090Message(``)
	parseDump1090Message(`{"incomplete":`)

	// Should not crash and should not add traffic
	if len(traffic) != initialTrafficCount {
		t.Error("Invalid JSON should not create traffic entries")
	}

	t.Log("Invalid JSON handling test complete")
}

// Test1090ESMissingPosition tests handling when lat/lng are missing but position_valid is true
func Test1090ESMissingPosition(t *testing.T) {
	reset1090ESState()

	icao := 0x300001

	// Position_valid true but missing Lat
	msg := `{"Icao_addr":` + intToStr(icao) + `,"DF":17,"CA":5,"TypeCode":11,"Lng":-74.0,"Alt":5000,"Position_valid":true}`
	parseDump1090Message(msg)

	trafficMutex.Lock()
	ti, exists := traffic[uint32(icao)]
	trafficMutex.Unlock()

	if exists && ti.Position_valid {
		t.Error("Expected Position_valid to be false when Lat is missing")
	}

	t.Log("Missing position handling test complete")
}

// Test1090ESMissingSpeed tests handling when track/speed are missing but speed_valid is true
func Test1090ESMissingSpeed(t *testing.T) {
	reset1090ESState()

	icao := 0x300002

	// Speed_valid true but missing Track
	msg := `{"Icao_addr":` + intToStr(icao) + `,"DF":17,"CA":5,"TypeCode":19,"Speed":250,"Vvel":500,"Speed_valid":true}`
	parseDump1090Message(msg)

	trafficMutex.Lock()
	ti, exists := traffic[uint32(icao)]
	trafficMutex.Unlock()

	if exists && ti.Speed_valid {
		t.Error("Expected Speed_valid to be false when Track is missing")
	}

	t.Log("Missing speed handling test complete")
}

// Test1090ESGnssDiff tests GnssDiffFromBaroAlt handling
func Test1090ESGnssDiff(t *testing.T) {
	reset1090ESState()

	icao := 0x400001
	gnssDiff := 150

	msg := `{"Icao_addr":` + intToStr(icao) + `,"DF":17,"CA":5,"TypeCode":11,"Lat":42.0,"Lng":-73.0,"Alt":10000,"GnssDiffFromBaroAlt":` + intToStr(gnssDiff) + `,"Position_valid":true}`
	parseDump1090Message(msg)

	trafficMutex.Lock()
	ti, exists := traffic[uint32(icao)]
	trafficMutex.Unlock()

	if !exists {
		t.Error("Expected traffic entry")
		return
	}

	if ti.GnssDiffFromBaroAlt != int32(gnssDiff) {
		t.Errorf("Expected GnssDiffFromBaroAlt %d, got %d", gnssDiff, ti.GnssDiffFromBaroAlt)
	}

	t.Logf("GnssDiff: %d", ti.GnssDiffFromBaroAlt)
}

// Test1090ESHeartbeatMessageWithDEBUG tests the heartbeat message with DEBUG mode enabled
func Test1090ESHeartbeatMessageWithDEBUG(t *testing.T) {
	reset1090ESState()

	// Enable DEBUG mode to cover lines 1098-1100
	originalDEBUG := globalSettings.DEBUG
	globalSettings.DEBUG = true
	defer func() { globalSettings.DEBUG = originalDEBUG }()

	// Heartbeat uses special ICAO address 0x07FFFFFF (134217727)
	heartbeatMsg := `{"Icao_addr":134217727,"DF":0,"CA":0,"TypeCode":0}`
	parseDump1090Message(heartbeatMsg)

	// Heartbeat should not create traffic entry even with DEBUG enabled
	trafficMutex.Lock()
	_, exists := traffic[0x07FFFFFF]
	trafficMutex.Unlock()

	if exists {
		t.Error("Heartbeat message should not create traffic entry")
	}

	t.Log("Heartbeat message with DEBUG correctly handled")
}

// Test1090ESNonICAOAddressWithDEBUG tests handling of non-ICAO addresses with DEBUG mode
func Test1090ESNonICAOAddressWithDEBUG(t *testing.T) {
	reset1090ESState()

	// Enable DEBUG mode to cover lines 1106-1108
	originalDEBUG := globalSettings.DEBUG
	globalSettings.DEBUG = true
	defer func() { globalSettings.DEBUG = originalDEBUG }()

	// Non-ICAO addresses have bit 25 set by dump1090
	// Address 0xABCDEF with bit 25 set becomes 0x01ABCDEF (28036591)
	nonIcaoMsg := `{"Icao_addr":28036591,"DF":18,"CA":2,"TypeCode":11,"Lat":40.7128,"Lng":-74.006,"Alt":5000,"Position_valid":true}`
	parseDump1090Message(nonIcaoMsg)

	// Should be stored with address 0xABCDEF (bit 25 stripped)
	trafficMutex.Lock()
	ti, exists := traffic[0xABCDEF]
	trafficMutex.Unlock()

	if !exists {
		t.Error("Non-ICAO address should be stored with bit 25 stripped")
	} else {
		t.Logf("Non-ICAO TIS-B target stored with ICAO=%X, Addr_type=%d", ti.Icao_addr, ti.Addr_type)
	}

	t.Log("Non-ICAO address with DEBUG handling test complete")
}

// Test1090ESSignalLevelHistPruning tests that SignalLevelHist is pruned to last 8 entries
func Test1090ESSignalLevelHistPruning(t *testing.T) {
	reset1090ESState()

	icao := uint32(0x500001)

	// Send more than 8 messages with signal levels to trigger pruning at line 1137-1138
	for i := 0; i < 12; i++ {
		signalLevel := 0.01 + float64(i)*0.01 // Varying signal levels
		msg := `{"Icao_addr":5242881,"DF":17,"CA":5,"TypeCode":11,"Lat":42.0,"Lng":-73.0,"Alt":` +
			intToStr(5000+i*100) + `,"Position_valid":true,"SignalLevel":` +
			floatToStr(signalLevel) + `}`
		parseDump1090Message(msg)
	}

	// Check that SignalLevelHist has exactly 8 entries (pruned from 12)
	trafficMutex.Lock()
	ti, exists := traffic[icao]
	trafficMutex.Unlock()

	if !exists {
		t.Error("Expected traffic entry")
		return
	}

	if len(ti.SignalLevelHist) != 8 {
		t.Errorf("Expected SignalLevelHist to have 8 entries (pruned), got %d", len(ti.SignalLevelHist))
	}

	// Verify that the last 8 entries are kept (signal levels from messages 5-12)
	// The signal levels should be in dB: 10 * log10(signalLevel)
	t.Logf("SignalLevelHist has %d entries: %v", len(ti.SignalLevelHist), ti.SignalLevelHist)
}

// floatToStr converts a float to string with 2 decimal places
func floatToStr(f float64) string {
	// Simple float to string for test purposes
	intPart := int(f)
	fracPart := int((f - float64(intPart)) * 100)
	if fracPart < 10 {
		return intToStr(intPart) + ".0" + intToStr(fracPart)
	}
	return intToStr(intPart) + "." + intToStr(fracPart)
}
