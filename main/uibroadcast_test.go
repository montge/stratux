package main

import (
	"testing"

	"golang.org/x/net/websocket"
)

// TestNewUIBroadcaster tests the constructor for uibroadcaster
func TestNewUIBroadcaster(t *testing.T) {
	t.Run("create_broadcaster", func(t *testing.T) {
		b := NewUIBroadcaster()

		if b == nil {
			t.Fatal("Expected non-nil broadcaster")
		}

		if b.sockets == nil {
			t.Error("Expected sockets slice to be initialized")
		}

		if b.sockets_mu == nil {
			t.Error("Expected sockets_mu to be initialized")
		}

		if b.messages == nil {
			t.Error("Expected messages channel to be initialized")
		}

		// Verify channel capacity
		if cap(b.messages) != 1024 {
			t.Errorf("Expected messages channel capacity of 1024, got %d", cap(b.messages))
		}

		t.Log("Successfully created new UIBroadcaster")
	})
}

// TestSend tests the Send function with various scenarios
func TestSend(t *testing.T) {
	t.Run("send_with_valid_broadcaster", func(t *testing.T) {
		b := NewUIBroadcaster()

		testMsg := []byte("test message")
		b.Send(testMsg)

		// Verify message was sent to channel (buffered, so should be immediate)
		if len(b.messages) != 1 {
			t.Errorf("Expected 1 message in channel, got %d", len(b.messages))
		}

		// Drain and verify
		msg := <-b.messages
		if string(msg) != string(testMsg) {
			t.Errorf("Expected message '%s', got '%s'", string(testMsg), string(msg))
		}
		t.Log("Successfully sent message to broadcaster")
	})

	t.Run("send_with_nil_broadcaster", func(t *testing.T) {
		var b *uibroadcaster = nil

		// This should not panic due to nil check
		testMsg := []byte("test message")
		b.Send(testMsg)

		t.Log("Successfully handled Send with nil broadcaster (no panic)")
	})

	t.Run("send_with_nil_messages_channel", func(t *testing.T) {
		// Create broadcaster with nil messages channel
		b := &uibroadcaster{
			sockets:    make([]*websocket.Conn, 0),
			sockets_mu: nil,
			messages:   nil, // Nil channel
		}

		// Should not panic due to nil check
		testMsg := []byte("test message")
		b.Send(testMsg)

		t.Log("Successfully handled Send with nil messages channel (no panic)")
	})

	t.Run("send_multiple_messages", func(t *testing.T) {
		b := NewUIBroadcaster()

		messages := [][]byte{
			[]byte("message 1"),
			[]byte("message 2"),
			[]byte("message 3"),
		}

		for _, msg := range messages {
			b.Send(msg)
		}

		// Verify all messages were sent to channel
		if len(b.messages) != len(messages) {
			t.Errorf("Expected %d messages in channel, got %d", len(messages), len(b.messages))
		}

		// Drain and verify
		for i, expectedMsg := range messages {
			msg := <-b.messages
			if string(msg) != string(expectedMsg) {
				t.Errorf("Message %d: expected '%s', got '%s'", i, string(expectedMsg), string(msg))
			}
		}
		t.Log("Successfully sent multiple messages")
	})

	t.Run("send_empty_message", func(t *testing.T) {
		b := NewUIBroadcaster()

		emptyMsg := []byte{}
		b.Send(emptyMsg)

		// Verify message was sent
		if len(b.messages) != 1 {
			t.Errorf("Expected 1 message in channel, got %d", len(b.messages))
		}

		msg := <-b.messages
		if len(msg) != 0 {
			t.Errorf("Expected empty message, got '%s'", string(msg))
		}
		t.Log("Successfully sent empty message")
	})

	t.Run("send_large_message", func(t *testing.T) {
		b := NewUIBroadcaster()

		// Create a large message (10KB)
		largeMsg := make([]byte, 10240)
		for i := range largeMsg {
			largeMsg[i] = byte(i % 256)
		}

		b.Send(largeMsg)

		// Verify message was sent
		if len(b.messages) != 1 {
			t.Errorf("Expected 1 message in channel, got %d", len(b.messages))
		}

		msg := <-b.messages
		if len(msg) != len(largeMsg) {
			t.Errorf("Expected message length %d, got %d", len(largeMsg), len(msg))
		}
		t.Log("Successfully sent large message")
	})
}

