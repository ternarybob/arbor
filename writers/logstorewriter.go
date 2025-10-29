// -----------------------------------------------------------------------
// Last Modified: Wednesday, 1st October 2025 4:00:00 pm
// Modified By: Bob McAllan
// -----------------------------------------------------------------------

package writers

import (
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
	writer IGoroutineWriter
}

// LogStoreWriter creates a new writer that stores logs in the provided ILogStore
func LogStoreWriter(store ILogStore, config models.WriterConfiguration) IWriter {
	internalLog := common.NewLogger().WithContext("function", "LogStoreWriter").GetLogger()

	// Create processor function that stores log events
	processor := func(entry models.LogEvent) error {
		return store.Store(entry)
	}

	// Create and start async writer with 1000 buffer size
	writer, err := newAsyncWriter(config, 1000, processor)
	if err != nil {
		internalLog.Warn().Err(err).Msg("Failed to create async writer")
		panic("Failed to create async writer: " + err.Error())
	}

	lsw := &logStoreWriter{
		store:  store,
		config: config,
		writer: writer,
	}

	return lsw
}

// Write implements IWriter interface
func (lsw *logStoreWriter) Write(data []byte) (int, error) {
	return lsw.writer.Write(data)
}

// WithLevel sets the minimum log level for this writer
func (lsw *logStoreWriter) WithLevel(level log.Level) IWriter {
	lsw.writer.WithLevel(level)
	lsw.config.Level = levels.FromLogLevel(level)
	return lsw
}

// GetFilePath returns empty string as store doesn't write to files directly
func (lsw *logStoreWriter) GetFilePath() string {
	return lsw.writer.GetFilePath()
}

// Close shuts down the writer
func (lsw *logStoreWriter) Close() error {
	return lsw.writer.Close()
}
