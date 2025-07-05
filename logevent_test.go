package arbor

import (
	"errors"
	"strings"
	"testing"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/models"
)

func TestNewLogEvent(t *testing.T) {
	logger := Logger().(*logger)

	testCases := []struct {
		name  string
		level log.Level
	}{
		{"Trace", log.TraceLevel},
		{"Debug", log.DebugLevel},
		{"Info", log.InfoLevel},
		{"Warn", log.WarnLevel},
		{"Error", log.ErrorLevel},
		{"Fatal", log.FatalLevel},
		{"Panic", log.PanicLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := newLogEvent(logger, tc.level)
			if event == nil {
				t.Error("newLogEvent should not return nil")
			}

			if event.level != tc.level {
				t.Errorf("Expected level %v, got %v", tc.level, event.level)
			}

			if event.logger != logger {
				t.Error("Event should reference the correct logger")
			}

			if event.fields == nil {
				t.Error("Fields map should be initialized")
			}
		})
	}
}

func TestLogEvent_Str(t *testing.T) {
	logger := Logger().(*logger)
	event := newLogEvent(logger, log.InfoLevel)

	// Test adding a string field
	result := event.Str("key1", "value1")
	if result == nil {
		t.Error("Str should not return nil")
	}

	// Should return the same instance for chaining
	if result != event {
		t.Error("Str should return the same instance for chaining")
	}

	// Test that field was added
	if event.fields["key1"] != "value1" {
		t.Error("Field should be added to the event")
	}

	// Test chaining multiple fields
	event.Str("key2", "value2").Str("key3", "value3")

	if len(event.fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(event.fields))
	}

	if event.fields["key2"] != "value2" {
		t.Error("Second field should be added")
	}

	if event.fields["key3"] != "value3" {
		t.Error("Third field should be added")
	}
}

func TestLogEvent_Err(t *testing.T) {
	logger := Logger().(*logger)
	event := newLogEvent(logger, log.ErrorLevel)

	testErr := errors.New("test error message")
	result := event.Err(testErr)

	if result == nil {
		t.Error("Err should not return nil")
	}

	// Should return the same instance for chaining
	if result != event {
		t.Error("Err should return the same instance for chaining")
	}

	// Test that error was set
	if event.err != testErr {
		t.Error("Error should be set on the event")
	}

	// Test chaining with other methods
	event.Str("component", "auth").Err(testErr)

	if event.fields["component"] != "auth" {
		t.Error("Should be able to chain Str and Err methods")
	}
}

