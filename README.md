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

- **Multi-Writer Architecture**:
  - Synchronous writers (Console, File) for immediate output
  - Async writers (LogStore, Context) with buffered non-blocking processing
  - Shared log store for queryable in-memory and optional persistent storage
- **Correlation ID Tracking**: Request tracing across application layers
- **Structured Logging**: Rich field support with fluent API
- **Log Level Management**: String-based and programmatic level configuration
- **In-Memory Log Store**: Fast queryable storage with optional BoltDB persistence
- **API Integration**: Built-in Gin framework support
- **Global Registry**: Cross-context logger access
- **Thread-Safe**: Concurrent access with proper synchronization
- **Performance Focused**: Non-blocking async writes, optimized for high-throughput API scenarios
- **Async Processing**: Non-blocking buffered writes with graceful shutdown and automatic buffer draining

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

## Async Writers with ChannelWriter

Arbor provides a powerful async buffered writer pattern through the `channelWriter` base. This architecture enables non-blocking log writes while maintaining reliability through automatic buffer draining and graceful shutdown.

### What is ChannelWriter?

The `channelWriter` is a reusable async writer implementation that:
- **Buffers log entries** in a channel (default 1000 entries)
- **Processes entries** in a background goroutine
- **Returns immediately** from Write() calls (~100μs latency)
- **Drains buffer** automatically during shutdown to prevent log loss
- **Handles overflow** gracefully by dropping entries with warnings

### Built-in Async Writers

Two writers use the channelWriter base:

#### LogStoreWriter
Writes logs to in-memory or persistent storage asynchronously:

```go
// Used internally by WithMemoryWriter
logger := arbor.Logger().WithMemoryWriter(models.WriterConfiguration{
    Type: models.LogWriterTypeMemory,
})

// Logs are buffered and written asynchronously
// No blocking on database writes
logger.Info().Msg("Stored asynchronously in memory/BoltDB")
```

#### ContextWriter
Streams logs for specific contexts (jobs, requests) to channels:

```go
// Setup channel to receive log batches
logChannel := make(chan []models.LogEvent, 10)
arbor.Logger().SetContextChannel(logChannel)

// Create context logger
contextLogger := logger.WithContextWriter("job-123")

// Logs go to all writers + async buffer for context channel
contextLogger.Info().Msg("Logged to console and buffered for channel")
```

### Lifecycle and Behavior

**Buffer Management:**
```go
// Buffer capacity: 1000 entries per writer
// Write latency: ~100μs (non-blocking)
// Overflow: Drops with warning log, no blocking
```

**Graceful Shutdown:**
```go
// On logger cleanup or application shutdown:
// 1. Stop accepting new entries
// 2. Process all buffered entries
// 3. Clean up resources

// For ContextWriter specifically:
defer common.Stop() // Flushes context buffer
```

**Performance Characteristics:**
- Write operations complete in ~100μs
- Background processing doesn't block logging
- Supports 10,000+ logs/second throughput
- Automatic level filtering before buffering
- Thread-safe concurrent writes

### Creating Custom Async Writers

You can build custom writers using channelWriter for:
- External services (Datadog, Splunk, CloudWatch)
- Custom databases (MongoDB, PostgreSQL, Elasticsearch)
- Specialized processing (aggregation, filtering, transformation)

See the "Custom Async Writers" section in Writer Architecture for detailed examples.

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
5.  **Non-Blocking Writes**: The context logger uses an async buffered writer (1000-entry capacity) to prevent blocking on slow context buffer operations, ensuring your application remains responsive even under high logging load.

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
- **Buffered async writes** - LogStoreWriter uses 1000-entry buffer for non-blocking writes with automatic overflow handling
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

### Channel-Based Log Streaming

Arbor provides two channel-based APIs for streaming logs: `SetContextChannel` for singleton context logging (covered in the "Context-Specific Logging" section above) and `SetChannel`/`SetChannelWithBuffer` for multiple independent named channels. This section focuses on the named channel API for general-purpose log streaming.

