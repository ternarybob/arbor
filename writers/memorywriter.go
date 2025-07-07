package writers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
	"go.etcd.io/bbolt"
)

const (
	CORRELATIONID_KEY = "CORRELATIONID"
	BUFFER_LIMIT      = 1000             // Maximum entries per correlation ID
	DEFAULT_TTL       = 10 * time.Minute // Default expiration time
	CLEANUP_INTERVAL  = 1 * time.Minute  // How often to clean up expired entries
	LOG_BUCKET        = "logs"           // BoltDB bucket name
)

var (
	// Global state for shared access
	indexCounter uint64 = 0
	counterMux   sync.Mutex
	cleanupMux   sync.Mutex
	dbInstances  map[string]*bbolt.DB = make(map[string]*bbolt.DB)
	dbMux        sync.RWMutex

	// Internal logger for debugging
	memoryLogLevel log.Level = log.DebugLevel
)

type memoryWriter struct {
	config        models.WriterConfiguration
	db            *bbolt.DB
	dbPath        string
	ttl           time.Duration
	cleanupTicker *time.Ticker
	cleanupStop   chan bool
}

type StoredLogEntry struct {
	LogEvent  models.LogEvent `json:"log_event"`
	ExpiresAt time.Time       `json:"expires_at"`
}

func MemoryWriter(config models.WriterConfiguration) IMemoryWriter {

	internalLog := common.NewLogger().WithContext("function", "MemoryWriter").GetLogger()

	// Get executable directory and create temp subdirectory
	execDir, err := os.Getwd()
	if err != nil {
		internalLog.Fatal().Err(err).Msg("Failed to get current working directory")
	}
	tempDir := filepath.Join(execDir, "temp")

	// Ensure temp directory exists
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		internalLog.Fatal().Err(err).Msg("Failed to create temp directory")
	}

	// Create date-based database filename (YYMMDD format)
	now := time.Now()
	dateStr := now.Format("060102") // YYMMDD format
	dbPath := filepath.Join(tempDir, fmt.Sprintf("arbor_logs_%s.db", dateStr))

	// Check if we already have this database instance
	dbMux.RLock()
	db, exists := dbInstances[dbPath]
	dbMux.RUnlock()

	if !exists {
		// Create new database instance with fatal error handling for conflicts
		newDB, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
			Timeout: 5 * time.Second,
		})
		if err != nil {
			internalLog.Fatal().Err(err).Str("db_path", dbPath).Msg("Failed to open BoltDB - database conflict detected")
		}

		// Create bucket if it doesn't exist
		err = newDB.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(LOG_BUCKET))
			return err
		})
		if err != nil {
			internalLog.Fatal().Err(err).Msg("Failed to create bucket")
		}

		// Store in global instances
		dbMux.Lock()
		dbInstances[dbPath] = newDB
		dbMux.Unlock()

		db = newDB
	}

	mw := &memoryWriter{
		config:      config,
		db:          db,
		dbPath:      dbPath,
		ttl:         DEFAULT_TTL,
		cleanupStop: make(chan bool),
	}

	// Start cleanup routine
	mw.startCleanup()

	return mw
}

// WithLevel sets the log level for the memory writer (required by IWriter interface)
func (mw *memoryWriter) WithLevel(level log.Level) IWriter {
	// Update the config level
	mw.config.Level = levels.FromLogLevel(level)
	return mw
}

func (mw *memoryWriter) Write(entry []byte) (int, error) {

	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.Write").GetLogger()
	ep := len(entry)

	if ep == 0 {
		return ep, nil
	}

	err := mw.writeline(entry)
	if err != nil {
		internalLog.Warn().Err(err).Msg("Error when actioning writeline")
	}

	return ep, nil
}

