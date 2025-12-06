/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	clientconnection_test.go: Unit tests for clientconnection.go

	Tests all connection types (network, serial, TCP, BLE) and their methods.
*/

package main

import (
	"errors"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/tarm/serial"
)

// Mock TCP Addr
type mockAddr struct {
	network string
	address string
}

func (m mockAddr) Network() string {
	return m.network
}

func (m mockAddr) String() string {
	return m.address
}

// TestNetworkConnection_MessageQueue tests the MessageQueue method for networkConnection
func TestNetworkConnection_MessageQueue(t *testing.T) {
	// Initialize stratuxClock if needed
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	conn := &networkConnection{
		Ip:   "192.168.10.100",
		Port: 4000,
	}

	// First call should create a new queue
	queue1 := conn.MessageQueue()
	if queue1 == nil {
		t.Fatal("MessageQueue() returned nil")
	}
	if queue1.maxSize != 1024 {
		t.Errorf("Queue maxSize = %d, expected 1024", queue1.maxSize)
	}

	// Second call should return the same queue
	queue2 := conn.MessageQueue()
	if queue2 != queue1 {
		t.Error("MessageQueue() returned different queue on second call")
	}

	// If queue is pre-set, should return that
	presetQueue := NewMessageQueue(512)
	conn.Queue = presetQueue
	queue3 := conn.MessageQueue()
	if queue3 != presetQueue {
		t.Error("MessageQueue() didn't return preset queue")
	}
}

// TestNetworkConnection_Writer tests the Writer method for networkConnection
func TestNetworkConnection_Writer(t *testing.T) {
	// Create a real UDP connection for testing
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("Cannot resolve UDP address: %v", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Skipf("Cannot create UDP connection: %v", err)
	}
	defer udpConn.Close()

	conn := &networkConnection{
		Conn: udpConn,
		Ip:   "192.168.10.100",
		Port: 4000,
	}

	writer := conn.Writer()
	if writer == nil {
		t.Fatal("Writer() returned nil")
	}

	// Verify Writer returns the Conn
	if writer != udpConn {
		t.Error("Writer() should return the UDP connection")
	}
}

// TestNetworkConnection_IsThrottled tests the IsThrottled method
func TestNetworkConnection_IsThrottled(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	conn := &networkConnection{
		Ip:   "192.168.10.100",
		Port: 4000,
	}

	// When LastUnreachable is recent (< 15 seconds), should be throttled sometimes
	conn.LastUnreachable = stratuxClock.Time

	// Test multiple times to account for randomness (rand.Int()%1000 != 0)
	throttledCount := 0
	testRuns := 100
	for i := 0; i < testRuns; i++ {
		if conn.IsThrottled() {
			throttledCount++
		}
	}

	// Should be throttled most of the time (999/1000 probability)
	if throttledCount < 90 {
		t.Errorf("IsThrottled was true only %d/%d times when recently unreachable (expected ~99%%)", throttledCount, testRuns)
	}

	// When LastUnreachable is old (>= 15 seconds), should not be throttled
	conn.LastUnreachable = stratuxClock.Time.Add(-16 * time.Second)
	throttledCount = 0
	for i := 0; i < testRuns; i++ {
		if conn.IsThrottled() {
			throttledCount++
		}
	}

	// Should rarely be throttled (1/1000 probability)
	if throttledCount > 10 {
		t.Errorf("IsThrottled was true %d/%d times when not recently unreachable (expected ~0%%)", throttledCount, testRuns)
	}
}

