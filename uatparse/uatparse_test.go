package uatparse

import (
	"strings"
	"testing"
)

// TestDlacDecode tests the dlac_decode function
func TestDlacDecode(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		data_len uint32
		expected string
	}{
		{
			name:     "Empty data",
			data:     []byte{},
			data_len: 0,
			expected: "",
		},
		{
			name:     "Single character - A",
			data:     []byte{0x04},
			data_len: 1,
			expected: "A",
		},
		{
			name:     "Single character - B",
			data:     []byte{0x08},
			data_len: 1,
			expected: "B",
		},
		{
			name:     "Multiple characters",
			data:     []byte{0x04, 0x00, 0x00, 0x00}, // Should decode to 5 characters due to step pattern
			data_len: 4,
			expected: "A\x03\x03\x03\x03",
		},
		{
			name:     "Space character",
			data:     []byte{0xFC}, // Index 63 in dlac_alpha is '?'
			data_len: 1,
			expected: "?",
		},
		{
			name:     "First character in alphabet (ETX)",
			data:     []byte{0x00},
			data_len: 1,
			expected: "\x03",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dlac_decode(tt.data, tt.data_len)
			if result != tt.expected {
				t.Errorf("dlac_decode() = %q (bytes: %v), expected %q (bytes: %v)",
					result, []byte(result), tt.expected, []byte(tt.expected))
			}
		})
	}
}

// TestDlacDecodeVariousInputs tests dlac_decode with various byte patterns
func TestDlacDecodeVariousInputs(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		len  uint32
	}{
		{
			name: "Two bytes",
			data: []byte{0xFF, 0xFF},
			len:  2,
		},
		{
			name: "Three bytes",
			data: []byte{0x10, 0x20, 0x30},
			len:  3,
		},
		{
			name: "Pattern with high bits",
			data: []byte{0xAA, 0x55, 0xAA, 0x55},
			len:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the function doesn't crash and returns something
			result := dlac_decode(tt.data, tt.len)
			// Result length depends on the step pattern, just verify no crash
			_ = result
		})
	}
}

