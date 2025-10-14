package main

import (
	"bytes"
	"strings"
	"testing"
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
