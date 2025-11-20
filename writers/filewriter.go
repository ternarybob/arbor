package writers

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/models"

	"github.com/phuslu/log"
)

const (
	// MaxLogSize default is 500KB - optimized for AI agent consumption
	// At ~150 bytes per log line, this allows approximately 3,300 lines per file
	// which is well within most AI context windows while maintaining readability
	MaxLogSize int64 = 500 * 1024 // 500 KB

	// DefaultMaxBackups keeps 20 backup files by default
	// With 500KB files, this provides ~10MB total log history (20 * 500KB)
	// spanning multiple days of activity while remaining manageable
	DefaultMaxBackups int = 20
)

type fileWriter struct {
	logger   log.Logger
	config   models.WriterConfiguration
	fileName string
}

func FileWriter(config models.WriterConfiguration) IWriter {

	maxBackups := config.MaxBackups
	if maxBackups < 1 {
		maxBackups = DefaultMaxBackups
	}

	maxSize := config.MaxSize
	if maxSize < 1 {
		maxSize = MaxLogSize
	}

	fileName := config.FileName
	if common.IsEmpty(fileName) {
		fileName = "logs/main.log"
	}

	fw := &fileWriter{
		config:   config,
		fileName: fileName,
	}

	// Use phuslu file writer with standard backup naming convention
	fw.initPhusluWriter(fileName, maxSize, maxBackups)

	return fw
}

func (fw *fileWriter) initPhusluWriter(fileName string, maxSize int64, maxBackups int) {
	phusluFileWriter := &log.FileWriter{
		Filename:     fileName,
		MaxSize:      maxSize,
		MaxBackups:   maxBackups,
		EnsureFolder: true,
		LocalTime:    true,
	}

	// Configure output format based on TextOutput setting.
	// Default is logfmt for AI-friendly, human-readable logs.
	format := fw.config.TextOutput
	if format == "" {
		format = models.TextOutputFormatLogfmt
	}

	var writer log.Writer = phusluFileWriter
	switch format {
	case models.TextOutputFormatJSON:
		// Structured JSON output (legacy behavior)
		writer = phusluFileWriter
	default:
		// Logfmt or any other text format uses the custom formatter
		writer = &log.ConsoleWriter{
			Writer:         phusluFileWriter,
			ColorOutput:    false, // No colors in file output
			EndWithMessage: true,
			Formatter:      fileFormatter,
		}
	}

	fw.logger = log.Logger{
		Level:      fw.config.Level.ToLogLevel(),
		TimeFormat: fw.config.TimeFormat,
		Writer:     writer,
	}
}

// FileLogEntry represents a parsed log entry for file writing
type FileLogEntry struct {
	Level   string      `json:"level"`
	Message string      `json:"message"`
	Time    string      `json:"time,omitempty"`
	Prefix  string      `json:"prefix,omitempty"`
	Extra   interface{} `json:"-"`
}

func (fw *fileWriter) WithLevel(level log.Level) IWriter {

	fw.logger.SetLevel(level)

	return fw
}

// GetFilePath returns the configured file path (phuslu creates timestamped files automatically)
func (fw *fileWriter) GetFilePath() string {
	// Return the base configured filename
	// Note: phuslu FileWriter actually creates timestamped files like: name.YYYY-MM-DDTHH-MM-SS.ext
	// but doesn't expose the current filename through its API
	return fw.fileName
}

func (fw *fileWriter) Write(data []byte) (n int, err error) {
	n = len(data)
	if n <= 0 {
		return n, nil
	}

	// Parse JSON log event from arbor
	var logEvent models.LogEvent
	if err := json.Unmarshal(data, &logEvent); err != nil {
		// If not JSON, fallback to direct output
		fw.logger.Warn().Msg("data not transposed to json -> fallback to string")
		fw.logger.Info().Msg(string(data))
		return n, nil
	}

	// Use phuslu logger with the parsed log event data
	var phusluEvent *log.Entry
	switch logEvent.Level {
	case log.TraceLevel:
		phusluEvent = fw.logger.Trace()
	case log.DebugLevel:
		phusluEvent = fw.logger.Debug()
	case log.InfoLevel:
		phusluEvent = fw.logger.Info()
	case log.WarnLevel:
		phusluEvent = fw.logger.Warn()
	case log.ErrorLevel:
		phusluEvent = fw.logger.Error()
	case log.FatalLevel:
		phusluEvent = fw.logger.Fatal()
	case log.PanicLevel:
		phusluEvent = fw.logger.Panic()
	default:
		phusluEvent = fw.logger.Info()
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

func (fw *fileWriter) Close() error {
	if fileWriter, ok := fw.logger.Writer.(*log.FileWriter); ok {
		return fileWriter.Close()
	}
	return nil
}

func fileFormatter(w io.Writer, a *log.FormatterArgs) (int, error) {
	// Map phuslu levels to 3-letter uppercase
	levelText := common.LevelStringTo3Letter(a.Level)

	// Format: logfmt-style
	//   time=<timestamp> level=<LVL> message="Message" key=value
	// Example:
	//   time=2025-11-19T22:08:28.123+11:00 level=INF message="Message" user=john

	var b strings.Builder

	// Timestamp
	if a.Time != "" {
		b.WriteString("time=")
		b.WriteString(a.Time)
	}

	// Level
	if levelText != "" {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("level=")
		b.WriteString(levelText)
	}

	// Message (always quoted to preserve spaces)
	if a.Message != "" {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("message=")
		b.WriteString(fmt.Sprintf("%q", a.Message))
	}

	// Additional key=value fields (logfmt-style)
	if len(a.KeyValues) > 0 {
		for _, kv := range a.KeyValues {
			if kv.Key == "" {
				continue
			}

			value := fmt.Sprint(kv.Value)
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(kv.Key)
			b.WriteByte('=')

			// Quote value if it contains spaces or quotes
			if strings.ContainsAny(value, " \"") {
				b.WriteString(fmt.Sprintf("%q", value))
			} else {
				b.WriteString(value)
			}
		}
	}

	b.WriteByte('\n')
	return io.WriteString(w, b.String())
}
