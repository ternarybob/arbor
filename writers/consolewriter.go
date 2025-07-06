package writers

import (
	"encoding/json"

	"github.com/ternarybob/arbor/models"
	"github.com/ternarybob/arbor/services"

	"github.com/gookit/color"
	"github.com/phuslu/log"
)

var (
	loglevel    log.Level  = log.WarnLevel
	internallog log.Logger = log.Logger{
		Level:  loglevel,
		Writer: &log.ConsoleWriter{},
	}
)

func init() {
	// Enable color output for Windows terminals
	color.ForceOpenColor()
}

type consoleWriter struct {
	logger     log.Logger
	config     models.WriterConfiguration
	ginService services.IGinService // Optional Gin formatting service
}

// ConsoleWriter creates a new ConsoleWriter with phuslu backend
func ConsoleWriter(config models.WriterConfiguration) IWriter {
	// Use phuslu's default console writer with colors
	phusluLogger := log.Logger{
		Level:      config.Level.ToLogLevel(),
		TimeFormat: config.TimeFormat,
		Writer: &log.ConsoleWriter{
			ColorOutput:    true,
			EndWithMessage: true,
		},
	}

	cw := &consoleWriter{
		logger: phusluLogger,
		config: config,
	}

	return cw
}

func (cw *consoleWriter) WithLevel(level log.Level) IWriter {
	cw.logger.SetLevel(level)
	return cw
}

func (cw *consoleWriter) Write(data []byte) (n int, err error) {
	n = len(data)
	if n <= 0 {
		return n, nil
	}

	// Parse JSON log event from arbor
	var logEvent models.LogEvent
	if err := json.Unmarshal(data, &logEvent); err != nil {
		// If not JSON, fallback to direct output
		cw.logger.Info().Msg(string(data))
		return n, nil
	}

	// Use phuslu logger with the parsed log event data
	var phusluEvent *log.Entry
	switch logEvent.Level {
	case log.TraceLevel:
		phusluEvent = cw.logger.Trace()
	case log.DebugLevel:
		phusluEvent = cw.logger.Debug()
	case log.InfoLevel:
		phusluEvent = cw.logger.Info()
	case log.WarnLevel:
		phusluEvent = cw.logger.Warn()
	case log.ErrorLevel:
		phusluEvent = cw.logger.Error()
	case log.FatalLevel:
		phusluEvent = cw.logger.Fatal()
	case log.PanicLevel:
		phusluEvent = cw.logger.Panic()
	default:
		phusluEvent = cw.logger.Info()
	}

	// Add arbor-specific fields to phuslu logger
	if logEvent.Prefix != "" {
		phusluEvent = phusluEvent.Str("prefix", logEvent.Prefix)
	}
	if logEvent.Function != "" {
		phusluEvent = phusluEvent.Str("function", logEvent.Function)
	}
	if logEvent.CorrelationID != "" {
		phusluEvent = phusluEvent.Str("correlationid", logEvent.CorrelationID)
	}

	// Add custom fields from arbor
	for key, value := range logEvent.Fields {
		phusluEvent = phusluEvent.Interface(key, value)
	}

	// Add error if present
	if logEvent.Error != "" {
		phusluEvent = phusluEvent.Str("error", logEvent.Error)
	}

	// Send the message through phuslu (uses phuslu's default console format)
	phusluEvent.Msg(logEvent.Message)

	return n, nil
}
