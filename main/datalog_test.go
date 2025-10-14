package main

import (
	"reflect"
	"testing"
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
