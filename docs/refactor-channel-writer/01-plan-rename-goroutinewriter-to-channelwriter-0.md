I have created the following plan after thorough exploration and analysis of the codebase. Follow the below plan verbatim. Trust the files and references. Do not re-verify what's written in the plan. Explore only when absolutely necessary. First implement all the proposed file changes and then I'll review all the changes together at the end.

### Observations

The task involves renaming `goroutineWriter` to `channelWriter` across the codebase. Through grep searches, I identified all affected files: 4 Go source files (`goroutinewriter.go`, `igoroutinewriter.go`, `contextwriter.go`, `logstorewriter.go`), 1 test file (`goroutinewriter_test.go`), 1 README, and 6 documentation files in `docs/refactor-goroutine-writer/`. The renaming includes the struct type (`goroutineWriter` → `channelWriter`), interface (`IGoroutineWriter` → `IChannelWriter`), factory function (`NewGoroutineWriter` → `NewChannelWriter`), and helper function (`newAsyncWriter` comment updates). The `websocketwriter.go` contains an architectural comment mentioning `goroutineWriter` that should be updated for consistency.

### Approach

Perform a systematic rename operation across all identified files. Start with the core files (rename the actual files first), then update type and interface definitions, followed by updating all references in dependent files. Update test file names and content, then documentation files. This approach ensures a clean, complete refactoring with no broken references.

### Reasoning

I listed the repository structure, read the four core files mentioned by the user (`goroutinewriter.go`, `igoroutinewriter.go`, `contextwriter.go`, `logstorewriter.go`), performed grep searches to find all references to `IGoroutineWriter`, `goroutineWriter`, and `NewGoroutineWriter` across Go files, markdown files, and test files. I also read `logger.go` and `websocketwriter.go` to understand the broader context and ensure no references were missed.

## Mermaid Diagram

sequenceDiagram
    participant Dev as Developer
    participant Files as File System
    participant Code as Source Code
    participant Tests as Test Suite
    participant Docs as Documentation

    Dev->>Files: Rename goroutinewriter.go → channelwriter.go
    Dev->>Files: Rename igoroutinewriter.go → ichannelwriter.go
    Dev->>Files: Rename goroutinewriter_test.go → channelwriter_test.go
    
    Dev->>Code: Update channelwriter.go
    Note over Code: Rename struct, functions,<br/>method receivers, comments
    
    Dev->>Code: Update ichannelwriter.go
    Note over Code: Rename interface type
    
    Dev->>Code: Update contextwriter.go
    Note over Code: Update field type reference
    
    Dev->>Code: Update logstorewriter.go
    Note over Code: Update field type reference
    
    Dev->>Code: Update websocketwriter.go
    Note over Code: Update architectural comment
    
    Dev->>Tests: Update channelwriter_test.go
    Note over Tests: Rename test functions,<br/>update function calls
    
    Dev->>Docs: Update README.md
    Note over Docs: Update sections, examples,<br/>code snippets
    
    Dev->>Docs: Update 6 refactor docs
    Note over Docs: Update file references,<br/>type names, diagrams
    
    Dev->>Dev: Verify all references updated
    Dev->>Dev: Run tests to confirm

## Proposed File Changes

### writers\goroutinewriter.go → writers\channelwriter.go

Rename the file from `goroutinewriter.go` to `channelwriter.go` to reflect the new naming convention.

### writers\channelwriter.go(NEW)

References: 

- writers\ichannelwriter.go(NEW)
- models\writerconfiguration.go

Update all internal references within the file:

1. **Struct type** (line 14): Rename `type goroutineWriter struct` to `type channelWriter struct`

2. **Factory function** (line 27): Rename `func NewGoroutineWriter(...)` to `func NewChannelWriter(...)`

3. **Return type** (line 27): Update return type from `(IGoroutineWriter, error)` to `(IChannelWriter, error)`

4. **Struct instantiation** (line 36): Update `return &goroutineWriter{...}` to `return &channelWriter{...}`

5. **Helper function comment** (line 46): Update comment from "creates and starts a goroutine writer" to "creates and starts a channel writer"

6. **Helper function** (line 49): Update return type from `(IGoroutineWriter, error)` to `(IChannelWriter, error)`

7. **Helper function body** (line 51): Update comment from "Create goroutine writer" to "Create channel writer", and update function call from `NewGoroutineWriter(...)` to `NewChannelWriter(...)`

