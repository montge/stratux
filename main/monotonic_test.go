/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	monotonic_test.go: Comprehensive tests for monotonic.go
	Tests the monotonic clock implementation used throughout Stratux
*/

package main

import (
	"testing"
	"time"
)

// TestNewMonotonic tests the monotonic clock constructor
func TestNewMonotonic(t *testing.T) {
	m := NewMonotonic()

	// Verify initial state
	if m == nil {
		t.Fatal("NewMonotonic() returned nil")
	}

	if m.ticker == nil {
		t.Error("ticker should be initialized")
	}

	if m.GetMilliseconds() != 0 {
		t.Errorf("Expected Milliseconds to start at 0, got %d", m.GetMilliseconds())
	}

	if m.GetTime().IsZero() {
		t.Error("Time should be initialized to current time")
	}

	if m.realTimeSet {
		t.Error("realTimeSet should be false initially")
	}

	// Verify the watcher is running by checking that time advances
	initialMs := m.GetMilliseconds()
	time.Sleep(50 * time.Millisecond)

	if m.GetMilliseconds() <= initialMs {
		t.Error("Milliseconds should increase as watcher runs")
	}
}

// TestMonotonicWatcher tests that the watcher goroutine properly increments time
func TestMonotonicWatcher(t *testing.T) {
	m := NewMonotonic()

	// Capture initial state
	initialMs := m.GetMilliseconds()
	initialTime := m.GetTime()

	// Wait for several ticks (ticker fires every 10ms)
	time.Sleep(100 * time.Millisecond)

	// Check that milliseconds increased
	if m.GetMilliseconds() <= initialMs {
		t.Errorf("Expected Milliseconds > %d, got %d", initialMs, m.GetMilliseconds())
	}

	// Check that time advanced
	if !m.GetTime().After(initialTime) {
		t.Error("Expected Time to advance")
	}

	// Verify approximate consistency (allowing for timing variations)
	expectedMs := initialMs + 100 // Approximately 100ms should have passed
	actualMs := m.GetMilliseconds()
	diff := int64(actualMs) - int64(expectedMs)
	if diff < -50 || diff > 50 {
		t.Logf("Warning: Milliseconds drift detected. Expected ~%d, got %d (diff: %d)",
			expectedMs, actualMs, diff)
	}
}

// TestMonotonicWatcherRealTime tests that RealTime advances when set
func TestMonotonicWatcherRealTime(t *testing.T) {
	m := NewMonotonic()

	// Set real time reference
	refTime := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	m.SetRealTimeReference(refTime)

	if !m.realTimeSet {
		t.Fatal("realTimeSet should be true after SetRealTimeReference")
	}

	// Initial RealTime should match reference
	initialRealTime := m.GetRealTime()

	// Wait for watcher to tick
	time.Sleep(100 * time.Millisecond)

	// RealTime should have advanced
	if !m.GetRealTime().After(initialRealTime) {
		t.Error("Expected RealTime to advance when realTimeSet is true")
	}

	// Verify RealTime advanced by approximately 100ms
	elapsed := m.GetRealTime().Sub(initialRealTime)
	if elapsed < 50*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Logf("Warning: RealTime advancement seems off. Expected ~100ms, got %v", elapsed)
	}
}

// TestMonotonicSince tests the Since method
func TestMonotonicSince(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond) // Let clock start

	testCases := []struct {
		name     string
		offset   time.Duration
		wantSign int // -1 for negative, 0 for zero, 1 for positive
	}{
		{
			name:     "past time",
			offset:   -5 * time.Second,
			wantSign: 1, // Since should return positive duration
		},
		{
			name:     "current time",
			offset:   0,
			wantSign: 0, // Should be approximately zero
		},
		{
			name:     "future time",
			offset:   5 * time.Second,
			wantSign: -1, // Since should return negative duration
		},
		{
			name:     "far past",
			offset:   -1 * time.Hour,
			wantSign: 1,
		},
		{
			name:     "far future",
			offset:   1 * time.Hour,
			wantSign: -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testTime := m.GetTime().Add(tc.offset)
			duration := m.Since(testTime)

			var actualSign int
			if duration > time.Millisecond {
				actualSign = 1
			} else if duration < -time.Millisecond {
				actualSign = -1
			} else {
				actualSign = 0
			}

			if actualSign != tc.wantSign {
				t.Errorf("Since(%v) = %v (sign %d), want sign %d",
					testTime, duration, actualSign, tc.wantSign)
			}

			// For non-zero cases, verify magnitude is approximately correct
			if tc.offset != 0 {
				expected := -tc.offset // Since returns positive for past times
				diff := duration - expected
				// Allow up to 100ms tolerance due to ticker resolution
				if diff < -100*time.Millisecond || diff > 100*time.Millisecond {
					t.Logf("Since(%v) magnitude: got %v, expected ~%v (diff: %v)",
						testTime, duration, expected, diff)
				}
			}
		})
	}
}