#### Overview

The named channel API allows you to register multiple independent channels, each with its own batching configuration and lifecycle:

- **SetContextChannel**: Singleton buffer shared by all `WithContextWriter` loggers. Use for capturing all logs for a specific job/request context. Managed via `common.Start()` and `common.Stop()`.
- **SetChannel/SetChannelWithBuffer**: Multiple independent named channels. Use for general streaming to WebSocket clients, external services, or custom consumers. Each channel has isolated batching and is managed via `SetChannel()` and `UnregisterChannel()`.

**Use Cases:**
- **SetContextChannel**: "Capture all logs for job-123 and store in database"
- **SetChannel**: "Stream all application logs to WebSocket clients" or "Send error logs to Slack alerting"

#### Basic Usage

```go
// Create channel for receiving log batches
// Buffer size of 10-100 recommended depending on throughput
logChannel := make(chan []models.LogEvent, 10)

// Register channel with default batching (5 events, 1 second)
// The "name" parameter uniquely identifies this channel
arbor.Logger().SetChannel("websocket-logs", logChannel)

// Or with custom batching for high-throughput scenarios
// batchSize: number of events before automatic flush
// flushInterval: maximum time before automatic flush
arbor.Logger().SetChannelWithBuffer("websocket-logs", logChannel, 100, 5*time.Second)

// Consumer goroutine to receive and process batches
go func() {
    for logBatch := range logChannel {
        // Batch contains up to batchSize events
        for _, logEvent := range logBatch {
            // Process each event
            fmt.Printf("[%s] %s: %s\n", logEvent.Level, logEvent.CorrelationID, logEvent.Message)
        }

        // Error handling for WebSocket broadcast
        if err := broadcastToClients(logBatch); err != nil {
            log.Printf("Broadcast error: %v", err)
        }
    }
}()

// Normal logging - all logs go to the channel
arbor.Info().Msg("This message will be batched and sent to the channel")
```

#### Advanced Patterns

**Multiple Independent Channels**

Register multiple named channels for different purposes:

```go
// Audit logs to database
auditChannel := make(chan []models.LogEvent, 20)
arbor.Logger().SetChannel("audit-logs", auditChannel)

// Metrics to monitoring system
metricsChannel := make(chan []models.LogEvent, 50)
arbor.Logger().SetChannel("metrics", metricsChannel)

// Critical alerts to notification service
alertsChannel := make(chan []models.LogEvent, 10)
arbor.Logger().SetChannelWithBuffer("alerts", alertsChannel, 1, 100*time.Millisecond)

// Each channel receives all logs independently
```

**Dynamic Channel Registration**

Add and remove channels at runtime:

```go
// Add channel when WebSocket client connects
func OnClientConnect(clientID string) {
    clientChannel := make(chan []models.LogEvent, 10)
    channelName := fmt.Sprintf("client-%s", clientID)

    arbor.Logger().SetChannel(channelName, clientChannel)

    // Start consumer for this client
    go streamToClient(clientID, clientChannel)
}

// Remove channel when client disconnects
func OnClientDisconnect(clientID string) {
    channelName := fmt.Sprintf("client-%s", clientID)
    arbor.Logger().UnregisterChannel(channelName)
    // Channel cleanup happens automatically
}
```

**High-Throughput Configuration**

For high-volume scenarios, use larger batches and longer intervals:

```go
// Process 500 events at once or flush every 10 seconds
// Reduces overhead but increases latency
highThroughputChannel := make(chan []models.LogEvent, 100)
arbor.Logger().SetChannelWithBuffer("high-volume", highThroughputChannel, 500, 10*time.Second)
```

**Real-Time Configuration**

For low-latency requirements, use small batches and short intervals:

