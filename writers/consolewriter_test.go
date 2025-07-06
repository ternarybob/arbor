package writers

import (
	"testing"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

func TestConsoleWriter_New(t *testing.T) {
	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeConsole,
		Level:      levels.InfoLevel,
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
		Level:      levels.InfoLevel,
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
		Level:      levels.InfoLevel,
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