func (mw *memoryWriter) writeline(event []byte) error {

	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.writeline").GetLogger()

	if len(event) <= 0 {
		return nil // Don't error on empty events
	}

	var logentry models.LogEvent

	if err := json.Unmarshal(event, &logentry); err != nil {
		internalLog.Warn().Err(err).Msgf("Error:%s Event:%s", err.Error(), string(event))
		return err
	}

	if isEmpty(logentry.CorrelationID) {
		internalLog.Trace().Msgf("CorrelationID is empty -> no write to memory store -> return")
		return nil
	}

	// Generate unique index
	counterMux.Lock()
	indexCounter++
	logentry.Index = indexCounter
	counterMux.Unlock()

	// Create stored entry with expiration
	storedEntry := StoredLogEntry{
		LogEvent:  logentry,
		ExpiresAt: time.Now().Add(mw.ttl),
	}

	// Store in BoltDB
	return mw.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(LOG_BUCKET))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", LOG_BUCKET)
		}

		// Create key using correlation ID and index
		key := fmt.Sprintf("%s:%010d", logentry.CorrelationID, logentry.Index)

		// Serialize the stored entry
		data, err := json.Marshal(storedEntry)
		if err != nil {
			return err
		}

		// Store in BoltDB
		err = bucket.Put([]byte(key), data)
		if err != nil {
			return err
		}

		internalLog.Trace().Msgf("CorrelationID:%s -> message:%s (key: %s)",
			logentry.CorrelationID, logentry.Message, key)

		return nil
	})
}

func (mw *memoryWriter) GetEntries(correlationid string) (map[string]string, error) {

	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.GetEntries").GetLogger()
	entries := make(map[string]string)

	if correlationid == "" {
		return entries, nil // Return empty instead of error
	}

	internalLog.Debug().Msgf("Getting log entries correlationid:%s", correlationid)

	err := mw.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(LOG_BUCKET))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", LOG_BUCKET)
		}

		// Iterate through all keys with the correlation ID prefix
		prefix := []byte(correlationid + ":")
		c := bucket.Cursor()

		for k, v := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, v = c.Next() {
			var storedEntry StoredLogEntry
			if err := json.Unmarshal(v, &storedEntry); err != nil {
				internalLog.Warn().Err(err).Msgf("Failed to unmarshal entry for key %s", string(k))
				continue
			}

			// Check if entry is expired
			if time.Now().After(storedEntry.ExpiresAt) {
				internalLog.Debug().Msgf("Entry expired for key %s", string(k))
				continue
			}

			index := formatIndex(storedEntry.LogEvent.Index)
			entries[index] = formatLogEvent(&storedEntry.LogEvent)
		}

		return nil
	})

	if err != nil {
		internalLog.Error().Err(err).Msgf("Error retrieving entries for correlationid:%s", correlationid)
		return entries, err
	}

	internalLog.Debug().Msgf("Found %d entries for correlationid:%s", len(entries), correlationid)
	return entries, nil
}

// GetAllEntries returns all log entries across all correlation IDs
func (mw *memoryWriter) GetAllEntries() (map[string]string, error) {

	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.GetAllEntries").GetLogger()
	entries := make(map[string]string)

	internalLog.Debug().Msg("Getting all log entries")

	err := mw.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(LOG_BUCKET))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", LOG_BUCKET)
		}

		// Iterate through all keys in the bucket
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var storedEntry StoredLogEntry
			if err := json.Unmarshal(v, &storedEntry); err != nil {
				internalLog.Warn().Err(err).Msgf("Failed to unmarshal entry for key %s", string(k))
				continue
			}

			// Check if entry is expired
			if time.Now().After(storedEntry.ExpiresAt) {
				internalLog.Debug().Msgf("Entry expired for key %s", string(k))
				continue
			}

			// Use the full key as the map key to ensure uniqueness across correlation IDs
			key := string(k)
			entries[key] = formatLogEvent(&storedEntry.LogEvent)
		}

		return nil
	})

	if err != nil {
		internalLog.Error().Err(err).Msg("Error retrieving all entries")
		return entries, err
	}

	internalLog.Debug().Msgf("Found %d total entries", len(entries))
	return entries, nil
}

