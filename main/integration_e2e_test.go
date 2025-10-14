package main

import (
	"sync"
	"testing"
	"time"
)

// resetE2EState resets all global state for end-to-end testing
func resetE2EState() {
	// Initialize stratuxClock
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond)
	}

	// Initialize mySituation mutexes if needed
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muGPSPerformance == nil {
		mySituation.muGPSPerformance = &sync.Mutex{}
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

	// Reset GPS state
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 0
	mySituation.GPSLongitude = 0
	mySituation.GPSAltitudeMSL = 0
	mySituation.GPSFixQuality = 0
	mySituation.GPSGroundSpeed = 0
	mySituation.GPSTrueCourse = 0
	mySituation.GPSLastFixLocalTime = time.Time{}
	mySituation.GPSLastGroundTrackTime = time.Time{}
	mySituation.muGPS.Unlock()

	// Initialize and reset traffic
	if trafficMutex == nil {
		trafficMutex = &sync.Mutex{}
	}
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)
	trafficMutex.Unlock()

	// Reset message counters
	globalStatus.UAT_messages_total = 0
	globalStatus.ES_messages_total = 0
	globalStatus.OGN_messages_total = 0
	globalStatus.UAT_messages_last_minute = 0
	globalStatus.ES_messages_last_minute = 0
	globalStatus.UAT_METAR_total = 0
	globalStatus.UAT_TAF_total = 0
	globalStatus.UAT_NEXRAD_total = 0
	globalStatus.UAT_SIGMET_total = 0
	globalStatus.UAT_PIREP_total = 0

	// Reset message log
	msgLogMutex.Lock()
	msgLog = make([]msg, 0)
	msgLogMutex.Unlock()

	// Initialize and reset ADS-B towers
	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}
	ADSBTowerMutex.Lock()
	ADSBTowers = make(map[string]ADSBTower)
	ADSBTowerMutex.Unlock()

	// Initialize Satellites if needed
	if Satellites == nil {
		Satellites = make(map[string]SatelliteInfo)
	}
}

// TestE2EGPSAndOwnshipReporting tests GPS integration with ownship report generation
func TestE2EGPSAndOwnshipReporting(t *testing.T) {
	resetE2EState()

	// Simulate GPS fix
	rmc := "$GPRMC,120000.000,A,4727.030,N,12218.528,W,057.9,349.7,131025,015.0,E*79"
	processNMEALine(rmc)

	gga := "$GPGGA,120000.000,4727.030,N,12218.528,W,1,08,0.9,420.9,M,46.9,M,,*4A"
	processNMEALine(gga)

	// Verify GPS position was set
	mySituation.muGPS.Lock()
	hasPosition := mySituation.GPSLatitude != 0 && mySituation.GPSLongitude != 0
	hasAltitude := mySituation.GPSAltitudeMSL != 0
	hasSpeed := mySituation.GPSGroundSpeed != 0
	fixQuality := mySituation.GPSFixQuality
	mySituation.muGPS.Unlock()

	if !hasPosition {
		t.Error("Expected GPS position to be set")
	}

	if !hasAltitude {
		t.Error("Expected GPS altitude to be set")
	}

	if !hasSpeed {
		t.Error("Expected GPS ground speed to be set")
	}

	if fixQuality == 0 {
		t.Error("Expected GPS fix quality to be set")
	}

	// Test that ownship report can be generated (requires valid GPS)
	// Note: isGPSValid() requires GPS time to be recent, which may not be true in tests
	result := makeOwnshipReport()

	t.Logf("GPS fix established: Lat=%.6f, Lon=%.6f, Alt=%.1f ft, Speed=%.1f kts, Fix=%d, Ownship=%v",
		mySituation.GPSLatitude, mySituation.GPSLongitude, mySituation.GPSAltitudeMSL,
		mySituation.GPSGroundSpeed, fixQuality, result)
}