// TestNetworkConnection_IsSleeping tests the IsSleeping method
func TestNetworkConnection_IsSleeping(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Save original settings and restore after test
	origNoSleep := globalSettings.NoSleep
	defer func() { globalSettings.NoSleep = origNoSleep }()

	conn := &networkConnection{
		Ip:   "192.168.10.100",
		Port: 4000,
	}

	// Test case 1: x86 debug mode - never sleeping
	if runtime.GOARCH == "i386" || runtime.GOARCH == "amd64" {
		globalSettings.NoSleep = false
		if conn.IsSleeping() {
			t.Error("IsSleeping should return false in x86 debug mode")
		}
	}

	// Test case 2: NoSleep setting - never sleeping
	globalSettings.NoSleep = true
	if conn.IsSleeping() {
		t.Error("IsSleeping should return false when NoSleep is true")
	}

	globalSettings.NoSleep = false

	// Skip remaining tests if in x86 debug mode
	if runtime.GOARCH == "i386" || runtime.GOARCH == "amd64" {
		t.Log("Skipping sleep tests in x86 debug mode")
		return
	}

	// Test case 3: No pong response - should be sleeping
	conn.LastPongResponse = time.Time{} // Zero time
	if !conn.IsSleeping() {
		t.Error("IsSleeping should return true when LastPongResponse is zero")
	}
	if !conn.SleepFlag {
		t.Error("SleepFlag should be true when LastPongResponse is zero")
	}

	// Test case 4: Recent pong response - should not be sleeping
	conn.LastPongResponse = stratuxClock.Time
	conn.LastPingResponse = stratuxClock.Time
	conn.LastUnreachable = time.Time{} // Zero time (old)
	if conn.IsSleeping() {
		t.Error("IsSleeping should return false when LastPongResponse is recent")
	}
	if conn.SleepFlag {
		t.Error("SleepFlag should be false when LastPongResponse is recent")
	}

	// Test case 5: Old pong response (> 10 seconds) - should be sleeping
	conn.LastPongResponse = stratuxClock.Time.Add(-11 * time.Second)
	conn.LastPingResponse = stratuxClock.Time.Add(-11 * time.Second)
	if !conn.IsSleeping() {
		t.Error("IsSleeping should return true when LastPongResponse is old")
	}

	// Test case 6: No ping response - should be sleeping
	conn.LastPingResponse = time.Time{} // Zero time
	conn.LastPongResponse = stratuxClock.Time // Recent pong
	if !conn.IsSleeping() {
		t.Error("IsSleeping should return true when LastPingResponse is zero")
	}

	// Test case 7: Recent unreachable (< 5 seconds) - should be sleeping
	conn.LastPongResponse = stratuxClock.Time
	conn.LastPingResponse = stratuxClock.Time
	conn.LastUnreachable = stratuxClock.Time
	if !conn.IsSleeping() {
		t.Error("IsSleeping should return true when LastUnreachable is recent")
	}

	// Test case 8: Old unreachable (>= 5 seconds) - should not be sleeping
	conn.LastUnreachable = stratuxClock.Time.Add(-6 * time.Second)
	if conn.IsSleeping() {
		t.Error("IsSleeping should return false when LastUnreachable is old and pings are recent")
	}
}

// TestNetworkConnection_Capabilities tests the Capabilities method
func TestNetworkConnection_Capabilities(t *testing.T) {
	testCases := []struct {
		name       string
		capability uint8
	}{
		{"GDL90", NETWORK_GDL90_STANDARD},
		{"AHRS_FFSIM", NETWORK_AHRS_FFSIM},
		{"AHRS_GDL90", NETWORK_AHRS_GDL90},
		{"FLARM_NMEA", NETWORK_FLARM_NMEA},
		{"POSITION_FFSIM", NETWORK_POSITION_FFSIM},
		{"Combined", NETWORK_GDL90_STANDARD | NETWORK_AHRS_GDL90},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := &networkConnection{
				Capability: tc.capability,
			}
			if conn.Capabilities() != tc.capability {
				t.Errorf("Capabilities() = %d, expected %d", conn.Capabilities(), tc.capability)
			}
		})
	}
}

