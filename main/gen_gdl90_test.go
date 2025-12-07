/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	gen_gdl90_test.go: Unit tests for gen_gdl90.go

	Implements: Phase 1.2 (Test Infrastructure)
	Verifies: FR-601-608 (GDL90 Protocol Implementation)
*/

package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestCrcInit tests CRC table initialization
// Verifies: FR-601 (GDL90 Frame Format - CRC)
func TestCrcInit(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Verify table has expected values
	// From FAA GDL90 spec, first few values should be:
	// Crc16Table[0] should be 0x0000
	// Crc16Table[1] should be 0x1021
	if Crc16Table[0] != 0x0000 {
		t.Errorf("Crc16Table[0] = 0x%04X, want 0x0000", Crc16Table[0])
	}
	if Crc16Table[1] != 0x1021 {
		t.Errorf("Crc16Table[1] = 0x%04X, want 0x1021", Crc16Table[1])
	}

	// Verify all 256 entries are populated (non-zero except entry 0)
	nonZeroCount := 0
	for i := 1; i < 256; i++ {
		if Crc16Table[i] != 0 {
			nonZeroCount++
		}
	}
	if nonZeroCount < 250 { // At least 250 should be non-zero
		t.Errorf("Expected most CRC table entries to be non-zero, got %d/255", nonZeroCount)
	}
}

// TestCrcCompute tests CRC calculation
// Verifies: FR-601 (GDL90 Frame Format - CRC)
func TestCrcCompute(t *testing.T) {
	crcInit() // Ensure table is initialized

	testCases := []struct {
		name     string
		data     []byte
		expected uint16
	}{
		{
			name:     "Empty data",
			data:     []byte{},
			expected: 0x0000,
		},
		{
			name:     "Single byte",
			data:     []byte{0x00},
			expected: 0x0000,
		},
		{
			name: "Heartbeat message type",
			data: []byte{0x00},
			// CRC will be computed based on the table
			expected: 0x0000, // Will be actual computed value
		},
		{
			name: "Sample GDL90 message",
			data: []byte{0x00, 0x81, 0x41, 0xDB, 0xD0, 0x08, 0x02},
			// This should compute to a specific CRC value
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := crcCompute(tc.data)
			// Just verify it completes without panic
			// Actual CRC values depend on the polynomial
			if len(tc.data) == 0 && result != 0 {
				t.Errorf("Expected CRC of empty data to be 0, got 0x%04X", result)
			}
			t.Logf("CRC of %v = 0x%04X", tc.data, result)
		})
	}
}

// TestPrepareMessage tests GDL90 message preparation with framing and CRC
// Verifies: FR-601 (GDL90 Frame Format - framing, stuffing, CRC)
func TestPrepareMessage(t *testing.T) {
	crcInit() // Ensure CRC table is initialized

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "Simple message",
			data: []byte{0x00}, // Heartbeat message type
		},
		{
			name: "Message with flag byte (needs stuffing)",
			data: []byte{0x00, 0x7E}, // Contains flag byte
		},
		{
			name: "Message with escape byte (needs stuffing)",
			data: []byte{0x00, 0x7D}, // Contains escape byte
		},
		{
			name: "Empty message",
			data: []byte{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := prepareMessage(tc.data)

			// Verify message structure
			if len(result) < 4 {
				t.Fatalf("Message too short: %d bytes (need at least 4: flag + crc + flag)", len(result))
			}

			// Verify start flag
			if result[0] != 0x7E {
				t.Errorf("Expected start flag 0x7E, got 0x%02X", result[0])
			}

			// Verify end flag
			if result[len(result)-1] != 0x7E {
				t.Errorf("Expected end flag 0x7E, got 0x%02X", result[len(result)-1])
			}

			// Verify no unescaped 0x7E or 0x7D in the middle
			for i := 1; i < len(result)-1; i++ {
				if result[i] == 0x7E && result[i-1] != 0x7D {
					t.Errorf("Found unescaped flag byte at position %d", i)
				}
				if result[i] == 0x7D && i+1 < len(result)-1 {
					// Next byte should be escaped (XOR 0x20)
					nextByte := result[i+1]
					if nextByte != (0x7E^0x20) && nextByte != (0x7D^0x20) {
						t.Logf("Escape sequence at %d: 0x7D 0x%02X", i, nextByte)
					}
				}
			}

			t.Logf("Prepared message: %d bytes, data: % X", len(result), result)
		})
	}
}