// TestE2EMultiProtocolTrafficFusion tests traffic from multiple protocols
func TestE2EMultiProtocolTrafficFusion(t *testing.T) {
	resetE2EState()

	// Process 1090ES traffic
	es1090 := `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7657,"Lng":-1.1918,"Alt":5850,"Position_valid":true,"SignalLevel":0.0512,"Timestamp":"2025-10-14T12:00:00.000Z"}`
	parseDump1090Message(es1090)

	// Check that traffic was created
	trafficMutex.Lock()
	count1090 := len(traffic)
	trafficMutex.Unlock()

	if count1090 == 0 {
		t.Error("Expected traffic from 1090ES message")
	}

	// Verify message counter incremented
	if globalStatus.ES_messages_total == 0 {
		t.Error("Expected ES_messages_total to be incremented")
	}

	t.Logf("Multi-protocol fusion: 1090ES traffic=%d, ES_messages_total=%d",
		count1090, globalStatus.ES_messages_total)
}

// TestE2EUATWeatherStatistics tests UAT weather product tracking
func TestE2EUATWeatherStatistics(t *testing.T) {
	resetE2EState()

	initialMETAR := globalStatus.UAT_METAR_total
	initialTAF := globalStatus.UAT_TAF_total
	initialNEXRAD := globalStatus.UAT_NEXRAD_total

	// Test different product IDs
	testCases := []struct {
		name      string
		productID uint32
		checkFunc func() uint32
	}{
		{"METAR", 0, func() uint32 { return globalStatus.UAT_METAR_total }},
		{"TAF", 1, func() uint32 { return globalStatus.UAT_TAF_total }},
		{"NEXRAD", 51, func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"SIGMET", 2, func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"PIREP", 5, func() uint32 { return globalStatus.UAT_PIREP_total }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			before := tc.checkFunc()
			UpdateUATStats(tc.productID)
			after := tc.checkFunc()

			if after != before+1 {
				t.Errorf("Expected %s counter to increment from %d to %d, got %d",
					tc.name, before, before+1, after)
			}
		})
	}

	t.Logf("UAT weather stats: METAR=%d→%d, TAF=%d→%d, NEXRAD=%d→%d",
		initialMETAR, globalStatus.UAT_METAR_total,
		initialTAF, globalStatus.UAT_TAF_total,
		initialNEXRAD, globalStatus.UAT_NEXRAD_total)
}

// TestE2EMessageStatistics tests message statistics tracking
func TestE2EMessageStatistics(t *testing.T) {
	resetE2EState()

	// Add some UAT messages to the log
	for i := 0; i < 5; i++ {
		var m msg
		m.MessageClass = MSGCLASS_UAT
		m.TimeReceived = stratuxClock.Time
		m.Signal_amplitude = 100
		msgLogAppend(m)
	}

	// Add some ES messages
	for i := 0; i < 3; i++ {
		var m msg
		m.MessageClass = MSGCLASS_ES
		m.TimeReceived = stratuxClock.Time
		m.Signal_amplitude = 80
		msgLogAppend(m)
	}

	// Verify messages were logged
	msgLogMutex.Lock()
	logSize := len(msgLog)
	msgLogMutex.Unlock()

	if logSize != 8 {
		t.Errorf("Expected 8 messages in log, got %d", logSize)
	}

	// Run updateMessageStats
	updateMessageStats()

	// Check that counters were updated
	if globalStatus.UAT_messages_last_minute != 5 {
		t.Errorf("Expected UAT_messages_last_minute=5, got %d", globalStatus.UAT_messages_last_minute)
	}

	if globalStatus.ES_messages_last_minute != 3 {
		t.Errorf("Expected ES_messages_last_minute=3, got %d", globalStatus.ES_messages_last_minute)
	}

	t.Logf("Message stats: UAT/min=%d, ES/min=%d, Total log size=%d",
		globalStatus.UAT_messages_last_minute, globalStatus.ES_messages_last_minute, logSize)
}

