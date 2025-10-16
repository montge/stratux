// flarm_nmea_input_test.go: Tests for FLARM NMEA input parsing
// Tests parseFlarmNmeaMessage, parseFlarmPFLAU, parseFlarmPFLAA functions

package main

import (
	"strings"
	"sync"
	"testing"
	"time"
)

// resetFlarmInputState resets state for FLARM input testing
func resetFlarmInputState() {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(20 * time.Millisecond)
	}

	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
	}
	if mySituation.muBaro == nil {
		mySituation.muBaro = &sync.Mutex{}
	}

	if trafficMutex == nil {
		trafficMutex = &sync.Mutex{}
	}

	// Reset traffic
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)
	trafficMutex.Unlock()

	// Set up valid GPS position for relative coordinate conversion
	mySituation.muGPS.Lock()
	mySituation.GPSFixQuality = 2 // DGPS fix
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = -122.3
	mySituation.GPSTrueCourse = 90                      // Heading east
	mySituation.GPSLastFixLocalTime = stratuxClock.Time // Set recent fix time
	mySituation.muGPS.Unlock()

	// Set GPS connected status
	globalStatus.GPS_connected = true

	// Set up baro altitude
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 5000.0
	mySituation.muBaro.Unlock()

	globalSettings.DEBUG = false
}

// TestParseFlarmPFLAU tests PFLAU (FLARM status) message parsing
func TestParseFlarmPFLAU(t *testing.T) {
	resetFlarmInputState()

	testCases := []struct {
		name          string
		pflauMsg      string
		expectedICAO  uint32
		expectedTail  string
		expectTraffic bool
	}{
		{
			name:          "Valid PFLAU with alarm",
			pflauMsg:      "PFLAU,1,1,2,1,2,45,2,152,1852,ABC123!N12345",
			expectedICAO:  0xABC123,
			expectedTail:  "N12345",
			expectTraffic: true,
		},
		{
			name:          "Valid PFLAU without tail",
			pflauMsg:      "PFLAU,1,1,2,1,1,-30,2,-100,500,DEF456",
			expectedICAO:  0xDEF456,
			expectedTail:  "",
			expectTraffic: true,
		},
		{
			name:          "PFLAU with zero distance (should create traffic)",
			pflauMsg:      "PFLAU,1,1,2,1,0,0,0,0,100,123456",
			expectedICAO:  0x123456,
			expectedTail:  "",
			expectTraffic: true,
		},
		{
			name:          "PFLAU too short (should be ignored)",
			pflauMsg:      "PFLAU,1,1,2,1",
			expectedICAO:  0,
			expectedTail:  "",
			expectTraffic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset traffic for each test
			trafficMutex.Lock()
			traffic = make(map[uint32]TrafficInfo)
			seenTraffic = make(map[uint32]bool)
			trafficMutex.Unlock()

			// Parse message
			fields := strings.Split(tc.pflauMsg, ",")
			parseFlarmPFLAU(fields)

			trafficMutex.Lock()
			defer trafficMutex.Unlock()

			if tc.expectTraffic {
				if len(traffic) == 0 {
					t.Error("Expected traffic to be created but traffic map is empty")
					return
				}

				// Find the traffic entry (could be at address or 1<<24|address)
				var ti TrafficInfo
				var found bool

				if t1, ok := traffic[tc.expectedICAO]; ok {
					ti = t1
					found = true
				} else if t2, ok := traffic[1<<24|tc.expectedICAO]; ok {
					ti = t2
					found = true
				}

				if !found {
					t.Errorf("Expected traffic with ICAO %06X but not found in map", tc.expectedICAO)
					return
				}

				if ti.Icao_addr != tc.expectedICAO {
					t.Errorf("Expected ICAO %06X, got %06X", tc.expectedICAO, ti.Icao_addr)
				}

				if tc.expectedTail != "" && ti.Tail != tc.expectedTail {
					t.Errorf("Expected tail %s, got %s", tc.expectedTail, ti.Tail)
				}

				if ti.Last_source != TRAFFIC_SOURCE_OGN {
					t.Errorf("Expected traffic source OGN (%d), got %d", TRAFFIC_SOURCE_OGN, ti.Last_source)
				}

				if ti.Lat == 0 && ti.Lng == 0 {
					t.Error("Traffic position not calculated (lat/lng both zero)")
				}

				t.Logf("PFLAU parsed: ICAO=%06X, Tail=%s, Lat=%.4f, Lng=%.4f, Alt=%d, Dist=%.0f",
					ti.Icao_addr, ti.Tail, ti.Lat, ti.Lng, ti.Alt, ti.Distance)
			} else {
				if len(traffic) > 0 {
					t.Errorf("Expected no traffic but got %d entries", len(traffic))
				}
			}
		})
	}
}

