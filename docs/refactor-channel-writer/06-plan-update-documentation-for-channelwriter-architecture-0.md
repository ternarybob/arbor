I have created the following plan after thorough exploration and analysis of the codebase. Follow the below plan verbatim. Trust the files and references. Do not re-verify what's written in the plan. Explore only when absolutely necessary. First implement all the proposed file changes and then I'll review all the changes together at the end.

### Observations

The current state shows: (1) README.md has a basic section on channel-based streaming (lines 521-547) but lacks comprehensive documentation on SetChannel/SetChannelWithBuffer, lifecycle management, error handling, and comparison with SetContextChannel, (2) CLAUDE.md (155 lines) documents the old architecture mentioning "Memory Writer: BoltDB persistence" and "Framework Integration" but doesn't cover the simplified ContextWriter design or the new channel API, (3) WARP.md is empty (1 line) and can be left as-is, (4) The implementation in logger.go shows sophisticated features like automatic cleanup of existing writers, parameter validation with defaults, and proper lifecycle management via channelBuffers map that should be documented, (5) ContextWriter was simplified from wrapping IChannelWriter to directly calling common.Log(), reducing complexity and eliminating an unnecessary async layer.

### Approach

Update documentation to reflect the refactored architecture: (1) Enhance README.md with comprehensive SetChannel/SetChannelWithBuffer documentation including advanced patterns, lifecycle management, and comparison with SetContextChannel, (2) Update CLAUDE.md to document the simplified ContextWriter design and new channel API architecture, (3) WARP.md remains empty and requires no changes. Focus on clarity, practical examples, and helping developers understand when to use each API.

### Reasoning

I explored the repository structure, read the three documentation files (README.md, CLAUDE.md, WARP.md), examined the implementation of SetChannel/SetChannelWithBuffer in logger.go, reviewed the simplified ContextWriter in contextwriter.go, studied the ChannelWriter base implementation in channelwriter.go, and verified the ILogger interface to understand the complete API surface. I confirmed that the README already has a basic section on "Real-Time Log Streaming with Channels" (lines 521-547) that was added in a previous plan, but it needs expansion with more comprehensive documentation.

## Proposed File Changes

### README.md(MODIFY)

References: 

- logger.go
- ilogger.go
- writers\channelwriter.go
- common\channelbuffer.go

Expand and enhance the existing "Real-Time Log Streaming with Channels" section (lines 521-547) with comprehensive documentation:

1. **Add section introduction** (before line 521): Add a brief overview explaining that Arbor provides two channel-based APIs: `SetContextChannel` for singleton context logging and `SetChannel`/`SetChannelWithBuffer` for multiple independent named channels.

2. **Restructure section** (lines 521-547): Break the existing content into subsections:
   - **Overview**: Explain the difference between SetContextChannel (singleton, for job/request contexts) and SetChannel (multiple named channels, for general streaming)
   - **Basic Usage**: Keep the existing example but enhance with comments
   - **Advanced Patterns**: Add new subsection
   - **Lifecycle Management**: Add new subsection
   - **Comparison Table**: Add new subsection

3. **Add "Basic Usage" subsection** (replace lines 525-544): Enhance the existing example with:
   - Clear explanation of the `name` parameter for identifying channels
   - Example showing both SetChannel (default batching) and SetChannelWithBuffer (custom batching)
   - Comments explaining batch size (number of events before flush) and flush interval (time-based flush)
   - Example of receiving batches in a goroutine with proper error handling
   - Note about channel buffer size recommendations (10-100 depending on throughput)

