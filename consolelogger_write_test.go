package arbor

import (
	"strings"
	"testing"
	"time"

	"github.com/ternarybob/arbor/writers"
)

func TestConsoleLogger_Write_GinLogDetection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Gin debug log",
			input:    "[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.",
			expected: true,
		},
		{
			name:     "Gin request log",
			input:    "[GIN] 2023/12/07 - 15:04:05 | 200 |     123.456ms |       ::1 | GET      \"/ping\"",
			expected: true,
		},
		{
			name:     "HTTP request pattern",
			input:    "| 200 | 1.234ms | 127.0.0.1 | GET",
			expected: true,
		},
		{
			name:     "POST request",
			input:    "| 201 | 2.456ms | 192.168.1.1 | POST",
			expected: true,
		},
		{
			name:     "Regular log message",
			input:    "This is a regular log message",
			expected: false,
		},
		{
			name:     "JSON log entry",
			input:    `{"level":"info","message":"test"}`,
			expected: false,
		},
	}

	logger := ConsoleLogger()
	ginDetector := writers.NewGinDetector(logger.GetLevel())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ginDetector.IsGinLog(tt.input)
			if result != tt.expected {
				t.Errorf("isGinLog() = %v, expected %v for input: %s", result, tt.expected, tt.input)
			}
		})
	}
}

func TestConsoleLogger_Write_Integration(t *testing.T) {
	// Create a logger with memory writer to capture output
	logger := ConsoleLogger().WithWriter("memory", writers.NewMemoryWriter())

	// Test Gin log
	ginLog := "[GIN] 2023/12/07 - 15:04:05 | 200 |     123.456ms |       ::1 | GET      \"/ping\""
	n, err := logger.Write([]byte(ginLog))

	if err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}

	if n != len(ginLog) {
		t.Errorf("Write() returned n = %d, expected %d", n, len(ginLog))
	}

	// Test regular log
	regularLog := "This is a regular log message"
	n, err = logger.Write([]byte(regularLog))

	if err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}

	if n != len(regularLog) {
		t.Errorf("Write() returned n = %d, expected %d", n, len(regularLog))
	}
}

func TestConsoleLogger_LevelFiltering(t *testing.T) {
	// Create logger with WARN level
	logger := ConsoleLogger().WithLevel(WarnLevel)
	ginDetector := writers.NewGinDetector(logger.GetLevel())

	tests := []struct {
		level    string
		expected bool
	}{
		{"fatal", true},
		{"error", true},
		{"warn", true},
		{"info", false},  // Should be filtered out
		{"debug", false}, // Should be filtered out
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			result := ginDetector.ShouldLogLevel(tt.level)
			if result != tt.expected {
				t.Errorf("shouldLogLevel(%s) = %v, expected %v", tt.level, result, tt.expected)
			}
		})
	}
}

func TestConsoleLogger_FormatGinConsoleOutput(t *testing.T) {
	logger := ConsoleLogger()
	ginDetector := writers.NewGinDetector(logger.GetLevel())

	testTime := time.Date(2023, 12, 7, 15, 4, 5, 0, time.UTC)
	logEntry := &writers.GinLogEvent{
		Level:     "info",
		Timestamp: testTime,
		Prefix:    "GIN",
		Message:   "Test message",
	}

	output := ginDetector.FormatConsoleOutput(logEntry)

	// Should contain level, timestamp, prefix, and message
	if !strings.Contains(output, "INF") {
		t.Error("Output should contain formatted level")
	}
	if !strings.Contains(output, "GIN") {
		t.Error("Output should contain prefix")
	}
	if !strings.Contains(output, "Test message") {
		t.Error("Output should contain message")
	}
	if !strings.Contains(output, "|") {
		t.Error("Output should contain pipe separators")
	}
}

func TestConsoleLogger_WriteImplementsInterface(t *testing.T) {
	logger := ConsoleLogger()

	// This should compile - testing that ConsoleLogger implements io.Writer
	var writer interface{} = logger
	if _, ok := writer.(interface{ Write([]byte) (int, error) }); !ok {
		t.Error("ConsoleLogger should implement io.Writer interface")
	}
}
