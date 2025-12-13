/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file, herein included
	as part of this header.

	config_paths.go: Configurable path variables for testability

	These variables allow tests to override the default paths used by
	various functions. In production, these remain at their default values.
	In tests, they can be temporarily changed to use temporary directories.
*/

package main

// Path variables that can be overridden in tests
// Default values match the original constants
var (
	// stratuxHome is the base directory for Stratux data files
	// Default: /opt/stratux
	stratuxHome = "/opt/stratux"

	// logDirPath is the directory for log files used by getStratuxLogFiles
	// Default: /var/log/
	logDirPath = "/var/log/"

	// varLogDirPath is the directory for log files used by managementinterface
	// Default: /var/log
	varLogDirPath = "/var/log"
)

// resetPathsToDefaults restores all path variables to their default values
// This should be called in test cleanup to avoid affecting other tests
func resetPathsToDefaults() {
	stratuxHome = "/opt/stratux"
	logDirPath = "/var/log/"
	varLogDirPath = "/var/log"
}