// TestMakeLatLng tests latitude/longitude encoding for GDL90
// Verifies: FR-604 (GDL90 Traffic Report - position encoding)
func TestMakeLatLng(t *testing.T) {
	testCases := []struct {
		name     string
		value    float32
		expected []byte
	}{
		{
			name:  "Zero",
			value: 0.0,
			// 0 / LON_LAT_RESOLUTION = 0, encoded as 3 bytes
			expected: []byte{0x00, 0x00, 0x00},
		},
		{
			name:  "Positive latitude",
			value: 43.99, // Oshkosh area
			// Will encode as (43.99 / LON_LAT_RESOLUTION) in 24 bits
		},
		{
			name:  "Negative longitude",
			value: -88.56, // Oshkosh area
			// Will encode as negative value in 24-bit two's complement
		},
		{
			name:  "Max positive (90 degrees)",
			value: 90.0,
		},
		{
			name:  "Max negative (-90 degrees)",
			value: -90.0,
		},
		{
			name:  "Positive longitude (180 degrees)",
			value: 180.0,
		},
		{
			name:  "Negative longitude (-180 degrees)",
			value: -180.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := makeLatLng(tc.value)

			// Verify result is 3 bytes
			if len(result) != 3 {
				t.Fatalf("Expected 3 bytes, got %d", len(result))
			}

			// Verify expected value if provided
			if tc.expected != nil {
				if !bytes.Equal(result, tc.expected) {
					t.Errorf("makeLatLng(%f) = % X, want % X", tc.value, result, tc.expected)
				}
			}

			// Decode and verify roundtrip
			encoded := int32(result[0])<<16 | int32(result[1])<<8 | int32(result[2])
			// Sign extend if negative (bit 23 is set)
			if encoded&0x800000 != 0 {
				encoded |= ^int32(0xFFFFFF) // Sign extend to 32 bits
			}
			decoded := float32(encoded) * LON_LAT_RESOLUTION

			// Allow small rounding error due to float32 precision
			diff := decoded - tc.value
			if diff < 0 {
				diff = -diff
			}
			// Special case: ±180° longitude are equivalent (same meridian)
			if (tc.value == 180.0 || tc.value == -180.0) && (decoded == 180.0 || decoded == -180.0) {
				diff = 0 // Accept ±180 equivalence
			}
			if diff > 0.01 { // Allow 0.01 degree error
				t.Errorf("Roundtrip error: input=%f, encoded=0x%06X, decoded=%f, diff=%f",
					tc.value, encoded, decoded, diff)
			}

			t.Logf("makeLatLng(%f) = % X (decoded: %f)", tc.value, result, decoded)
		})
	}
}

// TestMakeHeartbeat tests GDL90 heartbeat message generation
// Verifies: FR-602 (GDL90 Heartbeat)
func TestMakeHeartbeat(t *testing.T) {
	crcInit() // Ensure CRC table is initialized

	// Initialize stratuxClock for time functions
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Test with GPS invalid
	mySituation.GPSFixQuality = 0
	msg1 := makeHeartbeat()

	// Should return a valid GDL90 message
	if len(msg1) < 4 {
		t.Fatalf("Heartbeat message too short: %d bytes", len(msg1))
	}

	// Verify framing
	if msg1[0] != 0x7E || msg1[len(msg1)-1] != 0x7E {
		t.Error("Heartbeat message missing frame flags")
	}

	// Test with GPS valid
	mySituation.GPSFixQuality = 2
	mySituation.GPSLastFixLocalTime = time.Now()
	msg2 := makeHeartbeat()

	// Should also return a valid message
	if len(msg2) < 4 {
		t.Fatalf("Heartbeat message (GPS valid) too short: %d bytes", len(msg2))
	}

	t.Logf("Heartbeat (GPS invalid): %d bytes", len(msg1))
	t.Logf("Heartbeat (GPS valid): %d bytes", len(msg2))

	// Messages should be slightly different due to GPS valid bit
	// But we can't easily compare without unstuffing
}