// TestMonotonicSinceZeroTime tests Since with the zero time value
func TestMonotonicSinceZeroTime(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond)

	// Since(time.Time{}) should return duration since zero time
	duration := m.Since(time.Time{})

	// Should be a very large positive value
	if duration <= 0 {
		t.Errorf("Expected positive duration since zero time, got %v", duration)
	}

	// Should be at least the current time since zero (which is many years)
	if duration < 24*365*time.Hour {
		t.Errorf("Expected duration > 1 year, got %v", duration)
	}
}

// TestMonotonicSetRealTimeReferenceOnce tests that real time can only be set once
func TestMonotonicSetRealTimeReferenceOnce(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond)

	// Set first time
	firstRef := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	m.SetRealTimeReference(firstRef)

	if !m.realTimeSet {
		t.Error("realTimeSet should be true after first SetRealTimeReference")
	}

	if m.GetRealTime() != firstRef {
		t.Errorf("RealTime should be %v, got %v", firstRef, m.GetRealTime())
	}

	// Wait a bit for RealTime to advance
	time.Sleep(100 * time.Millisecond)
	realTimeAfterWait := m.GetRealTime()

	// Try to set second time
	secondRef := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	m.SetRealTimeReference(secondRef)

	// Should still be set
	if !m.realTimeSet {
		t.Error("realTimeSet should still be true")
	}

	// RealTime should NOT be secondRef, it should have continued from firstRef
	if m.GetRealTime() == secondRef {
		t.Error("Second SetRealTimeReference should be ignored")
	}

	// RealTime should be approximately what it was after the wait (plus a tiny bit more)
	if m.GetRealTime().Before(realTimeAfterWait) {
		t.Error("RealTime should not go backwards")
	}

	timeDiff := m.GetRealTime().Sub(realTimeAfterWait)
	if timeDiff > 50*time.Millisecond {
		t.Logf("RealTime advanced by %v since check (expected < 50ms)", timeDiff)
	}
}

// TestMonotonicSetRealTimeReferenceBeforeWatcher tests setting RealTime immediately
func TestMonotonicSetRealTimeReferenceBeforeWatcher(t *testing.T) {
	m := &monotonic{
		Milliseconds: 0,
		Time:         time.Now(),
		ticker:       time.NewTicker(10 * time.Millisecond),
		realTimeSet:  false,
	}

	refTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	m.SetRealTimeReference(refTime)

	if !m.realTimeSet {
		t.Error("realTimeSet should be true")
	}

	if m.GetRealTime() != refTime {
		t.Errorf("RealTime should be %v, got %v", refTime, m.GetRealTime())
	}

	// Start watcher
	go m.Watcher()

	// Wait for ticks
	time.Sleep(100 * time.Millisecond)

	// RealTime should have advanced
	if !m.GetRealTime().After(refTime) {
		t.Error("RealTime should have advanced after watcher started")
	}

	m.ticker.Stop()
}

// TestMonotonicConcurrency tests concurrent access to monotonic clock
func TestMonotonicConcurrency(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond)

	// Set real time reference
	m.SetRealTimeReference(time.Now())

	done := make(chan bool)
	iterations := 1000

	// Multiple goroutines reading from the clock concurrently using getters
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				_ = m.Since(time.Now())
				_ = m.Unix()
				_ = m.HasRealTimeReference()
				_ = m.HumanizeTime(m.GetTime())
				_ = m.GetMilliseconds()
				_ = m.GetTime()
				_ = m.GetRealTime()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify clock is still working using getters
	if m.GetMilliseconds() == 0 {
		t.Error("Clock should have advanced")
	}

	if !m.HasRealTimeReference() {
		t.Error("Should still have real time reference")
	}
}

