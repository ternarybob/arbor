# Memory Writer Interface Documentation

## Overview

The memory writer in the Arbor logging library now implements a dual interface structure:

1. **IWriter** - Basic writer interface with standard methods
2. **IMemoryWriter** - Extended interface with memory-specific operations

## Interface Definitions

### IWriter Interface
```go
type IWriter interface {
    WithLevel(level log.Level) IWriter
    Write(p []byte) (n int, err error)
}
```

### IMemoryWriter Interface
```go
type IMemoryWriter interface {
    IWriter
    
    // GetEntries retrieves all log entries for a specific correlation ID
    GetEntries(correlationID string) (map[string]string, error)
    
    // GetEntriesWithLevel retrieves log entries filtered by minimum log level
    GetEntriesWithLevel(correlationID string, minLevel log.Level) (map[string]string, error)
    
    // ClearEntries removes log entries for a specific correlation ID
    ClearEntries(correlationID string)
    
    // ClearAllEntries removes all stored log entries
    ClearAllEntries()
    
    // GetStoredCorrelationIDs returns all correlation IDs that have stored logs
    GetStoredCorrelationIDs() []string
}
```

## Usage Examples

### Creating a Memory Writer
```go
import (
    "github.com/phuslu/log"
    "github.com/ternarybob/arbor/models"
    "github.com/ternarybob/arbor/writers"
)

// Create configuration
config := models.WriterConfiguration{
    Level:      log.InfoLevel,
    TimeFormat: "2006-01-02 15:04:05",
}

// Create memory writer
memWriter := writers.MemoryWriter(config)

// Use as basic IWriter
memWriter.WithLevel(log.DebugLevel)

// Use memory-specific methods
entries, err := memWriter.GetEntries("correlation-123")
filteredEntries, err := memWriter.GetEntriesWithLevel("correlation-123", log.WarnLevel)
correlationIDs := memWriter.GetStoredCorrelationIDs()
memWriter.ClearEntries("correlation-123")
memWriter.ClearAllEntries()
```

### Integration with Logger
```go
import "github.com/ternarybob/arbor"

// Create logger with memory writer
logger := arbor.New().WithMemoryWriter(config)

// Log some entries
logger.Info().CorrelationID("test-123").Msg("Test message")
logger.Warn().CorrelationID("test-123").Msg("Warning message")

// Access memory-specific functionality
// Note: Need to cast to IMemoryWriter to access memory-specific methods
if memWriter, ok := logger.GetWriter("memory").(writers.IMemoryWriter); ok {
    entries, err := memWriter.GetEntries("test-123")
    // Process entries...
}
```

## Key Benefits

1. **Interface Segregation**: Basic writers only need to implement IWriter
2. **Memory-Specific Operations**: Memory writers provide additional functionality through IMemoryWriter
3. **Type Safety**: Clear distinction between basic and memory writers at compile time
4. **Flexibility**: Can use memory writer as either basic IWriter or full IMemoryWriter depending on needs

## Implementation Details

- **Global State**: Memory writer uses global state for shared access across instances
- **Thread Safety**: All operations are protected by mutexes for concurrent access
- **Buffer Management**: Automatic FIFO buffer management with configurable limits
- **Level Filtering**: Supports filtering log entries by minimum log level
- **Correlation ID Management**: Entries are organized by correlation ID for easy retrieval

## Method Descriptions

### GetEntries(correlationID string) (map[string]string, error)
Retrieves all log entries for a specific correlation ID as formatted strings.

### GetEntriesWithLevel(correlationID string, minLevel log.Level) (map[string]string, error)
Retrieves log entries filtered by minimum log level. Only entries at or above the specified level are returned.

### ClearEntries(correlationID string)
Removes all log entries associated with a specific correlation ID.

### ClearAllEntries()
Removes all stored log entries from memory. Useful for cleanup operations.

### GetStoredCorrelationIDs() []string
Returns a list of all correlation IDs that currently have stored log entries.

This interface structure provides a clean separation of concerns while maintaining compatibility with the existing logging infrastructure.
