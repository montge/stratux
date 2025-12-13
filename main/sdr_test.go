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
	"testing"
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
