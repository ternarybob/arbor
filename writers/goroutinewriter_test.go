package writers

import (
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

// ===========================
// HELPER FUNCTIONS
// ===========================

// createTestLogEvent creates a test LogEvent with common defaults
func createTestLogEvent(level log.Level, correlationID string, message string) models.LogEvent {
	return models.LogEvent{
		Level:         level,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Message:       message,
		Prefix:        "TEST",
		Function:      "test",
		Fields:        make(map[string]interface{}),
	}
}

// marshalLogEvent marshals a LogEvent to JSON for Write() calls
func marshalLogEvent(t *testing.T, event models.LogEvent) []byte {
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal test log event: %v", err)
	}
	return data
}

// createCountingProcessor creates a processor that increments an atomic counter
func createCountingProcessor(counter *atomic.Int64) func(models.LogEvent) error {
	return func(entry models.LogEvent) error {
		counter.Add(1)
		return nil
	}
}

// createDelayedProcessor creates a processor that increments counter and sleeps
func createDelayedProcessor(counter *atomic.Int64, delay time.Duration) func(models.LogEvent) error {
	return func(entry models.LogEvent) error {
		counter.Add(1)
		time.Sleep(delay)
		return nil
	}
}

// createCollectingProcessor creates a processor that collects entries in a slice
func createCollectingProcessor(collected *[]models.LogEvent, mu *sync.Mutex) func(models.LogEvent) error {
	return func(entry models.LogEvent) error {
		mu.Lock()
		*collected = append(*collected, entry)
		mu.Unlock()
		return nil
	}
}

// ===========================
// SECTION 1: UNIT TESTS FOR goroutineWriter BASE
// ===========================

func TestGoroutineWriter_NewWithValidProcessor(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	processor := func(models.LogEvent) error { return nil }

	writer, err := NewGoroutineWriter(config, 1000, processor)

	if writer == nil {
		t.Fatal("Expected writer to be non-nil")
	}
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify writer implements IGoroutineWriter
	_, ok := writer.(IGoroutineWriter)
	if !ok {
		t.Error("Expected writer to implement IGoroutineWriter interface")
	}

	// Verify not started yet
	if writer.IsRunning() {
		t.Error("Expected IsRunning() to be false before Start()")
	}
}

func TestGoroutineWriter_NewWithNilProcessor(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.InfoLevel}

	writer, err := NewGoroutineWriter(config, 1000, nil)

	if writer != nil {
		t.Error("Expected writer to be nil when processor is nil")
	}
	if err == nil {
		t.Fatal("Expected error when processor is nil")
	}
}

func TestGoroutineWriter_NewWithInvalidBufferSize(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	processor := func(models.LogEvent) error { return nil }

	// Test with zero buffer size
	writer, err := NewGoroutineWriter(config, 0, processor)
	if writer == nil || err != nil {
		t.Error("Expected writer creation to succeed with zero buffer size (should default to 1000)")
	}

	// Test with negative buffer size
	writer, err = NewGoroutineWriter(config, -100, processor)
	if writer == nil || err != nil {
		t.Error("Expected writer creation to succeed with negative buffer size (should default to 1000)")
	}
}

func TestGoroutineWriter_StartStop_Lifecycle(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	processor := func(models.LogEvent) error { return nil }

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	// Initially not running
	if writer.IsRunning() {
		t.Error("Expected IsRunning() to be false initially")
	}

	// Start
	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}
	if !writer.IsRunning() {
		t.Error("Expected IsRunning() to be true after Start()")
	}

	// Double-start should fail
	if err := writer.Start(); err == nil {
		t.Error("Expected error when starting already running writer")
	}

	// Stop
	if err := writer.Stop(); err != nil {
		t.Fatalf("Failed to stop writer: %v", err)
	}
	if writer.IsRunning() {
		t.Error("Expected IsRunning() to be false after Stop()")
	}

	// Double-stop should be idempotent
	if err := writer.Stop(); err != nil {
		t.Error("Expected Stop() to be idempotent")
	}
}

func TestGoroutineWriter_StartStopRestart(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	var counter atomic.Int64
	processor := createCountingProcessor(&counter)

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	// Start
	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	if !writer.IsRunning() {
		t.Error("Expected running after first Start()")
	}

	// Stop
	if err := writer.Stop(); err != nil {
		t.Fatalf("Failed to stop: %v", err)
	}
	if writer.IsRunning() {
		t.Error("Expected not running after Stop()")
	}

	// Restart
	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to restart: %v", err)
	}
	if !writer.IsRunning() {
		t.Error("Expected running after restart")
	}

	// Verify goroutine is processing
	event := createTestLogEvent(log.InfoLevel, "restart-test", "test message")
	data := marshalLogEvent(t, event)
	if _, err := writer.Write(data); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	time.Sleep(150 * time.Millisecond)
	if counter.Load() != 1 {
		t.Errorf("Expected 1 processed entry, got %d", counter.Load())
	}

	if err := writer.Stop(); err != nil {
		t.Errorf("Failed to stop writer: %v", err)
	}
}

