package main

import (
	"testing"
)

// TestMapAircraftType tests bidirectional aircraft type mapping
func TestMapAircraftType(t *testing.T) {
	// Sample mapping table (type1 -> type2)
	mapping := [][]int{
		{1, 10},   // Type 1 maps to 10
		{2, 20},   // Type 2 maps to 20
		{3, 30},   // Type 3 maps to 30
		{5, 50},   // Type 5 maps to 50
		{10, 100}, // Type 10 maps to 100
	}

	testCases := []struct {
		name     string
		mapping  [][]int
		forward  bool
		acType   int
		expected int
	}{
		// Forward mapping tests (first column -> second column)
		{
			name:     "Forward: Type 1 -> 10",
			mapping:  mapping,
			forward:  true,
			acType:   1,
			expected: 10,
		},
		{
			name:     "Forward: Type 2 -> 20",
			mapping:  mapping,
			forward:  true,
			acType:   2,
			expected: 20,
		},
		{
			name:     "Forward: Type 5 -> 50",
			mapping:  mapping,
			forward:  true,
			acType:   5,
			expected: 50,
		},
		{
			name:     "Forward: Type not found",
			mapping:  mapping,
			forward:  true,
			acType:   99,
			expected: -1,
		},
		{
			name:     "Forward: Type 0 (not in mapping)",
			mapping:  mapping,
			forward:  true,
			acType:   0,
			expected: -1,
		},

		// Reverse mapping tests (second column -> first column)
		{
			name:     "Reverse: Type 10 -> 1",
			mapping:  mapping,
			forward:  false,
			acType:   10,
			expected: 1,
		},
		{
			name:     "Reverse: Type 20 -> 2",
			mapping:  mapping,
			forward:  false,
			acType:   20,
			expected: 2,
		},
		{
			name:     "Reverse: Type 100 -> 10",
			mapping:  mapping,
			forward:  false,
			acType:   100,
			expected: 10,
		},
		{
			name:     "Reverse: Type not found",
			mapping:  mapping,
			forward:  false,
			acType:   999,
			expected: -1,
		},
		{
			name:     "Reverse: Type 0 (not in mapping)",
			mapping:  mapping,
			forward:  false,
			acType:   0,
			expected: -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapAircraftType(tc.mapping, tc.forward, tc.acType)
			if result != tc.expected {
				t.Errorf("mapAircraftType(mapping, forward=%v, type=%d) = %d, expected %d",
					tc.forward, tc.acType, result, tc.expected)
			}

			direction := "forward"
			if !tc.forward {
				direction = "reverse"
			}
			t.Logf("%s: type %d -> %d", direction, tc.acType, result)
		})
	}
}

// TestMapAircraftTypeEmptyMapping tests with empty mapping table
func TestMapAircraftTypeEmptyMapping(t *testing.T) {
	emptyMapping := [][]int{}

	testCases := []struct {
		name     string
		forward  bool
		acType   int
		expected int
	}{
		{
			name:     "Empty mapping - forward",
			forward:  true,
			acType:   1,
			expected: -1,
		},
		{
			name:     "Empty mapping - reverse",
			forward:  false,
			acType:   10,
			expected: -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapAircraftType(emptyMapping, tc.forward, tc.acType)
			if result != tc.expected {
				t.Errorf("mapAircraftType(empty, forward=%v, type=%d) = %d, expected %d",
					tc.forward, tc.acType, result, tc.expected)
			}
		})
	}
}

// TestMapAircraftTypeSingleEntry tests with single mapping entry
func TestMapAircraftTypeSingleEntry(t *testing.T) {
	singleMapping := [][]int{
		{7, 77},
	}

	testCases := []struct {
		name     string
		forward  bool
		acType   int
		expected int
	}{
		{
			name:     "Single entry - forward match",
			forward:  true,
			acType:   7,
			expected: 77,
		},
		{
			name:     "Single entry - reverse match",
			forward:  false,
			acType:   77,
			expected: 7,
		},
		{
			name:     "Single entry - forward no match",
			forward:  true,
			acType:   8,
			expected: -1,
		},
		{
			name:     "Single entry - reverse no match",
			forward:  false,
			acType:   78,
			expected: -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapAircraftType(singleMapping, tc.forward, tc.acType)
			if result != tc.expected {
				t.Errorf("mapAircraftType(single, forward=%v, type=%d) = %d, expected %d",
					tc.forward, tc.acType, result, tc.expected)
			}
		})
	}
}