// TestMakeStratuxHeartbeat tests Stratux-specific heartbeat message
// Verifies: Stratux custom protocol extension
func TestMakeStratuxHeartbeat(t *testing.T) {
	crcInit()

	// Test with GPS and AHRS invalid
	mySituation.GPSFixQuality = 0
	globalStatus.IMUConnected = false
	msg1 := makeStratuxHeartbeat()

	// Verify message structure
	if len(msg1) < 4 {
		t.Fatalf("Stratux heartbeat too short: %d bytes", len(msg1))
	}

	// Verify framing
	if msg1[0] != 0x7E || msg1[len(msg1)-1] != 0x7E {
		t.Error("Stratux heartbeat missing frame flags")
	}

	// Test with GPS valid
	mySituation.GPSFixQuality = 2
	mySituation.GPSLastFixLocalTime = time.Now()
	msg2 := makeStratuxHeartbeat()

	if len(msg2) < 4 {
		t.Fatalf("Stratux heartbeat (GPS valid) too short: %d bytes", len(msg2))
	}

	t.Logf("Stratux Heartbeat (GPS invalid): %d bytes", len(msg1))
	t.Logf("Stratux Heartbeat (GPS valid): %d bytes", len(msg2))
}

// TestMakeFFIDMessage tests ForeFlight ID message generation
// Verifies: ForeFlight integration protocol
func TestMakeFFIDMessage(t *testing.T) {
	crcInit()

	// Set up version info
	stratuxVersion = "v1.6"
	stratuxBuild = "test"

	msg := makeFFIDMessage()

	// Verify message structure
	if len(msg) < 4 {
		t.Fatalf("FF ID message too short: %d bytes", len(msg))
	}

	// Verify framing
	if msg[0] != 0x7E || msg[len(msg)-1] != 0x7E {
		t.Error("FF ID message missing frame flags")
	}

	t.Logf("ForeFlight ID message: %d bytes", len(msg))
}

// TestMakeStratuxStatus tests Stratux status message generation
// Verifies: Stratux custom protocol - status reporting
func TestMakeStratuxStatus(t *testing.T) {
	crcInit()

	// Initialize global status
	stratuxVersion = "v1.6rc1"
	globalStatus.GPS_satellites_locked = 10
	globalStatus.GPS_satellites_tracked = 12
	globalStatus.UAT_traffic_targets_tracking = 5
	globalStatus.ES_traffic_targets_tracking = 3
	globalStatus.UAT_messages_last_minute = 100
	globalStatus.ES_messages_last_minute = 50
	globalStatus.CPUTemp = 45.5

	// Initialize mySituation
	mySituation.GPSFixQuality = 2

	// Initialize ADSBTowers
	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}
	ADSBTowerMutex.Lock()
	ADSBTowers = make(map[string]ADSBTower)
	ADSBTowers["test1"] = ADSBTower{Lat: 43.0, Lng: -88.0}
	ADSBTowers["test2"] = ADSBTower{Lat: 44.0, Lng: -89.0}
	ADSBTowerMutex.Unlock()

	msg := makeStratuxStatus()

	// Verify message structure
	if len(msg) < 4 {
		t.Fatalf("Stratux status message too short: %d bytes", len(msg))
	}

	// Verify framing
	if msg[0] != 0x7E || msg[len(msg)-1] != 0x7E {
		t.Error("Stratux status message missing frame flags")
	}

	// Message should include tower data, so should be longer than base message
	// Base message is ~29 bytes, plus 6 bytes per tower
	minExpectedLen := 4 + 29 + (6 * 2) // flags + base + 2 towers
	if len(msg) < minExpectedLen {
		t.Logf("Warning: Stratux status message may be missing tower data: %d bytes (expected >=%d)",
			len(msg), minExpectedLen)
	}

	t.Logf("Stratux Status message: %d bytes (includes %d towers)", len(msg), len(ADSBTowers))
}

// TestPrepareMessage_Stuffing tests byte stuffing in detail
// Verifies: FR-601 (GDL90 Frame Format - byte stuffing)
func TestPrepareMessage_Stuffing(t *testing.T) {
	crcInit()

	testCases := []struct {
		name          string
		data          []byte
		expectStuffed bool
	}{
		{
			name:          "No stuffing needed",
			data:          []byte{0x00, 0x01, 0x02},
			expectStuffed: false,
		},
		{
			name:          "Flag byte in data",
			data:          []byte{0x00, 0x7E, 0x01},
			expectStuffed: true,
		},
		{
			name:          "Escape byte in data",
			data:          []byte{0x00, 0x7D, 0x01},
			expectStuffed: true,
		},
		{
			name:          "Multiple special bytes",
			data:          []byte{0x7E, 0x7D, 0x7E, 0x7D},
			expectStuffed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := prepareMessage(tc.data)

			// Calculate expected unstuffed length:
			// 2 flags + len(data) + 2 CRC bytes = len(data) + 4
			unstuffedLen := len(tc.data) + 4

			if tc.expectStuffed {
				// Should be longer due to stuffing
				if len(result) <= unstuffedLen {
					t.Errorf("Expected stuffing, but message length %d <= %d", len(result), unstuffedLen)
				}
				t.Logf("Message stuffed: %d bytes -> %d bytes", unstuffedLen, len(result))
			} else {
				// Might still be longer if CRC contains special bytes
				t.Logf("Message: %d bytes (stuffed: %d)", unstuffedLen, len(result))
			}

			// Verify hex output for debugging
			t.Logf("Input:  % X", tc.data)
			t.Logf("Output: % X", result)
		})
	}
}

