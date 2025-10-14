package main

import (
	"testing"
)

// TestMakeHeartbeatBasic tests the makeHeartbeat function for basic message structure
func TestMakeHeartbeatBasic(t *testing.T) {
	// Initialize CRC table (required for prepareMessage)
	crcInit()

	// Initialize stratuxClock (required for GPS validity checks)
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Call the function
	msg := makeHeartbeat()

	// Basic validations
	if len(msg) < 10 { // At least frame markers + data + CRC
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("makeHeartbeat() generated %d-byte message", len(msg))
}

// TestMakeStratuxHeartbeatBasic tests the makeStratuxHeartbeat function
func TestMakeStratuxHeartbeatBasic(t *testing.T) {
	// Initialize CRC table (required for prepareMessage)
	crcInit()

	// Initialize stratuxClock (required for GPS/AHRS validity checks)
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Call the function
	msg := makeStratuxHeartbeat()

	// Basic validations
	if len(msg) < 5 { // At least frame markers + data + CRC
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("makeStratuxHeartbeat() generated %d-byte message", len(msg))
}

// TestMakeFFIDMessageBasic tests the makeFFIDMessage function
func TestMakeFFIDMessageBasic(t *testing.T) {
	// Initialize CRC table (required for prepareMessage)
	crcInit()

	// Set version strings to ensure function works
	stratuxVersion = "v1.6"
	stratuxBuild = "test"

	// Call the function
	msg := makeFFIDMessage()

	// Basic validations
	if len(msg) < 40 { // At least frame markers + 39 data bytes + CRC
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("makeFFIDMessage() generated %d-byte message", len(msg))
}

// TestMakeFFIDMessageLongNames tests makeFFIDMessage with long version strings
func TestMakeFFIDMessageLongNames(t *testing.T) {
	// Initialize CRC table (required for prepareMessage)
	crcInit()

	// Set very long version strings to test truncation logic
	stratuxVersion = "v999.999.999.999"
	stratuxBuild = "verylongbuildstring"

	// Call the function
	msg := makeFFIDMessage()

	// Basic validations - should not panic and should generate valid message
	if len(msg) < 40 {
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("makeFFIDMessage() with long strings generated %d-byte message", len(msg))
}
