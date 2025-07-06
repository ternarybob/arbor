package arbor

import (
	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/levels"
)

// Re-export LogLevel type and constants from levels subpackage
// This allows users to use arbor.LogLevel instead of levels.LogLevel

type LogLevel = levels.LogLevel

const (
	TraceLevel LogLevel = levels.TraceLevel
	DebugLevel LogLevel = levels.DebugLevel
	InfoLevel  LogLevel = levels.InfoLevel
	WarnLevel  LogLevel = levels.WarnLevel
	ErrorLevel LogLevel = levels.ErrorLevel
	FatalLevel LogLevel = levels.FatalLevel
	PanicLevel LogLevel = levels.PanicLevel
	Disabled   LogLevel = levels.Disabled
)

// Re-export convenience functions from levels subpackage
func ParseLevelString(levelStr string) (log.Level, error) {
	return levels.ParseLevelString(levelStr)
}

func ParseLogLevel(level int) log.Level {
	return levels.ParseLogLevel(level)
}
