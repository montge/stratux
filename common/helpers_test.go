/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	helpers_test.go: Unit tests for common package
*/

package common

import (
	"math"
	"os"
	"os/user"
	"testing"
)

// TestIsRunningAsRoot tests root user detection
func TestIsRunningAsRoot(t *testing.T) {
	result := IsRunningAsRoot()

	// Get current user for verification
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	expected := usr.Username == "root" || usr.Uid == "0"

	if result != expected {
		t.Errorf("IsRunningAsRoot() = %v, want %v (user: %s, uid: %s)",
			result, expected, usr.Username, usr.Uid)
	}

	// Log the result for debugging
	if result {
		t.Log("Running as root")
	} else {
		t.Log("Not running as root")
	}

	// Also verify with environment variables
	if os.Geteuid() == 0 && !result {
		t.Error("Process has effective UID 0 but IsRunningAsRoot returned false")
	}
}

// TestLinReg tests linear regression calculation
func TestLinReg(t *testing.T) {
	testCases := []struct {
		name              string
		x                 []float64
		y                 []float64
		expectValid       bool
		expectedSlope     float64
		expectedIntercept float64
	}{
		{
			name:              "Perfect positive correlation",
			x:                 []float64{1, 2, 3, 4, 5},
			y:                 []float64{2, 4, 6, 8, 10},
			expectValid:       true,
			expectedSlope:     2.0,
			expectedIntercept: 0.0,
		},
		{
			name:              "Perfect negative correlation",
			x:                 []float64{1, 2, 3, 4, 5},
			y:                 []float64{5, 4, 3, 2, 1},
			expectValid:       true,
			expectedSlope:     -1.0,
			expectedIntercept: 6.0,
		},
		{
			name:        "Different lengths",
			x:           []float64{1, 2, 3},
			y:           []float64{1, 2},
			expectValid: false,
		},
		{
			name:        "Too few points",
			x:           []float64{1},
			y:           []float64{1},
			expectValid: false,
		},
		{
			name:        "Empty arrays",
			x:           []float64{},
			y:           []float64{},
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slope, intercept, valid := LinReg(tc.x, tc.y)

			if valid != tc.expectValid {
				t.Errorf("LinReg() valid = %v, want %v", valid, tc.expectValid)
			}

			if tc.expectValid {
				if math.Abs(slope-tc.expectedSlope) > 0.001 {
					t.Errorf("LinReg() slope = %f, want %f", slope, tc.expectedSlope)
				}
				if math.Abs(intercept-tc.expectedIntercept) > 0.001 {
					t.Errorf("LinReg() intercept = %f, want %f", intercept, tc.expectedIntercept)
				}
			}
		})
	}
}

// TestMean tests arithmetic mean calculation
func TestMean(t *testing.T) {
	testCases := []struct {
		name         string
		x            []float64
		expectValid  bool
		expectedMean float64
	}{
		{
			name:         "Positive numbers",
			x:            []float64{1, 2, 3, 4, 5},
			expectValid:  true,
			expectedMean: 3.0,
		},
		{
			name:         "Single value",
			x:            []float64{42},
			expectValid:  true,
			expectedMean: 42.0,
		},
		{
			name:        "Empty array",
			x:           []float64{},
			expectValid: false,
		},
		{
			name:         "Mixed positive and negative",
			x:            []float64{-2, -1, 0, 1, 2},
			expectValid:  true,
			expectedMean: 0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mean, valid := Mean(tc.x)

			if valid != tc.expectValid {
				t.Errorf("Mean() valid = %v, want %v", valid, tc.expectValid)
			}

			if tc.expectValid {
				if math.Abs(mean-tc.expectedMean) > 0.001 {
					t.Errorf("Mean() = %f, want %f", mean, tc.expectedMean)
				}
			}
		})
	}
}

// TestRadians tests degree to radian conversion
func TestRadians(t *testing.T) {
	testCases := []struct {
		degrees float64
		radians float64
	}{
		{0, 0},
		{90, math.Pi / 2},
		{180, math.Pi},
		{270, 3 * math.Pi / 2},
		{360, 2 * math.Pi},
		{-90, -math.Pi / 2},
	}

	for _, tc := range testCases {
		result := Radians(tc.degrees)
		if math.Abs(result-tc.radians) > 0.0001 {
			t.Errorf("Radians(%f) = %f, want %f", tc.degrees, result, tc.radians)
		}
	}
}

