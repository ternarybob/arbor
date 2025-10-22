# arbor

[![CI/CD Pipeline](https://github.com/ternarybob/arbor/actions/workflows/ci.yml/badge.svg)](https://github.com/ternarybob/arbor/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ternarybob/arbor.svg)](https://pkg.go.dev/github.com/ternarybob/arbor)
[![Go Report Card](https://goreportcard.com/badge/github.com/ternarybob/arbor)](https://goreportcard.com/report/github.com/ternarybob/arbor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive Go logging system designed for APIs with structured logging, multiple output writers, and advanced correlation tracking.

## Installation

```bash
go get github.com/ternarybob/arbor@latest
```

## Quick Start

```go
package main

import (
    "github.com/ternarybob/arbor"
    "github.com/ternarybob/arbor/models"
)

func main() {
    // Create logger with console output
    logger := arbor.Logger().
        WithConsoleWriter(models.WriterConfiguration{
            Type:       models.LogWriterTypeConsole,
            TimeFormat: "15:04:05.000",
        }).
        WithCorrelationId("app-startup")

    // Log messages
    logger.Info().Str("version", "1.0.0").Msg("Application started")
    logger.Warn().Msg("This is a warning")
    logger.Error().Str("error", "connection failed").Msg("Database connection error")
}
```

## Features

- **Multi-Writer Architecture**: Console, File, and Memory writers with shared log store
- **Correlation ID Tracking**: Request tracing across application layers
- **Structured Logging**: Rich field support with fluent API
- **Log Level Management**: String-based and programmatic level configuration
- **In-Memory Log Store**: Fast queryable storage with optional BoltDB persistence
- **WebSocket Support**: Real-time log streaming to connected clients
- **API Integration**: Built-in Gin framework support
- **Global Registry**: Cross-context logger access
- **Thread-Safe**: Concurrent access with proper synchronization
- **Performance Focused**: Non-blocking async writes, optimized for high-throughput API scenarios

## Basic Usage

### Simple Console Logging

```go
import "github.com/ternarybob/arbor"

// Use global logger
arbor.Info().Msg("Simple log message")
arbor.Error().Err(err).Msg("Error occurred")

// With structured fields
arbor.Info().
    Str("user", "john.doe").
    Int("attempts", 3).
    Msg("User login attempt")
```

### Multiple Writers Configuration

```go
logger := arbor.Logger().
    WithConsoleWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeConsole,
        TimeFormat: "15:04:05.000",
    }).
    WithFileWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeFile,
        FileName:   "logs/app.log",
        MaxSize:    500 * 1024, // 500KB (default) - AI-friendly size
        MaxBackups: 20,         // 20 files (default) - maintains ~10MB history
        TimeFormat: "2006-01-02 15:04:05.000",
        TextOutput: true, // Enable human-readable text format (default: false for JSON)
    }).
    WithMemoryWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeMemory,
        TimeFormat: "15:04:05.000",
    })

logger.Info().Msg("This goes to console, file, and memory")
```

## File Writer Configuration

The file writer supports both JSON and human-readable text output formats.

### JSON Output Format (Default)

```go
logger := arbor.Logger().
    WithFileWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeFile,
        FileName:   "logs/app.log",
        TimeFormat: "2006-01-02 15:04:05.000",
        TextOutput: false, // JSON format (default)
    })

logger.Info().Str("user", "john").Msg("User logged in")
```

**Output:**
```json
{"time":"2025-09-18 15:04:05.123","level":"info","user":"john","message":"User logged in"}
```

### Text Output Format

```go
logger := arbor.Logger().
    WithFileWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeFile,
        FileName:   "logs/app.log",
        TimeFormat: "15:04:05.000",
        TextOutput: true, // Human-readable text format
    })

logger.Info().Str("user", "john").Msg("User logged in")
```

**Output:**
```
15:04:05.123 INF > User logged in user=john
```

### File Writer Options

- **`FileName`**: Log file path (default: "logs/main.log")
- **`MaxSize`**: Maximum file size in bytes before rotation (default: 500KB)
- **`MaxBackups`**: Number of backup files to keep (default: 20)
- **`TextOutput`**: Enable human-readable format instead of JSON (default: false)
- **`TimeFormat`**: Timestamp format for log entries
- **`Level`**: Minimum log level to write

### AI-Friendly Log File Sizing

The default configuration is optimized for AI agent consumption:

- **500KB per file**: Approximately 3,300 log lines at ~150 bytes per line
- **20 backup files**: Maintains ~10MB total log history across all files
- **Automatic rotation**: New files created when size limit is reached
- **Timestamped backups**: Each rotated file includes timestamp in filename

This configuration ensures log files remain within AI context windows while maintaining sufficient history for debugging and analysis. For high-volume production systems, you may want to increase `MaxSize` and adjust `MaxBackups` accordingly:

```go
// Example: Larger files for high-volume systems
WithFileWriter(models.WriterConfiguration{
    Type:       models.LogWriterTypeFile,
    FileName:   "logs/app.log",
    MaxSize:    5 * 1024 * 1024, // 5MB per file
    MaxBackups: 10,               // Keep 10 backups (~50MB total)
})
```

## Log Levels

### String-Based Configuration

```go
// Configure from external config
logger := arbor.Logger().WithLevelFromString("debug")

// Supported levels: "trace", "debug", "info", "warn", "error", "fatal", "panic", "disabled"
```

### Programmatic Configuration

```go
logger := arbor.Logger().WithLevel(arbor.DebugLevel)

// Available levels: TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, PanicLevel
```

## Correlation ID Tracking

Correlation IDs enable request tracing across your application layers:

```go
// Set correlation ID for request tracking
logger := arbor.Logger().
    WithConsoleWriter(config).
    WithCorrelationId("req-12345")

logger.Info().Msg("Processing request")
logger.Debug().Str("step", "validation").Msg("Validating input")
logger.Info().Str("result", "success").Msg("Request completed")

// Auto-generate correlation ID
logger.WithCorrelationId("") // Generates UUID automatically

// Clear correlation ID
logger.ClearCorrelationId()
```

## Context-Specific Logging

For long-running processes, jobs, or any scenario where you need to stream all logs for a specific context (e.g., a `jobId`) to a durable store, `arbor` provides a context logging feature. This allows a consumer (e.g., a database writer) to receive all logs for multiple contexts on a single channel, in batches.

This approach is ideal for:
- Auditing all actions related to a specific job or entity.
- Persisting logs for long-running background tasks.
- Building custom log processing and analysis pipelines.

### How It Works

1.  **Consumer Sets a Channel**: At startup, your application's consumer creates a channel that accepts log batches (`chan []models.LogEvent`) and registers it with `arbor`.
2.  **Producers Log with Context**: Any part of your application, in any goroutine, can get a context-specific logger by calling `logger.WithContextWriter("your-job-id")`.
3.  **Additive Logging**: The context logger writes to all standard writers (like console and file) **and** sends a copy of the log to an internal buffer.
4.  **Batching and Streaming**: A background process batches the logs from the internal buffer and sends them as a slice to your consumer's channel. This is efficient and reduces channel contention.

### Setting up the Consumer

You can set up the context log consumer with default or custom buffering settings.

**Using Default Buffering**

This is the simplest way to get started. It uses a default batch size of 5 and a flush interval of 1 second.

```go
// 1. Create a channel to receive log batches.
logBatchChannel := make(chan []models.LogEvent, 10)

// 2. Configure arbor to send context logs to your channel with default settings.
arbor.Logger().SetContextChannel(logBatchChannel)
defer common.Stop() // Ensures the context buffer is flushed and stopped.

// 3. Start a consumer goroutine to process logs from the channel.
go func() {
    for logBatch := range logBatchChannel {
        // Process the batch of logs...
    }
}()
```

**Using Custom Buffering**

For more control over performance, you can specify the batch size and flush interval. This is useful for high-throughput applications where larger batches are more efficient.

```go
// 1. Create a channel.
logBatchChannel := make(chan []models.LogEvent, 10)

// 2. Configure with a larger batch size and longer interval.
arbor.Logger().SetContextChannelWithBuffer(logBatchChannel, 100, 5*time.Second)
defer common.Stop()

// 3. Start the consumer goroutine...
```

### Producer Example

This example demonstrates how a consumer can set up a channel and how multiple producers can log to it using a shared context ID.

```go
package main

import (
    "fmt"
    "sync"
    "time"

    "github.com/ternarybob/arbor"
    "github.com/ternarybob/arbor/common"
    "github.com/ternarybob/arbor/models"
)

func main() {
    // --- Consumer Setup ---

    // 1. Create a channel to receive log batches.
    logBatchChannel := make(chan []models.LogEvent, 10)

    // 2. Configure arbor to send context logs to your channel.
    // We use a small batch size and interval for demonstration purposes.
    arbor.Logger().SetContextChannelWithBuffer(logBatchChannel, 3, 500*time.Millisecond)
    defer common.Stop() // Ensures the context buffer is flushed and stopped.

    // 3. Start a consumer goroutine to process logs from the channel.
    var wgConsumer sync.WaitGroup
    wgConsumer.Add(1)
    go func() {
        defer wgConsumer.Done()
        for logBatch := range logBatchChannel {
            fmt.Printf("\n--- Received Batch of %d Logs ---\\n", len(logBatch))
            for _, log := range logBatch {
                // In a real application, you would write this to a database.
                fmt.Printf("  [DB] JobID: %s, Message: %s\n", log.CorrelationID, log.Message)
            }
            fmt.Println("------------------------------------")
        }
    }()

    // --- Producer Logic ---

    // 4. In various parts of your application, get a logger for a specific context.
    jobID := "job-xyz-789"
    logger := arbor.Logger().WithConsoleWriter(models.WriterConfiguration{})

    // Goroutine 1 simulates one part of the job.
    var wgProducers sync.WaitGroup
    wgProducers.Add(1)
    go func() {
        defer wgProducers.Done()
        jobLogger := logger.WithContextWriter(jobID)
        jobLogger.Info().Msg("Step 1: Validating input.")
        time.Sleep(10 * time.Millisecond)
        jobLogger.Info().Msg("Step 2: Processing data.")
    }()

    // Goroutine 2 simulates another part of the same job.
    wgProducers.Add(1)
    go func() {
        defer wgProducers.Done()
        jobLogger := logger.WithContextWriter(jobID)
        time.Sleep(20 * time.Millisecond)
        jobLogger.Warn().Msg("Step 3: A non-critical error occurred.")
        time.Sleep(10 * time.Millisecond)
        jobLogger.Info().Msg("Step 4: Job complete.")
    }()

    wgProducers.Wait()

    // 5. Stop the context buffer and wait for the consumer to finish.
    common.Stop()      // This will flush any remaining logs.
    close(logBatchChannel) // Close the channel to signal the consumer to exit.
    wgConsumer.Wait()
}
```

## Memory Logging & Retrieval

**Note:** For capturing logs related to a specific function or request, the recommended approach is to use `WithContextWriter` as described in the section above.

Arbor provides a powerful in-memory log store with optional BoltDB persistence for general-purpose debugging and log retrieval.

### Architecture

The memory writer uses a **shared log store** architecture:
- **Fast in-memory storage** (primary) for quick queries
- **Optional BoltDB persistence** (configurable)
- **Non-blocking async writes** - logging path remains fast
- **Automatic TTL cleanup** (10 min default, 1 min interval)

### Basic Configuration

```go
// In-memory only (fast, no persistence)
logger := arbor.Logger().
    WithMemoryWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeMemory,
        TimeFormat: "15:04:05.000",
    }).
    WithCorrelationId("debug-session")

// With optional BoltDB persistence
logger := arbor.Logger().
    WithMemoryWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeMemory,
        TimeFormat: "15:04:05.000",
        DBPath:     "temp/logs", // Enable persistence
    }).
    WithCorrelationId("debug-session")

// Log some messages
logger.Info().Msg("Starting process")
logger.Debug().Str("step", "initialization").Msg("Initializing components")
logger.Error().Str("error", "timeout").Msg("Operation failed")

// Retrieve logs by correlation ID (ordered by timestamp)
logs, err := logger.GetMemoryLogs("debug-session", arbor.DebugLevel)
if err != nil {
    log.Fatal(err)
}

// Display retrieved logs
for index, message := range logs {
    fmt.Printf("[%s]: %s\n", index, message)
}
```

### Memory Log Retrieval Options

```go
// Get all logs for correlation ID (timestamp ordered)
logs, _ := logger.GetMemoryLogsForCorrelation("correlation-id")

// Get logs with minimum level filter
logs, _ := logger.GetMemoryLogs("correlation-id", arbor.WarnLevel)

// Get most recent N entries
logs, _ := logger.GetMemoryLogsWithLimit(100)
```

### API Call Pattern - Snapshot at Request End

Perfect for API debugging where you want all logs for a request:

```go
func HandleRequest(c *gin.Context) {
    correlationID := c.GetHeader("X-Correlation-ID")
    logger := arbor.Logger().WithCorrelationId(correlationID)

    // Process request with logging
    logger.Info().Msg("Processing request")
    logger.Debug().Str("user", user.ID).Msg("Fetching user data")

    // ... business logic ...

    // At end of request - get snapshot ordered by timestamp
    if c.Query("debug") == "true" {
        logs, _ := logger.GetMemoryLogs(correlationID, arbor.TraceLevel)
        c.JSON(200, gin.H{
            "result": result,
            "logs":   logs, // All request logs in timestamp order
        })
    }
}
```

### WebSocket Streaming Pattern

Stream logs in real-time to WebSocket clients:

```go
import (
    "github.com/ternarybob/arbor/writers"
)

// Get memory writer and extract the shared store
memWriter := arbor.GetRegisteredMemoryWriter(arbor.WRITER_MEMORY)
store := memWriter.GetStore()

// Create WebSocket writer with 500ms poll interval
wsWriter := writers.WebSocketWriter(store, config, 500*time.Millisecond).(*writers.websocketWriter)

// Add WebSocket client
wsWriter.AddClient("client-id", yourWebSocketClient)

// Clients automatically receive new logs every 500ms
// Or query manually by timestamp:
newLogs, _ := wsWriter.GetLogsSince(time.Now().Add(-5 * time.Minute))
```

### Performance Characteristics

- **Console/File writes**: ~50-100μs (unchanged, fast path)
- **Memory store writes**: Non-blocking buffered (~100μs) + async persistence
- **Correlation queries**: ~50μs (in-memory map lookup)
- **Timestamp queries**: ~100μs (in-memory slice scan)
- **BoltDB persistence**: Async background writes (doesn't block logging)

## API Integration

### Gin Framework Integration

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/ternarybob/arbor"
    "github.com/ternarybob/arbor/models"
)

func main() {
    // Configure logger with memory writer for log retrieval
    logger := arbor.Logger().
        WithConsoleWriter(models.WriterConfiguration{
            Type:       models.LogWriterTypeConsole,
            TimeFormat: "15:04:05.000",
        }).
        WithMemoryWriter(models.WriterConfiguration{
            Type:       models.LogWriterTypeMemory,
            TimeFormat: "15:04:05.000",
        })

    // Create Gin engine with arbor integration
    r := gin.New()
    
    // Use arbor writer for Gin logs
    ginWriter := logger.GinWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeConsole,
        TimeFormat: "15:04:05.000",
    })
    
    r.Use(gin.LoggerWithWriter(ginWriter.(io.Writer)))
    
    // Your routes here
    r.GET("/health", func(c *gin.Context) {
        correlationID := c.GetHeader("X-Correlation-ID")
        requestLogger := logger.WithCorrelationId(correlationID)
        
        requestLogger.Info().Str("endpoint", "/health").Msg("Health check requested")
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    r.Run(":8080")
}
```

### Request Correlation Middleware

```go
func CorrelationMiddleware(logger arbor.ILogger) gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // Extract or generate correlation ID
        correlationID := c.GetHeader("X-Correlation-ID")
        if correlationID == "" {
            correlationID = generateUUID() // Your UUID generation
        }
        
        // Create request-scoped logger
        requestLogger := logger.WithCorrelationId(correlationID)
        
        // Store in context for handler access
        c.Set("logger", requestLogger)
        c.Header("X-Correlation-ID", correlationID)
        
        requestLogger.Info().
            Str("method", c.Request.Method).
            Str("path", c.Request.URL.Path).
            Msg("Request started")
        
        c.Next()
        
        requestLogger.Info().
            Int("status", c.Writer.Status()).
            Msg("Request completed")
    })
}
```

## Advanced Features

### Context Management

```go
// Add structured context
logger := arbor.Logger().
    WithContext("service", "user-management").
    WithContext("version", "1.2.0").
    WithPrefix("UserSvc")

// Copy logger with fresh context
cleanLogger := logger.Copy() // Same writers, no context data
```

## Configuration Examples

### From Environment Variables

```go
logLevel := os.Getenv("LOG_LEVEL")
if logLevel == "" {
    logLevel = "info"
}

logger := arbor.Logger().
    WithConsoleWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeConsole,
        TimeFormat: "15:04:05.000",
    }).
    WithLevelFromString(logLevel)
