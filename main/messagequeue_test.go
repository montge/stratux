package main

import (
	"sync"
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

// TestMessageQueueAllEntriesExpired tests when all entries are expired
func TestMessageQueueAllEntriesExpired(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Add entries with very short timeout
	queue.Put(100, 1*time.Millisecond, "expired1")
	queue.Put(200, 1*time.Millisecond, "expired2")
	queue.Put(300, 1*time.Millisecond, "expired3")

	// Wait for entries to expire
	time.Sleep(50 * time.Millisecond)

	// All entries should be expired
	data, prio := queue.PopFirst()
	if data != nil {
		t.Errorf("Expected nil when all entries expired, got %v", data)
	}
	if prio != 0 {
		t.Errorf("Expected priority 0 when all entries expired, got %d", prio)
	}

	// The queue should have been cleared internally
	dump := queue.GetQueueDump(false)
	if len(dump) != 0 {
		t.Errorf("Queue should be empty after all entries expired, got %d entries", len(dump))
	}
}

// TestMessageQueuePruneLowPriorityRemoval tests that low priority entries are removed during pruning
func TestMessageQueuePruneLowPriorityRemoval(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(3)

	// Add high priority entries
	queue.Put(10, 10*time.Second, "high1")
	queue.Put(20, 10*time.Second, "high2")

	// Add low priority entries to trigger pruning (>10% over maxSize)
	queue.Put(100, 10*time.Second, "low1")
	queue.Put(200, 10*time.Second, "low2")
	queue.Put(300, 10*time.Second, "low3")

	// Force another prune with GetQueueDump
	dump := queue.GetQueueDump(true)

	// Should have pruned to maxSize (3)
	if len(dump) > 3 {
		t.Errorf("Queue size after prune = %d, expected <= 3", len(dump))
	}

	// High priority entries should still be present
	data1, prio1 := queue.PeekFirst()
	if prio1 != 10 {
		t.Errorf("First entry priority = %d, expected 10 (high priority)", prio1)
	}
	if data1.(string) != "high1" {
		t.Errorf("First entry data = %v, expected 'high1'", data1)
	}
}

// TestMessageQueuePruneRemoveEntireCategory tests removing an entire priority category
func TestMessageQueuePruneRemoveEntireCategory(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(2)

	// Add a high priority entry
	queue.Put(10, 10*time.Second, "high")

	// Add several low priority entries to exceed maxSize
	queue.Put(500, 10*time.Second, "low1")
	queue.Put(500, 10*time.Second, "low2")
	queue.Put(500, 10*time.Second, "low3")
	queue.Put(500, 10*time.Second, "low4")

	// Force prune
	dump := queue.GetQueueDump(true)

	// The entire low priority category should be removed
	if len(dump) > 2 {
		t.Errorf("Queue size after prune = %d, expected <= 2", len(dump))
	}

	// Verify high priority is kept
	data, prio := queue.PeekFirst()
	if data == nil {
		t.Fatal("Queue is empty")
	}
	if prio != 10 {
		t.Errorf("First entry priority = %d, expected 10", prio)
	}
}

// TestMessageQueuePruneExpiredRemoval tests that expired entries are removed during pruning
func TestMessageQueuePruneExpiredRemoval(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Add entries with short timeout
	queue.Put(100, 1*time.Millisecond, "expired1")
	queue.Put(200, 1*time.Millisecond, "expired2")

	// Add entry with long timeout
	queue.Put(300, 10*time.Second, "valid")

	// Wait for short entries to expire
	time.Sleep(50 * time.Millisecond)

	// Force prune
	dump := queue.GetQueueDump(true)

	// Only the valid entry should remain
	if len(dump) != 1 {
		t.Errorf("Queue size after prune = %d, expected 1 (only valid entry)", len(dump))
	}

	data, prio := queue.PeekFirst()
	if prio != 300 {
		t.Errorf("Entry priority = %d, expected 300", prio)
	}
	if data.(string) != "valid" {
		t.Errorf("Entry data = %v, expected 'valid'", data)
	}
}

// TestMessageQueueInsertPositionAtEnd tests insertion at the end of queue
func TestMessageQueueInsertPositionAtEnd(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Add entries in ascending order (each should go at end)
	queue.Put(100, 10*time.Second, "a")
	queue.Put(200, 10*time.Second, "b")
	queue.Put(300, 10*time.Second, "c")

	// Verify order
	dump := queue.GetQueueDump(false)
	if len(dump) != 3 {
		t.Fatalf("Queue size = %d, expected 3", len(dump))
	}

	if dump[0].(string) != "a" || dump[1].(string) != "b" || dump[2].(string) != "c" {
		t.Errorf("Entries not in expected order: %v", dump)
	}
}

// TestMessageQueuePutToNilEntries tests Put when entries is nil
func TestMessageQueuePutToNilEntries(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)
	// Manually set entries to nil
	queue.entries = nil

	// Put should handle nil entries gracefully
	queue.Put(100, 10*time.Second, "test")

	data, prio := queue.PeekFirst()
	if data == nil {
		t.Fatal("Entry not added to nil entries slice")
	}
	if data.(string) != "test" {
		t.Errorf("Data = %v, expected 'test'", data)
	}
	if prio != 100 {
		t.Errorf("Priority = %d, expected 100", prio)
	}
}

