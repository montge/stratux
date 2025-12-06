package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
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

// TestLogFileSize tests the logFileSize function
func TestLogFileSize(t *testing.T) {
	// Save original log file handle
	origLogFileHandle := logFileHandle
	defer func() { logFileHandle = origLogFileHandle }()

	t.Run("nil_handle", func(t *testing.T) {
		logFileHandle = nil
		size := logFileSize()
		if size != 0 {
			t.Errorf("Expected 0 for nil handle, got %d", size)
		}
	})

	t.Run("valid_file", func(t *testing.T) {
		// Create a temp file
		tmpFile, err := ioutil.TempFile("", "test-log-*.log")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		// Write some content
		content := "Test log content\nLine 2\nLine 3\n"
		tmpFile.WriteString(content)

		logFileHandle = tmpFile

		size := logFileSize()
		expectedSize := int64(len(content))
		if size != expectedSize {
			t.Errorf("Expected size %d, got %d", expectedSize, size)
		}

		tmpFile.Close()
	})

	t.Run("closed_file", func(t *testing.T) {
		// Create and close a temp file
		tmpFile, err := ioutil.TempFile("", "test-log-closed-*.log")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		tmpFile.WriteString("Some content")
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		logFileHandle = tmpFile

		// logFileSize should return 0 on error from Stat
		size := logFileSize()
		if size != 0 {
			t.Errorf("Expected 0 for closed file handle, got %d", size)
		}
	})
}