```

### Configuration Struct

```go
type LogConfig struct {
    Level      string `json:"level"`
    Console    bool   `json:"console"`
    File       string `json:"file"`
    Memory     bool   `json:"memory"`
    TimeFormat string `json:"time_format"`
    TextOutput bool   `json:"text_output"`
}

func ConfigureLogger(config LogConfig) arbor.ILogger {
    logger := arbor.NewLogger()

    if config.Console {
        logger.WithConsoleWriter(models.WriterConfiguration{
            Type:       models.LogWriterTypeConsole,
            TimeFormat: config.TimeFormat,
        })
    }

    if config.File != "" {
        logger.WithFileWriter(models.WriterConfiguration{
            Type:       models.LogWriterTypeFile,
            FileName:   config.File,
            TimeFormat: config.TimeFormat,
            TextOutput: config.TextOutput, // Enable text format for files
        })
    }

    if config.Memory {
        logger.WithMemoryWriter(models.WriterConfiguration{
            Type:       models.LogWriterTypeMemory,
            TimeFormat: config.TimeFormat,
        })
    }

    return logger.WithLevelFromString(config.Level)
}
```

## Architecture & Performance

### Log Store Architecture

Arbor uses a **shared log store** pattern for memory-based writers:

```
┌─────────────┐
│  Log Event  │
└──────┬──────┘
       │
       ├──────────► Console Writer (direct, ~50μs)
       ├──────────► File Writer (direct, ~80μs)
       └──────────► Log Store Writer (buffered async)
                         │
                         ├──► In-Memory Store (primary, fast queries)
                         └──► BoltDB (optional persistence, async)
                              │
                              ├──► Memory Writer (correlation queries)
                              ├──► WebSocket Writer (timestamp polling)
                              └──► Future readers...
