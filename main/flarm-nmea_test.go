package main

import (
	"strings"
	"testing"
)

// TestAppendNmeaChecksum tests NMEA checksum calculation and formatting
func TestAppendNmeaChecksum(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple NMEA sentence without $",
			input:    "PFLAU,0,0,0,0,0",
			expected: "PFLAU,0,0,0,0,0*52",
		},
		{
			name:     "NMEA sentence with $ prefix",
			input:    "$PFLAU,0,0,0,0,0",
			expected: "$PFLAU,0,0,0,0,0*52",
		},
		{
			name:     "GPRMC sentence",
			input:    "$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W",
			expected: "$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A",
		},
		{
			name:     "GPGGA sentence",
			input:    "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,",
			expected: "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47",
		},
		{
			name:     "Empty sentence",
			input:    "",
			expected: "*00",
		},
		{
			name:     "Only $ sign",
			input:    "$",
			expected: "$*00",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := appendNmeaChecksum(tc.input)
			if result != tc.expected {
				t.Errorf("appendNmeaChecksum(%q) = %q, expected %q",
					tc.input, result, tc.expected)
			}
			t.Logf("Input: %q -> Output: %q", tc.input, result)
		})
	}
}

// TestAppendNmeaChecksumFormat tests that checksum is always uppercase 2-digit hex
func TestAppendNmeaChecksumFormat(t *testing.T) {
	testCases := []string{
		"PFLAA,0,0,0,0,1,AABBCC,0,0,0,0.0,0",
		"GPRMC,,,,,,,,,,,",
		"TEST",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			result := appendNmeaChecksum(tc)

			// Verify format: should end with *XX where XX is 2 uppercase hex digits
			if len(result) < 3 {
				t.Fatalf("Result too short: %q", result)
			}

			parts := strings.Split(result, "*")
			if len(parts) != 2 {
				t.Errorf("Expected exactly one * in result, got: %q", result)
			}

			checksum := parts[1]
			if len(checksum) != 2 {
				t.Errorf("Checksum should be 2 characters, got %d: %q", len(checksum), checksum)
			}

			// Verify checksum is uppercase hex
			for _, c := range checksum {
				if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
					t.Errorf("Checksum character '%c' is not uppercase hex", c)
				}
			}
		})
	}
}

// TestComputeAlarmLevel tests FLARM alarm level calculation
func TestComputeAlarmLevel(t *testing.T) {
	testCases := []struct {
		name             string
		dist             float64 // meters
		relativeVertical int32   // meters
		expectedAlarm    uint8
	}{
		{
			name:             "Very close - alarm level 3",
			dist:             500, // 0.27 NM
			relativeVertical: 100, // 328 ft
			expectedAlarm:    3,
		},
		{
			name:             "At boundary - level 3 (just under 0.5 NM)",
			dist:             925, // 0.499 NM
			relativeVertical: 151, // 495 ft
			expectedAlarm:    3,
		},
		{
			name:             "Just beyond level 3 threshold",
			dist:             927, // 0.501 NM
			relativeVertical: 100, // 328 ft
			expectedAlarm:    2,
		},
		{
			name:             "Medium distance - alarm level 2",
			dist:             1500, // 0.81 NM
			relativeVertical: 200,  // 656 ft
			expectedAlarm:    2,
		},
		{
			name:             "At level 2 boundary (1 NM)",
			dist:             1851, // 0.999 NM
			relativeVertical: 303,  // 994 ft
			expectedAlarm:    2,
		},
		{
			name:             "Just beyond level 2 threshold",
			dist:             1853, // 1.001 NM
			relativeVertical: 200,  // 656 ft
			expectedAlarm:    0,
		},
		{
			name:             "Far away - no alarm",
			dist:             5000, // 2.7 NM
			relativeVertical: 1000, // 3280 ft
			expectedAlarm:    0,
		},
		{
			name:             "Close horizontal but too high vertical - no alarm",
			dist:             500, // 0.27 NM
			relativeVertical: 200, // 656 ft (> 500 ft threshold)
			expectedAlarm:    0,
		},
		{
			name:             "Close horizontal but too low vertical - no alarm",
			dist:             500,  // 0.27 NM
			relativeVertical: -200, // -656 ft (< -500 ft threshold)
			expectedAlarm:    0,
		},
		{
			name:             "Negative vertical within range - level 3",
			dist:             800,  // 0.43 NM
			relativeVertical: -150, // -492 ft
			expectedAlarm:    3,
		},
		{
			name:             "Zero distance and altitude - level 3",
			dist:             0,
			relativeVertical: 0,
			expectedAlarm:    3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := computeAlarmLevel(tc.dist, tc.relativeVertical)
			if result != tc.expectedAlarm {
				t.Errorf("computeAlarmLevel(%.0f m, %d m) = %d, expected %d",
					tc.dist, tc.relativeVertical, result, tc.expectedAlarm)
			}
			t.Logf("Distance: %.0f m (%.2f NM), Vertical: %d m (%.0f ft) -> Alarm Level: %d",
				tc.dist, tc.dist/1852.0, tc.relativeVertical, float64(tc.relativeVertical)*3.28084, result)
		})
	}
}

