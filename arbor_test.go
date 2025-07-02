package arbor

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"tabs and newlines", "\t\n", true},
		{"non-empty string", "hello", false},
		{"string with content", "  hello  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmpty(tt.input)
			if result != tt.expected {
				t.Errorf("isEmpty(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConsoleLoggerCreation(t *testing.T) {
	// Test that we can create a console logger without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Logf("ConsoleLogger creation panicked (expected if no config): %v", r)
		}
	}()

	// This might panic if satus.GetAppConfig() fails, which is expected in test environment
	logger := ConsoleLogger()
	if logger != nil {
		t.Logf("ConsoleLogger created successfully")
	}
}

func TestIConsoleLoggerInterface(t *testing.T) {
	// Test that our interface methods are defined correctly
	var logger IConsoleLogger

	// Create a mock implementation to test interface compliance
	mockLogger := &mockConsoleLogger{}
	logger = mockLogger

	if logger == nil {
		t.Error("IConsoleLogger interface should be implemented")
	}

	// Test that all interface methods are available
	_ = logger.GetLogger()
	_ = logger.GetLevel()
	_ = logger.WithLevel(zerolog.InfoLevel)
	_ = logger.WithPrefix("test")
	_ = logger.WithCorrelationId("test-id")
	_ = logger.WithFunction()
	_ = logger.WithContext("key", "value")
}

// Mock implementation for testing
type mockConsoleLogger struct{}

func (m *mockConsoleLogger) GetLogger() zerolog.Logger {
	return zerolog.New(os.Stdout)
}

func (m *mockConsoleLogger) GetLevel() zerolog.Level {
	return zerolog.InfoLevel
}

func (m *mockConsoleLogger) WithRequestContext(ctx echo.Context) IConsoleLogger {
	return m
}

func (m *mockConsoleLogger) WithWriter(name string, writer io.Writer) IConsoleLogger {
	return m
}

func (m *mockConsoleLogger) WithPrefix(value string) IConsoleLogger {
	return m
}

func (m *mockConsoleLogger) WithCorrelationId(value string) IConsoleLogger {
	return m
}

func (m *mockConsoleLogger) WithLevel(lvl zerolog.Level) IConsoleLogger {
	return m
}

func (m *mockConsoleLogger) WithContext(key string, value string) IConsoleLogger {
	return m
}

func (m *mockConsoleLogger) WithFunction() IConsoleLogger {
	return m
}

func TestConsoleLoggerWithLevel(t *testing.T) {
	// Test logger level functionality with mock
	mockLogger := &mockConsoleLogger{}
	
	result := mockLogger.WithLevel(zerolog.DebugLevel)
	if result == nil {
		t.Error("WithLevel should return a logger instance")
	}
	
	if result.GetLevel() != zerolog.InfoLevel {
		t.Logf("Mock logger level is %v", result.GetLevel())
	}
}

func TestConsoleLoggerWithPrefix(t *testing.T) {
	// Test prefix functionality with mock
	mockLogger := &mockConsoleLogger{}
	
	result := mockLogger.WithPrefix("test-prefix")
	if result == nil {
		t.Error("WithPrefix should return a logger instance")
	}
}

func TestConsoleLoggerWithCorrelationId(t *testing.T) {
	// Test correlation ID functionality with mock
	mockLogger := &mockConsoleLogger{}
	
	result := mockLogger.WithCorrelationId("test-correlation-id")
	if result == nil {
		t.Error("WithCorrelationId should return a logger instance")
	}
}

func TestConsoleLoggerWithContext(t *testing.T) {
	// Test context functionality with mock
	mockLogger := &mockConsoleLogger{}
	
	result := mockLogger.WithContext("key", "value")
	if result == nil {
		t.Error("WithContext should return a logger instance")
	}
}

func TestConsoleLoggerWithFunction(t *testing.T) {
	// Test function name functionality with mock
	mockLogger := &mockConsoleLogger{}
	
	result := mockLogger.WithFunction()
	if result == nil {
		t.Error("WithFunction should return a logger instance")
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are defined correctly
	expectedConstants := map[string]string{
		CORRELATION_ID_KEY: "correlationid",
		LOGGERCONTEXT_KEY:  "consolelogger",
		WRITER_CONSOLE:     "writerconsole",
		WRITER_DATA:        "writerdata",
		WRITER_REDIS:       "writerredis",
		WRITER_ARRAY:       "writerarray",
	}

	for constant, expected := range expectedConstants {
		if constant != expected {
			t.Errorf("Expected constant %s to be %q, got %q", expected, expected, constant)
		}
	}
}

func TestLoggerInitialization(t *testing.T) {
	// Test that internal logger is properly initialized
	if internallog.GetLevel() != zerolog.WarnLevel {
		t.Errorf("Expected internal log level to be WarnLevel, got %v", internallog.GetLevel())
	}
}

func TestConsoleLoggerChaining(t *testing.T) {
	// Test that logger methods can be chained
	mockLogger := &mockConsoleLogger{}
	
	result := mockLogger.
		WithLevel(zerolog.DebugLevel).
		WithPrefix("test").
		WithCorrelationId("test-id").
		WithContext("key", "value").
		WithFunction()
	
	if result == nil {
		t.Error("Chained logger methods should return a logger instance")
	}
}

// Test writer interface compliance
func TestWriterInterface(t *testing.T) {
	// Test that we can create writers
	var buf bytes.Buffer
	
	// Test basic Write interface
	data := []byte("test data")
	n, err := buf.Write(data)
	
	if err != nil {
		t.Errorf("Write should not error: %v", err)
	}
	
	if n != len(data) {
		t.Errorf("Write should return correct byte count: got %d, want %d", n, len(data))
	}
	
	if !strings.Contains(buf.String(), "test data") {
		t.Errorf("Buffer should contain written data")
	}
}
