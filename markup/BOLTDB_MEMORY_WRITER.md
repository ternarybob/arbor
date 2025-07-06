# BoltDB Memory Writer Implementation

## Overview

The memory writer has been redesigned to use **BoltDB** for persistent, self-expiring, and self-maintaining log storage. This replaces the previous in-memory map implementation with a robust database solution that automatically handles log expiration without requiring manual cleanup functions.

## Key Features

### 1. **Self-Expiring Storage**
- Each log entry is stored with an expiration timestamp (`ExpiresAt`)
- Default TTL (Time To Live) is 24 hours
- Expired entries are automatically filtered out during retrieval
- No manual cleanup required

### 2. **Self-Maintaining Database**
- Automatic background cleanup routine runs every hour
- Removes expired and corrupted entries from the database
- Uses mutex locks to ensure thread-safe operations
- Graceful cleanup goroutine shutdown

### 3. **Persistent Storage**
- Uses BoltDB for reliable, file-based storage
- Database file stored in system temp directory: `arbor_logs.db`
- Survives application restarts (entries remain until TTL expires)
- Shared database instance across multiple memory writers

### 4. **Thread-Safe Operations**
- Global database instance management with RWMutex
- Safe concurrent access to BoltDB transactions
- Cleanup operations use separate mutex to prevent conflicts

## Architecture

### Storage Structure

```go
type StoredLogEntry struct {
    LogEvent  models.LogEvent `json:"log_event"`
    ExpiresAt time.Time       `json:"expires_at"`
}
```

### Key Design Pattern
- **Key Format**: `{correlationID}:{index}` (e.g., "user-123:0000000001")
- **Bucket**: Single bucket named "logs" contains all entries
- **Indexing**: Unique incrementing index ensures ordered retrieval

### Database Management
- Single shared BoltDB instance per database file
- Global registry (`dbInstances`) tracks open databases
- Database connections reused across memory writer instances
- Automatic bucket creation on first access

## API Changes

### Removed Methods
The following manual cleanup methods have been **removed** from `IMemoryWriter`:
- `ClearEntries(correlationID string)` 
- `ClearAllEntries()`

### Retained Methods
- `GetEntries(correlationID string) (map[string]string, error)`
- `GetEntriesWithLevel(correlationID string, minLevel log.Level) (map[string]string, error)`
- `GetStoredCorrelationIDs() []string`

### New Methods
- `Close() error` - Properly shutdown cleanup routines and resources

## Configuration

### Constants
```go
const (
    DEFAULT_TTL       = 24 * time.Hour // Default expiration time
    CLEANUP_INTERVAL  = 1 * time.Hour  // How often to clean up expired entries
    LOG_BUCKET        = "logs"         // BoltDB bucket name
)
```

### Database Location
- File: `{TempDir}/arbor_logs.db`
- Permissions: `0600` (owner read/write only)
- Timeout: 1 second for database open operations

## Usage Examples

### Basic Usage
```go
// Create memory writer
config := models.WriterConfiguration{}
writer := writers.MemoryWriter(config)
defer writer.Close() // Important: always close to cleanup resources

// Use with logger
logger := arbor.Logger().WithMemoryWriter(config)
logger = logger.WithCorrelationId("user-123")

// Log messages
logger.Info().Msg("User logged in")
logger.Warn().Msg("Password attempt failed")

// Retrieve logs
logs, err := logger.GetMemoryLogs("user-123", arbor.LogLevel(log.InfoLevel))
```

### Automatic Expiration
```go
// Logs are automatically expired after 24 hours
// No manual cleanup required

// Expired entries are filtered out during retrieval
logs, err := writer.GetEntries("correlation-id")
// Only returns non-expired entries
```

## Benefits

### 1. **Reliability**
- BoltDB provides ACID transactions
- Data survives application crashes/restarts
- Corrupted entries are automatically cleaned up

### 2. **Performance**
- Efficient B+tree storage structure
- Optimized key prefix scanning for correlation ID lookups
- Background cleanup doesn't block normal operations

### 3. **Memory Management**
- Bounded memory usage (entries automatically expire)
- No memory leaks from forgotten cleanup calls
- Efficient storage of large numbers of log entries

### 4. **Maintenance-Free**
- No manual intervention required
- Automatic cleanup of old data
- Self-healing (removes corrupted entries)

## Testing

### Comprehensive Test Suite
- **Basic functionality**: Write, read, level filtering
- **Expiration**: Automatic filtering of expired entries
- **Concurrency**: Thread-safe operations
- **Resource management**: Proper cleanup and shutdown
- **Integration**: Full logger integration testing

### Test Files
- `writers/memorywriter_boltdb_test.go` - BoltDB-specific tests
- `logger_memorywriter_integration_test.go` - End-to-end integration tests

## Migration Notes

### Breaking Changes
- `ClearEntries()` and `ClearAllEntries()` methods removed
- Constructor now returns `IMemoryWriter` interface instead of struct
- `NewMemoryWriter()` function replaced with `MemoryWriter(config)`

### Compatibility
- All existing `GetEntries*` methods maintain the same signatures
- Logger integration (`GetMemoryLogs`) unchanged
- Same correlation ID and level filtering behavior

## Performance Characteristics

### Storage
- **Write Performance**: ~1000-10000 writes/sec (depends on disk I/O)
- **Read Performance**: Very fast B+tree lookups
- **Memory Usage**: Minimal (only index and active transaction data)

### Cleanup
- **Background cleanup**: Every 1 hour (configurable)
- **Cleanup impact**: Minimal (uses separate mutex, batch operations)
- **Database size**: Automatically shrinks as expired entries are removed

## Error Handling

### Graceful Degradation
- Database open failures return `nil` writer (logged as error)
- Transaction failures are logged but don't crash the application
- Corrupted entries are skipped and marked for cleanup
- Network/disk issues are retried via BoltDB's built-in mechanisms

### Logging
- Internal operations logged via `memoryInternalLog`
- Cleanup operations logged at DEBUG level
- Errors logged at appropriate levels (WARN/ERROR)

## Future Enhancements

### Potential Improvements
1. **Configurable TTL**: Allow per-writer or per-entry TTL configuration
2. **Compression**: Compress log entries for storage efficiency
3. **Partitioning**: Multiple database files for very high-volume scenarios
4. **Metrics**: Expose storage metrics (entry count, database size, etc.)
5. **Custom cleanup schedules**: Configurable cleanup intervals

### Monitoring
Consider adding:
- Database size monitoring
- Entry count metrics
- Cleanup operation statistics
- Performance metrics (read/write latency)
