package main

import (
	"math"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stratux/stratux/uatparse"
)

// TestUpdateUATStats tests the UAT product statistics tracking
func TestUpdateUATStats(t *testing.T) {
	tests := []struct {
		name       string
		productID  uint32
		checkField string
		checkFunc  func() uint32
	}{
		// METAR products
		{"METAR ID 0", 0, "UAT_METAR_total", func() uint32 { return globalStatus.UAT_METAR_total }},
		{"METAR ID 20", 20, "UAT_METAR_total", func() uint32 { return globalStatus.UAT_METAR_total }},

		// TAF products
		{"TAF ID 1", 1, "UAT_TAF_total", func() uint32 { return globalStatus.UAT_TAF_total }},
		{"TAF ID 21", 21, "UAT_TAF_total", func() uint32 { return globalStatus.UAT_TAF_total }},

		// NEXRAD products (comprehensive list)
		{"NEXRAD ID 51", 51, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 52", 52, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 53", 53, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 54", 54, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 55", 55, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 56", 56, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 57", 57, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 58", 58, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 59", 59, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 60", 60, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 61", 61, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 62", 62, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 63", 63, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 64", 64, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 81", 81, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 82", 82, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 83", 83, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},

		// SIGMET/AIRMET products
		{"SIGMET ID 2", 2, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 3", 3, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 4", 4, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 6", 6, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 11", 11, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 12", 12, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 22", 22, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 23", 23, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 24", 24, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 26", 26, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 254", 254, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},

		// PIREP products
		{"PIREP ID 5", 5, "UAT_PIREP_total", func() uint32 { return globalStatus.UAT_PIREP_total }},
		{"PIREP ID 25", 25, "UAT_PIREP_total", func() uint32 { return globalStatus.UAT_PIREP_total }},

		// NOTAM product
		{"NOTAM ID 8", 8, "UAT_NOTAM_total", func() uint32 { return globalStatus.UAT_NOTAM_total }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset counter
			before := tc.checkFunc()

			// Call function
			UpdateUATStats(tc.productID)

			// Verify counter increased
			after := tc.checkFunc()
			if after != before+1 {
				t.Errorf("Expected %s to increment from %d to %d, got %d",
					tc.checkField, before, before+1, after)
			}
		})
	}
}

// TestUpdateUATStatsSpecialCases tests special handling for product 413 and unknown products
func TestUpdateUATStatsSpecialCases(t *testing.T) {
	t.Run("Product 413 no-op", func(t *testing.T) {
		// Product 413 should not increment any counter (early return)
		beforeMETAR := globalStatus.UAT_METAR_total
		beforeTAF := globalStatus.UAT_TAF_total
		beforeNEXRAD := globalStatus.UAT_NEXRAD_total
		beforeSIGMET := globalStatus.UAT_SIGMET_total
		beforePIREP := globalStatus.UAT_PIREP_total
		beforeNOTAM := globalStatus.UAT_NOTAM_total
		beforeOTHER := globalStatus.UAT_OTHER_total

		UpdateUATStats(413)

		// Verify nothing changed
		if globalStatus.UAT_METAR_total != beforeMETAR ||
			globalStatus.UAT_TAF_total != beforeTAF ||
			globalStatus.UAT_NEXRAD_total != beforeNEXRAD ||
			globalStatus.UAT_SIGMET_total != beforeSIGMET ||
			globalStatus.UAT_PIREP_total != beforePIREP ||
			globalStatus.UAT_NOTAM_total != beforeNOTAM ||
			globalStatus.UAT_OTHER_total != beforeOTHER {
			t.Error("Product ID 413 should not increment any counters (early return)")
		}
	})

	t.Run("Unknown product defaults to OTHER", func(t *testing.T) {
		before := globalStatus.UAT_OTHER_total

		// Test several unknown product IDs
		unknownIDs := []uint32{9999, 10000, 100, 500, 12345}
		for _, id := range unknownIDs {
			UpdateUATStats(id)
		}

		after := globalStatus.UAT_OTHER_total
		expected := before + uint32(len(unknownIDs))

		if after != expected {
			t.Errorf("Expected UAT_OTHER_total to be %d, got %d", expected, after)
		}
	})
}

