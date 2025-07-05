package arbor

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"

	"github.com/ternarybob/arbor/interfaces"
	"github.com/ternarybob/arbor/writers"

	"github.com/ternarybob/satus"

	"github.com/labstack/echo/v4"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/phuslu/log"
)

const (
	CORRELATION_ID_KEY string = "correlationid"
	LOGGERCONTEXT_KEY  string = "consolelogger"
	WRITER_CONSOLE     string = "writerconsole"
	WRITER_DATA        string = "writerdata"
	WRITER_REDIS       string = "writerredis"
	WRITER_ARRAY       string = "writerarray"
)

var (
	internallog log.Logger = log.Logger{
		Level:  log.WarnLevel,
		Writer: &log.ConsoleWriter{},
	}
	copieropts copier.Option = copier.Option{IgnoreEmpty: true, DeepCopy: false}
)

type consolelogger struct {
	logger        log.Logger
	writers       map[string]io.Writer
	currentPrefix string // Track current prefix for extension
}

func ConsoleLogger() interfaces.IConsoleLogger {

	var (
		cfg          *satus.AppConfig     = satus.GetAppConfig()
		namedwriters map[string]io.Writer = make(map[string]io.Writer)
	)

	loglevel, err := ParseLevel(cfg.Service.LogLevel)
	if err != nil {
		loglevel = InfoLevel
	}

	// Add Writers
	namedwriters[WRITER_CONSOLE] = writers.NewConsoleWriter()

	return &consolelogger{
		logger: log.Logger{
			Level:      loglevel,
			TimeFormat: "15:04:05.000",
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				EndWithMessage: true,
			},
		},
		writers: namedwriters,
	}

}

// func (d *consolelogger) WithRequestContext(ctx *gin.Context) IConsoleLogger {
func (d *consolelogger) WithRequestContext(ctx echo.Context) interfaces.IConsoleLogger {

	var (
		// Use direct logging instead of chained logger
		multiWriters []io.Writer
	)

	if ctx == nil {
		panic(fmt.Errorf("Context not available | nil"))
	}

	correlationid := ctx.Response().Header().Get(echo.HeaderXRequestID)

	if isEmpty(correlationid) {
		internallog.Warn().Msgf("Correlation Key Not Available -> New Logger NOT created")
		return d
	}

	if len(d.writers) == 0 {
		d.writers = make(map[string]io.Writer)
	}

	// Add Writers
	if _, ok := d.writers[WRITER_CONSOLE]; !ok {
		d.writers[WRITER_CONSOLE] = writers.NewConsoleWriter()
	}

	if _, ok := d.writers[WRITER_DATA]; !ok {
		d.writers[WRITER_DATA] = writers.NewMemoryWriter()
	}

	// Add to mulit writer
	for k, v := range d.writers {
		internallog.Debug().Msgf("Adding Writer name:%s type:%s", k, reflect.TypeOf(v))
		multiWriters = append(multiWriters, d.writers[k])
	}

	// Note: simplified to use basic ConsoleWriter since phuslu/log doesn't support MultiWriter like zerolog

	currentlevel := d.GetLevel()

	o := &consolelogger{
		logger: log.Logger{
			Level:      currentlevel,
			TimeFormat: "15:04:05.000",
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				EndWithMessage: true,
			},
		},
		writers: d.writers,
	}

	return o

}

func (d *consolelogger) WithWriter(name string, writer io.Writer) interfaces.IConsoleLogger {

	var (
		multiWriters []io.Writer
	)

	if len(d.writers) == 0 {
		d.writers = make(map[string]io.Writer)
	}

	// Ensure Default Writer
	if _, ok := d.writers[WRITER_CONSOLE]; !ok {
		d.writers[WRITER_CONSOLE] = writers.NewConsoleWriter()
	}

	// Add Writer
	d.writers[name] = writer

	// Add to mulit writer
	for k, v := range d.writers {
		internallog.Debug().Msgf("Adding Writer name:%s type:%s", k, reflect.TypeOf(v))
		multiWriters = append(multiWriters, v)
	}

	// Note: simplified to use basic ConsoleWriter since phuslu/log doesn't support MultiWriter like zerolog

	currentlevel := d.GetLevel()

	o := &consolelogger{
		logger: log.Logger{
			Level:      currentlevel,
			TimeFormat: "15:04:05.000",
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				EndWithMessage: true,
			},
		},
		writers: d.writers,
	}

	return o

}

