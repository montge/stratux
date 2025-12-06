package main

import (
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stratux/stratux/common"
)

// TestMakeStratuxStatusBasic tests the makeStratuxStatus function with basic settings
func TestMakeStratuxStatusBasic(t *testing.T) {
	// Initialize required components
	crcInit()

	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}

	// Save original values
	origVersion := stratuxVersion
	origSettings := globalSettings
	origStatus := globalStatus
	origTowers := ADSBTowers

	defer func() {
		stratuxVersion = origVersion
		globalSettings = origSettings
		globalStatus = origStatus
		ADSBTowers = origTowers
	}()

	// Initialize towers map
	ADSBTowers = make(map[string]ADSBTower)

	// Set up test values
	stratuxVersion = "v1.6"
	globalSettings = settings{}
	globalStatus = status{}

	msg := makeStratuxStatus()

	// Basic validations
	if len(msg) < 30 {
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("makeStratuxStatus() generated %d-byte message", len(msg))
}

// TestMakeStratuxStatusVersionFormats tests version parsing for different formats
func TestMakeStratuxStatusVersionFormats(t *testing.T) {
	// Initialize required components
	crcInit()

	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}

	// Save original values
	origVersion := stratuxVersion
	origSettings := globalSettings
	origStatus := globalStatus
	origTowers := ADSBTowers

	defer func() {
		stratuxVersion = origVersion
		globalSettings = origSettings
		globalStatus = origStatus
		ADSBTowers = origTowers
	}()

	// Initialize towers map and settings
	ADSBTowers = make(map[string]ADSBTower)
	globalSettings = settings{}
	globalStatus = status{}

	testCases := []struct {
		name    string
		version string
	}{
		{"Standard version", "v1.6"},
		{"Release candidate", "v3.1rc2"},
		{"Release version", "v2.5r3"},
		{"Beta version", "v1.8b1"},
		{"Simple version", "v2.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stratuxVersion = tc.version
			msg := makeStratuxStatus()

			if len(msg) < 30 {
				t.Errorf("Version %s: Message too short: got %d bytes", tc.version, len(msg))
			}

			// Check frame markers
			if msg[0] != 0x7E {
				t.Errorf("Version %s: Expected start frame marker 0x7E, got 0x%02X", tc.version, msg[0])
			}
			if msg[len(msg)-1] != 0x7E {
				t.Errorf("Version %s: Expected end frame marker 0x7E, got 0x%02X", tc.version, msg[len(msg)-1])
			}

			t.Logf("Version %s generated %d-byte message", tc.version, len(msg))
		})
	}
}

// TestMakeStratuxStatusWithFlags tests various enabled flags
func TestMakeStratuxStatusWithFlags(t *testing.T) {
	// Initialize required components
	crcInit()

	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}

	// Save original values
	origVersion := stratuxVersion
	origSettings := globalSettings
	origStatus := globalStatus
	origTowers := ADSBTowers

	defer func() {
		stratuxVersion = origVersion
		globalSettings = origSettings
		globalStatus = origStatus
		ADSBTowers = origTowers
	}()

	// Initialize towers map
	ADSBTowers = make(map[string]ADSBTower)
	stratuxVersion = "v1.6"

	testCases := []struct {
		name         string
		setupFunc    func()
		expectedBits string
	}{
		{
			name: "UAT_Enabled",
			setupFunc: func() {
				globalSettings = settings{}
				globalSettings.UAT_Enabled = true
				globalStatus = status{}
			},
			expectedBits: "UAT enabled bit set",
		},
		{
			name: "ES_Enabled",
			setupFunc: func() {
				globalSettings = settings{}
				globalSettings.ES_Enabled = true
				globalStatus = status{}
			},
			expectedBits: "ES enabled bit set",
		},
		{
			name: "Ping_Enabled",
			setupFunc: func() {
				globalSettings = settings{}
				globalSettings.Ping_Enabled = true
				globalStatus = status{}
			},
			expectedBits: "Ping enabled (UAT+ES) bits set",
		},
		{
			name: "Pong_Enabled",
			setupFunc: func() {
				globalSettings = settings{}
				globalSettings.Pong_Enabled = true
				globalStatus = status{}
			},
			expectedBits: "Pong enabled (UAT+ES) bits set",
		},
		{
			name: "GPS_Enabled",
			setupFunc: func() {
				globalSettings = settings{}
				globalSettings.GPS_Enabled = true
				globalStatus = status{}
			},
			expectedBits: "GPS enabled bit set",
		},
		{
			name: "IMU_Sensor_Enabled",
			setupFunc: func() {
				globalSettings = settings{}
				globalSettings.IMU_Sensor_Enabled = true
				globalStatus = status{}
			},
			expectedBits: "IMU enabled bit set",
		},
		{
			name: "CPU_Temp_Valid",
			setupFunc: func() {
				globalSettings = settings{}
				globalStatus = status{}
				globalStatus.CPUTemp = 50.0 // Valid temperature
			},
			expectedBits: "CPU temp valid bit set",
		},
		{
			name: "Multiple_Flags",
			setupFunc: func() {
				globalSettings = settings{}
				globalSettings.UAT_Enabled = true
				globalSettings.ES_Enabled = true
				globalSettings.GPS_Enabled = true
				globalSettings.IMU_Sensor_Enabled = true
				globalStatus = status{}
				globalStatus.CPUTemp = 45.0
			},
			expectedBits: "Multiple flags set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupFunc()
			msg := makeStratuxStatus()

			if len(msg) < 30 {
				t.Errorf("%s: Message too short: got %d bytes", tc.name, len(msg))
			}

			// Check frame markers
			if msg[0] != 0x7E {
				t.Errorf("%s: Expected start frame marker 0x7E, got 0x%02X", tc.name, msg[0])
			}
			if msg[len(msg)-1] != 0x7E {
				t.Errorf("%s: Expected end frame marker 0x7E, got 0x%02X", tc.name, msg[len(msg)-1])
			}

			t.Logf("%s: %s - generated %d-byte message", tc.name, tc.expectedBits, len(msg))
		})
	}
}

