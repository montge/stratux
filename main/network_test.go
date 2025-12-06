/*
	Copyright (c) 2025 Stratux Development Team
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file.

	network_test.go: Unit tests for network.go

	Tests for network connection management functions.
*/

package main

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/tarm/serial"
)

// TestOnConnectionClosed_NetworkConnection tests removing a network (UDP) connection
func TestOnConnectionClosed_NetworkConnection(t *testing.T) {
	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}

	// Create a network connection
	conn := &networkConnection{
		Ip:         "192.168.10.50",
		Port:       4000,
		Capability: NETWORK_GDL90_STANDARD,
		Queue:      NewMessageQueue(10),
	}

	key := conn.GetConnectionKey()

	// Add connection to the map
	netMutex.Lock()
	clientConnections[key] = conn
	netMutex.Unlock()

	// Verify connection exists
	netMutex.Lock()
	_, exists := clientConnections[key]
	netMutex.Unlock()
	if !exists {
		t.Fatal("Connection should exist in clientConnections before onConnectionClosed")
	}

	// Call onConnectionClosed
	onConnectionClosed(conn)

	// Verify connection was removed
	netMutex.Lock()
	_, exists = clientConnections[key]
	netMutex.Unlock()
	if exists {
		t.Error("onConnectionClosed should have removed network connection from clientConnections")
	}
}

// TestOnConnectionClosed_TCPConnection tests removing a TCP connection
func TestOnConnectionClosed_TCPConnection(t *testing.T) {
	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}

	// Create a TCP connection with a unique key
	conn := &tcpConnection{
		Key:        "TCP:192.168.10.60:8080",
		Capability: NETWORK_GDL90_STANDARD,
		Queue:      NewMessageQueue(10),
	}

	key := conn.GetConnectionKey()

	// Add connection to the map
	netMutex.Lock()
	clientConnections[key] = conn
	netMutex.Unlock()

	// Verify connection exists
	netMutex.Lock()
	_, exists := clientConnections[key]
	netMutex.Unlock()
	if !exists {
		t.Fatal("Connection should exist in clientConnections before onConnectionClosed")
	}

	// Call onConnectionClosed
	onConnectionClosed(conn)

	// Verify connection was removed
	netMutex.Lock()
	_, exists = clientConnections[key]
	netMutex.Unlock()
	if exists {
		t.Error("onConnectionClosed should have removed TCP connection from clientConnections")
	}
}

// TestOnConnectionClosed_SerialConnection tests removing a serial connection
func TestOnConnectionClosed_SerialConnection(t *testing.T) {
	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}

	// Create a serial connection
	conn := &serialConnection{
		DeviceString: "/dev/ttyUSB0",
		Baud:         38400,
		Capability:   NETWORK_GDL90_STANDARD,
		serialPort:   &serial.Port{}, // Mock port
		Queue:        NewMessageQueue(10),
	}

	key := conn.GetConnectionKey()

	// Add connection to the map
	netMutex.Lock()
	clientConnections[key] = conn
	netMutex.Unlock()

	// Verify connection exists
	netMutex.Lock()
	_, exists := clientConnections[key]
	netMutex.Unlock()
	if !exists {
		t.Fatal("Connection should exist in clientConnections before onConnectionClosed")
	}

	// Call onConnectionClosed
	onConnectionClosed(conn)

	// Verify connection was removed
	netMutex.Lock()
	_, exists = clientConnections[key]
	netMutex.Unlock()
	if exists {
		t.Error("onConnectionClosed should have removed serial connection from clientConnections")
	}
}

// TestOnConnectionClosed_NilConnection tests handling of nil connection (edge case)
func TestOnConnectionClosed_NilConnection(t *testing.T) {
	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}

	// Save the current state of clientConnections
	netMutex.Lock()
	initialCount := len(clientConnections)
	netMutex.Unlock()

	// Call onConnectionClosed with nil connection - should not panic
	onConnectionClosed(nil)

	// Verify no changes were made to clientConnections
	netMutex.Lock()
	finalCount := len(clientConnections)
	netMutex.Unlock()

	if finalCount != initialCount {
		t.Errorf("onConnectionClosed(nil) should not modify clientConnections, expected count %d, got %d", initialCount, finalCount)
	}
}

// TestOnConnectionClosed_NonExistentConnection tests removing a connection that's not in the map (edge case)
func TestOnConnectionClosed_NonExistentConnection(t *testing.T) {
	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}

	// Create a connection that is NOT in clientConnections
	conn := &networkConnection{
		Ip:         "192.168.10.99",
		Port:       5000,
		Capability: NETWORK_GDL90_STANDARD,
		Queue:      NewMessageQueue(10),
	}

	key := conn.GetConnectionKey()

	// Verify connection does NOT exist in map
	netMutex.Lock()
	_, exists := clientConnections[key]
	netMutex.Unlock()
	if exists {
		t.Fatal("Connection should not exist in clientConnections before test")
	}

	// Save the current state of clientConnections
	netMutex.Lock()
	initialCount := len(clientConnections)
	netMutex.Unlock()

	// Call onConnectionClosed - should not panic even though connection doesn't exist
	onConnectionClosed(conn)

	// Verify no changes were made (delete on non-existent key is a no-op in Go)
	netMutex.Lock()
	finalCount := len(clientConnections)
	_, stillDoesNotExist := clientConnections[key]
	netMutex.Unlock()

	if finalCount != initialCount {
		t.Errorf("onConnectionClosed should not change map size when removing non-existent connection, expected %d, got %d", initialCount, finalCount)
	}

	if stillDoesNotExist {
		t.Error("Non-existent connection should still not exist after onConnectionClosed")
	}
}