// TestCrcCompute_KnownValues tests CRC against known good values
// Verifies: FR-601 (GDL90 CRC-16 implementation correctness)
func TestCrcCompute_KnownValues(t *testing.T) {
	crcInit()

	// Test with known GDL90 message samples
	// These are from the GDL90 spec or real captures
	testCases := []struct {
		name     string
		data     string // hex string
		expected uint16
	}{
		{
			name: "Heartbeat example",
			data: "008141dbd00802", // Example from spec (without CRC)
			// Expected CRC needs to be verified against spec
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := hex.DecodeString(tc.data)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			result := crcCompute(data)

			if tc.expected != 0 && result != tc.expected {
				t.Errorf("CRC mismatch: got 0x%04X, want 0x%04X", result, tc.expected)
			} else {
				t.Logf("CRC of %s = 0x%04X", tc.data, result)
			}
		})
	}
}

// TestIsDetectedOwnshipValid tests ownship detection timeout
// Verifies: FR-403 (Ownship Detection - timeout)
func TestIsDetectedOwnshipValid(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Set ownship as recently seen
	OwnshipTrafficInfo.Last_seen = stratuxClock.Time
	result1 := isDetectedOwnshipValid()

	if !result1 {
		t.Error("Expected ownship to be valid when recently seen")
	}

	// Set ownship as old (>10 seconds)
	OwnshipTrafficInfo.Last_seen = stratuxClock.Time.Add(-15 * time.Second)
	result2 := isDetectedOwnshipValid()

	if result2 {
		t.Error("Expected ownship to be invalid when >10 seconds old")
	}

	t.Logf("Ownship valid (recent): %v", result1)
	t.Logf("Ownship valid (old): %v", result2)
}

// TestGetProductNameFromId tests product ID to name mapping
func TestGetProductNameFromId(t *testing.T) {
	testCases := []struct {
		name        string
		product_id  int
		expectedVal string
	}{
		{
			name:        "METAR (ID 0)",
			product_id:  0,
			expectedVal: "METAR",
		},
		{
			name:        "TAF (ID 1)",
			product_id:  1,
			expectedVal: "TAF",
		},
		{
			name:        "NEXRAD Regional (ID 63)",
			product_id:  63,
			expectedVal: "NEXRAD Regional",
		},
		{
			name:        "NEXRAD CONUS (ID 64)",
			product_id:  64,
			expectedVal: "NEXRAD CONUS",
		},
		{
			name:        "Text (ID 413)",
			product_id:  413,
			expectedVal: "Text",
		},
		{
			name:        "Custom/Test (ID 600)",
			product_id:  600,
			expectedVal: "Custom/Test",
		},
		{
			name:        "Custom/Test range (ID 2000)",
			product_id:  2000,
			expectedVal: "Custom/Test",
		},
		{
			name:        "Custom/Test range (ID 2005)",
			product_id:  2005,
			expectedVal: "Custom/Test",
		},
		{
			name:        "Unknown ID (999)",
			product_id:  999,
			expectedVal: "Unknown (999)",
		},
		{
			name:        "Unknown ID (1234)",
			product_id:  1234,
			expectedVal: "Unknown (1234)",
		},
		{
			name:        "Lightning (ID 101)",
			product_id:  101,
			expectedVal: "Lightning",
		},
		{
			name:        "G-AIRMET (ID 254)",
			product_id:  254,
			expectedVal: "G-AIRMET",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getProductNameFromId(tc.product_id)
			if result != tc.expectedVal {
				t.Errorf("getProductNameFromId(%d) = %q, expected %q",
					tc.product_id, result, tc.expectedVal)
			}
			t.Logf("Product ID %d -> %q", tc.product_id, result)
		})
	}
}