func TestGoroutineWriter_Write_BeforeStart(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	var counter atomic.Int64
	processor := createCountingProcessor(&counter)

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	// Do NOT call Start()
	event := createTestLogEvent(log.InfoLevel, "before-start", "test message")
	data := marshalLogEvent(t, event)

	n, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Expected Write() to succeed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	time.Sleep(150 * time.Millisecond)
	if counter.Load() != 0 {
		t.Errorf("Expected processor not called before Start(), got %d calls", counter.Load())
	}
}

func TestGoroutineWriter_Write_AfterStop(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	var counter atomic.Int64
	processor := createCountingProcessor(&counter)

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}
	if err := writer.Stop(); err != nil {
		t.Fatalf("Failed to stop writer: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	event := createTestLogEvent(log.InfoLevel, "after-stop", "test message")
	data := marshalLogEvent(t, event)

	n, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Expected Write() to succeed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	time.Sleep(150 * time.Millisecond)
	if counter.Load() != 0 {
		t.Errorf("Expected processor not called after Stop(), got %d calls", counter.Load())
	}
}

func TestGoroutineWriter_Write_Success(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	var counter atomic.Int64
	processor := createCountingProcessor(&counter)

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	event := createTestLogEvent(log.InfoLevel, "write-success", "test message")
	data := marshalLogEvent(t, event)

	n, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	time.Sleep(200 * time.Millisecond)
	if counter.Load() != 1 {
		t.Errorf("Expected 1 processed entry, got %d", counter.Load())
	}
}

func TestGoroutineWriter_Write_InvalidJSON(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	processor := func(models.LogEvent) error { return nil }

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	n, err := writer.Write([]byte("invalid json {{"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes written, got %d", n)
	}
}

func TestGoroutineWriter_Write_EmptyData(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	processor := func(models.LogEvent) error { return nil }

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	// Empty slice
	n, err := writer.Write([]byte(""))
	if err != nil {
		t.Errorf("Expected no error for empty data: %v", err)
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes written, got %d", n)
	}

	// Nil slice
	n, err = writer.Write(nil)
	if err != nil {
		t.Errorf("Expected no error for nil data: %v", err)
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes written for nil, got %d", n)
	}
}

func TestGoroutineWriter_LevelFiltering(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.WarnLevel}
	var counter atomic.Int64
	processor := createCountingProcessor(&counter)

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	// Write entries at different levels
	entries := []log.Level{log.DebugLevel, log.InfoLevel, log.WarnLevel, log.ErrorLevel}
	for i, level := range entries {
		event := createTestLogEvent(level, "level-test", "message")
		data := marshalLogEvent(t, event)
		if _, err := writer.Write(data); err != nil {
			t.Fatalf("Failed to write entry %d: %v", i, err)
		}
	}

	time.Sleep(300 * time.Millisecond)

	// Only Warn and Error should be processed
	if counter.Load() != 2 {
		t.Errorf("Expected 2 processed entries (Warn+Error), got %d", counter.Load())
	}
}

func TestGoroutineWriter_WithLevel_DynamicChange(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.InfoLevel}
	var counter atomic.Int64
	processor := createCountingProcessor(&counter)

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	// Write Debug entry (should be filtered)
	event := createTestLogEvent(log.DebugLevel, "dynamic-level", "filtered")
	data := marshalLogEvent(t, event)
	if _, err := writer.Write(data); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	time.Sleep(150 * time.Millisecond)
	if counter.Load() != 0 {
		t.Error("Expected Debug entry to be filtered initially")
	}

	// Lower threshold to Debug
	writer.WithLevel(log.DebugLevel)

	// Write another Debug entry (should now be processed)
	event = createTestLogEvent(log.DebugLevel, "dynamic-level", "processed")
	data = marshalLogEvent(t, event)
	if _, err := writer.Write(data); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	time.Sleep(150 * time.Millisecond)
	if counter.Load() != 1 {
		t.Errorf("Expected 1 processed entry after lowering level, got %d", counter.Load())
	}
}

