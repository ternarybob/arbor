package arbor

import (
	"github.com/ternarybob/arbor/models"
)

type ILogger interface {
	Write(p []byte) (n int, err error)

	WithPrefix(value string) ILogger

	WithCorrelationId(value string) ILogger

	WithLevel(lvl LogLevel) ILogger

	WithContext(key string, value string) ILogger

	WithFileWriter(config models.WriterConfiguration) ILogger

	// GetMemoryLogs(correlationid string, minLevel Level) (map[string]string, error)

}
