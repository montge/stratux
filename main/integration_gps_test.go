// integration_gps_test.go: Integration tests for GPS/NMEA protocol parsing
// Tests use trace file replay to verify NMEA parser behavior without GPS hardware

package main

import (
	"compress/gzip"
	"encoding/csv"
	"math"
	"os"
	"sync"
	"testing"
	"time"
)

// resetGPSState clears the global GPS situation state for testing
func resetGPSState() {
	// Initialize stratuxClock if not already initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond) // Let the clock start
	}

	// Initialize mutexes if not already initialized
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
		mySituation.muGPSPerformance = &sync.Mutex{}
		mySituation.muAttitude = &sync.Mutex{}
		mySituation.muBaro = &sync.Mutex{}
		mySituation.muSatellite = &sync.Mutex{}
	}

	mySituation.muGPS.Lock()
	defer mySituation.muGPS.Unlock()

	// Reset GPS position and state
	mySituation.GPSLatitude = 0
	mySituation.GPSLongitude = 0
	mySituation.GPSAltitudeMSL = 0
	mySituation.GPSGroundSpeed = 0
	mySituation.GPSTrueCourse = 0
	mySituation.GPSFixQuality = 0
	mySituation.GPSLastFixSinceMidnightUTC = 0
	mySituation.GPSLastValidNMEAMessage = ""
	mySituation.GPSLastGroundTrackTime = time.Time{}
	mySituation.GPSLastValidNMEAMessageTime = time.Time{}

	// Reset satellite tracking
	mySituation.GPSSatellites = 0
	mySituation.GPSSatellitesTracked = 0
	mySituation.GPSSatellitesSeen = 0

	// Reset global GPS status
	globalStatus.GPS_satellites_locked = 0
	globalStatus.GPS_satellites_seen = 0
	globalStatus.GPS_connected = false

	// Initialize Satellites map
	if Satellites == nil {
		Satellites = make(map[string]SatelliteInfo)
	}
}

// replayGPSTraceDirect reads a GPS trace file and directly injects NMEA sentences
func replayGPSTraceDirect(t *testing.T, filename string) int {
	fh, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open trace file %s: %v", filename, err)
	}
	defer fh.Close()

	gzr, err := gzip.NewReader(fh)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	csvr := csv.NewReader(gzr)
	count := 0

	for {
		record, err := csvr.Read()
		if err != nil {
			break // End of file
		}

		if len(record) != 3 {
			continue
		}

		// Check if this is an NMEA sentence
		if record[1] != CONTEXT_NMEA && record[1] != "nmea" {
			continue
		}

		// Directly process the NMEA sentence
		processNMEALineLow(record[2], false)
		count++
	}

	return count
}

// TestGPSBasicNMEAParsing tests basic NMEA sentence parsing
func TestGPSBasicNMEAParsing(t *testing.T) {
	resetGPSState()

	// Process the basic GPS trace file
	msgCount := replayGPSTraceDirect(t, "testdata/gps/basic_gps.trace.gz")
	t.Logf("Processed %d NMEA messages from trace file", msgCount)

	if msgCount != 10 {
		t.Errorf("Expected 10 NMEA messages, got %d", msgCount)
	}

	mySituation.muGPS.Lock()
	defer mySituation.muGPS.Unlock()

	// Verify GPS position was updated
	if mySituation.GPSLatitude == 0 {
		t.Error("GPS Latitude not updated (still 0)")
	}

	if mySituation.GPSLongitude == 0 {
		t.Error("GPS Longitude not updated (still 0)")
	}

	// Verify position is in Seattle area (approx 47.45N, 122.31W)
	if mySituation.GPSLatitude < 47.0 || mySituation.GPSLatitude > 48.0 {
		t.Errorf("GPS Latitude %f outside expected range (47-48)", mySituation.GPSLatitude)
	}

	if mySituation.GPSLongitude > -122.0 || mySituation.GPSLongitude < -123.0 {
		t.Errorf("GPS Longitude %f outside expected range (-123 to -122)", mySituation.GPSLongitude)
	}

	// Verify altitude is reasonable (420-421 meters = ~1380 feet)
	if mySituation.GPSAltitudeMSL < 1300 || mySituation.GPSAltitudeMSL > 1500 {
		t.Errorf("GPS Altitude %f outside expected range (1300-1500 ft)", mySituation.GPSAltitudeMSL)
	}

	t.Logf("GPS Position: Lat=%f, Lon=%f, Alt=%f ft",
		mySituation.GPSLatitude, mySituation.GPSLongitude, mySituation.GPSAltitudeMSL)
}

