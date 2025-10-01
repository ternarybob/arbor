// -----------------------------------------------------------------------
// Last Modified: Wednesday, 1st October 2025 3:50:00 pm
// Modified By: Bob McAllan
// -----------------------------------------------------------------------

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
	"github.com/ternarybob/arbor/models"
	"go.etcd.io/bbolt"
)

const (
	DEFAULT_TTL      = 10 * time.Minute
	CLEANUP_INTERVAL = 1 * time.Minute
	LOG_BUCKET       = "logs"
)

// InMemoryLogStore provides fast in-memory log storage with optional BoltDB persistence
type inMemoryLogStore struct {
	// In-memory storage (primary)
	entries    map[string][]models.LogEvent // correlationID -> events
	allEntries []models.LogEvent            // All entries for timestamp queries
	entriesMux sync.RWMutex

	// Configuration
	ttl               time.Duration
	enablePersistence bool
	dbPath            string

	// Optional BoltDB persistence
	db            *bbolt.DB
	persistBuffer chan models.LogEvent

	// Cleanup
	cleanupTicker *time.Ticker
	cleanupStop   chan bool
	indexCounter  uint64
}

// StoredLogEntry wraps a log event with expiration metadata
type StoredLogEntry struct {
	LogEvent  models.LogEvent `json:"log_event"`
	ExpiresAt time.Time       `json:"expires_at"`
}

// NewInMemoryLogStore creates a new in-memory log store
func NewInMemoryLogStore(config models.WriterConfiguration) (ILogStore, error) {
	internalLog := common.NewLogger().WithContext("function", "NewInMemoryLogStore").GetLogger()

	store := &inMemoryLogStore{
		entries:           make(map[string][]models.LogEvent),
		allEntries:        make([]models.LogEvent, 0),
		ttl:               DEFAULT_TTL,
		enablePersistence: config.DBPath != "",
		dbPath:            config.DBPath,
		persistBuffer:     make(chan models.LogEvent, 1000),
		cleanupStop:       make(chan bool),
		indexCounter:      0,
	}

	// Initialize BoltDB if persistence is enabled
	if store.enablePersistence {
		if err := store.initPersistence(); err != nil {
			internalLog.Error().Err(err).Msg("Failed to initialize persistence, continuing in-memory only")
			store.enablePersistence = false
		} else {
			// Start persistence worker
			go store.persistWorker()
			internalLog.Info().Str("path", store.dbPath).Msg("BoltDB persistence enabled")
		}
	}

	// Start cleanup routine
	store.startCleanup()

	return store, nil
}

// initPersistence sets up BoltDB
func (s *inMemoryLogStore) initPersistence() error {
	// Ensure directory exists
	dir := filepath.Dir(s.dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create persistence directory: %w", err)
	}

	// Create date-based database filename if not fully specified
	if !strings.HasSuffix(s.dbPath, ".db") {
		now := time.Now()
		dateStr := now.Format("060102") // YYMMDD format
		s.dbPath = filepath.Join(s.dbPath, fmt.Sprintf("arbor_logs_%s.db", dateStr))
	}

	// Open BoltDB
	db, err := bbolt.Open(s.dbPath, 0600, &bbolt.Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to open BoltDB: %w", err)
	}

	// Create bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(LOG_BUCKET))
		return err
	})
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	s.db = db
	return nil
}

// persistWorker handles async writes to BoltDB
func (s *inMemoryLogStore) persistWorker() {
	for entry := range s.persistBuffer {
		s.persistToDB(entry)
	}
}

// persistToDB writes a single entry to BoltDB
func (s *inMemoryLogStore) persistToDB(entry models.LogEvent) error {
	if s.db == nil {
		return nil
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(LOG_BUCKET))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", LOG_BUCKET)
		}

		storedEntry := StoredLogEntry{
			LogEvent:  entry,
			ExpiresAt: time.Now().Add(s.ttl),
		}

		key := fmt.Sprintf("%s:%010d", entry.CorrelationID, entry.Index)
		data, err := json.Marshal(storedEntry)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), data)
	})
}

// Store adds a log entry to in-memory store and optionally persists to BoltDB
func (s *inMemoryLogStore) Store(entry models.LogEvent) error {
	// Assign index
	s.entriesMux.Lock()
	s.indexCounter++
	entry.Index = s.indexCounter
	s.entriesMux.Unlock()

	// Store in memory (fast, primary storage)
	s.entriesMux.Lock()
	if entry.CorrelationID != "" {
		s.entries[entry.CorrelationID] = append(s.entries[entry.CorrelationID], entry)
	}
	s.allEntries = append(s.allEntries, entry)
	s.entriesMux.Unlock()

	// Async persist to BoltDB if enabled (non-blocking)
	if s.enablePersistence {
		select {
		case s.persistBuffer <- entry:
			// Buffered successfully
		default:
			// Buffer full, skip persistence for this entry
		}
	}

	return nil
}

// GetByCorrelation retrieves all logs for a correlation ID, ordered by timestamp
func (s *inMemoryLogStore) GetByCorrelation(correlationID string) ([]models.LogEvent, error) {
	if correlationID == "" {
		return []models.LogEvent{}, nil
	}

	s.entriesMux.RLock()
	entries, exists := s.entries[correlationID]
	s.entriesMux.RUnlock()

	if !exists {
		return []models.LogEvent{}, nil
	}

	// Return a copy sorted by timestamp
	result := make([]models.LogEvent, len(entries))
	copy(result, entries)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result, nil
}

