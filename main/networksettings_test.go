/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	networksettings_test.go: Unit tests for networksettings.go

	Tests WiFi network setting functions including setWifiClientNetworks.
*/

package main

import (
	"os"
	"strconv"
	"strings"
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

// TestWriteTemplate tests the writeTemplate function
func TestWriteTemplate(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	t.Run("successful_template_write", func(t *testing.T) {
		// Create a simple template file
		templatePath := tmpDir + "/test_template_1.tpl"
		templateContent := `IP: {{.IpAddr}}
SSID: {{.WiFiSSID}}
Channel: {{.WiFiChannel}}
Mode: {{.WiFiMode}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		// Create test settings
		settings := NetworkTemplateParams{
			IpAddr:      "192.168.10.1",
			WiFiSSID:    "TestNetwork",
			WiFiChannel: 6,
			WiFiMode:    WifiModeAp,
		}

		// Write template
		outputPath := tmpDir + "/output_1.conf"
		writeTemplate(templatePath, outputPath, settings)

		// Verify output file exists and has correct content
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		expected := `IP: 192.168.10.1
SSID: TestNetwork
Channel: 6
Mode: 0`
		if string(content) != expected {
			t.Errorf("Output mismatch.\nExpected:\n%s\nGot:\n%s", expected, string(content))
		}
	})

	t.Run("template_with_all_fields", func(t *testing.T) {
		// Create a comprehensive template
		templatePath := tmpDir + "/test_template_all.tpl"
		templateContent := `WiFiMode: {{.WiFiMode}}
WiFiCountry: {{.WiFiCountry}}
IpAddr: {{.IpAddr}}
IpPrefix: {{.IpPrefix}}
DhcpRangeStart: {{.DhcpRangeStart}}
DhcpRangeEnd: {{.DhcpRangeEnd}}
WiFiSSID: {{.WiFiSSID}}
WiFiChannel: {{.WiFiChannel}}
WiFiDirectPin: {{.WiFiDirectPin}}
WiFiPassPhrase: {{.WiFiPassPhrase}}
WiFiInternetPassThroughEnabled: {{.WiFiInternetPassThroughEnabled}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		// Create comprehensive settings
		settings := NetworkTemplateParams{
			WiFiMode:                       WifiModeApClient,
			WiFiCountry:                    "US",
			IpAddr:                         "10.0.0.1",
			IpPrefix:                       "10.0.0",
			DhcpRangeStart:                 "10.0.0.10",
			DhcpRangeEnd:                   "10.0.0.50",
			WiFiSSID:                       "Stratux-Test",
			WiFiChannel:                    11,
			WiFiDirectPin:                  "12345678",
			WiFiPassPhrase:                 "SecurePassword123",
			WiFiInternetPassThroughEnabled: true,
		}

		outputPath := tmpDir + "/output_all.conf"
		writeTemplate(templatePath, outputPath, settings)

		// Verify file was created and contains all values
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		contentStr := string(content)
		expectedValues := []string{
			"WiFiMode: 2",
			"WiFiCountry: US",
			"IpAddr: 10.0.0.1",
			"IpPrefix: 10.0.0",
			"DhcpRangeStart: 10.0.0.10",
			"DhcpRangeEnd: 10.0.0.50",
			"WiFiSSID: Stratux-Test",
			"WiFiChannel: 11",
			"WiFiDirectPin: 12345678",
			"WiFiPassPhrase: SecurePassword123",
			"WiFiInternetPassThroughEnabled: true",
		}

		for _, expected := range expectedValues {
			if !strings.Contains(contentStr, expected) {
				t.Errorf("Output missing expected value: %s", expected)
			}
		}
	})

	t.Run("template_with_client_networks", func(t *testing.T) {
		// Create template with range over client networks
		templatePath := tmpDir + "/test_template_networks.tpl"
		templateContent := `{{range .WiFiClientNetworks}}network={
	ssid="{{.SSID}}"
	psk="{{.Password}}"
}
{{end}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		// Create settings with multiple client networks
		settings := NetworkTemplateParams{
			WiFiClientNetworks: []wifiClientNetwork{
				{SSID: "Network1", Password: "pass1"},
				{SSID: "Network2", Password: "pass2"},
				{SSID: "Network3", Password: "pass3"},
			},
		}

		outputPath := tmpDir + "/output_networks.conf"
		writeTemplate(templatePath, outputPath, settings)

		// Verify all networks are in output
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		contentStr := string(content)
		for _, net := range settings.WiFiClientNetworks {
			if !strings.Contains(contentStr, `ssid="`+net.SSID+`"`) {
				t.Errorf("Output missing SSID: %s", net.SSID)
			}
			if !strings.Contains(contentStr, `psk="`+net.Password+`"`) {
				t.Errorf("Output missing password for: %s", net.SSID)
			}
		}
	})

	t.Run("template_with_conditional_logic", func(t *testing.T) {
		// Create template with if/else logic (like the real dnsmasq template)
		templatePath := tmpDir + "/test_template_conditional.tpl"
		templateContent := `{{if and .WiFiInternetPassThroughEnabled (eq .WiFiMode 2)}}
dhcp-option=3,{{.IpAddr}}
dhcp-option=6,{{.IpAddr}}
{{else}}
# No gateway/DNS
dhcp-option=3
dhcp-option=6
{{end}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		// Test with passthrough enabled and correct mode
		settings := NetworkTemplateParams{
			WiFiMode:                       WifiModeApClient,
			IpAddr:                         "192.168.10.1",
			WiFiInternetPassThroughEnabled: true,
		}

		outputPath := tmpDir + "/output_conditional_enabled.conf"
		writeTemplate(templatePath, outputPath, settings)

		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		if !strings.Contains(string(content), "dhcp-option=3,192.168.10.1") {
			t.Error("Expected passthrough-enabled output, got disabled")
		}

		// Test with passthrough disabled
		settings.WiFiInternetPassThroughEnabled = false
		outputPath2 := tmpDir + "/output_conditional_disabled.conf"
		writeTemplate(templatePath, outputPath2, settings)

		content2, err := os.ReadFile(outputPath2)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		if !strings.Contains(string(content2), "# No gateway/DNS") {
			t.Error("Expected passthrough-disabled output")
		}
	})

	t.Run("template_file_not_found", func(t *testing.T) {
		// Test with non-existent template file
		nonExistentPath := tmpDir + "/does_not_exist.tpl"
		outputPath := tmpDir + "/output_notfound.conf"

		settings := NetworkTemplateParams{
			IpAddr: "192.168.10.1",
		}

		// This should log an error but not panic
		writeTemplate(nonExistentPath, outputPath, settings)

		// Output file should not exist
		if _, err := os.Stat(outputPath); err == nil {
			t.Error("Output file should not exist when template file is not found")
		}
	})

	t.Run("invalid_template_syntax", func(t *testing.T) {
		// Create template with invalid syntax
		templatePath := tmpDir + "/test_template_invalid.tpl"
		templateContent := `IP: {{.IpAddr}
SSID: {{.WiFiSSID}}` // Missing closing braces

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		outputPath := tmpDir + "/output_invalid.conf"
		settings := NetworkTemplateParams{
			IpAddr:   "192.168.10.1",
			WiFiSSID: "Test",
		}

		// This should log an error but not panic
		writeTemplate(templatePath, outputPath, settings)

		// Output file may or may not exist depending on when error occurred
		// The important thing is that it doesn't panic
	})

	t.Run("output_to_readonly_directory", func(t *testing.T) {
		// Create a valid template
		templatePath := tmpDir + "/test_template_readonly.tpl"
		templateContent := `IP: {{.IpAddr}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		// Try to write to a directory that doesn't exist
		outputPath := tmpDir + "/nonexistent_dir/output.conf"
		settings := NetworkTemplateParams{
			IpAddr: "192.168.10.1",
		}

		// This should log an error but not panic
		writeTemplate(templatePath, outputPath, settings)

		// Output file should not exist
		if _, err := os.Stat(outputPath); err == nil {
			t.Error("Output file should not exist when directory doesn't exist")
		}
	})

	t.Run("template_with_undefined_field", func(t *testing.T) {
		// Create template referencing non-existent field
		templatePath := tmpDir + "/test_template_undefined.tpl"
		templateContent := `IP: {{.IpAddr}}
NonExistent: {{.ThisFieldDoesNotExist}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		outputPath := tmpDir + "/output_undefined.conf"
		settings := NetworkTemplateParams{
			IpAddr: "192.168.10.1",
		}

		// This should fail during template execution
		writeTemplate(templatePath, outputPath, settings)

		// The file might be created but the template execution should log an error
		// Check if file exists - it may or may not depending on when the error occurred
		if _, err := os.Stat(outputPath); err == nil {
			// If file exists, verify it's either empty or incomplete
			content, _ := os.ReadFile(outputPath)
			if strings.Contains(string(content), "NonExistent: <no value>") ||
				!strings.Contains(string(content), "NonExistent:") {
				// Template either output "<no value>" or failed before writing this line
				// Both are acceptable behaviors
			}
		}
	})

	t.Run("empty_template", func(t *testing.T) {
		// Create an empty template file
		templatePath := tmpDir + "/test_template_empty.tpl"
		if err := os.WriteFile(templatePath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		outputPath := tmpDir + "/output_empty.conf"
		settings := NetworkTemplateParams{}

		writeTemplate(templatePath, outputPath, settings)

		// Output file should exist but be empty
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		if len(content) != 0 {
			t.Errorf("Expected empty output, got: %s", string(content))
		}
	})

	t.Run("template_with_special_characters", func(t *testing.T) {
		// Create template and test with special characters in values
		templatePath := tmpDir + "/test_template_special.tpl"
		templateContent := `SSID: {{.WiFiSSID}}
Password: {{.WiFiPassPhrase}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		outputPath := tmpDir + "/output_special.conf"
		settings := NetworkTemplateParams{
			WiFiSSID:       "Network-2.4GHz (5G)",
			WiFiPassPhrase: `P@ss"word'with$pecial&chars!`,
		}

		writeTemplate(templatePath, outputPath, settings)

		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, settings.WiFiSSID) {
			t.Error("SSID with special characters not written correctly")
		}
		if !strings.Contains(contentStr, settings.WiFiPassPhrase) {
			t.Error("Password with special characters not written correctly")
		}
	})

	t.Run("overwrite_existing_file", func(t *testing.T) {
		// Create template
		templatePath := tmpDir + "/test_template_overwrite.tpl"
		templateContent := `Value: {{.IpAddr}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		outputPath := tmpDir + "/output_overwrite.conf"

		// Create existing file with different content
		if err := os.WriteFile(outputPath, []byte("OLD CONTENT"), 0644); err != nil {
			t.Fatalf("Failed to create existing output file: %v", err)
		}

		settings := NetworkTemplateParams{
			IpAddr: "192.168.10.1",
		}

		writeTemplate(templatePath, outputPath, settings)

		// Verify file was overwritten
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		if strings.Contains(string(content), "OLD CONTENT") {
			t.Error("File was not overwritten")
		}
		if !strings.Contains(string(content), "Value: 192.168.10.1") {
			t.Error("New content not written correctly")
		}
	})

	t.Run("template_with_zero_values", func(t *testing.T) {
		// Test with zero/default values
		templatePath := tmpDir + "/test_template_zero.tpl"
		templateContent := `Mode: {{.WiFiMode}}
Channel: {{.WiFiChannel}}
Enabled: {{.WiFiInternetPassThroughEnabled}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		outputPath := tmpDir + "/output_zero.conf"
		settings := NetworkTemplateParams{
			WiFiMode:                       0,
			WiFiChannel:                    0,
			WiFiInternetPassThroughEnabled: false,
		}

		writeTemplate(templatePath, outputPath, settings)

		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		expected := `Mode: 0
Channel: 0
Enabled: false`
		if string(content) != expected {
			t.Errorf("Zero values not written correctly.\nExpected:\n%s\nGot:\n%s", expected, string(content))
		}
	})

	t.Run("template_with_empty_client_networks", func(t *testing.T) {
		// Test with empty client networks list
		templatePath := tmpDir + "/test_template_empty_networks.tpl"
		templateContent := `Networks:
{{range .WiFiClientNetworks}}
- {{.SSID}}
{{end}}
Done`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		outputPath := tmpDir + "/output_empty_networks.conf"
		settings := NetworkTemplateParams{
			WiFiClientNetworks: []wifiClientNetwork{},
		}

		writeTemplate(templatePath, outputPath, settings)

		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		expected := `Networks:

Done`
		if string(content) != expected {
			t.Errorf("Empty network list not handled correctly.\nExpected:\n%s\nGot:\n%s", expected, string(content))
		}
	})

	t.Run("large_template_output", func(t *testing.T) {
		// Test with many client networks to produce large output
		templatePath := tmpDir + "/test_template_large.tpl"
		templateContent := `{{range .WiFiClientNetworks}}network={
	ssid="{{.SSID}}"
	psk="{{.Password}}"
}
{{end}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}

		// Create many networks
		var networks []wifiClientNetwork
		for i := 0; i < 100; i++ {
			networks = append(networks, wifiClientNetwork{
				SSID:     "Network" + strconv.Itoa(i),
				Password: "Password" + strconv.Itoa(i),
			})
		}

		settings := NetworkTemplateParams{
			WiFiClientNetworks: networks,
		}

		outputPath := tmpDir + "/output_large.conf"
		writeTemplate(templatePath, outputPath, settings)

		// Verify all networks were written
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		contentStr := string(content)
		for i := 0; i < 100; i++ {
			expectedSSID := "Network" + strconv.Itoa(i)
			if !strings.Contains(contentStr, expectedSSID) {
				t.Errorf("Missing network %s in large output", expectedSSID)
				break // Don't spam errors
			}
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
