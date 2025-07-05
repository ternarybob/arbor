# File Locking Prevention in FileWriter

## Overview

The FileWriter has been enhanced with a robust file locking prevention mechanism that automatically creates numbered file variants (001, 002, etc.) when the primary log file is locked or inaccessible.

## Problem Solved

Previously, if multiple processes or instances tried to write to the same log file simultaneously, or if a file was locked by another process, the FileWriter could fail to write log entries, potentially causing data loss.

## Solution Implementation

### Key Features

1. **Open-Write-Close Pattern**: Each write operation opens the file, writes data, and immediately closes it to minimize lock duration.

2. **Automatic Fallback**: If the primary file is locked, the system automatically tries numbered variants:
   - `logfile.log` → `logfile.001.log` → `logfile.002.log` → ... → `logfile.999.log`

3. **Dynamic Path Switching**: Once a numbered file is successfully created, the FileWriter switches to use that file for subsequent writes.

4. **Thread-Safe Operations**: All file operations are protected by mutexes to ensure thread safety.

### Implementation Details

#### Core Functions

1. **`writeToFile(data []byte) error`**
   - Main entry point for file writing
   - Tries primary file first, falls back to numbered variants if needed

2. **`attemptWrite(filePath string, data []byte) error`**
   - Attempts to write to a specific file path
   - Returns error if file is locked or inaccessible

3. **`writeToNumberedFile(data []byte) error`**
   - Creates numbered variants when primary file is locked
   - Tries variants from 001 to 999
   - Updates the FileWriter's current file path on success

#### File Naming Convention

- Primary file: `filename.ext`
- Numbered variants: `filename.001.ext`, `filename.002.ext`, etc.
- Supports up to 999 numbered variants

### Usage Examples

#### Basic Usage
```go
fw, err := filewriter.NewWithPath("app.log", 100, 10)
if err != nil {
    log.Fatal(err)
}
defer fw.Close()

// Write will automatically handle file locking
fw.Write([]byte(`{"level":"info","message":"Test message"}`))
```

#### Concurrent Usage
```go
// Multiple FileWriters can safely target the same file
fw1, _ := filewriter.NewWithPath("shared.log", 100, 10)
fw2, _ := filewriter.NewWithPath("shared.log", 100, 10)
fw3, _ := filewriter.NewWithPath("shared.log", 100, 10)

// All writers will work, creating numbered files as needed
fw1.Write([]byte(`{"message":"From writer 1"}`))
fw2.Write([]byte(`{"message":"From writer 2"}`))
fw3.Write([]byte(`{"message":"From writer 3"}`))
```

## Testing

The implementation includes comprehensive tests:

1. **TestFileLockingPrevention**: Basic file creation and writing
2. **TestNumberedFileCreation**: Verifies numbered file creation when primary is locked
3. **TestConcurrentWritesToSameFile**: Tests multiple concurrent writers

### Running Tests
```bash
cd arbor/filewriter
go test -v
```

### Running Demo
```bash
cd arbor/demo
go run file_locking_demo.go
```

## Benefits

1. **No Data Loss**: Messages are never lost due to file locking
2. **Automatic Recovery**: System automatically finds alternative files
3. **Transparent Operation**: Applications don't need to handle file locking logic
4. **Concurrent Safe**: Multiple processes can safely write to the same log file
5. **Scalable**: Supports up to 999 concurrent writers per file

## Monitoring

When a numbered file is created, the system logs an informational message:
```
File locked, switched to numbered variant: original=/path/to/file.log fallback=/path/to/file.001.log
```

This helps with monitoring and debugging file access patterns.

## Limitations

1. Maximum of 999 numbered variants per primary file
2. File naming follows a specific pattern (filename.NNN.ext)
3. Once switched to a numbered file, the FileWriter continues using that file

## Backward Compatibility

This enhancement is fully backward compatible. Existing code will continue to work without any changes, but will now benefit from automatic file locking prevention.
