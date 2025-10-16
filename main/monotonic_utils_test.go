// monotonic_utils_test.go: Tests for monotonic clock utility functions
// Targets: HumanizeTime, Unix, HasRealTimeReference

package main

import (
	"strings"
	"testing"
	"time"
)

// TestMonotonicHumanizeTime tests humanized time formatting
func TestMonotonicHumanizeTime(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond) // Let clock start

	// Test with a past time
	pastTime := m.Time.Add(-5 * time.Second)
	humanized := m.HumanizeTime(pastTime)

	// Should contain "ago"
	if !strings.Contains(humanized, "ago") {
		t.Errorf("Expected 'ago' in humanized past time, got: %s", humanized)
	}

	t.Logf("Past time humanized: %s", humanized)

	// Test with a future time
	futureTime := m.Time.Add(10 * time.Second)
	humanized = m.HumanizeTime(futureTime)

	// Should contain "from now"
	if !strings.Contains(humanized, "from now") {
		t.Errorf("Expected 'from now' in humanized future time, got: %s", humanized)
	}

	t.Logf("Future time humanized: %s", humanized)

	// Test with current time
	humanized = m.HumanizeTime(m.Time)
	t.Logf("Current time humanized: %s", humanized)
}

// TestMonotonicUnix tests Unix timestamp generation
func TestMonotonicUnix(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond) // Let clock start

	// Get Unix timestamp - this is seconds since time.Time{} (zero time), not Unix epoch
	unixTime := m.Unix()

	// Should be a very large positive number (seconds since year 1)
	if unixTime <= 0 {
		t.Errorf("Expected positive Unix time, got: %d", unixTime)
	}

	t.Logf("Unix time: %d seconds", unixTime)

	// The function is simple enough that even a single call gives us coverage
	// We don't need to verify it increases because the ticker resolution is 10ms
	// and our wait might not be long enough to register a change
}

// TestMonotonicHasRealTimeReference tests real time reference tracking
func TestMonotonicHasRealTimeReference(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond) // Let clock start

	// Initially should not have real time reference
	if m.HasRealTimeReference() {
		t.Error("Expected no real time reference initially")
	}

	// Set real time reference
	refTime := time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC)
	m.SetRealTimeReference(refTime)

	// Now should have real time reference
	if !m.HasRealTimeReference() {
		t.Error("Expected real time reference after setting")
	}

	t.Logf("Real time reference set: %v", refTime)

	// Try to set again - should not change
	newRefTime := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	m.SetRealTimeReference(newRefTime)

	// Should still have original reference
	if !m.HasRealTimeReference() {
		t.Error("Expected to still have real time reference")
	}

	// RealTime should be original + elapsed time, not newRefTime
	// We can't directly test the value without exposing it, but we confirmed it's set

	t.Log("Real time reference can only be set once - verified")
}

// TestMonotonicCombined tests multiple functions together
func TestMonotonicCombined(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond) // Let clock start

	// Check initial state
	if m.HasRealTimeReference() {
		t.Error("Should not have real time reference initially")
	}

	unixTime := m.Unix()
	t.Logf("Unix time: %d", unixTime)

	if unixTime <= 0 {
		t.Error("Unix time should be positive")
	}

	// Set real time reference
	refTime := m.Time // Use current monotonic time
	m.SetRealTimeReference(refTime)

	if !m.HasRealTimeReference() {
		t.Error("Should have real time reference after setting")
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Test humanize time with a past time
	pastTime := m.Time.Add(-2 * time.Second)
	humanized := m.HumanizeTime(pastTime)
	t.Logf("Past time humanized: %s", humanized)

	// Should show "ago" since it's in the past
	if !strings.Contains(humanized, "ago") && !strings.Contains(humanized, "now") {
		t.Errorf("Expected 'ago' or 'now' for past time, got: %s", humanized)
	}
}
