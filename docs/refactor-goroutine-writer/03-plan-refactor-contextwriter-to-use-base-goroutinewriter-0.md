I have created the following plan after thorough exploration and analysis of the codebase. Follow the below plan verbatim. Trust the files and references. Do not re-verify what's written in the plan. Explore only when absolutely necessary. First implement all the proposed file changes and then I'll review all the changes together at the end.

### Observations

The current ContextWriter in `writers/contextwriter.go` is a simple synchronous implementation with an empty struct, no-op lifecycle methods, and direct calls to `common.Log()`. The refactoring will follow the exact pattern established in `writers/logstorewriter.go`: compose `IGoroutineWriter`, create a processor closure calling `common.Log()`, auto-start in constructor, and delegate all `IWriter` methods. The call site in `logger.go` line 46 currently calls `NewContextWriter()` without parameters and must be updated to pass a default `models.WriterConfiguration` with `TraceLevel` to capture all context logs.

### Approach

Refactor `ContextWriter` to compose `goroutineWriter` following the proven `logStoreWriter` pattern. Update the struct to hold an `IGoroutineWriter` field, modify the constructor to accept `models.WriterConfiguration` and create a processor calling `common.Log()`, auto-start the goroutine for backward compatibility, and delegate all `IWriter` methods to the composed writer. Update the call site in `logger.go` to pass a default configuration with `TraceLevel`.

### Reasoning

I listed the repository structure, read the current `contextwriter.go` implementation, examined the base `goroutinewriter.go`, reviewed the global `contextbuffer.go` pattern, analyzed the usage in `logger.go` line 46, and studied the completed `logstorewriter.go` refactoring to understand the established composition pattern.

## Proposed File Changes

### writers\contextwriter.go(MODIFY)

References: 

- writers\goroutinewriter.go
- writers\igoroutinewriter.go
- writers\iwriter.go
- common\contextbuffer.go
- models\writerconfiguration.go
- models\logevent.go
- common\logger.go

Refactor the ContextWriter struct and methods to compose goroutineWriter following the pattern from `writers/logstorewriter.go`.

**Struct Changes (Line 12):**
- Replace the empty struct with fields to hold the composed writer
- Add `writer IGoroutineWriter` field to hold the composed goroutine writer instance
- Optionally add `config models.WriterConfiguration` field for consistency with logStoreWriter pattern (though not strictly required)

**Import Changes (Lines 3-9):**
- Remove `encoding/json` import (no longer needed, handled by goroutineWriter)
- Keep `github.com/phuslu/log` import (needed for log.Level type in WithLevel method)
- Keep `github.com/ternarybob/arbor/common` import (needed for common.Log() in processor and error handling)
- Keep `github.com/ternarybob/arbor/models` import (needed for models.WriterConfiguration and models.LogEvent types)

**Constructor Refactoring (Lines 14-17):**
- Update signature from `NewContextWriter()` to `NewContextWriter(config models.WriterConfiguration) IWriter`
- Create internal logger using `common.NewLogger().WithContext("function", "NewContextWriter").GetLogger()` for error reporting
- Define processor closure with signature `func(models.LogEvent) error` that calls `common.Log(entry)` and returns `nil` (since common.Log has no return value)
- Call `NewGoroutineWriter(config, 1000, processor)` with buffer size 1000 matching logStoreWriter pattern
- Handle error return from NewGoroutineWriter - if error occurs, log fatal error using internal logger and panic with descriptive message (matching logStoreWriter error handling pattern)
- Immediately call `writer.Start()` on the created goroutine writer to maintain backward compatibility (auto-start pattern)
- Handle error return from Start() - if error occurs, log warning using internal logger (shouldn't happen in normal flow)
- Construct ContextWriter struct with the IGoroutineWriter instance and config
- Return the ContextWriter instance as IWriter interface

**Write() Method Refactoring (Lines 20-28):**
- Replace entire implementation with simple delegation: `return cw.writer.Write(p)`
- Remove JSON unmarshaling logic (now handled by goroutineWriter)
- Remove direct `common.Log()` call (now handled by processor function)
- Remove internal logger creation
- Method signature remains: `func (cw *ContextWriter) Write(p []byte) (n int, err error)`

**WithLevel() Method Refactoring (Lines 30-33):**
- Replace no-op implementation with delegation: `cw.writer.WithLevel(level)`
- Return `cw` (not `cw.writer`) to maintain return type as IWriter and allow method chaining on ContextWriter instance
- Optionally update `cw.config.Level = levels.FromLogLevel(level)` for consistency with logStoreWriter (though redundant since base manages level)
- Method signature remains: `func (cw *ContextWriter) WithLevel(level log.Level) IWriter`

**GetFilePath() Method Refactoring (Lines 35-38):**
- Replace direct return with delegation: `return cw.writer.GetFilePath()`
- This will still return empty string as expected (context writers don't write to files)
- Method signature remains: `func (cw *ContextWriter) GetFilePath() string`

**Close() Method Refactoring (Lines 40-43):**
- Replace no-op implementation with delegation: `return cw.writer.Close()`
- This now properly stops the goroutine and drains the buffer for graceful shutdown
- Method signature remains: `func (cw *ContextWriter) Close() error`

**Behavior Changes:**
- Changes from synchronous to asynchronous with buffering (1000 entries)
- Adds level filtering capability (previously no-op)
- Adds graceful shutdown with buffer draining (previously no-op)
- Maintains integration with global context buffer via `common.Log()` in processor
- Non-blocking writes with overflow handling (drops entries when buffer full)

**Integration with Global Context Buffer:**
- Processor function calls `common.Log(event)` which adds event to global context buffer in `common/contextbuffer.go`
- Global buffer's goroutine continues to handle batching and flushing to output channel
- Two-stage buffering architecture maintained: ContextWriter buffer → common.Log() → global context buffer → output channel
- Async nature provides additional layer of non-blocking writes before global buffer

### logger.go(MODIFY)

References: 

- writers\contextwriter.go(MODIFY)
- models\writerconfiguration.go
- levels\levels.go

Update the WithContextWriter method to pass a models.WriterConfiguration to the NewContextWriter constructor.

**Location:** Line 46 within the `WithContextWriter(contextID string)` method

**Current Code:**
- `contextWriter := writers.NewContextWriter()`

**Required Change:**
- Create a default configuration with TraceLevel to ensure all context logs are captured regardless of level
- Update to: `contextWriter := writers.NewContextWriter(models.WriterConfiguration{Level: levels.TraceLevel})`
- This uses a struct literal to create the configuration inline with only the Level field set

**Rationale for TraceLevel:**
- Context logging is typically used for debugging specific operations (e.g., job execution, request tracing)
- Using TraceLevel ensures all logs are captured in the context buffer regardless of the global log level
- The global context buffer will still apply its own filtering when flushing to the output channel
- This matches the intent of context-specific logging: capture everything for that context

**Import Verification:**
- Ensure `github.com/ternarybob/arbor/levels` is imported (check existing imports around lines 3-15)
- If not present, add to import block
- `github.com/ternarybob/arbor/models` is already imported on line 9

**No Other Changes Required:**
- The rest of the WithContextWriter method remains unchanged (lines 47-61)
- The returned contextWriter is still added to the writer list
- The correlation ID tagging continues to work as before
- Method signature remains: `func (l *logger) WithContextWriter(contextID string) ILogger`

**Backward Compatibility:**
- This is a breaking change to the NewContextWriter constructor signature
- However, the call site is internal to the arbor package (not public API)
- The WithContextWriter public method signature remains unchanged
- External users of arbor are not affected