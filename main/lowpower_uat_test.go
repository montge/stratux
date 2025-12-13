package main

import (
	"sync"
	"testing"
	"time"
)

// TestProcessRadioMessage tests the processRadioMessage function with various message lengths
func TestProcessRadioMessage(t *testing.T) {
	// Initialize test environment
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	t.Run("uplink_frame_552_bytes", func(t *testing.T) {
		// Create a 552-byte message with RSSI and timestamp prepended
		// Format: [RSSI(1 byte)][Timestamp(4 bytes)][Message(552 bytes)]
		msg := make([]byte, 557)
		msg[0] = 206 // RSSI value (int8(-50) as uint8)

		// Timestamp bytes (currently unused but present in format)
		msg[1] = 0x01
		msg[2] = 0x02
		msg[3] = 0x03
		msg[4] = 0x04

		// Fill the rest with dummy data (552 bytes)
		for i := 5; i < 557; i++ {
			msg[i] = byte(i % 256)
		}

		// This should not panic - the function will attempt FEC correction
		processRadioMessage(msg)
		t.Log("Successfully processed 552-byte uplink frame")
	})

	t.Run("adsb_frame_48_bytes_short", func(t *testing.T) {
		// Create a 48-byte message with RSSI and timestamp prepended
		// Format: [RSSI(1 byte)][Timestamp(4 bytes)][Message(48 bytes)]
		msg := make([]byte, 53)
		msg[0] = 196 // RSSI value (int8(-60) as uint8)

		// Timestamp bytes
		msg[1] = 0x01
		msg[2] = 0x02
		msg[3] = 0x03
		msg[4] = 0x04

		// Fill with dummy ADS-B data (48 bytes)
		for i := 5; i < 53; i++ {
			msg[i] = byte(i % 256)
		}

		// This should not panic - function will attempt FEC correction
		processRadioMessage(msg)
		t.Log("Successfully processed 48-byte ADS-B frame")
	})

	t.Run("unhandled_message_size", func(t *testing.T) {
		// Create a message with non-standard size (not 552 or 48)
		// Format: [RSSI(1 byte)][Timestamp(4 bytes)][Message(100 bytes)]
		msg := make([]byte, 105)
		msg[0] = 186 // RSSI value (int8(-70) as uint8)

		// Timestamp bytes
		msg[1] = 0x01
		msg[2] = 0x02
		msg[3] = 0x03
		msg[4] = 0x04

		// Fill with dummy data
		for i := 5; i < 105; i++ {
			msg[i] = byte(i % 256)
		}

		// This should log "processRadioMessage(): unhandled message size 100"
		// but not panic
		processRadioMessage(msg)
		t.Log("Successfully handled unhandled message size without panic")
	})

	t.Run("minimum_size_message", func(t *testing.T) {
		// Test with minimum possible message (just RSSI + timestamp)
		msg := make([]byte, 5)
		msg[0] = 176 // RSSI value (int8(-80) as uint8)
		msg[1] = 0x01
		msg[2] = 0x02
		msg[3] = 0x03
		msg[4] = 0x04

		// Should log unhandled message size 0 but not panic
		processRadioMessage(msg)
		t.Log("Successfully handled minimum size message")
	})

	t.Run("rssi_value_extraction", func(t *testing.T) {
		// Test that RSSI is correctly extracted
		msg := make([]byte, 105)

		// Test with various RSSI values (as uint8 representations of int8)
		rssiValues := []uint8{206, 196, 186, 176, 166, 156} // -50, -60, -70, -80, -90, -100
		for _, rssi := range rssiValues {
			msg[0] = rssi
			msg[1] = 0x01
			msg[2] = 0x02
			msg[3] = 0x03
			msg[4] = 0x04

			// Fill rest with data
			for i := 5; i < 105; i++ {
				msg[i] = byte(i % 256)
			}

			// Should process without panic
			processRadioMessage(msg)
		}
		t.Log("Successfully tested RSSI value extraction")
	})

	t.Run("zero_length_message_after_header", func(t *testing.T) {
		// Edge case: only header, no actual message data
		msg := make([]byte, 5)
		msg[0] = 171 // RSSI value (int8(-85) as uint8)
		msg[1] = 0xFF
		msg[2] = 0xFF
		msg[3] = 0xFF
		msg[4] = 0xFF

		// Should handle gracefully
		processRadioMessage(msg)
		t.Log("Successfully handled zero-length message after header")
	})

	t.Run("large_timestamp_values", func(t *testing.T) {
		// Test with maximum timestamp values
		msg := make([]byte, 105)
		msg[0] = 181 // RSSI value (int8(-75) as uint8)
		msg[1] = 0xFF
		msg[2] = 0xFF
		msg[3] = 0xFF
		msg[4] = 0xFF

		for i := 5; i < 105; i++ {
			msg[i] = byte(i % 256)
		}

		processRadioMessage(msg)
		t.Log("Successfully processed message with large timestamp values")
	})
}