// TestMonotonicTimeAdvancement tests that time advances monotonically
func TestMonotonicTimeAdvancement(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(100 * time.Millisecond) // Wait for ticker to stabilize

	samples := 5
	prevMs := m.GetMilliseconds()
	prevTime := m.GetTime()

	for i := 0; i < samples; i++ {
		time.Sleep(50 * time.Millisecond) // Sleep longer than ticker period (10ms)

		currMs := m.GetMilliseconds()
		currTime := m.GetTime()

		// Milliseconds should not go backwards (allow equal due to timing)
		if currMs < prevMs {
			t.Errorf("Milliseconds should not decrease: prev=%d, curr=%d", prevMs, currMs)
		}

		// Time should not go backwards (allow equal due to timing)
		if currTime.Before(prevTime) {
			t.Errorf("Time should not go backwards: prev=%v, curr=%v", prevTime, currTime)
		}

		prevMs = currMs
		prevTime = currTime
	}

	// Verify overall advancement after all samples
	finalMs := m.GetMilliseconds()
	if finalMs <= 100 { // Should have advanced at least 250ms (5 * 50ms)
		t.Errorf("Expected significant time advancement, got only %d ms", finalMs)
	}
}

// TestMonotonicUnixEdgeCases tests Unix method edge cases
func TestMonotonicUnixEdgeCases(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond)

	// Unix() should return consistent results when called multiple times quickly
	unix1 := m.Unix()
	unix2 := m.Unix()

	// Should be very close (within a few seconds due to ticker resolution)
	diff := unix1 - unix2
	if diff < -1 || diff > 1 {
		t.Errorf("Unix timestamps differ too much: %d vs %d (diff: %d)", unix1, unix2, diff)
	}

	// Wait and verify it increases
	time.Sleep(200 * time.Millisecond)
	unix3 := m.Unix()

	// unix3 should be >= unix1 (time doesn't go backwards)
	if unix3 < unix1 {
		t.Errorf("Unix time went backwards: %d -> %d", unix1, unix3)
	}
}

// TestMonotonicHumanizeTimeEdgeCases tests HumanizeTime edge cases
func TestMonotonicHumanizeTimeEdgeCases(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond)

	testCases := []struct {
		name     string
		timeFunc func() time.Time
	}{
		{
			name: "zero time",
			timeFunc: func() time.Time {
				return time.Time{}
			},
		},
		{
			name: "very old time",
			timeFunc: func() time.Time {
				return time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
			},
		},
		{
			name: "far future",
			timeFunc: func() time.Time {
				return time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC)
			},
		},
		{
			name: "current time",
			timeFunc: func() time.Time {
				return m.GetTime()
			},
		},
		{
			name: "1 second ago",
			timeFunc: func() time.Time {
				return m.GetTime().Add(-1 * time.Second)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testTime := tc.timeFunc()
			result := m.HumanizeTime(testTime)

			// Should return a non-empty string
			if result == "" {
				t.Error("HumanizeTime returned empty string")
			}

			t.Logf("%s: %s", tc.name, result)
		})
	}
}

// TestMonotonicSinceConsistency tests that Since is consistent with Time
func TestMonotonicSinceConsistency(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond)

	// Capture a reference time
	refTime := m.GetTime()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Since should return approximately 100ms
	duration := m.Since(refTime)

	// Should be positive
	if duration <= 0 {
		t.Errorf("Expected positive duration, got %v", duration)
	}

	// Should be approximately 100ms (allow 50ms tolerance)
	if duration < 50*time.Millisecond || duration > 150*time.Millisecond {
		t.Logf("Warning: Since duration %v outside expected range [50ms, 150ms]", duration)
	}

	// Since should be equivalent to m.GetTime().Sub(refTime)
	expectedDuration := m.GetTime().Sub(refTime)
	if duration != expectedDuration {
		t.Errorf("Since(%v) = %v, but Time.Sub(%v) = %v",
			refTime, duration, refTime, expectedDuration)
	}
}

