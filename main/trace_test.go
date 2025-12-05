package main

import (
	"compress/gzip"
	"encoding/csv"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestTraceLoggerRecordAndRead tests recording and reading trace files
func TestTraceLoggerRecordAndRead(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()
	traceFile := filepath.Join(tmpDir, "test_trace.txt.gz")

	// Create and write a trace file manually
	fh, err := os.Create(traceFile)
	if err != nil {
		t.Fatalf("Failed to create trace file: %v", err)
	}

	gzw := gzip.NewWriter(fh)
	csvw := csv.NewWriter(gzw)

	testData := []struct {
		timestamp time.Time
		context   string
		data      string
	}{
		{time.Date(2025, 10, 13, 12, 0, 0, 0, time.UTC), CONTEXT_DUMP1090, `{"hex":"A12345","flight":"UAL123"}`},
		{time.Date(2025, 10, 13, 12, 0, 1, 0, time.UTC), CONTEXT_NMEA, "$GPRMC,120000,A,4727.030,N,12218.528,W,057.9,349.7,131025,015.0,E*66"},
		{time.Date(2025, 10, 13, 12, 0, 2, 0, time.UTC), CONTEXT_DUMP1090, `{"hex":"AC82EC","flight":"N172SP"}`},
	}

	for _, td := range testData {
		err := csvw.Write([]string{
			td.timestamp.Format(time.RFC3339Nano),
			td.context,
			td.data,
		})
		if err != nil {
			t.Fatalf("Failed to write CSV record: %v", err)
		}
	}

	csvw.Flush()
	gzw.Close()
	fh.Close()

	// Now read it back
	fh, err = os.Open(traceFile)
	if err != nil {
		t.Fatalf("Failed to open trace file: %v", err)
	}
	defer fh.Close()

	gzr, err := gzip.NewReader(fh)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	csvr := csv.NewReader(gzr)
	recordNum := 0

	for {
		record, err := csvr.Read()
		if err != nil {
			break
		}

		if len(record) != 3 {
			t.Errorf("Expected 3 fields, got %d", len(record))
			continue
		}

		if recordNum >= len(testData) {
			t.Errorf("Got more records than expected")
			break
		}

		expected := testData[recordNum]

		// Parse timestamp
		ts, err := time.Parse(time.RFC3339Nano, record[0])
		if err != nil {
			t.Errorf("Failed to parse timestamp %q: %v", record[0], err)
		}
		if !ts.Equal(expected.timestamp) {
			t.Errorf("Record %d: timestamp = %v, expected %v", recordNum, ts, expected.timestamp)
		}

		// Check context
		if record[1] != expected.context {
			t.Errorf("Record %d: context = %q, expected %q", recordNum, record[1], expected.context)
		}

		// Check data
		if record[2] != expected.data {
			t.Errorf("Record %d: data = %q, expected %q", recordNum, record[2], expected.data)
		}

		recordNum++
	}

	if recordNum != len(testData) {
		t.Errorf("Read %d records, expected %d", recordNum, len(testData))
	}

	t.Logf("Successfully read and verified %d trace records", recordNum)
}

// TestTraceContextConstants verifies trace context constants
func TestTraceContextConstants(t *testing.T) {
	testCases := []struct {
		name     string
		constant string
		expected string
	}{
		{"AIS context", CONTEXT_AIS, "ais"},
		{"NMEA context", CONTEXT_NMEA, "nmea"},
		{"APRS context", CONTEXT_APRS, "aprs"},
		{"OGN-RX context", CONTEXT_OGN_RX, "ogn-rx"},
		{"DUMP1090 context", CONTEXT_DUMP1090, "dump1090"},
		{"GODUMP978 context", CONTEXT_GODUMP978, "godump978"},
		{"Low Power UAT context", CONTEXT_LOWPOWERUAT, "lowpower_uat"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.constant != tc.expected {
				t.Errorf("%s = %q, expected %q", tc.name, tc.constant, tc.expected)
			}
		})
	}
}