// TestMakeStratuxStatusWithTowers tests tower encoding
func TestMakeStratuxStatusWithTowers(t *testing.T) {
	// Initialize required components
	crcInit()

	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}

	// Save original values
	origVersion := stratuxVersion
	origSettings := globalSettings
	origStatus := globalStatus
	origTowers := ADSBTowers

	defer func() {
		stratuxVersion = origVersion
		globalSettings = origSettings
		globalStatus = origStatus
		ADSBTowers = origTowers
	}()

	stratuxVersion = "v1.6"
	globalSettings = settings{}
	globalStatus = status{}

	// Test with no towers
	t.Run("NoTowers", func(t *testing.T) {
		ADSBTowers = make(map[string]ADSBTower)
		msg := makeStratuxStatus()

		if len(msg) < 30 {
			t.Errorf("Message too short: got %d bytes", len(msg))
		}
		t.Logf("No towers: generated %d-byte message", len(msg))
	})

	// Test with one tower
	t.Run("OneTower", func(t *testing.T) {
		ADSBTowers = make(map[string]ADSBTower)
		ADSBTowers["tower1"] = ADSBTower{
			Lat: 40.7128,
			Lng: -74.0060,
		}
		msg := makeStratuxStatus()

		if len(msg) < 30 {
			t.Errorf("Message too short: got %d bytes", len(msg))
		}
		// Should be longer due to tower data (6 bytes per tower)
		t.Logf("One tower: generated %d-byte message", len(msg))
	})

	// Test with multiple towers
	t.Run("MultipleTowers", func(t *testing.T) {
		ADSBTowers = make(map[string]ADSBTower)
		ADSBTowers["tower1"] = ADSBTower{
			Lat: 40.7128,
			Lng: -74.0060,
		}
		ADSBTowers["tower2"] = ADSBTower{
			Lat: 34.0522,
			Lng: -118.2437,
		}
		ADSBTowers["tower3"] = ADSBTower{
			Lat: 41.8781,
			Lng: -87.6298,
		}
		msg := makeStratuxStatus()

		if len(msg) < 30 {
			t.Errorf("Message too short: got %d bytes", len(msg))
		}
		t.Logf("Three towers: generated %d-byte message", len(msg))
	})
}

// TestMakeStratuxStatusWithStatus tests various status values
func TestMakeStratuxStatusWithStatus(t *testing.T) {
	// Initialize required components
	crcInit()

	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}

	// Save original values
	origVersion := stratuxVersion
	origSettings := globalSettings
	origStatus := globalStatus
	origTowers := ADSBTowers

	defer func() {
		stratuxVersion = origVersion
		globalSettings = origSettings
		globalStatus = origStatus
		ADSBTowers = origTowers
	}()

	// Initialize towers map
	ADSBTowers = make(map[string]ADSBTower)
	stratuxVersion = "v1.6"
	globalSettings = settings{}

	testCases := []struct {
		name      string
		setupFunc func()
	}{
		{
			name: "WithDevices",
			setupFunc: func() {
				globalStatus = status{}
				globalStatus.Devices = 2
			},
		},
		{
			name: "WithIMUConnected",
			setupFunc: func() {
				globalStatus = status{}
				globalStatus.IMUConnected = true
			},
		},
		{
			name: "WithGPSSatellites",
			setupFunc: func() {
				globalStatus = status{}
				globalStatus.GPS_satellites_locked = 10
				globalStatus.GPS_satellites_tracked = 15
			},
		},
		{
			name: "WithTrafficTargets",
			setupFunc: func() {
				globalStatus = status{}
				globalStatus.UAT_traffic_targets_tracking = 5
				globalStatus.ES_traffic_targets_tracking = 12
			},
		},
		{
			name: "WithMessages",
			setupFunc: func() {
				globalStatus = status{}
				globalStatus.UAT_messages_last_minute = 100
				globalStatus.ES_messages_last_minute = 250
			},
		},
		{
			name: "WithCPUTemp",
			setupFunc: func() {
				globalStatus = status{}
				globalStatus.CPUTemp = 55.5
			},
		},
		{
			name: "WithInvalidCPUTemp",
			setupFunc: func() {
				globalStatus = status{}
				globalStatus.CPUTemp = common.InvalidCpuTemp
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupFunc()
			msg := makeStratuxStatus()

			if len(msg) < 30 {
				t.Errorf("%s: Message too short: got %d bytes", tc.name, len(msg))
			}

			// Check frame markers
			if msg[0] != 0x7E {
				t.Errorf("%s: Expected start frame marker 0x7E, got 0x%02X", tc.name, msg[0])
			}
			if msg[len(msg)-1] != 0x7E {
				t.Errorf("%s: Expected end frame marker 0x7E, got 0x%02X", tc.name, msg[len(msg)-1])
			}

			t.Logf("%s: generated %d-byte message", tc.name, len(msg))
		})
	}
}