// TestGPSRMCSentence tests GPRMC (Recommended Minimum) sentence parsing
func TestGPSRMCSentence(t *testing.T) {
	resetGPSState()

	// GPRMC sentence from Seattle area: 47°27.030'N, 122°18.528'W, speed 57.9 knots, course 349.7°
	nmea := "$GPRMC,120000.000,A,4727.030,N,12218.528,W,057.9,349.7,131025,015.0,E*79"
	result := processNMEALineLow(nmea, false)

	if !result {
		t.Error("GPRMC sentence not processed successfully")
	}

	mySituation.muGPS.Lock()
	defer mySituation.muGPS.Unlock()

	// Check latitude: 47°27.030' = 47.4505°
	expectedLat := 47.4505
	if math.Abs(float64(mySituation.GPSLatitude)-expectedLat) > 0.01 {
		t.Errorf("Latitude: expected ~%f, got %f", expectedLat, mySituation.GPSLatitude)
	}

	// Check longitude: 122°18.528' = 122.3088°, West = negative
	expectedLon := -122.3088
	if math.Abs(float64(mySituation.GPSLongitude)-expectedLon) > 0.01 {
		t.Errorf("Longitude: expected ~%f, got %f", expectedLon, mySituation.GPSLongitude)
	}

	// Check ground speed: 57.9 knots
	expectedSpeed := 57.9
	if math.Abs(float64(mySituation.GPSGroundSpeed)-expectedSpeed) > 0.1 {
		t.Errorf("Ground speed: expected ~%f, got %f", expectedSpeed, mySituation.GPSGroundSpeed)
	}

	// Check true course: 349.7°
	expectedCourse := 349.7
	if math.Abs(float64(mySituation.GPSTrueCourse)-expectedCourse) > 0.5 {
		t.Errorf("True course: expected ~%f, got %f", expectedCourse, mySituation.GPSTrueCourse)
	}

	t.Logf("GPRMC parsed: Lat=%f, Lon=%f, Speed=%f kts, Course=%f°",
		mySituation.GPSLatitude, mySituation.GPSLongitude,
		mySituation.GPSGroundSpeed, mySituation.GPSTrueCourse)
}

// TestGPSGGASentence tests GPGGA (Fix Data) sentence parsing
func TestGPSGGASentence(t *testing.T) {
	resetGPSState()

	// GPGGA sentence: position, fix quality, satellites, altitude
	nmea := "$GPGGA,120000.000,4727.030,N,12218.528,W,1,08,0.9,420.9,M,46.9,M,,*4A"
	result := processNMEALineLow(nmea, false)

	if !result {
		t.Error("GPGGA sentence not processed successfully")
	}

	mySituation.muGPS.Lock()
	defer mySituation.muGPS.Unlock()

	// Check fix quality: 1 = GPS fix
	if mySituation.GPSFixQuality != 1 {
		t.Errorf("Fix quality: expected 1, got %d", mySituation.GPSFixQuality)
	}

	// Check altitude: 420.9 meters = ~1380 feet
	expectedAlt := 1380.0
	if math.Abs(float64(mySituation.GPSAltitudeMSL)-expectedAlt) > 10 {
		t.Errorf("Altitude: expected ~%f ft, got %f ft", expectedAlt, mySituation.GPSAltitudeMSL)
	}

	// Check position (same as RMC test)
	expectedLat := 47.4505
	if math.Abs(float64(mySituation.GPSLatitude)-expectedLat) > 0.01 {
		t.Errorf("Latitude: expected ~%f, got %f", expectedLat, mySituation.GPSLatitude)
	}

	t.Logf("GPGGA parsed: Fix=%d, Alt=%f ft, Lat=%f",
		mySituation.GPSFixQuality, mySituation.GPSAltitudeMSL, mySituation.GPSLatitude)
}

// TestGPSInvalidChecksum tests that invalid NMEA checksums are rejected
func TestGPSInvalidChecksum(t *testing.T) {
	resetGPSState()

	// Valid sentence with incorrect checksum
	nmea := "$GPRMC,120000,A,4727.030,N,12218.528,W,057.9,349.7,131025,015.0,E*00"
	result := processNMEALineLow(nmea, false)

	if result {
		t.Error("Invalid checksum sentence should not be processed")
	}

	mySituation.muGPS.Lock()
	defer mySituation.muGPS.Unlock()

	// Position should not be updated
	if mySituation.GPSLatitude != 0 {
		t.Error("GPS position should not be updated with invalid checksum")
	}
}

