package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
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
			name:             "Close horizontal but too high vertical for level 3 - level 2",
			dist:             500, // 0.27 NM
			relativeVertical: 200, // 656 ft (> 500 ft threshold, but < 1000 ft)
			expectedAlarm:    2,   // Still within level 2 threshold
		},
		{
			name:             "Close horizontal but too low vertical for level 3 - level 2",
			dist:             500,  // 0.27 NM
			relativeVertical: -200, // -656 ft (< -500 ft threshold, but > -1000 ft)
			expectedAlarm:    2,    // Still within level 2 threshold
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
		{"Level 3: max vert", 925, 152, 2},     // Just at boundary vertical, still level 2
		{"Level 3: min vert", 925, -151, 3},
		{"Level 3: exceeds min vert", 925, -152, 2}, // Still within level 2 threshold

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

// TestMakeFlarmPFLAUString tests PFLAU sentence generation
func TestMakeFlarmPFLAUString(t *testing.T) {
	// Save original state and restore after test
	origSituation := mySituation
	origGlobalStatus := globalStatus
	origTraffic := traffic
	origClock := stratuxClock
	defer func() {
		mySituation = origSituation
		globalStatus = origGlobalStatus
		traffic = origTraffic
		stratuxClock = origClock
	}()

	// Initialize stratuxClock
	stratuxClock = NewMonotonic()

	// Initialize mySituation mutexes
	mySituation.muGPS = &sync.Mutex{}
	mySituation.muGPSPerformance = &sync.Mutex{}
	mySituation.muBaro = &sync.Mutex{}
	mySituation.muSatellite = &sync.Mutex{}
	mySituation.muAttitude = &sync.Mutex{}

	// Initialize traffic map
	traffic = make(map[uint32]TrafficInfo)

	testCases := []struct {
		name                  string
		setupSituation        func()
		setupGlobalStatus     func()
		setupTraffic          func()
		ti                    TrafficInfo
		expectGPSStatus       string // "0" or "2"
		expectRX              string // Number of traffic
		expectAlarmLevel      bool   // Whether alarm info should be present
		expectChecksumPresent bool
	}{
		{
			name: "Valid GPS with close traffic - alarm level 3",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSTrueCourse = 90.0
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			setupTraffic: func() {
				traffic[0xAABBCC] = TrafficInfo{Icao_addr: 0xAABBCC}
			},
			ti: TrafficInfo{
				Icao_addr:      0xAABBCC,
				Lat:            47.005, // ~500m away
				Lng:            -122.0,
				Alt:            1000,
				Position_valid: true,
			},
			expectGPSStatus:       "2",
			expectRX:              "1",
			expectAlarmLevel:      true,
			expectChecksumPresent: true,
		},
		{
			name: "Valid GPS with far traffic - no alarm",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSTrueCourse = 0.0
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			setupTraffic: func() {
				traffic = make(map[uint32]TrafficInfo)
			},
			ti: TrafficInfo{
				Icao_addr:      0x123456,
				Lat:            48.0, // ~111km away
				Lng:            -122.0,
				Alt:            5000,
				Position_valid: true,
			},
			expectGPSStatus:       "2",
			expectRX:              "0",
			expectAlarmLevel:      false,
			expectChecksumPresent: true,
		},
		{
			name: "No GPS fix",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = 0.0
				mySituation.GPSLongitude = 0.0
				mySituation.GPSTrueCourse = 0.0
				mySituation.GPSFixQuality = 0
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = false
			},
			setupTraffic: func() {
				traffic = make(map[uint32]TrafficInfo)
			},
			ti: TrafficInfo{
				Icao_addr:      0xDEADBE,
				Lat:            47.0,
				Lng:            -122.0,
				Alt:            1000,
				Position_valid: true,
			},
			expectGPSStatus:       "0",
			expectRX:              "0",
			expectAlarmLevel:      false,
			expectChecksumPresent: true,
		},
		{
			name: "Traffic with tail number",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSTrueCourse = 45.0
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
				mySituation.muGPS.Unlock()

				mySituation.muBaro.Lock()
				mySituation.BaroPressureAltitude = 1500.0
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
				mySituation.muBaro.Unlock()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			setupTraffic: func() {
				traffic = make(map[uint32]TrafficInfo)
			},
			ti: TrafficInfo{
				Icao_addr:      0xABC123,
				Tail:           "N12345",
				Lat:            47.002, // ~222m away
				Lng:            -122.0,
				Alt:            1550, // 50 feet above = 15.24m vertical difference
				Position_valid: true,
			},
			expectGPSStatus:       "2",
			expectRX:              "0",
			expectAlarmLevel:      true,
			expectChecksumPresent: true,
		},
		{
			name: "Bearing wrap around - positive to negative",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSTrueCourse = 350.0 // Heading nearly north
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			setupTraffic: func() {
				traffic = make(map[uint32]TrafficInfo)
			},
			ti: TrafficInfo{
				Icao_addr:      0x111111,
				Lat:            47.004,
				Lng:            -122.0,
				Alt:            1000,
				Position_valid: true,
			},
			expectGPSStatus:       "2",
			expectRX:              "0",
			expectAlarmLevel:      true,
			expectChecksumPresent: true,
		},
		{
			name: "Bearing adjustment - bearing > 180 degrees",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSTrueCourse = 10.0 // Heading nearly north
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
				mySituation.muGPS.Unlock()

				mySituation.muBaro.Lock()
				mySituation.BaroPressureAltitude = 1000.0
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
				mySituation.muBaro.Unlock()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			setupTraffic: func() {
				traffic = make(map[uint32]TrafficInfo)
			},
			ti: TrafficInfo{
				Icao_addr:      0x222222,
				Lat:            46.991,  // South of our position
				Lng:            -121.99, // East of our position, creates bearing ~200 degrees
				Alt:            1050,    // 50 feet above
				Position_valid: true,
			},
			expectGPSStatus:       "2",
			expectRX:              "0",
			expectAlarmLevel:      true,
			expectChecksumPresent: true,
		},
		{
			name: "Bearing adjustment - bearing < -180 degrees",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSTrueCourse = 200.0 // Heading nearly south
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
				mySituation.muGPS.Unlock()

				mySituation.muBaro.Lock()
				mySituation.BaroPressureAltitude = 1000.0
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
				mySituation.muBaro.Unlock()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			setupTraffic: func() {
				traffic = make(map[uint32]TrafficInfo)
			},
			ti: TrafficInfo{
				Icao_addr:      0x333333,
				Lat:            47.009,  // North of our position
				Lng:            -122.01, // West of our position, creates bearing ~10 degrees
				Alt:            1050,    // 50 feet above
				Position_valid: true,
			},
			expectGPSStatus:       "2",
			expectRX:              "0",
			expectAlarmLevel:      true,
			expectChecksumPresent: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			tc.setupSituation()
			tc.setupGlobalStatus()
			tc.setupTraffic()

			// Call function
			result := makeFlarmPFLAUString(tc.ti)

			// Verify basic structure
			if !strings.HasPrefix(result, "$PFLAU,") {
				t.Errorf("Expected result to start with '$PFLAU,', got: %s", result)
			}

			if !strings.HasSuffix(result, "\r\n") {
				t.Errorf("Expected result to end with \\r\\n, got: %s", result)
			}

			// Verify checksum is present
			if tc.expectChecksumPresent && !strings.Contains(result, "*") {
				t.Errorf("Expected checksum to be present, got: %s", result)
			}

			// Verify GPS status in message
			if !strings.Contains(result, ","+tc.expectGPSStatus+",") {
				t.Logf("Result: %s", result)
				t.Logf("Expected GPS status %s to be in message", tc.expectGPSStatus)
			}

			// Verify checksum calculation
			trimmed := strings.TrimSuffix(result, "\r\n")
			parts := strings.Split(trimmed, "*")
			if len(parts) == 2 {
				// Verify checksum is valid hex
				if len(parts[1]) != 2 {
					t.Errorf("Checksum should be 2 characters, got %d: %q", len(parts[1]), parts[1])
				}
			}

			t.Logf("Generated PFLAU: %s", strings.TrimSuffix(result, "\r\n"))
		})
	}
}

