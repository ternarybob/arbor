// -----------------------------------------------------------------------
// Last Modified: Wednesday, 1st October 2025 4:10:00 pm
// Modified By: Bob McAllan
// -----------------------------------------------------------------------

package writers

import (
	"fmt"
	"sort"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

// memoryWriter provides query interface to the log store
// It doesn't write directly - LogStoreWriter handles that
type memoryWriter struct {
	store  ILogStore
	config models.WriterConfiguration
}

// MemoryWriter creates a new memory writer backed by a log store
// Note: This should be used alongside LogStoreWriter
// LogStoreWriter handles writes, MemoryWriter handles queries
func MemoryWriter(config models.WriterConfiguration) IMemoryWriter {
	internalLog := common.NewLogger().WithContext("function", "MemoryWriter").GetLogger()

	// Create the shared log store
	store, err := NewInMemoryLogStore(config)
	if err != nil {
		internalLog.Fatal().Err(err).Msg("Failed to create log store")
	}

	mw := &memoryWriter{
		store:  store,
		config: config,
	}

	return mw
}

// GetStore returns the underlying log store for use with LogStoreWriter
func (mw *memoryWriter) GetStore() ILogStore {
	return mw.store
}

// Write is a no-op for the memory writer
// Use LogStoreWriter for actual writing
func (mw *memoryWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

// WithLevel sets the log level (no-op for memory writer, filtering done at query time)
func (mw *memoryWriter) WithLevel(level log.Level) IWriter {
	mw.config.Level = levels.FromLogLevel(level)
	return mw
}

// GetFilePath returns empty string as memory writer doesn't write to files
func (mw *memoryWriter) GetFilePath() string {
	return ""
}

// GetEntries retrieves all log entries for a correlation ID, ordered by timestamp
func (mw *memoryWriter) GetEntries(correlationID string) (map[string]string, error) {
	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.GetEntries").GetLogger()

	if correlationID == "" {
		return make(map[string]string), nil
	}

	internalLog.Debug().Msgf("Getting log entries for correlationID:%s", correlationID)

	entries, err := mw.store.GetByCorrelation(correlationID)
	if err != nil {
		internalLog.Error().Err(err).Msgf("Error retrieving entries for correlationID:%s", correlationID)
		return make(map[string]string), err
	}

	// Format entries
	result := make(map[string]string)
	for _, entry := range entries {
		index := formatIndex(entry.Index)
		result[index] = formatLogEvent(&entry)
	}

	internalLog.Debug().Msgf("Found %d entries for correlationID:%s", len(result), correlationID)
	return result, nil
}

// GetAllEntries returns all log entries across all correlation IDs
func (mw *memoryWriter) GetAllEntries() (map[string]string, error) {
	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.GetAllEntries").GetLogger()

	internalLog.Debug().Msg("Getting all log entries")

	// Get all correlation IDs
	correlationIDs := mw.store.GetCorrelationIDs()

	result := make(map[string]string)
	for _, correlationID := range correlationIDs {
		entries, err := mw.store.GetByCorrelation(correlationID)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			key := fmt.Sprintf("%s:%010d", entry.CorrelationID, entry.Index)
			result[key] = formatLogEvent(&entry)
		}
	}

	internalLog.Debug().Msgf("Found %d total entries", len(result))
	return result, nil
}

// GetEntriesWithLevel returns log entries filtered by minimum log level
func (mw *memoryWriter) GetEntriesWithLevel(correlationID string, minLevel log.Level) (map[string]string, error) {
	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.GetEntriesWithLevel").GetLogger()

	if correlationID == "" {
		return make(map[string]string), nil
	}

	internalLog.Debug().
		Str("prefix", "GetEntriesWithLevel").
		Msgf("Getting log entries correlationID:%s minLevel:%s", correlationID, minLevel.String())

	entries, err := mw.store.GetByCorrelationWithLevel(correlationID, minLevel)
	if err != nil {
		internalLog.Error().Err(err).Msgf("Error retrieving entries for correlationID:%s", correlationID)
		return make(map[string]string), err
	}

	// Format entries
	result := make(map[string]string)
	for _, entry := range entries {
		index := formatIndex(entry.Index)
		result[index] = formatLogEvent(&entry)
	}

	internalLog.Debug().Msgf("Returning %d filtered entries (minLevel:%s)", len(result), minLevel.String())
	return result, nil
}

// GetStoredCorrelationIDs returns all correlation IDs that have stored logs
func (mw *memoryWriter) GetStoredCorrelationIDs() []string {
	return mw.store.GetCorrelationIDs()
}

// GetEntriesWithLimit retrieves the most recent log entries up to the specified limit
func (mw *memoryWriter) GetEntriesWithLimit(limit int) (map[string]string, error) {
	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.GetEntriesWithLimit").GetLogger()

	if limit <= 0 {
		return make(map[string]string), nil
	}

	internalLog.Debug().Msgf("Getting %d most recent log entries", limit)

	entries, err := mw.store.GetRecent(limit)
	if err != nil {
		internalLog.Error().Err(err).Msg("Error retrieving entries")
		return make(map[string]string), err
	}

	// Format entries
	result := make(map[string]string)
	for _, entry := range entries {
		key := fmt.Sprintf("%s:%010d", entry.CorrelationID, entry.Index)
		result[key] = formatLogEvent(&entry)
	}

	internalLog.Debug().Msgf("Returning %d entries (limit: %d)", len(result), limit)
	return result, nil
}

// GetEntriesSince retrieves all log entries since a given timestamp
func (mw *memoryWriter) GetEntriesSince(since time.Time) ([]models.LogEvent, error) {
	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.GetEntriesSince").GetLogger()

	internalLog.Debug().Msgf("Getting log entries since: %s", since.Format(time.RFC3339))

	entries, err := mw.store.GetSince(since)
	if err != nil {
		internalLog.Error().Err(err).Msg("Error retrieving entries")
		return []models.LogEvent{}, err
	}

	// Sort by timestamp
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	internalLog.Debug().Msgf("Returning %d entries since %s", len(entries), since.Format(time.RFC3339))
	return entries, nil
}

// Close closes the memory writer and underlying store
func (mw *memoryWriter) Close() error {
	if mw.store != nil {
		return mw.store.Close()
	}
	return nil
}

// formatLogEvent formats a log event for display
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

// formatIndex formats an index for consistent display
func formatIndex(index uint64) string {
	return fmt.Sprintf("%03d", index)
}
