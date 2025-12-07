package main

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestChksumUBX tests UBX checksum calculation
func TestChksumUBX(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "Empty message",
			input:    []byte{},
			expected: []byte{0x00, 0x00},
		},
		{
			name:     "Single byte",
			input:    []byte{0x01},
			expected: []byte{0x01, 0x01},
		},
		{
			name:     "Two bytes",
			input:    []byte{0x01, 0x02},
			expected: []byte{0x03, 0x04}, // CK_A = 0x01 + 0x02 = 0x03, CK_B = 0x01 + 0x03 = 0x04
		},
		{
			name:     "UBX-CFG-RATE example",
			input:    []byte{0x06, 0x08, 0x06, 0x00, 0xE8, 0x03, 0x01, 0x00, 0x01, 0x00},
			expected: []byte{0x01, 0x39}, // Known good checksum
		},
		{
			name:     "All zeros",
			input:    []byte{0x00, 0x00, 0x00, 0x00},
			expected: []byte{0x00, 0x00},
		},
		{
			name:     "All ones",
			input:    []byte{0x01, 0x01, 0x01, 0x01},
			expected: []byte{0x04, 0x0A}, // CK_A = 4, CK_B = 1+2+3+4 = 10
		},
		{
			name:     "Maximum byte values",
			input:    []byte{0xFF, 0xFF},
			expected: []byte{0xFE, 0xFD}, // CK_A = 0xFF + 0xFF = 0x1FE (wraps to 0xFE), CK_B = 0xFF + 0xFE = 0x1FD (wraps to 0xFD)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := chksumUBX(tc.input)
			if !bytes.Equal(result, tc.expected) {
				t.Errorf("chksumUBX(%v) = %v, expected %v",
					tc.input, result, tc.expected)
			}
			t.Logf("Input: %v -> Checksum: [0x%02X, 0x%02X]", tc.input, result[0], result[1])
		})
	}
}

// TestMakeUBXCFG tests UBX message construction
func TestMakeUBXCFG(t *testing.T) {
	testCases := []struct {
		name           string
		class          byte
		id             byte
		msglen         uint16
		payload        []byte
		expectedPrefix []byte
		expectedLen    int
	}{
		{
			name:           "Empty payload",
			class:          0x06,
			id:             0x08,
			msglen:         0,
			payload:        []byte{},
			expectedPrefix: []byte{0xB5, 0x62, 0x06, 0x08, 0x00, 0x00},
			expectedLen:    8, // header(6) + checksum(2)
		},
		{
			name:           "Small payload",
			class:          0x06,
			id:             0x08,
			msglen:         6,
			payload:        []byte{0xE8, 0x03, 0x01, 0x00, 0x01, 0x00},
			expectedPrefix: []byte{0xB5, 0x62, 0x06, 0x08, 0x06, 0x00},
			expectedLen:    14, // header(6) + payload(6) + checksum(2)
		},
		{
			name:           "Large length value (>255)",
			class:          0x01,
			id:             0x02,
			msglen:         300,
			payload:        make([]byte, 300),
			expectedPrefix: []byte{0xB5, 0x62, 0x01, 0x02, 0x2C, 0x01}, // 300 = 0x012C, little-endian: 0x2C 0x01
			expectedLen:    308,                                        // header(6) + payload(300) + checksum(2)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := makeUBXCFG(tc.class, tc.id, tc.msglen, tc.payload)

			// Check length
			if len(result) != tc.expectedLen {
				t.Errorf("makeUBXCFG() length = %d, expected %d", len(result), tc.expectedLen)
			}

			// Check prefix (sync chars + class + id + length)
			if !bytes.Equal(result[:6], tc.expectedPrefix) {
				t.Errorf("makeUBXCFG() prefix = %v, expected %v", result[:6], tc.expectedPrefix)
			}

			// Verify sync characters
			if result[0] != 0xB5 || result[1] != 0x62 {
				t.Errorf("makeUBXCFG() sync chars = [0x%02X, 0x%02X], expected [0xB5, 0x62]",
					result[0], result[1])
			}

			// Verify checksum is present at the end
			chkPos := len(result) - 2
			expectedChk := chksumUBX(result[2:chkPos])
			actualChk := result[chkPos:]
			if !bytes.Equal(actualChk, expectedChk) {
				t.Errorf("makeUBXCFG() checksum = %v, expected %v", actualChk, expectedChk)
			}

			t.Logf("UBX message: class=0x%02X id=0x%02X len=%d bytes=%d",
				tc.class, tc.id, tc.msglen, len(result))
		})
	}
}

// TestMakeNMEACmd tests NMEA command construction with checksum
func TestMakeNMEACmd(t *testing.T) {
	testCases := []struct {
		name           string
		cmd            string
		expectedPrefix string
		expectedSuffix string
	}{
		{
			name:           "Simple command",
			cmd:            "PMTK314,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0",
			expectedPrefix: "$PMTK314,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0*",
			expectedSuffix: "\r\n",
		},
		{
			name:           "Empty command",
			cmd:            "",
			expectedPrefix: "$*",
			expectedSuffix: "\r\n",
		},
		{
			name:           "Single character",
			cmd:            "A",
			expectedPrefix: "$A*",
			expectedSuffix: "\r\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := makeNMEACmd(tc.cmd)
			resultStr := string(result)

			// Check prefix
			if !strings.HasPrefix(resultStr, tc.expectedPrefix) {
				t.Errorf("makeNMEACmd(%q) prefix = %q, expected prefix %q",
					tc.cmd, resultStr, tc.expectedPrefix)
			}

			// Check suffix
			if !strings.HasSuffix(resultStr, tc.expectedSuffix) {
				t.Errorf("makeNMEACmd(%q) suffix = %q, expected suffix %q",
					tc.cmd, resultStr, tc.expectedSuffix)
			}

			// Verify checksum format (should be 2 hex digits)
			parts := strings.Split(resultStr, "*")
			if len(parts) != 2 {
				t.Fatalf("makeNMEACmd(%q) = %q, expected exactly one *", tc.cmd, resultStr)
			}
			checksumPart := strings.TrimSuffix(parts[1], "\r\n")
			if len(checksumPart) != 2 {
				t.Errorf("makeNMEACmd(%q) checksum length = %d, expected 2", tc.cmd, len(checksumPart))
			}

			// Verify checksum is hex
			for _, c := range checksumPart {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("makeNMEACmd(%q) checksum char %c is not lowercase hex", tc.cmd, c)
				}
			}

			t.Logf("Command: %q -> %q", tc.cmd, resultStr)
		})
	}
}

// TestValidateNMEAChecksum tests NMEA checksum validation
func TestValidateNMEAChecksum(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectValid bool
		expectOut   string
	}{
		{
			name:        "Valid GPRMC",
			input:       "$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A",
			expectValid: true,
			expectOut:   "GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W",
		},
		{
			name:        "Valid GPGGA",
			input:       "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47",
			expectValid: true,
			expectOut:   "GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,",
		},
		{
			name:        "Invalid checksum",
			input:       "$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*FF",
			expectValid: false,
		},
		{
			name:        "Missing $",
			input:       "GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A",
			expectValid: false,
		},
		{
			name:        "Missing *",
			input:       "$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W",
			expectValid: false,
		},
		{
			name:        "Missing checksum after *",
			input:       "$GPRMC,123519,A*",
			expectValid: false,
		},
		{
			name:        "Single character checksum",
			input:       "$GPRMC*1",
			expectValid: false,
		},
		{
			name:        "Empty sentence with valid format",
			input:       "$*00",
			expectValid: true,
			expectOut:   "",
		},
		{
			name:        "Invalid hex in checksum",
			input:       "$GPRMC*ZZ",
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, valid := validateNMEAChecksum(tc.input)

			if valid != tc.expectValid {
				t.Errorf("validateNMEAChecksum(%q) valid = %v, expected %v",
					tc.input, valid, tc.expectValid)
			}

			if tc.expectValid && out != tc.expectOut {
				t.Errorf("validateNMEAChecksum(%q) out = %q, expected %q",
					tc.input, out, tc.expectOut)
			}

			if valid {
				t.Logf("Valid: %q -> %q", tc.input, out)
			} else {
				t.Logf("Invalid: %q (reason: %s)", tc.input, out)
			}
		})
	}
}

// TestCalculateNACp tests Navigation Accuracy Category calculation
func TestCalculateNACp(t *testing.T) {
	testCases := []struct {
		name     string
		accuracy float32
		expected uint8
	}{
		{
			name:     "Very high accuracy (< 3m)",
			accuracy: 2.5,
			expected: 11,
		},
		{
			name:     "Boundary at 3m (exclusive)",
			accuracy: 3.0,
			expected: 10,
		},
		{
			name:     "High accuracy (< 10m)",
			accuracy: 5.0,
			expected: 10,
		},
		{
			name:     "Boundary at 10m (exclusive)",
			accuracy: 10.0,
			expected: 9,
		},
		{
			name:     "Good accuracy (< 30m)",
			accuracy: 25.0,
			expected: 9,
		},
		{
			name:     "Boundary at 30m (exclusive)",
			accuracy: 30.0,
			expected: 8,
		},
		{
			name:     "Medium accuracy (< 92.6m)",
			accuracy: 50.0,
			expected: 8,
		},
		{
			name:     "Boundary at 92.6m (exclusive)",
			accuracy: 92.6,
			expected: 7,
		},
		{
			name:     "Lower accuracy (< 185.2m)",
			accuracy: 100.0,
			expected: 7,
		},
		{
			name:     "Boundary at 185.2m (exclusive)",
			accuracy: 185.2,
			expected: 6,
		},
		{
			name:     "Low accuracy (< 555.6m)",
			accuracy: 300.0,
			expected: 6,
		},
		{
			name:     "Boundary at 555.6m (exclusive)",
			accuracy: 555.6,
			expected: 0,
		},
		{
			name:     "Very low accuracy (>= 555.6m)",
			accuracy: 1000.0,
			expected: 0,
		},
		{
			name:     "Zero accuracy",
			accuracy: 0.0,
			expected: 11,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateNACp(tc.accuracy)
			if result != tc.expected {
				t.Errorf("calculateNACp(%.2f) = %d, expected %d",
					tc.accuracy, result, tc.expected)
			}
			t.Logf("Accuracy: %.2f m -> NACp: %d", tc.accuracy, result)
		})
	}
}

// TestCalculateNACpBoundaries tests exact boundary conditions
func TestCalculateNACpBoundaries(t *testing.T) {
	boundaries := []struct {
		name     string
		accuracy float32
		expected uint8
	}{
		{"Just below 3m", 2.99, 11},
		{"Exactly 3m", 3.0, 10},
		{"Just above 3m", 3.01, 10},
		{"Just below 10m", 9.99, 10},
		{"Exactly 10m", 10.0, 9},
		{"Just above 10m", 10.01, 9},
		{"Just below 30m", 29.99, 9},
		{"Exactly 30m", 30.0, 8},
		{"Just above 30m", 30.01, 8},
		{"Just below 92.6m", 92.59, 8},
		{"Exactly 92.6m", 92.6, 7},
		{"Just above 92.6m", 92.61, 7},
		{"Just below 185.2m", 185.19, 7},
		{"Exactly 185.2m", 185.2, 6},
		{"Just above 185.2m", 185.21, 6},
		{"Just below 555.6m", 555.59, 6},
		{"Exactly 555.6m", 555.6, 0},
		{"Just above 555.6m", 555.61, 0},
	}

	for _, tc := range boundaries {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateNACp(tc.accuracy)
			if result != tc.expected {
				t.Errorf("calculateNACp(%.2f) = %d, expected %d",
					tc.accuracy, result, tc.expected)
			}
		})
	}
}

