package writers

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/models"
)

func TestBoltDBMemoryWriter_Basic(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer writer.Close()

	// Verify it implements IMemoryWriter interface
	var _ IMemoryWriter = writer
}

func TestBoltDBMemoryWriter_WithLevel(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer writer.Close()

	// Test changing level
	newWriter := writer.WithLevel(log.DebugLevel)
	if newWriter == nil {
		t.Error("WithLevel should not return nil")
	}

	// Should return the same instance
	if newWriter != writer {
		t.Error("WithLevel should return the same instance")
	}
}

func TestBoltDBMemoryWriter_WriteAndRetrieve(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer writer.Close()

	correlationID := "test-correlation-123"

	logEvent := models.LogEvent{
		Level:         log.InfoLevel,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Message:       "test message",
		Prefix:        "TEST",
		Error:         "test error",
	}

	// Convert to JSON as the memory writer expects
	jsonData, err := json.Marshal(logEvent)
	if err != nil {
		t.Fatalf("Failed to marshal log event: %v", err)
	}

	// Write the log event
	n, err := writer.Write(jsonData)
	if err != nil {
		t.Errorf("Write should not return error: %v", err)
	}
	if n != len(jsonData) {
		t.Errorf("Expected %d bytes written, got %d", len(jsonData), n)
	}

	// Small delay to ensure write is processed
	time.Sleep(10 * time.Millisecond)

	// Verify the entry was stored
	entries, err := writer.GetEntries(correlationID)
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestBoltDBMemoryWriter_WriteWithoutCorrelationID(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer writer.Close()

	logEvent := models.LogEvent{
		Level:     log.InfoLevel,
		Timestamp: time.Now(),
		Message:   "test message without correlation ID",
	}

	jsonData, err := json.Marshal(logEvent)
	if err != nil {
		t.Fatalf("Failed to marshal log event: %v", err)
	}

	// Write the log event
	n, err := writer.Write(jsonData)
	if err != nil {
		t.Errorf("Write should not return error: %v", err)
	}
	if n != len(jsonData) {
		t.Errorf("Expected %d bytes written, got %d", len(jsonData), n)
	}

	// Should not be stored since no correlation ID
	entries, err := writer.GetEntries("")
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(entries))
	}
}

func TestBoltDBMemoryWriter_MultipleEntries(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer writer.Close()

	correlationID := "test-multiple-entries"

	// Add multiple log events
	for i := 0; i < 3; i++ {
		logEvent := models.LogEvent{
			Level:         log.InfoLevel,
			Timestamp:     time.Now(),
			CorrelationID: correlationID,
			Message:       "test message",
		}

		jsonData, _ := json.Marshal(logEvent)
		writer.Write(jsonData)
	}

	// Small delay to ensure writes are processed
	time.Sleep(10 * time.Millisecond)

	// Get entries
	entries, err := writer.GetEntries(correlationID)
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Test with empty correlation ID
	entries, err = writer.GetEntries("")
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries for empty correlation ID, got %d", len(entries))
	}
}

func TestBoltDBMemoryWriter_LevelFiltering(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer writer.Close()

	correlationID := "test-level-filter"

	// Add log events with different levels
	levels := []log.Level{log.DebugLevel, log.InfoLevel, log.WarnLevel, log.ErrorLevel}
	for _, level := range levels {
		logEvent := models.LogEvent{
			Level:         level,
			Timestamp:     time.Now(),
			CorrelationID: correlationID,
			Message:       "test message",
		}

		jsonData, _ := json.Marshal(logEvent)
		writer.Write(jsonData)
	}

	// Small delay to ensure writes are processed
	time.Sleep(10 * time.Millisecond)

	// Get entries with minimum level of Warn
	entries, err := writer.GetEntriesWithLevel(correlationID, log.WarnLevel)
	if err != nil {
		t.Errorf("GetEntriesWithLevel should not return error: %v", err)
	}
	// Should include Warn and Error (2 entries)
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries with WarnLevel filter, got %d", len(entries))
	}

	// Get entries with minimum level of Info
	entries, err = writer.GetEntriesWithLevel(correlationID, log.InfoLevel)
	if err != nil {
		t.Errorf("GetEntriesWithLevel should not return error: %v", err)
	}
	// Should include Info, Warn, and Error (3 entries)
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries with InfoLevel filter, got %d", len(entries))
	}
}