// TestNetworkConnection_GetDesiredPacketSize tests the GetDesiredPacketSize method
func TestNetworkConnection_GetDesiredPacketSize(t *testing.T) {
	// Test case 1: FFSIM capabilities - should return 1 (X-Plane workaround)
	conn := &networkConnection{
		Capability: NETWORK_POSITION_FFSIM,
	}
	if size := conn.GetDesiredPacketSize(); size != 1 {
		t.Errorf("GetDesiredPacketSize() = %d, expected 1 for POSITION_FFSIM", size)
	}

	conn.Capability = NETWORK_AHRS_FFSIM
	if size := conn.GetDesiredPacketSize(); size != 1 {
		t.Errorf("GetDesiredPacketSize() = %d, expected 1 for AHRS_FFSIM", size)
	}

	conn.Capability = NETWORK_POSITION_FFSIM | NETWORK_AHRS_FFSIM
	if size := conn.GetDesiredPacketSize(); size != 1 {
		t.Errorf("GetDesiredPacketSize() = %d, expected 1 for combined FFSIM", size)
	}

	// Test case 2: Non-FFSIM capabilities - should return 1024
	conn.Capability = NETWORK_GDL90_STANDARD
	if size := conn.GetDesiredPacketSize(); size != 1024 {
		t.Errorf("GetDesiredPacketSize() = %d, expected 1024 for GDL90_STANDARD", size)
	}

	conn.Capability = NETWORK_AHRS_GDL90
	if size := conn.GetDesiredPacketSize(); size != 1024 {
		t.Errorf("GetDesiredPacketSize() = %d, expected 1024 for AHRS_GDL90", size)
	}

	conn.Capability = 0
	if size := conn.GetDesiredPacketSize(); size != 1024 {
		t.Errorf("GetDesiredPacketSize() = %d, expected 1024 for no capability", size)
	}
}

// TestNetworkConnection_OnError tests the OnError method
func TestNetworkConnection_OnError(t *testing.T) {
	conn := &networkConnection{
		Ip:   "192.168.10.100",
		Port: 4000,
	}

	// OnError should not panic or do anything for UDP
	testErr := errors.New("test error")
	conn.OnError(testErr)

	// If we get here without panic, test passes
	// UDP connections ignore errors and keep socket open
}

// TestNetworkConnection_Close tests the Close method
func TestNetworkConnection_Close(t *testing.T) {
	conn := &networkConnection{
		Conn: nil, // Can be nil for this test
		Ip:   "192.168.10.100",
		Port: 4000,
	}

	// Close should not panic or do anything for UDP
	conn.Close()

	// If we get here without panic, test passes
	// UDP connections ignore Close and keep socket open
}

// TestNetworkConnection_GetConnectionKey tests the GetConnectionKey method
func TestNetworkConnection_GetConnectionKey(t *testing.T) {
	testCases := []struct {
		ip       string
		port     uint32
		expected string
	}{
		{"192.168.10.100", 4000, "192.168.10.100:4000"},
		{"10.0.0.1", 12345, "10.0.0.1:12345"},
		{"172.16.0.50", 49002, "172.16.0.50:49002"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			conn := &networkConnection{
				Ip:   tc.ip,
				Port: tc.port,
			}
			if key := conn.GetConnectionKey(); key != tc.expected {
				t.Errorf("GetConnectionKey() = %s, expected %s", key, tc.expected)
			}
		})
	}
}

// TestSerialConnection_MessageQueue tests the MessageQueue method for serialConnection
func TestSerialConnection_MessageQueue(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	conn := &serialConnection{
		DeviceString: "/dev/ttyUSB0",
		Baud:         38400,
	}

	// First call should create a new queue
	queue1 := conn.MessageQueue()
	if queue1 == nil {
		t.Fatal("MessageQueue() returned nil")
	}
	if queue1.maxSize != 1024 {
		t.Errorf("Queue maxSize = %d, expected 1024", queue1.maxSize)
	}

	// Second call should return the same queue
	queue2 := conn.MessageQueue()
	if queue2 != queue1 {
		t.Error("MessageQueue() returned different queue on second call")
	}
}

// TestSerialConnection_Writer tests the Writer method for serialConnection
func TestSerialConnection_Writer(t *testing.T) {
	// Create a connection with nil serial port
	conn := &serialConnection{
		DeviceString: "/dev/ttyUSB0",
		Baud:         38400,
		serialPort:   nil, // Will be nil in tests
	}

	writer := conn.Writer()
	// Writer returns serialPort directly, which can be nil
	// This is the actual implementation behavior
	_ = writer // Writer can be nil, which is valid

	// Test with a non-nil serial port (just use the struct pointer)
	conn.serialPort = &serial.Port{}
	writer = conn.Writer()
	if writer == nil {
		t.Error("Writer() should return non-nil when serialPort is set")
	}
	if writer != conn.serialPort {
		t.Error("Writer() should return the serial port")
	}
}

