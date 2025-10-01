// -----------------------------------------------------------------------
// Last Modified: Wednesday, 1st October 2025 4:30:00 pm
// Modified By: Bob McAllan
// -----------------------------------------------------------------------

package writers

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

func TestMemoryWriter_Basic(t *testing.T) {
	config := models.WriterConfiguration{
		Type:  models.LogWriterTypeMemory,
		Level: levels.LogLevel(log.TraceLevel),
	}

	memWriter := MemoryWriter(config)
	if memWriter == nil {
		t.Fatal("MemoryWriter should not return nil")
	}
	defer memWriter.Close()

	// Get the store and create a store writer
	store := memWriter.GetStore()
	storeWriter := LogStoreWriter(store, config)

	correlationID := "test-123"

	logEvent := models.LogEvent{
		Level:         log.InfoLevel,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Message:       "test message",
		Prefix:        "TEST",
	}

	jsonData, _ := json.Marshal(logEvent)
	storeWriter.Write(jsonData)

	// Give async processing time to complete
	time.Sleep(50 * time.Millisecond)

	// Retrieve entries
	entries, err := memWriter.GetEntries(correlationID)
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}

	if len(entries) == 0 {
		t.Error("Expected at least one entry")
	}
}

func TestMemoryWriter_MultipleEntries(t *testing.T) {
	config := models.WriterConfiguration{
		Type:  models.LogWriterTypeMemory,
		Level: levels.LogLevel(log.TraceLevel),
	}

	memWriter := MemoryWriter(config)
	defer memWriter.Close()

	store := memWriter.GetStore()
	storeWriter := LogStoreWriter(store, config)

	correlationID := "test-multi"

	// Write multiple entries
	for i := 0; i < 5; i++ {
		logEvent := models.LogEvent{
			Level:         log.InfoLevel,
			Timestamp:     time.Now(),
			CorrelationID: correlationID,
			Message:       "test message",
		}

		jsonData, _ := json.Marshal(logEvent)
		storeWriter.Write(jsonData)
	}

	// Give async processing time
	time.Sleep(100 * time.Millisecond)

	entries, err := memWriter.GetEntries(correlationID)
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}

	if len(entries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(entries))
	}
}

func TestMemoryWriter_LevelFiltering(t *testing.T) {
	config := models.WriterConfiguration{
		Type:  models.LogWriterTypeMemory,
		Level: levels.LogLevel(log.TraceLevel),
	}

	memWriter := MemoryWriter(config)
	defer memWriter.Close()

	store := memWriter.GetStore()
	storeWriter := LogStoreWriter(store, config)

	correlationID := "test-level"

	// Write entries at different levels
	levels := []log.Level{log.DebugLevel, log.InfoLevel, log.WarnLevel, log.ErrorLevel}
	for _, level := range levels {
		logEvent := models.LogEvent{
			Level:         level,
			Timestamp:     time.Now(),
			CorrelationID: correlationID,
			Message:       "test message",
		}

		jsonData, _ := json.Marshal(logEvent)
		storeWriter.Write(jsonData)
	}

	time.Sleep(100 * time.Millisecond)

	// Get entries with minimum level of Warn
	entries, err := memWriter.GetEntriesWithLevel(correlationID, log.WarnLevel)
	if err != nil {
		t.Errorf("GetEntriesWithLevel should not return error: %v", err)
	}

	// Should only get Warn and Error (2 entries)
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries with Warn level or higher, got %d", len(entries))
	}
}

func TestMemoryWriter_GetEntriesSince(t *testing.T) {
	config := models.WriterConfiguration{
		Type:  models.LogWriterTypeMemory,
		Level: levels.LogLevel(log.TraceLevel),
	}

	memWriter := MemoryWriter(config)
	defer memWriter.Close()

	store := memWriter.GetStore()
	storeWriter := LogStoreWriter(store, config)

	correlationID := "test-since"
	startTime := time.Now()

	// Write entry before timestamp
	logEvent1 := models.LogEvent{
		Level:         log.InfoLevel,
		Timestamp:     startTime.Add(-1 * time.Second),
		CorrelationID: correlationID,
		Message:       "old message",
	}
	jsonData1, _ := json.Marshal(logEvent1)
	storeWriter.Write(jsonData1)

	time.Sleep(50 * time.Millisecond)

	// Mark the "since" timestamp
	sinceTime := time.Now()
	time.Sleep(50 * time.Millisecond)

	// Write entry after timestamp
	logEvent2 := models.LogEvent{
		Level:         log.InfoLevel,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Message:       "new message",
	}
	jsonData2, _ := json.Marshal(logEvent2)
	storeWriter.Write(jsonData2)

	time.Sleep(100 * time.Millisecond)

	// Get entries since timestamp
	entries, err := memWriter.GetEntriesSince(sinceTime)
	if err != nil {
		t.Errorf("GetEntriesSince should not return error: %v", err)
	}

	// Should only get the new message
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry since timestamp, got %d", len(entries))
	}

	if len(entries) > 0 && entries[0].Message != "new message" {
		t.Errorf("Expected 'new message', got '%s'", entries[0].Message)
	}
}

