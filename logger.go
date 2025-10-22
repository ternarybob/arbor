package arbor

import (
	"runtime"
	"strings"
	"time"

	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/models"
	"github.com/ternarybob/arbor/transformers"
	"github.com/ternarybob/arbor/writers"

	"github.com/google/uuid"
	"github.com/phuslu/log"
)

const (
	LOGGER_CONTEXT_KEY string = "logger"
	LEVEL_KEY          string = "level"
	CORRELATION_ID_KEY string = "correlationid"
	PREFIX_KEY         string = "prefix"
	GIN_LOG_KEY        string = "gin"
)

// logger is the main arbor logger implementation that supports multiple writers
type logger struct {
	writers     []writers.IWriter // Private writers for this logger instance
	contextData map[string]string // Track context key-value pairs
}

// SetContextChannel configures the context logger with a channel for receiving log batches.
// It uses default buffering settings (batch size 5, flush interval 1 second).
func (l *logger) SetContextChannel(ch chan []models.LogEvent) {
	common.Start(ch, 5, 1*time.Second)
}

// SetContextChannelWithBuffer configures the context logger with custom buffering settings.
func (l *logger) SetContextChannelWithBuffer(ch chan []models.LogEvent, batchSize int, flushInterval time.Duration) {
	common.Start(ch, batchSize, flushInterval)
}

// ForContext creates a logger for a specific context (e.g., a job ID).
// It logs to both the standard writers and the configured context channel.
func (l *logger) ForContext(contextID string) ILogger {
	// Create a new writer that will send logs to the context buffer.
	contextWriter := writers.NewContextWriter()

	// Get all the existing global writers.
	globalWriters := GetAllRegisteredWriters()
	newWriters := make([]writers.IWriter, 0, len(globalWriters)+1)
	for _, writer := range globalWriters {
		newWriters = append(newWriters, writer)
	}
	newWriters = append(newWriters, contextWriter)

	// Create a copy of the logger and configure it for the context.
	contextLogger := l.Copy().
		WithCorrelationId(contextID). // Use correlation ID to tag context logs.
		WithWriters(newWriters)

	return contextLogger
}

var (
	defaultLog ILogger
)

// Logger returns the default logger instance, creating it if it doesn't exist
func Logger() ILogger {
	if defaultLog == nil {
		defaultLog = createNewLogger()
	}
	return defaultLog
}

// NewLogger creates a new logger instance
// This is useful for testing or when you need isolated logger instances
func NewLogger() ILogger {
	return createNewLogger()
}

// createNewLogger is a helper function that creates a fresh logger instance
func createNewLogger() ILogger {
	// Create logger that will use registered writers
	// Start with INFO level as default before configuration is loaded
	logger := &logger{
		contextData: make(map[string]string),
	}
	logger.WithLevel(InfoLevel) // Initial level
	return logger
}

func (l *logger) WithWriters(writers []writers.IWriter) ILogger {
	l.writers = writers
	return l
}

func (l *logger) WithConsoleWriter(configuration models.WriterConfiguration) ILogger {

	internalLog := common.NewLogger().WithContext("function", "Logger.WithConsoleWriter").GetLogger()

	// Create and register the console writer
	consoleWriter := writers.ConsoleWriter(configuration)
	RegisterWriter(WRITER_CONSOLE, consoleWriter)

	internalLog.Trace().Msg("Console writer registered successfully.")

	return l

}

func (l *logger) WithFileWriter(configuration models.WriterConfiguration) ILogger {

	internalLog := common.NewLogger().WithContext("function", "Logger.WithFileWriter").GetLogger()

	// Create and register the file writer
	fileWriter := writers.FileWriter(configuration)
	RegisterWriter(WRITER_FILE, fileWriter)

	internalLog.Trace().Msg("File writer registered successfully.")

	return l

}

func (l *logger) WithMemoryWriter(configuration models.WriterConfiguration) ILogger {

	internalLog := common.NewLogger().WithContext("function", "Logger.WithMemoryWriter").GetLogger()

	// Create the memory writer (which creates the underlying store)
	memoryWriter := writers.MemoryWriter(configuration)

	// Get the store from the memory writer
	store := memoryWriter.GetStore()

	// Create a LogStoreWriter to handle actual writes to the store
	logStoreWriter := writers.LogStoreWriter(store, configuration)

	// Register both:
	// - LogStoreWriter handles writing log events to the store
	// - MemoryWriter provides query interface
	RegisterWriter(WRITER_MEMORY+"_store", logStoreWriter)
	RegisterWriter(WRITER_MEMORY, memoryWriter)

	internalLog.Trace().Msg("Memory writer and log store registered successfully.")

	return l

}

