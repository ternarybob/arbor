I have created the following plan after thorough exploration and analysis of the codebase. Follow the below plan verbatim. Trust the files and references. Do not re-verify what's written in the plan. Explore only when absolutely necessary. First implement all the proposed file changes and then I'll review all the changes together at the end.

### Observations

The channelwriter_test.go file (1179 lines) already has comprehensive tests with updated naming - no changes needed there. The logger_contextlogger_integration_test.go tests the context logger but was written when ContextWriter used IChannelWriter composition - it should still work with the simplified implementation since the external behavior is the same. The logger_test.go file has basic unit tests but lacks integration tests for the new SetChannel/SetChannelWithBuffer methods. The existing tests show patterns for: creating test log events (createTestLogEvent helper), using atomic counters for verification, proper cleanup with defer statements, and managing global registry state. The key challenge is ensuring tests don't interfere with each other due to global state (registry, channelBuffers map, singleton contextbuffer).

### Approach

The task involves updating test files to reflect the refactored architecture. The channelwriter_test.go already has comprehensive tests with updated naming. The main work is: (1) validating that the simplified ContextWriter (which now directly calls common.Log() instead of wrapping a ChannelWriter) still works correctly in the existing integration test, (2) adding new integration tests for SetChannel and SetChannelWithBuffer methods in logger_test.go, and (3) ensuring proper cleanup of global state between tests. The tests will follow existing patterns from channelwriter_test.go and logger_memorywriter_integration_test.go.

### Reasoning

I explored the codebase structure, read the existing test files (channelwriter_test.go, logger_contextlogger_integration_test.go, logger_test.go, logger_memorywriter_integration_test.go), examined the simplified ContextWriter implementation, reviewed the ChannelBuffer and contextbuffer implementations, checked the registry system for writer management, and identified helper functions and patterns used in existing tests. I also searched for cleanup patterns (UnregisterWriter, common.Stop, defer Close) to understand how tests manage global state.

## Mermaid Diagram

sequenceDiagram
    participant Test as Test Suite
    participant Logger as Logger
    participant Registry as Writer Registry
    participant ChanWriter as ChannelWriter
    participant ChanBuffer as ChannelBuffer
    participant TestChan as Test Channel

    Note over Test: TestLogger_SetChannel_Basic
    Test->>Logger: SetChannel("test-channel", logChan)
    Logger->>ChanBuffer: NewChannelBuffer(logChan, 5, 1s)
    ChanBuffer-->>Logger: buffer instance
    Logger->>ChanWriter: NewChannelWriter(config, processor)
    ChanWriter-->>Logger: writer instance
    Logger->>ChanWriter: Start()
    Logger->>Registry: RegisterWriter("test-channel", writer)
    
    Test->>Logger: Info().Msg("message 1")
    Logger->>ChanWriter: Write(jsonBytes)
    ChanWriter->>ChanBuffer: processor(logEvent)
    ChanBuffer->>ChanBuffer: Accumulate in buffer
    
    Note over Test: Log 5 messages (batch size)
    
    ChanBuffer->>TestChan: Send batch []LogEvent
    Test->>TestChan: Receive batch
    Test->>Test: Assert batch size = 5
    Test->>Test: Verify log contents
    
    Test->>Logger: UnregisterChannel("test-channel")
    Logger->>ChanBuffer: Stop()
    Logger->>ChanWriter: Close()
    Logger->>Registry: UnregisterWriter("test-channel")
    
    Note over Test: TestLogger_SetChannel_MultipleChannels
    Test->>Logger: SetChannel("channel-1", chan1)
    Test->>Logger: SetChannel("channel-2", chan2)
    Note over Test: Both channels receive all logs independently
    Test->>Logger: Log messages
    Test->>TestChan: Receive from chan1
    Test->>TestChan: Receive from chan2
    Test->>Test: Verify both received batches

## Proposed File Changes

### logger_contextlogger_integration_test.go(MODIFY)

References: 

- writers\contextwriter.go
- common\contextbuffer.go

Update the test to validate the simplified ContextWriter implementation:

1. **Add comment clarification** (after line 15): Add a comment explaining that this test validates the simplified ContextWriter which now directly calls `common.Log()` instead of wrapping a ChannelWriter. The external behavior remains the same - logs are batched and sent to the channel.

2. **Keep existing test structure**: The test already validates the correct behavior:
   - Creates a channel for receiving log batches
   - Configures context logger with `SetContextChannelWithBuffer`
   - Creates a context logger with `WithContextWriter`
   - Logs messages that trigger batch flushes
   - Verifies batches are received with correct correlation IDs
   - All assertions remain valid for the simplified implementation

3. **Add validation comment** (before line 60): Add a comment noting that the test validates that the simplified ContextWriter (which directly calls `common.Log()`) correctly integrates with the singleton context buffer to deliver batched logs to the channel.

No functional changes are needed - the test already validates the correct behavior. The simplified ContextWriter maintains the same external contract (implements IWriter, sends logs to context buffer), so existing assertions remain valid.