// TestParseFlarmPFLAA tests PFLAA (FLARM traffic) message parsing
func TestParseFlarmPFLAA(t *testing.T) {
	resetFlarmInputState()

	testCases := []struct {
		name          string
		pflaaMsg      string
		expectedICAO  uint32
		expectedTail  string
		expectedTrack float32
		expectedSpeed uint16
		expectTraffic bool
	}{
		{
			name:          "Valid PFLAA with position",
			pflaaMsg:      "PFLAA,2,1111,-750,152,1,ABC123!N12345,180,0,61,2.5,8",
			expectedICAO:  0xABC123,
			expectedTail:  "N12345",
			expectedTrack: 180,
			expectedSpeed: 118, // 61 m/s * 1.94384 = ~118 knots
			expectTraffic: true,
		},
		{
			name:          "Valid PFLAA without tail",
			pflaaMsg:      "PFLAA,0,2223,1502,304,2,DEF456,45,0,231,10.2,9",
			expectedICAO:  0xDEF456,
			expectedTail:  "",
			expectedTrack: 45,
			expectedSpeed: 449, // 231 m/s * 1.94384 = ~449 knots
			expectTraffic: true,
		},
		{
			name:          "PFLAA with negative coordinates",
			pflaaMsg:      "PFLAA,3,-556,-375,-91,1,789ABC,270,0,30,0.5,3",
			expectedICAO:  0x789ABC,
			expectedTail:  "",
			expectedTrack: 270,
			expectedSpeed: 58, // 30 m/s * 1.94384 = ~58 knots
			expectTraffic: true,
		},
		{
			name:          "PFLAA with glider (type 1)",
			pflaaMsg:      "PFLAA,0,1000,500,100,2,AABBCC,90,0,20,-2.5,1",
			expectedICAO:  0xAABBCC,
			expectedTail:  "",
			expectedTrack: 90,
			expectedSpeed: 38, // 20 m/s * 1.94384 = ~38 knots
			expectTraffic: true,
		},
		{
			name:          "PFLAA too short (should be ignored)",
			pflaaMsg:      "PFLAA,0,100,50",
			expectedICAO:  0,
			expectedTail:  "",
			expectTraffic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset traffic for each test
			trafficMutex.Lock()
			traffic = make(map[uint32]TrafficInfo)
			seenTraffic = make(map[uint32]bool)
			trafficMutex.Unlock()

			// Parse message
			fields := strings.Split(tc.pflaaMsg, ",")
			parseFlarmPFLAA(fields)

			trafficMutex.Lock()
			defer trafficMutex.Unlock()

			if tc.expectTraffic {
				if len(traffic) == 0 {
					t.Error("Expected traffic to be created but traffic map is empty")
					return
				}

				// Find the traffic entry (could be at different keys based on idType)
				var ti TrafficInfo
				var found bool

				// Try both possible keys (ICAO and non-ICAO)
				for key, traffic := range traffic {
					if traffic.Icao_addr == tc.expectedICAO {
						ti = traffic
						found = true
						t.Logf("Found traffic at key %08X", key)
						break
					}
				}

				if !found {
					t.Errorf("Expected traffic with ICAO %06X but not found in map", tc.expectedICAO)
					t.Logf("Traffic map contents:")
					for key, ti := range traffic {
						t.Logf("  Key %08X: ICAO=%06X", key, ti.Icao_addr)
					}
					return
				}

				if ti.Icao_addr != tc.expectedICAO {
					t.Errorf("Expected ICAO %06X, got %06X", tc.expectedICAO, ti.Icao_addr)
				}

				if tc.expectedTail != "" && ti.Tail != tc.expectedTail {
					t.Errorf("Expected tail %s, got %s", tc.expectedTail, ti.Tail)
				}

				if ti.Last_source != TRAFFIC_SOURCE_OGN {
					t.Errorf("Expected traffic source OGN (%d), got %d", TRAFFIC_SOURCE_OGN, ti.Last_source)
				}

				if ti.Track != tc.expectedTrack {
					t.Errorf("Expected track %.0f, got %.0f", tc.expectedTrack, ti.Track)
				}

				// Allow some tolerance for speed conversion
				if ti.Speed < tc.expectedSpeed-2 || ti.Speed > tc.expectedSpeed+2 {
					t.Errorf("Expected speed ~%d knots, got %d knots", tc.expectedSpeed, ti.Speed)
				}

				if !ti.Speed_valid {
					t.Error("Expected Speed_valid to be true")
				}

				if !ti.Position_valid {
					t.Error("Expected Position_valid to be true")
				}

				t.Logf("PFLAA parsed: ICAO=%06X, Tail=%s, Lat=%.4f, Lng=%.4f, Alt=%d, Track=%.0f, Speed=%d kts",
					ti.Icao_addr, ti.Tail, ti.Lat, ti.Lng, ti.Alt, ti.Track, ti.Speed)
			} else {
				if len(traffic) > 0 {
					t.Errorf("Expected no traffic but got %d entries", len(traffic))
				}
			}
		})
	}
}