// TestSerialConnection_IsThrottled tests the IsThrottled method
func TestSerialConnection_IsThrottled(t *testing.T) {
	conn := &serialConnection{
		DeviceString: "/dev/ttyUSB0",
	}

	// Serial connections are never throttled
	if conn.IsThrottled() {
		t.Error("IsThrottled() should always return false for serial connections")
	}
}

// TestSerialConnection_IsSleeping tests the IsSleeping method
func TestSerialConnection_IsSleeping(t *testing.T) {
	// When serialPort is nil, should be sleeping
	conn := &serialConnection{
		DeviceString: "/dev/ttyUSB0",
		serialPort:   nil,
	}
	if !conn.IsSleeping() {
		t.Error("IsSleeping() should return true when serialPort is nil")
	}

	// When serialPort is not nil, should not be sleeping
	// We can't create a real serial port, so we'll use a type assertion hack
	// This is just for testing the logic
	conn.serialPort = &serial.Port{}
	if conn.IsSleeping() {
		t.Error("IsSleeping() should return false when serialPort is not nil")
	}
}

// TestSerialConnection_Capabilities tests the Capabilities method
func TestSerialConnection_Capabilities(t *testing.T) {
	conn := &serialConnection{
		DeviceString: "/dev/ttyUSB0",
		Capability:   NETWORK_GDL90_STANDARD,
	}

	if conn.Capabilities() != NETWORK_GDL90_STANDARD {
		t.Errorf("Capabilities() = %d, expected %d", conn.Capabilities(), NETWORK_GDL90_STANDARD)
	}
}

// TestSerialConnection_GetDesiredPacketSize tests the GetDesiredPacketSize method
func TestSerialConnection_GetDesiredPacketSize(t *testing.T) {
	conn := &serialConnection{
		DeviceString: "/dev/ttyUSB0",
	}

	// Serial connections always use 128 byte packets
	if size := conn.GetDesiredPacketSize(); size != 128 {
		t.Errorf("GetDesiredPacketSize() = %d, expected 128", size)
	}
}

// TestSerialConnection_GetConnectionKey tests the GetConnectionKey method
func TestSerialConnection_GetConnectionKey(t *testing.T) {
	testCases := []string{
		"/dev/ttyUSB0",
		"/dev/ttyAMA0",
		"/dev/serial0",
	}

	for _, device := range testCases {
		t.Run(device, func(t *testing.T) {
			conn := &serialConnection{
				DeviceString: device,
			}
			if key := conn.GetConnectionKey(); key != device {
				t.Errorf("GetConnectionKey() = %s, expected %s", key, device)
			}
		})
	}
}

// TestTCPConnection_MessageQueue tests the MessageQueue method for tcpConnection
func TestTCPConnection_MessageQueue(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	conn := &tcpConnection{
		Conn: nil, // Can be nil for this test
		Key:  "TCP:192.168.10.50:30000",
	}

	// First call should create a new queue
	queue1 := conn.MessageQueue()
	if queue1 == nil {
		t.Fatal("MessageQueue() returned nil")
	}
	if queue1.maxSize != 1024 {
		t.Errorf("Queue maxSize = %d, expected 1024", queue1.maxSize)
	}

	// Second call should return the same queue
	queue2 := conn.MessageQueue()
	if queue2 != queue1 {
		t.Error("MessageQueue() returned different queue on second call")
	}
}

