I have created the following plan after thorough exploration and analysis of the codebase. Follow the below plan verbatim. Trust the files and references. Do not re-verify what's written in the plan. Explore only when absolutely necessary. First implement all the proposed file changes and then I'll review all the changes together at the end.

### Observations

The goroutineWriter base implementation in `writers/goroutinewriter.go` provides core async buffering infrastructure with lifecycle management (Start, Stop, IsRunning), buffered channel processing, processor functions, and graceful shutdown. The refactored `logStoreWriter` and `contextWriter` both compose this base successfully. Existing test patterns in `writers/memorywriter_test.go` show async processing with time.Sleep() for goroutine synchronization, integration tests combining store and writer, and correlation-based log retrieval. The `consolewriter_test.go` demonstrates table-driven tests and interface verification. The test suite needs to cover: (1) goroutineWriter base unit tests for lifecycle, buffering, overflow, shutdown, concurrency; (2) integration tests validating logStoreWriter and contextWriter behavior; (3) backward compatibility verification ensuring existing patterns still work.

### Approach

Create comprehensive test coverage in `writers/goroutinewriter_test.go` with three test categories: (1) Unit tests for goroutineWriter base covering lifecycle methods, buffer overflow, graceful shutdown, concurrent writes, and error handling; (2) Integration tests with refactored logStoreWriter and contextWriter to validate composition pattern; (3) Backward compatibility tests ensuring existing writer behavior is preserved. Use table-driven tests where appropriate, mock processor functions to verify behavior, and leverage time.Sleep() for async synchronization following established patterns in `memorywriter_test.go`.

### Reasoning

I listed the repository structure, read the goroutineWriter base implementation, examined refactored logStoreWriter and contextWriter, reviewed existing test patterns in memorywriter_test.go and consolewriter_test.go, and analyzed the interfaces (IWriter, IGoroutineWriter) and models (LogEvent, WriterConfiguration) to understand testing requirements and patterns.

## Proposed File Changes

### writers\goroutinewriter_test.go(NEW)

References: 

- writers\goroutinewriter.go
- writers\igoroutinewriter.go
- writers\iwriter.go
- writers\logstorewriter.go
- writers\contextwriter.go
- writers\memorywriter_test.go
- writers\consolewriter_test.go
- models\logevent.go
- models\writerconfiguration.go
- levels\levels.go
- writers\inmemorylogstore.go

Create comprehensive test suite for the goroutineWriter base implementation with unit tests, integration tests, and backward compatibility verification.

**Package and Imports:**
- Package: `writers`
- Required imports:
  - `encoding/json` - for marshaling test LogEvent data
  - `errors` - for creating test errors
  - `sync` - for sync.Mutex, sync.WaitGroup in concurrent tests
  - `sync/atomic` - for atomic counters in processor functions
  - `testing` - for test framework
  - `time` - for time.Sleep() async synchronization and timestamps
  - `github.com/phuslu/log` - for log.Level constants
  - `github.com/ternarybob/arbor/levels` - for levels.LogLevel and levels.TraceLevel
  - `github.com/ternarybob/arbor/models` - for models.LogEvent and models.WriterConfiguration

**SECTION 1: UNIT TESTS FOR goroutineWriter BASE**

**Test 1: TestGoroutineWriter_NewWithValidProcessor**
- Purpose: Verify constructor creates writer successfully with valid processor
- Steps:
  - Create config with levels.TraceLevel
  - Define simple processor: `func(models.LogEvent) error { return nil }`
  - Call NewGoroutineWriter(config, 1000, processor)
  - Assert: writer is not nil, error is nil
  - Assert: writer implements IGoroutineWriter interface using type assertion
  - Assert: IsRunning() returns false (not started yet)
  - Cleanup: No cleanup needed (writer not started)

**Test 2: TestGoroutineWriter_NewWithNilProcessor**
- Purpose: Verify constructor returns error when processor is nil
- Steps:
  - Create config with levels.InfoLevel
  - Call NewGoroutineWriter(config, 1000, nil)
  - Assert: writer is nil
  - Assert: error is not nil
  - Assert: error message contains "processor" or "nil"

**Test 3: TestGoroutineWriter_NewWithInvalidBufferSize**
- Purpose: Verify constructor uses default buffer size (1000) when invalid size provided
- Steps:
  - Create config and valid processor
  - Call NewGoroutineWriter(config, 0, processor) with zero buffer size
  - Assert: writer is not nil, error is nil
  - Note: Cannot directly verify buffer size (private field), but document expected behavior
  - Call NewGoroutineWriter(config, -100, processor) with negative buffer size
  - Assert: writer is not nil, error is nil (should default to 1000)

