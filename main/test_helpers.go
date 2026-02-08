/*
	Copyright (c) 2026 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	test_helpers.go: Shared test utilities to reduce code duplication

	This file contains common setup/teardown functions used across test files.
	Phase 8: Code Deduplication
*/

package main

import (
	"sync"
)

// =============================================================================
// Global Variable Initialization Helpers
// =============================================================================

// initStratuxClock initializes the stratuxClock if nil.
// This is the most common initialization pattern (346+ occurrences).
func initStratuxClock() {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
	}
}

// initNetMutex initializes the netMutex if nil.
// Used in 94+ test functions.
func initNetMutex() {
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
}

// initTrafficMutex initializes the trafficMutex if nil.
// Used in 58+ test functions.
func initTrafficMutex() {
	if trafficMutex == nil {
		trafficMutex = &sync.Mutex{}
	}
}

// initSystemErrsMutex initializes the systemErrsMutex if nil.
// Used in 46+ test functions.
func initSystemErrsMutex() {
	if systemErrsMutex == nil {
		systemErrsMutex = &sync.Mutex{}
	}
}

// initADSBTowerMutex initializes the ADSBTowerMutex if nil.
// Used in 23+ test functions.
func initADSBTowerMutex() {
	if ADSBTowerMutex == nil {
		ADSBTowerMutex = &sync.Mutex{}
	}
}

// initAllMutexes initializes all common mutexes.
// Convenience function for tests that need multiple mutexes.
func initAllMutexes() {
	initNetMutex()
	initTrafficMutex()
	initSystemErrsMutex()
	initADSBTowerMutex()
}

// =============================================================================
// Test State Setup Helpers
// =============================================================================

// initTestGlobals initializes all common global variables for testing.
// Call this at the start of tests that need a clean state.
func initTestGlobals() {
	initStratuxClock()
	initAllMutexes()
}

// initClientConnections initializes the clientConnections map.
// Safe to call multiple times.
func initClientConnections() {
	initNetMutex()
	netMutex.Lock()
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}
	netMutex.Unlock()
}

// clearClientConnections clears all client connections.
// Use for test cleanup.
func clearClientConnections() {
	initNetMutex()
	netMutex.Lock()
	clientConnections = make(map[string]connection)
	netMutex.Unlock()
}

// initTrafficMap initializes the traffic map.
// Safe to call multiple times.
func initTrafficMap() {
	initTrafficMutex()
	trafficMutex.Lock()
	if traffic == nil {
		traffic = make(map[uint32]TrafficInfo)
	}
	trafficMutex.Unlock()
}

// clearTrafficMap clears all traffic entries.
// Use for test cleanup.
func clearTrafficMap() {
	initTrafficMutex()
	trafficMutex.Lock()
	traffic = make(map[uint32]TrafficInfo)
	trafficMutex.Unlock()
}

// =============================================================================
// Settings Save/Restore Helpers
// =============================================================================

// SettingsSnapshot holds a copy of globalSettings for restoration.
type SettingsSnapshot struct {
	settings settings
}

// saveGlobalSettings saves the current globalSettings.
// Returns a snapshot that can be used to restore settings.
// Usage:
//
//	snapshot := saveGlobalSettings()
//	defer snapshot.Restore()
func saveGlobalSettings() *SettingsSnapshot {
	return &SettingsSnapshot{
		settings: globalSettings,
	}
}

// Restore restores globalSettings from the snapshot.
func (s *SettingsSnapshot) Restore() {
	globalSettings = s.settings
}

// =============================================================================
// Network Channel Helpers
// =============================================================================

// initNetworkChannels initializes network-related channels if nil.
func initNetworkChannels() {
	if networkGDL90Chan == nil {
		networkGDL90Chan = make(chan []byte, 1024)
	}
}

// drainNetworkGDL90Chan drains all messages from networkGDL90Chan.
// Use after tests that send messages.
func drainNetworkGDL90Chan() {
	if networkGDL90Chan == nil {
		return
	}
	for {
		select {
		case <-networkGDL90Chan:
		default:
			return
		}
	}
}

// =============================================================================
// MySituation Helpers
// =============================================================================

// initMySituation initializes mySituation mutexes if needed.
func initMySituation() {
	// The mutexes are embedded in the struct and initialized by Go's zero value
	// This function exists for documentation and future use
}

// =============================================================================
// Composite Test Setup
// =============================================================================

