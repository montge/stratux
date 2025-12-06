/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	sensors_test.go: Unit tests for sensors.go

	Tests cover:
	- isAHRSInvalidValue: Tests invalid value detection
	- makeOrientationQuaternion: Tests quaternion generation
	- ResetAHRSGLoad: Tests G-load reset functionality
*/

package main

import (
	"math"
	"testing"

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