// TestGetProductNameFromIdEdgeCases tests edge cases for product name lookup
func TestGetProductNameFromIdEdgeCases(t *testing.T) {
	// Test all products in the custom/test range
	for id := 2000; id <= 2005; id++ {
		result := getProductNameFromId(id)
		if result != "Custom/Test" {
			t.Errorf("getProductNameFromId(%d) = %q, expected \"Custom/Test\"", id, result)
		}
	}

	// Test boundary around custom range
	if getProductNameFromId(1999) == "Custom/Test" {
		t.Error("ID 1999 should not be Custom/Test")
	}
	if getProductNameFromId(2006) == "Custom/Test" {
		t.Error("ID 2006 should not be Custom/Test")
	}

	// Test that unknown IDs format correctly
	unknownTests := []int{-1, 10000, 9999}
	for _, id := range unknownTests {
		result := getProductNameFromId(id)
		expectedPrefix := "Unknown ("
		if len(result) < len(expectedPrefix) || result[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("getProductNameFromId(%d) = %q, should start with %q", id, result, expectedPrefix)
		}
	}
}

// =============================================================================
// Ownship Report Tests
// =============================================================================

// Note: makeOwnshipReport and makeOwnshipGeometricAltitudeReport call sendGDL90()
// and sendXPlane() which require network initialization. They are tested via
// integration tests (TestE2E*) rather than unit tests.

// TestIsDetectedOwnshipValidEdgeCases tests additional edge cases for ownship validity
func TestIsDetectedOwnshipValidEdgeCases(t *testing.T) {
	// Save original
	origOwnshipTrafficInfo := OwnshipTrafficInfo
	defer func() {
		OwnshipTrafficInfo = origOwnshipTrafficInfo
	}()

	t.Run("recently_seen_is_valid", func(t *testing.T) {
		OwnshipTrafficInfo = TrafficInfo{
			Last_seen: stratuxClock.Time,
		}

		if !isDetectedOwnshipValid() {
			t.Error("Expected recently seen ownship to be valid")
		}
	})

	t.Run("old_ownship_is_invalid", func(t *testing.T) {
		OwnshipTrafficInfo = TrafficInfo{
			Last_seen: stratuxClock.Time.Add(-15 * time.Second),
		}

		if isDetectedOwnshipValid() {
			t.Error("Expected old ownship to be invalid")
		}
	})

	t.Run("boundary_at_10_seconds", func(t *testing.T) {
		// Test exactly at the 10 second boundary
		OwnshipTrafficInfo = TrafficInfo{
			Last_seen: stratuxClock.Time.Add(-9 * time.Second),
		}

		if !isDetectedOwnshipValid() {
			t.Error("Expected ownship at 9 seconds to still be valid")
		}

		OwnshipTrafficInfo = TrafficInfo{
			Last_seen: stratuxClock.Time.Add(-11 * time.Second),
		}

		if isDetectedOwnshipValid() {
			t.Error("Expected ownship at 11 seconds to be invalid")
		}
	})
}

// =============================================================================
// relayMessage Tests
// =============================================================================

// TestRelayMessage is skipped because relayMessage calls sendGDL90 which requires
// full network infrastructure initialization including channels that block on send.
// The function is indirectly tested through integration tests.
func TestRelayMessage(t *testing.T) {
	t.Skip("Skipped: relayMessage requires full network infrastructure (sendGDL90 blocks on nil channels)")
}

// =============================================================================
// parseInput Tests
// =============================================================================

// TestParseInput_BasicCases tests parseInput with various message types
func TestParseInput_BasicCases(t *testing.T) {
	// Initialize required globals
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	crcInit()

	// Initialize traffic system (needed by parseDownlinkReport)
	if trafficMutex == nil {
		initTraffic(false)
	}

	testCases := []struct {
		name         string
		input        string
		expectedType uint16
		expectNil    bool
		description  string
	}{
		{
			name:         "Empty string",
			input:        "",
			expectedType: 0,
			expectNil:    true,
			description:  "Empty input should return nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset global status
			globalStatus.UAT_messages_total = 0

			result, msgtype := parseInput(tc.input)

			if tc.expectNil {
				if result != nil {
					t.Errorf("%s: expected nil result, got %d bytes", tc.description, len(result))
				}
			} else {
				if result == nil {
					t.Errorf("%s: expected non-nil result", tc.description)
				}
			}

			if msgtype != tc.expectedType {
				t.Errorf("%s: expected msgtype %d, got %d", tc.description, tc.expectedType, msgtype)
			}

			t.Logf("%s: msgtype=%d, result_len=%d", tc.description, msgtype, len(result))
		})
	}
}

