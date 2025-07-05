package arbor

import (
	"fmt"
	"strings"

	"github.com/phuslu/log"
)

type LogLevel uint32

const (
	TraceLevel LogLevel = LogLevel(log.TraceLevel)
	DebugLevel LogLevel = LogLevel(log.DebugLevel)
	InfoLevel  LogLevel = LogLevel(log.InfoLevel)
	WarnLevel  LogLevel = LogLevel(log.WarnLevel)
	ErrorLevel LogLevel = LogLevel(log.ErrorLevel)
	FatalLevel LogLevel = LogLevel(log.FatalLevel)
	PanicLevel LogLevel = LogLevel(log.PanicLevel)
	Disabled   LogLevel = LogLevel(0)
)

// ParseLevel converts a level string to a Level value.
func ParseLevelString(levelStr string) (log.Level, error) {
	switch strings.ToLower(levelStr) {
	case "trace":
		return log.TraceLevel, nil
	case "debug":
		return log.DebugLevel, nil
	case "info":
		return log.InfoLevel, nil
	case "warn", "warning":
		return log.WarnLevel, nil
	case "error":
		return log.ErrorLevel, nil
	case "fatal":
		return log.FatalLevel, nil
	case "panic":
		return log.PanicLevel, nil
	case "disabled", "off":
		return log.PanicLevel + 1, nil
	default:
		return log.InfoLevel, fmt.Errorf("unknown level: %s", levelStr)
	}
}

func ParseLogLevel(level int) log.Level {
	switch LogLevel(level) {
	case TraceLevel:
		return log.TraceLevel
	case DebugLevel:
		return log.DebugLevel
	case InfoLevel:
		return log.InfoLevel
	case WarnLevel:
		return log.WarnLevel
	case ErrorLevel:
		return log.ErrorLevel
	case FatalLevel:
		return log.FatalLevel
	case PanicLevel:
		return log.PanicLevel
	case Disabled:
		return 0
	default:
		return log.InfoLevel
	}
}