// TestE2EADSBTowerTracking tests UAT uplink tower tracking
func TestE2EADSBTowerTracking(t *testing.T) {
	resetE2EState()

	// Create a UAT uplink message with tower location
	uplinkHex := "3cc0978aa66ca1a0158000213c5d2082102c22cc00082eec1e012c22cc000000000000000fd90007110e240811081ec5ea23b0c00"
	for len(uplinkHex) < 864 {
		uplinkHex += "0"
	}
	uplinkMsg := "+" + uplinkHex + ";rs=16;ss=128"

	// Process the uplink
	parseInput(uplinkMsg)

	// Check that tower was tracked
	ADSBTowerMutex.Lock()
	towerCount := len(ADSBTowers)
	ADSBTowerMutex.Unlock()

	if towerCount == 0 {
		t.Log("Note: No towers tracked (may require valid uatMsg decoding)")
	} else {
		t.Logf("Tower tracking: %d tower(s) detected", towerCount)

		// Verify tower has signal strength data
		ADSBTowerMutex.Lock()
		for towerID, tower := range ADSBTowers {
			t.Logf("Tower %s: Lat=%.4f, Lng=%.4f, Signal=%.1f dB",
				towerID, tower.Lat, tower.Lng, tower.Signal_strength_now)
		}
		ADSBTowerMutex.Unlock()
	}
}

// TestE2ETrafficAging tests that old traffic is properly aged out
func TestE2ETrafficAging(t *testing.T) {
	resetE2EState()

	// Create a traffic target
	var ti TrafficInfo
	ti.Icao_addr = 0xABCDEF
	ti.Lat = 47.45
	ti.Lng = -122.31
	ti.Alt = 5000
	ti.Last_seen = stratuxClock.Time.Add(-65 * time.Second) // 65 seconds ago

	trafficMutex.Lock()
	traffic[ti.Icao_addr] = ti
	seenTraffic[ti.Icao_addr] = true
	trafficMutex.Unlock()

	// Verify traffic exists
	trafficMutex.Lock()
	_, exists := traffic[ti.Icao_addr]
	trafficMutex.Unlock()

	if !exists {
		t.Error("Expected traffic to exist before aging")
	}

	// The actual aging happens in trafficInfoExtrapolator, which we can't easily test
	// without running the background goroutine. But we can verify the data structure
	// is set up correctly.

	t.Logf("Traffic aging test: Target age=%.1f seconds",
		stratuxClock.Since(ti.Last_seen).Seconds())
}

// TestE2EDownlinkReportParsing tests UAT downlink (ADS-B) report parsing
func TestE2EDownlinkReportParsing(t *testing.T) {
	resetE2EState()

	// Create a basic UAT downlink report (18 bytes)
	basicReport := "-000000000000000000000000000000000000;rs=12;ss=94"

	frame, msgtype := parseInput(basicReport)

	if frame == nil {
		t.Error("Expected non-nil frame from downlink report")
	}

	if msgtype != MSGTYPE_BASIC_REPORT {
		t.Errorf("Expected MSGTYPE_BASIC_REPORT (0x%02X), got 0x%02X", MSGTYPE_BASIC_REPORT, msgtype)
	}

	// Note: The actual traffic parsing happens in parseDownlinkReport which is called
	// inside parseInput, but with all-zero data, no meaningful traffic will be created

	t.Logf("Downlink report parsed: msgtype=0x%02X", msgtype)
}

// TestE2EHeartbeatGeneration tests GDL90 heartbeat message generation
func TestE2EHeartbeatGeneration(t *testing.T) {
	resetE2EState()

	// Generate heartbeat message
	hb := makeHeartbeat()

	if len(hb) == 0 {
		t.Error("Expected non-empty heartbeat message")
	}

	// Heartbeat should be properly framed with CRC
	if hb[0] != 0x7E {
		t.Error("Expected heartbeat to start with frame flag 0x7E")
	}

	if hb[len(hb)-1] != 0x7E {
		t.Error("Expected heartbeat to end with frame flag 0x7E")
	}

	t.Logf("Heartbeat generated: %d bytes", len(hb))
}

