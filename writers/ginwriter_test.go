package writers

import (
	"strings"
	"testing"
	"time"

	"github.com/phuslu/log"
)

func TestGinLogDetector_IsGinLog(t *testing.T) {
	detector := NewGinDetector(log.InfoLevel)

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
			name:     "PUT request",
			input:    "| 200 | 1.789ms | 10.0.0.1 | PUT",
			expected: true,
		},
		{
			name:     "DELETE request",
			input:    "| 204 | 0.456ms | 192.168.1.100 | DELETE",
			expected: true,
		},
		{
			name:     "PATCH request",
			input:    "| 200 | 3.456ms | 172.16.0.1 | PATCH",
			expected: true,
		},
		{
			name:     "HEAD request",
			input:    "| 200 | 0.123ms | 127.0.0.1 | HEAD",
			expected: true,
		},
		{
			name:     "OPTIONS request",
			input:    "| 200 | 0.789ms | 192.168.0.1 | OPTIONS",
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
		{
			name:     "SQL query log",
			input:    "SELECT * FROM users WHERE id = 1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.IsGinLog(tt.input)
			if result != tt.expected {
				t.Errorf("IsGinLog() = %v, expected %v for input: %s", result, tt.expected, tt.input)
			}
		})
	}
}

func TestGinLogDetector_ParseGinLog(t *testing.T) {
	detector := NewGinDetector(log.InfoLevel)

	tests := []struct {
		name          string
		input         string
		expectedLevel string
		expectedMsg   string
	}{
		{
			name:          "Fatal log",
			input:         "[GIN-fatal] Fatal error occurred",
			expectedLevel: "fatal",
			expectedMsg:   "Fatal error occurred",
		},
		{
			name:          "Error log",
			input:         "[GIN-error] An error happened",
			expectedLevel: "error",
			expectedMsg:   "An error happened",
		},
		{
			name:          "Warning log",
			input:         "[GIN-warning] This is a warning",
			expectedLevel: "warn",
			expectedMsg:   "This is a warning",
		},
		{
			name:          "Information log",
			input:         "[GIN-information] Information message",
			expectedLevel: "info",
			expectedMsg:   "Information message",
		},
		{
			name:          "Debug log",
			input:         "[GIN-debug] Debug message",
			expectedLevel: "debug",
			expectedMsg:   "Debug message",
		},
		{
			name:          "Standard Gin log",
			input:         "[GIN] 2023/12/07 - 15:04:05 | 200 |     123.456ms |       ::1 | GET      \"/ping\"",
			expectedLevel: "info",
			expectedMsg:   "[GIN] 2023/12/07 - 15:04:05 | 200 |     123.456ms |       ::1 | GET      \"/ping\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.ParseGinLog([]byte(tt.input))
			
			if result.Level != tt.expectedLevel {
				t.Errorf("ParseGinLog() level = %v, expected %v", result.Level, tt.expectedLevel)
			}
			
			if result.Message != tt.expectedMsg {
				t.Errorf("ParseGinLog() message = %v, expected %v", result.Message, tt.expectedMsg)
			}
			
			if result.Prefix != "GIN" {
				t.Errorf("ParseGinLog() prefix = %v, expected GIN", result.Prefix)
			}
		})
	}
}

func TestGinLogDetector_ShouldLogLevel(t *testing.T) {
	tests := []struct {
		name         string
		detectorLevel log.Level
		logLevel     string
		expected     bool
	}{
		{
			name:         "Fatal should be logged at warn level",
			detectorLevel: log.WarnLevel,
			logLevel:     "fatal",
			expected:     true,
		},
		{
			name:         "Error should be logged at warn level",
			detectorLevel: log.WarnLevel,
			logLevel:     "error",
			expected:     true,
		},
		{
			name:         "Warn should be logged at warn level",
			detectorLevel: log.WarnLevel,
			logLevel:     "warn",
			expected:     true,
		},
		{
			name:         "Info should NOT be logged at warn level",
			detectorLevel: log.WarnLevel,
			logLevel:     "info",
			expected:     false,
		},
		{
			name:         "Debug should NOT be logged at warn level",
			detectorLevel: log.WarnLevel,
			logLevel:     "debug",
			expected:     false,
		},
		{
			name:         "Debug should be logged at debug level",
			detectorLevel: log.DebugLevel,
			logLevel:     "debug",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		detector := NewGinDetector(tt.detectorLevel)
			result := detector.ShouldLogLevel(tt.logLevel)
			if result != tt.expected {
				t.Errorf("ShouldLogLevel(%s) = %v, expected %v", tt.logLevel, result, tt.expected)
			}
		})
	}
}

