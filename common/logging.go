package common

import (
	"strings"

	"github.com/phuslu/log"
)

// LevelTo3Letter converts a phuslu log.Level to a 3-letter uppercase string.
func LevelTo3Letter(level log.Level) string {
	switch level {
	case log.TraceLevel:
		return "TRC"
	case log.DebugLevel:
		return "DBG"
	case log.InfoLevel:
		return "INF"
	case log.WarnLevel:
		return "WRN"
	case log.ErrorLevel:
		return "ERR"
	case log.FatalLevel:
		return "FTL"
	case log.PanicLevel:
		return "PNC"
	default:
		return "UNK"
	}
}

// LevelStringTo3Letter converts a string log level (e.g. "info", "warning") to a 3-letter uppercase string.
func LevelStringTo3Letter(level string) string {
	switch strings.ToLower(level) {
	case "trace":
		return "TRC"
	case "debug":
		return "DBG"
	case "info":
		return "INF"
	case "warn", "warning":
		return "WRN"
	case "error":
		return "ERR"
	case "fatal":
		return "FTL"
	case "panic":
		return "PNC"
	default:
		// If it's already 3 letters or unknown, just uppercase it
		if len(level) == 3 {
			return strings.ToUpper(level)
		}
		return "UNK"
	}
}
