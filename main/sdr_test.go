/*
	Copyright (c) 2015-2016 Christopher Young
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file, herein included
	as part of this header.

	sdr_test.go: Unit tests for SDR helper functions.
*/

package main

import (
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestGetPPM tests the PPM extraction from serial strings
func TestGetPPM(t *testing.T) {
	tests := []struct {
		name     string
		serial   string
		expected int
	}{
		{
			name:     "Valid stratux serial with PPM",
			serial:   "stratux:978:52",
			expected: 52,
		},
		{
			name:     "Valid stratux serial with negative PPM",
			serial:   "stratux:1090:-15",
			expected: -15,
		},
		{
			name:     "Valid stratux serial with zero PPM",
			serial:   "stratux:868:0",
			expected: 0,
		},
		{
			name:     "Stratux serial without PPM",
			serial:   "stratux:978",
			expected: globalSettings.PPM,
		},
		{
			name:     "Stratux serial with colon but no PPM",
			serial:   "stratux:1090:",
			expected: globalSettings.PPM,
		},
		{
			name:     "Stratux with optional 't' omitted",
			serial:   "sratux:978:25",
			expected: globalSettings.PPM,
		},
		{
			name:     "Invalid legacy format tatux",
			serial:   "tatux:1090:10",
			expected: globalSettings.PPM,
		},
		{
			name:     "Invalid legacy format tux",
			serial:   "tux:868:5",
			expected: globalSettings.PPM,
		},
		{
			name:     "Non-stratux serial",
			serial:   "00000001",
			expected: globalSettings.PPM,
		},
		{
			name:     "Empty serial",
			serial:   "",
			expected: globalSettings.PPM,
		},
		{
			name:     "Serial with invalid PPM (non-numeric)",
			serial:   "stratux:978:abc",
			expected: globalSettings.PPM,
		},
		{
			name:     "Large positive PPM",
			serial:   "stratux:1090:999",
			expected: 999,
		},
		{
			name:     "Large negative PPM",
			serial:   "stratux:868:-999",
			expected: -999,
		},
		{
			name:     "Serial with extra colons",
			serial:   "stratux:978:42:extra",
			expected: 42,
		},
		{
			name:     "Serial with only colon separator",
			serial:   "stratux:1090:",
			expected: globalSettings.PPM,
		},
		{
			name:     "Serial with PPM overflow value",
			serial:   "stratux:978:99999999999999999999",
			expected: globalSettings.PPM,
		},
		{
			// Regex requires 'x' at end: str?a?t?u?x - partial prefix without 'x' won't match
			name:     "Serial matching str prefix only (missing x)",
			serial:   "str:978:10",
			expected: globalSettings.PPM,
		},
		{
			name:     "Serial matching stra prefix only (missing x)",
			serial:   "stra:1090:20",
			expected: globalSettings.PPM,
		},
		{
			name:     "Serial matching strat prefix only (missing x)",
			serial:   "strat:868:30",
			expected: globalSettings.PPM,
		},
		{
			name:     "Serial matching stratu prefix only (missing x)",
			serial:   "stratu:162:40",
			expected: globalSettings.PPM,
		},
		{
			// stx matches: s t (skip r) (skip a) (skip t) (skip u) x
			name:     "Serial with minimal prefix stx",
			serial:   "stx:978:50",
			expected: 50,
		},
		{
			name:     "Serial with single digit frequency",
			serial:   "stratux:9:5",
			expected: 5,
		},
		{
			name:     "Serial with PPM but no trailing colon",
			serial:   "stratux:978:123",
			expected: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPPM(tt.serial)
			if result != tt.expected {
				t.Errorf("getPPM(%q) = %d, want %d", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestReCompile tests the regex compilation helper
func TestReCompile(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantNil bool
	}{
		{
			name:    "Valid simple pattern",
			pattern: "test",
			wantNil: false,
		},
		{
			name:    "Valid complex pattern",
			pattern: "str?a?t?u?x:978",
			wantNil: false,
		},
		{
			name:    "Valid pattern with numbers",
			pattern: "\\d+",
			wantNil: false,
		},
		{
			name:    "Invalid pattern - unclosed bracket",
			pattern: "[abc",
			wantNil: true,
		},
		{
			name:    "Invalid pattern - unclosed paren",
			pattern: "(abc",
			wantNil: true,
		},
		{
			name:    "Empty pattern",
			pattern: "",
			wantNil: false,
		},
		{
			name:    "Pattern with alternation",
			pattern: "a|b|c",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reCompile(tt.pattern)
			if (result == nil) != tt.wantNil {
				t.Errorf("reCompile(%q) nil=%v, want nil=%v", tt.pattern, result == nil, tt.wantNil)
			}
		})
	}
}

// TestRegexUATHasID tests UAT serial matching
func TestRegexUATHasID(t *testing.T) {
	tests := []struct {
		name     string
		serial   string
		expected bool
	}{
		{
			name:     "Exact match stratux:978",
			serial:   "stratux:978",
			expected: true,
		},
		{
			name:     "Match with PPM",
			serial:   "stratux:978:52",
			expected: true,
		},
		{
			name:     "Invalid format sratux",
			serial:   "sratux:978",
			expected: false,
		},
		{
			name:     "Invalid format tatux",
			serial:   "tatux:978",
			expected: false,
		},
		{
			name:     "Invalid format tux",
			serial:   "tux:978",
			expected: false,
		},
		{
			name:     "No match - wrong frequency",
			serial:   "stratux:1090",
			expected: false,
		},
		{
			name:     "No match - plain serial",
			serial:   "00000001",
			expected: false,
		},
		{
			name:     "No match - empty",
			serial:   "",
			expected: false,
		},
		{
			name:     "No match - 868",
			serial:   "stratux:868",
			expected: false,
		},
		{
			name:     "No match - 162",
			serial:   "stratux:162",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rUAT.hasID(tt.serial)
			if result != tt.expected {
				t.Errorf("rUAT.hasID(%q) = %v, want %v", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestRegexUATHasIDNilRegex tests UAT matching with nil regex
func TestRegexUATHasIDNilRegex(t *testing.T) {
	var nilRegex *regexUAT
	tests := []struct {
		name     string
		serial   string
		expected bool
	}{
		{
			name:     "Exact prefix match",
			serial:   "stratux:978",
			expected: true,
		},
		{
			name:     "Prefix match with PPM",
			serial:   "stratux:978:52",
			expected: true,
		},
		{
			name:     "No match - wrong frequency",
			serial:   "stratux:1090",
			expected: false,
		},
		{
			name:     "No match - legacy format",
			serial:   "sratux:978",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nilRegex.hasID(tt.serial)
			if result != tt.expected {
				t.Errorf("nilRegex.hasID(%q) = %v, want %v", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestRegexESHasID tests ES (1090MHz) serial matching
func TestRegexESHasID(t *testing.T) {
	tests := []struct {
		name     string
		serial   string
		expected bool
	}{
		{
			name:     "Exact match stratux:1090",
			serial:   "stratux:1090",
			expected: true,
		},
		{
			name:     "Match with PPM",
			serial:   "stratux:1090:52",
			expected: true,
		},
		{
			name:     "Invalid format sratux",
			serial:   "sratux:1090",
			expected: false,
		},
		{
			name:     "Invalid format tatux",
			serial:   "tatux:1090",
			expected: false,
		},
		{
			name:     "No match - wrong frequency",
			serial:   "stratux:978",
			expected: false,
		},
		{
			name:     "No match - plain serial",
			serial:   "00000001",
			expected: false,
		},
		{
			name:     "No match - empty",
			serial:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rES.hasID(tt.serial)
			if result != tt.expected {
				t.Errorf("rES.hasID(%q) = %v, want %v", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestRegexESHasIDNilRegex tests ES matching with nil regex
func TestRegexESHasIDNilRegex(t *testing.T) {
	var nilRegex *regexES
	tests := []struct {
		name     string
		serial   string
		expected bool
	}{
		{
			name:     "Exact prefix match",
			serial:   "stratux:1090",
			expected: true,
		},
		{
			name:     "Prefix match with PPM",
			serial:   "stratux:1090:52",
			expected: true,
		},
		{
			name:     "No match - wrong frequency",
			serial:   "stratux:978",
			expected: false,
		},
		{
			name:     "No match - legacy format",
			serial:   "sratux:1090",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nilRegex.hasID(tt.serial)
			if result != tt.expected {
				t.Errorf("nilRegex.hasID(%q) = %v, want %v", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestRegexOGNHasID tests OGN (868MHz) serial matching
func TestRegexOGNHasID(t *testing.T) {
	tests := []struct {
		name     string
		serial   string
		expected bool
	}{
		{
			name:     "Exact match stratux:868",
			serial:   "stratux:868",
			expected: true,
		},
		{
			name:     "Match with PPM",
			serial:   "stratux:868:52",
			expected: true,
		},
		{
			name:     "Invalid format sratux",
			serial:   "sratux:868",
			expected: false,
		},
		{
			name:     "Invalid format tatux",
			serial:   "tatux:868",
			expected: false,
		},
		{
			name:     "No match - wrong frequency",
			serial:   "stratux:978",
			expected: false,
		},
		{
			name:     "No match - plain serial",
			serial:   "00000001",
			expected: false,
		},
		{
			name:     "No match - empty",
			serial:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rOGN.hasID(tt.serial)
			if result != tt.expected {
				t.Errorf("rOGN.hasID(%q) = %v, want %v", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestRegexOGNHasIDNilRegex tests OGN matching with nil regex
func TestRegexOGNHasIDNilRegex(t *testing.T) {
	var nilRegex *regexOGN
	tests := []struct {
		name     string
		serial   string
		expected bool
	}{
		{
			name:     "Exact prefix match",
			serial:   "stratux:868",
			expected: true,
		},
		{
			name:     "Prefix match with PPM",
			serial:   "stratux:868:52",
			expected: true,
		},
		{
			name:     "No match - wrong frequency",
			serial:   "stratux:978",
			expected: false,
		},
		{
			name:     "No match - legacy format",
			serial:   "sratux:868",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nilRegex.hasID(tt.serial)
			if result != tt.expected {
				t.Errorf("nilRegex.hasID(%q) = %v, want %v", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestRegexOGNHasIDWithRegex tests OGN matching with a non-nil regex that matches
func TestRegexOGNHasIDWithRegex(t *testing.T) {
	// Create a regex that matches any serial containing "868"
	pattern := regexp.MustCompile(".*868.*")
	regexWithPattern := (*regexOGN)(pattern)

	tests := []struct {
		name     string
		serial   string
		expected bool
	}{
		{
			name:     "Match via regex - contains 868",
			serial:   "stratux:868:52",
			expected: true,
		},
		{
			name:     "Match via regex - custom with 868",
			serial:   "custom-868-device",
			expected: true,
		},
		{
			name:     "No match - doesn't contain 868",
			serial:   "stratux:978",
			expected: false,
		},
		{
			name:     "No match - empty string",
			serial:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := regexWithPattern.hasID(tt.serial)
			if result != tt.expected {
				t.Errorf("regexWithPattern.hasID(%q) = %v, want %v", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestRegexAISHasID tests AIS (162MHz) serial matching
func TestRegexAISHasID(t *testing.T) {
	tests := []struct {
		name     string
		serial   string
		expected bool
	}{
		{
			name:     "Exact match stratux:162",
			serial:   "stratux:162",
			expected: true,
		},
		{
			name:     "Match with PPM",
			serial:   "stratux:162:52",
			expected: true,
		},
		{
			name:     "Invalid format sratux",
			serial:   "sratux:162",
			expected: false,
		},
		{
			name:     "Invalid format tatux",
			serial:   "tatux:162",
			expected: false,
		},
		{
			name:     "No match - wrong frequency",
			serial:   "stratux:978",
			expected: false,
		},
		{
			name:     "No match - plain serial",
			serial:   "00000001",
			expected: false,
		},
		{
			name:     "No match - empty",
			serial:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rAIS.hasID(tt.serial)
			if result != tt.expected {
				t.Errorf("rAIS.hasID(%q) = %v, want %v", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestRegexAISHasIDNilRegex tests AIS matching with nil regex
func TestRegexAISHasIDNilRegex(t *testing.T) {
	var nilRegex *regexAIS
	tests := []struct {
		name     string
		serial   string
		expected bool
	}{
		{
			name:     "Exact prefix match",
			serial:   "stratux:162",
			expected: true,
		},
		{
			name:     "Prefix match with PPM",
			serial:   "stratux:162:52",
			expected: true,
		},
		{
			name:     "No match - wrong frequency",
			serial:   "stratux:978",
			expected: false,
		},
		{
			name:     "No match - legacy format",
			serial:   "sratux:162",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nilRegex.hasID(tt.serial)
			if result != tt.expected {
				t.Errorf("nilRegex.hasID(%q) = %v, want %v", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestDump1090TermMessage tests the Dump1090TermMessage struct
func TestDump1090TermMessage(t *testing.T) {
	msg := Dump1090TermMessage{
		Text:   "test message",
		Source: "stdout",
	}

	if msg.Text != "test message" {
		t.Errorf("Message text = %q, want %q", msg.Text, "test message")
	}
	if msg.Source != "stdout" {
		t.Errorf("Message source = %q, want %q", msg.Source, "stdout")
	}
}

// TestAISTermMessage tests the AISTermMessage struct
func TestAISTermMessage(t *testing.T) {
	msg := AISTermMessage{
		Text:   "test message",
		Source: "stderr",
	}

	if msg.Text != "test message" {
		t.Errorf("Message text = %q, want %q", msg.Text, "test message")
	}
	if msg.Source != "stderr" {
		t.Errorf("Message source = %q, want %q", msg.Source, "stderr")
	}
}

// TestSDRConstants tests that the UAT configuration constants are defined correctly
func TestSDRConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"TunerGain", TunerGain, 480},
		{"SampleRate", SampleRate, 2083334},
		{"NewRTLFreq", NewRTLFreq, 28800000},
		{"NewTunerFreq", NewTunerFreq, 28800000},
		{"CenterFreq", CenterFreq, 978000000},
		{"Bandwidth", Bandwidth, 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.expected)
			}
		})
	}
}

// TestGetPPMRegexPatterns tests that the regex patterns used in getPPM are correct
func TestGetPPMRegexPatterns(t *testing.T) {
	// Test the actual regex pattern used in getPPM
	pattern := "str?a?t?u?x:\\d+:?(-?\\d*)"
	r, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("Failed to compile getPPM regex: %v", err)
	}

	tests := []struct {
		name       string
		serial     string
		shouldFind bool
		wantPPM    string
	}{
		{
			name:       "Full stratux with PPM",
			serial:     "stratux:978:52",
			shouldFind: true,
			wantPPM:    "52",
		},
		{
			name:       "stratux without PPM",
			serial:     "stratux:978",
			shouldFind: true,
			wantPPM:    "",
		},
		{
			name:       "stratux with colon but no PPM",
			serial:     "stratux:978:",
			shouldFind: true,
			wantPPM:    "",
		},
		{
			name:       "Negative PPM",
			serial:     "stratux:1090:-15",
			shouldFind: true,
			wantPPM:    "-15",
		},
		{
			name:       "Invalid sratux format",
			serial:     "sratux:868:10",
			shouldFind: false,
			wantPPM:    "",
		},
		{
			name:       "Invalid tatux format",
			serial:     "tatux:162:5",
			shouldFind: false,
			wantPPM:    "",
		},
		{
			name:       "Invalid tux format",
			serial:     "tux:978:7",
			shouldFind: false,
			wantPPM:    "",
		},
		{
			name:       "Plain serial",
			serial:     "00000001",
			shouldFind: false,
			wantPPM:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := r.FindStringSubmatch(tt.serial)
			found := matches != nil

			if found != tt.shouldFind {
				t.Errorf("Pattern match for %q: found=%v, want=%v", tt.serial, found, tt.shouldFind)
				return
			}

			if found && matches[1] != tt.wantPPM {
				t.Errorf("PPM extraction from %q: got=%q, want=%q", tt.serial, matches[1], tt.wantPPM)
			}
		})
	}
}

// TestGetPPMEdgeCases tests additional edge cases for getPPM to achieve 100% coverage
func TestGetPPMEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		serial   string
		expected int
	}{
		{
			// Test when regex compile fails (simulated by invalid pattern in actual code)
			// This case is already handled by returning globalSettings.PPM
			name:     "Valid serial with PPM after successful regex compile",
			serial:   "stratux:978:42",
			expected: 42,
		},
		{
			// Test when FindStringSubmatch returns nil (no match)
			name:     "Serial that doesn't match regex pattern",
			serial:   "random-serial-number",
			expected: globalSettings.PPM,
		},
		{
			// Test when Atoi fails on the PPM value
			name:     "Serial with non-numeric PPM",
			serial:   "stratux:978:not-a-number",
			expected: globalSettings.PPM,
		},
		{
			// Test edge case with just minus sign
			name:     "Serial with just minus sign as PPM",
			serial:   "stratux:978:-",
			expected: globalSettings.PPM,
		},
		{
			// Test edge case with plus sign (strconv.Atoi fails on leading +)
			name:     "Serial with plus sign in PPM",
			serial:   "stratux:978:+10",
			expected: globalSettings.PPM,
		},
		{
			// Test very long PPM string that causes Atoi to fail
			name:     "Serial with extremely long PPM value",
			serial:   "stratux:978:12345678901234567890123456789012345678901234567890",
			expected: globalSettings.PPM,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPPM(tt.serial)
			if result != tt.expected {
				t.Errorf("getPPM(%q) = %d, want %d", tt.serial, result, tt.expected)
			}
		})
	}
}

// TestSDrKillBasic tests the sdrKill function in a basic scenario
func TestSDrKillBasic(t *testing.T) {
	// Save original state
	originalShutdown := sdrShutdown
	originalUATDev := UATDev
	originalESDev := ESDev
	originalOGNDev := OGNDev
	originalAISDev := AISDev

	// Reset to known state
	sdrShutdown = false
	UATDev = nil
	ESDev = nil
	OGNDev = nil
	AISDev = nil

	// Restore original state after test
	defer func() {
		sdrShutdown = originalShutdown
		UATDev = originalUATDev
		ESDev = originalESDev
		OGNDev = originalOGNDev
		AISDev = originalAISDev
	}()

	// Test with no devices initialized
	sdrKill()

	if !sdrShutdown {
		t.Error("sdrKill() should set sdrShutdown to true")
	}

	// Verify all devices are still nil
	if UATDev != nil || ESDev != nil || OGNDev != nil || AISDev != nil {
		t.Error("All devices should remain nil when none were initialized")
	}
}

// TestReCompileValidPatterns tests reCompile with various valid patterns
func TestReCompileValidPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		testStr string
		wantNil bool
		matches bool
	}{
		{
			name:    "UAT pattern matches correctly",
			pattern: "str?a?t?u?x:978",
			testStr: "stratux:978",
			wantNil: false,
			matches: true,
		},
		{
			name:    "ES pattern matches correctly",
			pattern: "str?a?t?u?x:1090",
			testStr: "stratux:1090",
			wantNil: false,
			matches: true,
		},
		{
			name:    "Word boundary pattern",
			pattern: "\\btest\\b",
			testStr: "test",
			wantNil: false,
			matches: true,
		},
		{
			name:    "Character class pattern",
			pattern: "[a-z]+",
			testStr: "abc",
			wantNil: false,
			matches: true,
		},
		{
			name:    "Anchored pattern",
			pattern: "^start.*end$",
			testStr: "start middle end",
			wantNil: false,
			matches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := reCompile(tt.pattern)
			if (r == nil) != tt.wantNil {
				t.Errorf("reCompile(%q) nil=%v, want nil=%v", tt.pattern, r == nil, tt.wantNil)
				return
			}

			if !tt.wantNil && r != nil {
				matched := r.MatchString(tt.testStr)
				if matched != tt.matches {
					t.Errorf("reCompile(%q).MatchString(%q) = %v, want %v", tt.pattern, tt.testStr, matched, tt.matches)
				}
			}
		})
	}
}

// TestReCompileInvalidPatterns tests reCompile with invalid patterns
func TestReCompileInvalidPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{
			name:    "Unclosed character class",
			pattern: "[abc",
		},
		{
			name:    "Unclosed group",
			pattern: "(abc",
		},
		{
			name:    "Invalid escape sequence",
			pattern: "\\k",
		},
		{
			name:    "Invalid repetition",
			pattern: "*abc",
		},
		{
			name:    "Unclosed named group",
			pattern: "(?P<name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := reCompile(tt.pattern)
			// For invalid patterns, reCompile should return nil
			if r != nil {
				t.Errorf("reCompile(%q) should return nil for invalid pattern, got non-nil", tt.pattern)
			}
		})
	}
}

// Benchmark tests for performance-critical functions
func BenchmarkGetPPM(b *testing.B) {
	serials := []string{
		"stratux:978:52",
		"stratux:1090:-15",
		"sratux:868:0",
		"00000001",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serial := serials[i%len(serials)]
		getPPM(serial)
	}
}

func BenchmarkRegexHasID(b *testing.B) {
	serials := []string{
		"stratux:978:52",
		"stratux:1090",
		"sratux:868",
		"00000001",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serial := serials[i%len(serials)]
		rUAT.hasID(serial)
	}
}

func BenchmarkReCompile(b *testing.B) {
	patterns := []string{
		"str?a?t?u?x:978",
		"str?a?t?u?x:1090",
		"\\d+",
		"test.*pattern",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern := patterns[i%len(patterns)]
		reCompile(pattern)
	}
}

// initHandleUatMessageTest initializes globals and returns a cleanup function
func initHandleUatMessageTest() func() {
	// Save original state
	originalClock := stratuxClock
	originalMaxSignalStrength := maxSignalStrength
	originalGlobalStatus := globalStatus
	originalGlobalSettings := globalSettings
	originalNetworkGDL90Chan := networkGDL90Chan
	originalNetMutex := netMutex
	originalClientConnections := clientConnections
	originalTrafficMutex := trafficMutex
	originalTraffic := traffic
	originalSeenTraffic := seenTraffic

	// Initialize required globals
	stratuxClock = NewMonotonic()
	time.Sleep(10 * time.Millisecond)         // Let the clock start
	networkGDL90Chan = make(chan []byte, 100) // Buffered channel to prevent blocking
	netMutex = &sync.Mutex{}
	clientConnections = make(map[string]connection)
	trafficMutex = &sync.Mutex{}
	traffic = make(map[uint32]TrafficInfo)
	seenTraffic = make(map[uint32]bool)

	// Return cleanup function
	return func() {
		stratuxClock = originalClock
		maxSignalStrength = originalMaxSignalStrength
		globalStatus = originalGlobalStatus
		globalSettings = originalGlobalSettings
		networkGDL90Chan = originalNetworkGDL90Chan
		netMutex = originalNetMutex
		clientConnections = originalClientConnections
		trafficMutex = originalTrafficMutex
		traffic = originalTraffic
		seenTraffic = originalSeenTraffic
	}
}

// TestHandleUatMessage tests the handleUatMessage function
func TestHandleUatMessage(t *testing.T) {
	defer initHandleUatMessageTest()()

	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "Valid UAT uplink message",
			input:       "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5;ss=1000",
			description: "Valid uplink message with signal strength",
		},
		{
			name:        "Valid UAT basic downlink message",
			input:       "-" + strings.Repeat("AA", 18) + ";rs=3;ss=500",
			description: "Valid basic downlink (18 bytes) with signal strength",
		},
		{
			name:        "Valid UAT long downlink message (34 bytes)",
			input:       "-" + strings.Repeat("BB", 34) + ";rs=4;ss=750",
			description: "Valid long downlink (34 bytes) with signal strength",
		},
		{
			name:        "Valid UAT long downlink message (48 bytes)",
			input:       "-" + strings.Repeat("CC", 48) + ";rs=6;ss=900",
			description: "Valid long downlink (48 bytes with Reed Solomon) with signal strength",
		},
		{
			name:        "Empty message",
			input:       "",
			description: "Empty string should be handled gracefully",
		},
		{
			name:        "Message without signal strength",
			input:       "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5",
			description: "Uplink message without ss= field",
		},
		{
			name:        "Downlink without signal strength",
			input:       "-" + strings.Repeat("DD", 18) + ";rs=2",
			description: "Downlink message without ss= field",
		},
		{
			name:        "Uplink with invalid signal strength",
			input:       "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5;ss=invalid",
			description: "Uplink with non-numeric signal strength",
		},
		{
			name:        "Downlink with high signal strength",
			input:       "-" + strings.Repeat("FF", 18) + ";rs=7;ss=9999",
			description: "Downlink with very high signal strength",
		},
		{
			name:        "Uplink with zero signal strength",
			input:       "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=1;ss=0",
			description: "Uplink with zero signal strength",
		},
		{
			name:        "Short uplink (should be padded)",
			input:       "+" + strings.Repeat("11", 100) + ";rs=4;ss=800",
			description: "Short uplink message that gets padded",
		},
		{
			name:        "Message with only prefix",
			input:       "+",
			description: "Message with only prefix character",
		},
		{
			name:        "Message with semicolon but no data",
			input:       ";",
			description: "Message starting with semicolon",
		},
		{
			name:        "Downlink message with no semicolon",
			input:       "-" + strings.Repeat("22", 18),
			description: "Downlink without semicolon separator",
		},
		{
			name:        "Uplink message with multiple semicolons",
			input:       "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5;ss=1000;extra=data",
			description: "Uplink with extra fields after semicolons",
		},
		{
			name:        "Message without ss prefix in third field",
			input:       "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5;rs=1000",
			description: "Uplink without proper ss= prefix in signal field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state before each test
			globalStatus.UAT_messages_total = 0
			maxSignalStrength = 0

			// Call handleUatMessage - should not panic
			handleUatMessage(tt.input)

			// Verify UAT_messages_total was incremented (except for empty message)
			if tt.input != "" && tt.input != ";" {
				if globalStatus.UAT_messages_total == 0 {
					t.Errorf("Expected UAT_messages_total to be incremented for non-empty message")
				}
			}
		})
	}
}

// TestHandleUatMessageWithValidMessages tests handleUatMessage with messages that should be relayed
func TestHandleUatMessageWithValidMessages(t *testing.T) {
	defer initHandleUatMessageTest()()

	tests := []struct {
		name           string
		input          string
		expectedRelay  bool
		expectedSignal int
	}{
		{
			name:           "Valid uplink with signal",
			input:          "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5;ss=5000",
			expectedRelay:  true,
			expectedSignal: 5000,
		},
		{
			name:           "Valid basic report",
			input:          "-" + strings.Repeat("AA", 18) + ";rs=3;ss=2000",
			expectedRelay:  true,
			expectedSignal: 2000,
		},
		{
			name:          "Empty input",
			input:         "",
			expectedRelay: false,
		},
		{
			name:          "Invalid format",
			input:         "invalid",
			expectedRelay: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			globalStatus.UAT_messages_total = 0
			maxSignalStrength = 0

			// Call the function
			handleUatMessage(tt.input)

			// For valid messages, check that maxSignalStrength was updated
			if tt.expectedRelay && tt.input[0] == '+' && tt.expectedSignal > 0 {
				if maxSignalStrength != tt.expectedSignal {
					t.Errorf("Expected maxSignalStrength=%d, got %d", tt.expectedSignal, maxSignalStrength)
				}
			}
		})
	}
}

// TestHandleUatMessageErrorPaths tests error handling in handleUatMessage
func TestHandleUatMessageErrorPaths(t *testing.T) {
	defer initHandleUatMessageTest()()

	errorCases := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "Empty string",
			input:       "",
			description: "Empty string returns nil from parseInput",
		},
		{
			name:        "Just semicolon",
			input:       ";",
			description: "Just semicolon should be handled",
		},
		{
			name:        "Invalid format",
			input:       "invalid",
			description: "Invalid format should be handled gracefully",
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state
			globalStatus.UAT_messages_total = 0

			// Should not panic
			handleUatMessage(tc.input)

			// Empty string won't increment counter
			if tc.input == "" || tc.input == ";" {
				if globalStatus.UAT_messages_total != 0 {
					t.Errorf("Empty/semicolon message should not increment UAT_messages_total")
				}
			}
		})
	}
}

// TestHandleUatMessageSignalStrengthTracking tests signal strength handling
func TestHandleUatMessageSignalStrengthTracking(t *testing.T) {
	defer initHandleUatMessageTest()()

	// Reset max signal strength
	maxSignalStrength = 0

	// Test increasing signal strength
	handleUatMessage("+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5;ss=1000")
	if maxSignalStrength != 1000 {
		t.Errorf("Expected maxSignalStrength=1000, got %d", maxSignalStrength)
	}

	// Test higher signal strength (should update)
	handleUatMessage("+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5;ss=2000")
	if maxSignalStrength != 2000 {
		t.Errorf("Expected maxSignalStrength=2000, got %d", maxSignalStrength)
	}

	// Test lower signal strength (should NOT update)
	handleUatMessage("+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5;ss=1500")
	if maxSignalStrength != 2000 {
		t.Errorf("Expected maxSignalStrength to remain 2000, got %d", maxSignalStrength)
	}

	// Test downlink message (should NOT update maxSignalStrength)
	handleUatMessage("-" + strings.Repeat("AA", 18) + ";rs=3;ss=5000")
	if maxSignalStrength != 2000 {
		t.Errorf("Downlink should not update maxSignalStrength, expected 2000, got %d", maxSignalStrength)
	}

	// Test zero signal strength
	prevMax := maxSignalStrength
	handleUatMessage("+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES) + ";rs=5;ss=0")
	if maxSignalStrength != prevMax {
		t.Errorf("Zero signal strength should not update max, expected %d, got %d", prevMax, maxSignalStrength)
	}
}

// TestHandleUatMessageMessageTypes tests different UAT message types
func TestHandleUatMessageMessageTypes(t *testing.T) {
	defer initHandleUatMessageTest()()

	messageTypes := []struct {
		name     string
		input    string
		msgBytes int
	}{
		{
			name:     "MSGTYPE_UPLINK (432 bytes)",
			input:    "+" + strings.Repeat("00", UPLINK_FRAME_DATA_BYTES),
			msgBytes: UPLINK_FRAME_DATA_BYTES,
		},
		{
			name:     "MSGTYPE_BASIC_REPORT (18 bytes)",
			input:    "-" + strings.Repeat("AA", 18),
			msgBytes: 18,
		},
		{
			name:     "MSGTYPE_LONG_REPORT (34 bytes)",
			input:    "-" + strings.Repeat("BB", 34),
			msgBytes: 34,
		},
		{
			name:     "MSGTYPE_LONG_REPORT (48 bytes)",
			input:    "-" + strings.Repeat("CC", 48),
			msgBytes: 48,
		},
	}

	for _, mt := range messageTypes {
		t.Run(mt.name, func(t *testing.T) {
			globalStatus.UAT_messages_total = 0

			// Add signal strength field
			fullInput := mt.input + ";rs=5;ss=1000"

			// Should not panic and should process the message
			handleUatMessage(fullInput)

			if globalStatus.UAT_messages_total == 0 {
				t.Errorf("Expected UAT_messages_total to be incremented")
			}
		})
	}
}
