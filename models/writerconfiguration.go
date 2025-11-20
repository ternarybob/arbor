package models

import (
	"io"

	"github.com/ternarybob/arbor/levels"
)

type LogWriterType string

const (
	LogWriterTypeConsole LogWriterType = "console"
	LogWriterTypeFile    LogWriterType = "file"
	LogWriterTypeMemory  LogWriterType = "memory"
)

// TextOutputFormat defines the format used for file writer output.
// "logfmt" is the default AI-friendly text format; "json" preserves structured JSON output.
type TextOutputFormat string

const (
	TextOutputFormatJSON   TextOutputFormat = "json"
	TextOutputFormatLogfmt TextOutputFormat = "logfmt"
)

type WriterConfiguration struct {
	Type             LogWriterType    `json:"type"`
	Writer           io.Writer        `json:"-"`
	Level            levels.LogLevel  `json:"level"`
	TimeFormat       string           `json:"timeformat"`
	FileName         string           `json:"filepath,omitempty"`
	LogNameFormat    string           `json:"lognameformat,omitempty"`
	MaxSize          int64            `json:"buffersize,omitempty"`
	MaxBackups       int              `json:"maxfiles,omitempty"`
	DisableTimestamp bool             `json:"disabletimestamp,omitempty"`
	TextOutput       TextOutputFormat `json:"textoutput,omitempty"`
	DBPath           string           `json:"dbpath,omitempty"`
}