// TestDegrees tests radian to degree conversion
func TestDegrees(t *testing.T) {
	testCases := []struct {
		radians float64
		degrees float64
	}{
		{0, 0},
		{math.Pi / 2, 90},
		{math.Pi, 180},
		{3 * math.Pi / 2, 270},
		{2 * math.Pi, 360},
		{-math.Pi / 2, -90},
	}

	for _, tc := range testCases {
		result := Degrees(tc.radians)
		if math.Abs(result-tc.degrees) > 0.0001 {
			t.Errorf("Degrees(%f) = %f, want %f", tc.radians, result, tc.degrees)
		}
	}
}

// TestDistRect tests rectangular distance calculation
func TestDistRect(t *testing.T) {
	// Test from Oshkosh to nearby point
	lat1, lon1 := 43.99, -88.56
	lat2, lon2 := 44.0, -88.55

	dist, bearing, distN, distE := DistRect(lat1, lon1, lat2, lon2)

	// Should have some distance
	if dist <= 0 {
		t.Errorf("DistRect() dist = %f, want > 0", dist)
	}

	// North component should be positive (moving north)
	if distN <= 0 {
		t.Errorf("DistRect() distN = %f, want > 0", distN)
	}

	// East component should be positive (moving east/less negative longitude)
	if distE <= 0 {
		t.Errorf("DistRect() distE = %f, want > 0", distE)
	}

	// Bearing should be in range [0, 360)
	if bearing < 0 || bearing >= 360 {
		t.Errorf("DistRect() bearing = %f, want [0, 360)", bearing)
	}

	t.Logf("Distance: %.1f m, Bearing: %.1f°, North: %.1f m, East: %.1f m",
		dist, bearing, distN, distE)
}

// TestDistance tests polar distance calculation
func TestDistance(t *testing.T) {
	// Test known distance: Oshkosh to Chicago (approx 120 nm)
	oshLat, oshLon := 43.99, -88.56
	chiLat, chiLon := 41.98, -87.90

	dist, bearing := Distance(oshLat, oshLon, chiLat, chiLon)

	// Should be approximately 222 km (120 nm)
	expectedDist := 222000.0                 // meters
	if math.Abs(dist-expectedDist) > 10000 { // Allow 10km error
		t.Errorf("Distance() = %f m, want ~%f m", dist, expectedDist)
	}

	// Bearing should be roughly southeast (90-180°)
	if bearing < 90 || bearing > 180 {
		t.Logf("Warning: Distance() bearing = %f°, expected southeast (90-180°)", bearing)
	}

	t.Logf("Distance Oshkosh->Chicago: %.1f km (%.1f nm), Bearing: %.1f°",
		dist/1000, dist/1852, bearing)
}

// TestCalcAltitude tests pressure altitude calculation
func TestCalcAltitude(t *testing.T) {
	testCases := []struct {
		name        string
		pressure    float64
		altOffset   int
		expectedAlt float64
		tolerance   float64
	}{
		{
			name:        "Sea level standard pressure",
			pressure:    1013.25,
			altOffset:   0,
			expectedAlt: 0,
			tolerance:   1.0,
		},
		{
			name:        "5000 ft pressure altitude",
			pressure:    843.08,
			altOffset:   0,
			expectedAlt: 5000,
			tolerance:   50,
		},
		{
			name:        "With altitude offset",
			pressure:    1013.25,
			altOffset:   100,
			expectedAlt: 100,
			tolerance:   1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalcAltitude(tc.pressure, tc.altOffset)

			if math.Abs(result-tc.expectedAlt) > tc.tolerance {
				t.Errorf("CalcAltitude(%f, %d) = %f, want %f ±%f",
					tc.pressure, tc.altOffset, result, tc.expectedAlt, tc.tolerance)
			}
		})
	}
}

