package arbor_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ternarybob/arbor"
	"github.com/ternarybob/arbor/models"
)

func TestWithFunctionLogger_Integration(t *testing.T) {
	// 1. Setup a global memory writer to act as a standard output.
	globalConfig := models.WriterConfiguration{Level: arbor.TraceLevel}
	arbor.Logger().WithMemoryWriter(globalConfig)

	// 2. Get the base logger.
	baseLogger := arbor.Logger()

	// 3. Create a function logger.
	funcLogger, extractor, cleanup, err := baseLogger.WithFunctionLogger("func-123", models.WriterConfiguration{Level: arbor.TraceLevel})
	require.NoError(t, err)
	require.NotNil(t, funcLogger)
	require.NotNil(t, extractor)
	require.NotNil(t, cleanup)

	// 4. Log messages to both loggers.
	baseLogger.Info().Str("logger", "global").Msg("This is a global message.")
	funcLogger.Info().Str("logger", "function").Msg("This is a function message.")

	time.Sleep(100 * time.Millisecond) // Allow time for async writes

	// 5. Verify the function logger's private store.
	funcLogs, err := extractor()
	require.NoError(t, err)
	assert.Len(t, funcLogs, 1, "Function logger's store should only contain its own message")
	for _, log := range funcLogs {
		assert.Contains(t, log, "This is a function message")
		assert.NotContains(t, log, "This is a global message")
	}

	// 6. Verify the global memory writer received BOTH messages.
	globalMemoryWriter := arbor.GetRegisteredMemoryWriter("memory")
	require.NotNil(t, globalMemoryWriter)

	globalLogs, err := globalMemoryWriter.GetAllEntries()
	require.NoError(t, err)

	var allGlobalLogs string
	for _, log := range globalLogs {
		allGlobalLogs += log + "\n"
	}
	assert.Contains(t, allGlobalLogs, "This is a global message")
	assert.Contains(t, allGlobalLogs, "This is a function message")

	// 7. Cleanup
	err = cleanup()
	assert.NoError(t, err)
}

func TestWithFunctionLogger_DuplicateIDError(t *testing.T) {
	baseLogger := arbor.Logger()
	correlationID := "duplicate-test-id"

	// 1. First call should succeed.
	_, _, cleanup1, err1 := baseLogger.WithFunctionLogger(correlationID, models.WriterConfiguration{})
	require.NoError(t, err1)
	require.NotNil(t, cleanup1, "First call should succeed")

	// 2. Second call with the same ID should fail.
	_, _, _, err2 := baseLogger.WithFunctionLogger(correlationID, models.WriterConfiguration{})
	assert.Error(t, err2, "Second call with same ID should fail")

	// 3. Cleanup the first logger.
	cleanup1()

	// 4. Third call with the same ID should now succeed again.
	_, _, cleanup2, err3 := baseLogger.WithFunctionLogger(correlationID, models.WriterConfiguration{})
	assert.NoError(t, err3)
	assert.NotNil(t, cleanup2, "Third call should succeed after cleanup")
	defer cleanup2()
}