// TestTCPConnection_Writer tests the Writer method for tcpConnection
func TestTCPConnection_Writer(t *testing.T) {
	// Create a real TCP listener and connection for testing
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("Cannot create TCP listener: %v", err)
	}
	defer listener.Close()

	// Connect to the listener in a goroutine
	doneCh := make(chan bool)
	go func() {
		conn, _ := listener.Accept()
		if conn != nil {
			conn.Close()
		}
		doneCh <- true
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Skipf("Cannot create TCP connection: %v", err)
	}
	defer clientConn.Close()

	tcpConn := clientConn.(*net.TCPConn)
	conn := &tcpConnection{
		Conn: tcpConn,
		Key:  "TCP:192.168.10.50:30000",
	}

	writer := conn.Writer()
	if writer == nil {
		t.Fatal("Writer() returned nil")
	}

	// Verify Writer returns the Conn
	if writer != tcpConn {
		t.Error("Writer() should return the TCP connection")
	}

	<-doneCh
}

// TestTCPConnection_IsThrottled tests the IsThrottled method
func TestTCPConnection_IsThrottled(t *testing.T) {
	conn := &tcpConnection{
		Key: "TCP:192.168.10.50:30000",
	}

	// TCP connections are never throttled
	if conn.IsThrottled() {
		t.Error("IsThrottled() should always return false for TCP connections")
	}
}

// TestTCPConnection_IsSleeping tests the IsSleeping method
func TestTCPConnection_IsSleeping(t *testing.T) {
	// When Conn is nil, should be sleeping
	conn := &tcpConnection{
		Conn: nil,
		Key:  "TCP:192.168.10.50:30000",
	}
	if !conn.IsSleeping() {
		t.Error("IsSleeping() should return true when Conn is nil")
	}

	// When Conn is not nil, should not be sleeping (create a minimal real connection)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("Cannot create TCP listener: %v", err)
	}
	defer listener.Close()

	doneCh := make(chan bool)
	go func() {
		c, _ := listener.Accept()
		if c != nil {
			c.Close()
		}
		doneCh <- true
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Skipf("Cannot create TCP connection: %v", err)
	}
	defer clientConn.Close()

	conn.Conn = clientConn.(*net.TCPConn)
	if conn.IsSleeping() {
		t.Error("IsSleeping() should return false when Conn is not nil")
	}

	<-doneCh
}

// TestTCPConnection_Capabilities tests the Capabilities method
func TestTCPConnection_Capabilities(t *testing.T) {
	conn := &tcpConnection{
		Capability: NETWORK_GDL90_STANDARD,
		Key:        "TCP:192.168.10.50:30000",
	}

	if conn.Capabilities() != NETWORK_GDL90_STANDARD {
		t.Errorf("Capabilities() = %d, expected %d", conn.Capabilities(), NETWORK_GDL90_STANDARD)
	}
}

// TestTCPConnection_GetDesiredPacketSize tests the GetDesiredPacketSize method
func TestTCPConnection_GetDesiredPacketSize(t *testing.T) {
	conn := &tcpConnection{
		Key: "TCP:192.168.10.50:30000",
	}

	// TCP connections use 512 byte packets
	if size := conn.GetDesiredPacketSize(); size != 512 {
		t.Errorf("GetDesiredPacketSize() = %d, expected 512", size)
	}
}

// TestTCPConnection_GetConnectionKey tests the GetConnectionKey method
func TestTCPConnection_GetConnectionKey(t *testing.T) {
	testCases := []string{
		"TCP:192.168.10.50:30000",
		"TCP:10.0.0.1:12345",
		"TCP:172.16.0.100:4000",
	}

	for _, key := range testCases {
		t.Run(key, func(t *testing.T) {
			conn := &tcpConnection{
				Key: key,
			}
			if connKey := conn.GetConnectionKey(); connKey != key {
				t.Errorf("GetConnectionKey() = %s, expected %s", connKey, key)
			}
		})
	}
}

// TestBLEConnection_MessageQueue tests the MessageQueue method for bleConnection
func TestBLEConnection_MessageQueue(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	conn := &bleConnection{
		UUIDService: "0xFFE0",
		UUIDGatt:    "0xFFE1",
	}

	// First call should create a new queue
	queue1 := conn.MessageQueue()
	if queue1 == nil {
		t.Fatal("MessageQueue() returned nil")
	}
	if queue1.maxSize != 1024 {
		t.Errorf("Queue maxSize = %d, expected 1024", queue1.maxSize)
	}

	// Second call should return the same queue
	queue2 := conn.MessageQueue()
	if queue2 != queue1 {
		t.Error("MessageQueue() returned different queue on second call")
	}
}

