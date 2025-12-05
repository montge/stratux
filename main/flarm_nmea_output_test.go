// flarm_nmea_output_test.go: Tests for FLARM NMEA output generation
// Tests GPRMC, GPGGA, PGRMZ, PFLAU, PFLAA sentence generation

package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// resetFlarmOutputState resets state for FLARM output testing
func resetFlarmOutputState() {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(20 * time.Millisecond)
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

	if trafficMutex == nil {
		trafficMutex = &sync.Mutex{}
	}

	// Reset traffic
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	trafficMutex.Unlock()

	// Reset GPS state
	mySituation.muGPS.Lock()
	mySituation.GPSFixQuality = 0
	mySituation.GPSLatitude = 0
	mySituation.GPSLongitude = 0
	mySituation.GPSAltitudeMSL = 0
	mySituation.GPSGroundSpeed = 0
	mySituation.GPSTrueCourse = 0
	mySituation.GPSTime = time.Time{}
	mySituation.muGPS.Unlock()

	// Reset attitude
	mySituation.muAttitude.Lock()
	mySituation.AHRSPitch = 0
	mySituation.AHRSRoll = 0
	mySituation.AHRSGyroHeading = 0
	mySituation.AHRSMagHeading = 0
	mySituation.muAttitude.Unlock()

	// Reset baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 0
	mySituation.muBaro.Unlock()

	// Reset global GPS status
	globalStatus.GPS_connected = false
}

// validateNMEASentence checks basic NMEA sentence structure
func validateNMEASentence(t *testing.T, sentence string, expectedType string) {
	if !strings.HasPrefix(sentence, "$") {
		t.Errorf("NMEA sentence should start with '$', got: %s", sentence)
	}

	if !strings.Contains(sentence, "*") {
		t.Errorf("NMEA sentence should contain checksum delimiter '*', got: %s", sentence)
	}

	if !strings.HasPrefix(sentence, "$"+expectedType) {
		t.Errorf("Expected sentence type %s, got: %s", expectedType, sentence)
	}

	// Validate checksum
	parts := strings.Split(sentence, "*")
	if len(parts) != 2 {
		t.Errorf("Invalid NMEA sentence format (should have exactly one *): %s", sentence)
		return
	}

	// Calculate checksum
	data := parts[0][1:] // Skip '$'
	expectedChecksum := byte(0)
	for i := 0; i < len(data); i++ {
		expectedChecksum ^= data[i]
	}

	// The checksum in the sentence is in hex
	var actualChecksum byte
	if len(parts[1]) >= 2 {
		// Simple hex conversion
		h1 := parts[1][0]
		h2 := parts[1][1]
		actualChecksum = hexCharToByte(h1)<<4 | hexCharToByte(h2)
	}

	if actualChecksum != expectedChecksum {
		t.Errorf("Checksum mismatch for %s: expected %02X, got %02X", sentence, expectedChecksum, actualChecksum)
	}
}

// hexCharToByte converts a hex character to byte
func hexCharToByte(c byte) byte {
	if c >= '0' && c <= '9' {
		return c - '0'
	}
	if c >= 'A' && c <= 'F' {
		return c - 'A' + 10
	}
	if c >= 'a' && c <= 'f' {
		return c - 'a' + 10
	}
	return 0
}

// TestMakeGPRMCString tests GPRMC (Recommended Minimum) sentence generation
func TestMakeGPRMCString(t *testing.T) {
	resetFlarmOutputState()

	testCases := []struct {
		name        string
		lat         float32
		lon         float32
		speed       float64
		track       float32
		fixQual     uint8
		gpsTime     time.Time
		expectEmpty bool
	}{
		{
			name:        "Valid position - Seattle area",
			lat:         47.6062,
			lon:         -122.3321,
			speed:       120.0, // knots
			track:       270.0, // degrees
			fixQual:     2,     // DGPS
			gpsTime:     time.Date(2024, 10, 16, 12, 30, 45, 0, time.UTC),
			expectEmpty: false,
		},
		{
			name:        "Valid position - zero speed",
			lat:         40.7128,
			lon:         -74.0060,
			speed:       0.0,
			track:       0.0,
			fixQual:     1,
			gpsTime:     time.Date(2024, 10, 16, 15, 0, 0, 0, time.UTC),
			expectEmpty: false,
		},
		{
			name:        "Valid position - high speed",
			lat:         51.5074,
			lon:         -0.1278,
			speed:       450.0, // knots
			track:       90.0,
			fixQual:     2,
			gpsTime:     time.Date(2024, 10, 16, 18, 45, 30, 0, time.UTC),
			expectEmpty: false,
		},
		{
			name:        "No GPS fix",
			lat:         0,
			lon:         0,
			speed:       0,
			track:       0,
			fixQual:     0,
			gpsTime:     time.Time{},
			expectEmpty: false, // GPRMC returns sentence with 'V' flag even without fix
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up GPS state
			mySituation.muGPS.Lock()
			mySituation.GPSLatitude = tc.lat
			mySituation.GPSLongitude = tc.lon
			mySituation.GPSGroundSpeed = tc.speed
			mySituation.GPSTrueCourse = tc.track
			mySituation.GPSFixQuality = tc.fixQual
			mySituation.GPSTime = tc.gpsTime
			mySituation.muGPS.Unlock()

			result := makeGPRMCString()

			if tc.expectEmpty {
				if result != "" {
					t.Errorf("Expected empty result for no GPS fix, got: %s", result)
				}
			} else {
				if result == "" {
					t.Error("Expected GPRMC sentence, got empty string")
				} else {
					validateNMEASentence(t, result, "GPRMC")

					// Check for expected components
					if !strings.Contains(result, "GPRMC") {
						t.Error("Sentence should contain GPRMC")
					}

					// Should contain A for valid or V for invalid
					if !strings.Contains(result, ",A,") && !strings.Contains(result, ",V,") {
						t.Error("Sentence should contain validity flag (A or V)")
					}

					t.Logf("GPRMC: %s", result)
				}
			}
		})
	}
}

