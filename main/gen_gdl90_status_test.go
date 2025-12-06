package main

import (
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
