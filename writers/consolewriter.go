package writers

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ternarybob/arbor/models"

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
	logger log.Logger
	config models.WriterConfiguration
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
			Formatter:      consoleFormatter,
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

// GetFilePath returns empty string as console writer doesn't write to files
func (cw *consoleWriter) GetFilePath() string {
	return ""
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
		cw.logger.Warn().Msg("data not transposed to json -> fallback to string")
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

func (cw *consoleWriter) Close() error {
	return nil
}

// ANSI Color Codes
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
)

func consoleFormatter(w io.Writer, a *log.FormatterArgs) (int, error) {
	var levelColor string
	var levelText string

	// Map phuslu levels to 3-letter uppercase and colors
	switch a.Level {
	case "trace":
		levelColor = colorGray
		levelText = "TRC"
	case "debug":
		levelColor = colorCyan
		levelText = "DBG"
	case "info":
		levelColor = colorGreen
		levelText = "INF"
	case "warn":
		levelColor = colorYellow
		levelText = "WRN"
	case "error":
		levelColor = colorRed
		levelText = "ERR"
	case "fatal":
		levelColor = colorMagenta
		levelText = "FTL"
	case "panic":
		levelColor = colorMagenta
		levelText = "PNC"
	default:
		levelColor = colorReset
		levelText = "???"
	}

	// Format: Time [Level] > Message KeyValues
	// Note: a.Time is pre-formatted by phuslu based on TimeFormat
	// If TimeFormat is empty, a.Time might be empty or default.
	// Arbor config usually sets a time format.

	// Reconstruct the standard phuslu console format but with our colors and level text
	// Default phuslu: Time Level Message KeyValues
	// We want: Time [Level] > Message KeyValues (based on previous observation of "INF >")
	// Wait, looking at previous output: "2025-11-19T18:50:10.246+11:00 INF > This is an info message"
	// The default phuslu console writer seems to do "Time Level > Message" or similar?
	// Let's check the default output again from repro:
	// "2025-11-19T18:50:10.246+11:00 INF > This is an info message"
	// So yes, it includes a ">".

	// However, a.Message contains the message.
	// a.KeyValues contains structured fields.

	// Let's try to match the exact format.
	// Time is in a.Time

	p := ""
	if a.Time != "" {
		p += fmt.Sprintf("%s%s%s ", colorGray, a.Time, colorReset)
	}

	// Level part
	p += fmt.Sprintf("%s%s%s", levelColor, levelText, colorReset)

	// Separator
	p += " > "

	// Message
	p += a.Message

	// KeyValues
	if len(a.KeyValues) > 0 {
		for _, kv := range a.KeyValues {
			p += fmt.Sprintf(" %s%s%s=%s%v%s", colorBlue, kv.Key, colorReset, colorGray, kv.Value, colorReset)
		}
	}

	p += "\n"

	return w.Write([]byte(p))
}