// TestArrayMinMax tests array min/max functions
func TestArrayMinMax(t *testing.T) {
	testData := []float64{3.5, 1.2, 5.7, 2.1, 4.9}

	min, validMin := ArrayMin(testData)
	max, validMax := ArrayMax(testData)

	if !validMin || !validMax {
		t.Fatal("Expected valid min/max for non-empty array")
	}

	if min != 1.2 {
		t.Errorf("ArrayMin() = %f, want 1.2", min)
	}

	if max != 5.7 {
		t.Errorf("ArrayMax() = %f, want 5.7", max)
	}

	// Test empty array
	_, validEmpty := ArrayMin([]float64{})
	if validEmpty {
		t.Error("Expected invalid result for empty array")
	}
}

// TestIsCPUTempValid tests CPU temperature validation
func TestIsCPUTempValid(t *testing.T) {
	testCases := []struct {
		temp    float32
		isValid bool
	}{
		{45.5, true},
		{0.1, true},
		{0.0, false},
		{-1.0, false},
		{InvalidCpuTemp, false},
	}

	for _, tc := range testCases {
		result := IsCPUTempValid(tc.temp)
		if result != tc.isValid {
			t.Errorf("IsCPUTempValid(%f) = %v, want %v", tc.temp, result, tc.isValid)
		}
	}
}

// TestIMinMax tests integer min/max functions
func TestIMinMax(t *testing.T) {
	if IMin(5, 3) != 3 {
		t.Error("IMin(5, 3) should be 3")
	}

	if IMin(3, 5) != 3 {
		t.Error("IMin(3, 5) should be 3")
	}

	if IMax(5, 3) != 5 {
		t.Error("IMax(5, 3) should be 5")
	}

	if IMax(3, 5) != 5 {
		t.Error("IMax(3, 5) should be 5")
	}
}

// TestRoundToInt16 tests float64 to int16 rounding
func TestRoundToInt16(t *testing.T) {
	testCases := []struct {
		input    float64
		expected int16
	}{
		{0.0, 0},
		{0.4, 0},
		{0.5, 1},
		{1.5, 2},
		{-0.4, 0},
		{-0.5, -1},
		{-1.5, -2},
		{1000.7, 1001},
	}

	for _, tc := range testCases {
		result := RoundToInt16(tc.input)
		if result != tc.expected {
			t.Errorf("RoundToInt16(%f) = %d, want %d", tc.input, result, tc.expected)
		}
	}
}

// TestLinRegWeighted tests weighted linear regression
func TestLinRegWeighted(t *testing.T) {
	testCases := []struct {
		name              string
		x                 []float64
		y                 []float64
		w                 []float64
		expectValid       bool
		expectedSlope     float64
		expectedIntercept float64
	}{
		{
			name:              "Equal weights (should match unweighted)",
			x:                 []float64{1, 2, 3, 4, 5},
			y:                 []float64{2, 4, 6, 8, 10},
			w:                 []float64{1, 1, 1, 1, 1},
			expectValid:       true,
			expectedSlope:     2.0,
			expectedIntercept: 0.0,
		},
		{
			name:        "Different array lengths",
			x:           []float64{1, 2, 3},
			y:           []float64{1, 2},
			w:           []float64{1, 1, 1},
			expectValid: false,
		},
		{
			name:        "Weights sum to zero",
			x:           []float64{1, 2, 3},
			y:           []float64{1, 2, 3},
			w:           []float64{1, -1, 0},
			expectValid: false,
		},
		{
			name:        "Too few points",
			x:           []float64{1},
			y:           []float64{1},
			w:           []float64{1},
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slope, intercept, valid := LinRegWeighted(tc.x, tc.y, tc.w)

			if valid != tc.expectValid {
				t.Errorf("LinRegWeighted() valid = %v, want %v", valid, tc.expectValid)
			}

			if tc.expectValid {
				if math.Abs(slope-tc.expectedSlope) > 0.001 {
					t.Errorf("LinRegWeighted() slope = %f, want %f", slope, tc.expectedSlope)
				}
				if math.Abs(intercept-tc.expectedIntercept) > 0.001 {
					t.Errorf("LinRegWeighted() intercept = %f, want %f", intercept, tc.expectedIntercept)
				}
			}
		})
	}
}