```go
// Process 1-5 events at once with sub-second flush
// Lower latency but higher processing overhead
realTimeChannel := make(chan []models.LogEvent, 20)
arbor.Logger().SetChannelWithBuffer("real-time", realTimeChannel, 1, 100*time.Millisecond)
```

**WebSocket Broadcasting with Connection Management**

Complete example with error handling and graceful shutdown:

```go
type WebSocketManager struct {
    clients map[string]*websocket.Conn
    mu      sync.RWMutex
    logChan chan []models.LogEvent
}

func NewWebSocketManager() *WebSocketManager {
    mgr := &WebSocketManager{
        clients: make(map[string]*websocket.Conn),
        logChan: make(chan []models.LogEvent, 50),
    }

    // Register channel with appropriate batching
    arbor.Logger().SetChannelWithBuffer("websocket", mgr.logChan, 20, 1*time.Second)

    // Start broadcast goroutine
    go mgr.broadcastLoop()

    return mgr
}

func (mgr *WebSocketManager) broadcastLoop() {
    for logBatch := range mgr.logChan {
        mgr.mu.RLock()
        for clientID, conn := range mgr.clients {
            if err := conn.WriteJSON(logBatch); err != nil {
                log.Printf("Failed to send to client %s: %v", clientID, err)
                // Handle disconnected clients
                go mgr.RemoveClient(clientID)
            }
        }
        mgr.mu.RUnlock()
    }
}

func (mgr *WebSocketManager) AddClient(clientID string, conn *websocket.Conn) {
    mgr.mu.Lock()
    mgr.clients[clientID] = conn
    mgr.mu.Unlock()
}

func (mgr *WebSocketManager) RemoveClient(clientID string) {
    mgr.mu.Lock()
    if conn, ok := mgr.clients[clientID]; ok {
        conn.Close()
        delete(mgr.clients, clientID)
    }
    mgr.mu.Unlock()
}

func (mgr *WebSocketManager) Shutdown() {
    // Unregister channel (stops the buffer and writer)
    arbor.Logger().UnregisterChannel("websocket")

    // Close all client connections
    mgr.mu.Lock()
    for _, conn := range mgr.clients {
        conn.Close()
    }
    mgr.clients = nil
    mgr.mu.Unlock()

    // Drain remaining batches with bounded wait
    timeout := time.After(2 * time.Second)
drainLoop:
    for {
        select {
        case batch := <-mgr.logChan:
            // Process final batches (already closed clients, just drain)
            _ = batch
        case <-time.After(100 * time.Millisecond):
            // No batch arrived, done draining
            break drainLoop
        case <-timeout:
            // Overall timeout exceeded
            break drainLoop
        }
    }

    // Exit without closing the channel - the sender (ChannelBuffer) owns it
    // and may still attempt a final send during flush
}
```

#### Lifecycle Management

**Cleanup with UnregisterChannel**

Properly stop and remove a channel logger:

```go
// Register channel
logChannel := make(chan []models.LogEvent, 10)
arbor.Logger().SetChannel("my-channel", logChannel)

// Later, when done with the channel
arbor.Logger().UnregisterChannel("my-channel")
// This stops the ChannelWriter and ChannelBuffer goroutines
// and removes the channel from the registry
```

**Automatic Cleanup on Replacement**

Calling `SetChannel` with an existing name automatically cleans up the old writer and buffer:

```go
// First registration
channel1 := make(chan []models.LogEvent, 10)
arbor.Logger().SetChannel("stream", channel1)

// Later, replace with new channel
// Old channel is automatically cleaned up
channel2 := make(chan []models.LogEvent, 10)
arbor.Logger().SetChannel("stream", channel2)
// channel1 is no longer receiving logs
```

**Graceful Shutdown**

Pattern for draining remaining batches during shutdown:

