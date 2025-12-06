package main

import (
	"reflect"
	"testing"
	"time"
)

// TestBoolMarshal tests boolean to string conversion for SQL
func TestBoolMarshal(t *testing.T) {
	testCases := []struct {
		name     string
		value    bool
		expected string
	}{
		{
			name:     "True value",
			value:    true,
			expected: "1",
		},
		{
			name:     "False value",
			value:    false,
			expected: "0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.ValueOf(tc.value)
			result := boolMarshal(v)
			if result != tc.expected {
				t.Errorf("boolMarshal(%v) = %q, expected %q", tc.value, result, tc.expected)
			}
			t.Logf("bool(%v) -> %q", tc.value, result)
		})
	}
}

// TestIntMarshal tests integer to string conversion for SQL
func TestIntMarshal(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "Positive int",
			value:    int(42),
			expected: "42",
		},
		{
			name:     "Negative int",
			value:    int(-42),
			expected: "-42",
		},
		{
			name:     "Zero",
			value:    int(0),
			expected: "0",
		},
		{
			name:     "Large positive int",
			value:    int(2147483647),
			expected: "2147483647",
		},
		{
			name:     "Large negative int",
			value:    int(-2147483648),
			expected: "-2147483648",
		},
		{
			name:     "int8",
			value:    int8(127),
			expected: "127",
		},
		{
			name:     "int16",
			value:    int16(32767),
			expected: "32767",
		},
		{
			name:     "int32",
			value:    int32(-12345),
			expected: "-12345",
		},
		{
			name:     "int64",
			value:    int64(9223372036854775807),
			expected: "9223372036854775807",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.ValueOf(tc.value)
			result := intMarshal(v)
			if result != tc.expected {
				t.Errorf("intMarshal(%v) = %q, expected %q", tc.value, result, tc.expected)
			}
			t.Logf("int(%v) -> %q", tc.value, result)
		})
	}
}

// TestUintMarshal tests unsigned integer to string conversion for SQL
func TestUintMarshal(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "Positive uint",
			value:    uint(42),
			expected: "42",
		},
		{
			name:     "Zero",
			value:    uint(0),
			expected: "0",
		},
		{
			name:     "Large uint",
			value:    uint(4294967295),
			expected: "4294967295",
		},
		{
			name:     "uint8",
			value:    uint8(255),
			expected: "255",
		},
		{
			name:     "uint16",
			value:    uint16(65535),
			expected: "65535",
		},
		{
			name:     "uint32",
			value:    uint32(12345),
			expected: "12345",
		},
		{
			name:     "uint64 max",
			value:    uint64(18446744073709551615),
			expected: "18446744073709551615",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.ValueOf(tc.value)
			result := uintMarshal(v)
			if result != tc.expected {
				t.Errorf("uintMarshal(%v) = %q, expected %q", tc.value, result, tc.expected)
			}
			t.Logf("uint(%v) -> %q", tc.value, result)
		})
	}
}

// TestFloatMarshal tests float to string conversion for SQL
func TestFloatMarshal(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "Positive float",
			value:    float32(42.5),
			expected: "42.5000000000",
		},
		{
			name:     "Negative float",
			value:    float32(-42.5),
			expected: "-42.5000000000",
		},
		{
			name:     "Zero",
			value:    float64(0.0),
			expected: "0.0000000000",
		},
		{
			name:     "Small decimal",
			value:    float64(0.123456789012345),
			expected: "0.1234567890",
		},
		{
			name:     "Large float",
			value:    float64(123456789.123456789),
			expected: "123456789.1234567910",
		},
		{
			name:     "Scientific notation input",
			value:    float64(1.23e10),
			expected: "12300000000.0000000000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.ValueOf(tc.value)
			result := floatMarshal(v)
			if result != tc.expected {
				t.Errorf("floatMarshal(%v) = %q, expected %q", tc.value, result, tc.expected)
			}
			t.Logf("float(%v) -> %q", tc.value, result)
		})
	}
}