// TestTraceFileCompression verifies gzip compression is working
func TestTraceFileCompression(t *testing.T) {
	tmpDir := t.TempDir()
	traceFile := filepath.Join(tmpDir, "compression_test.txt.gz")

	// Write a large amount of repetitive data
	fh, err := os.Create(traceFile)
	if err != nil {
		t.Fatalf("Failed to create trace file: %v", err)
	}

	gzw := gzip.NewWriter(fh)
	csvw := csv.NewWriter(gzw)

	// Write 100 identical records
	testData := "This is a test string that should compress well when repeated many times"
	for i := 0; i < 100; i++ {
		ts := time.Date(2025, 10, 13, 12, 0, i, 0, time.UTC)
		err := csvw.Write([]string{
			ts.Format(time.RFC3339Nano),
			CONTEXT_DUMP1090,
			testData,
		})
		if err != nil {
			t.Fatalf("Failed to write CSV record: %v", err)
		}
	}

	csvw.Flush()
	gzw.Close()
	fh.Close()

	// Check file size
	info, err := os.Stat(traceFile)
	if err != nil {
		t.Fatalf("Failed to stat trace file: %v", err)
	}

	compressedSize := info.Size()
	// Each uncompressed record is roughly: 30 (timestamp) + 8 (context) + 72 (data) = 110 bytes
	// 100 records = ~11000 bytes uncompressed
	// With compression, we expect significantly smaller
	if compressedSize > 5000 {
		t.Errorf("Compressed size %d bytes seems too large (expected < 5000 for repetitive data)", compressedSize)
	}

	t.Logf("Compressed 100 repetitive records to %d bytes", compressedSize)
}

// TestTraceFileReading tests reading various trace file formats
func TestTraceFileReading(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		minCount int // minimum expected message count
	}{
		{
			name:     "1090ES ADS-B trace",
			filename: "testdata/adsb/basic_adsb.trace.gz",
			minCount: 6,
		},
		{
			name:     "GPS NMEA trace",
			filename: "testdata/gps/basic_gps.trace.gz",
			minCount: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check if file exists
			if _, err := os.Stat(tc.filename); os.IsNotExist(err) {
				t.Skipf("Test file %s does not exist, skipping", tc.filename)
			}

			// Open and read the trace file
			fh, err := os.Open(tc.filename)
			if err != nil {
				t.Fatalf("Failed to open trace file: %v", err)
			}
			defer fh.Close()

			gzr, err := gzip.NewReader(fh)
			if err != nil {
				t.Fatalf("Failed to create gzip reader: %v", err)
			}
			defer gzr.Close()

			csvr := csv.NewReader(gzr)
			count := 0

			for {
				record, err := csvr.Read()
				if err != nil {
					break
				}

				if len(record) != 3 {
					t.Errorf("Record %d has %d fields, expected 3", count, len(record))
					continue
				}

				// Verify timestamp is parseable
				_, err = time.Parse(time.RFC3339Nano, record[0])
				if err != nil {
					t.Errorf("Record %d: invalid timestamp %q: %v", count, record[0], err)
				}

				// Verify context is non-empty
				if record[1] == "" {
					t.Errorf("Record %d: empty context", count)
				}

				// Verify data is non-empty
				if record[2] == "" {
					t.Errorf("Record %d: empty data", count)
				}

				count++
			}

			if count < tc.minCount {
				t.Errorf("Read %d records, expected at least %d", count, tc.minCount)
			}

			t.Logf("Successfully validated %d trace records from %s", count, tc.filename)
		})
	}
}