4. **Add "Advanced Patterns" subsection** (after line 544): Include examples for:
   - **Multiple Independent Channels**: Show how to register multiple named channels for different purposes (e.g., "audit-logs", "metrics", "alerts")
   - **Dynamic Channel Registration**: Show how to add/remove channels at runtime using SetChannel and UnregisterChannel
   - **High-Throughput Configuration**: Example with larger batch sizes (100-500) and longer intervals (5-10 seconds) for high-volume scenarios
   - **Real-Time Configuration**: Example with small batch sizes (1-5) and short intervals (100ms-1s) for low-latency requirements
   - **WebSocket Broadcasting**: Expand the existing WebSocket example with connection management, error handling, and graceful shutdown

5. **Add "Lifecycle Management" subsection**: Document:
   - **Cleanup with UnregisterChannel**: Show how to properly stop and remove a channel logger using `logger.UnregisterChannel(name)`
   - **Automatic Cleanup on Replacement**: Explain that calling SetChannel with an existing name automatically cleans up the old writer and buffer
   - **Graceful Shutdown**: Show pattern for closing channels and waiting for final batches: `logger.UnregisterChannel(name)` followed by `close(logChannel)` and consuming remaining batches
   - **Resource Management**: Note that each named channel creates a ChannelWriter (with goroutine) and ChannelBuffer (with goroutine), so cleanup is important

6. **Add "Comparison: SetChannel vs SetContextChannel" subsection**: Create a comparison table showing:
   - **SetContextChannel**: Singleton buffer, all WithContextWriter loggers share it, use for job/request tracking, managed via common.Start/Stop
   - **SetChannel**: Multiple independent buffers, each named channel is isolated, use for general streaming/broadcasting, managed via SetChannel/UnregisterChannel
   - **Use Cases**: SetContextChannel for "all logs for job-123", SetChannel for "stream all logs to WebSocket clients" or "send errors to alerting system"

7. **Add "Error Handling and Edge Cases" subsection**: Document:
   - **Nil Channel**: Calling SetChannel with nil channel will panic with clear error message
   - **Invalid Parameters**: Zero or negative batchSize/flushInterval will use defaults (5 events, 1 second)
   - **Buffer Overflow**: If ChannelWriter buffer fills (default 1000 entries), entries are dropped with warning log
   - **Channel Blocking**: If the output channel blocks (consumer too slow), the ChannelBuffer will timeout after 5 seconds and drop the batch

8. **Update section heading** (line 521): Change from "Real-Time Log Streaming with Channels" to "Channel-Based Log Streaming" to better reflect the dual API (context channels and named channels)

9. **Add cross-references**: Add references to the "Context-Specific Logging" section (lines 294-428) for SetContextChannel documentation and to the "Async Writers with ChannelWriter" section (lines 212-293) for understanding the underlying architecture.

The goal is to provide comprehensive, practical documentation that helps developers choose the right API and use it correctly with proper lifecycle management.

### CLAUDE.md(MODIFY)

References: 

- writers\contextwriter.go
- writers\channelwriter.go
- logger.go
- common\channelbuffer.go
- common\contextbuffer.go

Update the architecture documentation to reflect the refactored design:

1. **Update "Writer System" section** (lines 74-79): Revise the writer descriptions:
   - Update Console Writer description: "Colored output using phuslu backend, synchronous writes (~50-100μs)"
   - Update File Writer description: "Rotation, backup, size management, configurable JSON/text output format, synchronous writes (~50-100μs)"
   - Update Memory Writer description: "Async writes via LogStoreWriter (ChannelWriter base), BoltDB persistence with TTL and cleanup"
   - **Add ChannelWriter description**: "Reusable async buffered writer base (1000-entry buffer, non-blocking writes, automatic drain on shutdown)"
   - **Add ContextWriter description**: "Lightweight writer that directly sends log events to singleton context buffer (no internal async layer)"
   - **Add LogStoreWriter description**: "Async writer using ChannelWriter base to write to ILogStore (in-memory + optional BoltDB)"

