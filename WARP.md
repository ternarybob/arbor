# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

Arbor is a Go logging library for APIs with structured logging, multiple output writers, and correlation tracking. It prioritizes comprehensive logging capabilities over raw performance, using a multi-writer architecture (Console, File, Memory/BoltDB).

## Development Commands

### Testing
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./writers
go test ./models
go test ./transformers

# Run single test function
go test -run TestMemoryWriter_Write

# Clean test cache and artifacts
go clean -testcache
rm -rf temp/
```

### Building and Code Quality
```bash
# Build all packages
go build ./...

# Format code (required before commits)
go fmt ./...
gofmt -s -w .

# Lint and vet
go vet ./...

# Get dependencies
go mod download
go mod verify
```

### Working with BoltDB Tests
Memory writer tests create temporary BoltDB files in `temp/` directory:
```bash
# Clean up test databases
rm -rf temp/

# Run memory writer tests specifically
go test -v ./writers -run MemoryWriter

# Full cleanup
rm -rf temp/ && go clean -testcache
```

## Architecture Overview

### Multi-Writer Pattern
Single log events broadcast to all registered writers simultaneously:
- **Console Writer**: Colored terminal output via phuslu backend
- **File Writer**: JSON (default) or text format with rotation, backup, size management
- **Memory Writer**: BoltDB persistence with TTL cleanup for log retrieval

### Global Registry System
Thread-safe writer registration (`registry.go`) enables cross-context logger access without dependency injection. Writers are registered globally and accessed by name constants:
- `WRITER_CONSOLE` - Console output writer
- `WRITER_FILE` - File output writer  
- `WRITER_MEMORY` - Memory/BoltDB writer for queries

### Context and Correlation Tracking
Logger context is mutable and chainable:
- **Correlation IDs**: First-class support for request tracing across layers
- **Context Data**: Key-value pairs attached to log entries
- **WithContextWriter()**: Creates context-specific loggers that write to standard writers plus context channel
- **Copy()**: Creates fresh logger with same writers but clean context

### Memory Writer Architecture
- Date-based BoltDB files: `temp/arbor_logs_YYMMDD.db`
- Background TTL cleanup every 1 minute (10 min default TTL)
- Correlation ID-based storage and retrieval
- Thread-safe with shared database instances
- Non-blocking buffered writes (1000 entry buffer limit)

### Level Filtering
- Occurs at writer level (not logger level)
- Each writer maintains independent level configuration
- String parsing: "trace", "debug", "info", "warn", "error", "fatal", "panic", "disabled"
- Use `WithLevelFromString()` for config integration

## Key Packages

**Root (`/`)**: Logger interface/implementation (`ilogger.go`, `logger.go`), registry (`registry.go`, `iregistry.go`), log events (`logevent.go`, `ilogevent.go`), levels (`levels.go`)

**`writers/`**: Writer implementations and interfaces
- `IWriter`: Basic write functionality with level filtering
- `IMemoryWriter`: Extends IWriter with retrieval capabilities
- Console, File, Memory, LogStore, Context, WebSocket writers

**`models/`**: Data structures (`WriterConfiguration`, `LogEvent`, `GinLogEvent`)

**`levels/`**: Log level definitions and string parsing utilities

**`transformers/`**: Framework integrations (`gintransformer.go` converts Gin logs to arbor format)

**`common/`**: Shared utilities, internal logging, context buffer management

## Important Constants

- Memory Writer Buffer: 1000 entries per correlation ID
- TTL: 10 minutes default
- Cleanup Interval: 1 minute
- BoltDB Naming: `temp/arbor_logs_YYMMDD.db`

## Testing Patterns

- Tests create isolated instances to avoid global state conflicts
- Memory writer tests verify BoltDB persistence
- Registry thread-safety validation included
- Clean up `temp/` directory between test runs

## CI/CD

GitHub Actions workflow (`.github/workflows/ci.yml`):
- Runs on push to main, PRs, and version tags
- Tests, go vet, formatting checks, build validation
- Auto-releases on main branch pushes (patch increment)
- Manual releases via version tags (e.g., `v1.2.3`)

## Framework Integration

**Gin**: Use `GinWriter()` method and `transformers/gintransformer.go` to integrate Gin framework logs with arbor while preserving correlation context.
