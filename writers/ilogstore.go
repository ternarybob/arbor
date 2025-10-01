// -----------------------------------------------------------------------
// Last Modified: Wednesday, 1st October 2025 3:45:00 pm
// Modified By: Bob McAllan
// -----------------------------------------------------------------------

package writers

import (
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/models"
)

// ILogStore defines interface for queryable log storage
// Implementations can be in-memory only, or backed by persistence (BoltDB)
type ILogStore interface {
	// Store adds a log entry to the store (non-blocking, async safe)
	Store(entry models.LogEvent) error

	// GetByCorrelation retrieves all logs for a correlation ID, ordered by timestamp
	GetByCorrelation(correlationID string) ([]models.LogEvent, error)

	// GetByCorrelationWithLevel retrieves logs for a correlation ID filtered by minimum level
	GetByCorrelationWithLevel(correlationID string, minLevel log.Level) ([]models.LogEvent, error)

	// GetSince retrieves all logs since a given timestamp
	GetSince(since time.Time) ([]models.LogEvent, error)

	// GetRecent retrieves the N most recent log entries
	GetRecent(limit int) ([]models.LogEvent, error)

	// GetCorrelationIDs returns all active correlation IDs
	GetCorrelationIDs() []string

	// Close cleans up resources
	Close() error
}