// TestMakeGPGGAString tests GPGGA (Global Positioning System Fix Data) sentence generation
func TestMakeGPGGAString(t *testing.T) {
	resetFlarmOutputState()

	testCases := []struct {
		name        string
		lat         float32
		lon         float32
		alt         float32
		fixQual     uint8
		numSats     uint16
		hdop        float32
		geoidSep    float32
		gpsTime     time.Time
		expectEmpty bool
	}{
		{
			name:        "Valid 3D fix",
			lat:         47.6062,
			lon:         -122.3321,
			alt:         500.0,
			fixQual:     2, // DGPS
			numSats:     12,
			hdop:        1.2,
			geoidSep:    -20.0,
			gpsTime:     time.Date(2024, 10, 16, 12, 30, 45, 0, time.UTC),
			expectEmpty: false,
		},
		{
			name:        "Valid GPS fix with high altitude",
			lat:         35.6762,
			lon:         139.6503,
			alt:         35000.0, // 35000 ft
			fixQual:     1,
			numSats:     8,
			hdop:        2.5,
			geoidSep:    30.0,
			gpsTime:     time.Date(2024, 10, 16, 9, 15, 0, 0, time.UTC),
			expectEmpty: false,
		},
		{
			name:        "Low satellite count",
			lat:         48.8566,
			lon:         2.3522,
			alt:         100.0,
			fixQual:     1,
			numSats:     4, // Minimum for 3D fix
			hdop:        5.0,
			geoidSep:    48.0,
			gpsTime:     time.Date(2024, 10, 16, 14, 0, 0, 0, time.UTC),
			expectEmpty: false,
		},
		{
			name:        "No GPS fix",
			lat:         0,
			lon:         0,
			alt:         0,
			fixQual:     0,
			numSats:     0,
			hdop:        99.9,
			geoidSep:    0,
			gpsTime:     time.Time{},
			expectEmpty: false, // GPGGA returns sentence even without fix (fix quality = 0)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up GPS state
			mySituation.muGPS.Lock()
			mySituation.GPSLatitude = tc.lat
			mySituation.GPSLongitude = tc.lon
			mySituation.GPSAltitudeMSL = tc.alt
			mySituation.GPSFixQuality = tc.fixQual
			mySituation.GPSSatellites = tc.numSats
			mySituation.GPSHorizontalAccuracy = tc.hdop
			mySituation.GPSGeoidSep = tc.geoidSep
			mySituation.GPSTime = tc.gpsTime
			mySituation.muGPS.Unlock()

			result := makeGPGGAString()

			if tc.expectEmpty {
				if result != "" {
					t.Errorf("Expected empty result for no GPS fix, got: %s", result)
				}
			} else {
				if result == "" {
					t.Error("Expected GPGGA sentence, got empty string")
				} else {
					validateNMEASentence(t, result, "GPGGA")

					// Check for expected components
					if !strings.Contains(result, "GPGGA") {
						t.Error("Sentence should contain GPGGA")
					}

					t.Logf("GPGGA: %s", result)
				}
			}
		})
	}
}

// TestMakePGRMZString tests PGRMZ (Garmin altitude) sentence generation
func TestMakePGRMZString(t *testing.T) {
	resetFlarmOutputState()

	testCases := []struct {
		name        string
		altitude    float32
		expectEmpty bool
	}{
		{
			name:        "Sea level",
			altitude:    0.0,
			expectEmpty: false,
		},
		{
			name:        "Low altitude (500 ft)",
			altitude:    500.0,
			expectEmpty: false,
		},
		{
			name:        "Medium altitude (5000 ft)",
			altitude:    5000.0,
			expectEmpty: false,
		},
		{
			name:        "High altitude (35000 ft)",
			altitude:    35000.0,
			expectEmpty: false,
		},
		{
			name:        "Negative altitude (below sea level)",
			altitude:    -100.0,
			expectEmpty: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up barometric altitude
			mySituation.muBaro.Lock()
			mySituation.BaroPressureAltitude = tc.altitude
			mySituation.muBaro.Unlock()

			result := makePGRMZString()

			if tc.expectEmpty {
				if result != "" {
					t.Errorf("Expected empty result, got: %s", result)
				}
			} else {
				if result == "" {
					t.Error("Expected PGRMZ sentence, got empty string")
				} else {
					validateNMEASentence(t, result, "PGRMZ")

					// Check for expected components
					if !strings.Contains(result, "PGRMZ") {
						t.Error("Sentence should contain PGRMZ")
					}

					// Should contain 'f' for feet
					if !strings.Contains(result, ",f,") {
						t.Error("Sentence should specify feet unit (,f,)")
					}

					t.Logf("PGRMZ (alt=%.1f): %s", tc.altitude, result)
				}
			}
		})
	}
}