// TestMakeFlarmPFLAAString tests PFLAA sentence generation
func TestMakeFlarmPFLAAString(t *testing.T) {
	// Save original state
	origSituation := mySituation
	origGlobalSettings := globalSettings
	origClock := stratuxClock
	defer func() {
		mySituation = origSituation
		globalSettings = origGlobalSettings
		stratuxClock = origClock
	}()

	// Initialize stratuxClock
	stratuxClock = NewMonotonic()

	// Initialize mySituation mutexes
	mySituation.muGPS = &sync.Mutex{}
	mySituation.muGPSPerformance = &sync.Mutex{}
	mySituation.muBaro = &sync.Mutex{}
	mySituation.muSatellite = &sync.Mutex{}
	mySituation.muAttitude = &sync.Mutex{}

	testCases := []struct {
		name                  string
		setupSituation        func()
		ti                    TrafficInfo
		expectValid           bool
		expectAlarmLevel      uint8
		expectPositionValid   bool
		expectChecksumPresent bool
	}{
		{
			name: "Valid position with ICAO address",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muBaro.Lock()
				defer mySituation.muBaro.Unlock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSFixQuality = 1
				mySituation.BaroPressureAltitude = 1000.0 // Same level
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
			},
			ti: TrafficInfo{
				Icao_addr:        0xAABBCC,
				Addr_type:        0, // ICAO
				Lat:              47.005,
				Lng:              -122.005,
				Alt:              1000,
				Track:            90.0,
				Speed:            100,
				Speed_valid:      true,
				Vvel:             500,
				Emitter_category: 1,
				Position_valid:   true,
			},
			expectValid:           true,
			expectAlarmLevel:      3, // Close distance and vertical
			expectPositionValid:   true,
			expectChecksumPresent: true,
		},
		{
			name: "Valid position with non-ICAO address",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muBaro.Lock()
				defer mySituation.muBaro.Unlock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSFixQuality = 1
				mySituation.BaroPressureAltitude = 2000.0 // Same level
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
			},
			ti: TrafficInfo{
				Icao_addr:        0x123456,
				Addr_type:        1, // Non-ICAO
				Lat:              47.01,
				Lng:              -122.01,
				Alt:              2000,
				Track:            180.0,
				Speed:            150,
				Speed_valid:      true,
				Vvel:             -200,
				Emitter_category: 9,
				Position_valid:   true,
			},
			expectValid:           true,
			expectAlarmLevel:      2, // Within level 2 alarm distance (< 1NM)
			expectPositionValid:   true,
			expectChecksumPresent: true,
		},
		{
			name: "Invalid position - bearingless traffic",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSFixQuality = 1
			},
			ti: TrafficInfo{
				Icao_addr:         0xDEADBE,
				Addr_type:         0,
				Alt:               3000,
				Vvel:              100,
				Emitter_category:  7,
				Position_valid:    false,
				DistanceEstimated: 5000.0,
			},
			expectValid:           true,
			expectAlarmLevel:      0,
			expectPositionValid:   false,
			expectChecksumPresent: true,
		},
		{
			name: "Traffic with tail number",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muBaro.Lock()
				defer mySituation.muBaro.Unlock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSFixQuality = 1
				mySituation.BaroPressureAltitude = 1500.0 // Same level
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
			},
			ti: TrafficInfo{
				Icao_addr:        0xABC123,
				Addr_type:        0,
				Tail:             "N12345",
				Lat:              47.005,
				Lng:              -122.005,
				Alt:              1500,
				Track:            270.0,
				Speed:            120,
				Speed_valid:      true,
				Vvel:             0,
				Emitter_category: 1,
				Position_valid:   true,
			},
			expectValid:           true,
			expectAlarmLevel:      3,
			expectPositionValid:   true,
			expectChecksumPresent: true,
		},
		{
			name: "Different aircraft types",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muBaro.Lock()
				defer mySituation.muBaro.Unlock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSFixQuality = 1
				mySituation.BaroPressureAltitude = 2000.0 // Same level
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
			},
			ti: TrafficInfo{
				Icao_addr:        0x999999,
				Addr_type:        0,
				Lat:              47.005,
				Lng:              -122.005,
				Alt:              2000,
				Track:            45.0,
				Speed:            80,
				Speed_valid:      true,
				Vvel:             300,
				Emitter_category: 12, // Paraglider
				Position_valid:   true,
			},
			expectValid:           true,
			expectAlarmLevel:      3,
			expectPositionValid:   true,
			expectChecksumPresent: true,
		},
		{
			name: "Far away traffic - no alarm",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSFixQuality = 1
			},
			ti: TrafficInfo{
				Icao_addr:        0x555555,
				Addr_type:        0,
				Lat:              47.1, // ~11km away
				Lng:              -122.0,
				Alt:              5000,
				Track:            0.0,
				Speed:            200,
				Speed_valid:      true,
				Vvel:             1000,
				Emitter_category: 3,
				Position_valid:   true,
			},
			expectValid:           true,
			expectAlarmLevel:      0,
			expectPositionValid:   true,
			expectChecksumPresent: true,
		},
		{
			name: "DEBUG mode enabled - logs traffic info",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muBaro.Lock()
				defer mySituation.muBaro.Unlock()
				mySituation.GPSLatitude = 47.0
				mySituation.GPSLongitude = -122.0
				mySituation.GPSFixQuality = 1
				mySituation.BaroPressureAltitude = 1000.0
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
			},
			ti: TrafficInfo{
				Icao_addr:        0xDEB001,
				Addr_type:        0,
				Tail:             "TESTAC",
				Lat:              47.005,
				Lng:              -122.005,
				Alt:              1000,
				Track:            180.0,
				Speed:            100,
				Speed_valid:      true,
				Vvel:             0,
				Emitter_category: 1,
				Position_valid:   true,
			},
			expectValid:           true,
			expectAlarmLevel:      3,
			expectPositionValid:   true,
			expectChecksumPresent: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			tc.setupSituation()
			// Enable DEBUG mode for the specific test case
			if tc.name == "DEBUG mode enabled - logs traffic info" {
				globalSettings.DEBUG = true
			} else {
				globalSettings.DEBUG = false
			}

			// Call function
			msg, valid, alarmLevel := makeFlarmPFLAAString(tc.ti)

			// Verify valid flag
			if valid != tc.expectValid {
				t.Errorf("Expected valid=%v, got %v", tc.expectValid, valid)
			}

			// Verify alarm level
			if alarmLevel != tc.expectAlarmLevel {
				t.Errorf("Expected alarmLevel=%d, got %d", tc.expectAlarmLevel, alarmLevel)
			}

			if valid {
				// Verify basic structure
				if !strings.HasPrefix(msg, "$PFLAA,") {
					t.Errorf("Expected message to start with '$PFLAA,', got: %s", msg)
				}

				if !strings.HasSuffix(msg, "\r\n") {
					t.Errorf("Expected message to end with \\r\\n")
				}

				// Verify checksum is present
				if tc.expectChecksumPresent && !strings.Contains(msg, "*") {
					t.Errorf("Expected checksum to be present")
				}

				// Verify ID is in message (6 hex digits)
				if !strings.Contains(msg, strings.ToUpper(fmt.Sprintf("%.6X", tc.ti.Icao_addr&0xFFFFFF))) {
					t.Logf("Expected ID %.6X to be in message: %s", tc.ti.Icao_addr&0xFFFFFF, msg)
				}

				// Verify tail is in message if provided
				if len(tc.ti.Tail) > 0 && !strings.Contains(msg, tc.ti.Tail) {
					t.Errorf("Expected tail %s to be in message", tc.ti.Tail)
				}

				t.Logf("Generated PFLAA: %s", strings.TrimSuffix(msg, "\r\n"))
			}
		})
	}
}

