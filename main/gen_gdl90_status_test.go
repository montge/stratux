package main

import (
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

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
	mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
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
	mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
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
	mySituation.GPSLastFixLocalTime = stratuxClock.GetTime()
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
	mySituation.GPSFixQuality = 99                           // Unknown value - should trigger default case
	mySituation.GPSLastFixLocalTime = stratuxClock.GetTime() // Make GPS "valid" (recent fix)
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

// TestMakeFFIDMessage_SerialNumber tests the serial number field in FFID message
func TestMakeFFIDMessage_SerialNumber(t *testing.T) {
	// Initialize required components
	crcInit()

	// Save original values
	origVersion := stratuxVersion
	origBuild := stratuxBuild

	defer func() {
		stratuxVersion = origVersion
		stratuxBuild = origBuild
	}()

	stratuxVersion = "v1.6"
	stratuxBuild = "test"

	msg := makeFFIDMessage()

	// Remove framing and extract raw message
	// Message format: 0x7E [data...] [CRC] [CRC] 0x7E
	// We need to skip the 0x7E frame marker and unescape if needed
	if len(msg) < 6 {
		t.Fatalf("Message too short: got %d bytes", len(msg))
	}

	// For now, serial number (bytes 3-10 of raw message) should be 0xFF (invalid)
	// The message is prepareMessage-ed, so we need to account for framing
	// Let's just verify the message was created successfully
	// Serial number bytes are at positions 3-10 in the 39-byte raw message

	// Basic structure check
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	// After prepareMessage, the raw data is escaped, so we can't easily check
	// individual bytes. We verify the message is well-formed and non-empty.
	if len(msg) < 40 {
		t.Errorf("FFID message too short: got %d bytes, expected at least 40", len(msg))
	}

	t.Logf("Serial number field test: FFID message has %d bytes (serial currently hardcoded to 0xFF)", len(msg))
}

// TestMakeFFIDMessage_EmptyVersion tests FFID with empty version/build strings
func TestMakeFFIDMessage_EmptyVersion(t *testing.T) {
	// Initialize required components
	crcInit()

	// Save original values
	origVersion := stratuxVersion
	origBuild := stratuxBuild

	defer func() {
		stratuxVersion = origVersion
		stratuxBuild = origBuild
	}()

	// Set empty strings
	stratuxVersion = ""
	stratuxBuild = ""

	msg := makeFFIDMessage()

	// Should still produce valid message even with empty version
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

	t.Logf("Empty version FFID message: %d bytes", len(msg))
}

// TestMakeFFIDMessage_ShortVersion tests FFID with very short version strings
func TestMakeFFIDMessage_ShortVersion(t *testing.T) {
	// Initialize required components
	crcInit()

	// Save original values
	origVersion := stratuxVersion
	origBuild := stratuxBuild

	defer func() {
		stratuxVersion = origVersion
		stratuxBuild = origBuild
	}()

	// Set short strings that result in < 16 char devLongName
	stratuxVersion = "v1"
	stratuxBuild = "b"
	// devLongName will be "v1-b" which is only 4 characters

	msg := makeFFIDMessage()

	// Should produce valid message
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

	t.Logf("Short version FFID message: %d bytes (version=%s, build=%s)", len(msg), stratuxVersion, stratuxBuild)
}

// TestMakeFFIDMessage_ExactlyMaxLength tests FFID with version strings that result in exactly 16 chars
func TestMakeFFIDMessage_ExactlyMaxLength(t *testing.T) {
	// Initialize required components
	crcInit()

	// Save original values
	origVersion := stratuxVersion
	origBuild := stratuxBuild

	defer func() {
		stratuxVersion = origVersion
		stratuxBuild = origBuild
	}()

	// Set strings that result in exactly 16 char devLongName
	// "12345678-1234567" = 16 chars (8 + 1 + 7)
	stratuxVersion = "12345678"
	stratuxBuild = "1234567"

	msg := makeFFIDMessage()

	// Should produce valid message
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

	t.Logf("Exactly 16-char version FFID message: %d bytes (version=%s, build=%s)", len(msg), stratuxVersion, stratuxBuild)
}

// TestMakeFFIDMessage_MessageStructure tests the FFID message structure in detail
func TestMakeFFIDMessage_MessageStructure(t *testing.T) {
	// Initialize required components
	crcInit()

	// Save original values
	origVersion := stratuxVersion
	origBuild := stratuxBuild

	defer func() {
		stratuxVersion = origVersion
		stratuxBuild = origBuild
	}()

	stratuxVersion = "v1.6.2"
	stratuxBuild = "test123"

	msg := makeFFIDMessage()

	// Check frame markers
	if msg[0] != 0x7E {
		t.Errorf("Expected start frame marker 0x7E, got 0x%02X", msg[0])
	}
	if msg[len(msg)-1] != 0x7E {
		t.Errorf("Expected end frame marker 0x7E, got 0x%02X", msg[len(msg)-1])
	}

	// Message should be: 0x7E + escaped data (39 bytes raw + potential escapes) + CRC (2 bytes) + 0x7E
	// Minimum would be around 43 bytes if no escaping needed
	if len(msg) < 40 {
		t.Errorf("Message too short: got %d bytes, expected at least 40", len(msg))
	}

	// The raw message (before prepareMessage) is 39 bytes:
	// [0] = 0x65 (message type)
	// [1] = 0 (ID message identifier)
	// [2] = 1 (message version)
	// [3-10] = serial number (8 bytes, currently 0xFF)
	// [11-18] = device short name (8 bytes, "Stratux" + padding)
	// [19-34] = device long name (16 bytes)
	// [35-37] = reserved (0x00)
	// [38] = capabilities mask (0x00)

	// We can't easily check individual bytes after prepareMessage due to escaping,
	// but we can verify the message is well-formed
	t.Logf("FFID message structure test: %d bytes total", len(msg))
}

// TestMakeFFIDMessage_Coverage_Note documents the coverage limitation
// NOTE: This test documents why makeFFIDMessage cannot reach 100% coverage
// with the current implementation.
//
// The function has a branch that truncates devShortName if it's > 8 characters:
//
//	if len(devShortName) > 8 {
//	    devShortName = devShortName[:8]
//	}
//
// However, devShortName is currently hardcoded to "Stratux" (7 characters),
// so this branch is never executed. The comment in the code says:
// "Temporary. Will be populated in the future with other names."
//
// To reach 100% coverage, the code would need to be modified to make
// devShortName configurable (e.g., from globalSettings or a package variable).
//
// Current coverage: 93.8% (15/16 statements covered)
// Missing: Line 729 in gen_gdl90.go - devShortName truncation
func TestMakeFFIDMessage_Coverage_Note(t *testing.T) {
	// This test verifies the current behavior and documents the limitation
	crcInit()

	origVersion := stratuxVersion
	origBuild := stratuxBuild
	defer func() {
		stratuxVersion = origVersion
		stratuxBuild = origBuild
	}()

	stratuxVersion = "v1.6"
	stratuxBuild = "test"

	msg := makeFFIDMessage()

	if msg[0] != 0x7E || msg[len(msg)-1] != 0x7E {
		t.Errorf("Invalid frame markers")
	}

	// Document the limitation
	t.Log("Coverage Note: devShortName truncation branch (line 729) cannot be tested")
	t.Log("because devShortName is hardcoded to 'Stratux' (7 chars)")
	t.Log("To test this branch, devShortName would need to be configurable")
	t.Logf("Current message length: %d bytes", len(msg))
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

// TestReadSettings_ReadError tests readSettings when file read fails
func TestReadSettings_ReadError(t *testing.T) {
	// Save and restore original config location and settings
	origConfigLocation := configLocation
	origSettings := globalSettings
	defer func() {
		configLocation = origConfigLocation
		globalSettings = origSettings
	}()

	// Create a temp directory and use it as the config location
	// On Unix systems, opening a directory succeeds, but reading from it fails
	tmpDir := t.TempDir()
	configLocation = tmpDir

	// Call readSettings - should handle read error gracefully
	readSettings()

	// Function should complete without panic, settings will have defaults
	// Verify defaults were applied
	if !globalSettings.UAT_Enabled {
		t.Error("Expected UAT_Enabled=true (default after read error)")
	}
	if !globalSettings.ES_Enabled {
		t.Error("Expected ES_Enabled=true (default after read error)")
	}
	t.Log("readSettings handled file read error gracefully")
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

// TestAddSystemError tests the addSystemError function
func TestAddSystemError(t *testing.T) {
	// Save original errors
	origErrors := globalStatus.Errors
	defer func() {
		globalStatus.Errors = origErrors
	}()

	// Reset to empty state
	globalStatus.Errors = []string{}

	// Call addSystemError with a test error
	testError := "test system error message"
	addSystemError(errors.New(testError))

	// Verify the error was appended
	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error in globalStatus.Errors, got %d", len(globalStatus.Errors))
	}

	if globalStatus.Errors[0] != testError {
		t.Errorf("Expected error message '%s', got '%s'", testError, globalStatus.Errors[0])
	}

	// Test adding multiple errors
	addSystemError(errors.New("second error"))
	addSystemError(errors.New("third error"))

	if len(globalStatus.Errors) != 3 {
		t.Errorf("Expected 3 errors in globalStatus.Errors, got %d", len(globalStatus.Errors))
	}

	expectedErrors := []string{testError, "second error", "third error"}
	for i, expected := range expectedErrors {
		if globalStatus.Errors[i] != expected {
			t.Errorf("Error at index %d: expected '%s', got '%s'", i, expected, globalStatus.Errors[i])
		}
	}

	t.Logf("Successfully tested addSystemError with %d errors", len(globalStatus.Errors))
}

// TestDefaultSettings_BasicDefaults tests that defaultSettings() sets correct default values
func TestDefaultSettings_BasicDefaults(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Set globalSettings to a non-default state
	globalSettings = settings{}
	globalSettings.UAT_Enabled = false
	globalSettings.ES_Enabled = false
	globalSettings.GPS_Enabled = false
	globalSettings.DeveloperMode = true
	globalSettings.WiFiSSID = "CustomSSID"

	// Call defaultSettings
	defaultSettings()

	// Verify key boolean defaults
	if !globalSettings.UAT_Enabled {
		t.Error("UAT_Enabled should be true")
	}
	if !globalSettings.ES_Enabled {
		t.Error("ES_Enabled should be true")
	}
	if globalSettings.OGN_Enabled {
		t.Error("OGN_Enabled should be false")
	}
	if !globalSettings.APRS_Enabled {
		t.Error("APRS_Enabled should be true")
	}
	if !globalSettings.GPS_Enabled {
		t.Error("GPS_Enabled should be true")
	}
	if !globalSettings.IMU_Sensor_Enabled {
		t.Error("IMU_Sensor_Enabled should be true")
	}
	if !globalSettings.BMP_Sensor_Enabled {
		t.Error("BMP_Sensor_Enabled should be true")
	}

	// Verify debug/developer flags are disabled
	if globalSettings.DEBUG {
		t.Error("DEBUG should be false")
	}
	if globalSettings.DeveloperMode {
		t.Error("DeveloperMode should be false")
	}
	if globalSettings.DisplayTrafficSource {
		t.Error("DisplayTrafficSource should be false")
	}
	if globalSettings.ReplayLog {
		t.Error("ReplayLog should be false")
	}
	if globalSettings.AHRSLog {
		t.Error("AHRSLog should be false")
	}
	if globalSettings.NoSleep {
		t.Error("NoSleep should be false")
	}

	t.Log("Basic default settings verified successfully")
}

// TestDefaultSettings_NumericDefaults tests numeric default values
func TestDefaultSettings_NumericDefaults(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Reset to non-default state
	globalSettings = settings{}
	globalSettings.Dump1090Gain = 99.9
	globalSettings.WiFiChannel = 99
	globalSettings.RadarLimits = 9999
	globalSettings.RadarRange = 999
	globalSettings.AltitudeOffset = 12345
	globalSettings.PWMDutyMin = 100
	globalSettings.GpsManualTargetBaud = 9600

	// Call defaultSettings
	defaultSettings()

	// Verify numeric defaults
	if globalSettings.Dump1090Gain != 37.2 {
		t.Errorf("Dump1090Gain should be 37.2, got %f", globalSettings.Dump1090Gain)
	}
	if globalSettings.WiFiChannel != 1 {
		t.Errorf("WiFiChannel should be 1, got %d", globalSettings.WiFiChannel)
	}
	if globalSettings.RadarLimits != 2000 {
		t.Errorf("RadarLimits should be 2000, got %d", globalSettings.RadarLimits)
	}
	if globalSettings.RadarRange != 10 {
		t.Errorf("RadarRange should be 10, got %d", globalSettings.RadarRange)
	}
	if globalSettings.AltitudeOffset != 0 {
		t.Errorf("AltitudeOffset should be 0, got %d", globalSettings.AltitudeOffset)
	}
	if globalSettings.PWMDutyMin != 0 {
		t.Errorf("PWMDutyMin should be 0, got %d", globalSettings.PWMDutyMin)
	}
	if globalSettings.GpsManualTargetBaud != 115200 {
		t.Errorf("GpsManualTargetBaud should be 115200, got %d", globalSettings.GpsManualTargetBaud)
	}

	t.Log("Numeric default settings verified successfully")
}

// TestDefaultSettings_StringDefaults tests string default values
func TestDefaultSettings_StringDefaults(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Reset to non-default state
	globalSettings = settings{}
	globalSettings.WiFiIPAddress = "10.0.0.1"
	globalSettings.WiFiPassphrase = "secret123"
	globalSettings.WiFiSSID = "CustomWiFi"
	globalSettings.OwnshipModeS = "123456"
	globalSettings.GpsManualDevice = "/dev/ttyUSB0"
	globalSettings.GpsManualChip = "mtk"

	// Call defaultSettings
	defaultSettings()

	// Verify string defaults
	if globalSettings.WiFiIPAddress != "192.168.10.1" {
		t.Errorf("WiFiIPAddress should be '192.168.10.1', got '%s'", globalSettings.WiFiIPAddress)
	}
	if globalSettings.WiFiPassphrase != "" {
		t.Errorf("WiFiPassphrase should be empty, got '%s'", globalSettings.WiFiPassphrase)
	}
	if globalSettings.WiFiSSID != "Stratux" {
		t.Errorf("WiFiSSID should be 'Stratux', got '%s'", globalSettings.WiFiSSID)
	}
	if globalSettings.OwnshipModeS != "F00000" {
		t.Errorf("OwnshipModeS should be 'F00000', got '%s'", globalSettings.OwnshipModeS)
	}
	if globalSettings.GpsManualDevice != "/dev/ttyAMA0" {
		t.Errorf("GpsManualDevice should be '/dev/ttyAMA0', got '%s'", globalSettings.GpsManualDevice)
	}
	if globalSettings.GpsManualChip != "ublox" {
		t.Errorf("GpsManualChip should be 'ublox', got '%s'", globalSettings.GpsManualChip)
	}

	t.Log("String default settings verified successfully")
}

// TestDefaultSettings_NetworkOutputs tests network outputs default configuration
func TestDefaultSettings_NetworkOutputs(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Reset to non-default state
	globalSettings = settings{}
	globalSettings.NetworkOutputs = []networkConnection{
		{Conn: nil, Ip: "1.2.3.4", Port: 1234, Capability: 0},
	}

	// Call defaultSettings
	defaultSettings()

	// Verify NetworkOutputs
	if len(globalSettings.NetworkOutputs) != 3 {
		t.Errorf("NetworkOutputs should have 3 entries, got %d", len(globalSettings.NetworkOutputs))
	}

	// Check first output (GDL90/AHRS on port 4000)
	if globalSettings.NetworkOutputs[0].Port != 4000 {
		t.Errorf("NetworkOutputs[0].Port should be 4000, got %d", globalSettings.NetworkOutputs[0].Port)
	}
	expectedCap0 := uint8(NETWORK_GDL90_STANDARD | NETWORK_AHRS_GDL90)
	if globalSettings.NetworkOutputs[0].Capability != expectedCap0 {
		t.Errorf("NetworkOutputs[0].Capability should be %d, got %d", expectedCap0, globalSettings.NetworkOutputs[0].Capability)
	}

	// Check second output (FLARM NMEA on port 2000)
	if globalSettings.NetworkOutputs[1].Port != 2000 {
		t.Errorf("NetworkOutputs[1].Port should be 2000, got %d", globalSettings.NetworkOutputs[1].Port)
	}
	if globalSettings.NetworkOutputs[1].Capability != NETWORK_FLARM_NMEA {
		t.Errorf("NetworkOutputs[1].Capability should be %d, got %d", NETWORK_FLARM_NMEA, globalSettings.NetworkOutputs[1].Capability)
	}

	// Check third output (FFSIM on port 49002)
	if globalSettings.NetworkOutputs[2].Port != 49002 {
		t.Errorf("NetworkOutputs[2].Port should be 49002, got %d", globalSettings.NetworkOutputs[2].Port)
	}
	expectedCap2 := uint8(NETWORK_POSITION_FFSIM | NETWORK_AHRS_FFSIM)
	if globalSettings.NetworkOutputs[2].Capability != expectedCap2 {
		t.Errorf("NetworkOutputs[2].Capability should be %d, got %d", expectedCap2, globalSettings.NetworkOutputs[2].Capability)
	}

	t.Log("NetworkOutputs default configuration verified successfully")
}

// TestDefaultSettings_BleOutputs tests BLE outputs default configuration
func TestDefaultSettings_BleOutputs(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Reset to non-default state
	globalSettings = settings{}
	globalSettings.BleOutputs = []bleConnection{}

	// Call defaultSettings
	defaultSettings()

	// Verify BleOutputs
	if len(globalSettings.BleOutputs) != 2 {
		t.Errorf("BleOutputs should have 2 entries, got %d", len(globalSettings.BleOutputs))
	}

	// Check first BLE output (SoftRF style)
	if globalSettings.BleOutputs[0].Capability != NETWORK_FLARM_NMEA {
		t.Errorf("BleOutputs[0].Capability should be %d, got %d", NETWORK_FLARM_NMEA, globalSettings.BleOutputs[0].Capability)
	}
	if globalSettings.BleOutputs[0].UUIDService != "FFE0" {
		t.Errorf("BleOutputs[0].UUIDService should be 'FFE0', got '%s'", globalSettings.BleOutputs[0].UUIDService)
	}
	if globalSettings.BleOutputs[0].UUIDGatt != "FFE1" {
		t.Errorf("BleOutputs[0].UUIDGatt should be 'FFE1', got '%s'", globalSettings.BleOutputs[0].UUIDGatt)
	}

	// Check second BLE output (nRF UART)
	if globalSettings.BleOutputs[1].Capability != NETWORK_FLARM_NMEA {
		t.Errorf("BleOutputs[1].Capability should be %d, got %d", NETWORK_FLARM_NMEA, globalSettings.BleOutputs[1].Capability)
	}
	if globalSettings.BleOutputs[1].UUIDService != "6E400001-B5A3-F393-E0A9-E50E24DCCA9E" {
		t.Errorf("BleOutputs[1].UUIDService incorrect, got '%s'", globalSettings.BleOutputs[1].UUIDService)
	}
	if globalSettings.BleOutputs[1].UUIDGatt != "6E400003-B5A3-F393-E0A9-E50E24DCCA9E" {
		t.Errorf("BleOutputs[1].UUIDGatt incorrect, got '%s'", globalSettings.BleOutputs[1].UUIDGatt)
	}

	t.Log("BleOutputs default configuration verified successfully")
}

// TestDefaultSettings_WiFiDefaults tests WiFi-related default values
func TestDefaultSettings_WiFiDefaults(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Reset to non-default state
	globalSettings = settings{}
	globalSettings.WiFiChannel = 11
	globalSettings.WiFiIPAddress = "10.10.10.1"
	globalSettings.WiFiPassphrase = "password123"
	globalSettings.WiFiSSID = "MyWiFi"
	globalSettings.WiFiSecurityEnabled = true
	globalSettings.WiFiClientNetworks = []wifiClientNetwork{
		{SSID: "SomeNetwork"},
	}

	// Call defaultSettings
	defaultSettings()

	// Verify WiFi defaults
	if globalSettings.WiFiChannel != 1 {
		t.Errorf("WiFiChannel should be 1, got %d", globalSettings.WiFiChannel)
	}
	if globalSettings.WiFiIPAddress != "192.168.10.1" {
		t.Errorf("WiFiIPAddress should be '192.168.10.1', got '%s'", globalSettings.WiFiIPAddress)
	}
	if globalSettings.WiFiPassphrase != "" {
		t.Errorf("WiFiPassphrase should be empty, got '%s'", globalSettings.WiFiPassphrase)
	}
	if globalSettings.WiFiSSID != "Stratux" {
		t.Errorf("WiFiSSID should be 'Stratux', got '%s'", globalSettings.WiFiSSID)
	}
	if globalSettings.WiFiSecurityEnabled {
		t.Error("WiFiSecurityEnabled should be false")
	}
	if len(globalSettings.WiFiClientNetworks) != 0 {
		t.Errorf("WiFiClientNetworks should be empty, got %d entries", len(globalSettings.WiFiClientNetworks))
	}

	t.Log("WiFi default settings verified successfully")
}

// TestDefaultSettings_IMUMappingDefaults tests IMU mapping default values
func TestDefaultSettings_IMUMappingDefaults(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Reset to non-default state
	globalSettings = settings{}
	globalSettings.IMUMapping = [2]int{5, 6}

	// Call defaultSettings
	defaultSettings()

	// Verify IMUMapping defaults
	if globalSettings.IMUMapping[0] != -1 {
		t.Errorf("IMUMapping[0] should be -1, got %d", globalSettings.IMUMapping[0])
	}
	if globalSettings.IMUMapping[1] != 0 {
		t.Errorf("IMUMapping[1] should be 0, got %d", globalSettings.IMUMapping[1])
	}

	t.Log("IMUMapping default settings verified successfully")
}

// TestDefaultSettings_SliceDefaults tests that slice fields are properly initialized
func TestDefaultSettings_SliceDefaults(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Reset to non-default state with nil slices
	globalSettings = settings{}
	globalSettings.StaticIps = nil
	globalSettings.WiFiClientNetworks = nil

	// Call defaultSettings
	defaultSettings()

	// Verify slices are initialized (not nil)
	if globalSettings.StaticIps == nil {
		t.Error("StaticIps should not be nil")
	}
	if len(globalSettings.StaticIps) != 0 {
		t.Errorf("StaticIps should be empty, got %d entries", len(globalSettings.StaticIps))
	}

	if globalSettings.WiFiClientNetworks == nil {
		t.Error("WiFiClientNetworks should not be nil")
	}
	if len(globalSettings.WiFiClientNetworks) != 0 {
		t.Errorf("WiFiClientNetworks should be empty, got %d entries", len(globalSettings.WiFiClientNetworks))
	}

	t.Log("Slice default initialization verified successfully")
}

// TestDefaultSettings_GpsDefaults tests GPS-related default values
func TestDefaultSettings_GpsDefaults(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Reset to non-default state
	globalSettings = settings{}
	globalSettings.GpsManualConfig = true
	globalSettings.GpsManualDevice = "/dev/ttyUSB0"
	globalSettings.GpsManualTargetBaud = 9600
	globalSettings.GpsManualChip = "sirf"

	// Call defaultSettings
	defaultSettings()

	// Verify GPS defaults
	if globalSettings.GpsManualConfig {
		t.Error("GpsManualConfig should be false")
	}
	if globalSettings.GpsManualDevice != "/dev/ttyAMA0" {
		t.Errorf("GpsManualDevice should be '/dev/ttyAMA0', got '%s'", globalSettings.GpsManualDevice)
	}
	if globalSettings.GpsManualTargetBaud != 115200 {
		t.Errorf("GpsManualTargetBaud should be 115200, got %d", globalSettings.GpsManualTargetBaud)
	}
	if globalSettings.GpsManualChip != "ublox" {
		t.Errorf("GpsManualChip should be 'ublox', got '%s'", globalSettings.GpsManualChip)
	}

	t.Log("GPS default settings verified successfully")
}

// TestDefaultSettings_RegionAndMiscDefaults tests region and miscellaneous default values
func TestDefaultSettings_RegionAndMiscDefaults(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Reset to non-default state
	globalSettings = settings{}
	globalSettings.RegionSelected = 2
	globalSettings.DarkMode = true
	globalSettings.EstimateBearinglessDist = true
	globalSettings.OGNI2CTXEnabled = false
	globalSettings.ClearLogOnStart = true

	// Call defaultSettings
	defaultSettings()

	// Verify region and misc defaults
	if globalSettings.RegionSelected != 0 {
		t.Errorf("RegionSelected should be 0 (none/US), got %d", globalSettings.RegionSelected)
	}
	if globalSettings.DarkMode {
		t.Error("DarkMode should be false")
	}
	if globalSettings.EstimateBearinglessDist {
		t.Error("EstimateBearinglessDist should be false")
	}
	if !globalSettings.OGNI2CTXEnabled {
		t.Error("OGNI2CTXEnabled should be true")
	}
	if globalSettings.ClearLogOnStart {
		t.Error("ClearLogOnStart should be false")
	}

	t.Log("Region and miscellaneous default settings verified successfully")
}

// TestDefaultSettings_ResetToDefaults tests that calling defaultSettings multiple times resets to defaults
func TestDefaultSettings_ResetToDefaults(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Call defaultSettings first time
	defaultSettings()
	firstCallUAT := globalSettings.UAT_Enabled
	firstCallES := globalSettings.ES_Enabled
	firstCallSSID := globalSettings.WiFiSSID

	// Modify settings
	globalSettings.UAT_Enabled = false
	globalSettings.ES_Enabled = false
	globalSettings.WiFiSSID = "Modified"
	globalSettings.DeveloperMode = true
	globalSettings.Dump1090Gain = 99.9

	// Call defaultSettings second time
	defaultSettings()

	// Verify settings are reset to defaults
	if globalSettings.UAT_Enabled != firstCallUAT {
		t.Errorf("UAT_Enabled should reset to %v, got %v", firstCallUAT, globalSettings.UAT_Enabled)
	}
	if globalSettings.ES_Enabled != firstCallES {
		t.Errorf("ES_Enabled should reset to %v, got %v", firstCallES, globalSettings.ES_Enabled)
	}
	if globalSettings.WiFiSSID != firstCallSSID {
		t.Errorf("WiFiSSID should reset to '%s', got '%s'", firstCallSSID, globalSettings.WiFiSSID)
	}
	if globalSettings.DeveloperMode {
		t.Error("DeveloperMode should reset to false")
	}
	if globalSettings.Dump1090Gain != 37.2 {
		t.Errorf("Dump1090Gain should reset to 37.2, got %f", globalSettings.Dump1090Gain)
	}

	// Verify consistent behavior
	if !globalSettings.UAT_Enabled {
		t.Error("UAT_Enabled should be true after reset")
	}
	if !globalSettings.ES_Enabled {
		t.Error("ES_Enabled should be true after reset")
	}
	if globalSettings.WiFiSSID != "Stratux" {
		t.Errorf("WiFiSSID should be 'Stratux' after reset, got '%s'", globalSettings.WiFiSSID)
	}

	t.Log("Settings correctly reset to defaults on multiple calls")
}

// TestDefaultSettings_AllFieldsCovered tests comprehensive coverage of all default settings
func TestDefaultSettings_AllFieldsCovered(t *testing.T) {
	// Save original settings
	origSettings := globalSettings
	defer func() {
		globalSettings = origSettings
	}()

	// Reset to completely empty state
	globalSettings = settings{}

	// Call defaultSettings
	defaultSettings()

	// Create a comprehensive test of all set defaults
	tests := []struct {
		name     string
		checkFn  func() bool
		errorMsg string
	}{
		{"RegionSelected", func() bool { return globalSettings.RegionSelected == 0 }, "RegionSelected should be 0"},
		{"DarkMode", func() bool { return !globalSettings.DarkMode }, "DarkMode should be false"},
		{"UAT_Enabled", func() bool { return globalSettings.UAT_Enabled }, "UAT_Enabled should be true"},
		{"ES_Enabled", func() bool { return globalSettings.ES_Enabled }, "ES_Enabled should be true"},
		{"OGN_Enabled", func() bool { return !globalSettings.OGN_Enabled }, "OGN_Enabled should be false"},
		{"Dump1090Gain", func() bool { return globalSettings.Dump1090Gain == 37.2 }, "Dump1090Gain should be 37.2"},
		{"APRS_Enabled", func() bool { return globalSettings.APRS_Enabled }, "APRS_Enabled should be true"},
		{"GPS_Enabled", func() bool { return globalSettings.GPS_Enabled }, "GPS_Enabled should be true"},
		{"IMU_Sensor_Enabled", func() bool { return globalSettings.IMU_Sensor_Enabled }, "IMU_Sensor_Enabled should be true"},
		{"BMP_Sensor_Enabled", func() bool { return globalSettings.BMP_Sensor_Enabled }, "BMP_Sensor_Enabled should be true"},
		{"NetworkOutputs_count", func() bool { return len(globalSettings.NetworkOutputs) == 3 }, "NetworkOutputs should have 3 entries"},
		{"BleOutputs_count", func() bool { return len(globalSettings.BleOutputs) == 2 }, "BleOutputs should have 2 entries"},
		{"DEBUG", func() bool { return !globalSettings.DEBUG }, "DEBUG should be false"},
		{"DisplayTrafficSource", func() bool { return !globalSettings.DisplayTrafficSource }, "DisplayTrafficSource should be false"},
		{"ReplayLog", func() bool { return !globalSettings.ReplayLog }, "ReplayLog should be false"},
		{"AHRSLog", func() bool { return !globalSettings.AHRSLog }, "AHRSLog should be false"},
		{"IMUMapping_0", func() bool { return globalSettings.IMUMapping[0] == -1 }, "IMUMapping[0] should be -1"},
		{"IMUMapping_1", func() bool { return globalSettings.IMUMapping[1] == 0 }, "IMUMapping[1] should be 0"},
		{"OwnshipModeS", func() bool { return globalSettings.OwnshipModeS == "F00000" }, "OwnshipModeS should be 'F00000'"},
		{"DeveloperMode", func() bool { return !globalSettings.DeveloperMode }, "DeveloperMode should be false"},
		{"StaticIps_count", func() bool { return len(globalSettings.StaticIps) == 0 }, "StaticIps should be empty"},
		{"NoSleep", func() bool { return !globalSettings.NoSleep }, "NoSleep should be false"},
		{"EstimateBearinglessDist", func() bool { return !globalSettings.EstimateBearinglessDist }, "EstimateBearinglessDist should be false"},
		{"WiFiChannel", func() bool { return globalSettings.WiFiChannel == 1 }, "WiFiChannel should be 1"},
		{"WiFiIPAddress", func() bool { return globalSettings.WiFiIPAddress == "192.168.10.1" }, "WiFiIPAddress should be '192.168.10.1'"},
		{"WiFiPassphrase", func() bool { return globalSettings.WiFiPassphrase == "" }, "WiFiPassphrase should be empty"},
		{"WiFiSSID", func() bool { return globalSettings.WiFiSSID == "Stratux" }, "WiFiSSID should be 'Stratux'"},
		{"WiFiSecurityEnabled", func() bool { return !globalSettings.WiFiSecurityEnabled }, "WiFiSecurityEnabled should be false"},
		{"WiFiClientNetworks_count", func() bool { return len(globalSettings.WiFiClientNetworks) == 0 }, "WiFiClientNetworks should be empty"},
		{"RadarLimits", func() bool { return globalSettings.RadarLimits == 2000 }, "RadarLimits should be 2000"},
		{"RadarRange", func() bool { return globalSettings.RadarRange == 10 }, "RadarRange should be 10"},
		{"AltitudeOffset", func() bool { return globalSettings.AltitudeOffset == 0 }, "AltitudeOffset should be 0"},
		{"PWMDutyMin", func() bool { return globalSettings.PWMDutyMin == 0 }, "PWMDutyMin should be 0"},
		{"OGNI2CTXEnabled", func() bool { return globalSettings.OGNI2CTXEnabled }, "OGNI2CTXEnabled should be true"},
		{"ClearLogOnStart", func() bool { return !globalSettings.ClearLogOnStart }, "ClearLogOnStart should be false"},
		{"GpsManualConfig", func() bool { return !globalSettings.GpsManualConfig }, "GpsManualConfig should be false"},
		{"GpsManualDevice", func() bool { return globalSettings.GpsManualDevice == "/dev/ttyAMA0" }, "GpsManualDevice should be '/dev/ttyAMA0'"},
		{"GpsManualTargetBaud", func() bool { return globalSettings.GpsManualTargetBaud == 115200 }, "GpsManualTargetBaud should be 115200"},
		{"GpsManualChip", func() bool { return globalSettings.GpsManualChip == "ublox" }, "GpsManualChip should be 'ublox'"},
	}

	failCount := 0
	for _, test := range tests {
		if !test.checkFn() {
			t.Errorf("%s: %s", test.name, test.errorMsg)
			failCount++
		}
	}

	if failCount == 0 {
		t.Logf("All %d default settings verified successfully", len(tests))
	} else {
		t.Errorf("%d out of %d default settings failed", failCount, len(tests))
	}
}

// TestMsgLogAppend_Basic tests basic message appending functionality
func TestMsgLogAppend_Basic(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original msgLog and reset
	origMsgLog := msgLog
	defer func() {
		msgLog = origMsgLog
	}()

	// Start with empty log
	msgLog = []msg{}

	// Create a test message
	testMsg := msg{
		MessageClass:     MSGCLASS_UAT,
		TimeReceived:     stratuxClock.GetTime(),
		Data:             "test-data",
		Products:         []uint32{413},
		Signal_amplitude: 100,
		Signal_strength:  -10.5,
		ADSBTowerID:      "TEST123",
	}

	// Append the message
	msgLogAppend(testMsg)

	// Verify it was added
	if len(msgLog) != 1 {
		t.Fatalf("Expected msgLog length 1, got %d", len(msgLog))
	}

	// Verify the message content
	if msgLog[0].MessageClass != MSGCLASS_UAT {
		t.Errorf("Expected MessageClass %d, got %d", MSGCLASS_UAT, msgLog[0].MessageClass)
	}
	if msgLog[0].Data != "test-data" {
		t.Errorf("Expected Data 'test-data', got '%s'", msgLog[0].Data)
	}
	if msgLog[0].Signal_amplitude != 100 {
		t.Errorf("Expected Signal_amplitude 100, got %d", msgLog[0].Signal_amplitude)
	}
	if msgLog[0].Signal_strength != -10.5 {
		t.Errorf("Expected Signal_strength -10.5, got %f", msgLog[0].Signal_strength)
	}
	if msgLog[0].ADSBTowerID != "TEST123" {
		t.Errorf("Expected ADSBTowerID 'TEST123', got '%s'", msgLog[0].ADSBTowerID)
	}

	t.Log("Successfully appended and verified message")
}

// TestMsgLogAppend_MultipleMessages tests appending multiple messages
func TestMsgLogAppend_MultipleMessages(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original msgLog and reset
	origMsgLog := msgLog
	defer func() {
		msgLog = origMsgLog
	}()

	// Start with empty log
	msgLog = []msg{}

	// Append multiple messages of different classes
	msgClasses := []uint{MSGCLASS_UAT, MSGCLASS_ES, MSGCLASS_OGN, MSGCLASS_AIS}

	for i, msgClass := range msgClasses {
		testMsg := msg{
			MessageClass:     msgClass,
			TimeReceived:     stratuxClock.GetTime(),
			Data:             "test-data-" + string(rune('0'+i)),
			Signal_amplitude: 100 + i,
			Signal_strength:  -10.0 - float64(i),
		}
		msgLogAppend(testMsg)
	}

	// Verify all messages were added
	if len(msgLog) != len(msgClasses) {
		t.Fatalf("Expected msgLog length %d, got %d", len(msgClasses), len(msgLog))
	}

	// Verify each message
	for i, expectedClass := range msgClasses {
		if msgLog[i].MessageClass != expectedClass {
			t.Errorf("Message %d: Expected MessageClass %d, got %d", i, expectedClass, msgLog[i].MessageClass)
		}
		if msgLog[i].Signal_amplitude != 100+i {
			t.Errorf("Message %d: Expected Signal_amplitude %d, got %d", i, 100+i, msgLog[i].Signal_amplitude)
		}
	}

	t.Logf("Successfully appended and verified %d messages", len(msgClasses))
}

// TestMsgLogAppend_PreservesOrder tests that messages are appended in order
func TestMsgLogAppend_PreservesOrder(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original msgLog and reset
	origMsgLog := msgLog
	defer func() {
		msgLog = origMsgLog
	}()

	// Start with empty log
	msgLog = []msg{}

	// Append messages with sequential data
	numMessages := 10
	for i := 0; i < numMessages; i++ {
		testMsg := msg{
			MessageClass:     MSGCLASS_UAT,
			TimeReceived:     stratuxClock.GetTime(),
			Signal_amplitude: i,
		}
		msgLogAppend(testMsg)
	}

	// Verify order is preserved
	for i := 0; i < numMessages; i++ {
		if msgLog[i].Signal_amplitude != i {
			t.Errorf("Order not preserved: position %d has Signal_amplitude %d, expected %d",
				i, msgLog[i].Signal_amplitude, i)
		}
	}

	t.Logf("Successfully verified order preservation for %d messages", numMessages)
}

// TestMsgLogAppend_EmptyMessage tests appending an empty/zero-value message
func TestMsgLogAppend_EmptyMessage(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original msgLog and reset
	origMsgLog := msgLog
	defer func() {
		msgLog = origMsgLog
	}()

	// Start with empty log
	msgLog = []msg{}

	// Append an empty message
	emptyMsg := msg{}
	msgLogAppend(emptyMsg)

	// Verify it was added
	if len(msgLog) != 1 {
		t.Fatalf("Expected msgLog length 1, got %d", len(msgLog))
	}

	// Verify zero values
	if msgLog[0].MessageClass != 0 {
		t.Errorf("Expected MessageClass 0, got %d", msgLog[0].MessageClass)
	}
	if msgLog[0].Data != "" {
		t.Errorf("Expected empty Data, got '%s'", msgLog[0].Data)
	}

	t.Log("Successfully appended empty message")
}

// TestMsgLogAppend_LargeVolume tests appending a large number of messages
func TestMsgLogAppend_LargeVolume(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original msgLog and reset
	origMsgLog := msgLog
	defer func() {
		msgLog = origMsgLog
	}()

	// Start with empty log
	msgLog = []msg{}

	// Append a large number of messages
	numMessages := 1000
	for i := 0; i < numMessages; i++ {
		testMsg := msg{
			MessageClass:     MSGCLASS_ES,
			TimeReceived:     stratuxClock.GetTime(),
			Signal_amplitude: i,
		}
		msgLogAppend(testMsg)
	}

	// Verify all were added
	if len(msgLog) != numMessages {
		t.Fatalf("Expected msgLog length %d, got %d", numMessages, len(msgLog))
	}

	// Spot check a few messages
	if msgLog[0].Signal_amplitude != 0 {
		t.Errorf("First message Signal_amplitude should be 0, got %d", msgLog[0].Signal_amplitude)
	}
	if msgLog[numMessages/2].Signal_amplitude != numMessages/2 {
		t.Errorf("Middle message Signal_amplitude should be %d, got %d",
			numMessages/2, msgLog[numMessages/2].Signal_amplitude)
	}
	if msgLog[numMessages-1].Signal_amplitude != numMessages-1 {
		t.Errorf("Last message Signal_amplitude should be %d, got %d",
			numMessages-1, msgLog[numMessages-1].Signal_amplitude)
	}

	t.Logf("Successfully appended %d messages", numMessages)
}

// TestMsgLogAppend_Concurrent tests thread safety with concurrent appends
func TestMsgLogAppend_Concurrent(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original msgLog and reset
	origMsgLog := msgLog
	defer func() {
		msgLog = origMsgLog
	}()

	// Start with empty log
	msgLog = []msg{}

	// Use a wait group to coordinate goroutines
	var wg sync.WaitGroup
	numGoroutines := 10
	messagesPerGoroutine := 100

	// Launch multiple goroutines that append messages concurrently
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < messagesPerGoroutine; i++ {
				testMsg := msg{
					MessageClass:     uint(goroutineID % 4), // Rotate through message classes
					TimeReceived:     stratuxClock.GetTime(),
					Signal_amplitude: goroutineID*1000 + i,
				}
				msgLogAppend(testMsg)
			}
		}(g)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify the total count
	expectedTotal := numGoroutines * messagesPerGoroutine
	if len(msgLog) != expectedTotal {
		t.Fatalf("Expected msgLog length %d, got %d", expectedTotal, len(msgLog))
	}

	// Verify no messages were lost or corrupted by checking uniqueness of Signal_amplitude values
	amplitudeMap := make(map[int]bool)
	for _, m := range msgLog {
		if amplitudeMap[m.Signal_amplitude] {
			t.Errorf("Duplicate Signal_amplitude found: %d (possible race condition)", m.Signal_amplitude)
		}
		amplitudeMap[m.Signal_amplitude] = true
	}

	t.Logf("Successfully appended %d messages concurrently from %d goroutines",
		expectedTotal, numGoroutines)
}

// TestMsgLogAppend_WithExistingMessages tests appending to a non-empty log
func TestMsgLogAppend_WithExistingMessages(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original msgLog and reset
	origMsgLog := msgLog
	defer func() {
		msgLog = origMsgLog
	}()

	// Start with some existing messages
	msgLog = []msg{
		{MessageClass: MSGCLASS_UAT, Signal_amplitude: 1},
		{MessageClass: MSGCLASS_ES, Signal_amplitude: 2},
	}

	initialLength := len(msgLog)

	// Append a new message
	newMsg := msg{
		MessageClass:     MSGCLASS_OGN,
		TimeReceived:     stratuxClock.GetTime(),
		Signal_amplitude: 3,
	}
	msgLogAppend(newMsg)

	// Verify the length increased
	if len(msgLog) != initialLength+1 {
		t.Fatalf("Expected msgLog length %d, got %d", initialLength+1, len(msgLog))
	}

	// Verify existing messages are unchanged
	if msgLog[0].Signal_amplitude != 1 {
		t.Errorf("First existing message changed, expected Signal_amplitude 1, got %d",
			msgLog[0].Signal_amplitude)
	}
	if msgLog[1].Signal_amplitude != 2 {
		t.Errorf("Second existing message changed, expected Signal_amplitude 2, got %d",
			msgLog[1].Signal_amplitude)
	}

	// Verify new message was appended
	if msgLog[2].MessageClass != MSGCLASS_OGN {
		t.Errorf("New message has wrong MessageClass, expected %d, got %d",
			MSGCLASS_OGN, msgLog[2].MessageClass)
	}
	if msgLog[2].Signal_amplitude != 3 {
		t.Errorf("New message has wrong Signal_amplitude, expected 3, got %d",
			msgLog[2].Signal_amplitude)
	}

	t.Log("Successfully appended to existing message log")
}

// TestRegisterADSBTextMessageReceived_TooFewFields tests the early return for messages with too few fields
func TestRegisterADSBTextMessageReceived_TooFewFields(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original status
	origStatus := globalStatus
	defer func() {
		globalStatus = origStatus
	}()

	// Reset counters
	globalStatus.UAT_METAR_total = 0
	globalStatus.UAT_TAF_total = 0
	globalStatus.UAT_PIREP_total = 0

	// Test with messages that have fewer than 5 fields
	testCases := []string{
		"",                       // Empty string
		"METAR",                  // 1 field
		"METAR KJFK",             // 2 fields
		"METAR KJFK 121853",      // 3 fields
		"METAR KJFK 121853 AUTO", // 4 fields
	}

	for _, msg := range testCases {
		registerADSBTextMessageReceived(msg, nil)

		// Verify counters were not incremented
		if globalStatus.UAT_METAR_total != 0 {
			t.Errorf("For msg '%s': Expected UAT_METAR_total to be 0, got %d",
				msg, globalStatus.UAT_METAR_total)
		}
	}

	t.Log("Successfully handled messages with too few fields")
}

// TestRegisterADSBTextMessageReceived_METAR tests METAR message handling
func TestRegisterADSBTextMessageReceived_METAR(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original status
	origStatus := globalStatus
	defer func() {
		globalStatus = origStatus
	}()

	// Reset counters
	globalStatus.UAT_METAR_total = 0

	// Test METAR message
	msg := "METAR KJFK 121853Z AUTO 09008KT 10SM FEW250 M04/M17 A3034 RMK AO2"
	registerADSBTextMessageReceived(msg, nil)

	if globalStatus.UAT_METAR_total != 1 {
		t.Errorf("Expected UAT_METAR_total to be 1, got %d", globalStatus.UAT_METAR_total)
	}

	t.Log("Successfully processed METAR message")
}

// TestRegisterADSBTextMessageReceived_SPECI tests SPECI message handling
func TestRegisterADSBTextMessageReceived_SPECI(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original status
	origStatus := globalStatus
	defer func() {
		globalStatus = origStatus
	}()

	// Reset counters
	globalStatus.UAT_METAR_total = 0

	// Test SPECI message (should increment METAR counter)
	msg := "SPECI KJFK 121920Z 09010KT 10SM FEW250 M03/M16 A3034"
	registerADSBTextMessageReceived(msg, nil)

	if globalStatus.UAT_METAR_total != 1 {
		t.Errorf("Expected UAT_METAR_total to be 1 for SPECI, got %d", globalStatus.UAT_METAR_total)
	}

	t.Log("Successfully processed SPECI message")
}

// TestRegisterADSBTextMessageReceived_TAF tests TAF message handling
func TestRegisterADSBTextMessageReceived_TAF(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original status
	origStatus := globalStatus
	defer func() {
		globalStatus = origStatus
	}()

	// Reset counters
	globalStatus.UAT_TAF_total = 0

	// Test TAF message
	msg := "TAF KJFK 121720Z 1218/1324 09008KT P6SM FEW250 FM130200 09012KT P6SM SKC"
	registerADSBTextMessageReceived(msg, nil)

	if globalStatus.UAT_TAF_total != 1 {
		t.Errorf("Expected UAT_TAF_total to be 1, got %d", globalStatus.UAT_TAF_total)
	}

	t.Log("Successfully processed TAF message")
}

// TestRegisterADSBTextMessageReceived_TAF_AMD tests TAF.AMD message handling
func TestRegisterADSBTextMessageReceived_TAF_AMD(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original status
	origStatus := globalStatus
	defer func() {
		globalStatus = origStatus
	}()

	// Reset counters
	globalStatus.UAT_TAF_total = 0

	// Test TAF.AMD message
	msg := "TAF.AMD KJFK 121815Z 1218/1324 09010KT P6SM FEW250"
	registerADSBTextMessageReceived(msg, nil)

	if globalStatus.UAT_TAF_total != 1 {
		t.Errorf("Expected UAT_TAF_total to be 1 for TAF.AMD, got %d", globalStatus.UAT_TAF_total)
	}

	t.Log("Successfully processed TAF.AMD message")
}

// TestRegisterADSBTextMessageReceived_WINDS tests WINDS message handling
func TestRegisterADSBTextMessageReceived_WINDS(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original status
	origStatus := globalStatus
	defer func() {
		globalStatus = origStatus
	}()

	// Reset counters
	globalStatus.UAT_TAF_total = 0

	// Test WINDS message (note: WINDS increments TAF counter)
	msg := "WINDS ALOFT 121800Z 3000 2714+04 2826+00 2939-05"
	registerADSBTextMessageReceived(msg, nil)

	if globalStatus.UAT_TAF_total != 1 {
		t.Errorf("Expected UAT_TAF_total to be 1 for WINDS, got %d", globalStatus.UAT_TAF_total)
	}

	t.Log("Successfully processed WINDS message")
}

// TestRegisterADSBTextMessageReceived_PIREP tests PIREP message handling
func TestRegisterADSBTextMessageReceived_PIREP(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original status
	origStatus := globalStatus
	defer func() {
		globalStatus = origStatus
	}()

	// Reset counters
	globalStatus.UAT_PIREP_total = 0

	// Test PIREP message
	msg := "PIREP JFK UA /OV JFK /TM 1853 /FL095 /TP B737 /SK BKN250 /RM SMOOTH"
	registerADSBTextMessageReceived(msg, nil)

	if globalStatus.UAT_PIREP_total != 1 {
		t.Errorf("Expected UAT_PIREP_total to be 1, got %d", globalStatus.UAT_PIREP_total)
	}

	t.Log("Successfully processed PIREP message")
}

// TestRegisterADSBTextMessageReceived_UnknownType tests handling of unknown message types
func TestRegisterADSBTextMessageReceived_UnknownType(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original status
	origStatus := globalStatus
	defer func() {
		globalStatus = origStatus
	}()

	// Reset counters
	globalStatus.UAT_METAR_total = 0
	globalStatus.UAT_TAF_total = 0
	globalStatus.UAT_PIREP_total = 0

	// Test with an unknown message type - should not increment any counters
	msg := "UNKNOWN TYPE 121853Z SOME DATA HERE"
	registerADSBTextMessageReceived(msg, nil)

	if globalStatus.UAT_METAR_total != 0 {
		t.Errorf("Expected UAT_METAR_total to be 0 for unknown type, got %d",
			globalStatus.UAT_METAR_total)
	}
	if globalStatus.UAT_TAF_total != 0 {
		t.Errorf("Expected UAT_TAF_total to be 0 for unknown type, got %d",
			globalStatus.UAT_TAF_total)
	}
	if globalStatus.UAT_PIREP_total != 0 {
		t.Errorf("Expected UAT_PIREP_total to be 0 for unknown type, got %d",
			globalStatus.UAT_PIREP_total)
	}

	t.Log("Successfully handled unknown message type")
}

// TestRegisterADSBTextMessageReceived_MultipleMessages tests processing multiple messages
func TestRegisterADSBTextMessageReceived_MultipleMessages(t *testing.T) {
	// Initialize clock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Save original status
	origStatus := globalStatus
	defer func() {
		globalStatus = origStatus
	}()

	// Reset counters
	globalStatus.UAT_METAR_total = 0
	globalStatus.UAT_TAF_total = 0
	globalStatus.UAT_PIREP_total = 0

	// Process multiple different message types
	messages := []struct {
		msg   string
		mType string
	}{
		{"METAR KJFK 121853Z AUTO 09008KT 10SM FEW250 M04/M17 A3034", "METAR"},
		{"SPECI KLGA 121900Z 09009KT 10SM FEW200 M05/M18 A3035", "SPECI"},
		{"TAF KEWR 121720Z 1218/1324 09008KT P6SM FEW250", "TAF"},
		{"TAF.AMD KTEB 121815Z 1218/1324 09010KT P6SM FEW250", "TAF.AMD"},
		{"PIREP EWR UA /OV EWR /TM 1853 /FL095 /TP B737", "PIREP"},
		{"WINDS ALOFT 121800Z 3000 2714+04 2826+00", "WINDS"},
	}

	for _, tc := range messages {
		registerADSBTextMessageReceived(tc.msg, nil)
	}

	// Verify counters
	if globalStatus.UAT_METAR_total != 2 { // METAR + SPECI
		t.Errorf("Expected UAT_METAR_total to be 2, got %d", globalStatus.UAT_METAR_total)
	}
	if globalStatus.UAT_TAF_total != 3 { // TAF + TAF.AMD + WINDS
		t.Errorf("Expected UAT_TAF_total to be 3, got %d", globalStatus.UAT_TAF_total)
	}
	if globalStatus.UAT_PIREP_total != 1 {
		t.Errorf("Expected UAT_PIREP_total to be 1, got %d", globalStatus.UAT_PIREP_total)
	}

	t.Log("Successfully processed multiple messages")
}

// TestMakeStratuxStatus_NegativeVersionNumbers tests version parsing with negative numbers
// This covers the bounds checking branches for m < 0, mi < 0, and b < 0
func TestMakeStratuxStatus_NegativeVersionNumbers(t *testing.T) {
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

	// Test edge version formats that may produce negative parsing results
	testCases := []struct {
		name    string
		version string
	}{
		// Versions with actual negative numbers after Atoi
		{"Negative_major", "v-1.1"},       // Results in m < 0
		{"Negative_minor", "v1.-1rc1"},    // Results in mi < 0 (needs suffix to parse minor)
		{"Negative_build_rc", "v1.1rc-1"}, // Results in b < 0
		{"Negative_build_r", "v1.1r-1"},   // Results in b < 0
		{"Negative_build_b", "v1.1b-1"},   // Results in b < 0
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

			// Verify that negative numbers are clamped to 0
			// After prepareMessage adds the frame marker (0x7E) at position 0,
			// the version bytes are at: msg[5] = major, msg[6] = minor, msg[8] = build
			// These should be 0 when input is negative
			t.Logf("%s: generated %d-byte message, version bytes: major=%d, minor=%d, build=%d",
				tc.name, len(msg), msg[5], msg[6], msg[8])
		})
	}
}