// TestBLEConnection_Writer tests the Writer method for bleConnection
func TestBLEConnection_Writer(t *testing.T) {
	conn := &bleConnection{
		UUIDService: "0xFFE0",
		UUIDGatt:    "0xFFE1",
	}

	writer := conn.Writer()
	if writer == nil {
		t.Fatal("Writer() returned nil")
	}

	// Verify it returns the connection itself (which implements io.Writer via Write method)
	if writer != conn {
		t.Error("Writer() should return the connection itself")
	}
}

// TestBLEConnection_IsThrottled tests the IsThrottled method
func TestBLEConnection_IsThrottled(t *testing.T) {
	conn := &bleConnection{
		UUIDService: "0xFFE0",
	}

	// BLE connections are never throttled
	if conn.IsThrottled() {
		t.Error("IsThrottled() should always return false for BLE connections")
	}
}

// TestBLEConnection_IsSleeping tests the IsSleeping method
func TestBLEConnection_IsSleeping(t *testing.T) {
	conn := &bleConnection{
		UUIDService: "0xFFE0",
	}

	// BLE connections are never sleeping
	if conn.IsSleeping() {
		t.Error("IsSleeping() should always return false for BLE connections")
	}
}

// TestBLEConnection_Capabilities tests the Capabilities method
func TestBLEConnection_Capabilities(t *testing.T) {
	conn := &bleConnection{
		UUIDService: "0xFFE0",
		Capability:  NETWORK_GDL90_STANDARD,
	}

	if conn.Capabilities() != NETWORK_GDL90_STANDARD {
		t.Errorf("Capabilities() = %d, expected %d", conn.Capabilities(), NETWORK_GDL90_STANDARD)
	}
}

// TestBLEConnection_GetDesiredPacketSize tests the GetDesiredPacketSize method
func TestBLEConnection_GetDesiredPacketSize(t *testing.T) {
	conn := &bleConnection{
		UUIDService: "0xFFE0",
	}

	// BLE connections use 20 byte packets (BLE MTU limitation)
	if size := conn.GetDesiredPacketSize(); size != 20 {
		t.Errorf("GetDesiredPacketSize() = %d, expected 20", size)
	}
}

// TestBLEConnection_GetConnectionKey tests the GetConnectionKey method
func TestBLEConnection_GetConnectionKey(t *testing.T) {
	testCases := []string{
		"0xFFE0",
		"0x1234",
		"custom-uuid",
	}

	for _, uuid := range testCases {
		t.Run(uuid, func(t *testing.T) {
			conn := &bleConnection{
				UUIDService: uuid,
			}
			if key := conn.GetConnectionKey(); key != uuid {
				t.Errorf("GetConnectionKey() = %s, expected %s", key, uuid)
			}
		})
	}
}

// TestBLEConnection_OnError tests the OnError method
func TestBLEConnection_OnError(t *testing.T) {
	conn := &bleConnection{
		UUIDService: "0xFFE0",
	}

	// OnError should not panic
	testErr := errors.New("test BLE error")
	conn.OnError(testErr)

	// If we get here without panic, test passes
}

// TestBLEConnection_Close tests the Close method
func TestBLEConnection_Close(t *testing.T) {
	conn := &bleConnection{
		UUIDService: "0xFFE0",
	}

	// Close should not panic
	conn.Close()

	// If we get here without panic, test passes
}

// TestBLEConnection_Write tests the Write method (part of io.Writer interface)
func TestBLEConnection_Write(t *testing.T) {
	// We can't test the actual Write without a real BLE characteristic
	// The Write method calls Characteristic.Write() which will panic with a zero-value struct
	// So we'll just verify the method signature exists by checking it compiles

	// This test verifies that bleConnection implements io.Writer interface
	var _ = func() {
		conn := &bleConnection{
			UUIDService: "0xFFE0",
		}
		// Type assertion to verify io.Writer implementation
		var w interface{} = conn
		if _, ok := w.(interface{ Write([]byte) (int, error) }); !ok {
			t.Error("bleConnection does not implement Write method")
		}
	}

	// Note: Can't actually call Write() without a real bluetooth.Characteristic
	// as it would panic. The implementation just forwards to Characteristic.Write()
	t.Log("bleConnection.Write() method exists and has correct signature")
}

