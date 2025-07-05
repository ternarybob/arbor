# Arbor Fluent Logging

The Arbor logger now supports a fluent interface similar to zerolog, allowing for readable and efficient logging with method chaining.

## Features

- **Fluent Interface**: Chain methods for building log entries
- **Multiple Log Levels**: Support for Trace, Debug, Info, Warn, Error, Fatal, and Panic levels
- **Field Addition**: Add structured fields with `Str()` and `Err()` methods
- **Formatted Messages**: Support for both `Msg()` and `Msgf()` message methods
- **Context Support**: Integration with correlation IDs and prefixes
- **Global Functions**: Direct access to logging without creating logger instances

## Usage Examples

### Basic Logging

```go
import "github.com/ternarybob/arbor"

// Simple message
arbor.Info().Msg("Application started")

// With error
arbor.Error().Err(err).Msg("Something went wrong")

// With structured fields
arbor.Debug().Str("component", "auth").Msg("Debug information")
```

### Using Logger Instance

```go
logger := arbor.GetLogger()
logger.Warn().Str("connection", "database").Err(err).Msg("Connection failed")
```

### Formatted Messages

```go
port := 8080
host := "localhost"
arbor.Info().Msgf("Server starting on %s:%d", host, port)
```

### Chaining Multiple Fields

```go
arbor.Warn().
    Str("user", "john_doe").
    Str("action", "login").
    Err(errors.New("invalid credentials")).
    Msg("Authentication failed")
```

### Context-Aware Logging

```go
logger := arbor.GetLogger()
contextLogger := logger.WithCorrelationId("req-123").WithPrefix("API")
contextLogger.Info().Str("endpoint", "/users").Msg("Request processed")
```

## API Reference

### Global Functions

- `arbor.Trace()` - Create a trace level log event
- `arbor.Debug()` - Create a debug level log event  
- `arbor.Info()` - Create an info level log event
- `arbor.Warn()` - Create a warn level log event
- `arbor.Error()` - Create an error level log event
- `arbor.Fatal()` - Create a fatal level log event
- `arbor.Panic()` - Create a panic level log event
- `arbor.GetLogger()` - Get the default logger instance

### Logger Instance Methods

All the same level methods are available on logger instances:

```go
logger := arbor.GetLogger()
logger.Info().Msg("message")
```

### Log Event Methods

Once you call a level method (e.g., `Info()`), you get a log event that supports:

- `Str(key, value string)` - Add a string field
- `Err(err error)` - Add an error field
- `Msg(message string)` - Log the message (terminal method)
- `Msgf(format string, args ...interface{})` - Log a formatted message (terminal method)

### Context Methods

Available on logger instances:

- `WithCorrelationId(id string)` - Add correlation ID to all log entries
- `WithPrefix(prefix string)` - Add prefix to all log entries
- `WithLevel(level LogLevel)` - Set minimum log level
- `WithContext(key, value string)` - Add custom context
- `WithFileWriter(config models.WriterConfiguration)` - Add file writer

## Output Format

Log entries are formatted as pipe-delimited strings:

```
LEVEL|TIMESTAMP|PREFIX|FUNCTION|CORRELATION_ID|FIELD=VALUE|error=ERROR_MESSAGE|MESSAGE
```

Example:
```
INFO|22:23:07.627|API|main.main|req-123|endpoint=/users|Request processed
```

## Integration with Existing Code

The fluent logging interface is fully compatible with the existing Arbor logger. You can mix and match usage patterns:

```go
// Existing pattern
logger := arbor.Logger()
logger.WithPrefix("API").WithCorrelationId("123")

// New fluent pattern
logger.Info().Str("endpoint", "/health").Msg("Health check")
```

## Performance

The fluent interface is designed to be efficient:
- Fields are stored in a map until the terminal method (`Msg`/`Msgf`) is called
- No string concatenation happens until the log entry is actually written
- Compatible with all existing Arbor writers (console, file, memory)
