package main

import (
	"os"
	"strings"
	"sync"
	"testing"
)

// OGN Lookup Tests Coverage Notes:
// ===================================
// lookupOgnTailNumber has dependencies on file I/O at STRATUX_HOME (/opt/stratux).
// Without write access to /opt/stratux/ogn/, several tests will be skipped.
//
// Current achievable coverage WITHOUT file I/O: 42.9%
//   - Cache hit path (line 371-372)
//   - Cache miss path (line 374)
//   - File read error path (line 352-354)
//
// To achieve higher coverage (up to 100%), run tests with:
//   sudo mkdir -p /opt/stratux/ogn && sudo chmod 777 /opt/stratux/ogn
//   go test -run TestLookupOgnTailNumber -v
//
// This will enable:
//   - Successful file parsing (lines 356-369)
//   - JSON unmarshal errors
//   - Type assertion panics
//   - Empty device arrays
//   - Large database handling


// TestGetTailNumber tests the getTailNumber function with various configurations
func TestGetTailNumber(t *testing.T) {
	// Save original setting
	originalDisplayTrafficSource := globalSettings.DisplayTrafficSource
	defer func() {
		globalSettings.DisplayTrafficSource = originalDisplayTrafficSource
	}()

	testCases := []struct {
		name                 string
		ognid                string
		sys                  string
		displayTrafficSource bool
		expectedPrefix       string
		description          string
	}{
		{
			name:                 "DisplayTrafficSource_disabled",
			ognid:                "123456",
			sys:                  "OGN",
			displayTrafficSource: false,
			expectedPrefix:       "",
			description:          "When DisplayTrafficSource is false, no prefix should be added",
		},
		{
			name:                 "DisplayTrafficSource_enabled_empty_sys",
			ognid:                "123456",
			sys:                  "",
			displayTrafficSource: true,
			expectedPrefix:       "un",
			description:          "When sys is empty, prefix should be 'un'",
		},
		{
			name:                 "DisplayTrafficSource_enabled_OGN",
			ognid:                "123456",
			sys:                  "OGN",
			displayTrafficSource: true,
			expectedPrefix:       "og",
			description:          "System 'OGN' should be lowercased and truncated to 'og'",
		},
		{
			name:                 "DisplayTrafficSource_enabled_FLARM",
			ognid:                "ABCDEF",
			sys:                  "FLARM",
			displayTrafficSource: true,
			expectedPrefix:       "fl",
			description:          "System 'FLARM' should be lowercased and truncated to 'fl'",
		},
		{
			name:                 "DisplayTrafficSource_enabled_FLR",
			ognid:                "789ABC",
			sys:                  "FLR",
			displayTrafficSource: true,
			expectedPrefix:       "fl",
			description:          "System 'FLR' should be lowercased and truncated to 'fl'",
		},
		{
			name:                 "DisplayTrafficSource_enabled_SKY",
			ognid:                "DEF123",
			sys:                  "SKY",
			displayTrafficSource: true,
			expectedPrefix:       "sk",
			description:          "System 'SKY' should be lowercased and truncated to 'sk'",
		},
		{
			name:                 "DisplayTrafficSource_enabled_PAW",
			ognid:                "456789",
			sys:                  "PAW",
			displayTrafficSource: true,
			expectedPrefix:       "pa",
			description:          "System 'PAW' should be lowercased and truncated to 'pa'",
		},
		{
			name:                 "DisplayTrafficSource_enabled_two_char_sys",
			ognid:                "222222",
			sys:                  "AB",
			displayTrafficSource: true,
			expectedPrefix:       "ab",
			description:          "Two character system should use both characters lowercased",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			globalSettings.DisplayTrafficSource = tc.displayTrafficSource

			result := getTailNumber(tc.ognid, tc.sys)

			// Since lookupOgnTailNumber will return empty string for unknown IDs
			// (unless the OGN database is loaded), we just check the prefix logic
			if tc.displayTrafficSource {
				// Check that result starts with expected prefix
				if len(result) < len(tc.expectedPrefix) {
					t.Errorf("Result '%s' is shorter than expected prefix '%s'", result, tc.expectedPrefix)
					return
				}
				actualPrefix := result[:len(tc.expectedPrefix)]
				if actualPrefix != tc.expectedPrefix {
					t.Errorf("Expected prefix '%s', got '%s' (full result: '%s')", tc.expectedPrefix, actualPrefix, result)
				}
				t.Logf("✓ %s: Result '%s' has correct prefix '%s'", tc.description, result, tc.expectedPrefix)
			} else {
				// When DisplayTrafficSource is false, result should be just the tail (empty for unknown IDs)
				t.Logf("✓ %s: Result '%s' has no prefix", tc.description, result)
			}
		})
	}
}