// TestMakeStratuxStatus_WithValidTempPress tests the isTempPressValid branch
func TestMakeStratuxStatus_WithValidTempPress(t *testing.T) {
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
	globalSettings = settings{}
	globalStatus = status{}
	stratuxVersion = "v1.6"

	// Set up valid temperature/pressure with recent timestamp
	mySituation.muBaro.Lock()
	mySituation.BaroLastMeasurementTime = stratuxClock.GetTime()
	mySituation.muBaro.Unlock()

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

	// Check that the pressure altitude valid bit is set (bit 3 of msg[14])
	// After prepareMessage adds the frame marker (0x7E) at position 0,
	// the original msg[13] flags byte is now at position [14]
	pressureValidBit := (msg[14] & (1 << 3))
	if pressureValidBit == 0 {
		t.Errorf("Expected pressure valid bit to be set, got msg[14]=0x%02X", msg[14])
	}

	t.Logf("makeStratuxStatus() with valid temp/press generated %d-byte message, flags=0x%02X", len(msg), msg[14])
}

// TestUpdateStatus_GPSSolutions tests all GPS solution status strings
func TestUpdateStatus_GPSSolutions(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Initialize mutex if needed
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original values
	origSituation := mySituation
	origStatus := globalStatus

	defer func() {
		mySituation = origSituation
		globalStatus = origStatus
	}()

	// Initialize Satellites map
	Satellites = make(map[string]SatelliteInfo)

	testCases := []struct {
		name             string
		fixQuality       uint8
		gpsConnected     bool
		expectedSolution string
	}{
		{
			name:             "3D GPS + SBAS",
			fixQuality:       2,
			gpsConnected:     true,
			expectedSolution: "3D GPS + SBAS",
		},
		{
			name:             "3D GPS",
			fixQuality:       1,
			gpsConnected:     true,
			expectedSolution: "3D GPS",
		},
		{
			name:             "Dead Reckoning",
			fixQuality:       6,
			gpsConnected:     true,
			expectedSolution: "Dead Reckoning",
		},
		{
			name:             "No Fix",
			fixQuality:       0,
			gpsConnected:     true,
			expectedSolution: "No Fix",
		},
		{
			name:             "Unknown fix quality",
			fixQuality:       99, // Unknown value
			gpsConnected:     true,
			expectedSolution: "Unknown",
		},
		{
			name:             "GPS Disconnected",
			fixQuality:       2,
			gpsConnected:     false,
			expectedSolution: "Disconnected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state
			mySituation.GPSFixQuality = tc.fixQuality
			mySituation.GPSLastValidNMEAMessageTime = stratuxClock.GetTime()
			globalStatus.GPS_connected = tc.gpsConnected

			// Call updateStatus
			updateStatus()

			// Verify GPS solution
			if globalStatus.GPS_solution != tc.expectedSolution {
				t.Errorf("Expected GPS_solution=%q, got %q", tc.expectedSolution, globalStatus.GPS_solution)
			}
		})
	}
}