func TestLogEvent_LevelToString(t *testing.T) {
	logger := Logger().(*logger)
	event := newLogEvent(logger, log.InfoLevel)

	testCases := []struct {
		level    log.Level
		expected string
	}{
		{log.TraceLevel, "TRACE"},
		{log.DebugLevel, "DEBUG"},
		{log.InfoLevel, "INFO"},
		{log.WarnLevel, "WARN"},
		{log.ErrorLevel, "ERROR"},
		{log.FatalLevel, "FATAL"},
		{log.PanicLevel, "PANIC"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := event.levelToString(tc.level)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestLogEvent_FormatLogEntry(t *testing.T) {
	logger := Logger().(*logger)
	event := newLogEvent(logger, log.InfoLevel)

	// Add some fields and error
	event.Str("key1", "value1").Str("key2", "value2").Err(errors.New("test error"))

	// Create a mock log event
	logEvent := &models.LogEvent{
		Level:         log.InfoLevel,
		Message:       "test message",
		Prefix:        "TEST",
		CorrelationID: "correlation-123",
		Function:      "test.function",
		Fields:        event.fields,
		Error:         "test error",
	}

	result := event.formatLogEntry(logEvent)

	// Should not be empty
	if result == "" {
		t.Error("formatLogEntry should not return empty string")
	}

	// Should contain level
	if !strings.Contains(result, "INFO") {
		t.Error("Formatted entry should contain level")
	}

	// Should contain message
	if !strings.Contains(result, "test message") {
		t.Error("Formatted entry should contain message")
	}

	// Should contain prefix
	if !strings.Contains(result, "TEST") {
		t.Error("Formatted entry should contain prefix")
	}

	// Should contain correlation ID
	if !strings.Contains(result, "correlation-123") {
		t.Error("Formatted entry should contain correlation ID")
	}

	// Should contain fields
	if !strings.Contains(result, "key1=value1") {
		t.Error("Formatted entry should contain fields")
	}

	// Should contain error
	if !strings.Contains(result, "error=test error") {
		t.Error("Formatted entry should contain error")
	}

	// Should end with newline
	if !strings.HasSuffix(result, "\n") {
		t.Error("Formatted entry should end with newline")
	}
}

func TestLogEvent_FormatLogEntry_EmptyFields(t *testing.T) {
	logger := Logger().(*logger)
	event := newLogEvent(logger, log.InfoLevel)

	// Create a minimal log event
	logEvent := &models.LogEvent{
		Level:   log.InfoLevel,
		Message: "simple message",
	}

	result := event.formatLogEntry(logEvent)

	// Should still work with minimal fields
	if result == "" {
		t.Error("formatLogEntry should not return empty string even with minimal fields")
	}

	// Should contain level and message
	if !strings.Contains(result, "INFO") {
		t.Error("Should contain level")
	}

	if !strings.Contains(result, "simple message") {
		t.Error("Should contain message")
	}
}

func TestLogEvent_MethodChaining(t *testing.T) {
	logger := Logger().(*logger)
	event := newLogEvent(logger, log.InfoLevel)

	// Test complex method chaining
	result := event.
		Str("user", "john").
		Str("action", "login").
		Err(errors.New("invalid password"))

	if result != event {
		t.Error("All chained methods should return the same instance")
	}

	// Verify all fields were set
	if event.fields["user"] != "john" {
		t.Error("First Str field should be set")
	}

	if event.fields["action"] != "login" {
		t.Error("Second Str field should be set")
	}

	if event.err == nil {
		t.Error("Error should be set")
	}

	if event.err.Error() != "invalid password" {
		t.Error("Error message should match")
	}
}

func TestLogEvent_FieldOverwrite(t *testing.T) {
	logger := Logger().(*logger)
	event := newLogEvent(logger, log.InfoLevel)

	// Test that fields can be overwritten
	event.Str("key", "value1")
	if event.fields["key"] != "value1" {
		t.Error("First value should be set")
	}

	event.Str("key", "value2")
	if event.fields["key"] != "value2" {
		t.Error("Field should be overwritten with new value")
	}
}

func TestLogEvent_ErrorOverwrite(t *testing.T) {
	logger := Logger().(*logger)
	event := newLogEvent(logger, log.InfoLevel)

	// Test that errors can be overwritten
	err1 := errors.New("first error")
	err2 := errors.New("second error")

	event.Err(err1)
	if event.err != err1 {
		t.Error("First error should be set")
	}

	event.Err(err2)
	if event.err != err2 {
		t.Error("Error should be overwritten with new error")
	}
}

// Integration test to ensure the full flow works
func TestLogEvent_EndToEnd(t *testing.T) {
	// This test verifies that the complete logging flow works without panics
	logger := Logger().WithCorrelationId("test-correlation").WithPrefix("TEST")

	// Test that we can create events and call terminal methods without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Logging should not panic: %v", r)
		}
	}()

	// Test Msg method
	logger.Info().Str("key", "value").Msg("test message")

	// Test Msgf method
	logger.Warn().Err(errors.New("test error")).Msgf("formatted message %s %d", "test", 123)

	// Test with multiple fields
	logger.Error().
		Str("user", "john").
		Str("action", "delete").
		Str("resource", "file.txt").
		Err(errors.New("permission denied")).
		Msg("Operation failed")
}
