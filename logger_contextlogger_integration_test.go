package arbor_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ternarybob/arbor"
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/models"
)

func TestContextLogger_Integration(t *testing.T) {
	// This test validates the simplified ContextWriter which now directly calls common.Log()
	// instead of wrapping a ChannelWriter. The external behavior remains the same - logs are
	// batched and sent to the channel via the singleton context buffer.

	// 1. Create a channel to receive log batches.
	logChan := make(chan []models.LogEvent, 10)

	// 2. Configure the context logger with a small batch size and interval for testing.
	batchSize := 3
	flushInterval := 100 * time.Millisecond
	arbor.Logger().SetContextChannelWithBuffer(logChan, batchSize, flushInterval)
	defer common.Stop()

	// 3. Create a consumer that listens on the channel.
	var receivedLogs [][]models.LogEvent
	var wg sync.WaitGroup
	consumerStop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case batch := <-logChan:
				receivedLogs = append(receivedLogs, batch)
			case <-consumerStop:
				return
			}
		}
	}()

	// 4. Create a context logger.
	contextID := "job-123"
	contextLogger := arbor.Logger().WithContextWriter(contextID)

	// 5. Log enough messages to trigger a batch flush.
	contextLogger.Info().Msg("Message 1")
	contextLogger.Info().Msg("Message 2")
	contextLogger.Info().Msg("Message 3") // This should trigger a flush
	contextLogger.Warn().Msg("Message 4")

	// 6. Wait for processing.
	time.Sleep(2 * flushInterval) // Wait for more than the flush interval

	// 7. Stop the consumer and verify results.
	close(consumerStop)
	wg.Wait()

	// This test validates that the simplified ContextWriter (which directly calls common.Log())
	// correctly integrates with the singleton context buffer to deliver batched logs to the channel.
	require.GreaterOrEqual(t, len(receivedLogs), 1, "Should have received at least one batch")

	// 8. Flatten the received batches and verify content.
	var allLogs []models.LogEvent
	for _, batch := range receivedLogs {
		allLogs = append(allLogs, batch...)
	}

	assert.GreaterOrEqual(t, len(allLogs), 4, "Should have at least 4 log messages in total")

	// 9. Verify the correlation ID.
	for _, logEvent := range allLogs {
		assert.Equal(t, contextID, logEvent.CorrelationID, "Log event should have the correct context ID")
	}
}