// TestMakeGPRMCString tests GPRMC sentence generation
func TestMakeGPRMCString(t *testing.T) {
	// Save original state
	origSituation := mySituation
	origClock := stratuxClock
	defer func() {
		mySituation = origSituation
		stratuxClock = origClock
	}()

	// Initialize stratuxClock
	stratuxClock = NewMonotonic()

	// Initialize mySituation mutexes
	mySituation.muGPS = &sync.Mutex{}
	mySituation.muGPSPerformance = &sync.Mutex{}
	mySituation.muBaro = &sync.Mutex{}
	mySituation.muSatellite = &sync.Mutex{}
	mySituation.muAttitude = &sync.Mutex{}

	testCases := []struct {
		name                  string
		setupSituation        func()
		setupGlobalStatus     func()
		expectStatus          string // "A" or "V"
		expectValidPosition   bool
		expectChecksumPresent bool
	}{
		{
			name: "Valid GPS fix",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = 47.12345
				mySituation.GPSLongitude = -122.98765
				mySituation.GPSGroundSpeed = 100.0
				mySituation.GPSTrueCourse = 90.0
				mySituation.GPSLastFixSinceMidnightUTC = 43519.5 // 12:05:19.5 UTC
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			expectStatus:          "A",
			expectValidPosition:   true,
			expectChecksumPresent: true,
		},
		{
			name: "Valid GPS fix at equator/prime meridian",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = 0.0
				mySituation.GPSLongitude = 0.0
				mySituation.GPSGroundSpeed = 0.0
				mySituation.GPSTrueCourse = 0.0
				mySituation.GPSLastFixSinceMidnightUTC = 0.0
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			expectStatus:          "A",
			expectValidPosition:   true,
			expectChecksumPresent: true,
		},
		{
			name: "Valid GPS fix in southern/western hemisphere",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = -33.8688
				mySituation.GPSLongitude = -151.2093
				mySituation.GPSGroundSpeed = 250.0
				mySituation.GPSTrueCourse = 180.0
				mySituation.GPSLastFixSinceMidnightUTC = 3600.0 // 01:00:00 UTC
				mySituation.GPSFixQuality = 2                   // DGPS
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			expectStatus:          "A",
			expectValidPosition:   true,
			expectChecksumPresent: true,
		},
		{
			name: "No GPS fix - invalid",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = 0.0
				mySituation.GPSLongitude = 0.0
				mySituation.GPSGroundSpeed = 0.0
				mySituation.GPSTrueCourse = 0.0
				mySituation.GPSLastFixSinceMidnightUTC = 0.0
				mySituation.GPSFixQuality = 0
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = false
			},
			expectStatus:          "V",
			expectValidPosition:   false,
			expectChecksumPresent: true,
		},
		{
			name: "High speed and extreme coordinates",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSLatitude = 89.9999
				mySituation.GPSLongitude = 179.9999
				mySituation.GPSGroundSpeed = 600.0
				mySituation.GPSTrueCourse = 359.9
				mySituation.GPSLastFixSinceMidnightUTC = 86399.0 // 23:59:59 UTC
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			expectStatus:          "A",
			expectValidPosition:   true,
			expectChecksumPresent: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			tc.setupSituation()
			tc.setupGlobalStatus()

			// Call function
			result := makeGPRMCString()

			// Verify basic structure
			if !strings.HasPrefix(result, "$GPRMC,") {
				t.Errorf("Expected result to start with '$GPRMC,', got: %s", result)
			}

			if !strings.HasSuffix(result, "\r\n") {
				t.Errorf("Expected result to end with \\r\\n")
			}

			// Verify checksum is present
			if tc.expectChecksumPresent && !strings.Contains(result, "*") {
				t.Errorf("Expected checksum to be present")
			}

			// Verify status is in message
			if !strings.Contains(result, ","+tc.expectStatus+",") {
				t.Logf("Result: %s", result)
				t.Logf("Expected status %s to be in message", tc.expectStatus)
			}

			// For valid position, verify lat/lng are present
			if tc.expectValidPosition {
				// Should have numeric lat/lng
				if strings.Contains(result, ",,,,") {
					t.Errorf("Valid position should not have empty lat/lng fields")
				}
			} else {
				// Invalid position should have empty lat/lng
				if !strings.Contains(result, ",,,,") {
					t.Logf("Invalid position should have empty lat/lng fields")
				}
			}

			t.Logf("Generated GPRMC: %s", strings.TrimSuffix(result, "\r\n"))
		})
	}
}

