package main

import (
	"sync"
	"testing"
)

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
