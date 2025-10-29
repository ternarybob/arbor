I have created the following plan after thorough exploration and analysis of the codebase. Follow the below plan verbatim. Trust the files and references. Do not re-verify what's written in the plan. Explore only when absolutely necessary. First implement all the proposed file changes and then I'll review all the changes together at the end.

### Observations

The task requires adding two new methods to the `ILogger` interface that allow clients to set up multiple named channel loggers. Each channel should receive batched log events independently. The existing `SetContextChannel` methods use a singleton buffer in `common/contextbuffer.go`, but the new methods need per-channel buffers. The architecture follows the pattern used by `ContextWriter` and `LogStoreWriter`, where a `ChannelWriter` is created with a processor function that handles log events. The key challenge is implementing per-channel batching similar to the singleton context buffer but instantiable for each named channel.

### Approach

Create a new `channelBuffer` type in the `common` package that provides per-instance batching (unlike the singleton `contextbuffer`). Implement `SetChannel` and `SetChannelWithBuffer` methods in `logger.go` that create a `channelBuffer` instance, wrap it with a `ChannelWriter` using a processor function, and register the writer with the provided name. The `SetChannel` method uses default batching parameters (5 events, 1 second) while `SetChannelWithBuffer` accepts custom parameters. Each named channel operates independently with its own buffer and goroutine.

### Reasoning

I explored the codebase structure, read the existing `SetContextChannel` implementation and `common/contextbuffer.go` to understand the batching pattern, examined `ChannelWriter` and related writers (`ContextWriter`, `LogStoreWriter`) to understand the composition pattern, reviewed the registry system to understand writer registration, and studied the integration test to confirm the expected behavior of batched channel delivery.

## Mermaid Diagram

sequenceDiagram
    participant Client
    participant Logger
    participant ChannelBuffer
    participant ChannelWriter
    participant Registry
    participant BufferGoroutine

    Client->>Logger: SetChannelWithBuffer(name, ch, batchSize, interval)
    Logger->>ChannelBuffer: NewChannelBuffer(ch, batchSize, interval)
    ChannelBuffer->>BufferGoroutine: Start background goroutine
    Note over BufferGoroutine: Ticker flushes periodically
    
    Logger->>Logger: Create processor closure
    Note over Logger: processor = func(event) { buffer.Log(event) }
    
    Logger->>ChannelWriter: NewChannelWriter(config, 1000, processor)
    Logger->>ChannelWriter: Start()
    Note over ChannelWriter: Starts processing goroutine
    
    Logger->>Registry: RegisterWriter(name, writer)
    
    Client->>Logger: Info().Msg("log message")
    Logger->>ChannelWriter: Write(jsonBytes)
    ChannelWriter->>ChannelBuffer: processor(logEvent)
    ChannelBuffer->>ChannelBuffer: Accumulate in buffer
    
    alt Buffer full (batchSize reached)
        ChannelBuffer->>Client: Send batch to channel
    else Timer tick
        BufferGoroutine->>ChannelBuffer: Flush on interval
        ChannelBuffer->>Client: Send batch to channel
    end
    
    Client->>Client: Receive []LogEvent batch

## Proposed File Changes

### ilogger.go(MODIFY)

References: 

- models\logevent.go

Add two new method signatures to the `ILogger` interface:

1. **SetChannel** (after line 12): Add `SetChannel(name string, ch chan []models.LogEvent)` - this method allows clients to register a named channel logger with default batching settings (batch size 5, flush interval 1 second)

2. **SetChannelWithBuffer** (after line 12): Add `SetChannelWithBuffer(name string, ch chan []models.LogEvent, batchSize int, flushInterval time.Duration)` - this method allows clients to register a named channel logger with custom batching settings

These methods should be placed right after the existing `SetContextChannelWithBuffer` method to group related channel configuration methods together. The signatures mirror the existing context channel methods but add a `name` parameter for identifying the channel writer in the registry.

### common\channelbuffer.go(NEW)

References: 

- common\contextbuffer.go
- models\logevent.go

