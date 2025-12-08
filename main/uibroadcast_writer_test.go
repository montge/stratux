package main

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

// TestWriter tests the writer function with real websocket connections
func TestWriter(t *testing.T) {
	t.Run("writer_sends_to_single_socket", func(t *testing.T) {
		b := NewUIBroadcaster()

		// Create a test server and websocket connection
		receivedMsgs := make([]string, 0)
		var mu sync.Mutex

		server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
			for {
				var msg string
				err := websocket.Message.Receive(ws, &msg)
				if err != nil {
					return
				}
				mu.Lock()
				receivedMsgs = append(receivedMsgs, msg)
				mu.Unlock()
			}
		}))
		defer server.Close()

		// Connect to the test server
		u, _ := url.Parse(server.URL)
		u.Scheme = "ws"
		ws, err := websocket.Dial(u.String(), "", "http://localhost")
		if err != nil {
			t.Fatalf("Failed to create websocket: %v", err)
		}
		defer ws.Close()

		// Add the socket to broadcaster
		b.AddSocket(ws)

		// Send a message
		testMsg := []byte("test message")
		b.Send(testMsg)

		// Give time for processing
		time.Sleep(100 * time.Millisecond)

		// Verify socket is still in the list
		b.sockets_mu.Lock()
		sockCount := len(b.sockets)
		b.sockets_mu.Unlock()

		if sockCount != 1 {
			t.Errorf("Expected 1 socket remaining, got %d", sockCount)
		}

		t.Log("Successfully sent message to single socket")
	})

	t.Run("writer_sends_to_multiple_sockets", func(t *testing.T) {
		b := NewUIBroadcaster()

		// Create multiple test connections
		var wg sync.WaitGroup
		numSockets := 3

		for i := 0; i < numSockets; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
					buf := make([]byte, 1024)
					for {
						_, err := ws.Read(buf)
						if err != nil {
							return
						}
					}
				}))
				defer server.Close()

				u, _ := url.Parse(server.URL)
				u.Scheme = "ws"
				ws, err := websocket.Dial(u.String(), "", "http://localhost")
				if err != nil {
					t.Errorf("Socket %d: Failed to create websocket: %v", id, err)
					return
				}
				defer ws.Close()

				b.AddSocket(ws)
			}(i)
		}

		wg.Wait()
		time.Sleep(50 * time.Millisecond)

		// Verify all sockets were added
		b.sockets_mu.Lock()
		sockCount := len(b.sockets)
		b.sockets_mu.Unlock()

		if sockCount != numSockets {
			t.Errorf("Expected %d sockets, got %d", numSockets, sockCount)
		}

		// Send a message
		testMsg := []byte("broadcast message")
		b.Send(testMsg)

		// Give time for processing
		time.Sleep(100 * time.Millisecond)

		t.Log("Successfully broadcast message to multiple sockets")
	})

	t.Run("writer_removes_failed_socket", func(t *testing.T) {
		b := NewUIBroadcaster()

		// Create a websocket and close it immediately to simulate failure
		server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
			buf := make([]byte, 1024)
			ws.Read(buf)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		u.Scheme = "ws"
		ws, err := websocket.Dial(u.String(), "", "http://localhost")
		if err != nil {
			t.Fatalf("Failed to create websocket: %v", err)
		}

		b.AddSocket(ws)

		// Close the socket to make writes fail
		ws.Close()

		// Send a message (should fail and remove socket)
		testMsg := []byte("test message")
		b.Send(testMsg)

		// Give time for processing
		time.Sleep(100 * time.Millisecond)

		// Verify socket was removed
		b.sockets_mu.Lock()
		sockCount := len(b.sockets)
		b.sockets_mu.Unlock()

		if sockCount != 0 {
			t.Errorf("Expected 0 sockets remaining (failed socket removed), got %d", sockCount)
		}

		t.Log("Successfully removed failed socket")
	})

	t.Run("writer_handles_multiple_messages", func(t *testing.T) {
		b := NewUIBroadcaster()

		// Create a test server
		server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
			buf := make([]byte, 1024)
			for {
				_, err := ws.Read(buf)
				if err != nil {
					return
				}
			}
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		u.Scheme = "ws"
		ws, err := websocket.Dial(u.String(), "", "http://localhost")
		if err != nil {
			t.Fatalf("Failed to create websocket: %v", err)
		}
		defer ws.Close()

		b.AddSocket(ws)

		// Send multiple messages
		messages := [][]byte{
			[]byte("message 1"),
			[]byte("message 2"),
			[]byte("message 3"),
		}

		for _, msg := range messages {
			b.Send(msg)
		}

		// Give time for processing
		time.Sleep(150 * time.Millisecond)

		// Verify socket is still connected
		b.sockets_mu.Lock()
		sockCount := len(b.sockets)
		b.sockets_mu.Unlock()

		if sockCount != 1 {
			t.Errorf("Expected 1 socket, got %d", sockCount)
		}

		t.Log("Successfully processed multiple messages")
	})

	t.Run("writer_with_no_sockets", func(t *testing.T) {
		b := NewUIBroadcaster()

		// Send a message with no sockets (should not panic)
		testMsg := []byte("test message")
		b.Send(testMsg)

		// Give time for processing
		time.Sleep(50 * time.Millisecond)

		// Verify no sockets
		b.sockets_mu.Lock()
		sockCount := len(b.sockets)
		b.sockets_mu.Unlock()

		if sockCount != 0 {
			t.Errorf("Expected 0 sockets, got %d", sockCount)
		}

		t.Log("Successfully handled message with no sockets")
	})

	t.Run("writer_mixed_success_and_failure", func(t *testing.T) {
		b := NewUIBroadcaster()

		// Create good sockets
		goodSockets := 2
		for i := 0; i < goodSockets; i++ {
			server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
				buf := make([]byte, 1024)
				for {
					_, err := ws.Read(buf)
					if err != nil {
						return
					}
				}
			}))
			defer server.Close()

			u, _ := url.Parse(server.URL)
			u.Scheme = "ws"
			ws, err := websocket.Dial(u.String(), "", "http://localhost")
			if err != nil {
				t.Fatalf("Failed to create good websocket: %v", err)
			}
			defer ws.Close()

			b.AddSocket(ws)
		}

		// Create bad sockets (that will be closed)
		badSockets := 2
		for i := 0; i < badSockets; i++ {
			server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
				buf := make([]byte, 1024)
				ws.Read(buf)
			}))
			defer server.Close()

			u, _ := url.Parse(server.URL)
			u.Scheme = "ws"
			ws, err := websocket.Dial(u.String(), "", "http://localhost")
			if err != nil {
				t.Fatalf("Failed to create bad websocket: %v", err)
			}

			b.AddSocket(ws)
			ws.Close() // Close immediately to make it fail
		}

		time.Sleep(50 * time.Millisecond)

		// Send a message
		testMsg := []byte("mixed test")
		b.Send(testMsg)

		// Give time for processing
		time.Sleep(100 * time.Millisecond)

		// Verify only good sockets remain
		b.sockets_mu.Lock()
		sockCount := len(b.sockets)
		b.sockets_mu.Unlock()

		if sockCount != goodSockets {
			t.Errorf("Expected %d sockets remaining, got %d", goodSockets, sockCount)
		}

		t.Log(fmt.Sprintf("Successfully handled mixed sockets: %d good, %d bad", goodSockets, badSockets))
	})

	t.Run("writer_empty_message", func(t *testing.T) {
		b := NewUIBroadcaster()

		server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
			buf := make([]byte, 1024)
			for {
				_, err := ws.Read(buf)
				if err != nil {
					return
				}
			}
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		u.Scheme = "ws"
		ws, err := websocket.Dial(u.String(), "", "http://localhost")
		if err != nil {
			t.Fatalf("Failed to create websocket: %v", err)
		}
		defer ws.Close()

		b.AddSocket(ws)

		// Send empty message
		emptyMsg := []byte{}
		b.Send(emptyMsg)

		// Give time for processing
		time.Sleep(50 * time.Millisecond)

		// Verify socket is still connected
		b.sockets_mu.Lock()
		sockCount := len(b.sockets)
		b.sockets_mu.Unlock()

		if sockCount != 1 {
			t.Errorf("Expected 1 socket, got %d", sockCount)
		}

		t.Log("Successfully sent empty message")
	})
}