8. **Method receivers**: Update all method receivers from `(gw *goroutineWriter)` to `(cw *channelWriter)` for methods: `Start()`, `Stop()`, `processBuffer()`, `Write()`, `WithLevel()`, `GetFilePath()`, `Close()`, and `IsRunning()`

9. **Internal variable references**: Within each method, update references from `gw` to `cw` (e.g., `gw.runningMux` → `cw.runningMux`, `gw.buffer` → `cw.buffer`, etc.)

10. **Internal logger context** (line 98): Update the context string from `"goroutineWriter.processBuffer"` to `"channelWriter.processBuffer"`

11. **Internal logger context** (line 126): Update the context string from `"goroutineWriter.Write"` to `"channelWriter.Write"`

12. **Warning message** (line 150): Update the message from "Goroutine writer buffer full" to "Channel writer buffer full"

### writers\igoroutinewriter.go → writers\ichannelwriter.go

Rename the file from `igoroutinewriter.go` to `ichannelwriter.go` to reflect the new interface naming convention.

### writers\ichannelwriter.go(NEW)

References: 

- writers\iwriter.go

Update the interface definition:

1. **Interface name** (line 3): Rename `type IGoroutineWriter interface` to `type IChannelWriter interface`

The interface methods (`Start()`, `Stop()`, `IsRunning()`) and the embedded `IWriter` interface remain unchanged.

### writers\contextwriter.go(MODIFY)

References: 

- writers\ichannelwriter.go(NEW)
- writers\channelwriter.go(NEW)
- common\common.go

Update all references to the renamed interface and functions:

1. **Struct field** (line 12): Update field type from `writer IGoroutineWriter` to `writer IChannelWriter`

2. **Constructor comment** (line 27): Update comment from "Create and start async writer" to reflect that it creates a channel writer (optional, for clarity)

3. **Function call** (line 28): Update the call from `newAsyncWriter(...)` to use the new naming - the function itself doesn't need renaming but its internal implementation now uses `NewChannelWriter`

Note: The `newAsyncWriter` helper function is defined in `channelwriter.go` and will be updated there. This file only needs to update the type reference for the `writer` field.

### writers\logstorewriter.go(MODIFY)

References: 

- writers\ichannelwriter.go(NEW)
- writers\channelwriter.go(NEW)
- writers\ilogstore.go

Update all references to the renamed interface:

1. **Struct field** (line 20): Update field type from `writer IGoroutineWriter` to `writer IChannelWriter`

2. **Constructor comment** (line 32): Update comment from "Create and start async writer" to reflect channel writer terminology (optional, for clarity)

Note: Similar to `contextwriter.go`, the `newAsyncWriter` helper function call doesn't need to change here - it's defined in `channelwriter.go` and will be updated there.

### writers\websocketwriter.go(MODIFY)

References: 

- writers\channelwriter.go(NEW)

Update the architectural comment for consistency:

1. **Comment** (line 34-38): Update the comment from "with the write-driven goroutineWriter base" to "with the write-driven channelWriter base"

This is a documentation-only change to maintain consistency with the new naming convention. The comment explains why `websocketWriter` doesn't use the `channelWriter` base class.

### writers\goroutinewriter_test.go → writers\channelwriter_test.go

Rename the test file from `goroutinewriter_test.go` to `channelwriter_test.go` to match the renamed source file.

### writers\channelwriter_test.go(NEW)

References: 

- writers\channelwriter.go(NEW)
- writers\ichannelwriter.go(NEW)
- models\logevent.go

Update all test function names, comments, and references throughout the file:

1. **Section comment** (line 70): Update from "SECTION 1: UNIT TESTS FOR goroutineWriter BASE" to "SECTION 1: UNIT TESTS FOR channelWriter BASE"

2. **Test function names**: Rename all test functions from `TestGoroutineWriter_*` to `TestChannelWriter_*`. This includes:
   - `TestGoroutineWriter_NewWithValidProcessor` → `TestChannelWriter_NewWithValidProcessor`
   - `TestGoroutineWriter_NewWithNilProcessor` → `TestChannelWriter_NewWithNilProcessor`
   - `TestGoroutineWriter_StartStop` → `TestChannelWriter_StartStop`
   - `TestGoroutineWriter_ProcessBuffer` → `TestChannelWriter_ProcessBuffer`
   - `TestGoroutineWriter_BufferOverflow` → `TestChannelWriter_BufferOverflow`
   - `TestGoroutineWriter_GracefulShutdown` → `TestChannelWriter_GracefulShutdown`
   - `TestGoroutineWriter_ConcurrentWrites` → `TestChannelWriter_ConcurrentWrites`
   - `TestGoroutineWriter_ErrorHandling` → `TestChannelWriter_ErrorHandling`
   - And all other test functions following this pattern

