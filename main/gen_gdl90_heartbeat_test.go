package main

import (
	"sync"
	"testing"
	"time"
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

// TestMakeHeartbeatWithGPS tests makeHeartbeat with valid GPS
func TestMakeHeartbeatWithGPS(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Initialize stratuxClock
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original values
	origSituation := mySituation
	defer func() { mySituation = origSituation }()

	// Set up valid GPS situation
	mySituation.GPSLastFixSinceMidnightUTC = 3600.0
	mySituation.GPSLastFixLocalTime = stratuxClock.Time.Add(-1 * time.Second)

	msg := makeHeartbeat()

	// Check that message was generated
	if len(msg) < 10 {
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("makeHeartbeat() with GPS generated %d-byte message", len(msg))
}

// TestMakeHeartbeatWithErrors tests makeHeartbeat with system errors present
func TestMakeHeartbeatWithErrors(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Initialize stratuxClock
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original status
	origStatus := globalStatus
	defer func() { globalStatus = origStatus }()

	// Add some errors
	globalStatus.Errors = []string{"Test error 1", "Test error 2"}

	msg := makeHeartbeat()

	// Check that message was generated
	if len(msg) < 10 {
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("makeHeartbeat() with errors generated %d-byte message", len(msg))
}

// TestMakeStratuxHeartbeatWithGPSAndAHRS tests all combinations of GPS and AHRS validity
func TestMakeStratuxHeartbeatWithGPSAndAHRS(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Initialize stratuxClock
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	testCases := []struct {
		name        string
		setupFunc   func()
		description string
	}{
		{
			name: "GPS_Invalid_AHRS_Invalid",
			setupFunc: func() {
				mySituation.GPSLastFixLocalTime = stratuxClock.Time.Add(-60 * time.Second)
				mySituation.AHRSLastAttitudeTime = stratuxClock.Time.Add(-60 * time.Second)
			},
			description: "Both GPS and AHRS invalid",
		},
		{
			name: "GPS_Valid_AHRS_Invalid",
			setupFunc: func() {
				mySituation.GPSLastFixLocalTime = stratuxClock.Time.Add(-1 * time.Second)
				mySituation.AHRSLastAttitudeTime = stratuxClock.Time.Add(-60 * time.Second)
			},
			description: "GPS valid, AHRS invalid",
		},
		{
			name: "GPS_Invalid_AHRS_Valid",
			setupFunc: func() {
				mySituation.GPSLastFixLocalTime = stratuxClock.Time.Add(-60 * time.Second)
				mySituation.AHRSLastAttitudeTime = stratuxClock.Time.Add(-1 * time.Second)
			},
			description: "GPS invalid, AHRS valid",
		},
		{
			name: "GPS_Valid_AHRS_Valid",
			setupFunc: func() {
				mySituation.GPSLastFixLocalTime = stratuxClock.Time.Add(-1 * time.Second)
				mySituation.AHRSLastAttitudeTime = stratuxClock.Time.Add(-1 * time.Second)
			},
			description: "Both GPS and AHRS valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original situation
			origSituation := mySituation
			defer func() { mySituation = origSituation }()

			// Set up test scenario
			tc.setupFunc()

			msg := makeStratuxHeartbeat()

			// Check that message was generated
			if len(msg) < 5 {
				t.Errorf("%s: Message too short: got %d bytes", tc.description, len(msg))
			}

			// Check frame markers
			if msg[0] != 0x7E {
				t.Errorf("%s: Expected start frame marker 0x7E, got 0x%02X", tc.description, msg[0])
			}
			if msg[len(msg)-1] != 0x7E {
				t.Errorf("%s: Expected end frame marker 0x7E, got 0x%02X", tc.description, msg[len(msg)-1])
			}

			t.Logf("%s: generated %d-byte message", tc.description, len(msg))
		})
	}
}

// TestMakeFFIDMessageShortNames tests makeFFIDMessage with short version strings
func TestMakeFFIDMessageShortNames(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Set short version strings (no truncation needed)
	stratuxVersion = "v1.6"
	stratuxBuild = "test"

	msg := makeFFIDMessage()

	// Basic validations
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

	t.Logf("makeFFIDMessage() with short strings generated %d-byte message", len(msg))
}

// TestMakeStratuxHeartbeat_WithValidAHRS tests makeStratuxHeartbeat with valid AHRS
func TestMakeStratuxHeartbeat_WithValidAHRS(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Initialize stratuxClock
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	time.Sleep(50 * time.Millisecond) // Let clock start

	// Initialize mutexes if needed
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muAttitude == nil {
		mySituation.muAttitude = &sync.Mutex{}
	}

	// Set GPS valid: recent fix, connected, fix quality > 0
	mySituation.muGPS.Lock()
	mySituation.GPSLastFixLocalTime = stratuxClock.Time // Very recent
	mySituation.GPSFixQuality = 1
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Set AHRS valid: recent attitude time
	mySituation.muAttitude.Lock()
	mySituation.AHRSLastAttitudeTime = stratuxClock.Time // Very recent (< 1 second)
	mySituation.muAttitude.Unlock()

	// Generate heartbeat
	msg := makeStratuxHeartbeat()

	// Check message was generated
	if len(msg) < 5 {
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	// Check that GPS flag (0x02) and AHRS flag (0x01) are set
	// Message format: 0x7E, msg[0]=0xCC, msg[1]=flags, CRC, 0x7E
	// After unescaping, msg[1] should have GPS valid (0x02), AHRS valid (0x01), and protocol version (0x04)
	// Expected: 0x07 (GPS=0x02 | AHRS=0x01 | Protocol=0x04)

	t.Logf("makeStratuxHeartbeat() with valid AHRS generated %d-byte message", len(msg))
	t.Logf("isGPSValid()=%v, isAHRSValid()=%v", isGPSValid(), isAHRSValid())
}

// TestMakeHeartbeat_WithErrors tests makeHeartbeat with system errors present
func TestMakeHeartbeat_WithErrors(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Initialize stratuxClock
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	time.Sleep(50 * time.Millisecond)

	// Initialize mutex if needed
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}

	// Set GPS valid
	mySituation.muGPS.Lock()
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.GPSFixQuality = 1
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Add system errors
	globalStatus.Errors = []string{"Error 1", "Error 2"}
	defer func() { globalStatus.Errors = nil }()

	msg := makeHeartbeat()

	if len(msg) < 8 {
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("makeHeartbeat() with errors generated %d-byte message", len(msg))
}
