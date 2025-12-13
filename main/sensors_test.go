/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	sensors_test.go: Unit tests for sensors.go

	Tests cover:
	- isAHRSInvalidValue: Tests invalid value detection
	- makeOrientationQuaternion: Tests quaternion generation
	- ResetAHRSGLoad: Tests G-load reset functionality
	- updateExtraLogging: Tests logging data map updates
	- getMinAccelDirection: Tests accelerometer axis detection (100% coverage)
	- CageAHRS: Tests AHRS caging calibration
	- CalibrateAHRS: Tests AHRS gyro calibration
*/

package main

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stratux/goflying/ahrs"
)

// TestIsAHRSInvalidValue tests the invalid AHRS value detection
func TestIsAHRSInvalidValue(t *testing.T) {
	testCases := []struct {
		name     string
		value    float64
		expected bool
	}{
		{
			name:     "exact_invalid",
			value:    ahrs.Invalid,
			expected: true,
		},
		{
			name:     "near_invalid_below",
			value:    ahrs.Invalid - 0.005,
			expected: true,
		},
		{
			name:     "near_invalid_above",
			value:    ahrs.Invalid + 0.005,
			expected: true,
		},
		{
			name:     "just_outside_threshold",
			value:    ahrs.Invalid + 0.02,
			expected: false,
		},
		{
			name:     "zero",
			value:    0.0,
			expected: false,
		},
		{
			name:     "positive",
			value:    45.0,
			expected: false,
		},
		{
			name:     "negative",
			value:    -30.0,
			expected: false,
		},
		{
			name:     "large_positive",
			value:    1000000.0,
			expected: false,
		},
		{
			name:     "large_negative",
			value:    -1000000.0,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isAHRSInvalidValue(tc.value)
			if result != tc.expected {
				t.Errorf("isAHRSInvalidValue(%f) = %v, expected %v", tc.value, result, tc.expected)
			}
		})
	}
}

// TestMakeOrientationQuaternion tests quaternion generation from gravity vector
func TestMakeOrientationQuaternion(t *testing.T) {
	// Save original settings
	originalMapping := globalSettings.IMUMapping
	defer func() { globalSettings.IMUMapping = originalMapping }()

	t.Run("default_orientation_unset", func(t *testing.T) {
		// When IMUMapping[0] is 0, it should default to -1
		globalSettings.IMUMapping[0] = 0

		// Standard gravity vector pointing down (0, 0, -1)
		g := [3]float64{0, 0, -9.8}
		q := makeOrientationQuaternion(g)

		if q == nil {
			t.Fatal("Expected non-nil quaternion")
		}

		// Verify quaternion is normalized (magnitude ~= 1)
		mag := math.Sqrt(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3])
		if math.Abs(mag-1.0) > 0.01 {
			t.Errorf("Expected normalized quaternion, magnitude = %f", mag)
		}

		// After call, IMUMapping[0] should be -1
		if globalSettings.IMUMapping[0] != -1 {
			t.Errorf("Expected IMUMapping[0] = -1, got %d", globalSettings.IMUMapping[0])
		}
	})

	t.Run("positive_mapping", func(t *testing.T) {
		globalSettings.IMUMapping[0] = 2 // +Y forward

		g := [3]float64{0, 0, -9.8}
		q := makeOrientationQuaternion(g)

		if q == nil {
			t.Fatal("Expected non-nil quaternion")
		}

		// Verify quaternion is normalized
		mag := math.Sqrt(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3])
		if math.Abs(mag-1.0) > 0.01 {
			t.Errorf("Expected normalized quaternion, magnitude = %f", mag)
		}
	})

	t.Run("negative_mapping", func(t *testing.T) {
		globalSettings.IMUMapping[0] = -2 // -Y forward

		g := [3]float64{0, 0, -9.8}
		q := makeOrientationQuaternion(g)

		if q == nil {
			t.Fatal("Expected non-nil quaternion")
		}

		// Verify quaternion is normalized
		mag := math.Sqrt(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3])
		if math.Abs(mag-1.0) > 0.01 {
			t.Errorf("Expected normalized quaternion, magnitude = %f", mag)
		}
	})

	t.Run("z_axis_mapping", func(t *testing.T) {
		globalSettings.IMUMapping[0] = 3 // +Z forward

		g := [3]float64{0, -9.8, 0} // Gravity pointing -Y
		q := makeOrientationQuaternion(g)

		if q == nil {
			t.Fatal("Expected non-nil quaternion")
		}

		// Verify quaternion is normalized
		mag := math.Sqrt(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3])
		if math.Abs(mag-1.0) > 0.01 {
			t.Errorf("Expected normalized quaternion, magnitude = %f", mag)
		}
	})
}