**Test 4: TestGoroutineWriter_StartStop_Lifecycle**
- Purpose: Verify Start() and Stop() lifecycle methods work correctly
- Steps:
  - Create writer with valid processor
  - Assert: IsRunning() returns false initially
  - Call Start()
  - Assert: error is nil
  - Assert: IsRunning() returns true
  - Call Start() again (double-start)
  - Assert: error is not nil (already running)
  - Call Stop()
  - Assert: error is nil
  - Assert: IsRunning() returns false
  - Call Stop() again (double-stop)
  - Assert: error is nil (idempotent)

**Test 5: TestGoroutineWriter_StartStopRestart**
- Purpose: Verify writer can be restarted after Stop()
- Steps:
  - Create writer with valid processor
  - Call Start(), verify IsRunning() is true
  - Call Stop(), verify IsRunning() is false
  - Call Start() again (restart)
  - Assert: error is nil (should succeed)
  - Assert: IsRunning() returns true
  - Write a test log entry to verify goroutine is processing
  - Wait 50ms for processing
  - Call Stop() for cleanup

**Test 6: TestGoroutineWriter_Write_BeforeStart**
- Purpose: Verify Write() returns success but doesn't enqueue when not started
- Steps:
  - Create writer with processor that increments atomic counter
  - Do NOT call Start()
  - Create test LogEvent, marshal to JSON
  - Call Write(jsonData)
  - Assert: bytes written equals len(jsonData), error is nil
  - Wait 50ms
  - Assert: processor counter is 0 (not called because goroutine not running)

**Test 7: TestGoroutineWriter_Write_AfterStop**
- Purpose: Verify Write() returns success but doesn't enqueue after Stop()
- Steps:
  - Create writer with processor that increments atomic counter
  - Call Start()
  - Call Stop() immediately
  - Wait 50ms for shutdown to complete
  - Create test LogEvent, marshal to JSON
  - Call Write(jsonData)
  - Assert: bytes written equals len(jsonData), error is nil
  - Wait 50ms
  - Assert: processor counter is 0 (not called because goroutine stopped)

**Test 8: TestGoroutineWriter_Write_Success**
- Purpose: Verify Write() successfully processes log entries
- Steps:
  - Create atomic counter for processed entries
  - Create processor that increments counter and returns nil
  - Create writer, call Start()
  - Create test LogEvent with log.InfoLevel, marshal to JSON
  - Call Write(jsonData)
  - Assert: bytes written equals len(jsonData), error is nil
  - Wait 100ms for async processing
  - Assert: counter equals 1 (processor called once)
  - Cleanup: Call Stop()

**Test 9: TestGoroutineWriter_Write_InvalidJSON**
- Purpose: Verify Write() returns error for invalid JSON
- Steps:
  - Create writer with valid processor, call Start()
  - Call Write([]byte("invalid json {{"))
  - Assert: bytes written is 0, error is not nil
  - Assert: error message indicates JSON unmarshal failure
  - Cleanup: Call Stop()

**Test 10: TestGoroutineWriter_Write_EmptyData**
- Purpose: Verify Write() handles empty data gracefully
- Steps:
  - Create writer with processor, call Start()
  - Call Write([]byte("")) with empty slice
  - Assert: bytes written is 0, error is nil
  - Call Write(nil) with nil slice
  - Assert: bytes written is 0, error is nil
  - Cleanup: Call Stop()

**Test 11: TestGoroutineWriter_LevelFiltering**
- Purpose: Verify level filtering prevents low-level logs from being processed
- Steps:
  - Create atomic counter for processed entries
  - Create processor that increments counter
  - Create config with levels.WarnLevel (only Warn and above)
  - Create writer, call Start()
  - Write 4 log entries: DebugLevel, InfoLevel, WarnLevel, ErrorLevel (marshal each to JSON)
  - Wait 100ms for processing
  - Assert: counter equals 2 (only Warn and Error processed)
  - Cleanup: Call Stop()

**Test 12: TestGoroutineWriter_WithLevel_DynamicChange**
- Purpose: Verify WithLevel() dynamically changes filtering
- Steps:
  - Create atomic counter
  - Create processor that increments counter
  - Create config with levels.InfoLevel
  - Create writer, call Start()
  - Write DebugLevel entry (should be filtered)
  - Wait 50ms
  - Assert: counter is 0
  - Call writer.WithLevel(log.DebugLevel) to lower threshold
  - Write another DebugLevel entry (should now be processed)
  - Wait 50ms
  - Assert: counter is 1
  - Cleanup: Call Stop()

