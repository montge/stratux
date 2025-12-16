/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file, herein included
	as part of this header.

	interfaces.go: Interface abstractions for testability

	This file provides interface abstractions that allow mocking of
	system dependencies (disk usage, command execution) in unit tests.
*/

package main

import (
	"os/exec"

	"github.com/ricochet2200/go-disk-usage/du"
)

// =============================================================================
// DiskUsageProvider Interface
// =============================================================================

// DiskUsageProvider abstracts disk usage checking for testability.
// The default implementation uses the du package, but tests can
// provide a mock implementation that returns controlled values.
type DiskUsageProvider interface {
	// Free returns the number of free bytes on the filesystem
	Free(path string) uint64
}

// realDiskUsageProvider is the production implementation using the du package
type realDiskUsageProvider struct{}

func (r *realDiskUsageProvider) Free(path string) uint64 {
	usage := du.NewDiskUsage(path)
	return usage.Free()
}

// diskUsageProvider is the global instance used by the application.
// Tests can replace this with a mock implementation.
var diskUsageProvider DiskUsageProvider = &realDiskUsageProvider{}

// =============================================================================
// CommandRunner Interface
// =============================================================================

// CommandRunner abstracts shell command execution for testability.
// The default implementation uses os/exec, but tests can provide
// a mock implementation that simulates command output.
type CommandRunner interface {
	// Run executes a command and returns its output and any error
	Run(name string, args ...string) ([]byte, error)
}

// realCommandRunner is the production implementation using os/exec
type realCommandRunner struct{}

func (r *realCommandRunner) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// commandRunner is the global instance used by the application.
// Tests can replace this with a mock implementation.
var commandRunner CommandRunner = &realCommandRunner{}

// =============================================================================
// Helper Functions Using Interfaces
// =============================================================================

// getDiskFreeBytes returns the free bytes at the given path using the
// configured DiskUsageProvider. This allows tests to mock disk usage.
func getDiskFreeBytes(path string) uint64 {
	return diskUsageProvider.Free(path)
}

// runCommand executes a shell command using the configured CommandRunner.
// This allows tests to mock command execution.
func runCommand(name string, args ...string) ([]byte, error) {
	return commandRunner.Run(name, args...)
}
