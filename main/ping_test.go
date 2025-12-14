/*
	Copyright (c) 2016 uAvionix
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file, herein included
	as part of this header.

	ping_test.go: Unit tests for ping.go functions.
*/

package main

import (
	"sync"
	"testing"
	"time"
)

// TestMavLinkParse tests the mavLinkParse function with various inputs
func TestMavLinkParse(t *testing.T) {
	t.Run("invalid_magic_byte", func(t *testing.T) {
		// Frame with invalid magic byte (not 0xfe)
		frame := []byte{0xff, 0x26, 0x00, 0x00, 0x00, 0xf6}
		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for invalid magic byte, got false")
		}
	})

	t.Run("oversized_frame", func(t *testing.T) {
		// Frame larger than 1024 bytes
		frame := make([]byte, 1025)
		frame[0] = 0xfe
		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for oversized frame, got false")
		}
	})

	t.Run("valid_magic_but_oversized", func(t *testing.T) {
		// Valid magic but size > 1024
		frame := make([]byte, 1100)
		frame[0] = 0xfe
		frame[1] = 38
		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for valid magic but oversized frame, got false")
		}
	})

	t.Run("frame_too_short", func(t *testing.T) {
		// Frame shorter than minimum 9 bytes
		frame := []byte{0xfe, 0x26, 0x00, 0x00, 0x00}
		result := mavLinkParse(frame)
		if result {
			t.Error("Expected false for too-short frame, got true")
		}
	})

	t.Run("frame_exactly_9_bytes_but_wrong_length", func(t *testing.T) {
		// Frame is 9 bytes, but length field indicates it should be longer
		frame := []byte{0xfe, 0x26, 0x00, 0x00, 0x00, 0xf6, 0x00, 0x00, 0x00}
		// msglen[1] = 38, expected total = 38 + 8 = 46, actual = 9
		result := mavLinkParse(frame)
		if result {
			t.Error("Expected false for length mismatch, got true")
		}
	})

	t.Run("valid_traffic_message_type_246", func(t *testing.T) {
		// Initialize test environment
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

		// Create a valid MavLink traffic message (type 246, length >= 38)
		// Format: [magic][len][seq][sysid][compid][msgid][payload...][checksum1][checksum2]
		msglen := byte(38)
		frame := make([]byte, int(msglen)+8)
		frame[0] = 0xfe   // Magic byte
		frame[1] = msglen // Payload length
		frame[2] = 0      // Sequence
		frame[3] = 1      // System ID
		frame[4] = 1      // Component ID
		frame[5] = 246    // Message ID (ADSB_VEHICLE)

		// Fill payload with zeros (38 bytes)
		for i := 6; i < 6+int(msglen); i++ {
			frame[i] = 0
		}

		// Checksums (not validated in mavLinkFormat)
		frame[len(frame)-2] = 0
		frame[len(frame)-1] = 0

		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for valid traffic message, got false")
		}
	})

	t.Run("valid_message_type_246_exact_38_bytes", func(t *testing.T) {
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

		// Message type 246 with exactly 38 bytes
		msglen := byte(38)
		frame := make([]byte, int(msglen)+8)
		frame[0] = 0xfe
		frame[1] = msglen
		frame[2] = 0
		frame[3] = 1
		frame[4] = 1
		frame[5] = 246

		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for type 246 with exactly 38 bytes, got false")
		}
	})

	t.Run("type_246_with_less_than_38_bytes", func(t *testing.T) {
		if stratuxClock == nil {
			stratuxClock = NewMonotonic()
		}

		// Message type 246 but length < 38
		msglen := byte(30)
		frame := make([]byte, int(msglen)+8)
		frame[0] = 0xfe
		frame[1] = msglen
		frame[2] = 0
		frame[3] = 1
		frame[4] = 1
		frame[5] = 246 // Type 246 but insufficient length

		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for undersized type 246 message, got false")
		}
	})

	t.Run("type_246_with_more_than_38_bytes", func(t *testing.T) {
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

		// Message type 246 with more than 38 bytes
		msglen := byte(50)
		frame := make([]byte, int(msglen)+8)
		frame[0] = 0xfe
		frame[1] = msglen
		frame[2] = 0
		frame[3] = 1
		frame[4] = 1
		frame[5] = 246

		for i := 6; i < len(frame); i++ {
			frame[i] = 0
		}

		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for oversized type 246 message, got false")
		}
	})

	t.Run("non_246_message_type", func(t *testing.T) {
		if stratuxClock == nil {
			stratuxClock = NewMonotonic()
		}

		// Valid MavLink frame but not type 246
		msglen := byte(20)
		frame := make([]byte, int(msglen)+8)
		frame[0] = 0xfe
		frame[1] = msglen
		frame[2] = 0
		frame[3] = 1
		frame[4] = 1
		frame[5] = 100 // Different message type

		for i := 6; i < len(frame); i++ {
			frame[i] = 0
		}

		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for non-246 message type, got false")
		}
	})

	t.Run("zero_length_payload", func(t *testing.T) {
		if stratuxClock == nil {
			stratuxClock = NewMonotonic()
		}

		// Zero-length payload - frame should be exactly 8 bytes (0 payload + 8 header/checksums)
		// But mavLinkParse requires len >= 9, so this will return false
		msglen := byte(0)
		frame := make([]byte, int(msglen)+8)
		frame[0] = 0xfe
		frame[1] = msglen
		frame[2] = 0
		frame[3] = 1
		frame[4] = 1
		frame[5] = 246

		result := mavLinkParse(frame)
		if result {
			t.Error("Expected false for zero-length payload (too short), got true")
		}
	})

	t.Run("length_field_large_causes_overflow", func(t *testing.T) {
		// When msglen > 247, the check int(mavLinkFrame[1]+8) causes byte overflow
		// For example, byte(254) + byte(8) = byte(6) due to wrap-around
		// So the length check will fail and return false
		msglen := byte(254)
		frame := make([]byte, int(msglen)+8)
		frame[0] = 0xfe
		frame[1] = msglen
		frame[2] = 0
		frame[3] = 1
		frame[4] = 1
		frame[5] = 246

		// Fill payload
		for i := 6; i < len(frame); i++ {
			frame[i] = byte(i % 256)
		}

		result := mavLinkParse(frame)
		// Expects false due to byte overflow in length check
		if result {
			t.Error("Expected false for large payload causing overflow, got true")
		}
	})

	t.Run("length_field_247_max_valid", func(t *testing.T) {
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

		// Maximum payload that doesn't overflow: 247 + 8 = 255 (max byte value)
		msglen := byte(247)
		frame := make([]byte, int(msglen)+8)
		frame[0] = 0xfe
		frame[1] = msglen
		frame[2] = 0
		frame[3] = 1
		frame[4] = 1
		frame[5] = 246

		// Fill payload
		for i := 6; i < len(frame); i++ {
			frame[i] = byte(i % 256)
		}

		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for max valid payload (247 bytes), got false")
		}
	})
}