func TestGoroutineWriter_BufferOverflow(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	var counter atomic.Int64
	// Slow processor (100ms per entry)
	processor := createDelayedProcessor(&counter, 100*time.Millisecond)

	writer, err := NewGoroutineWriter(config, 10, processor) // Small buffer
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	// Write 60 entries rapidly
	for i := 0; i < 60; i++ {
		event := createTestLogEvent(log.InfoLevel, "overflow", "message")
		data := marshalLogEvent(t, event)
		n, err := writer.Write(data)
		if err != nil {
			// Expected - buffer may be full, writes may fail
		} else {
			// If write succeeded, verify correct byte count
			if n != len(data) {
				t.Errorf("Write returned n=%d, expected %d", n, len(data))
			}
		}
	}

	time.Sleep(3000 * time.Millisecond)

	// Not all entries should be processed (some dropped)
	processed := counter.Load()
	if processed >= 60 {
		t.Errorf("Expected some entries to be dropped, but all %d were processed", processed)
	}
	t.Logf("Processed %d out of 60 entries (overflow test)", processed)
}

func TestGoroutineWriter_GracefulShutdown_BufferDraining(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	var counter atomic.Int64
	processor := createDelayedProcessor(&counter, 10*time.Millisecond)

	writer, err := NewGoroutineWriter(config, 100, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	// Write 50 entries rapidly
	for i := 0; i < 50; i++ {
		event := createTestLogEvent(log.InfoLevel, "shutdown", "message")
		data := marshalLogEvent(t, event)
		if _, err := writer.Write(data); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	// Immediately stop (buffer should drain)
	if err := writer.Stop(); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	if counter.Load() != 50 {
		t.Errorf("Expected all 50 entries processed during shutdown, got %d", counter.Load())
	}

	if writer.IsRunning() {
		t.Error("Expected IsRunning() to be false after Stop()")
	}
}

func TestGoroutineWriter_ConcurrentWrites(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	var counter atomic.Int64
	processor := createCountingProcessor(&counter)

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	// Launch 10 goroutines, each writing 10 entries
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				event := createTestLogEvent(log.InfoLevel, "concurrent", "message")
				data := marshalLogEvent(t, event)
				if _, err := writer.Write(data); err != nil {
					t.Errorf("Write failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(600 * time.Millisecond)

	if counter.Load() != 100 {
		t.Errorf("Expected 100 processed entries, got %d", counter.Load())
	}
}

func TestGoroutineWriter_ProcessorError(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	processor := func(models.LogEvent) error {
		return errors.New("processor error")
	}

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	// Write 5 entries (all will cause processor errors)
	for i := 0; i < 5; i++ {
		event := createTestLogEvent(log.InfoLevel, "error-test", "message")
		data := marshalLogEvent(t, event)
		if _, err := writer.Write(data); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	time.Sleep(200 * time.Millisecond)

	// Goroutine should still be running
	if !writer.IsRunning() {
		t.Error("Expected goroutine to continue running after processor errors")
	}

	// Should be able to write more entries
	event := createTestLogEvent(log.InfoLevel, "after-error", "message")
	data := marshalLogEvent(t, event)
	n, err := writer.Write(data)
	if err != nil {
		t.Errorf("Expected Write() to succeed after processor errors: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}
}

func TestGoroutineWriter_Close_Idempotent(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	processor := func(models.LogEvent) error { return nil }

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Errorf("First Close() failed: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Errorf("Second Close() failed (should be idempotent): %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Errorf("Third Close() failed (should be idempotent): %v", err)
	}
}

func TestGoroutineWriter_GetFilePath(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	processor := func(models.LogEvent) error { return nil }

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	path := writer.GetFilePath()
	if path != "" {
		t.Errorf("Expected empty string, got %q", path)
	}
}

func TestGoroutineWriter_MultipleEntries_OrderPreserved(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	var collected []models.LogEvent
	var mu sync.Mutex
	processor := createCollectingProcessor(&collected, &mu)

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	// Write 10 entries with numbered messages
	for i := 0; i < 10; i++ {
		event := createTestLogEvent(log.InfoLevel, "order-test", "msg-"+string(rune('0'+i)))
		data := marshalLogEvent(t, event)
		if _, err := writer.Write(data); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	time.Sleep(400 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(collected) != 10 {
		t.Fatalf("Expected 10 entries, got %d", len(collected))
	}

	// Verify order
	for i, entry := range collected {
		expected := "msg-" + string(rune('0'+i))
		if entry.Message != expected {
			t.Errorf("Entry %d: expected message %q, got %q", i, expected, entry.Message)
		}
	}
}

// ===========================
// SECTION 2: INTEGRATION TESTS WITH REFACTORED WRITERS
// ===========================

func TestLogStoreWriter_Integration_WithGoroutineWriter(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	store, err := NewInMemoryLogStore(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	writer := LogStoreWriter(store, config)
	if writer == nil {
		t.Fatal("Expected writer to be non-nil")
	}
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("Failed to close writer: %v", err)
		}
	}()

	event := createTestLogEvent(log.InfoLevel, "integration-test", "test message")
	data := marshalLogEvent(t, event)

	n, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	time.Sleep(300 * time.Millisecond)

	entries, err := store.GetByCorrelation("integration-test")
	if err != nil {
		t.Fatalf("GetByCorrelation failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Message != "test message" {
		t.Errorf("Expected message 'test message', got %q", entries[0].Message)
	}
}

func TestLogStoreWriter_Integration_LevelFiltering(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.WarnLevel}
	store, err := NewInMemoryLogStore(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	writer := LogStoreWriter(store, config)
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("Failed to close writer: %v", err)
		}
	}()

	// Write entries at different levels
	levels := []log.Level{log.DebugLevel, log.InfoLevel, log.WarnLevel, log.ErrorLevel}
	for _, level := range levels {
		event := createTestLogEvent(level, "level-filter", "message")
		data := marshalLogEvent(t, event)
		n, err := writer.Write(data)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected %d bytes written, got %d", len(data), n)
		}
	}

	time.Sleep(300 * time.Millisecond)

	entries, err := store.GetByCorrelation("level-filter")
	if err != nil {
		t.Fatalf("GetByCorrelation failed: %v", err)
	}

	// Only Warn and Error should be stored
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries (Warn+Error), got %d", len(entries))
	}
}

func TestLogStoreWriter_Integration_GracefulShutdown(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	store, err := NewInMemoryLogStore(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	writer := LogStoreWriter(store, config)

	// Write 20 entries rapidly
	for i := 0; i < 20; i++ {
		event := createTestLogEvent(log.InfoLevel, "shutdown-test", "message")
		data := marshalLogEvent(t, event)
		n, err := writer.Write(data)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected %d bytes written, got %d", len(data), n)
		}
	}

	// Immediately close
	if err := writer.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	entries, err := store.GetByCorrelation("shutdown-test")
	if err != nil {
		t.Fatalf("GetByCorrelation failed: %v", err)
	}

	if len(entries) != 20 {
		t.Errorf("Expected all 20 entries, got %d", len(entries))
	}
}

func TestContextWriter_Integration_WithGoroutineWriter(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}

	writer := NewContextWriter(config)
	if writer == nil {
		t.Fatal("Expected writer to be non-nil")
	}
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("Failed to close writer: %v", err)
		}
	}()

	event := createTestLogEvent(log.InfoLevel, "context-test", "test message")
	data := marshalLogEvent(t, event)

	n, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	time.Sleep(200 * time.Millisecond)
	// Note: Cannot easily verify common.Log() was called without accessing global context buffer
}

func TestContextWriter_Integration_AsyncBehavior(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	writer := NewContextWriter(config)
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("Failed to close writer: %v", err)
		}
	}()

	start := time.Now()

	// Write 10 entries rapidly
	for i := 0; i < 10; i++ {
		event := createTestLogEvent(log.InfoLevel, "async-test", "message")
		data := marshalLogEvent(t, event)
		if _, err := writer.Write(data); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	elapsed := time.Since(start)

	// All writes should complete quickly (non-blocking)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected writes to be non-blocking (<100ms), took %v", elapsed)
	}

	time.Sleep(200 * time.Millisecond)
}