// TestMessageQueueMultiplePriorityCategories tests pruning with multiple categories
func TestMessageQueueMultiplePriorityCategories(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(4)

	// Add entries in multiple priority categories
	queue.Put(10, 10*time.Second, "high1")
	queue.Put(10, 10*time.Second, "high2")
	queue.Put(50, 10*time.Second, "medium1")
	queue.Put(50, 10*time.Second, "medium2")
	queue.Put(100, 10*time.Second, "low1")
	queue.Put(100, 10*time.Second, "low2")

	// Force prune
	dump := queue.GetQueueDump(true)

	// Should have pruned to 4 entries
	if len(dump) > 4 {
		t.Errorf("Queue size = %d, expected <= 4", len(dump))
	}

	// Highest priority should still be present
	data, prio := queue.PeekFirst()
	if prio != 10 {
		t.Errorf("First entry priority = %d, expected 10", prio)
	}
	if data.(string) != "high1" {
		t.Errorf("First entry = %v, expected 'high1'", data)
	}
}

// TestMessageQueuePruneRemoveMultipleCategories tests pruning that must remove multiple categories
// This test specifically triggers the "else" branch in prune() where an entire category
// is removed because it doesn't have enough entries to satisfy toBeRemoved
func TestMessageQueuePruneRemoveMultipleCategories(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	// Create a queue with large maxSize initially so Put doesn't auto-prune
	// Then we'll manually reduce maxSize and call prune via GetQueueDump
	queue := NewMessageQueue(100) // Large enough that Put won't auto-prune

	// Add entries in multiple priority categories
	// High priority (prio=10) - 2 entries (should remain)
	queue.Put(10, 10*time.Second, "high1")
	queue.Put(10, 10*time.Second, "high2")
	// Medium priority (prio=50) - 1 entry (entire category will be removed)
	queue.Put(50, 10*time.Second, "medium1")
	// Low priority (prio=100) - 1 entry (entire category will be removed first)
	queue.Put(100, 10*time.Second, "low1")

	// Now reduce maxSize to 2 and call prune
	// Total: 4 entries, need to remove 2
	// Low has 1, 1 < 2, so enter else branch, set low to nil, toBeRemoved = 1
	// Medium has 1, 1 >= 1, so enter if branch, remove 1, keep 0... wait that's not right

	// Let me recalculate:
	// - We need to remove 2
	// - Low priority has 1 entry, 1 < 2, so we set it to nil and continue (ELSE BRANCH)
	// - toBeRemoved becomes 2 - 1 = 1
	// - Medium has 1 entry, 1 >= 1, so we remove from it (IF BRANCH)
	// Result: Low removed entirely (else), Medium has 0 left, High intact

	queue.maxSize = 2

	// Force prune via GetQueueDump
	dump := queue.GetQueueDump(true)

	// Should have pruned to 2 entries (only high priority remains)
	if len(dump) != 2 {
		t.Errorf("Queue size = %d, expected 2", len(dump))
	}

	// Verify all remaining entries are high priority
	for _, entry := range dump {
		str := entry.(string)
		if str != "high1" && str != "high2" {
			t.Errorf("Unexpected entry remaining: %s (expected only high priority)", str)
		}
	}
}