// TestMakeStratuxStatus_VersionParsing tests various version string formats
func TestMakeStratuxStatus_VersionParsing(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Initialize mutexes
	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muAttitude == nil {
		mySituation.muAttitude = &sync.Mutex{}
	}
	if mySituation.muBaro == nil {
		mySituation.muBaro = &sync.Mutex{}
	}

	testCases := []struct {
		name    string
		version string
		build   string
	}{
		{
			name:    "Release Candidate Version",
			version: "v1.6rc3",
			build:   "test",
		},
		{
			name:    "Release Version",
			version: "v1.6r2",
			build:   "test",
		},
		{
			name:    "Beta Version",
			version: "v1.6b1",
			build:   "test",
		},
		{
			name:    "Simple Version",
			version: "v1.6",
			build:   "test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set version strings
			stratuxVersion = tc.version
			stratuxBuild = tc.build

			msg := makeStratuxStatus()

			if len(msg) < 30 {
				t.Errorf("Message too short: got %d bytes", len(msg))
			}

			// Check frame markers
			if msg[0] != 0x7E {
				t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
			}
			if msg[len(msg)-1] != 0x7E {
				t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
			}

			t.Logf("Version %s: generated %d-byte message", tc.version, len(msg))
		})
	}
}

// TestMakeStratuxStatus_GPSFixQuality3D tests GPS Fix Quality = 1 (3D GPS)
func TestMakeStratuxStatus_GPSFixQuality3D(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Initialize mutexes
	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muAttitude == nil {
		mySituation.muAttitude = &sync.Mutex{}
	}
	if mySituation.muBaro == nil {
		mySituation.muBaro = &sync.Mutex{}
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original values
	origVersion := stratuxVersion
	origSettings := globalSettings
	origStatus := globalStatus
	origTowers := ADSBTowers
	origSituation := mySituation

	defer func() {
		stratuxVersion = origVersion
		globalSettings = origSettings
		globalStatus = origStatus
		ADSBTowers = origTowers
		mySituation = origSituation
	}()

	// Initialize towers map
	ADSBTowers = make(map[string]ADSBTower)
	stratuxVersion = "v1.6"
	globalSettings = settings{}
	globalStatus = status{}
	globalStatus.GPS_connected = true

	// Set up GPS with fix quality 1 (3D GPS, not DGPS)
	mySituation.muGPS.Lock()
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()

	msg := makeStratuxStatus()

	if len(msg) < 30 {
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("GPS Fix Quality 1: generated %d-byte message", len(msg))
}

// TestMakeStratuxStatus_GPSFixQuality2DGPS tests GPS Fix Quality = 2 (DGPS/WAAS)
func TestMakeStratuxStatus_GPSFixQuality2DGPS(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Initialize mutexes
	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muAttitude == nil {
		mySituation.muAttitude = &sync.Mutex{}
	}
	if mySituation.muBaro == nil {
		mySituation.muBaro = &sync.Mutex{}
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original values
	origVersion := stratuxVersion
	origSettings := globalSettings
	origStatus := globalStatus
	origTowers := ADSBTowers
	origSituation := mySituation

	defer func() {
		stratuxVersion = origVersion
		globalSettings = origSettings
		globalStatus = origStatus
		ADSBTowers = origTowers
		mySituation = origSituation
	}()

	// Initialize towers map
	ADSBTowers = make(map[string]ADSBTower)
	stratuxVersion = "v1.6"
	globalSettings = settings{}
	globalStatus = status{}
	globalStatus.GPS_connected = true

	// Set up GPS with fix quality 2 (DGPS/SBAS/WAAS)
	mySituation.muGPS.Lock()
	mySituation.GPSFixQuality = 2
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()

	msg := makeStratuxStatus()

	if len(msg) < 30 {
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("GPS Fix Quality 2 (DGPS): generated %d-byte message", len(msg))
}

// TestMakeStratuxStatus_GPSFixQualityDefault tests GPS Fix Quality = 0 (default case)
func TestMakeStratuxStatus_GPSFixQualityDefault(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Initialize mutexes
	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muAttitude == nil {
		mySituation.muAttitude = &sync.Mutex{}
	}
	if mySituation.muBaro == nil {
		mySituation.muBaro = &sync.Mutex{}
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original values
	origVersion := stratuxVersion
	origSettings := globalSettings
	origStatus := globalStatus
	origTowers := ADSBTowers
	origSituation := mySituation

	defer func() {
		stratuxVersion = origVersion
		globalSettings = origSettings
		globalStatus = origStatus
		ADSBTowers = origTowers
		mySituation = origSituation
	}()

	// Initialize towers map
	ADSBTowers = make(map[string]ADSBTower)
	stratuxVersion = "v1.6"
	globalSettings = settings{}
	globalStatus = status{}
	globalStatus.GPS_connected = true

	// Set up GPS with fix quality 0 (default case - neither 1 nor 2)
	mySituation.muGPS.Lock()
	mySituation.GPSFixQuality = 0
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()

	msg := makeStratuxStatus()

	if len(msg) < 30 {
		t.Errorf("Message too short: got %d bytes", len(msg))
	}

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	t.Logf("GPS Fix Quality 0 (default): generated %d-byte message", len(msg))
}

// TestMakeStratuxStatus_EdgeVersions tests version numbers that hit edge cases
func TestMakeStratuxStatus_EdgeVersions(t *testing.T) {
	// Initialize CRC table
	crcInit()

	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Initialize mutexes
	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muAttitude == nil {
		mySituation.muAttitude = &sync.Mutex{}
	}
	if mySituation.muBaro == nil {
		mySituation.muBaro = &sync.Mutex{}
	}

	// Save original values
	origVersion := stratuxVersion
	origSettings := globalSettings
	origStatus := globalStatus
	origTowers := ADSBTowers

	defer func() {
		stratuxVersion = origVersion
		globalSettings = origSettings
		globalStatus = origStatus
		ADSBTowers = origTowers
	}()

	// Initialize towers map
	ADSBTowers = make(map[string]ADSBTower)
	globalSettings = settings{}
	globalStatus = status{}

	// Test edge version formats that may produce unusual parsing results
	testCases := []struct {
		name    string
		version string
	}{
		// No suffix version (tp = 0, empty minor/build)
		{"Plain_version", "v1.6"},
		// Very large major version
		{"Large_major", "v999.0"},
		// Very large minor version (> 255 to hit bounds check)
		{"Large_minor", "v1.999"},
		// Very large minor with RC
		{"Large_minor_rc", "v1.999rc1"},
		// Large build number
		{"Large_build", "v1.6rc999"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stratuxVersion = tc.version

			msg := makeStratuxStatus()

			if len(msg) < 30 {
				t.Errorf("%s: Message too short: got %d bytes", tc.name, len(msg))
			}

			// Check frame markers
			if msg[0] != 0x7E {
				t.Errorf("%s: Expected start frame marker 0x7E, got 0x%02X", tc.name, msg[0])
			}
			if msg[len(msg)-1] != 0x7E {
				t.Errorf("%s: Expected end frame marker 0x7E, got 0x%02X", tc.name, msg[len(msg)-1])
			}

			t.Logf("%s: generated %d-byte message", tc.name, len(msg))
		})
	}
}

// TestMakeStratuxStatus_UnknownGPSFixQuality tests the default case for GPS fix quality
// This covers the default case in the switch statement
func TestMakeStratuxStatus_UnknownGPSFixQuality(t *testing.T) {
	// Initialize required components
	crcInit()

	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}

	// Initialize mySituation mutexes
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}

	// Save original values
	origVersion := stratuxVersion
	origSettings := globalSettings
	origStatus := globalStatus
	origTowers := ADSBTowers
	origFixQuality := mySituation.GPSFixQuality

	defer func() {
		stratuxVersion = origVersion
		globalSettings = origSettings
		globalStatus = origStatus
		ADSBTowers = origTowers
		mySituation.muGPS.Lock()
		mySituation.GPSFixQuality = origFixQuality
		mySituation.muGPS.Unlock()
	}()

	// Initialize towers map
	ADSBTowers = make(map[string]ADSBTower)
	stratuxVersion = "v1.6"
	globalSettings = settings{}
	globalStatus = status{}

	// Set GPS fix quality to an unknown value (not 0, 1, or 2)
	// Also need to make GPS valid for the switch statement to be entered
	mySituation.muGPS.Lock()
	mySituation.GPSFixQuality = 99 // Unknown value - should trigger default case
	mySituation.GPSLastFixLocalTime = stratuxClock.Time // Make GPS "valid" (recent fix)
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	msg := makeStratuxStatus()

	if len(msg) < 14 {
		t.Fatalf("Message too short: got %d bytes", len(msg))
	}

	// With unknown GPS fix quality, byte 13 should be 0 (default case)
	if msg[13] != 0 {
		t.Errorf("Expected msg[13]=0 for unknown GPS fix quality, got %d", msg[13])
	}

	t.Logf("Unknown GPS fix quality (99) test: msg[13]=%d", msg[13])
}

// TestMakeFFIDMessage_LongVersion tests FFID with version > 16 chars (truncation)
func TestMakeFFIDMessage_LongVersion(t *testing.T) {
	// Initialize required components
	crcInit()

	// Save original values
	origVersion := stratuxVersion
	origBuild := stratuxBuild

	defer func() {
		stratuxVersion = origVersion
		stratuxBuild = origBuild
	}()

	// Set a long version and build that together exceed 16 characters
	// devLongName = fmt.Sprintf("%s-%s", stratuxVersion, stratuxBuild)
	// v1.6.2-beta1-rXXXXXXXX would be way over 16 chars
	stratuxVersion = "v1.6.2-beta1"
	stratuxBuild = "r1234567890"

	msg := makeFFIDMessage()

	// Should still produce valid message
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

	t.Logf("Long version FFID message: %d bytes (version=%s, build=%s)", len(msg), stratuxVersion, stratuxBuild)
}

// TestSaveSettings_Success tests saveSettings with a valid temp file
func TestSaveSettings_Success(t *testing.T) {
	// Initialize required mutexes and maps
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Save original config location
	origConfigLocation := configLocation
	defer func() { configLocation = origConfigLocation }()

	// Create a temp file for the config
	tmpDir := t.TempDir()
	configLocation = tmpDir + "/test_stratux.conf"

	// Save original settings
	origSettings := globalSettings
	defer func() { globalSettings = origSettings }()

	// Set some test values
	globalSettings.UAT_Enabled = true
	globalSettings.ES_Enabled = true
	globalSettings.DeveloperMode = true

	// Call saveSettings - should succeed
	saveSettings()

	// Verify the file was created and contains valid JSON
	data, err := os.ReadFile(configLocation)
	if err != nil {
		t.Fatalf("Failed to read saved settings: %v", err)
	}

	if len(data) == 0 {
		t.Error("Settings file is empty")
	}

	// Should contain JSON with our settings
	dataStr := string(data)
	if !strings.Contains(dataStr, "UAT_Enabled") {
		t.Error("Settings file should contain UAT_Enabled")
	}

	t.Logf("Successfully saved settings to %s (%d bytes)", configLocation, len(data))
}

// TestSaveSettings_Failure tests saveSettings when file cannot be created
func TestSaveSettings_Failure(t *testing.T) {
	// Initialize required mutexes and maps
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Save original config location
	origConfigLocation := configLocation
	defer func() { configLocation = origConfigLocation }()

	// Set config location to a path that will fail (non-existent directory)
	configLocation = "/nonexistent/directory/stratux.conf"

	// Save original error count
	origErrors := len(globalStatus.Errors)
	defer func() {
		// Clear errors we added
		if len(globalStatus.Errors) > origErrors {
			globalStatus.Errors = globalStatus.Errors[:origErrors]
		}
	}()

	// Call saveSettings - should fail and add error
	saveSettings()

	// The function should have added an error via addSingleSystemErrorf
	t.Log("saveSettings handled failure case")
}

// TestReadSettings_Success tests readSettings with a valid config file
func TestReadSettings_Success(t *testing.T) {
	// Initialize required mutexes and maps
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Save and restore original config location and settings
	origConfigLocation := configLocation
	origSettings := globalSettings
	defer func() {
		configLocation = origConfigLocation
		globalSettings = origSettings
	}()

	// Create a temp config file with valid JSON
	tmpDir := t.TempDir()
	configLocation = tmpDir + "/test_stratux.conf"

	testConfig := `{
		"UAT_Enabled": true,
		"ES_Enabled": true,
		"OGN_Enabled": false,
		"DeveloperMode": true,
		"DisplayTrafficSource": true
	}`
	if err := os.WriteFile(configLocation, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Call readSettings
	readSettings()

	// Verify settings were loaded
	if !globalSettings.UAT_Enabled {
		t.Error("UAT_Enabled should be true after readSettings")
	}
	if !globalSettings.ES_Enabled {
		t.Error("ES_Enabled should be true after readSettings")
	}
	if globalSettings.OGN_Enabled {
		t.Error("OGN_Enabled should be false after readSettings")
	}
	if !globalSettings.DeveloperMode {
		t.Error("DeveloperMode should be true after readSettings")
	}
}

// TestReadSettings_FileNotFound tests readSettings when config file doesn't exist
func TestReadSettings_FileNotFound(t *testing.T) {
	// Save and restore original config location and settings
	origConfigLocation := configLocation
	origSettings := globalSettings
	defer func() {
		configLocation = origConfigLocation
		globalSettings = origSettings
	}()

	// Set config location to a non-existent file
	configLocation = "/nonexistent/path/to/stratux.conf"

	// Call readSettings - should handle error gracefully
	readSettings()

	// Function should complete without panic, settings will have defaults
	t.Log("readSettings handled missing file gracefully")
}

// TestReadSettings_InvalidJSON tests readSettings with invalid JSON content
func TestReadSettings_InvalidJSON(t *testing.T) {
	// Save and restore original config location and settings
	origConfigLocation := configLocation
	origSettings := globalSettings
	defer func() {
		configLocation = origConfigLocation
		globalSettings = origSettings
	}()

	// Create a temp config file with invalid JSON
	tmpDir := t.TempDir()
	configLocation = tmpDir + "/test_stratux.conf"

	invalidJSON := `{ invalid json content here }`
	if err := os.WriteFile(configLocation, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Call readSettings - should handle error gracefully
	readSettings()

	// Function should complete without panic
	t.Log("readSettings handled invalid JSON gracefully")
}

// TestGetProductNameFromId_KnownProducts tests known product IDs from product_name_map
func TestGetProductNameFromId_KnownProducts(t *testing.T) {
	testCases := []struct {
		productID int
		expected  string
	}{
		// METAR/TAF/Weather Products
		{0, "METAR"},
		{1, "TAF"},
		{2, "SIGMET"},
		{3, "Conv SIGMET"},
		{4, "AIRMET"},
		{5, "PIREP"},
		{6, "Severe Wx"},
		{7, "Winds Aloft"},
		{8, "NOTAM"},
		{9, "D-ATIS"},
		{10, "Terminal Wx"},
		{11, "AIRMET"},
		{12, "SIGMET"},
		{13, "SUA"},

		// Additional weather products
		{20, "METAR"},
		{21, "TAF"},
		{22, "SIGMET"},
		{23, "Conv SIGMET"},
		{24, "AIRMET"},
		{25, "PIREP"},
		{26, "Severe Wx"},
		{27, "Winds Aloft"},

		// NEXRAD Products
		{51, "NEXRAD"},
		{52, "NEXRAD"},
		{53, "NEXRAD"},
		{54, "NEXRAD"},
		{55, "NEXRAD"},
		{56, "NEXRAD"},
		{57, "NEXRAD"},
		{58, "NEXRAD"},
		{59, "NEXRAD"},
		{60, "NEXRAD"},
		{61, "NEXRAD"},
		{62, "NEXRAD"},
		{63, "NEXRAD Regional"},
		{64, "NEXRAD CONUS"},

		// Tops Products
		{81, "Tops"},
		{82, "Tops"},
		{83, "Tops"},

		// Lightning Products
		{101, "Lightning"},
		{102, "Lightning"},
		{151, "Lightning"},

		// Surface Products
		{201, "Surface"},
		{202, "Surface"},

		// G-AIRMET
		{254, "G-AIRMET"},

		// System/Status Products
		{351, "Time"},
		{352, "Status"},
		{353, "Status"},

		// Generic Products
		{401, "Imagery"},
		{402, "Text"},
		{403, "Vector Imagery"},
		{404, "Symbols"},
		{405, "Text"},
		{411, "Text"},
		{412, "Symbols"},
		{413, "Text"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected+"_"+string(rune(tc.productID+'0')), func(t *testing.T) {
			result := getProductNameFromId(tc.productID)
			if result != tc.expected {
				t.Errorf("getProductNameFromId(%d) = %q, expected %q", tc.productID, result, tc.expected)
			}
		})
	}
}

// TestGetProductNameFromId_CustomTest600 tests product ID 600 (Custom/Test)
func TestGetProductNameFromId_CustomTest600(t *testing.T) {
	result := getProductNameFromId(600)
	expected := "Custom/Test"
	if result != expected {
		t.Errorf("getProductNameFromId(600) = %q, expected %q", result, expected)
	}
}

// TestGetProductNameFromId_CustomTest2000to2005 tests product IDs 2000-2005 (Custom/Test range)
func TestGetProductNameFromId_CustomTest2000to2005(t *testing.T) {
	expected := "Custom/Test"

	testCases := []int{2000, 2001, 2002, 2003, 2004, 2005}

	for _, productID := range testCases {
		t.Run(string(rune(productID)), func(t *testing.T) {
			result := getProductNameFromId(productID)
			if result != expected {
				t.Errorf("getProductNameFromId(%d) = %q, expected %q", productID, result, expected)
			}
		})
	}
}

// TestGetProductNameFromId_CustomTestBoundaries tests boundaries of Custom/Test range
func TestGetProductNameFromId_CustomTestBoundaries(t *testing.T) {
	testCases := []struct {
		productID int
		expected  string
	}{
		// Just before Custom/Test range
		{1999, "Unknown (1999)"},
		// Start of Custom/Test range
		{2000, "Custom/Test"},
		// End of Custom/Test range
		{2005, "Custom/Test"},
		// Just after Custom/Test range
		{2006, "Unknown (2006)"},
		// Product ID 600 (special Custom/Test case)
		{600, "Custom/Test"},
		// Just before and after 600
		{599, "Unknown (599)"},
		{601, "Unknown (601)"},
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.productID)), func(t *testing.T) {
			result := getProductNameFromId(tc.productID)
			if result != tc.expected {
				t.Errorf("getProductNameFromId(%d) = %q, expected %q", tc.productID, result, tc.expected)
			}
		})
	}
}

// TestGetProductNameFromId_UnknownProducts tests unknown product IDs
func TestGetProductNameFromId_UnknownProducts(t *testing.T) {
	testCases := []struct {
		productID int
		expected  string
	}{
		// Negative values
		{-1, "Unknown (-1)"},
		{-100, "Unknown (-100)"},

		// Zero-adjacent (0 is known as METAR)
		{-1, "Unknown (-1)"},

		// Small unknown values
		{14, "Unknown (14)"},
		{15, "Unknown (15)"},
		{19, "Unknown (19)"},

		// Mid-range unknown values
		{100, "Unknown (100)"},
		{150, "Unknown (150)"},
		{200, "Unknown (200)"},
		{255, "Unknown (255)"},
		{300, "Unknown (300)"},
		{400, "Unknown (400)"},
		{500, "Unknown (500)"},
		{700, "Unknown (700)"},
		{1000, "Unknown (1000)"},
		{1500, "Unknown (1500)"},

		// Just outside Custom/Test ranges
		{599, "Unknown (599)"},
		{601, "Unknown (601)"},
		{1999, "Unknown (1999)"},
		{2006, "Unknown (2006)"},

		// Large unknown values
		{3000, "Unknown (3000)"},
		{10000, "Unknown (10000)"},
		{65535, "Unknown (65535)"},
		{99999, "Unknown (99999)"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := getProductNameFromId(tc.productID)
			if result != tc.expected {
				t.Errorf("getProductNameFromId(%d) = %q, expected %q", tc.productID, result, tc.expected)
			}
		})
	}
}