// ===========================
// SECTION 3: BACKWARD COMPATIBILITY TESTS
// ===========================

func TestLogStoreWriter_BackwardCompatibility_ConstructorSignature(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	store, err := NewInMemoryLogStore(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	writer := LogStoreWriter(store, config)
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("Failed to close writer: %v", err)
		}
	}()

	// Verify returns IWriter
	_, ok := writer.(IWriter)
	if !ok {
		t.Error("Expected writer to implement IWriter interface")
	}

	// Verify can call all IWriter methods
	writer.WithLevel(log.InfoLevel)
	path := writer.GetFilePath()
	if path != "" {
		t.Errorf("Expected empty file path, got %q", path)
	}

	event := createTestLogEvent(log.InfoLevel, "compat-test", "message")
	data := marshalLogEvent(t, event)
	if _, err := writer.Write(data); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestLogStoreWriter_BackwardCompatibility_AutoStart(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	store, err := NewInMemoryLogStore(config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	writer := LogStoreWriter(store, config)
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("Failed to close writer: %v", err)
		}
	}()

	// Immediately write entry (no explicit Start() call)
	event := createTestLogEvent(log.InfoLevel, "autostart-test", "message")
	data := marshalLogEvent(t, event)
	if _, err := writer.Write(data); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	entries, err := store.GetByCorrelation("autostart-test")
	if err != nil {
		t.Fatalf("GetByCorrelation failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry (proves auto-start), got %d", len(entries))
	}
}