func TestGinLogDetector_FormatConsoleOutput(t *testing.T) {
	detector := NewGinDetector(log.InfoLevel)
	testTime := time.Date(2023, 12, 7, 15, 4, 5, 123000000, time.UTC)

	event := &GinLogEvent{
		Level:     "info",
		Timestamp: testTime,
		Prefix:    "GIN",
		Message:   "Test message",
		Error:     "Test error",
	}

	output := detector.FormatConsoleOutput(event)

	// Should contain level, timestamp, prefix, message, and error
	if !strings.Contains(output, "INF") {
		t.Error("Output should contain formatted level")
	}
	if !strings.Contains(output, "15:04:05.123") {
		t.Error("Output should contain timestamp")
	}
	if !strings.Contains(output, "GIN") {
		t.Error("Output should contain prefix")
	}
	if !strings.Contains(output, "Test message") {
		t.Error("Output should contain message")
	}
	if !strings.Contains(output, "Test error") {
		t.Error("Output should contain error")
	}
	if !strings.Contains(output, "|") {
		t.Error("Output should contain pipe separators")
	}
}

func TestGinLogDetector_ToJSON(t *testing.T) {
	detector := NewGinDetector(log.InfoLevel)
	testTime := time.Date(2023, 12, 7, 15, 4, 5, 0, time.UTC)

	event := &GinLogEvent{
		Level:         "info",
		Timestamp:     testTime,
		Prefix:        "GIN",
		CorrelationID: "test-correlation-id",
		Message:       "Test message",
		Error:         "Test error",
	}

	jsonBytes, err := detector.ToJSON(event)
	if err != nil {
		t.Fatalf("ToJSON() returned error: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Check that all fields are present in JSON
	if !strings.Contains(jsonStr, "\"level\":\"info\"") {
		t.Error("JSON should contain level field")
	}
	if !strings.Contains(jsonStr, "\"prefix\":\"GIN\"") {
		t.Error("JSON should contain prefix field")
	}
	if !strings.Contains(jsonStr, "\"correlationid\":\"test-correlation-id\"") {
		t.Error("JSON should contain correlationid field")
	}
	if !strings.Contains(jsonStr, "\"message\":\"Test message\"") {
		t.Error("JSON should contain message field")
	}
	if !strings.Contains(jsonStr, "\"error\":\"Test error\"") {
		t.Error("JSON should contain error field")
	}
	if !strings.Contains(jsonStr, "\"time\":\"2023-12-07T15:04:05Z\"") {
		t.Error("JSON should contain time field")
	}
}

func TestGinLogDetector_LevelPrint(t *testing.T) {
	tests := []struct {
		level    string
		colour   bool
		expected string
	}{
		{"fatal", false, "FTL"},
		{"error", false, "ERR"},
		{"warn", false, "WRN"},
		{"info", false, "INF"},
		{"debug", false, "DBG"},
		{"unknown", false, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			result := levelprint(tt.level, tt.colour)
			if !tt.colour && result != tt.expected {
				t.Errorf("levelprint(%s, %v) = %v, expected %v", tt.level, tt.colour, result, tt.expected)
			}
			if tt.colour && !strings.Contains(result, tt.expected) {
				t.Errorf("levelprint(%s, %v) = %v, should contain %v", tt.level, tt.colour, result, tt.expected)
			}
		})
	}
}

func TestGinLogDetector(t *testing.T) {
	detector := NewGinDetector(log.WarnLevel)
	
	if detector == nil {
		t.Error("GinLogDetector should return a non-nil detector")
	}
	
	if detector.level != log.WarnLevel {
		t.Errorf("GinLogDetector level = %v, expected %v", detector.level, log.WarnLevel)
	}
}