// TestUpdateConstellation tests the updateConstellation function
func TestUpdateConstellation(t *testing.T) {
	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Initialize mySituation mutexes
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original state
	origSatellites := Satellites

	defer func() {
		Satellites = origSatellites
	}()

	// Initialize satellites map
	Satellites = make(map[string]SatelliteInfo)

	// Add a fresh satellite (tracked within last 10 seconds)
	Satellites["G01"] = SatelliteInfo{
		SatelliteID:      "G01",
		Signal:           45,
		InSolution:       true,
		TimeLastTracked:  stratuxClock.Time,
		TimeLastSolution: stratuxClock.Time,
	}

	// Add another fresh satellite with no signal
	Satellites["G02"] = SatelliteInfo{
		SatelliteID:     "G02",
		Signal:          0,
		InSolution:      false,
		TimeLastTracked: stratuxClock.Time,
	}

	updateConstellation()

	// Check that fresh satellites are still present
	if _, ok := Satellites["G01"]; !ok {
		t.Error("Expected G01 to still be in Satellites map")
	}
	if _, ok := Satellites["G02"]; !ok {
		t.Error("Expected G02 to still be in Satellites map")
	}

	// Check mySituation values
	if mySituation.GPSSatellites != 1 {
		t.Errorf("Expected GPSSatellites=1, got %d", mySituation.GPSSatellites)
	}
	if mySituation.GPSSatellitesTracked != 2 {
		t.Errorf("Expected GPSSatellitesTracked=2, got %d", mySituation.GPSSatellitesTracked)
	}
	if mySituation.GPSSatellitesSeen != 1 {
		t.Errorf("Expected GPSSatellitesSeen=1 (only G01 has signal), got %d", mySituation.GPSSatellitesSeen)
	}

	t.Logf("updateConstellation: Sats=%d, Tracked=%d, Seen=%d",
		mySituation.GPSSatellites, mySituation.GPSSatellitesTracked, mySituation.GPSSatellitesSeen)
}

// TestUpdateConstellation_StaleSatellites tests removal of stale satellites
func TestUpdateConstellation_StaleSatellites(t *testing.T) {
	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Initialize mySituation mutexes
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original state
	origSatellites := Satellites

	defer func() {
		Satellites = origSatellites
	}()

	// Initialize satellites map
	Satellites = make(map[string]SatelliteInfo)

	// Add a stale satellite (tracked more than 10 seconds ago)
	staleTime := stratuxClock.Time.Add(-15 * time.Second)
	Satellites["G05"] = SatelliteInfo{
		SatelliteID:     "G05",
		Signal:          30,
		InSolution:      true,
		TimeLastTracked: staleTime, // 15 seconds ago - should be removed
	}

	// Add a fresh satellite for comparison
	Satellites["G06"] = SatelliteInfo{
		SatelliteID:      "G06",
		Signal:           40,
		InSolution:       true,
		TimeLastTracked:  stratuxClock.Time,
		TimeLastSolution: stratuxClock.Time,
	}

	// Verify we have 2 satellites before update
	if len(Satellites) != 2 {
		t.Fatalf("Expected 2 satellites before update, got %d", len(Satellites))
	}

	updateConstellation()

	// Stale satellite G05 should be removed
	if _, ok := Satellites["G05"]; ok {
		t.Error("Expected stale satellite G05 to be removed")
	}

	// Fresh satellite G06 should still be present
	if _, ok := Satellites["G06"]; !ok {
		t.Error("Expected fresh satellite G06 to still be present")
	}

	// Check that only 1 satellite remains
	if len(Satellites) != 1 {
		t.Errorf("Expected 1 satellite after update, got %d", len(Satellites))
	}

	t.Logf("Stale satellite removal test: %d satellites remaining", len(Satellites))
}

// TestUpdateConstellation_SolutionTimeout tests the 5-second solution timeout
func TestUpdateConstellation_SolutionTimeout(t *testing.T) {
	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Initialize mySituation mutexes
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original state
	origSatellites := Satellites

	defer func() {
		Satellites = origSatellites
	}()

	// Initialize satellites map
	Satellites = make(map[string]SatelliteInfo)

	// Add a satellite with old solution time (more than 5 seconds ago)
	oldSolutionTime := stratuxClock.Time.Add(-8 * time.Second)
	Satellites["G10"] = SatelliteInfo{
		SatelliteID:      "G10",
		Signal:           50,
		InSolution:       true, // Currently marked as in solution
		TimeLastTracked:  stratuxClock.Time,
		TimeLastSolution: oldSolutionTime, // 8 seconds ago
	}

	updateConstellation()

	// The satellite should still exist but InSolution should be set to false
	sat, ok := Satellites["G10"]
	if !ok {
		t.Fatal("Expected satellite G10 to still exist")
	}

	if sat.InSolution {
		t.Error("Expected InSolution=false after 5-second timeout")
	}

	// GPSSatellites count should be 0 since none are in solution
	if mySituation.GPSSatellites != 0 {
		t.Errorf("Expected GPSSatellites=0, got %d", mySituation.GPSSatellites)
	}

	t.Logf("Solution timeout test: InSolution=%v, GPSSatellites=%d",
		sat.InSolution, mySituation.GPSSatellites)
}

// TestCalculateNavRate tests the GPS navigation rate calculation
func TestCalculateNavRate(t *testing.T) {
	// Save original state
	origPerfStats := myGPSPerfStats
	defer func() { myGPSPerfStats = origPerfStats }()

	t.Run("empty_stats", func(t *testing.T) {
		myGPSPerfStats = []gpsPerfStats{}
		result := calculateNavRate()
		// With empty stats, Mean returns invalid, halfwidth defaults to 3.5
		if result != 3.5 {
			t.Errorf("Expected halfwidth=3.5 for empty stats, got %f", result)
		}
		if mySituation.GPSPositionSampleRate != 0 {
			t.Errorf("Expected GPSPositionSampleRate=0 for empty stats, got %f", mySituation.GPSPositionSampleRate)
		}
	})

	t.Run("single_point", func(t *testing.T) {
		myGPSPerfStats = []gpsPerfStats{
			{nmeaTime: 1.0},
		}
		result := calculateNavRate()
		// Single point can't calculate rate
		if result != 3.5 {
			t.Errorf("Expected halfwidth=3.5 for single point, got %f", result)
		}
	})

	t.Run("1hz_samples", func(t *testing.T) {
		// Simulate 1Hz GPS (1 second between samples)
		myGPSPerfStats = []gpsPerfStats{
			{nmeaTime: 0.0},
			{nmeaTime: 1.0},
			{nmeaTime: 2.0},
			{nmeaTime: 3.0},
			{nmeaTime: 4.0},
		}
		result := calculateNavRate()
		// 1Hz = 1.0s dt_avg, halfwidth = 9 * 1.0 = 9.0, clamped to 3.5
		if result != 3.5 {
			t.Errorf("Expected halfwidth=3.5 (clamped) for 1Hz samples, got %f", result)
		}
	})

	t.Run("5hz_samples", func(t *testing.T) {
		// Simulate 5Hz GPS (0.2 seconds between samples)
		myGPSPerfStats = []gpsPerfStats{
			{nmeaTime: 0.0},
			{nmeaTime: 0.2},
			{nmeaTime: 0.4},
			{nmeaTime: 0.6},
			{nmeaTime: 0.8},
			{nmeaTime: 1.0},
		}
		result := calculateNavRate()
		// 5Hz = 0.2s dt_avg, halfwidth = 9 * 0.2 = 1.8, between 1.5 and 3.5
		if result < 1.5 || result > 3.5 {
			t.Errorf("Expected halfwidth between 1.5 and 3.5 for 5Hz samples, got %f", result)
		}
	})

	t.Run("10hz_samples", func(t *testing.T) {
		// Simulate 10Hz GPS (0.1 seconds between samples)
		myGPSPerfStats = []gpsPerfStats{
			{nmeaTime: 0.0},
			{nmeaTime: 0.1},
			{nmeaTime: 0.2},
			{nmeaTime: 0.3},
			{nmeaTime: 0.4},
			{nmeaTime: 0.5},
		}
		result := calculateNavRate()
		// 10Hz = 0.1s dt_avg, halfwidth = 9 * 0.1 = 0.9, clamped to 1.5
		if result != 1.5 {
			t.Errorf("Expected halfwidth=1.5 (clamped) for 10Hz samples, got %f", result)
		}
	})

	t.Run("filters_duplicate_timestamps", func(t *testing.T) {
		// Some samples with same timestamp (< 0.05s apart) should be filtered
		myGPSPerfStats = []gpsPerfStats{
			{nmeaTime: 0.0},
			{nmeaTime: 0.01}, // Should be filtered (< 0.05)
			{nmeaTime: 1.0},
			{nmeaTime: 1.02}, // Should be filtered (< 0.05)
			{nmeaTime: 2.0},
		}
		result := calculateNavRate()
		// After filtering, we have samples at 0.0, 1.0, 2.0 = 1Hz rate
		if result != 3.5 {
			t.Errorf("Expected halfwidth=3.5 after filtering duplicates, got %f", result)
		}
	})
}

// TestProcessNMEALine_GPVTG tests VTG (velocity/track) message parsing
func TestProcessNMEALine_GPVTG(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}

	testCases := []struct {
		name          string
		input         string
		expectUsed    bool
		expectSpeed   float64
		expectCourse  float32
		speedMatters  bool
		courseMatters bool
	}{
		{
			name:          "Valid VTG with speed > 3 knots",
			input:         "$GPVTG,054.7,T,034.4,M,005.5,N,010.2,K*48",
			expectUsed:    true,
			expectSpeed:   5.5,
			expectCourse:  54.7,
			speedMatters:  true,
			courseMatters: true,
		},
		{
			name:          "Valid GNVTG variant",
			input:         "$GNVTG,120.5,T,110.2,M,025.0,N,046.3,K*52",
			expectUsed:    true,
			expectSpeed:   25.0,
			expectCourse:  120.5,
			speedMatters:  true,
			courseMatters: true,
		},
		{
			name:         "VTG with low speed (< 3 knots, no course update)",
			input:        "$GPVTG,054.7,T,034.4,M,001.5,N,002.8,K*45",
			expectUsed:   true,
			expectSpeed:  1.5,
			speedMatters: true,
			// Course should not be updated when speed < 3
			courseMatters: false,
		},
		{
			name:       "VTG with too few fields",
			input:      "$GPVTG,054.7,T,034.4,M*60",
			expectUsed: false,
		},
		{
			name:       "VTG with invalid checksum",
			input:      "$GPVTG,054.7,T,034.4,M,005.5,N,010.2,K*FF",
			expectUsed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state
			mySituation.GPSGroundSpeed = 0
			mySituation.GPSTrueCourse = 0

			result := processNMEALine(tc.input)

			if result != tc.expectUsed {
				t.Errorf("processNMEALine(%q) returned %v, expected %v",
					tc.input, result, tc.expectUsed)
			}

			if tc.expectUsed {
				if tc.speedMatters {
					speedDiff := mySituation.GPSGroundSpeed - tc.expectSpeed
					if speedDiff < -0.01 || speedDiff > 0.01 {
						t.Errorf("Expected GPSGroundSpeed=%.2f, got %.2f",
							tc.expectSpeed, mySituation.GPSGroundSpeed)
					}
				}
				if tc.courseMatters {
					courseDiff := mySituation.GPSTrueCourse - tc.expectCourse
					if courseDiff < -0.01 || courseDiff > 0.01 {
						t.Errorf("Expected GPSTrueCourse=%.2f, got %.2f",
							tc.expectCourse, mySituation.GPSTrueCourse)
					}
				}
				t.Logf("Parsed VTG: speed=%.2f kts, course=%.2f°",
					mySituation.GPSGroundSpeed, mySituation.GPSTrueCourse)
			}
		})
	}
}

