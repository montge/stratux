// +build ignore

package main

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

func main() {
	// Create sample UAT 978MHz trace file
	fname := "basic_uat.trace.gz"
	f, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	w := csv.NewWriter(gzw)
	defer w.Flush()

	// Base time
	baseTime := time.Date(2025, 10, 14, 12, 0, 0, 0, time.UTC)

	// Sample UAT messages from real dump978 data
	// Format: time offset in ms, UAT message (+ = uplink, - = downlink)
	// Messages include signal strength after semicolon: ;rs=<rssi>;ss=<signal>
	messages := []struct {
		offsetMs int64
		data     string
	}{
		// UPLINK MESSAGE (432 bytes) - Weather/FIS-B data from ground station
		// This is a UAT uplink containing weather products
		{0, "+3cc0978aa66ca1a0158000213c5d2082102c22cc00082eec1e012c22cc000000000000000fd90007110e240811081ec5ea23b0c00;rs=16;ss=128"},

		// DOWNLINK MESSAGE (18 bytes) - BASIC_REPORT - Short position report from aircraft
		// This is an ADS-B aircraft position report (basic)
		{500, "-000000000000000000000000000000000000;rs=12;ss=94"},

		// UPLINK MESSAGE (432 bytes) - Another weather uplink
		{1000, "+3c62ab89c854b370308000353f59682210000000ff005685d07c4d5060cb9c72d35833db9e36df57f2d70d707d77d27f5e30c837;rs=17;ss=132"},

		// DOWNLINK MESSAGE (34 bytes) - LONG_REPORT - Extended position report from aircraft
		// This is an ADS-B aircraft position report with velocity
		{1500, "-0000000000000000000000000000000000000000000000000000000000000000000000;rs=14;ss=102"},

		// UPLINK MESSAGE (432 bytes) - FIS-B weather product
		{2000, "+3cc0978aa66cbaa05a8000213d99c822102cc04e0000aa88842c38f50136d1840e02cc04ebb5bf8df3de0cb2c72d776a0846103835;rs=18;ss=135"},

		// DOWNLINK MESSAGE (48 bytes) - LONG_REPORT with Reed Solomon
		// Extended report with error correction
		{2500, "-0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000;rs=15;ss=98"},

		// UPLINK MESSAGE (432 bytes) - More weather data
		{3000, "+3cc0978aa66ca2a0290000353f5d182210000000ff00476cf47c4d5060cb9c74cf5833df2cb4db77f2d30cb07d77c27c14b3c32a17;rs=19;ss=140"},

		// DOWNLINK MESSAGE (18 bytes) - Another basic report
		{3500, "-000000000000000000000000000000000001;rs=13;ss=96"},
	}

	// Write messages
	for _, msg := range messages {
		ts := baseTime.Add(time.Duration(msg.offsetMs) * time.Millisecond)
		err := w.Write([]string{
			ts.Format(time.RFC3339Nano),
			"uat",
			msg.data,
		})
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Created %s with %d messages\n", fname, len(messages))
}
