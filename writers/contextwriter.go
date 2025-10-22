package writers

import (
	"encoding/json"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/models"
)

// ContextWriter is a writer that sends log events to the global context buffer.
type ContextWriter struct{}

// NewContextWriter creates a new ContextWriter.
func NewContextWriter() IWriter {
	return &ContextWriter{}
}

// Write unmarshals the log event and sends it to the context buffer.
func (cw *ContextWriter) Write(p []byte) (n int, err error) {
	var logEvent models.LogEvent
	if err := json.Unmarshal(p, &logEvent); err != nil {
		return 0, err
	}

	common.Log(logEvent)
	return len(p), nil
}

// WithLevel is a no-op for the ContextWriter.
func (cw *ContextWriter) WithLevel(level log.Level) IWriter {
	return cw
}

// GetFilePath returns an empty string.
func (cw *ContextWriter) GetFilePath() string {
	return ""
}

// Close is a no-op for the ContextWriter.
func (cw *ContextWriter) Close() error {
	return nil
}
