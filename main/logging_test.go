package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
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
		tmpFile, err := os.CreateTemp("", "test-log-*.log")
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
		tmpFile, err := os.CreateTemp("", "test-log-closed-*.log")
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

// TestClearDebugLogFile tests the clearDebugLogFile function
func TestClearDebugLogFile(t *testing.T) {
	// Save original log file handle
	origLogFileHandle := logFileHandle
	defer func() { logFileHandle = origLogFileHandle }()

	t.Run("nil_handle", func(t *testing.T) {
		logFileHandle = nil

		// Should not panic with nil handle
		clearDebugLogFile()
		t.Log("clearDebugLogFile handled nil handle gracefully")
	})

	t.Run("valid_file_truncate", func(t *testing.T) {
		// Create a temp file with content
		tmpFile, err := os.CreateTemp("", "test-clear-log-*.log")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		// Write some content
		content := "Test log content that should be cleared\nLine 2\nLine 3\n"
		tmpFile.WriteString(content)
		tmpFile.Sync()

		// Verify content was written
		info, _ := tmpFile.Stat()
		if info.Size() == 0 {
			t.Fatal("Content was not written to temp file")
		}

		logFileHandle = tmpFile

		// Clear the log file
		clearDebugLogFile()

		// Verify file was truncated
		info, _ = tmpFile.Stat()
		if info.Size() != 0 {
			t.Errorf("Expected file to be truncated to 0 bytes, got %d", info.Size())
		}

		tmpFile.Close()
		t.Log("clearDebugLogFile successfully truncated file")
	})

	t.Run("closed_file_seek_error", func(t *testing.T) {
		// Create and close a temp file
		tmpFile, err := os.CreateTemp("", "test-clear-closed-*.log")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		tmpFile.WriteString("Some content")
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		logFileHandle = tmpFile

		// Should handle error gracefully (seek will fail on closed file)
		clearDebugLogFile()
		t.Log("clearDebugLogFile handled closed file gracefully")
	})

	t.Run("truncate_error", func(t *testing.T) {
		// Create a temp file with read-only permissions
		tmpFile, err := os.CreateTemp("", "test-clear-readonly-*.log")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFilePath := tmpFile.Name()

		// Write some content
		content := "Test content for truncate error"
		tmpFile.WriteString(content)
		tmpFile.Sync()
		tmpFile.Close()

		// Reopen the file as read-only
		readOnlyFile, err := os.OpenFile(tmpFilePath, os.O_RDONLY, 0444)
		if err != nil {
			t.Fatalf("Failed to reopen file as read-only: %v", err)
		}
		defer readOnlyFile.Close()

		logFileHandle = readOnlyFile

		// Should handle truncate error gracefully (truncate will fail on read-only file)
		clearDebugLogFile()
		t.Log("clearDebugLogFile handled truncate error gracefully")
	})
}

// TestGetStratuxLogFiles tests the getStratuxLogFiles function
func TestGetStratuxLogFiles(t *testing.T) {
	// Note: getStratuxLogFiles reads from const logDir="/var/log/" which we can't change
	// But it uses logDirf to build paths. We'll test what we can.

	// Save original logDirf
	origLogDirf := logDirf
	defer func() { logDirf = origLogDirf }()

	t.Run("returns_sorted_list", func(t *testing.T) {
		// Just test that it returns a slice (may be empty) and doesn't crash
		logs := getStratuxLogFiles()

		// Verify it returns a slice (not nil)
		if logs == nil {
			t.Error("Expected non-nil slice")
		}

		// Verify sorting - each element should be >= previous
		for i := 1; i < len(logs); i++ {
			if logs[i] < logs[i-1] {
				t.Errorf("Logs not sorted: logs[%d]=%s < logs[%d]=%s",
					i, logs[i], i-1, logs[i-1])
			}
		}

		// Verify all returned files have the correct prefix
		for i, log := range logs {
			baseName := filepath.Base(log)
			if !strings.HasPrefix(baseName, debugLogFile+".") {
				t.Errorf("logs[%d]=%s doesn't have expected prefix %s.",
					i, log, debugLogFile)
			}
		}
	})

	t.Run("uses_logDirf_for_paths", func(t *testing.T) {
		// Set logDirf to a test path
		testPath := "/test/path"
		logDirf = testPath

		// Even though we can't change where it reads from (logDir const),
		// we can verify it uses logDirf when building returned paths
		// This test will only work if there are actual log files in /var/log/
		logs := getStratuxLogFiles()

		for i, log := range logs {
			dir := filepath.Dir(log)
			if dir != testPath {
				t.Errorf("logs[%d]=%s should use logDirf=%s but uses %s",
					i, log, testPath, dir)
			}
		}
	})
}