```go
func shutdown() {
    // Step 1: Unregister channel (stops the buffer and writer)
    arbor.Logger().UnregisterChannel("my-channel")

    // Step 2: Drain any remaining batches with bounded wait
    // Use a timeout to prevent indefinite blocking
    timeout := time.After(2 * time.Second)
    drainLoop:
    for {
        select {
        case batch := <-logChannel:
            // Process final batches
            processBatch(batch)
        case <-time.After(100 * time.Millisecond):
            // No batch arrived within timeout, done draining
            break drainLoop
        case <-timeout:
            // Overall timeout exceeded
            log.Println("Shutdown timeout exceeded, exiting")
            break drainLoop
        }
    }

    // Step 3: Exit consumer goroutine without closing the channel
    // The sender (ChannelBuffer) owns the channel and may attempt
    // a final send during flush. Never close a channel you don't own.
}
```

**Important**: The consumer must not close the channel because the sender (`common.ChannelBuffer`) owns it and may still attempt a final send during buffer flush. Receivers should only read from channels; closing is the sender's responsibility.

**Resource Management**

Each named channel creates two goroutines (ChannelWriter + ChannelBuffer), so cleanup is important:

```go
// Resource usage per channel:
// - 1 ChannelWriter goroutine (processes writes)
// - 1 ChannelBuffer goroutine (batches events)
// - ~150KB memory overhead (buffers)

// Always cleanup when done:
defer arbor.Logger().UnregisterChannel("channel-name")
```

#### Comparison: SetChannel vs SetContextChannel

| Feature | SetContextChannel | SetChannel/SetChannelWithBuffer |
|---------|-------------------|----------------------------------|
| **Buffer Type** | Singleton (one shared buffer) | Multiple independent buffers |
| **Scope** | All `WithContextWriter` loggers | All loggers |
| **Use Case** | Job/request tracking | General streaming/broadcasting |
| **Example** | "All logs for job-123" | "Stream to WebSocket clients" |
| **Lifecycle** | `common.Start()` / `common.Stop()` | `SetChannel()` / `UnregisterChannel()` |
| **Isolation** | Shared by correlation ID | Isolated by channel name |
| **Best For** | Context-specific capture | Service integrations, real-time streaming |

#### Error Handling and Edge Cases

**Nil Channel**

Calling `SetChannel` with a nil channel will panic with a clear error message:

```go
// This will panic
arbor.Logger().SetChannel("bad-channel", nil)
// Panic: Cannot create channel writer with nil channel
```

**Invalid Parameters**

Zero or negative `batchSize` or `flushInterval` will use safe defaults:

```go
// These all use defaults: batchSize=5, flushInterval=1s
arbor.Logger().SetChannelWithBuffer("ch1", logChan, 0, 1*time.Second)
arbor.Logger().SetChannelWithBuffer("ch2", logChan, -10, 1*time.Second)
arbor.Logger().SetChannelWithBuffer("ch3", logChan, 5, 0)
arbor.Logger().SetChannelWithBuffer("ch4", logChan, 5, -1*time.Second)
```

**Buffer Overflow**

If the ChannelWriter buffer fills (default 1000 entries), entries are dropped with warning logs:

```go
// If logging faster than channel consumer can process:
// - ChannelWriter buffer fills (1000 entries)
// - New writes complete in ~100μs (non-blocking)
// - Entries are dropped with warning log
// - Application continues normally without blocking
```

**Channel Blocking**

If the output channel blocks (consumer too slow), the ChannelBuffer will timeout and drop the batch:

```go
// If consumer is too slow and channel buffer fills:
// - ChannelBuffer attempts to send batch
// - 1 second timeout on channel send
// - Batch is dropped if timeout occurs
// - Warning logged about dropped batch
// - Next batch continues normally
```

**Best Practices:**
- Use buffered channels (10-100 capacity) for output
- Monitor consumer performance to avoid backpressure
- Implement proper error handling in consumers
- Always cleanup channels with `UnregisterChannel` when done

For more details on the context-specific logging API, see the "Context-Specific Logging" section above.

### Performance Characteristics