// GetEntriesWithLevel returns log entries filtered by minimum log level
func (mw *memoryWriter) GetEntriesWithLevel(correlationid string, minLevel log.Level) (map[string]string, error) {

	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.GetEntriesWithLevel").GetLogger()
	entries := make(map[string]string)

	if correlationid == "" {
		return entries, nil // Return empty instead of error
	}

	internalLog.Debug().
		Str("prefix", "GetEntriesWithLevel").
		Msgf("Getting log entries correlationid:%s minLevel:%s", correlationid, minLevel.String())

	err := mw.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(LOG_BUCKET))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", LOG_BUCKET)
		}

		// Iterate through all keys with the correlation ID prefix
		prefix := []byte(correlationid + ":")
		c := bucket.Cursor()

		for k, v := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, v = c.Next() {
			var storedEntry StoredLogEntry
			if err := json.Unmarshal(v, &storedEntry); err != nil {
				internalLog.Warn().Err(err).Msgf("Failed to unmarshal entry for key %s", string(k))
				continue
			}

			// Check if entry is expired
			if time.Now().After(storedEntry.ExpiresAt) {
				internalLog.Debug().Msgf("Entry expired for key %s", string(k))
				continue
			}

			// Filter by level
			eventLevel := storedEntry.LogEvent.Level
			if eventLevel >= minLevel {
				index := formatIndex(storedEntry.LogEvent.Index)
				entries[index] = formatLogEvent(&storedEntry.LogEvent)
			}
		}

		return nil
	})

	if err != nil {
		internalLog.Error().Err(err).Msgf("Error retrieving entries for correlationid:%s", correlationid)
		return entries, err
	}

	internalLog.Debug().Msgf("Returning %d filtered entries (minLevel:%s)", len(entries), minLevel.String())
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

// GetStoredCorrelationIDs returns all correlation IDs that have stored logs
func (mw *memoryWriter) GetStoredCorrelationIDs() []string {
	var ids []string
	idMap := make(map[string]bool)

	mw.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(LOG_BUCKET))
		if bucket == nil {
			return nil
		}

		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var storedEntry StoredLogEntry
			if err := json.Unmarshal(v, &storedEntry); err != nil {
				continue
			}

			// Check if entry is expired
			if time.Now().After(storedEntry.ExpiresAt) {
				continue
			}

			// Extract correlation ID from key (format: "correlationid:index")
			key := string(k)
			if colonIndex := strings.Index(key, ":"); colonIndex > 0 {
				correlationID := key[:colonIndex]
				if !idMap[correlationID] {
					idMap[correlationID] = true
					ids = append(ids, correlationID)
				}
			}
		}

		return nil
	})

	return ids
}

// GetEntriesWithLimit retrieves the most recent log entries up to the specified limit
func (mw *memoryWriter) GetEntriesWithLimit(limit int) (map[string]string, error) {

	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.GetEntriesWithLimit").GetLogger()

	if limit <= 0 {
		return make(map[string]string), nil
	}

	internalLog.Debug().Msgf("Getting %d most recent log entries", limit)

	// First, collect all entries with their timestamps
	type entryWithTime struct {
		key       string
		value     string
		timestamp time.Time
		index     uint64
	}

	var allEntries []entryWithTime

	err := mw.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(LOG_BUCKET))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", LOG_BUCKET)
		}

		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var storedEntry StoredLogEntry
			if err := json.Unmarshal(v, &storedEntry); err != nil {
				internalLog.Warn().Err(err).Msgf("Failed to unmarshal entry for key %s", string(k))
				continue
			}

			// Check if entry is expired
			if time.Now().After(storedEntry.ExpiresAt) {
				internalLog.Debug().Msgf("Entry expired for key %s", string(k))
				continue
			}

			allEntries = append(allEntries, entryWithTime{
				key:       string(k),
				value:     formatLogEvent(&storedEntry.LogEvent),
				timestamp: storedEntry.LogEvent.Timestamp,
				index:     storedEntry.LogEvent.Index,
			})
		}

		return nil
	})

	if err != nil {
		internalLog.Error().Err(err).Msg("Error retrieving entries")
		return make(map[string]string), err
	}

	// Sort by timestamp descending (most recent first), then by index descending as tiebreaker
	sort.Slice(allEntries, func(i, j int) bool {
		// First compare by timestamp (newer first)
		if !allEntries[i].timestamp.Equal(allEntries[j].timestamp) {
			return allEntries[i].timestamp.After(allEntries[j].timestamp)
		}
		// If timestamps are equal, compare by index (higher index first)
		return allEntries[i].index > allEntries[j].index
	})

	// Take only the first 'limit' entries
	entries := make(map[string]string)
	for i := 0; i < len(allEntries) && i < limit; i++ {
		entries[allEntries[i].key] = allEntries[i].value
	}

	internalLog.Debug().Msgf("Returning %d entries (limit: %d)", len(entries), limit)
	return entries, nil
}

