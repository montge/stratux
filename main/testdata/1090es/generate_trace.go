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
	// Create sample 1090ES trace file with dump1090 JSON format
	fname := "basic_1090es.trace.gz"
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

	// Sample dump1090 JSON messages
	// Format: dump1090-mutability JSON output on port 30006
	// Field names must match the dump1090Data struct exactly (capitalized, with underscores)
	// Note: Use Icao_addr, TypeCode (not type_code), SignalLevel (not signal_level), etc.
	messages := []struct {
		offsetMs int64
		data     string
	}{
		// ADS-B position message (DF17, TypeCode 11)
		{0, `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7657,"Lng":-1.1918,"Alt":5850,"Position_valid":true,"SignalLevel":0.0512,"Timestamp":"2025-10-14T12:00:00.000Z"}`},

		// Velocity message (DF17, TypeCode 19)
		{500, `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":19,"Speed":468,"Track":89,"Vvel":-64,"Speed_valid":true,"SignalLevel":0.0498,"Timestamp":"2025-10-14T12:00:00.500Z"}`},

		// Another aircraft - position
		{1000, `{"Icao_addr":10685854,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7542,"Lng":-1.1778,"Alt":8525,"Position_valid":true,"SignalLevel":0.0645,"Timestamp":"2025-10-14T12:00:01.000Z"}`},

		// Callsign/identification message (DF17, TypeCode 4)
		{1500, `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":4,"Tail":"EZY123  ","SignalLevel":0.0502,"Timestamp":"2025-10-14T12:00:01.500Z"}`},

		// TIS-B message (DF18, CA=2)
		{2000, `{"Icao_addr":2893118,"DF":18,"CA":2,"TypeCode":11,"Lat":51.7623,"Lng":-1.1889,"Alt":3200,"Position_valid":true,"SignalLevel":0.0512,"Timestamp":"2025-10-14T12:00:02.000Z"}`},

		// ADS-R message (DF18, CA=6)
		{2500, `{"Icao_addr":11230840,"DF":18,"CA":6,"TypeCode":11,"Lat":51.7589,"Lng":-1.1834,"Alt":4100,"Position_valid":true,"SignalLevel":0.0498,"Timestamp":"2025-10-14T12:00:02.500Z"}`},

		// Surveillance altitude reply (DF4)
		{3000, `{"Icao_addr":11230838,"DF":4,"Alt":5875,"SignalLevel":0.0425,"Timestamp":"2025-10-14T12:00:03.000Z"}`},

		// Mode S all-call reply (DF11)
		{3500, `{"Icao_addr":10685854,"DF":11,"CA":5,"SignalLevel":0.0312,"Timestamp":"2025-10-14T12:00:03.500Z"}`},

		// Message with squawk code (DF5)
		{4000, `{"Icao_addr":11230838,"DF":5,"Squawk":7700,"Alt":5900,"SignalLevel":0.0398,"Timestamp":"2025-10-14T12:00:04.000Z"}`},

		// High altitude aircraft
		{4500, `{"Icao_addr":11184758,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7465,"Lng":-1.1667,"Alt":35000,"Position_valid":true,"SignalLevel":0.0198,"Timestamp":"2025-10-14T12:00:04.500Z"}`},

		// Velocity for high altitude aircraft
		{5000, `{"Icao_addr":11184758,"DF":17,"CA":5,"TypeCode":19,"Speed":468,"Track":89,"Vvel":-64,"Speed_valid":true,"SignalLevel":0.0201,"Timestamp":"2025-10-14T12:00:05.000Z"}`},

		// On-ground position (DF17, with OnGround flag)
		{5500, `{"Icao_addr":10500126,"DF":17,"CA":5,"TypeCode":8,"Lat":51.7501,"Lng":-1.1723,"Alt":500,"OnGround":true,"Position_valid":true,"SignalLevel":0.0812,"Timestamp":"2025-10-14T12:00:05.500Z"}`},

		// NACp (navigation accuracy) example
		{6000, `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7660,"Lng":-1.1920,"Alt":5925,"NACp":8,"Position_valid":true,"SignalLevel":0.0515,"Timestamp":"2025-10-14T12:00:06.000Z"}`},

		// Emitter category example
		{6500, `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":4,"Tail":"EZY123  ","Emitter_category":7,"SignalLevel":0.0508,"Timestamp":"2025-10-14T12:00:06.500Z"}`},

		// GNSS altitude difference example
		{7000, `{"Icao_addr":11230838,"DF":17,"CA":5,"TypeCode":11,"Lat":51.7665,"Lng":-1.1925,"Alt":5950,"GnssDiffFromBaroAlt":150,"Position_valid":true,"SignalLevel":0.0501,"Timestamp":"2025-10-14T12:00:07.000Z"}`},
	}

	// Write messages
	for _, msg := range messages {
		ts := baseTime.Add(time.Duration(msg.offsetMs) * time.Millisecond)
		err := w.Write([]string{
			ts.Format(time.RFC3339Nano),
			"dump1090",
			msg.data,
		})
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Created %s with %d messages\\n", fname, len(messages))
}