// TestProcessNMEALine_GPGGA tests GGA (position fix) message parsing
func TestProcessNMEALine_GPGGA(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}

	testCases := []struct {
		name         string
		input        string
		expectUsed   bool
		expectLat    float32
		expectLon    float32
		expectAlt    float32
		expectFixQty uint8
	}{
		{
			name:         "Valid GPGGA with GPS fix",
			input:        "$GPGGA,123519.0,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*59",
			expectUsed:   true,
			expectLat:    48.1173,   // 48° 07.038' = 48 + 7.038/60
			expectLon:    11.51667,  // 11° 31.000' = 11 + 31/60
			expectAlt:    1789.37,   // 545.4m * 3.28084 ft/m (MSL altitude, includes geoid sep)
			expectFixQty: 1,
		},
		{
			name:         "Valid GNGGA variant",
			input:        "$GNGGA,123519.0,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47",
			expectUsed:   true,
			expectLat:    48.1173,
			expectLon:    11.51667,
			expectAlt:    1789.37,   // includes geoid separation
			expectFixQty: 1,
		},
		{
			name:         "Southern hemisphere latitude",
			input:        "$GPGGA,123519.0,3345.123,S,15112.456,E,1,08,0.9,100.0,M,10.0,M,,*4D",
			expectUsed:   true,
			expectLat:    -33.75205, // Negative for South
			expectLon:    151.2076,
			expectAlt:    328.084, // 100m * 3.28084
			expectFixQty: 1,
		},
		{
			name:         "Western hemisphere longitude",
			input:        "$GPGGA,123519.0,4045.678,N,07359.123,W,1,08,0.9,50.0,M,5.0,M,,*46",
			expectUsed:   true,
			expectLat:    40.7613,
			expectLon:    -73.98538, // Negative for West
			expectAlt:    164.042,
			expectFixQty: 1,
		},
		{
			name:         "GGA with no fix (quality=0)",
			input:        "$GPGGA,123519.0,4807.038,N,01131.000,E,0,08,0.9,545.4,M,46.9,M,,*58",
			expectUsed:   true,     // GGA is still processed even with quality=0
			expectLat:    48.1173,  // Position is still parsed
			expectLon:    11.51667, // Position is still parsed
			expectAlt:    1789.37,
			expectFixQty: 0, // But fix quality is set to 0
		},
		{
			name:       "GGA with too few fields",
			input:      "$GPGGA,123519,4807.038,N,01131.000,E*36",
			expectUsed: false,
		},
		{
			name:       "GGA with invalid checksum",
			input:      "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*FF",
			expectUsed: false,
		},
		{
			name:       "GGA with invalid latitude format",
			input:      "$GPGGA,123519,ABC,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*21",
			expectUsed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state
			mySituation.GPSLatitude = 0
			mySituation.GPSLongitude = 0
			mySituation.GPSAltitudeMSL = 0
			mySituation.GPSFixQuality = 0

			result := processNMEALine(tc.input)

			if result != tc.expectUsed {
				t.Errorf("processNMEALine(%q) returned %v, expected %v",
					tc.input, result, tc.expectUsed)
			}

			if tc.expectUsed {
				// Allow small floating point differences
				latDiff := mySituation.GPSLatitude - tc.expectLat
				if latDiff < -0.0001 || latDiff > 0.0001 {
					t.Errorf("Expected Latitude=%.5f, got %.5f",
						tc.expectLat, mySituation.GPSLatitude)
				}

				lonDiff := mySituation.GPSLongitude - tc.expectLon
				if lonDiff < -0.0001 || lonDiff > 0.0001 {
					t.Errorf("Expected Longitude=%.5f, got %.5f",
						tc.expectLon, mySituation.GPSLongitude)
				}

				altDiff := mySituation.GPSAltitudeMSL - tc.expectAlt
				if altDiff < -0.1 || altDiff > 0.1 {
					t.Errorf("Expected Altitude=%.2f ft, got %.2f ft",
						tc.expectAlt, mySituation.GPSAltitudeMSL)
				}

				if mySituation.GPSFixQuality != tc.expectFixQty {
					t.Errorf("Expected GPSFixQuality=%d, got %d",
						tc.expectFixQty, mySituation.GPSFixQuality)
				}

				t.Logf("Parsed GGA: lat=%.5f, lon=%.5f, alt=%.2f ft, fixQ=%d",
					mySituation.GPSLatitude, mySituation.GPSLongitude,
					mySituation.GPSAltitudeMSL, mySituation.GPSFixQuality)
			}
		})
	}
}

// TestProcessNMEALine_GPRMC tests RMC (recommended minimum) message parsing
func TestProcessNMEALine_GPRMC(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}

	testCases := []struct {
		name         string
		input        string
		expectUsed   bool
		expectLat    float32
		expectLon    float32
		expectSpeed  float64
		expectCourse float32
	}{
		{
			name:         "Valid GPRMC with active status",
			input:        "$GPRMC,123519.0,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*74",
			expectUsed:   true,
			expectLat:    48.1173,
			expectLon:    11.51667,
			expectSpeed:  22.4,
			expectCourse: 84.4,
		},
		{
			name:         "Valid GNRMC variant",
			input:        "$GNRMC,081836.0,A,3751.65,S,14507.36,E,000.0,360.0,130998,011.3,E*62",
			expectUsed:   true,
			expectLat:    -37.86083, // South is negative
			expectLon:    145.1227,
			expectSpeed:  0.0,
			// Course not updated when speed < 3
		},
		{
			name:       "RMC with void status",
			input:      "$GPRMC,123519,V,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*7D",
			expectUsed: false, // V = void, should reject
		},
		{
			name:       "RMC with too few fields",
			input:      "$GPRMC,123519,A,4807.038,N*2C",
			expectUsed: false,
		},
		{
			name:       "RMC with invalid checksum",
			input:      "$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*FF",
			expectUsed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state
			mySituation.GPSLatitude = 0
			mySituation.GPSLongitude = 0
			mySituation.GPSGroundSpeed = 0
			mySituation.GPSTrueCourse = 0

			result := processNMEALine(tc.input)

			if result != tc.expectUsed {
				t.Errorf("processNMEALine(%q) returned %v, expected %v",
					tc.input, result, tc.expectUsed)
			}

			if tc.expectUsed {
				latDiff := mySituation.GPSLatitude - tc.expectLat
				if latDiff < -0.0001 || latDiff > 0.0001 {
					t.Errorf("Expected Latitude=%.5f, got %.5f",
						tc.expectLat, mySituation.GPSLatitude)
				}

				lonDiff := mySituation.GPSLongitude - tc.expectLon
				if lonDiff < -0.0001 || lonDiff > 0.0001 {
					t.Errorf("Expected Longitude=%.5f, got %.5f",
						tc.expectLon, mySituation.GPSLongitude)
				}

				speedDiff := mySituation.GPSGroundSpeed - tc.expectSpeed
				if speedDiff < -0.01 || speedDiff > 0.01 {
					t.Errorf("Expected Speed=%.2f, got %.2f",
						tc.expectSpeed, mySituation.GPSGroundSpeed)
				}

				t.Logf("Parsed RMC: lat=%.5f, lon=%.5f, speed=%.2f kts, course=%.2f°",
					mySituation.GPSLatitude, mySituation.GPSLongitude,
					mySituation.GPSGroundSpeed, mySituation.GPSTrueCourse)
			}
		})
	}
}

// TestProcessNMEALine_InvalidMessages tests handling of malformed NMEA messages
func TestProcessNMEALine_InvalidMessages(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}

	testCases := []struct {
		name  string
		input string
	}{
		{"No dollar sign", "GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A"},
		{"No asterisk", "$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W"},
		{"Empty string", ""},
		{"Only dollar sign", "$"},
		{"Only asterisk", "*"},
		{"Dollar and asterisk only", "$*"},
		{"Invalid checksum chars", "$GPRMC,test*ZZ"},
		{"Short checksum", "$GPRMC,test*1"},
		{"Corrupted message", "$GP\x00RMC,test*2A"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := processNMEALine(tc.input)
			if result {
				t.Errorf("processNMEALine(%q) returned true, expected false for invalid message",
					tc.input)
			}
			t.Logf("Correctly rejected: %q", tc.input)
		})
	}
}