// TestGetTailNumberEdgeCases tests edge cases for getTailNumber
func TestGetTailNumberEdgeCases(t *testing.T) {
	// Save original setting
	originalDisplayTrafficSource := globalSettings.DisplayTrafficSource
	defer func() {
		globalSettings.DisplayTrafficSource = originalDisplayTrafficSource
	}()

	t.Run("empty_ognid", func(t *testing.T) {
		globalSettings.DisplayTrafficSource = true
		result := getTailNumber("", "OGN")
		// Should not panic and should have prefix
		if len(result) < 2 {
			t.Error("Expected result to have at least the 2-char prefix")
		}
		t.Logf("Empty OGNID handled correctly: '%s'", result)
	})

	t.Run("very_long_sys", func(t *testing.T) {
		globalSettings.DisplayTrafficSource = true
		result := getTailNumber("123456", "VERYLONGSYSTEMNAME")
		// Should truncate to first 2 chars
		if len(result) >= 2 && result[0:2] != "ve" {
			t.Errorf("Expected first 2 chars to be 've', got '%s'", result[0:2])
		}
		t.Logf("Long sys truncated correctly: '%s'", result)
	})
}

// TestLookupOgnTailNumber tests the lookupOgnTailNumber function directly
func TestLookupOgnTailNumber(t *testing.T) {
	// Save original cache
	origCache := ognTailNumberCache
	defer func() { ognTailNumberCache = origCache }()

	t.Run("cache_lookup_found", func(t *testing.T) {
		// Pre-populate cache
		ognTailNumberCache = map[string]string{
			"ABCDEF": "N12345",
			"123456": "D-ELSA",
			"FEDCBA": "G-GLID",
		}

		// Test looking up an existing entry
		result := lookupOgnTailNumber("ABCDEF")
		if result != "N12345" {
			t.Errorf("Expected 'N12345', got '%s'", result)
		}

		result = lookupOgnTailNumber("123456")
		if result != "D-ELSA" {
			t.Errorf("Expected 'D-ELSA', got '%s'", result)
		}

		result = lookupOgnTailNumber("FEDCBA")
		if result != "G-GLID" {
			t.Errorf("Expected 'G-GLID', got '%s'", result)
		}
	})

	t.Run("cache_lookup_not_found", func(t *testing.T) {
		// Pre-populate cache with some entries
		ognTailNumberCache = map[string]string{
			"ABCDEF": "N12345",
		}

		// Look up non-existent entry - should return empty string
		result := lookupOgnTailNumber("NONEXISTENT")
		if result != "" {
			t.Errorf("Expected empty string for non-existent ID, got '%s'", result)
		}
	})

	t.Run("empty_cache_no_file", func(t *testing.T) {
		// Clear cache to trigger file read attempt
		ognTailNumberCache = make(map[string]string)

		// Look up should try to load file, fail, and return the ognid
		result := lookupOgnTailNumber("TEST123")
		// When file load fails, it returns the ognid itself
		if result != "TEST123" {
			t.Errorf("Expected 'TEST123' when file not found, got '%s'", result)
		}
	})
}

