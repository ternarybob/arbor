package writers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/models"

	"github.com/phuslu/log"
)

const (
	MaxLogSize int64 = 10 * 1024 * 1024 // 10 MB
)

type fileWriter struct {
	logger     log.Logger
	config     models.WriterConfiguration
	customFile *os.File
	fileName   string
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

	if !common.IsEmpty(config.LogNameFormat) {
		now := time.Now()
		fileName = strings.ReplaceAll(config.LogNameFormat, "YY", now.Format("06"))
		fileName = strings.ReplaceAll(fileName, "MM", now.Format("01"))
		fileName = strings.ReplaceAll(fileName, "DD", now.Format("02"))
		fileName = strings.ReplaceAll(fileName, "TT", now.Format("15"))
		fileName = strings.ReplaceAll(fileName, "mm", now.Format("04"))
	}

	fw := &fileWriter{
		config:   config,
		fileName: fileName,
	}

	if config.DisableTimestamp {
		// Use custom file writer that respects exact filename
		err := fw.initCustomFileWriter()
		if err != nil {
			// Fallback to phuslu if custom writer fails
			fw.initPhusluWriter(fileName, maxSize, maxBackups)
		}
	} else {
		// Use phuslu file writer (with timestamp appending)
		fw.initPhusluWriter(fileName, maxSize, maxBackups)
	}

	return fw
}

func (fw *fileWriter) initCustomFileWriter() error {
	// Ensure directory exists
	dir := filepath.Dir(fw.fileName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Open file for appending
	file, err := os.OpenFile(fw.fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	fw.customFile = file
	return nil
}

func (fw *fileWriter) initPhusluWriter(fileName string, maxSize int64, maxBackups int) {
	phusluFileWriter := &log.FileWriter{
		Filename:     fileName,
		MaxSize:      maxSize,
		MaxBackups:   maxBackups,
		EnsureFolder: true,
		LocalTime:    true,
	}

	fw.logger = log.Logger{
		Level:      fw.config.Level.ToLogLevel(),
		TimeFormat: fw.config.TimeFormat,
		Writer:     phusluFileWriter,
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

	// If using custom file writer (DisableTimestamp = true), write directly to file
	if fw.config.DisableTimestamp && fw.customFile != nil {
		return fw.writeToCustomFile(data)
	}

	// Parse JSON log event from arbor
	var logEvent models.LogEvent
	if err := json.Unmarshal(data, &logEvent); err != nil {
		// If not JSON, fallback to direct output
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

func (fw *fileWriter) writeToCustomFile(data []byte) (n int, err error) {
	// Parse JSON log event from arbor
	var logEvent models.LogEvent
	if err := json.Unmarshal(data, &logEvent); err != nil {
		// If not JSON, write raw data
		return fw.customFile.Write(append(data, '\n'))
	}

	// Format the log entry similar to phuslu format but without timestamp appending
	logLine := fw.formatLogEntry(logEvent)
	return fw.customFile.Write([]byte(logLine + "\n"))
}

func (fw *fileWriter) formatLogEntry(logEvent models.LogEvent) string {
	// Create a JSON-formatted log entry similar to phuslu
	entry := map[string]interface{}{
		"time":    logEvent.Timestamp.Format(fw.config.TimeFormat),
		"level":   strings.ToLower(logEvent.Level.String()),
		"message": logEvent.Message,
	}

	if logEvent.Prefix != "" {
		entry["prefix"] = logEvent.Prefix
	}
	if logEvent.Function != "" {
		entry["function"] = logEvent.Function
	}
	if logEvent.CorrelationID != "" {
		entry["correlationid"] = logEvent.CorrelationID
	}
	if logEvent.Error != "" {
		entry["error"] = logEvent.Error
	}

	// Add custom fields
	for key, value := range logEvent.Fields {
		entry[key] = value
	}

	// Convert to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return logEvent.Message // Fallback to just the message
	}

	return string(jsonData)
}