func (d *consolelogger) WithCorrelationId(correlationid string) interfaces.IConsoleLogger {

	var (
		multiWriters []io.Writer
	)

	if isEmpty(correlationid) {

		uuid, err := uuid.NewRandom()
		if err != nil {
			internallog.Warn().Err(err).Msg("")
		}

		correlationid = uuid.String()
	}

	internallog.Trace().Msgf("Adding CorrelationId -> CorrelationId:%s", correlationid)

	// Ensure writers map exists
	if len(d.writers) == 0 {
		d.writers = make(map[string]io.Writer)
	}

	// Add console writer if not present
	if _, ok := d.writers[WRITER_CONSOLE]; !ok {
		d.writers[WRITER_CONSOLE] = writers.NewConsoleWriter()
	}

	// Add memory writer for correlation ID logging
	if _, ok := d.writers[WRITER_DATA]; !ok {
		d.writers[WRITER_DATA] = writers.NewMemoryWriter()
		internallog.Debug().Msgf("Added memory writer for correlation ID logging")
	}

	// Build multi-writer from all writers
	for k, v := range d.writers {
		internallog.Debug().Msgf("Adding Writer name:%s type:%s", k, reflect.TypeOf(v))
		multiWriters = append(multiWriters, v)
	}

	// Note: simplified to use basic ConsoleWriter since phuslu/log doesn't support MultiWriter like zerolog
	currentLevel := d.GetLevel()

	return &consolelogger{
		logger: log.Logger{
			Level:      currentLevel,
			TimeFormat: "15:04:05.000",
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				EndWithMessage: true,
			},
		},
		writers: d.writers,
	}

}

func (d *consolelogger) GetLevel() Level {
	return d.logger.Level
}

func (d *consolelogger) WithPrefix(value string) interfaces.IConsoleLogger {
	// Create a new logger instance to replace existing prefix
	var (
		multiWriters []io.Writer
	)

	// Ensure writers map exists
	if len(d.writers) == 0 {
		d.writers = make(map[string]io.Writer)
	}

	// Add console writer if not present
	if _, ok := d.writers[WRITER_CONSOLE]; !ok {
		d.writers[WRITER_CONSOLE] = writers.NewConsoleWriter()
	}

	// Build multi-writer from all writers
	for _, v := range d.writers {
		multiWriters = append(multiWriters, v)
	}

	// Note: simplified to use basic ConsoleWriter since phuslu/log doesn't support MultiWriter like zerolog
	currentLevel := d.GetLevel()

	// Create new logger with only the new prefix (replaces existing prefix)
	return &consolelogger{
		logger: log.Logger{
			Level:      currentLevel,
			TimeFormat: "15:04:05.000",
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				EndWithMessage: true,
			},
		},
		writers:       d.writers,
		currentPrefix: value, // Track the new prefix
	}
}

func (d *consolelogger) WithPrefixExtend(value string) interfaces.IConsoleLogger {
	// Use the tracked current prefix to extend it
	var (
		multiWriters []io.Writer
	)

	// Ensure writers map exists
	if len(d.writers) == 0 {
		d.writers = make(map[string]io.Writer)
	}

	// Add console writer if not present
	if _, ok := d.writers[WRITER_CONSOLE]; !ok {
		d.writers[WRITER_CONSOLE] = writers.NewConsoleWriter()
	}

	// Build multi-writer from all writers
	for _, v := range d.writers {
		multiWriters = append(multiWriters, v)
	}

	// Note: simplified to use basic ConsoleWriter since phuslu/log doesn't support MultiWriter like zerolog
	currentLevel := d.GetLevel()

	// Build extended prefix using tracked current prefix
	extendedPrefix := value
	if d.currentPrefix != "" {
		extendedPrefix = d.currentPrefix + "." + value
	}

	// Create new logger with extended prefix
	return &consolelogger{
		logger: log.Logger{
			Level:      currentLevel,
			TimeFormat: "15:04:05.000",
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				EndWithMessage: true,
			},
		},
		writers:       d.writers,
		currentPrefix: extendedPrefix, // Track the extended prefix
	}
}

func (d *consolelogger) WithLevel(lvl Level) interfaces.IConsoleLogger {

	var (
		output = &consolelogger{}
	)

	if err := copier.CopyWithOption(&output, &d, copieropts); err != nil {
		internallog.Warn().Err(err).Msgf("Unable to copy existing service -> reverted to inital")
		return d
	}

	output.logger.Level = lvl

	return output

}

func (d *consolelogger) WithFunction() interfaces.IConsoleLogger {
	// phuslu/log doesn't support runtime modification like this
	return d
}

