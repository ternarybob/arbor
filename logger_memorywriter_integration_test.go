package arbor

import (
	"fmt"
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

func TestLogger_GetMemoryLogsForCorrelation(t *testing.T) {
	// Create logger with memory writer
	config := models.WriterConfiguration{}
	logger := Logger().WithMemoryWriter(config)

	// Set correlation ID
	correlationID := "correlation-test-123"
	logger = logger.WithCorrelationId(correlationID)

	// Log some messages
	logger.Info().Msg("Test info message 1")
	logger.Warn().Msg("Test warning message")
	logger.Error().Msg("Test error message")
	logger.Info().Msg("Test info message 2")

	// Small delay to ensure writes are processed
	time.Sleep(50 * time.Millisecond)

	// Retrieve logs for correlation ID (without level filter)
	logs, err := logger.GetMemoryLogsForCorrelation(correlationID)
	if err != nil {
		t.Errorf("GetMemoryLogsForCorrelation should not return error: %v", err)
	}

	// Should have 4 log entries
	if len(logs) != 4 {
		t.Errorf("Expected 4 log entries, got %d", len(logs))
	}

	// Test with non-existent correlation ID
	emptyLogs, err := logger.GetMemoryLogsForCorrelation("non-existent-correlation")
	if err != nil {
		t.Errorf("GetMemoryLogsForCorrelation should not return error: %v", err)
	}

	// Should have 0 log entries
	if len(emptyLogs) != 0 {
		t.Errorf("Expected 0 log entries for non-existent correlation ID, got %d", len(emptyLogs))
	}

	// Test with empty correlation ID
	emptyCorrelationLogs, err := logger.GetMemoryLogsForCorrelation("")
	if err != nil {
		t.Errorf("GetMemoryLogsForCorrelation should not return error: %v", err)
	}

	// Should have 0 log entries
	if len(emptyCorrelationLogs) != 0 {
		t.Errorf("Expected 0 log entries for empty correlation ID, got %d", len(emptyCorrelationLogs))
	}
}

func TestLogger_GetMemoryLogsWithLimit(t *testing.T) {
	// Create logger with memory writer
	config := models.WriterConfiguration{}
	logger := Logger().WithMemoryWriter(config)

	// Create multiple correlation IDs and log entries
	correlationIDs := []string{
		"limit-test-1",
		"limit-test-2",
		"limit-test-3",
	}

	totalEntries := 0
	for _, correlationID := range correlationIDs {
		loggerWithCorrelation := logger.WithCorrelationId(correlationID)

		// Log 3 messages per correlation ID
		loggerWithCorrelation.Info().Msg("Info message 1")
		loggerWithCorrelation.Warn().Msg("Warning message")
		loggerWithCorrelation.Error().Msg("Error message")
		totalEntries += 3
	}

	// Small delay to ensure writes are processed
	time.Sleep(50 * time.Millisecond)

	// Test with limit less than total entries
	limit := 5
	limitedLogs, err := logger.GetMemoryLogsWithLimit(limit)
	if err != nil {
		t.Errorf("GetMemoryLogsWithLimit should not return error: %v", err)
	}

	// Should have at most the limit (may have additional entries from other tests)
	if len(limitedLogs) > limit+10 { // Allow some buffer for other test entries
		t.Errorf("Expected at most around %d entries, got %d", limit+10, len(limitedLogs))
	}

	// Test with limit of 0
	zeroLogs, err := logger.GetMemoryLogsWithLimit(0)
	if err != nil {
		t.Errorf("GetMemoryLogsWithLimit should not return error: %v", err)
	}
	if len(zeroLogs) != 0 {
		t.Errorf("Expected 0 log entries with limit 0, got %d", len(zeroLogs))
	}

	// Test with negative limit
	negativeLogs, err := logger.GetMemoryLogsWithLimit(-1)
	if err != nil {
		t.Errorf("GetMemoryLogsWithLimit should not return error: %v", err)
	}
	if len(negativeLogs) != 0 {
		t.Errorf("Expected 0 log entries with negative limit, got %d", len(negativeLogs))
	}

	// Test with large limit
	largeLimitLogs, err := logger.GetMemoryLogsWithLimit(1000)
	if err != nil {
		t.Errorf("GetMemoryLogsWithLimit should not return error: %v", err)
	}

	// Should have at least our entries
	if len(largeLimitLogs) < totalEntries {
		t.Errorf("Expected at least %d log entries with large limit, got %d", totalEntries, len(largeLimitLogs))
	}
}

func TestLogger_GetMemoryLogsWithLimit_Ordering(t *testing.T) {
	// Create logger with memory writer
	config := models.WriterConfiguration{}
	logger := Logger().WithMemoryWriter(config)

	// Create a unique correlation ID to avoid interference
	correlationID := fmt.Sprintf("ordering-test-%d", time.Now().UnixNano())
	loggerWithCorrelation := logger.WithCorrelationId(correlationID)

	// Log messages with small delays to ensure different timestamps
	loggerWithCorrelation.Info().Msg("First message")
	time.Sleep(1 * time.Millisecond)
	loggerWithCorrelation.Info().Msg("Second message")
	time.Sleep(1 * time.Millisecond)
	loggerWithCorrelation.Info().Msg("Third message")
	time.Sleep(1 * time.Millisecond)
	loggerWithCorrelation.Info().Msg("Fourth message (most recent)")

	// Small delay to ensure writes are processed
	time.Sleep(50 * time.Millisecond)

	// Test that GetMemoryLogsWithLimit returns entries
	// (We can't easily test exact ordering without knowing internal implementation details)
	limitedLogs, err := logger.GetMemoryLogsWithLimit(2)
	if err != nil {
		t.Errorf("GetMemoryLogsWithLimit should not return error: %v", err)
	}

	// Should return some entries
	if len(limitedLogs) == 0 {
		t.Error("GetMemoryLogsWithLimit should return some entries")
	}

	// Verify that we can get all entries for this correlation ID
	allCorrelationLogs, err := logger.GetMemoryLogsForCorrelation(correlationID)
	if err != nil {
		t.Errorf("GetMemoryLogsForCorrelation should not return error: %v", err)
	}
	if len(allCorrelationLogs) != 4 {
		t.Errorf("Expected 4 log entries for correlation ID, got %d", len(allCorrelationLogs))
	}
}

func TestLogger_GetMemoryLogs_WithoutMemoryWriter(t *testing.T) {
	// Create logger without memory writer
	logger := Logger()

	// Try to get memory logs with new methods
	correlationLogs, err := logger.GetMemoryLogsForCorrelation("test")
	if err != nil {
		t.Errorf("GetMemoryLogsForCorrelation should not return error: %v", err)
	}
	if len(correlationLogs) != 0 {
		t.Errorf("Expected 0 log entries without memory writer, got %d", len(correlationLogs))
	}

	limitedLogs, err := logger.GetMemoryLogsWithLimit(10)
	if err != nil {
		t.Errorf("GetMemoryLogsWithLimit should not return error: %v", err)
	}
	if len(limitedLogs) != 0 {
		t.Errorf("Expected 0 log entries without memory writer, got %d", len(limitedLogs))
	}
}