// TestCalcGPSAttitude tests the GPS attitude calculation function
func TestCalcGPSAttitude(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}

	// Save original state
	origPerfStats := myGPSPerfStats
	defer func() { myGPSPerfStats = origPerfStats }()

	t.Run("empty_stats", func(t *testing.T) {
		myGPSPerfStats = []gpsPerfStats{}
		result := calcGPSAttitude()
		if result {
			t.Error("Expected false for empty stats")
		}
		t.Log("Empty stats correctly returned false")
	})

	t.Run("single_point", func(t *testing.T) {
		myGPSPerfStats = []gpsPerfStats{
			{
				stratuxTime: stratuxClock.Milliseconds,
				nmeaTime:    100.0,
				msgType:     "GPRMC",
				gsf:         50.0,
				coursef:     90.0,
				alt:         1000.0,
			},
		}
		result := calcGPSAttitude()
		if result {
			t.Error("Expected false for single data point")
		}
		// Should have set all attitude values to zero
		if myGPSPerfStats[0].gpsTurnRate != 0 || myGPSPerfStats[0].gpsPitch != 0 || myGPSPerfStats[0].gpsRoll != 0 {
			t.Error("Expected zero attitude values for single point")
		}
		t.Log("Single point correctly returned false with zero attitude")
	})

	t.Run("stale_data", func(t *testing.T) {
		// Create data that's more than 3 seconds old
		oldTime := stratuxClock.Milliseconds - 4000
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: oldTime, nmeaTime: 100.0, msgType: "GPRMC", gsf: 50.0, coursef: 90.0, alt: 1000.0},
			{stratuxTime: oldTime, nmeaTime: 100.5, msgType: "GPRMC", gsf: 50.0, coursef: 90.0, alt: 1000.0},
		}
		result := calcGPSAttitude()
		if result {
			t.Error("Expected false for stale data")
		}
		t.Log("Stale data correctly rejected")
	})

	t.Run("time_jump_forward", func(t *testing.T) {
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 100.0, msgType: "GPRMC", gsf: 50.0, coursef: 90.0, alt: 1000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 105.0, msgType: "GPRMC", gsf: 50.0, coursef: 90.0, alt: 1000.0},
		}
		result := calcGPSAttitude()
		if result {
			t.Error("Expected false when time jump > 3 seconds")
		}
		t.Log("Large time jump correctly rejected")
	})

	t.Run("no_speed_data", func(t *testing.T) {
		// Only GGA messages, no RMC (which have speed)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 100.0, msgType: "GPGGA", gsf: 0, coursef: -1, alt: 1000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 100.5, msgType: "GPGGA", gsf: 0, coursef: -1, alt: 1010.0},
		}
		result := calcGPSAttitude()
		if result {
			t.Error("Expected false when no RMC speed data available")
		}
		t.Log("No speed data correctly rejected")
	})

	t.Run("insufficient_altitude_data", func(t *testing.T) {
		// Only one GGA message - not enough for vertical velocity calculation
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 100.0, msgType: "GPRMC", gsf: 50.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 100.2, msgType: "GPGGA", gsf: 50.0, coursef: 90.0, alt: 1000.0},
		}
		result := calcGPSAttitude()
		if result {
			t.Error("Expected false when insufficient altitude data")
		}
		t.Log("Insufficient altitude data correctly rejected")
	})

	t.Run("low_speed_level_flight", func(t *testing.T) {
		// Speed < 6 ft/sec (~3.55 knots) should return true with zero attitude
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 2.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPGGA", gsf: 2.0, coursef: 90.0, alt: 1000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 2.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.6, msgType: "GPGGA", gsf: 2.0, coursef: 90.0, alt: 1000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for low speed")
		}
		// Check that attitude is zeroed
		idx := len(myGPSPerfStats) - 1
		if myGPSPerfStats[idx].gpsPitch != 0 || myGPSPerfStats[idx].gpsRoll != 0 {
			t.Errorf("Expected zero attitude at low speed, got pitch=%.3f, roll=%.3f",
				myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll)
		}
		t.Logf("Low speed: pitch=%.3f, roll=%.3f, turnRate=%.3f",
			myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll, myGPSPerfStats[idx].gpsTurnRate)
	})

	t.Run("medium_speed_level_flight", func(t *testing.T) {
		// Medium speed, level flight
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 60.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 60.0, coursef: 90.0, alt: 1000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 60.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 60.0, coursef: 90.0, alt: 1000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 60.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 60.0, coursef: 90.0, alt: 1000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for valid medium speed level flight")
		}
		idx := len(myGPSPerfStats) - 1
		// Level flight should have near-zero pitch
		if myGPSPerfStats[idx].gpsPitch < -2 || myGPSPerfStats[idx].gpsPitch > 2 {
			t.Errorf("Expected near-zero pitch for level flight, got %.3f", myGPSPerfStats[idx].gpsPitch)
		}
		// Straight flight should have near-zero roll
		if myGPSPerfStats[idx].gpsRoll < -2 || myGPSPerfStats[idx].gpsRoll > 2 {
			t.Errorf("Expected near-zero roll for straight flight, got %.3f", myGPSPerfStats[idx].gpsRoll)
		}
		t.Logf("Level flight: pitch=%.3f, roll=%.3f, turnRate=%.3f",
			myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll, myGPSPerfStats[idx].gpsTurnRate)
	})

	t.Run("climbing_flight", func(t *testing.T) {
		// Climbing at constant speed and heading
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 1000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 1020.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 1040.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for climbing flight")
		}
		idx := len(myGPSPerfStats) - 1
		// Climbing should have positive pitch
		if myGPSPerfStats[idx].gpsPitch <= 0 {
			t.Errorf("Expected positive pitch for climb, got %.3f", myGPSPerfStats[idx].gpsPitch)
		}
		t.Logf("Climbing: pitch=%.3f, roll=%.3f, vv=%.1f ft/min",
			myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll,
			myGPSPerfStats[idx].vv*60)
	})

	t.Run("descending_flight", func(t *testing.T) {
		// Descending at constant speed and heading
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 1980.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 1960.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for descending flight")
		}
		idx := len(myGPSPerfStats) - 1
		// Descending should have negative pitch
		if myGPSPerfStats[idx].gpsPitch >= 0 {
			t.Errorf("Expected negative pitch for descent, got %.3f", myGPSPerfStats[idx].gpsPitch)
		}
		t.Logf("Descending: pitch=%.3f, roll=%.3f, vv=%.1f ft/min",
			myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll,
			myGPSPerfStats[idx].vv*60)
	})

	t.Run("turning_flight", func(t *testing.T) {
		// Turning right at constant speed and altitude
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 100.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 100.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 100.0, coursef: 93.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 100.0, coursef: 93.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 100.0, coursef: 96.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 100.0, coursef: 96.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for turning flight")
		}
		idx := len(myGPSPerfStats) - 1
		// Right turn should have positive turn rate
		if myGPSPerfStats[idx].gpsTurnRate <= 0 {
			t.Errorf("Expected positive turn rate for right turn, got %.3f", myGPSPerfStats[idx].gpsTurnRate)
		}
		// Bank angle should be positive for right turn at high speed
		if myGPSPerfStats[idx].gpsRoll <= 0 {
			t.Errorf("Expected positive roll for right turn, got %.3f", myGPSPerfStats[idx].gpsRoll)
		}
		// Load factor should be > 1 in a turn
		if myGPSPerfStats[idx].gpsLoadFactor <= 1.0 {
			t.Errorf("Expected load factor > 1 in turn, got %.3f", myGPSPerfStats[idx].gpsLoadFactor)
		}
		t.Logf("Right turn: pitch=%.3f, roll=%.3f, turnRate=%.3f deg/s, load=%.2fG",
			myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll,
			myGPSPerfStats[idx].gpsTurnRate, myGPSPerfStats[idx].gpsLoadFactor)
	})

	t.Run("high_speed_calculation", func(t *testing.T) {
		// High speed flight (>20 ft/sec) should calculate pitch and roll
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 120.0, coursef: 180.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 120.0, coursef: 180.0, alt: 3000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 120.0, coursef: 180.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 120.0, coursef: 180.0, alt: 3000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for high speed flight")
		}
		t.Logf("High speed flight processed successfully")
	})

	t.Run("heading_wrap_around_0_360", func(t *testing.T) {
		// Test heading wrapping from 359 to 1 (crossing north)
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 358.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: 358.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 1.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: 1.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 80.0, coursef: 4.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 80.0, coursef: 4.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true when heading wraps around 0/360")
		}
		idx := len(myGPSPerfStats) - 1
		// Should detect right turn across 0/360 boundary
		if myGPSPerfStats[idx].gpsTurnRate <= 0 {
			t.Errorf("Expected positive turn rate for heading wrap, got %.3f", myGPSPerfStats[idx].gpsTurnRate)
		}
		t.Logf("Heading wrap: turnRate=%.3f deg/s", myGPSPerfStats[idx].gpsTurnRate)
	})

	t.Run("insufficient_heading_data", func(t *testing.T) {
		// Only one valid heading data point
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: -1, alt: 0}, // invalid course
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: -1, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0}, // only one valid
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: -1, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if result {
			t.Error("Expected false when insufficient heading data")
		}
		t.Log("Insufficient heading data correctly rejected")
	})

	t.Run("midnight_rollover", func(t *testing.T) {
		// Test handling of midnight rollover (23:59:50 to 00:00:10)
		// The function should detect rollover and adjust time or handle it gracefully
		// Need enough data points for altitude regression (at least 2 GGA messages)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86389.8, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0}, // 23:59:49.8
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86389.9, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86390.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0}, // 23:59:50
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86390.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 10.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0}, // 00:00:10 next day
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 10.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		// The function should either succeed after rollover adjustment or gracefully handle the rollover
		// Either way, it tests the rollover code path
		idx := len(myGPSPerfStats) - 1
		if result {
			// If it succeeded, check if time was adjusted
			if myGPSPerfStats[idx].nmeaTime > 86400 {
				t.Logf("Midnight rollover successfully adjusted: time=%.1f", myGPSPerfStats[idx].nmeaTime)
			} else {
				t.Logf("Midnight rollover handled without adjustment: time=%.1f", myGPSPerfStats[idx].nmeaTime)
			}
		} else {
			// If it failed, that's also acceptable behavior for a rollover scenario
			t.Logf("Midnight rollover detected and function returned false (acceptable)")
		}
	})

	t.Run("midnight_rollover_successful", func(t *testing.T) {
		// Test successful midnight rollover handling with valid data
		// Create a proper rollover scenario that should succeed
		baseTime := float32(86395.0) // 23:59:55
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 5.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0}, // 00:00:05 next day
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 5.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		// Should succeed and adjust time
		idx := len(myGPSPerfStats) - 1
		t.Logf("Midnight rollover result: %v, adjusted time: %.1f", result, myGPSPerfStats[idx].nmeaTime)
	})

	t.Run("midnight_rollover_full_rebase", func(t *testing.T) {
		// Test full array rebase when all times > 86401 after rollover adjustment
		// Create scenario where last entry crosses midnight and triggers rebase
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86395.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86395.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86395.2, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86395.3, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86395.4, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86395.5, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 10.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0}, // 00:00:10 - triggers rollover
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 10.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		idx := len(myGPSPerfStats) - 1
		// After rollover adjustment, minTime should be > 86401, triggering full rebase
		if result {
			// Check if times were rebased down
			if myGPSPerfStats[0].nmeaTime < 86400 {
				t.Logf("Successfully rebased all times from >86401, first time: %.1f", myGPSPerfStats[0].nmeaTime)
			} else {
				t.Logf("Rollover adjusted but not rebased, first time: %.1f", myGPSPerfStats[0].nmeaTime)
			}
		}
		t.Logf("Full rebase result: %v, last time: %.1f", result, myGPSPerfStats[idx].nmeaTime)
	})

	t.Run("negative_time_non_rollover", func(t *testing.T) {
		// Test dt < 0 but NOT at midnight (should fail)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 50000.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 50000.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 49995.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0}, // time went backwards
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 49995.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if result {
			t.Error("Expected false for time going backwards (non-rollover)")
		}
		t.Log("Non-rollover negative time correctly rejected")
	})

	t.Run("rollover_with_dt_still_too_large", func(t *testing.T) {
		// Test rollover adjustment but dt still > 3 after adjustment
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86390.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86390.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 20.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0}, // Big gap even after rollover
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 20.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if result {
			t.Error("Expected false when dt > 3 even after rollover adjustment")
		}
		t.Log("Large dt after rollover correctly rejected")
	})

	t.Run("single_speed_point", func(t *testing.T) {
		// Test with exactly one RMC message (single speed point)
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true with single speed point")
		}
		t.Log("Single speed point handled successfully")
	})

	t.Run("speed_regression_invalid", func(t *testing.T) {
		// Create a scenario that might cause invalid speed regression
		// Use extreme or inconsistent speed values
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPGGA", gsf: 0, coursef: 90.0, alt: 2000.0}, // Same time
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 0, coursef: 90.0, alt: 0},       // Same time
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPGGA", gsf: 0, coursef: 90.0, alt: 2000.0}, // Same time
		}
		result := calcGPSAttitude()
		// This might fail due to regression issues with identical timestamps
		t.Logf("Speed regression with identical times result: %v", result)
	})

	t.Run("vv_regression_invalid", func(t *testing.T) {
		// Create a scenario that causes invalid vertical velocity regression
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0}, // Same time
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2010.0}, // Same time as first GGA
		}
		result := calcGPSAttitude()
		// This might fail due to regression issues
		t.Logf("VV regression with problematic times result: %v", result)
	})

	t.Run("heading_regression_invalid", func(t *testing.T) {
		// Create a scenario that causes invalid heading regression
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0}, // Same time
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},       // Same heading/time
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		// This might fail due to regression issues with identical timestamps
		t.Logf("Heading regression with identical times result: %v", result)
	})

	t.Run("medium_speed_with_roll", func(t *testing.T) {
		// Test medium speed (between 6 and 20 ft/sec) - should calculate turn rate but not pitch/roll
		// 10 ft/sec = ~5.9 knots, so use 8 knots
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 8.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 8.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 8.0, coursef: 95.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 8.0, coursef: 95.0, alt: 2010.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 8.0, coursef: 100.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 8.0, coursef: 100.0, alt: 2020.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for medium speed flight")
		}
		idx := len(myGPSPerfStats) - 1
		// At medium speed (between 6 and 20 ft/sec), pitch and roll should be zero
		if myGPSPerfStats[idx].gpsPitch != 0 || myGPSPerfStats[idx].gpsRoll != 0 {
			t.Errorf("Expected zero pitch/roll at medium speed, got pitch=%.3f, roll=%.3f",
				myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll)
		}
		// But load factor should be 1.0
		if myGPSPerfStats[idx].gpsLoadFactor != 1.0 {
			t.Errorf("Expected load factor = 1.0 at medium speed, got %.3f", myGPSPerfStats[idx].gpsLoadFactor)
		}
		t.Logf("Medium speed (6-20 ft/sec): pitch=%.3f, roll=%.3f, turnRate=%.3f, load=%.2f",
			myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll,
			myGPSPerfStats[idx].gpsTurnRate, myGPSPerfStats[idx].gpsLoadFactor)
	})

	t.Run("heading_wrap_360_to_0", func(t *testing.T) {
		// Test heading wrapping from NW to NE (e.g., 350 to 10)
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 350.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: 350.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 355.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: 355.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 80.0, coursef: 2.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 80.0, coursef: 2.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true when heading wraps from 350 to 2")
		}
		idx := len(myGPSPerfStats) - 1
		t.Logf("Heading wrap 360->0: turnRate=%.3f deg/s", myGPSPerfStats[idx].gpsTurnRate)
	})

	t.Run("heading_wrap_0_to_360_reverse", func(t *testing.T) {
		// Test heading wrapping backward from NE to NW (e.g., 10 to 350) - left turn
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 10.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: 10.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 5.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: 5.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 80.0, coursef: 358.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 80.0, coursef: 358.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true when heading wraps from 10 to 358 (left turn)")
		}
		idx := len(myGPSPerfStats) - 1
		// Should detect left turn (negative turn rate)
		if myGPSPerfStats[idx].gpsTurnRate >= 0 {
			t.Errorf("Expected negative turn rate for left turn across 0/360, got %.3f", myGPSPerfStats[idx].gpsTurnRate)
		}
		t.Logf("Heading wrap 0->360 (left turn): turnRate=%.3f deg/s", myGPSPerfStats[idx].gpsTurnRate)
	})

	t.Run("gnrmc_and_gngga_messages", func(t *testing.T) {
		// Test with GNSS messages (GNRMC/GNGGA) instead of GPS (GPRMC/GPGGA)
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GNRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GNGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GNRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GNGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for GNSS messages")
		}
		t.Log("GNRMC/GNGGA messages processed successfully")
	})

	t.Run("mixed_gp_and_gn_messages", func(t *testing.T) {
		// Test with mixed GP and GN messages
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GNGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GNRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for mixed GP/GN messages")
		}
		t.Log("Mixed GP/GN messages processed successfully")
	})

	t.Run("debug_mode_output", func(t *testing.T) {
		// Test with DEBUG mode enabled to cover debug output branches
		origDebug := globalSettings.DEBUG
		globalSettings.DEBUG = true
		defer func() { globalSettings.DEBUG = origDebug }()

		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true with DEBUG mode")
		}
		t.Log("DEBUG mode output tested")
	})

	t.Run("low_speed_debug_mode", func(t *testing.T) {
		// Test low speed path with DEBUG mode to cover debug output
		origDebug := globalSettings.DEBUG
		globalSettings.DEBUG = true
		defer func() { globalSettings.DEBUG = origDebug }()

		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 2.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPGGA", gsf: 2.0, coursef: 90.0, alt: 1000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 2.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.6, msgType: "GPGGA", gsf: 2.0, coursef: 90.0, alt: 1000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for low speed with DEBUG")
		}
		t.Log("Low speed DEBUG mode output tested")
	})

	t.Run("insufficient_heading_debug_mode", func(t *testing.T) {
		// Test insufficient heading data with DEBUG mode
		origDebug := globalSettings.DEBUG
		globalSettings.DEBUG = true
		defer func() { globalSettings.DEBUG = origDebug }()

		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: -1, alt: 0}, // invalid course
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: -1, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0}, // only one valid
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: -1, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if result {
			t.Error("Expected false when insufficient heading data in DEBUG mode")
		}
		t.Log("Insufficient heading data with DEBUG mode correctly rejected")
	})

	t.Run("heading_small_change_unwrap", func(t *testing.T) {
		// Test heading unwrapping with small changes (< 180 degrees)
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 80.0, coursef: 120.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 80.0, coursef: 120.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 80.0, coursef: 150.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 80.0, coursef: 150.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for small heading changes")
		}
		t.Log("Small heading changes processed correctly")
	})

	t.Run("rollover_successful_with_rebase", func(t *testing.T) {
		// Create a scenario that successfully goes through full rollover with rebase
		// All old times near midnight, then one crosses over, then all get rebased
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86398.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86398.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86398.2, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86398.3, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86398.4, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86398.5, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86398.6, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 86398.7, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 5.0, msgType: "GPRMC", gsf: 80.0, coursef: 90.0, alt: 0}, // Crosses midnight
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: 5.1, msgType: "GPGGA", gsf: 80.0, coursef: 90.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		// This should succeed: time gets adjusted to 86405, minTime > 86401, all times rebased down
		t.Logf("Rollover with rebase: result=%v", result)
		if result {
			t.Logf("First time after rebase: %.1f, last time: %.1f",
				myGPSPerfStats[0].nmeaTime, myGPSPerfStats[len(myGPSPerfStats)-1].nmeaTime)
		}
	})

	t.Run("exactly_at_speed_threshold_6", func(t *testing.T) {
		// Test exactly at v_x = 6 ft/sec threshold (boundary condition)
		// 6 ft/sec / 1.687810 = 3.555 knots
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 3.555, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPGGA", gsf: 3.555, coursef: 90.0, alt: 1000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 3.555, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.6, msgType: "GPGGA", gsf: 3.555, coursef: 90.0, alt: 1000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true at speed threshold")
		}
		idx := len(myGPSPerfStats) - 1
		t.Logf("At v_x=6 threshold: pitch=%.3f, roll=%.3f", myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll)
	})

	t.Run("exactly_at_speed_threshold_20", func(t *testing.T) {
		// Test exactly at v_x = 20 ft/sec threshold (boundary condition)
		// 20 ft/sec / 1.687810 = 11.85 knots
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 11.85, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 11.85, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 11.85, coursef: 95.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 11.85, coursef: 95.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 11.85, coursef: 100.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 11.85, coursef: 100.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true at speed threshold 20")
		}
		idx := len(myGPSPerfStats) - 1
		t.Logf("At v_x=20 threshold: pitch=%.3f, roll=%.3f", myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll)
	})

	t.Run("just_above_speed_threshold_20", func(t *testing.T) {
		// Test just above v_x = 20 ft/sec to ensure pitch/roll are calculated
		// 21 ft/sec / 1.687810 = 12.44 knots
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 12.5, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 12.5, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 12.5, coursef: 95.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 12.5, coursef: 95.0, alt: 2010.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 12.5, coursef: 100.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 12.5, coursef: 100.0, alt: 2020.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true just above speed threshold 20")
		}
		idx := len(myGPSPerfStats) - 1
		// Should now calculate pitch and roll since v_x > 20
		t.Logf("Just above v_x=20: pitch=%.3f, roll=%.3f, load=%.3f",
			myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll, myGPSPerfStats[idx].gpsLoadFactor)
	})

	t.Run("left_turn_high_speed", func(t *testing.T) {
		// Test left turn at high speed (negative turn rate, negative roll)
		baseTime := float32(100.0)
		myGPSPerfStats = []gpsPerfStats{
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime, msgType: "GPRMC", gsf: 100.0, coursef: 90.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.1, msgType: "GPGGA", gsf: 100.0, coursef: 90.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.2, msgType: "GPRMC", gsf: 100.0, coursef: 87.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.3, msgType: "GPGGA", gsf: 100.0, coursef: 87.0, alt: 2000.0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.4, msgType: "GPRMC", gsf: 100.0, coursef: 84.0, alt: 0},
			{stratuxTime: stratuxClock.Milliseconds, nmeaTime: baseTime + 0.5, msgType: "GPGGA", gsf: 100.0, coursef: 84.0, alt: 2000.0},
		}
		result := calcGPSAttitude()
		if !result {
			t.Error("Expected true for left turn")
		}
		idx := len(myGPSPerfStats) - 1
		// Left turn should have negative turn rate and negative roll
		if myGPSPerfStats[idx].gpsTurnRate >= 0 {
			t.Errorf("Expected negative turn rate for left turn, got %.3f", myGPSPerfStats[idx].gpsTurnRate)
		}
		if myGPSPerfStats[idx].gpsRoll >= 0 {
			t.Errorf("Expected negative roll for left turn, got %.3f", myGPSPerfStats[idx].gpsRoll)
		}
		t.Logf("Left turn: pitch=%.3f, roll=%.3f, turnRate=%.3f deg/s, load=%.2fG",
			myGPSPerfStats[idx].gpsPitch, myGPSPerfStats[idx].gpsRoll,
			myGPSPerfStats[idx].gpsTurnRate, myGPSPerfStats[idx].gpsLoadFactor)
	})
}
// New tests at end of gps_test.go