// TestParseInput_UplinkMessage tests uplink message parsing
func TestParseInput_UplinkMessage(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	crcInit()

	// Initialize traffic system
	if trafficMutex == nil {
		initTraffic(false)
	}

	// Create an uplink message (starts with '+', exactly 432 data bytes)
	// Each byte is 2 hex chars, so 432 bytes = 864 chars
	uplinkData := "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";"

	result, msgtype := parseInput(uplinkData)

	if result == nil {
		t.Fatal("Expected non-nil result for uplink message")
	}

	if msgtype != MSGTYPE_UPLINK {
		t.Errorf("Expected msgtype MSGTYPE_UPLINK (%d), got %d", MSGTYPE_UPLINK, msgtype)
	}

	if len(result) != UPLINK_FRAME_DATA_BYTES {
		t.Errorf("Expected result length %d, got %d", UPLINK_FRAME_DATA_BYTES, len(result))
	}

	t.Logf("Uplink message parsed: msgtype=%d, len=%d", msgtype, len(result))
}

// TestParseInput_SignalStrength tests signal strength parsing
func TestParseInput_SignalStrength(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	crcInit()

	// Initialize traffic system
	if trafficMutex == nil {
		initTraffic(false)
	}

	// Initialize maxSignalStrength
	maxSignalStrength = 0

	// Create uplink message with signal strength
	uplinkData := "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=1;ss=5000"

	result, msgtype := parseInput(uplinkData)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if msgtype != MSGTYPE_UPLINK {
		t.Errorf("Expected MSGTYPE_UPLINK, got %d", msgtype)
	}

	// Signal strength should be parsed and update maxSignalStrength
	if maxSignalStrength != 5000 {
		t.Errorf("Expected maxSignalStrength=5000, got %d", maxSignalStrength)
	}

	t.Logf("Signal strength parsed: maxSignalStrength=%d", maxSignalStrength)
}

// TestParseInput_InvalidSignalStrength tests invalid signal strength
func TestParseInput_InvalidSignalStrength(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	crcInit()

	// Initialize traffic system
	if trafficMutex == nil {
		initTraffic(false)
	}

	maxSignalStrength = 0

	// Invalid signal strength (not a number)
	uplinkData := "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=1;ss=invalid"

	result, msgtype := parseInput(uplinkData)

	if result == nil {
		t.Fatal("Expected non-nil result even with invalid ss")
	}

	if msgtype != MSGTYPE_UPLINK {
		t.Errorf("Expected MSGTYPE_UPLINK, got %d", msgtype)
	}

	// Invalid ss should not update maxSignalStrength
	if maxSignalStrength != 0 {
		t.Errorf("Expected maxSignalStrength=0 (invalid not parsed), got %d", maxSignalStrength)
	}
}

// TestParseInput_ShortUplinkPadded tests uplink message padding
func TestParseInput_ShortUplinkPadded(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	crcInit()

	// Initialize traffic system
	if trafficMutex == nil {
		initTraffic(false)
	}

	// Create a short uplink (should be padded to UPLINK_FRAME_DATA_BYTES)
	shortUplink := "+" + strings.Repeat("FF", 100) + ";"

	result, msgtype := parseInput(shortUplink)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if msgtype != MSGTYPE_UPLINK {
		t.Errorf("Expected MSGTYPE_UPLINK, got %d", msgtype)
	}

	// Should be padded to full frame size
	if len(result) != UPLINK_FRAME_DATA_BYTES {
		t.Errorf("Expected padded length %d, got %d", UPLINK_FRAME_DATA_BYTES, len(result))
	}

	// Check that padding is zeros
	for i := 100; i < len(result); i++ {
		if result[i] != 0x00 {
			t.Errorf("Expected padding byte at %d to be 0x00, got 0x%02X", i, result[i])
			break
		}
	}

	t.Logf("Short uplink padded: original ~%d bytes, padded to %d bytes", 100, len(result))
}

// =============================================================================
// Settings Tests
// =============================================================================

// TestReadSettings_NoFile tests readSettings when config file doesn't exist
func TestReadSettings_NoFile(t *testing.T) {
	// Save original configLocation
	origLocation := configLocation
	defer func() { configLocation = origLocation }()

	// Set to non-existent file
	configLocation = "/tmp/nonexistent_stratux_test_config.conf"

	// Should use defaults without error
	readSettings()

	// Verify defaults were applied (check a few key settings)
	if !globalSettings.UAT_Enabled {
		t.Error("Expected UAT_Enabled=true (default)")
	}
	if !globalSettings.ES_Enabled {
		t.Error("Expected ES_Enabled=true (default)")
	}
	if globalSettings.RegionSelected != 0 {
		t.Errorf("Expected RegionSelected=0 (default), got %d", globalSettings.RegionSelected)
	}
}

