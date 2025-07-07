package arbor

import (
	"github.com/ternarybob/arbor/models"
)

type ILogger interface {
	// Write(p []byte) (n int, err error)

	WithPrefix(value string) ILogger

	WithCorrelationId(value string) ILogger

	WithLevel(lvl LogLevel) ILogger

	WithContext(key string, value string) ILogger

	WithFileWriter(config models.WriterConfiguration) ILogger

	WithMemoryWriter(config models.WriterConfiguration) ILogger

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
}
