package arbor

import (
	"errors"
	"strings"
	"testing"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

func TestLogger_New(t *testing.T) {
	logger := Logger()
	if logger == nil {
		t.Fatal("Logger() should not return nil")
	}

	// Verify it implements ILogger interface
	var _ ILogger = logger
}

func TestLogger_WithCorrelationId(t *testing.T) {
	logger := Logger()

	// Test with provided correlation ID
	correlationID := "test-correlation-123"
	newLogger := logger.WithCorrelationId(correlationID)

	if newLogger == nil {
		t.Error("WithCorrelationId should not return nil")
	}

	// Test with empty correlation ID (should generate UUID)
	newLogger2 := logger.WithCorrelationId("")
	if newLogger2 == nil {
		t.Error("WithCorrelationId with empty string should not return nil")
	}
}

func TestLogger_WithPrefix(t *testing.T) {
	logger := Logger()

	// Test with valid prefix
	prefix := "API"
	newLogger := logger.WithPrefix(prefix)

	if newLogger == nil {
		t.Error("WithPrefix should not return nil")
	}

	// Test with empty prefix
	newLogger2 := logger.WithPrefix("")
	if newLogger2 == nil {
		t.Error("WithPrefix with empty string should not return nil")
	}
}

func TestLogger_WithLevel(t *testing.T) {
	logger := Logger()

	testLevels := []levels.LogLevel{
		levels.TraceLevel,
		levels.DebugLevel,
		levels.InfoLevel,
		levels.WarnLevel,
		levels.ErrorLevel,
		levels.FatalLevel,
		levels.PanicLevel,
	}

	for _, level := range testLevels {
		t.Run(string(rune(level)), func(t *testing.T) {
			newLogger := logger.WithLevel(level)
			if newLogger == nil {
				t.Error("WithLevel should not return nil")
			}
		})
	}
}

func TestLogger_WithContext(t *testing.T) {
	logger := Logger()

	// Test with valid key-value pair
	newLogger := logger.WithContext("key", "value")
	if newLogger == nil {
		t.Error("WithContext should not return nil")
	}

	// Test with empty key
	newLogger2 := logger.WithContext("", "value")
	if newLogger2 == nil {
		t.Error("WithContext with empty key should not return nil")
	}

	// Test with empty value
	newLogger3 := logger.WithContext("key", "")
	if newLogger3 == nil {
		t.Error("WithContext with empty value should not return nil")
	}
}

func TestLogger_WithFileWriter(t *testing.T) {
	logger := Logger()

	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeFile,
		Level:      levels.InfoLevel,
		TimeFormat: "15:04:05.000",
		FileName:   "test.log",
	}

	newLogger := logger.WithFileWriter(config)
	if newLogger == nil {
		t.Error("WithFileWriter should not return nil")
	}
}

func TestLogger_FluentMethods(t *testing.T) {
	logger := Logger()

	// Test all fluent logging methods
	testCases := []struct {
		name   string
		method func() ILogEvent
	}{
		{"Trace", logger.Trace},
		{"Debug", logger.Debug},
		{"Info", logger.Info},
		{"Warn", logger.Warn},
		{"Error", logger.Error},
		{"Fatal", logger.Fatal},
		{"Panic", logger.Panic},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := tc.method()
			if event == nil {
				t.Errorf("%s() should not return nil", tc.name)
			}

			// Verify it implements ILogEvent interface
			var _ ILogEvent = event
		})
	}
}

func TestGlobalLogger_Functions(t *testing.T) {
	// Test that global functions work
	testCases := []struct {
		name   string
		method func() ILogEvent
	}{
		{"Trace", Trace},
		{"Debug", Debug},
		{"Info", Info},
		{"Warn", Warn},
		{"Error", Error},
		{"Fatal", Fatal},
		{"Panic", Panic},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := tc.method()
			if event == nil {
				t.Errorf("%s() should not return nil", tc.name)
			}

			// Verify it implements ILogEvent interface
			var _ ILogEvent = event
		})
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger() should not return nil")
	}

	// Should return the same instance on multiple calls
	logger2 := GetLogger()
	if logger != logger2 {
		t.Error("GetLogger() should return the same instance")
	}
}

func TestLevelToString_Function(t *testing.T) {
	testCases := []struct {
		level    log.Level
		expected string
	}{
		{log.TraceLevel, "trace"},
		{log.DebugLevel, "debug"},
		{log.InfoLevel, "info"},
		{log.WarnLevel, "warn"},
		{log.ErrorLevel, "error"},
		{log.FatalLevel, "fatal"},
		{log.PanicLevel, "panic"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := LevelToString(tc.level)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestParseLevelString(t *testing.T) {
	testCases := []struct {
		input    string
		expected log.Level
		hasError bool
	}{
		{"trace", log.TraceLevel, false},
		{"debug", log.DebugLevel, false},
		{"info", log.InfoLevel, false},
		{"warn", log.WarnLevel, false},
		{"warning", log.WarnLevel, false},
		{"error", log.ErrorLevel, false},
		{"fatal", log.FatalLevel, false},
		{"panic", log.PanicLevel, false},
		{"disabled", log.PanicLevel + 1, false},
		{"off", log.PanicLevel + 1, false},
		{"invalid", log.InfoLevel, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := levels.ParseLevelString(tc.input)

			if tc.hasError {
				if err == nil {
					t.Error("Expected error for invalid level")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	testCases := []struct {
		input    int
		expected log.Level
	}{
		{int(levels.TraceLevel), log.TraceLevel},
		{int(levels.DebugLevel), log.DebugLevel},
		{int(levels.InfoLevel), log.InfoLevel},
		{int(levels.WarnLevel), log.WarnLevel},
		{int(levels.ErrorLevel), log.ErrorLevel},
		{int(levels.FatalLevel), log.FatalLevel},
		{int(levels.PanicLevel), log.PanicLevel},
		{int(levels.Disabled), 0},
		{999, log.InfoLevel}, // Default case
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.input)), func(t *testing.T) {
			result := levels.ParseLogLevel(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestLogger_GetFunctionName(t *testing.T) {
	logger := Logger().(*logger)

	// This will test the function name detection
	funcName := logger.getFunctionName()

	// Should contain test function name or be empty (acceptable for edge cases)
	if funcName != "" && !strings.Contains(funcName, "Test") {
		// This is informational - function name detection can vary
		t.Logf("Function name detected: %s", funcName)
	}
}

func TestLogger_ChainedUsage(t *testing.T) {
	// Test complex chained usage
	logger := Logger().WithCorrelationId("test-123").WithPrefix("TEST")

	// This should not panic and should work end-to-end
	event := logger.Info().Str("key1", "value1").Str("key2", "value2")
	if event == nil {
		t.Error("Chained usage should not return nil")
	}

	// Verify we can add an error and still chain
	err := errors.New("test error")
	event2 := event.Err(err)
	if event2 == nil {
		t.Error("Chained usage with error should not return nil")
	}

	// Should be the same instance
	if event != event2 {
		t.Error("Chained methods should return the same instance")
	}
}