// TestStringMarshal tests string passthrough for SQL
func TestStringMarshal(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "Simple string",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "Empty string",
			value:    "",
			expected: "",
		},
		{
			name:     "String with spaces",
			value:    "hello world",
			expected: "hello world",
		},
		{
			name:     "String with special characters",
			value:    "test@#$%^&*()",
			expected: "test@#$%^&*()",
		},
		{
			name:     "String with newlines",
			value:    "line1\nline2",
			expected: "line1\nline2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.ValueOf(tc.value)
			result := stringMarshal(v)
			if result != tc.expected {
				t.Errorf("stringMarshal(%q) = %q, expected %q", tc.value, result, tc.expected)
			}
			t.Logf("string(%q) -> %q", tc.value, result)
		})
	}
}

// TestNotsupportedMarshal tests unsupported type handling
func TestNotsupportedMarshal(t *testing.T) {
	testCases := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "Complex number",
			value: complex(1, 2),
		},
		{
			name:  "Nil interface",
			value: (*int)(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.ValueOf(tc.value)
			result := notsupportedMarshal(v)
			if result != "" {
				t.Errorf("notsupportedMarshal(%v) = %q, expected empty string", tc.value, result)
			}
			t.Logf("notsupported(%v) -> %q", tc.value, result)
		})
	}
}

// Struct with String() method for testing structCanBeMarshalled and structMarshal
type TestStructWithString struct {
	Value string
}

func (t TestStructWithString) String() string {
	return "TestStruct:" + t.Value
}

// Struct without String() method for testing
type TestStructWithoutString struct {
	Value string
}

// TestStructCanBeMarshalled tests struct marshallability detection
func TestStructCanBeMarshalled(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{
			name:     "Struct with String method",
			value:    TestStructWithString{Value: "test"},
			expected: true,
		},
		{
			name:     "Struct without String method",
			value:    TestStructWithoutString{Value: "test"},
			expected: false,
		},
		{
			name:     "Pointer to struct with String method",
			value:    &TestStructWithString{Value: "test"},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.ValueOf(tc.value)
			result := structCanBeMarshalled(v)
			if result != tc.expected {
				t.Errorf("structCanBeMarshalled(%T) = %v, expected %v",
					tc.value, result, tc.expected)
			}
			t.Logf("struct(%T) -> canMarshal=%v", tc.value, result)
		})
	}
}

// TestStructMarshal tests struct marshalling via String() method
func TestStructMarshal(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "Struct with String method",
			value:    TestStructWithString{Value: "hello"},
			expected: "TestStruct:hello",
		},
		{
			name:     "Struct without String method",
			value:    TestStructWithoutString{Value: "hello"},
			expected: "",
		},
		{
			name:     "Pointer to struct with String method",
			value:    &TestStructWithString{Value: "world"},
			expected: "TestStruct:world",
		},
		{
			name:     "Struct with empty value",
			value:    TestStructWithString{Value: ""},
			expected: "TestStruct:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.ValueOf(tc.value)
			result := structMarshal(v)
			if result != tc.expected {
				t.Errorf("structMarshal(%T{%v}) = %q, expected %q",
					tc.value, tc.value, result, tc.expected)
			}
			t.Logf("struct(%T) -> %q", tc.value, result)
		})
	}
}

// TestMarshalFunctionsIntegration tests integration with different types
func TestMarshalFunctionsIntegration(t *testing.T) {
	testCases := []struct {
		name        string
		value       interface{}
		marshalFunc func(reflect.Value) string
		expected    string
	}{
		{
			name:        "Bool true via intMarshal",
			value:       true,
			marshalFunc: boolMarshal,
			expected:    "1",
		},
		{
			name:        "Int via intMarshal",
			value:       int(999),
			marshalFunc: intMarshal,
			expected:    "999",
		},
		{
			name:        "Uint via uintMarshal",
			value:       uint(123),
			marshalFunc: uintMarshal,
			expected:    "123",
		},
		{
			name:        "Float via floatMarshal",
			value:       float64(3.14159),
			marshalFunc: floatMarshal,
			expected:    "3.1415900000",
		},
		{
			name:        "String via stringMarshal",
			value:       "integration test",
			marshalFunc: stringMarshal,
			expected:    "integration test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.ValueOf(tc.value)
			result := tc.marshalFunc(v)
			if result != tc.expected {
				t.Errorf("%s(%v) = %q, expected %q",
					tc.name, tc.value, result, tc.expected)
			}
		})
	}
}