- **Synchronous writes (Console/File)**: ~50-100μs, blocking but fast
- **Async writes (LogStore/Context)**: ~100μs non-blocking + background processing
  - Buffer capacity: 1000 entries per writer
  - Overflow behavior: Drops entries with warning log
  - Shutdown: Automatic buffer draining prevents log loss
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

## Writer Architecture

Arbor uses different writer patterns optimized for specific use cases. Understanding these patterns helps you choose the right configuration for your application.

### Synchronous Writers (Console, File)

**Pattern:** Direct write to output (stdout or file)

These writers provide immediate output with minimal overhead:

- **Performance**: ~50-100μs per log entry
- **Blocking**: Yes, but very fast (acceptable for most use cases)
- **Use Cases**:
  - Development debugging with immediate console feedback
  - Production file logging for audit trails
  - Scenarios where log order guarantee is critical

**Example:**

```go
logger := arbor.Logger().
    WithConsoleWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeConsole,
        TimeFormat: "15:04:05.000",
    }).
    WithFileWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeFile,
        FileName:   "logs/app.log",
        MaxSize:    500 * 1024,
        MaxBackups: 20,
    })

logger.Info().Msg("Immediate output to console and file")
```

### Async Writers (LogStore, Context)

**Pattern:** Buffered channel + background goroutine processing

These writers provide non-blocking writes with automatic buffer management:

- **Performance**: ~100μs non-blocking write + async background processing
- **Blocking**: No - Write() returns immediately
- **Buffer Capacity**: 1000 entries per writer
- **Overflow Behavior**: Drops entries with warning log when buffer is full
- **Shutdown**: Automatic buffer draining ensures no log loss during graceful shutdown

**Data Flow:**

```
Log Event → channelWriter Base (async, buffered)
    ├──► LogStoreWriter → ILogStore → In-Memory/BoltDB
    └──► ContextWriter → Global Context Buffer → Channel
```

**Benefits:**

- Non-blocking writes prevent slow storage from blocking logging path
- 1000-entry buffer absorbs traffic bursts without dropping logs
- Graceful shutdown with automatic buffer draining prevents log loss
- Level filtering applied before buffering for efficiency
- Thread-safe concurrent writes with minimal lock contention

**Example:**

```go
// LogStore writer for queryable memory logs
logger := arbor.Logger().
    WithMemoryWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeMemory,
        TimeFormat: "15:04:05.000",
    })

// Context writer for streaming logs to channel
logChannel := make(chan []models.LogEvent, 10)
arbor.Logger().SetContextChannel(logChannel)
contextLogger := logger.WithContextWriter("job-123")

contextLogger.Info().Msg("Non-blocking write completes in ~100μs")
```

### Custom Async Writers (ChannelWriter)

For advanced use cases, you can create custom async writers using the `channelWriter` base. This is useful when you need to integrate with custom storage backends, external services, or implement specialized log processing.

**Pattern:** The `channelWriter` provides a reusable async buffered writer with automatic lifecycle management.

**Use Cases:**
- Custom database writers (MongoDB, PostgreSQL, Elasticsearch)
- External service integrations (Datadog, Splunk, CloudWatch)
- Specialized log processing (aggregation, filtering, transformation)
- High-throughput scenarios requiring non-blocking writes

#### Basic Custom Writer Example

