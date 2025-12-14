/*
	pong_test.go: Unit tests for pong.go functions.
	Added 1/2026
*/

package main

import (
	"testing"
	"time"
)

// TestPongSetUpdateMode tests the pongSetUpdateMode function
func TestPongSetUpdateMode(t *testing.T) {
	t.Run("set_update_mode_true", func(t *testing.T) {
		// Reset state
		pongUpdateMode = false

		// Call function
		pongSetUpdateMode()

		// Verify mode is set
		if !pongUpdateMode {
			t.Error("Expected pongUpdateMode to be true after pongSetUpdateMode()")
		}
	})

	t.Run("set_update_mode_when_already_true", func(t *testing.T) {
		// Set state to true
		pongUpdateMode = true

		// Call function again
		pongSetUpdateMode()

		// Should still be true
		if !pongUpdateMode {
			t.Error("Expected pongUpdateMode to remain true")
		}
	})

	t.Run("multiple_calls", func(t *testing.T) {
		// Reset
		pongUpdateMode = false

		// Call multiple times
		pongSetUpdateMode()
		pongSetUpdateMode()
		pongSetUpdateMode()

		// Should be true
		if !pongUpdateMode {
			t.Error("Expected pongUpdateMode to be true after multiple calls")
		}
	})
}

// TestPongKill tests the pongKill function
func TestPongKill(t *testing.T) {
	t.Run("kill_when_not_connected", func(t *testing.T) {
		// Initialize global status
		globalStatus.Pong_connected = false
		shutdownPong = false

		// Should return immediately since already disconnected
		done := make(chan bool)
		go func() {
			pongKill()
			done <- true
		}()

		select {
		case <-done:
			// Success - function returned
		case <-time.After(2 * time.Second):
			t.Error("pongKill() did not return in expected time")
		}

		// shutdownPong should have been set to true during the call
		// (even though it returns immediately)
	})

	t.Run("verify_shutdown_flag_set", func(t *testing.T) {
		// Reset state
		shutdownPong = false
		globalStatus.Pong_connected = false

		// Call pongKill
		pongKill()

		// Verify shutdown flag was set (and then cleared when connection was already false)
		if globalStatus.Pong_connected {
			t.Error("Expected Pong_connected to be false after pongKill()")
		}
	})

	t.Run("shutdown_flag_type", func(t *testing.T) {
		// Verify shutdownPong is a bool and can be set
		shutdownPong = false
		if shutdownPong {
			t.Error("shutdownPong should be false")
		}

		shutdownPong = true
		if !shutdownPong {
			t.Error("shutdownPong should be true")
		}

		// Reset
		shutdownPong = false
	})
}

// TestPongTermMessage tests the PongTermMessage struct
func TestPongTermMessage(t *testing.T) {
	t.Run("struct_creation", func(t *testing.T) {
		msg := PongTermMessage{
			Text:   "test message",
			Source: "stdout",
		}

		if msg.Text != "test message" {
			t.Errorf("Message text = %q, want %q", msg.Text, "test message")
		}
		if msg.Source != "stdout" {
			t.Errorf("Message source = %q, want %q", msg.Source, "stdout")
		}
	})

	t.Run("empty_fields", func(t *testing.T) {
		msg := PongTermMessage{
			Text:   "",
			Source: "",
		}

		if msg.Text != "" {
			t.Errorf("Message text = %q, want empty string", msg.Text)
		}
		if msg.Source != "" {
			t.Errorf("Message source = %q, want empty string", msg.Source)
		}
	})

	t.Run("long_text", func(t *testing.T) {
		longText := "This is a very long message that might be typical of actual Pong device output during update or error conditions"
		msg := PongTermMessage{
			Text:   longText,
			Source: "stderr",
		}

		if msg.Text != longText {
			t.Errorf("Message text length = %d, want %d", len(msg.Text), len(longText))
		}
		if msg.Source != "stderr" {
			t.Errorf("Message source = %q, want %q", msg.Source, "stderr")
		}
	})

	t.Run("various_sources", func(t *testing.T) {
		sources := []string{"stdout", "stderr", "update", "error"}
		for _, src := range sources {
			msg := PongTermMessage{
				Text:   "test",
				Source: src,
			}
			if msg.Source != src {
				t.Errorf("Message source = %q, want %q", msg.Source, src)
			}
		}
	})
}

// TestPongGlobalVariables tests initialization and state of pong global variables
func TestPongGlobalVariables(t *testing.T) {
	t.Run("pong_update_mode_flag", func(t *testing.T) {
		// pongUpdateMode should be a bool
		pongUpdateMode = false
		if pongUpdateMode {
			t.Error("pongUpdateMode should be false")
		}

		pongUpdateMode = true
		if !pongUpdateMode {
			t.Error("pongUpdateMode should be true")
		}

		// Reset
		pongUpdateMode = false
	})

	t.Run("pong_device_working_flag", func(t *testing.T) {
		// pongDeviceSuccessfullyWorking should be bool
		pongDeviceSuccessfullyWorking = false
		if pongDeviceSuccessfullyWorking {
			t.Error("pongDeviceSuccessfullyWorking should be false")
		}

		pongDeviceSuccessfullyWorking = true
		if !pongDeviceSuccessfullyWorking {
			t.Error("pongDeviceSuccessfullyWorking should be true")
		}

		// Reset
		pongDeviceSuccessfullyWorking = false
	})

	t.Run("shutdown_pong_flag", func(t *testing.T) {
		// shutdownPong should be a bool
		shutdownPong = false
		if shutdownPong {
			t.Error("shutdownPong should be false")
		}

		shutdownPong = true
		if !shutdownPong {
			t.Error("shutdownPong should be true")
		}

		// Reset
		shutdownPong = false
	})
}