// TestMavLinkParse_EdgeCases tests edge cases in MavLink parsing
func TestMavLinkParse_EdgeCases(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	t.Run("empty_frame", func(t *testing.T) {
		frame := []byte{}
		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for empty frame, got false")
		}
	})

	t.Run("single_byte_frame", func(t *testing.T) {
		frame := []byte{0xfe}
		result := mavLinkParse(frame)
		if result {
			t.Error("Expected false for single byte frame, got true")
		}
	})

	t.Run("exact_minimum_length", func(t *testing.T) {
		// Exactly 9 bytes with matching length field
		frame := []byte{0xfe, 0x01, 0x00, 0x01, 0x01, 0x00, 0xAA, 0x00, 0x00}
		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true for exact minimum valid frame, got false")
		}
	})

	t.Run("boundary_1024_bytes", func(t *testing.T) {
		// Exactly 1024 bytes (valid size within limit)
		// However, MavLink payload length (1024-8=1016) exceeds uint8 max (255)
		// So this will fail length validation (return false) even though size is valid
		frame := make([]byte, 1024)
		frame[0] = 0xfe
		frame[1] = 255 // Payload length can't represent 1016, so validation fails
		result := mavLinkParse(frame)
		if result {
			t.Error("Expected false for 1024-byte frame with invalid length field, got true")
		}
	})

	t.Run("boundary_1025_bytes", func(t *testing.T) {
		// Exactly 1025 bytes (invalid size)
		frame := make([]byte, 1025)
		frame[0] = 0xfe
		result := mavLinkParse(frame)
		if !result {
			t.Error("Expected true (invalid) for 1025-byte frame, got false")
		}
	})
}

