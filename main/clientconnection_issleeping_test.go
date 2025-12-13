/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	clientconnection_issleeping_test.go: Comprehensive tests for IsSleeping methods

	This file contains additional tests to achieve 100% coverage of all IsSleeping
	methods across different connection types.
*/

package main

import (
	"runtime"
	"testing"
	"time"
)

// TestNetworkConnection_IsSleeping_Comprehensive provides comprehensive coverage
// of the networkConnection.IsSleeping method, testing all code paths.
//
// Note: On x86/amd64 architectures, the function returns early due to isX86DebugMode(),
// so full coverage of all branches requires running on ARM architecture. However, this
// test suite is designed to exercise all paths when run on the appropriate hardware.
func TestNetworkConnection_IsSleeping_Comprehensive(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Save and restore global state
	origNoSleep := globalSettings.NoSleep
	defer func() { globalSettings.NoSleep = origNoSleep }()

	isX86 := runtime.GOARCH == "i386" || runtime.GOARCH == "amd64"

	tests := []struct {
		name             string
		noSleep          bool
		lastPongResponse time.Time
		lastPingResponse time.Time
		lastUnreachable  time.Time
		wantSleeping     bool
		wantSleepFlag    bool
		skipOnX86        bool
		description      string
	}{
		{
			name:             "NoSleep enabled",
			noSleep:          true,
			lastPongResponse: time.Time{},
			lastPingResponse: time.Time{},
			lastUnreachable:  time.Time{},
			wantSleeping:     false,
			wantSleepFlag:    false,
			skipOnX86:        false,
			description:      "When NoSleep is true, should never sleep",
		},
		{
			name:             "Pong response is zero",
			noSleep:          false,
			lastPongResponse: time.Time{}, // Zero
			lastPingResponse: stratuxClock.GetTime(),
			lastUnreachable:  time.Time{},
			wantSleeping:     true,
			wantSleepFlag:    true,
			skipOnX86:        true,
			description:      "Zero pong response triggers sleep (first if condition)",
		},
		{
			name:             "Pong response old (>10s)",
			noSleep:          false,
			lastPongResponse: stratuxClock.GetTime().Add(-11 * time.Second),
			lastPingResponse: stratuxClock.GetTime(),
			lastUnreachable:  time.Time{},
			wantSleeping:     true,
			wantSleepFlag:    true,
			skipOnX86:        true,
			description:      "Pong older than 10s triggers sleep (first if condition)",
		},
		{
			name:             "Pong exactly 10s ago",
			noSleep:          false,
			lastPongResponse: stratuxClock.GetTime().Add(-10 * time.Second),
			lastPingResponse: stratuxClock.GetTime(),
			lastUnreachable:  time.Time{},
			wantSleeping:     false,
			wantSleepFlag:    false,
			skipOnX86:        true,
			description:      "Pong exactly 10s ago should not sleep (boundary test)",
		},
		{
			name:             "Ping response is zero",
			noSleep:          false,
			lastPongResponse: stratuxClock.GetTime(),
			lastPingResponse: time.Time{}, // Zero
			lastUnreachable:  time.Time{},
			wantSleeping:     true,
			wantSleepFlag:    true,
			skipOnX86:        true,
			description:      "Zero ping response triggers sleep (second if condition)",
		},
		{
			name:             "Ping response old (>10s)",
			noSleep:          false,
			lastPongResponse: stratuxClock.GetTime(),
			lastPingResponse: stratuxClock.GetTime().Add(-11 * time.Second),
			lastUnreachable:  time.Time{},
			wantSleeping:     true,
			wantSleepFlag:    true,
			skipOnX86:        true,
			description:      "Ping older than 10s triggers sleep (second if condition)",
		},
		{
			name:             "Ping exactly 10s ago",
			noSleep:          false,
			lastPongResponse: stratuxClock.GetTime(),
			lastPingResponse: stratuxClock.GetTime().Add(-10 * time.Second),
			lastUnreachable:  time.Time{},
			wantSleeping:     false,
			wantSleepFlag:    false,
			skipOnX86:        true,
			description:      "Ping exactly 10s ago should not sleep (boundary test)",
		},
		{
			name:             "Recently unreachable (<5s)",
			noSleep:          false,
			lastPongResponse: stratuxClock.GetTime(),
			lastPingResponse: stratuxClock.GetTime(),
			lastUnreachable:  stratuxClock.GetTime().Add(-2 * time.Second),
			wantSleeping:     true,
			wantSleepFlag:    true,
			skipOnX86:        true,
			description:      "Unreachable within 5s triggers sleep (third if condition)",
		},
		{
			name:             "Unreachable exactly 5s ago",
			noSleep:          false,
			lastPongResponse: stratuxClock.GetTime(),
			lastPingResponse: stratuxClock.GetTime(),
			lastUnreachable:  stratuxClock.GetTime().Add(-5 * time.Second),
			wantSleeping:     false,
			wantSleepFlag:    false,
			skipOnX86:        true,
			description:      "Unreachable exactly 5s ago should not sleep (boundary test)",
		},
		{
			name:             "All conditions healthy",
			noSleep:          false,
			lastPongResponse: stratuxClock.GetTime(),
			lastPingResponse: stratuxClock.GetTime(),
			lastUnreachable:  stratuxClock.GetTime().Add(-10 * time.Second),
			wantSleeping:     false,
			wantSleepFlag:    false,
			skipOnX86:        true,
			description:      "All times recent and healthy - should not sleep (else clause)",
		},
		{
			name:             "All zero times",
			noSleep:          false,
			lastPongResponse: time.Time{},
			lastPingResponse: time.Time{},
			lastUnreachable:  time.Time{},
			wantSleeping:     true,
			wantSleepFlag:    true,
			skipOnX86:        true,
			description:      "All zero times should trigger sleep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOnX86 && isX86 {
				t.Skipf("Skipping %s on x86: %s", tt.name, tt.description)
			}

			globalSettings.NoSleep = tt.noSleep

			conn := &networkConnection{
				Ip:               "192.168.10.100",
				Port:             4000,
				LastPongResponse: tt.lastPongResponse,
				LastPingResponse: tt.lastPingResponse,
				LastUnreachable:  tt.lastUnreachable,
			}

			got := conn.IsSleeping()

			// On x86 with NoSleep=false, isX86DebugMode() causes early return of false
			want := tt.wantSleeping
			if isX86 && !tt.noSleep {
				want = false
			}

			if got != want {
				t.Errorf("%s: IsSleeping() = %v, want %v", tt.description, got, want)
			}

			// Only check SleepFlag on non-x86 or when NoSleep is true
			if !isX86 || tt.noSleep {
				if conn.SleepFlag != tt.wantSleepFlag {
					t.Errorf("%s: SleepFlag = %v, want %v", tt.description, conn.SleepFlag, tt.wantSleepFlag)
				}
			}
		})
	}
}