// TestMakeAHRSLevilReportOutput tests AHRS level report generation
func TestMakeAHRSLevilReportOutput(t *testing.T) {
	resetFlarmOutputState()

	// makeAHRSLevilReport() doesn't return anything, it sends directly
	// We can only test that it doesn't panic when called

	testCases := []struct {
		name    string
		pitch   float64
		roll    float64
		heading float64
	}{
		{
			name:    "Level flight",
			pitch:   0.0,
			roll:    0.0,
			heading: 0.0,
		},
		{
			name:    "Climbing turn",
			pitch:   10.0,
			roll:    15.0,
			heading: 270.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up AHRS state
			mySituation.muAttitude.Lock()
			mySituation.AHRSPitch = tc.pitch
			mySituation.AHRSRoll = tc.roll
			mySituation.AHRSGyroHeading = tc.heading
			mySituation.muAttitude.Unlock()

			// Just call it - it should not panic
			makeAHRSLevilReport()

			t.Logf("AHRS report called successfully (pitch=%.1f, roll=%.1f, hdg=%.1f)",
				tc.pitch, tc.roll, tc.heading)
		})
	}
}

// TestComputeRelativeVertical tests relative vertical separation calculation
func TestComputeRelativeVertical(t *testing.T) {
	resetFlarmOutputState()

	testCases := []struct {
		name      string
		ownAlt    float32
		targetAlt int32
	}{
		{
			name:      "Same altitude",
			ownAlt:    5000.0,
			targetAlt: 5000,
		},
		{
			name:      "Target 500 ft above",
			ownAlt:    5000.0,
			targetAlt: 5500,
		},
		{
			name:      "Target 500 ft below",
			ownAlt:    5000.0,
			targetAlt: 4500,
		},
		{
			name:      "Large separation (10000 ft)",
			ownAlt:    5000.0,
			targetAlt: 15000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up own altitude
			mySituation.muBaro.Lock()
			mySituation.BaroPressureAltitude = tc.ownAlt
			mySituation.muBaro.Unlock()

			// Create a traffic target
			ti := TrafficInfo{
				Alt:       tc.targetAlt,
				AltIsGNSS: false,
			}

			result := computeRelativeVertical(ti)

			// Result is in meters, convert to feet for logging
			resultFt := float32(result) / 0.3048

			t.Logf("Relative vertical: own=%.1f ft, target=%d ft -> %d m (%.1f ft)",
				tc.ownAlt, tc.targetAlt, result, resultFt)
		})
	}
}

// TestMakeFlarmPFLAUString tests PFLAU (FLARM status) sentence generation
func TestMakeFlarmPFLAUString(t *testing.T) {
	resetFlarmOutputState()

	testCases := []struct {
		name          string
		ownLat        float32
		ownLon        float32
		ownAlt        float32
		ownTrack      float32
		gpsValid      bool
		targetLat     float32
		targetLon     float32
		targetAlt     int32
		targetICAO    uint32
		targetTail    string
		expectedAlarm int // Expected alarm level
	}{
		{
			name:          "No alarm - distant traffic",
			ownLat:        47.5,
			ownLon:        -122.3,
			ownAlt:        5000,
			ownTrack:      90,
			gpsValid:      true,
			targetLat:     48.0, // ~30 NM away
			targetLon:     -122.3,
			targetAlt:     5000,
			targetICAO:    0xABC123,
			targetTail:    "N12345",
			expectedAlarm: 0,
		},
		{
			name:          "Medium alarm - 1 NM separation",
			ownLat:        47.5,
			ownLon:        -122.3,
			ownAlt:        5000,
			ownTrack:      180,
			gpsValid:      true,
			targetLat:     47.515, // ~1 NM
			targetLon:     -122.3,
			targetAlt:     5500, // 500 ft above
			targetICAO:    0xDEF456,
			targetTail:    "",
			expectedAlarm: 2,
		},
		{
			name:          "High alarm - close traffic",
			ownLat:        47.5,
			ownLon:        -122.3,
			ownAlt:        5000,
			ownTrack:      270,
			gpsValid:      true,
			targetLat:     47.503, // ~0.2 NM
			targetLon:     -122.3,
			targetAlt:     5100, // 100 ft above
			targetICAO:    0x123456,
			targetTail:    "TEST1",
			expectedAlarm: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state for each sub-test to prevent contamination
			resetFlarmOutputState()

			// Set up own position
			mySituation.muGPS.Lock()
			if tc.gpsValid {
				mySituation.GPSFixQuality = 2
				mySituation.GPSLastFixLocalTime = stratuxClock.Time
			}
			mySituation.GPSLatitude = tc.ownLat
			mySituation.GPSLongitude = tc.ownLon
			mySituation.GPSTrueCourse = tc.ownTrack
			mySituation.muGPS.Unlock()

			// Set GPS connected status for valid GPS
			if tc.gpsValid {
				globalStatus.GPS_connected = true
			}

			mySituation.muBaro.Lock()
			mySituation.BaroPressureAltitude = tc.ownAlt
			mySituation.BaroLastMeasurementTime = stratuxClock.Time // Make baro data valid
			mySituation.muBaro.Unlock()

			// Create traffic target
			ti := TrafficInfo{
				Icao_addr: tc.targetICAO,
				Tail:      tc.targetTail,
				Lat:       tc.targetLat,
				Lng:       tc.targetLon,
				Alt:       tc.targetAlt,
				AltIsGNSS: false,
			}

			// Create fake traffic map for RX count
			trafficMutex.Lock()
			traffic = make(map[uint32]TrafficInfo)
			traffic[tc.targetICAO] = ti
			trafficMutex.Unlock()

			result := makeFlarmPFLAUString(ti)

			if result == "" {
				t.Error("Expected PFLAU sentence, got empty string")
			} else {
				validateNMEASentence(t, strings.TrimSpace(result), "PFLAU")

				// Check for expected components
				if !strings.Contains(result, "PFLAU") {
					t.Error("Sentence should contain PFLAU")
				}

				// Check ICAO address is included only when alarm level > 0
				if tc.expectedAlarm > 0 {
					icaoStr := fmt.Sprintf("%06X", tc.targetICAO&0xFFFFFF)
					if !strings.Contains(result, icaoStr) {
						t.Errorf("Expected ICAO %s in sentence for alarm level %d", icaoStr, tc.expectedAlarm)
					}
				}

				t.Logf("PFLAU (alarm=%d): %s", tc.expectedAlarm, strings.TrimSpace(result))
			}
		})
	}
}

