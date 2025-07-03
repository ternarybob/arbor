package filewriter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	loglevel    zerolog.Level  = zerolog.WarnLevel
	internallog zerolog.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(loglevel)
)

type WriteTask struct {
	data []byte
}

type FileWriter struct {
	mutex    sync.Mutex
	file     *os.File
	queue    chan WriteTask
	wg       sync.WaitGroup
	err      error
	Log      []string
	logFormat   string // "standard" or "json"
	pattern  string // file naming pattern
	filePath string // current file path
}

type LogEvent struct {
	Level         string                 `json:"level"`
	Timestamp     time.Time              `json:"time"`
	Prefix        string                 `json:"prefix"`
	CorrelationID string                 `json:"correlationid"`
	Message       string                 `json:"message"`
	Error         string                 `json:"error"`
	// Additional fields to handle zerolog output
	Fields        map[string]interface{} `json:"-"`
}

func New(file *os.File, bufferSize int) *FileWriter {

	// fileWriter, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)

	w := &FileWriter{
		file:  file,
		queue: make(chan WriteTask, bufferSize),
		wg:    sync.WaitGroup{},
	}
	w.wg.Add(1)
	go w.writeLoop()
	return w

}

// NewWithPath creates a new FileWriter with the specified file path, creating directories if needed
// It also implements file rotation based on the specified max number of log files.
func NewWithPath(filePath string, bufferSize, maxFiles int) (*FileWriter, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	fw := New(file, bufferSize)
	fw.rotateFiles(filePath, maxFiles)

	return fw, nil
}

// NewWithPatternAndFormat creates a new FileWriter with custom naming pattern and format
// pattern: file naming pattern with placeholders like {YYMMDD}, {SERVICE}, etc.
// format: "standard" for console-like format, "json" for JSON format
func NewWithPatternAndFormat(filePath, pattern, format string, bufferSize, maxFiles int) (*FileWriter, error) {
	// If pattern is provided, expand it to create the actual filename
	if pattern != "" {
		dir := filepath.Dir(filePath)
		baseName := expandFileNamePattern(pattern, "")
		filePath = filepath.Join(dir, baseName)
	}

	// Create directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Open file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	// Create FileWriter with enhanced fields
	fw := &FileWriter{
		file:     file,
		queue:    make(chan WriteTask, bufferSize),
		wg:       sync.WaitGroup{},
		logFormat: format,
		pattern:  pattern,
		filePath: filePath,
	}

	fw.wg.Add(1)
	go fw.writeLoopWithFormat()

	// Handle file rotation
	fw.rotateFiles(filePath, maxFiles)

	return fw, nil
}

// rotateFiles rotates the log files to ensure no more than maxFiles are kept
func (w *FileWriter) rotateFiles(filePath string, maxFiles int) {
	// Get directory for rotation
	dir := filepath.Dir(filePath)

	// Create pattern to match log files
	pattern := dir + string(filepath.Separator) + "*" + ".log"

	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Println("Error fetching log files for rotation:", err)
		return
	}

	// Ensure files are sorted, oldest first (by name, which should be date-based)
	sort.Strings(files)

	// Remove old log files if we exceed maxFiles
	for len(files) >= maxFiles {
		if err := os.Remove(files[0]); err != nil {
			fmt.Printf("Error removing old log file %s: %v\n", files[0], err)
		}
		files = files[1:]
	}
}

func (w *FileWriter) Write(e []byte) (n int, err error) {

	n = len(e)
	if n <= 0 {
		return n, err
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.err != nil {
		return // Silently ignore write if already in error state
	}

	select {
	case w.queue <- WriteTask{data: e}:
	default:
		fmt.Println("Write queue full, data dropped:", string(e))
	}

	return n, nil
}

func (w *FileWriter) writeline(event []byte) error {

	log := internallog.With().Str("prefix", "writeline").Logger()

	if len(event) <= 0 {
		log.Warn().Msg("Entry is Empty")
		return fmt.Errorf("Entry is Empty")
	}

	var logentry LogEvent

	if err := json.Unmarshal(event, &logentry); err != nil {

		log.Warn().Err(err).Msgf("error:%s entry:%s", err.Error(), string(event))

		return err
	}

	_, err := fmt.Printf("%s\n", w.formatLine(&logentry, true))
	if err != nil {
		return err
	}

	return nil
}

// Close waits for pending writes and closes the writer
func (w *FileWriter) Close() error {
	close(w.queue)
	w.wg.Wait()
	return w.err
}

// writeLoop continuously processes tasks from the queue and writes to the file
func (w *FileWriter) writeLoop() {
	defer w.wg.Done()
	for task := range w.queue {
		_, err := w.file.Write(task.data)

		if err != nil {
			fmt.Println("Write error:", err)
			w.mutex.Lock()
			w.err = err // Update error state if desired for future writes
			w.mutex.Unlock()
			// No retry logic, continue processing remaining tasks
		}
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

// writeLoopWithFormat processes tasks with format-specific handling
func (w *FileWriter) writeLoopWithFormat() {
	defer w.wg.Done()
	for task := range w.queue {
		var output []byte
		var err error

		if w.logFormat == "json" {
			// Write JSON format directly
			output = task.data
		} else {
			// Convert JSON to standard format
			output, err = w.convertJSONToStandardFormat(task.data)
			if err != nil {
				fmt.Printf("Format conversion error: %v\n", err)
				output = task.data // Fallback to original data
			}
		}

		_, writeErr := w.file.Write(output)
		if writeErr != nil {
			fmt.Println("Write error:", writeErr)
			w.mutex.Lock()
			w.err = writeErr
			w.mutex.Unlock()
		}
	}
}

// convertJSONToStandardFormat converts JSON log data to standard pipe-separated format
func (w *FileWriter) convertJSONToStandardFormat(data []byte) ([]byte, error) {
	// First try to parse as a generic map to handle dynamic fields
	var genericEntry map[string]interface{}
	if err := json.Unmarshal(data, &genericEntry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal log entry: %w", err)
	}

	// Extract fields with defaults
	logEntry := LogEvent{}
	if level, ok := genericEntry["level"].(string); ok {
		logEntry.Level = level
	}
	if timeStr, ok := genericEntry["time"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
			logEntry.Timestamp = parsedTime
		} else {
			logEntry.Timestamp = time.Now()
		}
	} else {
		logEntry.Timestamp = time.Now()
	}
	if prefix, ok := genericEntry["prefix"].(string); ok {
		logEntry.Prefix = prefix
	}
	if correlationID, ok := genericEntry["correlationid"].(string); ok {
		logEntry.CorrelationID = correlationID
	}
	if message, ok := genericEntry["message"].(string); ok {
		logEntry.Message = message
	}
	if errorMsg, ok := genericEntry["error"].(string); ok {
		logEntry.Error = errorMsg
	}

	// Format as standard log line
	formatted := w.formatLine(&logEntry, false) // false = no color codes for file
	return []byte(formatted + "\n"), nil
}

func (w *FileWriter) formatLine(l *LogEvent, colour bool) string {

	timestamp := l.Timestamp.Format(time.Stamp)

	output := fmt.Sprintf("%s|%s", levelprint(l.Level, colour), timestamp)

	if l.CorrelationID != "" {
		output += fmt.Sprintf("|%s", l.CorrelationID)
	}

	if l.Prefix != "" {
		output += fmt.Sprintf("|%-55s", l.Prefix)
	}

	if l.Message != "" {
		output += fmt.Sprintf("|%s", l.Message)
	}

	if l.Error != "" {
		output += fmt.Sprintf("|%s", l.Error)
	}

	return output
}