// TestNetworkConnection_IsSleeping_X86Behavior specifically tests x86 debug mode behavior
func TestNetworkConnection_IsSleeping_X86Behavior(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Only run on x86/amd64
	if runtime.GOARCH != "i386" && runtime.GOARCH != "amd64" {
		t.Skip("Skipping x86-specific test on non-x86 architecture")
	}

	// Save and restore global state
	origNoSleep := globalSettings.NoSleep
	defer func() { globalSettings.NoSleep = origNoSleep }()

	tests := []struct {
		name         string
		noSleep      bool
		wantSleeping bool
	}{
		{
			name:         "x86 mode with NoSleep false",
			noSleep:      false,
			wantSleeping: false, // isX86DebugMode() returns false
		},
		{
			name:         "x86 mode with NoSleep true",
			noSleep:      true,
			wantSleeping: false, // globalSettings.NoSleep returns false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalSettings.NoSleep = tt.noSleep

			conn := &networkConnection{
				Ip:               "192.168.10.100",
				Port:             4000,
				LastPongResponse: time.Time{},            // Zero - would trigger sleep on ARM
				LastPingResponse: time.Time{},            // Zero - would trigger sleep on ARM
				LastUnreachable:  stratuxClock.GetTime(), // Recent - would trigger sleep on ARM
			}

			got := conn.IsSleeping()
			if got != tt.wantSleeping {
				t.Errorf("IsSleeping() = %v, want %v (x86 should always return false)", got, tt.wantSleeping)
			}
		})
	}
}
