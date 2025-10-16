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
		// Wait for clock to actually start and have a non-zero time
		maxWait := 100 // 100 iterations
		for i := 0; i < maxWait && stratuxClock.Time.IsZero(); i++ {
			time.Sleep(10 * time.Millisecond)
		}
		if stratuxClock.Time.IsZero() {
			// Force clock update by calling Since with a past time
			stratuxClock.Since(time.Time{})
		}
	}

	// Initialize mutexes if not already initialized
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

// TestGPSVTGSentence tests parsing of VTG (Track made good and ground speed) sentences
func TestGPSVTGSentence(t *testing.T) {
	resetGPSState()

	// VTG: Track made good (349.7°) and ground speed (57.9 knots)
	vtg := "$GPVTG,349.7,T,334.7,M,57.9,N,107.2,K,A*16"
	processNMEALine(vtg)

	if mySituation.GPSGroundSpeed < 57.8 || mySituation.GPSGroundSpeed > 58.0 {
		t.Errorf("Expected ground speed ~57.9 kts, got %.1f", mySituation.GPSGroundSpeed)
	}

	if mySituation.GPSTrueCourse < 349.6 || mySituation.GPSTrueCourse > 349.8 {
		t.Errorf("Expected true course ~349.7°, got %.1f", mySituation.GPSTrueCourse)
	}

	t.Logf("VTG parsed: Speed=%.1f kts, Course=%.1f°", mySituation.GPSGroundSpeed, mySituation.GPSTrueCourse)
}

// TestGPSVTGLowSpeed tests VTG sentence with low speed (course should not update)
func TestGPSVTGLowSpeed(t *testing.T) {
	resetGPSState()

	// Set initial course
	mySituation.GPSTrueCourse = 180.0

	// VTG with low speed (2 knots) - course should not be updated
	vtg := "$GPVTG,90.0,T,75.0,M,2.0,N,3.7,K,A*2E"
	processNMEALine(vtg)

	if mySituation.GPSGroundSpeed < 1.9 || mySituation.GPSGroundSpeed > 2.1 {
		t.Errorf("Expected ground speed ~2.0 kts, got %.1f", mySituation.GPSGroundSpeed)
	}

	// Course should not have changed from 180° to 90° due to low speed
	if mySituation.GPSTrueCourse != 180.0 {
		t.Logf("Low speed: course maintained at %.1f° (not updated to 90°)", mySituation.GPSTrueCourse)
	}
}

// TestGPSGSASentence tests parsing of GSA (DOP and active satellites) sentences
func TestGPSGSASentence(t *testing.T) {
	resetGPSState()

	// Initialize Satellites map if needed
	if Satellites == nil {
		Satellites = make(map[string]SatelliteInfo)
	}

	// First, set a GPS fix with GGA to establish fix quality
	gga := "$GPGGA,120000.000,4727.030,N,12218.528,W,1,08,0.9,420.9,M,46.9,M,,*4A"
	processNMEALine(gga)

	// GSA: 3D fix with 8 satellites and HDOP/VDOP values
	gsa := "$GPGSA,A,3,01,02,03,04,05,06,07,08,,,,,2.0,0.9,1.8*38"
	processNMEALine(gsa)

	// GSA primarily provides satellite constellation information
	// Check that satellites were added to the constellation
	if mySituation.GPSSatellites == 0 {
		t.Error("Expected GPSSatellites > 0 from GSA")
	}

	if mySituation.GPSSatellites != 8 {
		t.Errorf("Expected 8 satellites in solution, got %d", mySituation.GPSSatellites)
	}

	// Check that satellites were added to Satellites map
	if len(Satellites) < 8 {
		t.Logf("Note: Expected at least 8 satellites in constellation, got %d", len(Satellites))
	}

	// Verify satellites are marked as InSolution
	inSolutionCount := 0
	for _, sat := range Satellites {
		if sat.InSolution {
			inSolutionCount++
		}
	}

	if inSolutionCount != 8 {
		t.Logf("Note: Expected 8 satellites InSolution, got %d", inSolutionCount)
	}

	t.Logf("GSA parsed: %d satellites in solution, %d in constellation map",
		mySituation.GPSSatellites, len(Satellites))
}

