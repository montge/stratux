package main

import (
	"testing"
	"time"
)

// TestNewMessageQueue tests the MessageQueue constructor
func TestNewMessageQueue(t *testing.T) {
	queue := NewMessageQueue(100)

	if queue == nil {
		t.Fatal("NewMessageQueue returned nil")
	}
	if queue.maxSize != 100 {
		t.Errorf("maxSize = %d, expected 100", queue.maxSize)
	}
	if queue.entries == nil {
		t.Error("entries should not be nil")
	}
	if queue.DataAvailable == nil {
		t.Error("DataAvailable channel should not be nil")
	}
	if queue.Closed {
		t.Error("Closed should be false initially")
	}
}

// TestMessageQueuePutAndPeek tests adding and peeking entries
func TestMessageQueuePutAndPeek(t *testing.T) {
	// Ensure stratuxClock is initialized
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Add a single entry
	queue.Put(100, 10*time.Second, "test1")

	data, prio := queue.PeekFirst()
	if data == nil {
		t.Fatal("PeekFirst returned nil")
	}
	if data.(string) != "test1" {
		t.Errorf("PeekFirst data = %v, expected 'test1'", data)
	}
	if prio != 100 {
		t.Errorf("PeekFirst priority = %d, expected 100", prio)
	}

	// Peek should not remove the entry
	data2, _ := queue.PeekFirst()
	if data2 == nil {
		t.Error("Second PeekFirst returned nil, entry was removed")
	}
}

// TestMessageQueuePutAndPop tests adding and popping entries
func TestMessageQueuePutAndPop(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Add entries
	queue.Put(100, 10*time.Second, "test1")
	queue.Put(200, 10*time.Second, "test2")

	// Pop first entry
	data, prio := queue.PopFirst()
	if data == nil {
		t.Fatal("PopFirst returned nil")
	}
	if data.(string) != "test1" {
		t.Errorf("PopFirst data = %v, expected 'test1'", data)
	}
	if prio != 100 {
		t.Errorf("PopFirst priority = %d, expected 100", prio)
	}

	// Pop should have removed the first entry
	data2, prio2 := queue.PopFirst()
	if data2.(string) != "test2" {
		t.Errorf("Second PopFirst data = %v, expected 'test2'", data2)
	}
	if prio2 != 200 {
		t.Errorf("Second PopFirst priority = %d, expected 200", prio2)
	}
}

// TestMessageQueuePriorityOrdering tests that entries are returned in priority order
func TestMessageQueuePriorityOrdering(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(100)

	// Add entries in non-sorted order
	queue.Put(300, 10*time.Second, "low")
	queue.Put(100, 10*time.Second, "high")
	queue.Put(200, 10*time.Second, "medium")

	// Should be returned in priority order (lowest first)
	data1, prio1 := queue.PopFirst()
	if prio1 != 100 || data1.(string) != "high" {
		t.Errorf("First entry: prio=%d data=%v, expected prio=100 data='high'", prio1, data1)
	}

	data2, prio2 := queue.PopFirst()
	if prio2 != 200 || data2.(string) != "medium" {
		t.Errorf("Second entry: prio=%d data=%v, expected prio=200 data='medium'", prio2, data2)
	}

	data3, prio3 := queue.PopFirst()
	if prio3 != 300 || data3.(string) != "low" {
		t.Errorf("Third entry: prio=%d data=%v, expected prio=300 data='low'", prio3, data3)
	}
}

// TestMessageQueueEmptyQueue tests behavior with empty queue
func TestMessageQueueEmptyQueue(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	data, prio := queue.PopFirst()
	if data != nil {
		t.Errorf("PopFirst on empty queue returned %v, expected nil", data)
	}
	if prio != 0 {
		t.Errorf("PopFirst priority = %d, expected 0", prio)
	}

	data2, prio2 := queue.PeekFirst()
	if data2 != nil {
		t.Errorf("PeekFirst on empty queue returned %v, expected nil", data2)
	}
	if prio2 != 0 {
		t.Errorf("PeekFirst priority = %d, expected 0", prio2)
	}
}

// TestMessageQueueSamePriorityFIFO tests that same-priority entries maintain FIFO order
func TestMessageQueueSamePriorityFIFO(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(100)

	// Add multiple entries with same priority
	queue.Put(100, 10*time.Second, "first")
	queue.Put(100, 10*time.Second, "second")
	queue.Put(100, 10*time.Second, "third")

	// Should be returned in insertion order (FIFO)
	data1, _ := queue.PopFirst()
	if data1.(string) != "first" {
		t.Errorf("First entry = %v, expected 'first'", data1)
	}

	data2, _ := queue.PopFirst()
	if data2.(string) != "second" {
		t.Errorf("Second entry = %v, expected 'second'", data2)
	}

	data3, _ := queue.PopFirst()
	if data3.(string) != "third" {
		t.Errorf("Third entry = %v, expected 'third'", data3)
	}
}