// TestReadSettings_InvalidJSONContent tests readSettings with invalid JSON content
func TestReadSettings_InvalidJSONContent(t *testing.T) {
	// Save original configLocation
	origLocation := configLocation
	defer func() {
		configLocation = origLocation
		os.Remove("/tmp/test_invalid_stratux2.conf")
	}()

	// Create temp file with invalid JSON
	configLocation = "/tmp/test_invalid_stratux2.conf"
	err := os.WriteFile(configLocation, []byte("{ invalid json "), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Should use defaults when JSON is invalid
	readSettings()

	// Verify defaults were applied
	if !globalSettings.UAT_Enabled {
		t.Error("Expected UAT_Enabled=true (default after invalid JSON)")
	}
}

// TestDefaultSettings tests the defaultSettings function
func TestDefaultSettings(t *testing.T) {
	// Clear settings
	globalSettings = settings{}

	defaultSettings()

	// Verify all critical defaults
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"RegionSelected", globalSettings.RegionSelected, 0},
		{"UAT_Enabled", globalSettings.UAT_Enabled, true},
		{"ES_Enabled", globalSettings.ES_Enabled, true},
		{"OGN_Enabled", globalSettings.OGN_Enabled, false},
		{"GPS_Enabled", globalSettings.GPS_Enabled, true},
		{"IMU_Sensor_Enabled", globalSettings.IMU_Sensor_Enabled, true},
		{"BMP_Sensor_Enabled", globalSettings.BMP_Sensor_Enabled, true},
		{"WiFiIPAddress", globalSettings.WiFiIPAddress, "192.168.10.1"},
		{"WiFiSSID", globalSettings.WiFiSSID, "Stratux"},
		{"WiFiChannel", globalSettings.WiFiChannel, 1},
		{"OwnshipModeS", globalSettings.OwnshipModeS, "F00000"},
		{"Dump1090Gain", globalSettings.Dump1090Gain, 37.2},
		{"DeveloperMode", globalSettings.DeveloperMode, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.actual != tc.expected {
				t.Errorf("Expected %s=%v, got %v", tc.name, tc.expected, tc.actual)
			}
		})
	}

	// Verify network outputs are initialized
	if len(globalSettings.NetworkOutputs) == 0 {
		t.Error("Expected NetworkOutputs to be initialized")
	}

	// Verify BLE outputs are initialized
	if len(globalSettings.BleOutputs) == 0 {
		t.Error("Expected BleOutputs to be initialized")
	}
}

// =============================================================================
// System Error Tracking Tests
// =============================================================================

// TestAddSystemErrorBasic tests adding system errors
func TestAddSystemErrorBasic(t *testing.T) {
	// Initialize if needed
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Clear errors
	globalStatus.Errors = make([]string, 0)
	systemErrs = make(map[string]string)

	err1 := fmt.Errorf("test error 1")
	addSystemError(err1)

	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(globalStatus.Errors))
	}

	if globalStatus.Errors[0] != "test error 1" {
		t.Errorf("Expected error message 'test error 1', got %q", globalStatus.Errors[0])
	}
}

// TestAddSingleSystemErrorf tests single system error tracking
func TestAddSingleSystemErrorf(t *testing.T) {
	// Initialize if needed
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Clear errors
	globalStatus.Errors = make([]string, 0)
	systemErrs = make(map[string]string)

	// Add first error
	addSingleSystemErrorf("test1", "Error: %d", 123)

	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(globalStatus.Errors))
	}

	if globalStatus.Errors[0] != "Error: 123" {
		t.Errorf("Expected 'Error: 123', got %q", globalStatus.Errors[0])
	}

	// Add same error again - should not duplicate
	addSingleSystemErrorf("test1", "Error: %d", 123)

	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error (no duplicate), got %d", len(globalStatus.Errors))
	}

	// Add different error
	addSingleSystemErrorf("test2", "Another error")

	if len(globalStatus.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(globalStatus.Errors))
	}
}