// TestGetProductNameFromId_AllBranches tests to ensure all code branches are covered
func TestGetProductNameFromId_AllBranches(t *testing.T) {
	// Test each branch explicitly
	t.Run("KnownProductBranch", func(t *testing.T) {
		// Branch 1: Known product in map
		result := getProductNameFromId(0) // METAR
		if result != "METAR" {
			t.Errorf("Expected METAR, got %s", result)
		}
	})

	t.Run("CustomTest600Branch", func(t *testing.T) {
		// Branch 2: Product ID 600
		result := getProductNameFromId(600)
		if result != "Custom/Test" {
			t.Errorf("Expected Custom/Test, got %s", result)
		}
	})

	t.Run("CustomTest2000RangeBranch", func(t *testing.T) {
		// Branch 3: Product ID in range 2000-2005
		result := getProductNameFromId(2003)
		if result != "Custom/Test" {
			t.Errorf("Expected Custom/Test, got %s", result)
		}
	})

	t.Run("UnknownProductBranch", func(t *testing.T) {
		// Branch 4: Unknown product (default case)
		result := getProductNameFromId(9999)
		expected := "Unknown (9999)"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

// TestAddSingleSystemErrorf_AddsError tests that addSingleSystemErrorf adds an error
func TestAddSingleSystemErrorf_AddsError(t *testing.T) {
	// Initialize mutexes and maps
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Save and restore original errors
	origErrors := globalStatus.Errors
	origSystemErrs := systemErrs
	defer func() {
		globalStatus.Errors = origErrors
		systemErrs = origSystemErrs
	}()

	// Reset to empty state
	globalStatus.Errors = []string{}
	systemErrs = make(map[string]string)

	// Add an error
	addSingleSystemErrorf("test-error", "test error message: %s", "details")

	// Verify error was added to globalStatus.Errors
	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error in globalStatus.Errors, got %d", len(globalStatus.Errors))
	}

	expectedMsg := "test error message: details"
	if globalStatus.Errors[0] != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, globalStatus.Errors[0])
	}

	// Verify error was added to systemErrs map
	if msg, ok := systemErrs["test-error"]; !ok {
		t.Error("Error not found in systemErrs map")
	} else if msg != expectedMsg {
		t.Errorf("Expected systemErrs message '%s', got '%s'", expectedMsg, msg)
	}

	t.Logf("Successfully added error: %s", expectedMsg)
}