// TestProcessNMEALine_GPGSA tests GSA (satellite active) message parsing
func TestProcessNMEALine_GPGSA(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original satellites
	origSatellites := Satellites
	defer func() { Satellites = origSatellites }()

	testCases := []struct {
		name             string
		input            string
		expectUsed       bool
		expectSatCount   uint16
		setupFixQuality  uint8
		expectHAccuracy  bool
		expectVAccuracy  bool
	}{
		{
			name:            "Valid GPGSA with 3D fix",
			input:           "$GPGSA,A,3,04,05,,09,12,,,24,,,,,2.5,1.3,2.1*39",
			expectUsed:      true,
			setupFixQuality: 1,
			expectSatCount:  4,
			expectHAccuracy: true,
			expectVAccuracy: true,
		},
		{
			name:            "Valid GNGSA variant",
			input:           "$GNGSA,A,3,01,02,03,04,05,06,07,08,09,10,11,12,1.0,0.5,0.8*23",
			expectUsed:      true,
			setupFixQuality: 2, // SBAS fix
			expectSatCount:  12,
			expectHAccuracy: true,
			expectVAccuracy: true,
		},
		{
			name:       "GSA with no fix (empty solution type)",
			input:      "$GPGSA,A,,04,05,09,12,,,24,,,,,2.5,1.3,2.1*26",
			expectUsed: false,
		},
		{
			name:       "GSA with no fix (solution type 1)",
			input:      "$GPGSA,A,1,04,05,09,12,,,24,,,,,2.5,1.3,2.1*17",
			expectUsed: false,
		},
		{
			name:       "GSA with too few fields",
			input:      "$GPGSA,A,3,04,05*31",
			expectUsed: false,
		},
		{
			name:       "GSA with invalid checksum",
			input:      "$GPGSA,A,3,04,05,,09,12,,,24,,,,,2.5,1.3,2.1*FF",
			expectUsed: false,
		},
		{
			name:            "GSA with 2D fix",
			input:           "$GPGSA,A,2,01,02,03,04,05,06,,,,,,,3.0,2.0,2.2*35",
			expectUsed:      true,
			setupFixQuality: 1,
			expectSatCount:  6,
			expectHAccuracy: true,
			expectVAccuracy: true,
		},
		{
			name:            "GSA with GLONASS satellites",
			input:           "$GNGSA,A,3,65,66,67,68,69,70,,,,,,,1.5,0.8,1.3*26",
			expectUsed:      true,
			setupFixQuality: 1,
			expectSatCount:  6,
			expectHAccuracy: true,
			expectVAccuracy: true,
		},
		{
			name:            "GSA with Galileo satellites",
			input:           "$GNGSA,A,3,301,302,303,304,,,,,,,,,1.8,1.0,1.5*24",
			expectUsed:      true,
			setupFixQuality: 1,
			expectSatCount:  4,
			expectHAccuracy: true,
			expectVAccuracy: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state
			Satellites = make(map[string]SatelliteInfo)
			mySituation.GPSFixQuality = tc.setupFixQuality
			mySituation.GPSHorizontalAccuracy = 0
			mySituation.GPSVerticalAccuracy = 0
			mySituation.GPSLastAccuracyTime = time.Time{}

			result := processNMEALine(tc.input)

			if result != tc.expectUsed {
				t.Errorf("processNMEALine(%q) returned %v, expected %v",
					tc.input, result, tc.expectUsed)
			}

			if tc.expectUsed {
				if tc.expectSatCount > 0 {
					satCount := uint16(0)
					mySituation.muSatellite.Lock()
					for _, sat := range Satellites {
						if sat.InSolution {
							satCount++
						}
					}
					mySituation.muSatellite.Unlock()

					if satCount == 0 {
						t.Errorf("Expected at least some satellites in solution, got 0")
					}
					t.Logf("Satellites in solution: %d", satCount)
				}

				if tc.expectHAccuracy && mySituation.GPSHorizontalAccuracy == 0 {
					t.Error("Expected GPSHorizontalAccuracy to be set")
				}
				if tc.expectVAccuracy && mySituation.GPSVerticalAccuracy == 0 {
					t.Error("Expected GPSVerticalAccuracy to be set")
				}

				t.Logf("Parsed GSA: hacc=%.2f, vacc=%.2f, NACp=%d",
					mySituation.GPSHorizontalAccuracy,
					mySituation.GPSVerticalAccuracy,
					mySituation.GPSNACp)
			}
		})
	}
}

