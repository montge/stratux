package main

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestFmap tests the fmap range mapping function
func TestFmap(t *testing.T) {
	tests := []struct {
		name                                             string
		x, inMin, inMax, outMin, outMax, expected float64
	}{
		{
			name:     "Map 5 from [0,10] to [0,100] should give 50",
			x:        5,
			inMin:    0,
			inMax:    10,
			outMin:   0,
			outMax:   100,
			expected: 50,
		},
		{
			name:     "Map 0 from [0,10] to [0,100] should give 0",
			x:        0,
			inMin:    0,
			inMax:    10,
			outMin:   0,
			outMax:   100,
			expected: 0,
		},
		{
			name:     "Map 10 from [0,10] to [0,100] should give 100",
			x:        10,
			inMin:    0,
			inMax:    10,
			outMin:   0,
			outMax:   100,
			expected: 100,
		},
		{
			name:     "Map 25 from [0,100] to [32,212] (Celsius to Fahrenheit)",
			x:        25,
			inMin:    0,
			inMax:    100,
			outMin:   32,
			outMax:   212,
			expected: 77,
		},
		{
			name:     "Map 0 from [0,100] to [32,212] should give 32",
			x:        0,
			inMin:    0,
			inMax:    100,
			outMin:   32,
			outMax:   212,
			expected: 32,
		},
		{
			name:     "Map 100 from [0,100] to [32,212] should give 212",
			x:        100,
			inMin:    0,
			inMax:    100,
			outMin:   32,
			outMax:   212,
			expected: 212,
		},
		{
			name:     "Map 50 from [0,100] to [0,255] (PWM scaling)",
			x:        50,
			inMin:    0,
			inMax:    100,
			outMin:   0,
			outMax:   255,
			expected: 127.5,
		},
		{
			name:     "Map 75 from [0,100] to [20,100] (fan min duty 20%)",
			x:        75,
			inMin:    0,
			inMax:    100,
			outMin:   20,
			outMax:   100,
			expected: 80,
		},
		{
			name:     "Map negative value: -5 from [-10,10] to [0,100]",
			x:        -5,
			inMin:    -10,
			inMax:    10,
			outMin:   0,
			outMax:   100,
			expected: 25,
		},
		{
			name:     "Map to negative range: 5 from [0,10] to [-100,0]",
			x:        5,
			inMin:    0,
			inMax:    10,
			outMin:   -100,
			outMax:   0,
			expected: -50,
		},
		{
			name:     "Identity mapping: 5 from [0,10] to [0,10]",
			x:        5,
			inMin:    0,
			inMax:    10,
			outMin:   0,
			outMax:   10,
			expected: 5,
		},
		{
			name:     "Map decimal values: 2.5 from [0,5] to [0,10]",
			x:        2.5,
			inMin:    0,
			inMax:    5,
			outMin:   0,
			outMax:   10,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmap(tt.x, tt.inMin, tt.inMax, tt.outMin, tt.outMax)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("fmap(%f, %f, %f, %f, %f) = %f, want %f",
					tt.x, tt.inMin, tt.inMax, tt.outMin, tt.outMax, result, tt.expected)
			}
		})
	}
}

// TestFmapEdgeCases tests edge cases and boundary conditions
func TestFmapEdgeCases(t *testing.T) {
	t.Run("Map value at input minimum", func(t *testing.T) {
		result := fmap(0, 0, 100, 0, 255)
		if result != 0 {
			t.Errorf("Expected 0, got %f", result)
		}
	})

	t.Run("Map value at input maximum", func(t *testing.T) {
		result := fmap(100, 0, 100, 0, 255)
		if result != 255 {
			t.Errorf("Expected 255, got %f", result)
		}
	})

	t.Run("Map value beyond input range (extrapolation)", func(t *testing.T) {
		result := fmap(150, 0, 100, 0, 255)
		expected := 382.5 // Linear extrapolation
		if math.Abs(result-expected) > 0.001 {
			t.Errorf("Expected %f, got %f", expected, result)
		}
	})

	t.Run("Map with very small ranges", func(t *testing.T) {
		result := fmap(0.5, 0, 1, 0, 100)
		if math.Abs(result-50) > 0.001 {
			t.Errorf("Expected 50, got %f", result)
		}
	})
}