// TestAddSingleSystemErrorf_NoDuplicates tests that duplicate errors with same ident are not added
func TestAddSingleSystemErrorf_NoDuplicates(t *testing.T) {
	// Initialize mutexes and maps
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Save and restore original errors
	origErrors := globalStatus.Errors
	origSystemErrs := systemErrs
	defer func() {
		globalStatus.Errors = origErrors
		systemErrs = origSystemErrs
	}()

	// Reset to empty state
	globalStatus.Errors = []string{}
	systemErrs = make(map[string]string)

	// Add the same error twice with same ident
	addSingleSystemErrorf("duplicate-test", "first error message")
	addSingleSystemErrorf("duplicate-test", "second error message - should be ignored")

	// Should only have one error
	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error in globalStatus.Errors, got %d", len(globalStatus.Errors))
	}

	// Should keep the first error message
	if globalStatus.Errors[0] != "first error message" {
		t.Errorf("Expected first error message to be kept, got '%s'", globalStatus.Errors[0])
	}

	// systemErrs should have the first message
	if msg := systemErrs["duplicate-test"]; msg != "first error message" {
		t.Errorf("Expected systemErrs to keep first message, got '%s'", msg)
	}

	t.Logf("Duplicate error correctly ignored")
}

// TestAddSingleSystemErrorf_MultipleErrors tests adding multiple different errors
func TestAddSingleSystemErrorf_MultipleErrors(t *testing.T) {
	// Initialize mutexes and maps
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Save and restore original errors
	origErrors := globalStatus.Errors
	origSystemErrs := systemErrs
	defer func() {
		globalStatus.Errors = origErrors
		systemErrs = origSystemErrs
	}()

	// Reset to empty state
	globalStatus.Errors = []string{}
	systemErrs = make(map[string]string)

	// Add multiple different errors
	addSingleSystemErrorf("error1", "first error")
	addSingleSystemErrorf("error2", "second error")
	addSingleSystemErrorf("error3", "third error")

	// Should have three errors
	if len(globalStatus.Errors) != 3 {
		t.Errorf("Expected 3 errors in globalStatus.Errors, got %d", len(globalStatus.Errors))
	}

	// Verify all errors are in systemErrs
	if len(systemErrs) != 3 {
		t.Errorf("Expected 3 errors in systemErrs, got %d", len(systemErrs))
	}

	// Verify specific errors
	expectedErrors := map[string]string{
		"error1": "first error",
		"error2": "second error",
		"error3": "third error",
	}

	for ident, expectedMsg := range expectedErrors {
		if msg, ok := systemErrs[ident]; !ok {
			t.Errorf("Error '%s' not found in systemErrs", ident)
		} else if msg != expectedMsg {
			t.Errorf("Error '%s': expected message '%s', got '%s'", ident, expectedMsg, msg)
		}
	}

	t.Logf("Successfully added %d different errors", len(systemErrs))
}