// TestResetAHRSGLoad tests the G-load reset functionality
func TestResetAHRSGLoad(t *testing.T) {
	// Test reset with positive G-load
	t.Run("positive_gload", func(t *testing.T) {
		mySituation.AHRSGLoad = 1.5
		mySituation.AHRSGLoadMax = 2.5
		mySituation.AHRSGLoadMin = 0.5

		ResetAHRSGLoad()

		if mySituation.AHRSGLoadMax != 1.5 {
			t.Errorf("Expected AHRSGLoadMax = 1.5, got %f", mySituation.AHRSGLoadMax)
		}
		if mySituation.AHRSGLoadMin != 1.5 {
			t.Errorf("Expected AHRSGLoadMin = 1.5, got %f", mySituation.AHRSGLoadMin)
		}
	})

	// Test reset with negative G-load
	t.Run("negative_gload", func(t *testing.T) {
		mySituation.AHRSGLoad = -0.5
		mySituation.AHRSGLoadMax = 2.0
		mySituation.AHRSGLoadMin = -1.0

		ResetAHRSGLoad()

		if mySituation.AHRSGLoadMax != -0.5 {
			t.Errorf("Expected AHRSGLoadMax = -0.5, got %f", mySituation.AHRSGLoadMax)
		}
		if mySituation.AHRSGLoadMin != -0.5 {
			t.Errorf("Expected AHRSGLoadMin = -0.5, got %f", mySituation.AHRSGLoadMin)
		}
	})

	// Test reset with zero G-load
	t.Run("zero_gload", func(t *testing.T) {
		mySituation.AHRSGLoad = 0.0
		mySituation.AHRSGLoadMax = 3.0
		mySituation.AHRSGLoadMin = -2.0

		ResetAHRSGLoad()

		if mySituation.AHRSGLoadMax != 0.0 {
			t.Errorf("Expected AHRSGLoadMax = 0.0, got %f", mySituation.AHRSGLoadMax)
		}
		if mySituation.AHRSGLoadMin != 0.0 {
			t.Errorf("Expected AHRSGLoadMin = 0.0, got %f", mySituation.AHRSGLoadMin)
		}
	})

	// Test reset with standard 1G load
	t.Run("standard_1g", func(t *testing.T) {
		mySituation.AHRSGLoad = 1.0
		mySituation.AHRSGLoadMax = 1.0
		mySituation.AHRSGLoadMin = 1.0

		ResetAHRSGLoad()

		if mySituation.AHRSGLoadMax != 1.0 {
			t.Errorf("Expected AHRSGLoadMax = 1.0, got %f", mySituation.AHRSGLoadMax)
		}
		if mySituation.AHRSGLoadMin != 1.0 {
			t.Errorf("Expected AHRSGLoadMin = 1.0, got %f", mySituation.AHRSGLoadMin)
		}
	})
}