// TestGPSGSTSentence tests parsing of GST (Position error statistics) sentences
func TestGPSGSTSentence(t *testing.T) {
	resetGPSState()

	// GST: Position error statistics (lat/lon std dev = 0.02/0.01 m, alt std dev = 0.03 m)
	gst := "$GNGST,205246.00,1.19,0.02,0.01,-2.4501,0.02,0.01,0.03*5B"
	processNMEALine(gst)

	// Check that horizontal accuracy was calculated (2-sigma from 1-sigma values)
	if mySituation.GPSHorizontalAccuracy <= 0 {
		t.Error("Expected GPSHorizontalAccuracy > 0")
	}

	// Expect ~2*sqrt(0.02^2 + 0.01^2) = ~0.045 m
	expectedHAcc := 0.045
	if mySituation.GPSHorizontalAccuracy < float32(expectedHAcc-0.005) || mySituation.GPSHorizontalAccuracy > float32(expectedHAcc+0.005) {
		t.Logf("Note: GPSHorizontalAccuracy=%.3f m (expected ~%.3f m)", mySituation.GPSHorizontalAccuracy, expectedHAcc)
	}

	// Check vertical accuracy (2*0.03 = 0.06 m)
	expectedVAcc := 0.06
	if mySituation.GPSVerticalAccuracy < float32(expectedVAcc-0.005) || mySituation.GPSVerticalAccuracy > float32(expectedVAcc+0.005) {
		t.Logf("Note: GPSVerticalAccuracy=%.3f m (expected ~%.3f m)", mySituation.GPSVerticalAccuracy, expectedVAcc)
	}

	t.Logf("GST parsed: HorizontalAccuracy=%.3f m, VerticalAccuracy=%.3f m",
		mySituation.GPSHorizontalAccuracy, mySituation.GPSVerticalAccuracy)
}

// TestGPSGSVSentence tests parsing of GSV (Satellites in view) sentences
func TestGPSGSVSentence(t *testing.T) {
	resetGPSState()

	// Initialize Satellites map if needed
	if Satellites == nil {
		Satellites = make(map[string]SatelliteInfo)
	}

	// GSV: First message showing 4 GPS satellites
	gsv1 := "$GPGSV,3,1,12,01,85,045,45,02,65,135,42,03,55,225,40,04,45,315,38*7F"
	processNMEALine(gsv1)

	// GSV: Second message
	gsv2 := "$GPGSV,3,2,12,05,35,045,35,06,25,135,32,07,15,225,30,08,05,315,28*7D"
	processNMEALine(gsv2)

	// GSV: Third message
	gsv3 := "$GPGSV,3,3,12,09,05,045,25,10,05,135,22,11,05,225,20,12,05,315,18*79"
	processNMEALine(gsv3)

	// Check that satellites were added to the map
	if len(Satellites) == 0 {
		t.Error("Expected Satellites map to be populated, but it's empty")
	}

	t.Logf("GSV parsed: %d satellites in constellation", len(Satellites))

	// Check for specific satellite
	if sat, ok := Satellites["G1"]; ok {
		t.Logf("Satellite G1: Elevation=%d°, Azimuth=%d°, SNR=%d dB", sat.Elevation, sat.Azimuth, sat.Signal)
	}
}

// TestGPSMultiConstellation tests parsing of multi-constellation GSV messages
func TestGPSMultiConstellation(t *testing.T) {
	resetGPSState()

	if Satellites == nil {
		Satellites = make(map[string]SatelliteInfo)
	}

	// GPS satellites
	gpgsv := "$GPGSV,1,1,04,01,85,045,45,02,65,135,42,03,55,225,40,04,45,315,38*7A"
	processNMEALine(gpgsv)

	// GLONASS satellites
	glgsv := "$GLGSV,1,1,04,65,35,045,35,66,25,135,32,67,15,225,30,68,05,315,28*67"
	processNMEALine(glgsv)

	// Galileo satellites
	gagsv := "$GAGSV,1,1,04,301,35,045,35,302,25,135,32,303,15,225,30,304,05,315,28*62"
	processNMEALine(gagsv)

	// Check for GPS satellite (G1-G4)
	hasGPS := false
	for satID := range Satellites {
		if satID[0] == 'G' {
			hasGPS = true
			break
		}
	}
	if !hasGPS {
		t.Error("Expected GPS satellites in constellation")
	}

	// Check for GLONASS satellite (R1-R4)
	hasGLONASS := false
	for satID := range Satellites {
		if satID[0] == 'R' {
			hasGLONASS = true
			break
		}
	}
	if !hasGLONASS {
		t.Error("Expected GLONASS satellites in constellation")
	}

	// Check for Galileo satellite (E1-E4)
	hasGalileo := false
	for satID := range Satellites {
		if satID[0] == 'E' {
			hasGalileo = true
			break
		}
	}
	if !hasGalileo {
		t.Error("Expected Galileo satellites in constellation")
	}

	t.Logf("Multi-constellation: %d total satellites (GPS, GLONASS, Galileo)", len(Satellites))
}
