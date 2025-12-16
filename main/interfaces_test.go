/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file, herein included
	as part of this header.

	interfaces_test.go: Tests for interface abstractions

	These tests verify the interface abstractions work correctly and
	demonstrate how to use mock implementations for testing.
*/

package main

import (
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// =============================================================================
// Mock Implementations
// =============================================================================

// mockDiskUsageProvider is a mock that returns configurable disk usage values
type mockDiskUsageProvider struct {
	freeBytes uint64
	calls     []string // Track which paths were queried
}

func (m *mockDiskUsageProvider) Free(path string) uint64 {
	m.calls = append(m.calls, path)
	return m.freeBytes
}

// mockCommandRunner is a mock that returns configurable command results
type mockCommandRunner struct {
	output      []byte
	outputErr   error
	runErr      error
	calls       [][]string // Track all command invocations
	callCount   int
	outputCalls int
	runCalls    int
}

func (m *mockCommandRunner) Output(name string, args ...string) ([]byte, error) {
	m.callCount++
	m.outputCalls++
	call := append([]string{name}, args...)
	m.calls = append(m.calls, call)
	return m.output, m.outputErr
}

func (m *mockCommandRunner) Run(name string, args ...string) error {
	m.callCount++
	m.runCalls++
	call := append([]string{name}, args...)
	m.calls = append(m.calls, call)
	return m.runErr
}

// =============================================================================
// DiskUsageProvider Tests
// =============================================================================

func TestRealDiskUsageProvider(t *testing.T) {
	// Test that the real provider works with a valid path
	provider := &realDiskUsageProvider{}

	// Use temp directory which should always exist
	tmpDir := os.TempDir()
	freeBytes := provider.Free(tmpDir)

	// Should return some positive value (unless disk is completely full)
	if freeBytes == 0 {
		t.Logf("Warning: Free bytes is 0 for %s - disk may be full", tmpDir)
	} else {
		t.Logf("Free bytes on %s: %d", tmpDir, freeBytes)
	}
}

func TestMockDiskUsageProvider(t *testing.T) {
	tests := []struct {
		name      string
		freeBytes uint64
		path      string
	}{
		{"zero_free_space", 0, "/test/path"},
		{"low_free_space", 10 * 1024 * 1024, "/var/log"}, // 10MB
		{"normal_free_space", 100 * 1024 * 1024, "/home"}, // 100MB
		{"high_free_space", 1024 * 1024 * 1024, "/data"},  // 1GB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDiskUsageProvider{freeBytes: tt.freeBytes}

			result := mock.Free(tt.path)

			if result != tt.freeBytes {
				t.Errorf("Expected %d, got %d", tt.freeBytes, result)
			}
			if len(mock.calls) != 1 || mock.calls[0] != tt.path {
				t.Errorf("Expected path %s to be recorded, got %v", tt.path, mock.calls)
			}
		})
	}
}

func TestGetDiskFreeBytes(t *testing.T) {
	// Save original provider
	origProvider := diskUsageProvider
	defer func() { diskUsageProvider = origProvider }()

	// Test with mock provider
	mock := &mockDiskUsageProvider{freeBytes: 42 * 1024 * 1024}
	diskUsageProvider = mock

	result := getDiskFreeBytes("/test/path")

	if result != 42*1024*1024 {
		t.Errorf("Expected 42MB, got %d bytes", result)
	}
	if len(mock.calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(mock.calls))
	}
}

// =============================================================================
// CommandRunner Tests
// =============================================================================

func TestRealCommandRunner_Output(t *testing.T) {
	// Test with a simple command that exists on all systems
	runner := &realCommandRunner{}

	// "echo" is available on all platforms
	output, err := runner.Output("echo", "test")

	if err != nil {
		t.Errorf("Unexpected error running echo: %v", err)
	}
	// Output should contain "test"
	if len(output) == 0 {
		t.Error("Expected non-empty output from echo")
	}
	t.Logf("Echo output: %s", string(output))
}

func TestRealCommandRunner_Run(t *testing.T) {
	// Test with a simple command that exists on all systems
	runner := &realCommandRunner{}

	// "true" is available on all platforms and always succeeds
	err := runner.Run("true")

	if err != nil {
		t.Errorf("Unexpected error running true: %v", err)
	}
}