// TestReadSettingsDefaults tests that readSettings sets correct defaults when file doesn't exist
func TestReadSettingsDefaults(t *testing.T) {
	// Note: configLocation is a const, so we test defaults by calling readSettings()
	// which will fail to open the file and fall back to defaults

	// Call readSettings which should set defaults when file doesn't exist
	readSettings()

	// Verify defaults are set
	if myFanControl.TempTarget != defaultTempTarget {
		t.Errorf("Expected TempTarget=%f, got %f", defaultTempTarget, myFanControl.TempTarget)
	}
	if myFanControl.PWMDutyMin != defaultPwmDutyMin {
		t.Errorf("Expected PWMDutyMin=%d, got %d", defaultPwmDutyMin, myFanControl.PWMDutyMin)
	}
	if myFanControl.PWMFrequency != defaultPwmFrequency {
		t.Errorf("Expected PWMFrequency=%d, got %d", defaultPwmFrequency, myFanControl.PWMFrequency)
	}
	if myFanControl.PWMPin != defaultPin {
		t.Errorf("Expected PWMPin=%d, got %d", defaultPin, myFanControl.PWMPin)
	}

	t.Logf("Defaults verified: TempTarget=%f, PWMDutyMin=%d, PWMFrequency=%d, PWMPin=%d",
		myFanControl.TempTarget, myFanControl.PWMDutyMin, myFanControl.PWMFrequency, myFanControl.PWMPin)
}

// TestFanControlStructMarshaling tests JSON marshaling/unmarshaling of FanControl struct
func TestFanControlStructMarshaling(t *testing.T) {
	original := FanControl{
		TempTarget:           55.0,
		TempCurrent:          48.5,
		PWMDutyMin:           25,
		PWMFrequency:         64000,
		PWMDuty80PStartDelay: 500,
		PWMDutyCurrent:       75,
		PWMPin:               18,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("Failed to marshal FanControl: %v", err)
	}

	t.Logf("Marshaled JSON: %s", string(jsonData))

	// Unmarshal back
	var decoded FanControl
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal FanControl: %v", err)
	}

	// Verify all fields
	if decoded.TempTarget != original.TempTarget {
		t.Errorf("TempTarget mismatch: got %f, want %f", decoded.TempTarget, original.TempTarget)
	}
	if decoded.TempCurrent != original.TempCurrent {
		t.Errorf("TempCurrent mismatch: got %f, want %f", decoded.TempCurrent, original.TempCurrent)
	}
	if decoded.PWMDutyMin != original.PWMDutyMin {
		t.Errorf("PWMDutyMin mismatch: got %d, want %d", decoded.PWMDutyMin, original.PWMDutyMin)
	}
	if decoded.PWMFrequency != original.PWMFrequency {
		t.Errorf("PWMFrequency mismatch: got %d, want %d", decoded.PWMFrequency, original.PWMFrequency)
	}
	if decoded.PWMDuty80PStartDelay != original.PWMDuty80PStartDelay {
		t.Errorf("PWMDuty80PStartDelay mismatch: got %d, want %d", decoded.PWMDuty80PStartDelay, original.PWMDuty80PStartDelay)
	}
	if decoded.PWMDutyCurrent != original.PWMDutyCurrent {
		t.Errorf("PWMDutyCurrent mismatch: got %d, want %d", decoded.PWMDutyCurrent, original.PWMDutyCurrent)
	}
	if decoded.PWMPin != original.PWMPin {
		t.Errorf("PWMPin mismatch: got %d, want %d", decoded.PWMPin, original.PWMPin)
	}
}

