I have created the following plan after thorough exploration and analysis of the codebase. Follow the below plan verbatim. Trust the files and references. Do not re-verify what's written in the plan. Explore only when absolutely necessary. First implement all the proposed file changes and then I'll review all the changes together at the end.

### Observations

The websocketWriter in `writers/websocketwriter.go` implements a fundamentally different pattern from goroutineWriter: it polls an ILogStore on a ticker (500ms default), retrieves batches of logs via GetSince(), and broadcasts them to multiple WebSocket clients in parallel goroutines. The Write() method is a no-op (lines 157-159). In contrast, goroutineWriter is designed for write-driven workflows where Write() is the primary entry point, entries are buffered in a channel, and a processor function handles each entry. The patterns are architecturally incompatible: websocketWriter has outbound data flow (poll → retrieve → broadcast) while goroutineWriter has inbound data flow (write → buffer → process). Forcing composition would add unused Write() buffering overhead, require workarounds for ticker-based polling, and obscure the actual polling behavior without providing meaningful code reuse.

### Approach

After thorough architectural analysis, the websocketWriter should **NOT** be refactored to use the goroutineWriter base due to fundamental pattern incompatibility. The websocketWriter implements a poll-driven, broadcast-oriented pattern (retrieves logs from store and pushes to clients), while goroutineWriter implements a write-driven, buffering pattern (receives logs and processes them). However, minor standalone improvements can enhance the websocketWriter's robustness without forcing incompatible composition.

### Reasoning

I listed the repository structure, read the websocketwriter.go implementation, examined the goroutineWriter base class, reviewed the ILogStore interface, analyzed the refactored logStoreWriter and contextWriter patterns, and compared the architectural differences between write-driven and poll-driven patterns to assess compatibility.

## Proposed File Changes

### writers\websocketwriter.go(MODIFY)

References: 

- writers\goroutinewriter.go
- writers\ilogstore.go
- writers\logstorewriter.go
- writers\contextwriter.go

**EVALUATION RESULT: NO REFACTORING TO USE goroutineWriter BASE REQUIRED**

After comprehensive architectural analysis, the websocketWriter should remain as-is and NOT be refactored to use the goroutineWriter base. The patterns are fundamentally incompatible.

**Architectural Incompatibility Analysis:**

1. **Data Flow Direction Mismatch:**
   - goroutineWriter: Inbound (receives writes via Write() → buffers → processes)
   - websocketWriter: Outbound (polls store → retrieves → broadcasts)
   - The websocketWriter's Write() method is a no-op (lines 157-159) and would not utilize goroutineWriter's core buffering functionality

2. **Triggering Mechanism Mismatch:**
   - goroutineWriter: Event-driven (processes when entries arrive in buffer channel)
   - websocketWriter: Time-driven (polls on ticker interval, line 88)
   - The goroutineWriter's processBuffer() loop waits for channel events, but websocketWriter needs ticker-based polling which is incompatible

3. **Processing Model Mismatch:**
   - goroutineWriter processor: func(models.LogEvent) error - processes single entries
   - websocketWriter broadcast: Retrieves []models.LogEvent batches, broadcasts to multiple clients in parallel goroutines (lines 133-141), handles client failures and removal
   - The processor function signature doesn't support batch operations or client management

4. **Purpose Mismatch:**
   - goroutineWriter: Async write buffering to prevent blocking the logging path
   - websocketWriter: Real-time log streaming to connected WebSocket clients
   - These are orthogonal concerns that should remain separate

**Current Implementation Quality Assessment:**

Strengths:
- Clean separation of concerns (polling in pollAndBroadcast, broadcasting in broadcastNewLogs, client management in AddClient/RemoveClient)
- Proper concurrency handling with RWMutex for clients map (lines 120-126)
- Parallel client sends in goroutines (lines 133-141) with automatic failure handling
- Graceful shutdown with stopPoll channel (line 95) and client cleanup (lines 177-182)
- Configurable poll interval with sensible default (lines 40-42)
- Query methods for clients (GetLogsSince, GetLogsByCorrelation)

**OPTIONAL STANDALONE IMPROVEMENTS** (if desired, independent of goroutineWriter):

These improvements enhance robustness without changing the core architecture. Implement based on actual requirements:

