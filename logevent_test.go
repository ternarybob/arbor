package arbor

import (
	"errors"
	"testing"

	"github.com/phuslu/log"
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

func TestLogEvent_Bool(t *testing.T) {
	logger := Logger().(*logger)
	event := newLogEvent(logger, log.InfoLevel)

	// Test adding a boolean field
	result := event.Bool("isEnabled", true)
	if result == nil {
		t.Error("Bool should not return nil")
	}

	// Should return the same instance for chaining
	if result != event {
		t.Error("Bool should return the same instance for chaining")
	}

	// Test that field was added with correct value
	if event.fields["isEnabled"] != true {
		t.Error("Boolean field should be added to the event with value true")
	}

	// Test chaining multiple boolean fields
	event.Bool("isActive", false).Bool("hasPermission", true)

	if len(event.fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(event.fields))
	}

	if event.fields["isActive"] != false {
		t.Error("Second boolean field should be false")
	}

	if event.fields["hasPermission"] != true {
		t.Error("Third boolean field should be true")
	}

	// Test chaining with other field types
	event.Str("component", "auth").Bool("authenticated", true)
	if event.fields["component"] != "auth" {
		t.Error("Should be able to chain Str and Bool methods")
	}
	if event.fields["authenticated"] != true {
		t.Error("Boolean field should be set when chained with other methods")
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

	// Test with multiple fields including bool
	logger.Error().
		Str("user", "john").
		Str("action", "delete").
		Str("resource", "file.txt").
		Bool("authenticated", true).
		Bool("authorized", false).
		Err(errors.New("permission denied")).
		Msg("Operation failed")
}