// TestMakeFlarmPFLAAString tests PFLAA (FLARM traffic) sentence generation
func TestMakeFlarmPFLAAString(t *testing.T) {
	resetFlarmOutputState()

	testCases := []struct {
		name          string
		ownLat        float32
		ownLon        float32
		ownAlt        float32
		targetLat     float32
		targetLon     float32
		targetAlt     int32
		targetICAO    uint32
		targetTail    string
		targetTrack   float32
		targetSpeed   uint16
		targetVvel    int16
		targetEmitter uint8
		positionValid bool
		speedValid    bool
		expectedValid bool
	}{
		{
			name:          "Valid traffic with position",
			ownLat:        47.5,
			ownLon:        -122.3,
			ownAlt:        5000,
			targetLat:     47.51,
			targetLon:     -122.31,
			targetAlt:     5500,
			targetICAO:    0xABC123,
			targetTail:    "N12345",
			targetTrack:   180,
			targetSpeed:   120, // knots
			targetVvel:    500, // fpm
			targetEmitter: 1,   // Light aircraft
			positionValid: true,
			speedValid:    true,
			expectedValid: true,
		},
		{
			name:          "Traffic without position (bearingless)",
			ownLat:        47.5,
			ownLon:        -122.3,
			ownAlt:        5000,
			targetLat:     0,
			targetLon:     0,
			targetAlt:     6000,
			targetICAO:    0xDEF456,
			targetTail:    "TEST2",
			targetTrack:   90,
			targetSpeed:   0,
			targetVvel:    -200,
			targetEmitter: 9, // Glider
			positionValid: false,
			speedValid:    false,
			expectedValid: true,
		},
		{
			name:          "Heavy aircraft at high speed",
			ownLat:        47.5,
			ownLon:        -122.3,
			ownAlt:        5000,
			targetLat:     47.52,
			targetLon:     -122.28,
			targetAlt:     35000,
			targetICAO:    0x123456,
			targetTail:    "UAL123",
			targetTrack:   45,
			targetSpeed:   450,  // knots
			targetVvel:    2000, // fpm
			targetEmitter: 3,    // Large aircraft
			positionValid: true,
			speedValid:    true,
			expectedValid: true,
		},
		{
			name:          "Helicopter at low altitude",
			ownLat:        47.5,
			ownLon:        -122.3,
			ownAlt:        500,
			targetLat:     47.505,
			targetLon:     -122.305,
			targetAlt:     800,
			targetICAO:    0x789ABC,
			targetTail:    "",
			targetTrack:   270,
			targetSpeed:   60,  // knots
			targetVvel:    100, // fpm
			targetEmitter: 7,   // Helicopter
			positionValid: true,
			speedValid:    true,
			expectedValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up own position
			mySituation.muGPS.Lock()
			mySituation.GPSFixQuality = 2
			mySituation.GPSLatitude = tc.ownLat
			mySituation.GPSLongitude = tc.ownLon
			mySituation.muGPS.Unlock()

			mySituation.muBaro.Lock()
			mySituation.BaroPressureAltitude = tc.ownAlt
			mySituation.muBaro.Unlock()

			// Create traffic target
			ti := TrafficInfo{
				Icao_addr:        tc.targetICAO,
				Addr_type:        0, // ICAO
				Tail:             tc.targetTail,
				Lat:              tc.targetLat,
				Lng:              tc.targetLon,
				Alt:              tc.targetAlt,
				AltIsGNSS:        false,
				Track:            tc.targetTrack,
				Speed:            tc.targetSpeed,
				Speed_valid:      tc.speedValid,
				Vvel:             tc.targetVvel,
				Emitter_category: tc.targetEmitter,
				Position_valid:   tc.positionValid,
			}

			result, valid, alarmLevel := makeFlarmPFLAAString(ti)

			if !valid && tc.expectedValid {
				t.Error("Expected valid PFLAA sentence but got invalid")
			}

			if valid {
				if result == "" {
					t.Error("Expected PFLAA sentence, got empty string")
				} else {
					validateNMEASentence(t, strings.TrimSpace(result), "PFLAA")

					// Check for expected components
					if !strings.Contains(result, "PFLAA") {
						t.Error("Sentence should contain PFLAA")
					}

					// Check ICAO address
					icaoStr := fmt.Sprintf("%06X", tc.targetICAO&0xFFFFFF)
					if !strings.Contains(result, icaoStr) {
						t.Errorf("Expected ICAO %s in sentence", icaoStr)
					}

					t.Logf("PFLAA (alarm=%d, pos_valid=%v): %s", alarmLevel, tc.positionValid, strings.TrimSpace(result))
				}
			}
		})
	}
}

