/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	networksettings_test.go: Unit tests for networksettings.go

	Tests WiFi network setting functions including setWifiClientNetworks.
*/

package main

import (
	"testing"
)

// TestSetWifiClientNetworks tests the setWifiClientNetworks function
func TestSetWifiClientNetworks(t *testing.T) {
	// Save original settings and restore after test
	origNetworks := globalSettings.WiFiClientNetworks
	origHasChanged := hasChanged
	defer func() {
		globalSettings.WiFiClientNetworks = origNetworks
		hasChanged = origHasChanged
	}()

	t.Run("empty_to_single_network", func(t *testing.T) {
		// Setup: empty networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{}
		hasChanged = false

		// Test: set single network
		newNetworks := []wifiClientNetwork{
			{SSID: "TestNetwork1", Password: "password123"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be true
		if !hasChanged {
			t.Error("hasChanged should be true when adding network to empty list")
		}
		if len(globalSettings.WiFiClientNetworks) != 1 {
			t.Errorf("Expected 1 network, got %d", len(globalSettings.WiFiClientNetworks))
		}
		if globalSettings.WiFiClientNetworks[0].SSID != "TestNetwork1" {
			t.Errorf("Expected SSID 'TestNetwork1', got '%s'", globalSettings.WiFiClientNetworks[0].SSID)
		}
		if globalSettings.WiFiClientNetworks[0].Password != "password123" {
			t.Errorf("Expected Password 'password123', got '%s'", globalSettings.WiFiClientNetworks[0].Password)
		}
	})

	t.Run("single_to_multiple_networks", func(t *testing.T) {
		// Setup: single network
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{
			{SSID: "Network1", Password: "pass1"},
		}
		hasChanged = false

		// Test: set multiple networks
		newNetworks := []wifiClientNetwork{
			{SSID: "Network1", Password: "pass1"},
			{SSID: "Network2", Password: "pass2"},
			{SSID: "Network3", Password: "pass3"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be true (length changed)
		if !hasChanged {
			t.Error("hasChanged should be true when network count changes")
		}
		if len(globalSettings.WiFiClientNetworks) != 3 {
			t.Errorf("Expected 3 networks, got %d", len(globalSettings.WiFiClientNetworks))
		}
	})

	t.Run("same_networks_no_change", func(t *testing.T) {
		// Setup: two networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{
			{SSID: "Net1", Password: "pw1"},
			{SSID: "Net2", Password: "pw2"},
		}
		hasChanged = false

		// Test: set same networks
		newNetworks := []wifiClientNetwork{
			{SSID: "Net1", Password: "pw1"},
			{SSID: "Net2", Password: "pw2"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be false (no actual change)
		if hasChanged {
			t.Error("hasChanged should be false when networks are identical")
		}
	})

	t.Run("modify_ssid_only", func(t *testing.T) {
		// Setup: single network
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{
			{SSID: "OldSSID", Password: "password"},
		}
		hasChanged = false

		// Test: change SSID only
		newNetworks := []wifiClientNetwork{
			{SSID: "NewSSID", Password: "password"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be true
		if !hasChanged {
			t.Error("hasChanged should be true when SSID changes")
		}
		if globalSettings.WiFiClientNetworks[0].SSID != "NewSSID" {
			t.Errorf("Expected SSID 'NewSSID', got '%s'", globalSettings.WiFiClientNetworks[0].SSID)
		}
	})

	t.Run("modify_password_only", func(t *testing.T) {
		// Setup: single network
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{
			{SSID: "MyNetwork", Password: "oldpass"},
		}
		hasChanged = false

		// Test: change password only
		newNetworks := []wifiClientNetwork{
			{SSID: "MyNetwork", Password: "newpass"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be true
		if !hasChanged {
			t.Error("hasChanged should be true when password changes")
		}
		if globalSettings.WiFiClientNetworks[0].Password != "newpass" {
			t.Errorf("Expected password 'newpass', got '%s'", globalSettings.WiFiClientNetworks[0].Password)
		}
	})

	t.Run("multiple_to_empty", func(t *testing.T) {
		// Setup: multiple networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{
			{SSID: "Net1", Password: "pass1"},
			{SSID: "Net2", Password: "pass2"},
		}
		hasChanged = false

		// Test: clear all networks
		newNetworks := []wifiClientNetwork{}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be true (length changed)
		if !hasChanged {
			t.Error("hasChanged should be true when clearing networks")
		}
		if len(globalSettings.WiFiClientNetworks) != 0 {
			t.Errorf("Expected 0 networks, got %d", len(globalSettings.WiFiClientNetworks))
		}
	})

	t.Run("empty_to_empty", func(t *testing.T) {
		// Setup: empty networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{}
		hasChanged = false

		// Test: set empty networks
		newNetworks := []wifiClientNetwork{}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be false
		if hasChanged {
			t.Error("hasChanged should be false when both are empty")
		}
	})

	t.Run("special_characters_in_ssid", func(t *testing.T) {
		// Setup: empty networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{}
		hasChanged = false

		// Test: SSID with special characters
		newNetworks := []wifiClientNetwork{
			{SSID: "My WiFi Network-5GHz (2.4)", Password: "pass"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: should handle special characters
		if !hasChanged {
			t.Error("hasChanged should be true when adding network with special chars")
		}
		if globalSettings.WiFiClientNetworks[0].SSID != "My WiFi Network-5GHz (2.4)" {
			t.Errorf("Expected SSID with special chars, got '%s'", globalSettings.WiFiClientNetworks[0].SSID)
		}
	})

	t.Run("special_characters_in_password", func(t *testing.T) {
		// Setup: empty networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{}
		hasChanged = false

		// Test: password with special characters
		newNetworks := []wifiClientNetwork{
			{SSID: "TestNet", Password: "P@ssw0rd!#$%^&*()"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: should handle special characters
		if !hasChanged {
			t.Error("hasChanged should be true")
		}
		if globalSettings.WiFiClientNetworks[0].Password != "P@ssw0rd!#$%^&*()" {
			t.Errorf("Expected password with special chars, got '%s'", globalSettings.WiFiClientNetworks[0].Password)
		}
	})

	t.Run("unicode_in_ssid", func(t *testing.T) {
		// Setup: empty networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{}
		hasChanged = false

		// Test: SSID with unicode characters
		newNetworks := []wifiClientNetwork{
			{SSID: "Café WiFi ☕", Password: "coffee123"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: should handle unicode
		if !hasChanged {
			t.Error("hasChanged should be true")
		}
		if globalSettings.WiFiClientNetworks[0].SSID != "Café WiFi ☕" {
			t.Errorf("Expected unicode SSID, got '%s'", globalSettings.WiFiClientNetworks[0].SSID)
		}
	})

	t.Run("empty_ssid_and_password", func(t *testing.T) {
		// Setup: non-empty networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{
			{SSID: "Test", Password: "pass"},
		}
		hasChanged = false

		// Test: network with empty SSID and password
		newNetworks := []wifiClientNetwork{
			{SSID: "", Password: ""},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: should still process (validation happens elsewhere)
		if !hasChanged {
			t.Error("hasChanged should be true when SSID changes to empty")
		}
		if globalSettings.WiFiClientNetworks[0].SSID != "" {
			t.Errorf("Expected empty SSID, got '%s'", globalSettings.WiFiClientNetworks[0].SSID)
		}
	})

	t.Run("modify_middle_network", func(t *testing.T) {
		// Setup: three networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{
			{SSID: "Net1", Password: "pass1"},
			{SSID: "Net2", Password: "pass2"},
			{SSID: "Net3", Password: "pass3"},
		}
		hasChanged = false

		// Test: modify middle network
		newNetworks := []wifiClientNetwork{
			{SSID: "Net1", Password: "pass1"},
			{SSID: "Net2Modified", Password: "pass2new"},
			{SSID: "Net3", Password: "pass3"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be true
		if !hasChanged {
			t.Error("hasChanged should be true when middle network changes")
		}
		if globalSettings.WiFiClientNetworks[1].SSID != "Net2Modified" {
			t.Errorf("Expected 'Net2Modified', got '%s'", globalSettings.WiFiClientNetworks[1].SSID)
		}
	})

	t.Run("long_ssid_and_password", func(t *testing.T) {
		// Setup: empty networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{}
		hasChanged = false

		// Test: very long SSID and password (max WPA2 password is 63 chars, SSID is 32)
		longSSID := "VeryLongNetworkNameWith32Chars"
		longPassword := "ThisIsAVeryLongPasswordThatCouldBeUpTo63CharactersLongForWPA2"
		newNetworks := []wifiClientNetwork{
			{SSID: longSSID, Password: longPassword},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: should handle long strings
		if !hasChanged {
			t.Error("hasChanged should be true")
		}
		if globalSettings.WiFiClientNetworks[0].SSID != longSSID {
			t.Errorf("Expected long SSID, got '%s'", globalSettings.WiFiClientNetworks[0].SSID)
		}
		if globalSettings.WiFiClientNetworks[0].Password != longPassword {
			t.Errorf("Expected long password, got '%s'", globalSettings.WiFiClientNetworks[0].Password)
		}
	})

	t.Run("reorder_networks", func(t *testing.T) {
		// Setup: two networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{
			{SSID: "Net1", Password: "pass1"},
			{SSID: "Net2", Password: "pass2"},
		}
		hasChanged = false

		// Test: same networks but different order
		newNetworks := []wifiClientNetwork{
			{SSID: "Net2", Password: "pass2"},
			{SSID: "Net1", Password: "pass1"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be true (order matters - first network checked is [0])
		if !hasChanged {
			t.Error("hasChanged should be true when network order changes")
		}
		if globalSettings.WiFiClientNetworks[0].SSID != "Net2" {
			t.Errorf("Expected first network to be 'Net2', got '%s'", globalSettings.WiFiClientNetworks[0].SSID)
		}
	})

	t.Run("nil_to_empty_slice", func(t *testing.T) {
		// Setup: nil networks
		globalSettings.WiFiClientNetworks = nil
		hasChanged = false

		// Test: set empty slice
		newNetworks := []wifiClientNetwork{}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be false (both have length 0)
		if hasChanged {
			t.Error("hasChanged should be false when going from nil to empty slice")
		}
	})

	t.Run("nil_to_networks", func(t *testing.T) {
		// Setup: nil networks
		globalSettings.WiFiClientNetworks = nil
		hasChanged = false

		// Test: set actual networks
		newNetworks := []wifiClientNetwork{
			{SSID: "NewNet", Password: "newpass"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: hasChanged should be true (length changes from 0 to 1)
		if !hasChanged {
			t.Error("hasChanged should be true when going from nil to networks")
		}
		if len(globalSettings.WiFiClientNetworks) != 1 {
			t.Errorf("Expected 1 network, got %d", len(globalSettings.WiFiClientNetworks))
		}
	})

	t.Run("whitespace_in_credentials", func(t *testing.T) {
		// Setup: empty networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{}
		hasChanged = false

		// Test: SSID and password with leading/trailing whitespace
		newNetworks := []wifiClientNetwork{
			{SSID: "  Network  ", Password: "  password  "},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: should preserve whitespace (trimming happens elsewhere if needed)
		if !hasChanged {
			t.Error("hasChanged should be true")
		}
		if globalSettings.WiFiClientNetworks[0].SSID != "  Network  " {
			t.Errorf("Expected SSID with whitespace, got '%s'", globalSettings.WiFiClientNetworks[0].SSID)
		}
		if globalSettings.WiFiClientNetworks[0].Password != "  password  " {
			t.Errorf("Expected password with whitespace, got '%s'", globalSettings.WiFiClientNetworks[0].Password)
		}
	})

	t.Run("multiple_identical_networks", func(t *testing.T) {
		// Setup: empty networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{}
		hasChanged = false

		// Test: multiple networks with same credentials (edge case)
		newNetworks := []wifiClientNetwork{
			{SSID: "SameNet", Password: "samepass"},
			{SSID: "SameNet", Password: "samepass"},
			{SSID: "SameNet", Password: "samepass"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: should accept duplicates (validation happens elsewhere)
		if !hasChanged {
			t.Error("hasChanged should be true")
		}
		if len(globalSettings.WiFiClientNetworks) != 3 {
			t.Errorf("Expected 3 networks, got %d", len(globalSettings.WiFiClientNetworks))
		}
	})

	t.Run("detect_first_network_change", func(t *testing.T) {
		// Setup: multiple networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{
			{SSID: "First", Password: "pass1"},
			{SSID: "Second", Password: "pass2"},
		}
		hasChanged = false

		// Test: change only first network
		newNetworks := []wifiClientNetwork{
			{SSID: "FirstModified", Password: "pass1"},
			{SSID: "Second", Password: "pass2"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: should detect change early and return
		if !hasChanged {
			t.Error("hasChanged should be true when first network SSID changes")
		}
	})

	t.Run("detect_last_network_change", func(t *testing.T) {
		// Setup: multiple networks
		globalSettings.WiFiClientNetworks = []wifiClientNetwork{
			{SSID: "First", Password: "pass1"},
			{SSID: "Second", Password: "pass2"},
			{SSID: "Third", Password: "pass3"},
		}
		hasChanged = false

		// Test: change only last network
		newNetworks := []wifiClientNetwork{
			{SSID: "First", Password: "pass1"},
			{SSID: "Second", Password: "pass2"},
			{SSID: "Third", Password: "modifiedpass3"},
		}
		setWifiClientNetworks(newNetworks)

		// Verify: should detect change in last network
		if !hasChanged {
			t.Error("hasChanged should be true when last network password changes")
		}
	})
}
// TestApplyNetworkSettings tests the applyNetworkSettings function
func TestApplyNetworkSettings(t *testing.T) {
	origIPAddress := globalSettings.WiFiIPAddress
	origHasChanged := hasChanged
	defer func() {
		globalSettings.WiFiIPAddress = origIPAddress
		hasChanged = origHasChanged
	}()

	t.Run("ip_in_dhcp_range", func(t *testing.T) {
		hasChanged = false
		globalSettings.WiFiIPAddress = "192.168.1.25"
		applyNetworkSettings(true, true)
	})

	t.Run("onlyWriteFiles_true", func(t *testing.T) {
		hasChanged = true
		globalSettings.WiFiIPAddress = "192.168.10.1"
		applyNetworkSettings(false, true)
	})
}
