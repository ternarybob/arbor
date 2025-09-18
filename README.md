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

- **Multi-Writer Architecture**: Console, File, and Memory (BoltDB) writers
- **Correlation ID Tracking**: Request tracing across application layers
- **Structured Logging**: Rich field support with fluent API
- **Log Level Management**: String-based and programmatic level configuration
- **Memory Logging**: Persistent storage with TTL and automatic cleanup
- **API Integration**: Built-in Gin framework support
- **Global Registry**: Cross-context logger access
- **Thread-Safe**: Concurrent access with proper synchronization
- **Performance Focused**: Optimized for high-throughput API scenarios

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
        MaxSize:    10 * 1024 * 1024, // 10MB
        MaxBackups: 5,
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
- **`MaxSize`**: Maximum file size in bytes before rotation (default: 10MB)
- **`MaxBackups`**: Number of backup files to keep (default: 5)
- **`TextOutput`**: Enable human-readable format instead of JSON (default: false)
- **`TimeFormat`**: Timestamp format for log entries
- **`Level`**: Minimum log level to write

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

## Memory Logging & Retrieval

Store logs in BoltDB for later retrieval and debugging:

```go
// Configure memory writer
logger := arbor.Logger().
    WithMemoryWriter(models.WriterConfiguration{
        Type:       models.LogWriterTypeMemory,
        TimeFormat: "15:04:05.000",
    }).
    WithCorrelationId("debug-session")

// Log some messages
logger.Info().Msg("Starting process")
logger.Debug().Str("step", "initialization").Msg("Initializing components")
logger.Error().Str("error", "timeout").Msg("Operation failed")

// Retrieve logs by correlation ID
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
// Get all logs for correlation ID
logs, _ := logger.GetMemoryLogsForCorrelation("correlation-id")

// Get logs with minimum level filter
logs, _ := logger.GetMemoryLogs("correlation-id", arbor.WarnLevel)

// Get most recent N entries
logs, _ := logger.GetMemoryLogsWithLimit(100)
```

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

## Global Registry Pattern

Access loggers across different application contexts:

```go
// Register logger during application startup
func initLogging() {
    mainLogger := arbor.Logger().
        WithConsoleWriter(consoleConfig).
        WithMemoryWriter(memoryConfig).
        WithCorrelationId("main-app")
    
    // Register for global access
    arbor.RegisterLogger("main", mainLogger)
    arbor.RegisterLogger("api", apiLogger)
}

// Access from any context (e.g., service layer)
func ProcessRequest(correlationID string) error {
    logger := arbor.GetRegisteredLogger("main")
    if logger == nil {
        logger = arbor.GetRegisteredLogger("default") // fallback
    }
    
    requestLogger := logger.WithCorrelationId(correlationID)
    requestLogger.Info().Msg("Processing business logic")
    
    return nil
}

// Retrieve logs in API handlers
func GetLogsHandler(c *gin.Context) {
    correlationID := c.Query("correlation_id")
    
    logger := arbor.GetRegisteredLogger("main")
    logs, err := logger.GetMemoryLogs(correlationID, arbor.DebugLevel)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"logs": logs})
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

### Broadcast Logging

```go
// Register multiple loggers
arbor.RegisterLogger("console", consoleLogger)
arbor.RegisterLogger("file", fileLogger) 
arbor.RegisterLogger("memory", memoryLogger)

// Broadcast to all registered loggers
arbor.BroadcastInfo().Msg("Application started")
arbor.BroadcastError().Err(err).Msg("Critical system error")
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

## Performance Considerations

- **Memory Writer**: Uses BoltDB for persistence with configurable TTL
- **Cleanup**: Automatic cleanup of expired entries every minute
- **Thread Safety**: All operations are thread-safe with minimal lock contention
- **Buffer Limits**: Memory writer maintains buffer limits per correlation ID
- **Log Levels**: Level filtering occurs at writer level for efficiency

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