// Close closes the memory writer and any underlying resources
func (mw *memoryWriter) Close() error {
	// Stop cleanup routine
	if mw.cleanupTicker != nil {
		mw.cleanupTicker.Stop()
		mw.cleanupTicker = nil
	}
	if mw.cleanupStop != nil {
		select {
		case <-mw.cleanupStop:
			// Channel already closed
		default:
			close(mw.cleanupStop)
		}
		mw.cleanupStop = nil
	}

	// Close the database and remove from instances
	if mw.db != nil {
		dbMux.Lock()
		defer dbMux.Unlock()

		// Close database
		mw.db.Close()

		// Remove from global instances
		delete(dbInstances, mw.dbPath)

		// Note: Database files are kept for persistence across application restarts
		// They will be reused if the application starts again on the same day
		// Old files can be cleaned up manually or via external cleanup scripts
	}

	return nil
}

// startCleanup starts the automatic cleanup routine
func (mw *memoryWriter) startCleanup() {

	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.startCleanup").GetLogger()

	mw.cleanupTicker = time.NewTicker(CLEANUP_INTERVAL)
	cleanupTicker := mw.cleanupTicker
	go func() {
		defer func() {
			if r := recover(); r != nil {
				internalLog.Debug().Msgf("Cleanup goroutine exited: %v", r)
			}
		}()
		for {
			select {
			case <-cleanupTicker.C:
				mw.cleanupExpiredEntries()
			case <-mw.cleanupStop:
				return
			}
		}
	}()
}

// cleanupExpiredEntries removes expired log entries from BoltDB
func (mw *memoryWriter) cleanupExpiredEntries() {

	internalLog := common.NewLogger().WithContext("function", "MemoryWriter.cleanupExpiredEntries").GetLogger()

	cleanupMux.Lock()
	defer cleanupMux.Unlock()

	now := time.Now()
	var keysToDelete [][]byte

	// First pass: identify expired keys
	mw.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(LOG_BUCKET))
		if bucket == nil {
			return nil
		}

		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var storedEntry StoredLogEntry
			if err := json.Unmarshal(v, &storedEntry); err != nil {
				// If we can't unmarshal, consider it corrupted and mark for deletion
				keysToDelete = append(keysToDelete, append([]byte(nil), k...))
				continue
			}

			if now.After(storedEntry.ExpiresAt) {
				keysToDelete = append(keysToDelete, append([]byte(nil), k...))
			}
		}

		return nil
	})

	// Second pass: delete expired keys
	if len(keysToDelete) > 0 {
		mw.db.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte(LOG_BUCKET))
			if bucket == nil {
				return nil
			}

			for _, key := range keysToDelete {
				bucket.Delete(key)
			}

			return nil
		})

		internalLog.Debug().Msgf("Cleaned up %d expired entries", len(keysToDelete))
	}
}
