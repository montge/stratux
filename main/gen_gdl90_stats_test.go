package main

import (
	"testing"
)

// TestUpdateUATStats tests the UAT product statistics tracking
func TestUpdateUATStats(t *testing.T) {
	tests := []struct {
		name       string
		productID  uint32
		checkField string
		checkFunc  func() uint32
	}{
		// METAR products
		{"METAR ID 0", 0, "UAT_METAR_total", func() uint32 { return globalStatus.UAT_METAR_total }},
		{"METAR ID 20", 20, "UAT_METAR_total", func() uint32 { return globalStatus.UAT_METAR_total }},

		// TAF products
		{"TAF ID 1", 1, "UAT_TAF_total", func() uint32 { return globalStatus.UAT_TAF_total }},
		{"TAF ID 21", 21, "UAT_TAF_total", func() uint32 { return globalStatus.UAT_TAF_total }},

		// NEXRAD products (comprehensive list)
		{"NEXRAD ID 51", 51, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 52", 52, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 53", 53, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 54", 54, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 55", 55, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 56", 56, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 57", 57, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 58", 58, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 59", 59, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 60", 60, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 61", 61, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 62", 62, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 63", 63, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 64", 64, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 81", 81, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 82", 82, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},
		{"NEXRAD ID 83", 83, "UAT_NEXRAD_total", func() uint32 { return globalStatus.UAT_NEXRAD_total }},

		// SIGMET/AIRMET products
		{"SIGMET ID 2", 2, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 3", 3, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 4", 4, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 6", 6, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 11", 11, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 12", 12, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 22", 22, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 23", 23, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 24", 24, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 26", 26, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},
		{"SIGMET ID 254", 254, "UAT_SIGMET_total", func() uint32 { return globalStatus.UAT_SIGMET_total }},

		// PIREP products
		{"PIREP ID 5", 5, "UAT_PIREP_total", func() uint32 { return globalStatus.UAT_PIREP_total }},
		{"PIREP ID 25", 25, "UAT_PIREP_total", func() uint32 { return globalStatus.UAT_PIREP_total }},

		// NOTAM product
		{"NOTAM ID 8", 8, "UAT_NOTAM_total", func() uint32 { return globalStatus.UAT_NOTAM_total }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset counter
			before := tc.checkFunc()

			// Call function
			UpdateUATStats(tc.productID)

			// Verify counter increased
			after := tc.checkFunc()
			if after != before+1 {
				t.Errorf("Expected %s to increment from %d to %d, got %d",
					tc.checkField, before, before+1, after)
			}
		})
	}
}

// TestUpdateUATStatsSpecialCases tests special handling for product 413 and unknown products
func TestUpdateUATStatsSpecialCases(t *testing.T) {
	t.Run("Product 413 no-op", func(t *testing.T) {
		// Product 413 should not increment any counter (early return)
		beforeMETAR := globalStatus.UAT_METAR_total
		beforeTAF := globalStatus.UAT_TAF_total
		beforeNEXRAD := globalStatus.UAT_NEXRAD_total
		beforeSIGMET := globalStatus.UAT_SIGMET_total
		beforePIREP := globalStatus.UAT_PIREP_total
		beforeNOTAM := globalStatus.UAT_NOTAM_total
		beforeOTHER := globalStatus.UAT_OTHER_total

		UpdateUATStats(413)

		// Verify nothing changed
		if globalStatus.UAT_METAR_total != beforeMETAR ||
			globalStatus.UAT_TAF_total != beforeTAF ||
			globalStatus.UAT_NEXRAD_total != beforeNEXRAD ||
			globalStatus.UAT_SIGMET_total != beforeSIGMET ||
			globalStatus.UAT_PIREP_total != beforePIREP ||
			globalStatus.UAT_NOTAM_total != beforeNOTAM ||
			globalStatus.UAT_OTHER_total != beforeOTHER {
			t.Error("Product ID 413 should not increment any counters (early return)")
		}
	})

	t.Run("Unknown product defaults to OTHER", func(t *testing.T) {
		before := globalStatus.UAT_OTHER_total

		// Test several unknown product IDs
		unknownIDs := []uint32{9999, 10000, 100, 500, 12345}
		for _, id := range unknownIDs {
			UpdateUATStats(id)
		}

		after := globalStatus.UAT_OTHER_total
		expected := before + uint32(len(unknownIDs))

		if after != expected {
			t.Errorf("Expected UAT_OTHER_total to be %d, got %d", expected, after)
		}
	})
}