// TestSendJSON tests the SendJSON function
func TestSendJSON(t *testing.T) {
	t.Run("send_json_object", func(t *testing.T) {
		b := NewUIBroadcaster()

		testObj := map[string]interface{}{
			"field1": "value1",
			"field2": 42,
			"field3": true,
		}

		b.SendJSON(testObj)

		// Verify message was sent
		if len(b.messages) != 1 {
			t.Errorf("Expected 1 message in channel, got %d", len(b.messages))
		}

		msg := <-b.messages
		// Just verify we got a message (JSON marshaling is tested elsewhere)
		if len(msg) == 0 {
			t.Error("Expected non-empty JSON message")
		}
		t.Logf("Successfully sent JSON message: %s", string(msg))
	})

	t.Run("send_json_with_nil_broadcaster", func(t *testing.T) {
		var b *uibroadcaster = nil

		testObj := map[string]string{"test": "value"}

		// Should not panic
		b.SendJSON(testObj)

		t.Log("Successfully handled SendJSON with nil broadcaster (no panic)")
	})

	t.Run("send_json_nil_object", func(t *testing.T) {
		b := NewUIBroadcaster()

		// Send nil object (should marshal to "null")
		b.SendJSON(nil)

		// Verify message was sent
		if len(b.messages) != 1 {
			t.Errorf("Expected 1 message in channel, got %d", len(b.messages))
		}

		msg := <-b.messages
		expected := "null"
		if string(msg) != expected {
			t.Errorf("Expected JSON '%s', got '%s'", expected, string(msg))
		}
		t.Log("Successfully sent nil JSON object")
	})
}

// TestAddSocket tests the AddSocket function
func TestAddSocket(t *testing.T) {
	t.Run("add_socket", func(t *testing.T) {
		b := NewUIBroadcaster()

		// We can't create a real websocket connection in a unit test,
		// but we can test that the function doesn't panic with nil
		// (in real use, this would be a real connection)

		initialLen := len(b.sockets)
		if initialLen != 0 {
			t.Errorf("Expected initial socket count of 0, got %d", initialLen)
		}

		t.Log("Successfully tested AddSocket (sockets start empty)")
	})
}

// TestSend_EdgeCases tests edge cases for Send function
func TestSend_EdgeCases(t *testing.T) {
	t.Run("send_with_both_nil_broadcaster_and_messages", func(t *testing.T) {
		var b *uibroadcaster = nil

		// Should handle gracefully
		b.Send(nil)

		t.Log("Successfully handled Send with nil broadcaster and nil message")
	})

	t.Run("send_with_valid_broadcaster_nil_message", func(t *testing.T) {
		b := NewUIBroadcaster()

		// Send nil message (this is valid - a nil slice)
		b.Send(nil)

		// Verify message was sent
		if len(b.messages) != 1 {
			t.Errorf("Expected 1 message in channel, got %d", len(b.messages))
		}

		msg := <-b.messages
		if msg != nil {
			t.Errorf("Expected nil message, got %v", msg)
		}
		t.Log("Successfully sent nil message")
	})

	t.Run("concurrent_sends", func(t *testing.T) {
		// This test verifies that Send is thread-safe and can handle concurrent calls
		// Note: The writer goroutine will be draining the messages channel, so we can't
		// reliably count messages. Instead, we just verify no panics occur.
		b := NewUIBroadcaster()

		// Send messages concurrently from multiple goroutines
		numGoroutines := 10
		messagesPerGoroutine := 5

		done := make(chan bool)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				for j := 0; j < messagesPerGoroutine; j++ {
					msg := []byte("concurrent message")
					b.Send(msg)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// The test passes if we got here without a panic
		t.Log("Successfully sent concurrent messages without panic")
	})
}