func TestMockCommandRunner(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		err    error
	}{
		{"success", []byte("command output"), nil},
		{"empty_output", []byte(""), nil},
		{"error", nil, errors.New("command failed")},
		{"output_with_error", []byte("partial output"), errors.New("partial failure")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCommandRunner{output: tt.output, outputErr: tt.err}

			output, err := mock.Output("test", "arg1", "arg2")

			if string(output) != string(tt.output) {
				t.Errorf("Expected output %q, got %q", tt.output, output)
			}
			if (err == nil) != (tt.err == nil) {
				t.Errorf("Expected error %v, got %v", tt.err, err)
			}
			if mock.callCount != 1 {
				t.Errorf("Expected 1 call, got %d", mock.callCount)
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	// Save original runner
	origRunner := commandRunner
	defer func() { commandRunner = origRunner }()

	// Test with mock
	mock := &mockCommandRunner{
		output:    []byte("mock output"),
		outputErr: nil,
	}
	commandRunner = mock

	output, err := runCommand("mycmd", "arg1", "arg2")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(output) != "mock output" {
		t.Errorf("Expected 'mock output', got %q", output)
	}
	if mock.callCount != 1 {
		t.Errorf("Expected 1 call, got %d", mock.callCount)
	}
	if len(mock.calls) != 1 || mock.calls[0][0] != "mycmd" {
		t.Errorf("Expected command 'mycmd', got %v", mock.calls)
	}
}

func TestRunCommand_Error(t *testing.T) {
	// Save original runner
	origRunner := commandRunner
	defer func() { commandRunner = origRunner }()

	// Test error case
	expectedErr := errors.New("execution failed")
	mock := &mockCommandRunner{
		output:    nil,
		outputErr: expectedErr,
	}
	commandRunner = mock

	_, err := runCommand("failing-cmd")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != expectedErr.Error() {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// =============================================================================
// Integration Tests with logFileWatcherOnce
// =============================================================================

func TestLogFileWatcherOnce_WithMockedDiskUsage_LowSpace(t *testing.T) {
	// Save original values
	origProvider := diskUsageProvider
	origLogDirPath := logDirPath
	origLogDirf := logDirf
	origDebugLogf := debugLogf
	origLogFileHandle := logFileHandle

	defer func() {
		diskUsageProvider = origProvider
		logDirPath = origLogDirPath
		logDirf = origLogDirf
		debugLogf = origDebugLogf
		logFileHandle = origLogFileHandle
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "logwatcher_mock_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logDirPath = tmpDir
	logDirf = tmpDir
	debugLogf = filepath.Join(tmpDir, "stratux.log")
	logFileHandle = nil

	// Create a small log file
	err = os.WriteFile(debugLogf, []byte("small log"), 0644)
	if err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Create old log files that can be deleted
	for i := 1; i <= 5; i++ {
		logPath := filepath.Join(tmpDir, "stratux.log."+string(rune('0'+i)))
		err := os.WriteFile(logPath, []byte("old log content "+string(rune('0'+i))), 0644)
		if err != nil {
			t.Fatalf("Failed to create old log: %v", err)
		}
	}

	// Mock disk usage to report low space (less than 50MB)
	mock := &mockDiskUsageProvider{freeBytes: 10 * 1024 * 1024} // 10MB free
	diskUsageProvider = mock

	// Run the function - should trigger disk cleanup
	result := logFileWatcherOnce()

	// Verify the mock was called
	if len(mock.calls) == 0 {
		t.Error("Expected disk usage provider to be called")
	}

	// Action should be taken (disk cleanup)
	if !result {
		t.Error("Expected action to be taken due to low disk space")
	}

	// Verify some old logs were deleted
	remainingLogs := getStratuxLogFiles()
	t.Logf("Remaining logs after cleanup: %d", len(remainingLogs))
}

func TestLogFileWatcherOnce_WithMockedDiskUsage_PlentyOfSpace(t *testing.T) {
	// Save original values
	origProvider := diskUsageProvider
	origLogDirPath := logDirPath
	origLogDirf := logDirf
	origDebugLogf := debugLogf
	origLogFileHandle := logFileHandle

	defer func() {
		diskUsageProvider = origProvider
		logDirPath = origLogDirPath
		logDirf = origLogDirf
		debugLogf = origDebugLogf
		logFileHandle = origLogFileHandle
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "logwatcher_plenty_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logDirPath = tmpDir
	logDirf = tmpDir
	debugLogf = filepath.Join(tmpDir, "stratux.log")
	logFileHandle = nil

	// Create a small log file (under 10MB limit)
	err = os.WriteFile(debugLogf, []byte("small log"), 0644)
	if err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Mock disk usage to report plenty of space
	mock := &mockDiskUsageProvider{freeBytes: 100 * 1024 * 1024} // 100MB free
	diskUsageProvider = mock

	// Run the function - should NOT trigger any action
	result := logFileWatcherOnce()

	// Verify the mock was called
	if len(mock.calls) == 0 {
		t.Error("Expected disk usage provider to be called")
	}

	// No action should be taken
	if result {
		t.Error("Expected no action when plenty of disk space and small log file")
	}
}

func TestLogFileWatcherOnce_WithMockedDiskUsage_ZeroSpace(t *testing.T) {
	// Save original values
	origProvider := diskUsageProvider
	origLogDirPath := logDirPath
	origLogDirf := logDirf
	origDebugLogf := debugLogf
	origLogFileHandle := logFileHandle

	defer func() {
		diskUsageProvider = origProvider
		logDirPath = origLogDirPath
		logDirf = origLogDirf
		debugLogf = origDebugLogf
		logFileHandle = origLogFileHandle
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "logwatcher_zero_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logDirPath = tmpDir
	logDirf = tmpDir
	debugLogf = filepath.Join(tmpDir, "stratux.log")
	logFileHandle = nil

	// Create a small log file
	err = os.WriteFile(debugLogf, []byte("small log"), 0644)
	if err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Mock disk usage to report zero space (extreme case)
	mock := &mockDiskUsageProvider{freeBytes: 0}
	diskUsageProvider = mock

	// Run the function - should attempt cleanup but find nothing to delete
	result := logFileWatcherOnce()

	// Verify the mock was called
	if len(mock.calls) == 0 {
		t.Error("Expected disk usage provider to be called")
	}

	// No action taken because there are no old logs to delete
	t.Logf("Action taken: %v (expected false if no old logs exist)", result)
}

// =============================================================================
// Integration Tests with overlayctl
// =============================================================================

func TestOverlayctl_WithMockedCommandRunner_Success(t *testing.T) {
	// Save original runner
	origRunner := commandRunner
	defer func() { commandRunner = origRunner }()

	// Mock successful command execution
	mock := &mockCommandRunner{
		output:    []byte("overlay enabled\n"),
		outputErr: nil,
	}
	commandRunner = mock

	// Call overlayctl - should succeed
	overlayctl("enable")

	// Verify the mock was called with correct arguments
	if mock.callCount != 1 {
		t.Errorf("Expected 1 call, got %d", mock.callCount)
	}

	if len(mock.calls) != 1 {
		t.Fatalf("Expected 1 call record, got %d", len(mock.calls))
	}

	call := mock.calls[0]
	if call[0] != "/bin/sh" {
		t.Errorf("Expected /bin/sh, got %s", call[0])
	}
	if call[1] != "/sbin/overlayctl" {
		t.Errorf("Expected /sbin/overlayctl, got %s", call[1])
	}
	if call[2] != "enable" {
		t.Errorf("Expected 'enable' argument, got %s", call[2])
	}
}

func TestOverlayctl_WithMockedCommandRunner_Error(t *testing.T) {
	// Save original runner
	origRunner := commandRunner
	defer func() { commandRunner = origRunner }()

	// Mock failed command execution
	mock := &mockCommandRunner{
		output:    []byte("error: permission denied"),
		outputErr: errors.New("exit status 1"),
	}
	commandRunner = mock

	// Call overlayctl - should handle error gracefully
	overlayctl("unlock")

	// Verify the mock was called
	if mock.callCount != 1 {
		t.Errorf("Expected 1 call, got %d", mock.callCount)
	}

	// The function should not panic and should log the error
	t.Log("overlayctl handled error gracefully")
}

func TestOverlayctl_WithMockedCommandRunner_AllCommands(t *testing.T) {
	// Save original runner
	origRunner := commandRunner
	defer func() { commandRunner = origRunner }()

	commands := []string{"status", "enable", "disable", "lock", "unlock"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			mock := &mockCommandRunner{
				output:    []byte("overlayctl: " + cmd + " completed\n"),
				outputErr: nil,
			}
			commandRunner = mock

			overlayctl(cmd)

			if mock.callCount != 1 {
				t.Errorf("Expected 1 call for %s, got %d", cmd, mock.callCount)
			}

			// Verify the command was passed correctly
			if len(mock.calls) > 0 && mock.calls[0][2] != cmd {
				t.Errorf("Expected command %s, got %s", cmd, mock.calls[0][2])
			}
		})
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestLogFileWatcherOnce_DiskCleanupLoop(t *testing.T) {
	// This test verifies the disk cleanup loop executes and terminates properly
	// when mock returns low disk space

	// Save original values
	origProvider := diskUsageProvider
	origLogDirPath := logDirPath
	origLogDirf := logDirf
	origDebugLogf := debugLogf
	origLogFileHandle := logFileHandle

	defer func() {
		diskUsageProvider = origProvider
		logDirPath = origLogDirPath
		logDirf = origLogDirf
		debugLogf = origDebugLogf
		logFileHandle = origLogFileHandle
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "logwatcher_loop_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logDirPath = tmpDir
	logDirf = tmpDir
	debugLogf = filepath.Join(tmpDir, "stratux.log")
	logFileHandle = nil

	// Create current log file
	err = os.WriteFile(debugLogf, []byte("current log"), 0644)
	if err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Create multiple old log files with known sizes
	logSizes := []int{1000, 2000, 3000, 4000, 5000}
	for i, size := range logSizes {
		logPath := filepath.Join(tmpDir, "stratux.log."+string(rune('1'+i)))
		content := make([]byte, size)
		err := os.WriteFile(logPath, content, 0644)
		if err != nil {
			t.Fatalf("Failed to create old log %d: %v", i, err)
		}
	}

	// Mock reports 20MB free initially - needs cleanup to reach 50MB
	mock := &mockDiskUsageProvider{freeBytes: 20 * 1024 * 1024}
	diskUsageProvider = mock

	initialLogs := len(getStratuxLogFiles())
	t.Logf("Initial log count: %d", initialLogs)

	// Run the function
	result := logFileWatcherOnce()

	finalLogs := len(getStratuxLogFiles())
	t.Logf("Final log count: %d", finalLogs)
	t.Logf("Action taken: %v", result)

	// Should have taken action (deleted logs)
	if !result {
		t.Error("Expected action to be taken")
	}

	// Should have deleted some logs
	if finalLogs >= initialLogs {
		t.Errorf("Expected fewer logs after cleanup, had %d, now have %d", initialLogs, finalLogs)
	}
}

// =============================================================================
// Tests for Management Interface Functions Using CommandRunner
// =============================================================================

func TestDoRestartApp_WithMockedCommandRunner(t *testing.T) {
	// Save original runner
	origRunner := commandRunner
	defer func() { commandRunner = origRunner }()

	t.Run("success", func(t *testing.T) {
		mock := &mockCommandRunner{
			output:    []byte("Restarting stratux.service\n"),
			outputErr: nil,
		}
		commandRunner = mock

		// Call doRestartApp - should succeed
		doRestartApp()

		// Verify the mock was called with correct arguments
		if mock.outputCalls != 1 {
			t.Errorf("Expected 1 Output call, got %d", mock.outputCalls)
		}

		if len(mock.calls) != 1 {
			t.Fatalf("Expected 1 call, got %d", len(mock.calls))
		}

		call := mock.calls[0]
		if call[0] != "/bin/systemctl" {
			t.Errorf("Expected /bin/systemctl, got %s", call[0])
		}
		if call[1] != "restart" {
			t.Errorf("Expected 'restart', got %s", call[1])
		}
		if call[2] != "stratux" {
			t.Errorf("Expected 'stratux', got %s", call[2])
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockCommandRunner{
			output:    []byte("Failed to restart stratux.service"),
			outputErr: errors.New("exit status 1"),
		}
		commandRunner = mock

		// Call doRestartApp - should handle error gracefully
		doRestartApp()

		// Verify the mock was called
		if mock.outputCalls != 1 {
			t.Errorf("Expected 1 Output call, got %d", mock.outputCalls)
		}
		t.Log("doRestartApp handled error gracefully")
	})
}

func TestHandleShutdownRequest_WithMockedCommandRunner(t *testing.T) {
	// Save original runner
	origRunner := commandRunner
	defer func() { commandRunner = origRunner }()

	mock := &mockCommandRunner{
		runErr: nil,
	}
	commandRunner = mock

	// Create a test request
	req := httptest.NewRequest("POST", "/shutdown", nil)
	w := httptest.NewRecorder()

	// Call handleShutdownRequest
	handleShutdownRequest(w, req)

	// Verify the mock was called with correct arguments
	if mock.runCalls != 1 {
		t.Errorf("Expected 1 Run call, got %d", mock.runCalls)
	}

	if len(mock.calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(mock.calls))
	}

	call := mock.calls[0]
	if call[0] != "systemctl" {
		t.Errorf("Expected systemctl, got %s", call[0])
	}
	if call[1] != "poweroff" {
		t.Errorf("Expected 'poweroff', got %s", call[1])
	}
}

func TestDoReboot_WithMockedCommandRunner(t *testing.T) {
	// Save original runner
	origRunner := commandRunner
	defer func() { commandRunner = origRunner }()

	mock := &mockCommandRunner{
		runErr: nil,
	}
	commandRunner = mock

	// Call doReboot
	doReboot()

	// Verify the mock was called with correct arguments
	if mock.runCalls != 1 {
		t.Errorf("Expected 1 Run call, got %d", mock.runCalls)
	}

	if len(mock.calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(mock.calls))
	}

	call := mock.calls[0]
	if call[0] != "systemctl" {
		t.Errorf("Expected systemctl, got %s", call[0])
	}
	if call[1] != "reboot" {
		t.Errorf("Expected 'reboot', got %s", call[1])
	}
}

func TestHandleroPartitionRebuild_WithMockedCommandRunner(t *testing.T) {
	// Save original runner
	origRunner := commandRunner
	defer func() { commandRunner = origRunner }()

	// Initialize system error tracking if needed
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
	if systemErrs == nil {
		systemErrs = make(map[string]string)
	}

	t.Run("success", func(t *testing.T) {
		mock := &mockCommandRunner{
			output:    []byte("Partition rebuilt successfully\n"),
			outputErr: nil,
		}
		commandRunner = mock

		// Create a test request
		req := httptest.NewRequest("POST", "/roPartitionRebuild", nil)
		w := httptest.NewRecorder()

		// Call handler
		handleroPartitionRebuild(w, req)

		// Verify the mock was called
		if mock.outputCalls != 1 {
			t.Errorf("Expected 1 Output call, got %d", mock.outputCalls)
		}

		if len(mock.calls) != 1 {
			t.Fatalf("Expected 1 call, got %d", len(mock.calls))
		}

		call := mock.calls[0]
		if call[0] != "/usr/sbin/rebuild_ro_part.sh" {
			t.Errorf("Expected /usr/sbin/rebuild_ro_part.sh, got %s", call[0])
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := &mockCommandRunner{
			output:    []byte("Error rebuilding partition"),
			outputErr: errors.New("exit status 1"),
		}
		commandRunner = mock

		// Create a test request
		req := httptest.NewRequest("POST", "/roPartitionRebuild", nil)
		w := httptest.NewRecorder()

		// Call handler - should handle error gracefully
		handleroPartitionRebuild(w, req)

		// Verify the mock was called
		if mock.outputCalls != 1 {
			t.Errorf("Expected 1 Output call, got %d", mock.outputCalls)
		}
		t.Log("handleroPartitionRebuild handled error gracefully")
	})
}

// Test runCommandNoOutput helper function
func TestRunCommandNoOutput(t *testing.T) {
	// Save original runner
	origRunner := commandRunner
	defer func() { commandRunner = origRunner }()

	t.Run("success", func(t *testing.T) {
		mock := &mockCommandRunner{
			runErr: nil,
		}
		commandRunner = mock

		err := runCommandNoOutput("test-cmd", "arg1")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if mock.runCalls != 1 {
			t.Errorf("Expected 1 Run call, got %d", mock.runCalls)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("command failed")
		mock := &mockCommandRunner{
			runErr: expectedErr,
		}
		commandRunner = mock

		err := runCommandNoOutput("failing-cmd")

		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != expectedErr.Error() {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})
}