// TestFormatDLACData tests the formatDLACData function
func TestFormatDLACData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "Single line with \\x1E separator",
			input:    "Line1\x1ELine2\x1ELine3",
			expected: []string{"Line1", "Line2", "Line3"},
		},
		{
			name:     "Single line with \\x03 separator",
			input:    "Line1\x03Line2\x03Line3",
			expected: []string{"Line1", "Line2", "Line3"},
		},
		{
			name:     "Mixed separators",
			input:    "Line1\x1ELine2\x03Line3",
			expected: []string{"Line1", "Line2", "Line3"},
		},
		{
			name:     "No separators",
			input:    "SingleLine",
			expected: []string{"SingleLine"},
		},
		{
			name:     "Trailing separator \\x1E",
			input:    "Line1\x1E",
			expected: []string{"Line1", ""},
		},
		{
			name:     "Leading separator \\x03",
			input:    "\x03Line1",
			expected: []string{"", "Line1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDLACData(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("formatDLACData() returned %d elements, expected %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("formatDLACData()[%d] = %q, expected %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestAirmetParseDate tests the airmetParseDate function
func TestAirmetParseDate(t *testing.T) {
	tests := []struct {
		name             string
		data             []byte
		date_time_format uint8
		expected         string
	}{
		{
			name:             "No date/time (format 0)",
			data:             []byte{},
			date_time_format: 0,
			expected:         "",
		},
		{
			name:             "Month, Day, Hours, Minutes (format 1)",
			data:             []byte{12, 25, 14, 30},
			date_time_format: 1,
			expected:         "12-25 14:30",
		},
		{
			name:             "Day, Hours, Minutes (format 2)",
			data:             []byte{15, 9, 45},
			date_time_format: 2,
			expected:         "15 09:45",
		},
		{
			name:             "Hours, Minutes (format 3)",
			data:             []byte{23, 59},
			date_time_format: 3,
			expected:         "23:59",
		},
		{
			name:             "Invalid format returns empty",
			data:             []byte{1, 2, 3, 4},
			date_time_format: 99,
			expected:         "",
		},
		{
			name:             "Format 1 with zeros",
			data:             []byte{0, 0, 0, 0},
			date_time_format: 1,
			expected:         "00-00 00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := airmetParseDate(tt.data, tt.date_time_format)
			if result != tt.expected {
				t.Errorf("airmetParseDate() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestAirmetLatLng tests the airmetLatLng function
func TestAirmetLatLng(t *testing.T) {
	tests := []struct {
		name        string
		lat_raw     int32
		lng_raw     int32
		alt         bool
		expectedLat float64
		expectedLng float64
	}{
		{
			name:        "Zero coordinates, normal mode",
			lat_raw:     0,
			lng_raw:     0,
			alt:         false,
			expectedLat: 0.0,
			expectedLng: 0.0,
		},
		{
			name:        "Positive coordinates, normal mode",
			lat_raw:     100000,
			lng_raw:     200000,
			alt:         false,
			expectedLat: 68.7,
			expectedLng: 137.4,
		},
		{
			name:        "Coordinates requiring wrapping, normal mode",
			lat_raw:     200000,
			lng_raw:     300000,
			alt:         false,
			expectedLat: -42.6,  // 137.4 - 180
			expectedLng: -153.9, // 206.1 - 360
		},
		{
			name:        "Zero coordinates, alt mode",
			lat_raw:     0,
			lng_raw:     0,
			alt:         true,
			expectedLat: 0.0,
			expectedLng: 0.0,
		},
		{
			name:        "Positive coordinates, alt mode",
			lat_raw:     50000,
			lng_raw:     100000,
			alt:         true,
			expectedLat: 68.65,
			expectedLng: 137.3,
		},
		{
			name:        "Large coordinates requiring wrapping, alt mode",
			lat_raw:     100000,
			lng_raw:     150000,
			alt:         true,
			expectedLat: -42.7,  // 137.3 - 180
			expectedLng: -154.05, // 205.95 - 360
		},
		{
			name:        "Negative raw values",
			lat_raw:     -10000,
			lng_raw:     -20000,
			alt:         false,
			expectedLat: -6.87,
			expectedLng: -13.74,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng := airmetLatLng(tt.lat_raw, tt.lng_raw, tt.alt)
			// Use approximate comparison due to floating point precision
			if !almostEqual(lat, tt.expectedLat, 0.1) {
				t.Errorf("airmetLatLng() lat = %f, expected %f", lat, tt.expectedLat)
			}
			if !almostEqual(lng, tt.expectedLng, 0.1) {
				t.Errorf("airmetLatLng() lng = %f, expected %f", lng, tt.expectedLng)
			}
		})
	}
}

// TestNew tests the New function for parsing UAT messages
func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
			errorMsg:    "Invalid format",
		},
		{
			name:        "Missing semicolon",
			input:       "+AABBCCDD",
			expectError: true,
			errorMsg:    "Invalid format",
		},
		{
			name:        "Downlink message (starts with -)",
			input:       "-" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=1;ss=10",
			expectError: true,
			errorMsg:    "expecting uplink frame",
		},
		{
			name:        "Invalid character (not + or -)",
			input:       "X" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=1",
			expectError: true,
			errorMsg:    "expecting uplink frame",
		},
		{
			name:        "Short message gets padded",
			input:       "+AABBCC;rs=1",
			expectError: false, // New() pads short messages with zeros
		},
		{
			name:        "Valid uplink message with signal strength",
			input:       "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5;ss=20",
			expectError: false,
		},
		{
			name:        "Valid uplink message without optional fields",
			input:       "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";",
			expectError: false,
		},
		{
			name:        "Valid uplink message with partial metadata",
			input:       "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";ss=15",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := New(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("New() expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("New() error = %q, expected error containing %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("New() unexpected error = %v", err)
				}
				if msg == nil {
					t.Errorf("New() returned nil message")
				}
			}
		})
	}
}

// TestNewSignalStrengthParsing tests signal strength and RS error parsing
func TestNewSignalStrengthParsing(t *testing.T) {
	// Valid uplink message with signal strength
	input := "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=3;ss=25"
	msg, err := New(input)

	if err != nil {
		t.Fatalf("New() unexpected error = %v", err)
	}

	if msg.RS_Err != 3 {
		t.Errorf("New() RS_Err = %d, expected 3", msg.RS_Err)
	}

	if msg.SignalStrength != 25 {
		t.Errorf("New() SignalStrength = %d, expected 25", msg.SignalStrength)
	}
}

// TestNewDefaultValues tests that unspecified values get default -1
func TestNewDefaultValues(t *testing.T) {
	input := "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";"
	msg, err := New(input)

	if err != nil {
		t.Fatalf("New() unexpected error = %v", err)
	}

	if msg.RS_Err != -1 {
		t.Errorf("New() RS_Err = %d, expected -1", msg.RS_Err)
	}

	if msg.SignalStrength != -1 {
		t.Errorf("New() SignalStrength = %d, expected -1", msg.SignalStrength)
	}
}

// TestGetTextReportsNoDecoded tests GetTextReports when not yet decoded
func TestGetTextReportsNoDecoded(t *testing.T) {
	msg := &UATMsg{
		decoded: false,
		msg:     make([]byte, UPLINK_FRAME_DATA_BYTES),
	}

	// This should trigger DecodeUplink, which will fail due to invalid data
	// but should not panic
	reports, err := msg.GetTextReports()

	if err != nil {
		// Expected to fail with invalid test data
		if len(reports) != 0 {
			t.Errorf("GetTextReports() with decode error should return empty array, got %d reports", len(reports))
		}
	}
}

// TestGetTextReportsAlreadyDecoded tests GetTextReports when already decoded
func TestGetTextReportsAlreadyDecoded(t *testing.T) {
	msg := &UATMsg{
		decoded: true,
		Frames: []*UATFrame{
			{
				Text_data: []string{"Text1", "Text2", ""},
			},
			{
				Text_data: []string{"Text3"},
			},
		},
	}

	reports, err := msg.GetTextReports()

	if err != nil {
		t.Fatalf("GetTextReports() unexpected error = %v", err)
	}

	// Should skip empty strings
	expected := []string{"Text1", "Text2", "Text3"}
	if len(reports) != len(expected) {
		t.Errorf("GetTextReports() returned %d reports, expected %d", len(reports), len(expected))
	}

	for i, exp := range expected {
		if i >= len(reports) || reports[i] != exp {
			t.Errorf("GetTextReports()[%d] = %q, expected %q", i, reports[i], exp)
		}
	}
}

// Helper function for floating point comparison
func almostEqual(a, b, epsilon float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}
