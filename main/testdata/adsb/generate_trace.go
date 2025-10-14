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
	// Create sample 1090ES trace file
	fname := "basic_adsb.trace.gz"
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

	// Sample aircraft: UAL123 flying at 35000 ft
	// Using Stratux's modified dump1090 JSON format (not standard dump1090 format)
	// Key differences: "Tail" not "flight", "Alt" not "alt_baro", "Speed" not "gs"
	messages := []struct {
		offsetMs int64
		data     string
	}{
		{0, `{"Icao_addr":10560325,"DF":17,"CA":0,"TypeCode":11,"SubtypeCode":0,"SignalLevel":0.5,"Tail":"UAL123","Alt":35000,"AltIsGNSS":false,"Speed_valid":true,"Speed":450,"Track":270,"Lat":47.4502,"Lng":-122.3088,"Position_valid":true,"Vvel":0,"OnGround":false,"NACp":10}`},
		{1000, `{"Icao_addr":10560325,"DF":17,"CA":0,"TypeCode":11,"SubtypeCode":0,"SignalLevel":0.6,"Tail":"UAL123","Alt":35000,"AltIsGNSS":false,"Speed_valid":true,"Speed":450,"Track":270,"Lat":47.4503,"Lng":-122.3188,"Position_valid":true,"Vvel":0,"OnGround":false,"NACp":10}`},
		{2000, `{"Icao_addr":10560325,"DF":17,"CA":0,"TypeCode":11,"SubtypeCode":0,"SignalLevel":0.4,"Tail":"UAL123","Alt":35000,"AltIsGNSS":false,"Speed_valid":true,"Speed":450,"Track":270,"Lat":47.4504,"Lng":-122.3288,"Position_valid":true,"Vvel":0,"OnGround":false,"NACp":10}`},
	}

	// Sample aircraft 2: N172SP - general aviation
	messages = append(messages, []struct {
		offsetMs int64
		data     string
	}{
		{500, `{"Icao_addr":11305708,"DF":17,"CA":0,"TypeCode":11,"SubtypeCode":0,"SignalLevel":0.8,"Tail":"N172SP","Alt":5500,"AltIsGNSS":false,"Speed_valid":true,"Speed":120,"Track":90,"Lat":47.4600,"Lng":-122.2900,"Position_valid":true,"Vvel":0,"OnGround":false,"NACp":9}`},
		{1500, `{"Icao_addr":11305708,"DF":17,"CA":0,"TypeCode":11,"SubtypeCode":0,"SignalLevel":0.9,"Tail":"N172SP","Alt":5520,"AltIsGNSS":false,"Speed_valid":true,"Speed":121,"Track":91,"Lat":47.4605,"Lng":-122.2800,"Position_valid":true,"Vvel":0,"OnGround":false,"NACp":9}`},
		{2500, `{"Icao_addr":11305708,"DF":17,"CA":0,"TypeCode":11,"SubtypeCode":0,"SignalLevel":0.7,"Tail":"N172SP","Alt":5540,"AltIsGNSS":false,"Speed_valid":true,"Speed":122,"Track":92,"Lat":47.4610,"Lng":-122.2700,"Position_valid":true,"Vvel":0,"OnGround":false,"NACp":9}`},
	}...)

	// Sort by offsetMs
	for i := 0; i < len(messages)-1; i++ {
		for j := i + 1; j < len(messages); j++ {
			if messages[i].offsetMs > messages[j].offsetMs {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
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

	fmt.Printf("Created %s with %d messages\n", fname, len(messages))
}