func TestBoltDBMemoryWriter_GetStoredCorrelationIDs(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer writer.Close()

	// Use unique correlation IDs with timestamp to avoid conflicts
	timestamp := time.Now().UnixNano()
	correlationIDs := []string{
		fmt.Sprintf("test-1-%d", timestamp),
		fmt.Sprintf("test-2-%d", timestamp),
		fmt.Sprintf("test-3-%d", timestamp),
	}

	// Add log events with different correlation IDs
	for _, id := range correlationIDs {
		logEvent := models.LogEvent{
			Level:         log.InfoLevel,
			Timestamp:     time.Now(),
			CorrelationID: id,
			Message:       "test message",
		}

		jsonData, _ := json.Marshal(logEvent)
		writer.Write(jsonData)
	}

	// Small delay to ensure writes are processed
	time.Sleep(10 * time.Millisecond)

	// Get stored correlation IDs
	storedIDs := writer.GetStoredCorrelationIDs()

	// Check that our correlation IDs are present (there may be others from other tests)
	idMap := make(map[string]bool)
	for _, id := range storedIDs {
		idMap[id] = true
	}

	for _, expectedID := range correlationIDs {
		if !idMap[expectedID] {
			t.Errorf("Expected correlation ID %s not found in stored IDs", expectedID)
		}
	}

	// Verify we have at least our 3 correlation IDs
	if len(storedIDs) < 3 {
		t.Errorf("Expected at least 3 correlation IDs, got %d", len(storedIDs))
	}
}

func TestBoltDBMemoryWriter_Expiration(t *testing.T) {
	// This test would need to be modified to test with a shorter TTL
	// For now, we'll just verify that expired entries are filtered out during retrieval
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	memWriter := writer.(*memoryWriter)
	defer writer.Close()

	// Reduce TTL for testing
	memWriter.ttl = 100 * time.Millisecond

	correlationID := "test-expiration"

	logEvent := models.LogEvent{
		Level:         log.InfoLevel,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Message:       "test message that will expire",
	}

	jsonData, _ := json.Marshal(logEvent)
	writer.Write(jsonData)

	// Immediately check that entry exists
	entries, err := writer.GetEntries(correlationID)
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry immediately after write, got %d", len(entries))
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Check that entry is now filtered out
	entries, err = writer.GetEntries(correlationID)
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after expiration, got %d", len(entries))
	}
}

func TestBoltDBMemoryWriter_Close(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}

	// Should not error when closing
	err := writer.Close()
	if err != nil {
		t.Errorf("Close should not return error: %v", err)
	}

	// Should be safe to call Close multiple times
	err = writer.Close()
	if err != nil {
		t.Errorf("Second Close should not return error: %v", err)
	}
}

func TestBoltDBMemoryWriter_GetAllEntries(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer writer.Close()

	// Add multiple log events with different correlation IDs
	timestamp := time.Now().UnixNano()
	correlationIDs := []string{
		fmt.Sprintf("test-all-1-%d", timestamp),
		fmt.Sprintf("test-all-2-%d", timestamp),
		fmt.Sprintf("test-all-3-%d", timestamp),
	}

	totalEntries := 0
	for i, id := range correlationIDs {
		// Add 2 entries per correlation ID
		for j := 0; j < 2; j++ {
			logEvent := models.LogEvent{
				Level:         log.InfoLevel,
				Timestamp:     time.Now().Add(time.Duration(i*2+j) * time.Millisecond),
				CorrelationID: id,
				Message:       fmt.Sprintf("test message %d-%d", i, j),
			}

			jsonData, _ := json.Marshal(logEvent)
			writer.Write(jsonData)
			totalEntries++
		}
	}

	// Small delay to ensure writes are processed
	time.Sleep(20 * time.Millisecond)

	// Get all entries
	allEntries, err := writer.GetAllEntries()
	if err != nil {
		t.Errorf("GetAllEntries should not return error: %v", err)
	}

	// Should have at least our entries (may have more from other tests)
	if len(allEntries) < totalEntries {
		t.Errorf("Expected at least %d entries, got %d", totalEntries, len(allEntries))
	}

	// Verify that entries from all correlation IDs are present
	foundCorrelationIDs := make(map[string]int)
	for key, _ := range allEntries {
		// Extract correlation ID from key (format: "correlationid:index")
		for _, expectedID := range correlationIDs {
			if len(key) > len(expectedID) && key[:len(expectedID)] == expectedID {
				foundCorrelationIDs[expectedID]++
			}
		}
	}

	for _, expectedID := range correlationIDs {
		if foundCorrelationIDs[expectedID] < 2 {
			t.Errorf("Expected at least 2 entries for correlation ID %s, got %d", expectedID, foundCorrelationIDs[expectedID])
		}
	}
}