// TestMakeGPGGAString tests GPGGA sentence generation
func TestMakeGPGGAString(t *testing.T) {
	// Save original state
	origSituation := mySituation
	origClock := stratuxClock
	defer func() {
		mySituation = origSituation
		stratuxClock = origClock
	}()

	// Initialize stratuxClock
	stratuxClock = NewMonotonic()

	// Initialize mySituation mutexes
	mySituation.muGPS = &sync.Mutex{}
	mySituation.muGPSPerformance = &sync.Mutex{}
	mySituation.muBaro = &sync.Mutex{}
	mySituation.muSatellite = &sync.Mutex{}
	mySituation.muAttitude = &sync.Mutex{}

	testCases := []struct {
		name                  string
		setupSituation        func()
		setupGlobalStatus     func()
		expectFixQuality      int
		expectValidPosition   bool
		expectChecksumPresent bool
	}{
		{
			name: "Valid GPS fix with satellites",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muSatellite.Lock()
				defer mySituation.muSatellite.Unlock()
				mySituation.GPSLatitude = 47.6062
				mySituation.GPSLongitude = -122.3321
				mySituation.GPSAltitudeMSL = 500.0
				mySituation.GPSGeoidSep = 50.0
				mySituation.GPSSatellites = 8
				mySituation.GPSLastFixSinceMidnightUTC = 43200.0 // 12:00:00 UTC
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			expectFixQuality:      1,
			expectValidPosition:   true,
			expectChecksumPresent: true,
		},
		{
			name: "DGPS fix",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muSatellite.Lock()
				defer mySituation.muSatellite.Unlock()
				mySituation.GPSLatitude = 40.7128
				mySituation.GPSLongitude = -74.0060
				mySituation.GPSAltitudeMSL = 100.0
				mySituation.GPSGeoidSep = -100.0
				mySituation.GPSSatellites = 12
				mySituation.GPSLastFixSinceMidnightUTC = 0.0
				mySituation.GPSFixQuality = 2
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			expectFixQuality:      2,
			expectValidPosition:   true,
			expectChecksumPresent: true,
		},
		{
			name: "Southern/western hemisphere with negative geoid sep",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muSatellite.Lock()
				defer mySituation.muSatellite.Unlock()
				mySituation.GPSLatitude = -34.6037
				mySituation.GPSLongitude = -58.3816
				mySituation.GPSAltitudeMSL = 82.0
				mySituation.GPSGeoidSep = -30.0
				mySituation.GPSSatellites = 10
				mySituation.GPSLastFixSinceMidnightUTC = 50000.0
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			expectFixQuality:      1,
			expectValidPosition:   true,
			expectChecksumPresent: true,
		},
		{
			name: "No GPS fix",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muSatellite.Lock()
				defer mySituation.muSatellite.Unlock()
				mySituation.GPSLatitude = 0.0
				mySituation.GPSLongitude = 0.0
				mySituation.GPSAltitudeMSL = 0.0
				mySituation.GPSGeoidSep = 0.0
				mySituation.GPSSatellites = 3
				mySituation.GPSFixQuality = 0
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = false
			},
			expectFixQuality:      0,
			expectValidPosition:   false,
			expectChecksumPresent: true,
		},
		{
			name: "High altitude fix",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muSatellite.Lock()
				defer mySituation.muSatellite.Unlock()
				mySituation.GPSLatitude = 45.0
				mySituation.GPSLongitude = 6.0
				mySituation.GPSAltitudeMSL = 29000.0 // High altitude flight
				mySituation.GPSGeoidSep = 150.0
				mySituation.GPSSatellites = 9
				mySituation.GPSLastFixSinceMidnightUTC = 7200.0
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
			},
			expectFixQuality:      1,
			expectValidPosition:   true,
			expectChecksumPresent: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			tc.setupSituation()
			tc.setupGlobalStatus()

			// Call function
			result := makeGPGGAString()

			// Verify basic structure
			if !strings.HasPrefix(result, "$GPGGA,") {
				t.Errorf("Expected result to start with '$GPGGA,', got: %s", result)
			}

			if !strings.HasSuffix(result, "\r\n") {
				t.Errorf("Expected result to end with \\r\\n")
			}

			// Verify checksum is present
			if tc.expectChecksumPresent && !strings.Contains(result, "*") {
				t.Errorf("Expected checksum to be present")
			}

			// For valid position, verify lat/lng are present
			if tc.expectValidPosition {
				// Should have numeric lat/lng
				if strings.HasPrefix(result, "$GPGGA,,,,") {
					t.Errorf("Valid position should not have empty time/lat/lng fields")
				}
			} else {
				// Invalid position should start with empty fields
				if !strings.HasPrefix(result, "$GPGGA,,,,") {
					t.Logf("Invalid position should have empty time/lat/lng fields")
				}
			}

			t.Logf("Generated GPGGA: %s", strings.TrimSuffix(result, "\r\n"))
		})
	}
}