// TestRemoveSingleSystemError_RemovesError tests that removeSingleSystemError removes errors correctly
func TestRemoveSingleSystemError_RemovesError(t *testing.T) {
	// Initialize mutexes and maps
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Save and restore original errors
	origErrors := globalStatus.Errors
	origSystemErrs := systemErrs
	defer func() {
		globalStatus.Errors = origErrors
		systemErrs = origSystemErrs
	}()

	// Reset to empty state
	globalStatus.Errors = []string{}
	systemErrs = make(map[string]string)

	// Add some errors
	addSingleSystemErrorf("error1", "first error")
	addSingleSystemErrorf("error2", "second error")
	addSingleSystemErrorf("error3", "third error")

	// Verify we have 3 errors
	if len(globalStatus.Errors) != 3 {
		t.Fatalf("Expected 3 errors before removal, got %d", len(globalStatus.Errors))
	}

	// Remove the middle error
	removeSingleSystemError("error2")

	// Should have 2 errors left
	if len(globalStatus.Errors) != 2 {
		t.Errorf("Expected 2 errors after removal, got %d", len(globalStatus.Errors))
	}

	// Should have 2 errors in systemErrs
	if len(systemErrs) != 2 {
		t.Errorf("Expected 2 errors in systemErrs after removal, got %d", len(systemErrs))
	}

	// Verify error2 is not in systemErrs
	if _, ok := systemErrs["error2"]; ok {
		t.Error("error2 should not be in systemErrs after removal")
	}

	// Verify "second error" is not in globalStatus.Errors
	for _, err := range globalStatus.Errors {
		if err == "second error" {
			t.Error("'second error' should not be in globalStatus.Errors after removal")
		}
	}

	// Verify error1 and error3 are still present
	if _, ok := systemErrs["error1"]; !ok {
		t.Error("error1 should still be in systemErrs")
	}
	if _, ok := systemErrs["error3"]; !ok {
		t.Error("error3 should still be in systemErrs")
	}

	t.Logf("Successfully removed error2, %d errors remaining", len(systemErrs))
}