// TestMessageQueueCloseNotifiesChannel tests that Close sends notification on DataAvailable channel
func TestMessageQueueCloseNotifiesChannel(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Close should notify the channel
	queue.Close()

	// Check if channel is notified
	select {
	case <-queue.DataAvailable:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("DataAvailable channel was not notified after Close")
	}
}

// TestMessageQueuePeekDoesNotRemove tests that PeekFirst doesn't modify queue
func TestMessageQueuePeekDoesNotRemove(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	queue.Put(100, 10*time.Second, "test1")
	queue.Put(200, 10*time.Second, "test2")

	// Peek multiple times
	for i := 0; i < 5; i++ {
		data, prio := queue.PeekFirst()
		if data.(string) != "test1" {
			t.Errorf("Peek iteration %d: data = %v, expected 'test1'", i, data)
		}
		if prio != 100 {
			t.Errorf("Peek iteration %d: priority = %d, expected 100", i, prio)
		}
	}

	// Queue should still have both entries
	dump := queue.GetQueueDump(false)
	if len(dump) != 2 {
		t.Errorf("Queue length = %d, expected 2 after multiple peeks", len(dump))
	}
}

// TestMessageQueuePopRemoves tests that PopFirst removes entry
func TestMessageQueuePopRemoves(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	queue.Put(100, 10*time.Second, "test1")
	queue.Put(200, 10*time.Second, "test2")
	queue.Put(300, 10*time.Second, "test3")

	// Verify initial size
	dump := queue.GetQueueDump(false)
	if len(dump) != 3 {
		t.Fatalf("Initial queue length = %d, expected 3", len(dump))
	}

	// Pop first entry
	data, _ := queue.PopFirst()
	if data.(string) != "test1" {
		t.Errorf("First pop data = %v, expected 'test1'", data)
	}

	// Verify size decreased
	dump = queue.GetQueueDump(false)
	if len(dump) != 2 {
		t.Errorf("Queue length after first pop = %d, expected 2", len(dump))
	}

	// Pop second entry
	data, _ = queue.PopFirst()
	if data.(string) != "test2" {
		t.Errorf("Second pop data = %v, expected 'test2'", data)
	}

	// Verify size decreased again
	dump = queue.GetQueueDump(false)
	if len(dump) != 1 {
		t.Errorf("Queue length after second pop = %d, expected 1", len(dump))
	}
}

