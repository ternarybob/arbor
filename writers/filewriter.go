package writers

import (
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/models"
	"github.com/ternarybob/arbor/services"

	"github.com/phuslu/log"
)

const (
	MaxLogSize int64 = 10 * 1024 * 1024 // 10 MB
)

type fileWriter struct {
	logger     log.Logger
	config     models.WriterConfiguration
	ginService services.IGinService // Optional Gin formatting service
}

func FileWriter(config models.WriterConfiguration) IWriter {

	maxBackups := config.MaxBackups
	if maxBackups < 1 {
		maxBackups = 5
	}

	maxSize := config.MaxSize
	if maxSize < 1 {
		maxSize = MaxLogSize
	}

	fileName := config.FileName
	if common.IsEmpty(fileName) {
		fileName = "logs/main.log"
	}

	fw := &consoleWriter{
		logger: log.Logger{
			Level:      config.Level,
			TimeFormat: config.TimeFormat,
			Writer: &log.FileWriter{
				Filename:     fileName,
				MaxSize:      maxSize,
				MaxBackups:   maxBackups,
				EnsureFolder: true,
				LocalTime:    true,
			},
		},
		config: config,
	}

	return fw
}

// FileLogEntry represents a parsed log entry for file writing
type FileLogEntry struct {
	Level   string      `json:"level"`
	Message string      `json:"message"`
	Time    string      `json:"time,omitempty"`
	Prefix  string      `json:"prefix,omitempty"`
	Extra   interface{} `json:"-"`
}

func (fw *fileWriter) WithLevel(level log.Level) IWriter {

	fw.logger.SetLevel(level)

	return fw
}

func (fw *fileWriter) Write(e []byte) (n int, err error) {
	n = len(e)
	if n <= 0 {
		return n, err
	}

	return n, nil
}
