package memorywriter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	consolewriter "github.com/ternarybob/arbor/consolewriter"

	"github.com/rs/zerolog"
)

type MemoryWriter struct {
	Out io.Writer
}

type LogEvent struct {
	Index         uint64    `json:"index"`
	Level         string    `json:"level"`
	Timestamp     time.Time `json:"time"`
	CorrelationID string    `json:"correlationid"`
	Prefix        string    `json:"prefix"`
	Message       string    `json:"message"`
	Error         string    `json:"error"`
}

const (
	CORRELATIONID_KEY = "CORRELATIONID"
	BUFFER_LIMIT      = 1000 // Maximum entries per correlation ID
)

var (
	loglevel    zerolog.Level  = zerolog.InfoLevel
	internallog zerolog.Logger = zerolog.New(&consolewriter.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger().Level(loglevel)
	
	// In-memory storage with mutex for thread safety
	logStore = make(map[string][]LogEvent)
	storeMux sync.RWMutex
	indexCounter uint64 = 0
	counterMux sync.Mutex
)

func New() *MemoryWriter {
	return &MemoryWriter{
		Out: os.Stdout,
	}
}

func (w *MemoryWriter) Write(entry []byte) (int, error) {
	log := internallog.With().Str("prefix", "Write").Logger()
	ep := len(entry)

	if ep == 0 {
		return ep, nil
	}

	err := w.writeline(entry)
	if err != nil {
		log.Warn().Err(err).Msg("")
	}

	return ep, nil
}

func (w *MemoryWriter) writeline(event []byte) error {
	var (
		log zerolog.Logger = internallog.With().Str("prefix", "writeline").Logger()
	)

	if len(event) <= 0 {
		return nil // Don't error on empty events
	}

	var logentry LogEvent

	if err := json.Unmarshal(event, &logentry); err != nil {
		log.Warn().Err(err).Msgf("Error:%s Event:%s", err.Error(), string(event))
		return err
	}

	if isEmpty(logentry.CorrelationID) {
		log.Debug().Msgf("CorrelationID is empty -> no write to memory store")
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
		logStore[logentry.CorrelationID] = make([]LogEvent, 0, BUFFER_LIMIT)
	}

	// Add to store with buffer limit
	entries := logStore[logentry.CorrelationID]
	if len(entries) >= BUFFER_LIMIT {
		// Remove oldest entry (FIFO)
		entries = entries[1:]
	}
	entries = append(entries, logentry)
	logStore[logentry.CorrelationID] = entries

	log.Trace().Msgf("CorrelationID:%s -> message:%s (total entries: %d)", 
		logentry.CorrelationID, logentry.Message, len(entries))

	return nil
}

func GetEntries(correlationid string) (map[string]string, error) {
	var (
		log     zerolog.Logger    = internallog.With().Str("prefix", "GetEntries").Logger()
		entries map[string]string = make(map[string]string)
	)

	if correlationid == "" {
		return entries, nil // Return empty instead of error
	}

	log.Debug().Msgf("Getting log entries correlationid:%s", correlationid)

	storeMux.RLock()
	defer storeMux.RUnlock()

	logEvents, exists := logStore[correlationid]
	if !exists || len(logEvents) == 0 {
		log.Debug().Msgf("No log entries found for correlationid:%s", correlationid)
		return entries, nil
	}

	log.Debug().Msgf("Found %d entries for correlationid:%s", len(logEvents), correlationid)

	// Convert to formatted strings
	for _, logEvent := range logEvents {
		index := formatIndex(logEvent.Index)
		entries[index] = logEvent.format()
	}

	return entries, nil
}

// GetEntriesWithLevel returns log entries filtered by minimum log level
func GetEntriesWithLevel(correlationid string, minLevel zerolog.Level) (map[string]string, error) {
	var (
		log     zerolog.Logger    = internallog.With().Str("prefix", "GetEntriesWithLevel").Logger()
		entries map[string]string = make(map[string]string)
	)

	if correlationid == "" {
		return entries, nil // Return empty instead of error
	}

	log.Debug().Msgf("Getting log entries correlationid:%s minLevel:%s", correlationid, minLevel.String())

	storeMux.RLock()
	defer storeMux.RUnlock()

	logEvents, exists := logStore[correlationid]
	if !exists || len(logEvents) == 0 {
		log.Debug().Msgf("No log entries found for correlationid:%s", correlationid)
		return entries, nil
	}

	log.Debug().Msgf("Found %d entries for correlationid:%s", len(logEvents), correlationid)

	// Convert to formatted strings, filtering by level
	for _, logEvent := range logEvents {
		// Parse the log level from string
		eventLevel, err := zerolog.ParseLevel(logEvent.Level)
		if err != nil {
			// If we can't parse the level, include it (safer to include than exclude)
			log.Debug().Msgf("Could not parse level '%s', including entry", logEvent.Level)
			eventLevel = zerolog.TraceLevel // Most verbose level to ensure inclusion
		}

		// Only include entries at or above the minimum level
		// Note: Lower numeric values = higher priority (error=3, warn=2, info=1, debug=0, trace=-1)
		if eventLevel >= minLevel {
			index := formatIndex(logEvent.Index)
			entries[index] = logEvent.format()
		}
	}

	log.Debug().Msgf("Returning %d filtered entries (minLevel:%s)", len(entries), minLevel.String())
	return entries, nil
}

func (l *LogEvent) format() string {
	epoch := l.Timestamp.Format(time.Stamp)
	
	// Use simple level formatting
	levelStr := l.Level
	switch l.Level {
	case "debug":
		levelStr = "DBG"
	case "info":
		levelStr = "INF"
	case "warn":
		levelStr = "WRN"
	case "error":
		levelStr = "ERR"
	case "fatal":
		levelStr = "FTL"
	case "panic":
		levelStr = "PNC"
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

func isEmpty(a string) bool {
	return len(a) == 0
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
	
	logStore = make(map[string][]LogEvent)
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