// TestUpdateMessageStats tests the message statistics collection function
func TestUpdateMessageStats(t *testing.T) {
	// Initialize required globals
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}
	// msgLogMutex is a value type (sync.Mutex), already initialized

	// Clear state before test
	ADSBTowerMutex.Lock()
	ADSBTowers = make(map[string]ADSBTower)
	ADSBTowerMutex.Unlock()

	msgLogMutex.Lock()
	msgLog = make([]msg, 0)
	msgLogMutex.Unlock()

	globalStatus.UAT_messages_last_minute = 0
	globalStatus.ES_messages_last_minute = 0
	globalStatus.OGN_messages_last_minute = 0
	globalStatus.AIS_messages_last_minute = 0
	globalStatus.UAT_messages_max = 0
	globalStatus.ES_messages_max = 0
	globalStatus.OGN_messages_max = 0
	globalStatus.AIS_messages_max = 0

	t.Run("empty_message_log", func(t *testing.T) {
		updateMessageStats()

		if globalStatus.UAT_messages_last_minute != 0 {
			t.Errorf("Expected UAT_messages_last_minute=0, got %d", globalStatus.UAT_messages_last_minute)
		}
		if globalStatus.ES_messages_last_minute != 0 {
			t.Errorf("Expected ES_messages_last_minute=0, got %d", globalStatus.ES_messages_last_minute)
		}
	})

	t.Run("uat_messages", func(t *testing.T) {
		// Add UAT messages
		msgLogMutex.Lock()
		msgLog = []msg{
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_UAT},
		}
		msgLogMutex.Unlock()

		updateMessageStats()

		if globalStatus.UAT_messages_last_minute != 3 {
			t.Errorf("Expected UAT_messages_last_minute=3, got %d", globalStatus.UAT_messages_last_minute)
		}
	})

	t.Run("es_messages", func(t *testing.T) {
		msgLogMutex.Lock()
		msgLog = []msg{
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_ES},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_ES},
		}
		msgLogMutex.Unlock()

		updateMessageStats()

		if globalStatus.ES_messages_last_minute != 2 {
			t.Errorf("Expected ES_messages_last_minute=2, got %d", globalStatus.ES_messages_last_minute)
		}
	})

	t.Run("ogn_messages", func(t *testing.T) {
		msgLogMutex.Lock()
		msgLog = []msg{
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_OGN},
		}
		msgLogMutex.Unlock()

		updateMessageStats()

		if globalStatus.OGN_messages_last_minute != 1 {
			t.Errorf("Expected OGN_messages_last_minute=1, got %d", globalStatus.OGN_messages_last_minute)
		}
	})

	t.Run("ais_messages", func(t *testing.T) {
		msgLogMutex.Lock()
		msgLog = []msg{
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_AIS},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_AIS},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_AIS},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_AIS},
		}
		msgLogMutex.Unlock()

		updateMessageStats()

		if globalStatus.AIS_messages_last_minute != 4 {
			t.Errorf("Expected AIS_messages_last_minute=4, got %d", globalStatus.AIS_messages_last_minute)
		}
	})

	t.Run("mixed_messages", func(t *testing.T) {
		msgLogMutex.Lock()
		msgLog = []msg{
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_ES},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_OGN},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_AIS},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_UAT},
		}
		msgLogMutex.Unlock()

		updateMessageStats()

		if globalStatus.UAT_messages_last_minute != 2 {
			t.Errorf("Expected UAT_messages_last_minute=2, got %d", globalStatus.UAT_messages_last_minute)
		}
		if globalStatus.ES_messages_last_minute != 1 {
			t.Errorf("Expected ES_messages_last_minute=1, got %d", globalStatus.ES_messages_last_minute)
		}
		if globalStatus.OGN_messages_last_minute != 1 {
			t.Errorf("Expected OGN_messages_last_minute=1, got %d", globalStatus.OGN_messages_last_minute)
		}
		if globalStatus.AIS_messages_last_minute != 1 {
			t.Errorf("Expected AIS_messages_last_minute=1, got %d", globalStatus.AIS_messages_last_minute)
		}
	})

	t.Run("old_messages_filtered", func(t *testing.T) {
		// Add old messages (> 1 minute old)
		oldTime := stratuxClock.GetTime().Add(-2 * time.Minute)
		msgLogMutex.Lock()
		msgLog = []msg{
			{TimeReceived: oldTime, MessageClass: MSGCLASS_UAT},
			{TimeReceived: oldTime, MessageClass: MSGCLASS_ES},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_UAT}, // Recent
		}
		msgLogMutex.Unlock()

		updateMessageStats()

		// Only recent message should count
		if globalStatus.UAT_messages_last_minute != 1 {
			t.Errorf("Expected UAT_messages_last_minute=1 (old filtered), got %d", globalStatus.UAT_messages_last_minute)
		}
		if globalStatus.ES_messages_last_minute != 0 {
			t.Errorf("Expected ES_messages_last_minute=0 (old filtered), got %d", globalStatus.ES_messages_last_minute)
		}
	})

	t.Run("max_counters_updated", func(t *testing.T) {
		globalStatus.UAT_messages_max = 0
		globalStatus.ES_messages_max = 0

		msgLogMutex.Lock()
		msgLog = []msg{
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_ES},
			{TimeReceived: stratuxClock.GetTime(), MessageClass: MSGCLASS_ES},
		}
		msgLogMutex.Unlock()

		updateMessageStats()

		if globalStatus.UAT_messages_max != 3 {
			t.Errorf("Expected UAT_messages_max=3, got %d", globalStatus.UAT_messages_max)
		}
		if globalStatus.ES_messages_max != 2 {
			t.Errorf("Expected ES_messages_max=2, got %d", globalStatus.ES_messages_max)
		}
	})

	t.Run("adsb_tower_tracking", func(t *testing.T) {
		ADSBTowerMutex.Lock()
		ADSBTowers = make(map[string]ADSBTower)
		ADSBTowerMutex.Unlock()

		msgLogMutex.Lock()
		msgLog = []msg{
			{
				TimeReceived:     stratuxClock.GetTime(),
				MessageClass:     MSGCLASS_UAT,
				ADSBTowerID:      "TOWER1",
				Signal_strength:  -20.0,
				Signal_amplitude: 100,
				uatMsg:           &uatparse.UATMsg{Lat: 40.0, Lon: -75.0},
			},
			{
				TimeReceived:     stratuxClock.GetTime(),
				MessageClass:     MSGCLASS_UAT,
				ADSBTowerID:      "TOWER1",
				Signal_strength:  -15.0,
				Signal_amplitude: 150,
				uatMsg:           &uatparse.UATMsg{Lat: 40.0, Lon: -75.0},
			},
		}
		msgLogMutex.Unlock()

		updateMessageStats()

		ADSBTowerMutex.Lock()
		tower, exists := ADSBTowers["TOWER1"]
		ADSBTowerMutex.Unlock()

		if !exists {
			t.Error("Expected tower TOWER1 to be tracked")
		} else {
			if tower.Messages_last_minute != 2 {
				t.Errorf("Expected tower Messages_last_minute=2, got %d", tower.Messages_last_minute)
			}
			if tower.Signal_strength_max != -15.0 {
				t.Errorf("Expected tower Signal_strength_max=-15.0, got %f", tower.Signal_strength_max)
			}
		}
	})

	t.Run("adsb_tower_signal_strength_calculation", func(t *testing.T) {
		ADSBTowerMutex.Lock()
		ADSBTowers = make(map[string]ADSBTower)
		ADSBTowerMutex.Unlock()

		// First call to updateMessageStats to create tower with messages
		msgLogMutex.Lock()
		msgLog = []msg{
			{
				TimeReceived:     stratuxClock.GetTime(),
				MessageClass:     MSGCLASS_UAT,
				ADSBTowerID:      "TOWER2",
				Signal_strength:  -25.0,
				Signal_amplitude: 200,
				uatMsg:           &uatparse.UATMsg{Lat: 41.0, Lon: -76.0},
			},
			{
				TimeReceived:     stratuxClock.GetTime(),
				MessageClass:     MSGCLASS_UAT,
				ADSBTowerID:      "TOWER2",
				Signal_strength:  -30.0,
				Signal_amplitude: 300,
				uatMsg:           &uatparse.UATMsg{Lat: 41.0, Lon: -76.0},
			},
		}
		msgLogMutex.Unlock()

		updateMessageStats()

		// Verify tower was created with stats
		ADSBTowerMutex.Lock()
		tower, exists := ADSBTowers["TOWER2"]
		ADSBTowerMutex.Unlock()

		if !exists {
			t.Fatal("Expected tower TOWER2 to be created")
		}

		// Verify the signal strength calculation (else branch at line 964)
		// Signal_strength_last_minute should be calculated as:
		// 10 * (math.Log10(float64((Energy_last_minute / Messages_last_minute))) - 6)
		expectedEnergy := uint64(200*200 + 300*300) // 40000 + 90000 = 130000
		expectedMessages := uint64(2)
		expectedAvgPower := float64(expectedEnergy / expectedMessages) // 65000
		expectedSignalStrength := 10 * (math.Log10(expectedAvgPower) - 6)

		if tower.Messages_last_minute != 2 {
			t.Errorf("Expected tower Messages_last_minute=2, got %d", tower.Messages_last_minute)
		}
		if tower.Energy_last_minute != expectedEnergy {
			t.Errorf("Expected tower Energy_last_minute=%d, got %d", expectedEnergy, tower.Energy_last_minute)
		}
		if tower.Signal_strength_last_minute != expectedSignalStrength {
			t.Errorf("Expected tower Signal_strength_last_minute=%f, got %f", expectedSignalStrength, tower.Signal_strength_last_minute)
		}
	})

	t.Run("adsb_tower_zero_messages_or_energy", func(t *testing.T) {
		// Test the case where Messages_last_minute == 0 or Energy_last_minute == 0
		// This should trigger the -999 signal strength (line 962)
		ADSBTowerMutex.Lock()
		ADSBTowers = map[string]ADSBTower{
			"TOWER_ZERO": {
				Lat:                         42.0,
				Lng:                         -77.0,
				Messages_last_minute:        5,
				Energy_last_minute:          50000,
				Signal_strength_last_minute: -20.0,
			},
		}
		ADSBTowerMutex.Unlock()

		// Clear message log so tower gets zero messages this minute
		msgLogMutex.Lock()
		msgLog = []msg{}
		msgLogMutex.Unlock()

		updateMessageStats()

		ADSBTowerMutex.Lock()
		tower, exists := ADSBTowers["TOWER_ZERO"]
		ADSBTowerMutex.Unlock()

		if !exists {
			t.Fatal("Expected tower TOWER_ZERO to still exist")
		}

		// After updateMessageStats with no messages, the tower should have zero stats
		if tower.Messages_last_minute != 0 {
			t.Errorf("Expected tower Messages_last_minute=0, got %d", tower.Messages_last_minute)
		}
		if tower.Energy_last_minute != 0 {
			t.Errorf("Expected tower Energy_last_minute=0, got %d", tower.Energy_last_minute)
		}
		if tower.Signal_strength_last_minute != -999 {
			t.Errorf("Expected tower Signal_strength_last_minute=-999 (no data), got %f", tower.Signal_strength_last_minute)
		}
	})
}