// TestContext holds test state that can be cleaned up.
type TestContext struct {
	settingsSnapshot *SettingsSnapshot
}

// NewTestContext creates a new test context with saved state.
// Usage:
//
//	ctx := NewTestContext()
//	defer ctx.Cleanup()
func NewTestContext() *TestContext {
	initTestGlobals()
	return &TestContext{
		settingsSnapshot: saveGlobalSettings(),
	}
}

// Cleanup restores all saved state.
func (ctx *TestContext) Cleanup() {
	if ctx.settingsSnapshot != nil {
		ctx.settingsSnapshot.Restore()
	}
	drainNetworkGDL90Chan()
}

// =============================================================================
// Test Fixtures - Reusable Test Data
// =============================================================================

// NewTestTrafficInfo creates a TrafficInfo with common test values.
// Customize fields after creation as needed.
func NewTestTrafficInfo(icao uint32) TrafficInfo {
	initStratuxClock()
	return TrafficInfo{
		Icao_addr:         icao,
		Reg:               "",
		Tail:              "",
		Squawk:            1200,
		OnGround:          false,
		Lat:               43.0,
		Lng:               -88.0,
		Position_valid:    true,
		Alt:               5000,
		Last_seen:         stratuxClock.Time,
		Last_alt:          stratuxClock.Time,
		Last_source:       TRAFFIC_SOURCE_1090ES,
		BearingDist_valid: true,
		Bearing:           90.0,
		Distance:          5.0,
	}
}

// NewTestNetworkConnection creates a networkConnection with common test values.
func NewTestNetworkConnection(ip string, port uint32) *networkConnection {
	return &networkConnection{
		Ip:         ip,
		Port:       port,
		Capability: NETWORK_GDL90_STANDARD,
		Queue:      NewMessageQueue(100),
	}
}

// NewTestTCPConnection creates a tcpConnection with common test values.
func NewTestTCPConnection(key string) *tcpConnection {
	return &tcpConnection{
		Key:        key,
		Capability: NETWORK_FLARM_NMEA,
		Queue:      NewMessageQueue(100),
	}
}

// NewTestSerialConnection creates a serialConnection with common test values.
func NewTestSerialConnection(device string) *serialConnection {
	return &serialConnection{
		DeviceString: device,
		Baud:         38400,
		Capability:   NETWORK_GDL90_STANDARD,
		Queue:        NewMessageQueue(100),
	}
}

// =============================================================================
// GPS/Situation Test Fixtures
// =============================================================================

// SetupTestGPSPosition sets up mySituation with a valid GPS position.
// Returns a cleanup function to restore the original state.
func SetupTestGPSPosition(lat, lng float32) func() {
	initStratuxClock()

	// Save original values
	origLat := mySituation.GPSLatitude
	origLng := mySituation.GPSLongitude
	origFix := mySituation.GPSFixQuality
	origConnected := globalStatus.GPS_connected

	// Set test values
	mySituation.muGPS.Lock()
	mySituation.GPSLatitude = lat
	mySituation.GPSLongitude = lng
	mySituation.GPSFixQuality = 2
	mySituation.GPSLastFixLocalTime = stratuxClock.Time
	mySituation.muGPS.Unlock()

	globalStatus.GPS_connected = true

	// Return cleanup function
	return func() {
		mySituation.muGPS.Lock()
		mySituation.GPSLatitude = origLat
		mySituation.GPSLongitude = origLng
		mySituation.GPSFixQuality = origFix
		mySituation.muGPS.Unlock()
		globalStatus.GPS_connected = origConnected
	}
}

// =============================================================================
// Mock Implementation Notes
// =============================================================================

// Mock implementations are intentionally kept in their respective test files
// rather than consolidated here because:
// 1. Each mock serves a specific test purpose with unique requirements
// 2. Mocks often need test-specific fields (error counts, channels, etc.)
// 3. Keeping mocks close to their tests improves maintainability
// 4. The connection interface is simple (8 methods), making local mocks straightforward
//
// Current mock implementations:
// - mockConnection (network_test.go) - basic connection mock
// - mockConnectionWithWriter (network_test.go) - mock with io.Writer
// - mockConnectionForErrorTest (network_test.go) - mock for error tracking
// - mockWebSocketConn (managementinterface_test.go) - WebSocket mock
// - mockIMUReader variants (sensors_test.go) - IMU sensor mocks
// - mockPressureReader (sensors_test.go) - pressure sensor mock