3. **Function calls**: Update all calls from `NewGoroutineWriter(...)` to `NewChannelWriter(...)` (appears on lines 77, 101, 116, 122, 132, 174, 225, 253, 288, 323, 350, 388, 425, 473, 517, 554, 598, 644, 670, 687, 1035, 1078, 1144)

4. **Interface type assertions**: Update type assertions from `IGoroutineWriter` to `IChannelWriter` (line 87, 89)

5. **Comments**: Update all comments that reference "goroutine writer" to "channel writer" throughout the test file

6. **Error messages**: Update error messages that mention "goroutine writer" to "channel writer"

The test logic and structure remain unchanged - only naming references need updating.

### README.md(MODIFY)

Update all documentation references to reflect the new naming:

1. **Section heading** (line 214): Update from "## Async Writers with GoroutineWriter" to "## Async Writers with ChannelWriter"

2. **Subsection heading** (line 218): Update from "### What is GoroutineWriter?" to "### What is ChannelWriter?"

3. **Description** (line 220): Update from "The `goroutineWriter` is a reusable async writer" to "The `channelWriter` is a reusable async writer"

4. **Section heading** (line 864): Update from "### Custom Async Writers (GoroutineWriter)" to "### Custom Async Writers (ChannelWriter)"

5. **Description** (line 866): Update from "using the `goroutineWriter` base" to "using the `channelWriter` base"

6. **Struct field** (line 886): Update from `writer writers.IGoroutineWriter` to `writer writers.IChannelWriter`

7. **Comment** (line 892): Update from "// Create goroutine writer with 1000 buffer size" to "// Create channel writer with 1000 buffer size"

8. **Function call** (line 900): Update from `writers.NewGoroutineWriter(config, 1000, processor)` to `writers.NewChannelWriter(config, 1000, processor)`

9. **Function call** (line 948): Update from `writers.NewGoroutineWriter(config, 1000, processor)` to `writers.NewChannelWriter(config, 1000, processor)`

10. **Function call** (line 1023): Update from `writers.NewGoroutineWriter(config, 1000, slowProcessor)` to `writers.NewChannelWriter(config, 1000, slowProcessor)`

Ensure all code examples and explanatory text consistently use the new "channel writer" terminology instead of "goroutine writer".

### docs\refactor-goroutine-writer\01-plan-create-base-goroutine-writer-interface-and-implementation-0.md(MODIFY)

Update all references to reflect the new naming convention:

1. **Title** (line 7): Update from "Base Goroutine Writer" to "Base Channel Writer"

2. **Overview** (line 10): Update from "The base `goroutineWriter` implementation" to "The base `channelWriter` implementation" and update file references from `goroutinewriter.go` and `igoroutinewriter.go` to `channelwriter.go` and `ichannelwriter.go`

3. **Section heading** (line 16): Update from "Interface Design (`igoroutinewriter.go`)" to "Interface Design (`ichannelwriter.go`)"

4. **Description** (line 19): Update from "The `IGoroutineWriter` interface" to "The `IChannelWriter` interface"

5. **Section heading** (line 26): Update from "Base Implementation (`goroutinewriter.go`)" to "Base Implementation (`channelwriter.go`)"

Update all other occurrences of "goroutineWriter", "IGoroutineWriter", "goroutine writer", and file name references throughout the document to use the new "channel" terminology.

### docs\refactor-goroutine-writer\02-plan-refactor-logstorewriter-to-use-base-goroutinewriter-0.md(MODIFY)

Update all references to reflect the new naming convention:

1. **Approach** (line 9): Update from "Refactor `logStoreWriter` to compose `goroutineWriter`" to "Refactor `logStoreWriter` to compose `channelWriter`" and update from "The struct will hold an `IGoroutineWriter` instance" to "The struct will hold an `IChannelWriter` instance"

2. **Reasoning** (line 13): Update file references from `goroutineWriter` to `channelWriter` and from `IGoroutineWriter` to `IChannelWriter`

3. **Mermaid diagram** (line 26-30): Update participant names from `GW` (GoroutineWriter) to `CW` (ChannelWriter), update function call from `NewGoroutineWriter` to `NewChannelWriter`, and update return type from `IGoroutineWriter` to `IChannelWriter`