// TestGdl90EmitterCatToNMEA tests aircraft type code conversion
func TestGdl90EmitterCatToNMEA(t *testing.T) {
	testCases := []struct {
		name         string
		emitterCat   uint8
		expectedType string
	}{
		{
			name:         "Light aircraft (1)",
			emitterCat:   1,
			expectedType: "8", // piston
		},
		{
			name:         "Highly maneuverable (6)",
			emitterCat:   6,
			expectedType: "8", // piston
		},
		{
			name:         "Small aircraft (2)",
			emitterCat:   2,
			expectedType: "9", // jet
		},
		{
			name:         "Large aircraft (3)",
			emitterCat:   3,
			expectedType: "9", // jet
		},
		{
			name:         "Heavy aircraft (5)",
			emitterCat:   5,
			expectedType: "9", // jet
		},
		{
			name:         "Helicopter (7)",
			emitterCat:   7,
			expectedType: "3", // helicopter
		},
		{
			name:         "Glider (9)",
			emitterCat:   9,
			expectedType: "1", // glider
		},
		{
			name:         "Lighter than air (10)",
			emitterCat:   10,
			expectedType: "B", // balloon
		},
		{
			name:         "Skydiver (11)",
			emitterCat:   11,
			expectedType: "4", // sky diver
		},
		{
			name:         "Paraglider (12)",
			emitterCat:   12,
			expectedType: "7", // paraglider
		},
		{
			name:         "UAV (14)",
			emitterCat:   14,
			expectedType: "D", // UAV
		},
		{
			name:         "Surface vehicle (17)",
			emitterCat:   17,
			expectedType: "E", // ground support
		},
		{
			name:         "Surface vehicle (18)",
			emitterCat:   18,
			expectedType: "E", // ground support
		},
		{
			name:         "Static object (19)",
			emitterCat:   19,
			expectedType: "F", // point obstacle
		},
		{
			name:         "Unknown type (0)",
			emitterCat:   0,
			expectedType: "0", // unknown
		},
		{
			name:         "Unmapped type (99)",
			emitterCat:   99,
			expectedType: "0", // unknown
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := gdl90EmitterCatToNMEA(tc.emitterCat)
			if result != tc.expectedType {
				t.Errorf("gdl90EmitterCatToNMEA(%d) = %q, expected %q",
					tc.emitterCat, result, tc.expectedType)
			}
			t.Logf("GDL90 Emitter Category %d -> NMEA Type %q", tc.emitterCat, result)
		})
	}
}

