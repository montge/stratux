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
	messages := []struct {
		offsetMs int64
		data     string
	}{
		{0, `{"hex":"A12345","flight":"UAL123  ","alt_baro":35000,"gs":450,"track":270,"lat":47.4502,"lon":-122.3088,"nic":8,"rc":186,"seen_pos":0.1,"version":2,"nac_p":10,"nac_v":2,"sil":3,"sil_type":"perhour","gva":2,"sda":2,"mlat":[],"tisb":[],"messages":150,"seen":0.1,"rssi":-25.5}`},
		{1000, `{"hex":"A12345","flight":"UAL123  ","alt_baro":35000,"gs":450,"track":270,"lat":47.4503,"lon":-122.3188,"nic":8,"rc":186,"seen_pos":0.1,"version":2,"nac_p":10,"nac_v":2,"sil":3,"sil_type":"perhour","gva":2,"sda":2,"mlat":[],"tisb":[],"messages":152,"seen":0.1,"rssi":-25.3}`},
		{2000, `{"hex":"A12345","flight":"UAL123  ","alt_baro":35000,"gs":450,"track":270,"lat":47.4504,"lon":-122.3288,"nic":8,"rc":186,"seen_pos":0.1,"version":2,"nac_p":10,"nac_v":2,"sil":3,"sil_type":"perhour","gva":2,"sda":2,"mlat":[],"tisb":[],"messages":154,"seen":0.1,"rssi":-25.7}`},
	}

	// Sample aircraft 2: N172SP - general aviation
	messages = append(messages, []struct {
		offsetMs int64
		data     string
	}{
		{500, `{"hex":"AC82EC","flight":"N172SP  ","alt_baro":5500,"gs":120,"track":90,"lat":47.4600,"lon":-122.2900,"nic":7,"rc":370,"seen_pos":0.2,"version":2,"nac_p":9,"nac_v":1,"sil":3,"sil_type":"perhour","messages":80,"seen":0.2,"rssi":-18.2}`},
		{1500, `{"hex":"AC82EC","flight":"N172SP  ","alt_baro":5520,"gs":121,"track":91,"lat":47.4605,"lon":-122.2800,"nic":7,"rc":370,"seen_pos":0.2,"version":2,"nac_p":9,"nac_v":1,"sil":3,"sil_type":"perhour","messages":82,"seen":0.2,"rssi":-18.5}`},
		{2500, `{"hex":"AC82EC","flight":"N172SP  ","alt_baro":5540,"gs":122,"track":92,"lat":47.4610,"lon":-122.2700,"nic":7,"rc":370,"seen_pos":0.2,"version":2,"nac_p":9,"nac_v":1,"sil":3,"sil_type":"perhour","messages":84,"seen":0.2,"rssi":-18.1}`},
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
