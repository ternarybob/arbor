// -----------------------------------------------------------------------
// Last Modified: Wednesday, 1st October 2025 4:00:00 pm
// Modified By: Bob McAllan
// -----------------------------------------------------------------------

package writers

import (
	"encoding/json"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

// logStoreWriter bridges the IWriter interface to ILogStore
// It buffers writes asynchronously to prevent blocking the logging path
type logStoreWriter struct {
	store  ILogStore
	config models.WriterConfiguration
	buffer chan models.LogEvent
	done   chan bool
}

// LogStoreWriter creates a new writer that stores logs in the provided ILogStore
func LogStoreWriter(store ILogStore, config models.WriterConfiguration) IWriter {
	lsw := &logStoreWriter{
		store:  store,
		config: config,
		buffer: make(chan models.LogEvent, 1000), // Buffered to prevent blocking
		done:   make(chan bool),
	}

	// Start async processor
	go lsw.processBuffer()

	return lsw
}

// processBuffer handles async writes to the store
func (lsw *logStoreWriter) processBuffer() {
	internalLog := common.NewLogger().WithContext("function", "LogStoreWriter.processBuffer").GetLogger()

	for {
		select {
		case entry := <-lsw.buffer:
			if err := lsw.store.Store(entry); err != nil {
				internalLog.Warn().Err(err).Msg("Failed to store log entry")
			}
		case <-lsw.done:
			// Drain remaining buffer
			for len(lsw.buffer) > 0 {
				entry := <-lsw.buffer
				lsw.store.Store(entry)
			}
			return
		}
	}
}

// Write implements IWriter interface
func (lsw *logStoreWriter) Write(data []byte) (int, error) {
	internalLog := common.NewLogger().WithContext("function", "LogStoreWriter.Write").GetLogger()

	n := len(data)
	if n == 0 {
		return n, nil
	}

	// Parse log event
	var logEvent models.LogEvent
	if err := json.Unmarshal(data, &logEvent); err != nil {
		internalLog.Warn().Err(err).Msg("Failed to unmarshal log event")
		return 0, err
	}

	// Check level filtering
	if logEvent.Level < lsw.config.Level.ToLogLevel() {
		return n, nil // Skip this log entry
	}

	// Non-blocking write to buffer
	select {
	case lsw.buffer <- logEvent:
		return n, nil
	default:
		// Buffer full - log warning and drop
		internalLog.Warn().Msg("Log store buffer full, dropping entry")
		return n, nil
	}
}

// WithLevel sets the minimum log level for this writer
func (lsw *logStoreWriter) WithLevel(level log.Level) IWriter {
	lsw.config.Level = levels.FromLogLevel(level)
	return lsw
}

// GetFilePath returns empty string as store doesn't write to files directly
func (lsw *logStoreWriter) GetFilePath() string {
	return ""
}

// Close shuts down the writer
func (lsw *logStoreWriter) Close() error {
	close(lsw.done)
	// Wait briefly for buffer to drain
	// Note: In production, consider using WaitGroup for proper synchronization
	return nil
}