// TestOnConnectionClosed_MultipleConnections tests that only the specified connection is removed
func TestOnConnectionClosed_MultipleConnections(t *testing.T) {
	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	// Clear the map for this test to ensure clean state
	clientConnections = make(map[string]connection)

	// Create multiple connections
	conn1 := &networkConnection{
		Ip:         "192.168.10.101",
		Port:       4001,
		Capability: NETWORK_GDL90_STANDARD,
		Queue:      NewMessageQueue(10),
	}

	conn2 := &tcpConnection{
		Key:        "TCP:192.168.10.102:8080",
		Capability: NETWORK_GDL90_STANDARD,
		Queue:      NewMessageQueue(10),
	}

	conn3 := &serialConnection{
		DeviceString: "/dev/ttyUSB1",
		Baud:         38400,
		Capability:   NETWORK_GDL90_STANDARD,
		serialPort:   &serial.Port{},
		Queue:        NewMessageQueue(10),
	}

	// Add all connections to the map
	netMutex.Lock()
	clientConnections[conn1.GetConnectionKey()] = conn1
	clientConnections[conn2.GetConnectionKey()] = conn2
	clientConnections[conn3.GetConnectionKey()] = conn3
	netMutex.Unlock()

	// Verify all connections exist
	netMutex.Lock()
	initialCount := len(clientConnections)
	netMutex.Unlock()
	if initialCount != 3 {
		t.Fatalf("Expected 3 connections initially, got %d", initialCount)
	}

	// Remove only conn2
	onConnectionClosed(conn2)

	// Verify conn2 was removed but conn1 and conn3 remain
	netMutex.Lock()
	_, exists1 := clientConnections[conn1.GetConnectionKey()]
	_, exists2 := clientConnections[conn2.GetConnectionKey()]
	_, exists3 := clientConnections[conn3.GetConnectionKey()]
	finalCount := len(clientConnections)
	netMutex.Unlock()

	if !exists1 {
		t.Error("conn1 should still exist after removing conn2")
	}
	if exists2 {
		t.Error("conn2 should have been removed")
	}
	if !exists3 {
		t.Error("conn3 should still exist after removing conn2")
	}
	if finalCount != 2 {
		t.Errorf("Expected 2 connections after removal, got %d", finalCount)
	}
}

// TestOnConnectionClosed_ConcurrentAccess tests thread safety of onConnectionClosed
func TestOnConnectionClosed_ConcurrentAccess(t *testing.T) {
	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	clientConnections = make(map[string]connection)

	// Create multiple connections
	const numConnections = 100
	connections := make([]*networkConnection, numConnections)

	for i := 0; i < numConnections; i++ {
		conn := &networkConnection{
			Ip:         "192.168.10." + string(rune(i)),
			Port:       uint32(5000 + i),
			Capability: NETWORK_GDL90_STANDARD,
			Queue:      NewMessageQueue(10),
		}
		connections[i] = conn

		netMutex.Lock()
		clientConnections[conn.GetConnectionKey()] = conn
		netMutex.Unlock()
	}

	// Remove all connections concurrently
	var wg sync.WaitGroup
	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(conn connection) {
			defer wg.Done()
			// Add a small random delay to increase likelihood of concurrent access
			time.Sleep(time.Microsecond * time.Duration(i%10))
			onConnectionClosed(conn)
		}(connections[i])
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify all connections were removed
	netMutex.Lock()
	finalCount := len(clientConnections)
	netMutex.Unlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 connections after concurrent removal, got %d", finalCount)
	}
}

// TestOnConnectionClosed_RealTCPConnection tests with an actual TCP connection
func TestOnConnectionClosed_RealTCPConnection(t *testing.T) {
	// Initialize required global variables
	if netMutex == nil {
		netMutex = &sync.Mutex{}
	}
	if clientConnections == nil {
		clientConnections = make(map[string]connection)
	}

	// Create a real TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create TCP listener: %v", err)
	}
	defer listener.Close()

	// Accept connection in a goroutine
	serverConnChan := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		serverConnChan <- conn
	}()

	// Connect to the listener
	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to dial TCP connection: %v", err)
	}

	// Wait for server to accept
	serverConn := <-serverConnChan

	// Create tcpConnection wrapper
	tcpConn := &tcpConnection{
		Conn:       serverConn.(*net.TCPConn),
		Key:        "TCP:" + serverConn.RemoteAddr().String(),
		Capability: NETWORK_GDL90_STANDARD,
		Queue:      NewMessageQueue(10),
	}

	key := tcpConn.GetConnectionKey()

	// Add to clientConnections
	netMutex.Lock()
	clientConnections[key] = tcpConn
	netMutex.Unlock()

	// Verify connection exists
	netMutex.Lock()
	_, exists := clientConnections[key]
	netMutex.Unlock()
	if !exists {
		t.Fatal("Connection should exist in clientConnections before onConnectionClosed")
	}

	// Call onConnectionClosed
	onConnectionClosed(tcpConn)

	// Verify connection was removed
	netMutex.Lock()
	_, exists = clientConnections[key]
	netMutex.Unlock()
	if exists {
		t.Error("onConnectionClosed should have removed TCP connection from clientConnections")
	}

	// Cleanup
	clientConn.Close()
}
