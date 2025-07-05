package writers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/models"
)

func TestNewMemoryWriter(t *testing.T) {
	writer := NewMemoryWriter()
	if writer == nil {
		t.Fatal("NewMemoryWriter should not return nil")
	}

	// Verify it implements IWriter interface
	var _ IWriter = writer
}

func TestMemoryWriter_WithLevel(t *testing.T) {
	writer := NewMemoryWriter()

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

func TestMemoryWriter_Write(t *testing.T) {
	writer := NewMemoryWriter()

	testCases := []struct {
		name     string
		input    []byte
		expected int
	}{
		{
			name:     "normal message",
			input:    []byte("test message"),
			expected: 12,
		},
		{
			name:     "empty message",
			input:    []byte(""),
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, err := writer.Write(tc.input)
			if err != nil {
				t.Errorf("Write should not return error: %v", err)
			}
			if n != tc.expected {
				t.Errorf("Expected %d bytes written, got %d", tc.expected, n)
			}
		})
	}
}

func TestMemoryWriter_WriteLogEvent(t *testing.T) {
	// Clear any existing entries
	ClearAllEntries()

	writer := NewMemoryWriter()
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

	// Verify the entry was stored
	entries, err := GetEntries(correlationID)
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestMemoryWriter_WriteWithoutCorrelationID(t *testing.T) {
	// Clear any existing entries
	ClearAllEntries()

	writer := NewMemoryWriter()

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
	entries, err := GetEntries("")
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(entries))
	}
}

func TestGetEntries(t *testing.T) {
	// Clear any existing entries
	ClearAllEntries()

	writer := NewMemoryWriter()
	correlationID := "test-get-entries"

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

	// Get entries
	entries, err := GetEntries(correlationID)
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Test with empty correlation ID
	entries, err = GetEntries("")
	if err != nil {
		t.Errorf("GetEntries should not return error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries for empty correlation ID, got %d", len(entries))
	}
}

func TestGetEntriesWithLevel(t *testing.T) {
	// Clear any existing entries
	ClearAllEntries()

	writer := NewMemoryWriter()
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

	// Get entries with minimum level of Warn
	entries, err := GetEntriesWithLevel(correlationID, log.WarnLevel)
	if err != nil {
		t.Errorf("GetEntriesWithLevel should not return error: %v", err)
	}
	// Should include Warn and Error (2 entries)
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries with WarnLevel filter, got %d", len(entries))
	}

	// Get entries with minimum level of Info
	entries, err = GetEntriesWithLevel(correlationID, log.InfoLevel)
	if err != nil {
		t.Errorf("GetEntriesWithLevel should not return error: %v", err)
	}
	// Should include Info, Warn, and Error (3 entries)
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries with InfoLevel filter, got %d", len(entries))
	}
}

func TestFormatLogEvent(t *testing.T) {
	logEvent := &models.LogEvent{
		Level:     log.InfoLevel,
		Timestamp: time.Now(),
		Message:   "test message",
		Prefix:    "TEST",
		Error:     "test error",
	}

	result := formatLogEvent(logEvent)
	if result == "" {
		t.Error("FormatLogEvent should not return empty string")
	}

	// Should contain level abbreviation
	if !contains(result, "INF") {
		t.Error("Formatted log should contain level abbreviation")
	}

	// Should contain message
	if !contains(result, "test message") {
		t.Error("Formatted log should contain message")
	}
}

func TestClearEntries(t *testing.T) {
	// Clear any existing entries
	ClearAllEntries()

	writer := NewMemoryWriter()
	correlationID := "test-clear"

	// Add a log event
	logEvent := models.LogEvent{
		Level:         log.InfoLevel,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Message:       "test message",
	}

	jsonData, _ := json.Marshal(logEvent)
	writer.Write(jsonData)

	// Verify entry exists
	entries, _ := GetEntries(correlationID)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry before clear, got %d", len(entries))
	}

	// Clear entries for this correlation ID
	ClearEntries(correlationID)

	// Verify entry is gone
	entries, _ = GetEntries(correlationID)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(entries))
	}
}

func TestGetStoredCorrelationIDs(t *testing.T) {
	// Clear any existing entries
	ClearAllEntries()

	writer := NewMemoryWriter()
	correlationIDs := []string{"test-1", "test-2", "test-3"}

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

	// Get stored correlation IDs
	storedIDs := GetStoredCorrelationIDs()
	if len(storedIDs) != 3 {
		t.Errorf("Expected 3 correlation IDs, got %d", len(storedIDs))
	}

	// Verify all IDs are present
	for _, expectedID := range correlationIDs {
		found := false
		for _, storedID := range storedIDs {
			if storedID == expectedID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find correlation ID %s in stored IDs", expectedID)
		}
	}
}

func TestBufferLimit(t *testing.T) {
	// Clear any existing entries
	ClearAllEntries()

	writer := NewMemoryWriter()
	correlationID := "test-buffer-limit"

	// Add more entries than the buffer limit
	for i := 0; i < BUFFER_LIMIT+10; i++ {
		logEvent := models.LogEvent{
			Level:         log.InfoLevel,
			Timestamp:     time.Now(),
			CorrelationID: correlationID,
			Message:       "test message",
		}

		jsonData, _ := json.Marshal(logEvent)
		writer.Write(jsonData)
	}

	// Should not exceed buffer limit
	entries, _ := GetEntries(correlationID)
	if len(entries) > BUFFER_LIMIT {
		t.Errorf("Entries should not exceed buffer limit of %d, got %d", BUFFER_LIMIT, len(entries))
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
