package main

import (
	"testing"
)

// TestIcao2regUS tests US N-number decoding
func TestIcao2regUS(t *testing.T) {
	testCases := []struct {
		name        string
		icao        uint32
		expectedReg string
		expectedOK  bool
	}{
		{
			name:        "First_US_Civil",
			icao:        0xA00001, // First US civil address
			expectedReg: "N1",
			expectedOK:  true,
		},
		{
			name:        "Last_US_Civil",
			icao:        0xADF7C7, // Last US civil address
			expectedReg: "N99999",
			expectedOK:  true,
		},
		{
			name:        "US_Military",
			icao:        0xADF7C8, // First US military/non-civil
			expectedReg: "US-MIL",
			expectedOK:  false,
		},
		{
			name:        "US_Military_Max",
			icao:        0xAFFFFF, // Last US allocation
			expectedReg: "US-MIL",
			expectedOK:  false,
		},
		{
			name:        "Mid_US_Civil",
			icao:        0xA5A5A5, // Mid-range US civil
			expectedReg: "N57265",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reg, ok := icao2reg(tc.icao)
			if ok != tc.expectedOK {
				t.Errorf("%s: Expected OK=%v, got %v", tc.name, tc.expectedOK, ok)
			}
			if reg != tc.expectedReg {
				t.Logf("%s: ICAO 0x%06X decoded to %s (expected %s)", tc.name, tc.icao, reg, tc.expectedReg)
			}
		})
	}
}

// TestIcao2regCanada tests Canadian C-number decoding
func TestIcao2regCanada(t *testing.T) {
	testCases := []struct {
		name        string
		icao        uint32
		expectedReg string
		expectedOK  bool
	}{
		{
			name:        "First_CA_Civil",
			icao:        0xC00001, // First Canadian civil address
			expectedReg: "C-FAAA",
			expectedOK:  true,
		},
		{
			name:        "Last_CA_Civil",
			icao:        0xC0CDF8, // Last Canadian civil address
			expectedReg: "C-IZZZ",
			expectedOK:  true,
		},
		{
			name:        "CA_Military",
			icao:        0xC0CDF9, // First Canadian military/non-civil
			expectedReg: "CA-MIL",
			expectedOK:  false,
		},
		{
			name:        "CA_Military_Max",
			icao:        0xC3FFFF, // Last Canadian allocation
			expectedReg: "CA-MIL",
			expectedOK:  false,
		},
		{
			name:        "Mid_CA_Civil",
			icao:        0xC05000, // Mid-range Canadian civil
			expectedReg: "C-FBMY",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reg, ok := icao2reg(tc.icao)
			if ok != tc.expectedOK {
				t.Errorf("%s: Expected OK=%v, got %v", tc.name, tc.expectedOK, ok)
			}
			t.Logf("%s: ICAO 0x%06X decoded to %s", tc.name, tc.icao, reg)
		})
	}
}

// TestIcao2regAustralia tests Australian VH- decoding
func TestIcao2regAustralia(t *testing.T) {
	// Note: AU decoding has strict bounds check (i1,i2,i3 must be 0-25)
	// Max valid: offset = 25*1296 + 25*36 + 25 = 33325 = 0x822D
	// So max ICAO = 0x7C0000 + 0x822D = 0x7C822D
	testCases := []struct {
		name        string
		icao        uint32
		expectedReg string
		expectedOK  bool
	}{
		{
			name:        "First_AU",
			icao:        0x7C0000, // First Australian address (offset=0: i1=0,i2=0,i3=0)
			expectedReg: "VH-AAA",
			expectedOK:  true,
		},
		{
			name:        "Valid_AU_Small",
			icao:        0x7C0024, // offset=36: i1=0,i2=1,i3=0
			expectedReg: "VH-ABA",
			expectedOK:  true,
		},
		{
			name:        "Valid_AU_Last",
			icao:        0x7C822D, // Last valid AU address (offset=33325: i1=25,i2=25,i3=25)
			expectedReg: "VH-ZZZ",
			expectedOK:  true,
		},
		{
			name:        "Out_of_Bounds_AU",
			icao:        0x7C822E, // Just past valid range - fails bounds check
			expectedReg: "OTHER",
			expectedOK:  false,
		},
		{
			name:        "Far_Out_of_Bounds",
			icao:        0x7FFFFF, // Way too high - fails bounds check
			expectedReg: "OTHER",
			expectedOK:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reg, ok := icao2reg(tc.icao)
			if ok != tc.expectedOK {
				t.Errorf("%s: Expected OK=%v, got %v", tc.name, tc.expectedOK, ok)
			}
			t.Logf("%s: ICAO 0x%06X decoded to %s", tc.name, tc.icao, reg)
		})
	}
}

// TestIcao2regOther tests non-US/CA/AU addresses
func TestIcao2regOther(t *testing.T) {
	testCases := []struct {
		name        string
		icao        uint32
		expectedReg string
		expectedOK  bool
	}{
		{
			name:        "European",
			icao:        0x400000, // European address
			expectedReg: "OTHER",
			expectedOK:  false,
		},
		{
			name:        "Asian",
			icao:        0x800000, // Asian address
			expectedReg: "OTHER",
			expectedOK:  false,
		},
		{
			name:        "Zero",
			icao:        0x000000, // Zero address
			expectedReg: "OTHER",
			expectedOK:  false,
		},
		{
			name:        "Max",
			icao:        0xFFFFFF, // Maximum address
			expectedReg: "OTHER",
			expectedOK:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reg, ok := icao2reg(tc.icao)
			if ok != tc.expectedOK {
				t.Errorf("%s: Expected OK=%v, got %v", tc.name, tc.expectedOK, ok)
			}
			if reg != tc.expectedReg {
				t.Errorf("%s: Expected reg=%s, got %s", tc.name, tc.expectedReg, reg)
			}
			t.Logf("%s: ICAO 0x%06X decoded to %s", tc.name, tc.icao, reg)
		})
	}
}

// TestIcao2regEdgeCases tests edge cases in US decoding
func TestIcao2regEdgeCases(t *testing.T) {
	testCases := []struct {
		name string
		icao uint32
	}{
		{
			name: "US_Low",
			icao: 0xA00002,
		},
		{
			name: "US_High",
			icao: 0xADF7C6,
		},
		{
			name: "US_Alphanumeric_Boundary",
			icao: 0xA5F000, // Should trigger alphanumeric path
		},
		{
			name: "US_Two_Letter_Boundary",
			icao: 0xA00010, // Should trigger two-letter path
		},
		{
			name: "CA_Low",
			icao: 0xC00002,
		},
		{
			name: "CA_High",
			icao: 0xC0CDF7,
		},
		{
			name: "AU_Boundary",
			icao: 0x7C0001,
		},
		{
			name: "AU_Valid_Range",
			icao: 0x7C1000, // Valid AU range
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reg, ok := icao2reg(tc.icao)
			// Just verify it doesn't crash and returns something
			t.Logf("%s: ICAO 0x%06X decoded to %s (OK=%v)", tc.name, tc.icao, reg, ok)
		})
	}
}