// GetByCorrelationWithLevel retrieves logs filtered by minimum level
func (s *inMemoryLogStore) GetByCorrelationWithLevel(correlationID string, minLevel log.Level) ([]models.LogEvent, error) {
	allEntries, err := s.GetByCorrelation(correlationID)
	if err != nil {
		return nil, err
	}

	// Filter by level
	filtered := make([]models.LogEvent, 0, len(allEntries))
	for _, entry := range allEntries {
		if entry.Level >= minLevel {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// GetSince retrieves all logs since a given timestamp
func (s *inMemoryLogStore) GetSince(since time.Time) ([]models.LogEvent, error) {
	s.entriesMux.RLock()
	defer s.entriesMux.RUnlock()

	result := make([]models.LogEvent, 0)
	for _, entry := range s.allEntries {
		if entry.Timestamp.After(since) {
			result = append(result, entry)
		}
	}

	// Sort by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result, nil
}

// GetRecent retrieves the N most recent log entries
func (s *inMemoryLogStore) GetRecent(limit int) ([]models.LogEvent, error) {
	if limit <= 0 {
		return []models.LogEvent{}, nil
	}

	s.entriesMux.RLock()
	defer s.entriesMux.RUnlock()

	// Sort all entries by timestamp descending
	sorted := make([]models.LogEvent, len(s.allEntries))
	copy(sorted, s.allEntries)

	sort.Slice(sorted, func(i, j int) bool {
		if !sorted[i].Timestamp.Equal(sorted[j].Timestamp) {
			return sorted[i].Timestamp.After(sorted[j].Timestamp)
		}
		return sorted[i].Index > sorted[j].Index
	})

	// Return up to limit entries
	if len(sorted) > limit {
		sorted = sorted[:limit]
	}

	return sorted, nil
}

// GetCorrelationIDs returns all active correlation IDs
func (s *inMemoryLogStore) GetCorrelationIDs() []string {
	s.entriesMux.RLock()
	defer s.entriesMux.RUnlock()

	ids := make([]string, 0, len(s.entries))
	for id := range s.entries {
		ids = append(ids, id)
	}

	return ids
}

// startCleanup starts the automatic cleanup routine
func (s *inMemoryLogStore) startCleanup() {
	s.cleanupTicker = time.NewTicker(CLEANUP_INTERVAL)
	go func() {
		for {
			select {
			case <-s.cleanupTicker.C:
				s.cleanupExpiredEntries()
			case <-s.cleanupStop:
				return
			}
		}
	}()
}

// cleanupExpiredEntries removes entries older than TTL
func (s *inMemoryLogStore) cleanupExpiredEntries() {
	internalLog := common.NewLogger().WithContext("function", "inMemoryLogStore.cleanupExpiredEntries").GetLogger()

	now := time.Now()
	cutoff := now.Add(-s.ttl)

	s.entriesMux.Lock()
	defer s.entriesMux.Unlock()

	// Clean correlation-based entries
	for correlationID, entries := range s.entries {
		filtered := make([]models.LogEvent, 0, len(entries))
		for _, entry := range entries {
			if entry.Timestamp.After(cutoff) {
				filtered = append(filtered, entry)
			}
		}

		if len(filtered) == 0 {
			delete(s.entries, correlationID)
		} else {
			s.entries[correlationID] = filtered
		}
	}

	// Clean all entries
	filtered := make([]models.LogEvent, 0, len(s.allEntries))
	for _, entry := range s.allEntries {
		if entry.Timestamp.After(cutoff) {
			filtered = append(filtered, entry)
		}
	}
	s.allEntries = filtered

	// Clean BoltDB if persistence enabled
	if s.enablePersistence && s.db != nil {
		go s.cleanupBoltDB(cutoff)
	}

	internalLog.Debug().Msgf("Cleaned up expired entries (cutoff: %s)", cutoff.Format(time.RFC3339))
}

// cleanupBoltDB removes expired entries from BoltDB
func (s *inMemoryLogStore) cleanupBoltDB(cutoff time.Time) {
	if s.db == nil {
		return
	}

	var keysToDelete [][]byte

	// Identify expired keys
	s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(LOG_BUCKET))
		if bucket == nil {
			return nil
		}

		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var storedEntry StoredLogEntry
			if err := json.Unmarshal(v, &storedEntry); err != nil {
				keysToDelete = append(keysToDelete, append([]byte(nil), k...))
				continue
			}

			if time.Now().After(storedEntry.ExpiresAt) || storedEntry.LogEvent.Timestamp.Before(cutoff) {
				keysToDelete = append(keysToDelete, append([]byte(nil), k...))
			}
		}

		return nil
	})

	// Delete expired keys
	if len(keysToDelete) > 0 {
		s.db.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte(LOG_BUCKET))
			if bucket == nil {
				return nil
			}

			for _, key := range keysToDelete {
				bucket.Delete(key)
			}

			return nil
		})
	}
}

// Close shuts down the log store
func (s *inMemoryLogStore) Close() error {
	// Stop cleanup
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}
	if s.cleanupStop != nil {
		close(s.cleanupStop)
	}

	// Close persist buffer
	if s.persistBuffer != nil {
		close(s.persistBuffer)
	}

	// Close BoltDB
	if s.db != nil {
		return s.db.Close()
	}

	return nil
}
