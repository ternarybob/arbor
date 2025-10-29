I have created the following plan after thorough exploration and analysis of the codebase. Follow the below plan verbatim. Trust the files and references. Do not re-verify what's written in the plan. Explore only when absolutely necessary. First implement all the proposed file changes and then I'll review all the changes together at the end.

### Observations

The current `ContextWriter` uses composition with `IChannelWriter` to achieve async behavior, creating an extra layer of buffering and goroutine management. However, this is redundant because `common.contextbuffer` already provides singleton batching for context logs. The refactoring will simplify the architecture by making `ContextWriter` a lightweight synchronous writer that directly feeds the singleton context buffer. The `common.Log()` function is already non-blocking (quick mutex-protected append), so removing the async layer won't impact performance. All existing tests expect `IWriter` interface compliance and non-blocking behavior, which will be preserved.

### Approach

Refactor `ContextWriter` to eliminate the `IChannelWriter` composition layer. Replace the internal `writer` field with direct implementation of all `IWriter` methods. The `Write()` method will unmarshal JSON, perform log level filtering, and call `common.Log()` synchronously. Remove all delegation methods (`WithLevel`, `GetFilePath`, `Close`) and implement them directly with appropriate behavior for a context writer. Store the configuration directly in the struct for level filtering. This simplifies the architecture while maintaining backward compatibility and test compliance.

### Reasoning

I explored the repository structure, read the three core files mentioned (`contextwriter.go`, `logger.go`, `contextbuffer.go`), examined the `IWriter` interface and `channelwriter.go` to understand the composition pattern, reviewed existing tests in `channelwriter_test.go` to understand expected behavior, and verified that `common.Log()` is non-blocking by examining its implementation in `contextbuffer.go`.

## Mermaid Diagram

sequenceDiagram
    participant Logger as Logger
    participant CW as ContextWriter
    participant CB as common.contextbuffer
    participant Client as Client Channel

    Note over CW: Before: ContextWriter → ChannelWriter → common.Log()
    Note over CW: After: ContextWriter → common.Log() (direct)

    Logger->>CW: Write(jsonBytes)
    CW->>CW: Unmarshal JSON to LogEvent
    CW->>CW: Check log level filter
    alt Level passes filter
        CW->>CB: common.Log(logEvent)
        CB->>CB: Append to singleton buffer
        alt Buffer full or timer tick
            CB->>Client: Send batch []LogEvent
        end
    else Level filtered
        CW->>Logger: Return (skip)
    end
    CW->>Logger: Return bytes written

    Note over CW: Simplified: No internal goroutine<br/>No internal channel<br/>No async layer<br/>Direct synchronous call

## Proposed File Changes

### writers\contextwriter.go(MODIFY)

References: 

- writers\iwriter.go
- common\contextbuffer.go
- writers\channelwriter.go
- models\logevent.go
- levels\levels.go

Refactor the entire file to remove `IChannelWriter` composition and implement `IWriter` directly:

1. **Remove imports**: Remove the unused import for `IChannelWriter` if it becomes unused after refactoring.

2. **Simplify struct** (line 11-14): Replace the current struct with:
   - Remove `writer IChannelWriter` field
   - Keep `config models.WriterConfiguration` field
   - Add `configMux sync.RWMutex` field for thread-safe level changes

3. **Simplify constructor** (line 16-39): Refactor `NewContextWriter` to:
   - Remove the internal logger creation (lines 18-19)
   - Remove the processor closure definition (lines 21-25)
   - Remove the `newAsyncWriter` call and error handling (lines 27-32)
   - Directly return `&ContextWriter{config: config}` after validating the config
   - No goroutine startup needed since this becomes a synchronous writer

4. **Reimplement Write method** (line 41-44): Replace delegation with direct implementation:
   - Accept `p []byte` parameter
   - Return early if `len(p) == 0` with `(0, nil)`
   - Unmarshal JSON into `models.LogEvent` using `json.Unmarshal(p, &logEvent)`
   - Return error if unmarshaling fails: `(0, err)`
   - Acquire read lock on `configMux` to get `minLevel := cw.config.Level.ToLogLevel()`
   - Release read lock
   - Check if `logEvent.Level < minLevel`, if so return `(len(p), nil)` (filtered)
   - Call `common.Log(logEvent)` to send to singleton context buffer
   - Return `(len(p), nil)` on success
   - This implementation mirrors the logic in `channelwriter.go` Write method (lines 121-153) but without the async channel layer

5. **Reimplement WithLevel method** (line 46-51): Replace delegation with direct implementation:
   - Accept `level log.Level` parameter
   - Acquire write lock on `configMux`
   - Set `cw.config.Level = levels.FromLogLevel(level)`
   - Release write lock
   - Return `cw` for method chaining
   - This matches the pattern in `channelwriter.go` WithLevel method (lines 155-160)

6. **Reimplement GetFilePath method** (line 53-56): Replace delegation with direct implementation:
   - Simply return empty string `""` since context writers don't write to files
   - Remove the delegation to `cw.writer.GetFilePath()`

7. **Reimplement Close method** (line 58-61): Replace delegation with direct implementation:
   - Return `nil` immediately since there's no goroutine or resources to clean up
   - The singleton context buffer lifecycle is managed separately via `common.Stop()`
   - Remove the delegation to `cw.writer.Close()`

8. **Add necessary imports**: Ensure the following imports are present:
   - `encoding/json` for unmarshaling
   - `sync` for `RWMutex`
   - Keep existing imports: `github.com/phuslu/log`, `github.com/ternarybob/arbor/common`, `github.com/ternarybob/arbor/levels`, `github.com/ternarybob/arbor/models`

9. **Update package comment** (line 10): Update the comment to reflect the simplified design: "ContextWriter is a lightweight writer that sends log events directly to the global singleton context buffer managed by common.contextbuffer."

The refactored implementation will be synchronous and lightweight, with no internal goroutines or channels. All batching and async behavior is handled by the singleton `common.contextbuffer`. This maintains backward compatibility with the `IWriter` interface while simplifying the architecture.