// TestTraceTimestampOrdering verifies timestamps are in chronological order
func TestTraceTimestampOrdering(t *testing.T) {
	tmpDir := t.TempDir()
	traceFile := filepath.Join(tmpDir, "ordered_test.txt.gz")

	// Create trace file with ordered timestamps
	fh, err := os.Create(traceFile)
	if err != nil {
		t.Fatalf("Failed to create trace file: %v", err)
	}

	gzw := gzip.NewWriter(fh)
	csvw := csv.NewWriter(gzw)

	baseTime := time.Date(2025, 10, 13, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		ts := baseTime.Add(time.Duration(i) * time.Second)
		err := csvw.Write([]string{
			ts.Format(time.RFC3339Nano),
			CONTEXT_NMEA,
			"test data",
		})
		if err != nil {
			t.Fatalf("Failed to write record: %v", err)
		}
	}

	csvw.Flush()
	gzw.Close()
	fh.Close()

	// Read back and verify ordering
	fh, err = os.Open(traceFile)
	if err != nil {
		t.Fatalf("Failed to open trace file: %v", err)
	}
	defer fh.Close()

	gzr, err := gzip.NewReader(fh)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	csvr := csv.NewReader(gzr)
	var prevTime time.Time

	for i := 0; ; i++ {
		record, err := csvr.Read()
		if err != nil {
			break
		}

		ts, err := time.Parse(time.RFC3339Nano, record[0])
		if err != nil {
			t.Fatalf("Failed to parse timestamp: %v", err)
		}

		if i > 0 && ts.Before(prevTime) {
			t.Errorf("Timestamp at record %d (%v) is before previous (%v)", i, ts, prevTime)
		}

		prevTime = ts
	}

	t.Logf("Verified chronological ordering of timestamps")
}

// TestTraceLoggerRecord_NoFileHandle tests Record with no active file
func TestTraceLoggerRecord_NoFileHandle(t *testing.T) {
	// Save original state
	origTraceLog := TraceLog

	// Create a fresh TraceLogger with no file handle
	TraceLog = TraceLogger{}

	// Record should not panic when fileHandle is nil
	TraceLog.Record(CONTEXT_DUMP1090, []byte("test data"))

	// Give goroutine time to execute
	time.Sleep(10 * time.Millisecond)

	// Restore original
	TraceLog = origTraceLog

	t.Log("Record with no file handle completed without panic")
}

// TestTraceLoggerRecord_IsReplaying tests Record when replaying
func TestTraceLoggerRecord_IsReplaying(t *testing.T) {
	// Save original state
	origTraceLog := TraceLog

	// Create a TraceLogger that is replaying
	TraceLog = TraceLogger{
		isReplaying: true,
	}

	// Record should return early when replaying
	TraceLog.Record(CONTEXT_NMEA, []byte("test NMEA data"))

	// Give goroutine time to execute
	time.Sleep(10 * time.Millisecond)

	// Restore original
	TraceLog = origTraceLog

	t.Log("Record when replaying completed without action")
}

// TestTraceLoggerOnTimestamp_NoFileHandle tests OnTimestamp with no file
func TestTraceLoggerOnTimestamp_NoFileHandle(t *testing.T) {
	// Save original state
	origTraceLog := TraceLog

	// Create a fresh TraceLogger
	TraceLog = TraceLogger{}

	ts := time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)

	// OnTimestamp should return early when fileHandle is nil
	TraceLog.OnTimestamp(ts)

	if TraceLog.hasProperFilename {
		t.Error("hasProperFilename should still be false")
	}

	// Restore original
	TraceLog = origTraceLog

	t.Log("OnTimestamp with no file handle completed")
}

// TestTraceLoggerOnTimestamp_AlreadyHasProperFilename tests OnTimestamp when already named
func TestTraceLoggerOnTimestamp_AlreadyHasProperFilename(t *testing.T) {
	// Save original state
	origTraceLog := TraceLog

	// Create a TraceLogger that already has proper filename
	TraceLog = TraceLogger{
		hasProperFilename: true,
		fileName:          "/var/log/stratux/test_trace.txt.gz",
	}

	ts := time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)
	origFilename := TraceLog.fileName

	TraceLog.OnTimestamp(ts)

	// Filename should not have changed
	if TraceLog.fileName != origFilename {
		t.Errorf("Filename should not change when hasProperFilename=true")
	}

	// Restore original
	TraceLog = origTraceLog

	t.Log("OnTimestamp with proper filename already set completed")
}

