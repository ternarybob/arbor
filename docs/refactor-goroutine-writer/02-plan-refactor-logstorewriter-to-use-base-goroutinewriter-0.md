I have created the following plan after thorough exploration and analysis of the codebase. Follow the below plan verbatim. Trust the files and references. Do not re-verify what's written in the plan. Explore only when absolutely necessary. First implement all the proposed file changes and then I'll review all the changes together at the end.

### Observations

The current `logStoreWriter` in `writers/logstorewriter.go` contains duplicate goroutine infrastructure (buffer channel, done channel, processBuffer goroutine) that is now available in the base `goroutineWriter` implementation. The refactoring requires replacing this duplicate code with composition while maintaining backward compatibility with the existing `LogStoreWriter()` constructor signature and automatic goroutine startup behavior. The tests in `writers/memorywriter_test.go` show that the constructor is called without any explicit `Start()` invocation, meaning the goroutine must start automatically within the constructor.

### Approach

Refactor `logStoreWriter` to compose `goroutineWriter` instead of implementing its own goroutine infrastructure. The struct will hold an `IGoroutineWriter` instance and delegate most operations to it. The constructor will create the `goroutineWriter` with a processor function that calls `store.Store()`, then immediately start it to maintain backward compatibility. The struct will implement `IWriter` interface by delegating to the embedded writer while maintaining access to the `ILogStore` reference.

### Reasoning

I explored the repository structure, read the current `logStoreWriter` implementation, examined the new `goroutineWriter` base class, reviewed the interface hierarchy (`IWriter` and `IGoroutineWriter`), and analyzed the usage patterns in `memorywriter_test.go` to understand the expected behavior and backward compatibility requirements.

## Mermaid Diagram

sequenceDiagram
    participant Client as Test/Logger
    participant LSW as logStoreWriter
    participant GW as goroutineWriter
    participant Processor as Processor Func
    participant Store as ILogStore

    Client->>LSW: LogStoreWriter(store, config)
    LSW->>LSW: Create processor: func(e) { store.Store(e) }
    LSW->>GW: NewGoroutineWriter(config, 1000, processor)
    GW-->>LSW: IGoroutineWriter instance
    LSW->>GW: Start()
    GW->>GW: Launch processBuffer() goroutine
    LSW-->>Client: IWriter instance

    Client->>LSW: Write(jsonData)
    LSW->>GW: Write(jsonData)
    GW->>GW: Unmarshal & filter
    GW->>GW: Send to buffer channel
    GW-->>LSW: bytes written
    LSW-->>Client: bytes written

    Note over GW,Processor: Background goroutine
    GW->>GW: Read from buffer
    GW->>Processor: processor(logEvent)
    Processor->>Store: Store(logEvent)
    Store-->>Processor: error/nil
    Processor-->>GW: error/nil

    Client->>LSW: Close()
    LSW->>GW: Close()
    GW->>GW: Stop() - signal shutdown
    GW->>GW: Drain buffer
    GW->>Processor: Process remaining entries
    Processor->>Store: Store(logEvent)
    GW-->>LSW: nil
    LSW-->>Client: nil

## Proposed File Changes

### writers\logstorewriter.go(MODIFY)

References: 

- writers\goroutinewriter.go
- writers\igoroutinewriter.go
- writers\iwriter.go
- writers\ilogstore.go
- models\writerconfiguration.go
- models\logevent.go
- common\logger.go

Refactor the `logStoreWriter` struct to use composition with `goroutineWriter` instead of implementing its own goroutine infrastructure.

**Struct Changes:**
- Remove the `buffer chan models.LogEvent` field (now handled by `goroutineWriter`)
- Remove the `done chan bool` field (now handled by `goroutineWriter`)
- Add `writer IGoroutineWriter` field to hold the composed goroutine writer instance
- Keep the `store ILogStore` field (still needed for the processor function)
- Keep the `config models.WriterConfiguration` field (may be needed for future use)

**Constructor Refactoring (`LogStoreWriter` function):**
- Create a processor function that captures the `store` reference and calls `store.Store(entry)` for each log event
- The processor function signature should be `func(models.LogEvent) error` to match `goroutineWriter` expectations
- Call `NewGoroutineWriter(config, 1000, processorFunc)` to create the base writer (buffer size 1000 matches current implementation)
- Handle the error return from `NewGoroutineWriter()` - if error occurs, log warning using `common.NewLogger()` and return a no-op writer or panic (decide based on current error handling philosophy)
- Immediately call `writer.Start()` on the created goroutine writer to maintain backward compatibility (current implementation starts goroutine in constructor)
- Store the `IGoroutineWriter` instance in the struct's `writer` field
- Return the `logStoreWriter` instance as `IWriter` interface

**Remove `processBuffer()` Method:**
- Delete the entire `processBuffer()` method (lines 42-60) as this functionality is now handled by `goroutineWriter.processBuffer()`

**Refactor `Write()` Method:**
- Replace the entire implementation with a simple delegation: `return lsw.writer.Write(data)`
- Remove the internal logger creation, JSON unmarshaling, level filtering, and channel operations (all now handled by `goroutineWriter`)
- The method signature remains: `func (lsw *logStoreWriter) Write(data []byte) (int, error)`

**Refactor `WithLevel()` Method:**
- Delegate to the composed writer: `lsw.writer.WithLevel(level)`
- Return `lsw` (not `lsw.writer`) to maintain the return type as `IWriter` and allow method chaining on the `logStoreWriter` instance
- Optionally update `lsw.config.Level` to keep local config in sync: `lsw.config.Level = levels.FromLogLevel(level)`

**Refactor `GetFilePath()` Method:**
- Delegate to the composed writer: `return lsw.writer.GetFilePath()`
- This will return empty string (as expected, since stores don't write to files)

**Refactor `Close()` Method:**
- Delegate to the composed writer: `return lsw.writer.Close()`
- Remove the manual `close(lsw.done)` call (now handled by `goroutineWriter.Close()` which calls `Stop()`)
- Remove the comment about WaitGroup (now properly implemented in `goroutineWriter`)

**Import Changes:**
- Remove `encoding/json` import (no longer needed, handled by `goroutineWriter`)
- Remove `github.com/phuslu/log` import if only used for internal logging (check if still needed)
- Remove `github.com/ternarybob/arbor/levels` import if only used in `Write()` (check if still needed for `WithLevel()`)
- Keep `github.com/ternarybob/arbor/common` import if used for error handling in constructor
- Keep `github.com/ternarybob/arbor/models` import (needed for types)

**Error Handling Considerations:**
- In the constructor, if `NewGoroutineWriter()` returns an error (e.g., nil processor), decide whether to:
  - Return a no-op writer that silently drops logs
  - Panic (current code doesn't handle constructor errors)
  - Log a fatal error and exit
- If `writer.Start()` returns an error (e.g., already running), log a warning but continue (shouldn't happen in normal flow)

**Backward Compatibility Verification:**
- The constructor signature remains: `LogStoreWriter(store ILogStore, config models.WriterConfiguration) IWriter`
- The return type remains `IWriter` (not `IGoroutineWriter`)
- The goroutine starts automatically in the constructor (via `Start()` call)
- All `IWriter` methods (`Write`, `WithLevel`, `GetFilePath`, `Close`) remain functional
- The behavior is identical: async writes with 1000-entry buffer, level filtering, graceful shutdown

**Testing Impact:**
- No changes required to `writers/memorywriter_test.go` - all tests should pass without modification
- The async behavior remains the same (tests use `time.Sleep()` to wait for processing)
- The `Close()` behavior remains the same (graceful shutdown with buffer draining)