// TestIsDataLogReady tests the isDataLogReady function
func TestIsDataLogReady(t *testing.T) {
	// Save original state
	origReadyToWrite := dataLogReadyToWrite
	defer func() {
		dataLogReadyToWrite = origReadyToWrite
	}()

	t.Run("Returns true when dataLogReadyToWrite is true", func(t *testing.T) {
		dataLogReadyToWrite = true
		result := isDataLogReady()
		if !result {
			t.Error("Expected isDataLogReady() to return true when dataLogReadyToWrite is true")
		}
		t.Log("isDataLogReady() correctly returned true")
	})

	t.Run("Returns false when dataLogReadyToWrite is false", func(t *testing.T) {
		dataLogReadyToWrite = false
		result := isDataLogReady()
		if result {
			t.Error("Expected isDataLogReady() to return false when dataLogReadyToWrite is false")
		}
		t.Log("isDataLogReady() correctly returned false")
	})
}

// TestLogMsg tests the logMsg function which logs message data
func TestLogMsg(t *testing.T) {
	// Save original settings
	origReplayLog := globalSettings.ReplayLog
	origReadyToWrite := dataLogReadyToWrite
	defer func() {
		globalSettings.ReplayLog = origReplayLog
		dataLogReadyToWrite = origReadyToWrite
		// Drain the channel if there's anything left
		if dataLogChan != nil {
			for len(dataLogChan) > 0 {
				<-dataLogChan
			}
		}
	}()

	// Create a test message
	testMsg := msg{
		MessageClass: MSGCLASS_UAT,
		Data:         "test message data",
	}

	t.Run("LogMsg when ReplayLog enabled and datalog ready", func(t *testing.T) {
		// Setup: enable replay logging and make datalog ready
		globalSettings.ReplayLog = true
		dataLogReadyToWrite = true

		// Initialize the channel if it doesn't exist
		if dataLogChan == nil {
			dataLogChan = make(chan DataLogRow, 10)
		}

		// Call logMsg
		logMsg(testMsg)

		// Verify message was sent to channel
		select {
		case row := <-dataLogChan:
			if row.tbl != "messages" {
				t.Errorf("Expected table 'messages', got %q", row.tbl)
			}
			receivedMsg, ok := row.data.(msg)
			if !ok {
				t.Errorf("Expected data type msg, got %T", row.data)
			}
			if receivedMsg.Data != testMsg.Data {
				t.Errorf("Expected message data %q, got %q", testMsg.Data, receivedMsg.Data)
			}
			if receivedMsg.MessageClass != testMsg.MessageClass {
				t.Errorf("Expected MessageClass %d, got %d", testMsg.MessageClass, receivedMsg.MessageClass)
			}
			t.Logf("Successfully logged message: %+v", receivedMsg)
		default:
			t.Error("Expected message to be sent to dataLogChan, but channel was empty")
		}
	})

	t.Run("LogMsg when ReplayLog disabled", func(t *testing.T) {
		// Setup: disable replay logging
		globalSettings.ReplayLog = false
		dataLogReadyToWrite = true

		// Initialize the channel if it doesn't exist
		if dataLogChan == nil {
			dataLogChan = make(chan DataLogRow, 10)
		}

		// Call logMsg
		logMsg(testMsg)

		// Verify message was NOT sent to channel
		select {
		case row := <-dataLogChan:
			t.Errorf("Expected no message, but got message with table %q", row.tbl)
		default:
			t.Log("Message correctly not logged when ReplayLog is disabled")
		}
	})

	t.Run("LogMsg when datalog not ready", func(t *testing.T) {
		// Setup: enable replay logging but datalog not ready
		globalSettings.ReplayLog = true
		dataLogReadyToWrite = false

		// Initialize the channel if it doesn't exist
		if dataLogChan == nil {
			dataLogChan = make(chan DataLogRow, 10)
		}

		// Call logMsg
		logMsg(testMsg)

		// Verify message was NOT sent to channel
		select {
		case row := <-dataLogChan:
			t.Errorf("Expected no message, but got message with table %q", row.tbl)
		default:
			t.Log("Message correctly not logged when datalog is not ready")
		}
	})

	t.Run("LogMsg when both disabled", func(t *testing.T) {
		// Setup: both disabled
		globalSettings.ReplayLog = false
		dataLogReadyToWrite = false

		// Initialize the channel if it doesn't exist
		if dataLogChan == nil {
			dataLogChan = make(chan DataLogRow, 10)
		}

		// Call logMsg
		logMsg(testMsg)

		// Verify message was NOT sent to channel
		select {
		case row := <-dataLogChan:
			t.Errorf("Expected no message, but got message with table %q", row.tbl)
		default:
			t.Log("Message correctly not logged when both conditions are false")
		}
	})
}