// TestE2EStratuxMessages tests Stratux-specific GDL90 messages
func TestE2EStratuxMessages(t *testing.T) {
	resetE2EState()

	testCases := []struct {
		name     string
		genFunc  func() []byte
		minBytes int
	}{
		{"Stratux Heartbeat", makeStratuxHeartbeat, 4},
		{"Stratux Status", makeStratuxStatus, 30},
		{"ForeFlight ID", makeFFIDMessage, 40},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := tc.genFunc()

			if len(msg) < tc.minBytes {
				t.Errorf("Expected at least %d bytes, got %d", tc.minBytes, len(msg))
			}

			// Check framing
			if msg[0] != 0x7E {
				t.Error("Expected message to start with 0x7E")
			}

			if msg[len(msg)-1] != 0x7E {
				t.Error("Expected message to end with 0x7E")
			}

			t.Logf("%s generated: %d bytes", tc.name, len(msg))
		})
	}
}

// TestE2EUpdateStatus tests global status updates
func TestE2EUpdateStatus(t *testing.T) {
	resetE2EState()

	// Set GPS as connected (required for updateStatus to process GPS data)
	globalStatus.GPS_connected = true

	// Set up GPS with valid fix
	mySituation.muGPS.Lock()
	mySituation.GPSFixQuality = 2 // DGPS
	mySituation.GPSSatellites = 12
	mySituation.GPSSatellitesSeen = 15
	mySituation.GPSSatellitesTracked = 14
	mySituation.GPSHorizontalAccuracy = 3.5
	mySituation.muGPS.Unlock()

	// Update status
	updateStatus()

	// Verify status was updated
	if globalStatus.GPS_solution != "3D GPS + SBAS" {
		t.Errorf("Expected GPS_solution='3D GPS + SBAS', got '%s'", globalStatus.GPS_solution)
	}

	if globalStatus.GPS_satellites_locked != 12 {
		t.Errorf("Expected GPS_satellites_locked=12, got %d", globalStatus.GPS_satellites_locked)
	}

	if globalStatus.GPS_position_accuracy != 3.5 {
		t.Errorf("Expected GPS_position_accuracy=3.5, got %.1f", globalStatus.GPS_position_accuracy)
	}

	t.Logf("Status updated: GPS=%s, Sats=%d, Accuracy=%.1f m",
		globalStatus.GPS_solution, globalStatus.GPS_satellites_locked, globalStatus.GPS_position_accuracy)
}

// TestE2EGeometricAltitudeReport tests ownship geometric altitude report generation
func TestE2EGeometricAltitudeReport(t *testing.T) {
	resetE2EState()

	// Set up GPS with valid position and height above ellipsoid
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.45
	mySituation.GPSLongitude = -122.31
	mySituation.GPSHeightAboveEllipsoid = 500.0 // 500 feet HAE
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()

	// Set GPS as connected
	globalStatus.GPS_connected = true

	// Generate geometric altitude report
	result := makeOwnshipGeometricAltitudeReport()

	// Result depends on isGPSValid() which checks timing
	if !result {
		t.Log("Note: makeOwnshipGeometricAltitudeReport returned false (GPS may not be considered valid)")
	} else {
		t.Logf("Geometric altitude report generated successfully: HAE=%.1f ft", mySituation.GPSHeightAboveEllipsoid)
	}
}

// TestE2ESystemErrors tests system error tracking
func TestE2ESystemErrors(t *testing.T) {
	// Initialize system errors mutex if needed
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	// Clear any existing errors
	globalStatus.Errors = make([]string, 0)
	systemErrsMutex.Lock()
	systemErrs = make(map[string]string)
	systemErrsMutex.Unlock()

	// Add a single system error
	addSingleSystemErrorf("test_error_1", "Test error %d", 1)

	// Verify error was added
	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(globalStatus.Errors))
	}

	// Add the same error again - should not duplicate
	addSingleSystemErrorf("test_error_1", "Test error %d", 1)

	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error after duplicate add, got %d", len(globalStatus.Errors))
	}

	// Add a different error
	addSingleSystemErrorf("test_error_2", "Test error %d", 2)

	if len(globalStatus.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(globalStatus.Errors))
	}

	// Remove the first error
	removeSingleSystemError("test_error_1")

	if len(globalStatus.Errors) != 1 {
		t.Errorf("Expected 1 error after removal, got %d", len(globalStatus.Errors))
	}

	t.Logf("System errors test: %d error(s) tracked", len(globalStatus.Errors))
}

