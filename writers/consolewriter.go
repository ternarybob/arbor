package writers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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

type ConsoleWriter struct {
	Out      io.Writer
	minLevel log.Level
}

type LogEvent struct {
	Index         uint64    `json:"index"`
	Level         string    `json:"level"`
	Timestamp     time.Time `json:"time"`
	CorrelationID string    `json:"correlationid"`
	Prefix        string    `json:"prefix"`
	Message       string    `json:"message"`
	Error         string    `json:"error"`
}

func NewConsoleWriter() *ConsoleWriter {

	return &ConsoleWriter{
		Out:      os.Stdout,
		minLevel: log.TraceLevel, // Default to most verbose level
	}

}

func (w *ConsoleWriter) Write(e []byte) (n int, err error) {

	n = len(e)
	if n <= 0 {
		return n, err
	}

	// fmt.Printf("%d", n)

	err = w.writeline(e)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (w *ConsoleWriter) writeline(event []byte) error {

	// Use direct logging instead of chained logger

	if len(event) <= 0 {
		internallog.Warn().Str("prefix", "writeline").Msg("Entry is Empty")
		return fmt.Errorf("Entry is Empty")
	}

	var logentry LogEvent

	if err := json.Unmarshal(event, &logentry); err != nil {

		internallog.Warn().Str("prefix", "writeline").Err(err).Msgf("error:%s entry:%s", err.Error(), string(event))

		return err
	}

	// Check if this level should be logged based on minimum level
	if !w.shouldLogLevel(logentry.Level) {
		return nil // Skip this log entry
	}

	_, err := fmt.Fprintf(w.Out, "%s\n", w.format(&logentry, true))
	if err != nil {
		return err
	}

	return nil
}

func (w *ConsoleWriter) format(l *LogEvent, colour bool) string {

	timestamp := l.Timestamp.Format("15:04:05.000")

	output := fmt.Sprintf("%s|%s", levelprint(l.Level, colour), timestamp)

	if l.Prefix != "" {
		output += fmt.Sprintf("|%s", l.Prefix)
	}

	if l.Message != "" {
		output += fmt.Sprintf("|%s", l.Message)
	}

	if l.Error != "" {
		output += fmt.Sprintf("|%s", l.Error)
	}

	return output
}

// Close implements the IWriter interface - no-op for console writer
func (w *ConsoleWriter) Close() error {
	// Console writer doesn't need cleanup
	return nil
}

// SetMinLevel implements the ILevelWriter interface
func (w *ConsoleWriter) SetMinLevel(level interface{}) error {
	if logLevel, ok := level.(log.Level); ok {
		w.minLevel = logLevel
		return nil
	}
	return fmt.Errorf("invalid level type: expected log.Level, got %T", level)
}

// shouldLogLevel checks if the log level should be written based on minimum level
func (w *ConsoleWriter) shouldLogLevel(levelStr string) bool {
	// Parse the log level from string
	var eventLevel log.Level
	switch strings.ToLower(levelStr) {
	case "trace":
		eventLevel = log.TraceLevel
	case "debug":
		eventLevel = log.DebugLevel
	case "info":
		eventLevel = log.InfoLevel
	case "warn", "warning":
		eventLevel = log.WarnLevel
	case "error":
		eventLevel = log.ErrorLevel
	case "fatal":
		eventLevel = log.FatalLevel
	case "panic":
		eventLevel = log.PanicLevel
	default:
		eventLevel = log.InfoLevel // Default to info level
	}

	// Only include entries at or above the minimum level
	return eventLevel >= w.minLevel
}

// HandleRegularLog processes regular (non-Gin) log entries
// It handles both JSON and plain text logs appropriately
func HandleRegularLog(p []byte, writers map[string]interface{}) (n int, err error) {
	n = len(p)
	
	// Try to parse as JSON first
	var jsonLog map[string]interface{}
	if json.Unmarshal(p, &jsonLog) == nil {
		// It's valid JSON, write to all writers
		for writerName, writer := range writers {
			if w, ok := writer.(interface{ Write([]byte) (int, error) }); ok {
				_, err := w.Write(p)
				if err != nil {
					// Note: We can't use internallog here since it's in the main package
					// The calling code should handle logging errors
					return n, fmt.Errorf("failed to write regular JSON log to writer %s: %w", writerName, err)
				}
			}
		}
	} else {
		// It's plain text, write directly to stdout and create JSON for other writers
		_, writeErr := fmt.Fprintf(os.Stdout, "%s\n", strings.TrimSpace(string(p)))
		if writeErr != nil {
			return n, fmt.Errorf("failed to write plain text log to console: %w", writeErr)
		}
		
		// Create a simple JSON log entry for other writers
		logEntry := LogEvent{
			Level:     "info",
			Timestamp: time.Now(),
			Prefix:    "APP",
			Message:   strings.TrimSpace(string(p)),
		}
		
		// Write JSON to non-console writers
		for writerName, writer := range writers {
			if writerName != "writerconsole" {
				if w, ok := writer.(interface{ Write([]byte) (int, error) }); ok {
					jsonOutput, err := json.Marshal(logEntry)
					if err != nil {
						return n, fmt.Errorf("failed to marshal plain text log to JSON: %w", err)
					}
					_, err = w.Write(jsonOutput)
					if err != nil {
						return n, fmt.Errorf("failed to write plain text log to writer %s: %w", writerName, err)
					}
				}
			}
		}
	}
	
	return n, nil
}