**Test 13: TestGoroutineWriter_BufferOverflow**
- Purpose: Verify buffer overflow handling (drops entries and logs warning)
- Steps:
  - Create processor that sleeps 100ms per entry (slow processor)
  - Create writer with small buffer size (e.g., 10), call Start()
  - Write 100 log entries rapidly in a loop (faster than processor can handle)
  - Wait for processing to complete (2 seconds)
  - Verify: Not all 100 entries were processed (some dropped due to buffer full)
  - Note: Cannot easily verify warning logs without capturing log output, but document expected behavior
  - Cleanup: Call Stop()

**Test 14: TestGoroutineWriter_GracefulShutdown_BufferDraining**
- Purpose: Verify Stop() drains all buffered entries before returning
- Steps:
  - Create atomic counter
  - Create processor that increments counter and sleeps 10ms
  - Create writer with buffer size 100, call Start()
  - Write 50 log entries rapidly
  - Immediately call Stop() (before all entries processed)
  - Assert: Stop() returns nil (waits for drain to complete)
  - Assert: counter equals 50 (all entries processed during shutdown)
  - Verify: IsRunning() returns false after Stop()

**Test 15: TestGoroutineWriter_ConcurrentWrites**
- Purpose: Verify thread-safety of concurrent Write() calls
- Steps:
  - Create atomic counter
  - Create processor that increments counter
  - Create writer, call Start()
  - Launch 10 goroutines, each writing 10 log entries (100 total)
  - Use sync.WaitGroup to wait for all goroutines to complete writes
  - Wait 500ms for processing
  - Assert: counter equals 100 (all entries processed)
  - Cleanup: Call Stop()

**Test 16: TestGoroutineWriter_ProcessorError**
- Purpose: Verify processor errors are logged but don't crash goroutine
- Steps:
  - Create processor that returns error for every entry: `return errors.New("processor error")`
  - Create writer, call Start()
  - Write 5 log entries
  - Wait 100ms for processing
  - Verify: goroutine still running (IsRunning() returns true)
  - Verify: Can write more entries successfully
  - Cleanup: Call Stop()
  - Note: Cannot easily verify warning logs without capturing output

**Test 17: TestGoroutineWriter_Close_Idempotent**
- Purpose: Verify Close() can be called multiple times safely
- Steps:
  - Create writer with valid processor, call Start()
  - Call Close()
  - Assert: error is nil
  - Call Close() again
  - Assert: error is nil (idempotent)
  - Call Close() third time
  - Assert: error is nil (still idempotent)

**Test 18: TestGoroutineWriter_GetFilePath**
- Purpose: Verify GetFilePath() returns empty string
- Steps:
  - Create writer with valid processor
  - Call GetFilePath()
  - Assert: returns empty string ""

**Test 19: TestGoroutineWriter_MultipleEntries_OrderPreserved**
- Purpose: Verify entries are processed in FIFO order
- Steps:
  - Create slice to collect processed entries (protected by mutex)
  - Create processor that appends entry.Message to slice
  - Create writer, call Start()
  - Write 10 entries with messages "msg-0", "msg-1", ..., "msg-9"
  - Wait 200ms for processing
  - Assert: slice contains 10 entries
  - Assert: entries are in order ("msg-0", "msg-1", ..., "msg-9")
  - Cleanup: Call Stop()

**SECTION 2: INTEGRATION TESTS WITH REFACTORED WRITERS**

**Test 20: TestLogStoreWriter_Integration_WithGoroutineWriter**
- Purpose: Verify logStoreWriter correctly composes goroutineWriter and stores entries
- Steps:
  - Create in-memory log store using NewInMemoryLogStore(config) from `writers/inmemorylogstore.go`
  - Create config with levels.TraceLevel
  - Create logStoreWriter using LogStoreWriter(store, config)
  - Verify: writer is not nil
  - Create test LogEvent with correlation ID "integration-test"
  - Marshal to JSON and call Write()
  - Wait 100ms for async processing
  - Query store using store.GetByCorrelation("integration-test") method from ILogStore interface
  - Assert: retrieved 1 entry
  - Assert: entry message matches original
  - Cleanup: Call writer.Close(), store.Close()

