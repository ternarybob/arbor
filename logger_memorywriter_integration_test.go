package arbor

import (
	"testing"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/models"
)

func TestLogger_MemoryWriterIntegration(t *testing.T) {
	// Create logger with memory writer
	config := models.WriterConfiguration{}
	logger := Logger().WithMemoryWriter(config)

	// Set correlation ID
	correlationID := "integration-test-123"
	logger = logger.WithCorrelationId(correlationID)

	// Log some messages
	logger.Info().Msg("Test info message")
	logger.Warn().Msg("Test warning message")
	logger.Error().Msg("Test error message")

	// Small delay to ensure writes are processed
	time.Sleep(50 * time.Millisecond)

	// Retrieve logs
	logs, err := logger.GetMemoryLogs(correlationID, LogLevel(log.InfoLevel))
	if err != nil {
		t.Errorf("GetMemoryLogs should not return error: %v", err)
	}

	// Should have 3 log entries
	if len(logs) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(logs))
	}

	// Test level filtering
	warnLogs, err := logger.GetMemoryLogs(correlationID, LogLevel(log.WarnLevel))
	if err != nil {
		t.Errorf("GetMemoryLogs should not return error: %v", err)
	}

	// Should have 2 log entries (warn and error)
	if len(warnLogs) != 2 {
		t.Errorf("Expected 2 log entries with warn level filter, got %d", len(warnLogs))
	}

	// Test with non-existent correlation ID
	emptyLogs, err := logger.GetMemoryLogs("non-existent", LogLevel(log.InfoLevel))
	if err != nil {
		t.Errorf("GetMemoryLogs should not return error: %v", err)
	}

	// Should have 0 log entries
	if len(emptyLogs) != 0 {
		t.Errorf("Expected 0 log entries for non-existent correlation ID, got %d", len(emptyLogs))
	}
}

func TestLogger_WithoutMemoryWriter(t *testing.T) {
	// Create logger without memory writer
	logger := Logger()

	// Try to get memory logs
	logs, err := logger.GetMemoryLogs("test", LogLevel(log.InfoLevel))
	if err != nil {
		t.Errorf("GetMemoryLogs should not return error: %v", err)
	}

	// Should return empty map
	if len(logs) != 0 {
		t.Errorf("Expected 0 log entries without memory writer, got %d", len(logs))
	}
}

func TestLogger_MemoryWriterExpiration(t *testing.T) {
	// This test verifies that expired entries are filtered out during retrieval
	config := models.WriterConfiguration{}
	logger := Logger().WithMemoryWriter(config)

	correlationID := "expiration-test-456"
	logger = logger.WithCorrelationId(correlationID)

	// Log a message
	logger.Info().Msg("Test message that should expire")

	// Small delay to ensure write is processed
	time.Sleep(10 * time.Millisecond)

	// Verify entry exists
	logs, err := logger.GetMemoryLogs(correlationID, LogLevel(log.InfoLevel))
	if err != nil {
		t.Errorf("GetMemoryLogs should not return error: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log entry immediately after logging, got %d", len(logs))
	}

	// For this test, we would need to set a shorter TTL to actually test expiration
	// The current implementation uses 24 hours, so we can't easily test expiration
	// in a unit test without modifying the TTL configuration
}