// TestParseFlarmNmeaMessage tests the router function for PFLAU/PFLAA
func TestParseFlarmNmeaMessage(t *testing.T) {
	resetFlarmInputState()

	testCases := []struct {
		name          string
		nmeaMsg       string
		expectTraffic bool
	}{
		{
			name:          "Route to PFLAU parser",
			nmeaMsg:       "PFLAU,1,1,2,1,2,45,2,152,1852,ABC123!TEST",
			expectTraffic: true,
		},
		{
			name:          "Route to PFLAA parser",
			nmeaMsg:       "PFLAA,2,1111,-750,152,1,DEF456,180,0,61,2.5,8",
			expectTraffic: true,
		},
		{
			name:          "Unknown message type (should be ignored)",
			nmeaMsg:       "PFLAX,1,2,3,4,5",
			expectTraffic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset traffic
			trafficMutex.Lock()
			traffic = make(map[uint32]TrafficInfo)
			seenTraffic = make(map[uint32]bool)
			trafficMutex.Unlock()

			// Parse message
			fields := strings.Split(tc.nmeaMsg, ",")
			parseFlarmNmeaMessage(fields)

			trafficMutex.Lock()
			trafficCount := len(traffic)
			trafficMutex.Unlock()

			if tc.expectTraffic && trafficCount == 0 {
				t.Error("Expected traffic to be created but traffic map is empty")
			}

			if !tc.expectTraffic && trafficCount > 0 {
				t.Errorf("Expected no traffic but got %d entries", trafficCount)
			}
		})
	}
}

// TestParseFlarmPFLAUWithoutGPS tests PFLAU parsing when GPS is invalid
func TestParseFlarmPFLAUWithoutGPS(t *testing.T) {
	resetFlarmInputState()

	// Set GPS to invalid
	mySituation.muGPS.Lock()
	mySituation.GPSFixQuality = 0
	mySituation.muGPS.Unlock()

	pflauMsg := "PFLAU,1,1,0,1,2,45,2,152,1852,ABC123"
	fields := strings.Split(pflauMsg, ",")

	// Reset traffic
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	trafficMutex.Unlock()

	parseFlarmPFLAU(fields)

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	// Should not create traffic without valid GPS
	if len(traffic) > 0 {
		t.Error("Expected no traffic without valid GPS, but traffic was created")
	}
}

