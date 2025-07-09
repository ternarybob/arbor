# Level Configuration

Arbor supports configuring log levels both programmatically and from string values. This allows for flexible configuration management where the consuming application (like `satus`) can handle configuration loading and pass string values to arbor.

## Setting Log Levels

### 1. Using LogLevel Constants

```go
logger := arbor.Logger().WithLevel(arbor.InfoLevel)
```

### 2. Using String Values

```go
logger := arbor.Logger().WithLevelFromString("info")
```

## Supported Level Strings

The `WithLevelFromString` method supports the following level strings (case-insensitive):

- `"trace"` - Most verbose, includes all log messages
- `"debug"` - Debug messages and above
- `"info"` - Informational messages and above (default)
- `"warn"` or `"warning"` - Warning messages and above
- `"error"` - Error messages and above
- `"fatal"` - Fatal messages and above
- `"panic"` - Only panic messages
- `"disabled"` or `"off"` - No logging

## Error Handling

If an invalid level string is provided, the logger will:
1. Log a warning message about the invalid level
2. Fall back to `InfoLevel` as the default
3. Continue to function normally

## Usage in Applications

This design allows consuming applications to handle their own configuration loading:

```go
// Application code (e.g., in satus)
cfg := loadConfigFromFile("config.toml")  // Application handles config loading

// Apply to arbor logger
logger := arbor.Logger().
    WithConsoleWriter(models.WriterConfiguration{
        Type: models.LogWriterTypeConsole,
    }).
    WithLevelFromString(cfg.Logging.LogLevel)  // Pass string value from config
```

## Order Independence

The logger starts with `InfoLevel` by default and can be updated at any time:

```go
// Logger starts with INFO level
logger := arbor.Logger()

// Configure writers
logger.WithConsoleWriter(/* ... */)

// Update level later from configuration
logger.WithLevelFromString("debug")
```

This ensures no dependency on the order of configuration loading vs logger initialization.

## Chaining

`WithLevelFromString` returns the logger instance and can be chained with other methods:

```go
logger := arbor.Logger().
    WithConsoleWriter(config).
    WithLevelFromString("debug").
    WithCorrelationId("12345").
    WithPrefix("MyApp")
```
