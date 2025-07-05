# Writer Interfaces and Naming Conventions

This directory contains the interface definitions for all log writers in the arbor package.

## Interface Hierarchy

### IWriter (Base Interface)
```go
type IWriter interface {
    io.Writer
    Close() error
}
```
- **Purpose**: Base interface that all writers must implement
- **Methods**: 
  - `Write([]byte) (int, error)` - from `io.Writer`
  - `Close() error` - for cleanup and resource management

### ILevelWriter (Level Filtering)
```go
type ILevelWriter interface {
    IWriter
    SetMinLevel(level interface{}) error
}
```
- **Purpose**: Writers that support log level filtering
- **Additional Methods**: `SetMinLevel(level interface{}) error`

### IBufferedWriter (Buffering Support)
```go
type IBufferedWriter interface {
    IWriter
    Flush() error
    SetBufferSize(size int) error
}
```
- **Purpose**: Writers that support buffering operations
- **Additional Methods**: 
  - `Flush() error` - force buffered data to be written
  - `SetBufferSize(size int) error` - configure buffer size

### IRotatableWriter (Log Rotation)
```go
type IRotatableWriter interface {
    IWriter
    Rotate() error
    SetMaxFiles(maxFiles int)
}
```
- **Purpose**: Writers that support log file rotation
- **Additional Methods**:
  - `Rotate() error` - manually trigger rotation
  - `SetMaxFiles(maxFiles int)` - set maximum number of files to keep

### IFullFeaturedWriter (Complete Functionality)
```go
type IFullFeaturedWriter interface {
    IWriter
    ILevelWriter
    IBufferedWriter
    IRotatableWriter
}
```
- **Purpose**: Combined interface for writers with all features

## Writer Implementations

### ConsoleWriter
- **Interface Compliance**: `IWriter`
- **Constructor**: `writers.NewConsoleWriter() *ConsoleWriter`
- **Purpose**: Writes formatted output to console/stdout

### MemoryWriter  
- **Interface Compliance**: `IWriter`
- **Constructor**: `writers.NewMemoryWriter() *MemoryWriter`
- **Purpose**: Stores log entries in memory for retrieval

### FileWriter
- **Interface Compliance**: `IWriter`, `IBufferedWriter`, `IRotatableWriter`
- **Constructors**: 
  - `writers.NewFileWriter(filePath string, bufferSize, maxFiles int) (*FileWriter, error)`
  - `writers.NewFileWriterWithLevel(filePath string, bufferSize, maxFiles int, minLevel log.Level) (*FileWriter, error)`
  - `writers.NewFileWriterWithPattern(filePath, pattern, format string, bufferSize, maxFiles int) (*FileWriter, error)`
  - `writers.NewFileWriterWithPatternAndLevel(filePath, pattern, format string, bufferSize, maxFiles int, minLevel log.Level) (*FileWriter, error)`
- **Purpose**: Writes logs to files with rotation and level filtering support

## Naming Conventions

### File Names
- All lowercase: `consolewriter.go`, `filewriter.go`, `memorywriter.go`
- Test files: `*_test.go` suffix
- Interface files: `iwriter.go`

### Constructor Functions
- Pattern: `New{WriterType}()` or `New{WriterType}With{Feature}()`
- Examples:
  - ✅ `NewConsoleWriter()`
  - ✅ `NewMemoryWriter()` 
  - ✅ `NewFileWriter()`
  - ✅ `NewFileWriterWithLevel()`
  - ✅ `NewFileWriterWithPattern()`

### Type Names
- Pattern: `{Purpose}Writer`
- Examples:
  - ✅ `ConsoleWriter`
  - ✅ `MemoryWriter`
  - ✅ `FileWriter`

### Interface Names  
- Pattern: `I{Purpose}Writer`
- Examples:
  - ✅ `IWriter`
  - ✅ `ILevelWriter`
  - ✅ `IBufferedWriter`
  - ✅ `IRotatableWriter`

### Method Names
- Use standard Go conventions
- Interface methods are concise and focused
- Implementation methods follow the same patterns

## Usage Examples

```go
// Using through interface
var writer interfaces.IWriter = writers.NewConsoleWriter()
writer.Write([]byte("test"))
writer.Close()

// Using full-featured writer
fileWriter, err := writers.NewFileWriter("/path/to/log.txt", 1024, 10)
if err != nil {
    return err
}
defer fileWriter.Close()

// Use as buffered writer
var buffered interfaces.IBufferedWriter = fileWriter
buffered.Flush()

// Use as rotatable writer  
var rotatable interfaces.IRotatableWriter = fileWriter
rotatable.SetMaxFiles(5)
rotatable.Rotate()
```

## Benefits

1. **Consistency**: All writers follow the same naming patterns
2. **Interface Compliance**: Writers implement appropriate interfaces for their capabilities
3. **Extensibility**: New writers can easily implement the same interfaces
4. **Testability**: Interface-based design enables easy mocking and testing
5. **Modularity**: Clear separation of concerns through focused interfaces
