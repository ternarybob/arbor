# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Core Principles

### Code Quality Standards
- **Single Responsibility**: Functions do one thing well
- **Clear Naming**: Descriptive, intention-revealing names
- **Function Size**: Maximum 80 lines, ideally 20-40
- **Error Handling**: Comprehensive validation and error management
- **No Dead Code**: Remove unused imports, variables, functions

### Professional Output
- **Human-Authored Appearance**: No AI attribution or generation markers
- **Production Ready**: Code passes enterprise review standards
- **Clean Architecture**: Follow SOLID principles and design patterns
- **Consistent Style**: Language-specific conventions and formatting

## Project Overview

Arbor is a comprehensive Go logging library designed for APIs with structured logging, multiple output writers, and advanced correlation tracking. It provides a multi-writer architecture supporting Console, File, and Memory (BoltDB) writers with focus on logging comprehensiveness over raw performance.

## Development Commands

### Building and Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test file
go test -v ./writers/memorywriter_boltdb_test.go

# Run tests for specific package
go test ./writers

# Run single test function
go test -run TestMemoryWriter_Write

# Build the module
go build ./...

# Format code
go fmt ./...

# Lint and vet
go vet ./...
```

### Working with BoltDB Tests
```bash
# Memory writer tests use BoltDB - clean up test databases
rm -rf temp/

# Run memory writer specific tests
go test -v ./writers -run MemoryWriter

