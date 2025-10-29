package writers

import (
	"encoding/json"
	"sync"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

// Deprecated: ContextWriter is deprecated and will be removed in a future major version.
// The WithContextWriter method now only adds a correlation ID without creating a ContextWriter.
// Use SetChannel/SetChannelWithBuffer with correlation ID filtering in consumers instead.
//
// ContextWriter is a lightweight writer that sends log events directly to the global singleton context buffer managed by common.contextbuffer.
type ContextWriter struct {
	config    models.WriterConfiguration
	configMux sync.RWMutex
}

// NewContextWriter creates a new ContextWriter.
func NewContextWriter(config models.WriterConfiguration) IWriter {
	return &ContextWriter{
		config: config,
	}
}

// Write implements IWriter by unmarshaling JSON, filtering by log level, and sending to the singleton context buffer.
func (cw *ContextWriter) Write(p []byte) (n int, err error) {
	// Return early for empty input
	if len(p) == 0 {
		return 0, nil
	}

	// Unmarshal JSON into LogEvent
	var logEvent models.LogEvent
	if err := json.Unmarshal(p, &logEvent); err != nil {
		return 0, err
	}

	// Check log level filter
	cw.configMux.RLock()
	minLevel := cw.config.Level.ToLogLevel()
	cw.configMux.RUnlock()

	// Filter out logs below minimum level
	if logEvent.Level < minLevel {
		return len(p), nil
	}

	// Send directly to singleton context buffer
	common.Log(logEvent)

	return len(p), nil
}

// WithLevel sets the minimum log level for this writer.
func (cw *ContextWriter) WithLevel(level log.Level) IWriter {
	cw.configMux.Lock()
	cw.config.Level = levels.FromLogLevel(level)
	cw.configMux.Unlock()
	return cw
}

// GetFilePath returns an empty string (context writers don't write to files).
func (cw *ContextWriter) GetFilePath() string {
	return ""
}

// Close returns nil immediately since there are no resources to clean up.
// The singleton context buffer lifecycle is managed separately via common.Stop().
func (cw *ContextWriter) Close() error {
	return nil
}