```go
import (
    "github.com/ternarybob/arbor/writers"
    "github.com/ternarybob/arbor/models"
)

// Custom writer that sends logs to an external API
type APIWriter struct {
    writer writers.IChannelWriter
    config models.WriterConfiguration
    apiClient *YourAPIClient
}

func NewAPIWriter(config models.WriterConfiguration, apiURL string) (writers.IWriter, error) {
    apiClient := NewAPIClient(apiURL)

    // Define processor that handles each log entry
    processor := func(entry models.LogEvent) error {
        return apiClient.SendLog(entry)
    }

    // Create channel writer with 1000 buffer size
    writer, err := writers.NewChannelWriter(config, 1000, processor)
    if err != nil {
        return nil, err
    }

    // Start the background goroutine
    if err := writer.Start(); err != nil {
        return nil, err
    }

    return &APIWriter{
        writer: writer,
        config: config,
        apiClient: apiClient,
    }, nil
}

// Implement IWriter interface
func (w *APIWriter) Write(data []byte) (int, error) {
    return w.writer.Write(data)
}

func (w *APIWriter) WithLevel(level log.Level) writers.IWriter {
    w.writer.WithLevel(level)
    return w
}

func (w *APIWriter) GetFilePath() string {
    return "" // Not file-based
}

func (w *APIWriter) Close() error {
    // Gracefully shut down - drains buffer before closing
    return w.writer.Close()
}
```

#### WebSocketWriter using ChannelWriter

This example demonstrates creating a WebSocketWriter that broadcasts log events to connected WebSocket clients using the ChannelWriter pattern:

```go
import (
    "encoding/json"
    "sync"

    "github.com/gorilla/websocket"
    "github.com/ternarybob/arbor"
    "github.com/ternarybob/arbor/models"
    "github.com/ternarybob/arbor/writers"
)

// WebSocketWriter broadcasts log events to WebSocket clients
type WebSocketWriter struct {
    writer  writers.IChannelWriter
    config  models.WriterConfiguration
    manager *ConnectionManager
}

// ConnectionManager handles thread-safe WebSocket client connections
type ConnectionManager struct {
    clients map[string]*websocket.Conn
    mu      sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
    return &ConnectionManager{
        clients: make(map[string]*websocket.Conn),
    }
}

func (cm *ConnectionManager) AddClient(clientID string, conn *websocket.Conn) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    cm.clients[clientID] = conn
}

func (cm *ConnectionManager) RemoveClient(clientID string) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    if conn, ok := cm.clients[clientID]; ok {
        conn.Close()
        delete(cm.clients, clientID)
    }
}

func (cm *ConnectionManager) Broadcast(logEvent models.LogEvent) error {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    // Broadcast to all connected clients
    for clientID, conn := range cm.clients {
        if err := conn.WriteJSON(logEvent); err != nil {
            // Handle disconnected clients asynchronously
            go cm.RemoveClient(clientID)
            continue
        }
    }
    return nil
}

// NewWebSocketWriter creates a ChannelWriter-based WebSocket broadcaster
func NewWebSocketWriter(config models.WriterConfiguration, manager *ConnectionManager) (writers.IWriter, error) {
    // Define processor that broadcasts each log entry to WebSocket clients
    processor := func(entry models.LogEvent) error {
        return manager.Broadcast(entry)
    }

    // Create channel writer with 1000 buffer size
    writer, err := writers.NewChannelWriter(config, 1000, processor)
    if err != nil {
        return nil, err
    }

    // Start the background goroutine
    if err := writer.Start(); err != nil {
        return nil, err
    }

    return &WebSocketWriter{
        writer:  writer,
        config:  config,
        manager: manager,
    }, nil
}

// Implement IWriter interface
func (w *WebSocketWriter) Write(data []byte) (int, error) {
    return w.writer.Write(data)
}

func (w *WebSocketWriter) WithLevel(level log.Level) writers.IWriter {
    w.writer.WithLevel(level)
    return w
}

func (w *WebSocketWriter) GetFilePath() string {
    return "" // Not file-based
}

func (w *WebSocketWriter) Close() error {
    // Gracefully shut down - drains buffer before closing
    return w.writer.Close()
}

// Usage example
func SetupWebSocketLogging() {
    // Create connection manager
    connManager := NewConnectionManager()

    // Create WebSocket writer
    wsWriter, err := NewWebSocketWriter(models.WriterConfiguration{
        Type:  models.LogWriterTypeConsole, // Use console type for compatibility
        Level: arbor.InfoLevel,
    }, connManager)
    if err != nil {
        log.Fatal(err)
    }
    defer wsWriter.Close()

    // Register with logger
    allWriters := append(arbor.GetAllRegisteredWriters(), wsWriter)
    logger := arbor.Logger().WithWriters(allWriters)

    // Add WebSocket clients as they connect
    // connManager.AddClient(clientID, conn)

    // All logs now broadcast to WebSocket clients
    logger.Info().Msg("This log will be sent to all WebSocket clients")
}
```