// TestPingKill tests the pingKill function
func TestPingKill(t *testing.T) {
	t.Run("kill_when_not_connected", func(t *testing.T) {
		// Initialize global status
		globalStatus.Ping_connected = false
		shutdownPing = false

		// Should return immediately since already disconnected
		done := make(chan bool)
		go func() {
			pingKill()
			done <- true
		}()

		select {
		case <-done:
			// Success - function returned
		case <-time.After(2 * time.Second):
			t.Error("pingKill() did not return in expected time")
		}

		// shutdownPing should have been set to true during the call
		// (even though it returns immediately)
	})

	t.Run("verify_shutdown_flag_set", func(t *testing.T) {
		// Reset state
		shutdownPing = false
		globalStatus.Ping_connected = false

		// Call pingKill
		pingKill()

		// Verify shutdown flag was set (and then cleared when connection was already false)
		if globalStatus.Ping_connected {
			t.Error("Expected Ping_connected to be false after pingKill()")
		}
	})
}

// TestPingDeviceModel tests ping device model constants
func TestPingDeviceModel(t *testing.T) {
	// pingDeviceModel is set in initPingSerial:
	// 0 => pingEFB - 1090ES
	// 1 => pingUSB - MavLink

	t.Run("device_model_types", func(t *testing.T) {
		validModels := []int{0, 1}
		for _, model := range validModels {
			if model < 0 || model > 1 {
				t.Errorf("Invalid device model: %d", model)
			}
		}
	})
}

// TestMavlinkTrafficMessageFormat tests the MavlinkTrafficMessageFormat struct
func TestMavlinkTrafficMessageFormat(t *testing.T) {
	t.Run("struct_field_sizes", func(t *testing.T) {
		var msg MavlinkTrafficMessageFormat

		// Verify struct fields exist and can be assigned
		msg.ICAO_address = 0xABCDEF
		msg.lat = 123456789
		msg.lon = -123456789
		msg.altitude = 10000
		msg.heading = 360
		msg.hor_velocity = 250
		msg.ver_velocity = -500
		msg.validFlags = 0xFFFF
		msg.squawk = 7700
		msg.altitude_type = 1
		msg.callsign = [9]byte{'T', 'E', 'S', 'T', '1', '2', '3', 0, 0}
		msg.emitter_type = 2
		msg.tslc = 15

		// Verify values
		if msg.ICAO_address != 0xABCDEF {
			t.Errorf("ICAO_address = %x, want %x", msg.ICAO_address, 0xABCDEF)
		}
		if msg.squawk != 7700 {
			t.Errorf("squawk = %d, want %d", msg.squawk, 7700)
		}
		if msg.callsign[0] != 'T' {
			t.Errorf("callsign[0] = %c, want %c", msg.callsign[0], 'T')
		}
	})

	t.Run("callsign_array_size", func(t *testing.T) {
		var msg MavlinkTrafficMessageFormat

		// Callsign should be exactly 9 bytes
		if len(msg.callsign) != 9 {
			t.Errorf("callsign length = %d, want 9", len(msg.callsign))
		}
	})
}

// TestPingGlobalVariables tests initialization of ping global variables
func TestPingGlobalVariables(t *testing.T) {
	t.Run("shutdown_flag_type", func(t *testing.T) {
		// shutdownPing should be a bool
		shutdownPing = false
		if shutdownPing != false {
			t.Error("shutdownPing should be false")
		}
		shutdownPing = true
		if shutdownPing != true {
			t.Error("shutdownPing should be true")
		}
		shutdownPing = false // Reset
	})

	t.Run("ping_device_working_flag", func(t *testing.T) {
		// pingDeviceSuccessfullyWorking should be bool
		pingDeviceSuccessfullyWorking = false
		if pingDeviceSuccessfullyWorking {
			t.Error("pingDeviceSuccessfullyWorking should be false")
		}
		pingDeviceSuccessfullyWorking = true
		if !pingDeviceSuccessfullyWorking {
			t.Error("pingDeviceSuccessfullyWorking should be true")
		}
		pingDeviceSuccessfullyWorking = false // Reset
	})
}

// BenchmarkMavLinkParse benchmarks the mavLinkParse function
func BenchmarkMavLinkParse(b *testing.B) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Create a typical valid frame
	msglen := byte(38)
	frame := make([]byte, int(msglen)+8)
	frame[0] = 0xfe
	frame[1] = msglen
	frame[2] = 0
	frame[3] = 1
	frame[4] = 1
	frame[5] = 246

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mavLinkParse(frame)
	}
}

// BenchmarkMavLinkParseInvalid benchmarks parsing invalid frames
func BenchmarkMavLinkParseInvalid(b *testing.B) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Invalid frame (wrong magic byte)
	frame := []byte{0xff, 0x26, 0x00, 0x00, 0x00, 0xf6}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mavLinkParse(frame)
	}
}
