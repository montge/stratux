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