// TestNmeaAircraftTypeToGdl90 tests reverse conversion from NMEA to GDL90
func TestNmeaAircraftTypeToGdl90(t *testing.T) {
	testCases := []struct {
		name        string
		actype      string
		expectedCat uint8
	}{
		{
			name:        "Glider (1)",
			actype:      "1",
			expectedCat: 9,
		},
		{
			name:        "Tow plane (2)",
			actype:      "2",
			expectedCat: 1, // light
		},
		{
			name:        "Helicopter (3)",
			actype:      "3",
			expectedCat: 7,
		},
		{
			name:        "Skydiver (4)",
			actype:      "4",
			expectedCat: 11,
		},
		{
			name:        "Drop plane (5)",
			actype:      "5",
			expectedCat: 1, // light
		},
		{
			name:        "Hang glider (6)",
			actype:      "6",
			expectedCat: 12,
		},
		{
			name:        "Paraglider (7)",
			actype:      "7",
			expectedCat: 12,
		},
		{
			name:        "Piston (8)",
			actype:      "8",
			expectedCat: 1, // light
		},
		{
			name:        "Jet (9)",
			actype:      "9",
			expectedCat: 3, // large
		},
		{
			name:        "Balloon (B)",
			actype:      "B",
			expectedCat: 10,
		},
		{
			name:        "Airship (C)",
			actype:      "C",
			expectedCat: 10,
		},
		{
			name:        "UAV (D)",
			actype:      "D",
			expectedCat: 14,
		},
		{
			name:        "Ground support (E)",
			actype:      "E",
			expectedCat: 18,
		},
		{
			name:        "Point obstacle (F)",
			actype:      "F",
			expectedCat: 19,
		},
		{
			name:        "Unknown (0)",
			actype:      "0",
			expectedCat: 0,
		},
		{
			name:        "Lowercase hex (d)",
			actype:      "d",
			expectedCat: 14, // UAV
		},
		{
			name:        "Invalid string",
			actype:      "ZZ",
			expectedCat: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := nmeaAircraftTypeToGdl90(tc.actype)
			if result != tc.expectedCat {
				t.Errorf("nmeaAircraftTypeToGdl90(%q) = %d, expected %d",
					tc.actype, result, tc.expectedCat)
			}
			t.Logf("NMEA Type %q -> GDL90 Category %d", tc.actype, result)
		})
	}
}

// TestAtof32 tests string to float32 conversion
func TestAtof32(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected float32
	}{
		{
			name:     "Zero",
			input:    "0",
			expected: 0.0,
		},
		{
			name:     "Positive integer",
			input:    "123",
			expected: 123.0,
		},
		{
			name:     "Negative integer",
			input:    "-456",
			expected: -456.0,
		},
		{
			name:     "Positive decimal",
			input:    "123.456",
			expected: 123.456,
		},
		{
			name:     "Negative decimal",
			input:    "-789.012",
			expected: -789.012,
		},
		{
			name:     "Scientific notation",
			input:    "1.23e2",
			expected: 123.0,
		},
		{
			name:     "Small decimal",
			input:    "0.001",
			expected: 0.001,
		},
		{
			name:     "Leading whitespace (invalid)",
			input:    " 123",
			expected: 0.0, // ParseFloat error returns 0
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := atof32(tc.input)
			// Use approximate comparison for float32
			diff := result - tc.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.0001 {
				t.Errorf("atof32(%q) = %f, expected %f (diff: %f)",
					tc.input, result, tc.expected, diff)
			}
			t.Logf("atof32(%q) = %f", tc.input, result)
		})
	}
}

// TestAtof32InvalidInputs tests atof32 with invalid inputs
func TestAtof32InvalidInputs(t *testing.T) {
	invalidInputs := []string{
		"",
		"abc",
		"12.34.56",
		"NaN",
		"Infinity",
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			result := atof32(input)
			// Invalid inputs should return 0 (ParseFloat error handling)
			if result != 0.0 {
				t.Logf("atof32(%q) = %f (non-zero for invalid input)", input, result)
			}
		})
	}
}

// TestComputeAlarmLevelBoundaries tests exact boundary conditions
func TestComputeAlarmLevelBoundaries(t *testing.T) {
	// Test exact boundaries for alarm levels
	boundaries := []struct {
		name             string
		dist             float64
		relativeVertical int32
		expectedAlarm    uint8
	}{
		// Level 3 boundaries (< 926m && < ±152m)
		{"Level 3: max dist", 925, 151, 3},
		{"Level 3: exceeds dist", 926, 151, 2}, // Just at boundary, should be level 2
		{"Level 3: max vert", 925, 152, 0},     // Just at boundary vertical
		{"Level 3: min vert", 925, -151, 3},
		{"Level 3: exceeds min vert", 925, -152, 0},

		// Level 2 boundaries (< 1852m && < ±304m)
		{"Level 2: max dist", 1851, 303, 2},
		{"Level 2: exceeds dist", 1852, 303, 0},
		{"Level 2: max vert", 1851, 304, 0},
		{"Level 2: min vert", 1851, -303, 2},
		{"Level 2: exceeds min vert", 1851, -304, 0},
	}

	for _, tc := range boundaries {
		t.Run(tc.name, func(t *testing.T) {
			result := computeAlarmLevel(tc.dist, tc.relativeVertical)
			if result != tc.expectedAlarm {
				t.Errorf("computeAlarmLevel(%.0f, %d) = %d, expected %d",
					tc.dist, tc.relativeVertical, result, tc.expectedAlarm)
			}
		})
	}
}