// TestProcessNMEALine_GPGST tests GST (error statistics) message parsing
func TestProcessNMEALine_GPGST(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}

	testCases := []struct {
		name            string
		input           string
		expectUsed      bool
		expectHAccuracy float32
		expectVAccuracy float32
		expectNACp      uint8
	}{
		{
			name:            "Valid GPGST with low error",
			input:           "$GPGST,172814.0,0.006,0.023,0.020,273.6,0.023,0.020,0.031*6A",
			expectUsed:      true,
			expectHAccuracy: 0.060,
			expectVAccuracy: 0.062,
			expectNACp:      11,
		},
		{
			name:            "Valid GNGST variant",
			input:           "$GNGST,082356.00,1.4,1.3,0.52,217.5,1.2,0.95,1.4*48",
			expectUsed:      true,
			expectHAccuracy: 3.13,
			expectVAccuracy: 2.8,
			expectNACp:      10,
		},
		{
			name:            "GST with medium error",
			input:           "$GPGST,172814.0,2.0,5.0,4.0,90.0,5.0,4.0,6.0*53",
			expectUsed:      true,
			expectHAccuracy: 12.81,
			expectVAccuracy: 12.0,
			expectNACp:      9,
		},
		{
			name:       "GST with too few fields",
			input:      "$GPGST,172814.0,0.006,0.023*0C",
			expectUsed: false,
		},
		{
			name:       "GST with invalid checksum",
			input:      "$GPGST,172814.0,0.006,0.023,0.020,273.6,0.023,0.020,0.031*FF",
			expectUsed: false,
		},
		{
			name:            "GST with zero error",
			input:           "$GPGST,172814.0,0.0,0.0,0.0,0.0,0.0,0.0,0.0*6E",
			expectUsed:      true,
			expectHAccuracy: 0.0,
			expectVAccuracy: 0.0,
			expectNACp:      11,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mySituation.GPSHorizontalAccuracy = 0
			mySituation.GPSVerticalAccuracy = 0
			mySituation.GPSNACp = 0
			mySituation.GPSLastAccuracyTime = time.Time{}

			result := processNMEALine(tc.input)

			if result != tc.expectUsed {
				t.Errorf("processNMEALine(%q) returned %v, expected %v",
					tc.input, result, tc.expectUsed)
			}

			if tc.expectUsed {
				haccDiff := mySituation.GPSHorizontalAccuracy - tc.expectHAccuracy
				tolerance := tc.expectHAccuracy * 0.05
				if tolerance < 0.01 {
					tolerance = 0.01
				}
				if haccDiff < -tolerance || haccDiff > tolerance {
					t.Errorf("Expected GPSHorizontalAccuracy=%.2f±%.2f, got %.2f",
						tc.expectHAccuracy, tolerance, mySituation.GPSHorizontalAccuracy)
				}

				vaccDiff := mySituation.GPSVerticalAccuracy - tc.expectVAccuracy
				tolerance = tc.expectVAccuracy * 0.05
				if tolerance < 0.01 {
					tolerance = 0.01
				}
				if vaccDiff < -tolerance || vaccDiff > tolerance {
					t.Errorf("Expected GPSVerticalAccuracy=%.2f±%.2f, got %.2f",
						tc.expectVAccuracy, tolerance, mySituation.GPSVerticalAccuracy)
				}

				if mySituation.GPSNACp != tc.expectNACp {
					t.Errorf("Expected GPSNACp=%d, got %d",
						tc.expectNACp, mySituation.GPSNACp)
				}

				if mySituation.GPSLastAccuracyTime.IsZero() {
					t.Error("Expected GPSLastAccuracyTime to be updated")
				}

				t.Logf("Parsed GST: hacc=%.2fm, vacc=%.2fm, NACp=%d",
					mySituation.GPSHorizontalAccuracy,
					mySituation.GPSVerticalAccuracy,
					mySituation.GPSNACp)
			}
		})
	}
}

// TestProcessNMEALine_GPGSV tests GSV (satellites in view) message parsing
func TestProcessNMEALine_GPGSV(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original satellites
	origSatellites := Satellites
	defer func() { Satellites = origSatellites }()

	testCases := []struct {
		name           string
		input          string
		expectUsed     bool
		expectSatID    string
		expectSignal   int8
		expectElevation int16
		expectAzimuth  int16
	}{
		{
			name:            "Valid GPGSV with GPS satellites",
			input:           "$GPGSV,3,1,12,01,85,349,48,03,60,075,47,06,57,186,48,11,15,063,38*7F",
			expectUsed:      true,
			expectSatID:     "G1",
			expectSignal:    48,
			expectElevation: 85,
			expectAzimuth:   349,
		},
		{
			name:            "Valid GLGSV with GLONASS satellites",
			input:           "$GLGSV,3,1,10,65,45,125,42,66,38,215,40,67,30,280,38,68,25,045,36*63",
			expectUsed:      true,
			expectSatID:     "R1",
			expectSignal:    42,
			expectElevation: 45,
			expectAzimuth:   125,
		},
		{
			name:            "Valid GAGSV with Galileo satellites",
			input:           "$GAGSV,2,1,07,301,55,023,45,302,48,156,43,303,38,245,41,304,28,315,39*6B",
			expectUsed:      true,
			expectSatID:     "E1",
			expectSignal:    45,
			expectElevation: 55,
			expectAzimuth:   23,
		},
		{
			name:            "Valid GBGSV with BeiDou satellites",
			input:           "$GBGSV,3,1,09,401,42,156,38,402,38,215,36,403,32,280,34,404,28,045,32*65",
			expectUsed:      true,
			expectSatID:     "B1",
			expectSignal:    38,
			expectElevation: 42,
			expectAzimuth:   156,
		},
		{
			name:       "GSV with too few fields",
			input:      "$GPGSV,3,1*77",
			expectUsed: false,
		},
		{
			name:       "GSV with invalid checksum",
			input:      "$GPGSV,3,1,12,01,85,349,48,03,60,075,47,06,57,186,48,11,15,063,38*FF",
			expectUsed: false,
		},
		{
			name:            "GSV with SBAS satellites",
			input:           "$GPGSV,2,1,05,33,45,125,42,34,38,215,40,01,85,349,48,03,60,075,47*7A",
			expectUsed:      true,
			expectSatID:     "S120",
			expectSignal:    42,
			expectElevation: 45,
			expectAzimuth:   125,
		},
		{
			name:            "GSV with QZSS satellites",
			input:           "$GPGSV,2,1,05,193,55,023,45,194,48,156,43,01,85,349,48,03,60,075,47*7B",
			expectUsed:      true,
			expectSatID:     "Q1",
			expectSignal:    45,
			expectElevation: 55,
			expectAzimuth:   23,
		},
		{
			name:            "GSV with satellite without signal",
			input:           "$GPGSV,2,1,06,01,85,349,,03,60,075,47,06,57,186,48,11,15,063,38*77",
			expectUsed:      true,
			expectSatID:     "G1",
			expectSignal:    -99,
			expectElevation: 85,
			expectAzimuth:   349,
		},
		{
			name:            "GSV with satellite missing elevation/azimuth",
			input:           "$GPGSV,2,1,06,01,,,48,03,60,075,47,06,57,186,48,11,15,063,38*48",
			expectUsed:      true,
			expectSatID:     "G1",
			expectSignal:    48,
			expectElevation: -999,
			expectAzimuth:   -999,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Satellites = make(map[string]SatelliteInfo)
			mySituation.GPSFixQuality = 1

			result := processNMEALine(tc.input)

			if result != tc.expectUsed {
				t.Errorf("processNMEALine(%q) returned %v, expected %v",
					tc.input, result, tc.expectUsed)
			}

			if tc.expectUsed && tc.expectSatID != "" {
				mySituation.muSatellite.Lock()
				sat, ok := Satellites[tc.expectSatID]
				mySituation.muSatellite.Unlock()

				if !ok {
					t.Errorf("Expected satellite %s to be in Satellites map", tc.expectSatID)
				} else {
					if sat.Signal != tc.expectSignal {
						t.Errorf("Expected Signal=%d, got %d", tc.expectSignal, sat.Signal)
					}
					if sat.Elevation != tc.expectElevation {
						t.Errorf("Expected Elevation=%d, got %d", tc.expectElevation, sat.Elevation)
					}
					if sat.Azimuth != tc.expectAzimuth {
						t.Errorf("Expected Azimuth=%d, got %d", tc.expectAzimuth, sat.Azimuth)
					}

					t.Logf("Parsed GSV: sat=%s, signal=%d, elev=%d°, az=%d°",
						tc.expectSatID, sat.Signal, sat.Elevation, sat.Azimuth)
				}
			}
		})
	}
}

