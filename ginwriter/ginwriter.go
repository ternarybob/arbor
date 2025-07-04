package ginwriter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/phuslu/log"
)

type GinWriter struct {
	Out   io.Writer
	Level log.Level
}

var (
	loglevel    log.Level  = log.WarnLevel
	prefix      string     = "GinWriter"
	internallog log.Logger = log.Logger{
		Level:  loglevel,
		Writer: &log.ConsoleWriter{},
	}
)

// New creates a new GinWriter instance
func New() *GinWriter {
	return &GinWriter{
		Out:   os.Stdout,
		Level: log.InfoLevel,
	}
}

type LogEvent struct {
	Level         string    `json:"level"`
	Timestamp     time.Time `json:"time"`
	Prefix        string    `json:"prefix"`
	CorrelationID string    `json:"correlationid"`
	Message       string    `json:"message"`
	Error         string    `json:"error"`
}


func (w *GinWriter) Write(e []byte) (n int, err error) {

	n = len(e)

	if n == 0 {
		return n, nil
	}

	// fmt.Printf(string(e))

	err = w.writeline(e)
	if err != nil {
		return 0, err
	}

	return 0, nil
}

func (w *GinWriter) writeline(event []byte) error {

	var (
		// Use direct logging instead of chained logger
		logentry LogEvent
	)

	if len(event) <= 0 {
		return fmt.Errorf("[%s] entry is Empty", prefix)
	}

	logentry.Prefix = "GIN"
	logentry.Timestamp = time.Now()
	logentry.Message = strings.TrimSuffix(string(event), "\n")

	logstring := string(event)

	switch {
	case stringContains(logstring, "GIN-fatal"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-fatal] ", "")
		logentry.Level = "fatal"
	case stringContains(logstring, "GIN-error"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-error] ", "")
		logentry.Level = "error"
	case stringContains(logstring, "GIN-warning"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-warning] ", "")
		logentry.Level = "warn"
	case stringContains(logstring, "GIN-information"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-information] ", "")
		logentry.Level = "info"
	case stringContains(logstring, "GIN-debug"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-debug] ", "")
		logentry.Level = "debug"
	default:
	// Default to info level for standard Gin logs
		logentry.Level = "info"
	}

	// Check if we should log this level
	if !shouldLogLevel(logentry.Level, w.Level) {
		return nil
	}

	// Convert to JSON and write to output
	jsonOutput, err := json.Marshal(logentry)
	if err != nil {
		internallog.Warn().Str("prefix", "writeline").Err(err).Msg("Failed to marshal log entry")
		return err
	}

	_, err = fmt.Fprintf(w.Out, "%s\n", string(jsonOutput))
	if err != nil {
		internallog.Warn().Str("prefix", "writeline").Err(err).Msg("Failed to write log entry")
		return err
	}

	return nil
}

// Level filtering based on string comparison with phuslu log levels
func shouldLogLevel(level string, writerLevel log.Level) bool {
	switch strings.ToLower(level) {
	case "fatal":
		return writerLevel <= log.FatalLevel
	case "error":
		return writerLevel <= log.ErrorLevel
	case "warn":
		return writerLevel <= log.WarnLevel
	case "info":
		return writerLevel <= log.InfoLevel
	case "debug":
		return writerLevel <= log.DebugLevel
	default:
		return true
	}
}

func toJson(input interface{}) string {

	output, err := json.MarshalIndent(input, "", "\t")

	if err != nil {
		internallog.Error().Str("prefix", "toJson").Msgf("Object marshaling error: %v", err)
		return ""
	}

	return string(output)
}

func isEmpty(input string) bool {
	return (len(strings.TrimSpace(input)) <= 0)
}

func stringContains(this string, contains string) bool {

	if strings.Contains(strings.ToLower(this), strings.ToLower(contains)) {
		return true
	}

	return false
}