4. **References** (line 61-62): Update file references from `goroutinewriter.go` and `igoroutinewriter.go` to `channelwriter.go` and `ichannelwriter.go`

5. **Struct changes** (line 67-68): Update comments from "now handled by `goroutineWriter`" to "now handled by `channelWriter`"

6. **Field addition** (line 74): Update from "`writer IGoroutineWriter` field" to "`writer IChannelWriter` field" and update comment from "composed goroutine writer" to "composed channel writer"

7. **Constructor** (line 80-81): Update from "`goroutineWriter` expectations" to "`channelWriter` expectations" and update function call from `NewGoroutineWriter` to `NewChannelWriter`

Update all other occurrences throughout the document.

### docs\refactor-goroutine-writer\03-plan-refactor-contextwriter-to-use-base-goroutinewriter-0.md(MODIFY)

Update all references to reflect the new naming convention:

1. **Observations** (line 5): Update from "compose `IGoroutineWriter`" to "compose `IChannelWriter`"

2. **Approach** (line 9): Update from "Refactor `ContextWriter` to compose `goroutineWriter`" to "Refactor `ContextWriter` to compose `channelWriter`" and update from "hold an `IGoroutineWriter` field" to "hold an `IChannelWriter` field"

3. **Reasoning** (line 13): Update file references from `goroutinewriter.go` to `channelwriter.go`

4. **References** (line 21-22): Update file references from `goroutinewriter.go` and `igoroutinewriter.go` to `channelwriter.go` and `ichannelwriter.go`

5. **Struct changes** (line 33): Update from "`writer IGoroutineWriter` field" to "`writer IChannelWriter` field" and update comment from "composed goroutine writer" to "composed channel writer"

6. **Constructor** (line 46-49): Update function call from `NewGoroutineWriter` to `NewChannelWriter` and update error handling comments from "NewGoroutineWriter" to "NewChannelWriter"

Update all other occurrences of "goroutineWriter", "IGoroutineWriter", and related terminology throughout the document.

### docs\refactor-goroutine-writer\04-plan-evaluate-and-optionally-refactor-websocketwriter-0.md(MODIFY)

Update all references to reflect the new naming convention:

1. **References** (line 21): Update file reference from `goroutinewriter.go` to `channelwriter.go`

Update any other occurrences of "goroutineWriter" or "IGoroutineWriter" in the document to use "channelWriter" and "IChannelWriter" respectively. This document discusses why WebSocketWriter doesn't use the base writer pattern, so ensure the architectural discussion references the correct new names.

### docs\refactor-goroutine-writer\05-plan-add-comprehensive-tests-for-goroutinewriter-0.md(MODIFY)

Update all references to reflect the new naming convention:

1. **Observations** (line 5): Update from "The goroutineWriter base implementation" to "The channelWriter base implementation" and update file reference from `writers/goroutinewriter.go` to `writers/channelwriter.go`

2. **Approach** (line 9): Update from "Create comprehensive test coverage in `writers/goroutinewriter_test.go`" to "Create comprehensive test coverage in `writers/channelwriter_test.go`" and update from "Unit tests for goroutineWriter base" to "Unit tests for channelWriter base"

3. **Reasoning** (line 13): Update from "read the goroutineWriter base implementation" to "read the channelWriter base implementation" and update interface reference from "IGoroutineWriter" to "IChannelWriter"

4. **File path** (line 17): Update from `writers\goroutinewriter_test.go(NEW)` to `writers\channelwriter_test.go(NEW)`

5. **References** (line 21-22): Update file references from `goroutinewriter.go` and `igoroutinewriter.go` to `channelwriter.go` and `ichannelwriter.go`

6. **Section comment** (line 27): Update from "SECTION 1: UNIT TESTS FOR goroutineWriter BASE" to "SECTION 1: UNIT TESTS FOR channelWriter BASE"

7. **Test names** (line 50-56): Update all test function names from `TestGoroutineWriter_*` to `TestChannelWriter_*` and update function calls from `NewGoroutineWriter` to `NewChannelWriter`

Update all other occurrences throughout the document.

### docs\refactor-goroutine-writer\06-plan-review-and-validate-goroutinewriter-refactoring-0.md(MODIFY)

Update all references to reflect the new naming convention:

1. **References** (line 21): Update file reference from `writers\goroutinewriter.go` to `writers\channelwriter.go`

Update any other occurrences of "goroutineWriter", "IGoroutineWriter", or file name references throughout the document to use the new "channel" terminology. This is a review document, so ensure all references to the refactored components use the updated names.