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
	// Create sample GPS NMEA trace file
	fname := "basic_gps.trace.gz"
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
	baseTime := time.Date(2025, 10, 13, 12, 0, 0, 0, time.UTC)

	// Sample GPS NMEA messages - simulating aircraft at Seattle area
	// Format: time offset in ms, NMEA sentence
	// Note: Checksums calculated correctly using XOR of all characters between $ and *
	messages := []struct {
		offsetMs int64
		data     string
	}{
		// RMC - Recommended Minimum Navigation Information
		{0, "$GPRMC,120000.000,A,4727.030,N,12218.528,W,057.9,349.7,131025,015.0,E*79"},
		{100, "$GPGGA,120000.000,4727.030,N,12218.528,W,1,08,0.9,420.9,M,46.9,M,,*4A"},
		{200, "$GPGSA,A,3,01,02,03,04,05,06,07,08,,,,,2.0,0.9,1.8*38"},
		{300, "$GPGSV,3,1,12,01,85,045,45,02,65,135,42,03,55,225,40,04,45,315,38*7F"},

		{1000, "$GPRMC,120001.000,A,4727.031,N,12218.530,W,057.9,349.7,131025,015.0,E*70"},
		{1100, "$GPGGA,120001.000,4727.031,N,12218.530,W,1,08,0.9,421.0,M,46.9,M,,*4B"},

		{2000, "$GPRMC,120002.000,A,4727.032,N,12218.532,W,057.9,349.7,131025,015.0,E*72"},
		{2100, "$GPGGA,120002.000,4727.032,N,12218.532,W,1,08,0.9,421.1,M,46.9,M,,*48"},

		{3000, "$GPRMC,120003.000,A,4727.033,N,12218.534,W,057.9,349.7,131025,015.0,E*74"},
		{3100, "$GPGGA,120003.000,4727.033,N,12218.534,W,1,08,0.9,421.2,M,46.9,M,,*4D"},
	}

	// Write messages
	for _, msg := range messages {
		ts := baseTime.Add(time.Duration(msg.offsetMs) * time.Millisecond)
		err := w.Write([]string{
			ts.Format(time.RFC3339Nano),
			"nmea",
			msg.data,
		})
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Created %s with %d messages\n", fname, len(messages))
}
