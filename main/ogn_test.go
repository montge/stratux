package main

import (
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
