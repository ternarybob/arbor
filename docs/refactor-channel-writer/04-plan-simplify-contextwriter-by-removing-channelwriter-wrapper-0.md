I have created the following plan after thorough exploration and analysis of the codebase. Follow the below plan verbatim. Trust the files and references. Do not re-verify what's written in the plan. Explore only when absolutely necessary. First implement all the proposed file changes and then I'll review all the changes together at the end.

### Observations

The WebSocket writer (`writers/websocketwriter.go`) is a 258-line file implementing a poll-based pattern that retrieves logs from an `ILogStore` and broadcasts them to connected clients. It includes three types: `IWebSocketWriter` interface, `WebSocketClient` interface, and `SimpleWebSocketClient` test helper. The only test is `TestWebSocketWriter_Basic` in `memorywriter_test.go` (lines 279-350). Documentation references appear in 6 files: README.md (multiple sections), WARP.md (line 103), and 4 refactor documentation files in `docs/`. The user's intent is to simplify the library by removing this specialized writer and encouraging clients to implement WebSocket functionality using the more flexible `SetChannel`/`SetChannelWithBuffer` API that was just added.

### Approach

This is a cleanup task to remove the WebSocket writer implementation and update documentation to reflect that WebSocket logging should be implemented client-side using the new ChannelWriter API. The removal involves: (1) deleting the entire `websocketwriter.go` file, (2) removing the WebSocket test from `memorywriter_test.go`, and (3) updating all documentation files to remove WebSocket references and add guidance on client-side implementation using `SetChannel`/`SetChannelWithBuffer` methods.

### Reasoning

I explored the repository structure, read the WebSocket writer implementation and its test, performed grep searches to identify all references to WebSocket-related code across Go files and documentation, and reviewed the README and WARP documentation to understand the current architecture descriptions that need updating.

## Proposed File Changes

### writers\websocketwriter.go(DELETE)

Delete the entire file. This file contains:

1. **WebSocketClient interface** (lines 19-23): Defines the contract for WebSocket client implementations with `SendJSON` and `Close` methods

2. **IWebSocketWriter interface** (lines 25-32): Extends `IWriter` with WebSocket-specific methods: `AddClient`, `RemoveClient`, `GetLogsSince`, `GetLogsByCorrelation`

3. **websocketWriter struct** (lines 39-50): Internal implementation with fields for store, config, clients map, polling interval, and synchronization primitives

4. **WebSocketWriter factory function** (lines 52-75): Creates and starts a WebSocket writer with configurable poll interval

5. **Client management methods** (lines 77-98): `AddClient` and `RemoveClient` for managing connected WebSocket clients

6. **Polling logic** (lines 100-184): `pollAndBroadcast` goroutine that polls the store on a timer and `broadcastNewLogs` that sends batches to all clients

7. **Query methods** (lines 186-194): `GetLogsSince` and `GetLogsByCorrelation` for manual log retrieval

8. **IWriter implementation** (lines 196-228): `Write` (no-op), `WithLevel`, `GetFilePath`, and `Close` methods

9. **SimpleWebSocketClient test helper** (lines 230-257): Basic implementation for testing with configurable send/close functions

This entire implementation is being removed in favor of client-side implementations using the new `SetChannel`/`SetChannelWithBuffer` API.

### writers\memorywriter_test.go(MODIFY)

Remove the `TestWebSocketWriter_Basic` test function (lines 279-350). This test:

1. **Setup** (lines 280-294): Creates an in-memory log store and WebSocket writer with 50ms poll interval for testing

2. **Client setup** (lines 296-317): Creates a test client using `NewSimpleWebSocketClient` with a custom `SendJSON` function that collects received logs

3. **Test execution** (lines 319-341): Adds the client, writes a log event to the store, and waits for the broadcast with a 200ms timeout

4. **Assertions** (lines 343-349): Verifies that logs were received via the WebSocket broadcast mechanism