// TestMapAircraftTypeMultipleMatches tests when first match should be returned
func TestMapAircraftTypeMultipleMatches(t *testing.T) {
	// Mapping with duplicate entries (first should be returned)
	duplicateMapping := [][]int{
		{1, 10},
		{1, 11}, // Duplicate key - should be ignored
		{2, 20},
	}

	testCases := []struct {
		name     string
		forward  bool
		acType   int
		expected int
	}{
		{
			name:     "Duplicate forward - returns first match",
			forward:  true,
			acType:   1,
			expected: 10, // First match should be returned
		},
		{
			name:     "Non-duplicate forward",
			forward:  true,
			acType:   2,
			expected: 20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapAircraftType(duplicateMapping, tc.forward, tc.acType)
			if result != tc.expected {
				t.Errorf("mapAircraftType(duplicates, forward=%v, type=%d) = %d, expected %d",
					tc.forward, tc.acType, result, tc.expected)
			}
			t.Logf("type %d -> %d (first match)", tc.acType, result)
		})
	}
}

// TestMapAircraftTypeNegativeValues tests handling of negative type values
func TestMapAircraftTypeNegativeValues(t *testing.T) {
	mapping := [][]int{
		{-1, -10},
		{1, 10},
		{-5, 50},
	}

	testCases := []struct {
		name     string
		forward  bool
		acType   int
		expected int
	}{
		{
			name:     "Negative forward match",
			forward:  true,
			acType:   -1,
			expected: -10,
		},
		{
			name:     "Negative reverse match",
			forward:  false,
			acType:   -10,
			expected: -1,
		},
		{
			name:     "Mixed: negative to positive",
			forward:  true,
			acType:   -5,
			expected: 50,
		},
		{
			name:     "Positive forward",
			forward:  true,
			acType:   1,
			expected: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapAircraftType(mapping, tc.forward, tc.acType)
			if result != tc.expected {
				t.Errorf("mapAircraftType(negatives, forward=%v, type=%d) = %d, expected %d",
					tc.forward, tc.acType, result, tc.expected)
			}
		})
	}
}

// TestMapAircraftTypeSymmetry tests that forward and reverse mappings are symmetric
func TestMapAircraftTypeSymmetry(t *testing.T) {
	mapping := [][]int{
		{1, 10},
		{2, 20},
		{3, 30},
	}

	// For each mapping entry, test that forward(x) = y implies reverse(y) = x
	for _, m := range mapping {
		from := m[0]
		to := m[1]

		// Test forward
		forwardResult := mapAircraftType(mapping, true, from)
		if forwardResult != to {
			t.Errorf("Forward mapping failed: %d -> %d (expected %d)", from, forwardResult, to)
		}

		// Test reverse
		reverseResult := mapAircraftType(mapping, false, to)
		if reverseResult != from {
			t.Errorf("Reverse mapping failed: %d -> %d (expected %d)", to, reverseResult, from)
		}

		// Test symmetry
		if forwardResult != to || reverseResult != from {
			t.Errorf("Symmetry broken for mapping [%d, %d]", from, to)
		} else {
			t.Logf("Symmetry verified: %d <-> %d", from, to)
		}
	}
}

// TestMapAircraftTypeZeroValues tests explicit zero value handling
func TestMapAircraftTypeZeroValues(t *testing.T) {
	mapping := [][]int{
		{0, 100}, // Zero maps to 100
		{1, 0},   // 1 maps to zero
		{2, 200},
	}

	testCases := []struct {
		name     string
		forward  bool
		acType   int
		expected int
	}{
		{
			name:     "Zero forward",
			forward:  true,
			acType:   0,
			expected: 100,
		},
		{
			name:     "Zero reverse",
			forward:  false,
			acType:   100,
			expected: 0,
		},
		{
			name:     "Maps to zero forward",
			forward:  true,
			acType:   1,
			expected: 0,
		},
		{
			name:     "Maps from zero reverse",
			forward:  false,
			acType:   0,
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapAircraftType(mapping, tc.forward, tc.acType)
			if result != tc.expected {
				t.Errorf("mapAircraftType(zeros, forward=%v, type=%d) = %d, expected %d",
					tc.forward, tc.acType, result, tc.expected)
			}
		})
	}
}