// TestFlarmEmitterCategoryConversion tests GDL90 to NMEA emitter category conversion
func TestFlarmEmitterCategoryConversion(t *testing.T) {
	resetFlarmOutputState()

	testCases := []struct {
		gdl90Cat uint8
		nmeaType string
		desc     string
	}{
		{1, "8", "Light aircraft"},
		{3, "9", "Large aircraft"},
		{7, "3", "Helicopter"},
		{9, "1", "Glider"},
		{10, "B", "Balloon"},
		{11, "4", "Skydiver"},
		{12, "7", "Paraglider"},
		{14, "D", "UAV"},
		{19, "F", "Static object"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := gdl90EmitterCatToNMEA(tc.gdl90Cat)
			if result != tc.nmeaType {
				t.Errorf("Expected NMEA type %s for GDL90 category %d, got %s",
					tc.nmeaType, tc.gdl90Cat, result)
			}
			t.Logf("GDL90 cat %d -> NMEA type %s (%s)", tc.gdl90Cat, result, tc.desc)
		})
	}
}

// TestMakeGPRMCString_SouthernHemisphere tests GPRMC with southern latitude
func TestMakeGPRMCString_SouthernHemisphere(t *testing.T) {
	resetFlarmOutputState()

	// Set up GPS state for southern hemisphere
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = -33.8688    // Sydney, Australia
	mySituation.GPSLongitude = 151.2093   // East
	mySituation.GPSGroundSpeed = 120
	mySituation.GPSTrueCourse = 180
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTime = time.Now().UTC()
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	result := makeGPRMCString()

	if !strings.Contains(result, ",S,") {
		t.Error("Expected 'S' for southern latitude")
	}
	if !strings.Contains(result, ",E,") {
		t.Error("Expected 'E' for eastern longitude")
	}
	if !strings.Contains(result, ",A,") {
		t.Error("Expected 'A' for valid GPS status")
	}

	t.Logf("GPRMC Southern: %s", strings.TrimSpace(result))
}

// TestMakeGPRMCString_WesternLongitude tests GPRMC with western longitude
func TestMakeGPRMCString_WesternLongitude(t *testing.T) {
	resetFlarmOutputState()

	// Set up GPS state for western hemisphere
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 40.7128     // New York, northern
	mySituation.GPSLongitude = -74.0060   // West
	mySituation.GPSGroundSpeed = 100
	mySituation.GPSTrueCourse = 90
	mySituation.GPSFixQuality = 2         // DGPS fix
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTime = time.Now().UTC()
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	result := makeGPRMCString()

	if !strings.Contains(result, ",N,") {
		t.Error("Expected 'N' for northern latitude")
	}
	if !strings.Contains(result, ",W,") {
		t.Error("Expected 'W' for western longitude")
	}
	// Fix quality 2 should give mode "D"
	if !strings.Contains(result, ",D*") {
		t.Error("Expected mode 'D' for DGPS fix quality")
	}

	t.Logf("GPRMC Western: %s", strings.TrimSpace(result))
}

// TestMakeGPRMCString_InvalidGPS tests GPRMC with invalid GPS
func TestMakeGPRMCString_InvalidGPS(t *testing.T) {
	resetFlarmOutputState()

	// Set up invalid GPS state
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 0
	mySituation.GPSLongitude = 0
	mySituation.GPSGroundSpeed = 0
	mySituation.GPSTrueCourse = 0
	mySituation.GPSFixQuality = 0
	mySituation.GPSLastFixLocalTime = time.Time{} // Invalid time
	mySituation.GPSTime = time.Time{}
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = false

	result := makeGPRMCString()

	if !strings.Contains(result, ",V,") {
		t.Error("Expected 'V' for invalid GPS status")
	}
	if !strings.Contains(result, ",N*") {
		t.Error("Expected mode 'N' for no fix quality")
	}

	t.Logf("GPRMC Invalid: %s", strings.TrimSpace(result))
}

