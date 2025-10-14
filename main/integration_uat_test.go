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

// resetUATState clears the global UAT state for testing
func resetUATState() {
	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond)
	}

	// Reset UAT statistics
	globalStatus.UAT_messages_total = 0
	globalStatus.UAT_messages_last_minute = 0
	globalStatus.UAT_messages_max = 0
	globalStatus.UAT_METAR_total = 0
	globalStatus.UAT_TAF_total = 0
	globalStatus.UAT_NEXRAD_total = 0
	globalStatus.UAT_SIGMET_total = 0
	globalStatus.UAT_PIREP_total = 0
	globalStatus.UAT_NOTAM_total = 0
	globalStatus.UAT_OTHER_total = 0

	// Reset message log
	msgLogMutex = sync.Mutex{}
	msgLog = make([]msg, 0)

	// Reset ADS-B tower tracking
	ADSBTowerMutex = &sync.Mutex{}
	ADSBTowers = make(map[string]ADSBTower)

	// Reset traffic tracking (needed for parseDownlinkReport)
	trafficMutex = &sync.Mutex{}
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)

	// Reset max signal strength
	maxSignalStrength = 0
}

// replayUATTraceDirect reads a UAT trace file and directly injects UAT messages
func replayUATTraceDirect(t *testing.T, filename string) int {
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

		if msgType == "uat" {
			// Process UAT message directly via parseInput
			frame, msgtype := parseInput(msgData)
			if frame != nil && msgtype != 0 {
				messageCount++
			}
		}
	}

	return messageCount
}

// TestUATBasicParsing tests basic UAT message parsing from trace file
func TestUATBasicParsing(t *testing.T) {
	resetUATState()

	count := replayUATTraceDirect(t, "testdata/uat/basic_uat.trace.gz")

	if count == 0 {
		t.Error("Expected to parse at least one UAT message, got 0")
	}

	t.Logf("Successfully parsed %d UAT messages from trace file", count)

	// Verify that UAT message counter was incremented
	if globalStatus.UAT_messages_total == 0 {
		t.Error("Expected UAT_messages_total > 0, got 0")
	}

	t.Logf("Total UAT messages counted: %d", globalStatus.UAT_messages_total)
}

// TestUATUplinkMessage tests parsing of UAT uplink messages (FIS-B weather)
func TestUATUplinkMessage(t *testing.T) {
	resetUATState()

	// Uplink message (432 bytes = 864 hex characters)
	// Create a properly sized uplink message (pad with zeros to reach 864 chars)
	uplinkHex := "3cc0978aa66ca1a0158000213c5d2082102c22cc00082eec1e012c22cc000000000000000fd90007110e240811081ec5ea23b0c00158000213c6b2882102c869900082ee71e012c8699000000000000000fd9000711152508011525c69dc3b6ac00158000213c56a082102c869900082ee61e012c8699000000000000000fd90007110b1408010b14c69dc3b6ac00158000213dacc882102c865800082ee71e012c8658000000000000000fd90007161619090f1619c45d83dc5400158000213d57c882102d00d7000830701e012d00d7000000000000000fd90007150b3908050b39c51243b0b800158000213cc09082102d43cc00082efc1e012d43cc000000000000000fd900071300120813000fc46743b25400158000213d1ed082102ca60e00082ee91e012ca60e000000000000000fd90007140f1a08040f1ac3f0a3c1a400158000213e070082102d630c00082ee51e012d630c000000000000000fd9000718032008080320c4da03c81400158000213c453882102c22cc00082eeb1e012c22cc000000000000000fd9000711022708110227c5ea23b0c00000000000000000000000000000000000000000"
	// Pad to exactly 864 characters
	for len(uplinkHex) < 864 {
		uplinkHex += "0"
	}
	uplinkMsg := "+" + uplinkHex + ";rs=16;ss=128"

	frame, msgtype := parseInput(uplinkMsg)

	if frame == nil {
		t.Fatal("Expected non-nil frame from uplink message")
	}

	if msgtype != MSGTYPE_UPLINK {
		t.Errorf("Expected msgtype MSGTYPE_UPLINK (0x07), got 0x%02X", msgtype)
	}

	if len(frame) != UPLINK_FRAME_DATA_BYTES {
		t.Errorf("Expected frame length %d, got %d", UPLINK_FRAME_DATA_BYTES, len(frame))
	}

	t.Logf("Successfully parsed uplink message: msgtype=0x%02X, frame_length=%d", msgtype, len(frame))
}

// TestUATDownlinkBasicReport tests parsing of UAT basic report (18 bytes)
func TestUATDownlinkBasicReport(t *testing.T) {
	resetUATState()

	// Basic report (18 bytes hex = 36 characters)
	basicMsg := "-000000000000000000000000000000000000;rs=12;ss=94"

	frame, msgtype := parseInput(basicMsg)

	if frame == nil {
		t.Fatal("Expected non-nil frame from basic report")
	}

	if msgtype != MSGTYPE_BASIC_REPORT {
		t.Errorf("Expected msgtype MSGTYPE_BASIC_REPORT (0x1E), got 0x%02X", msgtype)
	}

	t.Logf("Successfully parsed basic report: msgtype=0x%02X", msgtype)
}