**Test 21: TestLogStoreWriter_Integration_LevelFiltering**
- Purpose: Verify logStoreWriter respects level filtering from goroutineWriter
- Steps:
  - Create in-memory store
  - Create config with levels.WarnLevel (only Warn and above)
  - Create logStoreWriter
  - Write 4 entries: DebugLevel, InfoLevel, WarnLevel, ErrorLevel
  - Wait 100ms
  - Query all entries from store using GetAll() or similar method
  - Assert: store contains only 2 entries (Warn and Error)
  - Cleanup: Close writer and store

**Test 22: TestLogStoreWriter_Integration_GracefulShutdown**
- Purpose: Verify logStoreWriter drains buffer on Close()
- Steps:
  - Create in-memory store
  - Create logStoreWriter with TraceLevel
  - Write 20 log entries rapidly
  - Immediately call Close() (before processing completes)
  - Query store for all entries
  - Assert: store contains all 20 entries (buffer drained during shutdown)
  - Cleanup: Close store

**Test 23: TestContextWriter_Integration_WithGoroutineWriter**
- Purpose: Verify contextWriter correctly composes goroutineWriter and calls common.Log()
- Steps:
  - Create config with levels.TraceLevel
  - Create contextWriter using NewContextWriter(config)
  - Verify: writer is not nil
  - Create test LogEvent with correlation ID "context-test"
  - Marshal to JSON and call Write()
  - Wait 100ms for async processing
  - Note: Cannot easily verify common.Log() was called without accessing global context buffer from `common/contextbuffer.go`
  - Verify: Write() returns success (bytes written, nil error)
  - Cleanup: Call writer.Close()

**Test 24: TestContextWriter_Integration_AsyncBehavior**
- Purpose: Verify contextWriter provides async buffering before global context buffer
- Steps:
  - Create contextWriter with TraceLevel
  - Record start time
  - Write 10 log entries rapidly in a loop
  - Record end time
  - Verify: All Write() calls return immediately (non-blocking)
  - Assert: time taken for 10 writes is less than 10ms (proving async)
  - Wait 100ms for processing
  - Cleanup: Call Close()

**SECTION 3: BACKWARD COMPATIBILITY TESTS**

**Test 25: TestLogStoreWriter_BackwardCompatibility_ConstructorSignature**
- Purpose: Verify LogStoreWriter constructor signature unchanged
- Steps:
  - Create in-memory store
  - Create config
  - Call LogStoreWriter(store, config) - verify compiles and returns IWriter
  - Verify: returned writer implements IWriter interface
  - Verify: can call Write(), WithLevel(), GetFilePath(), Close() methods
  - Cleanup: Close writer and store

**Test 26: TestLogStoreWriter_BackwardCompatibility_AutoStart**
- Purpose: Verify logStoreWriter auto-starts goroutine in constructor (backward compatibility)
- Steps:
  - Create in-memory store
  - Create logStoreWriter using LogStoreWriter(store, config)
  - Immediately write log entry (no explicit Start() call)
  - Wait 100ms
  - Query store
  - Assert: entry was processed (proves goroutine auto-started)
  - Cleanup: Close writer and store

**Test 27: TestContextWriter_BackwardCompatibility_ConstructorSignature**
- Purpose: Verify NewContextWriter accepts WriterConfiguration parameter
- Steps:
  - Create config with levels.TraceLevel
  - Call NewContextWriter(config) - verify compiles and returns IWriter
  - Verify: returned writer implements IWriter interface
  - Verify: can call all IWriter methods
  - Cleanup: Call Close()

**SECTION 4: EDGE CASE AND CONCURRENCY TESTS**

**Test 28: TestGoroutineWriter_ConcurrentStartStop**
- Purpose: Verify thread-safety of concurrent Start()/Stop() calls
- Steps:
  - Create writer with valid processor
  - Launch 5 goroutines that call Start() concurrently
  - Launch 5 goroutines that call Stop() concurrently
  - Use WaitGroup to wait for all goroutines
  - Verify: No panics or deadlocks occur
  - Verify: Final state is consistent (either running or stopped)
  - Cleanup: Ensure Stop() is called

**Test 29: TestGoroutineWriter_ConcurrentWithLevel**
- Purpose: Verify thread-safety of WithLevel() during concurrent writes
- Steps:
  - Create atomic counter
  - Create processor that increments counter
  - Create writer with InfoLevel, call Start()
  - Launch goroutine that writes DebugLevel entries in a loop (100 entries)
  - Launch goroutine that calls WithLevel(log.DebugLevel) after 50ms
  - Wait for completion
  - Verify: Some DebugLevel entries were filtered (before WithLevel), some processed (after WithLevel)
  - Verify: No data races or panics
  - Cleanup: Call Stop()