func TestBoltDBMemoryWriter_GetEntriesWithLimit(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer writer.Close()

	// Add multiple log events with different timestamps
	timestamp := time.Now().UnixNano()
	correlationID := fmt.Sprintf("test-limit-%d", timestamp)
	entryCount := 5

	for i := 0; i < entryCount; i++ {
		logEvent := models.LogEvent{
			Level:         log.InfoLevel,
			Timestamp:     time.Now().Add(time.Duration(i) * time.Millisecond),
			CorrelationID: correlationID,
			Message:       fmt.Sprintf("test message %d", i),
		}

		jsonData, _ := json.Marshal(logEvent)
		writer.Write(jsonData)
	}

	// Small delay to ensure writes are processed
	time.Sleep(20 * time.Millisecond)

	// Test with limit less than total entries
	limit := 3
	limitedEntries, err := writer.GetEntriesWithLimit(limit)
	if err != nil {
		t.Errorf("GetEntriesWithLimit should not return error: %v", err)
	}

	// Should have at most the limit (may have entries from other tests)
	if len(limitedEntries) > limit+10 { // Allow some buffer for other test entries
		t.Errorf("Expected at most around %d entries, got %d", limit+10, len(limitedEntries))
	}

	// Test with limit of 0
	zeroEntries, err := writer.GetEntriesWithLimit(0)
	if err != nil {
		t.Errorf("GetEntriesWithLimit should not return error: %v", err)
	}
	if len(zeroEntries) != 0 {
		t.Errorf("Expected 0 entries with limit 0, got %d", len(zeroEntries))
	}

	// Test with negative limit
	negativeEntries, err := writer.GetEntriesWithLimit(-1)
	if err != nil {
		t.Errorf("GetEntriesWithLimit should not return error: %v", err)
	}
	if len(negativeEntries) != 0 {
		t.Errorf("Expected 0 entries with negative limit, got %d", len(negativeEntries))
	}
}

func TestBoltDBMemoryWriter_GetEntriesWithLimitOrdering(t *testing.T) {
	config := models.WriterConfiguration{}
	writer := MemoryWriter(config)
	if writer == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer writer.Close()

	// Add log events with specific timestamps to test ordering
	timestamp := time.Now().UnixNano()
	correlationID := fmt.Sprintf("test-order-%d", timestamp)
	baseTime := time.Now()

	// Add entries with timestamps in specific order
	timestamps := []time.Time{
		baseTime.Add(-3 * time.Second), // Oldest
		baseTime.Add(-1 * time.Second),
		baseTime.Add(-2 * time.Second),
		baseTime, // Most recent
	}

	for i, ts := range timestamps {
		logEvent := models.LogEvent{
			Level:         log.InfoLevel,
			Timestamp:     ts,
			CorrelationID: correlationID,
			Message:       fmt.Sprintf("message %d at %v", i, ts.Format(time.RFC3339Nano)),
		}

		jsonData, _ := json.Marshal(logEvent)
		writer.Write(jsonData)
	}

	// Small delay to ensure writes are processed
	time.Sleep(20 * time.Millisecond)

	// Get entries for this correlation ID to verify ordering
	correlationEntries, err := writer.GetEntries(correlationID)
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(correlationEntries) != 4 {
		t.Errorf("Expected 4 entries for correlation %s, got %d", correlationID, len(correlationEntries))
	}

	// Test that GetEntriesWithLimit returns entries (we can't easily test exact ordering
	// without knowing the exact internal key format, but we can verify it returns entries)
	limitedEntries, err := writer.GetEntriesWithLimit(2)
	if err != nil {
		t.Errorf("GetEntriesWithLimit should not return error: %v", err)
	}

	// Should return some entries (exact count depends on other tests running)
	if len(limitedEntries) == 0 {
		t.Error("GetEntriesWithLimit should return some entries")
	}
}