// TestTriCubeWeight tests tricube weight function
func TestTriCubeWeight(t *testing.T) {
	testCases := []struct {
		center    float64
		halfwidth float64
		x         float64
		expected  float64
	}{
		{0, 1, 0, 1.0},         // At center, weight = 1
		{0, 1, 1, 0.0},         // At edge, weight = 0
		{0, 1, 2, 0.0},         // Beyond halfwidth, weight = 0
		{5, 2, 5, 1.0},         // Different center
		{5, 2, 6, 0.669922},    // x_t=0.5, weight=(1-0.5^3)^3 = 0.875^3
		{0, 1, 0.5, 0.669922},  // x_t=0.5, same formula
		{0, 1, -0.5, 0.669922}, // Symmetric (negative side)
	}

	for _, tc := range testCases {
		result := TriCubeWeight(tc.center, tc.halfwidth, tc.x)
		if math.Abs(result-tc.expected) > 0.001 {
			t.Errorf("TriCubeWeight(%f, %f, %f) = %f, want %f",
				tc.center, tc.halfwidth, tc.x, result, tc.expected)
		}
	}
}

// TestArrayRange tests array range calculation
func TestArrayRange(t *testing.T) {
	testCases := []struct {
		name          string
		data          []float64
		expectedRange float64
		expectValid   bool
	}{
		{
			name:          "Positive numbers",
			data:          []float64{1, 2, 3, 4, 5},
			expectedRange: 4.0,
			expectValid:   true,
		},
		{
			name:          "Mixed values",
			data:          []float64{-5, 0, 5},
			expectedRange: 10.0,
			expectValid:   true,
		},
		{
			name:        "Empty array",
			data:        []float64{},
			expectValid: false,
		},
		{
			name:          "Single value",
			data:          []float64{42},
			expectedRange: 0.0,
			expectValid:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, valid := ArrayRange(tc.data)

			if valid != tc.expectValid {
				t.Errorf("ArrayRange() valid = %v, want %v", valid, tc.expectValid)
			}

			if tc.expectValid && math.Abs(result-tc.expectedRange) > 0.001 {
				t.Errorf("ArrayRange() = %f, want %f", result, tc.expectedRange)
			}
		})
	}
}

// TestStdev tests standard deviation calculation
func TestStdev(t *testing.T) {
	testCases := []struct {
		name        string
		data        []float64
		expectedStd float64
		expectValid bool
	}{
		{
			name:        "Constant values",
			data:        []float64{5, 5, 5, 5, 5},
			expectedStd: 0.0,
			expectValid: true,
		},
		{
			name:        "Simple sequence",
			data:        []float64{2, 4, 4, 4, 5, 5, 7, 9},
			expectedStd: 2.138,
			expectValid: true,
		},
		{
			name:        "Too few points",
			data:        []float64{5},
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, valid := Stdev(tc.data)

			if valid != tc.expectValid {
				t.Errorf("Stdev() valid = %v, want %v", valid, tc.expectValid)
			}

			if tc.expectValid && math.Abs(result-tc.expectedStd) > 0.01 {
				t.Errorf("Stdev() = %f, want %f", result, tc.expectedStd)
			}
		})
	}
}

// TestRadiansRel tests relative angle conversion with wrapping
func TestRadiansRel(t *testing.T) {
	testCases := []struct {
		degrees float64
		radians float64
	}{
		{0, 0},
		{90, math.Pi / 2},
		{180, math.Pi},
		{-180, -math.Pi},
		{270, -math.Pi / 2}, // Wraps to -90
		{360, 0},            // Wraps to 0
		{450, math.Pi / 2},  // Wraps to 90
		{-270, math.Pi / 2}, // Wraps to 90
	}

	for _, tc := range testCases {
		result := RadiansRel(tc.degrees)
		if math.Abs(result-tc.radians) > 0.0001 {
			t.Errorf("RadiansRel(%f) = %f, want %f", tc.degrees, result, tc.radians)
		}
	}
}