func TestContextWriter_BackwardCompatibility_ConstructorSignature(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}

	writer := NewContextWriter(config)
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("Failed to close writer: %v", err)
		}
	}()

	// Verify returns IWriter
	_, ok := writer.(IWriter)
	if !ok {
		t.Error("Expected writer to implement IWriter interface")
	}

	// Verify can call all IWriter methods
	writer.WithLevel(log.InfoLevel)
	path := writer.GetFilePath()
	if path != "" {
		t.Errorf("Expected empty file path, got %q", path)
	}

	event := createTestLogEvent(log.InfoLevel, "context-compat", "message")
	data := marshalLogEvent(t, event)
	n, err := writer.Write(data)
	if err != nil || n != len(data) {
		t.Errorf("Write failed: n=%d, err=%v", n, err)
	}

	if err := writer.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

// ===========================
// SECTION 4: EDGE CASE AND CONCURRENCY TESTS
// ===========================

func TestGoroutineWriter_ConcurrentStartStop(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	processor := func(models.LogEvent) error { return nil }

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	var wg sync.WaitGroup

	// Launch 5 goroutines calling Start()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			writer.Start()
		}()
	}

	// Launch 5 goroutines calling Stop()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			writer.Stop()
		}()
	}

	wg.Wait()

	// Verify no panics or deadlocks occurred
	// Final state should be consistent
	running := writer.IsRunning()
	t.Logf("Final state: running=%v", running)

	// Ensure cleanup
	if err := writer.Stop(); err != nil {
		t.Errorf("Failed to stop writer: %v", err)
	}
}

func TestGoroutineWriter_ConcurrentWithLevel(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.InfoLevel}
	var counter atomic.Int64
	processor := createCountingProcessor(&counter)

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	var wg sync.WaitGroup

	// Goroutine writing Debug entries
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			event := createTestLogEvent(log.DebugLevel, "concurrent-level", "debug")
			data := marshalLogEvent(t, event)
			if _, err := writer.Write(data); err != nil {
				t.Errorf("Write failed: %v", err)
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Goroutine changing level to Debug after 50ms
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		writer.WithLevel(log.DebugLevel)
	}()

	wg.Wait()
	time.Sleep(400 * time.Millisecond)

	// Some Debug entries should be filtered, some processed
	processed := counter.Load()
	if processed == 0 {
		t.Error("Expected some Debug entries to be processed after WithLevel()")
	}
	if processed >= 100 {
		t.Error("Expected some Debug entries to be filtered before WithLevel()")
	}

	t.Logf("Processed %d out of 100 Debug entries (concurrent WithLevel test)", processed)
}

func TestGoroutineWriter_ProcessorError_ContinuesProcessing(t *testing.T) {
	config := models.WriterConfiguration{Level: levels.TraceLevel}
	var counter atomic.Int64

	// Processor returns error for first 3 entries, then succeeds
	processor := func(entry models.LogEvent) error {
		count := counter.Add(1)
		if count <= 3 {
			return errors.New("first 3 entries error")
		}
		return nil
	}

	writer, err := NewGoroutineWriter(config, 1000, processor)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() {
		if err := writer.Stop(); err != nil {
			t.Errorf("Failed to stop writer: %v", err)
		}
	}()

	if err := writer.Start(); err != nil {
		t.Fatalf("Failed to start writer: %v", err)
	}

	// Write 5 entries
	for i := 0; i < 5; i++ {
		event := createTestLogEvent(log.InfoLevel, "error-continue", "message")
		data := marshalLogEvent(t, event)
		if _, err := writer.Write(data); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	time.Sleep(400 * time.Millisecond)

	// All 5 entries should be attempted
	if counter.Load() != 5 {
		t.Errorf("Expected 5 attempts (errors logged but processing continued), got %d", counter.Load())
	}

	// Goroutine should still be running
	if !writer.IsRunning() {
		t.Error("Expected goroutine to continue running after processor errors")
	}
}