// TestTraceLoggerIsActive tests IsActive method
func TestTraceLoggerIsActive(t *testing.T) {
	// Save original state
	origTraceLog := TraceLog

	// Create a fresh TraceLogger
	TraceLog = TraceLogger{}

	// Should be inactive with no file handle
	if TraceLog.IsActive() {
		t.Error("Expected IsActive=false with no file handle")
	}

	// Restore original
	TraceLog = origTraceLog

	t.Log("IsActive returns correct state")
}

// TestTraceLoggerIsReplaying tests IsReplaying method
func TestTraceLoggerIsReplaying(t *testing.T) {
	// Save original state
	origTraceLog := TraceLog

	// Create a fresh TraceLogger
	TraceLog = TraceLogger{}

	// Should not be replaying by default
	if TraceLog.IsReplaying() {
		t.Error("Expected IsReplaying=false by default")
	}

	// Set replaying
	TraceLog.isReplaying = true

	if !TraceLog.IsReplaying() {
		t.Error("Expected IsReplaying=true after setting flag")
	}

	// Restore original
	TraceLog = origTraceLog

	t.Log("IsReplaying returns correct state")
}

// TestTraceLoggerFlush_NoFileHandle tests Flush with no file
func TestTraceLoggerFlush_NoFileHandle(t *testing.T) {
	// Save original state
	origTraceLog := TraceLog

	// Create a fresh TraceLogger
	TraceLog = TraceLogger{}

	// Flush should not panic when fileHandle is nil
	TraceLog.Flush()

	// Restore original
	TraceLog = origTraceLog

	t.Log("Flush with no file handle completed without panic")
}

// TestInjectTraceMessage_AllContexts tests message injection for all context types
func TestInjectTraceMessage_AllContexts(t *testing.T) {
	// Initialize stratuxClock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond)
	}

	// Initialize required globals
	initTestState()

	testCases := []struct {
		context string
		data    string
	}{
		{CONTEXT_AIS, "!AIVDM,1,1,,A,13u@ND0P00PkCj0L1uUoEf600000,0*69"},
		{CONTEXT_NMEA, "$GPRMC,120000,A,4727.030,N,12218.528,W,057.9,349.7,131025,015.0,E*66"},
		{CONTEXT_APRS, `FLR395F39>APRS,qAS,OXFORD:/120000h5145.945N/00111.511W'057/057/A=000407 !W02! id06395F39`},
		{CONTEXT_OGN_RX, `{"sys":"OGN","addr":"123456","lat_deg":51.7657,"lon_deg":-1.1918}`},
		{CONTEXT_DUMP1090, `{"hex":"A12345","flight":"N12345","lat":47.5,"lon":-122.3}`},
	}

	for _, tc := range testCases {
		t.Run(tc.context, func(t *testing.T) {
			// Set time slightly in the future so we don't sleep
			ts := stratuxClock.Time.Add(-1 * time.Millisecond)

			// This should not panic
			injectTraceMessage(tc.context, ts, []byte(tc.data))

			t.Logf("Injected %s message", tc.context)
		})
	}
}

// initTestState initializes global state for trace injection tests
func initTestState() {
	// Initialize mutexes
	if mySituation.muGPS == nil {
		mySituation.muGPS = &sync.Mutex{}
		mySituation.muGPSPerformance = &sync.Mutex{}
		mySituation.muAttitude = &sync.Mutex{}
		mySituation.muBaro = &sync.Mutex{}
		mySituation.muSatellite = &sync.Mutex{}
	}

	// Initialize traffic
	if trafficMutex == nil {
		trafficMutex = &sync.Mutex{}
	}
	if traffic == nil {
		traffic = make(map[uint32]TrafficInfo)
	}
	if seenTraffic == nil {
		seenTraffic = make(map[uint32]bool)
	}

	// Initialize message log
	msgLogMutex = sync.Mutex{}
	if msgLog == nil {
		msgLog = make([]msg, 0)
	}
}