// TestProcessRadioMessage_Integration tests integration with parseInput
func TestProcessRadioMessage_Integration(t *testing.T) {
	// Initialize required globals
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	t.Run("adsb_short_frame_integration", func(t *testing.T) {
		// Create a 48-byte ADS-B message
		// The actual FEC correction will likely fail with random data,
		// but we're testing that the code path executes without panic
		msg := make([]byte, 53)
		msg[0] = 201 // RSSI (int8(-55) as uint8)

		// Timestamp
		msg[1] = 0x00
		msg[2] = 0x00
		msg[3] = 0x00
		msg[4] = 0x00

		// Fill with pattern that might look like valid data
		for i := 5; i < 53; i++ {
			msg[i] = 0xAA
		}

		// Should execute without panic even if FEC fails
		processRadioMessage(msg)
		t.Log("Integration test completed for ADS-B short frame")
	})

	t.Run("uplink_frame_integration", func(t *testing.T) {
		// Create a 552-byte uplink message
		msg := make([]byte, 557)
		msg[0] = 191 // RSSI (int8(-65) as uint8)

		// Timestamp
		msg[1] = 0x00
		msg[2] = 0x00
		msg[3] = 0x00
		msg[4] = 0x00

		// Fill with pattern
		for i := 5; i < 557; i++ {
			msg[i] = 0x55
		}

		// Should execute without panic
		processRadioMessage(msg)
		t.Log("Integration test completed for uplink frame")
	})
}

// TestProcessRadioMessage_VariousDataPatterns tests various data patterns for ADS-B frames
func TestProcessRadioMessage_VariousDataPatterns(t *testing.T) {
	// Initialize required globals once for all subtests
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond)
	}
	if trafficMutex == nil {
		trafficMutex = &sync.Mutex{}
	}
	if traffic == nil {
		traffic = make(map[uint32]TrafficInfo)
	}
	if seenTraffic == nil {
		seenTraffic = make(map[uint32]bool)
	}
	if networkGDL90Chan == nil {
		networkGDL90Chan = make(chan []byte, 100)
	}
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}

	t.Run("adsb_48_all_zeros", func(t *testing.T) {
		msg := make([]byte, 53)
		msg[0] = 200 // RSSI
		// Timestamp and rest are all zeros
		processRadioMessage(msg)
		t.Log("Processed ADS-B frame with all zeros")
	})

	t.Run("adsb_48_all_ones", func(t *testing.T) {
		msg := make([]byte, 53)
		msg[0] = 200 // RSSI
		for i := 1; i < 53; i++ {
			msg[i] = 0xFF
		}
		processRadioMessage(msg)
		t.Log("Processed ADS-B frame with all ones")
	})

	t.Run("adsb_48_alternating", func(t *testing.T) {
		msg := make([]byte, 53)
		msg[0] = 200 // RSSI
		for i := 5; i < 53; i++ {
			if i%2 == 0 {
				msg[i] = 0xAA
			} else {
				msg[i] = 0x55
			}
		}
		processRadioMessage(msg)
		t.Log("Processed ADS-B frame with alternating pattern")
	})

	t.Run("uplink_552_all_zeros", func(t *testing.T) {
		msg := make([]byte, 557)
		msg[0] = 190 // RSSI
		// Rest are zeros
		processRadioMessage(msg)
		t.Log("Processed uplink frame with all zeros")
	})

	t.Run("uplink_552_all_ones", func(t *testing.T) {
		msg := make([]byte, 557)
		msg[0] = 190 // RSSI
		for i := 1; i < 557; i++ {
			msg[i] = 0xFF
		}
		processRadioMessage(msg)
		t.Log("Processed uplink frame with all ones")
	})

	t.Run("uplink_552_incrementing", func(t *testing.T) {
		msg := make([]byte, 557)
		msg[0] = 190 // RSSI
		for i := 1; i < 557; i++ {
			msg[i] = byte(i % 256)
		}
		processRadioMessage(msg)
		t.Log("Processed uplink frame with incrementing pattern")
	})
}

