package main

import (
	"bytes"
	"log"
	"testing"
)

// TestLoggingFunctions tests the logging wrapper functions
func TestLoggingFunctions(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput) // Restore original output

	t.Run("logInf", func(t *testing.T) {
		buf.Reset()
		logInf("Test info message: %s", "hello")
		output := buf.String()
		if output == "" {
			t.Error("Expected logInf to produce output")
		}
		t.Logf("logInf output: %s", output)
	})

	t.Run("logErr", func(t *testing.T) {
		buf.Reset()
		logErr("Test error message: %d", 42)
		output := buf.String()
		if output == "" {
			t.Error("Expected logErr to produce output")
		}
		t.Logf("logErr output: %s", output)
	})

	t.Run("logDbg_disabled", func(t *testing.T) {
		originalDebug := globalSettings.DEBUG
		defer func() { globalSettings.DEBUG = originalDebug }()

		globalSettings.DEBUG = false
		buf.Reset()
		logDbg("This should not appear: %s", "hidden")
		output := buf.String()
		if output != "" {
			t.Error("Expected logDbg to produce no output when DEBUG=false")
		}
		t.Log("logDbg with DEBUG=false correctly produced no output")
	})

	t.Run("logDbg_enabled", func(t *testing.T) {
		originalDebug := globalSettings.DEBUG
		defer func() { globalSettings.DEBUG = originalDebug }()

		globalSettings.DEBUG = true
		buf.Reset()
		logDbg("This should appear: %s", "visible")
		output := buf.String()
		if output == "" {
			t.Error("Expected logDbg to produce output when DEBUG=true")
		}
		t.Logf("logDbg with DEBUG=true output: %s", output)
	})
}