```

### Performance Characteristics

- **Direct Writers** (Console/File): ~50-100μs per log, no blocking
- **Log Store Writes**: Buffered channel (1000 entries), non-blocking
- **In-Memory Queries**: ~50-100μs for correlation/timestamp lookups
- **Optional Persistence**: Async BoltDB writes, doesn't block logging path
- **Cleanup**: Automatic TTL expiration every 1 minute (10 min default TTL)
- **Thread Safety**: RWMutex for concurrent access with minimal lock contention
- **Level Filtering**: Occurs at writer level for efficiency

### Design Principles

- **Separation of Concerns**: Write path (fast) vs. Query path (acceptable latency)
- **Non-Blocking**: Buffered async writes prevent slow storage from blocking logs
- **In-Memory Primary**: Fast queries without disk I/O for active sessions
- **Optional Persistence**: BoltDB backup for crash recovery and long-term storage
- **Extensible**: Easy to add new store-based readers (metrics, search, alerts)

## CI/CD

This project uses GitHub Actions for continuous integration and deployment:

- **Automated Testing**: Runs tests on every push and pull request
- **Code Quality Checks**: Enforces `go fmt`, `go vet`, and build validation
- **Auto-Release**: Automatically creates releases on main branch pushes
- **Tagged Releases**: Manual version control via git tags

## Documentation

Full documentation is available at [pkg.go.dev](https://pkg.go.dev/github.com/ternarybob/arbor).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Libraries

This library is part of the T3B ecosystem:

- [funktion](https://github.com/ternarybob/funktion) - Core utility functions
- [satus](https://github.com/ternarybob/satus) - Configuration and status management  
- [arbor](https://github.com/ternarybob/arbor) - Structured logging system
- [omnis](https://github.com/ternarybob/omnis) - Web framework integrations