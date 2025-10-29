package common

import (
	"sync"
	"time"

	"github.com/ternarybob/arbor/models"
)

// ChannelBuffer provides per-instance batching for log events sent to a channel.
// Unlike contextbuffer.go which is a singleton, this allows multiple independent buffers.
type ChannelBuffer struct {
	buffer        []models.LogEvent
	bufferMux     sync.Mutex
	outputChan    chan []models.LogEvent
	batchSize     int
	flushInterval time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewChannelBuffer creates and starts a new channel buffer instance.
// Returns nil if the output channel is nil.
func NewChannelBuffer(out chan []models.LogEvent, size int, interval time.Duration) *ChannelBuffer {
	// Defensive nil check for output channel
	if out == nil {
		return nil
	}

	// Validate and set defaults for size
	if size <= 0 {
		size = 5 // Default batch size
	}

	// Validate and set defaults for interval
	if interval <= 0 {
		interval = 1 * time.Second // Default flush interval
	}

	cb := &ChannelBuffer{
		outputChan:    out,
		batchSize:     size,
		flushInterval: interval,
		buffer:        make([]models.LogEvent, 0, size),
		stopChan:      make(chan struct{}),
	}
	go cb.run()
	return cb
}

// Stop signals the buffer to flush any remaining logs and stop.
func (cb *ChannelBuffer) Stop() {
	close(cb.stopChan)
	cb.wg.Wait()
}

// Log adds a log event to the buffer.
func (cb *ChannelBuffer) Log(event models.LogEvent) {
	cb.bufferMux.Lock()
	defer cb.bufferMux.Unlock()

	cb.buffer = append(cb.buffer, event)
	if len(cb.buffer) >= cb.batchSize {
		cb.flush()
	}
}

func (cb *ChannelBuffer) run() {
	cb.wg.Add(1)
	defer cb.wg.Done()

	ticker := time.NewTicker(cb.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cb.bufferMux.Lock()
			cb.flush()
			cb.bufferMux.Unlock()
		case <-cb.stopChan:
			cb.bufferMux.Lock()
			cb.flush()
			cb.bufferMux.Unlock()
			return
		}
	}
}

func (cb *ChannelBuffer) flush() {
	if len(cb.buffer) > 0 {
		// Create a copy of the buffer to send
		logBatch := make([]models.LogEvent, len(cb.buffer))
		copy(logBatch, cb.buffer)

		// Send the batch to the output channel without blocking the buffer mutex
		go func() {
			select {
			case cb.outputChan <- logBatch:
				// Sent successfully
			case <-time.After(1 * time.Second): // Timeout to prevent blocking forever
				// Log this issue to the internal logger if you have one
			}
		}()

		// Clear the buffer
		cb.buffer = make([]models.LogEvent, 0, cb.batchSize)
	}
}