// TestUpdateStatus tests the GPS status string generation
func TestUpdateStatus(t *testing.T) {
	// Initialize required globals
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}
	if Satellites == nil {
		Satellites = make(map[string]SatelliteInfo)
	}

	// Set GPS as connected for these tests
	globalStatus.GPS_connected = true
	// Also set recent NMEA message time so isGPSConnected() returns true
	mySituation.GPSLastValidNMEAMessageTime = stratuxClock.GetTime()

	testCases := []struct {
		name           string
		gpsFixQuality  uint8
		expectedStatus string
	}{
		{
			name:           "3D GPS + SBAS",
			gpsFixQuality:  2,
			expectedStatus: "3D GPS + SBAS",
		},
		{
			name:           "3D GPS",
			gpsFixQuality:  1,
			expectedStatus: "3D GPS",
		},
		{
			name:           "Dead Reckoning",
			gpsFixQuality:  6,
			expectedStatus: "Dead Reckoning",
		},
		{
			name:           "No Fix",
			gpsFixQuality:  0,
			expectedStatus: "No Fix",
		},
		{
			name:           "Unknown fix quality 3",
			gpsFixQuality:  3,
			expectedStatus: "Unknown",
		},
		{
			name:           "Unknown fix quality 4",
			gpsFixQuality:  4,
			expectedStatus: "Unknown",
		},
		{
			name:           "Unknown fix quality 5",
			gpsFixQuality:  5,
			expectedStatus: "Unknown",
		},
		{
			name:           "Unknown fix quality 7",
			gpsFixQuality:  7,
			expectedStatus: "Unknown",
		},
		{
			name:           "Unknown fix quality 8",
			gpsFixQuality:  8,
			expectedStatus: "Unknown",
		},
		{
			name:           "Unknown fix quality 9",
			gpsFixQuality:  9,
			expectedStatus: "Unknown",
		},
		{
			name:           "Unknown fix quality 99",
			gpsFixQuality:  99,
			expectedStatus: "Unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mySituation.GPSFixQuality = tc.gpsFixQuality
			updateStatus()

			if globalStatus.GPS_solution != tc.expectedStatus {
				t.Errorf("Expected GPS_solution=%q, got %q", tc.expectedStatus, globalStatus.GPS_solution)
			}
		})
	}
}

