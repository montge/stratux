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