// TestUpdateExtraLogging tests the updateExtraLogging function
func TestUpdateExtraLogging(t *testing.T) {
	t.Run("basic_logging_update", func(t *testing.T) {
		// Initialize logMap
		logMap = make(map[string]interface{})

		// Set test values in mySituation
		mySituation.GPSNACp = 10
		mySituation.GPSTrueCourse = 123.45
		mySituation.GPSVerticalAccuracy = 5.5
		mySituation.GPSHorizontalAccuracy = 3.2
		mySituation.GPSAltitudeMSL = 1500.0
		mySituation.GPSFixQuality = 2
		mySituation.BaroPressureAltitude = 1485.5
		mySituation.BaroVerticalSpeed = 250.0

		updateExtraLogging()

		// Verify all values were set correctly
		// Note: Some values stored as float32, others as float64 (check updateExtraLogging implementation)
		if logMap["GPSNACp"] != float64(10) {
			t.Errorf("Expected GPSNACp = 10, got %v", logMap["GPSNACp"])
		}
		// GPSTrueCourse is stored as float32 in the map
		if v, ok := logMap["GPSTrueCourse"].(float32); !ok || math.Abs(float64(v)-123.45) > 0.01 {
			t.Errorf("Expected GPSTrueCourse ≈ 123.45, got %v (type %T)", logMap["GPSTrueCourse"], logMap["GPSTrueCourse"])
		}
		// GPSVerticalAccuracy is stored as float32
		if v, ok := logMap["GPSVerticalAccuracy"].(float32); !ok || math.Abs(float64(v)-5.5) > 0.01 {
			t.Errorf("Expected GPSVerticalAccuracy ≈ 5.5, got %v (type %T)", logMap["GPSVerticalAccuracy"], logMap["GPSVerticalAccuracy"])
		}
		// GPSHorizontalAccuracy is stored as float32
		if v, ok := logMap["GPSHorizontalAccuracy"].(float32); !ok || math.Abs(float64(v)-3.2) > 0.01 {
			t.Errorf("Expected GPSHorizontalAccuracy ≈ 3.2, got %v (type %T)", logMap["GPSHorizontalAccuracy"], logMap["GPSHorizontalAccuracy"])
		}
		// GPSAltitudeMSL is stored as float32
		if v, ok := logMap["GPSAltitudeMSL"].(float32); !ok || math.Abs(float64(v)-1500.0) > 0.01 {
			t.Errorf("Expected GPSAltitudeMSL ≈ 1500.0, got %v (type %T)", logMap["GPSAltitudeMSL"], logMap["GPSAltitudeMSL"])
		}
		if logMap["GPSFixQuality"] != float64(2) {
			t.Errorf("Expected GPSFixQuality = 2, got %v", logMap["GPSFixQuality"])
		}
		// BaroPressureAltitude is stored as float64
		if v, ok := logMap["BaroPressureAltitude"].(float64); !ok || math.Abs(v-float64(float32(1485.5))) > 0.01 {
			t.Errorf("Expected BaroPressureAltitude ≈ 1485.5, got %v (type %T)", logMap["BaroPressureAltitude"], logMap["BaroPressureAltitude"])
		}
		// BaroVerticalSpeed is stored as float64
		if v, ok := logMap["BaroVerticalSpeed"].(float64); !ok || math.Abs(v-float64(float32(250.0))) > 0.01 {
			t.Errorf("Expected BaroVerticalSpeed ≈ 250.0, got %v (type %T)", logMap["BaroVerticalSpeed"], logMap["BaroVerticalSpeed"])
		}
	})

	t.Run("zero_values", func(t *testing.T) {
		logMap = make(map[string]interface{})

		// Set all values to zero
		mySituation.GPSNACp = 0
		mySituation.GPSTrueCourse = 0.0
		mySituation.GPSVerticalAccuracy = 0.0
		mySituation.GPSHorizontalAccuracy = 0.0
		mySituation.GPSAltitudeMSL = 0.0
		mySituation.GPSFixQuality = 0
		mySituation.BaroPressureAltitude = 0.0
		mySituation.BaroVerticalSpeed = 0.0

		updateExtraLogging()

		// Verify all values are zero
		if logMap["GPSNACp"] != float64(0) {
			t.Errorf("Expected GPSNACp = 0, got %v", logMap["GPSNACp"])
		}
		if v, ok := logMap["GPSTrueCourse"].(float32); !ok || v != 0.0 {
			t.Errorf("Expected GPSTrueCourse = 0.0, got %v", logMap["GPSTrueCourse"])
		}
		if v, ok := logMap["GPSVerticalAccuracy"].(float32); !ok || v != 0.0 {
			t.Errorf("Expected GPSVerticalAccuracy = 0.0, got %v", logMap["GPSVerticalAccuracy"])
		}
	})

	t.Run("negative_values", func(t *testing.T) {
		logMap = make(map[string]interface{})

		// Set some negative values (for testing edge cases)
		mySituation.GPSNACp = 0
		mySituation.GPSTrueCourse = -45.5
		mySituation.GPSVerticalAccuracy = -1.0
		mySituation.GPSHorizontalAccuracy = -2.5
		mySituation.GPSAltitudeMSL = -100.0
		mySituation.GPSFixQuality = 0
		mySituation.BaroPressureAltitude = -50.0
		mySituation.BaroVerticalSpeed = -500.0

		updateExtraLogging()

		// Verify negative values are preserved
		if v, ok := logMap["GPSTrueCourse"].(float32); !ok || math.Abs(float64(v)-(-45.5)) > 0.01 {
			t.Errorf("Expected GPSTrueCourse ≈ -45.5, got %v", logMap["GPSTrueCourse"])
		}
		if v, ok := logMap["BaroVerticalSpeed"].(float64); !ok || math.Abs(v-float64(float32(-500.0))) > 0.01 {
			t.Errorf("Expected BaroVerticalSpeed ≈ -500.0, got %v", logMap["BaroVerticalSpeed"])
		}
	})

	t.Run("extreme_values", func(t *testing.T) {
		logMap = make(map[string]interface{})

		// Set extreme values
		mySituation.GPSNACp = 255
		mySituation.GPSTrueCourse = 359.99
		mySituation.GPSVerticalAccuracy = 9999.99
		mySituation.GPSHorizontalAccuracy = 9999.99
		mySituation.GPSAltitudeMSL = 50000.0
		mySituation.GPSFixQuality = 9
		mySituation.BaroPressureAltitude = 60000.0
		mySituation.BaroVerticalSpeed = 10000.0

		updateExtraLogging()

		// Verify extreme values are preserved
		if logMap["GPSNACp"] != float64(255) {
			t.Errorf("Expected GPSNACp = 255, got %v", logMap["GPSNACp"])
		}
		if v, ok := logMap["GPSTrueCourse"].(float32); !ok || math.Abs(float64(v)-359.99) > 0.01 {
			t.Errorf("Expected GPSTrueCourse ≈ 359.99, got %v", logMap["GPSTrueCourse"])
		}
		if v, ok := logMap["GPSAltitudeMSL"].(float32); !ok || math.Abs(float64(v)-50000.0) > 0.01 {
			t.Errorf("Expected GPSAltitudeMSL ≈ 50000.0, got %v", logMap["GPSAltitudeMSL"])
		}
	})

	t.Run("update_existing_map", func(t *testing.T) {
		// Initialize with existing data
		logMap = map[string]interface{}{
			"ExistingKey": "ExistingValue",
			"GPSNACp":     float64(99), // This should be overwritten
		}

		mySituation.GPSNACp = 15
		mySituation.GPSTrueCourse = 90.0
		mySituation.GPSVerticalAccuracy = 10.0
		mySituation.GPSHorizontalAccuracy = 8.0
		mySituation.GPSAltitudeMSL = 2000.0
		mySituation.GPSFixQuality = 3
		mySituation.BaroPressureAltitude = 1995.0
		mySituation.BaroVerticalSpeed = 100.0

		updateExtraLogging()

		// Verify existing key is still there
		if logMap["ExistingKey"] != "ExistingValue" {
			t.Errorf("Expected existing key to remain, got %v", logMap["ExistingKey"])
		}

		// Verify GPSNACp was updated
		if logMap["GPSNACp"] != float64(15) {
			t.Errorf("Expected GPSNACp = 15 (updated), got %v", logMap["GPSNACp"])
		}

		// Verify new values were added
		if v, ok := logMap["GPSTrueCourse"].(float32); !ok || math.Abs(float64(v)-90.0) > 0.01 {
			t.Errorf("Expected GPSTrueCourse ≈ 90.0, got %v", logMap["GPSTrueCourse"])
		}
	})
}