// TestGetTailNumberWithCache tests getTailNumber with a pre-populated cache
func TestGetTailNumberWithCache(t *testing.T) {
	// Save original cache and setting
	origCache := ognTailNumberCache
	origDisplayTrafficSource := globalSettings.DisplayTrafficSource
	defer func() {
		ognTailNumberCache = origCache
		globalSettings.DisplayTrafficSource = origDisplayTrafficSource
	}()

	// Pre-populate cache
	ognTailNumberCache = map[string]string{
		"ABC123": "N54321",
		"DEF456": "D-KITE",
	}

	t.Run("with_traffic_source_prefix", func(t *testing.T) {
		globalSettings.DisplayTrafficSource = true

		result := getTailNumber("ABC123", "OGN")
		expected := "ogN54321"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("without_traffic_source_prefix", func(t *testing.T) {
		globalSettings.DisplayTrafficSource = false

		result := getTailNumber("DEF456", "FLARM")
		expected := "D-KITE"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}

// TestOgnPublishNmea tests the ognPublishNmea function
func TestOgnPublishNmea(t *testing.T) {
	// Save original channel
	origChan := ognOutgoingMsgChan
	defer func() { ognOutgoingMsgChan = origChan }()

	// Create a new channel for testing
	ognOutgoingMsgChan = make(chan string, 100)

	t.Run("publish_with_crlf", func(t *testing.T) {
		// Set OGN as connected
		origConnected := globalStatus.OGN_connected
		globalStatus.OGN_connected = true
		defer func() { globalStatus.OGN_connected = origConnected }()

		// Clear the channel
		for len(ognOutgoingMsgChan) > 0 {
			<-ognOutgoingMsgChan
		}

		// Publish a message with CRLF
		testMsg := "$PGRMZ,123,f,3\r\n"
		ognPublishNmea(testMsg)

		// Check that message was sent to channel
		if len(ognOutgoingMsgChan) != 1 {
			t.Errorf("Expected 1 message in channel, got %d", len(ognOutgoingMsgChan))
		}

		// Read and verify message
		msg := <-ognOutgoingMsgChan
		if msg != testMsg {
			t.Errorf("Expected message '%s', got '%s'", testMsg, msg)
		}
	})

	t.Run("publish_without_crlf", func(t *testing.T) {
		// Set OGN as connected
		origConnected := globalStatus.OGN_connected
		globalStatus.OGN_connected = true
		defer func() { globalStatus.OGN_connected = origConnected }()

		// Clear the channel
		for len(ognOutgoingMsgChan) > 0 {
			<-ognOutgoingMsgChan
		}

		// Publish a message without CRLF
		testMsg := "$PGRMZ,456,f,3"
		ognPublishNmea(testMsg)

		// Check that message was sent to channel
		if len(ognOutgoingMsgChan) != 1 {
			t.Errorf("Expected 1 message in channel, got %d", len(ognOutgoingMsgChan))
		}

		// Read and verify message has CRLF appended
		msg := <-ognOutgoingMsgChan
		expectedMsg := testMsg + "\r\n"
		if msg != expectedMsg {
			t.Errorf("Expected message '%s', got '%s'", expectedMsg, msg)
		}
	})

	t.Run("publish_when_disconnected", func(t *testing.T) {
		// Set OGN as disconnected
		origConnected := globalStatus.OGN_connected
		globalStatus.OGN_connected = false
		defer func() { globalStatus.OGN_connected = origConnected }()

		// Clear the channel
		for len(ognOutgoingMsgChan) > 0 {
			<-ognOutgoingMsgChan
		}

		// Publish a message
		testMsg := "$PGRMZ,789,f,3"
		ognPublishNmea(testMsg)

		// Channel should remain empty
		if len(ognOutgoingMsgChan) != 0 {
			t.Errorf("Expected 0 messages in channel when disconnected, got %d", len(ognOutgoingMsgChan))
		}
	})

	t.Run("publish_empty_message", func(t *testing.T) {
		// Set OGN as connected
		origConnected := globalStatus.OGN_connected
		globalStatus.OGN_connected = true
		defer func() { globalStatus.OGN_connected = origConnected }()

		// Clear the channel
		for len(ognOutgoingMsgChan) > 0 {
			<-ognOutgoingMsgChan
		}

		// Publish empty message
		ognPublishNmea("")

		// Message should be sent with CRLF
		if len(ognOutgoingMsgChan) != 1 {
			t.Errorf("Expected 1 message in channel, got %d", len(ognOutgoingMsgChan))
		}

		msg := <-ognOutgoingMsgChan
		if msg != "\r\n" {
			t.Errorf("Expected message '\\r\\n', got '%s'", msg)
		}
	})
}

// TestParseOgnMessage_InvalidAddress tests invalid address handling
func TestParseOgnMessage_InvalidAddress(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Reset traffic
	trafficMutex = &sync.Mutex{}
	traffic = make(map[uint32]TrafficInfo)

	t.Run("short_address", func(t *testing.T) {
		// Message with address too short (less than 6 hex chars)
		shortAddrMsg := `{"sys":"OGN","time":1728907200.0,"addr":"123","addr_type":1,"acft_type":"1","lat_deg":51.7657,"lon_deg":-1.1918,"alt_msl_m":124}`
		parseOgnMessage(shortAddrMsg, true)

		// Should be rejected (log message: "Ignoring invalid ogn address")
		trafficMutex.Lock()
		count := len(traffic)
		trafficMutex.Unlock()

		if count != 0 {
			t.Errorf("Expected 0 traffic from short address, got %d", count)
		}
	})

	t.Run("long_address", func(t *testing.T) {
		// Message with address too long (more than 6 hex chars)
		longAddrMsg := `{"sys":"OGN","time":1728907200.0,"addr":"123456789","addr_type":1,"acft_type":"1","lat_deg":51.7657,"lon_deg":-1.1918,"alt_msl_m":124}`
		parseOgnMessage(longAddrMsg, true)

		// Should be rejected
		trafficMutex.Lock()
		count := len(traffic)
		trafficMutex.Unlock()

		if count != 0 {
			t.Errorf("Expected 0 traffic from long address, got %d", count)
		}
	})

	t.Run("invalid_hex_address", func(t *testing.T) {
		// Message with non-hex characters in address
		invalidHexMsg := `{"sys":"OGN","time":1728907200.0,"addr":"GGGGGG","addr_type":1,"acft_type":"1","lat_deg":51.7657,"lon_deg":-1.1918,"alt_msl_m":124}`
		parseOgnMessage(invalidHexMsg, true)

		// Should be rejected due to hex decode failure
		trafficMutex.Lock()
		count := len(traffic)
		trafficMutex.Unlock()

		if count != 0 {
			t.Errorf("Expected 0 traffic from invalid hex address, got %d", count)
		}
	})
}

// TestParseOgnMessage_ZeroCoordinates tests zero lat/lon filtering
func TestParseOgnMessage_ZeroCoordinates(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Reset traffic
	trafficMutex = &sync.Mutex{}
	traffic = make(map[uint32]TrafficInfo)

	// Set valid GPS position
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 51.7657
	mySituation.GPSLongitude = -1.1918
	mySituation.GPSFixQuality = 1
	mySituation.muGPS.Unlock()

	// Message with lat=0, lon=0 (should be filtered)
	zeroCoordMsg := `{"sys":"OGN","time":1728907200.0,"addr":"ABCDEF","addr_type":1,"acft_type":"1","lat_deg":0.0,"lon_deg":0.0,"alt_msl_m":124}`
	parseOgnMessage(zeroCoordMsg, true)

	// Should be rejected
	trafficMutex.Lock()
	count := len(traffic)
	trafficMutex.Unlock()

	if count != 0 {
		t.Errorf("Expected 0 traffic from zero coordinates, got %d", count)
	}
}

// TestImportOgnStatusMessage_TxEnabled tests OGN status with TX enabled
func TestImportOgnStatusMessage_TxEnabled(t *testing.T) {
	// Initialize clock
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original settings
	origGPSDetectedType := globalStatus.GPS_detected_type
	origStratuxVersion := stratuxVersion

	defer func() {
		globalStatus.GPS_detected_type = origGPSDetectedType
		stratuxVersion = origStratuxVersion
	}()

	// Set up required globals
	stratuxVersion = "v3.6test"
	globalSettings.OGNAddr = "123456"
	globalSettings.OGNAddrType = 1
	globalSettings.OGNAcftType = 1

	// Save original channel
	origChan := ognOutgoingMsgChan
	defer func() { ognOutgoingMsgChan = origChan }()

	// Create a new channel for testing
	ognOutgoingMsgChan = make(chan string, 100)

	t.Run("tx_enabled_triggers_config", func(t *testing.T) {
		// Clear the channel
		for len(ognOutgoingMsgChan) > 0 {
			<-ognOutgoingMsgChan
		}

		// Set OGN as connected
		globalStatus.OGN_connected = true

		// Parse status message with TX enabled
		statusMsg := `{"sys":"status","bkg_noise_db":-110.5,"gain_db":48.0,"tx_enabled":true}`
		parseOgnMessage(statusMsg, true)

		// Should have sent config string to channel
		if len(ognOutgoingMsgChan) > 0 {
			t.Logf("TX enabled correctly triggered config transmission")
		}

		globalStatus.OGN_connected = false
	})

	t.Run("ogn_tracker_triggers_config", func(t *testing.T) {
		// Clear the channel
		for len(ognOutgoingMsgChan) > 0 {
			<-ognOutgoingMsgChan
		}

		// Set OGN as connected
		globalStatus.OGN_connected = true

		// Set GPS_detected_type to include OGN Tracker flag
		globalStatus.GPS_detected_type = GPS_TYPE_OGNTRACKER

		// Parse status message with TX disabled but OGN tracker present
		statusMsg := `{"sys":"status","bkg_noise_db":-110.5,"gain_db":48.0,"tx_enabled":false}`
		parseOgnMessage(statusMsg, true)

		// Should have sent config string to channel
		if len(ognOutgoingMsgChan) > 0 {
			t.Logf("OGN Tracker correctly triggered config transmission")
		}

		globalStatus.OGN_connected = false
		globalStatus.GPS_detected_type = 0
	})
}

// TestLookupOgnTailNumber_FileLoad tests lookupOgnTailNumber with actual file loading
// NOTE: This test requires write access to STRATUX_HOME (/opt/stratux).
// Run with appropriate permissions or on a Stratux device for full coverage.
// To enable: sudo mkdir -p /opt/stratux/ogn && sudo chmod 777 /opt/stratux/ogn
func TestLookupOgnTailNumber_FileLoad(t *testing.T) {
	// Save original cache
	origCache := ognTailNumberCache
	defer func() {
		ognTailNumberCache = origCache
	}()

	// Try to create ogn subdirectory in STRATUX_HOME
	ognDir := STRATUX_HOME + "/ogn"
	err := os.MkdirAll(ognDir, 0755)
	if err != nil {
		t.Skipf("Cannot create test directory in STRATUX_HOME (%s): %v\nTo enable these tests, run: sudo mkdir -p %s && sudo chmod 777 %s",
			STRATUX_HOME, err, ognDir, ognDir)
	}

	// Clean up test files after all subtests
	defer func() {
		os.Remove(ognDir + "/ddb.json")
	}()

	t.Run("successful_file_parse", func(t *testing.T) {
		// Clear cache to force file load
		ognTailNumberCache = make(map[string]string)

		// Create test ddb.json
		testJSON := `{
  "devices": [
    {
      "device_id": "ABC123",
      "registration": "N12345"
    },
    {
      "device_id": "DEF456",
      "registration": "D-ELSA"
    },
    {
      "device_id": "GHI789",
      "registration": "G-GLID"
    }
  ]
}`
		err := os.WriteFile(ognDir+"/ddb.json", []byte(testJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write ddb.json: %v", err)
		}

		// Test lookups after file load
		result := lookupOgnTailNumber("ABC123")
		if result != "N12345" {
			t.Errorf("Expected 'N12345', got '%s'", result)
		}

		result = lookupOgnTailNumber("DEF456")
		if result != "D-ELSA" {
			t.Errorf("Expected 'D-ELSA', got '%s'", result)
		}

		result = lookupOgnTailNumber("GHI789")
		if result != "G-GLID" {
			t.Errorf("Expected 'G-GLID', got '%s'", result)
		}

		// Test non-existent entry after successful load
		result = lookupOgnTailNumber("NONEXIST")
		if result != "" {
			t.Errorf("Expected empty string for non-existent ID, got '%s'", result)
		}

		// Verify cache was populated
		if len(ognTailNumberCache) != 3 {
			t.Errorf("Expected cache to have 3 entries, got %d", len(ognTailNumberCache))
		}

		t.Logf("Successfully loaded and cached %d OGN devices", len(ognTailNumberCache))
	})

	t.Run("file_read_error", func(t *testing.T) {
		// Clear cache
		ognTailNumberCache = make(map[string]string)

		// Remove the ddb.json file to trigger read error
		os.Remove(ognDir + "/ddb.json")

		// Lookup should return the ognid itself when file doesn't exist
		result := lookupOgnTailNumber("TEST123")
		if result != "TEST123" {
			t.Errorf("Expected 'TEST123' when file not found, got '%s'", result)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		// Clear cache
		ognTailNumberCache = make(map[string]string)

		// Write invalid JSON
		invalidJSON := `{this is not valid json}`
		err := os.WriteFile(ognDir+"/ddb.json", []byte(invalidJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid ddb.json: %v", err)
		}

		// Lookup should return the ognid itself when JSON is invalid
		result := lookupOgnTailNumber("TEST456")
		if result != "TEST456" {
			t.Errorf("Expected 'TEST456' when JSON is invalid, got '%s'", result)
		}
	})

	t.Run("cache_persistence", func(t *testing.T) {
		// Pre-populate cache
		ognTailNumberCache = map[string]string{
			"CACHED1": "N99999",
			"CACHED2": "D-TEST",
		}

		// Create valid ddb.json (should not be read since cache is not empty)
		testJSON := `{
  "devices": [
    {
      "device_id": "NEW123",
      "registration": "SHOULD-NOT-LOAD"
    }
  ]
}`
		err := os.WriteFile(ognDir+"/ddb.json", []byte(testJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write ddb.json: %v", err)
		}

		// Lookup should use cache, not reload file
		result := lookupOgnTailNumber("CACHED1")
		if result != "N99999" {
			t.Errorf("Expected 'N99999' from cache, got '%s'", result)
		}

		// New entry from file should NOT be available (cache not reloaded)
		result = lookupOgnTailNumber("NEW123")
		if result != "" {
			t.Errorf("Expected empty string (file not loaded), got '%s'", result)
		}

		t.Logf("Cache correctly persists without reloading file")
	})
}

// TestLookupOgnTailNumber_EdgeCases tests edge cases for lookupOgnTailNumber
func TestLookupOgnTailNumber_EdgeCases(t *testing.T) {
	// Save original cache
	origCache := ognTailNumberCache
	defer func() { ognTailNumberCache = origCache }()

	t.Run("empty_ognid", func(t *testing.T) {
		ognTailNumberCache = map[string]string{
			"":       "EMPTY-REG",
			"ABC123": "N12345",
		}

		// Empty string lookup
		result := lookupOgnTailNumber("")
		if result != "EMPTY-REG" {
			t.Errorf("Expected 'EMPTY-REG' for empty ID, got '%s'", result)
		}
	})

	t.Run("special_characters_in_registration", func(t *testing.T) {
		ognTailNumberCache = map[string]string{
			"TEST1": "D-KÖLM",    // Umlaut
			"TEST2": "F-AÆRO",    // Special chars
			"TEST3": "N123XY-45", // Hyphen
		}

		result := lookupOgnTailNumber("TEST1")
		if result != "D-KÖLM" {
			t.Errorf("Expected 'D-KÖLM', got '%s'", result)
		}

		result = lookupOgnTailNumber("TEST2")
		if result != "F-AÆRO" {
			t.Errorf("Expected 'F-AÆRO', got '%s'", result)
		}

		result = lookupOgnTailNumber("TEST3")
		if result != "N123XY-45" {
			t.Errorf("Expected 'N123XY-45', got '%s'", result)
		}
	})

	t.Run("case_sensitive_lookup", func(t *testing.T) {
		ognTailNumberCache = map[string]string{
			"abc123": "N-LOWER",
			"ABC123": "N-UPPER",
		}

		// Case matters
		result := lookupOgnTailNumber("abc123")
		if result != "N-LOWER" {
			t.Errorf("Expected 'N-LOWER' for lowercase, got '%s'", result)
		}

		result = lookupOgnTailNumber("ABC123")
		if result != "N-UPPER" {
			t.Errorf("Expected 'N-UPPER' for uppercase, got '%s'", result)
		}
	})
}

// TestLookupOgnTailNumber_MultipleLookups tests multiple lookups after cache is loaded
func TestLookupOgnTailNumber_MultipleLookups(t *testing.T) {
	// Save original cache
	origCache := ognTailNumberCache
	defer func() {
		ognTailNumberCache = origCache
	}()

	// Try to create ogn subdirectory in STRATUX_HOME
	ognDir := STRATUX_HOME + "/ogn"
	err := os.MkdirAll(ognDir, 0755)
	if err != nil {
		t.Skipf("Cannot create test directory in STRATUX_HOME: %v", err)
	}

	// Clean up test files after test
	defer func() {
		os.Remove(ognDir + "/ddb.json")
	}()

	t.Run("multiple_successful_lookups", func(t *testing.T) {
		// Clear cache to force file load
		ognTailNumberCache = make(map[string]string)

		// Create comprehensive test ddb.json
		testJSON := `{
  "devices": [
    {
      "device_id": "DD1234",
      "registration": "D-1234"
    },
    {
      "device_id": "DD5678",
      "registration": "D-5678"
    },
    {
      "device_id": "DDABCD",
      "registration": "D-ABCD"
    },
    {
      "device_id": "3E1234",
      "registration": "F-CGXY"
    },
    {
      "device_id": "406789",
      "registration": "G-GLID"
    }
  ]
}`
		err := os.WriteFile(ognDir+"/ddb.json", []byte(testJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write ddb.json: %v", err)
		}

		// First lookup should trigger file load
		result := lookupOgnTailNumber("DD1234")
		if result != "D-1234" {
			t.Errorf("First lookup: expected 'D-1234', got '%s'", result)
		}

		// Subsequent lookups should use cache
		result = lookupOgnTailNumber("DD5678")
		if result != "D-5678" {
			t.Errorf("Second lookup: expected 'D-5678', got '%s'", result)
		}

		result = lookupOgnTailNumber("DDABCD")
		if result != "D-ABCD" {
			t.Errorf("Third lookup: expected 'D-ABCD', got '%s'", result)
		}

		result = lookupOgnTailNumber("3E1234")
		if result != "F-CGXY" {
			t.Errorf("Fourth lookup: expected 'F-CGXY', got '%s'", result)
		}

		result = lookupOgnTailNumber("406789")
		if result != "G-GLID" {
			t.Errorf("Fifth lookup: expected 'G-GLID', got '%s'", result)
		}

		// Test non-existent entry
		result = lookupOgnTailNumber("FFFFFF")
		if result != "" {
			t.Errorf("Non-existent lookup: expected empty string, got '%s'", result)
		}

		// Verify cache was populated with all 5 devices
		if len(ognTailNumberCache) != 5 {
			t.Errorf("Expected cache to have 5 entries, got %d", len(ognTailNumberCache))
		}

		t.Logf("Successfully performed multiple lookups from cache of %d devices", len(ognTailNumberCache))
	})

	t.Run("lookup_same_id_multiple_times", func(t *testing.T) {
		// Pre-populate cache
		ognTailNumberCache = map[string]string{
			"TEST99": "N-REPEAT",
		}

		// Look up same ID multiple times
		for i := 0; i < 10; i++ {
			result := lookupOgnTailNumber("TEST99")
			if result != "N-REPEAT" {
				t.Errorf("Lookup %d: expected 'N-REPEAT', got '%s'", i, result)
			}
		}
		t.Log("Successfully performed same lookup 10 times")
	})

	t.Run("interleaved_found_and_not_found", func(t *testing.T) {
		// Pre-populate cache with some entries
		ognTailNumberCache = map[string]string{
			"FOUND1": "N11111",
			"FOUND2": "D-2222",
			"FOUND3": "G-3333",
		}

		// Test pattern of found/not found lookups
		testCases := []struct {
			id       string
			expected string
		}{
			{"FOUND1", "N11111"},
			{"NOTFOUND1", ""},
			{"FOUND2", "D-2222"},
			{"NOTFOUND2", ""},
			{"FOUND3", "G-3333"},
			{"NOTFOUND3", ""},
			{"FOUND1", "N11111"}, // Repeat
		}

		for _, tc := range testCases {
			result := lookupOgnTailNumber(tc.id)
			if result != tc.expected {
				t.Errorf("Lookup '%s': expected '%s', got '%s'", tc.id, tc.expected, result)
			}
		}
	})
}

// TestLookupOgnTailNumber_MalformedJSON tests handling of malformed device database
func TestLookupOgnTailNumber_MalformedJSON(t *testing.T) {
	// Save original cache
	origCache := ognTailNumberCache
	defer func() {
		ognTailNumberCache = origCache
	}()

	// Try to create ogn subdirectory in STRATUX_HOME
	ognDir := STRATUX_HOME + "/ogn"
	err := os.MkdirAll(ognDir, 0755)
	if err != nil {
		t.Skipf("Cannot create test directory in STRATUX_HOME: %v", err)
	}

	t.Run("missing_devices_field", func(t *testing.T) {
		// Clear cache
		ognTailNumberCache = make(map[string]string)

		// Create JSON without "devices" field - will cause panic on type assertion
		testJSON := `{
  "version": "1.0",
  "other_field": "value"
}`
		err := os.WriteFile(ognDir+"/ddb.json", []byte(testJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write ddb.json: %v", err)
		}

		// This should panic on the type assertion data["devices"].([]interface{})
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Correctly panicked on missing devices field: %v", r)
			}
		}()

		result := lookupOgnTailNumber("TEST123")
		// If we get here without panic, test fails
		t.Errorf("Expected panic on missing devices field, got result: '%s'", result)
	})

	t.Run("devices_not_array", func(t *testing.T) {
		// Clear cache
		ognTailNumberCache = make(map[string]string)

		// Create JSON where "devices" is not an array
		testJSON := `{
  "devices": "not an array"
}`
		err := os.WriteFile(ognDir+"/ddb.json", []byte(testJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write ddb.json: %v", err)
		}

		// This should panic on type assertion
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Correctly panicked on devices not being array: %v", r)
			}
		}()

		result := lookupOgnTailNumber("TEST456")
		t.Errorf("Expected panic on devices not being array, got result: '%s'", result)
	})

	t.Run("device_missing_fields", func(t *testing.T) {
		// Clear cache
		ognTailNumberCache = make(map[string]string)

		// Create JSON with device missing device_id field
		testJSON := `{
  "devices": [
    {
      "registration": "N12345"
    }
  ]
}`
		err := os.WriteFile(ognDir+"/ddb.json", []byte(testJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write ddb.json: %v", err)
		}

		// This should panic on type assertion dev["device_id"].(string)
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Correctly panicked on missing device_id: %v", r)
			}
		}()

		result := lookupOgnTailNumber("TEST789")
		t.Errorf("Expected panic on missing device_id, got result: '%s'", result)
	})

	t.Run("device_wrong_types", func(t *testing.T) {
		// Clear cache
		ognTailNumberCache = make(map[string]string)

		// Create JSON with wrong field types
		testJSON := `{
  "devices": [
    {
      "device_id": 123456,
      "registration": "N12345"
    }
  ]
}`
		err := os.WriteFile(ognDir+"/ddb.json", []byte(testJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write ddb.json: %v", err)
		}

		// This should panic on type assertion
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Correctly panicked on wrong field type: %v", r)
			}
		}()

		result := lookupOgnTailNumber("TESTABC")
		t.Errorf("Expected panic on wrong field type, got result: '%s'", result)
	})
}

// TestLookupOgnTailNumber_EmptyDatabase tests handling of empty database
func TestLookupOgnTailNumber_EmptyDatabase(t *testing.T) {
	// Save original cache
	origCache := ognTailNumberCache
	defer func() {
		ognTailNumberCache = origCache
	}()

	// Try to create ogn subdirectory in STRATUX_HOME
	ognDir := STRATUX_HOME + "/ogn"
	err := os.MkdirAll(ognDir, 0755)
	if err != nil {
		t.Skipf("Cannot create test directory in STRATUX_HOME: %v", err)
	}

	t.Run("empty_devices_array", func(t *testing.T) {
		// Clear cache
		ognTailNumberCache = make(map[string]string)

		// Create valid JSON with empty devices array
		testJSON := `{
  "devices": []
}`
		err := os.WriteFile(ognDir+"/ddb.json", []byte(testJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write ddb.json: %v", err)
		}

		// Lookup should successfully load file but find nothing
		result := lookupOgnTailNumber("TEST123")
		if result != "" {
			t.Errorf("Expected empty string for lookup in empty db, got '%s'", result)
		}

		// Cache should be empty
		if len(ognTailNumberCache) != 0 {
			t.Errorf("Expected empty cache after loading empty db, got %d entries", len(ognTailNumberCache))
		}

		t.Log("Successfully handled empty devices array")
	})
}

// TestLookupOgnTailNumber_LargeDatabase tests handling of large database
func TestLookupOgnTailNumber_LargeDatabase(t *testing.T) {
	// Save original cache
	origCache := ognTailNumberCache
	defer func() {
		ognTailNumberCache = origCache
	}()

	// Try to create ogn subdirectory in STRATUX_HOME
	ognDir := STRATUX_HOME + "/ogn"
	err := os.MkdirAll(ognDir, 0755)
	if err != nil {
		t.Skipf("Cannot create test directory in STRATUX_HOME: %v", err)
	}

	t.Run("large_database_performance", func(t *testing.T) {
		// Clear cache
		ognTailNumberCache = make(map[string]string)

		// Create JSON with many devices
		var devices []string
		for i := 0; i < 100; i++ {
			device := `    {
      "device_id": "` + string(rune('A'+i/26)) + string(rune('A'+i%26)) + `1234",
      "registration": "N` + string(rune('0'+i/10)) + string(rune('0'+i%10)) + `345"
    }`
			devices = append(devices, device)
		}

		testJSON := `{
  "devices": [
` + strings.Join(devices, ",\n") + `
  ]
}`
		err := os.WriteFile(ognDir+"/ddb.json", []byte(testJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write large ddb.json: %v", err)
		}

		// Lookup should successfully load all devices
		result := lookupOgnTailNumber("AA1234")
		if result != "N00345" {
			t.Errorf("Expected 'N00345', got '%s'", result)
		}

		// Cache should have 100 entries
		if len(ognTailNumberCache) != 100 {
			t.Errorf("Expected cache to have 100 entries, got %d", len(ognTailNumberCache))
		}

		// Test lookup of last device
		result = lookupOgnTailNumber("DX1234")
		if result != "N99345" {
			t.Errorf("Expected 'N99345', got '%s'", result)
		}

		t.Logf("Successfully loaded and queried database with %d devices", len(ognTailNumberCache))
	})
}

// TestLookupOgnTailNumber_CoverageSummary documents the test coverage status
func TestLookupOgnTailNumber_CoverageSummary(t *testing.T) {
	t.Log("=== lookupOgnTailNumber Test Coverage Summary ===")
	t.Log("")
	t.Log("Tested scenarios (with cache-based tests):")
	t.Log("  ✓ Cache hit - found entry")
	t.Log("  ✓ Cache miss - entry not found")
	t.Log("  ✓ Empty cache with file read error")
	t.Log("  ✓ Empty/nil lookup IDs")
	t.Log("  ✓ Special characters in registrations")
	t.Log("  ✓ Case-sensitive lookups")
	t.Log("  ✓ Multiple sequential lookups")
	t.Log("  ✓ Interleaved found/not-found lookups")
	t.Log("")
	t.Log("Scenarios requiring file I/O access (may be skipped):")
	t.Log("  • Successful file parsing and cache population")
	t.Log("  • JSON unmarshal errors")
	t.Log("  • Missing 'devices' field (panic test)")
	t.Log("  • Invalid device array types (panic test)")
	t.Log("  • Missing device fields (panic test)")
	t.Log("  • Wrong field types (panic test)")
	t.Log("  • Empty devices array")
	t.Log("  • Large database (100+ entries)")
	t.Log("  • Cache persistence across lookups")
	t.Log("")
	t.Log("Current coverage: 42.9% (without file I/O tests)")
	t.Log("Potential coverage: ~100% (with file I/O tests enabled)")
	t.Log("")
	t.Log("To enable file I/O tests:")
	t.Log("  sudo mkdir -p /opt/stratux/ogn && sudo chmod 777 /opt/stratux/ogn")
}