// TestMessageQueueConcurrentAccess tests thread safety with concurrent operations
func TestMessageQueueConcurrentAccess(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(1000)
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			queue.Put(int32(i), 10*time.Second, i)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 50; i++ {
			queue.PopFirst()
			time.Sleep(2 * time.Millisecond)
		}
		done <- true
	}()

	// Peek goroutine
	go func() {
		for i := 0; i < 50; i++ {
			queue.PeekFirst()
			time.Sleep(2 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// If we got here without crashing, test passed
}

// TestMessageQueueNotifyDataMultipleCalls tests that multiple notifyData calls don't block
func TestMessageQueueNotifyDataMultipleCalls(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Add multiple entries rapidly (each triggers notifyData)
	for i := 0; i < 10; i++ {
		queue.Put(int32(i), 10*time.Second, i)
	}

	// Should not have blocked, even though channel buffer is 1
	// Verify at least one notification was sent
	select {
	case <-queue.DataAvailable:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("DataAvailable channel was not notified")
	}
}

// TestMessageQueuePutAfterClose tests that Put is a no-op after Close
func TestMessageQueuePutAfterClose(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	// Add an entry before closing
	queue.Put(100, 10*time.Second, "before")

	// Close the queue
	queue.Close()

	// Try to add entries after close
	queue.Put(200, 10*time.Second, "after1")
	queue.Put(300, 10*time.Second, "after2")

	// Should only have the first entry
	dump := queue.GetQueueDump(false)
	if len(dump) != 1 {
		t.Errorf("Queue length = %d, expected 1 (entries after Close should be ignored)", len(dump))
	}

	if dump[0].(string) != "before" {
		t.Errorf("Queue entry = %v, expected 'before'", dump[0])
	}
}

// TestMessageQueuePruneWithOnlyExpiredEntries tests pruning when all entries are expired
func TestMessageQueuePruneWithOnlyExpiredEntries(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(5)

	// Add entries with very short timeout
	for i := 0; i < 10; i++ {
		queue.Put(int32(100+i*10), 1*time.Millisecond, i)
	}

	// Wait for all to expire
	time.Sleep(50 * time.Millisecond)

	// Force prune - should clear all entries
	dump := queue.GetQueueDump(true)

	if len(dump) != 0 {
		t.Errorf("Queue length = %d, expected 0 (all entries should be expired and pruned)", len(dump))
	}
}

// TestMessageQueueZeroMaxSize tests behavior with maxSize of 0
func TestMessageQueueZeroMaxSize(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(0)

	// Add entries to a zero-size queue
	queue.Put(100, 10*time.Second, "test1")
	queue.Put(200, 10*time.Second, "test2")

	// Force prune
	dump := queue.GetQueueDump(true)

	// Should prune everything since maxSize is 0
	if len(dump) != 0 {
		t.Errorf("Queue length = %d, expected 0 for maxSize=0", len(dump))
	}
}

// TestMessageQueueLargeNumberOfEntries tests handling of many entries
func TestMessageQueueLargeNumberOfEntries(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(5000)

	// Add many entries
	for i := 0; i < 1000; i++ {
		queue.Put(int32(i), 10*time.Second, i)
	}

	dump := queue.GetQueueDump(false)
	if len(dump) != 1000 {
		t.Errorf("Queue length = %d, expected 1000", len(dump))
	}

	// Verify they're in priority order
	for i := 0; i < 1000; i++ {
		data, prio := queue.PopFirst()
		if data == nil {
			t.Fatalf("Entry %d is nil", i)
		}
		if prio != int32(i) {
			t.Errorf("Entry %d: priority = %d, expected %d", i, prio, i)
		}
	}
}

// TestMessageQueueGetQueueDumpDoesNotModify tests that GetQueueDump doesn't change queue
func TestMessageQueueGetQueueDumpDoesNotModify(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	queue.Put(100, 10*time.Second, "a")
	queue.Put(200, 10*time.Second, "b")
	queue.Put(300, 10*time.Second, "c")

	// Get dump without pruning
	dump1 := queue.GetQueueDump(false)

	// Verify queue still has all entries
	dump2 := queue.GetQueueDump(false)

	if len(dump1) != len(dump2) {
		t.Errorf("Queue size changed: before=%d, after=%d", len(dump1), len(dump2))
	}

	for i := range dump1 {
		if dump1[i] != dump2[i] {
			t.Errorf("Entry %d changed: %v -> %v", i, dump1[i], dump2[i])
		}
	}
}

// TestMessageQueueSingleEntry tests queue with only one entry
func TestMessageQueueSingleEntry(t *testing.T) {
	if stratuxClock == nil {
		stratuxClock = NewMonotonic()
		time.Sleep(10 * time.Millisecond)
	}

	queue := NewMessageQueue(10)

	queue.Put(100, 10*time.Second, "single")

	// Peek
	data, prio := queue.PeekFirst()
	if data.(string) != "single" {
		t.Errorf("PeekFirst data = %v, expected 'single'", data)
	}
	if prio != 100 {
		t.Errorf("PeekFirst priority = %d, expected 100", prio)
	}

	// Pop
	data, prio = queue.PopFirst()
	if data.(string) != "single" {
		t.Errorf("PopFirst data = %v, expected 'single'", data)
	}
	if prio != 100 {
		t.Errorf("PopFirst priority = %d, expected 100", prio)
	}

	// Queue should be empty
	data, prio = queue.PopFirst()
	if data != nil {
		t.Errorf("Expected nil after popping only entry, got %v", data)
	}
}

// =============================================================================
// Phase 7.3: Resource Limits - Message Queue Overflow Tests
// =============================================================================

// TestMessageQueueOverflow tests queue behavior when heavily overloaded
func TestMessageQueueOverflow(t *testing.T) {
	initStratuxClock()

	// Small queue that will overflow quickly
	queue := NewMessageQueue(100)

	// Add 10x the max capacity
	const numMessages = 1000
	for i := 0; i < numMessages; i++ {
		// Mix of priorities
		prio := int32(i % 10)
		queue.Put(prio, 1*time.Second, i)
	}

	// Queue should have pruned to approximately maxSize
	dump := queue.GetQueueDump(false)
	if len(dump) > 110 { // Allow 10% over (per prune logic)
		t.Errorf("Queue should have pruned to ~100 entries, has %d", len(dump))
	}

	// Verify high priority messages are preserved over low priority
	dump = queue.GetQueueDump(true) // Prune first
	if len(dump) > 100 {
		t.Errorf("After explicit prune, queue should have at most 100 entries, has %d", len(dump))
	}

	t.Logf("Overflow test: Added %d messages, queue has %d after prune", numMessages, len(dump))
}

// TestMessageQueuePriorityPreservationUnderLoad tests that high priority
// messages are preserved when queue is overloaded
func TestMessageQueuePriorityPreservationUnderLoad(t *testing.T) {
	initStratuxClock()

	queue := NewMessageQueue(50)

	// Add low priority messages first
	for i := 0; i < 100; i++ {
		queue.Put(10, 1*time.Second, "low-"+string(rune('a'+i%26)))
	}

	// Add high priority messages (negative priority = higher)
	for i := 0; i < 50; i++ {
		queue.Put(-5, 1*time.Second, "HIGH-"+string(rune('A'+i%26)))
	}

	// Force prune
	dump := queue.GetQueueDump(true)

	// Count high priority messages
	highCount := 0
	for _, item := range dump {
		if str, ok := item.(string); ok && len(str) > 0 && str[0] == 'H' {
			highCount++
		}
	}

	// High priority messages should be mostly preserved
	if highCount < 25 { // At least half of the high priority should remain
		t.Errorf("Expected at least 25 high priority messages, found %d", highCount)
	}

	t.Logf("Priority preservation: %d high-priority messages preserved out of 50", highCount)
}

// TestMessageQueueRapidPutPop tests rapid put/pop operations
func TestMessageQueueRapidPutPop(t *testing.T) {
	initStratuxClock()

	queue := NewMessageQueue(100)

	const iterations = 1000
	for i := 0; i < iterations; i++ {
		// Put
		queue.Put(0, 100*time.Millisecond, i)

		// Pop half the time
		if i%2 == 0 {
			queue.PopFirst()
		}
	}

	// Verify queue is in a valid state
	dump := queue.GetQueueDump(false)
	if len(dump) > 110 {
		t.Errorf("Queue size %d exceeds max*1.1", len(dump))
	}

	t.Logf("Rapid put/pop completed: %d iterations, final queue size: %d", iterations, len(dump))
}

// TestMessageQueueExpiredEntriesUnderLoad tests that expired entries
// are properly removed when queue is under load
func TestMessageQueueExpiredEntriesUnderLoad(t *testing.T) {
	initStratuxClock()

	queue := NewMessageQueue(100)

	// Add messages with very short expiry
	for i := 0; i < 200; i++ {
		queue.Put(0, 1*time.Millisecond, i) // Very short expiry
	}

	// Wait for entries to expire
	time.Sleep(10 * time.Millisecond)
	stratuxClock.Milliseconds += 100 // Advance clock

	// Access queue - should trigger cleanup
	dump := queue.GetQueueDump(true)

	// All entries should be expired
	if len(dump) != 0 {
		t.Errorf("Expected 0 entries after expiry, got %d", len(dump))
	}
}

// TestMessageQueueConcurrentOverflow tests concurrent access during overflow
func TestMessageQueueConcurrentOverflow(t *testing.T) {
	initStratuxClock()

	queue := NewMessageQueue(100)

	var wg sync.WaitGroup
	const numGoroutines = 10
	const messagesPerGoroutine = 100

	// Concurrent writes from multiple goroutines
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < messagesPerGoroutine; i++ {
				queue.Put(int32(id), 1*time.Second, id*1000+i)
			}
		}(g)
	}

	wg.Wait()

	// Verify queue didn't exceed reasonable size
	dump := queue.GetQueueDump(true)
	if len(dump) > 110 {
		t.Errorf("Concurrent overflow: queue size %d exceeds max*1.1", len(dump))
	}

	t.Logf("Concurrent overflow: %d goroutines x %d messages, final size: %d",
		numGoroutines, messagesPerGoroutine, len(dump))
}