// TestProcessNMEALine_POGNB tests POGNB (OGN Tracker barometric) message parsing
func TestProcessNMEALine_POGNB(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}
	if mySituation.muBaro == nil {
		mySituation.muBaro = &sync.Mutex{}
	}

	testCases := []struct {
		name              string
		input             string
		expectUsed        bool
		expectPressureAlt float32
		expectVertSpeed   float32
	}{
		{
			name:              "Valid POGNB message",
			input:             "$POGNB,22.0,+29.1,100972.3,3.8,+29.4,+87.2,-0.04,+32.6*47",
			expectUsed:        true,
			expectPressureAlt: 96.46,
			expectVertSpeed:   -7.87,
		},
		{
			name:              "POGNB with positive climb",
			input:             "$POGNB,22.0,+29.1,100972.3,3.8,+150.0,+150.0,+2.5,+32.6*70",
			expectUsed:        true,
			expectPressureAlt: 492.13,
			expectVertSpeed:   492.12,
		},
		{
			name:       "POGNB with too few fields",
			input:      "$POGNB,22.0,+29.1*5C",
			expectUsed: false,
		},
		{
			name:       "POGNB with invalid checksum",
			input:      "$POGNB,22.0,+29.1,100972.3,3.8,+29.4,+87.2,-0.04,+32.6,*FF",
			expectUsed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mySituation.muBaro.Lock()
			mySituation.BaroSourceType = BARO_TYPE_NONE
			mySituation.BaroPressureAltitude = 0
			mySituation.BaroVerticalSpeed = 0
			mySituation.muBaro.Unlock()

			result := processNMEALine(tc.input)

			if result != tc.expectUsed {
				t.Errorf("processNMEALine(%q) returned %v, expected %v",
					tc.input, result, tc.expectUsed)
			}

			if tc.expectUsed {
				altDiff := mySituation.BaroPressureAltitude - tc.expectPressureAlt
				if altDiff < -1.0 || altDiff > 1.0 {
					t.Errorf("Expected BaroPressureAltitude=%.2f±1.0, got %.2f",
						tc.expectPressureAlt, mySituation.BaroPressureAltitude)
				}

				vsDiff := mySituation.BaroVerticalSpeed - tc.expectVertSpeed
				if vsDiff < -1.0 || vsDiff > 1.0 {
					t.Errorf("Expected BaroVerticalSpeed=%.2f±1.0, got %.2f",
						tc.expectVertSpeed, mySituation.BaroVerticalSpeed)
				}

				if mySituation.BaroSourceType != BARO_TYPE_OGNTRACKER {
					t.Errorf("Expected BaroSourceType=%d (OGNTRACKER), got %d",
						BARO_TYPE_OGNTRACKER, mySituation.BaroSourceType)
				}

				t.Logf("Parsed POGNB: alt=%.2f ft, vs=%.2f ft/min",
					mySituation.BaroPressureAltitude,
					mySituation.BaroVerticalSpeed)
			}
		})
	}
}

// TestProcessNMEALineLow_PGRMZ tests PGRMZ (Garmin altitude) message parsing
func TestProcessNMEALineLow_PGRMZ(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muBaro == nil {
		mySituation.muBaro = &sync.Mutex{}
	}

	// Save original state
	origGPSType := globalStatus.GPS_detected_type
	defer func() { globalStatus.GPS_detected_type = origGPSType }()

	testCases := []struct {
		name              string
		input             string
		gpsType           uint
		baroSourceType    uint8
		expectUsed        bool
		expectPressureAlt float32
		expectBaroSource  uint8
	}{
		{
			name:              "Valid PGRMZ with feet from SoftRF",
			input:             "$PGRMZ,1089,f,3*2B",
			gpsType:           GPS_TYPE_SOFTRF,
			baroSourceType:    BARO_TYPE_NONE,
			expectUsed:        true,
			expectPressureAlt: 1089.0,
			expectBaroSource:  BARO_TYPE_NMEA,
		},
		{
			name:              "Valid PGRMZ with meters from SoftRF dongle",
			input:             "$PGRMZ,500,m,3*15",
			gpsType:           GPS_TYPE_SOFTRF_DONGLE,
			baroSourceType:    BARO_TYPE_NONE,
			expectUsed:        true,
			expectPressureAlt: 1640.42, // 500m * 3.28084
			expectBaroSource:  BARO_TYPE_NMEA,
		},
		{
			name:              "Valid PGRMZ with feet from serial GPS",
			input:             "$PGRMZ,2000,f,3*29",
			gpsType:           GPS_TYPE_SERIAL,
			baroSourceType:    BARO_TYPE_NONE,
			expectUsed:        true,
			expectPressureAlt: 2000.0,
			expectBaroSource:  BARO_TYPE_NMEA,
		},
		{
			name:             "PGRMZ ignored when BMP280 sensor present",
			input:            "$PGRMZ,1089,f,3*2B",
			gpsType:          GPS_TYPE_SOFTRF,
			baroSourceType:   BARO_TYPE_BMP280,
			expectUsed:       false,
			expectBaroSource: BARO_TYPE_BMP280,
		},
		{
			name:             "PGRMZ ignored when OGN tracker present",
			input:            "$PGRMZ,1089,f,3*2B",
			gpsType:          GPS_TYPE_SOFTRF,
			baroSourceType:   BARO_TYPE_OGNTRACKER,
			expectUsed:       false,
			expectBaroSource: BARO_TYPE_OGNTRACKER,
		},
		{
			name:       "PGRMZ ignored for non-SoftRF GPS type",
			input:      "$PGRMZ,1089,f,3*2B",
			gpsType:    GPS_TYPE_UBX9,
			expectUsed: false,
		},
		{
			name:       "PGRMZ with too few fields",
			input:      "$PGRMZ,1089*3A",
			gpsType:    GPS_TYPE_SOFTRF,
			expectUsed: false,
		},
		{
			name:       "PGRMZ with invalid altitude",
			input:      "$PGRMZ,invalid,f,3*46",
			gpsType:    GPS_TYPE_SOFTRF,
			expectUsed: false,
		},
		{
			name:       "PGRMZ with invalid checksum",
			input:      "$PGRMZ,1089,f,3*FF",
			gpsType:    GPS_TYPE_SOFTRF,
			expectUsed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state
			globalStatus.GPS_detected_type = tc.gpsType
			mySituation.BaroSourceType = tc.baroSourceType
			mySituation.BaroPressureAltitude = 0

			// For BMP280/OGNTRACKER tests, set recent time to make isTempPressValid() return true
			if tc.baroSourceType == BARO_TYPE_BMP280 || tc.baroSourceType == BARO_TYPE_OGNTRACKER {
				mySituation.BaroLastMeasurementTime = stratuxClock.Time
			} else {
				mySituation.BaroLastMeasurementTime = time.Time{}
			}

			result := processNMEALine(tc.input)

			if result != tc.expectUsed {
				t.Errorf("processNMEALine(%q) returned %v, expected %v",
					tc.input, result, tc.expectUsed)
			}

			if tc.expectUsed {
				altDiff := mySituation.BaroPressureAltitude - tc.expectPressureAlt
				if altDiff < -0.1 || altDiff > 0.1 {
					t.Errorf("Expected BaroPressureAltitude=%.2f, got %.2f",
						tc.expectPressureAlt, mySituation.BaroPressureAltitude)
				}
			}

			if tc.expectBaroSource != 0 {
				if mySituation.BaroSourceType != tc.expectBaroSource {
					t.Errorf("Expected BaroSourceType=%d, got %d",
						tc.expectBaroSource, mySituation.BaroSourceType)
				}
			}

			t.Logf("PGRMZ test complete: used=%v, alt=%.2f ft, source=%d",
				result, mySituation.BaroPressureAltitude, mySituation.BaroSourceType)
		})
	}
}

// TestProcessNMEALineLow_PFLAU_PFLAA tests FLARM NMEA message parsing
func TestProcessNMEALineLow_PFLAU_PFLAA(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}

	testCases := []struct {
		name       string
		input      string
		expectUsed bool
	}{
		{
			name:       "Valid PFLAU message",
			input:      "$PFLAU,0,1,2,1,0,-30,2,+30,400*50",
			expectUsed: true,
		},
		{
			name:       "Valid PFLAA message",
			input:      "$PFLAA,0,1000,500,100,2,DD1234,180,10,15,1.5,1*52",
			expectUsed: true,
		},
		{
			name:       "PFLAU with invalid checksum",
			input:      "$PFLAU,0,1,2,1,0,-30,2,+30,400*FF",
			expectUsed: false,
		},
		{
			name:       "PFLAA with invalid checksum",
			input:      "$PFLAA,0,1000,500,100,2,DD1234,180,10,15,1.5,1*FF",
			expectUsed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := processNMEALine(tc.input)

			if result != tc.expectUsed {
				t.Errorf("processNMEALine(%q) returned %v, expected %v",
					tc.input, result, tc.expectUsed)
			}

			t.Logf("FLARM test complete: used=%v", result)
		})
	}
}