func TestMemoryWriter_GetRecent(t *testing.T) {
	config := models.WriterConfiguration{
		Type:  models.LogWriterTypeMemory,
		Level: levels.LogLevel(log.TraceLevel),
	}

	memWriter := MemoryWriter(config)
	defer memWriter.Close()

	store := memWriter.GetStore()
	storeWriter := LogStoreWriter(store, config)

	// Write 10 entries
	for i := 0; i < 10; i++ {
		logEvent := models.LogEvent{
			Level:         log.InfoLevel,
			Timestamp:     time.Now(),
			CorrelationID: "test-recent",
			Message:       "test message",
		}

		jsonData, _ := json.Marshal(logEvent)
		storeWriter.Write(jsonData)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	time.Sleep(100 * time.Millisecond)

	// Get only 5 most recent
	entries, err := memWriter.GetEntriesWithLimit(5)
	if err != nil {
		t.Errorf("GetEntriesWithLimit should not return error: %v", err)
	}

	if len(entries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(entries))
	}
}

func TestMemoryWriter_WithPersistence(t *testing.T) {
	config := models.WriterConfiguration{
		Type:   models.LogWriterTypeMemory,
		Level:  levels.LogLevel(log.TraceLevel),
		DBPath: "temp/test_logs",
	}

	memWriter := MemoryWriter(config)
	defer memWriter.Close()

	store := memWriter.GetStore()
	storeWriter := LogStoreWriter(store, config)

	correlationID := "test-persist"

	logEvent := models.LogEvent{
		Level:         log.InfoLevel,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Message:       "persistent message",
	}

	jsonData, _ := json.Marshal(logEvent)
	storeWriter.Write(jsonData)

	time.Sleep(100 * time.Millisecond)

	entries, err := memWriter.GetEntries(correlationID)
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}

	if len(entries) == 0 {
		t.Error("Expected at least one entry")
	}
}

func TestWebSocketWriter_Basic(t *testing.T) {
	config := models.WriterConfiguration{
		Type:  models.LogWriterTypeMemory,
		Level: levels.LogLevel(log.TraceLevel),
	}

	// Create store
	store, err := NewInMemoryLogStore(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create WebSocket writer with faster poll interval for testing
	wsWriter := WebSocketWriter(store, config, 50*time.Millisecond).(*websocketWriter)
	defer wsWriter.Close()

	// Add a test client
	received := make([]models.LogEvent, 0)
	var receivedMux sync.Mutex
	receivedChan := make(chan bool, 1)

	client := NewSimpleWebSocketClient(
		func(data interface{}) error {
			if logs, ok := data.([]models.LogEvent); ok {
				receivedMux.Lock()
				received = append(received, logs...)
				receivedMux.Unlock()
				select {
				case receivedChan <- true:
				default:
				}
			}
			return nil
		},
		func() error {
			return nil
		},
	)

	// Add client first
	wsWriter.AddClient("test-client", client)

	// Give the polling goroutine time to start
	time.Sleep(20 * time.Millisecond)

	// Write log to store
	logEvent := models.LogEvent{
		Level:         log.InfoLevel,
		Timestamp:     time.Now(),
		CorrelationID: "ws-test",
		Message:       "websocket test message",
	}

	store.Store(logEvent)

	// Wait for broadcast (should happen within 2 poll intervals)
	select {
	case <-receivedChan:
		// Success
	case <-time.After(200 * time.Millisecond):
		t.Error("Timeout waiting for WebSocket broadcast")
	}

	receivedMux.Lock()
	receivedCount := len(received)
	receivedMux.Unlock()

	if receivedCount == 0 {
		t.Error("Expected to receive log entries via WebSocket")
	}
}