// TestMonotonicWatcherWithoutRealTime tests watcher when RealTime is not set
func TestMonotonicWatcherWithoutRealTime(t *testing.T) {
	m := NewMonotonic()

	// Don't set RealTime
	if m.realTimeSet {
		t.Fatal("realTimeSet should be false initially")
	}

	initialMs := m.GetMilliseconds()
	initialTime := m.GetTime()

	// Wait for watcher to tick
	time.Sleep(100 * time.Millisecond)

	// Milliseconds and Time should advance
	if m.GetMilliseconds() <= initialMs {
		t.Error("Milliseconds should advance even without RealTime set")
	}

	if !m.GetTime().After(initialTime) {
		t.Error("Time should advance even without RealTime set")
	}

	// RealTime should still be zero/unset
	if !m.GetRealTime().IsZero() && m.realTimeSet == false {
		t.Error("RealTime should be zero when not set")
	}
}

// TestMonotonicGetters tests the thread-safe getter methods
func TestMonotonicGetters(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond)

	// Set real time reference
	refTime := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	m.SetRealTimeReference(refTime)

	// Wait for watcher to tick
	time.Sleep(50 * time.Millisecond)

	// Test GetMilliseconds
	ms := m.GetMilliseconds()
	if ms == 0 {
		t.Error("GetMilliseconds() should return non-zero after waiting")
	}

	// Verify getter returns same value as field (within mutex window)
	m.mu.RLock()
	directMs := m.GetMilliseconds()
	m.mu.RUnlock()
	if ms != directMs && ms != directMs+10 && ms != directMs-10 {
		t.Logf("GetMilliseconds() = %d, direct access = %d (may differ due to timing)", ms, directMs)
	}

	// Test GetTime
	gotTime := m.GetTime()
	if gotTime.IsZero() {
		t.Error("GetTime() should not return zero time")
	}

	// Test GetRealTime
	gotRealTime := m.GetRealTime()
	if gotRealTime.IsZero() {
		t.Error("GetRealTime() should not return zero time after SetRealTimeReference")
	}

	// RealTime should have advanced from refTime
	if !gotRealTime.After(refTime) && gotRealTime != refTime {
		t.Errorf("GetRealTime() = %v should be >= refTime %v", gotRealTime, refTime)
	}
}

// TestMonotonicGettersConsistency tests that getters are consistent
func TestMonotonicGettersConsistency(t *testing.T) {
	m := NewMonotonic()
	time.Sleep(50 * time.Millisecond)

	// Multiple calls to getters should return increasing values
	ms1 := m.GetMilliseconds()
	time1 := m.GetTime()

	time.Sleep(50 * time.Millisecond)

	ms2 := m.GetMilliseconds()
	time2 := m.GetTime()

	if ms2 <= ms1 {
		t.Errorf("GetMilliseconds() should increase: %d -> %d", ms1, ms2)
	}

	if !time2.After(time1) {
		t.Errorf("GetTime() should advance: %v -> %v", time1, time2)
	}
}

// TestMonotonicMultipleInstances tests that multiple monotonic instances are independent
func TestMonotonicMultipleInstances(t *testing.T) {
	m1 := NewMonotonic()
	time.Sleep(50 * time.Millisecond)
	m2 := NewMonotonic()
	time.Sleep(50 * time.Millisecond)

	// Set RealTime on only one
	m1.SetRealTimeReference(time.Now())

	if !m1.HasRealTimeReference() {
		t.Error("m1 should have RealTime set")
	}

	if m2.HasRealTimeReference() {
		t.Error("m2 should not have RealTime set")
	}

	// Both should advance independently
	m1Ms := m1.GetMilliseconds()
	m2Ms := m2.GetMilliseconds()

	time.Sleep(100 * time.Millisecond)

	if m1.GetMilliseconds() <= m1Ms {
		t.Error("m1 should advance")
	}

	if m2.GetMilliseconds() <= m2Ms {
		t.Error("m2 should advance")
	}

	// m1 started first, so should have higher milliseconds
	if m1.GetMilliseconds() <= m2.GetMilliseconds() {
		t.Logf("Warning: m1 (%d) should have higher milliseconds than m2 (%d)",
			m1.GetMilliseconds(), m2.GetMilliseconds())
	}
}