// TestFanControlStructJSONFields tests that JSON field names are correct
func TestFanControlStructJSONFields(t *testing.T) {
	fc := FanControl{
		TempTarget:           50.0,
		TempCurrent:          45.0,
		PWMDutyMin:           20,
		PWMFrequency:         64000,
		PWMDuty80PStartDelay: 500,
		PWMDutyCurrent:       50,
		PWMPin:               18,
	}

	jsonData, err := json.Marshal(&fc)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(jsonData)
	t.Logf("JSON output: %s", jsonStr)

	// Verify expected field names are present
	expectedFields := []string{
		"TempTarget",
		"TempCurrent",
		"PWMDutyMin",
		"PWMFrequency",
		"PWMDuty80PStartDelay",
		"PWMDutyCurrent",
		"PWMPin",
	}

	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("Expected field '%s' not found in JSON", field)
		}
	}
}

// TestHandleStatusRequest tests the HTTP status handler
func TestHandleStatusRequest(t *testing.T) {
	// Set up test state
	myFanControl = FanControl{
		TempTarget:           55.0,
		TempCurrent:          48.5,
		PWMDutyMin:           25,
		PWMFrequency:         64000,
		PWMDuty80PStartDelay: 500,
		PWMDutyCurrent:       75,
		PWMPin:               18,
	}

	// Create a request to pass to the handler
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleStatusRequest)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check that response is valid JSON
	var decoded FanControl
	err = json.Unmarshal(rr.Body.Bytes(), &decoded)
	if err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	// Verify the decoded values match what we set
	if decoded.TempTarget != myFanControl.TempTarget {
		t.Errorf("TempTarget mismatch: got %f, want %f", decoded.TempTarget, myFanControl.TempTarget)
	}
	if decoded.TempCurrent != myFanControl.TempCurrent {
		t.Errorf("TempCurrent mismatch: got %f, want %f", decoded.TempCurrent, myFanControl.TempCurrent)
	}
	if decoded.PWMDutyCurrent != myFanControl.PWMDutyCurrent {
		t.Errorf("PWMDutyCurrent mismatch: got %d, want %d", decoded.PWMDutyCurrent, myFanControl.PWMDutyCurrent)
	}

	t.Logf("Status response: %s", rr.Body.String())
}

// TestHandleStatusRequestMultipleCalls tests that handler works across multiple calls
func TestHandleStatusRequestMultipleCalls(t *testing.T) {
	testCases := []FanControl{
		{TempTarget: 50.0, TempCurrent: 45.0, PWMDutyCurrent: 30},
		{TempTarget: 55.0, TempCurrent: 52.0, PWMDutyCurrent: 60},
		{TempTarget: 60.0, TempCurrent: 58.0, PWMDutyCurrent: 90},
	}

	handler := http.HandlerFunc(handleStatusRequest)

	for i, tc := range testCases {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			// Set global state
			myFanControl = tc

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			var decoded FanControl
			err = json.Unmarshal(rr.Body.Bytes(), &decoded)
			if err != nil {
				t.Errorf("Response is not valid JSON: %v", err)
			}

			if decoded.TempTarget != tc.TempTarget {
				t.Errorf("TempTarget mismatch: got %f, want %f", decoded.TempTarget, tc.TempTarget)
			}
			if decoded.PWMDutyCurrent != tc.PWMDutyCurrent {
				t.Errorf("PWMDutyCurrent mismatch: got %d, want %d", decoded.PWMDutyCurrent, tc.PWMDutyCurrent)
			}
		})
	}
}

// TestFanControlZeroValues tests that zero values are handled correctly
func TestFanControlZeroValues(t *testing.T) {
	fc := FanControl{} // All fields zero

	jsonData, err := json.Marshal(&fc)
	if err != nil {
		t.Fatalf("Failed to marshal zero-value FanControl: %v", err)
	}

	var decoded FanControl
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify zero values are preserved
	if decoded.TempTarget != 0.0 {
		t.Errorf("Expected TempTarget=0, got %f", decoded.TempTarget)
	}
	if decoded.PWMDutyCurrent != 0 {
		t.Errorf("Expected PWMDutyCurrent=0, got %d", decoded.PWMDutyCurrent)
	}

	t.Logf("Zero-value JSON: %s", string(jsonData))
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
