package writers

import (
	"encoding/json"

	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/models"

	"github.com/phuslu/log"
)

const (
	MaxLogSize int64 = 10 * 1024 * 1024 // 10 MB
)

type fileWriter struct {
	logger   log.Logger
	config   models.WriterConfiguration
	fileName string
}

func FileWriter(config models.WriterConfiguration) IWriter {

	maxBackups := config.MaxBackups
	if maxBackups < 1 {
		maxBackups = 5
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

	// Configure output format based on TextOutput setting
	var writer log.Writer = phusluFileWriter
	if fw.config.TextOutput {
		// Use ConsoleWriter for human-readable text format, but output to file
		writer = &log.ConsoleWriter{
			Writer:         phusluFileWriter,
			ColorOutput:    false, // No colors in file output
			EndWithMessage: true,
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