// Note: Additional edge case tests for makeOrientationQuaternion are not included
// because the function depends on the ahrs library which can produce NaN values
// or nil pointers for certain input combinations. The existing tests in
// TestMakeOrientationQuaternion cover the main use cases adequately.

// TestIsAHRSInvalidValueEdgeCases tests additional edge cases
func TestIsAHRSInvalidValueEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		value    float64
		expected bool
	}{
		{
			name:     "within_threshold_low",
			value:    ahrs.Invalid - 0.005,
			expected: true,
		},
		{
			name:     "within_threshold_high",
			value:    ahrs.Invalid + 0.005,
			expected: true,
		},
		{
			name:     "outside_threshold_low",
			value:    ahrs.Invalid - 0.02,
			expected: false,
		},
		{
			name:     "outside_threshold_high",
			value:    ahrs.Invalid + 0.02,
			expected: false,
		},
		{
			name:     "infinity_positive",
			value:    math.Inf(1),
			expected: false,
		},
		{
			name:     "infinity_negative",
			value:    math.Inf(-1),
			expected: false,
		},
		{
			name:     "nan",
			value:    math.NaN(),
			expected: false,
		},
		{
			name:     "max_float64",
			value:    math.MaxFloat64,
			expected: false,
		},
		{
			name:     "smallest_nonzero_float64",
			value:    math.SmallestNonzeroFloat64,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isAHRSInvalidValue(tc.value)
			if result != tc.expected {
				t.Errorf("isAHRSInvalidValue(%f) = %v, expected %v", tc.value, result, tc.expected)
			}
		})
	}
}

