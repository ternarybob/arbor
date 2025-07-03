package arbor

import (
	"fmt"
	"github.com/phuslu/log"
	"strings"
)

// Level type for arbor logging levels
type Level = log.Level

// Level constants that mirror phuslu/log levels
const (
	// TraceLevel defines trace log level.
	TraceLevel Level = log.TraceLevel
	// DebugLevel defines debug log level.
	DebugLevel Level = log.DebugLevel
	// InfoLevel defines info log level.
	InfoLevel Level = log.InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel Level = log.WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel Level = log.ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel Level = log.FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel Level = log.PanicLevel
	// Disabled disables the logger.
	Disabled Level = log.PanicLevel + 1
)

// ParseLevel converts a level string to a Level value.
func ParseLevel(levelStr string) (Level, error) {
	switch strings.ToLower(levelStr) {
	case "trace":
		return TraceLevel, nil
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "panic":
		return PanicLevel, nil
	case "disabled", "off":
		return Disabled, nil
	default:
		return InfoLevel, fmt.Errorf("unknown level: %s", levelStr)
	}
}
