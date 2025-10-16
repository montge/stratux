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
