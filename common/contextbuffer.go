package common

import (
	"sync"
	"time"

	"github.com/ternarybob/arbor/models"
)

var (
	buffer        []models.LogEvent
	bufferMux     sync.Mutex
	outputChan    chan []models.LogEvent
	batchSize     int
	flushInterval time.Duration
	stopChan      chan struct{}
	once          sync.Once
)

// Start initializes and starts the context log buffer.
// It should only be called once.
func Start(out chan []models.LogEvent, size int, interval time.Duration) {
	once.Do(func() {
		outputChan = out
		batchSize = size
		flushInterval = interval
		buffer = make([]models.LogEvent, 0, batchSize)
		stopChan = make(chan struct{})
		go run()
	})
}

// Stop signals the buffer to flush any remaining logs and stop.
func Stop() {
	if stopChan != nil {
		close(stopChan)
	}
}

// Log adds a log event to the buffer.
func Log(event models.LogEvent) {
	bufferMux.Lock()
	defer bufferMux.Unlock()

	buffer = append(buffer, event)
	if len(buffer) >= batchSize {
		flush()
	}
}

func run() {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bufferMux.Lock()
			flush()
			bufferMux.Unlock()
		case <-stopChan:
			bufferMux.Lock()
			flush()
			bufferMux.Unlock()
			return
		}
	}
}

func flush() {
	if len(buffer) > 0 {
		// Create a copy of the buffer to send
		logBatch := make([]models.LogEvent, len(buffer))
		copy(logBatch, buffer)

		// Send the batch to the output channel without blocking the buffer mutex
		go func() {
			select {
			case outputChan <- logBatch:
				// Sent successfully
			case <-time.After(1 * time.Second): // Timeout to prevent blocking forever
				// Log this issue to the internal logger if you have one
			}
		}()

		// Clear the buffer
		buffer = make([]models.LogEvent, 0, batchSize)
	}
}
