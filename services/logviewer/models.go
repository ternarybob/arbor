package logviewer

import (
	"encoding/json"
	"time"

	"github.com/phuslu/log"
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
		"level":         levelToString(e.Level),
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

// levelToString converts log.Level integer to 3-letter string format
func levelToString(level log.Level) string {
	switch level {
	case log.TraceLevel:
		return "TRC"
	case log.DebugLevel:
		return "DBG"
	case log.InfoLevel:
		return "INF"
	case log.WarnLevel:
		return "WAR"
	case log.ErrorLevel:
		return "ERR"
	case log.FatalLevel:
		return "FTL"
	case log.PanicLevel:
		return "PNC"
	default:
		return "UNK"
	}
}