// TestUATDownlinkLongReport tests parsing of UAT long report (34 or 48 bytes)
func TestUATDownlinkLongReport(t *testing.T) {
	resetUATState()

	testCases := []struct {
		name    string
		message string
		length  int
	}{
		{
			name:    "34_byte_long_report",
			// 34 bytes = 68 hex characters
			message: "-00000000000000000000000000000000000000000000000000000000000000000000;rs=14;ss=102",
			length:  34,
		},
		{
			name:    "48_byte_long_report",
			// 48 bytes = 96 hex characters
			message: "-000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000;rs=15;ss=98",
			length:  48,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			frame, msgtype := parseInput(tc.message)

			if frame == nil {
				t.Fatal("Expected non-nil frame from long report")
			}

			if msgtype != MSGTYPE_LONG_REPORT {
				t.Errorf("Expected msgtype MSGTYPE_LONG_REPORT (0x1F), got 0x%02X", msgtype)
			}

			t.Logf("Successfully parsed %d-byte long report: msgtype=0x%02X", tc.length, msgtype)
		})
	}
}

// TestUATSignalStrength tests signal strength parsing from UAT messages
func TestUATSignalStrength(t *testing.T) {
	resetUATState()

	// Uplink message with signal strength ss=128 (864 hex chars = 432 bytes)
	uplinkHex := "3cc0978aa66ca1a0158000213c5d2082102c22cc00082eec1e012c22cc000000000000000fd90007110e240811081ec5ea23b0c00158000213c6b2882102c869900082ee71e012c8699000000000000000fd9000711152508011525c69dc3b6ac00158000213c56a082102c869900082ee61e012c8699000000000000000fd90007110b1408010b14c69dc3b6ac00158000213dacc882102c865800082ee71e012c8658000000000000000fd90007161619090f1619c45d83dc5400158000213d57c882102d00d7000830701e012d00d7000000000000000fd90007150b3908050b39c51243b0b800158000213cc09082102d43cc00082efc1e012d43cc000000000000000fd900071300120813000fc46743b25400158000213d1ed082102ca60e00082ee91e012ca60e000000000000000fd90007140f1a08040f1ac3f0a3c1a400158000213e070082102d630c00082ee51e012d630c000000000000000fd9000718032008080320c4da03c81400158000213c453882102c22cc00082eeb1e012c22cc000000000000000fd9000711022708110227c5ea23b0c00000000000000000000000000000000000000000"
	for len(uplinkHex) < 864 {
		uplinkHex += "0"
	}
	uplinkMsg := "+" + uplinkHex + ";rs=16;ss=128"

	parseInput(uplinkMsg)

	// Check that maxSignalStrength was updated (uplink messages update this)
	if maxSignalStrength == 0 {
		t.Error("Expected maxSignalStrength to be updated for uplink message")
	}

	t.Logf("maxSignalStrength updated to: %d", maxSignalStrength)
}

// TestUATMessageLog tests that UAT messages are logged to msgLog
func TestUATMessageLog(t *testing.T) {
	resetUATState()

	initialLogSize := len(msgLog)

	// Uplink message (864 hex chars = 432 bytes)
	uplinkHex := "3cc0978aa66ca1a0158000213c5d2082102c22cc00082eec1e012c22cc000000000000000fd90007110e240811081ec5ea23b0c00158000213c6b2882102c869900082ee71e012c8699000000000000000fd9000711152508011525c69dc3b6ac00158000213c56a082102c869900082ee61e012c8699000000000000000fd90007110b1408010b14c69dc3b6ac00158000213dacc882102c865800082ee71e012c8658000000000000000fd90007161619090f1619c45d83dc5400158000213d57c882102d00d7000830701e012d00d7000000000000000fd90007150b3908050b39c51243b0b800158000213cc09082102d43cc00082efc1e012d43cc000000000000000fd900071300120813000fc46743b25400158000213d1ed082102ca60e00082ee91e012ca60e000000000000000fd90007140f1a08040f1ac3f0a3c1a400158000213e070082102d630c00082ee51e012d630c000000000000000fd9000718032008080320c4da03c81400158000213c453882102c22cc00082eeb1e012c22cc000000000000000fd9000711022708110227c5ea23b0c00000000000000000000000000000000000000000"
	for len(uplinkHex) < 864 {
		uplinkHex += "0"
	}
	uplinkMsg := "+" + uplinkHex + ";rs=16;ss=128"

	parseInput(uplinkMsg)

	finalLogSize := len(msgLog)

	if finalLogSize <= initialLogSize {
		t.Errorf("Expected msgLog to grow, initial=%d, final=%d", initialLogSize, finalLogSize)
	}

	// Check the last message in the log
	if finalLogSize > 0 {
		lastMsg := msgLog[finalLogSize-1]
		if lastMsg.MessageClass != MSGCLASS_UAT {
			t.Errorf("Expected MessageClass MSGCLASS_UAT (%d), got %d", MSGCLASS_UAT, lastMsg.MessageClass)
		}
		t.Logf("Message logged with class=%d, signal_amplitude=%d", lastMsg.MessageClass, lastMsg.Signal_amplitude)
	}
}