2. **Add "Channel-Based Logging APIs" subsection** (after line 105): Document the new channel APIs:
   - **SetContextChannel/SetContextChannelWithBuffer**: Singleton context buffer for job/request tracking, all WithContextWriter loggers share the same buffer, managed via `common.Start()` and `common.Stop()`
   - **SetChannel/SetChannelWithBuffer**: Multiple independent named channels for general streaming, each channel has its own ChannelWriter and ChannelBuffer, managed via `SetChannel()` and `UnregisterChannel()`
   - **Use Cases**: SetContextChannel for "capture all logs for job-123", SetChannel for "stream logs to WebSocket clients" or "send errors to Slack"

3. **Update "Key Design Patterns" section** (lines 96-105): Add new patterns:
   - **Simplified ContextWriter**: Direct synchronous writes to singleton context buffer (no internal ChannelWriter wrapper), reduces complexity while maintaining non-blocking behavior via singleton buffer
   - **Named Channel Writers**: Multiple independent channel loggers with per-channel batching and lifecycle management, enables flexible streaming to multiple consumers
   - **Dual Channel APIs**: Singleton context channel for job tracking vs. multiple named channels for general streaming, provides flexibility for different use cases

4. **Update "Memory Writer Architecture" section** (lines 109-113): Clarify the async write path:
   - Update line 109: "Uses date-based BoltDB files (`temp/arbor_logs_YYMMDD.db`)"
   - Update line 110: "Implements TTL with background cleanup every minute"
   - Update line 111: "Thread-safe with shared database instances"
   - Update line 112: "Correlation ID-based log storage and retrieval"
   - **Add line 113**: "Async writes via LogStoreWriter (ChannelWriter base with 1000-entry buffer)"

5. **Add "ContextWriter Architecture" subsection** (after line 113): Document the simplified design:
   - **Design**: Lightweight synchronous writer that directly calls `common.Log()` to send events to the singleton context buffer
   - **No Internal Async Layer**: Unlike the old design (which wrapped ChannelWriter), the new ContextWriter has no internal goroutine or channel
   - **Batching**: Handled by the singleton context buffer (`common.contextbuffer`), not by ContextWriter itself
   - **Lifecycle**: No cleanup needed for ContextWriter instances; singleton buffer is managed via `common.Start()` and `common.Stop()`
   - **Performance**: Write operations are fast (~10-50μs) since they only append to the singleton buffer under mutex

6. **Add "ChannelWriter Base" subsection** (after the new ContextWriter section): Document the reusable async writer:
   - **Purpose**: Reusable base for async writers (LogStoreWriter, named channel writers created by SetChannel)
   - **Architecture**: Buffered channel (default 1000 entries) + background goroutine + processor function
   - **Lifecycle**: Start/Stop methods for goroutine control, automatic buffer drain on Close()
   - **Overflow Behavior**: Non-blocking writes, drops entries with warning when buffer is full
   - **Used By**: LogStoreWriter (for memory writer), named channel writers (created by SetChannel/SetChannelWithBuffer)

7. **Update "Thread Safety" section** (lines 125-128): Add notes about the new components:
   - Add: "ChannelWriter uses RWMutex for config access and separate mutex for running state"
   - Add: "ContextWriter uses RWMutex for thread-safe level changes"
   - Add: "Named channel buffers tracked in global map with RWMutex protection"

8. **Update "Package Structure" section** (lines 142-155): Update the common package description:
   - Update line 149: "**`common/`**: Shared utilities, internal logging, singleton context buffer (`contextbuffer.go`), and per-instance channel buffer (`channelbuffer.go`)"

9. **Update "Key Constants and Configuration" section** (lines 151-155): Add new constants:
   - Add: "**Named Channel Writers**: Default batch size 5 events, default flush interval 1 second, queue size calculated as max(1000, batchSize * 100)"
   - Add: "**ChannelWriter Buffer**: Default 1000 entries, non-blocking writes, automatic drain on shutdown"

The goal is to provide clear architectural documentation that helps developers understand the refactored design, the simplified ContextWriter, and the new named channel API.