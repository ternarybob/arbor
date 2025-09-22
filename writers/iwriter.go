package writers

import "github.com/phuslu/log"

type IWriter interface {
	WithLevel(level log.Level) IWriter
	Write(p []byte) (n int, err error)
	GetFilePath() string // Returns the file path if this is a file writer, empty string otherwise
}
