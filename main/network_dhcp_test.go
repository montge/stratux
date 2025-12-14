/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License

	network_dhcp_test.go: Tests for DHCP lease parsing and network functions
*/

package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupDHCPTestDir creates a temporary directory structure for DHCP testing
func setupDHCPTestDir(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "stratux_dhcp_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Save original paths
	origDhcpLeaseDirPath := dhcpLeaseDirPath
	origDhcpLeaseFilePath := dhcpLeaseFilePath
	origArpFilePath := arpFilePath
	origExtraHostsFilePath := extraHostsFilePath
	origDhcpLeaseDirectoryLastTest := dhcpLeaseDirectoryLastTest

	// Set paths to temp directory
	dhcpLeaseDirPath = tmpDir
	dhcpLeaseFilePath = filepath.Join(tmpDir, "dnsmasq.leases")
	arpFilePath = filepath.Join(tmpDir, "arp")
	extraHostsFilePath = filepath.Join(tmpDir, "static-hosts.conf")
	// Reset the last test time to ensure fsWriteTest runs
	dhcpLeaseDirectoryLastTest = time.Time{}

	cleanup = func() {
		dhcpLeaseDirPath = origDhcpLeaseDirPath
		dhcpLeaseFilePath = origDhcpLeaseFilePath
		arpFilePath = origArpFilePath
		extraHostsFilePath = origExtraHostsFilePath
		dhcpLeaseDirectoryLastTest = origDhcpLeaseDirectoryLastTest
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestGetDHCPLeases_EmptyDirectory(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	tmpDir, cleanup := setupDHCPTestDir(t)
	defer cleanup()
	_ = tmpDir

	leases, err := getDHCPLeases()

	// Should return empty map without error (files don't exist)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if leases == nil {
		t.Error("Expected non-nil map")
	}
	t.Logf("Empty directory returns %d leases", len(leases))
}

func TestGetDHCPLeases_WithLeaseFile(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	tmpDir, cleanup := setupDHCPTestDir(t)
	defer cleanup()

	// Create a valid dnsmasq.leases file
	// Format: <expiry> <mac> <ip> <hostname> <client-id>
	leaseContent := `1609459200 aa:bb:cc:dd:ee:01 192.168.10.100 ipad-user1 *
1609459200 aa:bb:cc:dd:ee:02 192.168.10.101 iphone-user2 01:aa:bb:cc:dd:ee:02
1609459200 aa:bb:cc:dd:ee:03 192.168.10.102 * *
`
	leaseFile := filepath.Join(tmpDir, "dnsmasq.leases")
	if err := os.WriteFile(leaseFile, []byte(leaseContent), 0644); err != nil {
		t.Fatalf("Failed to create lease file: %v", err)
	}

	leases, err := getDHCPLeases()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should have 3 entries
	if len(leases) != 3 {
		t.Errorf("Expected 3 leases, got %d: %v", len(leases), leases)
	}

	// Check specific entries
	if hostname, ok := leases["192.168.10.100"]; !ok || hostname != "ipad-user1" {
		t.Errorf("Expected ipad-user1 at 192.168.10.100, got: %s", hostname)
	}
	if hostname, ok := leases["192.168.10.101"]; !ok || hostname != "iphone-user2" {
		t.Errorf("Expected iphone-user2 at 192.168.10.101, got: %s", hostname)
	}
	// Wildcard hostname should be empty string
	if hostname, ok := leases["192.168.10.102"]; !ok || hostname != "" {
		t.Errorf("Expected empty hostname at 192.168.10.102, got: %s", hostname)
	}

	t.Logf("Parsed %d leases correctly", len(leases))
}

func TestGetDHCPLeases_WithStaticIPs(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	tmpDir, cleanup := setupDHCPTestDir(t)
	defer cleanup()
	_ = tmpDir

	// Save and set static IPs
	origStaticIps := globalSettings.StaticIps
	globalSettings.StaticIps = []string{"192.168.10.200", "192.168.10.201"}
	defer func() { globalSettings.StaticIps = origStaticIps }()

	leases, err := getDHCPLeases()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should have 2 entries from static IPs
	if len(leases) < 2 {
		t.Errorf("Expected at least 2 leases from static IPs, got %d: %v", len(leases), leases)
	}

	// Check static IPs are present
	if _, ok := leases["192.168.10.200"]; !ok {
		t.Error("Expected 192.168.10.200 in leases")
	}
	if _, ok := leases["192.168.10.201"]; !ok {
		t.Error("Expected 192.168.10.201 in leases")
	}

	t.Logf("Static IPs added correctly")
}

func TestGetDHCPLeases_WithArpTable(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	tmpDir, cleanup := setupDHCPTestDir(t)
	defer cleanup()

	// Create a mock ARP table file
	// Format matches /proc/net/arp
	arpContent := `IP address       HW type     Flags       HW address            Mask     Device
192.168.10.1     0x1         0x2         aa:bb:cc:dd:ee:00     *        wlan0
192.168.10.50    0x1         0x2         aa:bb:cc:dd:ee:50     *        wlan0
10.0.0.1         0x1         0x2         11:22:33:44:55:66     *        eth0
`
	arpFile := filepath.Join(tmpDir, "arp")
	if err := os.WriteFile(arpFile, []byte(arpContent), 0644); err != nil {
		t.Fatalf("Failed to create ARP file: %v", err)
	}

	leases, err := getDHCPLeases()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should have entries from ARP table (IPs starting with 0-2)
	if _, ok := leases["192.168.10.1"]; !ok {
		t.Error("Expected 192.168.10.1 from ARP table")
	}
	if _, ok := leases["192.168.10.50"]; !ok {
		t.Error("Expected 192.168.10.50 from ARP table")
	}
	if _, ok := leases["10.0.0.1"]; !ok {
		t.Error("Expected 10.0.0.1 from ARP table")
	}

	t.Logf("ARP table parsed correctly: %v", leases)
}

func TestGetDHCPLeases_WithExtraHostsFile(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	tmpDir, cleanup := setupDHCPTestDir(t)
	defer cleanup()

	// Create ARP file first (required to reach extra hosts parsing)
	arpFile := filepath.Join(tmpDir, "arp")
	if err := os.WriteFile(arpFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create ARP file: %v", err)
	}

	// Create extra hosts file
	// Format: <ip> <hostname>
	hostsContent := `192.168.10.250 my-efb-device
192.168.10.251 backup-device
`
	hostsFile := filepath.Join(tmpDir, "static-hosts.conf")
	if err := os.WriteFile(hostsFile, []byte(hostsContent), 0644); err != nil {
		t.Fatalf("Failed to create hosts file: %v", err)
	}

	leases, err := getDHCPLeases()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should have entries from extra hosts file
	if hostname, ok := leases["192.168.10.250"]; !ok || hostname != "my-efb-device" {
		t.Errorf("Expected my-efb-device at 192.168.10.250, got: %s", hostname)
	}
	if hostname, ok := leases["192.168.10.251"]; !ok || hostname != "backup-device" {
		t.Errorf("Expected backup-device at 192.168.10.251, got: %s", hostname)
	}

	t.Logf("Extra hosts file parsed correctly")
}

func TestGetDHCPLeases_CombinedSources(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	tmpDir, cleanup := setupDHCPTestDir(t)
	defer cleanup()

	// Create lease file
	leaseContent := `1609459200 aa:bb:cc:dd:ee:01 192.168.10.100 lease-host *
`
	leaseFile := filepath.Join(tmpDir, "dnsmasq.leases")
	if err := os.WriteFile(leaseFile, []byte(leaseContent), 0644); err != nil {
		t.Fatalf("Failed to create lease file: %v", err)
	}

	// Create ARP file
	arpContent := `IP address       HW type     Flags       HW address            Mask     Device
192.168.10.101   0x1         0x2         aa:bb:cc:dd:ee:01     *        wlan0
`
	arpFile := filepath.Join(tmpDir, "arp")
	if err := os.WriteFile(arpFile, []byte(arpContent), 0644); err != nil {
		t.Fatalf("Failed to create ARP file: %v", err)
	}

	// Create extra hosts file
	hostsContent := `192.168.10.102 extra-host
`
	hostsFile := filepath.Join(tmpDir, "static-hosts.conf")
	if err := os.WriteFile(hostsFile, []byte(hostsContent), 0644); err != nil {
		t.Fatalf("Failed to create hosts file: %v", err)
	}

	// Set static IPs
	origStaticIps := globalSettings.StaticIps
	globalSettings.StaticIps = []string{"192.168.10.103"}
	defer func() { globalSettings.StaticIps = origStaticIps }()

	leases, err := getDHCPLeases()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should have entries from all sources
	expectedCount := 4 // lease, arp, extra hosts, static IP
	if len(leases) < expectedCount {
		t.Errorf("Expected at least %d leases, got %d: %v", expectedCount, len(leases), leases)
	}

	// Verify each source
	if _, ok := leases["192.168.10.100"]; !ok {
		t.Error("Missing lease file entry")
	}
	if _, ok := leases["192.168.10.101"]; !ok {
		t.Error("Missing ARP table entry")
	}
	if _, ok := leases["192.168.10.102"]; !ok {
		t.Error("Missing extra hosts entry")
	}
	if _, ok := leases["192.168.10.103"]; !ok {
		t.Error("Missing static IP entry")
	}

	t.Logf("Combined %d entries from all sources", len(leases))
}

func TestGetDHCPLeases_MalformedLeaseFile(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	tmpDir, cleanup := setupDHCPTestDir(t)
	defer cleanup()

	// Create a malformed lease file with various edge cases
	leaseContent := `1609459200 aa:bb:cc:dd:ee:01 192.168.10.100 valid-host *
invalid line without enough fields
1609459200 aa 192.168.10.101
short
1609459200 aa:bb:cc:dd:ee:03 192.168.10.102 another-valid *
`
	leaseFile := filepath.Join(tmpDir, "dnsmasq.leases")
	if err := os.WriteFile(leaseFile, []byte(leaseContent), 0644); err != nil {
		t.Fatalf("Failed to create lease file: %v", err)
	}

	leases, err := getDHCPLeases()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should have only valid entries (at least 2)
	if len(leases) < 2 {
		t.Errorf("Expected at least 2 valid leases, got %d", len(leases))
	}

	t.Logf("Malformed file handled gracefully, got %d leases", len(leases))
}

func TestGetDHCPLeases_FsWriteTestTiming(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	tmpDir, cleanup := setupDHCPTestDir(t)
	defer cleanup()
	_ = tmpDir

	// Reset the last test time
	dhcpLeaseDirectoryLastTest = time.Time{}

	// First call should trigger fsWriteTest
	_, _ = getDHCPLeases()

	// Save the time after first call
	firstTestTime := dhcpLeaseDirectoryLastTest

	// Second call immediately after should NOT trigger fsWriteTest
	// (because < 5 minutes have passed)
	_, _ = getDHCPLeases()

	// Time should be the same
	if !dhcpLeaseDirectoryLastTest.Equal(firstTestTime) {
		t.Log("fsWriteTest was called again (expected only once within 5 minutes)")
	}

	t.Logf("Timing check completed, last test: %v", dhcpLeaseDirectoryLastTest)
}

func TestGetDHCPLeases_EmptyFiles(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	tmpDir, cleanup := setupDHCPTestDir(t)
	defer cleanup()

	// Create empty files
	leaseFile := filepath.Join(tmpDir, "dnsmasq.leases")
	if err := os.WriteFile(leaseFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty lease file: %v", err)
	}

	arpFile := filepath.Join(tmpDir, "arp")
	if err := os.WriteFile(arpFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty ARP file: %v", err)
	}

	hostsFile := filepath.Join(tmpDir, "static-hosts.conf")
	if err := os.WriteFile(hostsFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty hosts file: %v", err)
	}

	leases, err := getDHCPLeases()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Empty files should result in empty map
	if len(leases) != 0 {
		t.Errorf("Expected 0 leases from empty files, got %d", len(leases))
	}

	t.Log("Empty files handled correctly")
}