func (l *logger) WithCorrelationId(correlationID string) ILogger {

	internalLog := common.NewLogger().WithContext("function", "Logger.WithCorrelationId").GetLogger()

	if common.IsEmpty(correlationID) {

		uuid, err := uuid.NewRandom()
		if err != nil {
			internalLog.Warn().Err(err).Msg("")
		}

		correlationID = uuid.String()
	}

	internalLog.Debug().Msgf("Adding CorrelationId -> CorrelationId:%s", correlationID)

	l.WithContext(CORRELATION_ID_KEY, correlationID)

	return l

}

// ClearCorrelationId removes the correlation ID from the logger context
func (l *logger) ClearCorrelationId() ILogger {
	internalLog := common.NewLogger().WithContext("function", "Logger.ClearCorrelationId").GetLogger()

	// Remove the correlation ID from context data
	if l.contextData != nil {
		delete(l.contextData, CORRELATION_ID_KEY)
		internalLog.Debug().Msg("Cleared correlation ID from logger context")
	}

	return l
}

// ClearContext removes all context data from the logger
func (l *logger) ClearContext() ILogger {
	internalLog := common.NewLogger().WithContext("function", "Logger.ClearContext").GetLogger()

	// Clear all context data
	if l.contextData != nil {
		l.contextData = make(map[string]string)
		internalLog.Debug().Msg("Cleared all context data from logger")
	}

	return l
}

func (l *logger) WithPrefix(value string) ILogger {

	internalLog := common.NewLogger().WithContext("function", "Logger.WithPrefix").GetLogger()

	if common.IsEmpty(value) {
		return l
	}

	internalLog.Trace().Msgf("Replacing Prefix:%s", value)

	l.WithContext(PREFIX_KEY, value)

	return l
}

// WithLevelFromString applies a log level from a string configuration
func (l *logger) WithLevelFromString(levelStr string) ILogger {
	internalLog := common.NewLogger().WithContext("function", "Logger.WithLevelFromString").GetLogger()

	// Parse and apply log level from string
	phusluLevel, err := ParseLevelString(levelStr)
	if err != nil {
		internalLog.Warn().Err(err).Msgf("Invalid log level '%s', using INFO", levelStr)
		l.WithLevel(InfoLevel)
	} else {
		// Convert phuslu log.Level to our LogLevel
		arborLevel := LogLevel(phusluLevel)
		l.WithLevel(arborLevel)
		internalLog.Debug().Msgf("Set log level to: %s", levelStr)
	}

	return l
}

func (l *logger) WithLevel(level LogLevel) ILogger {

	internalLog := common.NewLogger().WithContext("function", "Logger.WithLevel").GetLogger()
	lvl := ParseLogLevel(int(level))

	// If the logger has its own writers, use them. Otherwise, use the global registry.
	if l.writers != nil {
		internalLog.Trace().Msg("Iterating over private writers")
		for _, writer := range l.writers {
			writer.WithLevel(lvl)
		}
	} else {
		internalLog.Trace().Msg("Iterating over registered writers")
		// Get all registered writers
		registeredWriters := GetAllRegisteredWriters()
		if len(registeredWriters) == 0 {
			internalLog.Trace().Msg("No writers registered.")
			return l
		}

		for key, writer := range registeredWriters {
			internalLog.Trace().Msgf("Key: \"%s\", Writer Type: %T\n", key, writer)
			writer.WithLevel(lvl)
		}
	}

	return l

}

func (l *logger) WithContext(key string, value string) ILogger {

	internalLog := common.NewLogger().WithContext("function", "Logger.WithContext").GetLogger()

	if common.IsEmpty(key) || common.IsEmpty(value) {
		internalLog.Trace().Msgf("Key or Value empty -> returning")
		return l
	}

	// Ensure contextData map is initialized
	if l.contextData == nil {
		l.contextData = make(map[string]string)
		internalLog.Trace().Msg("contextData was nil, initialized it.")
	}

	// Check if the specific 'key' already exists
	if _, exists := l.contextData[key]; exists {
		l.contextData[key] = value
		internalLog.Trace().Msgf("Updated existing context key '%s' to: %s", key, value)
	} else {
		l.contextData[key] = value
		internalLog.Trace().Msgf("Added new context key '%s' with value: %s", key, value)
	}

	return l
}