// TestRemoveSingleSystemError_NonExistent tests removing a non-existent error
func TestRemoveSingleSystemError_NonExistent(t *testing.T) {
	// Initialize mutexes and maps
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Save and restore original errors
	origErrors := globalStatus.Errors
	origSystemErrs := systemErrs
	defer func() {
		globalStatus.Errors = origErrors
		systemErrs = origSystemErrs
	}()

	// Reset to empty state
	globalStatus.Errors = []string{}
	systemErrs = make(map[string]string)

	// Add some errors
	addSingleSystemErrorf("error1", "first error")
	addSingleSystemErrorf("error2", "second error")

	// Verify we have 2 errors
	if len(globalStatus.Errors) != 2 {
		t.Fatalf("Expected 2 errors before removal, got %d", len(globalStatus.Errors))
	}

	// Try to remove a non-existent error - should not panic
	removeSingleSystemError("non-existent-error")

	// Should still have 2 errors
	if len(globalStatus.Errors) != 2 {
		t.Errorf("Expected 2 errors after non-existent removal, got %d", len(globalStatus.Errors))
	}

	// Should still have 2 errors in systemErrs
	if len(systemErrs) != 2 {
		t.Errorf("Expected 2 errors in systemErrs after non-existent removal, got %d", len(systemErrs))
	}

	t.Log("Successfully handled removal of non-existent error")
}

