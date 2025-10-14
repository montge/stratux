//go:build ignore
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
	// Create sample APRS trace file
	fname := "basic_aprs.trace.gz"
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

	// Sample APRS messages from OGN APRS-IS network
	// Format: time offset in ms, APRS text message
	// APRS format: PROTOCOL+ID>APRS,qAS,STATION:/HHMMSSh+DDMM.MMMx/DDDMM.MMMxTTTT/SSS/A=AAAAAA !WPP! idXXXXXXXX
	// Where: x = N/S/E/W, TTTT = track, SSS = speed (knots), AAAAAA = altitude (feet)
	messages := []struct {
		offsetMs int64
		data     string
	}{
		// FLARM message 1 - Oxford area (51.7657°N = 5145.94'N, 1.1918°W = 00111.51'W)
		// Track: 057°, Speed: 057 knots, Altitude: 407 feet
		{0, `FLR395F39>APRS,qAS,OXFORD:/120000h5145.945N/00111.511W'057/057/A=000407 !W02! id06395F39`},

		// OGN message 1 - Moving north
		{500, `OGN395F39>APRS,qAS,OXFORD:/120000h5146.021N/00111.537W'057/062/A=000415 !W12! id0D395F39`},

		// ICAO message 1 - Commercial aircraft
		{1000, `ICADD4B12>APRS,qAS,OXFORD:/120001h5146.206N/00111.674W'124/099/A=000478 !W25! id10DD4B12`},

		// Skylines message
		{1500, `SKYA12345>APRS,qAS,OXFORD:/120001h5145.738N/00111.333W'215/073/A=000433 !W18! id02A12345`},

		// PAW (PilotAware) message
		{2000, `PAW123ABC>APRS,qAS,OXFORD:/120002h5145.252N/00110.668W'045/110/A=000476 !W35! id03123ABC`},

		// Message without optional fields (no track/speed/altitude)
		{2500, `FLR395F40>APRS,qAS,OXFORD:/120002h5145.534N/00111.004W' !W02!`},

		// FLARM message with different precision
		{3000, `FLR395F41>APRS,qAS,OXFORD:/120003h5145.006N/00110.338W'180/079/A=000455 !W05! id06395F41`},

		// OGN message with high precision
		{3500, `OGN395F42>APRS,qAS,OXFORD:/120003h5144.790N/00110.002W'270/074/A=000500 !W52! id1A395F42`},

		// Message from FAN (FANET) protocol
		{4000, `FAN234567>APRS,qAS,OXFORD:/120004h5144.501N/00109.667W'090/065/A=000521 !W15! id05234567`},

		// Invalid message (should be ignored) - missing required fields
		{4500, `INVALID>MESSAGE`},

		// Ground station beacon (TCPIP*,qAC - should be ignored by parseAprsMessage)
		{5000, `OXFORD>APRS,TCPIP*,qAC,GLIDERN1:/120005h5146.000N/00112.000W'`},

		// Message with minimal track/speed
		{5500, `FLR395F43>APRS,qAS,OXFORD:/120005h5144.234N/00109.334W'000/001/A=000538 !W02! id06395F43`},
	}

	// Write messages
	for _, msg := range messages {
		ts := baseTime.Add(time.Duration(msg.offsetMs) * time.Millisecond)
		err := w.Write([]string{
			ts.Format(time.RFC3339Nano),
			"aprs",
			msg.data,
		})
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Created %s with %d messages\n", fname, len(messages))
}