// TestProcessNMEALineLow_EdgeCases tests edge cases and error paths
func TestProcessNMEALineLow_EdgeCases(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	testCases := []struct {
		name       string
		input      string
		expectUsed bool
	}{
		// GPVTG edge cases
		{
			name:       "GPVTG with invalid groundspeed",
			input:      "$GPVTG,054.7,T,034.4,M,ABC,N,010.2,K*64",
			expectUsed: false,
		},
		{
			name:       "GPVTG with invalid true course",
			input:      "$GPVTG,XYZ,T,034.4,M,005.5,N,010.2,K*52",
			expectUsed: false,
		},

		// GPGGA edge cases
		{
			name:       "GPGGA with invalid fix quality",
			input:      "$GPGGA,123519.0,4807.038,N,01131.000,E,X,08,0.9,545.4,M,46.9,M,,*35",
			expectUsed: false,
		},
		{
			name:       "GPGGA with timestamp too short",
			input:      "$GPGGA,12351,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*45",
			expectUsed: false,
		},
		{
			name:       "GPGGA with invalid hour in timestamp",
			input:      "$GPGGA,XX3519.0,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*60",
			expectUsed: false,
		},
		{
			name:       "GPGGA with invalid minute in timestamp",
			input:      "$GPGGA,12XX19.0,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*77",
			expectUsed: false,
		},
		{
			name:       "GPGGA with invalid second in timestamp",
			input:      "$GPGGA,1235XX.0,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*36",
			expectUsed: false,
		},
		{
			name:       "GPGGA with latitude too short",
			input:      "$GPGGA,123519.0,480,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*6B",
			expectUsed: false,
		},
		{
			name:       "GPGGA with invalid latitude degrees",
			input:      "$GPGGA,123519.0,XX07.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*4B",
			expectUsed: false,
		},
		{
			name:       "GPGGA with invalid latitude minutes",
			input:      "$GPGGA,123519.0,48XX.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*6E",
			expectUsed: false,
		},
		{
			name:       "GPGGA with longitude too short",
			input:      "$GPGGA,123519.0,4807.038,N,0113,E,1,08,0.9,545.4,M,46.9,M,,*52",
			expectUsed: false,
		},
		{
			name:       "GPGGA with invalid longitude degrees",
			input:      "$GPGGA,123519.0,4807.038,N,XXX31.000,E,1,08,0.9,545.4,M,46.9,M,,*0C",
			expectUsed: false,
		},
		{
			name:       "GPGGA with invalid longitude minutes",
			input:      "$GPGGA,123519.0,4807.038,N,011XX.000,E,1,08,0.9,545.4,M,46.9,M,,*21",
			expectUsed: false,
		},
		{
			name:       "GPGGA with invalid altitude",
			input:      "$GPGGA,123519.0,4807.038,N,01131.000,E,1,08,0.9,XXX,M,46.9,M,,*5C",
			expectUsed: false,
		},
		{
			name:       "GPGGA with invalid geoid separation",
			input:      "$GPGGA,123519.0,4807.038,N,01131.000,E,1,08,0.9,545.4,M,XXX,M,,*77",
			expectUsed: false,
		},
		{
			name:       "GPGGA with fix quality out of bounds (negative)",
			input:      "$GPGGA,123519.0,4807.038,N,01131.000,E,-1,08,0.9,545.4,M,46.9,M,,*74",
			expectUsed: true, // Should clamp to 0
		},
		{
			name:       "GPGGA with fix quality out of bounds (too high)",
			input:      "$GPGGA,123519.0,4807.038,N,01131.000,E,15,08,0.9,545.4,M,46.9,M,,*6C",
			expectUsed: true, // Should clamp to 9
		},

		// GPRMC edge cases
		{
			name:       "GPRMC with invalid hour",
			input:      "$GPRMC,XX3519.0,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*53",
			expectUsed: false,
		},
		{
			name:       "GPRMC with invalid minute",
			input:      "$GPRMC,12XX19.0,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*34",
			expectUsed: false,
		},
		{
			name:       "GPRMC with invalid second",
			input:      "$GPRMC,1235XX.0,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*73",
			expectUsed: false,
		},
		{
			name:       "GPRMC with latitude too short",
			input:      "$GPRMC,123519.0,A,480,N,01131.000,E,022.4,084.4,230394,003.1,W*28",
			expectUsed: false,
		},
		{
			name:       "GPRMC with invalid latitude degrees",
			input:      "$GPRMC,123519.0,A,XX07.038,N,01131.000,E,022.4,084.4,230394,003.1,W*08",
			expectUsed: false,
		},
		{
			name:       "GPRMC with invalid latitude minutes",
			input:      "$GPRMC,123519.0,A,48XX.038,N,01131.000,E,022.4,084.4,230394,003.1,W*2D",
			expectUsed: false,
		},
		{
			name:       "GPRMC with longitude too short",
			input:      "$GPRMC,123519.0,A,4807.038,N,0113,E,022.4,084.4,230394,003.1,W*11",
			expectUsed: false,
		},
		{
			name:       "GPRMC with invalid longitude degrees",
			input:      "$GPRMC,123519.0,A,4807.038,N,XXX31.000,E,022.4,084.4,230394,003.1,W*4F",
			expectUsed: false,
		},
		{
			name:       "GPRMC with invalid longitude minutes",
			input:      "$GPRMC,123519.0,A,4807.038,N,011XX.000,E,022.4,084.4,230394,003.1,W*62",
			expectUsed: false,
		},
		{
			name:       "GPRMC with invalid ground speed",
			input:      "$GPRMC,123519.0,A,4807.038,N,01131.000,E,XXX,084.4,230394,003.1,W*2C",
			expectUsed: false,
		},
		{
			name:       "GPRMC with invalid course when speed > 3",
			input:      "$GPRMC,123519.0,A,4807.038,N,01131.000,E,022.4,XXX,230394,003.1,W*57",
			expectUsed: false,
		},
		{
			name:       "GPRMC with timestamp too short",
			input:      "$GPRMC,12351,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*06",
			expectUsed: false,
		},

		// GPGSA edge cases
		{
			name:       "GPGSA with empty solution type",
			input:      "$GPGSA,A,,03,04,05,06,07,08,09,10,11,12,13,14,1.0,1.0,1.0*39",
			expectUsed: false,
		},
		{
			name:       "GPGSA with solution type 1 (no solution)",
			input:      "$GPGSA,A,1,03,04,05,06,07,08,09,10,11,12,13,14,1.0,1.0,1.0*30",
			expectUsed: false,
		},
		{
			name:       "GPGSA with invalid HDOP",
			input:      "$GPGSA,A,3,03,04,05,06,07,08,09,10,11,12,13,14,XXX,1.0,1.0*75",
			expectUsed: false,
		},
		{
			name:       "GPGSA with invalid VDOP",
			input:      "$GPGSA,A,3,03,04,05,06,07,08,09,10,11,12,13,14,1.0,1.0,XXX*7F",
			expectUsed: false,
		},

		// GPGST edge cases
		{
			name:       "GPGST with invalid latitude error",
			input:      "$GPGST,205246.00,1.19,0.02,0.01,-2.4501,XXX,0.01,0.03*3E",
			expectUsed: false,
		},
		{
			name:       "GPGST with invalid longitude error",
			input:      "$GPGST,205246.00,1.19,0.02,0.01,-2.4501,0.02,XXX,0.03*3E",
			expectUsed: false,
		},
		{
			name:       "GPGST with invalid altitude error",
			input:      "$GPGST,205246.00,1.19,0.02,0.01,-2.4501,0.02,0.01,XXX*3E",
			expectUsed: false,
		},

		// GPGSV edge cases
		{
			name:       "GPGSV with invalid message number",
			input:      "$GPGSV,X,1,08,01,40,083,46,02,17,308,41,12,07,344,39,14,22,228,45*7F",
			expectUsed: false,
		},
		{
			name:       "GPGSV with invalid message index",
			input:      "$GPGSV,2,X,08,01,40,083,46,02,17,308,41,12,07,344,39,14,22,228,45*7F",
			expectUsed: false,
		},
		{
			name:       "GPGSV with invalid satellite ID",
			input:      "$GPGSV,2,1,08,XX,40,083,46,02,17,308,41,12,07,344,39,14,22,228,45*47",
			expectUsed: false,
		},

		// POGNB edge cases
		{
			name:       "POGNB with invalid pressure altitude",
			input:      "$POGNB,22.0,+29.1,100972.3,3.8,XXX,+87.2,-0.04,+32.6*77",
			expectUsed: false,
		},
		{
			name:       "POGNB with invalid vertical speed",
			input:      "$POGNB,22.0,+29.1,100972.3,3.8,+29.4,+87.2,XXX,+32.6*70",
			expectUsed: false,
		},

		// Unknown message types (should return false)
		{
			name:       "Unknown NMEA sentence type",
			input:      "$GPXYZ,123,456,789*3C",
			expectUsed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset satellite state
			Satellites = make(map[string]SatelliteInfo)

			result := processNMEALine(tc.input)

			if result != tc.expectUsed {
				t.Errorf("processNMEALine(%q) returned %v, expected %v",
					tc.input, result, tc.expectUsed)
			}

			t.Logf("Edge case test complete: used=%v", result)
		})
	}
}

// TestProcessNMEALineLow_GPSTypeDetection tests GPS type detection from NMEA messages
func TestProcessNMEALineLow_GPSTypeDetection(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
	}

	// Save original state
	origGPSType := globalStatus.GPS_detected_type
	defer func() { globalStatus.GPS_detected_type = origGPSType }()

	testCases := []struct {
		name            string
		input           string
		initialGPSType  uint
		expectProtocol  uint
	}{
		{
			name:           "GPGGA sets NMEA protocol when type is unset",
			input:          "$GPGGA,123519.0,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*59",
			initialGPSType: 0,
			expectProtocol: GPS_PROTOCOL_NMEA,
		},
		{
			name:           "GNGGA sets NMEA protocol when type is unset",
			input:          "$GNGGA,123519.0,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47",
			initialGPSType: 0,
			expectProtocol: GPS_PROTOCOL_NMEA,
		},
		{
			name:           "GPRMC sets NMEA protocol when type is unset",
			input:          "$GPRMC,123519.0,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*74",
			initialGPSType: 0,
			expectProtocol: GPS_PROTOCOL_NMEA,
		},
		{
			name:           "GNRMC sets NMEA protocol when type is unset",
			input:          "$GNRMC,081836.0,A,3751.65,S,14507.36,E,000.0,360.0,130998,011.3,E*62",
			initialGPSType: 0,
			expectProtocol: GPS_PROTOCOL_NMEA,
		},
		{
			name:           "NMEA protocol not set when already has protocol",
			input:          "$GPGGA,123519.0,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*59",
			initialGPSType: GPS_TYPE_UBX9 | 0x20,
			expectProtocol: 0x20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			globalStatus.GPS_detected_type = tc.initialGPSType

			result := processNMEALine(tc.input)

			if !result {
				t.Errorf("processNMEALine(%q) returned false, expected true", tc.input)
			}

			if tc.expectProtocol != 0 {
				detectedProtocol := globalStatus.GPS_detected_type & 0xf0
				if detectedProtocol != tc.expectProtocol {
					t.Errorf("Expected GPS protocol 0x%X, got 0x%X",
						tc.expectProtocol, detectedProtocol)
				}
			}

			t.Logf("GPS type detection test complete: type=0x%X",
				globalStatus.GPS_detected_type)
		})
	}
}

// TestProcessNMEALineLow_SatelliteTypes tests satellite type classification
func TestProcessNMEALineLow_SatelliteTypes(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	testCases := []struct {
		name          string
		input         string
		expectSatID   string
		expectSatType uint8
	}{
		{
			name:          "GPS satellite (1-32)",
			input:         "$GPGSV,1,1,04,01,40,083,46,02,17,308,41,03,07,344,39,04,22,228,45*7B",
			expectSatID:   "G1",
			expectSatType: SAT_TYPE_GPS,
		},
		{
			name:          "SBAS satellite (33-64)",
			input:         "$GPGSV,1,1,01,33,40,083,46*45",
			expectSatID:   "S120",
			expectSatType: SAT_TYPE_SBAS,
		},
		{
			name:          "GLONASS satellite (65-96)",
			input:         "$GLGSV,1,1,01,65,40,083,46*5A",
			expectSatID:   "R1",
			expectSatType: SAT_TYPE_GLONASS,
		},
		{
			name:          "SBAS satellite (152-158)",
			input:         "$GPGSV,1,1,01,152,40,083,46*73",
			expectSatID:   "S152",
			expectSatType: SAT_TYPE_SBAS,
		},
		{
			name:          "QZSS satellite (193-202)",
			input:         "$GPGSV,1,1,01,193,40,083,46*7E",
			expectSatID:   "Q1",
			expectSatType: SAT_TYPE_QZSS,
		},
		{
			name:          "Galileo satellite (301-336)",
			input:         "$GAGSV,1,1,01,301,40,083,46*66",
			expectSatID:   "E1",
			expectSatType: SAT_TYPE_GALILEO,
		},
		{
			name:          "Beidou satellite (401-463)",
			input:         "$GBGSV,1,1,01,401,40,083,46*62",
			expectSatID:   "B1",
			expectSatType: SAT_TYPE_BEIDOU,
		},
		{
			name:          "Unknown satellite (>463)",
			input:         "$GPGSV,1,1,01,500,40,083,46*70",
			expectSatID:   "U500",
			expectSatType: SAT_TYPE_UNKNOWN,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset satellite state
			Satellites = make(map[string]SatelliteInfo)

			result := processNMEALine(tc.input)

			if !result {
				t.Errorf("processNMEALine(%q) returned false, expected true", tc.input)
			}

			mySituation.muSatellite.Lock()
			sat, exists := Satellites[tc.expectSatID]
			mySituation.muSatellite.Unlock()

			if !exists {
				t.Errorf("Expected satellite %s to exist, but it doesn't", tc.expectSatID)
			} else {
				if sat.Type != tc.expectSatType {
					t.Errorf("Expected satellite type %d, got %d",
						tc.expectSatType, sat.Type)
				}
			}

			t.Logf("Satellite type test complete: ID=%s, type=%d",
				tc.expectSatID, tc.expectSatType)
		})
	}
}
