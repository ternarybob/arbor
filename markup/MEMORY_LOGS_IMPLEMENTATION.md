# Memory Logs Implementation

## Summary

Successfully implemented the `GetMemoryLogs` function in the arbor logger package with the following features:

## Changes Made

### 1. Interface Updates (`ilogger.go`)
- Added `WithMemoryWriter(config models.WriterConfiguration) ILogger` method to the ILogger interface
- Added `GetMemoryLogs(correlationid string, minLevel LogLevel) (map[string]string, error)` method to the ILogger interface

### 2. Logger Implementation (`logger.go`)
- Implemented `WithMemoryWriter()` method that adds a memory writer to the logger's writers map
- Implemented `GetMemoryLogs()` method that:
  - Checks if a memory writer is configured (returns empty if not)
  - Converts the LogLevel parameter to the internal log.Level type
  - Calls the writers.GetEntriesWithLevel() function to retrieve filtered log entries

### 3. Memory Writer Factory (`writers/memorywriter.go`)
- Added `CreateMemoryWriter(config models.WriterConfiguration) IWriter` factory function
- This function creates a new MemoryWriter instance that implements the IWriter interface

### 4. Log Event Handling (`logevent.go`)
- Modified the `writeLog()` method to handle memory writers differently:
  - Memory writers receive JSON-serialized LogEvent objects
  - Other writers (console, file) receive formatted string output
- Added `encoding/json` import for JSON marshaling

### 5. Memory Writer Storage (`writers/memorywriter.go`)
The existing memory writer already had:
- Thread-safe in-memory storage using `logStore = make(map[string][]models.LogEvent)`
- Simple formatting for log entries stored in memory
- Level-based filtering via `GetEntriesWithLevel()` function
- Buffer limits (1000 entries per correlation ID) with FIFO behavior

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/ternarybob/arbor"
    "github.com/ternarybob/arbor/models"
    "github.com/phuslu/log"
)

func main() {
    // Create logger with memory writer
    logger := arbor.Logger().
        WithMemoryWriter(models.WriterConfiguration{
            Type:       models.LogWriterTypeMemory,
            Level:      log.InfoLevel,
            TimeFormat: "15:04:05.000",
        }).
        WithCorrelationId("my-correlation-123")

    // Log some messages
    logger.Info().Msg("This is an info message")
    logger.Warn().Msg("This is a warning message")
    logger.Error().Msg("This is an error message")

    // Retrieve memory logs with minimum level filter
    entries, err := logger.GetMemoryLogs("my-correlation-123", arbor.InfoLevel)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Found %d log entries:\n", len(entries))
    for key, value := range entries {
        fmt.Printf("[%s]: %s\n", key, value)
    }
}
```

## Key Features

1. **Level Filtering**: GetMemoryLogs accepts a minimum log level and only returns entries at or above that level
2. **Correlation ID Based**: Logs are stored and retrieved by correlation ID
3. **Thread Safe**: Memory storage uses mutexes for concurrent access
4. **Buffer Management**: Automatically manages memory with configurable buffer limits
5. **Simple Format**: Log entries are formatted as simple pipe-delimited strings for easy reading
6. **Fallback Behavior**: Returns empty results gracefully when no memory writer is configured or correlation ID doesn't exist

## Format of Retrieved Entries

Memory logs are returned as a map[string]string where:
- **Key**: 3-digit formatted index (e.g., "001", "002", "003")
- **Value**: Formatted log entry like: "INF|Jan _2 15:04:05|prefix|message|error"

Example output:
```
001: INF|Jul 6 08:20:56|This is an info message
002: WRN|Jul 6 08:20:56|This is a warning message  
003: ERR|Jul 6 08:20:56|This is an error message
```

## Integration

The implementation integrates seamlessly with the existing arbor logger architecture:
- Uses the same LogLevel constants as other parts of the system
- Follows the same fluent interface pattern
- Maintains compatibility with existing console and file writers
- Reuses existing models.LogEvent structure for consistency

## Testing

Comprehensive tests have been added to verify:
- Memory writer configuration and integration
- Log level filtering functionality
- Correlation ID-based retrieval
- Edge cases (empty correlation ID, non-existent correlation ID, no memory writer)
- Thread safety and buffer management

All existing tests continue to pass, ensuring backward compatibility.