// TestUpdateStatus_Disconnected tests the GPS disconnected state
func TestUpdateStatus_Disconnected(t *testing.T) {
	// Initialize required globals
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}
	if Satellites == nil {
		Satellites = make(map[string]SatelliteInfo)
	}

	// Test disconnected GPS
	globalStatus.GPS_connected = false
	mySituation.GPSFixQuality = 2 // Even with good fix quality

	updateStatus()

	if globalStatus.GPS_solution != "Disconnected" {
		t.Errorf("Expected GPS_solution='Disconnected' when GPS_connected=false, got %q", globalStatus.GPS_solution)
	}
}

// TestUpdateStatus_AHRSLogFiles tests AHRS log file scanning with sensor files in /var/log
// This test covers lines 1014-1017 in gen_gdl90.go (the AHRS log file scanning code).
// These lines are only executed when actual sensor files matching "sensors_*.csv" exist in /var/log.
//
// Coverage breakdown for updateStatus function (lines 971-1022):
// - Lines 972-982: GPS fix quality status strings (covered by TestUpdateStatus)
// - Lines 984-996: GPS disconnection detection (covered by TestUpdateStatus_Disconnected and TestUpdateStatus_CompleteCoverage)
// - Lines 998-1009: GPS status field updates, uptime, disk usage (covered by TestUpdateStatus_CompleteCoverage)
// - Lines 1011-1013: AHRS log scanning initialization (covered by all tests)
// - Lines 1014-1017: AHRS log file size calculation (covered ONLY when sensor files exist - requires sudo)
// - Lines 1020-1021: Final AHRS size assignment (covered by all tests)
//
// Without sudo/root: 91.7% coverage (missing only lines 1014-1017)
// With sudo/root: ~100% coverage (all lines covered)
//
// To achieve 100% coverage, run:
//
//	sudo bash run_coverage_test.sh
func TestUpdateStatus_AHRSLogFiles(t *testing.T) {
	// Initialize required globals
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}
	if Satellites == nil {
		Satellites = make(map[string]SatelliteInfo)
	}

	// Set GPS as connected
	globalStatus.GPS_connected = true
	mySituation.GPSLastValidNMEAMessageTime = stratuxClock.GetTime()
	mySituation.GPSFixQuality = 1

	// Create test sensor log files in /var/log
	testFiles := []string{
		"/var/log/sensors_test_coverage1.csv",
		"/var/log/sensors_test_coverage2.csv",
	}

	filesCreatedByUs := false

	// Try to create test files - skip test if we don't have permission
	// But don't skip if files already exist (may have been created by setup script)
	for _, fname := range testFiles {
		content := []byte("timestamp,accel_x,accel_y,accel_z\n1234567890,0.1,0.2,0.3\n")
		if err := os.WriteFile(fname, content, 0644); err != nil {
			// Check if files exist anyway (created by external setup script)
			if _, statErr := os.Stat(fname); statErr == nil {
				t.Logf("Using pre-existing test file: %s", fname)
				continue
			}
			// Can't write and files don't exist - skip this test
			t.Skipf("Skipping test - cannot write to /var/log: %v (Hint: run 'sudo bash run_coverage_test.sh' for 100%% coverage)", err)
			return
		}
		filesCreatedByUs = true
	}

	// Clean up test files when done (only if we created them)
	if filesCreatedByUs {
		defer func() {
			for _, fname := range testFiles {
				os.Remove(fname)
			}
		}()
	}

	// Record initial size
	sizeBefore := globalStatus.AHRS_LogFiles_Size

	// Call updateStatus - it will scan /var/log for sensors_*.csv files
	updateStatus()

	// Verify the size increased (our test files should be counted)
	// The size should include our test files
	expectedMinSize := int64(len("timestamp,accel_x,accel_y,accel_z\n1234567890,0.1,0.2,0.3\n") * 2)
	if globalStatus.AHRS_LogFiles_Size < expectedMinSize {
		t.Errorf("Expected AHRS_LogFiles_Size to be at least %d (our test files), got %d",
			expectedMinSize, globalStatus.AHRS_LogFiles_Size)
	}

	t.Logf("AHRS_LogFiles_Size: before=%d, after=%d (expected at least %d from test files)",
		sizeBefore, globalStatus.AHRS_LogFiles_Size, expectedMinSize)
}

