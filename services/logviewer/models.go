package logviewer

import (
	"encoding/json"
	"time"

	"github.com/ternarybob/arbor/common"
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
	Time string `json:"time"` // Shadowing Timestamp with string
}

// MarshalJSON customizes JSON marshaling to convert level integer to string
func (e LogEntry) MarshalJSON() ([]byte, error) {
	// Create a map with all fields
	data := map[string]interface{}{
		"index":         e.Index,
		"level":         common.LevelTo3Letter(e.Level),
		"time":          e.Time,
		"correlationid": e.CorrelationID,
		"prefix":        e.Prefix,
		"message":       e.Message,
		"error":         e.Error,
		"function":      e.Function,
		"fields":        e.Fields,
	}
	return json.Marshal(data)
}