Create a new file that implements a per-instance batching buffer (unlike the singleton `contextbuffer.go`). This file should define:

1. **channelBuffer struct** with fields:
   - `buffer []models.LogEvent` - accumulates log events
   - `bufferMux sync.Mutex` - protects buffer access
   - `outputChan chan []models.LogEvent` - client's channel for receiving batches
   - `batchSize int` - number of events before auto-flush
   - `flushInterval time.Duration` - time between periodic flushes
   - `stopChan chan struct{}` - signals shutdown
   - `wg sync.WaitGroup` - tracks goroutine lifecycle

2. **NewChannelBuffer** constructor function that takes `(out chan []models.LogEvent, size int, interval time.Duration)` and returns `*channelBuffer`. Initialize all fields, create the buffer with capacity, and start the background goroutine via `go cb.run()`.

3. **Stop** method that closes `stopChan`, waits for the goroutine via `wg.Wait()`, and performs final flush.

4. **Log** method that accepts a `models.LogEvent`, adds it to the buffer under mutex protection, and triggers flush if `len(buffer) >= batchSize`.

5. **run** private method that:
   - Increments `wg` at start and defers `wg.Done()`
   - Creates a ticker with `flushInterval`
   - Loops selecting on ticker and stopChan
   - On ticker: locks mutex, calls flush, unlocks
   - On stop: locks mutex, calls flush, unlocks, returns

6. **flush** private method (called under mutex) that:
   - Checks if buffer has events
   - Creates a copy of the buffer slice
   - Sends the copy to `outputChan` in a goroutine with timeout (similar to `contextbuffer.go`)
   - Clears the buffer

This design mirrors `common/contextbuffer.go` but as an instantiable type rather than singleton, allowing multiple independent channel buffers.

### logger.go(MODIFY)

References: 

- common\channelbuffer.go(NEW)
- writers\channelwriter.go
- registry.go
- writers\contextwriter.go
- models\writerconfiguration.go
- levels\levels.go

Implement the two new methods in the `logger` struct (after the existing `SetContextChannelWithBuffer` method around line 41):

1. **SetChannel method** (signature: `func (l *logger) SetChannel(name string, ch chan []models.LogEvent)`):
   - Call `SetChannelWithBuffer` with default parameters: `l.SetChannelWithBuffer(name, ch, 5, 1*time.Second)`
   - This provides a convenient method with sensible defaults matching the existing `SetContextChannel` behavior

2. **SetChannelWithBuffer method** (signature: `func (l *logger) SetChannelWithBuffer(name string, ch chan []models.LogEvent, batchSize int, flushInterval time.Duration)`):
   - Create internal logger for error reporting: `internalLog := common.NewLogger().WithContext("function", "Logger.SetChannelWithBuffer").GetLogger()`
   - Create a new channel buffer instance: `channelBuf := common.NewChannelBuffer(ch, batchSize, flushInterval)`
   - Define a processor closure that adds events to the buffer: `processor := func(entry models.LogEvent) error { channelBuf.Log(entry); return nil }`
   - Create a minimal `WriterConfiguration` with trace level: `config := models.WriterConfiguration{Level: levels.TraceLevel}`
   - Create and start the channel writer: `writer, err := writers.NewChannelWriter(config, 1000, processor)` followed by `err = writer.Start()`
   - Handle errors appropriately (log fatal and panic if creation/start fails, similar to `NewContextWriter` in `writers/contextwriter.go`)
   - Register the writer in the global registry: `RegisterWriter(name, writer)`
   - Log success: `internalLog.Trace().Msgf("Channel writer '%s' registered successfully", name)`

The implementation follows the same pattern as `WithFileWriter`, `WithConsoleWriter`, and `WithMemoryWriter` - create a writer, configure it, and register it with a name. The key difference is that each named channel gets its own `channelBuffer` instance for independent batching.

**Important**: Store a reference to the `channelBuffer` instance if cleanup is needed later, though the current requirements don't specify a cleanup API. Consider adding a comment noting that cleanup/unregistration should be handled by the client via `UnregisterWriter(name)` if needed.