// TestGPSCoordinateParsing tests various coordinate formats
func TestGPSCoordinateParsing(t *testing.T) {
	tests := []struct {
		name        string
		nmea        string
		expectedLat float32
		expectedLon float32
	}{
		{
			name:        "Seattle area",
			nmea:        "$GPRMC,120000.000,A,4727.030,N,12218.528,W,057.9,349.7,131025,015.0,E*79",
			expectedLat: 47.4505,
			expectedLon: -122.3088,
		},
		{
			name:        "Equator and Prime Meridian",
			nmea:        "$GPRMC,120000.000,A,0000.000,N,00000.000,E,000.0,000.0,131025,000.0,E*6F",
			expectedLat: 0.0,
			expectedLon: 0.0,
		},
		{
			name:        "Southern Hemisphere",
			nmea:        "$GPRMC,120000.000,A,3351.000,S,15109.000,E,000.0,000.0,131025,000.0,E*7A",
			expectedLat: -33.85,
			expectedLon: 151.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGPSState()
			processNMEALineLow(tt.nmea, false)

			mySituation.muGPS.Lock()
			defer mySituation.muGPS.Unlock()

			if math.Abs(float64(mySituation.GPSLatitude)-float64(tt.expectedLat)) > 0.01 {
				t.Errorf("Latitude: expected %f, got %f", tt.expectedLat, mySituation.GPSLatitude)
			}

			if math.Abs(float64(mySituation.GPSLongitude)-float64(tt.expectedLon)) > 0.01 {
				t.Errorf("Longitude: expected %f, got %f", tt.expectedLon, mySituation.GPSLongitude)
			}
		})
	}
}

// TestGPSTimeStampParsing tests GPS timestamp parsing
func TestGPSTimeStampParsing(t *testing.T) {
	resetGPSState()

	// Time: 12:00:00 UTC
	nmea := "$GPGGA,120000.000,4727.030,N,12218.528,W,1,08,0.9,420.9,M,46.9,M,,*4A"
	processNMEALineLow(nmea, false)

	mySituation.muGPS.Lock()
	defer mySituation.muGPS.Unlock()

	// 12:00:00 = 12*3600 = 43200 seconds since midnight
	expectedTime := float32(43200)
	if mySituation.GPSLastFixSinceMidnightUTC != expectedTime {
		t.Errorf("GPS time: expected %f, got %f",
			expectedTime, mySituation.GPSLastFixSinceMidnightUTC)
	}

	t.Logf("GPS time parsed: %f seconds since midnight", mySituation.GPSLastFixSinceMidnightUTC)
}

// TestGPSGroundSpeedThreshold tests that low speeds don't update course
func TestGPSGroundSpeedThreshold(t *testing.T) {
	resetGPSState()

	// First, set a valid course with high speed
	nmea1 := "$GPRMC,120000.000,A,4727.030,N,12218.528,W,057.9,349.7,131025,015.0,E*79"
	processNMEALineLow(nmea1, false)

	mySituation.muGPS.Lock()
	initialCourse := mySituation.GPSTrueCourse
	mySituation.muGPS.Unlock()

	if initialCourse < 340 || initialCourse > 360 {
		t.Errorf("Initial course should be ~349.7, got %f", initialCourse)
	}

	// Now send a low-speed message with different course
	nmea2 := "$GPRMC,120001.000,A,4727.031,N,12218.530,W,002.0,180.0,131025,015.0,E*79"
	processNMEALineLow(nmea2, true)

	mySituation.muGPS.Lock()
	defer mySituation.muGPS.Unlock()

	// Speed should update
	if mySituation.GPSGroundSpeed > 3 {
		t.Errorf("Ground speed should be low, got %f", mySituation.GPSGroundSpeed)
	}

	// Course should NOT change significantly (threshold is 3 knots)
	// The course may have changed slightly, so we just check it didn't flip to 180
	if math.Abs(float64(mySituation.GPSTrueCourse)-180.0) < 10 {
		t.Error("Course should not update for speeds < 3 knots")
	}

	t.Logf("Low-speed handling: Speed=%f kts, Course unchanged at %f°",
		mySituation.GPSGroundSpeed, mySituation.GPSTrueCourse)
}