// TestDegreesRel tests relative angle conversion from radians
func TestDegreesRel(t *testing.T) {
	testCases := []struct {
		radians float64
		degrees float64
	}{
		{0, 0},
		{math.Pi / 2, 90},
		{math.Pi, 180},
		{-math.Pi, -180},
		{3 * math.Pi / 2, -90}, // Wraps to -90
		{2 * math.Pi, 0},       // Wraps to 0
		{5 * math.Pi / 2, 90},  // Wraps to 90
		{-3 * math.Pi / 2, 90}, // Wraps to 90
	}

	for _, tc := range testCases {
		result := DegreesRel(tc.radians)
		if math.Abs(result-tc.degrees) > 0.0001 {
			t.Errorf("DegreesRel(%f) = %f, want %f", tc.radians, result, tc.degrees)
		}
	}
}

// TestDegreesHdg tests heading angle conversion
func TestDegreesHdg(t *testing.T) {
	testCases := []struct {
		radians float64
		heading float64
	}{
		{0, 0},
		{math.Pi / 2, 90},
		{math.Pi, 180},
		{3 * math.Pi / 2, 270},
		{2 * math.Pi, 360},
		{-math.Pi / 2, 270},    // Negative wraps to 270
		{-math.Pi, 180},        // Negative wraps to 180
		{5 * math.Pi / 2, 450}, // Beyond 360 (doesn't wrap for headings)
	}

	for _, tc := range testCases {
		result := DegreesHdg(tc.radians)
		if math.Abs(result-tc.heading) > 0.0001 {
			t.Errorf("DegreesHdg(%f) = %f, want %f", tc.radians, result, tc.heading)
		}
	}
}

// TestDistRectNorth tests north-south distance calculation
func TestDistRectNorth(t *testing.T) {
	// 1 degree of latitude is approximately 111 km
	lat1, lat2 := 43.0, 44.0
	result := DistRectNorth(lat1, lat2)

	expectedDist := 111000.0 // Approximately 111 km
	if math.Abs(result-expectedDist) > 2000 {
		t.Errorf("DistRectNorth(%f, %f) = %f, want ~%f", lat1, lat2, result, expectedDist)
	}

	// Test negative direction (south)
	resultSouth := DistRectNorth(lat2, lat1)
	if resultSouth >= 0 {
		t.Errorf("DistRectNorth(%f, %f) should be negative (moving south)", lat2, lat1)
	}

	t.Logf("North distance: %.1f m", result)
}

// TestDistRectEast tests east-west distance calculation
func TestDistRectEast(t *testing.T) {
	// At 45° latitude, 1 degree longitude ≈ 78.8 km
	lat := 45.0
	lon1, lon2 := -89.0, -88.0

	result := DistRectEast(lat, lon1, lat, lon2)

	expectedDist := 78800.0 // Approximately 78.8 km at 45° latitude
	if math.Abs(result-expectedDist) > 2000 {
		t.Errorf("DistRectEast(%f, %f, %f, %f) = %f, want ~%f",
			lat, lon1, lat, lon2, result, expectedDist)
	}

	// Test negative direction (west)
	resultWest := DistRectEast(lat, lon2, lat, lon1)
	if resultWest >= 0 {
		t.Errorf("DistRectEast should be negative when moving west")
	}

	t.Logf("East distance at 45° lat: %.1f m", result)
}

// TestLinReg_EdgeCase tests linear regression edge case (infinite slope)
func TestLinReg_EdgeCase(t *testing.T) {
	// All x values the same (infinite slope)
	x := []float64{5, 5, 5, 5}
	y := []float64{1, 2, 3, 4}

	_, _, valid := LinReg(x, y)

	if valid {
		t.Error("LinReg should return invalid for infinite slope (all x values equal)")
	}
}

// TestLinRegWeighted_InfiniteSlope tests weighted regression edge case
func TestLinRegWeighted_InfiniteSlope(t *testing.T) {
	// All x values the same (infinite slope)
	x := []float64{5, 5, 5, 5}
	y := []float64{1, 2, 3, 4}
	w := []float64{1, 1, 1, 1}

	_, _, valid := LinRegWeighted(x, y, w)

	if valid {
		t.Error("LinRegWeighted should return invalid for infinite slope")
	}
}
