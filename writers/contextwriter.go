package writers

import (
	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

// ContextWriter is a writer that sends log events to the global context buffer.
type ContextWriter struct {
	writer IGoroutineWriter
	config models.WriterConfiguration
}

// NewContextWriter creates a new ContextWriter.
func NewContextWriter(config models.WriterConfiguration) IWriter {
	// Create internal logger for error reporting
	internalLog := common.NewLogger().WithContext("function", "NewContextWriter").GetLogger()

	// Define processor closure that calls common.Log()
	processor := func(entry models.LogEvent) error {
		common.Log(entry)
		return nil
	}

	// Create and start async writer with 1000 buffer size
	writer, err := newAsyncWriter(config, 1000, processor)
	if err != nil {
		internalLog.Fatal().Err(err).Msg("Failed to create async writer")
		panic("Failed to create async writer: " + err.Error())
	}

	// Construct ContextWriter struct
	return &ContextWriter{
		writer: writer,
		config: config,
	}
}

// Write delegates to the composed goroutine writer.
func (cw *ContextWriter) Write(p []byte) (n int, err error) {
	return cw.writer.Write(p)
}

// WithLevel delegates to the composed goroutine writer.
func (cw *ContextWriter) WithLevel(level log.Level) IWriter {
	cw.writer.WithLevel(level)
	cw.config.Level = levels.FromLogLevel(level)
	return cw
}

// GetFilePath returns an empty string (context writers don't write to files).
func (cw *ContextWriter) GetFilePath() string {
	return cw.writer.GetFilePath()
}

// Close stops the goroutine and drains the buffer.
func (cw *ContextWriter) Close() error {
	return cw.writer.Close()
}
