package logviewer

import (
	"time"

	"github.com/ternarybob/arbor/models"
)

// LogFile represents a log file in the directory.
type LogFile struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

// LogEntry represents a single log entry parsed from the log file.
// It mirrors models.LogEvent but is defined here for the service API.
type LogEntry struct {
	models.LogEvent
}