// TestMessageQueuePruning tests that the queue prunes correctly when exceeding maxSize
func TestMessageQueuePruning(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(5)

	// Add entries to exceed maxSize by 10% to trigger pruning
	for i := 0; i < 10; i++ {
		// Lower priority entries should be pruned first
		prio := int32(100 + i*10)
		queue.Put(prio, 10*time.Second, i)
	}

	// Queue should have pruned to maxSize or close to it
	dump := queue.GetQueueDump(false)
	if len(dump) > 6 { // Allow 10% overage
		t.Errorf("Queue size = %d, expected <= 6 after pruning", len(dump))
	}

	// Verify that higher priority entries are still present
	data, prio := queue.PeekFirst()
	if data == nil {
		t.Fatal("Queue is empty after pruning")
	}
	if prio < 100 || prio > 200 {
		t.Errorf("First entry priority = %d, expected low priority (100-200)", prio)
	}
}

// TestMessageQueueGetQueueDump tests the GetQueueDump method
func TestMessageQueueGetQueueDump(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	queue.Put(100, 10*time.Second, "a")
	queue.Put(200, 10*time.Second, "b")
	queue.Put(300, 10*time.Second, "c")

	dump := queue.GetQueueDump(false)
	if len(dump) != 3 {
		t.Errorf("GetQueueDump length = %d, expected 3", len(dump))
	}

	// Should be in priority order
	if dump[0].(string) != "a" {
		t.Errorf("dump[0] = %v, expected 'a'", dump[0])
	}
	if dump[1].(string) != "b" {
		t.Errorf("dump[1] = %v, expected 'b'", dump[1])
	}
	if dump[2].(string) != "c" {
		t.Errorf("dump[2] = %v, expected 'c'", dump[2])
	}
}

// TestMessageQueueClose tests the Close method
func TestMessageQueueClose(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	if queue.Closed {
		t.Error("Queue should not be closed initially")
	}

	queue.Close()

	if !queue.Closed {
		t.Error("Queue should be closed after Close()")
	}

	// Verify that Put does nothing after close
	queue.Put(100, 10*time.Second, "test")
	dump := queue.GetQueueDump(false)
	if len(dump) != 0 {
		t.Errorf("Closed queue accepted Put, length = %d", len(dump))
	}

	// Verify Close is idempotent
	queue.Close()
	if !queue.Closed {
		t.Error("Queue should still be closed after second Close()")
	}
}

// TestMessageQueueExpiredEntries tests that expired entries are handled correctly
func TestMessageQueueExpiredEntries(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Add entries with very short timeout
	queue.Put(100, 1*time.Millisecond, "expired")

	// Wait for entry to expire
	time.Sleep(50 * time.Millisecond)

	// Add a non-expired entry
	queue.Put(200, 10*time.Second, "valid")

	// PopFirst should skip the expired entry and return the valid one
	data, prio := queue.PopFirst()
	if data == nil {
		t.Fatal("PopFirst returned nil")
	}
	if data.(string) != "valid" {
		t.Errorf("PopFirst data = %v, expected 'valid' (expired entry should be skipped)", data)
	}
	if prio != 200 {
		t.Errorf("PopFirst priority = %d, expected 200", prio)
	}
}

// TestMessageQueueFindInsertPosition tests the binary search for insertion
func TestMessageQueueFindInsertPosition(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Add entries to create a sorted queue
	queue.Put(100, 10*time.Second, "a")
	queue.Put(300, 10*time.Second, "c")
	queue.Put(500, 10*time.Second, "e")

	// Test findInsertPosition directly
	pos := queue.findInsertPosition(200)
	if pos != 1 {
		t.Errorf("findInsertPosition(200) = %d, expected 1", pos)
	}

	pos = queue.findInsertPosition(50)
	if pos != 0 {
		t.Errorf("findInsertPosition(50) = %d, expected 0", pos)
	}

	pos = queue.findInsertPosition(600)
	if pos != 3 {
		t.Errorf("findInsertPosition(600) = %d, expected 3", pos)
	}
}

// TestMessageQueueMixedPriorities tests complex scenarios with mixed priorities
func TestMessageQueueMixedPriorities(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(100)

	// Add entries with different priorities in random order
	priorities := []int32{500, 100, 300, 200, 400, 150, 250, 350, 450}
	for _, prio := range priorities {
		queue.Put(prio, 10*time.Second, prio)
	}

	// Verify they come out in sorted order
	prevPrio := int32(-1)
	for i := 0; i < len(priorities); i++ {
		data, prio := queue.PopFirst()
		if data == nil {
			t.Fatalf("Entry %d: PopFirst returned nil", i)
		}
		if prio <= prevPrio {
			t.Errorf("Entry %d: priority = %d, previous = %d (not in ascending order)", i, prio, prevPrio)
		}
		prevPrio = prio
	}
}

// TestMessageQueueGetQueueDumpWithPrune tests GetQueueDump with pruning enabled
func TestMessageQueueGetQueueDumpWithPrune(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(5)

	// Add more entries than maxSize with low priorities
	for i := 0; i < 10; i++ {
		queue.Put(int32(100+i*10), 10*time.Second, i)
	}

	// Force pruning
	dump := queue.GetQueueDump(true)

	// After pruning, should be at or near maxSize
	if len(dump) > queue.maxSize {
		t.Errorf("After pruning, queue size = %d, expected <= %d", len(dump), queue.maxSize)
	}
}

// TestMessageQueueDataAvailableChannel tests the DataAvailable notification
func TestMessageQueueDataAvailableChannel(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Add an entry and check if channel is notified
	queue.Put(100, 10*time.Second, "test")

	// Non-blocking check if channel has data
	select {
	case <-queue.DataAvailable:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("DataAvailable channel was not notified after Put")
	}
}
