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

// URL path constants for HTTP routing
const (
	// mapdataStylesURLPath is the URL path prefix for map style resources
	mapdataStylesURLPath = "/mapdata/styles/"
)

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

	// dhcpLeaseDirPath is the directory for DHCP lease files
	// Default: /var/lib/misc
	dhcpLeaseDirPath = "/var/lib/misc"

	// dhcpLeaseFilePath is the path to the DHCP lease file
	// Default: /var/lib/misc/dnsmasq.leases
	dhcpLeaseFilePath = "/var/lib/misc/dnsmasq.leases"

	// arpFilePath is the path to the ARP table file
	// Default: /proc/net/arp
	arpFilePath = "/proc/net/arp"

	// extraHostsFilePath is the path to static hosts file
	// Default: /etc/stratux-static-hosts.conf
	extraHostsFilePath = "/etc/stratux-static-hosts.conf"
)

// resetPathsToDefaults restores all path variables to their default values
// This should be called in test cleanup to avoid affecting other tests
func resetPathsToDefaults() {
	stratuxHome = "/opt/stratux"
	logDirPath = "/var/log/"
	varLogDirPath = "/var/log"
	dhcpLeaseDirPath = "/var/lib/misc"
	dhcpLeaseFilePath = "/var/lib/misc/dnsmasq.leases"
	arpFilePath = "/proc/net/arp"
	extraHostsFilePath = "/etc/stratux-static-hosts.conf"
}

// getMapdataPath returns the path to the mapdata directory
// Uses the configurable stratuxHome variable
func getMapdataPath() string {
	return stratuxHome + "/mapdata/"
}

// getMapdataStylesPath returns the path to the mapdata styles directory
// Uses the configurable stratuxHome variable
func getMapdataStylesPath() string {
	return stratuxHome + "/mapdata/styles/"
}

// getOgnPath returns the path to the OGN data directory
// Uses the configurable stratuxHome variable
func getOgnPath() string {
	return stratuxHome + "/ogn/"
}