### logger_test.go(MODIFY)

References: 

- logger.go
- ilogger.go
- common\channelbuffer.go
- registry.go
- writers\channelwriter_test.go

Add comprehensive integration tests for the new SetChannel and SetChannelWithBuffer methods at the end of the file (after line 400):

1. **TestLogger_SetChannel_Basic** - Test basic SetChannel functionality:
   - Create a buffered channel for receiving log batches: `logChan := make(chan []models.LogEvent, 10)`
   - Register channel with default settings: `Logger().SetChannel("test-channel", logChan)`
   - Defer cleanup: `defer Logger().UnregisterChannel("test-channel")`
   - Verify writer is registered: `assert.NotNil(GetRegisteredWriter("test-channel"))`
   - Create logger and log 5 messages (default batch size) to trigger flush
   - Use goroutine with timeout to receive batch from channel
   - Assert batch contains expected number of messages
   - Verify log event fields (message, level, etc.)

2. **TestLogger_SetChannelWithBuffer_CustomBatching** - Test custom batching parameters:
   - Create channel: `logChan := make(chan []models.LogEvent, 10)`
   - Register with custom settings: `Logger().SetChannelWithBuffer("custom-channel", logChan, 3, 50*time.Millisecond)`
   - Defer cleanup: `defer Logger().UnregisterChannel("custom-channel")`
   - Log 3 messages to trigger batch flush (custom batch size)
   - Verify batch is received within expected timeframe
   - Log 2 more messages and wait for timer-based flush (50ms)
   - Verify second batch is received

3. **TestLogger_SetChannel_MultipleChannels** - Test multiple independent channels:
   - Create two channels: `chan1`, `chan2`
   - Register both: `SetChannel("channel-1", chan1)` and `SetChannel("channel-2", chan2)`
   - Defer cleanup for both
   - Create two loggers with different correlation IDs
   - Log to both loggers
   - Verify both channels receive their respective batches
   - Verify batches are independent (each channel gets all logs, not split)

4. **TestLogger_SetChannel_WithCorrelationID** - Test channel logging with correlation IDs:
   - Create channel and register: `SetChannel("corr-channel", logChan)`
   - Defer cleanup
   - Create logger with correlation ID: `logger := Logger().WithCorrelationId("test-job-123")`
   - Log multiple messages
   - Receive batch from channel
   - Verify all log events have the correct correlation ID

5. **TestLogger_SetChannel_LevelFiltering** - Test that channel writers respect log levels:
   - Create channel and register with INFO level: `SetChannelWithBuffer("level-channel", logChan, 10, 100*time.Millisecond)`
   - Defer cleanup
   - Log messages at different levels (Debug, Info, Warn, Error)
   - Wait for timer flush
   - Verify only Info and above are in the batch (Debug filtered out)

6. **TestLogger_UnregisterChannel** - Test cleanup functionality:
   - Create channel and register: `SetChannel("temp-channel", logChan)`
   - Verify writer is registered
   - Call `UnregisterChannel("temp-channel")`
   - Verify writer is no longer registered: `assert.Nil(GetRegisteredWriter("temp-channel"))`
   - Verify logging after unregister doesn't panic

7. **TestLogger_SetChannel_ReplaceExisting** - Test replacing an existing channel:
   - Create first channel and register: `SetChannel("replace-channel", chan1)`
   - Verify registered
   - Create second channel and register with same name: `SetChannel("replace-channel", chan2)`
   - Defer cleanup
   - Log messages
   - Verify only chan2 receives logs (chan1 should not)

8. **TestLogger_SetChannel_NilChannel** - Test error handling for nil channel:
   - Use defer/recover to catch panic: `defer func() { if r := recover(); r == nil { t.Error("Expected panic") } }()`
   - Call `SetChannel("nil-channel", nil)`
   - Verify panic occurs with appropriate message

9. **TestLogger_SetChannelWithBuffer_InvalidParameters** - Test parameter validation:
   - Create channel
   - Test with zero batch size: `SetChannelWithBuffer("zero-batch", logChan, 0, 1*time.Second)` - should use default (5)
   - Test with negative batch size: should use default
   - Test with zero interval: should use default (1 second)
   - Test with negative interval: should use default
   - Defer cleanup for all
   - Verify all channels work correctly with defaults

10. **TestLogger_SetChannel_ConcurrentWrites** - Test concurrent logging to channel:
   - Create channel and register: `SetChannel("concurrent-channel", logChan)`
   - Defer cleanup
   - Use WaitGroup to coordinate goroutines
   - Launch 10 goroutines, each logging 5 messages
   - Collect all batches from channel
   - Verify total count is 50 messages
   - Verify no messages are lost or duplicated

All tests should follow patterns from existing tests:
- Use `require` for fatal assertions, `assert` for non-fatal
- Import `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require`
- Use proper cleanup with `defer`
- Use timeouts when receiving from channels to prevent test hangs
- Create unique channel names to avoid conflicts between tests
- Use helper pattern for receiving batches with timeout