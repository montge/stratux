package main

import (
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
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_UAT},
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
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_ES},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_ES},
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
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_OGN},
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
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_AIS},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_AIS},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_AIS},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_AIS},
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
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_ES},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_OGN},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_AIS},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_UAT},
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
		oldTime := stratuxClock.Time.Add(-2 * time.Minute)
		msgLogMutex.Lock()
		msgLog = []msg{
			{TimeReceived: oldTime, MessageClass: MSGCLASS_UAT},
			{TimeReceived: oldTime, MessageClass: MSGCLASS_ES},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_UAT}, // Recent
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
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_UAT},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_ES},
			{TimeReceived: stratuxClock.Time, MessageClass: MSGCLASS_ES},
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
				TimeReceived:     stratuxClock.Time,
				MessageClass:     MSGCLASS_UAT,
				ADSBTowerID:      "TOWER1",
				Signal_strength:  -20.0,
				Signal_amplitude: 100,
				uatMsg:           &uatparse.UATMsg{Lat: 40.0, Lon: -75.0},
			},
			{
				TimeReceived:     stratuxClock.Time,
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
	mySituation.GPSLastValidNMEAMessageTime = stratuxClock.Time

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
			name:           "Unknown fix quality",
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