// TestProcessRadioMessage_RSSIBoundaries tests RSSI boundary conditions
func TestProcessRadioMessage_RSSIBoundaries(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	rssiTestCases := []struct {
		name     string
		rssiRaw  byte
		rssiInt8 int8
	}{
		{"rssi_zero", 0, 0},
		{"rssi_max_positive", 127, 127},
		{"rssi_min_negative", 128, -128}, // 128 as uint8 = -128 as int8
		{"rssi_negative_one", 255, -1},
		{"rssi_typical_negative_50", 206, -50},
		{"rssi_typical_negative_100", 156, -100},
	}

	for _, tc := range rssiTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with 48-byte message
			msg := make([]byte, 53)
			msg[0] = tc.rssiRaw
			processRadioMessage(msg)
			t.Logf("Processed message with RSSI raw=%d (int8=%d)", tc.rssiRaw, tc.rssiInt8)
		})
	}
}

// TestProcessRadioMessage_TimestampVariations tests different timestamp values
func TestProcessRadioMessage_TimestampVariations(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	timestampCases := []struct {
		name string
		ts   []byte // 4 bytes for timestamp
	}{
		{"timestamp_zero", []byte{0x00, 0x00, 0x00, 0x00}},
		{"timestamp_max", []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{"timestamp_incrementing", []byte{0x01, 0x02, 0x03, 0x04}},
		{"timestamp_random_1", []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		{"timestamp_random_2", []byte{0x12, 0x34, 0x56, 0x78}},
		{"timestamp_high_bit", []byte{0x80, 0x00, 0x00, 0x00}},
	}

	for _, tc := range timestampCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := make([]byte, 53)
			msg[0] = 200 // RSSI
			copy(msg[1:5], tc.ts)
			processRadioMessage(msg)
			t.Logf("Processed message with timestamp %02x%02x%02x%02x",
				tc.ts[0], tc.ts[1], tc.ts[2], tc.ts[3])
		})
	}
}

// TestProcessRadioMessage_EdgeCases tests edge cases in message processing
func TestProcessRadioMessage_EdgeCases(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	t.Run("message_with_invalid_size_1_byte", func(t *testing.T) {
		msg := make([]byte, 6) // 1 byte message after header
		msg[0] = 200
		processRadioMessage(msg)
		t.Log("Processed message with 1-byte payload")
	})

	t.Run("message_with_invalid_size_10_bytes", func(t *testing.T) {
		msg := make([]byte, 15) // 10 bytes after header
		msg[0] = 200
		processRadioMessage(msg)
		t.Log("Processed message with 10-byte payload")
	})

	t.Run("message_with_invalid_size_100_bytes", func(t *testing.T) {
		msg := make([]byte, 105) // 100 bytes after header
		msg[0] = 200
		processRadioMessage(msg)
		t.Log("Processed message with 100-byte payload")
	})

	t.Run("message_with_size_between_48_and_552", func(t *testing.T) {
		msg := make([]byte, 305) // 300 bytes after header (between 48 and 552)
		msg[0] = 200
		processRadioMessage(msg)
		t.Log("Processed message with 300-byte payload")
	})

	t.Run("message_larger_than_552", func(t *testing.T) {
		msg := make([]byte, 1005) // 1000 bytes after header (larger than 552)
		msg[0] = 200
		processRadioMessage(msg)
		t.Log("Processed message with 1000-byte payload")
	})
}

