package arbor

import (
	"github.com/ternarybob/arbor/models"
	"testing"
)

func TestLogger_WithLevelFromString(t *testing.T) {
	tests := []struct {
		name        string
		levelString string
		expectError bool
	}{
		{"Valid trace level", "trace", false},
		{"Valid debug level", "debug", false},
		{"Valid info level", "info", false},
		{"Valid warn level", "warn", false},
		{"Valid error level", "error", false},
		{"Valid fatal level", "fatal", false},
		{"Valid panic level", "panic", false},
		{"Valid disabled level", "disabled", false},
		{"Case insensitive", "INFO", false},
		{"Invalid level", "invalid", true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a logger with console writer for testing
			logger := NewLogger().WithConsoleWriter(models.WriterConfiguration{
				Type: models.LogWriterTypeConsole,
			})

			// Apply the level from string
			result := logger.WithLevelFromString(tt.levelString)

			// The method should always return a logger instance
			if result == nil {
				t.Error("WithLevelFromString should never return nil")
			}

			// Test that the logger can still be used after level setting
			result.Info().Msg("Test message after level setting")
		})
	}
}

func TestLogger_WithLevelFromString_ChainedUsage(t *testing.T) {
	// Test that WithLevelFromString can be chained with other methods
	logger := NewLogger().
		WithConsoleWriter(models.WriterConfiguration{
			Type: models.LogWriterTypeConsole,
		}).
		WithLevelFromString("debug").
		WithCorrelationId("test-correlation").
		WithPrefix("TEST")

	// The logger should be functional after chaining
	logger.Debug().Msg("This is a debug message")
	logger.Info().Msg("This is an info message")
	logger.Warn().Msg("This is a warning message")
}