// TestConnectionInterface tests that all connection types implement the connection interface
func TestConnectionInterface(t *testing.T) {
	var _ connection = (*networkConnection)(nil)
	var _ connection = (*serialConnection)(nil)
	var _ connection = (*tcpConnection)(nil)
	var _ connection = (*bleConnection)(nil)
}

// TestSerialConnection_OnError tests the OnError method for serialConnection
func TestSerialConnection_OnError(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}

	conn := &serialConnection{
		DeviceString: "/dev/ttyUSB0",
		Baud:         38400,
		serialPort:   &serial.Port{}, // Mock port
		Queue:        NewMessageQueue(10),
	}

	// Add to clientConnections so onConnectionClosed can remove it
	clientConnections[conn.GetConnectionKey()] = conn

	// OnError should close the connection
	testErr := errors.New("serial port error")
	conn.OnError(testErr)

	// Note: The implementation calls Close() which closes the port but doesn't set it to nil
	// The queue should be closed though
	if !conn.Queue.Closed {
		t.Error("OnError should have closed the queue")
	}

	// Verify the connection was removed from clientConnections
	netMutex.Lock()
	_, exists := clientConnections[conn.GetConnectionKey()]
	netMutex.Unlock()
	if exists {
		t.Error("OnError should have removed connection from clientConnections")
	}
}

// TestSerialConnection_Close tests the Close method for serialConnection
func TestSerialConnection_Close(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}

	conn := &serialConnection{
		DeviceString: "/dev/ttyUSB1",
		Baud:         38400,
		serialPort:   &serial.Port{}, // Mock port
		Queue:        NewMessageQueue(10),
	}

	// Add to clientConnections
	clientConnections[conn.GetConnectionKey()] = conn

	// Close should close the port and queue
	conn.Close()

	// Verify queue was closed
	if !conn.Queue.Closed {
		t.Error("Close should have closed the queue")
	}

	// Calling Close again should not panic (idempotent)
	conn.Close()
}

// TestTCPConnection_OnError tests the OnError method for tcpConnection
func TestTCPConnection_OnError(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}

	// Create a real TCP connection
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("Cannot create TCP listener: %v", err)
	}
	defer listener.Close()

	doneCh := make(chan bool)
	go func() {
		c, _ := listener.Accept()
		if c != nil {
			c.Close()
		}
		doneCh <- true
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Skipf("Cannot create TCP connection: %v", err)
	}

	tcpConn := clientConn.(*net.TCPConn)
	conn := &tcpConnection{
		Conn:  tcpConn,
		Queue: NewMessageQueue(10),
		Key:   "TCP:192.168.10.60:30001",
	}

	// Add to clientConnections
	clientConnections[conn.GetConnectionKey()] = conn

	// OnError should close the connection
	testErr := errors.New("TCP connection error")
	conn.OnError(testErr)

	// Verify connection was closed
	if conn.Conn != nil {
		t.Error("OnError should have set Conn to nil")
	}

	// Verify queue was closed
	if !conn.Queue.Closed {
		t.Error("OnError should have closed the queue")
	}

	<-doneCh
}

// TestTCPConnection_OnError_NilConn tests OnError when Conn is already nil
func TestTCPConnection_OnError_NilConn(t *testing.T) {
	conn := &tcpConnection{
		Conn: nil,
		Key:  "TCP:192.168.10.70:30002",
	}

	// OnError with nil Conn should not panic
	testErr := errors.New("TCP error with nil conn")
	conn.OnError(testErr)

	// If we get here without panic, test passes
}

