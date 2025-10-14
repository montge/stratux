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
	// Create sample OGN trace file
	fname := "basic_ogn.trace.gz"
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

	// Sample OGN messages from ogn-rx-eu
	// Format: time offset in ms, JSON message
	// Messages include both traffic messages and status messages
	messages := []struct {
		offsetMs int64
		data     string
	}{
		// STATUS MESSAGE - Background noise, gain, tx status
		{0, `{"sys":"status","bkg_noise_db":-110.5,"gain_db":48.0,"tx_enabled":false}`},

		// TRAFFIC MESSAGE 1 - ICAO address, glider (acft_type=1)
		// Location: 51.7657°N, 1.1918°W (near Oxford, UK)
		{500, `{"sys":"OGN","time":1728907200.5,"addr":"395F39","addr_type":1,"acft_type":"1","lat_deg":51.7657533,"lon_deg":-1.1918533,"alt_msl_m":124.5,"alt_std_m":63.2,"track_deg":57.0,"speed_mps":15.4,"climb_mps":-0.5,"turn_dps":0.0,"DOP":1.5,"snr_db":12.3}`},

		// TRAFFIC MESSAGE 2 - FLARM address, powered aircraft (acft_type=8)
		{1000, `{"sys":"FLR","time":1728907201.0,"addr":"DD4B12","addr_type":2,"acft_type":"8","lat_deg":51.7701,"lon_deg":-1.1956,"alt_msl_m":145.8,"alt_std_m":84.5,"track_deg":124.0,"speed_mps":25.7,"climb_mps":1.2,"turn_dps":-5.0,"DOP":1.2,"snr_db":15.8}`},

		// TRAFFIC MESSAGE 3 - OGN address with registration
		{1500, `{"sys":"OGN","time":1728907201.5,"addr":"395F39","addr_type":1,"reg":"G-ABCD","acft_type":"1","lat_deg":51.7665,"lon_deg":-1.1925,"alt_msl_m":126.3,"alt_std_m":65.1,"track_deg":58.0,"speed_mps":16.1,"climb_mps":-0.4,"turn_dps":1.0,"DOP":1.4,"snr_db":13.1}`},

		// TRAFFIC MESSAGE 4 - Skylines address (acft_type=1)
		{2000, `{"sys":"SKY","time":1728907202.0,"addr":"A12345","addr_type":2,"acft_type":"1","lat_deg":51.7623,"lon_deg":-1.1889,"alt_msl_m":132.1,"alt_std_m":70.8,"track_deg":215.0,"speed_mps":18.9,"climb_mps":0.8,"turn_dps":3.5,"DOP":1.8,"snr_db":11.5}`},

		// TRAFFIC MESSAGE 5 - With hardware type (Stratux)
		{2500, `{"sys":"OGN","time":1728907202.5,"addr":"395F40","addr_type":1,"acft_type":"8","hard":"STX","lat_deg":51.7589,"lon_deg":-1.1834,"alt_msl_m":138.7,"alt_std_m":77.4,"track_deg":310.0,"speed_mps":22.3,"climb_mps":1.5,"turn_dps":-2.0,"DOP":1.3,"snr_db":14.2}`},

		// TRAFFIC MESSAGE 6 - PAW (PilotAware) - no explicit addr_type
		{3000, `{"sys":"PAW","time":1728907203.0,"addr":"123ABC","addr_type":0,"acft_type":"8","lat_deg":51.7542,"lon_deg":-1.1778,"alt_msl_m":145.2,"alt_std_m":83.9,"track_deg":45.0,"speed_mps":28.6,"climb_mps":2.1,"turn_dps":0.0,"DOP":1.1,"snr_db":16.3}`},

		// REGISTRATION UPDATE - message without coordinates, only registration
		{3500, `{"sys":"OGN","addr":"395F39","reg":"G-WXYZ"}`},

		// TRAFFIC MESSAGE 7 - With HAE altitude instead of MSL
		{4000, `{"sys":"OGN","time":1728907204.0,"addr":"395F41","addr_type":1,"acft_type":"1","lat_deg":51.7501,"lon_deg":-1.1723,"alt_hae_m":145.3,"track_deg":180.0,"speed_mps":20.5,"climb_mps":0.0,"turn_dps":0.0,"DOP":1.6,"snr_db":10.8}`},

		// TRAFFIC MESSAGE 8 - Emitter category as hex (acft_cat)
		{4500, `{"sys":"OGN","time":1728907204.5,"addr":"395F42","addr_type":1,"acft_cat":"1A","lat_deg":51.7465,"lon_deg":-1.1667,"alt_msl_m":152.4,"alt_std_m":91.1,"track_deg":270.0,"speed_mps":19.2,"climb_mps":-1.0,"turn_dps":4.5,"DOP":1.4,"snr_db":12.9}`},

		// STATUS MESSAGE - Update at end
		{5000, `{"sys":"status","bkg_noise_db":-109.8,"gain_db":48.0,"tx_enabled":false}`},
	}

	// Write messages
	for _, msg := range messages {
		ts := baseTime.Add(time.Duration(msg.offsetMs) * time.Millisecond)
		err := w.Write([]string{
			ts.Format(time.RFC3339Nano),
			"ogn",
			msg.data,
		})
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Created %s with %d messages\n", fname, len(messages))
}