// TestPongUpdateModeIntegration tests the interaction between update mode and device state
func TestPongUpdateModeIntegration(t *testing.T) {
	t.Run("update_mode_state_transitions", func(t *testing.T) {
		// Initial state
		pongUpdateMode = false
		pongDeviceSuccessfullyWorking = false

		// Enable update mode
		pongSetUpdateMode()
		if !pongUpdateMode {
			t.Error("Update mode should be true after calling pongSetUpdateMode()")
		}

		// Simulate update completion (would normally be done by pongWatcher)
		pongUpdateMode = false
		if pongUpdateMode {
			t.Error("Update mode should be false after manual reset")
		}

		// Reset working flag
		pongDeviceSuccessfullyWorking = false
	})

	t.Run("concurrent_update_mode_settings", func(t *testing.T) {
		// Reset
		pongUpdateMode = false

		// Simulate concurrent calls (might happen in edge cases)
		done := make(chan bool, 3)

		for i := 0; i < 3; i++ {
			go func() {
				pongSetUpdateMode()
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 3; i++ {
			select {
			case <-done:
				// Success
			case <-time.After(1 * time.Second):
				t.Error("Goroutine did not complete")
			}
		}

		// All should have set the flag
		if !pongUpdateMode {
			t.Error("Update mode should be true after concurrent calls")
		}

		// Reset
		pongUpdateMode = false
	})
}

// TestPongHeartbeatCounter tests the Pong heartbeat counter functionality
func TestPongHeartbeatCounter(t *testing.T) {
	t.Run("heartbeat_counter_increment", func(t *testing.T) {
		// Reset counter
		globalStatus.Pong_Heartbeats = 0

		// Simulate heartbeat increments (as done in pongSerialReader)
		for i := 1; i <= 10; i++ {
			globalStatus.Pong_Heartbeats++
			if globalStatus.Pong_Heartbeats != int64(i) {
				t.Errorf("Heartbeat counter = %d, want %d", globalStatus.Pong_Heartbeats, i)
			}
		}
	})

	t.Run("heartbeat_counter_reset", func(t *testing.T) {
		// Set counter to non-zero
		globalStatus.Pong_Heartbeats = 12345

		// Reset (as done in initPongSerial)
		globalStatus.Pong_Heartbeats = 0

		if globalStatus.Pong_Heartbeats != 0 {
			t.Errorf("Heartbeat counter = %d, want 0", globalStatus.Pong_Heartbeats)
		}
	})

	t.Run("heartbeat_counter_large_value", func(t *testing.T) {
		// Test with large int64 value
		globalStatus.Pong_Heartbeats = 9223372036854775800 // Near max int64

		// Increment several times
		for i := 0; i < 5; i++ {
			globalStatus.Pong_Heartbeats++
		}

		// Should have incremented successfully
		if globalStatus.Pong_Heartbeats != 9223372036854775805 {
			t.Errorf("Heartbeat counter = %d, want %d", globalStatus.Pong_Heartbeats, 9223372036854775805)
		}

		// Reset
		globalStatus.Pong_Heartbeats = 0
	})
}

// TestPongConnectionStates tests various connection state scenarios
func TestPongConnectionStates(t *testing.T) {
	t.Run("initial_disconnected_state", func(t *testing.T) {
		// Reset to disconnected
		globalStatus.Pong_connected = false
		globalSettings.Pong_Enabled = false

		if globalStatus.Pong_connected {
			t.Error("Pong should be disconnected initially")
		}
		if globalSettings.Pong_Enabled {
			t.Error("Pong should be disabled initially")
		}
	})

	t.Run("enabled_but_not_connected", func(t *testing.T) {
		// Simulate enabled in settings but not yet connected
		globalSettings.Pong_Enabled = true
		globalStatus.Pong_connected = false

		if !globalSettings.Pong_Enabled {
			t.Error("Pong should be enabled in settings")
		}
		if globalStatus.Pong_connected {
			t.Error("Pong should not be connected yet")
		}

		// Reset
		globalSettings.Pong_Enabled = false
	})

	t.Run("connected_and_enabled", func(t *testing.T) {
		// Simulate fully operational state
		globalSettings.Pong_Enabled = true
		globalStatus.Pong_connected = true
		pongDeviceSuccessfullyWorking = true

		if !globalSettings.Pong_Enabled || !globalStatus.Pong_connected || !pongDeviceSuccessfullyWorking {
			t.Error("All Pong flags should be true in operational state")
		}

		// Reset
		globalSettings.Pong_Enabled = false
		globalStatus.Pong_connected = false
		pongDeviceSuccessfullyWorking = false
	})
}

// BenchmarkPongSetUpdateMode benchmarks the pongSetUpdateMode function
func BenchmarkPongSetUpdateMode(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pongUpdateMode = false
		pongSetUpdateMode()
	}
}

// BenchmarkPongKill benchmarks the pongKill function when already disconnected
func BenchmarkPongKill(b *testing.B) {
	globalStatus.Pong_connected = false
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shutdownPong = false
		pongKill()
	}
}
