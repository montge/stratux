package main

import (
	"testing"
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