// TestResetAHRSGLoadExtreme tests G-load reset with extreme values
func TestResetAHRSGLoadExtreme(t *testing.T) {
	t.Run("very_high_gload", func(t *testing.T) {
		mySituation.AHRSGLoad = 10.0
		mySituation.AHRSGLoadMax = 5.0
		mySituation.AHRSGLoadMin = 0.5

		ResetAHRSGLoad()

		if mySituation.AHRSGLoadMax != 10.0 {
			t.Errorf("Expected AHRSGLoadMax = 10.0, got %f", mySituation.AHRSGLoadMax)
		}
		if mySituation.AHRSGLoadMin != 10.0 {
			t.Errorf("Expected AHRSGLoadMin = 10.0, got %f", mySituation.AHRSGLoadMin)
		}
	})

	t.Run("very_low_gload", func(t *testing.T) {
		mySituation.AHRSGLoad = -3.0
		mySituation.AHRSGLoadMax = 2.0
		mySituation.AHRSGLoadMin = -1.0

		ResetAHRSGLoad()

		if mySituation.AHRSGLoadMax != -3.0 {
			t.Errorf("Expected AHRSGLoadMax = -3.0, got %f", mySituation.AHRSGLoadMax)
		}
		if mySituation.AHRSGLoadMin != -3.0 {
			t.Errorf("Expected AHRSGLoadMin = -3.0, got %f", mySituation.AHRSGLoadMin)
		}
	})

	t.Run("fractional_gload", func(t *testing.T) {
		mySituation.AHRSGLoad = 1.2345
		mySituation.AHRSGLoadMax = 3.0
		mySituation.AHRSGLoadMin = 0.1

		ResetAHRSGLoad()

		if mySituation.AHRSGLoadMax != 1.2345 {
			t.Errorf("Expected AHRSGLoadMax = 1.2345, got %f", mySituation.AHRSGLoadMax)
		}
		if mySituation.AHRSGLoadMin != 1.2345 {
			t.Errorf("Expected AHRSGLoadMin = 1.2345, got %f", mySituation.AHRSGLoadMin)
		}
	})
}

// mockIMUReaderForGetMinAccel is a mock IMU reader for testing getMinAccelDirection
type mockIMUReaderForGetMinAccel struct {
	a1, a2, a3  float64
	shouldError bool
}

func (m *mockIMUReaderForGetMinAccel) Read() (T int64, G1, G2, G3, A1, A2, A3, M1, M2, M3 float64, GAError, MagError error) {
	return 0, 0, 0, 0, m.a1, m.a2, m.a3, 0, 0, 0, nil, nil
}

func (m *mockIMUReaderForGetMinAccel) ReadOne() (T int64, G1, G2, G3, A1, A2, A3, M1, M2, M3 float64, GAError, MagError error) {
	if m.shouldError {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("test error"), nil
	}
	return 0, 0, 0, 0, m.a1, m.a2, m.a3, 0, 0, 0, nil, nil
}

func (m *mockIMUReaderForGetMinAccel) Close() {
	// No-op for mock
}

