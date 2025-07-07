package arbor

import (
	"runtime"
	"strings"

	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
	"github.com/ternarybob/arbor/writers"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/phuslu/log"
)

const (
	LOGGER_CONTEXT_KEY string = "logger"
	LEVEL_KEY          string = "level"
	CORRELATION_ID_KEY string = "correlationid"
	PREFIX_KEY         string = "prefix"
	WRITER_CONSOLE     string = "writerconsole"
	WRITER_FILE        string = "writerfile"
	WRITER_MEMORY      string = "writermemory"
	GIN_LOG_KEY        string = "gin"
)

var (
	internallog log.Logger = log.Logger{
		Level:  log.WarnLevel,
		Writer: &log.ConsoleWriter{},
	}
	copieropts           copier.Option              = copier.Option{IgnoreEmpty: true, DeepCopy: false}
	defaultLoggerOptions models.WriterConfiguration = models.WriterConfiguration{
		Type:       models.LogWriterTypeConsole,
		Level:      levels.InfoLevel,
		TimeFormat: "01-02 15:04:05.000",
	}
)

// logger is the main arbor logger implementation that supports multiple writers
type logger struct {
	writers     map[string]writers.IWriter
	contextData map[string]string // Track context key-value pairs
}

// Logger creates a new arbor logger with console writer as default
func Logger() ILogger {

	var (
		configuredWriters map[string]writers.IWriter = make(map[string]writers.IWriter)
	)

	// Add Console Writer
	configuredWriters[WRITER_CONSOLE] = writers.ConsoleWriter(defaultLoggerOptions)

	return &logger{
		writers:     configuredWriters,
		contextData: make(map[string]string),
	}

}

func (l *logger) WithCorrelationId(correlationID string) ILogger {

	if common.IsEmpty(correlationID) {

		uuid, err := uuid.NewRandom()
		if err != nil {
			internallog.Warn().Err(err).Msg("")
		}

		correlationID = uuid.String()
	}

	internallog.Trace().Msgf("Adding CorrelationId -> CorrelationId:%s", correlationID)

	l.WithContext(CORRELATION_ID_KEY, correlationID)

	return l

}

func (l *logger) WithPrefix(value string) ILogger {

	if common.IsEmpty(value) {
		return l
	}

	internallog.Trace().Msgf("Replacing Prefix:%s", value)

	l.WithContext(PREFIX_KEY, value)

	return l
}

func (l *logger) WithLevel(level LogLevel) ILogger {

	internallog.Trace().Msg("Iterating over writers")

	// Ensure writers map exists before iterating
	if len(l.writers) == 0 {
		internallog.Trace().Msg("No writers configured.")
		return l
	}

	lvl := ParseLogLevel(int(level))

	for key, writer := range l.writers {
		internallog.Trace().Msgf("Key: \"%s\", Writer Type: %T\n", key, writer)
		writer.WithLevel(lvl)
	}

	return l

}

func (l *logger) WithContext(key string, value string) ILogger {
	if common.IsEmpty(key) || common.IsEmpty(value) {
		internallog.Trace().Msgf("Key or Value empty -> returning")
		return l
	}

	// Ensure contextData map is initialized
	if l.contextData == nil {
		l.contextData = make(map[string]string)
		internallog.Trace().Msg("contextData was nil, initialized it.")
	}

	// Check if the specific 'key' already exists
	if _, exists := l.contextData[key]; exists {
		l.contextData[key] = value
		internallog.Trace().Msgf("Updated existing context key '%s' to: %s", key, value)
	} else {
		l.contextData[key] = value
		internallog.Trace().Msgf("Added new context key '%s' with value: %s", key, value)
	}

	return l
}

func (l *logger) WithFileWriter(configuration models.WriterConfiguration) ILogger {

	// 1. Check if contextData is nil (not yet created/initialized).
	if l.writers == nil {
		l.writers = make(map[string]writers.IWriter)
		internallog.Trace().Msg("writers was nil, initialized it.")
	}

	// Add File Writer
	l.writers[WRITER_FILE] = writers.FileWriter(configuration)

	return l

}

func (l *logger) WithMemoryWriter(configuration models.WriterConfiguration) ILogger {

	// 1. Check if writers is nil (not yet created/initialized).
	if l.writers == nil {
		l.writers = make(map[string]writers.IWriter)
		internallog.Trace().Msg("writers was nil, initialized it.")
	}

	// Add Memory Writer
	l.writers[WRITER_MEMORY] = writers.MemoryWriter(configuration)

	return l

}

// func (l *logger) Write(p []byte) (n int, err error) {

// 	return n, nil

// }

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

// Global logger instance
var defaultLogger ILogger

// init initializes the default logger
func init() {
	defaultLogger = Logger()
}

// GetLogger returns the default logger instance
func GetLogger() ILogger {
	return defaultLogger
}

// Global convenience functions for direct logging
func Trace() ILogEvent {
	return defaultLogger.Trace()
}

func Debug() ILogEvent {
	return defaultLogger.Debug()
}

func Info() ILogEvent {
	return defaultLogger.Info()
}

func Warn() ILogEvent {
	return defaultLogger.Warn()
}

func Error() ILogEvent {
	return defaultLogger.Error()
}

func Fatal() ILogEvent {
	return defaultLogger.Fatal()
}

func Panic() ILogEvent {
	return defaultLogger.Panic()
}

func (l *logger) GetMemoryLogs(correlationid string, minLevel LogLevel) (map[string]string, error) {
	// Check if memory writer is configured
	if l.writers == nil {
		return make(map[string]string), nil
	}

	memoryWriter, hasMemoryWriter := l.writers[WRITER_MEMORY]
	if !hasMemoryWriter {
		return make(map[string]string), nil
	}

	// Cast to IMemoryWriter and call the method
	if mw, ok := memoryWriter.(writers.IMemoryWriter); ok {
		return mw.GetEntriesWithLevel(correlationid, minLevel.ToLogLevel())
	}

	return make(map[string]string), nil
}

func (l *logger) GetMemoryLogsForCorrelation(correlationid string) (map[string]string, error) {
	// Check if memory writer is configured
	if l.writers == nil {
		return make(map[string]string), nil
	}

	memoryWriter, hasMemoryWriter := l.writers[WRITER_MEMORY]
	if !hasMemoryWriter {
		return make(map[string]string), nil
	}

	// Cast to IMemoryWriter and call the method
	if mw, ok := memoryWriter.(writers.IMemoryWriter); ok {
		return mw.GetEntries(correlationid)
	}

	return make(map[string]string), nil
}

func (l *logger) GetMemoryLogsWithLimit(limit int) (map[string]string, error) {
	// Check if memory writer is configured
	if l.writers == nil {
		return make(map[string]string), nil
	}

	memoryWriter, hasMemoryWriter := l.writers[WRITER_MEMORY]
	if !hasMemoryWriter {
		return make(map[string]string), nil
	}

	// Cast to IMemoryWriter and call the method
	if mw, ok := memoryWriter.(writers.IMemoryWriter); ok {
		return mw.GetEntriesWithLimit(limit)
	}

	return make(map[string]string), nil
}
