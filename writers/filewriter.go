package writers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/phuslu/log"
)

type FileWriter struct {
	logger     *log.Logger
	filePath   string
	maxFiles   int
	fileWriter *os.File
	minLevel   log.Level
}

func NewFileWriter(filePath string, bufferSize, maxFiles int) (*FileWriter, error) {
	return NewFileWriterWithLevel(filePath, bufferSize, maxFiles, log.TraceLevel)
}

func NewFileWriterWithLevel(filePath string, bufferSize, maxFiles int, minLevel log.Level) (*FileWriter, error) {
	// Create directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Create or open file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Create Logger with console writer that writes to file
	logger := &log.Logger{
		Level: minLevel,
		Writer: &log.ConsoleWriter{
			Writer:      file,
			ColorOutput: false,
		},
	}

	fw := &FileWriter{
		logger:     logger,
		filePath:   filePath,
		maxFiles:   maxFiles,
		fileWriter: file,
		minLevel:   minLevel,
	}

	// Handle file rotation
	fw.rotateFiles(filePath, maxFiles)

	return fw, nil
}

// FileLogEntry represents a parsed log entry for file writing
type FileLogEntry struct {
	Level   string      `json:"level"`
	Message string      `json:"message"`
	Time    string      `json:"time,omitempty"`
	Prefix  string      `json:"prefix,omitempty"`
	Extra   interface{} `json:"-"`
}

// parseLogLevel converts string level to phuslu/log Level
func parseLogLevel(levelStr string) log.Level {
	switch strings.ToLower(levelStr) {
	case "trace", "trc":
		return log.TraceLevel
	case "debug", "dbg":
		return log.DebugLevel
	case "info", "inf":
		return log.InfoLevel
	case "warn", "warning", "wrn":
		return log.WarnLevel
	case "error", "err":
		return log.ErrorLevel
	case "fatal", "ftl":
		return log.FatalLevel
	case "panic":
		return log.PanicLevel
	default:
		return log.InfoLevel
	}
}

// Write implements io.Writer interface
func (fw *FileWriter) Write(p []byte) (n int, err error) {
	input := strings.TrimSpace(string(p))
	if input == "" {
		return len(p), nil
	}

	// Try to parse as JSON
	var entry FileLogEntry
	if jsonErr := json.Unmarshal([]byte(input), &entry); jsonErr == nil {
		// Successfully parsed JSON - log at the specified level
		level := parseLogLevel(entry.Level)

		// Check if this level should be logged based on minimum level
		if level < fw.minLevel {
			return len(p), nil // Skip this log entry
		}

		// Log at the appropriate level with message and fields
		switch level {
		case log.TraceLevel:
			fw.logger.Trace().Str("prefix", entry.Prefix).Str("original_time", entry.Time).Msg(entry.Message)
		case log.DebugLevel:
			fw.logger.Debug().Str("prefix", entry.Prefix).Str("original_time", entry.Time).Msg(entry.Message)
		case log.InfoLevel:
			fw.logger.Info().Str("prefix", entry.Prefix).Str("original_time", entry.Time).Msg(entry.Message)
		case log.WarnLevel:
			fw.logger.Warn().Str("prefix", entry.Prefix).Str("original_time", entry.Time).Msg(entry.Message)
		case log.ErrorLevel:
			fw.logger.Error().Str("prefix", entry.Prefix).Str("original_time", entry.Time).Msg(entry.Message)
		case log.FatalLevel:
			// Use Error level instead of Fatal to avoid program exit
			fw.logger.Error().Str("level", "fatal").Str("prefix", entry.Prefix).Str("original_time", entry.Time).Msg(entry.Message)
		case log.PanicLevel:
			// Use Error level instead of Panic to avoid program exit
			fw.logger.Error().Str("level", "panic").Str("prefix", entry.Prefix).Str("original_time", entry.Time).Msg(entry.Message)
		default:
			fw.logger.Info().Str("prefix", entry.Prefix).Str("original_time", entry.Time).Msg(entry.Message)
		}
	} else {
		// Not JSON, log as-is at info level
		fw.logger.Info().Msg(input)
	}

	return len(p), nil
}

// Close properly closes the file
func (fw *FileWriter) Close() error {
	if err := fw.fileWriter.Close(); err != nil {
		return err
	}
	return nil
}

