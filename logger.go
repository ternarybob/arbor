package arbor

import (
	"runtime"
	"strings"

	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/writers"

	"github.com/ternarybob/arbor/models"

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

// type LogLevel interfaces.LogLevel

// const (
// 	TraceLevel = common.TraceLevel
// 	DebugLevel = common.DebugLevel
// 	InfoLevel  = common.InfoLevel
// 	WarnLevel  = common.WarnLevel
// 	ErrorLevel = common.ErrorLevel
// 	FatalLevel = common.FatalLevel
// 	PanicLevel = common.PanicLevel
// 	Disabled   = common.Disabled
// )

var (
	internallog log.Logger = log.Logger{
		Level:  log.WarnLevel,
		Writer: &log.ConsoleWriter{},
	}
	copieropts           copier.Option              = copier.Option{IgnoreEmpty: true, DeepCopy: false}
	defaultLoggerOptions models.WriterConfiguration = models.WriterConfiguration{
		Type:       models.LogWriterTypeConsole,
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
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
	if l.writers == nil || len(l.writers) == 0 {
		internallog.Trace().Msg("No writers configured or writers map is nil.")
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

	internallog.Trace().Msgf("Adding key:value %s:%s", key, value)

	// 1. Check if contextData is nil (not yet created/initialized).
	if l.contextData == nil {
		l.contextData = make(map[string]string)
		internallog.Trace().Msg("contextData was nil, initialized it.")
	}

	// 2. Check if the CORRELATION_ID_KEY already exists in contextData.
	if _, exists := l.contextData[CORRELATION_ID_KEY]; exists {
		// If it exists, update the value.
		l.contextData[key] = value
		internallog.Trace().Msgf("Updated existing %s to: %s\n", key, value)
	} else {
		// If it does not exist, add the new key-value pair.
		l.contextData[key] = value
		internallog.Trace().Msgf("Added new %s with value: %s\n", key, value)
	}

	return l
}

func (l *logger) WithFileWriter(configuration models.WriterConfiguration) ILogger {

	// 1. Check if contextData is nil (not yet created/initialized).
	if l.writers == nil {
		l.writers = make(map[string]writers.IWriter)
		internallog.Trace().Msg("writers was nil, initialized it.")
	}

	// Add Console Writer
	l.writers[WRITER_FILE] = writers.FileWriter(configuration)

	return l

}
func (l *logger) Write(p []byte) (n int, err error) {

	return n, nil

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

// func (d *logger) GetMemoryLogs(correlationid string, minLevel Level) (map[string]string, error) {
// 	return writers.GetEntriesWithLevel(correlationid, minLevel)
// }