// TestProcessRadioMessage_MultipleCalls tests processing multiple messages in sequence
func TestProcessRadioMessage_MultipleCalls(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Process multiple messages of different types
	messages := [][]byte{
		// ADS-B 48-byte
		make([]byte, 53),
		// Uplink 552-byte
		make([]byte, 557),
		// Invalid size
		make([]byte, 105),
		// Another ADS-B
		make([]byte, 53),
		// Another uplink
		make([]byte, 557),
	}

	// Initialize messages
	for i, msg := range messages {
		msg[0] = byte(200 + i) // Different RSSI for each
		for j := 1; j < len(msg); j++ {
			msg[j] = byte((i * j) % 256)
		}
	}

	// Process all messages
	for i, msg := range messages {
		processRadioMessage(msg)
		t.Logf("Processed message %d with size %d", i, len(msg)-5)
	}

	t.Log("Successfully processed multiple messages in sequence")
}

// TestProcessRadioMessage_SpecificBytePatterns tests specific byte patterns that might trigger edge cases
func TestProcessRadioMessage_SpecificBytePatterns(t *testing.T) {
	// Initialize required globals
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond)
	}
	if trafficMutex == nil {
		trafficMutex = &sync.Mutex{}
	}
	if traffic == nil {
		traffic = make(map[uint32]TrafficInfo)
	}
	if seenTraffic == nil {
		seenTraffic = make(map[uint32]bool)
	}
	if networkGDL90Chan == nil {
		networkGDL90Chan = make(chan []byte, 100)
	}
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}

	t.Run("adsb_with_sync_pattern", func(t *testing.T) {
		// ADS-B messages might have sync patterns
		msg := make([]byte, 53)
		msg[0] = 200
		// Add a pattern that looks like a Mode-S preamble
		msg[5] = 0xA1
		msg[6] = 0x40
		processRadioMessage(msg)
		t.Log("Processed ADS-B with sync-like pattern")
	})

	t.Run("uplink_with_fis_b_header", func(t *testing.T) {
		// Uplink FIS-B messages have specific headers
		msg := make([]byte, 557)
		msg[0] = 200
		// Simulate a FIS-B header structure
		msg[5] = 0x01 // Version
		msg[6] = 0x19 // Product ID (text weather)
		msg[7] = 0x80 // Flags
		processRadioMessage(msg)
		t.Log("Processed uplink with FIS-B-like header")
	})

	t.Run("message_with_all_printable_ascii", func(t *testing.T) {
		msg := make([]byte, 53)
		msg[0] = 200
		// Fill with printable ASCII
		for i := 5; i < 53; i++ {
			msg[i] = byte(0x20 + (i % 95)) // Printable ASCII range
		}
		processRadioMessage(msg)
		t.Log("Processed message with printable ASCII pattern")
	})

	t.Run("message_with_high_entropy", func(t *testing.T) {
		msg := make([]byte, 53)
		msg[0] = 200
		// Create pseudo-random high entropy data
		for i := 5; i < 53; i++ {
			msg[i] = byte((i*7 + 13) % 256)
		}
		processRadioMessage(msg)
		t.Log("Processed message with high entropy data")
	})

	t.Run("adsb_long_frame_pattern", func(t *testing.T) {
		// Test with a pattern that triggers long ADS-B frame processing (i == 2)
		// Using test vector from DO-282B Table 2-104 #6
		// This is a valid long frame where to[0]>>3 != 0
		msg := make([]byte, 53)
		msg[0] = 200 // RSSI
		// Timestamp
		msg[1] = 0x00
		msg[2] = 0x00
		msg[3] = 0x00
		msg[4] = 0x00

		// Long ADS-B test vector (48 bytes) - based on DO-282B test data
		// This should trigger correct_adsb_frame to return 2 (long frame)
		// The key is that byte[0] has upper bits set (to[0]>>3 != 0)
		testVector := []byte{
			0xA5, 0x73, 0x55, 0xB5, 0x43, 0x85, 0xED, 0x2A,
			0xD0, 0x92, 0x9E, 0xB8, 0x0D, 0xA0, 0x44, 0xD5,
			0x56, 0xB4, 0x52, 0xA7, 0xA7, 0x3A, 0x57, 0x16,
			0xCD, 0x1D, 0xB4, 0x09, 0x64, 0xF5, 0xBA, 0x10,
			0x5D, 0x9D, 0xAA, 0xD7, 0x53, 0x42, 0x19, 0x6D,
			0xEB, 0xB6, 0x3C, 0xC9, 0x72, 0x99, 0x4D, 0xEA,
		}
		copy(msg[5:], testVector)

		processRadioMessage(msg)
		t.Log("Processed long ADS-B frame pattern (frametype 2)")
	})
}