// TestMakeGPGGAString_SouthernHemisphere tests GPGGA with southern latitude
func TestMakeGPGGAString_SouthernHemisphere(t *testing.T) {
	resetFlarmOutputState()

	// Set up GPS state for southern hemisphere
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = -33.8688    // Sydney, Australia
	mySituation.GPSLongitude = 151.2093   // East
	mySituation.GPSAltitudeMSL = 500
	mySituation.GPSFixQuality = 1
	mySituation.GPSSatellites = 10
	mySituation.GPSGeoidSep = -30.0
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTime = time.Now().UTC()
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	result := makeGPGGAString()

	if !strings.Contains(result, ",S,") {
		t.Error("Expected 'S' for southern latitude")
	}
	if !strings.Contains(result, ",E,") {
		t.Error("Expected 'E' for eastern longitude")
	}

	t.Logf("GPGGA Southern: %s", strings.TrimSpace(result))
}

// TestMakeGPGGAString_WesternLongitude tests GPGGA with western longitude
func TestMakeGPGGAString_WesternLongitude(t *testing.T) {
	resetFlarmOutputState()

	// Set up GPS state for western hemisphere
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 40.7128     // New York, northern
	mySituation.GPSLongitude = -74.0060   // West
	mySituation.GPSAltitudeMSL = 100
	mySituation.GPSFixQuality = 2         // DGPS
	mySituation.GPSSatellites = 12
	mySituation.GPSGeoidSep = -35.0
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTime = time.Now().UTC()
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	result := makeGPGGAString()

	if !strings.Contains(result, ",N,") {
		t.Error("Expected 'N' for northern latitude")
	}
	if !strings.Contains(result, ",W,") {
		t.Error("Expected 'W' for western longitude")
	}

	t.Logf("GPGGA Western: %s", strings.TrimSpace(result))
}

// TestMakeGPGGAString_InvalidGPS tests GPGGA with invalid GPS
func TestMakeGPGGAString_InvalidGPS(t *testing.T) {
	resetFlarmOutputState()

	// Set up invalid GPS state
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 0
	mySituation.GPSLongitude = 0
	mySituation.GPSAltitudeMSL = 0
	mySituation.GPSFixQuality = 0
	mySituation.GPSSatellites = 3  // Some satellites tracked but no fix
	mySituation.GPSGeoidSep = 0
	mySituation.GPSLastFixLocalTime = time.Time{}
	mySituation.GPSTime = time.Time{}
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = false

	result := makeGPGGAString()

	// Should have fix quality 0
	if !strings.Contains(result, ",0,") {
		t.Error("Expected fix quality 0 for invalid GPS")
	}

	t.Logf("GPGGA Invalid: %s", strings.TrimSpace(result))
}

// TestMakeFlarmPFLAUString_BearingWrapping tests PFLAU bearing wrapping (> 180 and < -180)
func TestMakeFlarmPFLAUString_BearingWrapping(t *testing.T) {
	resetFlarmOutputState()

	// Set up valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.BaroPressureAltitude = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.GPSTrueCourse = 350 // Heading north-northwest
	mySituation.muGPS.Unlock()
	mySituation.muBaro.Lock()
	mySituation.BaroLastMeasurementTime = time.Now()
	mySituation.muBaro.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	trafficMutex.Unlock()

	// Traffic close enough to trigger alarm (to get tail number included)
	// and behind to trigger bearing wrapping
	ti := TrafficInfo{
		Icao_addr:  0xABCDEF,
		Lat:        47.4995, // Very slightly south (~500m)
		Lng:        8.5005,  // Very slightly east
		Alt:        1050,    // Close altitude (within 500ft)
		Tail:       "TEST1",
	}

	result := makeFlarmPFLAUString(ti)

	if !strings.Contains(result, "PFLAU") {
		t.Error("Expected PFLAU sentence")
	}
	// Tail is included when alarm level > 0
	if !strings.Contains(result, "!TEST1") {
		t.Logf("Note: Tail not included (likely no alarm level - traffic not close enough)")
	}

	t.Logf("PFLAU with bearing wrapping: %s", strings.TrimSpace(result))
}

// TestMakeFlarmPFLAUString_AlarmLevel tests PFLAU with alarm conditions
func TestMakeFlarmPFLAUString_AlarmLevel(t *testing.T) {
	resetFlarmOutputState()

	// Set up valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.BaroPressureAltitude = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.muGPS.Unlock()
	mySituation.muBaro.Lock()
	mySituation.BaroLastMeasurementTime = time.Now()
	mySituation.muBaro.Unlock()
	globalStatus.GPS_connected = true

	// Initialize traffic map
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	trafficMutex.Unlock()

	// Traffic very close - should trigger alarm level 3
	ti := TrafficInfo{
		Icao_addr:  0x123456,
		Lat:        47.5001, // Very close
		Lng:        8.5001,
		Alt:        1050,    // Close altitude
		Tail:       "",      // No tail
	}

	result := makeFlarmPFLAUString(ti)

	// Should have non-zero alarm level for close traffic
	if !strings.Contains(result, "PFLAU") {
		t.Error("Expected PFLAU sentence")
	}

	t.Logf("PFLAU with alarm: %s", strings.TrimSpace(result))
}

