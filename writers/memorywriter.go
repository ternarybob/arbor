package writers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/models"
)

type MemoryWriter struct {
	Out io.Writer
}

const (
	CORRELATIONID_KEY = "CORRELATIONID"
	BUFFER_LIMIT      = 1000 // Maximum entries per correlation ID
)

var (
	// In-memory storage with mutex for thread safety
	logStore     = make(map[string][]models.LogEvent)
	storeMux     sync.RWMutex
	indexCounter uint64 = 0
	counterMux   sync.Mutex
)

func NewMemoryWriter() *MemoryWriter {
	return &MemoryWriter{
		Out: os.Stdout,
	}
}

func (w *MemoryWriter) Write(entry []byte) (int, error) {
	ep := len(entry)

	if ep == 0 {
		return ep, nil
	}

	err := w.writeline(entry)
	if err != nil {
		internallog.Warn().Str("prefix", "Write").Err(err).Msg("")
	}

	return ep, nil
}

// WithLevel implements the IWriter interface
func (w *MemoryWriter) WithLevel(level log.Level) IWriter {
	// Memory writer doesn't filter by level, just returns itself
	return w
}

// Close implements the IWriter interface - no-op for memory writer
func (w *MemoryWriter) Close() error {
	// Memory writer doesn't need cleanup
	return nil
}

func (w *MemoryWriter) writeline(event []byte) error {
	var (
	// Use direct logging instead of stored logger
	)

	if len(event) <= 0 {
		return nil // Don't error on empty events
	}

	var logentry models.LogEvent

	if err := json.Unmarshal(event, &logentry); err != nil {
		internallog.Warn().Str("prefix", "writeline").Err(err).Msgf("Error:%s Event:%s", err.Error(), string(event))
		return err
	}

	if isEmpty(logentry.CorrelationID) {
		internallog.Debug().Str("prefix", "writeline").Msgf("CorrelationID is empty -> no write to memory store")
		return nil
	}

	// Generate unique index
	counterMux.Lock()
	indexCounter++
	logentry.Index = indexCounter
	counterMux.Unlock()

	// Store in memory
	storeMux.Lock()
	defer storeMux.Unlock()

	if _, exists := logStore[logentry.CorrelationID]; !exists {
		logStore[logentry.CorrelationID] = make([]models.LogEvent, 0, BUFFER_LIMIT)
	}

	// Add to store with buffer limit
	entries := logStore[logentry.CorrelationID]
	if len(entries) >= BUFFER_LIMIT {
		// Remove oldest entry (FIFO)
		entries = entries[1:]
	}
	entries = append(entries, logentry)
	logStore[logentry.CorrelationID] = entries

	internallog.Trace().Str("prefix", "writeline").Msgf("CorrelationID:%s -> message:%s (total entries: %d)",
		logentry.CorrelationID, logentry.Message, len(entries))

	return nil
}

func GetEntries(correlationid string) (map[string]string, error) {
	var (
		// Use direct logging instead of stored logger
		entries map[string]string = make(map[string]string)
	)

	if correlationid == "" {
		return entries, nil // Return empty instead of error
	}

	internallog.Debug().Str("prefix", "GetEntries").Msgf("Getting log entries correlationid:%s", correlationid)

	storeMux.RLock()
	defer storeMux.RUnlock()

	logEvents, exists := logStore[correlationid]
	if !exists || len(logEvents) == 0 {
		internallog.Debug().Str("prefix", "GetEntries").Msgf("No log entries found for correlationid:%s", correlationid)
		return entries, nil
	}

	internallog.Debug().Str("prefix", "GetEntries").Msgf("Found %d entries for correlationid:%s", len(logEvents), correlationid)

	// Convert to formatted strings
	for _, logEvent := range logEvents {
		index := formatIndex(logEvent.Index)
		entries[index] = formatLogEvent(&logEvent)
	}

	return entries, nil
}

// GetEntriesWithLevel returns log entries filtered by minimum log level
func GetEntriesWithLevel(correlationid string, minLevel log.Level) (map[string]string, error) {
	var (
		// Use direct logging instead of stored logger
		entries map[string]string = make(map[string]string)
	)

	if correlationid == "" {
		return entries, nil // Return empty instead of error
	}

	internallog.Debug().Str("prefix", "GetEntriesWithLevel").Msgf("Getting log entries correlationid:%s minLevel:%s", correlationid, minLevel.String())

	storeMux.RLock()
	defer storeMux.RUnlock()

	logEvents, exists := logStore[correlationid]
	if !exists || len(logEvents) == 0 {
		internallog.Debug().Str("prefix", "GetEntriesWithLevel").Msgf("No log entries found for correlationid:%s", correlationid)
		return entries, nil
	}

	internallog.Debug().Str("prefix", "GetEntriesWithLevel").Msgf("Found %d entries for correlationid:%s", len(logEvents), correlationid)

	// Convert to formatted strings, filtering by level
	for _, logEvent := range logEvents {
		// The logEvent.Level is now already a log.Level type
		eventLevel := logEvent.Level

		// Only include entries at or above the minimum level
		// Note: Lower numeric values = higher priority (error=3, warn=2, info=1, debug=0, trace=-1)
		if eventLevel >= minLevel {
			index := formatIndex(logEvent.Index)
			entries[index] = formatLogEvent(&logEvent)
		}
	}

	internallog.Debug().Str("prefix", "GetEntriesWithLevel").Msgf("Returning %d filtered entries (minLevel:%s)", len(entries), minLevel.String())
	return entries, nil
}

func formatLogEvent(l *models.LogEvent) string {
	epoch := l.Timestamp.Format(time.Stamp)

	// Use simple level formatting
	var levelStr string
	switch l.Level {
	case log.DebugLevel:
		levelStr = "DBG"
	case log.InfoLevel:
		levelStr = "INF"
	case log.WarnLevel:
		levelStr = "WRN"
	case log.ErrorLevel:
		levelStr = "ERR"
	case log.FatalLevel:
		levelStr = "FTL"
	case log.PanicLevel:
		levelStr = "PNC"
	case log.TraceLevel:
		levelStr = "TRC"
	default:
		levelStr = "INF"
	}

	output := levelStr + "|" + epoch

	if l.Prefix != "" {
		output += "|" + l.Prefix
	}

	if l.Message != "" {
		output += "|" + l.Message
	}

	if l.Error != "" {
		output += "|" + l.Error
	}

	return output
}

func formatIndex(index uint64) string {
	// Format as 3-digit string for consistent ordering
	return fmt.Sprintf("%03d", index)
}

// isEmpty checks if a string is empty or contains only whitespace
func isEmpty(a string) bool {
	return len(strings.TrimSpace(a)) == 0
}

// ClearEntries removes log entries for a specific correlation ID
func ClearEntries(correlationid string) {
	if correlationid == "" {
		return
	}

	storeMux.Lock()
	defer storeMux.Unlock()

	delete(logStore, correlationid)
}

// ClearAllEntries removes all stored log entries (useful for cleanup)
func ClearAllEntries() {
	storeMux.Lock()
	defer storeMux.Unlock()

	logStore = make(map[string][]models.LogEvent)
}

// GetStoredCorrelationIDs returns all correlation IDs that have stored logs
func GetStoredCorrelationIDs() []string {
	storeMux.RLock()
	defer storeMux.RUnlock()

	ids := make([]string, 0, len(logStore))
	for id := range logStore {
		ids = append(ids, id)
	}
	return ids
}