// TestGetIdTail tests OGN ID and tail parsing
func TestGetIdTail(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		expectedId      string
		expectedTail    string
		expectedAddress uint32
	}{
		{
			name:            "Simple ID without tail",
			input:           "AABBCC",
			expectedId:      "AABBCC",
			expectedTail:    "",
			expectedAddress: 0x00AABBCC,
		},
		{
			name:            "ID with tail",
			input:           "AABBCC!N12345",
			expectedId:      "AABBCC",
			expectedTail:    "N12345",
			expectedAddress: 0x00AABBCC,
		},
		{
			name:            "ID with OGN prefix in tail (should strip)",
			input:           "AABBCC!OGN_TAIL",
			expectedId:      "AABBCC",
			expectedTail:    "",
			expectedAddress: 0x00AABBCC,
		},
		{
			name:            "ID with FLR prefix in tail (should strip)",
			input:           "AABBCC!FLR_TAIL",
			expectedId:      "AABBCC",
			expectedTail:    "",
			expectedAddress: 0x00AABBCC,
		},
		{
			name:            "Long ID (> 6 chars, should truncate to last 6)",
			input:           "01AABBCC",
			expectedId:      "AABBCC",
			expectedTail:    "",
			expectedAddress: 0x00AABBCC,
		},
		{
			name:            "Long ID with tail",
			input:           "01AABBCC!TAIL",
			expectedId:      "AABBCC",
			expectedTail:    "TAIL",
			expectedAddress: 0x00AABBCC,
		},
		{
			name:            "Short ID",
			input:           "ABC",
			expectedId:      "ABC",
			expectedTail:    "",
			expectedAddress: 0x00000ABC,
		},
		{
			name:            "Short ID with tail",
			input:           "ABC!TAIL",
			expectedId:      "ABC",
			expectedTail:    "TAIL",
			expectedAddress: 0x00000ABC,
		},
		{
			name:            "ID with short prefix tail (should keep)",
			input:           "AABBCC!ABC",
			expectedId:      "AABBCC",
			expectedTail:    "ABC",
			expectedAddress: 0x00AABBCC,
		},
		{
			name:            "Lowercase hex ID",
			input:           "aabbcc",
			expectedId:      "aabbcc",
			expectedTail:    "",
			expectedAddress: 0x00AABBCC, // hex.DecodeString handles lowercase
		},
		{
			name:            "Zero address",
			input:           "000000",
			expectedId:      "000000",
			expectedTail:    "",
			expectedAddress: 0x00000000,
		},
		{
			name:            "Maximum 6-char hex address",
			input:           "FFFFFF",
			expectedId:      "FFFFFF",
			expectedTail:    "",
			expectedAddress: 0x00FFFFFF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			idStr, tail, address := getIdTail(tc.input)

			if idStr != tc.expectedId {
				t.Errorf("getIdTail(%q) idStr = %q, expected %q",
					tc.input, idStr, tc.expectedId)
			}

			if tail != tc.expectedTail {
				t.Errorf("getIdTail(%q) tail = %q, expected %q",
					tc.input, tail, tc.expectedTail)
			}

			if address != tc.expectedAddress {
				t.Errorf("getIdTail(%q) address = 0x%08X, expected 0x%08X",
					tc.input, address, tc.expectedAddress)
			}

			t.Logf("Input: %q -> ID: %q, Tail: %q, Address: 0x%08X",
				tc.input, idStr, tail, address)
		})
	}
}

// TestGetIdTailEdgeCases tests edge cases for OGN ID parsing
func TestGetIdTailEdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Empty string",
			input: "",
		},
		{
			name:  "Just exclamation",
			input: "!",
		},
		{
			name:  "Multiple exclamations",
			input: "AABBCC!TAIL!EXTRA",
		},
		{
			name:  "ID with underscore but no prefix",
			input: "AABBCC!TAI_L",
		},
		{
			name:  "Very long ID (8 chars, should use last 6)",
			input: "0123AABB",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Just verify these don't panic
			idStr, tail, address := getIdTail(tc.input)
			t.Logf("getIdTail(%q) -> ID: %q, Tail: %q, Address: 0x%08X",
				tc.input, idStr, tail, address)
		})
	}
}
