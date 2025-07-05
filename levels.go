package arbor

import (
	"fmt"
	"strings"

	"github.com/ternarybob/arbor/interfaces"
)

// Re-export Level type and constants from interfaces package for backward compatibility
type Level = interfaces.Level

const (
	// TraceLevel defines trace log level.
	TraceLevel = interfaces.TraceLevel
	// DebugLevel defines debug log level.
	DebugLevel = interfaces.DebugLevel
	// InfoLevel defines info log level.
	InfoLevel = interfaces.InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel = interfaces.WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel = interfaces.ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel = interfaces.FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel = interfaces.PanicLevel
	// Disabled disables the logger.
	Disabled = interfaces.Disabled
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