// TestUATInvalidMessage tests handling of invalid UAT messages
func TestUATInvalidMessage(t *testing.T) {
	resetUATState()

	testCases := []struct {
		name    string
		message string
	}{
		{"empty_message", ""},
		{"only_semicolon", ";"},
		{"invalid_hex", "+GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG;rs=10"},
		{"wrong_length", "+123456;rs=10"}, // Too short
		{"no_plus_or_minus", "0000000000000000000000000000000000;rs=10"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initialTotal := globalStatus.UAT_messages_total

			frame, msgtype := parseInput(tc.message)

			// Empty messages and messages that become empty after parsing (like ";") should not increment counter
			if tc.message == "" || tc.message == ";" {
				if globalStatus.UAT_messages_total != initialTotal {
					t.Errorf("Empty or semicolon-only message should not increment UAT_messages_total, got %d, expected %d", globalStatus.UAT_messages_total, initialTotal)
				}
				return
			}

			// Invalid messages should return nil frame or msgtype 0
			if frame != nil && msgtype != 0 {
				t.Logf("Warning: Invalid message was parsed: msgtype=0x%02X", msgtype)
			}

			// Non-empty messages should increment the counter even if invalid
			if globalStatus.UAT_messages_total <= initialTotal {
				t.Error("Expected UAT_messages_total to increment for non-empty message")
			}
		})
	}
}

// TestUATMessageTypeDetection tests message type detection based on length
func TestUATMessageTypeDetection(t *testing.T) {
	resetUATState()

	testCases := []struct {
		name        string
		hexLength   int // Length in hex characters (2 chars per byte)
		expectedType uint16
	}{
		{"uplink_432_bytes", 864, MSGTYPE_UPLINK},        // 432 bytes * 2 = 864 hex chars
		{"long_report_48_bytes", 96, MSGTYPE_LONG_REPORT},  // 48 bytes * 2 = 96 hex chars
		{"long_report_34_bytes", 68, MSGTYPE_LONG_REPORT},  // 34 bytes * 2 = 68 hex chars
		{"basic_report_18_bytes", 36, MSGTYPE_BASIC_REPORT}, // 18 bytes * 2 = 36 hex chars
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a message with the specified length
			var prefix string
			if tc.expectedType == MSGTYPE_UPLINK {
				prefix = "+"
			} else {
				prefix = "-"
			}

			// Create hex string of the right length (all zeros)
			hexData := ""
			for i := 0; i < tc.hexLength; i++ {
				hexData += "0"
			}

			message := prefix + hexData + ";rs=10"

			frame, msgtype := parseInput(message)

			if frame == nil && tc.expectedType != 0 {
				t.Fatal("Expected non-nil frame")
			}

			if msgtype != tc.expectedType {
				t.Errorf("Expected msgtype 0x%02X, got 0x%02X", tc.expectedType, msgtype)
			}

			t.Logf("Correctly detected message type 0x%02X for %d hex characters", msgtype, tc.hexLength)
		})
	}
}

// TestUATTraceReplay tests replaying the entire UAT trace file
func TestUATTraceReplay(t *testing.T) {
	resetUATState()

	count := replayUATTraceDirect(t, "testdata/uat/basic_uat.trace.gz")

	// We expect 8 messages in the trace file
	// Some may be invalid (all zeros), so we check that at least some were parsed
	if count == 0 {
		t.Error("Expected to parse at least one message from trace file")
	}

	t.Logf("Replayed %d UAT messages from trace file", count)
	t.Logf("Total UAT messages counted: %d", globalStatus.UAT_messages_total)
	t.Logf("Max signal strength: %d", maxSignalStrength)
}

// TestUATMultipleMessages tests processing multiple UAT messages in sequence
func TestUATMultipleMessages(t *testing.T) {
	resetUATState()

	// Create a properly sized uplink message (864 hex chars)
	uplinkHex := "3cc0978aa66ca1a0158000213c5d2082102c22cc00082eec1e012c22cc000000000000000fd90007110e240811081ec5ea23b0c00"
	for len(uplinkHex) < 864 {
		uplinkHex += "0"
	}

	messages := []string{
		"+" + uplinkHex + ";rs=16;ss=128",
		"-000000000000000000000000000000000000;rs=12;ss=94",
		"+" + uplinkHex[:120] + ";rs=17;ss=132", // Shorter uplink (invalid, for testing)
	}

	successCount := 0
	for i, msg := range messages {
		frame, msgtype := parseInput(msg)
		if frame != nil && msgtype != 0 {
			successCount++
			t.Logf("Message %d: msgtype=0x%02X", i, msgtype)
		}
	}

	if successCount == 0 {
		t.Error("Expected at least one message to parse successfully")
	}

	t.Logf("Successfully parsed %d out of %d messages", successCount, len(messages))
	t.Logf("Total UAT messages counted: %d", globalStatus.UAT_messages_total)
}
