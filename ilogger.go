package arbor

import (
	"time"

	"github.com/ternarybob/arbor/models"
	"github.com/ternarybob/arbor/writers"
)

type ILogger interface {
	// Deprecated: Use SetChannel("context", ch) instead. This method will be removed in a future version.
	// SetContextChannel internally calls SetChannel with a fixed name "context".
	SetContextChannel(ch chan []models.LogEvent)

	// Deprecated: Use SetChannelWithBuffer("context", ch, batchSize, flushInterval) instead. This method will be removed in a future version.
	// SetContextChannelWithBuffer internally calls SetChannelWithBuffer with a fixed name "context".
	SetContextChannelWithBuffer(ch chan []models.LogEvent, batchSize int, flushInterval time.Duration)

	SetChannel(name string, ch chan []models.LogEvent)
	SetChannelWithBuffer(name string, ch chan []models.LogEvent, batchSize int, flushInterval time.Duration)
	UnregisterChannel(name string)
	WithContextWriter(contextID string) ILogger
	WithWriters(writers []writers.IWriter) ILogger
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

	// GetLogFilePath returns the configured log file path if a file writer is registered
	GetLogFilePath() string
}