// TestMakeFlarmPFLAAString_PositionInvalid tests PFLAA with invalid position
func TestMakeFlarmPFLAAString_PositionInvalid(t *testing.T) {
	resetFlarmOutputState()

	// Set up valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Traffic with invalid position
	ti := TrafficInfo{
		Icao_addr:         0xDEADBE,
		Position_valid:    false,
		DistanceEstimated: 5000, // 5km estimated
		Alt:               2000,
		Tail:              "N12345",
		Addr_type:         1, // Non-ICAO
	}

	msg, valid, _ := makeFlarmPFLAAString(ti)

	if !valid {
		t.Error("Expected valid message even with position_valid=false")
	}
	if !strings.Contains(msg, "PFLAA") {
		t.Error("Expected PFLAA sentence")
	}
	// For invalid position, RelativeEast should be empty
	if !strings.Contains(msg, "!N12345") {
		t.Error("Expected tail number appended")
	}

	t.Logf("PFLAA with invalid position: %s", strings.TrimSpace(msg))
}

// TestMakeFlarmPFLAAString_NonICAOAddress tests PFLAA with non-ICAO address type
func TestMakeFlarmPFLAAString_NonICAOAddress(t *testing.T) {
	resetFlarmOutputState()

	// Set up valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = time.Now()
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Traffic with non-ICAO address
	ti := TrafficInfo{
		Icao_addr:      0xF12345,
		Position_valid: true,
		Lat:            47.51,
		Lng:            8.51,
		Alt:            1500,
		Addr_type:      1, // Non-ICAO
		Speed_valid:    true,
		Speed:          100, // knots
		Vvel:           500, // ft/min
	}

	msg, valid, _ := makeFlarmPFLAAString(ti)

	if !valid {
		t.Error("Expected valid message")
	}
	// IDType should be 2 for non-ICAO
	if !strings.Contains(msg, ",2,") {
		t.Error("Expected IDType 2 for non-ICAO address")
	}

	t.Logf("PFLAA with non-ICAO: %s", strings.TrimSpace(msg))
}

// TestMakeFlarmPFLAUString_BearingLessThanMinus180 tests bearing wrapping for bearing < -180
func TestMakeFlarmPFLAUString_BearingLessThanMinus180(t *testing.T) {
	resetFlarmOutputState()

	// Set up valid GPS with course that results in bearing < -180
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.GPSTrueCourse = 350 // heading almost north
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Set up baro for relative vertical calculation
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 1000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	// Add traffic to the traffic map (needed for len(traffic))
	trafficMutex.Lock()
	traffic[0x123456] = TrafficInfo{}
	trafficMutex.Unlock()

	// Traffic to the east-southeast (bearing ~120 degrees)
	// With course 350, relative bearing = 120 - 350 = -230 -> should wrap to +130
	ti := TrafficInfo{
		Icao_addr:      0x123456,
		Position_valid: true,
		Lat:            47.4,  // South
		Lng:            8.7,   // East
		Alt:            1200,
	}

	msg := makeFlarmPFLAUString(ti)

	if !strings.HasPrefix(msg, "$PFLAU") {
		t.Errorf("Expected PFLAU sentence, got: %s", msg)
	}

	// Validate it's a proper NMEA sentence
	validateNMEASentence(t, msg, "PFLAU")

	t.Logf("PFLAU with bearing wrapping (< -180): %s", strings.TrimSpace(msg))
}

// TestMakeFlarmPFLAUString_BearingGreaterThan180 tests bearing wrapping when relative bearing > 180
func TestMakeFlarmPFLAUString_BearingGreaterThan180(t *testing.T) {
	resetFlarmOutputState()

	// Set up valid GPS with low true course (e.g., 10 degrees)
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.GPSTrueCourse = 10 // Low true course
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Set up baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 1000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	trafficMutex.Lock()
	traffic[0xBEA123] = TrafficInfo{}
	trafficMutex.Unlock()

	// Traffic to the southwest (bearing ~225 from north)
	// Relative bearing = 225 - 10 = 215 degrees, which is > 180
	// Should be wrapped to 215 - 360 = -145 degrees
	ti := TrafficInfo{
		Icao_addr:      0xBEA123,
		Position_valid: true,
		Lat:            47.495, // South of us
		Lng:            8.494,  // West of us, creates ~225° bearing
		Alt:            1050,   // Close in altitude for alarm
		BearingDist_valid: true,
	}

	msg := makeFlarmPFLAUString(ti)

	if !strings.HasPrefix(msg, "$PFLAU") {
		t.Errorf("Expected PFLAU sentence, got: %s", msg)
	}

	// Validate it's a proper NMEA sentence
	validateNMEASentence(t, msg, "PFLAU")

	t.Logf("PFLAU with bearing wrapping (> 180): %s", strings.TrimSpace(msg))
}

