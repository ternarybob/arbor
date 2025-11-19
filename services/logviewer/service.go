package logviewer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/phuslu/log"
	arborLevels "github.com/ternarybob/arbor/levels"
)

// Service provides methods for viewing log files.
type Service struct {
	LogDirectory string
}

// NewService creates a new LogViewer Service.
func NewService(logDirectory string) *Service {
	return &Service{
		LogDirectory: logDirectory,
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
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
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

	return entries, nil
}