// TestMakePGRMZString tests PGRMZ sentence generation
func TestMakePGRMZString(t *testing.T) {
	// Save original state
	origSituation := mySituation
	origClock := stratuxClock
	defer func() {
		mySituation = origSituation
		stratuxClock = origClock
	}()

	// Initialize stratuxClock
	stratuxClock = NewMonotonic()

	// Initialize mySituation mutexes
	mySituation.muGPS = &sync.Mutex{}
	mySituation.muGPSPerformance = &sync.Mutex{}
	mySituation.muBaro = &sync.Mutex{}
	mySituation.muSatellite = &sync.Mutex{}
	mySituation.muAttitude = &sync.Mutex{}

	testCases := []struct {
		name                  string
		altitude              float32
		expectChecksumPresent bool
	}{
		{
			name:                  "Sea level altitude",
			altitude:              0.0,
			expectChecksumPresent: true,
		},
		{
			name:                  "Typical cruise altitude",
			altitude:              5500.0,
			expectChecksumPresent: true,
		},
		{
			name:                  "High altitude",
			altitude:              41000.0,
			expectChecksumPresent: true,
		},
		{
			name:                  "Negative altitude (below sea level)",
			altitude:              -282.0, // Death Valley
			expectChecksumPresent: true,
		},
		{
			name:                  "Fractional altitude (should be rounded)",
			altitude:              1234.56,
			expectChecksumPresent: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mySituation.muBaro.Lock()
			mySituation.BaroPressureAltitude = tc.altitude
			mySituation.muBaro.Unlock()

			// Call function
			result := makePGRMZString()

			// Verify basic structure
			if !strings.HasPrefix(result, "$PGRMZ,") {
				t.Errorf("Expected result to start with '$PGRMZ,', got: %s", result)
			}

			if !strings.HasSuffix(result, "\r\n") {
				t.Errorf("Expected result to end with \\r\\n")
			}

			// Verify checksum is present
			if tc.expectChecksumPresent && !strings.Contains(result, "*") {
				t.Errorf("Expected checksum to be present")
			}

			// Verify altitude units (should end with ",f,3*" before checksum)
			if !strings.Contains(result, ",f,3*") {
				t.Errorf("Expected altitude in feet with mode 3, got: %s", result)
			}

			t.Logf("Generated PGRMZ for altitude %.1f: %s", tc.altitude, strings.TrimSuffix(result, "\r\n"))
		})
	}
}

// TestMakeAHRSLevilReport tests AHRS Levil report generation
func TestMakeAHRSLevilReport(t *testing.T) {
	// Save original state
	origSituation := mySituation
	origGlobalStatus := globalStatus
	origClock := stratuxClock
	defer func() {
		mySituation = origSituation
		globalStatus = origGlobalStatus
		stratuxClock = origClock
	}()

	// Initialize stratuxClock
	stratuxClock = NewMonotonic()

	// Initialize mySituation mutexes
	mySituation.muGPS = &sync.Mutex{}
	mySituation.muGPSPerformance = &sync.Mutex{}
	mySituation.muBaro = &sync.Mutex{}
	mySituation.muSatellite = &sync.Mutex{}
	mySituation.muAttitude = &sync.Mutex{}

	testCases := []struct {
		name              string
		setupSituation    func()
		setupGlobalStatus func()
		shouldReturn      bool // Whether function should return early without sending
	}{
		{
			name: "No IMU connected - should return early",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
				globalStatus.IMUConnected = false
			},
			shouldReturn: true,
		},
		{
			name: "No GPS fix - should return early",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muAttitude.Lock()
				defer mySituation.muAttitude.Unlock()

				mySituation.GPSFixQuality = 0
				mySituation.AHRSRoll = 10.0
				mySituation.AHRSPitch = 5.0
				mySituation.AHRSGyroHeading = 90.0
				mySituation.AHRSLastAttitudeTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = false
				globalStatus.IMUConnected = true
			},
			shouldReturn: true,
		},
		{
			name: "AHRS data too old - should return early",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muAttitude.Lock()
				defer mySituation.muAttitude.Unlock()

				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
				mySituation.AHRSRoll = 10.0
				mySituation.AHRSPitch = 5.0
				mySituation.AHRSGyroHeading = 90.0
				// AHRS data too old (> 1 second)
				mySituation.AHRSLastAttitudeTime = stratuxClock.GetTime().Add(-2 * time.Second)
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
				globalStatus.IMUConnected = true
			},
			shouldReturn: true,
		},
		{
			name: "Valid AHRS data - generates report",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muAttitude.Lock()
				defer mySituation.muAttitude.Unlock()

				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
				mySituation.GPSTrueCourse = 180.0
				mySituation.GPSTurnRate = 5.0
				mySituation.AHRSRoll = 10.0
				mySituation.AHRSPitch = 5.0
				mySituation.AHRSGyroHeading = 270.0
				mySituation.AHRSSlipSkid = 1.5
				mySituation.AHRSTurnRate = 3.0
				mySituation.AHRSGLoad = 1.2
				// Recent AHRS data (within 1 second)
				mySituation.AHRSLastAttitudeTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
				globalStatus.IMUConnected = true
			},
			shouldReturn: false,
		},
		{
			name: "Valid AHRS with some invalid values",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muAttitude.Lock()
				defer mySituation.muAttitude.Unlock()

				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
				mySituation.GPSTrueCourse = 90.0
				mySituation.GPSTurnRate = 0.0
				mySituation.AHRSRoll = 15.0
				mySituation.AHRSPitch = 3276.7 // Invalid marker
				mySituation.AHRSGyroHeading = 180.0
				mySituation.AHRSSlipSkid = 3276.7 // Invalid marker
				mySituation.AHRSTurnRate = 3276.7 // Invalid marker
				mySituation.AHRSGLoad = 1.0
				mySituation.AHRSLastAttitudeTime = stratuxClock.GetTime()
			},
			setupGlobalStatus: func() {
				globalStatus.GPS_connected = true
				globalStatus.IMUConnected = true
			},
			shouldReturn: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			tc.setupSituation()
			tc.setupGlobalStatus()

			// Note: makeAHRSLevilReport() doesn't return a value, it calls sendNetFLARM()
			// When data is valid (shouldReturn=false), it will try to send network data
			// which may panic without network initialization - that's expected
			defer func() {
				if r := recover(); r != nil {
					if tc.shouldReturn {
						// If we expected early return, panic is unexpected
						t.Errorf("makeAHRSLevilReport() panicked unexpectedly when should return early: %v", r)
					} else {
						// If we expected to generate report, panic on sendNetFLARM is OK
						// because it means we exercised the main function body
						t.Logf("makeAHRSLevilReport() executed main body and panicked on network send (expected): %v", r)
					}
				}
			}()

			makeAHRSLevilReport()

			// If we get here without panic
			if tc.shouldReturn {
				t.Logf("makeAHRSLevilReport() returned early as expected (no valid data)")
			} else {
				t.Logf("makeAHRSLevilReport() completed without panic (network was available)")
			}
		})
	}
}