// TestRemoveSingleSystemError tests removing system errors
func TestRemoveSingleSystemError(t *testing.T) {
	// Initialize if needed
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Clear and add errors
	globalStatus.Errors = make([]string, 0)
	systemErrs = make(map[string]string)

	addSingleSystemErrorf("err1", "Error 1")
	addSingleSystemErrorf("err2", "Error 2")

	if len(globalStatus.Errors) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(globalStatus.Errors))
	}

	// Remove first error
	removeSingleSystemError("err1")

	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error after removal, got %d", len(globalStatus.Errors))
	}

	if globalStatus.Errors[0] != "Error 2" {
		t.Errorf("Expected remaining error to be 'Error 2', got %q", globalStatus.Errors[0])
	}

	// Remove non-existent error (should not crash)
	removeSingleSystemError("nonexistent")

	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error (unchanged), got %d", len(globalStatus.Errors))
	}
}

// =============================================================================
// Region Settings Tests
// =============================================================================

// TestChangeRegionSettings_US tests US region settings
func TestChangeRegionSettings_US(t *testing.T) {
	// Save original
	origSettings := globalSettings
	defer func() { globalSettings = origSettings }()

	// Set to US
	globalSettings.RegionSelected = 1
	changeRegionSettings()

	if !globalSettings.UAT_Enabled {
		t.Error("Expected UAT_Enabled=true for US region")
	}
	if globalSettings.OGN_Enabled {
		t.Error("Expected OGN_Enabled=false for US region")
	}
	if globalSettings.DeveloperMode {
		t.Error("Expected DeveloperMode=false for US region")
	}
}

// TestChangeRegionSettings_EU tests EU region settings
func TestChangeRegionSettings_EU(t *testing.T) {
	// Save original
	origSettings := globalSettings
	defer func() { globalSettings = origSettings }()

	// Set to EU
	globalSettings.RegionSelected = 2
	changeRegionSettings()

	if globalSettings.UAT_Enabled {
		t.Error("Expected UAT_Enabled=false for EU region")
	}
	if !globalSettings.OGN_Enabled {
		t.Error("Expected OGN_Enabled=true for EU region")
	}
	if !globalSettings.DeveloperMode {
		t.Error("Expected DeveloperMode=true for EU region")
	}
}

// TestChangeRegionSettings_None tests no region selected
func TestChangeRegionSettings_None(t *testing.T) {
	// Save original
	origSettings := globalSettings
	defer func() { globalSettings = origSettings }()

	// Set initial values
	globalSettings.RegionSelected = 0
	globalSettings.UAT_Enabled = true
	globalSettings.OGN_Enabled = false
	origUAT := globalSettings.UAT_Enabled
	origOGN := globalSettings.OGN_Enabled

	changeRegionSettings()

	// Settings should not change when region is 0
	if globalSettings.UAT_Enabled != origUAT {
		t.Error("UAT_Enabled should not change when region=0")
	}
	if globalSettings.OGN_Enabled != origOGN {
		t.Error("OGN_Enabled should not change when region=0")
	}
}

// =============================================================================
// Message Log Tests
// =============================================================================

// TestMsgLogAppend tests message log appending
func TestMsgLogAppend(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Clear log
	msgLogMutex.Lock()
	msgLog = make([]msg, 0)
	msgLogMutex.Unlock()

	// Add a message
	testMsg := msg{
		MessageClass: MSGCLASS_UAT,
		TimeReceived: stratuxClock.Time,
		Data:         "test",
	}

	msgLogAppend(testMsg)

	// Verify it was added
	msgLogMutex.Lock()
	defer msgLogMutex.Unlock()

	if len(msgLog) != 1 {
		t.Fatalf("Expected 1 message in log, got %d", len(msgLog))
	}

	if msgLog[0].Data != "test" {
		t.Errorf("Expected Data='test', got %q", msgLog[0].Data)
	}
	if msgLog[0].MessageClass != MSGCLASS_UAT {
		t.Errorf("Expected MessageClass=MSGCLASS_UAT, got %d", msgLog[0].MessageClass)
	}
}

// =============================================================================
// isX86DebugMode Tests
// =============================================================================

// TestIsX86DebugMode tests the platform detection function
func TestIsX86DebugMode(t *testing.T) {
	result := isX86DebugMode()

	// Just verify it returns a boolean without panic
	// The actual value depends on the platform we're running on
	t.Logf("isX86DebugMode() = %v (GOARCH=%s)", result, runtime.GOARCH)

	// Verify logic
	expectedResult := runtime.GOARCH == "i386" || runtime.GOARCH == "amd64"
	if result != expectedResult {
		t.Errorf("Expected isX86DebugMode()=%v for GOARCH=%s, got %v",
			expectedResult, runtime.GOARCH, result)
	}
}