**Key Features:**
- **Thread-safe connection management**: RWMutex protects concurrent client access
- **Automatic cleanup**: Disconnected clients are removed asynchronously during broadcast failures
- **Non-blocking writes**: ChannelWriter buffers logs, preventing slow clients from blocking the logger
- **Error handling**: Gracefully handles client disconnections without affecting other clients
- **Standard IWriter interface**: Integrates seamlessly with arbor's writer system

**Notes:**
- The processor function unmarshal and broadcasts the `LogEvent` to all connected clients
- Connection management is separated from the writer for better modularity
- Buffer size of 1000 entries prevents memory issues during client slowdowns
- Always call `Close()` during shutdown to drain the buffer and prevent log loss

#### Manual Lifecycle Control

If you need fine-grained control over the goroutine lifecycle:

```go
// Create without auto-start
processor := func(entry models.LogEvent) error {
    // Your processing logic
    return nil
}

writer, err := writers.NewChannelWriter(config, 1000, processor)
if err != nil {
    log.Fatal(err)
}

// Start when ready
if err := writer.Start(); err != nil {
    log.Fatal(err)
}

// Check if running
if writer.IsRunning() {
    fmt.Println("Writer is processing logs")
}

// Stop processing (drains buffer)
if err := writer.Stop(); err != nil {
    log.Printf("Error stopping writer: %v", err)
}

// Close and cleanup
writer.Close()
```

#### Helper Function for Simple Async Writers

For most cases, use the `newAsyncWriter` helper which creates and starts in one call:

```go
// This is used internally by LogStoreWriter and ContextWriter
processor := func(entry models.LogEvent) error {
    // Process log entry
    return db.Store(entry)
}

// Creates, starts, and returns ready-to-use writer
writer, err := writers.newAsyncWriter(config, 1000, processor)
if err != nil {
    log.Fatal(err)
}
// Writer is already running and processing logs
```

#### Error Handling in Processors

The processor function receives each log entry and should handle errors appropriately:

```go
processor := func(entry models.LogEvent) error {
    // Retry logic for transient failures
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        err := sendToExternalService(entry)
        if err == nil {
            return nil
        }

        if isTransientError(err) && i < maxRetries-1 {
            time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
            continue
        }

        // Log error and return (channelWriter will log the failure)
        return fmt.Errorf("failed after %d retries: %w", i+1, err)
    }
    return nil
}
```

#### Buffer Overflow Behavior

When the buffer is full (1000 entries by default):

```go
// Buffer full scenario
writer, _ := writers.NewChannelWriter(config, 1000, slowProcessor)
writer.Start()

// If buffer fills up:
// - New writes complete immediately (~100μs)
// - Entry is dropped with warning log
// - No blocking occurs
// - Application continues normally
```

#### Graceful Shutdown Pattern

Always close writers to ensure no log loss:

```go
func main() {
    writer := setupCustomWriter()
    defer writer.Close() // Drains buffer before exiting

    // Application logic...

    // On shutdown, Close() will:
    // 1. Stop accepting new entries
    // 2. Process all buffered entries
    // 3. Clean up resources
}
```

#### Performance Characteristics

- **Write latency**: ~100μs (non-blocking, returns immediately)
- **Buffer capacity**: Configurable (default 1000 entries)
- **Throughput**: Supports 10,000+ logs/sec depending on processor speed
- **Memory overhead**: ~150KB per writer (buffer + goroutine)
- **Buffer drain**: Automatic on Close() - ensures zero log loss during shutdown

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