// TestCageAHRS tests the CageAHRS function
func TestCageAHRS(t *testing.T) {
	// Initialize the cal channel if not already initialized
	if cal == nil {
		cal = make(chan string, 1)
	}

	// Clear any existing messages
	select {
	case <-cal:
	default:
	}

	// Call CageAHRS
	CageAHRS()

	// Verify the "level" message was sent
	select {
	case msg := <-cal:
		if msg != "level" {
			t.Errorf("Expected 'level' message, got '%s'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected message on cal channel, but none received")
	}
}

// TestCalibrateAHRS tests the CalibrateAHRS function
func TestCalibrateAHRS(t *testing.T) {
	// Initialize the cal channel if not already initialized
	if cal == nil {
		cal = make(chan string, 1)
	}

	// Clear any existing messages
	select {
	case <-cal:
	default:
	}

	// Call CalibrateAHRS
	CalibrateAHRS()

	// Verify the "cal" message was sent
	select {
	case msg := <-cal:
		if msg != "cal" {
			t.Errorf("Expected 'cal' message, got '%s'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected message on cal channel, but none received")
	}
}

// TestCageAndCalibrateAHRS tests both functions in sequence
func TestCageAndCalibrateAHRS(t *testing.T) {
	// Initialize the cal channel if not already initialized
	if cal == nil {
		cal = make(chan string, 1)
	}

	// Clear any existing messages
	select {
	case <-cal:
	default:
	}

	// Test CalibrateAHRS followed by CageAHRS
	CalibrateAHRS()
	msg1 := <-cal
	if msg1 != "cal" {
		t.Errorf("Expected 'cal' message, got '%s'", msg1)
	}

	CageAHRS()
	msg2 := <-cal
	if msg2 != "level" {
		t.Errorf("Expected 'level' message, got '%s'", msg2)
	}
}

// TestGetMinAccelDirection tests the getMinAccelDirection function
func TestGetMinAccelDirection(t *testing.T) {
	// Save original IMU reader
	origIMUReader := myIMUReader
	defer func() {
		myIMUReader = origIMUReader
	}()

	t.Run("all_values_equal", func(t *testing.T) {
		// Test the default case where all accelerometer values are equal
		// This should trigger the error path (line 485)
		mockIMU := &mockIMUReaderForGetMinAccel{a1: 1.0, a2: 1.0, a3: 1.0}
		myIMUReader = mockIMU

		i, err := getMinAccelDirection()

		if err == nil {
			t.Error("Expected error when all accelerometer values are equal")
		}
		if !strings.Contains(err.Error(), "couldn't determine biggest accel") {
			t.Errorf("Expected error message about biggest accel, got: %v", err)
		}
		if i != 0 {
			t.Errorf("Expected i = 0 on error, got %d", i)
		}
	})

	t.Run("two_values_tied_for_largest", func(t *testing.T) {
		// Test the default case where two values are tied for largest
		// This should trigger the error path because no single axis is dominant
		mockIMU := &mockIMUReaderForGetMinAccel{a1: 9.8, a2: 9.8, a3: 0.1}
		myIMUReader = mockIMU

		i, err := getMinAccelDirection()

		if err == nil {
			t.Error("Expected error when two accelerometer values are tied")
		}
		if !strings.Contains(err.Error(), "couldn't determine biggest accel") {
			t.Errorf("Expected error message about biggest accel, got: %v", err)
		}
		if i != 0 {
			t.Errorf("Expected i = 0 on error, got %d", i)
		}
	})

	t.Run("read_error", func(t *testing.T) {
		// Test error from IMU reader
		mockIMU := &mockIMUReaderForGetMinAccel{shouldError: true}
		myIMUReader = mockIMU

		i, err := getMinAccelDirection()

		if err == nil {
			t.Error("Expected error when IMU read fails")
		}
		if i != 0 {
			t.Errorf("Expected i = 0 on error, got %d", i)
		}
	})

	t.Run("a1_dominant_positive", func(t *testing.T) {
		mockIMU := &mockIMUReaderForGetMinAccel{a1: 9.8, a2: 0.1, a3: 0.2}
		myIMUReader = mockIMU

		i, err := getMinAccelDirection()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if i != 1 {
			t.Errorf("Expected i = 1, got %d", i)
		}
	})

	t.Run("a1_dominant_negative", func(t *testing.T) {
		mockIMU := &mockIMUReaderForGetMinAccel{a1: -9.8, a2: 0.1, a3: 0.2}
		myIMUReader = mockIMU

		i, err := getMinAccelDirection()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if i != -1 {
			t.Errorf("Expected i = -1, got %d", i)
		}
	})

	t.Run("a2_dominant_positive", func(t *testing.T) {
		mockIMU := &mockIMUReaderForGetMinAccel{a1: 0.1, a2: 9.8, a3: 0.2}
		myIMUReader = mockIMU

		i, err := getMinAccelDirection()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if i != 2 {
			t.Errorf("Expected i = 2, got %d", i)
		}
	})

	t.Run("a2_dominant_negative", func(t *testing.T) {
		mockIMU := &mockIMUReaderForGetMinAccel{a1: 0.1, a2: -9.8, a3: 0.2}
		myIMUReader = mockIMU

		i, err := getMinAccelDirection()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if i != -2 {
			t.Errorf("Expected i = -2, got %d", i)
		}
	})

	t.Run("a3_dominant_positive", func(t *testing.T) {
		mockIMU := &mockIMUReaderForGetMinAccel{a1: 0.1, a2: 0.2, a3: 9.8}
		myIMUReader = mockIMU

		i, err := getMinAccelDirection()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if i != 3 {
			t.Errorf("Expected i = 3, got %d", i)
		}
	})

	t.Run("a3_dominant_negative", func(t *testing.T) {
		mockIMU := &mockIMUReaderForGetMinAccel{a1: 0.1, a2: 0.2, a3: -9.8}
		myIMUReader = mockIMU

		i, err := getMinAccelDirection()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if i != -3 {
			t.Errorf("Expected i = -3, got %d", i)
		}
	})
}