**Test 30: TestGoroutineWriter_ProcessorError_ContinuesProcessing**
- Purpose: Verify processor errors don't stop the goroutine
- Steps:
  - Create atomic counter
  - Create processor that returns error for first 3 entries, then succeeds
  - Create writer, call Start()
  - Write 5 log entries
  - Wait 200ms for processing
  - Assert: counter equals 5 (all entries attempted, errors logged but processing continued)
  - Verify: IsRunning() still true
  - Cleanup: Call Stop()

**SECTION 5: HELPER FUNCTIONS**

**Helper: createTestLogEvent(level log.Level, correlationID string, message string) models.LogEvent**
- Purpose: Create test LogEvent with common defaults
- Returns: models.LogEvent with specified level, correlationID, message, and current timestamp
- Sets: Prefix to "TEST", Function to "test", empty Fields map

**Helper: marshalLogEvent(t *testing.T, event models.LogEvent) []byte**
- Purpose: Marshal LogEvent to JSON for Write() calls
- Returns: JSON byte slice
- Calls t.Fatal() on marshal error (test helper, should never fail)

**Helper: createCountingProcessor(counter *atomic.Int64) func(models.LogEvent) error**
- Purpose: Create processor that increments atomic counter
- Returns: Processor function that increments counter and returns nil
- Useful for verifying entries were processed

**Helper: createDelayedProcessor(counter *atomic.Int64, delay time.Duration) func(models.LogEvent) error**
- Purpose: Create processor that increments counter and sleeps
- Returns: Processor function that increments counter, sleeps for delay, returns nil
- Useful for testing buffer overflow and slow processor scenarios

**Helper: createCollectingProcessor(collected *[]models.LogEvent, mu *sync.Mutex) func(models.LogEvent) error**
- Purpose: Create processor that collects entries in a slice
- Returns: Processor function that appends entry to slice (protected by mutex) and returns nil
- Useful for verifying entry order and content

**TEST ORGANIZATION:**
- Group related tests using subtests where appropriate
- Use table-driven tests for similar test cases with different inputs (e.g., level filtering with multiple levels)
- Follow naming convention: Test<Component>_<Scenario> or Test<Component>_<Scenario>_<ExpectedBehavior>
- Add comments explaining the purpose and expected behavior of each test
- Use defer for cleanup (Close(), Stop()) to ensure resources are released even if test fails

**ASYNC SYNCHRONIZATION STRATEGY:**
- Use time.Sleep() for async processing synchronization (following pattern from `memorywriter_test.go` lines 49, 90, 130, etc.)
- Typical wait times:
  - 50ms for single entry processing
  - 100ms for multiple entries (< 10)
  - 200-500ms for larger batches or concurrent operations
  - 2 seconds for stress tests with slow processors
- Alternative: Use channels to signal completion from processor, but time.Sleep() is simpler and matches existing patterns

**ERROR HANDLING IN TESTS:**
- Use t.Fatal() for setup errors that prevent test from continuing
- Use t.Error() or t.Errorf() for assertion failures
- Use defer for cleanup to ensure resources released on failure
- Verify error messages contain expected keywords when testing error cases

**COVERAGE GOALS:**
- Lifecycle: Start, Stop, IsRunning, restart, double-start, double-stop
- Write operations: success, invalid JSON, empty data, before start, after stop
- Level filtering: static config, dynamic WithLevel, multiple levels
- Buffer management: overflow, draining on shutdown, FIFO order
- Concurrency: concurrent writes, concurrent Start/Stop, concurrent WithLevel
- Error handling: processor errors, invalid input, edge cases
- Integration: logStoreWriter composition, contextWriter composition
- Backward compatibility: constructor signatures, auto-start behavior

**TESTING BEST PRACTICES:**
- Each test should be independent (no shared state between tests)
- Use descriptive test names that explain what is being tested
- Include comments explaining non-obvious behavior or timing requirements
- Clean up resources (Close, Stop) using defer
- Verify both positive cases (success) and negative cases (errors)
- Test boundary conditions (empty data, nil, zero buffer size)
- Test concurrent access to verify thread-safety
- Follow existing test patterns from `memorywriter_test.go` and `consolewriter_test.go`