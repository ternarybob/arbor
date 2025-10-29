package arbor_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ternarybob/arbor"
	"github.com/ternarybob/arbor/models"
)

func TestContextLogger_Integration(t *testing.T) {
	// This test validates context-specific logging using the unified SetChannel API with correlation ID filtering.
	// WithContextWriter now only adds a correlation ID, and consumers filter batches by correlation ID.

	// 1. Create a channel to receive log batches.
	logChan := make(chan []models.LogEvent, 10)

	// 2. Configure a named channel logger with a small batch size and interval for testing.
	batchSize := 3
	flushInterval := 100 * time.Millisecond
	channelName := "test-context-channel"
	arbor.Logger().SetChannelWithBuffer(channelName, logChan, batchSize, flushInterval)
	defer arbor.Logger().UnregisterChannel(channelName)

	// 3. Create a consumer that listens on the channel and filters by correlation ID.
	contextID := "job-123"
	var receivedLogs [][]models.LogEvent
	var wg sync.WaitGroup
	consumerStop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case batch := <-logChan:
				// Filter batch to only include logs with the target correlation ID
				filteredBatch := make([]models.LogEvent, 0)
				for _, event := range batch {
					if event.CorrelationID == contextID {
						filteredBatch = append(filteredBatch, event)
					}
				}
				if len(filteredBatch) > 0 {
					receivedLogs = append(receivedLogs, filteredBatch)
				}
			case <-consumerStop:
				return
			}
		}
	}()

	// 4. Create a context logger (now only adds correlation ID, no separate writer).
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

	// This test validates that WithContextWriter (which now only adds correlation ID)
	// correctly tags logs, and the unified SetChannel API delivers them to consumers
	// who can filter by correlation ID.
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