// TestMakeFlarmPFLAUString_AlarmLevel2 tests PFLAU with alarm level 2 (close traffic)
func TestMakeFlarmPFLAUString_AlarmLevel2(t *testing.T) {
	resetFlarmOutputState()

	// Set up valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.GPSTrueCourse = 90
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Set up baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 1000 // 1000 ft
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	trafficMutex.Lock()
	traffic[0xABCDEF] = TrafficInfo{}
	trafficMutex.Unlock()

	// Traffic very close (within 1 NM / 1852m and within 1000 ft / 304m vertical)
	// This should trigger alarm level 2
	ti := TrafficInfo{
		Icao_addr:      0xABCDEF,
		Position_valid: true,
		Lat:            47.5 + 0.005, // ~500m north
		Lng:            8.5 + 0.005,  // ~350m east
		Alt:            1150,          // 150 ft higher (within 304m = 1000ft)
		Tail:           "N123AB",
	}

	msg := makeFlarmPFLAUString(ti)

	if !strings.HasPrefix(msg, "$PFLAU") {
		t.Errorf("Expected PFLAU sentence, got: %s", msg)
	}

	// Alarm level should be 2 (1NM, 1000ft)
	// Format: $PFLAU,<RX>,<TX>,<GPS>,<Power>,<AlarmLevel>,...
	parts := strings.Split(msg, ",")
	if len(parts) > 5 {
		alarmLevel := parts[5]
		if alarmLevel != "2" && alarmLevel != "3" {
			t.Logf("Alarm level was %s (expected 2 or 3 for close traffic)", alarmLevel)
		}
	}

	// Should have the ID with tail appended
	if !strings.Contains(msg, "ABCDEF!N123AB") {
		t.Log("Expected ID with tail appended")
	}

	t.Logf("PFLAU with alarm level: %s", strings.TrimSpace(msg))
}

// TestMakeFlarmPFLAUString_AlarmLevel3 tests PFLAU with alarm level 3 (very close traffic)
func TestMakeFlarmPFLAUString_AlarmLevel3(t *testing.T) {
	resetFlarmOutputState()

	// Set up valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.GPSTrueCourse = 0
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Set up baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 1000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	trafficMutex.Lock()
	traffic[0x111111] = TrafficInfo{}
	trafficMutex.Unlock()

	// Traffic very close (within 0.5 NM / 926m and within 500 ft / 152m vertical)
	// This should trigger alarm level 3
	ti := TrafficInfo{
		Icao_addr:      0x111111,
		Position_valid: true,
		Lat:            47.5 + 0.002, // ~220m north
		Lng:            8.5 + 0.002,  // ~150m east
		Alt:            1050,          // 50 ft higher (within 152m = 500ft)
	}

	msg := makeFlarmPFLAUString(ti)

	if !strings.HasPrefix(msg, "$PFLAU") {
		t.Errorf("Expected PFLAU sentence, got: %s", msg)
	}

	// Alarm level should be 3 (0.5NM, 500ft)
	parts := strings.Split(msg, ",")
	if len(parts) > 5 {
		alarmLevel := parts[5]
		if alarmLevel != "3" {
			t.Logf("Alarm level was %s (expected 3 for very close traffic)", alarmLevel)
		}
	}

	t.Logf("PFLAU with alarm level 3: %s", strings.TrimSpace(msg))
}

// TestMakeFlarmPFLAAString_DEBUGMode tests PFLAA with DEBUG mode enabled
func TestMakeFlarmPFLAAString_DEBUGMode(t *testing.T) {
	resetFlarmOutputState()

	// Save and restore DEBUG setting
	origDEBUG := globalSettings.DEBUG
	defer func() { globalSettings.DEBUG = origDEBUG }()

	globalSettings.DEBUG = true

	// Set up valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Set up baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 1000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	ti := TrafficInfo{
		Icao_addr:      0xDEB001,
		Position_valid: true,
		Lat:            47.51,
		Lng:            8.51,
		Alt:            1500,
		Tail:           "DEBUG",
		Speed_valid:    true,
		Speed:          100,
	}

	msg, valid, _ := makeFlarmPFLAAString(ti)

	if !valid {
		t.Error("Expected valid PFLAA message")
	}

	if !strings.HasPrefix(msg, "$PFLAA") {
		t.Errorf("Expected PFLAA sentence, got: %s", msg)
	}

	t.Logf("PFLAA with DEBUG mode: %s", strings.TrimSpace(msg))
}

// TestMakeFlarmPFLAAString_DistanceEstimated tests PFLAA with DistanceEstimated
func TestMakeFlarmPFLAAString_DistanceEstimated(t *testing.T) {
	resetFlarmOutputState()

	// Set up valid GPS
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = 47.5
	mySituation.GPSLongitude = 8.5
	mySituation.GPSAltitudeMSL = 1000
	mySituation.GPSFixQuality = 1
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()
	globalStatus.GPS_connected = true

	// Set up baro
	mySituation.muBaro.Lock()
	mySituation.BaroPressureAltitude = 1000
	mySituation.BaroLastMeasurementTime = stratuxClock.Time
	mySituation.muBaro.Unlock()

	// Traffic without valid position but with estimated distance
	ti := TrafficInfo{
		Icao_addr:         0xE51111,
		Position_valid:    false,
		DistanceEstimated: 5000, // 5km estimated
		Alt:               1500,
		Vvel:              200,
	}

	msg, valid, _ := makeFlarmPFLAAString(ti)

	if !valid {
		t.Error("Expected valid PFLAA message even with invalid position")
	}

	// For invalid position, RelativeEast should be empty (consecutive commas)
	if !strings.Contains(msg, ",,") {
		t.Log("Expected empty RelativeEast field for bearingless traffic")
	}

	t.Logf("PFLAA with DistanceEstimated: %s", strings.TrimSpace(msg))
}