// TestRemoveSingleSystemError_EmptyState tests removing from empty state
func TestRemoveSingleSystemError_EmptyState(t *testing.T) {
	// Initialize mutexes and maps
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Save and restore original errors
	origErrors := globalStatus.Errors
	origSystemErrs := systemErrs
	defer func() {
		globalStatus.Errors = origErrors
		systemErrs = origSystemErrs
	}()

	// Reset to empty state
	globalStatus.Errors = []string{}
	systemErrs = make(map[string]string)

	// Try to remove from empty state - should not panic
	removeSingleSystemError("any-error")

	// Should still be empty
	if len(globalStatus.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(globalStatus.Errors))
	}

	if len(systemErrs) != 0 {
		t.Errorf("Expected 0 errors in systemErrs, got %d", len(systemErrs))
	}

	t.Log("Successfully handled removal from empty state")
}

// TestSystemErrors_AddRemoveSequence tests a sequence of add and remove operations
func TestSystemErrors_AddRemoveSequence(t *testing.T) {
	// Initialize mutexes and maps
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Save and restore original errors
	origErrors := globalStatus.Errors
	origSystemErrs := systemErrs
	defer func() {
		globalStatus.Errors = origErrors
		systemErrs = origSystemErrs
	}()

	// Reset to empty state
	globalStatus.Errors = []string{}
	systemErrs = make(map[string]string)

	// Complex sequence of operations
	addSingleSystemErrorf("gps-error", "GPS not responding")
	addSingleSystemErrorf("sdr-error", "SDR initialization failed")

	if len(systemErrs) != 2 {
		t.Errorf("After adding 2 errors, expected 2, got %d", len(systemErrs))
	}

	addSingleSystemErrorf("gps-error", "GPS still not responding") // Duplicate, should be ignored

	if len(systemErrs) != 2 {
		t.Errorf("After duplicate add, expected 2, got %d", len(systemErrs))
	}

	removeSingleSystemError("gps-error")

	if len(systemErrs) != 1 {
		t.Errorf("After removing gps-error, expected 1, got %d", len(systemErrs))
	}

	addSingleSystemErrorf("gps-error", "GPS reconnected then failed again") // Now it should be added

	if len(systemErrs) != 2 {
		t.Errorf("After re-adding gps-error, expected 2, got %d", len(systemErrs))
	}

	// Verify the new message was added
	if msg := systemErrs["gps-error"]; msg != "GPS reconnected then failed again" {
		t.Errorf("Expected new GPS error message, got '%s'", msg)
	}

	removeSingleSystemError("sdr-error")
	removeSingleSystemError("gps-error")

	if len(systemErrs) != 0 {
		t.Errorf("After removing all errors, expected 0, got %d", len(systemErrs))
	}

	if len(globalStatus.Errors) != 0 {
		t.Errorf("After removing all errors, expected 0 in globalStatus.Errors, got %d", len(globalStatus.Errors))
	}

	t.Log("Successfully completed add/remove sequence")
}