// TestCheckTimestamp tests the timestamp management function
func TestCheckTimestamp(t *testing.T) {
	// Initialize required globals
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond)
	}

	// Save original state
	origTimestamps := dataLogTimestamps
	origCurTimestamp := dataLogCurTimestamp
	defer func() {
		dataLogTimestamps = origTimestamps
		dataLogCurTimestamp = origCurTimestamp
	}()

	t.Run("first_call_creates_timestamp", func(t *testing.T) {
		// Reset state
		dataLogTimestamps = []StratuxTimestamp{
			{
				id:                   0,
				Time_type_preference: 0,
				StratuxClock_value:   stratuxClock.Time.Add(-2 * time.Second), // Old enough to trigger new timestamp
				PreferredTime_value:  stratuxClock.Time.Add(-2 * time.Second),
			},
		}
		dataLogCurTimestamp = 0

		result := checkTimestamp()

		// Should return false because a new timestamp was created
		if result {
			t.Error("Expected checkTimestamp to return false when creating new timestamp")
		}

		// Should have added a new timestamp
		if len(dataLogTimestamps) != 2 {
			t.Errorf("Expected 2 timestamps, got %d", len(dataLogTimestamps))
		}

		// New timestamp should have StratuxClock type (0)
		newTs := dataLogTimestamps[1]
		if newTs.Time_type_preference != 0 {
			t.Errorf("Expected Time_type_preference=0 (stratuxClock), got %d", newTs.Time_type_preference)
		}
	})

	t.Run("recent_timestamp_returns_true", func(t *testing.T) {
		// Reset state with very recent timestamp
		dataLogTimestamps = []StratuxTimestamp{
			{
				id:                   0,
				Time_type_preference: 0,
				StratuxClock_value:   stratuxClock.Time, // Just now
				PreferredTime_value:  stratuxClock.Time,
			},
		}
		dataLogCurTimestamp = 0

		result := checkTimestamp()

		// Should return true because timestamp is still valid
		if !result {
			t.Error("Expected checkTimestamp to return true when timestamp is still valid")
		}

		// Should NOT have added a new timestamp
		if len(dataLogTimestamps) != 1 {
			t.Errorf("Expected 1 timestamp (unchanged), got %d", len(dataLogTimestamps))
		}
	})

	t.Run("gps_extrapolation", func(t *testing.T) {
		// Save GPS state
		origGPSLastGPSTime := mySituation.GPSLastGPSTimeStratuxTime
		origGPSConnected := globalStatus.GPS_connected
		defer func() {
			mySituation.GPSLastGPSTimeStratuxTime = origGPSLastGPSTime
			globalStatus.GPS_connected = origGPSConnected
		}()

		// Set up GPS clock as valid (isGPSClockValid checks GPSLastGPSTimeStratuxTime)
		mySituation.GPSLastGPSTimeStratuxTime = stratuxClock.Time
		globalStatus.GPS_connected = true

		// Reset state with GPS timestamp
		baseTime := time.Now().UTC()
		dataLogTimestamps = []StratuxTimestamp{
			{
				id:                   0,
				Time_type_preference: 1, // GPS time
				StratuxClock_value:   stratuxClock.Time.Add(-2 * time.Second),
				GPSClock_value:       baseTime.Add(-2 * time.Second),
				PreferredTime_value:  baseTime.Add(-2 * time.Second),
			},
		}
		dataLogCurTimestamp = 0

		result := checkTimestamp()

		// Should return false because new timestamp created
		if result {
			t.Error("Expected checkTimestamp to return false when creating new timestamp")
		}

		// Should have added a new timestamp
		if len(dataLogTimestamps) < 2 {
			t.Errorf("Expected at least 2 timestamps, got %d", len(dataLogTimestamps))
		}

		// New timestamp may be extrapolated (type 2) if GPS was valid
		newTs := dataLogTimestamps[len(dataLogTimestamps)-1]
		t.Logf("New timestamp type: %d", newTs.Time_type_preference)
	})
}