**Optional Improvement 1: Add WaitGroup for Goroutine Synchronization (Recommended)**
- **Issue:** Close() signals stopPoll (line 174) but doesn't wait for pollAndBroadcast goroutine to exit, potentially leaving goroutine running briefly after Close() returns
- **Location:** Struct definition (line 26), constructor (line 54), pollAndBroadcast (line 85), Close method (line 173)
- **Changes needed:**
  - Add wg sync.WaitGroup field to websocketWriter struct after line 33
  - In WebSocketWriter constructor after line 51 (before go wsw.pollAndBroadcast()), add wsw.wg.Add(1)
  - In pollAndBroadcast method at line 86 (after internalLog creation), add defer wsw.wg.Done() for proper cleanup
  - In Close method after line 174 (after close(wsw.stopPoll)), add wsw.wg.Wait() to ensure goroutine completes before returning
- **Benefit:** Ensures goroutine lifecycle is properly synchronized with Close() method, preventing potential resource leaks
- **Priority:** Medium - improves shutdown guarantees

**Optional Improvement 2: Add Mutex Protection for lastSent Field**
- **Issue:** lastSent is written in broadcastNewLogs (line 117) and read in store.GetSince call (line 107), both in same goroutine so no actual race currently, but if GetLogsSince() method (line 147) is called concurrently from external goroutines, there could be a race condition
- **Location:** Struct definition (line 31), broadcastNewLogs (line 117), pollAndBroadcast (line 107), GetLogsSince (line 148)
- **Changes needed:**
  - Add lastSentMux sync.RWMutex field to websocketWriter struct after line 33
  - In broadcastNewLogs, wrap wsw.lastSent = time.Now() (line 117) with lastSentMux.Lock() and Unlock()
  - In broadcastNewLogs, wrap reading wsw.lastSent for store.GetSince call (line 107) with lastSentMux.RLock() and RUnlock()
  - In GetLogsSince method (line 148), if implementation reads wsw.lastSent, wrap with RLock/RUnlock
- **Benefit:** Prevents potential race conditions if GetLogsSince is called from multiple goroutines
- **Priority:** Low - only needed if GetLogsSince is called concurrently from external code

**Optional Improvement 3: Track Client Send Goroutines**
- **Issue:** Spawns goroutines for each client send (lines 133-141) without tracking completion, Close() may return before all sends complete
- **Location:** broadcastNewLogs method (lines 133-141)
- **Changes needed:**
  - Create local var sendWg sync.WaitGroup before the for loop at line 133
  - Inside the loop before launching each goroutine, add sendWg.Add(1)
  - Inside each goroutine function (line 134), add defer sendWg.Done() at the start
  - After the loop at line 141, optionally add sendWg.Wait() if you need to ensure all sends complete before returning (note: this may impact performance by blocking until all clients receive logs)
- **Benefit:** Ensures all client sends complete before method returns, useful for testing and graceful shutdown
- **Priority:** Low - only needed if you require guaranteed send completion before proceeding

**Optional Improvement 4: Add Architectural Documentation Comment**
- **Location:** Top of websocketWriter struct definition (before line 25)
- **Changes needed:**
  - Add comprehensive comment explaining the polling-based architecture and why it doesn't use goroutineWriter base
  - Example text: "websocketWriter implements a polling-based pattern that is architecturally incompatible with the write-driven goroutineWriter base. It polls an ILogStore on a timer and broadcasts to WebSocket clients, rather than receiving and buffering writes. The Write() method is intentionally a no-op. This is a deliberate design choice to support real-time log streaming with a pull-based model."
- **Benefit:** Documents architectural decision for future maintainers, prevents confusion about why this writer doesn't follow the goroutineWriter pattern
- **Priority:** Medium - improves code documentation and maintainability

**FINAL RECOMMENDATION:**

**Primary Action:** Mark this phase as COMPLETE with NO REFACTORING NEEDED. Document the evaluation result and architectural incompatibility.

**Optional Actions:** Implement Improvement 1 (WaitGroup) for better shutdown synchronization. The other improvements are optional and should be evaluated based on actual concurrency requirements.

**Rationale for No Refactoring:**
- The websocketWriter is well-implemented for its polling-based, broadcast-oriented purpose
- Forcing composition with goroutineWriter would increase complexity without providing meaningful benefits
- The current implementation has no code duplication with goroutineWriter (different patterns)
- Maintaining separate implementations preserves clarity and adheres to Single Responsibility Principle
- The goroutineWriter base successfully serves write-based writers (logStoreWriter, contextWriter); websocketWriter's exclusion validates appropriate abstraction scope