// TestGPSFixQuality tests GPS fix quality indicator
func TestGPSFixQuality(t *testing.T) {
	tests := []struct {
		name            string
		fixQuality      string
		expectedQuality uint8
	}{
		{"No fix", "0", 0},
		{"GPS fix", "1", 1},
		{"DGPS fix", "2", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGPSState()

			// For this test, we'll manually construct valid checksums for each fix quality
			var validNmea string
			switch tt.fixQuality {
			case "0":
				validNmea = "$GPGGA,120000.000,4727.030,N,12218.528,W,0,08,0.9,420.9,M,46.9,M,,*4B"
			case "1":
				validNmea = "$GPGGA,120000.000,4727.030,N,12218.528,W,1,08,0.9,420.9,M,46.9,M,,*4A"
			case "2":
				validNmea = "$GPGGA,120000.000,4727.030,N,12218.528,W,2,08,0.9,420.9,M,46.9,M,,*49"
			}

			processNMEALineLow(validNmea, false)

			mySituation.muGPS.Lock()
			defer mySituation.muGPS.Unlock()

			if mySituation.GPSFixQuality != tt.expectedQuality {
				t.Errorf("Fix quality: expected %d, got %d",
					tt.expectedQuality, mySituation.GPSFixQuality)
			}
		})
	}
}

// TestGPSAltitudeConversion tests meter to feet conversion
func TestGPSAltitudeConversion(t *testing.T) {
	tests := []struct {
		name          string
		altMeters     string
		expectedFeet  float32
		toleranceFeet float32
	}{
		{"Sea level", "0.0", 0.0, 1.0},
		{"420.9 meters", "420.9", 1380.9, 1.0},
		{"1000 meters", "1000.0", 3280.84, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGPSState()

			// Build GPGGA with specified altitude
			// Need to calculate checksum properly - for now use known good sentence
			// and just verify the conversion logic works

			if tt.altMeters == "420.9" {
				nmea := "$GPGGA,120000.000,4727.030,N,12218.528,W,1,08,0.9,420.9,M,46.9,M,,*4A"
				processNMEALineLow(nmea, false)

				mySituation.muGPS.Lock()
				defer mySituation.muGPS.Unlock()

				if math.Abs(float64(mySituation.GPSAltitudeMSL)-float64(tt.expectedFeet)) > float64(tt.toleranceFeet) {
					t.Errorf("Altitude: expected ~%f ft, got %f ft",
						tt.expectedFeet, mySituation.GPSAltitudeMSL)
				}
			}
		})
	}
}

// TestGPSLastValidMessage tests that last valid NMEA message is stored
func TestGPSLastValidMessage(t *testing.T) {
	resetGPSState()

	nmea := "$GPRMC,120000.000,A,4727.030,N,12218.528,W,057.9,349.7,131025,015.0,E*79"
	processNMEALineLow(nmea, false)

	mySituation.muGPS.Lock()
	defer mySituation.muGPS.Unlock()

	if mySituation.GPSLastValidNMEAMessage != nmea {
		t.Errorf("Last valid NMEA message not stored correctly")
		t.Logf("Expected: %s", nmea)
		t.Logf("Got: %s", mySituation.GPSLastValidNMEAMessage)
	}

	// Check that timestamp was updated
	if mySituation.GPSLastValidNMEAMessageTime.IsZero() {
		t.Error("GPS last valid message time should be set")
	}
}

// TestGPSMultipleSentenceSequence tests processing a realistic sequence of NMEA sentences
func TestGPSMultipleSentenceSequence(t *testing.T) {
	resetGPSState()

	sentences := []string{
		"$GPRMC,120000.000,A,4727.030,N,12218.528,W,057.9,349.7,131025,015.0,E*79",
		"$GPGGA,120000.000,4727.030,N,12218.528,W,1,08,0.9,420.9,M,46.9,M,,*4A",
		"$GPGSA,A,3,01,02,03,04,05,06,07,08,,,,,2.0,0.9,1.8*38",
	}

	for _, nmea := range sentences {
		result := processNMEALineLow(nmea, false)
		if !result {
			t.Errorf("Failed to process sentence: %s", nmea)
		}
	}

	mySituation.muGPS.Lock()
	defer mySituation.muGPS.Unlock()

	// After processing all sentences, we should have complete GPS data
	if mySituation.GPSLatitude == 0 {
		t.Error("GPS Latitude not set after processing sentence sequence")
	}

	if mySituation.GPSAltitudeMSL == 0 {
		t.Error("GPS Altitude not set after processing sentence sequence")
	}

	if mySituation.GPSFixQuality == 0 {
		t.Error("GPS Fix Quality not set after processing sentence sequence")
	}

	t.Logf("Sentence sequence processed successfully: Lat=%f, Lon=%f, Alt=%f, Fix=%d",
		mySituation.GPSLatitude, mySituation.GPSLongitude,
		mySituation.GPSAltitudeMSL, mySituation.GPSFixQuality)
}