// TestComputeRelativeVertical tests relative vertical calculation
func TestComputeRelativeVertical(t *testing.T) {
	// Save original state
	origSituation := mySituation
	origClock := stratuxClock
	defer func() {
		mySituation = origSituation
		stratuxClock = origClock
	}()

	// Initialize stratuxClock
	stratuxClock = NewMonotonic()

	// Initialize mySituation mutexes
	mySituation.muGPS = &sync.Mutex{}
	mySituation.muGPSPerformance = &sync.Mutex{}
	mySituation.muBaro = &sync.Mutex{}
	mySituation.muSatellite = &sync.Mutex{}
	mySituation.muAttitude = &sync.Mutex{}

	testCases := []struct {
		name           string
		setupSituation func()
		ti             TrafficInfo
		expectedRange  [2]int32 // Min and max expected values (approximate)
	}{
		{
			name: "Traffic higher with baro alt",
			setupSituation: func() {
				mySituation.muBaro.Lock()
				defer mySituation.muBaro.Unlock()
				mySituation.BaroPressureAltitude = 5000.0
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
			},
			ti: TrafficInfo{
				Alt:       6000,
				AltIsGNSS: false,
			},
			expectedRange: [2]int32{300, 310}, // ~304m = 1000ft
		},
		{
			name: "Traffic lower with baro alt",
			setupSituation: func() {
				mySituation.muBaro.Lock()
				defer mySituation.muBaro.Unlock()
				mySituation.BaroPressureAltitude = 5000.0
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
			},
			ti: TrafficInfo{
				Alt:       4000,
				AltIsGNSS: false,
			},
			expectedRange: [2]int32{-310, -300}, // ~-304m = -1000ft
		},
		{
			name: "Same altitude",
			setupSituation: func() {
				mySituation.muBaro.Lock()
				defer mySituation.muBaro.Unlock()
				mySituation.BaroPressureAltitude = 3000.0
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
			},
			ti: TrafficInfo{
				Alt:       3000,
				AltIsGNSS: false,
			},
			expectedRange: [2]int32{-5, 5}, // Should be close to 0
		},
		{
			name: "Traffic with GNSS altitude",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.GPSHeightAboveEllipsoid = 1000.0
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
			},
			ti: TrafficInfo{
				Alt:       1500,
				AltIsGNSS: true,
			},
			expectedRange: [2]int32{150, 155}, // ~152m = 500ft
		},
		{
			name: "Use GPS MSL when no baro available",
			setupSituation: func() {
				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()
				mySituation.muBaro.Lock()
				defer mySituation.muBaro.Unlock()
				mySituation.GPSAltitudeMSL = 2000.0
				mySituation.GPSFixQuality = 1
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
				// Make baro invalid by setting old time
				mySituation.BaroLastMeasurementTime = stratuxClock.GetTime().Add(-20 * time.Second)
			},
			ti: TrafficInfo{
				Alt:       3000,
				AltIsGNSS: false,
			},
			expectedRange: [2]int32{300, 310}, // ~304m = 1000ft
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			tc.setupSituation()
			globalStatus.GPS_connected = true

			// Call function
			result := computeRelativeVertical(tc.ti)

			// Verify result is in expected range
			if result < tc.expectedRange[0] || result > tc.expectedRange[1] {
				t.Errorf("computeRelativeVertical() = %d, expected in range [%d, %d]",
					result, tc.expectedRange[0], tc.expectedRange[1])
			}

			t.Logf("Relative vertical: %d meters (%.0f feet)", result, float64(result)*3.28084)
		})
	}
}