// TestUpdateStatus_CompleteCoverage tests all remaining code paths in updateStatus
func TestUpdateStatus_CompleteCoverage(t *testing.T) {
	// Initialize required globals
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
	if mySituation.muSatellite == nil {
		mySituation.muSatellite = &sync.Mutex{}
	}
	if Satellites == nil {
		Satellites = make(map[string]SatelliteInfo)
	}

	t.Run("AHRS log files scanning without root", func(t *testing.T) {
		// Set GPS as connected
		globalStatus.GPS_connected = true
		mySituation.GPSLastValidNMEAMessageTime = stratuxClock.GetTime()
		mySituation.GPSFixQuality = 1

		// This tests the AHRS log file scanning code (lines 1011-1021)
		// Even without sensor files, the code should execute without error
		// and set AHRS_LogFiles_Size to 0 or the sum of any existing files
		sizeBefore := globalStatus.AHRS_LogFiles_Size

		updateStatus()

		// AHRS_LogFiles_Size should be set (even if to 0)
		// The important thing is that the code path executes without panic
		if globalStatus.AHRS_LogFiles_Size < 0 {
			t.Errorf("AHRS_LogFiles_Size should not be negative, got %d", globalStatus.AHRS_LogFiles_Size)
		}

		// The value might be 0 or might include existing sensor files
		t.Logf("AHRS_LogFiles_Size before=%d, after=%d", sizeBefore, globalStatus.AHRS_LogFiles_Size)
	})

	t.Run("isGPSConnected false path", func(t *testing.T) {
		// Set GPS_connected to true but make isGPSConnected() return false
		// by setting an old NMEA message time
		globalStatus.GPS_connected = true
		mySituation.GPSLastValidNMEAMessageTime = stratuxClock.GetTime().Add(-10 * time.Minute)
		mySituation.GPSFixQuality = 2
		mySituation.GPSSatellites = 5
		mySituation.GPSSatellitesSeen = 10
		mySituation.GPSSatellitesTracked = 8

		updateStatus()

		// Should detect disconnection via isGPSConnected()
		if globalStatus.GPS_solution != "Disconnected" {
			t.Errorf("Expected GPS_solution='Disconnected' when isGPSConnected() returns false, got %q", globalStatus.GPS_solution)
		}
		if globalStatus.GPS_connected != false {
			t.Errorf("Expected GPS_connected=false after disconnect detection")
		}
		if mySituation.GPSSatellites != 0 {
			t.Errorf("Expected GPSSatellites reset to 0, got %d", mySituation.GPSSatellites)
		}
	})

	t.Run("GPS_connected false initially", func(t *testing.T) {
		// Test the first condition in line 984: !(globalStatus.GPS_connected)
		globalStatus.GPS_connected = false
		mySituation.GPSLastValidNMEAMessageTime = stratuxClock.GetTime() // Valid time but GPS_connected is false
		mySituation.GPSFixQuality = 2
		mySituation.GPSSatellites = 5

		updateStatus()

		// Should be disconnected due to GPS_connected being false
		if globalStatus.GPS_solution != "Disconnected" {
			t.Errorf("Expected GPS_solution='Disconnected' when GPS_connected=false, got %q", globalStatus.GPS_solution)
		}
		if mySituation.GPSSatellites != 0 {
			t.Errorf("Expected GPSSatellites reset to 0, got %d", mySituation.GPSSatellites)
		}
	})

	t.Run("All GPS status fields updated", func(t *testing.T) {
		// Test that all GPS status fields are properly copied from mySituation
		globalStatus.GPS_connected = true
		mySituation.GPSLastValidNMEAMessageTime = stratuxClock.GetTime()
		mySituation.GPSFixQuality = 1
		mySituation.GPSSatellites = 7
		mySituation.GPSSatellitesSeen = 12
		mySituation.GPSSatellitesTracked = 9
		mySituation.GPSHorizontalAccuracy = 2.5

		updateStatus()

		// Verify all fields are updated (lines 998-1001)
		if globalStatus.GPS_satellites_locked != 7 {
			t.Errorf("Expected GPS_satellites_locked=7, got %d", globalStatus.GPS_satellites_locked)
		}
		if globalStatus.GPS_satellites_seen != 12 {
			t.Errorf("Expected GPS_satellites_seen=12, got %d", globalStatus.GPS_satellites_seen)
		}
		if globalStatus.GPS_satellites_tracked != 9 {
			t.Errorf("Expected GPS_satellites_tracked=9, got %d", globalStatus.GPS_satellites_tracked)
		}
		if globalStatus.GPS_position_accuracy != 2.5 {
			t.Errorf("Expected GPS_position_accuracy=2.5, got %f", globalStatus.GPS_position_accuracy)
		}
	})

	t.Run("Uptime and disk usage updated", func(t *testing.T) {
		// Test that uptime and disk usage fields are updated (lines 1004-1009)
		globalStatus.GPS_connected = true
		mySituation.GPSLastValidNMEAMessageTime = stratuxClock.GetTime()
		mySituation.GPSFixQuality = 1

		beforeUptime := globalStatus.Uptime

		updateStatus()

		// Uptime should be set to current clock value (line 1004)
		if globalStatus.Uptime != int64(stratuxClock.GetMilliseconds()) {
			t.Errorf("Expected Uptime to match stratuxClock.GetMilliseconds()")
		}

		// UptimeClock should be set (line 1005)
		if globalStatus.UptimeClock != stratuxClock.GetTime() {
			t.Errorf("Expected UptimeClock to match stratuxClock.GetTime()")
		}

		// DiskBytesFree should be set (line 1008)
		if globalStatus.DiskBytesFree == 0 {
			t.Logf("Warning: DiskBytesFree is 0, this might be a test environment issue")
		}

		// Logfile_Size should be set (line 1009)
		// Note: logFileSize() returns 0 if logFileHandle is nil
		if globalStatus.Logfile_Size < 0 {
			t.Errorf("Logfile_Size should not be negative, got %d", globalStatus.Logfile_Size)
		}

		t.Logf("Uptime before=%d, after=%d, DiskBytesFree=%d, Logfile_Size=%d",
			beforeUptime, globalStatus.Uptime, globalStatus.DiskBytesFree, globalStatus.Logfile_Size)
	})
}
