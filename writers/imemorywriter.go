package writers

import "github.com/phuslu/log"

// IMemoryWriter extends IWriter with memory-specific operations
type IMemoryWriter interface {
	IWriter

	// GetEntries retrieves all log entries for a specific correlation ID
	GetEntries(correlationID string) (map[string]string, error)

	// GetAllEntries retrieves all log entries across all correlation IDs
	GetAllEntries() (map[string]string, error)

	// GetEntriesWithLevel retrieves log entries filtered by minimum log level
	GetEntriesWithLevel(correlationID string, minLevel log.Level) (map[string]string, error)

	// GetStoredCorrelationIDs returns all correlation IDs that have stored logs
	GetStoredCorrelationIDs() []string

	// GetEntriesWithLimit retrieves the most recent log entries up to the specified limit
	GetEntriesWithLimit(limit int) (map[string]string, error)

	// Close closes the memory writer and any underlying resources
	Close() error
}
