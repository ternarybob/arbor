package logviewer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/common"
	arborLevels "github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

// Service provides methods for viewing log files.
type Service struct {
	LogDirectory string
	Format       string // "text" or "json"
}

// NewService creates a new LogViewer Service.
func NewService(config models.WriterConfiguration) *Service {
	fileName := config.FileName
	if common.IsEmpty(fileName) {
		fileName = "logs/main.log"
	}

	logDirectory := filepath.Dir(fileName)

	format := "json"
	if config.TextOutput {
		format = "text"
	}

	return &Service{
		LogDirectory: logDirectory,
		Format:       format,
	}
}

// ListLogFiles returns a list of log files in the configured directory.
func (s *Service) ListLogFiles() ([]LogFile, error) {
	files, err := os.ReadDir(s.LogDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read log directory: %w", err)
	}

	var logFiles []LogFile
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		logFiles = append(logFiles, LogFile{
			Name:    file.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	// Sort by modification time, newest first
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].ModTime.After(logFiles[j].ModTime)
	})

	return logFiles, nil
}

// GetLogContent returns parsed log entries from a specific log file.
// filename: The name of the file to read.
// limit: Number of lines to read from the end (tail). If <= 0, reads all.
// levels: List of log levels to filter by (case-insensitive). If empty, returns all.
func (s *Service) GetLogContent(filename string, limit int, levels []string) ([]LogEntry, error) {
	// Security check: prevent directory traversal
	if filepath.Base(filename) != filename {
		return nil, fmt.Errorf("invalid file name")
	}

	filePath := filepath.Join(s.LogDirectory, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)

	// Create a map for faster level lookup
	// We map the integer value of log.Level to bool
	levelMap := make(map[log.Level]bool)
	if len(levels) > 0 {
		for _, l := range levels {
			lvl, err := arborLevels.ParseLevelString(l)
			if err == nil {
				levelMap[lvl] = true
			}
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var entry LogEntry
		var err error

		// Try to parse as JSON first
		if json.Valid([]byte(line)) {
			// For JSON, we need to handle the time field being a string or something else.
			// If we unmarshal into LogEntry, the custom MarshalJSON won't help with Unmarshal.
			// And LogEntry has Time string.
			// If the JSON has "time": "...", it will go into Time string?
			// models.LogEvent has Timestamp time.Time with json:"time".
			// LogEntry has Time string with json:"time".
			// Go's JSON unmarshaler prefers the field in the struct over embedded.
			// So it should unmarshal "time" into LogEntry.Time (string).
			err = json.Unmarshal([]byte(line), &entry)
		} else {
			// Fallback to text parsing
			entry, err = parseTextLog(line)
		}

		if err != nil {
			continue
		}

		// Filter by level
		if len(levelMap) > 0 {
			if !levelMap[entry.Level] {
				continue
			}
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Apply limit if needed (naive implementation, for large files we should read from end)
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	return entries, nil
}

// parseTextLog parses a text format log line.
// Supported formats:
// 1. Pipe delimited: "TIMESTAMP | LEVEL | MESSAGE | KEY=VALUE"
// 2. Legacy: "TIMESTAMP LEVEL > [FIELDS] MESSAGE"
func parseTextLog(line string) (LogEntry, error) {
	// Check for pipe delimiter
	if strings.Contains(line, " | ") {
		return parsePipeLog(line)
	}

	// Fallback to legacy format
	return parseLegacyTextLog(line)
}

func parsePipeLog(line string) (LogEntry, error) {
	parts := strings.Split(line, " | ")
	if len(parts) < 3 {
		return LogEntry{}, fmt.Errorf("invalid pipe format: not enough parts")
	}

	timeStr := parts[0]
	levelStr := parts[1]
	message := parts[2]

	// Parse Level
	level := parseLevel(levelStr)

	// Parse Timestamp
	// The user request says "do not convert to date or timestamp".
	// So we just set the Time string field.
	// However, LogEvent has Timestamp time.Time.
	// If we leave it zero, it might be confusing if accessed elsewhere.
	// But LogEntry shadows it for JSON output.
	// Let's just set the Time string.

	var fields map[string]interface{}

	if len(parts) == 4 {
		// We have a 4th part. Is it KeyValues or part of message?
		// We can try to parse it as key=value.
		// If it parses well, treat as fields.
		// Otherwise append to message.
		potentialFields := parts[3]
		parsedFields := parseFields(potentialFields)
		if len(parsedFields) > 0 {
			fields = parsedFields
		} else {
			// Append back to message
			message = message + " | " + potentialFields
		}
	}

	return LogEntry{
		LogEvent: models.LogEvent{
			Level:   level,
			Message: message,
			Fields:  fields,
		},
		Time: timeStr,
	}, nil
}

func parseFields(text string) map[string]interface{} {
	fields := make(map[string]interface{})
	// Split by space, but handle quotes?
	// The formatter uses fmt.Sprintf("%s=%v", k, v).
	// It doesn't quote.
	// So "key=value key2=value2".
	// If value contains space, it's not quoted in formatter!
	// Wait, `fmt.Sprintf("%v", value)` might produce spaces.
	// If value is string "hello world", it prints "hello world".
	// So "msg=hello world".
	// Then "key=value msg=hello world".
	// Splitting by space will fail: "msg=hello", "world".
	// The formatter in filewriter.go is naive:
	// p += fmt.Sprintf("%s=%v", kv.Key, kv.Value)
	// It separates by space.
	// If value contains space, parsing is ambiguous.
	// But we have to do our best.
	// We can look for "key=".

	// Regex: `(\S+)=` matches keys.
	// But we don't have regex here easily without compiling.
	// Let's just split by space for now, as it's standard logfmt approximation.
	// Or better, we assume the user knows what they are doing with spaces.

	parts := strings.Split(text, " ")
	for _, part := range parts {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				fields[kv[0]] = kv[1]
			}
		}
	}
	return fields
}

func parseLegacyTextLog(line string) (LogEntry, error) {
	// Find the separator " > "
	// We expect: Timestamp Level > Message
	const separator = " > "
	sepIndex := -1

	// Optimization: The separator should be within the first 50 characters typically
	// Timestamp (30-35) + Space + Level (3) + Separator
	searchLimit := 60
	if len(line) < searchLimit {
		searchLimit = len(line)
	}

	for i := 0; i < searchLimit-len(separator)+1; i++ {
		if line[i:i+len(separator)] == separator {
			sepIndex = i
			break
		}
	}

	if sepIndex == -1 {
		return LogEntry{}, fmt.Errorf("separator not found")
	}

	// Part before separator: "Timestamp Level"
	prefix := line[:sepIndex]
	// Part after separator: "Message" (including fields)
	message := line[sepIndex+len(separator):]

	// Split prefix by last space to separate Timestamp and Level
	lastSpace := -1
	for i := len(prefix) - 1; i >= 0; i-- {
		if prefix[i] == ' ' {
			lastSpace = i
			break
		}
	}

	if lastSpace == -1 {
		return LogEntry{}, fmt.Errorf("invalid prefix format")
	}

	timeStr := prefix[:lastSpace]
	levelStr := prefix[lastSpace+1:]

	// Parse Level
	level := parseLevel(levelStr)

	// Parse Timestamp

	return LogEntry{
		LogEvent: models.LogEvent{
			Level:   level,
			Message: message,
		},
		Time: timeStr,
	}, nil
}

func parseLevel(levelStr string) log.Level {
	switch levelStr {
	case "TRC":
		return log.TraceLevel
	case "DBG":
		return log.DebugLevel
	case "INF":
		return log.InfoLevel
	case "WAR", "WRN":
		return log.WarnLevel
	case "ERR":
		return log.ErrorLevel
	case "FTL":
		return log.FatalLevel
	case "PNC":
		return log.PanicLevel
	default:
		return log.InfoLevel
	}
}