// New creates a new FileWriter using the old signature for compatibility
func New(file *os.File, bufferSize int) *FileWriter {
	filePath := file.Name()
	file.Close()

	fw, err := NewFileWriter(filePath, bufferSize, 10) // default 10 max files
	if err != nil {
		// fallback to a basic file writer on error
		return &FileWriter{
			filePath: filePath,
			maxFiles: 10,
		}
	}
	return fw
}

// NewFileWriterWithPattern creates a new FileWriter with custom naming pattern
func NewFileWriterWithPattern(filePath, pattern, format string, bufferSize, maxFiles int) (*FileWriter, error) {
	return NewFileWriterWithPatternAndLevel(filePath, pattern, format, bufferSize, maxFiles, log.TraceLevel)
}

// NewFileWriterWithPatternAndLevel creates a new FileWriter with custom naming pattern and minimum log level
func NewFileWriterWithPatternAndLevel(filePath, pattern, format string, bufferSize, maxFiles int, minLevel log.Level) (*FileWriter, error) {
	// If pattern is provided, expand it to create the actual filename
	if pattern != "" {
		dir := filepath.Dir(filePath)
		baseName := expandFileNamePattern(pattern, "")
		filePath = filepath.Join(dir, baseName)
	}

	return NewFileWriterWithLevel(filePath, bufferSize, maxFiles, minLevel)
}

// SetMinLevel sets the minimum log level for this writer
func (fw *FileWriter) SetMinLevel(level log.Level) {
	fw.minLevel = level
	fw.logger.Level = level
}

// Flush implements IBufferedWriter interface - forces any buffered data to be written
func (fw *FileWriter) Flush() error {
	// phuslu/log doesn't provide explicit flushing, but we can sync the file
	if fw.fileWriter != nil {
		return fw.fileWriter.Sync()
	}
	return nil
}

// SetBufferSize implements IBufferedWriter interface - no-op for this implementation
func (fw *FileWriter) SetBufferSize(size int) error {
	// This implementation doesn't use configurable buffering
	return nil
}

// Rotate implements IRotatableWriter interface - manually triggers log rotation
func (fw *FileWriter) Rotate() error {
	// Force a rotation by calling rotateFiles with maxFiles - 1
	fw.rotateFiles(fw.filePath, fw.maxFiles-1)
	return nil
}

// SetMaxFiles implements IRotatableWriter interface - sets the maximum number of files to keep
func (fw *FileWriter) SetMaxFiles(maxFiles int) {
	fw.maxFiles = maxFiles
}

// rotateFiles rotates the log files to ensure no more than maxFiles are kept
func (fw *FileWriter) rotateFiles(filePath string, maxFiles int) {
	// Get directory for rotation
	dir := filepath.Dir(filePath)

	// Create pattern to match log files
	pattern := dir + string(filepath.Separator) + "*" + ".log"

	files, err := filepath.Glob(pattern)
	if err != nil {
		fw.logger.Error().Msgf("Error fetching log files for rotation: %v", err)
		return
	}

	// Ensure files are sorted, oldest first (by name, which should be date-based)
	sort.Strings(files)

	// Remove old log files if we exceed maxFiles
	for len(files) >= maxFiles {
		if err := os.Remove(files[0]); err != nil {
			fw.logger.Error().Msgf("Error removing old log file %s: %v", files[0], err)
		}
		files = files[1:]
	}
}

// expandFileNamePattern expands placeholders in filename patterns
func expandFileNamePattern(pattern, serviceName string) string {
	now := time.Now()

	expanded := strings.ReplaceAll(pattern, "{SERVICE}", serviceName)
	expanded = strings.ReplaceAll(expanded, "{YYMMDD}", now.Format("060102"))
	expanded = strings.ReplaceAll(expanded, "{YYMMDD-HH}", now.Format("060102-15"))
	expanded = strings.ReplaceAll(expanded, "{YYMMDD-HHMMSS}", now.Format("060102-150405"))
	expanded = strings.ReplaceAll(expanded, "{TT}", now.Format("15"))
	expanded = strings.ReplaceAll(expanded, "{YYYY}", now.Format("2006"))
	expanded = strings.ReplaceAll(expanded, "{MM}", now.Format("01"))
	expanded = strings.ReplaceAll(expanded, "{DD}", now.Format("02"))
	expanded = strings.ReplaceAll(expanded, "{HH}", now.Format("15"))
	expanded = strings.ReplaceAll(expanded, "{MMSS}", now.Format("0405"))

	return expanded
}
