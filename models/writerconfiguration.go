package models

import (
	"io"

	"github.com/phuslu/log"
)

type LogWriterType string

const (
	LogWriterTypeConsole LogWriterType = "console"
	LogWriterTypeFile    LogWriterType = "file"
	LogWriterTypeMemory  LogWriterType = "memory"
)

type WriterConfiguration struct {
	Type       LogWriterType `json:"type"`
	Writer     io.Writer     `json:"-"`
	Level      log.Level     `json:"level"`
	TimeFormat string        `json:"timeformat"`
	FileName   string        `json:"filepath,omitempty"`
	MaxSize    int64         `json:"buffersize,omitempty"`
	MaxBackups int           `json:"maxfiles,omitempty"`
}