Keep all other tests in the file unchanged:
- `TestMemoryWriter_Basic` (lines 19-60)
- `TestMemoryWriter_MultipleEntries` (lines 62-100)
- `TestMemoryWriter_LevelFiltering` (lines 102-142)
- `TestMemoryWriter_GetEntriesSince` (lines 144-201)
- `TestMemoryWriter_GetRecent` (lines 203-240)
- `TestMemoryWriter_WithPersistence` (lines 242-277)

After removal, the file will contain only memory writer tests, which is appropriate since the file is named `memorywriter_test.go`.

### README.md(MODIFY)

References: 

- writers\channelwriter.go
- logger.go

Update multiple sections to remove WebSocket writer references and add guidance on client-side implementation:

1. **Features section** (line 47): Remove the bullet point "WebSocket writer for real-time streaming to connected clients" from the Multi-Writer Architecture list

2. **Features section** (line 53): Remove the standalone bullet point "**WebSocket Support**: Real-time log streaming to connected clients"

3. **WebSocket Streaming Pattern section** (lines 523-545): Replace the entire section with a new section titled "### Real-Time Log Streaming with Channels" that explains:
   - How to use `SetChannel` or `SetChannelWithBuffer` to create a channel logger
   - How clients can receive batched log events on their channel
   - How to implement WebSocket broadcasting in client code by reading from the channel and sending to WebSocket clients
   - Example code showing:
     ```go
     // Create channel for receiving log batches
     logChannel := make(chan []models.LogEvent, 10)
     
     // Register channel with default batching (5 events, 1 second)
     arbor.Logger().SetChannel("websocket-logs", logChannel)
     
     // Or with custom batching for high-throughput scenarios
     arbor.Logger().SetChannelWithBuffer("websocket-logs", logChannel, 100, 5*time.Second)
     
     // Client implements WebSocket broadcasting
     go func() {
         for logBatch := range logChannel {
             // Broadcast to your WebSocket clients
             for _, client := range websocketClients {
                 client.SendJSON(logBatch)
             }
         }
     }()
     ```
   - Note that this approach gives clients full control over WebSocket connection management, client tracking, error handling, and broadcasting logic

4. **Performance Characteristics section** (line 557): Remove the bullet point "**WebSocket polling**: Retrieves batches every 500ms (configurable)"

5. **Log Store Architecture diagram** (lines 732-733): Remove the line "├──► WebSocket Writer (timestamp polling)" from the ASCII diagram

6. **Poll-Based Writer (WebSocket) subsection** (lines 835-862): Remove the entire subsection including the heading, description, and example code. This section describes the old WebSocket writer pattern that is being removed.

7. **Architecture description** (line 39): Update the comment in the diagram from "WebSocket writer for real-time streaming to connected clients" to reflect that clients can implement this using channel writers

The goal is to remove all references to the built-in WebSocket writer while providing clear guidance on how clients can achieve the same functionality using the more flexible channel-based API.

### WARP.md(MODIFY)

Update the writers list to remove WebSocket writer reference:

1. **Writers list** (line 103): Change the line "Console, File, Memory, LogStore, Context, WebSocket writers" to "Console, File, Memory, LogStore, Context writers"

This is a simple one-line change to remove the WebSocket writer from the list of available writer implementations. The rest of the WARP.md file remains unchanged as it doesn't contain detailed WebSocket documentation.

### docs\refactor-goroutine-writer\04-plan-evaluate-and-optionally-refactor-websocketwriter-0.md(DELETE)

This document is entirely about evaluating whether to refactor the WebSocket writer to use the goroutineWriter (now channelWriter) base. Since the WebSocket writer is being removed entirely, this document is no longer relevant.

**Options:**
1. **Delete the file entirely** - Cleanest approach since the evaluation is moot
2. **Add a deprecation notice** - Keep the file with a header explaining the WebSocket writer was removed in favor of client-side implementation

**Recommendation**: Delete the file entirely. The document served its purpose (concluding that WebSocket writer should NOT use the base pattern), and now that the WebSocket writer is removed, the evaluation is no longer relevant. The refactor documentation directory should only contain documents about active refactorings.

