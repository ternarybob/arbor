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
- Console Writer: Colored output using phuslu backend
- File Writer: Rotation, backup, size management, configurable JSON/text output format
- Memory Writer: BoltDB persistence with TTL and cleanup

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

## Important Implementation Details

### Memory Writer Architecture
- Uses date-based BoltDB files (`temp/arbor_logs_YYMMDD.db`)
- Implements TTL with background cleanup every minute
- Thread-safe with shared database instances
- Correlation ID-based log storage and retrieval

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
- **`common/`**: Shared utilities and internal logging

### Key Constants and Configuration
- **Memory Writer**: Buffer limit 1000 entries per correlation ID, 10-minute TTL, 1-minute cleanup interval
- **BoltDB Files**: Date-based naming (`temp/arbor_logs_YYMMDD.db`)
- **File Writer**: Supports JSON (default) and text output formats via `TextOutput` configuration
- **Global State**: Shared database instances with mutex protection