// TestSetDataLogTimeWithGPS tests the GPS time logging function
func TestSetDataLogTimeWithGPS(t *testing.T) {
	// Initialize required globals
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(50 * time.Millisecond)
	}

	// Save original state
	origTimestamps := dataLogTimestamps
	origCurTimestamp := dataLogCurTimestamp
	origDataLogStarted := dataLogStarted
	origGPSLastGPSTime := mySituation.GPSLastGPSTimeStratuxTime
	defer func() {
		dataLogTimestamps = origTimestamps
		dataLogCurTimestamp = origCurTimestamp
		dataLogStarted = origDataLogStarted
		mySituation.GPSLastGPSTimeStratuxTime = origGPSLastGPSTime
	}()

	t.Run("logs_gps_time_when_conditions_met", func(t *testing.T) {
		// Setup: datalog started and GPS clock valid
		dataLogStarted = true
		mySituation.GPSLastGPSTimeStratuxTime = stratuxClock.Time // Make GPS clock valid
		dataLogTimestamps = []StratuxTimestamp{}
		dataLogCurTimestamp = -1

		// Create test situation with GPS time
		gpsTime := time.Now().UTC()
		testSit := SituationData{
			GPSTime: gpsTime,
		}

		setDataLogTimeWithGPS(testSit)

		// Should have added a timestamp
		if len(dataLogTimestamps) != 1 {
			t.Errorf("Expected 1 timestamp, got %d", len(dataLogTimestamps))
		}

		if len(dataLogTimestamps) > 0 {
			ts := dataLogTimestamps[0]
			if ts.Time_type_preference != 1 {
				t.Errorf("Expected Time_type_preference=1 (GPS), got %d", ts.Time_type_preference)
			}
			if !ts.GPSClock_value.Equal(gpsTime) {
				t.Errorf("Expected GPSClock_value=%v, got %v", gpsTime, ts.GPSClock_value)
			}
			if !ts.PreferredTime_value.Equal(gpsTime) {
				t.Errorf("Expected PreferredTime_value=%v, got %v", gpsTime, ts.PreferredTime_value)
			}
		}
	})

	t.Run("no_log_when_datalog_not_started", func(t *testing.T) {
		// Setup: datalog NOT started
		dataLogStarted = false
		mySituation.GPSLastGPSTimeStratuxTime = stratuxClock.Time // GPS clock valid
		dataLogTimestamps = []StratuxTimestamp{}
		dataLogCurTimestamp = -1

		testSit := SituationData{
			GPSTime: time.Now().UTC(),
		}

		setDataLogTimeWithGPS(testSit)

		// Should NOT have added a timestamp
		if len(dataLogTimestamps) != 0 {
			t.Errorf("Expected 0 timestamps when datalog not started, got %d", len(dataLogTimestamps))
		}
	})

	t.Run("no_log_when_gps_clock_invalid", func(t *testing.T) {
		// Setup: datalog started but GPS clock invalid
		dataLogStarted = true
		mySituation.GPSLastGPSTimeStratuxTime = time.Time{} // Invalid (zero time)
		dataLogTimestamps = []StratuxTimestamp{}
		dataLogCurTimestamp = -1

		testSit := SituationData{
			GPSTime: time.Now().UTC(),
		}

		setDataLogTimeWithGPS(testSit)

		// Should NOT have added a timestamp
		if len(dataLogTimestamps) != 0 {
			t.Errorf("Expected 0 timestamps when GPS clock invalid, got %d", len(dataLogTimestamps))
		}
	})

	t.Run("updates_current_timestamp_index", func(t *testing.T) {
		// Setup
		dataLogStarted = true
		mySituation.GPSLastGPSTimeStratuxTime = stratuxClock.Time
		dataLogTimestamps = []StratuxTimestamp{
			{id: 0}, // Existing timestamp
		}
		dataLogCurTimestamp = 0

		testSit := SituationData{
			GPSTime: time.Now().UTC(),
		}

		setDataLogTimeWithGPS(testSit)

		// Current timestamp should now be 1 (the new one)
		if dataLogCurTimestamp != 1 {
			t.Errorf("Expected dataLogCurTimestamp=1, got %d", dataLogCurTimestamp)
		}
	})
}