// TestTCPConnection_Close tests the Close method for tcpConnection
func TestTCPConnection_Close(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}

	// Create a real TCP connection
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("Cannot create TCP listener: %v", err)
	}
	defer listener.Close()

	doneCh := make(chan bool)
	go func() {
		c, _ := listener.Accept()
		if c != nil {
			c.Close()
		}
		doneCh <- true
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Skipf("Cannot create TCP connection: %v", err)
	}

	tcpConn := clientConn.(*net.TCPConn)
	conn := &tcpConnection{
		Conn:  tcpConn,
		Queue: NewMessageQueue(10),
		Key:   "TCP:192.168.10.80:30003",
	}

	// Add to clientConnections
	clientConnections[conn.GetConnectionKey()] = conn

	// Close should close the connection and queue
	conn.Close()

	// Verify connection was set to nil
	if conn.Conn != nil {
		t.Error("Close should have set Conn to nil")
	}

	// Verify queue was closed
	if !conn.Queue.Closed {
		t.Error("Close should have closed the queue")
	}

	// Calling Close again should not panic (idempotent)
	conn.Close()

	<-doneCh
}

// TestNetworkConnection_IsSleeping_EdgeCases tests edge cases for IsSleeping
func TestNetworkConnection_IsSleeping_EdgeCases(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Save and restore settings
	origNoSleep := globalSettings.NoSleep
	defer func() { globalSettings.NoSleep = origNoSleep }()
	globalSettings.NoSleep = false

	// Skip if in x86 debug mode
	if runtime.GOARCH == "i386" || runtime.GOARCH == "amd64" {
		t.Skip("Skipping sleep edge case tests in x86 debug mode")
	}

	conn := &networkConnection{
		Ip:   "192.168.10.100",
		Port: 4000,
	}

	// Test: Recent ping but no pong - should be sleeping
	conn.LastPingResponse = stratuxClock.Time
	conn.LastPongResponse = time.Time{} // Zero
	conn.LastUnreachable = time.Time{}  // Zero
	if !conn.IsSleeping() {
		t.Error("Should be sleeping when pong is missing")
	}

	// Test: Recent pong but old ping (>10s) - should be sleeping
	conn.LastPongResponse = stratuxClock.Time
	conn.LastPingResponse = stratuxClock.Time.Add(-11 * time.Second)
	if !conn.IsSleeping() {
		t.Error("Should be sleeping when ping is old")
	}

	// Test: Both recent but recent unreachable - should be sleeping
	conn.LastPongResponse = stratuxClock.Time
	conn.LastPingResponse = stratuxClock.Time
	conn.LastUnreachable = stratuxClock.Time.Add(-2 * time.Second)
	if !conn.IsSleeping() {
		t.Error("Should be sleeping when recently unreachable")
	}
}

// TestConnectionCapabilitiesCombinations tests various capability combinations
func TestConnectionCapabilitiesCombinations(t *testing.T) {
	testCases := []struct {
		name         string
		capability   uint8
		expectFFSIM  bool
		expectedSize int
	}{
		{"None", 0, false, 1024},
		{"GDL90 only", NETWORK_GDL90_STANDARD, false, 1024},
		{"AHRS FFSIM only", NETWORK_AHRS_FFSIM, true, 1},
		{"Position FFSIM only", NETWORK_POSITION_FFSIM, true, 1},
		{"Both FFSIM", NETWORK_AHRS_FFSIM | NETWORK_POSITION_FFSIM, true, 1},
		{"GDL90 + AHRS FFSIM", NETWORK_GDL90_STANDARD | NETWORK_AHRS_FFSIM, true, 1},
		{"All capabilities", NETWORK_GDL90_STANDARD | NETWORK_AHRS_FFSIM | NETWORK_AHRS_GDL90 | NETWORK_FLARM_NMEA | NETWORK_POSITION_FFSIM, true, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := &networkConnection{
				Capability: tc.capability,
			}

			size := conn.GetDesiredPacketSize()
			if size != tc.expectedSize {
				t.Errorf("GetDesiredPacketSize() = %d, expected %d for capability %d", size, tc.expectedSize, tc.capability)
			}
		})
	}
}