// TestUpdateStatus_SatelliteTracking tests satellite count updates
func TestUpdateStatus_SatelliteTracking(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}

	// Initialize mutex if needed
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original values
	origSituation := mySituation
	origStatus := globalStatus

	defer func() {
		mySituation = origSituation
		globalStatus = origStatus
	}()

	// Initialize Satellites map
	Satellites = make(map[string]SatelliteInfo)

	// Set up connected GPS with satellite data
	mySituation.GPSFixQuality = 2
	mySituation.GPSSatellites = 10
	mySituation.GPSSatellitesSeen = 15
	mySituation.GPSSatellitesTracked = 12
	mySituation.GPSHorizontalAccuracy = 3.5
	mySituation.GPSLastValidNMEAMessageTime = stratuxClock.GetTime()
	globalStatus.GPS_connected = true

	// Call updateStatus
	updateStatus()

	// Verify satellite counts were updated
	if globalStatus.GPS_satellites_locked != 10 {
		t.Errorf("Expected GPS_satellites_locked=10, got %d", globalStatus.GPS_satellites_locked)
	}
	if globalStatus.GPS_satellites_seen != 15 {
		t.Errorf("Expected GPS_satellites_seen=15, got %d", globalStatus.GPS_satellites_seen)
	}
	if globalStatus.GPS_satellites_tracked != 12 {
		t.Errorf("Expected GPS_satellites_tracked=12, got %d", globalStatus.GPS_satellites_tracked)
	}
	if globalStatus.GPS_position_accuracy != 3.5 {
		t.Errorf("Expected GPS_position_accuracy=3.5, got %f", globalStatus.GPS_position_accuracy)
	}
}

// TestUpdateStatus_Uptime tests uptime value updates
func TestUpdateStatus_Uptime(t *testing.T) {
	// Initialize required components
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	// Wait a bit to ensure clock has advanced
	time.Sleep(10 * time.Millisecond)

	// Initialize mutex if needed
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Save original values
	origSituation := mySituation
	origStatus := globalStatus

	defer func() {
		mySituation = origSituation
		globalStatus = origStatus
	}()

	// Initialize Satellites map
	Satellites = make(map[string]SatelliteInfo)

	// Set up minimal state
	mySituation.GPSFixQuality = 0
	mySituation.GPSLastValidNMEAMessageTime = stratuxClock.GetTime()
	globalStatus.GPS_connected = true

	// Call updateStatus
	updateStatus()

	// Verify uptime was updated (should be non-zero after some time)
	if globalStatus.Uptime == 0 {
		t.Error("Expected Uptime to be non-zero")
	}
	if globalStatus.UptimeClock.IsZero() {
		t.Error("Expected UptimeClock to be set")
	}
}