// Copy creates a copy of the logger with the same configuration but fresh/clean context
// This is useful when you want a logger with the same writers but without any context data
func (l *logger) Copy() ILogger {
	// Create a new logger instance with fresh context (no context data copied)
	newLogger := createNewLogger()

	return newLogger
}

func (d *logger) getFunctionName() string {
	// Try different caller depths to find the actual calling function
	// Skip arbor package methods and get to the caller's function
	for i := 1; i <= 10; i++ {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		funcName := fn.Name()
		// Skip internal arbor methods, runtime, and testing functions
		if !strings.Contains(funcName, "arbor.") &&
			!strings.Contains(funcName, "runtime.") &&
			!strings.Contains(funcName, "testing.") &&
			funcName != "" {
			return funcName
		}
	}

	// If we can't find a good function name, return empty
	return ""
}

// Fluent logging methods
func (l *logger) Trace() ILogEvent {
	return newLogEvent(l, log.TraceLevel)
}

func (l *logger) Debug() ILogEvent {
	return newLogEvent(l, log.DebugLevel)
}

func (l *logger) Info() ILogEvent {
	return newLogEvent(l, log.InfoLevel)
}

func (l *logger) Warn() ILogEvent {
	return newLogEvent(l, log.WarnLevel)
}

func (l *logger) Error() ILogEvent {
	return newLogEvent(l, log.ErrorLevel)
}

func (l *logger) Fatal() ILogEvent {
	return newLogEvent(l, log.FatalLevel)
}

func (l *logger) Panic() ILogEvent {
	return newLogEvent(l, log.PanicLevel)
}

// GetLogger returns the default logger instance from the registry
func GetLogger() ILogger {
	return Logger()
}

// Global convenience functions for direct logging
func Trace() ILogEvent {
	return GetLogger().Trace()
}

func Debug() ILogEvent {
	return GetLogger().Debug()
}

func Info() ILogEvent {
	return GetLogger().Info()
}

func Warn() ILogEvent {
	return GetLogger().Warn()
}

func Error() ILogEvent {
	return GetLogger().Error()
}

func Fatal() ILogEvent {
	return GetLogger().Fatal()
}

func Panic() ILogEvent {
	return GetLogger().Panic()
}

func (l *logger) GetMemoryLogsForCorrelation(correlationid string) (map[string]string, error) {
	return l.GetMemoryLogs(correlationid, TraceLevel)
}

func (l *logger) GetMemoryLogs(correlationid string, minLevel LogLevel) (map[string]string, error) {

	internalLog := common.NewLogger().WithContext("function", "Logger.GetMemoryLogs").GetLogger()

	internalLog.Context = log.NewContext(nil).Str("function", "GetMemoryLogs").Value()

	// Get memory writer from registry
	memoryWriter := GetRegisteredMemoryWriter(WRITER_MEMORY)
	if memoryWriter == nil {
		internalLog.Warn().Msg("Memory writer not registered -> return")
		return make(map[string]string), nil
	}

	internalLog.Debug().Msg("Getting Memory writer entries -> GetEntriesWithLevel")
	return memoryWriter.GetEntriesWithLevel(correlationid, minLevel.ToLogLevel())
}

func (l *logger) GetMemoryLogsWithLimit(limit int) (map[string]string, error) {

	internalLog := common.NewLogger().WithContext("function", "Logger.GetMemoryLogsWithLimit").GetLogger()

	// Get memory writer from registry
	memoryWriter := GetRegisteredMemoryWriter(WRITER_MEMORY)
	if memoryWriter == nil {
		internalLog.Warn().Msg("Memory writer not registered -> return")
		return make(map[string]string), nil
	}

	internalLog.Debug().Msgf("Getting Memory writer entries -> GetEntriesWithLimit(%v)", limit)
	return memoryWriter.GetEntriesWithLimit(limit)
}

func (l *logger) GinWriter(config models.WriterConfiguration) interface{} {
	internalLog := common.NewLogger().WithContext("function", "Logger.GinWriter").GetLogger()

	// Create Gin transformer with provided configuration and registry function
	ginTransformer := transformers.NewGinTransformer(config, GetAllRegisteredWriters)
	internalLog.Debug().Msg("Created Gin transformer")

	return ginTransformer
}

// GetLogFilePath returns the configured log file path if a file writer is registered
func (l *logger) GetLogFilePath() string {
	// Get file writer from registry
	fileWriter := GetRegisteredWriter(WRITER_FILE)
	if fileWriter == nil {
		return "" // No file writer registered
	}

	// Use the GetFilePath method from the writer interface
	return fileWriter.GetFilePath()
}