func (d *consolelogger) WithContext(key string, value string) interfaces.IConsoleLogger {

	// phuslu/log doesn't support runtime modification like this

	return d

}

func (d *consolelogger) WithCorrelationid(value string) {

	// phuslu/log doesn't support runtime modification like this

}

func (d *consolelogger) getFunctionName() string {

	// Assuming called from within the package runtime.Caller(2)
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}

	return fn.Name()

}

func (d *consolelogger) GetLogger() *log.Logger {

	var (
		output = &consolelogger{}
	)

	if err := copier.CopyWithOption(&output, &d, copieropts); err != nil {
		internallog.Warn().Err(err).Msgf("Unable to copy existing service -> reverted to inital")
		return &d.logger
	}

	return &output.logger

}

func (d *consolelogger) WithFileWriterPath(name string, filePath string, bufferSize, maxFiles int) (interfaces.IConsoleLogger, error) {
	// Create file writer with directory creation
	fileWriter, err := writers.NewFileWriter(filePath, bufferSize, maxFiles)
	if err != nil {
		return nil, err
	}

	// Use the existing WithWriter method
	return d.WithWriter(name, fileWriter), nil
}

func (d *consolelogger) WithFileWriterCustom(name string, fileWriter io.Writer) (interfaces.IConsoleLogger, error) {
	// Use the existing WithWriter method
	return d.WithWriter(name, fileWriter), nil
}

func (d *consolelogger) WithFileWriterPattern(name string, pattern string, format string, bufferSize, maxFiles int) (interfaces.IConsoleLogger, error) {
	// Create enhanced file writer with pattern and format
	fileWriter, err := writers.NewFileWriterWithPattern("", pattern, format, bufferSize, maxFiles)
	if err != nil {
		return nil, err
	}

	// Use the existing WithWriter method
	return d.WithWriter(name, fileWriter), nil
}

func (d *consolelogger) GetMemoryLogs(correlationid string, minLevel Level) (map[string]string, error) {
	return writers.GetEntriesWithLevel(correlationid, minLevel)
}

// Write implements io.Writer interface for direct usage with Gin
func (d *consolelogger) Write(p []byte) (n int, err error) {
	n = len(p)

	if n == 0 {
		return n, nil
	}

	logContent := string(p)

	// Create gin detector with current level
	ginDetector := writers.NewGinDetector(d.GetLevel())

	if ginDetector.IsGinLog(logContent) {
		return d.handleGinLog(p)
	}

	// Handle as regular log entry using writers package
	return d.handleRegularLog(p)
}

// handleGinLog processes Gin log entries with proper formatting
func (d *consolelogger) handleGinLog(p []byte) (n int, err error) {
	n = len(p)

	// Create gin detector and parse log
	ginDetector := writers.NewGinDetector(d.GetLevel())
	logentry := ginDetector.ParseGinLog(p)

	// Check if we should log this level
	if !ginDetector.ShouldLogLevel(logentry.Level) {
		return n, nil
	}

	// Write formatted output directly to stdout for console display
	formattedOutput := ginDetector.FormatConsoleOutput(logentry)
	_, writeErr := fmt.Fprintf(os.Stdout, "%s\n", formattedOutput)
	if writeErr != nil {
		internallog.Warn().Err(writeErr).Msg("Failed to write Gin log to console")
	}

	// Write JSON to all other writers (file, memory, etc.)
	for writerName, writer := range d.writers {
		if writerName != WRITER_CONSOLE && writer != nil {
			jsonOutput, err := ginDetector.ToJSON(logentry)
			if err != nil {
				internallog.Warn().Err(err).Msg("Failed to marshal Gin log to JSON")
				continue
			}
			_, err = writer.Write(jsonOutput)
			if err != nil {
				internallog.Warn().Str("writer", writerName).Err(err).Msg("Failed to write Gin log to writer")
			}
		}
	}

	return n, nil
}

// handleRegularLog processes regular log entries using writers package
func (d *consolelogger) handleRegularLog(p []byte) (n int, err error) {
	// Convert io.Writer map to interface{} map for writers.HandleRegularLog
	writerMap := make(map[string]interface{})
	for name, writer := range d.writers {
		writerMap[name] = writer
	}

	n, err = writers.HandleRegularLog(p, writerMap)
	if err != nil {
		// Log the error using internal logger
		internallog.Warn().Err(err).Msg("Failed to handle regular log")
		// Don't return the error, just log it and continue
		return len(p), nil
	}

	return n, nil
}
