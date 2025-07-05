package writers

import (
	"testing"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/models"
)

func TestConsoleWriter_New(t *testing.T) {
	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeConsole,
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
	}

	writer := ConsoleWriter(config)
	if writer == nil {
		t.Fatal("ConsoleWriter should not return nil")
	}

	// Verify it implements IWriter interface
	var _ IWriter = writer
}

func TestConsoleWriter_WithLevel(t *testing.T) {
	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeConsole,
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
	}

	writer := ConsoleWriter(config)
	
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

func TestConsoleWriter_Write(t *testing.T) {
	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeConsole,
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
	}

	writer := ConsoleWriter(config)
	
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
		{
			name:     "json message",
			input:    []byte(`{"level":"info","msg":"test"}`),
			expected: 29,
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

func TestLevelToString(t *testing.T) {
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
			result := levelToString(tc.level)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestConsoleWriter_Format(t *testing.T) {
	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeConsole,
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
	}

	cw := ConsoleWriter(config).(*consoleWriter)
	
	logEvent := &models.LogEvent{
		Level:   log.InfoLevel,
		Message: "test message",
		Prefix:  "TEST",
		Error:   "test error",
	}

	// Test with color
	result := cw.format(logEvent, true)
	if result == "" {
		t.Error("Format should not return empty string")
	}

	// Test without color
	result = cw.format(logEvent, false)
	if result == "" {
		t.Error("Format should not return empty string")
	}
}
