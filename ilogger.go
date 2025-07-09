package arbor

import (
	"github.com/ternarybob/arbor/models"
)

type ILogger interface {
	WithConsoleWriter(config models.WriterConfiguration) ILogger

	WithFileWriter(config models.WriterConfiguration) ILogger

	WithMemoryWriter(config models.WriterConfiguration) ILogger

	WithPrefix(value string) ILogger

	WithCorrelationId(value string) ILogger

	ClearCorrelationId() ILogger

	// ClearContext removes all context data from the logger
	ClearContext() ILogger

	WithLevel(lvl LogLevel) ILogger

	// WithLevelFromString applies a log level from a string configuration
	WithLevelFromString(levelStr string) ILogger

	WithContext(key string, value string) ILogger

	// Copy creates a copy of the logger with the same configuration but clean/empty context
	// This is useful when you want a fresh logger that shares the same writers but has no correlation ID, prefix, or other context
	Copy() ILogger

	// Fluent logging methods
	Trace() ILogEvent
	Debug() ILogEvent
	Info() ILogEvent
	Warn() ILogEvent
	Error() ILogEvent
	Fatal() ILogEvent
	Panic() ILogEvent

	GetMemoryLogs(correlationid string, minLevel LogLevel) (map[string]string, error)

	// GetMemoryLogsForCorrelation retrieves all log entries for a specific correlation ID
	GetMemoryLogsForCorrelation(correlationid string) (map[string]string, error)

	// GetMemoryLogsWithLimit retrieves the most recent log entries up to the specified limit
	GetMemoryLogsWithLimit(limit int) (map[string]string, error)

	// GinWriter returns an io.Writer that integrates Gin logs with arbor's registered writers
	GinWriter(config models.WriterConfiguration) interface{}
}