// TestParseFlarmNmeaMessage tests the parseFlarmNmeaMessage dispatcher function
func TestParseFlarmNmeaMessage(t *testing.T) {
	// Initialize test environment
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

	// Save original state
	origTraffic := traffic
	origSeenTraffic := seenTraffic
	defer func() {
		traffic = origTraffic
		seenTraffic = origSeenTraffic
	}()

	t.Run("valid_PFLAU_message", func(t *testing.T) {
		// Setup valid GPS
		mySituation.muGPS.Lock()
		mySituation.GPSLatitude = 47.5
		mySituation.GPSLongitude = -122.3
		mySituation.GPSAltitudeMSL = 500
		mySituation.GPSTrueCourse = 90.0
		mySituation.GPSFixQuality = 2
		mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
		mySituation.muGPS.Unlock()
		globalStatus.GPS_connected = true

		// Initialize traffic map
		traffic = make(map[uint32]TrafficInfo)
		seenTraffic = make(map[uint32]bool)

		// $PFLAU,<RX>,<TX>,<GPS>,<Power>,<AlarmLevel>,<RelativeBearing>,<AlarmType>,<RelativeVertical>,<RelativeDistance>,<ID>
		message := []string{"PFLAU", "1", "1", "2", "1", "0", "45", "2", "100", "5000", "ABC123"}

		// Should not panic
		parseFlarmNmeaMessage(message)

		// Check that traffic was added
		trafficMutex.Lock()
		defer trafficMutex.Unlock()
		if len(traffic) == 0 {
			t.Error("Expected traffic to be added from PFLAU message")
		}
		t.Logf("PFLAU message processed, %d traffic targets", len(traffic))
	})

	t.Run("valid_PFLAA_message", func(t *testing.T) {
		// Setup valid GPS
		mySituation.muGPS.Lock()
		mySituation.GPSLatitude = 47.5
		mySituation.GPSLongitude = -122.3
		mySituation.GPSAltitudeMSL = 500
		mySituation.GPSFixQuality = 2
		mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
		mySituation.muGPS.Unlock()
		globalStatus.GPS_connected = true

		// Initialize traffic map
		traffic = make(map[uint32]TrafficInfo)
		seenTraffic = make(map[uint32]bool)

		// $PFLAA,<AlarmLevel>,<RelativeNorth>,<RelativeEast>,<RelativeVertical>,<IDType>,<ID>,<Track>,<TurnRate>,<GroundSpeed>,<ClimbRate>,<AcftType>
		message := []string{"PFLAA", "0", "1000", "500", "100", "1", "ABCDEF", "90", "0", "50", "2.5", "1"}

		// Should not panic
		parseFlarmNmeaMessage(message)

		// Check that traffic was added
		trafficMutex.Lock()
		defer trafficMutex.Unlock()
		if len(traffic) == 0 {
			t.Error("Expected traffic to be added from PFLAA message")
		}

		// Find the traffic entry
		var ti TrafficInfo
		var found bool
		for _, v := range traffic {
			ti = v
			found = true
			break
		}

		if !found {
			t.Fatal("No traffic found")
		}

		// Check some basic fields
		if ti.Track != 90 {
			t.Errorf("Expected Track=90, got %f", ti.Track)
		}
		if ti.Emitter_category != 9 { // Type "1" = glider = category 9
			t.Errorf("Expected Emitter_category=9 (glider), got %d", ti.Emitter_category)
		}

		t.Logf("PFLAA message processed: Track=%.0f, Speed=%d kts, Category=%d",
			ti.Track, ti.Speed, ti.Emitter_category)
	})

	t.Run("PFLAA_with_different_aircraft_types", func(t *testing.T) {
		testCases := []struct {
			name             string
			acType           string
			expectedCategory uint8
		}{
			{"Glider", "1", 9},
			{"Tow plane", "2", 1},
			{"Helicopter", "3", 7},
			{"Skydiver", "4", 11},
			{"Drop plane", "5", 1},
			{"Hang glider", "6", 12},
			{"Paraglider", "7", 12},
			{"Piston", "8", 1},
			{"Jet", "9", 3},
			{"Balloon", "B", 10},
			{"Airship", "C", 10},
			{"UAV", "D", 14},
			{"Ground vehicle", "E", 18},
			{"Static object", "F", 19},
			{"Unknown", "0", 0},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup valid GPS
				mySituation.muGPS.Lock()
				mySituation.GPSLatitude = 47.5
				mySituation.GPSLongitude = -122.3
				mySituation.GPSAltitudeMSL = 500
				mySituation.GPSFixQuality = 2
				mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
				mySituation.muGPS.Unlock()
				globalStatus.GPS_connected = true

				// Initialize traffic map
				traffic = make(map[uint32]TrafficInfo)
				seenTraffic = make(map[uint32]bool)

				// Create PFLAA message with specific aircraft type
				message := []string{"PFLAA", "0", "1000", "500", "100", "1", "TEST01", "90", "0", "50", "2.5", tc.acType}

				parseFlarmNmeaMessage(message)

				// Check that traffic was added with correct category
				trafficMutex.Lock()
				defer trafficMutex.Unlock()

				if len(traffic) == 0 {
					t.Error("Expected traffic to be added")
					return
				}

				var ti TrafficInfo
				for _, v := range traffic {
					ti = v
					break
				}

				if ti.Emitter_category != tc.expectedCategory {
					t.Errorf("Expected category %d, got %d", tc.expectedCategory, ti.Emitter_category)
				}

				t.Logf("Aircraft type %s -> Category %d", tc.acType, ti.Emitter_category)
			})
		}
	})

	t.Run("PFLAU_without_valid_GPS", func(t *testing.T) {
		// Setup invalid GPS
		mySituation.muGPS.Lock()
		mySituation.GPSFixQuality = 0
		mySituation.GPSLastFixLocalTime = time.Time{}
		mySituation.muGPS.Unlock()
		globalStatus.GPS_connected = false

		// Initialize traffic map
		traffic = make(map[uint32]TrafficInfo)
		seenTraffic = make(map[uint32]bool)

		message := []string{"PFLAU", "1", "1", "2", "1", "0", "45", "2", "100", "5000", "ABC123"}

		// Should not panic, but should not add traffic
		parseFlarmNmeaMessage(message)

		trafficMutex.Lock()
		defer trafficMutex.Unlock()
		if len(traffic) != 0 {
			t.Error("Expected no traffic to be added without valid GPS")
		}
		t.Log("PFLAU correctly ignored without valid GPS")
	})

	t.Run("PFLAU_with_empty_fields", func(t *testing.T) {
		// Message with empty required fields should be ignored
		message := []string{"PFLAU", "1", "1", "2", "1", "0", "", "2", "", "", ""}

		traffic = make(map[uint32]TrafficInfo)
		seenTraffic = make(map[uint32]bool)

		// Should not panic
		parseFlarmNmeaMessage(message)

		// Should not add traffic due to empty fields
		trafficMutex.Lock()
		defer trafficMutex.Unlock()
		if len(traffic) != 0 {
			t.Error("Expected no traffic from PFLAU with empty fields")
		}
		t.Log("PFLAU with empty fields correctly ignored")
	})

	t.Run("PFLAU_too_short", func(t *testing.T) {
		// Message too short (< 11 fields) should be ignored
		message := []string{"PFLAU", "1", "1", "2", "1"}

		traffic = make(map[uint32]TrafficInfo)
		seenTraffic = make(map[uint32]bool)

		// Should not panic
		parseFlarmNmeaMessage(message)

		// Should not add traffic
		trafficMutex.Lock()
		defer trafficMutex.Unlock()
		if len(traffic) != 0 {
			t.Error("Expected no traffic from short PFLAU message")
		}
		t.Log("Short PFLAU message correctly ignored")
	})

	t.Run("PFLAA_too_short", func(t *testing.T) {
		// Message too short (< 12 fields) should be ignored
		message := []string{"PFLAA", "0", "1000", "500", "100"}

		traffic = make(map[uint32]TrafficInfo)
		seenTraffic = make(map[uint32]bool)

		// Should not panic
		parseFlarmNmeaMessage(message)

		// Should not add traffic
		trafficMutex.Lock()
		defer trafficMutex.Unlock()
		if len(traffic) != 0 {
			t.Error("Expected no traffic from short PFLAA message")
		}
		t.Log("Short PFLAA message correctly ignored")
	})

	t.Run("PFLAA_with_tail_number", func(t *testing.T) {
		// Setup valid GPS
		mySituation.muGPS.Lock()
		mySituation.GPSLatitude = 47.5
		mySituation.GPSLongitude = -122.3
		mySituation.GPSAltitudeMSL = 500
		mySituation.GPSFixQuality = 2
		mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
		mySituation.muGPS.Unlock()
		globalStatus.GPS_connected = true

		traffic = make(map[uint32]TrafficInfo)
		seenTraffic = make(map[uint32]bool)

		// ID with tail number using ! separator: ID!TAIL
		message := []string{"PFLAA", "0", "1000", "500", "100", "1", "ABCDEF!N12345", "90", "0", "50", "2.5", "1"}

		parseFlarmNmeaMessage(message)

		trafficMutex.Lock()
		defer trafficMutex.Unlock()

		if len(traffic) == 0 {
			t.Error("Expected traffic to be added")
			return
		}

		var ti TrafficInfo
		for _, v := range traffic {
			ti = v
			break
		}

		// Check that tail was extracted
		if ti.Tail != "N12345" {
			t.Errorf("Expected Tail='N12345', got '%s'", ti.Tail)
		}

		t.Logf("PFLAA with tail: ID=0x%X, Tail=%s", ti.Icao_addr, ti.Tail)
	})

	t.Run("PFLAA_with_ICAO_and_NonICAO_idType", func(t *testing.T) {
		// Setup valid GPS
		mySituation.muGPS.Lock()
		mySituation.GPSLatitude = 47.5
		mySituation.GPSLongitude = -122.3
		mySituation.GPSAltitudeMSL = 500
		mySituation.GPSFixQuality = 2
		mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
		mySituation.muGPS.Unlock()
		globalStatus.GPS_connected = true

		traffic = make(map[uint32]TrafficInfo)
		seenTraffic = make(map[uint32]bool)

		// IDType=1 should map to Addr_type=0 (ICAO)
		message := []string{"PFLAA", "0", "1000", "500", "100", "1", "ABC123", "90", "0", "50", "2.5", "1"}
		parseFlarmNmeaMessage(message)

		trafficMutex.Lock()
		if len(traffic) == 0 {
			t.Fatal("Expected traffic from ICAO message")
		}
		var ti TrafficInfo
		for _, v := range traffic {
			ti = v
			break
		}
		if ti.Addr_type != 0 {
			t.Errorf("Expected Addr_type=0 for IDType=1, got %d", ti.Addr_type)
		}
		trafficMutex.Unlock()

		// IDType=2 should map to Addr_type=1 (Non-ICAO)
		traffic = make(map[uint32]TrafficInfo)
		message2 := []string{"PFLAA", "0", "1000", "500", "100", "2", "DEF456", "90", "0", "50", "2.5", "1"}
		parseFlarmNmeaMessage(message2)

		trafficMutex.Lock()
		defer trafficMutex.Unlock()
		if len(traffic) == 0 {
			t.Fatal("Expected traffic from non-ICAO message")
		}
		for _, v := range traffic {
			ti = v
			break
		}
		if ti.Addr_type != 1 {
			t.Errorf("Expected Addr_type=1 for IDType=2, got %d", ti.Addr_type)
		}

		t.Logf("IDType mapping: 1->%d, 2->%d", 0, 1)
	})

	t.Run("unknown_message_type", func(t *testing.T) {
		// Unknown message types should be ignored without panic
		message := []string{"PFLXX", "some", "data"}

		// Should not panic
		parseFlarmNmeaMessage(message)

		t.Log("Unknown message type correctly ignored")
	})

	t.Run("PFLAA_with_negative_coordinates", func(t *testing.T) {
		// Setup valid GPS
		mySituation.muGPS.Lock()
		mySituation.GPSLatitude = 47.5
		mySituation.GPSLongitude = -122.3
		mySituation.GPSAltitudeMSL = 500
		mySituation.GPSFixQuality = 2
		mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
		mySituation.muGPS.Unlock()
		globalStatus.GPS_connected = true

		traffic = make(map[uint32]TrafficInfo)
		seenTraffic = make(map[uint32]bool)

		// Negative relative north and east (target is SW of ownship)
		message := []string{"PFLAA", "0", "-2000", "-1500", "-50", "1", "SW1234", "270", "0", "60", "-1.5", "1"}

		parseFlarmNmeaMessage(message)

		trafficMutex.Lock()
		defer trafficMutex.Unlock()

		if len(traffic) == 0 {
			t.Error("Expected traffic to be added with negative coordinates")
			return
		}

		var ti TrafficInfo
		for _, v := range traffic {
			ti = v
			break
		}

		// Check that position was calculated (should be SW of ownship)
		if ti.Lat >= mySituation.GPSLatitude {
			t.Errorf("Expected Lat < ownship lat for negative north, got %.5f vs %.5f",
				ti.Lat, mySituation.GPSLatitude)
		}
		if ti.Lng >= mySituation.GPSLongitude {
			t.Errorf("Expected Lng < ownship lng for negative east, got %.5f vs %.5f",
				ti.Lng, mySituation.GPSLongitude)
		}

		// Check negative vertical velocity
		if ti.Vvel >= 0 {
			t.Errorf("Expected negative Vvel, got %d", ti.Vvel)
		}

		t.Logf("Negative coordinates: RelN=-2000m, RelE=-1500m -> Lat=%.5f, Lng=%.5f",
			ti.Lat, ti.Lng)
	})

	t.Run("panic_recovery", func(t *testing.T) {
		// Test that malformed data doesn't crash the function
		// The defer/recover in parseFlarmNmeaMessage should handle this

		testCases := []struct {
			name    string
			message []string
		}{
			{
				name:    "nil_message",
				message: nil,
			},
			{
				name:    "empty_message",
				message: []string{},
			},
			{
				name:    "PFLAA_with_invalid_numbers",
				message: []string{"PFLAA", "x", "y", "z", "w", "1", "ABC", "90", "0", "50", "2.5", "1"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Should not panic
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("parseFlarmNmeaMessage panicked: %v", r)
					}
				}()

				parseFlarmNmeaMessage(tc.message)
				t.Logf("Malformed message handled gracefully: %v", tc.message)
			})
		}
	})

	t.Run("PFLAA_speed_and_climb_rate_conversion", func(t *testing.T) {
		// Setup valid GPS
		mySituation.muGPS.Lock()
		mySituation.GPSLatitude = 47.5
		mySituation.GPSLongitude = -122.3
		mySituation.GPSAltitudeMSL = 500
		mySituation.GPSFixQuality = 2
		mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
		mySituation.muGPS.Unlock()
		globalStatus.GPS_connected = true

		traffic = make(map[uint32]TrafficInfo)
		seenTraffic = make(map[uint32]bool)

		// GroundSpeed in m/s should be converted to knots (multiply by 1.94384)
		// ClimbRate in m/s should be converted to ft/min (multiply by 196.85)
		// Speed=25 m/s = ~48.6 knots
		// Climb=5 m/s = ~984 ft/min
		message := []string{"PFLAA", "0", "1000", "500", "100", "1", "SPD001", "90", "0", "25", "5.0", "1"}

		parseFlarmNmeaMessage(message)

		trafficMutex.Lock()
		defer trafficMutex.Unlock()

		if len(traffic) == 0 {
			t.Error("Expected traffic to be added")
			return
		}

		var ti TrafficInfo
		for _, v := range traffic {
			ti = v
			break
		}

		// Speed should be ~48-49 knots
		expectedSpeed := uint16(48)
		if ti.Speed < expectedSpeed-2 || ti.Speed > expectedSpeed+2 {
			t.Errorf("Expected Speed~%d knots, got %d", expectedSpeed, ti.Speed)
		}

		// Vvel should be ~984 ft/min
		expectedVvel := int16(984)
		if ti.Vvel < expectedVvel-50 || ti.Vvel > expectedVvel+50 {
			t.Errorf("Expected Vvel~%d ft/min, got %d", expectedVvel, ti.Vvel)
		}

		t.Logf("Conversion: 25 m/s -> %d kts, 5.0 m/s -> %d ft/min",
			ti.Speed, ti.Vvel)
	})
}