// TestParseFlarmPFLAACoordinateConversion tests coordinate conversion accuracy
func TestParseFlarmPFLAACoordinateConversion(t *testing.T) {
	resetFlarmInputState()

	// Test with known relative coordinates
	// Own position: 47.5N, 122.3W
	// Relative: North=1111m, East=-750m
	pflaaMsg := "PFLAA,2,1111,-750,152,1,TEST01,180,0,61,2.5,8"
	fields := strings.Split(pflaaMsg, ",")

	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	trafficMutex.Unlock()

	parseFlarmPFLAA(fields)

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	if len(traffic) == 0 {
		t.Fatal("Expected traffic to be created")
	}

	// Get the traffic entry
	var ti TrafficInfo
	for _, t := range traffic {
		ti = t
		break
	}

	// Verify position is north and west of ownship
	if ti.Lat <= mySituation.GPSLatitude {
		t.Errorf("Expected target north of ownship, got Lat=%.4f (own=%.4f)",
			ti.Lat, mySituation.GPSLatitude)
	}

	if ti.Lng >= mySituation.GPSLongitude {
		t.Errorf("Expected target west of ownship, got Lng=%.4f (own=%.4f)",
			ti.Lng, mySituation.GPSLongitude)
	}

	t.Logf("Coordinate conversion: Own=(%.4f,%.4f), RelN=%dm, RelE=%dm -> Target=(%.4f,%.4f)",
		mySituation.GPSLatitude, mySituation.GPSLongitude,
		1111, -750,
		ti.Lat, ti.Lng)
}

// TestParseFlarmPFLAAEmitterCategory tests aircraft type conversion
func TestParseFlarmPFLAAEmitterCategory(t *testing.T) {
	resetFlarmInputState()

	testCases := []struct {
		nmeaType    string
		expectedCat uint8
		desc        string
	}{
		{"1", 9, "Glider"},
		{"8", 1, "Piston (light)"},
		{"9", 3, "Jet (large)"},
		{"3", 7, "Helicopter"},
		{"D", 14, "UAV"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			trafficMutex.Lock()
			traffic = make(map[uint32]TrafficInfo)
			trafficMutex.Unlock()

			// Create PFLAA with specific aircraft type
			pflaaMsg := "PFLAA,0,1000,500,100,1,TEST02,90,0,20,0.0," + tc.nmeaType
			fields := strings.Split(pflaaMsg, ",")

			parseFlarmPFLAA(fields)

			trafficMutex.Lock()
			defer trafficMutex.Unlock()

			if len(traffic) == 0 {
				t.Fatal("Expected traffic to be created")
			}

			var ti TrafficInfo
			for _, t := range traffic {
				ti = t
				break
			}

			if ti.Emitter_category != tc.expectedCat {
				t.Errorf("Expected emitter category %d for NMEA type %s, got %d",
					tc.expectedCat, tc.nmeaType, ti.Emitter_category)
			}

			t.Logf("NMEA type %s -> GDL90 category %d (%s)", tc.nmeaType, ti.Emitter_category, tc.desc)
		})
	}
}

// TestParseFlarmPFLAUExisting1090ESTraffic tests that FLARM doesn't override recent 1090ES traffic
func TestParseFlarmPFLAUExisting1090ESTraffic(t *testing.T) {
	resetFlarmInputState()

	icao := uint32(0xABC123)

	// Create existing 1090ES traffic
	trafficMutex.Lock()
	traffic[icao] = TrafficInfo{
		Icao_addr:   icao,
		Tail:        "N12345",
		Lat:         47.51,
		Lng:         -122.31,
		Alt:         5500,
		Last_source: TRAFFIC_SOURCE_1090ES,
		Age:         2, // Recently seen (< 5 seconds)
	}
	trafficMutex.Unlock()

	// Try to update with FLARM
	pflauMsg := "PFLAU,1,1,2,1,2,45,2,152,1852,ABC123!DIFFERENT"
	fields := strings.Split(pflauMsg, ",")
	parseFlarmPFLAU(fields)

	trafficMutex.Lock()
	defer trafficMutex.Unlock()

	ti := traffic[icao]

	// Should still be 1090ES, not updated by FLARM
	if ti.Last_source != TRAFFIC_SOURCE_1090ES {
		t.Errorf("Expected traffic to remain 1090ES source, got %d", ti.Last_source)
	}

	if ti.Tail != "N12345" {
		t.Errorf("Expected tail to remain N12345 (not updated by FLARM), got %s", ti.Tail)
	}

	t.Log("Verified that recent 1090ES traffic is not overridden by FLARM")
}