# Clean up all test artifacts
rm -rf temp/ && go clean -testcache
```

## Architecture Overview

### Core Components

**Logger (`logger.go`, `ilogger.go`)**
- Main logger implementation with fluent API
- Manages context data (correlation IDs, prefixes)
- Coordinates with global writer registry

**Writer System (`writers/`)**
- `IWriter` interface: Basic write functionality with level filtering
- `IMemoryWriter` interface: Extends IWriter with retrieval capabilities
- Console Writer: Colored output using phuslu backend, synchronous writes (~50-100μs)
- File Writer: Rotation, backup, size management, configurable JSON/text output format, synchronous writes (~50-100μs)
- Memory Writer: Async writes via LogStoreWriter (ChannelWriter base), BoltDB persistence with TTL and cleanup
- ChannelWriter: Reusable async buffered writer base (1000-entry buffer, non-blocking writes, automatic drain on shutdown)
- ContextWriter: Lightweight writer that directly sends log events to singleton context buffer (no internal async layer)
- LogStoreWriter: Async writer using ChannelWriter base to write to ILogStore (in-memory + optional BoltDB)

**Global Registry (`registry.go`, `iregistry.go`)**
- Thread-safe writer registration and management
- Enables cross-context logger access
- Supports broadcast logging to all registered writers

**Log Events (`logevent.go`, `ilogevent.go`)**
- Fluent API for building structured log entries
- Marshals to JSON for writer consumption
- Handles field types (strings, integers, errors, durations)

**Level Management (`levels.go`, `levels/levels.go`)**
- String-based level parsing for configuration integration
- Conversion between arbor LogLevel and phuslu log.Level types

### Key Design Patterns

**Registry Pattern**: Global writer registry allows logger access across application contexts without dependency injection

**Multi-Writer**: Single log event broadcasts to all registered writers (console, file, memory) simultaneously

**Correlation Tracking**: First-class correlation ID support for request tracing across application layers

**Memory Persistence**: BoltDB-backed storage with automatic TTL cleanup for log retrieval and debugging

**Framework Integration**: Gin transformer (`transformers/gintransformer.go`) converts framework logs to arbor format while preserving correlation context

### Channel-Based Logging APIs

Arbor provides two channel-based APIs for streaming logs to consumers:

**SetContextChannel/SetContextChannelWithBuffer**: Singleton context buffer for job/request tracking
- All `WithContextWriter` loggers share the same buffer
- Use for capturing all logs related to a specific context (e.g., "all logs for job-123")
- Managed via `common.Start()` and `common.Stop()`
- Batches logs by correlation ID before sending to channel

**SetChannel/SetChannelWithBuffer**: Multiple independent named channels for general streaming
- Each channel has its own ChannelWriter and ChannelBuffer
- Use for streaming logs to WebSocket clients, external services, or custom consumers
- Managed via `SetChannel()` and `UnregisterChannel()`
- Each channel receives all logs independently (not filtered by correlation ID)

**Use Cases**:
- **SetContextChannel**: "Capture all logs for job-123 and store in database"
- **SetChannel**: "Stream all application logs to WebSocket clients" or "Send error logs to Slack"

### Key Design Patterns (Updated)

**Registry Pattern**: Global writer registry allows logger access across application contexts without dependency injection

**Multi-Writer**: Single log event broadcasts to all registered writers (console, file, memory) simultaneously

**Correlation Tracking**: First-class correlation ID support for request tracing across application layers

**Memory Persistence**: BoltDB-backed storage with automatic TTL cleanup for log retrieval and debugging

**Framework Integration**: Gin transformer (`transformers/gintransformer.go`) converts framework logs to arbor format while preserving correlation context

**Simplified ContextWriter**: Direct synchronous writes to singleton context buffer (no internal ChannelWriter wrapper), reduces complexity while maintaining non-blocking behavior via singleton buffer

**Named Channel Writers**: Multiple independent channel loggers with per-channel batching and lifecycle management, enables flexible streaming to multiple consumers

**Dual Channel APIs**: Singleton context channel for job tracking vs. multiple named channels for general streaming, provides flexibility for different use cases

## Important Implementation Details

### Memory Writer Architecture
- Uses date-based BoltDB files (`temp/arbor_logs_YYMMDD.db`)
- Implements TTL with background cleanup every minute
- Thread-safe with shared database instances
- Correlation ID-based log storage and retrieval
- Async writes via LogStoreWriter (ChannelWriter base with 1000-entry buffer)

### ContextWriter Architecture

The ContextWriter was simplified in the recent refactoring to reduce complexity:

**Design**: Lightweight synchronous writer that directly calls `common.Log()` to send events to the singleton context buffer
- **No Internal Async Layer**: Unlike the old design (which wrapped ChannelWriter), the new ContextWriter has no internal goroutine or channel
- **Batching**: Handled by the singleton context buffer (`common.contextbuffer`), not by ContextWriter itself
- **Lifecycle**: No cleanup needed for ContextWriter instances; singleton buffer is managed via `common.Start()` and `common.Stop()`
- **Performance**: Write operations are fast (~10-50μs) since they only append to the singleton buffer under mutex
- **Thread-Safe**: Uses RWMutex for level changes and safe concurrent writes

### ChannelWriter Base

The ChannelWriter provides a reusable async buffered writer pattern:

**Purpose**: Reusable base for async writers (LogStoreWriter, named channel writers created by SetChannel)
- **Architecture**: Buffered channel (default 1000 entries) + background goroutine + processor function
- **Lifecycle**: Start/Stop methods for goroutine control, automatic buffer drain on Close()
- **Overflow Behavior**: Non-blocking writes, drops entries with warning when buffer is full
- **Used By**:
  - LogStoreWriter (for memory writer)
  - Named channel writers (created by SetChannel/SetChannelWithBuffer)
- **Performance**: ~100μs non-blocking writes, supports 10,000+ logs/sec throughput

### Level Filtering
- Occurs at writer level for efficiency
- Each writer maintains its own level configuration
- String parsing supports: trace, debug, info, warn, error, fatal, panic, disabled

### Context Management
- Logger context is mutable and chainable
- `Copy()` method creates fresh logger with same writers but clean context
- Context includes correlation ID, prefix, and custom key-value pairs

### Thread Safety
- All operations use appropriate mutexes (RWMutex for registries)
- BoltDB instances are shared and managed globally
- Writers handle concurrent access safely
- ChannelWriter uses RWMutex for config access and separate mutex for running state
- ContextWriter uses RWMutex for thread-safe level changes
- Named channel buffers tracked in global map with RWMutex protection

## Testing Approach

Tests are organized by component with comprehensive coverage:
- Unit tests for individual components
- Integration tests for writer interactions  
- Memory writer tests include BoltDB persistence verification
- Level parsing and conversion testing
- Registry thread-safety validation

Most tests create isolated instances to avoid global state conflicts. Memory writer tests may create temporary BoltDB files that should be cleaned up.

## Package Structure

### Core Packages
- **Root (`/`)**: Main logger interface and implementation, registry, log events
- **`writers/`**: All writer implementations (console, file, memory) and interfaces
- **`models/`**: Data structures (WriterConfiguration, LogEvent, GinLogEvent)
- **`levels/`**: Log level definitions and string parsing utilities
- **`transformers/`**: Framework integrations (Gin transformer)
- **`common/`**: Shared utilities, internal logging, singleton context buffer (`contextbuffer.go`), and per-instance channel buffer (`channelbuffer.go`)

### Key Constants and Configuration
- **Memory Writer**: Buffer limit 1000 entries per correlation ID, 10-minute TTL, 1-minute cleanup interval
- **BoltDB Files**: Date-based naming (`temp/arbor_logs_YYMMDD.db`)
- **File Writer**: Supports JSON (default) and text output formats via `TextOutput` configuration
- **Global State**: Shared database instances with mutex protection
- **Named Channel Writers**: Default batch size 5 events, default flush interval 1 second, queue size calculated as max(1000, batchSize * 100)
- **ChannelWriter Buffer**: Default 1000 entries, non-blocking writes, automatic drain on shutdown