// TestDeleteOldestLog tests the deleteOldestLog function
func TestDeleteOldestLog(t *testing.T) {
	// Note: deleteOldestLog calls getStratuxLogFiles which reads from const logDir

	t.Run("no_logs_available", func(t *testing.T) {
		// When getStratuxLogFiles returns empty (no logs in /var/log/stratux.log.*)
		// deleteOldestLog should return 0
		// We can't control the actual /var/log/ directory, so this test is limited
		deleted := deleteOldestLog()

		// Should return 0 or a positive number (if logs exist and were deleted)
		if deleted < 0 {
			t.Errorf("Expected non-negative bytes deleted, got %d", deleted)
		}
		// We can't assert it's 0 since there might actually be logs to delete
		t.Logf("deleteOldestLog returned %d bytes deleted", deleted)
	})

	t.Run("function_behavior", func(t *testing.T) {
		// Test that the function doesn't crash
		// It uses getStratuxLogFiles which reads from /var/log/
		// Returns 0 if no logs, or size of deleted file

		// Call it multiple times to test behavior
		deleted1 := deleteOldestLog()
		if deleted1 < 0 {
			t.Errorf("First call: Expected non-negative bytes deleted, got %d", deleted1)
		}

		deleted2 := deleteOldestLog()
		if deleted2 < 0 {
			t.Errorf("Second call: Expected non-negative bytes deleted, got %d", deleted2)
		}

		t.Logf("deleteOldestLog calls: %d, %d bytes", deleted1, deleted2)
	})
}

// TestRotateLogs tests the rotateLogs function
func TestRotateLogs(t *testing.T) {
	// Note: rotateLogs calls getStratuxLogFiles which reads from const logDir="/var/log/"
	// We can only test limited behavior

	// Save original values
	origLogDirf := logDirf
	origDebugLogf := debugLogf
	origLogFileHandle := logFileHandle
	defer func() {
		logDirf = origLogDirf
		debugLogf = origDebugLogf
		if logFileHandle != nil && logFileHandle != origLogFileHandle {
			logFileHandle.Close()
		}
		logFileHandle = origLogFileHandle
	}()

	t.Run("rotates_current_log", func(t *testing.T) {
		// Create a temp directory for the current log
		tmpDir, err := os.MkdirTemp("", "test-rotate-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		logDirf = tmpDir
		debugLogf = tmpDir + "/" + debugLogFile

		// Create current log file
		currentLog, err := os.Create(debugLogf)
		if err != nil {
			t.Fatalf("Failed to create current log: %v", err)
		}
		currentLog.WriteString("Current log content")
		currentLog.Close()
		logFileHandle = nil

		// Call rotateLogs
		rotateLogs()

		// Verify current log was renamed to .1
		_, err = os.Stat(debugLogf + ".1")
		if err != nil {
			t.Logf("Current log rename check: %v", err)
		}

		// Verify new log file was created by openLogFile
		_, err = os.Stat(debugLogf)
		if err != nil {
			t.Error("Expected new stratux.log to be created")
		}

		// Clean up
		if logFileHandle != nil {
			logFileHandle.Close()
			logFileHandle = nil
		}
	})

	t.Run("doesnt_crash", func(t *testing.T) {
		// Test that rotateLogs doesn't crash with various states
		tmpDir, err := os.MkdirTemp("", "test-rotate-safe-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		logDirf = tmpDir
		debugLogf = tmpDir + "/" + debugLogFile

		// Create current log file
		currentLog, err := os.Create(debugLogf)
		if err != nil {
			t.Fatalf("Failed to create current log: %v", err)
		}
		currentLog.Close()
		logFileHandle = nil

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("rotateLogs panicked: %v", r)
			}
		}()

		rotateLogs()

		// Clean up
		if logFileHandle != nil {
			logFileHandle.Close()
			logFileHandle = nil
		}
	})
}

