package arbor

import (
	"testing"
	"github.com/phuslu/log"
)

func TestTimestampFormat(t *testing.T) {
	// Create a simple logger with short timestamp format
	testLogger := log.Logger{
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
		Writer: &log.ConsoleWriter{
			ColorOutput:    true,
			EndWithMessage: true,
		},
	}
	
	t.Log("Testing timestamp format - should show HH:MM:SS.sss format")
	testLogger.Info().Msg("This message should have short timestamp format: HH:MM:SS.sss")
	testLogger.Warn().Msg("This warning should also have short timestamp format")
	testLogger.Error().Msg("This error should also have short timestamp format")
}