If you prefer to keep historical context, add a header:
```markdown
# [DEPRECATED] Evaluate and Optionally Refactor WebSocketWriter

**Status**: This document is deprecated. The WebSocketWriter was removed from the library in favor of client-side implementations using `SetChannel`/`SetChannelWithBuffer` methods.

**Rationale**: The WebSocket writer's poll-based pattern was architecturally incompatible with the channelWriter base, and the functionality is better implemented client-side where developers have full control over WebSocket connection management, broadcasting logic, and error handling.

---

[Original content follows...]
```

### docs\refactor-goroutine-writer\06-plan-review-and-validate-goroutinewriter-refactoring-0.md(MODIFY)

Update all references to WebSocket writer in this comprehensive review document:

1. **Observations section** (line 5): Remove the phrase "poll-based writer (websocket) for real-time streaming," from the list of writer patterns. Update to: "All writers follow appropriate patterns: synchronous writers (console, file) for fast I/O, async writers (logStore, context) using channelWriter base for non-blocking buffered processing, and query-only writer (memory) for log retrieval."

2. **References section** (line 24): Remove the line "- writers\websocketwriter.go" from the file list

3. **Multi-Writer Architecture bullet** (line 39): Remove the line "- WebSocket writer for real-time streaming to connected clients"

4. **Poll-Based Writer (WebSocket) subsection** (lines 79-86): Remove the entire subsection including:
   - Pattern description
   - Performance characteristics
   - Use cases
   - Code example
   - Note about pull-based vs push-based model

5. **Performance section** (line 110): Remove the bullet point "**WebSocket polling**: Retrieves batches every 500ms (configurable)"

6. **Rationale** (line 113): Update the rationale to remove the phrase "and adds WebSocket polling information"

This document is a comprehensive review of the goroutineWriter (now channelWriter) refactoring. Since the WebSocket writer is being removed, all references to it should be cleaned up to reflect the current architecture.

### docs\refactor-goroutine-writer\01-plan-create-base-goroutine-writer-interface-and-implementation-0.md(MODIFY)

Update the WebSocket writer evaluation note:

1. **Phase 4 heading and note** (lines 238-239): Update the section to reflect that the WebSocket writer was removed:

   Change from:
   ```markdown
   ### Phase 4: WebSocket Writer Evaluation ⚠️
   **Note:** WebSocket writer has a different pattern (polls store vs receives writes). May not benefit from this base. Evaluation phase is appropriate.
   ```

   To:
   ```markdown
   ### Phase 4: WebSocket Writer Evaluation ⚠️
   **Note:** [REMOVED] The WebSocket writer was removed from the library. Its poll-based pattern was incompatible with the channelWriter base, and the functionality is better implemented client-side using `SetChannel`/`SetChannelWithBuffer` methods.
   ```

This document describes the initial plan for creating the base goroutineWriter (now channelWriter) implementation. The note about WebSocket writer evaluation should be updated to reflect the final outcome.

### docs\refactor-channel-writer\01-plan-rename-goroutinewriter-to-channelwriter-0.md(MODIFY)

Update references to WebSocket writer in this rename documentation:

1. **Observations section** (line 5): Update the sentence that mentions websocketwriter.go. Change "The `websocketwriter.go` contains an architectural comment mentioning `goroutineWriter` that should be updated for consistency." to "The WebSocket writer was removed from the library in favor of client-side implementations."

2. **How Did I Get Here section** (line 13): Remove the phrase "I also read `logger.go` and `websocketwriter.go` to understand the broader context and ensure no references were missed."

3. **Mermaid diagram** (lines 40-41): Remove the two lines:
   ```
   Dev->>Code: Update websocketwriter.go
   Note over Code: Update architectural comment
   ```

4. **websocketwriter.go file change** (lines 144-154): Remove the entire file change section for `writers\websocketwriter.go(MODIFY)`

5. **Documentation file changes** (lines 273-279): Remove the section for `docs\refactor-goroutine-writer\04-plan-evaluate-and-optionally-refactor-websocketwriter-0.md(MODIFY)` since that file is being deleted

This document describes the plan for renaming goroutineWriter to channelWriter. Since the WebSocket writer is being removed, all references to updating it should be removed from this plan.