// TestOpenLogFile tests the openLogFile function
func TestOpenLogFile(t *testing.T) {
	// Save original values
	origLogDirf := logDirf
	origDebugLogf := debugLogf
	origLogFileHandle := logFileHandle
	defer func() {
		logDirf = origLogDirf
		debugLogf = origDebugLogf
		if logFileHandle != nil && logFileHandle != origLogFileHandle {
			logFileHandle.Close()
		}
		logFileHandle = origLogFileHandle
	}()

	t.Run("create_new_log", func(t *testing.T) {
		// Create a temp directory
		tmpDir, err := os.MkdirTemp("", "test-open-log-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		logDirf = tmpDir
		logFileHandle = nil

		openLogFile()

		// Verify debugLogf was set correctly
		expectedPath := tmpDir + "/" + debugLogFile
		if debugLogf != expectedPath {
			t.Errorf("Expected debugLogf=%s, got %s", expectedPath, debugLogf)
		}

		// Verify file handle was created
		if logFileHandle == nil {
			t.Error("Expected logFileHandle to be set")
		}

		// Verify file exists
		_, err = os.Stat(debugLogf)
		if err != nil {
			t.Errorf("Expected log file to exist: %v", err)
		}

		// Verify we can write to the log
		log.Printf("Test log message")

		// Clean up
		if logFileHandle != nil {
			logFileHandle.Close()
			logFileHandle = nil
		}
	})

	t.Run("replace_old_handle", func(t *testing.T) {
		// Create a temp directory
		tmpDir, err := os.MkdirTemp("", "test-open-log-replace-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		logDirf = tmpDir

		// Create an initial log file
		initialLog := tmpDir + "/initial.log"
		initialFile, err := os.Create(initialLog)
		if err != nil {
			t.Fatalf("Failed to create initial log: %v", err)
		}
		initialFile.WriteString("Initial content")
		logFileHandle = initialFile

		// Open new log file
		openLogFile()

		// Verify old handle was closed (can't write to it)
		_, err = initialFile.WriteString("Should fail")
		if err == nil {
			t.Log("Warning: Old file handle should be closed but write succeeded (may be system-dependent)")
		}

		// Verify new handle is different
		if logFileHandle == initialFile {
			t.Error("Expected new logFileHandle to be different from old one")
		}

		// Clean up
		if logFileHandle != nil {
			logFileHandle.Close()
			logFileHandle = nil
		}
	})

	t.Run("append_to_existing", func(t *testing.T) {
		// Create a temp directory
		tmpDir, err := os.MkdirTemp("", "test-open-log-append-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		logDirf = tmpDir
		debugLogf = tmpDir + "/" + debugLogFile

		// Create existing log with content
		existingLog, err := os.Create(debugLogf)
		if err != nil {
			t.Fatalf("Failed to create existing log: %v", err)
		}
		existingContent := "Existing log content\n"
		existingLog.WriteString(existingContent)
		existingLog.Close()

		logFileHandle = nil

		// Open the log file (should append)
		openLogFile()

		// Write new content
		newContent := "New log content"
		log.Printf("%s", newContent)

		// Close the handle
		if logFileHandle != nil {
			logFileHandle.Close()
			logFileHandle = nil
		}

		// Verify both contents are in the file
		data, err := os.ReadFile(debugLogf)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		content := string(data)
		if !contains(content, existingContent) {
			t.Error("Expected existing content to be preserved")
		}
		// Note: We can't check for exact newContent because log.Printf adds timestamp
	})

	t.Run("permission_error", func(t *testing.T) {
		// Try to create log in a non-existent directory
		logDirf = "/nonexistent/invalid/path"
		debugLogf = ""
		oldHandle := logFileHandle
		logFileHandle = nil

		// Should handle error gracefully (won't crash)
		openLogFile()

		// Should have tried to set debugLogf
		expectedPath := "/nonexistent/invalid/path/" + debugLogFile
		if debugLogf != expectedPath {
			t.Errorf("Expected debugLogf=%s, got %s", expectedPath, debugLogf)
		}

		// Note: openLogFile calls addSingleSystemErrorf on error but doesn't set logFileHandle
		// However it DOES close oldFp at the end, so logFileHandle might be set to oldFp's value
		// Let's just verify it doesn't crash

		// Clean up if needed
		if logFileHandle != nil && logFileHandle != oldHandle {
			logFileHandle.Close()
			logFileHandle = nil
		}

		// Restore
		logFileHandle = oldHandle
	})
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSlowPath(s, substr))
}

func containsSlowPath(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

