# Global Log Registry

The Global Log Registry provides a high-performance solution for accessing BoltDB-stored logs from different contexts without passing logger instances around. This is particularly useful for web services where you need to access logs in render services or other components.

## Key Features

- **High Performance**: Direct access to existing logger instances with zero additional I/O overhead
- **Thread-Safe**: Uses RWMutex for concurrent access
- **Broadcast Logging**: Write to all registered loggers simultaneously
- **Default Console Logger**: Automatically registered for backward compatibility
- **No Breaking Changes**: All existing code continues to work

## Architecture

The registry maintains a global map of named loggers that can be accessed from any context:

```go
// Register loggers during application startup
arbor.RegisterLogger("main", mainLogger)
arbor.RegisterLogger("webservice", webServiceLogger)

// Access from any context
logger := arbor.GetRegisteredLogger("main")
logs, err := logger.GetMemoryLogs(correlationID, arbor.DebugLevel)
```

## Usage Patterns

### 1. Web Service Pattern (Recommended for your use case)

**Application Startup:**
```go
// Create logger with memory writer for BoltDB storage
memoryConfig := models.WriterConfiguration{
    Type:       models.LogWriterTypeMemory,
    TimeFormat: "01-02 15:04:05.000",
}

mainLogger := arbor.Logger().
    WithMemoryWriter(memoryConfig).
    WithCorrelationId("webservice")

// Register the logger
arbor.RegisterLogger("main", mainLogger)
```

**In RenderService (omnis/service_render.go):**
```go
func (s renderservice) getApiResponse(code int) *ApiResponse {
    // Get logger with memory writer from registry
    logger := arbor.GetRegisteredLogger("main")
    if logger == nil {
        logger = arbor.GetRegisteredLogger("default") // fallback
    }
    
    if logger != nil {
        logs, err := logger.GetMemoryLogs(cid, arbor.DebugLevel)
        // Use logs in API response
    }
}
```

### 2. Multiple Logger Pattern

```go
// Register different loggers for different purposes
arbor.RegisterLogger("audit", auditLogger)      // File-based audit logs
arbor.RegisterLogger("memory", memoryLogger)    // BoltDB memory logs
arbor.RegisterLogger("console", consoleLogger)  // Console-only logs

// Access specific logger based on need
auditLogger := arbor.GetRegisteredLogger("audit")
memoryLogger := arbor.GetRegisteredLogger("memory")
```

### 3. Broadcast Logging

Write to all registered loggers simultaneously:

```go
// These messages go to ALL registered loggers
arbor.BroadcastInfo().Msg("Application started")
arbor.BroadcastError().Msgf("Critical error: %v", err)
arbor.BroadcastWarn().Msg("Configuration warning")
```

## API Reference

### Registration Functions

```go
// Register a logger with a name
func RegisterLogger(name string, logger ILogger)

// Get a logger by name (returns nil if not found)
func GetRegisteredLogger(name string) ILogger

// Get all registered logger names
func GetRegisteredLoggerNames() []string

// Get count of registered loggers
func GetLoggerCount() int

// Remove a logger from registry
func UnregisterLogger(name string)
```

### Broadcast Functions

```go
// Broadcast logging functions
func BroadcastTrace() *broadcastLogEvent
func BroadcastDebug() *broadcastLogEvent
func BroadcastInfo() *broadcastLogEvent
func BroadcastWarn() *broadcastLogEvent
func BroadcastError() *broadcastLogEvent
func BroadcastFatal() *broadcastLogEvent
func BroadcastPanic() *broadcastLogEvent
```

### Default Loggers

The registry automatically registers:
- `"default"` - The default console logger
- `"console"` - Alias for the default console logger

## Performance Benefits

1. **Zero Additional I/O**: Uses existing logger instances and BoltDB connections
2. **Minimal Memory Overhead**: Just a map lookup operation
3. **No Abstraction Penalty**: Direct access to logger methods
4. **Thread-Safe Concurrent Access**: RWMutex allows multiple readers
5. **Reuses Existing Infrastructure**: Leverages existing MemoryWriter BoltDB connections

## Migration Guide

### Before (Problematic)
```go
// This was calling the wrong log package
logs, err := log.GetMemoryLogs(cid, levels.DebugLevel) // ERROR: doesn't exist
```

### After (Working)
```go
// Use the registry to get the appropriate logger
logger := arbor.GetRegisteredLogger("main")
if logger != nil {
    logs, err := logger.GetMemoryLogs(cid, arbor.DebugLevel)
}
```

## Best Practices

1. **Register loggers during application startup** before any concurrent access
2. **Use descriptive names** like "main", "audit", "webservice" rather than generic names
3. **Always check for nil** when getting loggers from the registry
4. **Use broadcast logging sparingly** - it writes to ALL loggers
5. **Register memory-enabled loggers** with names that indicate their purpose

## Example: Complete Web Service Setup

```go
package main

import (
    "github.com/ternarybob/arbor"
    "github.com/ternarybob/arbor/models"
)

func main() {
    // 1. Create logger with memory writer
    memoryConfig := models.WriterConfiguration{
        Type:       models.LogWriterTypeMemory,
        TimeFormat: "01-02 15:04:05.000",
    }
    
    webLogger := arbor.Logger().
        WithMemoryWriter(memoryConfig).
        WithCorrelationId("web-service")
    
    // 2. Register the logger
    arbor.RegisterLogger("webservice", webLogger)
    
    // 3. Start your web service
    // Now RenderService can access logs via:
    // arbor.GetRegisteredLogger("webservice")
    
    startWebServer()
}
```

This approach provides the highest performance solution for accessing BoltDB logs across different contexts while maintaining clean architecture and backward compatibility.