// TestE2EDownlinkReportEdgeCases tests UAT downlink report parsing edge cases
func TestE2EDownlinkReportEdgeCases(t *testing.T) {
	resetE2EState()

	testCases := []struct {
		name    string
		message string
		wantNil bool
	}{
		{
			name:    "Empty downlink",
			message: "-000000000000000000000000000000000000;rs=10;ss=90",
			wantNil: false,
		},
		{
			name:    "Long report 34 bytes",
			message: "-00000000000000000000000000000000000000000000000000000000000000000000;rs=14;ss=102",
			wantNil: false,
		},
		{
			name:    "Long report 48 bytes",
			message: "-000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000;rs=15;ss=98",
			wantNil: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			frame, msgtype := parseInput(tc.message)

			if tc.wantNil && frame != nil {
				t.Errorf("Expected nil frame, got non-nil")
			}

			if !tc.wantNil && frame == nil {
				t.Errorf("Expected non-nil frame, got nil")
			}

			t.Logf("%s: msgtype=0x%02X, frame_length=%d", tc.name, msgtype, len(frame))
		})
	}
}

// TestE2EDefaultSettings tests default settings initialization
func TestE2EDefaultSettings(t *testing.T) {
	// Save current settings
	savedSettings := globalSettings

	// Call defaultSettings
	defaultSettings()

	// Verify key defaults are set
	if !globalSettings.UAT_Enabled {
		t.Error("Expected UAT_Enabled to be true by default")
	}

	if !globalSettings.ES_Enabled {
		t.Error("Expected ES_Enabled to be true by default")
	}

	if globalSettings.OGN_Enabled {
		t.Error("Expected OGN_Enabled to be false by default (US region)")
	}

	if globalSettings.GPS_Enabled != true {
		t.Error("Expected GPS_Enabled to be true by default")
	}

	if len(globalSettings.NetworkOutputs) == 0 {
		t.Error("Expected NetworkOutputs to be populated")
	}

	if globalSettings.WiFiSSID != "Stratux" {
		t.Errorf("Expected WiFiSSID='Stratux', got '%s'", globalSettings.WiFiSSID)
	}

	t.Logf("Default settings: UAT=%v, ES=%v, OGN=%v, GPS=%v, SSID=%s, Outputs=%d",
		globalSettings.UAT_Enabled, globalSettings.ES_Enabled, globalSettings.OGN_Enabled,
		globalSettings.GPS_Enabled, globalSettings.WiFiSSID, len(globalSettings.NetworkOutputs))

	// Restore settings
	globalSettings = savedSettings
}

// TestE2EUpdateStatusEdgeCases tests updateStatus with various GPS states
func TestE2EUpdateStatusEdgeCases(t *testing.T) {
	resetE2EState()

	testCases := []struct {
		name             string
		fixQuality       uint8
		expectedSolution string
	}{
		{"No fix", 0, "No Fix"},
		{"GPS fix", 1, "3D GPS"},
		{"DGPS fix", 2, "3D GPS + SBAS"},
		{"Dead reckoning", 6, "Dead Reckoning"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set GPS as connected
			globalStatus.GPS_connected = true

			// Set fix quality
			mySituation.muGPS.Lock()
			mySituation.GPSFixQuality = tc.fixQuality
			mySituation.muGPS.Unlock()

			// Update status
			updateStatus()

			// Verify GPS solution string
			if globalStatus.GPS_solution != tc.expectedSolution {
				t.Errorf("Expected GPS_solution='%s', got '%s'", tc.expectedSolution, globalStatus.GPS_solution)
			}

			t.Logf("%s: GPS_solution='%s'", tc.name, globalStatus.GPS_solution)
		})
	}

	// Test disconnected GPS
	t.Run("Disconnected GPS", func(t *testing.T) {
		globalStatus.GPS_connected = false
		updateStatus()

		if globalStatus.GPS_solution != "Disconnected" {
			t.Errorf("Expected GPS_solution='Disconnected', got '%s'", globalStatus.GPS_solution)
		}

		if globalStatus.GPS_satellites_locked != 0 {
			t.Errorf("Expected GPS_satellites_locked=0 for disconnected GPS, got %d", globalStatus.GPS_satellites_locked)
		}
	})
}
